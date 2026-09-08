package validation

import (
	"errors"
	"regexp"
	"strings"

	"ecommerce-api-go/internal/models"
)

// PaymentValidator handles payment request validation
type PaymentValidator struct {
	zipRegex *regexp.Regexp
}

// NewPaymentValidator creates a new payment validator
func NewPaymentValidator() *PaymentValidator {
	return &PaymentValidator{
		zipRegex: regexp.MustCompile(`^\d{5}(-\d{4})?$`),
	}
}

// ValidatePaymentRequest validates the incoming payment request
func (v *PaymentValidator) ValidatePaymentRequest(req models.PaymentRequest) error {
	// Required field validation
	if strings.TrimSpace(req.ProductId) == "" {
		return errors.New("product_id is required")
	}

	if strings.TrimSpace(req.FirstName) == "" {
		return errors.New("first_name is required")
	}

	if strings.TrimSpace(req.LastName) == "" {
		return errors.New("last_name is required")
	}

	if strings.TrimSpace(req.Address1) == "" {
		return errors.New("address_1 is required")
	}

	if strings.TrimSpace(req.City) == "" {
		return errors.New("city is required")
	}

	if strings.TrimSpace(req.State) == "" {
		return errors.New("state is required")
	}

	if strings.TrimSpace(req.Country) == "" {
		return errors.New("country is required")
	}

	// Country validation (for now, only US)
	if strings.ToUpper(req.Country) != "US" {
		return errors.New("only US payments are currently supported")
	}

	// ZIP code validation for US
	if !v.zipRegex.MatchString(strings.TrimSpace(req.Zip)) {
		return errors.New("invalid ZIP code format (expected: 12345 or 12345-6789)")
	}

	// State validation (basic - should be 2 characters for US)
	state := strings.TrimSpace(strings.ToUpper(req.State))
	if len(state) != 2 {
		return errors.New("state must be 2 characters (e.g., CA, NY, TX)")
	}

	// Product validation
	if !v.IsValidProduct(req.ProductId) {
		return errors.New("invalid product_id")
	}

	return nil
}

// IsValidProduct checks if the product ID is valid
func (v *PaymentValidator) IsValidProduct(productId string) bool {
	validProducts := map[string]bool{
		"Forever Pants":  true,
		"Forever Shirt":  true,
		"Forever Shorts": true,
	}
	return validProducts[productId]
}
