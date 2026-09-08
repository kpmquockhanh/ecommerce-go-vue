package validation

import (
	"testing"

	"ecommerce-api-go/internal/models"
)

func TestValidatePaymentRequest_Success(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Forever Pants",
		FirstName: "John",
		LastName:  "Doe",
		Address1:  "123 Main St",
		City:      "Springfield",
		State:     "IL",
		Zip:       "62701",
		Country:   "US",
	}

	err := v.ValidatePaymentRequest(req)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidatePaymentRequest_MissingProductId(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "",
		FirstName: "John",
		LastName:  "Doe",
		Address1:  "123 Main St",
		City:      "Springfield",
		State:     "IL",
		Zip:       "62701",
		Country:   "US",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for missing product_id")
	}
}

func TestValidatePaymentRequest_MissingFirstName(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Forever Pants",
		FirstName: "",
		LastName:  "Doe",
		Address1:  "123 Main St",
		City:      "Springfield",
		State:     "IL",
		Zip:       "62701",
		Country:   "US",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for missing first_name")
	}
}

func TestValidatePaymentRequest_MissingLastName(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Forever Pants",
		FirstName: "John",
		LastName:  "",
		Address1:  "123 Main St",
		City:      "Springfield",
		State:     "IL",
		Zip:       "62701",
		Country:   "US",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for missing last_name")
	}
}

func TestValidatePaymentRequest_MissingAddress(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Forever Pants",
		FirstName: "John",
		LastName:  "Doe",
		Address1:  "",
		City:      "Springfield",
		State:     "IL",
		Zip:       "62701",
		Country:   "US",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for missing address")
	}
}

func TestValidatePaymentRequest_MissingCity(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Forever Pants",
		FirstName: "John",
		LastName:  "Doe",
		Address1:  "123 Main St",
		City:      "",
		State:     "IL",
		Zip:       "62701",
		Country:   "US",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for missing city")
	}
}

func TestValidatePaymentRequest_MissingState(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Forever Pants",
		FirstName: "John",
		LastName:  "Doe",
		Address1:  "123 Main St",
		City:      "Springfield",
		State:     "",
		Zip:       "62701",
		Country:   "US",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for missing state")
	}
}

func TestValidatePaymentRequest_MissingCountry(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Forever Pants",
		FirstName: "John",
		LastName:  "Doe",
		Address1:  "123 Main St",
		City:      "Springfield",
		State:     "IL",
		Zip:       "62701",
		Country:   "",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for missing country")
	}
}

func TestValidatePaymentRequest_NonUSCountry(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Forever Pants",
		FirstName: "John",
		LastName:  "Doe",
		Address1:  "123 Main St",
		City:      "London",
		State:     "LD",
		Zip:       "SW1A1AA",
		Country:   "UK",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for non-US country")
	}
}

func TestValidatePaymentRequest_InvalidZip(t *testing.T) {
	v := NewPaymentValidator()
	tests := []struct {
		name string
		zip  string
	}{
		{"too short", "123"},
		{"too long", "123456789"},
		{"letters", "ABCDE"},
		{"with special chars", "12345-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.PaymentRequest{
				ProductId: "Forever Pants",
				FirstName: "John",
				LastName:  "Doe",
				Address1:  "123 Main St",
				City:      "Springfield",
				State:     "IL",
				Zip:       tt.zip,
				Country:   "US",
			}

			err := v.ValidatePaymentRequest(req)
			if err == nil {
				t.Errorf("expected error for zip %s", tt.zip)
			}
		})
	}
}

func TestValidatePaymentRequest_ValidZipFormats(t *testing.T) {
	v := NewPaymentValidator()
	tests := []struct {
		name string
		zip  string
	}{
		{"5 digits", "62701"},
		{"9 digits with hyphen", "62701-1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.PaymentRequest{
				ProductId: "Forever Pants",
				FirstName: "John",
				LastName:  "Doe",
				Address1:  "123 Main St",
				City:      "Springfield",
				State:     "IL",
				Zip:       tt.zip,
				Country:   "US",
			}

			err := v.ValidatePaymentRequest(req)
			if err != nil {
				t.Errorf("expected no error for zip %s, got %v", tt.zip, err)
			}
		})
	}
}

func TestValidatePaymentRequest_InvalidState(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Forever Pants",
		FirstName: "John",
		LastName:  "Doe",
		Address1:  "123 Main St",
		City:      "Springfield",
		State:     "Illinois",
		Zip:       "62701",
		Country:   "US",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for state too long")
	}
}

func TestValidatePaymentRequest_InvalidProduct(t *testing.T) {
	v := NewPaymentValidator()
	req := models.PaymentRequest{
		ProductId: "Invalid Product",
		FirstName: "John",
		LastName:  "Doe",
		Address1:  "123 Main St",
		City:      "Springfield",
		State:     "IL",
		Zip:       "62701",
		Country:   "US",
	}

	err := v.ValidatePaymentRequest(req)
	if err == nil {
		t.Error("expected error for invalid product")
	}
}

func TestIsValidProduct(t *testing.T) {
	v := NewPaymentValidator()

	tests := []struct {
		productID string
		valid     bool
	}{
		{"Forever Pants", true},
		{"Forever Shirt", true},
		{"Forever Shorts", true},
		{"Invalid Product", false},
		{"", false},
		{"forever pants", false},
	}

	for _, tt := range tests {
		t.Run(tt.productID, func(t *testing.T) {
			result := v.IsValidProduct(tt.productID)
			if result != tt.valid {
				t.Errorf("IsValidProduct(%q) = %v, want %v", tt.productID, result, tt.valid)
			}
		})
	}
}
