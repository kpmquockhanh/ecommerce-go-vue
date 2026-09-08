package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type Product struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	Description     string     `json:"description"`
	Price           int        `json:"price"`
	CompareAtPrice  *int       `json:"compare_at_price,omitempty"`
	Status          string     `json:"status"`
	Categories      []Category `json:"categories"`
	Tags            []string   `json:"tags"`
	Images          []string   `json:"images"`
	PrimaryImage    string     `json:"primary_image"`
	Weight          *int       `json:"weight,omitempty"`
	Length          *int       `json:"length,omitempty"`
	Width           *int       `json:"width,omitempty"`
	Height          *int       `json:"height,omitempty"`
	MetaTitle       *string    `json:"meta_title,omitempty"`
	MetaDescription *string    `json:"meta_description,omitempty"`
	OgImage         *string    `json:"og_image,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ProductVariant struct {
	ID           int                  `json:"id"`
	ProductID    int                  `json:"product_id"`
	Label        string               `json:"label"`
	OptionValues []VariantOptionValue `json:"option_values"`
	Stock        int                  `json:"stock"`
	SKU          string               `json:"sku"`
	Weight       *int                 `json:"weight,omitempty"`
	Length       *int                 `json:"length,omitempty"`
	Width        *int                 `json:"width,omitempty"`
	Height       *int                 `json:"height,omitempty"`
}

type VariantOptionValue struct {
	OptionGroupID   int    `json:"option_group_id"`
	OptionGroupName string `json:"option_group_name"`
	OptionValueID   int    `json:"option_value_id"`
	Value           string `json:"value"`
}

type ProductWithOptions struct {
	Product
	OptionGroups   []ProductOptionGroup        `json:"option_groups"`
	Customizations []ProductCustomizationField `json:"customizations"`
	Variants       []ProductVariant            `json:"variants"`
	AvgRating      float64                     `json:"avg_rating"`
	ReviewCount    int                         `json:"review_count"`
}

type ProductOptionGroup struct {
	ID        int                  `json:"id"`
	ProductID int                  `json:"product_id"`
	Name      string               `json:"name"`
	SortOrder int                  `json:"sort_order"`
	Values    []ProductOptionValue `json:"values"`
}

type ProductOptionValue struct {
	ID            int    `json:"id"`
	GroupID       int    `json:"group_id"`
	Value         string `json:"value"`
	PriceModifier int    `json:"price_modifier"`
}

type ProductCustomizationField struct {
	ID           int    `json:"id"`
	ProductID    int    `json:"product_id"`
	Name         string `json:"name"`
	FieldType    string `json:"field_type"`
	Required     bool   `json:"required"`
	DefaultValue string `json:"default_value,omitempty"`
	SortOrder    int    `json:"sort_order"`
}

type ProductListItem struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description"`
	Price           int       `json:"price"`
	CompareAtPrice  *int      `json:"compare_at_price,omitempty"`
	Status          string    `json:"status"`
	CategoryNames   []string  `json:"category_names"`
	Tags            []string  `json:"tags"`
	Images          []string  `json:"images"`
	PrimaryImage    string    `json:"primary_image"`
	Weight          *int      `json:"weight,omitempty"`
	Length          *int      `json:"length,omitempty"`
	Width           *int      `json:"width,omitempty"`
	Height          *int      `json:"height,omitempty"`
	MetaTitle       *string   `json:"meta_title,omitempty"`
	MetaDescription *string   `json:"meta_description,omitempty"`
	OgImage         *string   `json:"og_image,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProductListResponse struct {
	Products   []ProductListItem `json:"products"`
	Pagination Pagination        `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type CreateProductRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Price           int      `json:"price"`
	CompareAtPrice  *int     `json:"compare_at_price,omitempty"`
	Images          []string `json:"images"`
	CategoryIDs     []int    `json:"category_ids"`
	Tags            []string `json:"tags"`
	Weight          *int     `json:"weight,omitempty"`
	Length          *int     `json:"length,omitempty"`
	Width           *int     `json:"width,omitempty"`
	Height          *int     `json:"height,omitempty"`
	MetaTitle       *string  `json:"meta_title,omitempty"`
	MetaDescription *string  `json:"meta_description,omitempty"`
	OgImage         *string  `json:"og_image,omitempty"`
}

type UpdateProductRequest struct {
	Name            *string   `json:"name"`
	Description     *string   `json:"description"`
	Price           *int      `json:"price"`
	CompareAtPrice  *int      `json:"compare_at_price,omitempty"`
	Images          *[]string `json:"images"`
	CategoryIDs     *[]int    `json:"category_ids"`
	Tags            *[]string `json:"tags"`
	Status          *string   `json:"status"`
	Weight          *int      `json:"weight,omitempty"`
	Length          *int      `json:"length,omitempty"`
	Width           *int      `json:"width,omitempty"`
	Height          *int      `json:"height,omitempty"`
	MetaTitle       *string   `json:"meta_title,omitempty"`
	MetaDescription *string   `json:"meta_description,omitempty"`
	OgImage         *string   `json:"og_image,omitempty"`
}

