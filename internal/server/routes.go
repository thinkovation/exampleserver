package server

import (
	"net/http"

	"exampleserver/internal/auth"
	"exampleserver/internal/handlers"
	"exampleserver/pkg/logger"
)

func (s *Server) setupRoutes() {
	// Create JWT service for token generation
	jwtService := auth.NewJWTService(s.config.JWTSecret)

	// Create authenticators and middleware
	jwtAuth := auth.NewJWTAuthenticator(s.config.JWTSecret, "")
	apiAuth := auth.NewAPIKeyAuthenticator(nil)
	authChain := auth.NewChain(apiAuth, jwtAuth)
	authMiddleware := auth.NewMiddleware(authChain, s.logger)

	// Create handlers
	authHandler := handlers.NewAuth(jwtService)
	customersHandler := handlers.NewCustomers()
	loggerHandler := logger.NewHTTPHandler(logger.Default())

	// Static file server for public directory
	fs := http.FileServer(http.Dir("public"))
	s.router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))

	// API routes
	s.router.HandleFunc("/api/login", authHandler.Login).Methods("POST")
	s.router.Handle("/api/customers", authMiddleware.RequireAuth(http.HandlerFunc(customersHandler.List))).Methods("GET")
	s.router.HandleFunc("/api/loggersettings/debug", loggerHandler.SetDebug).Methods("POST")
	s.router.HandleFunc("/api/logging/log", loggerHandler.GetLogs).Methods("GET", "POST")
	s.router.HandleFunc("/api/logs", loggerHandler.PutWebook)

}
