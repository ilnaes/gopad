package server

import (
	"net/http"

	"github.com/gorilla/mux"
	c "github.com/ilnaes/gopad/internal/common"
)

func Run(port int) {
	server := Server{
		docs: make(map[int64]*c.Doc, 0),
		log:  []c.Op{},
	}
	go server.run()

	r := mux.NewRouter()

	r.HandleFunc("/edit/{id}", server.edit).Methods("POST")
	http.Handle("/", r)
}
