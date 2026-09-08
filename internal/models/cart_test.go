package models

import (
	"testing"
)

func TestCart_TotalCalculation(t *testing.T) {
	cart := Cart{
		Items: []CartItem{
			{ID: 1, Quantity: 2, Subtotal: 2000},
			{ID: 2, Quantity: 1, Subtotal: 1500},
			{ID: 3, Quantity: 3, Subtotal: 3000},
		},
	}

	total := 0
	for _, item := range cart.Items {
		total += item.Subtotal
	}

	if total != 6500 {
		t.Errorf("expected total 6500, got %d", total)
	}
}

func TestCart_EmptyCart(t *testing.T) {
	cart := Cart{
		Items:     []CartItem{},
		Total:     0,
		ItemCount: 0,
	}

	if cart.ItemCount != 0 {
		t.Errorf("expected 0 items, got %d", cart.ItemCount)
	}
	if cart.Total != 0 {
		t.Errorf("expected total 0, got %d", cart.Total)
	}
}

func TestCartItem_Subtotal(t *testing.T) {
	item := CartItem{
		Quantity: 3,
		Product: Product{
			Price: 1500,
		},
	}

	subtotal := item.Product.Price * item.Quantity
	if subtotal != 4500 {
		t.Errorf("expected subtotal 4500, got %d", subtotal)
	}
}

func TestAddToCartRequest_Fields(t *testing.T) {
	variantID := 5
	req := AddToCartRequest{
		ProductID: 10,
		VariantID: &variantID,
		Quantity:  2,
		CustomizationData: map[string]string{
			"color": "red",
			"size":  "XL",
		},
	}

	if req.ProductID != 10 {
		t.Errorf("expected product ID 10, got %d", req.ProductID)
	}
	if req.VariantID == nil || *req.VariantID != 5 {
		t.Error("expected variant ID 5")
	}
	if req.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", req.Quantity)
	}
	if req.CustomizationData["color"] != "red" {
		t.Errorf("expected color red, got %s", req.CustomizationData["color"])
	}
}

func TestUpdateCartItemRequest_Fields(t *testing.T) {
	req := UpdateCartItemRequest{Quantity: 5}
	if req.Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", req.Quantity)
	}
}

func TestCartItem_WithVariant(t *testing.T) {
	stock := 10
	item := CartItem{
		ID:       1,
		Quantity: 1,
		Subtotal: 2500,
		Variant: &ProductVariant{
			ID:    1,
			Label: "Large / Red",
			Stock: stock,
		},
	}

	if item.Variant == nil {
		t.Error("expected variant to be present")
	}
	if item.Variant.Label != "Large / Red" {
		t.Errorf("expected label 'Large / Red', got %s", item.Variant.Label)
	}
}
