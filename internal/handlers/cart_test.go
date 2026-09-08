package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"

	"ecommerce-api-go/internal/middleware"
	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/repositories"
)

type mockCartRepo struct {
	items     []repositories.CartItemRow
	nextID    int
	upsertErr error
	deleteErr error
}

func newMockCartRepo() *mockCartRepo {
	return &mockCartRepo{nextID: 1}
}

func (m *mockCartRepo) FindByUserID(ctx context.Context, userID int) ([]repositories.CartItemRow, error) {
	return m.items, nil
}

func (m *mockCartRepo) FindByUserIDForUpdate(ctx context.Context, tx pgx.Tx, userID int) ([]repositories.CartItemRow, error) {
	return m.items, nil
}

func (m *mockCartRepo) FindBySessionID(ctx context.Context, sessionID string) ([]repositories.CartItemRow, error) {
	return m.items, nil
}

func (m *mockCartRepo) Upsert(ctx context.Context, userID *int, sessionID *string, productID int, variantID *int, quantity int, customizationData []byte) error {
	return m.upsertErr
}

func (m *mockCartRepo) UpdateQuantity(ctx context.Context, itemID int, userID *int, sessionID *string, quantity int) error {
	return nil
}

func (m *mockCartRepo) Delete(ctx context.Context, itemID int, userID *int, sessionID *string) error {
	return m.deleteErr
}

func (m *mockCartRepo) GetGuestItems(ctx context.Context, sessionID string) ([]repositories.GuestCartItem, error) {
	return nil, nil
}

func (m *mockCartRepo) MergeGuestCart(ctx context.Context, userID int, sessionID string) (int, error) {
	return 2, nil
}

func (m *mockCartRepo) ClearByUserID(ctx context.Context, userID int) error {
	return nil
}

func (m *mockCartRepo) InsertItems(ctx context.Context, userID int, items []repositories.CartItemRow) error {
	return nil
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return middleware.AuthMiddleware(next)
}

func optionalAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return middleware.OptionalAuthMiddleware(next)
}

