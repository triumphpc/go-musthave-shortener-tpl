package main

import (
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storage"
	"log"
	"net/http"
)

func main() {
	// Allocation storage for urls
	s := make(storage.Storage)
	// Make Routes
	rtr := mux.NewRouter()
	rtr.HandleFunc("/{id:.+}", handlers.Get(s))
	rtr.HandleFunc("/", handlers.Save(s))
	http.Handle("/", rtr)

	// Init server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
