package handlers

import (
	"encoding/json"
	"net/http"

	"exampleserver/internal/auth"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type Auth struct {
	jwtService *auth.JWTService
}

func NewAuth(jwtService *auth.JWTService) *Auth {
	return &Auth{
		jwtService: jwtService,
	}
}

func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual authentication logic here
	// For now, we'll just check if username and password are not empty
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// In a real application, you would validate credentials here
	// For now, we'll just generate a token with the username
	token, err := a.jwtService.GenerateToken("user-123", req.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
