package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/repositories"
)

type mockCategoryRepoForHandler struct {
	categories map[int]*models.Category
	nextID     int
}

func newMockCategoryRepoForHandler() *mockCategoryRepoForHandler {
	return &mockCategoryRepoForHandler{
		categories: make(map[int]*models.Category),
		nextID:     1,
	}
}

func (m *mockCategoryRepoForHandler) List(ctx context.Context) ([]models.Category, error) {
	var cats []models.Category
	for _, c := range m.categories {
		cats = append(cats, *c)
	}
	return cats, nil
}

func (m *mockCategoryRepoForHandler) FindByID(ctx context.Context, id int) (*models.Category, error) {
	c, ok := m.categories[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return c, nil
}

func (m *mockCategoryRepoForHandler) Create(ctx context.Context, name string) (*models.Category, error) {
	cat := &models.Category{
		ID:   m.nextID,
		Name: name,
	}
	m.categories[m.nextID] = cat
	m.nextID++
	return cat, nil
}

func (m *mockCategoryRepoForHandler) Update(ctx context.Context, id int, name string) error {
	c, ok := m.categories[id]
	if !ok {
		return repositories.ErrNotFound
	}
	c.Name = name
	return nil
}

func (m *mockCategoryRepoForHandler) Delete(ctx context.Context, id int) error {
	if _, ok := m.categories[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(m.categories, id)
	return nil
}

func (m *mockCategoryRepoForHandler) CountProducts(ctx context.Context, categoryID int) (int, error) {
	return 0, nil
}

func (m *mockCategoryRepoForHandler) RemoveCategoryFromProducts(ctx context.Context, categoryID int) error {
	return nil
}

func TestListCategories_Success(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()

	handler.ListCategories(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["categories"] == nil {
		t.Error("expected categories in response")
	}
}

func TestListPublicCategories_FilterEmpty(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()

	handler.ListPublicCategories(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCreateCategory_Success(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	body := map[string]string{"name": "Electronics"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/categories/create", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateCategory(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var cat models.Category
	json.NewDecoder(w.Body).Decode(&cat)
	if cat.Name != "Electronics" {
		t.Errorf("expected name Electronics, got %s", cat.Name)
	}
}

func TestCreateCategory_EmptyName(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	body := map[string]string{"name": "  "}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/categories/create", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateCategory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateCategory_InvalidJSON(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("POST", "/api/admin/categories/create", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateCategory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateCategory_Success(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	cat, _ := repo.Create(context.Background(), "Old Name")
	handler := NewCategoryHandler(repo)

	body := map[string]string{"name": "New Name"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/categories/"+strconv.Itoa(cat.ID), bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateCategory(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateCategory_NotFound(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	body := map[string]string{"name": "New Name"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/categories/999", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateCategory(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestUpdateCategory_InvalidID(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	body := map[string]string{"name": "New Name"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/categories/abc", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateCategory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteCategory_Success(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	cat, _ := repo.Create(context.Background(), "To Delete")
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("DELETE", "/api/admin/categories/"+strconv.Itoa(cat.ID), nil)
	w := httptest.NewRecorder()

	handler.DeleteCategory(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteCategory_NotFound(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("DELETE", "/api/admin/categories/999", nil)
	w := httptest.NewRecorder()

	handler.DeleteCategory(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestDeleteCategory_InvalidID(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("DELETE", "/api/admin/categories/abc", nil)
	w := httptest.NewRecorder()

	handler.DeleteCategory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestListCategories_MethodNotAllowed(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("POST", "/api/categories", nil)
	w := httptest.NewRecorder()

	handler.ListCategories(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestCreateCategory_MethodNotAllowed(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("GET", "/api/admin/categories/create", nil)
	w := httptest.NewRecorder()

	handler.CreateCategory(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestUpdateCategory_MethodNotAllowed(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("GET", "/api/admin/categories/1", nil)
	w := httptest.NewRecorder()

	handler.UpdateCategory(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestDeleteCategory_MethodNotAllowed(t *testing.T) {
	repo := newMockCategoryRepoForHandler()
	handler := NewCategoryHandler(repo)

	req := httptest.NewRequest("GET", "/api/admin/categories/1", nil)
	w := httptest.NewRecorder()

	handler.DeleteCategory(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
