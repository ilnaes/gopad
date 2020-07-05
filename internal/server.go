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

type Server struct {
	Docs       map[int64]*DocMeta
	CommitLog  []Request // Paxos stand-in for now
	LastCommit int       // last req in log to have been handled
	LastSave   int       // last req in log to have been saved to db

	cl   sync.Mutex   // protects CommitLog
	docs sync.RWMutex // W protects all docs, R needs to be held when locking just one doc

	db *mongo.Collection
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
		tmp := s.CommitLog[s.LastCommit:]
		s.LastCommit = len(s.CommitLog)
		s.cl.Unlock()

		s.docs.Lock()
		for _, r := range tmp {
			s.handle(r)
		}
		s.docs.Unlock()

		req := make([]interface{}, len(tmp))
		for i, x := range tmp {
			req[i] = interface{}(x)
		}

		s.db.InsertMany(context.TODO(), req)
		s.LastSave += len(tmp)
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

	s := Server{
		Docs:      make(map[int64]*DocMeta, 0),
		CommitLog: make([]Request, 0),
	}

	opt := options.Client().ApplyURI(MONGODB_URI)
	client, _ := mongo.Connect(context.TODO(), opt)

	s.db = client.Database("gopad").Collection("log")

	cur, err := s.db.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(context.TODO()) {
		// create a value into which the single document can be decoded
		var elem Request
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		s.CommitLog = append(s.CommitLog, elem)
	}

	s.LastSave = len(s.CommitLog)
	s.LastCommit = len(s.CommitLog)

	// populate docs from log
	for _, req := range s.CommitLog {
		if _, ok := s.Docs[req.DocId]; !ok {
			s.Docs[req.DocId] = &DocMeta{
				Doc: Doc{
					Body:  []byte{},
					View:  0,
					DocId: req.DocId,
				},
				Log:         [][]Op{},
				NextSeq:     make(map[int64]int, 0),
				AppliedSeqs: make(map[int64]int, 0),
				UserViews:   make(map[int64]int, 0),
				DocId:       req.DocId,
			}
		}

		s.handle(req)
	}

	return &s
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
