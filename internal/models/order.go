package models

import "time"

type Order struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	Status           string    `json:"status"`
	Total            int       `json:"total"`
	ShippingAddress  Address   `json:"shipping_address"`
	StripePaymentIntent string `json:"stripe_payment_intent"`
	PaymentStatus    string    `json:"payment_status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID           int    `json:"id"`
	ProductID    int    `json:"product_id"`
	VariantID    *int   `json:"variant_id"`
	ProductName  string `json:"product_name"`
	VariantLabel string `json:"variant_label"`
	Quantity     int    `json:"quantity"`
	Price        int    `json:"price"`
	Subtotal     int    `json:"subtotal"`
}

type OrderWithItems struct {
	Order
	Items []OrderItem `json:"items"`
}

type OrderListResponse struct {
	Orders     []Order    `json:"orders"`
	Pagination Pagination `json:"pagination"`
}

type Address struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address1  string `json:"address_1"`
	Address2  string `json:"address_2"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip       string `json:"zip"`
	Country   string `json:"country"`
}

type CheckoutInitRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
}

type CheckoutInitResponse struct {
	ClientSecret string `json:"client_secret"`
}

type CheckoutConfirmRequest struct {
	ShippingAddress Address `json:"shipping_address"`
	IdempotencyKey  string  `json:"idempotency_key"`
}

type CheckoutConfirmResponse struct {
	OrderID int `json:"order_id"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}
