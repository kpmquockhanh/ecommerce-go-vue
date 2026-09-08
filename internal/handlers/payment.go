package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"

	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/validation"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	validator *validation.PaymentValidator
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		validator: validation.NewPaymentValidator(),
	}
}

// CreatePaymentIntent handles payment intent creation
func (h *PaymentHandler) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	// Method validation
	if r.Method != "POST" {
		log.Printf("Method not allowed: %s", r.Method)
		h.respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON parsing error: %v", err)
		h.respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validator.ValidatePaymentRequest(req); err != nil {
		log.Printf("Validation error: %v", err)
		h.respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate amount
	amount := h.calculateOrderAmount(req.ProductId)
	log.Printf("Creating payment intent for product %s, amount: $%.2f",
		req.ProductId, float64(amount)/100)

	// Create Stripe payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(stripe.CurrencyUSD),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	paymentIntent, err := paymentintent.New(params)
	if err != nil {
		log.Printf("Stripe error: %v", err)
		h.respondWithError(w, "Payment processing failed", http.StatusInternalServerError)
		return
	}

	log.Printf("Payment intent created successfully: %s", paymentIntent.ID)

	// Send response
	response := models.PaymentResponse{
		ClientSecret: paymentIntent.ClientSecret,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Response encoding error: %v", err)
		h.respondWithError(w, "Internal server error", http.StatusInternalServerError)
	}
}

// calculateOrderAmount calculates the order amount based on product ID
func (h *PaymentHandler) calculateOrderAmount(productId string) int64 {
	switch productId {
	case "Forever Pants":
		return 26000 // $260.00
	case "Forever Shirt":
		return 15500 // $155.00
	case "Forever Shorts":
		return 30000 // $300.00
	default:
		return 0 // Invalid product
	}
}

// respondWithError sends a structured error response
func (h *PaymentHandler) respondWithError(w http.ResponseWriter, message string, statusCode int) {
	respondWithError(w, message, statusCode)
}
