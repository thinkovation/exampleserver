package main

import (
	"log"

	"exampleserver/internal/server"
	"exampleserver/internal/services"
	"exampleserver/internal/stats"
	"exampleserver/pkg/config"
	"exampleserver/pkg/logger"
)

func main() {
	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize shared logger
	if err := logger.Initialize("logger.yaml"); err != nil {
		log.Fatal(err)
	}

	// Log startup information
	logger.Info("Starting server...")

	// Create service manager
	serviceManager := services.NewManager()

	// Create and add stats service
	statsService := stats.NewStatsService(cfg.StatsInterval, logger.Default())
	serviceManager.AddService(statsService)

	// Create and start server
	srv := server.New(cfg, logger.Default())
	if err := srv.Start(); err != nil {
		logger.Fatal("Server error: %v", err)
	}
}
