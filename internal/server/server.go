package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/config"
)

// Server represents the HTTP server.
type Server struct {
	config *config.Config
	logger *logrus.Logger
	server *http.Server
}

// New creates a new server instance.
func New(cfg *config.Config, logger *logrus.Logger) (*Server, error) {
	mux := http.NewServeMux()
	
	// Add basic endpoints
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})
	
	mux.HandleFunc(cfg.Server.MetricsPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics will be implemented in later tasks\n"))
	})

	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: mux,
	}

	return &Server{
		config: cfg,
		logger: logger,
		server: server,
	}, nil
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.WithField("address", s.config.Server.Address).Info("Starting HTTP server")
	
	go func() {
		<-ctx.Done()
		s.logger.Info("Context cancelled, shutting down server")
		s.server.Shutdown(context.Background())
	}()
	
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}