package server

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	c "github.com/ilnaes/gopad/internal/common"
)

const editpage = `<html>
    <head>
        <script type="text/javascript" src="/static/main.js"></script>
    </head>
    <body>
        <center>
            <textarea id="textbox" name="textbox" rows="45" cols="150"></textarea>
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
	Doc c.Doc

	Log         [][]c.Op // one update is a collection of ops from one diff
	SeenSeqs    map[int64]int
	AppliedSeqs map[int64]int
	NextDiscard int
}

type Server struct {
	Docs      map[int64]*DocMeta
	CommitLog []c.Request // Paxos stand-in for now

	sync.Mutex
}

// processes a request
// called while holding lock
func (s *Server) handle(r c.Request) {
	doc := s.Docs[r.DocId]

	var ops [][]c.Op

	// trim previously applied ops
	for i, op := range r.Ops {
		if op[0].Seq > doc.AppliedSeqs[r.Uid] {
			ops = r.Ops[i:]
			break
		}
	}

	// xform
	for _, op := range ops {
		for _, op1 := range doc.Log[len(doc.Log)-(doc.Doc.View-r.View):] {
			op = c.Xform(op1, op)
		}

		doc.Log = append(doc.Log, op)
		doc.Doc.ApplyOps(op)
	}

	doc.AppliedSeqs[r.Uid] = doc.Log[len(doc.Log)-1][0].Seq
}

// deletes old ops from logs
func (s *Server) prune() {
	for {
		s.Lock()
		for _, d := range s.Docs {
			d.Log = d.Log[d.NextDiscard:]
			d.NextDiscard = len(d.Log)
		}
		s.Unlock()

		time.Sleep(PruneInterval)
	}
}

// applies commited requests to documents
func (s *Server) update() {
	// go s.prune()

	for {
		time.Sleep(UpdateInterval)
		s.Lock()

		for _, r := range s.CommitLog {
			s.handle(r)
		}

		s.CommitLog = []c.Request{}

		s.Unlock()
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

	// random uids for now
	uid := rand.Int63()
	if err != nil {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	if _, ok := s.Docs[docId]; !ok {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
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

	if _, ok := s.Docs[id]; !ok {
		s.Docs[id] = &DocMeta{
			Doc: c.Doc{
				Body:  []byte{},
				View:  0,
				DocId: id,
			},

			Log:         [][]c.Op{},
			SeenSeqs:    make(map[int64]int, 0),
			AppliedSeqs: make(map[int64]int, 0),
		}
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, editpage)
}
