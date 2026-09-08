package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"ecommerce-api-go/internal/database"
	"ecommerce-api-go/internal/models"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns the health status of the service
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	status := "healthy"
	message := "Server is up and running!"

	// Check database connectivity
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := database.DB.Ping(ctx); err != nil {
		status = "degraded"
		message = "Database connection failed"
		log.Printf("Health check: database ping failed: %v", err)
	}

	healthInfo := models.HealthResponse{
		Status:    status,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(healthInfo); err != nil {
		log.Printf("Health check encoding error: %v", err)
		http.Error(w, "Health check failed", http.StatusInternalServerError)
		return
	}

	log.Printf("Health check completed in %v", time.Since(start))
}
