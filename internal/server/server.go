package server

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	c "github.com/ilnaes/gopad/internal/common"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	docs map[int64]*c.Doc
	log  []c.Op
	mu   sync.Mutex
}

func (s *Server) run() {
}

// pushes info out to client
func (s *Server) push(conn *websocket.Conn) {
}

// reads input from client
func (s *Server) pull(conn *websocket.Conn) {
	for {
		var m []c.Op
		if err := conn.ReadJSON(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		s.mu.Lock()
		s.log = append(s.log, m...)
		s.mu.Unlock()
	}
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	if _, ok := s.docs[id]; !ok {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go s.pull(conn)
}

func (s *Server) edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, "Malformed id", http.StatusBadRequest)
		return
	}

	if _, ok := s.docs[id]; !ok {
		s.docs[id] = &c.Doc{Body: []byte{}, View: 0, Id: id}
	}

	http.ServeFile(w, r, "/static/edit.html")
}
