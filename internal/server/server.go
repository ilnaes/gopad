package server

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	c "github.com/ilnaes/gopad/internal/common"
)

const (
	UpdateDelay = 250 * time.Millisecond
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type DocMeta struct {
	Doc c.Doc

	Log         []c.Op
	SeenSeqs    map[int64]int
	AppliedSeqs map[int64]int
}

type Server struct {
	Docs         map[int64]*DocMeta
	CommitLog    []c.Request
	Log          []c.Op
	Discardpoint int
	Commitpoint  int

	sync.Mutex
}

func (s *Server) update() {
	for {
		time.Sleep(UpdateDelay)
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

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	docId, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	uid, err := strconv.ParseInt(mux.Vars(r)["uid"], 10, 64)
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
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	if _, ok := s.Docs[id]; !ok {
		s.Docs[id] = &DocMeta{
			Doc: c.Doc{
				Body:  [][]byte{},
				View:  0,
				DocId: id,
			},

			Log:         []c.Op{},
			SeenSeqs:    make(map[int64]int, 0),
			AppliedSeqs: make(map[int64]int, 0),
		}
	}

	http.ServeFile(w, r, "/static/edit.html")
}
