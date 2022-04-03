package main

import (
	"context"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/handlers/middlewares"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/worker"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/routes"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Global variables
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	// Print build info
	printBuildInfo()
	// Init project config
	c := configs.Instance()
	// Allocation handler and storage
	h := handlers.New(c.Logger, c.Storage)

	// Init context
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	// Pool workers
	p, poolClose := worker.New(ctx, c.Logger, c.Storage)
	// Init routes
	rtr := routes.Router(h, c, p)
	http.Handle("/", rtr)
	// Init handle
	mux := middlewares.Conveyor(
		rtr, middlewares.NewCompressor(c.Logger).GzipMiddleware,
		middlewares.NewCookie(c.Logger).CookieMiddleware,
	)

	// HTTP server
	if c.EnableHTTPS == "false" {
		srv := startHTTPServer(c, mux)
		releaseResources(ctx, c, srv, poolClose)
	} else {
		// HTTPS server
		srv := startHTTPSServer(c, mux)
		releaseResources(ctx, c, srv, poolClose)
	}
}

// startHTTPSServer run HTTPS server
func startHTTPSServer(c *configs.Config, h http.Handler) *http.Server {
	serverAddress, err := c.Param(configs.ServerAddress)
	if err != nil {
		c.Logger.Fatal("app error exit", zap.Error(err))
	}

	// Start HTTPS server
	manager := &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(serverAddress),
	}

	srv := &http.Server{
		Addr:      ":443",
		Handler:   h,
		TLSConfig: manager.TLSConfig(),
	}

	go func() {
		err := srv.ListenAndServeTLS("server.crt", "server.key")
		if err != nil {
			c.Logger.Info("app error exit", zap.Error(err))
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}()

	c.Logger.Info("Start server address: 443")

	return srv
}

// startHTTPServer run HTTP server
func startHTTPServer(c *configs.Config, h http.Handler) *http.Server {
	serverAddress, err := c.Param(configs.ServerAddress)
	if err != nil {
		c.Logger.Fatal("app error exit", zap.Error(err))
	}

	srv := &http.Server{
		Addr:    serverAddress,
		Handler: h,
	}

	go func() {
		err = srv.ListenAndServe()
		if err != nil {
			c.Logger.Info("app error exit", zap.Error(err))
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}()

	c.Logger.Info("Start server address: " + serverAddress)

	return srv
}

// printBuildInfo print info about package
func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

// releaseResources free resources
func releaseResources(ctx context.Context, c *configs.Config, srv *http.Server, poolClose func()) {
	<-ctx.Done()
	if ctx.Err() != nil {
		fmt.Printf("Error:%v\n", ctx.Err())
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

	time.Sleep(1 * time.Second)
	// Server shutdown
	if err := srv.Shutdown(ctx); err != nil {
		c.Logger.Info("app error exit", zap.Error(err))
	}
	c.Logger.Info("Done")
}
