package main

import (
	"context"
	_ "github.com/lib/pq"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/middlewares"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/mypool"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/routes"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

func main() {
	// Init project config
	c := configs.Instance()
	// Allocation handler and storage
	h := handlers.New(c.Logger, c.Storage)
	// Worker for background tasks
	ctx, cancel := context.WithCancel(context.Background())
	//// Pool workers
	//p, poolClose := worker.New(ctx, c.Logger, c.Storage)

	pool := mypool.New(c.Logger, 1000)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := pool.Run(ctx, runtime.NumCPU()); err != nil {
			c.Logger.Error("Worker pool returned error", zap.Error(err))
			cancel()
		}
	}()

	// Init routes
	rtr := routes.Router(h, c, pool, c.Storage)
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
			middlewares.New(c.Logger).CookieMiddleware,
		),
	}
	// Goroutine to run server
	go func() {
		c.Logger.Info("app error exit", zap.Error(srv.ListenAndServe()))
	}()
	c.Logger.Info("The service is ready to listen and serve.")

	// Context with cancel func
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Add context for Graceful shutdown
	select {
	case killSignal := <-interrupt:
		switch killSignal {
		case os.Interrupt:
			c.Logger.Info("Got SIGINT...")
		case syscall.SIGTERM:
			c.Logger.Info("Got SIGTERM...")
		}
	case <-ctx.Done():
	}

	cancel()

	c.Logger.Info("The service is shutting down...")
	// database close
	if c.Database != nil {
		c.Logger.Info("Closing connect to db")
		err := c.Database.Close()
		if err != nil {
			c.Logger.Info("Closing don't close")
		}
	}
	// Server shutdown
	if err := srv.Shutdown(ctx); err != nil {
		c.Logger.Info("app error exit", zap.Error(err))
	}

	wg.Wait()

	c.Logger.Info("Done")
}
