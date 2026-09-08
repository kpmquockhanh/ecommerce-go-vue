package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/repositories"
)

type CategoryHandler struct {
	categoryRepo repositories.CategoryRepository
}

func NewCategoryHandler(categoryRepo repositories.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := h.categoryRepo.List(r.Context())
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	if categories == nil {
		categories = []models.Category{}
	}

	respondWithJSON(w, map[string]interface{}{
		"categories": categories,
	}, http.StatusOK)
}

func (h *CategoryHandler) ListPublicCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := h.categoryRepo.List(r.Context())
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	// Filter to only categories with at least one published product
	var filtered []models.Category
	for _, cat := range categories {
		if cat.ProductCount > 0 {
			filtered = append(filtered, cat)
		}
	}

	if filtered == nil {
		filtered = []models.Category{}
	}

	respondWithJSON(w, map[string]interface{}{
		"categories": filtered,
	}, http.StatusOK)
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respondWithError(w, "Category name is required", http.StatusBadRequest)
		return
	}

	category, err := h.categoryRepo.Create(r.Context(), req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			respondWithError(w, "Category already exists", http.StatusConflict)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, category, http.StatusCreated)
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/categories/")
	idStr = strings.TrimSuffix(idStr, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respondWithError(w, "Category name is required", http.StatusBadRequest)
		return
	}

	if err := h.categoryRepo.Update(r.Context(), id, req.Name); err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Category not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "unique") {
			respondWithError(w, "Category name already exists", http.StatusConflict)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Category updated"}, http.StatusOK)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/categories/")
	idStr = strings.TrimSuffix(idStr, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	count, err := h.categoryRepo.CountProducts(r.Context(), id)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	if count > 0 {
		respondWithJSON(w, map[string]interface{}{
			"error":         "Category has products assigned. Remove it from products first or reassign.",
			"product_count": count,
		}, http.StatusConflict)
		return
	}

	if err := h.categoryRepo.Delete(r.Context(), id); err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Category not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Category deleted"}, http.StatusOK)
}
