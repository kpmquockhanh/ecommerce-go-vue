package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ecommerce-api-go/internal/middleware"
	"ecommerce-api-go/internal/models"
)

func adminAuthReq(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	claims := &middleware.Claims{UserID: 1, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestListUsers_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.Create(context.Background(), "user1@test.com", "hash", "User", "One")
	userRepo.Create(context.Background(), "user2@test.com", "hash", "User", "Two")
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := adminAuthReq("GET", "/api/admin/users?page=1&limit=10")
	w := httptest.NewRecorder()

	handler.ListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["users"] == nil {
		t.Error("expected users in response")
	}
}

func TestGetUser_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.Create(context.Background(), "user@test.com", "hash", "User", "Test")
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := adminAuthReq("GET", "/api/admin/users?id=1")
	w := httptest.NewRecorder()

	handler.GetUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var user models.User
	json.NewDecoder(w.Body).Decode(&user)
	if user.Email != "user@test.com" {
		t.Errorf("expected email user@test.com, got %s", user.Email)
	}
}

func TestGetUser_MissingID(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := adminAuthReq("GET", "/api/admin/users")
	w := httptest.NewRecorder()

	handler.GetUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := adminAuthReq("GET", "/api/admin/users?id=999")
	w := httptest.NewRecorder()

	handler.GetUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestGetUser_InvalidID(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := adminAuthReq("GET", "/api/admin/users?id=abc")
	w := httptest.NewRecorder()

	handler.GetUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateUserRole_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	adminUser, _ := userRepo.Create(context.Background(), "admin@example.com", "hash", "Admin", "User")
	user, _ := userRepo.Create(context.Background(), "user@test.com", "hash", "User", "Test")
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	body := map[string]string{"role": "admin"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/users/"+strconv.Itoa(user.ID), bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: adminUser.ID, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.AdminMiddleware(handler.UpdateUserRole)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateUserRole_CannotChangeOwnRole(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	body := map[string]string{"role": "user"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/users/1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.AdminMiddleware(handler.UpdateUserRole)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateUserRole_InvalidRole(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.Create(context.Background(), "user@test.com", "hash", "User", "Test")
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	body := map[string]string{"role": "superadmin"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/users/2", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.AdminMiddleware(handler.UpdateUserRole)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateUserRole_Unauthorized(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	body := map[string]string{"role": "admin"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/users/2", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateUserRole(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	adminUser, _ := userRepo.Create(context.Background(), "admin@example.com", "hash", "Admin", "User")
	user, _ := userRepo.Create(context.Background(), "user@test.com", "hash", "User", "Test")
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := httptest.NewRequest("DELETE", "/api/admin/users/"+strconv.Itoa(user.ID), nil)
	claims := &middleware.Claims{UserID: adminUser.ID, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.AdminMiddleware(handler.DeleteUser)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteUser_CannotDeleteSelf(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := adminAuthReq("DELETE", "/api/admin/users/1")
	w := httptest.NewRecorder()

	middleware.AdminMiddleware(handler.DeleteUser)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := adminAuthReq("DELETE", "/api/admin/users/999")
	w := httptest.NewRecorder()

	middleware.AdminMiddleware(handler.DeleteUser)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestDeleteUser_InvalidID(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := adminAuthReq("DELETE", "/api/admin/users/abc")
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDashboardStats_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := adminAuthReq("GET", "/api/admin/stats")
	w := httptest.NewRecorder()

	handler.DashboardStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["total_orders"] == nil {
		t.Error("expected total_orders in response")
	}
}

func TestDashboardStats_MethodNotAllowed(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := httptest.NewRequest("POST", "/api/admin/stats", nil)
	w := httptest.NewRecorder()

	handler.DashboardStats(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestListUsers_MethodNotAllowed(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := httptest.NewRequest("POST", "/api/admin/users", nil)
	w := httptest.NewRecorder()

	handler.ListUsers(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestUpdateUserRole_MethodNotAllowed(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := httptest.NewRequest("GET", "/api/admin/users/1", nil)
	w := httptest.NewRecorder()

	handler.UpdateUserRole(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestDeleteUser_MethodNotAllowed(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	req := httptest.NewRequest("GET", "/api/admin/users/1", nil)
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestUpdateUserRole_UserNotFound(t *testing.T) {
	userRepo := newMockUserRepo()
	orderRepo := newMockOrderRepo()
	productRepo := newMockProductRepo()
	handler := NewAdminHandler(userRepo, orderRepo, productRepo)

	body := map[string]string{"role": "admin"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/users/999", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	claims := &middleware.Claims{UserID: 1, Email: "admin@example.com", Role: "admin"}
	token := generateTestToken(claims.UserID, claims.Email, claims.Role)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.AdminMiddleware(handler.UpdateUserRole)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
