package internal

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func Run(port int) {
	addr := "127.0.0.1"
	server := NewServer(addr, port)
	go server.update()

	r := mux.NewRouter()

	r.HandleFunc("/edit/{docid}", server.edit)
	r.HandleFunc("/ws/{docid}", server.ws)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
