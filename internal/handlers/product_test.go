package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/repositories"
)

type mockProductRepo struct {
	products map[int]*models.Product
	slugMap  map[string]*models.Product
	nextID   int

	validateOptionValuesErr          error
	validateOptionValuesCalled       bool
	hasVariantCombinationResult      bool
	hasVariantCombinationErr         error
	hasVariantCombinationCalled      bool
	lastHasVariantCombinationExclude int
	hasVariantCombinationFunc        func(optionValueIDs []int) bool

	countVariantsResult int
	countVariantsErr    error

	optionGroupsResult []models.ProductOptionGroup
	optionGroupsErr    error

	createVariantCalls   []models.CreateVariantRequest
	createVariantErrFunc func(v *models.CreateVariantRequest) error
	createVariantNextID  int

	updateVariantCalls   []models.UpdateVariantRequest
	updateVariantErrFunc func(variantID int, v *models.UpdateVariantRequest) error

	variantProductIDs   map[int]int
	variantProductIDErr error
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{
		products: make(map[int]*models.Product),
		slugMap:  make(map[string]*models.Product),
		nextID:   1,
	}
}

func (m *mockProductRepo) FindBySlug(ctx context.Context, slug string) (*models.Product, error) {
	if p, ok := m.slugMap[slug]; ok {
		return p, nil
	}
	return nil, repositories.ErrNotFound
}

func (m *mockProductRepo) FindByID(ctx context.Context, id int) (*models.Product, error) {
	if p, ok := m.products[id]; ok {
		return p, nil
	}
	return nil, repositories.ErrNotFound
}

func (m *mockProductRepo) Create(ctx context.Context, p *models.CreateProductRequest, slug string) (*models.Product, error) {
	product := &models.Product{
		ID:          m.nextID,
		Name:        p.Name,
		Slug:        slug,
		Description: p.Description,
		Price:       p.Price,
		Images:      p.Images,
		Status:      "published",
		Categories:  []models.Category{},
	}
	m.products[m.nextID] = product
	m.slugMap[slug] = product
	m.nextID++
	return product, nil
}

