package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-api-go/internal/middleware"
	"ecommerce-api-go/internal/models"
)

type mockReviewRepoForHandler struct {
	reviews   []models.Review
	nextID    int
	createErr error
	existMap  map[int]bool
}

func newMockReviewRepoForHandler() *mockReviewRepoForHandler {
	return &mockReviewRepoForHandler{
		nextID:   1,
		existMap: map[int]bool{1: true},
	}
}

func (m *mockReviewRepoForHandler) Create(ctx context.Context, productID, userID, rating int, comment string) (*models.Review, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	review := &models.Review{
		ID:        m.nextID,
		ProductID: productID,
		UserID:    userID,
		Rating:    rating,
		Comment:   comment,
	}
	m.nextID++
	return review, nil
}

func (m *mockReviewRepoForHandler) ListByProduct(ctx context.Context, productID, limit, offset int) ([]models.Review, error) {
	return m.reviews, nil
}

func (m *mockReviewRepoForHandler) GetStats(ctx context.Context, productID int) (float64, int, error) {
	return 4.5, len(m.reviews), nil
}

func (m *mockReviewRepoForHandler) ExistsProduct(ctx context.Context, productID int) (bool, error) {
	if v, ok := m.existMap[productID]; ok {
		return v, nil
	}
	return false, nil
}

func TestCreateReview_Success(t *testing.T) {
	reviewRepo := newMockReviewRepoForHandler()
	productRepo := newMockProductRepo()
	handler := NewReviewHandler(reviewRepo, productRepo)

	body := models.CreateReviewRequest{
		ProductID: 1,
		Rating:    5,
		Comment:   "Great product!",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	authMiddleware(handler.CreateReview)(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var review models.Review
	json.NewDecoder(w.Body).Decode(&review)
	if review.Rating != 5 {
		t.Errorf("expected rating 5, got %d", review.Rating)
	}
}

func TestCreateReview_InvalidRating(t *testing.T) {
	reviewRepo := newMockReviewRepoForHandler()
	productRepo := newMockProductRepo()
	handler := NewReviewHandler(reviewRepo, productRepo)

	tests := []struct {
		name   string
		rating int
	}{
		{"zero rating", 0},
		{"negative rating", -1},
		{"rating too high", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := models.CreateReviewRequest{
				ProductID: 1,
				Rating:    tt.rating,
				Comment:   "Test",
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
			token := generateTestToken(claims.UserID, claims.Email, claims.Role)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			authMiddleware(handler.CreateReview)(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestCreateReview_Unauthorized(t *testing.T) {
	reviewRepo := newMockReviewRepoForHandler()
	productRepo := newMockProductRepo()
	handler := NewReviewHandler(reviewRepo, productRepo)

	body := models.CreateReviewRequest{
		ProductID: 1,
		Rating:    5,
		Comment:   "Great!",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateReview(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestCreateReview_ProductNotFound(t *testing.T) {
	reviewRepo := newMockReviewRepoForHandler()
	productRepo := newMockProductRepo()
	handler := NewReviewHandler(reviewRepo, productRepo)

	body := models.CreateReviewRequest{
		ProductID: 999,
		Rating:    5,
		Comment:   "Great!",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	authMiddleware(handler.CreateReview)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListReviews_Success(t *testing.T) {
	reviewRepo := newMockReviewRepoForHandler()
	reviewRepo.reviews = []models.Review{
		{ID: 1, ProductID: 1, Rating: 5, Comment: "Great!"},
		{ID: 2, ProductID: 1, Rating: 4, Comment: "Good"},
	}
	productRepo := newMockProductRepo()
	handler := NewReviewHandler(reviewRepo, productRepo)

	req := httptest.NewRequest("GET", "/api/reviews/product/1?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.ListReviews(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp models.ReviewListResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Reviews) != 2 {
		t.Errorf("expected 2 reviews, got %d", len(resp.Reviews))
	}
}

func TestListReviews_InvalidProductID(t *testing.T) {
	reviewRepo := newMockReviewRepoForHandler()
	productRepo := newMockProductRepo()
	handler := NewReviewHandler(reviewRepo, productRepo)

	req := httptest.NewRequest("GET", "/api/reviews/product/abc", nil)
	w := httptest.NewRecorder()

	handler.ListReviews(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateReview_InvalidJSON(t *testing.T) {
	reviewRepo := newMockReviewRepoForHandler()
	productRepo := newMockProductRepo()
	handler := NewReviewHandler(reviewRepo, productRepo)

	req := httptest.NewRequest("POST", "/api/reviews", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "test@example.com", Role: "user"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	authMiddleware(handler.CreateReview)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
