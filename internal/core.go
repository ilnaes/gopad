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

	r.HandleFunc("/edit/{docid}", server.edit).Methods("GET")
	r.HandleFunc("/ws/{docid}", server.ws)

	r.HandleFunc("/login", server.login).Methods("POST")
	r.HandleFunc("/register", server.register).Methods("PUT")
	r.PathPrefix("/dist/").Handler(http.StripPrefix("/dist/", http.FileServer(http.Dir("dist/")))).Methods("GET")

	r.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world!")
	})
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "dist/index.html")
	}).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
