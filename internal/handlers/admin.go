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

type AdminHandler struct {
	userRepo    repositories.UserRepository
	orderRepo   repositories.OrderRepository
	productRepo repositories.ProductRepository
}

func NewAdminHandler(userRepo repositories.UserRepository, orderRepo repositories.OrderRepository, productRepo repositories.ProductRepository) *AdminHandler {
	return &AdminHandler{userRepo: userRepo, orderRepo: orderRepo, productRepo: productRepo}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
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

	total, err := h.userRepo.Count(r.Context())
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	users, err := h.userRepo.List(r.Context(), limit, offset)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []models.User{}
	}

	respondWithJSON(w, map[string]interface{}{
		"users": users,
		"pagination": models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: (total + limit - 1) / limit,
		},
	}, http.StatusOK)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondWithError(w, "User ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), id)
	if err != nil {
		respondWithError(w, "User not found", http.StatusNotFound)
		return
	}

	respondWithJSON(w, user, http.StatusOK)
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	idStr = strings.TrimSuffix(idStr, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		respondWithError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.UserID == id {
		respondWithError(w, "Cannot change your own role", http.StatusBadRequest)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Role = strings.TrimSpace(req.Role)
	if req.Role != "admin" && req.Role != "user" {
		respondWithError(w, "Role must be 'admin' or 'user'", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.UpdateRole(r.Context(), id, req.Role); err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "User not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	user, _ := h.userRepo.FindByID(r.Context(), id)
	respondWithJSON(w, user, http.StatusOK)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	idStr = strings.TrimSuffix(idStr, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		respondWithError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.UserID == id {
		respondWithError(w, "Cannot delete your own account", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.Delete(r.Context(), id); err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "User not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "User deleted"}, http.StatusOK)
}

func (h *AdminHandler) DashboardStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderStats, err := h.orderRepo.GetStats(r.Context())
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	totalUsers, err := h.userRepo.Count(r.Context())
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	_, totalProducts, err := h.productRepo.List(r.Context(), repositories.ProductFilters{Limit: 1})
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]interface{}{
		"total_orders":     orderStats.TotalOrders,
		"total_revenue":    orderStats.TotalRevenue,
		"pending_orders":   orderStats.PendingOrders,
		"delivered_orders": orderStats.DeliveredOrders,
		"total_users":      totalUsers,
		"total_products":   totalProducts,
	}, http.StatusOK)
}
