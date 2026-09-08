package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce-api-go/internal/middleware"
	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/repositories"
)

type mockOrderRepo struct {
	orders      []repositories.OrderSummary
	adminOrders []repositories.AdminOrder
	order       *models.Order
	items       []models.OrderItem
	stats       *repositories.DashboardStats
	nextID      int
	statusMap   map[int]string
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		statusMap: make(map[int]string),
		stats: &repositories.DashboardStats{
			TotalOrders:     10,
			TotalRevenue:    50000,
			PendingOrders:   3,
			DeliveredOrders: 7,
		},
	}
}

func (m *mockOrderRepo) Checkout(ctx context.Context, userID int, params repositories.CheckoutParams) (int, error) {
	m.nextID++
	m.statusMap[m.nextID] = "pending"
	return m.nextID, nil
}

func (m *mockOrderRepo) FindByID(ctx context.Context, orderID, userID int) (*models.Order, error) {
	if m.order != nil {
		return m.order, nil
	}
	return &models.Order{
		ID:     orderID,
		UserID: userID,
		Status: "pending",
		Total:  2000,
	}, nil
}

func (m *mockOrderRepo) FindByPaymentIntent(ctx context.Context, piID string) (*models.Order, error) {
	return nil, repositories.ErrNotFound
}

func (m *mockOrderRepo) ListByUserID(ctx context.Context, userID, limit, offset int) ([]repositories.OrderSummary, int, error) {
	return m.orders, len(m.orders), nil
}

func (m *mockOrderRepo) ListAll(ctx context.Context, status string, limit, offset int) ([]repositories.AdminOrder, int, error) {
	return m.adminOrders, len(m.adminOrders), nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, orderID int, status string) error {
	m.statusMap[orderID] = status
	return nil
}

func (m *mockOrderRepo) MarkPaid(ctx context.Context, orderID int) (int64, error) {
	m.statusMap[orderID] = "paid"
	return 2000, nil
}

func (m *mockOrderRepo) MarkPaymentFailed(ctx context.Context, orderID int) error {
	return nil
}

func (m *mockOrderRepo) GetItems(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	return m.items, nil
}

func (m *mockOrderRepo) GetItemsForStockRestore(ctx context.Context, orderID int) ([]repositories.StockRestoreItem, error) {
	return nil, nil
}

func (m *mockOrderRepo) GetStatus(ctx context.Context, orderID int) (string, error) {
	if s, ok := m.statusMap[orderID]; ok {
		return s, nil
	}
	return "pending", nil
}

func (m *mockOrderRepo) GetUserIDAndTotal(ctx context.Context, orderID int) (int, int, error) {
	return 1, 2000, nil
}

func (m *mockOrderRepo) GetStats(ctx context.Context) (*repositories.DashboardStats, error) {
	return m.stats, nil
}

type mockIdempotencyRepo struct{}

func (m *mockIdempotencyRepo) FindByKey(ctx context.Context, key string, userID int) (*repositories.IdempotencyResult, error) {
	return nil, repositories.ErrNotFound
}

func (m *mockIdempotencyRepo) Store(ctx context.Context, key string, userID, orderID int) error {
	return nil
}

func (m *mockIdempotencyRepo) Upsert(ctx context.Context, key string, userID, orderID int, paymentIntentID string) (*repositories.IdempotencyResult, error) {
	return &repositories.IdempotencyResult{OrderID: orderID, PaymentIntentID: paymentIntentID}, nil
}

type mockCheckoutSessionRepo struct{}

func (m *mockCheckoutSessionRepo) Create(ctx context.Context, params repositories.CreateCheckoutSessionParams) (*repositories.CheckoutSession, error) {
	return &repositories.CheckoutSession{
		ID:              1,
		UserID:          params.UserID,
		IdempotencyKey:  params.IdempotencyKey,
		PaymentIntentID: params.PaymentIntentID,
		Status:          "initialized",
	}, nil
}

