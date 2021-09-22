package main

import (
	"context"
	_ "github.com/lib/pq"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/middlewares"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/routes"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Init logger
	l, err := logger.Instance()
	if err != nil {
		log.Fatal(err)
	}
	// Db instance
	dbh, _ := db.New()
	// Allocation handler and storage
	h, err := handlers.New(dbh)
	if err != nil {
		l.Fatal("app error exit", zap.Error(err))
	}
	// Get routes
	rtr := routes.Router(h)
	http.Handle("/", rtr)

	// Context with cancel func
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Get base URL
	serverAddress, err := configs.Instance().Param(configs.ServerAddress)
	if err != nil {
		l.Fatal("app error exit", zap.Error(err))
	}
	l.Info("Start server address: " + serverAddress)

	// Init server
	srv := &http.Server{
		Addr: serverAddress,
		// Send request to conveyor
		Handler: middlewares.Conveyor(rtr, middlewares.GzipMiddleware, middlewares.CookieMiddleware),
	}
	// Goroutine
	go func() {
		l.Fatal("app error exit", zap.Error(srv.ListenAndServe()))
	}()
	l.Info("The service is ready to listen and serve.")

	// Add context for Graceful shutdown
	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		l.Info("Got SIGINT...")
	case syscall.SIGTERM:
		l.Info("Got SIGTERM...")
	}

	// database close
	if dbh != nil {
		l.Info("Closing connect to db")
		err := dbh.Close()
		if err != nil {
			l.Info("Closing don't close")
		}
	}

	l.Info("The service is shutting down...")
	if err = srv.Shutdown(context.Background()); err != nil {
		l.Fatal("app error exit", zap.Error(err))
	}
	l.Info("Done")
}