func TestGetCart_Empty(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	req := httptest.NewRequest("GET", "/api/cart", nil)
	req.Header.Set("X-Session-ID", "test-session")
	w := httptest.NewRecorder()

	handler.GetCart(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var cart models.Cart
	json.NewDecoder(w.Body).Decode(&cart)
	if cart.ItemCount != 0 {
		t.Errorf("expected 0 items, got %d", cart.ItemCount)
	}
}

func TestGetCart_WithItems(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	cartRepo.items = []repositories.CartItemRow{
		{
			CartItemID:   1,
			Quantity:     2,
			ProductID:    1,
			ProductPrice: 1000,
			ProductName:  "Test Product",
			ProductSlug:  "test-product",
			ProductImages: []string{"/img/test.jpg"},
		},
	}

	req := httptest.NewRequest("GET", "/api/cart", nil)
	req.Header.Set("X-Session-ID", "test-session")
	w := httptest.NewRecorder()

	handler.GetCart(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var cart models.Cart
	json.NewDecoder(w.Body).Decode(&cart)
	if cart.ItemCount != 1 {
		t.Errorf("expected 1 item, got %d", cart.ItemCount)
	}
	if cart.Total != 2000 {
		t.Errorf("expected total 2000, got %d", cart.Total)
	}
}

func TestGetCart_WithVariant(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	stock := 10
	label := "Large / Red"
	cartRepo.items = []repositories.CartItemRow{
		{
			CartItemID:    1,
			Quantity:      1,
			ProductID:     1,
			ProductPrice:  2000,
			ProductName:   "Variant Product",
			ProductSlug:   "variant-product",
			ProductImages: []string{},
			VariantID:     &stock,
			VariantStock:  &stock,
			VariantLabel:  &label,
		},
	}

	req := httptest.NewRequest("GET", "/api/cart", nil)
	req.Header.Set("X-Session-ID", "test-session")
	w := httptest.NewRecorder()

	handler.GetCart(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var cart models.Cart
	json.NewDecoder(w.Body).Decode(&cart)
	if cart.Items[0].Variant == nil {
		t.Error("expected variant to be present")
	}
}

func TestGetCart_NoIdentifier(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	req := httptest.NewRequest("GET", "/api/cart", nil)
	w := httptest.NewRecorder()

	handler.GetCart(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var cart models.Cart
	json.NewDecoder(w.Body).Decode(&cart)
	if cart.ItemCount != 0 {
		t.Errorf("expected 0 items, got %d", cart.ItemCount)
	}
}

func TestAddItem_Success(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	// Pre-create a product
	body := models.CreateProductRequest{Name: "Test Product", Price: 1000}
	jsonBody, _ := json.Marshal(body)
	reqCreate := httptest.NewRequest("POST", "/api/admin/products", bytes.NewReader(jsonBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	handlerProduct := NewProductHandler(productRepo, &mockReviewRepo{}, &mockCategoryRepo{})
	handlerProduct.CreateProduct(wCreate, reqCreate)

	addBody := models.AddToCartRequest{
		ProductID: 1,
		Quantity:  1,
	}
	jsonAddBody, _ := json.Marshal(addBody)
	req := httptest.NewRequest("POST", "/api/cart/items", bytes.NewReader(jsonAddBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", "test-session")
	w := httptest.NewRecorder()

	handler.AddItem(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAddItem_ProductNotFound(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	body := models.AddToCartRequest{
		ProductID: 999,
		Quantity:  1,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/cart/items", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", "test-session")
	w := httptest.NewRecorder()

	handler.AddItem(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestAddItem_NoIdentifier(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	body := models.AddToCartRequest{
		ProductID: 1,
		Quantity:  1,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/cart/items", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.AddItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAddItem_DefaultQuantity(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	// Pre-create a product
	body := models.CreateProductRequest{Name: "Default Qty Product", Price: 500}
	jsonBody, _ := json.Marshal(body)
	reqCreate := httptest.NewRequest("POST", "/api/admin/products", bytes.NewReader(jsonBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	handlerProduct := NewProductHandler(productRepo, &mockReviewRepo{}, &mockCategoryRepo{})
	handlerProduct.CreateProduct(wCreate, reqCreate)

	addBody := models.AddToCartRequest{
		ProductID: 1,
		Quantity:  0,
	}
	jsonAddBody, _ := json.Marshal(addBody)
	req := httptest.NewRequest("POST", "/api/cart/items", bytes.NewReader(jsonAddBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", "test-session")
	w := httptest.NewRecorder()

	handler.AddItem(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestUpdateItem_Success(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	body := models.UpdateCartItemRequest{Quantity: 5}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/cart/items/1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", "test-session")
	w := httptest.NewRecorder()

	handler.UpdateItem(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateItem_InvalidID(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	body := models.UpdateCartItemRequest{Quantity: 5}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/cart/items/abc", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateItem_InvalidQuantity(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	body := models.UpdateCartItemRequest{Quantity: 0}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/cart/items/1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRemoveItem_Success(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	req := httptest.NewRequest("DELETE", "/api/cart/items/1", nil)
	req.Header.Set("X-Session-ID", "test-session")
	w := httptest.NewRecorder()

	handler.RemoveItem(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRemoveItem_InvalidID(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	req := httptest.NewRequest("DELETE", "/api/cart/items/abc", nil)
	w := httptest.NewRecorder()

	handler.RemoveItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestMergeGuestCart_Success(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	body := map[string]string{"session_id": "guest-session-123"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/cart/merge", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	authMiddleware(handler.MergeGuestCart)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMergeGuestCart_EmptySession(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	body := map[string]string{"session_id": ""}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/cart/merge", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	authMiddleware(handler.MergeGuestCart)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMergeGuestCart_Unauthorized(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	body := map[string]string{"session_id": "guest-session"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/cart/merge", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.MergeGuestCart(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAddItem_InvalidJSON(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	req := httptest.NewRequest("POST", "/api/cart/items", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", "test-session")
	w := httptest.NewRecorder()

	handler.AddItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetCart_MethodNotAllowed(t *testing.T) {
	cartRepo := newMockCartRepo()
	productRepo := newMockProductRepo()
	handler := NewCartHandler(cartRepo, productRepo)

	req := httptest.NewRequest("POST", "/api/cart", nil)
	w := httptest.NewRecorder()

	handler.GetCart(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
