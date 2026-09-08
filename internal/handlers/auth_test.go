package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ecommerce-api-go/internal/middleware"
	"ecommerce-api-go/internal/models"
	"ecommerce-api-go/internal/repositories"
)

type mockUserRepo struct {
	users      map[int]*models.User
	emailIndex map[string]*models.User
	passIndex  map[string]string
	nextID     int
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:      make(map[int]*models.User),
		emailIndex: make(map[string]*models.User),
		passIndex:  make(map[string]string),
		nextID:     1,
	}
}

func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := m.emailIndex[email]
	return ok, nil
}

func (m *mockUserRepo) Create(ctx context.Context, email, passwordHash, firstName, lastName string) (*models.User, error) {
	user := &models.User{
		ID:        m.nextID,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      "user",
		CreatedAt: time.Now(),
	}
	m.users[m.nextID] = user
	m.emailIndex[email] = user
	m.passIndex[email] = passwordHash
	m.nextID++
	return user, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, string, error) {
	user, ok := m.emailIndex[email]
	if !ok {
		return nil, "", repositories.ErrNotFound
	}
	return user, m.passIndex[email], nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepo) UpdateProfile(ctx context.Context, id int, firstName, lastName string) error {
	user, ok := m.users[id]
	if !ok {
		return repositories.ErrNotFound
	}
	user.FirstName = firstName
	user.LastName = lastName
	return nil
}

func (m *mockUserRepo) UpdateRole(ctx context.Context, id int, role string) error {
	user, ok := m.users[id]
	if !ok {
		return repositories.ErrNotFound
	}
	user.Role = role
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id int) error {
	if _, ok := m.users[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	var users []models.User
	for _, u := range m.users {
		users = append(users, *u)
	}
	return users, nil
}

func (m *mockUserRepo) Count(ctx context.Context) (int, error) {
	return len(m.users), nil
}

func init() {
	middleware.JWTSecret = []byte("test-secret-key-for-testing")
}

func generateTestToken(userID int, email, role string) string {
	claims := &middleware.Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString(middleware.JWTSecret)
	return s
}

func authHeaders(user models.User) map[string]string {
	token := generateTestToken(user.ID, user.Email, user.Role)
	return map[string]string{"Authorization": "Bearer " + token}
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	body := models.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.AuthResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Token == "" {
		t.Error("expected token in response")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.User.Email)
	}
	if resp.User.FirstName != "John" {
		t.Errorf("expected first name John, got %s", resp.User.FirstName)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	tests := []struct {
		name       string
		body       models.RegisterRequest
		wantStatus int
	}{
		{
			name:       "empty email",
			body:       models.RegisterRequest{Email: "", Password: "password123", FirstName: "John", LastName: "Doe"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty password",
			body:       models.RegisterRequest{Email: "test@example.com", Password: "", FirstName: "John", LastName: "Doe"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "short password",
			body:       models.RegisterRequest{Email: "test@example.com", Password: "123", FirstName: "John", LastName: "Doe"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	body := models.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	body2 := models.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password456",
		FirstName: "Jane",
		LastName:  "Doe",
	}
	jsonBody2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.Register(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w2.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	regBody := models.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	jsonBody, _ := json.Marshal(regBody)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	loginBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonLogin, _ := json.Marshal(loginBody)
	req2 := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(jsonLogin))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.Login(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp models.AuthResponse
	json.NewDecoder(w2.Body).Decode(&resp)
	if resp.Token == "" {
		t.Error("expected token in response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	regBody := models.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	jsonBody, _ := json.Marshal(regBody)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	loginBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	jsonLogin, _ := json.Marshal(loginBody)
	req2 := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(jsonLogin))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.Login(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w2.Code)
	}
}

func TestLogin_NonexistentEmail(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	loginBody := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	jsonLogin, _ := json.Marshal(loginBody)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(jsonLogin))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGetProfile_Success(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	regBody := models.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	jsonBody, _ := json.Marshal(regBody)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	var regResp models.AuthResponse
	json.NewDecoder(w.Body).Decode(&regResp)

	req2 := httptest.NewRequest("GET", "/api/auth/me", nil)
	for k, v := range authHeaders(regResp.User) {
		req2.Header.Set(k, v)
	}
	w2 := httptest.NewRecorder()
	middleware.AuthMiddleware(handler.GetProfile)(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	var user models.User
	json.NewDecoder(w2.Body).Decode(&user)
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
}

func TestGetProfile_Unauthorized(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()
	handler.GetProfile(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	regBody := models.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	jsonBody, _ := json.Marshal(regBody)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	var regResp models.AuthResponse
	json.NewDecoder(w.Body).Decode(&regResp)

	updateBody := models.UpdateProfileRequest{
		FirstName: "Jane",
		LastName:  "Smith",
	}
	jsonUpdate, _ := json.Marshal(updateBody)
	req2 := httptest.NewRequest("PUT", "/api/auth/me/update", bytes.NewReader(jsonUpdate))
	req2.Header.Set("Content-Type", "application/json")
	for k, v := range authHeaders(regResp.User) {
		req2.Header.Set(k, v)
	}
	w2 := httptest.NewRecorder()
	middleware.AuthMiddleware(handler.UpdateProfile)(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var user models.User
	json.NewDecoder(w2.Body).Decode(&user)
	if user.FirstName != "Jane" {
		t.Errorf("expected first name Jane, got %s", user.FirstName)
	}
}

func TestUpdateProfile_MissingNames(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	regBody := models.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	jsonBody, _ := json.Marshal(regBody)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	var regResp models.AuthResponse
	json.NewDecoder(w.Body).Decode(&regResp)

	updateBody := models.UpdateProfileRequest{
		FirstName: "",
		LastName:  "Smith",
	}
	jsonUpdate, _ := json.Marshal(updateBody)
	req2 := httptest.NewRequest("PUT", "/api/auth/me/update", bytes.NewReader(jsonUpdate))
	req2.Header.Set("Content-Type", "application/json")
	for k, v := range authHeaders(regResp.User) {
		req2.Header.Set(k, v)
	}
	w2 := httptest.NewRecorder()
	middleware.AuthMiddleware(handler.UpdateProfile)(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w2.Code)
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRegister_MethodNotAllowed(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewAuthHandler(repo)

	req := httptest.NewRequest("GET", "/api/auth/register", nil)
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
