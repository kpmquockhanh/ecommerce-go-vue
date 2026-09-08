package models

// PaymentRequest represents the incoming payment request
type PaymentRequest struct {
	ProductId string `json:"product_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address1  string `json:"address_1"`
	Address2  string `json:"address_2"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip       string `json:"zip"`
	Country   string `json:"country"`
}

// PaymentResponse represents the API response
type PaymentResponse struct {
	ClientSecret string `json:"clientSecret"`
	Error        string `json:"error,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}
