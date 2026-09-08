package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/text/unicode/norm"

	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/repositories"
	"ecommerce-api-go/internal/storage"
)

type ProductHandler struct {
	productRepo  repositories.ProductRepository
	reviewRepo   repositories.ReviewRepository
	categoryRepo repositories.CategoryRepository
}

func NewProductHandler(productRepo repositories.ProductRepository, reviewRepo repositories.ReviewRepository, categoryRepo repositories.CategoryRepository) *ProductHandler {
	return &ProductHandler{productRepo: productRepo, reviewRepo: reviewRepo, categoryRepo: categoryRepo}
}

// maxVariantsPerProduct bounds how many variants (manually created or
// generated) a single product may accumulate, to prevent cartesian-product
// blowup from option groups with many values.
const maxVariantsPerProduct = 500

func slugify(s string, fallback string) string {
	slug := strings.ToLower(strings.ReplaceAll(s, " ", "-"))
	slug = norm.NFKD.String(slug)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = fallback
	}
	return slug
}

// presignProductImages converts stored S3 object keys into browser-usable
// presigned URLs for public product responses. It leaves keys untouched when
// storage isn't configured or a given key fails to presign, so image order
// and count (relied on for index-based image switching in the storefront)
// stay stable.
func presignProductImages(ctx context.Context, images []string, primary string) ([]string, string) {
	if storage.Client == nil {
		return images, primary
	}

	presigned := make([]string, len(images))
	for i, img := range images {
		url, err := storage.GetPresignedURL(ctx, img)
		if err != nil {
			log.Printf("failed to presign product image %q: %v", img, err)
			presigned[i] = img
			continue
		}
		presigned[i] = url
	}

	primaryURL := primary
	if primary != "" {
		if url, err := storage.GetPresignedURL(ctx, primary); err != nil {
			log.Printf("failed to presign primary product image %q: %v", primary, err)
		} else {
			primaryURL = url
		}
	}

	return presigned, primaryURL
}

func (h *ProductHandler) generateSlug(ctx context.Context, name string) (string, error) {
	slug := slugify(name, "product")
	candidate := slug
	for i := 2; i < 10000; i++ {
		existing, _ := h.productRepo.FindBySlug(ctx, candidate)
		if existing == nil {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", slug, i)
	}
	return "", fmt.Errorf("could not generate unique slug")
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	filters := repositories.ProductFilters{
		Limit:  limit,
		Offset: offset,
	}

	if categoriesParam := query.Get("categories"); categoriesParam != "" {
		parts := strings.Split(categoriesParam, ",")
		var ids []int
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if id, err := strconv.Atoi(part); err == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			filters.CategoryIDs = ids
		}
	}
	if tagsParam := query.Get("tags"); tagsParam != "" {
		parts := strings.Split(tagsParam, ",")
		var tags []string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				tags = append(tags, strings.ToLower(part))
			}
		}
		if len(tags) > 0 {
			filters.Tags = tags
		}
	}
	if minPrice := query.Get("min_price"); minPrice != "" {
		if v, err := strconv.Atoi(minPrice); err == nil {
			filters.MinPrice = &v
		}
	}
	if maxPrice := query.Get("max_price"); maxPrice != "" {
		if v, err := strconv.Atoi(maxPrice); err == nil {
			filters.MaxPrice = &v
		}
	}
	if search := query.Get("search"); search != "" {
		filters.Search = search
	}
	filters.Sort = query.Get("sort")

	products, total, err := h.productRepo.List(r.Context(), filters)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	var enriched []models.ProductListItem
	for _, p := range products {
		p.Images, p.PrimaryImage = presignProductImages(r.Context(), p.Images, p.PrimaryImage)
		item := models.ProductListItem{
			ID:              p.ID,
			Name:            p.Name,
			Slug:            p.Slug,
			Description:     p.Description,
			Price:           p.Price,
			CompareAtPrice:  p.CompareAtPrice,
			Status:          p.Status,
			Tags:            p.Tags,
			Images:          p.Images,
			PrimaryImage:    p.PrimaryImage,
			Weight:          p.Weight,
			Length:          p.Length,
			Width:           p.Width,
			Height:          p.Height,
			MetaTitle:       p.MetaTitle,
			MetaDescription: p.MetaDescription,
			OgImage:         p.OgImage,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
		}

		categories, err := h.productRepo.GetCategories(r.Context(), p.ID)
		if err == nil {
			for _, c := range categories {
				item.CategoryNames = append(item.CategoryNames, c.Name)
			}
		}
		if item.CategoryNames == nil {
			item.CategoryNames = []string{}
		}

		enriched = append(enriched, item)
	}

	if enriched == nil {
		enriched = []models.ProductListItem{}
	}

	respondWithJSON(w, models.ProductListResponse{
		Products: enriched,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, http.StatusOK)
}

