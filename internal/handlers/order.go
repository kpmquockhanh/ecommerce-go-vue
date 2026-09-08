package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"

	"ecommerce-api-go/internal/middleware"
	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/queue"
	"ecommerce-api-go/internal/repositories"
)

type OrderHandler struct {
	orderRepo         repositories.OrderRepository
	cartRepo          repositories.CartRepository
	userRepo          repositories.UserRepository
	idempotencyRepo   repositories.IdempotencyRepository
	checkoutSessionRepo repositories.CheckoutSessionRepository
	queue             *queue.Queue
}

func NewOrderHandler(orderRepo repositories.OrderRepository, cartRepo repositories.CartRepository, userRepo repositories.UserRepository, idempotencyRepo repositories.IdempotencyRepository, checkoutSessionRepo repositories.CheckoutSessionRepository, q *queue.Queue) *OrderHandler {
	return &OrderHandler{orderRepo: orderRepo, cartRepo: cartRepo, userRepo: userRepo, idempotencyRepo: idempotencyRepo, checkoutSessionRepo: checkoutSessionRepo, queue: q}
}

func (h *OrderHandler) InitCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		respondWithError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.CheckoutInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = fmt.Sprintf("checkout-%d-%d", claims.UserID, time.Now().UnixMilli())
	}

	existing, err := h.checkoutSessionRepo.FindByKey(r.Context(), idempotencyKey, claims.UserID)
	if err == nil && existing.PaymentIntentID != "" {
		pi, err := paymentintent.Get(existing.PaymentIntentID, nil)
		if err == nil && pi.Status != stripe.PaymentIntentStatusCanceled {
			respondWithJSON(w, models.CheckoutInitResponse{
				ClientSecret: pi.ClientSecret,
			}, http.StatusOK)
			return
		}
	}

	cartRows, err := h.cartRepo.FindByUserID(r.Context(), claims.UserID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if len(cartRows) == 0 {
		respondWithError(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	var cartItems []repositories.CartSnapshotItem
	total := 0
	for _, row := range cartRows {
		subtotal := row.ProductPrice * row.Quantity
		total += subtotal

		variantLabel := ""
		if row.VariantLabel != nil {
			variantLabel = *row.VariantLabel
		}

		var custData map[string]string
		if len(row.CustomizationData) > 0 {
			json.Unmarshal(row.CustomizationData, &custData)
		}

		cartItems = append(cartItems, repositories.CartSnapshotItem{
			ProductID:         row.ProductID,
			VariantID:         row.VariantID,
			ProductName:       row.ProductName,
			VariantLabel:      variantLabel,
			Quantity:          row.Quantity,
			Price:             row.ProductPrice,
			CustomizationData: custData,
		})
	}

	piParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(total)),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	pi, err := paymentintent.New(piParams)
	if err != nil {
		respondWithServerError(w, "Failed to create payment intent", err, http.StatusInternalServerError)
		return
	}

	_, err = h.checkoutSessionRepo.Create(r.Context(), repositories.CreateCheckoutSessionParams{
		UserID:          claims.UserID,
		IdempotencyKey:  idempotencyKey,
		PaymentIntentID: pi.ID,
		CartItems:       cartItems,
		Total:           total,
	})
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, models.CheckoutInitResponse{
		ClientSecret: pi.ClientSecret,
	}, http.StatusCreated)
}

