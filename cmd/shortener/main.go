package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/middlewares"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/worker"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/routes"
)

func main() {
	// Init project config
	c := configs.Instance()
	// Allocation handler and storage
	h := handlers.New(c.Logger, c.Storage)
	// Worker for background tasks
	ctx := context.Background()
	// Pool workers
	p, poolClose := worker.New(ctx, c.Logger, c.Storage)
	// Init routes
	rtr := routes.Router(h, c, p)
	http.Handle("/", rtr)
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
			middlewares.NewCookie(c.Logger).CookieMiddleware,
		),
	}
	// Goroutine to run server
	go func() {
		c.Logger.Info("app error exit", zap.Error(srv.ListenAndServe()))
	}()
	c.Logger.Info("The service is ready to listen and serve.")

	// Shut down handler
	shutDownServer(ctx, c, srv, poolClose)
}

func shutDownServer(ctx context.Context, c *configs.Config, srv *http.Server, poolClose func()) {
	// Context with cancel func
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Add context for Graceful shutdown
	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		c.Logger.Info("Got SIGINT...")
	case syscall.SIGTERM:
		c.Logger.Info("Got SIGTERM...")
	}

	c.Logger.Info("The service is shutting down...")
	// database close
	if c.Database != nil {
		c.Logger.Info("Closing connect to db")
		err := c.Database.Close()
		if err != nil {
			c.Logger.Info("Closing don't close")
		}
	}
	// Close pool worker
	poolClose()
	// Server shutdown
	if err := srv.Shutdown(ctx); err != nil {
		c.Logger.Info("app error exit", zap.Error(err))
	}
	c.Logger.Info("Done")
}