func (m *mockProductRepo) Update(ctx context.Context, id int, p *models.UpdateProductRequest) (*models.Product, error) {
	product, ok := m.products[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	if p.Name != nil {
		product.Name = *p.Name
	}
	if p.Description != nil {
		product.Description = *p.Description
	}
	if p.Price != nil {
		product.Price = *p.Price
	}
	if p.Images != nil {
		product.Images = *p.Images
	}
	if p.Status != nil {
		product.Status = *p.Status
	}
	return product, nil
}

func (m *mockProductRepo) SoftDelete(ctx context.Context, id int) error {
	if _, ok := m.products[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(m.products, id)
	return nil
}

func (m *mockProductRepo) List(ctx context.Context, filters repositories.ProductFilters) ([]models.Product, int, error) {
	var allProducts []models.Product
	for _, p := range m.products {
		if p.Status == "published" {
			allProducts = append(allProducts, *p)
		}
	}

	total := len(allProducts)
	if filters.Offset >= total {
		return []models.Product{}, total, nil
	}

	end := filters.Offset + filters.Limit
	if end > total {
		end = total
	}

	return allProducts[filters.Offset:end], total, nil
}

func (m *mockProductRepo) AdminList(ctx context.Context, filters repositories.ProductFilters) ([]models.Product, int, error) {
	var allProducts []models.Product
	for _, p := range m.products {
		allProducts = append(allProducts, *p)
	}

	total := len(allProducts)
	if filters.Offset >= total {
		return []models.Product{}, total, nil
	}

	end := filters.Offset + filters.Limit
	if end > total {
		end = total
	}

	return allProducts[filters.Offset:end], total, nil
}

func (m *mockProductRepo) GetImages(ctx context.Context, productID int) ([]string, error) {
	if p, ok := m.products[productID]; ok {
		return p.Images, nil
	}
	return nil, repositories.ErrNotFound
}

func (m *mockProductRepo) RemoveImageFromProduct(ctx context.Context, productID int, imagePath string) error {
	p, ok := m.products[productID]
	if !ok {
		return repositories.ErrNotFound
	}
	for i, img := range p.Images {
		if img == imagePath {
			p.Images = append(p.Images[:i], p.Images[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockProductRepo) AddImageToProduct(ctx context.Context, productID int, imagePath string) error {
	p, ok := m.products[productID]
	if !ok {
		return repositories.ErrNotFound
	}
	p.Images = append(p.Images, imagePath)
	return nil
}

func (m *mockProductRepo) GetVariants(ctx context.Context, productID int) ([]models.ProductVariant, error) {
	return nil, nil
}

func (m *mockProductRepo) GetReviewStats(ctx context.Context, productID int) (float64, int, error) {
	return 0, 0, nil
}

func (m *mockProductRepo) LinkCategories(ctx context.Context, productID int, categoryIDs []int) error {
	return nil
}

func (m *mockProductRepo) GetCategories(ctx context.Context, productID int) ([]models.Category, error) {
	return []models.Category{}, nil
}

func (m *mockProductRepo) CreateVariant(ctx context.Context, productID int, v *models.CreateVariantRequest) (*models.ProductVariant, error) {
	m.createVariantCalls = append(m.createVariantCalls, *v)
	if m.createVariantErrFunc != nil {
		if err := m.createVariantErrFunc(v); err != nil {
			return nil, err
		}
	}
	m.createVariantNextID++
	return &models.ProductVariant{
		ID:        m.createVariantNextID,
		ProductID: productID,
		Stock:     v.Stock,
		SKU:       v.SKU,
	}, nil
}

func (m *mockProductRepo) CountVariants(ctx context.Context, productID int) (int, error) {
	return m.countVariantsResult, m.countVariantsErr
}

func (m *mockProductRepo) GetVariantProductID(ctx context.Context, variantID int) (int, error) {
	if m.variantProductIDErr != nil {
		return 0, m.variantProductIDErr
	}
	if productID, ok := m.variantProductIDs[variantID]; ok {
		return productID, nil
	}
	return 0, repositories.ErrNotFound
}

func (m *mockProductRepo) UpdateVariant(ctx context.Context, variantID int, v *models.UpdateVariantRequest) (*models.ProductVariant, error) {
	m.updateVariantCalls = append(m.updateVariantCalls, *v)
	if m.updateVariantErrFunc != nil {
		if err := m.updateVariantErrFunc(variantID, v); err != nil {
			return nil, err
		}
	}
	return &models.ProductVariant{
		ID:    variantID,
		Stock: safeDerefInt(v.Stock, 0),
	}, nil
}

func (m *mockProductRepo) DeleteVariant(ctx context.Context, variantID int) error {
	return nil
}

func (m *mockProductRepo) GetTags(ctx context.Context, productID int) ([]string, error) {
	return []string{}, nil
}

func (m *mockProductRepo) AddTags(ctx context.Context, productID int, tags []string) error {
	return nil
}

func (m *mockProductRepo) RemoveTag(ctx context.Context, productID int, tag string) error {
	return nil
}

func (m *mockProductRepo) GetImageOrder(ctx context.Context, productID int) ([]models.ProductImageOrder, error) {
	return []models.ProductImageOrder{}, nil
}

func (m *mockProductRepo) UpdateImageOrder(ctx context.Context, productID int, images []models.ImageOrderEntry) error {
	return nil
}

func (m *mockProductRepo) GetOptionGroupsByProduct(ctx context.Context, productID int) ([]models.ProductOptionGroup, error) {
	if m.optionGroupsErr != nil {
		return nil, m.optionGroupsErr
	}
	if m.optionGroupsResult != nil {
		return m.optionGroupsResult, nil
	}
	return []models.ProductOptionGroup{}, nil
}

func (m *mockProductRepo) CreateOptionGroup(ctx context.Context, productID int, req *models.CreateOptionGroupRequest) (*models.ProductOptionGroup, error) {
	return &models.ProductOptionGroup{
		ID:        1,
		ProductID: productID,
		Name:      req.Name,
		Values:    []models.ProductOptionValue{},
	}, nil
}

func (m *mockProductRepo) UpdateOptionGroup(ctx context.Context, groupID int, req *models.UpdateOptionGroupRequest) (*models.ProductOptionGroup, error) {
	return &models.ProductOptionGroup{
		ID:   groupID,
		Name: safeDeref(req.Name, "default"),
	}, nil
}

func (m *mockProductRepo) DeleteOptionGroup(ctx context.Context, groupID int) error {
	return nil
}

func (m *mockProductRepo) CreateOptionValue(ctx context.Context, groupID int, req *models.CreateOptionValueRequest) (*models.ProductOptionValue, error) {
	pm := 0
	if req.PriceModifier != nil {
		pm = *req.PriceModifier
	}
	return &models.ProductOptionValue{
		ID:            1,
		GroupID:       groupID,
		Value:         req.Value,
		PriceModifier: pm,
	}, nil
}

func (m *mockProductRepo) UpdateOptionValue(ctx context.Context, valueID int, req *models.UpdateOptionValueRequest) (*models.ProductOptionValue, error) {
	return &models.ProductOptionValue{
		ID:            valueID,
		Value:         safeDeref(req.Value, ""),
		PriceModifier: safeDerefInt(req.PriceModifier, 0),
	}, nil
}

func (m *mockProductRepo) DeleteOptionValue(ctx context.Context, valueID int) error {
	return nil
}

func (m *mockProductRepo) ReorderOptionValues(ctx context.Context, groupID int, valueIDs []int) error {
	return nil
}

func (m *mockProductRepo) ValidateOptionValues(ctx context.Context, productID int, optionValueIDs []int) error {
	m.validateOptionValuesCalled = true
	return m.validateOptionValuesErr
}

func (m *mockProductRepo) HasVariantCombination(ctx context.Context, productID int, optionValueIDs []int, excludeVariantID int) (bool, error) {
	m.hasVariantCombinationCalled = true
	m.lastHasVariantCombinationExclude = excludeVariantID
	if m.hasVariantCombinationFunc != nil {
		return m.hasVariantCombinationFunc(optionValueIDs), m.hasVariantCombinationErr
	}
	return m.hasVariantCombinationResult, m.hasVariantCombinationErr
}

func (m *mockProductRepo) GetCustomizationFieldsByProduct(ctx context.Context, productID int) ([]models.ProductCustomizationField, error) {
	return []models.ProductCustomizationField{}, nil
}

func (m *mockProductRepo) CreateCustomizationField(ctx context.Context, productID int, req *models.CreateCustomizationFieldRequest) (*models.ProductCustomizationField, error) {
	return &models.ProductCustomizationField{
		ID:        1,
		ProductID: productID,
		Name:      req.Name,
		FieldType: req.FieldType,
	}, nil
}

func (m *mockProductRepo) UpdateCustomizationField(ctx context.Context, fieldID int, req *models.UpdateCustomizationFieldRequest) (*models.ProductCustomizationField, error) {
	return &models.ProductCustomizationField{
		ID:   fieldID,
		Name: safeDeref(req.Name, ""),
	}, nil
}

func (m *mockProductRepo) DeleteCustomizationField(ctx context.Context, fieldID int) error {
	return nil
}

func (m *mockProductRepo) ReorderCustomizationFields(ctx context.Context, productID int, fieldIDs []int) error {
	return nil
}

func safeDeref(s *string, def string) string {
	if s != nil {
		return *s
	}
	return def
}

func safeDerefInt(i *int, def int) int {
	if i != nil {
		return *i
	}
	return def
}

type mockReviewRepo struct{}

func (m *mockReviewRepo) Create(ctx context.Context, productID, userID, rating int, comment string) (*models.Review, error) {
	return nil, nil
}

func (m *mockReviewRepo) ListByProduct(ctx context.Context, productID, limit, offset int) ([]models.Review, error) {
	return nil, nil
}

func (m *mockReviewRepo) GetStats(ctx context.Context, productID int) (float64, int, error) {
	return 0, 0, nil
}

func (m *mockReviewRepo) ExistsProduct(ctx context.Context, productID int) (bool, error) {
	return false, nil
}

type mockCategoryRepo struct{}

func (m *mockCategoryRepo) List(ctx context.Context) ([]models.Category, error) {
	return []models.Category{}, nil
}

func (m *mockCategoryRepo) FindByID(ctx context.Context, id int) (*models.Category, error) {
	return nil, repositories.ErrNotFound
}

func (m *mockCategoryRepo) Create(ctx context.Context, name string) (*models.Category, error) {
	return nil, nil
}

func (m *mockCategoryRepo) Update(ctx context.Context, id int, name string) error {
	return nil
}

func (m *mockCategoryRepo) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *mockCategoryRepo) CountProducts(ctx context.Context, categoryID int) (int, error) {
	return 0, nil
}

func (m *mockCategoryRepo) RemoveCategoryFromProducts(ctx context.Context, categoryID int) error {
	return nil
}

func TestGenerateSlug(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "Shirt", "shirt"},
		{"name with spaces", "Blue Shirt", "blue-shirt"},
		{"special characters", "Shirt & Pants!", "shirt-pants"},
		{"multiple spaces", "  Blue   Shirt  ", "blue-shirt"},
		{"unicode characters", "Café Résumé", "cafe-resume"},
		{"numbers", "Shirt 2.0", "shirt-20"},
		{"only special chars", "!@#$%", "product"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug, err := handler.generateSlug(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if slug != tt.expected {
				t.Errorf("generateSlug(%q) = %q, want %q", tt.input, slug, tt.expected)
			}
		})
	}
}

func TestGenerateSlug_Collision(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	ctx := context.Background()

	slug1, _ := handler.generateSlug(ctx, "Shirt")
	if slug1 != "shirt" {
		t.Fatalf("expected 'shirt', got %q", slug1)
	}

	repo.slugMap["shirt"] = &models.Product{ID: 1, Slug: "shirt"}

	slug2, _ := handler.generateSlug(ctx, "Shirt")
	if slug2 != "shirt-2" {
		t.Errorf("expected 'shirt-2', got %q", slug2)
	}

	repo.slugMap["shirt-2"] = &models.Product{ID: 2, Slug: "shirt-2"}

	slug3, _ := handler.generateSlug(ctx, "Shirt")
	if slug3 != "shirt-3" {
		t.Errorf("expected 'shirt-3', got %q", slug3)
	}
}

func TestCreateProduct_Handler(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := models.CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product",
		Price:       1000,
		CategoryIDs: []int{},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/admin/products", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateProduct(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var product models.Product
	json.NewDecoder(w.Body).Decode(&product)
	if product.Name != "Test Product" {
		t.Errorf("expected name 'Test Product', got %q", product.Name)
	}
	if product.Slug != "test-product" {
		t.Errorf("expected slug 'test-product', got %q", product.Slug)
	}
}

func TestCreateProduct_Validation(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	tests := []struct {
		name       string
		body       models.CreateProductRequest
		wantStatus int
	}{
		{
			name:       "empty name",
			body:       models.CreateProductRequest{Name: "", Price: 1000},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "zero price",
			body:       models.CreateProductRequest{Name: "Test", Price: 0},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "negative price",
			body:       models.CreateProductRequest{Name: "Test", Price: -1},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "price over max",
			body:       models.CreateProductRequest{Name: "Test", Price: 100000000},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/admin/products", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateProduct(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestUpdateProduct_Handler(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	createBody := models.CreateProductRequest{
		Name:  "Original",
		Price: 1000,
	}
	jsonBody, _ := json.Marshal(createBody)
	req := httptest.NewRequest("POST", "/api/admin/products", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.CreateProduct(w, req)

	var created models.Product
	json.NewDecoder(w.Body).Decode(&created)

	newName := "Updated Product"
	updateBody := models.UpdateProductRequest{
		Name: &newName,
	}
	jsonBody, _ = json.Marshal(updateBody)
	req = httptest.NewRequest("PUT", "/api/admin/products/1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler.UpdateProduct(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var updated models.Product
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Name != "Updated Product" {
		t.Errorf("expected name 'Updated Product', got %q", updated.Name)
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	newName := "Updated"
	updateBody := models.UpdateProductRequest{
		Name: &newName,
	}
	jsonBody, _ := json.Marshal(updateBody)
	req := httptest.NewRequest("PUT", "/api/admin/products/999", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateProduct(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestListProducts_Handler(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	for i := 0; i < 5; i++ {
		body := models.CreateProductRequest{
			Name:  "Product " + strings.Repeat(string(rune('A'+i)), 3),
			Price: 1000 * (i + 1),
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/admin/products", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.CreateProduct(w, req)
	}

	req := httptest.NewRequest("GET", "/api/products?page=1&limit=2", nil)
	w := httptest.NewRecorder()
	handler.ListProducts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response models.ProductListResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(response.Products))
	}
	if response.Pagination.Total != 5 {
		t.Errorf("expected total 5, got %d", response.Pagination.Total)
	}
	if response.Pagination.TotalPages != 3 {
		t.Errorf("expected total pages 3, got %d", response.Pagination.TotalPages)
	}
}

func TestCreateVariant_ValidationError(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.validateOptionValuesErr = errors.New("must select exactly one value for each of this product's 2 option group(s)")
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := models.CreateVariantRequest{SKU: "SKU-1", Stock: 5}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/products/1/variants", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	handler.CreateVariant(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if !repo.validateOptionValuesCalled {
		t.Error("expected ValidateOptionValues to be called")
	}
	if repo.hasVariantCombinationCalled {
		t.Error("expected HasVariantCombination not to be called after validation failure")
	}
}

func TestCreateVariant_DuplicateCombination(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.hasVariantCombinationResult = true
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := models.CreateVariantRequest{
		SKU:          "SKU-1",
		Stock:        5,
		OptionValues: []models.CreateVariantOptionValueRequest{{OptionValueID: 10}},
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/products/1/variants", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	handler.CreateVariant(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
	if repo.lastHasVariantCombinationExclude != 0 {
		t.Errorf("expected exclude 0 on create, got %d", repo.lastHasVariantCombinationExclude)
	}
}

func TestCreateVariant_Success(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := models.CreateVariantRequest{SKU: "SKU-1", Stock: 5}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/products/1/variants", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	handler.CreateVariant(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
	if !repo.validateOptionValuesCalled {
		t.Error("expected ValidateOptionValues to always be called, even with no option values")
	}
	if !repo.hasVariantCombinationCalled {
		t.Error("expected HasVariantCombination to always be called, even with no option values")
	}
}

func TestUpdateVariant_SkipsValidationWhenOptionValuesOmitted(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := map[string]interface{}{"stock": 7}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/products/1/variants/12", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateVariant(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if repo.validateOptionValuesCalled {
		t.Error("expected ValidateOptionValues not to be called when option_values is omitted")
	}
	if repo.hasVariantCombinationCalled {
		t.Error("expected HasVariantCombination not to be called when option_values is omitted")
	}
}

func TestUpdateVariant_ValidatesAndExcludesSelfWhenOptionValuesProvided(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := models.UpdateVariantRequest{
		OptionValues: []models.UpdateVariantOptionValueRequest{{OptionValueID: 10}},
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/products/1/variants/12", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateVariant(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if !repo.validateOptionValuesCalled {
		t.Error("expected ValidateOptionValues to be called when option_values is provided")
	}
	if !repo.hasVariantCombinationCalled {
		t.Error("expected HasVariantCombination to be called when option_values is provided")
	}
	if repo.lastHasVariantCombinationExclude != 12 {
		t.Errorf("expected duplicate check to exclude the variant being updated (12), got %d", repo.lastHasVariantCombinationExclude)
	}
}

func TestUpdateVariant_DuplicateCombination(t *testing.T) {
	repo := newMockProductRepo()
	repo.hasVariantCombinationResult = true
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := models.UpdateVariantRequest{
		OptionValues: []models.UpdateVariantOptionValueRequest{{OptionValueID: 10}},
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/products/1/variants/12", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateVariant(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
}

func TestCreateVariant_RejectsAtVariantCap(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.countVariantsResult = maxVariantsPerProduct
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := models.CreateVariantRequest{SKU: "SKU-1", Stock: 5}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/products/1/variants", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	handler.CreateVariant(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if repo.validateOptionValuesCalled {
		t.Error("expected ValidateOptionValues not to be called once the cap is reached")
	}
}

func optionGroup(id int, name string, valueIDs []int, values []string) models.ProductOptionGroup {
	vals := make([]models.ProductOptionValue, len(valueIDs))
	for i, vid := range valueIDs {
		vals[i] = models.ProductOptionValue{ID: vid, GroupID: id, Value: values[i]}
	}
	return models.ProductOptionGroup{ID: id, Name: name, Values: vals}
}

func TestGenerateVariants_ProductNotFound(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestGenerateVariants_NoOptionGroups(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1, Slug: "widget"}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGenerateVariants_ExceedsCap(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1, Slug: "widget"}
	valueIDs := make([]int, 200)
	values := make([]string, 200)
	for i := range valueIDs {
		valueIDs[i] = i + 1
		values[i] = fmt.Sprintf("v%d", i+1)
	}
	repo.optionGroupsResult = []models.ProductOptionGroup{
		optionGroup(1, "A", valueIDs, values),
		optionGroup(2, "B", valueIDs, values),
	}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if len(repo.createVariantCalls) != 0 {
		t.Errorf("expected no variants created when the cap would be exceeded, got %d", len(repo.createVariantCalls))
	}
}

func TestGenerateVariants_CreatesMissingSkipsExisting(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1, Slug: "widget"}
	repo.optionGroupsResult = []models.ProductOptionGroup{
		optionGroup(1, "Size", []int{10, 11}, []string{"S", "M"}),
		optionGroup(2, "Color", []int{20, 21}, []string{"Red", "Blue"}),
	}
	// Combination S+Red (10,20) already exists; the other 3 do not.
	repo.hasVariantCombinationFunc = func(ids []int) bool {
		has10, has20 := false, false
		for _, id := range ids {
			if id == 10 {
				has10 = true
			}
			if id == 20 {
				has20 = true
			}
		}
		return has10 && has20
	}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp models.GenerateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.CreatedCount != 3 {
		t.Errorf("expected 3 created, got %d", resp.CreatedCount)
	}
	if resp.SkippedCount != 1 {
		t.Errorf("expected 1 skipped, got %d", resp.SkippedCount)
	}
	if len(repo.createVariantCalls) != 3 {
		t.Fatalf("expected 3 CreateVariant calls, got %d", len(repo.createVariantCalls))
	}
	for _, call := range repo.createVariantCalls {
		if call.Stock != 0 {
			t.Errorf("expected generated variant stock 0, got %d", call.Stock)
		}
		if !strings.HasPrefix(call.SKU, "widget-") {
			t.Errorf("expected SKU to start with product slug, got %q", call.SKU)
		}
		if len(call.OptionValues) != 2 {
			t.Errorf("expected 2 option values per generated variant, got %d", len(call.OptionValues))
		}
	}
}

func TestGenerateVariants_RetriesSKUOnCollision(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1, Slug: "widget"}
	repo.optionGroupsResult = []models.ProductOptionGroup{
		optionGroup(1, "Color", []int{20}, []string{"Red"}),
	}
	firstAttempt := true
	repo.createVariantErrFunc = func(v *models.CreateVariantRequest) error {
		if firstAttempt {
			firstAttempt = false
			return errors.New(`duplicate key value violates unique constraint "product_variants_sku_key"`)
		}
		return nil
	}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if len(repo.createVariantCalls) != 2 {
		t.Fatalf("expected 2 CreateVariant attempts (collision + retry), got %d", len(repo.createVariantCalls))
	}
	if repo.createVariantCalls[1].SKU != "widget-red-2" {
		t.Errorf("expected retried SKU to have a numeric suffix, got %q", repo.createVariantCalls[1].SKU)
	}
}

func TestGenerateVariants_ExistingPlusMissingExceedsCap(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1, Slug: "widget"}
	repo.optionGroupsResult = []models.ProductOptionGroup{
		optionGroup(1, "Size", []int{10, 11}, []string{"S", "M"}),
	}
	repo.countVariantsResult = maxVariantsPerProduct - 1
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if len(repo.createVariantCalls) != 0 {
		t.Errorf("expected no variants created when existing+missing would exceed the cap, got %d", len(repo.createVariantCalls))
	}
}

func TestGenerateVariants_OptionGroupsLookupError(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1, Slug: "widget"}
	repo.optionGroupsErr = errors.New("db unavailable")
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGenerateVariants_CountVariantsError(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1, Slug: "widget"}
	repo.optionGroupsResult = []models.ProductOptionGroup{
		optionGroup(1, "Size", []int{10, 11}, []string{"S", "M"}),
	}
	repo.countVariantsErr = errors.New("db unavailable")
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGenerateVariants_HasVariantCombinationError(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1, Slug: "widget"}
	repo.optionGroupsResult = []models.ProductOptionGroup{
		optionGroup(1, "Size", []int{10, 11}, []string{"S", "M"}),
	}
	repo.hasVariantCombinationErr = errors.New("db unavailable")
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
	if len(repo.createVariantCalls) != 0 {
		t.Errorf("expected no variants created when the existence check errors, got %d", len(repo.createVariantCalls))
	}
}

func TestGenerateVariants_NonCollisionCreateErrorReportsPartialProgress(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1, Slug: "widget"}
	repo.optionGroupsResult = []models.ProductOptionGroup{
		optionGroup(1, "Size", []int{10, 11, 12}, []string{"S", "M", "L"}),
	}
	calls := 0
	repo.createVariantErrFunc = func(v *models.CreateVariantRequest) error {
		calls++
		if calls == 2 {
			return errors.New("connection reset by peer")
		}
		return nil
	}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	var resp models.GenerateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.CreatedCount != 1 {
		t.Errorf("expected partial progress of 1 created variant to be reported, got %d", resp.CreatedCount)
	}
	if resp.Error == "" {
		t.Error("expected an error message describing the partial failure")
	}
	// A non-unique-constraint error must not be retried with a suffixed SKU.
	if len(repo.createVariantCalls) != 2 {
		t.Errorf("expected exactly 2 CreateVariant attempts (1 success, 1 non-retried failure), got %d", len(repo.createVariantCalls))
	}
}

func TestCreateVariant_CountVariantsError(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.countVariantsErr = errors.New("db unavailable")
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := models.CreateVariantRequest{SKU: "SKU-1", Stock: 5}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/products/1/variants", bytes.NewReader(jsonBody))
	w := httptest.NewRecorder()

	handler.CreateVariant(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGenerateVariants_TruncatesLongSKU(t *testing.T) {
	repo := newMockProductRepo()
	longSlug := strings.Repeat("a", 120)
	repo.products[1] = &models.Product{ID: 1, Slug: longSlug}
	repo.optionGroupsResult = []models.ProductOptionGroup{
		optionGroup(1, "Color", []int{20}, []string{"Red"}),
	}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateVariants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if len(repo.createVariantCalls) != 1 {
		t.Fatalf("expected 1 CreateVariant call, got %d", len(repo.createVariantCalls))
	}
	if len(repo.createVariantCalls[0].SKU) > 100 {
		t.Errorf("expected generated SKU to stay within the 100-char column limit, got %d chars", len(repo.createVariantCalls[0].SKU))
	}
}

func bulkUpdateRequest(t *testing.T, body models.BulkUpdateVariantsRequest) *http.Request {
	t.Helper()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return httptest.NewRequest("PATCH", "/api/admin/products/1/variants/bulk", bytes.NewReader(jsonBody))
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func TestBulkUpdateVariants_ProductNotFound(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{{VariantID: 1, Stock: intPtr(5)}},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestBulkUpdateVariants_EmptyItems(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{Items: []models.BulkUpdateVariantItem{}})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestBulkUpdateVariants_RejectsOptionValues(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	body := `{"items":[{"variant_id":1,"stock":5},{"variant_id":2,"option_values":[{"option_value_id":10}]}]}`
	req := httptest.NewRequest("PATCH", "/api/admin/products/1/variants/bulk", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if len(repo.updateVariantCalls) != 0 {
		t.Errorf("expected no updates to be applied once any item includes option_values, got %d", len(repo.updateVariantCalls))
	}
}

func TestBulkUpdateVariants_Success(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.variantProductIDs = map[int]int{1: 1, 2: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{
			{VariantID: 1, Stock: intPtr(10)},
			{VariantID: 2, SKU: strPtr("NEW-SKU-2")},
		},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp models.BulkUpdateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.SucceededCount != 2 || resp.FailedCount != 0 {
		t.Errorf("expected 2 succeeded, 0 failed, got %d/%d", resp.SucceededCount, resp.FailedCount)
	}
	if len(repo.updateVariantCalls) != 2 {
		t.Fatalf("expected 2 UpdateVariant calls, got %d", len(repo.updateVariantCalls))
	}
	if *repo.updateVariantCalls[0].Stock != 10 {
		t.Errorf("expected stock 10 passed through, got %d", *repo.updateVariantCalls[0].Stock)
	}
	if *repo.updateVariantCalls[1].SKU != "NEW-SKU-2" {
		t.Errorf("expected SKU passed through, got %q", *repo.updateVariantCalls[1].SKU)
	}
}

func TestBulkUpdateVariants_PartialFailureVariantNotFound(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.variantProductIDs = map[int]int{1: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{
			{VariantID: 1, Stock: intPtr(10)},
			{VariantID: 99, Stock: intPtr(5)},
		},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp models.BulkUpdateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.SucceededCount != 1 || resp.FailedCount != 1 {
		t.Errorf("expected 1 succeeded, 1 failed, got %d/%d", resp.SucceededCount, resp.FailedCount)
	}
	if len(repo.updateVariantCalls) != 1 {
		t.Errorf("expected only the valid item to reach UpdateVariant, got %d calls", len(repo.updateVariantCalls))
	}
	if resp.Results[1].Success || resp.Results[1].Error == "" {
		t.Errorf("expected item 2 to fail with an error message, got %+v", resp.Results[1])
	}
}

func TestBulkUpdateVariants_RejectsCrossProductVariant(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.variantProductIDs = map[int]int{5: 2} // belongs to product 2, not 1
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{{VariantID: 5, Stock: intPtr(10)}},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	var resp models.BulkUpdateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.SucceededCount != 0 || resp.FailedCount != 1 {
		t.Errorf("expected 0 succeeded, 1 failed, got %d/%d", resp.SucceededCount, resp.FailedCount)
	}
	if len(repo.updateVariantCalls) != 0 {
		t.Errorf("expected no update applied for a variant belonging to another product, got %d calls", len(repo.updateVariantCalls))
	}
}

func TestBulkUpdateVariants_EmptySKURejected(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.variantProductIDs = map[int]int{1: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{{VariantID: 1, SKU: strPtr("   ")}},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	var resp models.BulkUpdateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.FailedCount != 1 {
		t.Errorf("expected 1 failed, got %d", resp.FailedCount)
	}
	if len(repo.updateVariantCalls) != 0 {
		t.Errorf("expected no UpdateVariant call for a blank SKU, got %d", len(repo.updateVariantCalls))
	}
}

func TestBulkUpdateVariants_SKUConflict(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.variantProductIDs = map[int]int{1: 1}
	repo.updateVariantErrFunc = func(variantID int, v *models.UpdateVariantRequest) error {
		return errors.New(`duplicate key value violates unique constraint "product_variants_sku_key"`)
	}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{{VariantID: 1, SKU: strPtr("TAKEN")}},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	var resp models.BulkUpdateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.FailedCount != 1 || resp.Results[0].Error != "SKU already exists" {
		t.Errorf("expected a SKU-conflict failure, got %+v", resp.Results)
	}
}

func TestBulkUpdateVariants_MissingVariantID(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{{Stock: intPtr(5)}},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	var resp models.BulkUpdateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.FailedCount != 1 || resp.Results[0].Error != "variant_id is required" {
		t.Errorf("expected a variant_id-required failure, got %+v", resp.Results)
	}
	if len(repo.updateVariantCalls) != 0 {
		t.Errorf("expected no repository call for a missing variant_id, got %d", len(repo.updateVariantCalls))
	}
}

func TestBulkUpdateVariants_RejectsMethodNotAllowed(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("POST", "/api/admin/products/1/variants/bulk", nil)
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestBulkUpdateVariants_InvalidProductID(t *testing.T) {
	repo := newMockProductRepo()
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("PATCH", "/api/admin/products/abc/variants/bulk", nil)
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestBulkUpdateVariants_InvalidJSON(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := httptest.NewRequest("PATCH", "/api/admin/products/1/variants/bulk", strings.NewReader("{not json"))
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestBulkUpdateVariants_RejectsOversizedBatch(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	items := make([]models.BulkUpdateVariantItem, maxVariantsPerProduct+1)
	for i := range items {
		items[i] = models.BulkUpdateVariantItem{VariantID: i + 1, Stock: intPtr(1)}
	}
	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{Items: items})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if len(repo.updateVariantCalls) != 0 {
		t.Errorf("expected no updates to be applied for an oversized batch, got %d", len(repo.updateVariantCalls))
	}
}

func TestBulkUpdateVariants_VariantLookupInternalError(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.variantProductIDErr = errors.New("db unavailable")
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{{VariantID: 1, Stock: intPtr(5)}},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	var resp models.BulkUpdateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.FailedCount != 1 || resp.Results[0].Error != "internal server error" {
		t.Errorf("expected an internal-server-error failure, got %+v", resp.Results)
	}
	if len(repo.updateVariantCalls) != 0 {
		t.Errorf("expected no UpdateVariant call when ownership lookup errors, got %d", len(repo.updateVariantCalls))
	}
}

func TestBulkUpdateVariants_ForwardsDimensions(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.variantProductIDs = map[int]int{1: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{
			{VariantID: 1, Weight: intPtr(100), Length: intPtr(10), Width: intPtr(20), Height: intPtr(30)},
		},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	if len(repo.updateVariantCalls) != 1 {
		t.Fatalf("expected 1 UpdateVariant call, got %d", len(repo.updateVariantCalls))
	}
	call := repo.updateVariantCalls[0]
	if call.Weight == nil || *call.Weight != 100 ||
		call.Length == nil || *call.Length != 10 ||
		call.Width == nil || *call.Width != 20 ||
		call.Height == nil || *call.Height != 30 {
		t.Errorf("expected weight/length/width/height to be forwarded, got %+v", call)
	}
}

func TestBulkUpdateVariants_DuplicateVariantIDProcessedIndependently(t *testing.T) {
	repo := newMockProductRepo()
	repo.products[1] = &models.Product{ID: 1}
	repo.variantProductIDs = map[int]int{1: 1}
	handler := NewProductHandler(repo, &mockReviewRepo{}, &mockCategoryRepo{})

	req := bulkUpdateRequest(t, models.BulkUpdateVariantsRequest{
		Items: []models.BulkUpdateVariantItem{
			{VariantID: 1, Stock: intPtr(5)},
			{VariantID: 1, Stock: intPtr(9)},
		},
	})
	w := httptest.NewRecorder()

	handler.BulkUpdateVariants(w, req)

	var resp models.BulkUpdateVariantsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.SucceededCount != 2 {
		t.Errorf("expected both entries for the duplicate variant_id to be processed and succeed, got %d", resp.SucceededCount)
	}
	if len(repo.updateVariantCalls) != 2 {
		t.Fatalf("expected 2 independent UpdateVariant calls, got %d", len(repo.updateVariantCalls))
	}
	if *repo.updateVariantCalls[0].Stock != 5 || *repo.updateVariantCalls[1].Stock != 9 {
		t.Errorf("expected both stock values applied in order, got %d then %d", *repo.updateVariantCalls[0].Stock, *repo.updateVariantCalls[1].Stock)
	}
}
