package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-api-go/internal/middleware"
	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/repositories"
)

type ReviewHandler struct {
	reviewRepo  repositories.ReviewRepository
	productRepo repositories.ProductRepository
}

func NewReviewHandler(reviewRepo repositories.ReviewRepository, productRepo repositories.ProductRepository) *ReviewHandler {
	return &ReviewHandler{reviewRepo: reviewRepo, productRepo: productRepo}
}

func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		respondWithError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		respondWithError(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	productID := req.ProductID
	if productID == 0 {
		respondWithError(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	exists, err := h.reviewRepo.ExistsProduct(r.Context(), productID)
	if err != nil || !exists {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	review, err := h.reviewRepo.Create(r.Context(), productID, claims.UserID, req.Rating, req.Comment)
	if err != nil {
		respondWithError(w, "You have already reviewed this product", http.StatusConflict)
		return
	}

	review.UserName = claims.Email

	respondWithJSON(w, review, http.StatusCreated)
}

func (h *ReviewHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/reviews/product/")
	productID, err := strconv.Atoi(path)
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
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

	var total int
	exists, err := h.reviewRepo.ExistsProduct(r.Context(), productID)
	if err != nil || !exists {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	_, count, err := h.reviewRepo.GetStats(r.Context(), productID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	total = count

	reviews, err := h.reviewRepo.ListByProduct(r.Context(), productID, limit, offset)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	avgRating, _, _ := h.reviewRepo.GetStats(r.Context(), productID)

	if reviews == nil {
		reviews = []models.Review{}
	}

	respondWithJSON(w, models.ReviewListResponse{
		Reviews:   reviews,
		AvgRating: avgRating,
		Count:     count,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: (total + limit - 1) / limit,
		},
	}, http.StatusOK)
}
