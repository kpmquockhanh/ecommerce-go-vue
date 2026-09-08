package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v82"

	"ecommerce-api-go/internal/queue"
	"ecommerce-api-go/internal/repositories"
)

type WebhookHandler struct {
	webhookSecret string
	orderRepo     repositories.OrderRepository
	cartRepo      repositories.CartRepository
	stockRepo     repositories.StockRepository
	userRepo      repositories.UserRepository
	queue         *queue.Queue
}

func NewWebhookHandler(webhookSecret string, orderRepo repositories.OrderRepository, cartRepo repositories.CartRepository, stockRepo repositories.StockRepository, userRepo repositories.UserRepository, q *queue.Queue) *WebhookHandler {
	return &WebhookHandler{webhookSecret: webhookSecret, orderRepo: orderRepo, cartRepo: cartRepo, stockRepo: stockRepo, userRepo: userRepo, queue: q}
}

func (h *WebhookHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Webhook: failed to read body: %v", err)
		respondWithError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	if sigHeader == "" {
		log.Printf("Webhook: missing Stripe-Signature header")
		respondWithError(w, "Missing signature", http.StatusBadRequest)
		return
	}

	event, err := stripe.ConstructEvent(body, sigHeader, h.webhookSecret)
	if err != nil {
		log.Printf("Webhook: signature verification failed: %v", err)
		respondWithError(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	log.Printf("Webhook: received event %s (type: %s)", event.ID, event.Type)

	ctx := r.Context()

	switch event.Type {
	case "payment_intent.succeeded":
		h.handlePaymentIntentSucceeded(ctx, w, event)
	case "payment_intent.payment_failed":
		h.handlePaymentIntentPaymentFailed(ctx, w, event)
	default:
		log.Printf("Webhook: unhandled event type %s", event.Type)
		respondWithJSON(w, map[string]string{"status": "ignored"}, http.StatusOK)
	}
}

func (h *WebhookHandler) handlePaymentIntentSucceeded(ctx context.Context, w http.ResponseWriter, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		log.Printf("Webhook: failed to parse PaymentIntent: %v", err)
		respondWithError(w, "Invalid event data", http.StatusBadRequest)
		return
	}

	order, err := h.orderRepo.FindByPaymentIntent(ctx, pi.ID)
	if err != nil {
		log.Printf("Webhook: order not found for PI %s: %v", pi.ID, err)
		respondWithJSON(w, map[string]string{"status": "order not found"}, http.StatusOK)
		return
	}

	rowsAffected, err := h.orderRepo.MarkPaid(ctx, order.ID)
	if err != nil {
		log.Printf("Webhook: failed to update order %d: %v", order.ID, err)
		respondWithError(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	log.Printf("Webhook: order %d updated to paid (rows affected: %d)", order.ID, rowsAffected)

	if err := h.cartRepo.ClearByUserID(ctx, order.UserID); err != nil {
		log.Printf("Webhook: failed to clear cart for user %d: %v", order.UserID, err)
	}

	respondWithJSON(w, map[string]string{"status": "processed"}, http.StatusOK)
}

func (h *WebhookHandler) handlePaymentIntentPaymentFailed(ctx context.Context, w http.ResponseWriter, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		log.Printf("Webhook: failed to parse PaymentIntent: %v", err)
		respondWithError(w, "Invalid event data", http.StatusBadRequest)
		return
	}

	order, err := h.orderRepo.FindByPaymentIntent(ctx, pi.ID)
	if err != nil {
		log.Printf("Webhook: order not found for PI %s: %v", pi.ID, err)
		respondWithJSON(w, map[string]string{"status": "order not found"}, http.StatusOK)
		return
	}

	if err := h.orderRepo.MarkPaymentFailed(ctx, order.ID); err != nil {
		log.Printf("Webhook: failed to update order %d: %v", order.ID, err)
		respondWithError(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	log.Printf("Webhook: order %d payment marked as failed", order.ID)

	h.restoreStock(ctx, order.ID)
	h.restoreCart(ctx, order.UserID, order.ID)

	if h.queue != nil {
		user, err := h.userRepo.FindByID(ctx, order.UserID)
		if err == nil {
			jobErr := h.queue.Publish(ctx, "emails", queue.Job{
				Type: queue.JobPaymentFailed,
				Payload: queue.PaymentFailedPayload{
					OrderID:   order.ID,
					UserID:    order.UserID,
					Email:     user.Email,
					FirstName: user.FirstName,
					LastName:  user.LastName,
					Total:     order.Total,
					Reason:    "Payment declined by bank",
				},
			})
			if jobErr != nil {
				log.Printf("Webhook: failed to publish payment failed job: %v", jobErr)
			}
		}
	}

	respondWithJSON(w, map[string]string{"status": "processed"}, http.StatusOK)
}

func (h *WebhookHandler) restoreStock(ctx context.Context, orderID int) {
	items, err := h.orderRepo.GetItemsForStockRestore(ctx, orderID)
	if err != nil {
		log.Printf("Webhook: failed to fetch order items for stock restore: %v", err)
		return
	}

	for _, item := range items {
		if err := h.stockRepo.Increment(ctx, item.VariantID, item.Quantity); err != nil {
			log.Printf("Webhook: failed to restore stock for variant %d: %v", item.VariantID, err)
		} else {
			log.Printf("Webhook: restored %d stock for variant %d", item.Quantity, item.VariantID)
		}
	}
}

func (h *WebhookHandler) restoreCart(ctx context.Context, userID int, orderID int) {
	orderItems, err := h.orderRepo.GetItems(ctx, orderID)
	if err != nil {
		log.Printf("Webhook: failed to fetch order items for cart restore: %v", err)
		return
	}

	var cartItems []repositories.CartItemRow
	for _, item := range orderItems {
		cartItems = append(cartItems, repositories.CartItemRow{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
		})
	}

	if err := h.cartRepo.InsertItems(ctx, userID, cartItems); err != nil {
		log.Printf("Webhook: failed to restore cart for user %d: %v", userID, err)
		return
	}

	log.Printf("Webhook: restored %d cart items for user %d", len(cartItems), userID)
}
