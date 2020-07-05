package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/joho/godotenv"
)

const editpage = `<html>
    <head>
        <script src="/static/main.js" type="module"></script>
    </head>
    <body>
        <center>
            <textarea id="textbox" name="textbox" rows="45" cols="150" disabled></textarea>
        </center>
    </body>
</html>`

const (
	UpdateInterval = 250 * time.Millisecond
	PruneInterval  = 30 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type DocMeta struct {
	Doc Doc

	Log         [][]Op        // one update is a collection of ops from one diff
	NextSeq     map[int64]int // expected next seq from user
	AppliedSeqs map[int64]int // all seqs up to this from user have been applied
	UserViews   map[int64]int // user reported to have seen this
	NextDiscard int
	DocId       int64

	mu sync.Mutex // protects individual doc, must hold RLock of server.docs
}

type Server struct {
	Docs      map[int64]*DocMeta
	CommitLog []Request // Paxos stand-in for now

	cl   sync.Mutex   // protects CommitLog
	docs sync.RWMutex // W protects all docs, R needs to be held when locking just one doc

	doc_db *mongo.Collection
	cl_db  *mongo.Collection
}

// processes a request
// called while holding s.docs lock
func (s *Server) handle(r Request) {
	doc := s.Docs[r.DocId]

	if r.View > doc.UserViews[r.Uid] {
		doc.UserViews[r.Uid] = r.View
	}

	var ops [][]Op

	// trim previously applied ops
	for i, op := range r.Ops {
		if op[0].Seq >= doc.AppliedSeqs[r.Uid] {
			ops = r.Ops[i:]
			break
		}
	}

	// xform
	view := doc.Doc.View
	for _, op := range ops {
		for _, op1 := range doc.Log[len(doc.Log)-(view-op[0].View):] {
			if op[0].Uid != op1[0].Uid {
				op = Xform(op1, op)
			}
		}
		doc.Doc.ApplyOps(op)
	}

	doc.Log = append(doc.Log, ops...)
	doc.AppliedSeqs[r.Uid] = doc.Log[len(doc.Log)-1][0].Seq + 1
}

// applies commited requests to documents
func (s *Server) update() {
	for {
		time.Sleep(UpdateInterval)

		s.cl.Lock()
		tmp := s.CommitLog
		s.CommitLog = []Request{}
		s.cl.Unlock()

		s.docs.Lock()
		for _, r := range tmp {
			s.handle(r)
		}

		for _, d := range s.Docs {
			filter := bson.D{{Key: "docid", Value: d.DocId}}
			s.doc_db.ReplaceOne(context.TODO(), filter, d)
		}

		s.docs.Unlock()
	}
}

func (s *Server) NewClient(docId, uid int64, conn *websocket.Conn) Client {
	return Client{
		s:     s,
		doc:   s.Docs[docId],
		conn:  conn,
		uid:   uid,
		alive: true,
	}
}

// set up websocket
func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	docId, err := strconv.ParseInt(mux.Vars(r)["docid"], 10, 64)
	if err != nil {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	s.docs.RLock()
	if _, ok := s.Docs[docId]; !ok {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		s.docs.RUnlock()
		return
	}
	s.docs.RUnlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// wait for uid message
	_, res, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}
	uid, err := strconv.ParseInt(string(res), 10, 64)
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	c := s.NewClient(docId, uid, conn)
	c.interact()
}

func (s *Server) edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["docid"], 10, 64)
	if err != nil {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	s.docs.Lock()

	if _, ok := s.Docs[id]; !ok {
		s.Docs[id] = &DocMeta{
			Doc: Doc{
				Body:  []byte{},
				View:  0,
				DocId: id,
			},

			Log:         [][]Op{},
			NextSeq:     make(map[int64]int, 0),
			AppliedSeqs: make(map[int64]int, 0),
			UserViews:   make(map[int64]int, 0),
			DocId:       id,
		}

		s.doc_db.InsertOne(context.TODO(), s.Docs[id])
	}
	s.docs.Unlock()

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, editpage)
}

func NewServer(port int) *Server {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")

	s := new(Server)

	opt := options.Client().ApplyURI(MONGODB_URI)
	client, _ := mongo.Connect(context.TODO(), opt)

	s.doc_db = client.Database("gopad").Collection("documents")
	s.cl_db = client.Database("gopad").Collection("log")

	s.CommitLog = []Request{}
	s.Docs = make(map[int64]*DocMeta, 0)
	return s
}

// deletes old ops from logs
func (s *Server) prune() {
	for {
		s.cl.Lock()
		for _, d := range s.Docs {
			d.Log = d.Log[d.NextDiscard:]
			d.NextDiscard = len(d.Log)
		}
		s.cl.Unlock()

		time.Sleep(PruneInterval)
	}
}
