package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Response encoding error: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	respondWithJSON(w, map[string]string{"error": message}, statusCode)
}

func respondWithServerError(w http.ResponseWriter, message string, err error, statusCode int) {
	log.Printf("ERROR [%d]: %s: %v", statusCode, message, err)
	respondWithError(w, message, statusCode)
}
