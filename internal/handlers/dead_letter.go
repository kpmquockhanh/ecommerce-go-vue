package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-api-go/internal/queue"
	"ecommerce-api-go/internal/repositories"
)

type DeadLetterHandler struct {
	deadLetterRepo repositories.DeadLetterRepository
	queue          *queue.Queue
}

func NewDeadLetterHandler(deadLetterRepo repositories.DeadLetterRepository, q *queue.Queue) *DeadLetterHandler {
	return &DeadLetterHandler{deadLetterRepo: deadLetterRepo, queue: q}
}

func (h *DeadLetterHandler) ListDeadLetters(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	total, err := h.deadLetterRepo.Count(r.Context())
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	deadLetters, err := h.deadLetterRepo.List(r.Context(), limit, offset)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	if deadLetters == nil {
		deadLetters = []repositories.DeadLetter{}
	}

	respondWithJSON(w, map[string]interface{}{
		"dead_letters": deadLetters,
		"pagination": map[string]int{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	}, http.StatusOK)
}

func (h *DeadLetterHandler) RetryDeadLetter(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/dead-letters/")
	idStr = strings.TrimSuffix(idStr, "/retry")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid dead letter ID", http.StatusBadRequest)
		return
	}

	dl, err := h.deadLetterRepo.FindByID(r.Context(), id)
	if err != nil {
		respondWithError(w, "Dead letter not found", http.StatusNotFound)
		return
	}

	var payload interface{}
	if err := json.Unmarshal(dl.Payload, &payload); err != nil {
		respondWithServerError(w, "Invalid payload", err, http.StatusInternalServerError)
		return
	}

	job := queue.Job{
		Type:    queue.JobType(dl.JobType),
		Payload: payload,
	}

	if err := h.queue.Publish(r.Context(), dl.QueueName, job); err != nil {
		respondWithServerError(w, "Failed to retry job", err, http.StatusInternalServerError)
		return
	}

	if err := h.deadLetterRepo.MarkRetried(r.Context(), id); err != nil {
		respondWithServerError(w, "Failed to update dead letter status", err, http.StatusInternalServerError)
	}

	log.Printf("DLQ: retried job id=%d type=%s queue=%s", id, dl.JobType, dl.QueueName)
	respondWithJSON(w, map[string]string{"message": "Job retried successfully"}, http.StatusOK)
}

func (h *DeadLetterHandler) DeleteDeadLetter(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/dead-letters/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid dead letter ID", http.StatusBadRequest)
		return
	}

	if err := h.deadLetterRepo.Delete(r.Context(), id); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Dead letter deleted"}, http.StatusOK)
}
