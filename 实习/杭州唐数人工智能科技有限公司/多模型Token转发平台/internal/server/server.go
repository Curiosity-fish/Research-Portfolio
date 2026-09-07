package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Server wraps the HTTP server and exposes lifecycle helpers.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// New creates a new Server with the given address and Gin engine.
func New(addr string, engine *gin.Engine, logger *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           engine,
			ReadTimeout:       30 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			// WriteTimeout stays 0: it bounds the entire response lifetime,
			// which would kill long-lived SSE streams used by AI relaying.
			// Streaming handlers should set per-write deadlines via
			// http.NewResponseController instead.
			WriteTimeout: 0,
			IdleTimeout:  120 * time.Second,
		},
		logger: logger,
	}
}

// Start begins serving HTTP requests in a goroutine.
// It returns an error if the listener cannot be created.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.httpServer.Addr, err)
	}

	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.logger.Error("http server failed", slog.Any("error", err))
		}
	}()
	return nil
}

// Shutdown gracefully stops the server with the provided timeout.
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// WaitForShutdown blocks until an interrupt or termination signal is received,
// then shuts down the server gracefully.
func (s *Server) WaitForShutdown(timeout time.Duration) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("shutdown signal received")
	if err := s.Shutdown(timeout); err != nil {
		s.logger.Error("graceful shutdown failed", slog.Any("error", err))
		return err
	}
	s.logger.Info("server stopped")
	return nil
}
