package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"exampleserver/internal/stats"
	"exampleserver/pkg/config"
	"exampleserver/pkg/logger"

	"github.com/gorilla/mux"
)

type Server struct {
	config       *config.Config
	router       *mux.Router
	server       *http.Server
	statsService *stats.StatsService
	logger       logger.LoggerInterface
}

func New(cfg *config.Config, logger logger.LoggerInterface) *Server {
	s := &Server{
		config:       cfg,
		router:       mux.NewRouter(),
		statsService: stats.NewStatsService(cfg.StatsInterval, logger),
		logger:       logger,
	}

	s.setupRoutes()

	s.server = &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	// Check if port is already in use
	addr := ":" + s.config.Port
	if ln, err := net.Listen("tcp", addr); err != nil {
		return fmt.Errorf("port %s is not available: %w", s.config.Port, err)
	} else {
		ln.Close()
	}

	// Create a root context for the server
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// WaitGroup to track all goroutines
	var wg sync.WaitGroup

	// Start stats service
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.statsService.Start(rootCtx); err != nil && err != context.Canceled {
			s.logger.Error("Stats service error: %v", err)
		}
	}()

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Start the server in a goroutine
	serverError := make(chan error, 1)
	go func() {
		s.logger.Info("Server starting on port %s", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverError <- err
		}
	}()

	// Wait for shutdown signal or server error
	var shutdownErr error
	select {
	case err := <-serverError:
		shutdownErr = fmt.Errorf("server error: %w", err)
		rootCancel() // Cancel all goroutines
	case <-sig:
		s.logger.Info("Shutdown signal received")
		rootCancel() // Cancel all goroutines

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Trigger graceful shutdown
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			shutdownErr = fmt.Errorf("error during shutdown: %w", err)
		}
	}

	// Wait for all goroutines to finish
	s.logger.Info("Waiting for all goroutines to finish...")
	wg.Wait()
	s.logger.Info("All goroutines finished")

	return shutdownErr
}
