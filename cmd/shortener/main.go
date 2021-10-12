package main

import (
	"context"
	_ "github.com/lib/pq"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/middlewares"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/worker"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/routes"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Init project config
	c := configs.Instance()
	// Allocation handler and storage
	h := handlers.New(c.Logger, c.Storage)
	// Worker for background tasks
	ctx := context.Background()
	// Pool workers
	p := worker.New(ctx, c.Logger, c.Storage)
	// Init routes
	rtr := routes.Router(h, c, p)
	http.Handle("/", rtr)

	// Context with cancel func
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Get base URL
	serverAddress, err := c.Param(configs.ServerAddress)
	if err != nil {
		c.Logger.Fatal("app error exit", zap.Error(err))
	}
	c.Logger.Info("Start server address: " + serverAddress)

	// Init server
	srv := &http.Server{
		Addr: serverAddress,
		// Send request to conveyor example
		Handler: middlewares.Conveyor(
			rtr, middlewares.NewCompressor(c.Logger).GzipMiddleware,
			middlewares.New(c.Logger).CookieMiddleware,
		),
	}
	// Goroutine
	go func() {
		c.Logger.Fatal("app error exit", zap.Error(srv.ListenAndServe()))
	}()
	c.Logger.Info("The service is ready to listen and serve.")

	// Add context for Graceful shutdown
	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		c.Logger.Info("Got SIGINT...")
	case syscall.SIGTERM:
		c.Logger.Info("Got SIGTERM...")
	}

	// database close
	if c.Database != nil {
		c.Logger.Info("Closing connect to db")
		err := c.Database.Close()
		if err != nil {
			c.Logger.Info("Closing don't close")
		}
	}

	c.Logger.Info("The service is shutting down...")
	if err = srv.Shutdown(ctx); err != nil {
		c.Logger.Fatal("app error exit", zap.Error(err))
	}
	c.Logger.Info("Done")
}