func (m *mockCheckoutSessionRepo) FindByKey(ctx context.Context, idempotencyKey string, userID int) (*repositories.CheckoutSession, error) {
	return nil, repositories.ErrNotFound
}

func (m *mockCheckoutSessionRepo) UpdateStatus(ctx context.Context, idempotencyKey string, userID int, status string) error {
	return nil
}

func (m *mockCheckoutSessionRepo) UpdateOrderID(ctx context.Context, idempotencyKey string, userID int, orderID int) error {
	return nil
}

func (m *mockCheckoutSessionRepo) FindAbandoned(ctx context.Context, olderThan time.Duration, limit int) ([]repositories.CheckoutSession, error) {
	return nil, nil
}

func (m *mockCheckoutSessionRepo) ExpireOldSessions(ctx context.Context, olderThan time.Duration) (int, error) {
	return 0, nil
}

func orderAuthReq(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestListOrders_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	req := orderAuthReq("GET", "/api/orders?page=1&limit=10")
	w := httptest.NewRecorder()
	authMiddleware(handler.ListOrders)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["orders"] == nil {
		t.Error("expected orders in response")
	}
}

func TestListOrders_Unauthorized(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	req := httptest.NewRequest("GET", "/api/orders", nil)
	w := httptest.NewRecorder()
	handler.ListOrders(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGetOrder_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	req := orderAuthReq("GET", "/api/orders/1")
	w := httptest.NewRecorder()
	authMiddleware(handler.GetOrder)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetOrder_InvalidID(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	req := orderAuthReq("GET", "/api/orders/abc")
	w := httptest.NewRecorder()
	authMiddleware(handler.GetOrder)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAdminListOrders_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	req := httptest.NewRequest("GET", "/api/admin/orders?page=1&limit=10", nil)
	claims := &middleware.Claims{UserID: 1, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.AdminListOrders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAdminUpdateOrderStatus_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	body := models.UpdateOrderStatusRequest{Status: "paid"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/orders/1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.AdminUpdateOrderStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminUpdateOrderStatus_InvalidStatus(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	body := models.UpdateOrderStatusRequest{Status: "invalid"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/orders/1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.AdminUpdateOrderStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAdminUpdateOrderStatus_InvalidTransition(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	body := models.UpdateOrderStatusRequest{Status: "delivered"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/orders/1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.AdminUpdateOrderStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestInitCheckout_Unauthorized(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	req := httptest.NewRequest("POST", "/api/checkout/init", nil)
	w := httptest.NewRecorder()
	handler.InitCheckout(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestInitCheckout_EmptyCart(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	body := models.CheckoutInitRequest{IdempotencyKey: "test-key"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/checkout/init", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	authMiddleware(handler.InitCheckout)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestConfirmCheckout_Unauthorized(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	req := httptest.NewRequest("POST", "/api/checkout/confirm", nil)
	w := httptest.NewRecorder()
	handler.ConfirmCheckout(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestConfirmCheckout_MissingFields(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	body := models.CheckoutConfirmRequest{
		ShippingAddress: models.Address{},
		IdempotencyKey:  "test-key",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/checkout/confirm", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	authMiddleware(handler.ConfirmCheckout)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestConfirmCheckout_NoIdempotencyKey(t *testing.T) {
	orderRepo := newMockOrderRepo()
	cartRepo := newMockCartRepo()
	userRepo := newMockUserRepo()
	idempRepo := &mockIdempotencyRepo{}
	sessionRepo := &mockCheckoutSessionRepo{}
	handler := NewOrderHandler(orderRepo, cartRepo, userRepo, idempRepo, sessionRepo, nil)

	body := models.CheckoutConfirmRequest{
		ShippingAddress: models.Address{
			FirstName: "John",
			LastName:  "Doe",
			Address1:  "123 Main St",
			City:      "Springfield",
			State:     "IL",
			Zip:       "62701",
			Country:   "US",
		},
		IdempotencyKey: "",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/checkout/confirm", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	authMiddleware(handler.ConfirmCheckout)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
