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
	proto "github.com/triumphpc/go-musthave-shortener-tpl/pkg/api"
	grpcshortener "github.com/triumphpc/go-musthave-shortener-tpl/pkg/shortener"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
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

	// make gRPC without registered service
	s := grpc.NewServer()

	// HTTP server
	if c.EnableHTTPS == "false" {
		srv := startHTTPServer(c, mux, stop)
		// gRPC service
		rungRPC(c, p, s, stop)
		releaseResources(ctx, c, srv, poolClose, s)
	} else {
		// HTTPS server
		srv := startHTTPSServer(c, mux, stop)
		// gRPC service
		rungRPC(c, p, s, stop)
		releaseResources(ctx, c, srv, poolClose, s)
	}

}

// Run gRPC server
func rungRPC(c *configs.Config, p *worker.Pool, s *grpc.Server, stop context.CancelFunc) {
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		stop()
		c.Logger.Fatal(err.Error())
	}
	// service register
	proto.RegisterShortenerServer(s, grpcshortener.New(c.Logger, c.Storage, c.Database, p))

	c.Logger.Info("gRPC server started on :3200")

	// get request from gRPC
	go func() {
		if err := s.Serve(listen); err != nil {
			stop()
			c.Logger.Fatal(err.Error())
		}
	}()
}

// startHTTPSServer run HTTPS server
func startHTTPSServer(c *configs.Config, h http.Handler, stop context.CancelFunc) *http.Server {
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
			stop()
		}
	}()

	c.Logger.Info("Start server address: 443")

	return srv
}

// startHTTPServer run HTTP server
func startHTTPServer(c *configs.Config, h http.Handler, stop context.CancelFunc) *http.Server {
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
			stop()
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
func releaseResources(ctx context.Context, c *configs.Config, srv *http.Server, poolClose func(), s *grpc.Server) {
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
	// Close gRPC server
	s.Stop()
	c.Logger.Info("gRPC server stopped")

	// Close pool worker
	poolClose()
	// Server shutdown
	if err := srv.Shutdown(ctx); err != nil {
		c.Logger.Info("app error exit", zap.Error(err))
	}
	c.Logger.Info("Done")
}
