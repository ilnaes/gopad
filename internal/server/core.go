package server

import (
	"net/http"

	"github.com/gorilla/mux"
	c "github.com/ilnaes/gopad/internal/common"
)

func Run(port int) {
	server := Server{
		Docs:      make(map[int64]*DocMeta, 0),
		CommitLog: []c.Request{},
	}
	go server.update()

	r := mux.NewRouter()

	r.HandleFunc("/edit/{docid}/{uid}", server.edit).Methods("POST")
	http.Handle("/", r)
}
