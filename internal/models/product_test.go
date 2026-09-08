package models

import (
	"strings"
	"testing"
)

func TestValidateProductName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "Shirt", false},
		{"valid name with spaces", "Blue Shirt", false},
		{"empty name", "", true},
		{"whitespace only", "   ", true},
		{"tab only", "\t", true},
		{"name at limit", strings.Repeat("a", 255), false},
		{"name over limit", strings.Repeat("a", 256), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProductName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProductName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateProductPrice(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid price", 1000, false},
		{"min price", 1, false},
		{"max price", 99999999, false},
		{"zero price", 0, true},
		{"negative price", -1, true},
		{"over max price", 100000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProductPrice(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProductPrice(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