func (h *ProductHandler) AdminListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	filters := repositories.ProductFilters{
		Limit:  limit,
		Offset: offset,
	}

	if status := query.Get("status"); status != "" {
		filters.Status = status
	}
	if categoriesParam := query.Get("categories"); categoriesParam != "" {
		parts := strings.Split(categoriesParam, ",")
		var ids []int
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if id, err := strconv.Atoi(part); err == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			filters.CategoryIDs = ids
		}
	}
	if tagsParam := query.Get("tags"); tagsParam != "" {
		parts := strings.Split(tagsParam, ",")
		var tags []string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				tags = append(tags, strings.ToLower(part))
			}
		}
		if len(tags) > 0 {
			filters.Tags = tags
		}
	}
	if minPrice := query.Get("min_price"); minPrice != "" {
		if v, err := strconv.Atoi(minPrice); err == nil {
			filters.MinPrice = &v
		}
	}
	if maxPrice := query.Get("max_price"); maxPrice != "" {
		if v, err := strconv.Atoi(maxPrice); err == nil {
			filters.MaxPrice = &v
		}
	}
	if search := query.Get("search"); search != "" {
		filters.Search = search
	}
	filters.Sort = query.Get("sort")

	products, total, err := h.productRepo.AdminList(r.Context(), filters)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	var enriched []models.ProductListItem
	for _, p := range products {
		item := models.ProductListItem{
			ID:              p.ID,
			Name:            p.Name,
			Slug:            p.Slug,
			Description:     p.Description,
			Price:           p.Price,
			CompareAtPrice:  p.CompareAtPrice,
			Status:          p.Status,
			Tags:            p.Tags,
			Images:          p.Images,
			PrimaryImage:    p.PrimaryImage,
			Weight:          p.Weight,
			Length:          p.Length,
			Width:           p.Width,
			Height:          p.Height,
			MetaTitle:       p.MetaTitle,
			MetaDescription: p.MetaDescription,
			OgImage:         p.OgImage,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
		}

		categories, err := h.productRepo.GetCategories(r.Context(), p.ID)
		if err == nil {
			for _, c := range categories {
				item.CategoryNames = append(item.CategoryNames, c.Name)
			}
		}
		if item.CategoryNames == nil {
			item.CategoryNames = []string{}
		}

		enriched = append(enriched, item)
	}

	if enriched == nil {
		enriched = []models.ProductListItem{}
	}

	respondWithJSON(w, models.ProductListResponse{
		Products: enriched,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, http.StatusOK)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/api/products/")
	if slug == "" {
		respondWithError(w, "Product slug is required", http.StatusBadRequest)
		return
	}

	p, err := h.productRepo.FindBySlug(r.Context(), slug)
	if err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	variants, err := h.productRepo.GetVariants(r.Context(), p.ID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if variants == nil {
		variants = []models.ProductVariant{}
	}

	optionGroups, err := h.productRepo.GetOptionGroupsByProduct(r.Context(), p.ID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if optionGroups == nil {
		optionGroups = []models.ProductOptionGroup{}
	}

	customizations, err := h.productRepo.GetCustomizationFieldsByProduct(r.Context(), p.ID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if customizations == nil {
		customizations = []models.ProductCustomizationField{}
	}

	avgRating, reviewCount, _ := h.reviewRepo.GetStats(r.Context(), p.ID)

	p.Images, p.PrimaryImage = presignProductImages(r.Context(), p.Images, p.PrimaryImage)

	result := models.ProductWithOptions{
		Product:        *p,
		OptionGroups:   optionGroups,
		Customizations: customizations,
		Variants:       variants,
		AvgRating:      avgRating,
		ReviewCount:    reviewCount,
	}

	respondWithJSON(w, result, http.StatusOK)
}

func (h *ProductHandler) AdminGetProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := h.extractProductID(r.URL.Path, "/api/admin/products/", "")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	p, err := h.productRepo.FindByID(r.Context(), id)
	if err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	variants, err := h.productRepo.GetVariants(r.Context(), p.ID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if variants == nil {
		variants = []models.ProductVariant{}
	}

	optionGroups, err := h.productRepo.GetOptionGroupsByProduct(r.Context(), p.ID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if optionGroups == nil {
		optionGroups = []models.ProductOptionGroup{}
	}

	customizations, err := h.productRepo.GetCustomizationFieldsByProduct(r.Context(), p.ID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if customizations == nil {
		customizations = []models.ProductCustomizationField{}
	}

	avgRating, reviewCount, _ := h.reviewRepo.GetStats(r.Context(), p.ID)

	result := models.ProductWithOptions{
		Product:        *p,
		OptionGroups:   optionGroups,
		Customizations: customizations,
		Variants:       variants,
		AvgRating:      avgRating,
		ReviewCount:    reviewCount,
	}

	respondWithJSON(w, result, http.StatusOK)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if err := models.ValidateProductName(req.Name); err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := models.ValidateProductPrice(req.Price); err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Tags != nil {
		if err := models.ValidateTags(req.Tags); err != nil {
			respondWithError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if len(req.CategoryIDs) > 0 {
		for _, catID := range req.CategoryIDs {
			if _, err := h.categoryRepo.FindByID(r.Context(), catID); err != nil {
				respondWithError(w, "category not found", http.StatusBadRequest)
				return
			}
		}
	}

	slug, err := h.generateSlug(r.Context(), req.Name)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	if req.Images == nil {
		req.Images = []string{}
	}
	if req.CategoryIDs == nil {
		req.CategoryIDs = []int{}
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}

	p, err := h.productRepo.Create(r.Context(), &req, slug)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, p, http.StatusCreated)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
	idStr = strings.TrimSuffix(idStr, "/")
	// Handle nested paths like /api/admin/products/1/variants
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		req.Name = &name
		if err := models.ValidateProductName(*req.Name); err != nil {
			respondWithError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if req.Price != nil {
		if err := models.ValidateProductPrice(*req.Price); err != nil {
			respondWithError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		req.Description = &desc
	}
	if req.Tags != nil {
		if err := models.ValidateTags(*req.Tags); err != nil {
			respondWithError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if req.CategoryIDs != nil {
		for _, catID := range *req.CategoryIDs {
			if _, err := h.categoryRepo.FindByID(r.Context(), catID); err != nil {
				respondWithError(w, "category not found", http.StatusBadRequest)
				return
			}
		}
	}

	p, err := h.productRepo.Update(r.Context(), id, &req)
	if err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Product not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, p, http.StatusOK)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.SoftDelete(r.Context(), id); err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Product not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Product deleted"}, http.StatusOK)
}

// Variant CRUD (Story 2)

func (h *ProductHandler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/admin/products/", "/variants")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req models.CreateVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	req.SKU = strings.TrimSpace(req.SKU)

	if req.SKU == "" {
		respondWithError(w, "SKU is required", http.StatusBadRequest)
		return
	}

	if _, err := h.productRepo.FindByID(r.Context(), productID); err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	count, err := h.productRepo.CountVariants(r.Context(), productID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if count >= maxVariantsPerProduct {
		respondWithError(w, fmt.Sprintf("Maximum %d variants per product", maxVariantsPerProduct), http.StatusBadRequest)
		return
	}

	// Validate option values belong to this product and cover every option group exactly once
	ids := make([]int, len(req.OptionValues))
	for i, ov := range req.OptionValues {
		ids[i] = ov.OptionValueID
	}
	if err := h.productRepo.ValidateOptionValues(r.Context(), productID, ids); err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check for duplicate combination
	exists, err := h.productRepo.HasVariantCombination(r.Context(), productID, ids, 0)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if exists {
		respondWithError(w, "Variant with this option combination already exists", http.StatusConflict)
		return
	}

	variant, err := h.productRepo.CreateVariant(r.Context(), productID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			respondWithError(w, "SKU already exists", http.StatusConflict)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, variant, http.StatusCreated)
}

// variantCombo is one cartesian-product combination of option values, in
// option-group order.
type variantCombo struct {
	optionValues []models.CreateVariantOptionValueRequest
	optionIDs    []int
	labels       []string
}

// cartesianCombinations expands a product's option groups into every
// combination of one value per group, in group and value order.
func cartesianCombinations(groups []models.ProductOptionGroup) []variantCombo {
	combos := []variantCombo{{}}
	for _, g := range groups {
		var next []variantCombo
		for _, combo := range combos {
			for _, v := range g.Values {
				next = append(next, variantCombo{
					optionValues: append(append([]models.CreateVariantOptionValueRequest{}, combo.optionValues...), models.CreateVariantOptionValueRequest{OptionValueID: v.ID}),
					optionIDs:    append(append([]int{}, combo.optionIDs...), v.ID),
					labels:       append(append([]string{}, combo.labels...), v.Value),
				})
			}
		}
		combos = next
	}
	return combos
}

// createGeneratedVariant creates a variant for combo with a system-generated
// SKU (product slug + slugified option values), retrying with a numeric
// suffix on SKU collision.
func (h *ProductHandler) createGeneratedVariant(ctx context.Context, productID int, productSlug string, combo variantCombo) (*models.ProductVariant, error) {
	parts := make([]string, 0, len(combo.labels)+1)
	parts = append(parts, productSlug)
	for _, label := range combo.labels {
		parts = append(parts, slugify(label, "value"))
	}
	base := strings.Join(parts, "-")
	// sku is VARCHAR(100); leave room for a "-NNNN" collision suffix.
	if len(base) > 95 {
		base = strings.TrimRight(base[:95], "-")
	}

	req := &models.CreateVariantRequest{
		Stock:        0,
		OptionValues: combo.optionValues,
		SKU:          base,
	}

	for i := 2; i < 10000; i++ {
		variant, err := h.productRepo.CreateVariant(ctx, productID, req)
		if err == nil {
			return variant, nil
		}
		if !strings.Contains(err.Error(), "unique") {
			return nil, err
		}
		req.SKU = fmt.Sprintf("%s-%d", base, i)
	}
	return nil, fmt.Errorf("could not generate unique SKU for combination")
}

func (h *ProductHandler) GenerateVariants(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/admin/products/", "/variants/generate")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.productRepo.FindByID(r.Context(), productID)
	if err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	groups, err := h.productRepo.GetOptionGroupsByProduct(r.Context(), productID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if len(groups) == 0 {
		respondWithError(w, "Product has no option groups to generate variants from", http.StatusBadRequest)
		return
	}

	var total int64 = 1
	for _, g := range groups {
		if len(g.Values) == 0 {
			respondWithError(w, "Every option group must have at least one value to generate variants", http.StatusBadRequest)
			return
		}
		total *= int64(len(g.Values))
		if total > maxVariantsPerProduct {
			break
		}
	}
	if total > maxVariantsPerProduct {
		respondWithError(w, fmt.Sprintf("Generating all option combinations (%d) would exceed the maximum of %d variants per product", total, maxVariantsPerProduct), http.StatusBadRequest)
		return
	}

	existingCount, err := h.productRepo.CountVariants(r.Context(), productID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	var missing []variantCombo
	skipped := 0
	for _, combo := range cartesianCombinations(groups) {
		exists, err := h.productRepo.HasVariantCombination(r.Context(), productID, combo.optionIDs, 0)
		if err != nil {
			respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
			return
		}
		if exists {
			skipped++
			continue
		}
		missing = append(missing, combo)
	}

	if existingCount+len(missing) > maxVariantsPerProduct {
		respondWithError(w, fmt.Sprintf("Generating %d new variants would exceed the maximum of %d variants per product", len(missing), maxVariantsPerProduct), http.StatusBadRequest)
		return
	}

	created := make([]models.ProductVariant, 0, len(missing))
	for _, combo := range missing {
		variant, err := h.createGeneratedVariant(r.Context(), productID, product.Slug, combo)
		if err != nil {
			log.Printf("ERROR [%d]: generate variants: %v", http.StatusInternalServerError, err)
			respondWithJSON(w, models.GenerateVariantsResponse{
				Created:      created,
				CreatedCount: len(created),
				SkippedCount: skipped,
				Error:        fmt.Sprintf("stopped after creating %d of %d new variants: internal server error", len(created), len(missing)),
			}, http.StatusInternalServerError)
			return
		}
		created = append(created, *variant)
	}

	respondWithJSON(w, models.GenerateVariantsResponse{
		Created:      created,
		CreatedCount: len(created),
		SkippedCount: skipped,
	}, http.StatusOK)
}

func (h *ProductHandler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	variantID, err := h.extractID(r.URL.Path, "/variants/")
	if err != nil {
		respondWithError(w, "Invalid variant ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.SKU != nil {
		s := strings.TrimSpace(*req.SKU)
		if s == "" {
			respondWithError(w, "SKU cannot be empty", http.StatusBadRequest)
			return
		}
		req.SKU = &s
	}

	if req.OptionValues != nil {
		productID, err := h.extractIDBetween(r.URL.Path, "/products/")
		if err != nil {
			respondWithError(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		ids := make([]int, len(req.OptionValues))
		for i, ov := range req.OptionValues {
			ids[i] = ov.OptionValueID
		}
		if err := h.productRepo.ValidateOptionValues(r.Context(), productID, ids); err != nil {
			respondWithError(w, err.Error(), http.StatusBadRequest)
			return
		}

		exists, err := h.productRepo.HasVariantCombination(r.Context(), productID, ids, variantID)
		if err != nil {
			respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
			return
		}
		if exists {
			respondWithError(w, "Variant with this option combination already exists", http.StatusConflict)
			return
		}
	}

	variant, err := h.productRepo.UpdateVariant(r.Context(), variantID, &req)
	if err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Variant not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "unique") {
			respondWithError(w, "SKU already exists", http.StatusConflict)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, variant, http.StatusOK)
}

func (h *ProductHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	variantID, err := h.extractID(r.URL.Path, "/variants/")
	if err != nil {
		respondWithError(w, "Invalid variant ID", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.DeleteVariant(r.Context(), variantID); err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Variant not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Variant deleted"}, http.StatusOK)
}

func (h *ProductHandler) BulkUpdateVariants(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/admin/products/", "/variants/bulk")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if _, err := h.productRepo.FindByID(r.Context(), productID); err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	var req models.BulkUpdateVariantsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		respondWithError(w, "items must not be empty", http.StatusBadRequest)
		return
	}
	if len(req.Items) > maxVariantsPerProduct {
		respondWithError(w, fmt.Sprintf("Cannot bulk-update more than %d items at once", maxVariantsPerProduct), http.StatusBadRequest)
		return
	}

	for _, item := range req.Items {
		if item.OptionValues != nil {
			respondWithError(w, "Bulk update does not support option_values; use the single-variant update endpoint", http.StatusBadRequest)
			return
		}
	}

	results := make([]models.BulkUpdateVariantResult, 0, len(req.Items))
	succeeded, failed := 0, 0
	for _, item := range req.Items {
		result := h.applyBulkVariantUpdate(r.Context(), productID, item)
		results = append(results, result)
		if result.Success {
			succeeded++
		} else {
			failed++
		}
	}

	respondWithJSON(w, models.BulkUpdateVariantsResponse{
		Results:        results,
		SucceededCount: succeeded,
		FailedCount:    failed,
	}, http.StatusOK)
}

// applyBulkVariantUpdate validates and applies one item of a bulk update
// request, never returning an error itself — every outcome, including
// failure, is reported in the returned result.
func (h *ProductHandler) applyBulkVariantUpdate(ctx context.Context, productID int, item models.BulkUpdateVariantItem) models.BulkUpdateVariantResult {
	if item.VariantID <= 0 {
		return models.BulkUpdateVariantResult{VariantID: item.VariantID, Error: "variant_id is required"}
	}

	ownerID, err := h.productRepo.GetVariantProductID(ctx, item.VariantID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return models.BulkUpdateVariantResult{VariantID: item.VariantID, Error: "variant not found"}
		}
		log.Printf("ERROR: bulk update variant %d lookup: %v", item.VariantID, err)
		return models.BulkUpdateVariantResult{VariantID: item.VariantID, Error: "internal server error"}
	}
	if ownerID != productID {
		return models.BulkUpdateVariantResult{VariantID: item.VariantID, Error: "variant does not belong to this product"}
	}

	update := &models.UpdateVariantRequest{
		Stock:  item.Stock,
		Weight: item.Weight,
		Length: item.Length,
		Width:  item.Width,
		Height: item.Height,
	}
	if item.SKU != nil {
		sku := strings.TrimSpace(*item.SKU)
		if sku == "" {
			return models.BulkUpdateVariantResult{VariantID: item.VariantID, Error: "SKU cannot be empty"}
		}
		update.SKU = &sku
	}

	if _, err := h.productRepo.UpdateVariant(ctx, item.VariantID, update); err != nil {
		if err == repositories.ErrNotFound {
			return models.BulkUpdateVariantResult{VariantID: item.VariantID, Error: "variant not found"}
		}
		if strings.Contains(err.Error(), "unique") {
			return models.BulkUpdateVariantResult{VariantID: item.VariantID, Error: "SKU already exists"}
		}
		log.Printf("ERROR: bulk update variant %d: %v", item.VariantID, err)
		return models.BulkUpdateVariantResult{VariantID: item.VariantID, Error: "internal server error"}
	}

	return models.BulkUpdateVariantResult{VariantID: item.VariantID, Success: true}
}

// Tags (Story 3)

func (h *ProductHandler) AddTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/admin/products/", "/tags")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := models.ValidateTags(req.Tags); err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := h.productRepo.FindByID(r.Context(), productID); err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	if err := h.productRepo.AddTags(r.Context(), productID, req.Tags); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	tags, _ := h.productRepo.GetTags(r.Context(), productID)
	respondWithJSON(w, map[string]interface{}{"tags": tags}, http.StatusOK)
}

func (h *ProductHandler) RemoveTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
	parts := strings.Split(path, "/tags/")
	if len(parts) != 2 {
		respondWithError(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(parts[0])
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	tag := strings.TrimSpace(parts[1])
	if tag == "" {
		respondWithError(w, "Tag is required", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.RemoveTag(r.Context(), productID, tag); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	tags, _ := h.productRepo.GetTags(r.Context(), productID)
	respondWithJSON(w, map[string]interface{}{"tags": tags}, http.StatusOK)
}

// Image ordering (Story 5)

func (h *ProductHandler) UpdateImageOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/admin/products/", "/images/order")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateImageOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	primaryCount := 0
	for _, img := range req.Images {
		if img.IsPrimary {
			primaryCount++
		}
	}
	if primaryCount > 1 {
		respondWithError(w, "Only one image can be primary", http.StatusBadRequest)
		return
	}

	if _, err := h.productRepo.FindByID(r.Context(), productID); err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	if err := h.productRepo.UpdateImageOrder(r.Context(), productID, req.Images); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	orders, _ := h.productRepo.GetImageOrder(r.Context(), productID)
	respondWithJSON(w, map[string]interface{}{"images": orders}, http.StatusOK)
}

// Option Groups & Values (Story 2)

func (h *ProductHandler) GetOptionGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/products/", "/options")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if _, err := h.productRepo.FindByID(r.Context(), productID); err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	groups, err := h.productRepo.GetOptionGroupsByProduct(r.Context(), productID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if groups == nil {
		groups = []models.ProductOptionGroup{}
	}

	respondWithJSON(w, groups, http.StatusOK)
}

func (h *ProductHandler) CreateOptionGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/admin/products/", "/option-groups")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if _, err := h.productRepo.FindByID(r.Context(), productID); err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	var req models.CreateOptionGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respondWithError(w, "Option group name is required", http.StatusBadRequest)
		return
	}

	groups, _ := h.productRepo.GetOptionGroupsByProduct(r.Context(), productID)
	if len(groups) >= 6 {
		respondWithError(w, "Maximum 6 option groups per product", http.StatusBadRequest)
		return
	}

	for _, g := range groups {
		if g.Name == req.Name {
			respondWithError(w, "Option group name already exists", http.StatusConflict)
			return
		}
	}

	group, err := h.productRepo.CreateOptionGroup(r.Context(), productID, &req)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, group, http.StatusCreated)
}

func (h *ProductHandler) UpdateOptionGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groupID, err := h.extractID(r.URL.Path, "/option-groups/")
	if err != nil {
		respondWithError(w, "Invalid option group ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateOptionGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		req.Name = &name
		if name == "" {
			respondWithError(w, "Option group name cannot be empty", http.StatusBadRequest)
			return
		}
	}

	group, err := h.productRepo.UpdateOptionGroup(r.Context(), groupID, &req)
	if err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Option group not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, group, http.StatusOK)
}

func (h *ProductHandler) DeleteOptionGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groupID, err := h.extractID(r.URL.Path, "/option-groups/")
	if err != nil {
		respondWithError(w, "Invalid option group ID", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.DeleteOptionGroup(r.Context(), groupID); err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Option group not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Option group deleted"}, http.StatusOK)
}

func (h *ProductHandler) CreateOptionValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groupID, err := h.extractIDBetween(r.URL.Path, "/option-groups/")
	if err != nil {
		respondWithError(w, "Invalid option group ID", http.StatusBadRequest)
		return
	}

	var req models.CreateOptionValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	req.Value = strings.TrimSpace(req.Value)
	if req.Value == "" {
		respondWithError(w, "Option value is required", http.StatusBadRequest)
		return
	}

	v, err := h.productRepo.CreateOptionValue(r.Context(), groupID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			respondWithError(w, "Option value already exists in this group", http.StatusConflict)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, v, http.StatusCreated)
}

func (h *ProductHandler) UpdateOptionValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	valueID, err := h.extractID(r.URL.Path, "/values/")
	if err != nil {
		respondWithError(w, "Invalid option value ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateOptionValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Value != nil {
		v := strings.TrimSpace(*req.Value)
		req.Value = &v
		if v == "" {
			respondWithError(w, "Option value cannot be empty", http.StatusBadRequest)
			return
		}
	}

	v, err := h.productRepo.UpdateOptionValue(r.Context(), valueID, &req)
	if err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Option value not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "unique") {
			respondWithError(w, "Option value already exists in this group", http.StatusConflict)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, v, http.StatusOK)
}

func (h *ProductHandler) DeleteOptionValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	valueID, err := h.extractID(r.URL.Path, "/values/")
	if err != nil {
		respondWithError(w, "Invalid option value ID", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.DeleteOptionValue(r.Context(), valueID); err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Option value not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Option value deleted"}, http.StatusOK)
}

func (h *ProductHandler) ReorderOptionValues(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groupID, err := h.extractIDBetween(r.URL.Path, "/option-groups/")
	if err != nil {
		respondWithError(w, "Invalid option group ID", http.StatusBadRequest)
		return
	}

	var req models.ReorderOptionValuesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(req.ValueIDs) == 0 {
		respondWithError(w, "value_ids is required", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.ReorderOptionValues(r.Context(), groupID, req.ValueIDs); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	group, err := h.productRepo.UpdateOptionGroup(r.Context(), groupID, &models.UpdateOptionGroupRequest{})
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, group, http.StatusOK)
}

// Customization Fields (Story 5)

func (h *ProductHandler) GetCustomizationFields(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/products/", "/customizations")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if _, err := h.productRepo.FindByID(r.Context(), productID); err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	fields, err := h.productRepo.GetCustomizationFieldsByProduct(r.Context(), productID)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}
	if fields == nil {
		fields = []models.ProductCustomizationField{}
	}

	respondWithJSON(w, fields, http.StatusOK)
}

func (h *ProductHandler) CreateCustomizationField(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/admin/products/", "/customizations")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if _, err := h.productRepo.FindByID(r.Context(), productID); err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	var req models.CreateCustomizationFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respondWithError(w, "Field name is required", http.StatusBadRequest)
		return
	}

	validTypes := map[string]bool{"text": true, "textarea": true, "file": true}
	if !validTypes[req.FieldType] {
		respondWithError(w, "field_type must be text, textarea, or file", http.StatusBadRequest)
		return
	}

	fields, _ := h.productRepo.GetCustomizationFieldsByProduct(r.Context(), productID)
	if len(fields) >= 5 {
		respondWithError(w, "Maximum 5 customization fields per product", http.StatusBadRequest)
		return
	}

	field, err := h.productRepo.CreateCustomizationField(r.Context(), productID, &req)
	if err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, field, http.StatusCreated)
}

func (h *ProductHandler) UpdateCustomizationField(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fieldID, err := h.extractID(r.URL.Path, "/customizations/")
	if err != nil {
		respondWithError(w, "Invalid customization field ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateCustomizationFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		req.Name = &name
		if name == "" {
			respondWithError(w, "Field name cannot be empty", http.StatusBadRequest)
			return
		}
	}
	if req.FieldType != nil {
		validTypes := map[string]bool{"text": true, "textarea": true, "file": true}
		if !validTypes[*req.FieldType] {
			respondWithError(w, "field_type must be text, textarea, or file", http.StatusBadRequest)
			return
		}
	}

	field, err := h.productRepo.UpdateCustomizationField(r.Context(), fieldID, &req)
	if err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Customization field not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, field, http.StatusOK)
}

func (h *ProductHandler) DeleteCustomizationField(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fieldID, err := h.extractID(r.URL.Path, "/customizations/")
	if err != nil {
		respondWithError(w, "Invalid customization field ID", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.DeleteCustomizationField(r.Context(), fieldID); err != nil {
		if err == repositories.ErrNotFound {
			respondWithError(w, "Customization field not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{"message": "Customization field deleted"}, http.StatusOK)
}

func (h *ProductHandler) ReorderCustomizationFields(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID, err := h.extractProductID(r.URL.Path, "/api/admin/products/", "/customizations/reorder")
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req models.ReorderCustomizationFieldsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(req.FieldIDs) == 0 {
		respondWithError(w, "field_ids is required", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.ReorderCustomizationFields(r.Context(), productID, req.FieldIDs); err != nil {
		respondWithServerError(w, "Internal server error", err, http.StatusInternalServerError)
		return
	}

	fields, _ := h.productRepo.GetCustomizationFieldsByProduct(r.Context(), productID)
	respondWithJSON(w, fields, http.StatusOK)
}

// Helpers

func (h *ProductHandler) extractProductID(path, prefix, suffix string) (int, error) {
	rest := strings.TrimPrefix(path, prefix)
	rest = strings.TrimSuffix(rest, suffix)
	rest = strings.TrimSuffix(rest, "/")
	return strconv.Atoi(rest)
}

func (h *ProductHandler) extractID(path, marker string) (int, error) {
	idx := strings.LastIndex(path, marker)
	if idx == -1 {
		return 0, fmt.Errorf("not found")
	}
	idStr := path[idx+len(marker):]
	idStr = strings.TrimSuffix(idStr, "/")
	return strconv.Atoi(idStr)
}

// extractIDBetween returns the numeric path segment immediately following marker,
// for paths where further segments follow (e.g. "/products/5/variants/12" with
// marker "/products/" yields 5).
func (h *ProductHandler) extractIDBetween(path, marker string) (int, error) {
	idx := strings.Index(path, marker)
	if idx == -1 {
		return 0, fmt.Errorf("not found")
	}
	rest := path[idx+len(marker):]
	if slash := strings.Index(rest, "/"); slash != -1 {
		rest = rest[:slash]
	}
	return strconv.Atoi(rest)
}
