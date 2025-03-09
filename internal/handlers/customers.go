package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"exampleserver/internal/auth"
	"exampleserver/pkg/logger"
)

type Customer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CustomersResponse struct {
	Customers []Customer `json:"customers"`
}

type Customers struct{}

func NewCustomers() *Customers {
	return &Customers{}
}

func (c *Customers) List(w http.ResponseWriter, r *http.Request) {
	logger.WithFields(map[string]interface{}{
		"handler": "customers",
		"method":  "List",
	}).Debug("Listing customers")

	// Get claims from context
	claims, ok := auth.GetClaims(r.Context())
	if ok {
		fmt.Printf("Request claims: %+v\n", claims)
	} else {
		fmt.Println("No claims found in request context")
	}

	// TODO: Implement actual customer fetching logic
	customers := []Customer{
		{ID: "1", Name: "John Doe"},
		{ID: "2", Name: "Jane Smith"},
	}

	response := CustomersResponse{
		Customers: customers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