type CreateVariantRequest struct {
	Stock        int                               `json:"stock"`
	SKU          string                            `json:"sku"`
	Weight       *int                              `json:"weight,omitempty"`
	Length       *int                              `json:"length,omitempty"`
	Width        *int                              `json:"width,omitempty"`
	Height       *int                              `json:"height,omitempty"`
	OptionValues []CreateVariantOptionValueRequest `json:"option_values"`
}

type CreateVariantOptionValueRequest struct {
	OptionValueID int `json:"option_value_id"`
}

type UpdateVariantRequest struct {
	Stock        *int                              `json:"stock"`
	SKU          *string                           `json:"sku"`
	Weight       *int                              `json:"weight,omitempty"`
	Length       *int                              `json:"length,omitempty"`
	Width        *int                              `json:"width,omitempty"`
	Height       *int                              `json:"height,omitempty"`
	OptionValues []UpdateVariantOptionValueRequest `json:"option_values,omitempty"`
}

type UpdateVariantOptionValueRequest struct {
	OptionValueID int `json:"option_value_id"`
}

// Option Group / Value CRUD DTOs

type CreateOptionGroupRequest struct {
	Name   string                     `json:"name"`
	Values []CreateOptionValueRequest `json:"values"`
}

type CreateOptionValueRequest struct {
	Value         string `json:"value"`
	PriceModifier *int   `json:"price_modifier,omitempty"`
}

type UpdateOptionGroupRequest struct {
	Name      *string `json:"name,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}

type UpdateOptionValueRequest struct {
	Value         *string `json:"value,omitempty"`
	PriceModifier *int    `json:"price_modifier,omitempty"`
	SortOrder     *int    `json:"sort_order,omitempty"`
}

type ReorderOptionValuesRequest struct {
	ValueIDs []int `json:"value_ids"`
}

type GenerateVariantsResponse struct {
	Created      []ProductVariant `json:"created"`
	CreatedCount int              `json:"created_count"`
	SkippedCount int              `json:"skipped_count"`
	Error        string           `json:"error,omitempty"`
}

type BulkUpdateVariantItem struct {
	VariantID    int             `json:"variant_id"`
	Stock        *int            `json:"stock,omitempty"`
	SKU          *string         `json:"sku,omitempty"`
	Weight       *int            `json:"weight,omitempty"`
	Length       *int            `json:"length,omitempty"`
	Width        *int            `json:"width,omitempty"`
	Height       *int            `json:"height,omitempty"`
	OptionValues json.RawMessage `json:"option_values,omitempty"`
}

type BulkUpdateVariantsRequest struct {
	Items []BulkUpdateVariantItem `json:"items"`
}

type BulkUpdateVariantResult struct {
	VariantID int    `json:"variant_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

type BulkUpdateVariantsResponse struct {
	Results        []BulkUpdateVariantResult `json:"results"`
	SucceededCount int                       `json:"succeeded_count"`
	FailedCount    int                       `json:"failed_count"`
}

// Customization Field CRUD DTOs

type CreateCustomizationFieldRequest struct {
	Name         string `json:"name"`
	FieldType    string `json:"field_type"`
	Required     *bool  `json:"required,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
	SortOrder    *int   `json:"sort_order,omitempty"`
}

type UpdateCustomizationFieldRequest struct {
	Name         *string `json:"name,omitempty"`
	FieldType    *string `json:"field_type,omitempty"`
	Required     *bool   `json:"required,omitempty"`
	DefaultValue *string `json:"default_value,omitempty"`
	SortOrder    *int    `json:"sort_order,omitempty"`
}

type ReorderCustomizationFieldsRequest struct {
	FieldIDs []int `json:"field_ids"`
}

type ProductImageOrder struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	ImagePath string `json:"image_path"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

type UpdateImageOrderRequest struct {
	Images []ImageOrderEntry `json:"images"`
}

type ImageOrderEntry struct {
	ImagePath string `json:"image_path"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

func ValidateProductName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if utf8.RuneCountInString(name) > 255 {
		return fmt.Errorf("name must be 255 characters or fewer")
	}
	return nil
}

func ValidateProductPrice(price int) error {
	if price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if price > 99999999 {
		return fmt.Errorf("price must be 99999999 or fewer")
	}
	return nil
}

func ValidateTags(tags []string) error {
	if len(tags) > 12 {
		return fmt.Errorf("maximum 12 tags allowed")
	}
	for _, tag := range tags {
		if len(strings.TrimSpace(tag)) == 0 {
			return fmt.Errorf("tag cannot be empty")
		}
		if len(tag) > 50 {
			return fmt.Errorf("tag must be 50 characters or fewer")
		}
	}
	return nil
}
