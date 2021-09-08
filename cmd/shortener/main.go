package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Allocation handler and storage
	h, err := handlers.New()
	if err != nil {
		log.Fatal(err)
	}

	// Make Routes
	rtr := mux.NewRouter()
	rtr.HandleFunc("/api/shorten", h.SaveJSON)
	rtr.HandleFunc("/{id:.+}", h.Get)
	rtr.HandleFunc("/", h.Save)

	http.Handle("/", rtr)

	// context with cancel func
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Get base URL
	serverAddress, err := configs.Instance().Param(configs.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Start server address:", serverAddress)

	// Init server
	srv := &http.Server{
		Addr: serverAddress,
	}
	// Goroutine
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()
	log.Print("The service is ready to listen and serve.")

	// Add context for Graceful shutdown
	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	log.Print("The service is shutting down...")
	if err = srv.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
	log.Print("Done")
}
