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

type CartHandler struct {
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
}

func NewCartHandler(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository) *CartHandler {
	return &CartHandler{cartRepo: cartRepo, productRepo: productRepo}
}

func (h *CartHandler) getCartIdentifier(r *http.Request) (userID *int, sessionID *string) {
	claims := middleware.GetUserFromContext(r)
	if claims != nil {
		userID = &claims.UserID
		return
	}

	sid := r.Header.Get("X-Session-ID")
	if sid != "" {
		sessionID = &sid
	}
	return
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, sessionID := h.getCartIdentifier(r)
	if userID == nil && sessionID == nil {
		respondWithJSON(w, models.Cart{Items: []models.CartItem{}, Total: 0, ItemCount: 0}, http.StatusOK)
		return
	}

	var dbRows []repositories.CartItemRow
	var err error
	if userID != nil {
		dbRows, err = h.cartRepo.FindByUserID(r.Context(), *userID)
	} else {
		dbRows, err = h.cartRepo.FindBySessionID(r.Context(), *sessionID)
	}
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	var items []models.CartItem
	total := 0

	for _, row := range dbRows {
		subtotal := row.ProductPrice * row.Quantity
		total += subtotal

		presignedImages, _ := presignProductImages(r.Context(), row.ProductImages, "")

		item := models.CartItem{
			ID:       row.CartItemID,
			Quantity: row.Quantity,
			Subtotal: subtotal,
			Product: models.Product{
				ID:     row.ProductID,
				Name:   row.ProductName,
				Slug:   row.ProductSlug,
				Price:  row.ProductPrice,
				Images: presignedImages,
			},
		}

		if row.VariantID != nil {
			variantLabel := ""
			if row.VariantLabel != nil {
				variantLabel = *row.VariantLabel
			}
			item.Variant = &models.ProductVariant{
				ID:    *row.VariantID,
				Label: variantLabel,
				Stock: *row.VariantStock,
			}
		}

		if len(row.CustomizationData) > 0 {
			var custData map[string]string
			if err := json.Unmarshal(row.CustomizationData, &custData); err == nil {
				item.CustomizationData = custData
			}
		}

		items = append(items, item)
	}

	if items == nil {
		items = []models.CartItem{}
	}

	respondWithJSON(w, models.Cart{
		Items:     items,
		Total:     total,
		ItemCount: len(items),
	}, http.StatusOK)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Quantity < 1 {
		req.Quantity = 1
	}

	userID, sessionID := h.getCartIdentifier(r)
	if userID == nil && sessionID == nil {
		respondWithError(w, "Authentication or session required", http.StatusBadRequest)
		return
	}

	_, err := h.productRepo.FindByID(r.Context(), req.ProductID)
	if err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	variantID := req.VariantID
	if variantID != nil && *variantID == 0 {
		variantID = nil
	}

	var custData []byte
	if len(req.CustomizationData) > 0 {
		var err error
		custData, err = json.Marshal(req.CustomizationData)
		if err != nil {
			respondWithError(w, "Invalid customization data", http.StatusBadRequest)
			return
		}
	}

	if err := h.cartRepo.Upsert(r.Context(), userID, sessionID, req.ProductID, variantID, req.Quantity, custData); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Item added to cart"}, http.StatusCreated)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/cart/items/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid cart item ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Quantity < 1 {
		respondWithError(w, "Quantity must be at least 1", http.StatusBadRequest)
		return
	}

	userID, sessionID := h.getCartIdentifier(r)

	if err := h.cartRepo.UpdateQuantity(r.Context(), id, userID, sessionID, req.Quantity); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Cart item updated"}, http.StatusOK)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/cart/items/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid cart item ID", http.StatusBadRequest)
		return
	}

	userID, sessionID := h.getCartIdentifier(r)

	if err := h.cartRepo.Delete(r.Context(), id, userID, sessionID); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Cart item removed"}, http.StatusOK)
}

func (h *CartHandler) MergeGuestCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		respondWithError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		respondWithJSON(w, map[string]string{"message": "No guest cart to merge"}, http.StatusOK)
		return
	}

	merged, err := h.cartRepo.MergeGuestCart(r.Context(), claims.UserID, req.SessionID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]interface{}{
		"message": "Guest cart merged",
		"merged":  merged,
	}, http.StatusOK)
}