func (h *OrderHandler) ConfirmCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		respondWithError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.CheckoutConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	addr := req.ShippingAddress
	if addr.FirstName == "" || addr.LastName == "" || addr.Address1 == "" ||
		addr.City == "" || addr.State == "" || addr.Zip == "" || addr.Country == "" {
		respondWithError(w, "All shipping address fields are required", http.StatusBadRequest)
		return
	}

	if req.IdempotencyKey == "" {
		respondWithError(w, "Idempotency key required", http.StatusBadRequest)
		return
	}

	session, err := h.checkoutSessionRepo.FindByKey(r.Context(), req.IdempotencyKey, claims.UserID)
	if err != nil {
		respondWithError(w, "Checkout not initialized", http.StatusBadRequest)
		return
	}

	if session.Status == "confirmed" {
		respondWithError(w, "Checkout already completed", http.StatusBadRequest)
		return
	}

	if session.Status == "expired" {
		respondWithError(w, "Checkout expired, please start again", http.StatusBadRequest)
		return
	}

	pi, err := paymentintent.Get(session.PaymentIntentID, nil)
	if err != nil {
		respondWithError(w, "Payment intent not found", http.StatusBadRequest)
		return
	}
	if pi.Status == stripe.PaymentIntentStatusCanceled {
		respondWithError(w, "Payment intent was cancelled, please start again", http.StatusBadRequest)
		return
	}

	cartRows, err := h.cartRepo.FindByUserID(r.Context(), claims.UserID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if len(cartRows) == 0 {
		respondWithError(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	var cartItems []repositories.CheckoutCartItem
	for _, row := range cartRows {
		variantLabel := ""
		if row.VariantLabel != nil {
			variantLabel = *row.VariantLabel
		}

		cartItems = append(cartItems, repositories.CheckoutCartItem{
			ProductID:        row.ProductID,
			VariantID:        row.VariantID,
			Quantity:         row.Quantity,
			Price:            row.ProductPrice,
			ProductName:      row.ProductName,
			VariantLabel:     variantLabel,
			CustomizationData: row.CustomizationData,
		})
	}

	total := 0
	for _, ci := range cartItems {
		total += ci.Price * ci.Quantity
	}

	if int(pi.Amount) != total {
		respondWithError(w, fmt.Sprintf("Cart total changed from $%.2f to $%.2f, please refresh", float64(pi.Amount)/100, float64(total)/100), http.StatusBadRequest)
		return
	}

	orderID, err := h.orderRepo.Checkout(r.Context(), claims.UserID, repositories.CheckoutParams{
		Total:           total,
		ShippingAddress: addr,
		PaymentIntentID: session.PaymentIntentID,
		CartItems:       cartItems,
		IdempotencyKey:  req.IdempotencyKey,
	})
	if err != nil {
		respondWithServerError(w, "Failed to create order", err, http.StatusInternalServerError)
		return
	}

	if err := h.checkoutSessionRepo.UpdateOrderID(r.Context(), req.IdempotencyKey, claims.UserID, orderID); err != nil {
		log.Printf("ConfirmCheckout: failed to update checkout session: %v", err)
	}

	respondWithJSON(w, models.CheckoutConfirmResponse{
		OrderID: orderID,
	}, http.StatusCreated)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		respondWithError(w, "Authentication required", http.StatusUnauthorized)
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

	orders, total, err := h.orderRepo.ListByUserID(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	if orders == nil {
		orders = []repositories.OrderSummary{}
	}

	respondWithJSON(w, map[string]interface{}{
		"orders": orders,
		"pagination": models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: (total + limit - 1) / limit,
		},
	}, http.StatusOK)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		respondWithError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.orderRepo.FindByID(r.Context(), orderID, claims.UserID)
	if err != nil {
		respondWithError(w, "Order not found", http.StatusNotFound)
		return
	}

	items, err := h.orderRepo.GetItems(r.Context(), orderID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	if items == nil {
		items = []models.OrderItem{}
	}

	respondWithJSON(w, models.OrderWithItems{
		Order: *order,
		Items: items,
	}, http.StatusOK)
}

func (h *OrderHandler) AdminListOrders(w http.ResponseWriter, r *http.Request) {
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

	status := query.Get("status")

	orders, total, err := h.orderRepo.ListAll(r.Context(), status, limit, offset)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	if orders == nil {
		orders = []repositories.AdminOrder{}
	}

	respondWithJSON(w, map[string]interface{}{
		"orders": orders,
		"pagination": models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: (total + limit - 1) / limit,
		},
	}, http.StatusOK)
}

func (h *OrderHandler) AdminUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/orders/")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{
		"pending": true, "paid": true, "shipped": true, "delivered": true, "cancelled": true,
	}
	if !validStatuses[req.Status] {
		respondWithError(w, "Invalid status", http.StatusBadRequest)
		return
	}

	allowedTransitions := map[string][]string{
		"pending":   {"paid", "cancelled"},
		"paid":      {"shipped", "cancelled"},
		"shipped":   {"delivered"},
		"delivered": {},
		"cancelled": {},
	}

	currentStatus, err := h.orderRepo.GetStatus(r.Context(), orderID)
	if err != nil {
		respondWithError(w, "Order not found", http.StatusNotFound)
		return
	}

	valid := false
	for _, target := range allowedTransitions[currentStatus] {
		if target == req.Status {
			valid = true
			break
		}
	}
	if !valid {
		respondWithError(w, fmt.Sprintf("Cannot transition from '%s' to '%s'", currentStatus, req.Status), http.StatusBadRequest)
		return
	}

	if err := h.orderRepo.UpdateStatus(r.Context(), orderID, req.Status); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	if h.queue != nil {
		userID, total, err := h.orderRepo.GetUserIDAndTotal(r.Context(), orderID)
		if err == nil {
			user, err := h.userRepo.FindByID(r.Context(), userID)
			if err == nil {
				jobErr := h.queue.Publish(r.Context(), "emails", queue.Job{
					Type: queue.JobOrderStatusChanged,
					Payload: queue.OrderStatusChangedPayload{
						OrderID:   orderID,
						UserID:    userID,
						Email:     user.Email,
						FirstName: user.FirstName,
						LastName:  user.LastName,
						OldStatus: currentStatus,
						NewStatus: req.Status,
					},
				})
				if jobErr != nil {
					log.Printf("AdminUpdateOrderStatus: failed to publish status change job: %v", jobErr)
				}
				_ = total
			}
		}
	}

	respondWithJSON(w, map[string]string{"message": "Order status updated"}, http.StatusOK)
}
