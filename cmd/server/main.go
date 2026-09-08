package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/stripe/stripe-go/v82"
	"golang.org/x/time/rate"

	"ecommerce-api-go/internal/checkout"
	"ecommerce-api-go/internal/config"
	"ecommerce-api-go/internal/database"
	"ecommerce-api-go/internal/handlers"
	"ecommerce-api-go/internal/mail"
	"ecommerce-api-go/internal/middleware"
	"ecommerce-api-go/internal/queue"
	"ecommerce-api-go/internal/repositories"
	"ecommerce-api-go/internal/storage"
	"ecommerce-api-go/internal/worker"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	if err := database.Migrate(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := database.Seed(ctx); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Initialize S3 storage (optional)
	if cfg.S3Endpoint != "" {
		if err := storage.Init(storage.Config{
			EndPoint:       cfg.S3Endpoint,
			PublicEndPoint: cfg.S3PublicEndpoint,
			AccessKey:      cfg.S3AccessKey,
			SecretKey:      cfg.S3SecretKey,
			Bucket:         cfg.S3Bucket,
			UseSSL:         cfg.S3UseSSL,
			PublicUseSSL:   cfg.S3PublicUseSSL,
		}); err != nil {
			log.Printf("Warning: S3 storage not available: %v", err)
		}
	}

	// Initialize JWT
	middleware.JWTSecret = []byte(cfg.JWTSecret)

	// Initialize Stripe
	stripe.Key = cfg.StripeSecretKey

	// Initialize repositories
	db := database.DB
	userRepo := repositories.NewUserRepository(db)
	productRepo := repositories.NewProductRepository(db)
	cartRepo := repositories.NewCartRepository(db)
	reviewRepo := repositories.NewReviewRepository(db)
	stockRepo := repositories.NewStockRepository(db)
	deadLetterRepo := repositories.NewDeadLetterRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	idempotencyRepo := repositories.NewIdempotencyRepository(db)
	checkoutSessionRepo := repositories.NewCheckoutSessionRepository(db)

	// Initialize queue and worker (optional — won't crash if RabbitMQ is unavailable)
	var q *queue.Queue
	var w *worker.Worker
	var mailer *mail.Mailer

	q, err := queue.New(cfg.RabbitMQURL, deadLetterRepo)
	if err != nil {
		log.Printf("Warning: RabbitMQ not available, background jobs disabled: %v", err)
	} else {
		mailer = mail.NewMailer(mail.Config{
			SMTPHost:     cfg.SMTPHost,
			SMTPPort:     cfg.SMTPPort,
			SMTPUsername: cfg.SMTPUsername,
			SMTPPassword: cfg.SMTPPassword,
			FromEmail:    cfg.FromEmail,
			FromName:     cfg.FromName,
		})
		w = worker.New(q, mailer)
	}

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	paymentHandler := handlers.NewPaymentHandler()
	authHandler := handlers.NewAuthHandler(userRepo)
	productHandler := handlers.NewProductHandler(productRepo, reviewRepo, categoryRepo)
	cartHandler := handlers.NewCartHandler(cartRepo, productRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, cartRepo, userRepo, idempotencyRepo, checkoutSessionRepo, q)
	reviewHandler := handlers.NewReviewHandler(reviewRepo, productRepo)
	imageHandler := handlers.NewImageHandler(productRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, orderRepo, productRepo)
	webhookHandler := handlers.NewWebhookHandler(cfg.StripeWebhookSecret, orderRepo, cartRepo, stockRepo, userRepo, q)
	configHandler := handlers.NewConfigHandler(cfg.StripePublishableKey)
	deadLetterHandler := handlers.NewDeadLetterHandler(deadLetterRepo, q)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter(rate.Limit(10), 100)

	// Auth routes
	http.HandleFunc("/api/auth/register", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(authHandler.Register)))
	http.HandleFunc("/api/auth/login", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(authHandler.Login)))
	http.HandleFunc("/api/auth/me", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AuthMiddleware(authHandler.GetProfile))))
	http.HandleFunc("/api/auth/me/update", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AuthMiddleware(authHandler.UpdateProfile))))

	// Product routes (public)
	http.HandleFunc("/api/products", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(productHandler.ListProducts)))
	http.HandleFunc("/api/products/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if strings.HasSuffix(path, "/options") && r.Method == "GET" {
				productHandler.GetOptionGroups(w, r)
				return
			}
			if strings.HasSuffix(path, "/customizations") && r.Method == "GET" {
				productHandler.GetCustomizationFields(w, r)
				return
			}
			productHandler.GetProduct(w, r)
		})))

	// Public category route (storefront navigation)
	http.HandleFunc("/api/categories", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(categoryHandler.ListPublicCategories)))

	// Admin product routes
	http.HandleFunc("/api/admin/products", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				productHandler.AdminListProducts(w, r)
			case "POST":
				productHandler.CreateProduct(w, r)
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			}
		}))))

	// Cart routes (optional auth for guest support)
	http.HandleFunc("/api/cart", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.OptionalAuthMiddleware(cartHandler.GetCart))))
	http.HandleFunc("/api/cart/items", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.OptionalAuthMiddleware(cartHandler.AddItem))))
	http.HandleFunc("/api/cart/items/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.OptionalAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "PUT":
				cartHandler.UpdateItem(w, r)
			case "DELETE":
				cartHandler.RemoveItem(w, r)
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			}
		}))))
	http.HandleFunc("/api/cart/merge", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AuthMiddleware(cartHandler.MergeGuestCart))))

	// Order routes (auth required)
	http.HandleFunc("/api/checkout/init", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AuthMiddleware(
			rateLimiter.Middleware(orderHandler.InitCheckout)))))
	http.HandleFunc("/api/checkout/confirm", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AuthMiddleware(
			rateLimiter.Middleware(orderHandler.ConfirmCheckout)))))
	http.HandleFunc("/api/orders", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AuthMiddleware(orderHandler.ListOrders))))
	http.HandleFunc("/api/orders/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AuthMiddleware(orderHandler.GetOrder))))

	// Admin order routes
	http.HandleFunc("/api/admin/orders", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(orderHandler.AdminListOrders))))
	http.HandleFunc("/api/admin/orders/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(orderHandler.AdminUpdateOrderStatus))))

	// Admin user routes
	http.HandleFunc("/api/admin/users", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(adminHandler.ListUsers))))
	http.HandleFunc("/api/admin/users/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				adminHandler.GetUser(w, r)
			case "PUT":
				adminHandler.UpdateUserRole(w, r)
			case "DELETE":
				adminHandler.DeleteUser(w, r)
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			}
		}))))

	// Admin dashboard stats
	http.HandleFunc("/api/admin/stats", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(adminHandler.DashboardStats))))

	// Admin dead letter routes
	http.HandleFunc("/api/admin/dead-letters", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(deadLetterHandler.ListDeadLetters))))
	http.HandleFunc("/api/admin/dead-letters/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "POST":
				deadLetterHandler.RetryDeadLetter(w, r)
			case "DELETE":
				deadLetterHandler.DeleteDeadLetter(w, r)
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			}
		}))))

	// Admin category routes
	http.HandleFunc("/api/admin/categories", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(categoryHandler.ListCategories))))
	http.HandleFunc("/api/admin/categories/create", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(categoryHandler.CreateCategory))))
	http.HandleFunc("/api/admin/categories/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "PUT":
				categoryHandler.UpdateCategory(w, r)
			case "DELETE":
				categoryHandler.DeleteCategory(w, r)
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			}
		}))))

	// Review routes (uses numeric product ID, different from slug-based product routes)
	http.HandleFunc("/api/reviews/product/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(reviewHandler.ListReviews)))
	http.HandleFunc("/api/reviews", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AuthMiddleware(reviewHandler.CreateReview))))

	// Image routes (admin only for upload/delete)
	http.HandleFunc("/api/images/upload", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(imageHandler.UploadImage))))
	http.HandleFunc("/api/images/product/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(imageHandler.GetProductImages)))
	http.HandleFunc("/api/images/delete", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(imageHandler.DeleteImage))))

	// Admin product image upload (auto-links to product)
	http.HandleFunc("/api/admin/products/", middleware.CORSMiddleware(
		middleware.LoggingMiddleware(middleware.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// POST .../images — upload product image
			if r.Method == "POST" && strings.HasSuffix(path, "/images") {
				imageHandler.UploadProductImage(w, r)
				return
			}

			// POST .../tags — add tags
			if r.Method == "POST" && strings.HasSuffix(path, "/tags") {
				productHandler.AddTags(w, r)
				return
			}

			// DELETE .../tags/{tag} — remove tag
			if r.Method == "DELETE" && strings.Contains(path, "/tags/") {
				productHandler.RemoveTag(w, r)
				return
			}

			// PUT .../images/order — reorder images
			if r.Method == "PUT" && strings.HasSuffix(path, "/images/order") {
				productHandler.UpdateImageOrder(w, r)
				return
			}

			// POST .../variants/generate — auto-generate variants from option groups
			if r.Method == "POST" && strings.HasSuffix(path, "/variants/generate") {
				productHandler.GenerateVariants(w, r)
				return
			}

			// POST .../variants — create variant
			if r.Method == "POST" && strings.HasSuffix(path, "/variants") {
				productHandler.CreateVariant(w, r)
				return
			}

			// PATCH .../variants/bulk — bulk update variants
			if r.Method == "PATCH" && strings.HasSuffix(path, "/variants/bulk") {
				productHandler.BulkUpdateVariants(w, r)
				return
			}

			// PUT .../variants/{id} — update variant
			if r.Method == "PUT" && strings.Contains(path, "/variants/") {
				productHandler.UpdateVariant(w, r)
				return
			}

			// DELETE .../variants/{id} — delete variant
			if r.Method == "DELETE" && strings.Contains(path, "/variants/") {
				productHandler.DeleteVariant(w, r)
				return
			}

			// POST .../option-groups — create option group
			if r.Method == "POST" && strings.HasSuffix(path, "/option-groups") {
				productHandler.CreateOptionGroup(w, r)
				return
			}

			// PUT .../option-groups/{id} — update option group
			if r.Method == "PUT" && strings.Contains(path, "/option-groups/") && !strings.Contains(path, "/values/") && !strings.Contains(path, "/reorder") {
				productHandler.UpdateOptionGroup(w, r)
				return
			}

			// DELETE .../option-groups/{id} — delete option group
			if r.Method == "DELETE" && strings.Contains(path, "/option-groups/") && !strings.Contains(path, "/values/") {
				productHandler.DeleteOptionGroup(w, r)
				return
			}

			// POST .../option-groups/{id}/values — add option value
			if r.Method == "POST" && strings.Contains(path, "/option-groups/") && strings.HasSuffix(path, "/values") {
				productHandler.CreateOptionValue(w, r)
				return
			}

			// PUT .../option-groups/{id}/values/{id} — update option value
			if r.Method == "PUT" && strings.Contains(path, "/option-groups/") && strings.Contains(path, "/values/") && !strings.HasSuffix(path, "/reorder") {
				productHandler.UpdateOptionValue(w, r)
				return
			}

			// DELETE .../option-groups/{id}/values/{id} — delete option value
			if r.Method == "DELETE" && strings.Contains(path, "/option-groups/") && strings.Contains(path, "/values/") {
				productHandler.DeleteOptionValue(w, r)
				return
			}

			// PUT .../option-groups/{id}/reorder — reorder option values
			if r.Method == "PUT" && strings.Contains(path, "/option-groups/") && strings.HasSuffix(path, "/reorder") {
				productHandler.ReorderOptionValues(w, r)
				return
			}

			// POST .../customizations — create customization field
			if r.Method == "POST" && strings.HasSuffix(path, "/customizations") {
				productHandler.CreateCustomizationField(w, r)
				return
			}

			// PUT .../customizations/{id} — update customization field
			if r.Method == "PUT" && strings.Contains(path, "/customizations/") && !strings.HasSuffix(path, "/reorder") {
				productHandler.UpdateCustomizationField(w, r)
				return
			}

			// DELETE .../customizations/{id} — delete customization field
			if r.Method == "DELETE" && strings.Contains(path, "/customizations/") {
				productHandler.DeleteCustomizationField(w, r)
				return
			}

			// PUT .../customizations/reorder — reorder customization fields
			if r.Method == "PUT" && strings.HasSuffix(path, "/customizations/reorder") {
				productHandler.ReorderCustomizationFields(w, r)
				return
			}

			// GET — get product by ID
			// PUT — update product
			// DELETE — delete product
			switch r.Method {
			case "GET":
				productHandler.AdminGetProduct(w, r)
			case "PUT":
				productHandler.UpdateProduct(w, r)
			case "DELETE":
				productHandler.DeleteProduct(w, r)
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			}
		}))))

	// Existing routes
	http.HandleFunc("/create-payment-intent",
		middleware.CORSMiddleware(
			middleware.LoggingMiddleware(
				rateLimiter.Middleware(paymentHandler.CreatePaymentIntent))))

	// Stripe webhook (no CORS — server-to-server call)
	http.HandleFunc("/webhooks/stripe",
		middleware.LoggingMiddleware(webhookHandler.HandleStripeWebhook))

	http.HandleFunc("/health",
		middleware.LoggingMiddleware(healthHandler.Health))

	// Public config route (Stripe publishable key for frontend)
	http.HandleFunc("/api/config",
		middleware.LoggingMiddleware(configHandler.GetPublicConfig))

	// Start worker in background (if queue is available)
	if w != nil {
		workerCtx, workerCancel := context.WithCancel(context.Background())
		go func() {
			if err := w.Start(workerCtx); err != nil && err != context.Canceled {
				log.Printf("Worker error: %v", err)
			}
		}()
		defer workerCancel()
	}

	// Start checkout cleanup job
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	cleanupJob := checkout.NewCleanupJob(checkoutSessionRepo, 15*time.Minute, 24*time.Hour)
	go cleanupJob.Start(cleanupCtx)
	defer cleanupCancel()

	// Graceful shutdown
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: nil,
	}

	// Start server in goroutine
	go func() {
		if cfg.IsProduction {
			log.Printf("Server starting on port 443 (HTTPS - Production Mode)")
			server.Addr = ":443"
			server.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			server.ReadTimeout = 15 * time.Second
			server.WriteTimeout = 15 * time.Second
			server.IdleTimeout = 60 * time.Second
			if err := server.ListenAndServeTLS(cfg.TLSCertPath, cfg.TLSKeyPath); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		} else {
			log.Printf("Server starting on port %s (HTTP - Development Mode)", cfg.Port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Close queue connection
	if q != nil {
		q.Close()
	}

	// Shutdown server with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
