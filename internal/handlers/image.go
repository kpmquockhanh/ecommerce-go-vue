package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-api-go/internal/middleware"
	"ecommerce-api-go/internal/repositories"
	"ecommerce-api-go/internal/storage"
)

type ImageHandler struct {
	productRepo repositories.ProductRepository
}

func NewImageHandler(productRepo repositories.ProductRepository) *ImageHandler {
	return &ImageHandler{productRepo: productRepo}
}

func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.Role != "admin" {
		respondWithError(w, "Admin access required", http.StatusForbidden)
		return
	}

	if storage.Client == nil {
		respondWithError(w, "S3 storage not configured", http.StatusServiceUnavailable)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondWithError(w, "File too large (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		respondWithError(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size <= 0 {
		respondWithError(w, "Empty or invalid file", http.StatusBadRequest)
		return
	}
	if header.Size > 10<<20 {
		respondWithError(w, "File too large (max 10MB)", http.StatusBadRequest)
		return
	}

	objectName, err := storage.UploadImage(r.Context(), file, header)
	if err != nil {
		msg := "Failed to upload image"
		if strings.Contains(err.Error(), "unsupported file type") {
			msg = "Unsupported file type. Allowed: jpg, jpeg, png, webp, gif"
		} else if strings.Contains(err.Error(), "empty") {
			msg = "Empty file"
		} else if strings.Contains(err.Error(), "too large") {
			msg = "File too large (max 10MB)"
		}
		respondWithError(w, msg, http.StatusBadRequest)
		return
	}

	presignedURL, err := storage.GetPresignedURL(r.Context(), objectName)
	if err != nil {
		respondWithServerError(w, "Failed to generate image URL", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{
		"path": objectName,
		"url":  presignedURL,
	}, http.StatusOK)
}

func (h *ImageHandler) GetProductImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/images/product/")
	if path == "" || strings.Contains(path, "/") {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	p, err := h.productRepo.FindByID(r.Context(), id)
	if err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	if p.Status != "published" {
		claims := middleware.GetUserFromContext(r)
		if claims == nil || claims.Role != "admin" {
			respondWithError(w, "Product not found", http.StatusNotFound)
			return
		}
	}

	images, err := h.productRepo.GetImages(r.Context(), id)
	if err != nil {
		respondWithError(w, "Product not found", http.StatusNotFound)
		return
	}

	var urls []string
	if storage.Client != nil {
		for _, img := range images {
			url, err := storage.GetPresignedURL(r.Context(), img)
			if err == nil {
				urls = append(urls, url)
			}
		}
	} else {
		urls = images
	}

	respondWithJSON(w, map[string]interface{}{
		"images": urls,
	}, http.StatusOK)
}

func (h *ImageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.Role != "admin" {
		respondWithError(w, "Admin access required", http.StatusForbidden)
		return
	}

	if storage.Client == nil {
		respondWithError(w, "S3 storage not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Path      string `json:"path"`
		ProductID *int   `json:"product_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		respondWithError(w, "Image path is required", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(req.Path, "products/") {
		respondWithError(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	if err := storage.DeleteImage(r.Context(), req.Path); err != nil {
		respondWithServerError(w, "Failed to delete image", err, http.StatusInternalServerError)
		return
	}

	if req.ProductID != nil {
		if err := h.productRepo.RemoveImageFromProduct(r.Context(), *req.ProductID, req.Path); err != nil && err != repositories.ErrNotFound {
			respondWithServerError(w, "Failed to remove image from product", err, http.StatusInternalServerError)
			return
		}
	}

	respondWithJSON(w, map[string]string{"message": "Image deleted"}, http.StatusOK)
}

func (h *ImageHandler) UploadProductImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.Role != "admin" {
		respondWithError(w, "Admin access required", http.StatusForbidden)
		return
	}

	if storage.Client == nil {
		respondWithError(w, "S3 storage not configured", http.StatusServiceUnavailable)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
	idStr = strings.TrimSuffix(idStr, "/images")
	productID, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondWithError(w, "File too large (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		respondWithError(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size <= 0 {
		respondWithError(w, "Empty or invalid file", http.StatusBadRequest)
		return
	}
	if header.Size > 10<<20 {
		respondWithError(w, "File too large (max 10MB)", http.StatusBadRequest)
		return
	}

	objectName, err := storage.UploadImage(r.Context(), file, header)
	if err != nil {
		msg := "Failed to upload image"
		if strings.Contains(err.Error(), "unsupported file type") {
			msg = "Unsupported file type. Allowed: jpg, jpeg, png, webp, gif"
		} else if strings.Contains(err.Error(), "empty") {
			msg = "Empty file"
		} else if strings.Contains(err.Error(), "too large") {
			msg = "File too large (max 10MB)"
		}
		respondWithError(w, msg, http.StatusBadRequest)
		return
	}

	if err := h.productRepo.AddImageToProduct(r.Context(), productID, objectName); err != nil {
		_ = storage.DeleteImage(r.Context(), objectName)
		if err == repositories.ErrNotFound {
			respondWithError(w, "Product not found", http.StatusNotFound)
			return
		}
		respondWithServerError(w, "Failed to link image to product", err, http.StatusInternalServerError)
		return
	}

	presignedURL, err := storage.GetPresignedURL(r.Context(), objectName)
	if err != nil {
		_ = storage.DeleteImage(r.Context(), objectName)
		_ = h.productRepo.RemoveImageFromProduct(r.Context(), productID, objectName)
		respondWithServerError(w, "Failed to generate image URL", err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, map[string]string{
		"path": objectName,
		"url":  presignedURL,
	}, http.StatusOK)
}
