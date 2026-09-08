package models

type CartItem struct {
	ID               int                `json:"id"`
	Product          Product            `json:"product"`
	Variant          *ProductVariant    `json:"variant"`
	Quantity         int                `json:"quantity"`
	Subtotal         int                `json:"subtotal"`
	CustomizationData map[string]string `json:"customization_data,omitempty"`
}

type Cart struct {
	ID        int        `json:"id"`
	Items     []CartItem `json:"items"`
	Total     int        `json:"total"`
	ItemCount int        `json:"item_count"`
}

type AddToCartRequest struct {
	ProductID         int               `json:"product_id"`
	VariantID         *int              `json:"variant_id"`
	Quantity          int               `json:"quantity"`
	CustomizationData map[string]string `json:"customization_data,omitempty"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"`
}
