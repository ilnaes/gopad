package internal

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func Run(port int) {
	server := Server{
		Docs:      make(map[int64]*DocMeta, 0),
		CommitLog: []Request{},
	}
	go server.update()

	r := mux.NewRouter()

	r.HandleFunc("/edit/{docid}", server.edit)
	r.HandleFunc("/ws/{docid}", server.ws)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
