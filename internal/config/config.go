package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripePublishableKey string
	DatabaseURL          string
	JWTSecret            string
	S3Endpoint           string
	S3PublicEndpoint     string
	S3AccessKey          string
	S3SecretKey          string
	S3Bucket             string
	S3UseSSL             bool
	S3PublicUseSSL       bool
	Port                 string
	IsProduction         bool
	TLSCertPath          string
	TLSKeyPath           string

	// RabbitMQ
	RabbitMQURL string

	// Mail
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if exists (for development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	smtpPort, _ := strconv.Atoi(getEnvOrDefault("SMTP_PORT", "587"))

	cfg := &Config{
		StripeSecretKey:      os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:  os.Getenv("STRIPE_WEBHOOK_SECRET"),
		StripePublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		DatabaseURL:          getEnvOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable"),
		JWTSecret:            getEnvOrDefault("JWT_SECRET", "your-secret-key-change-in-production"),
		S3Endpoint:           getEnvOrDefault("S3_ENDPOINT", "localhost:9000"),
		S3PublicEndpoint:     os.Getenv("S3_PUBLIC_ENDPOINT"),
		S3AccessKey:          getEnvOrDefault("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:          getEnvOrDefault("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:             getEnvOrDefault("S3_BUCKET", "ecommerce-images"),
		S3UseSSL:             os.Getenv("S3_USE_SSL") == "true",
		S3PublicUseSSL:       os.Getenv("S3_PUBLIC_USE_SSL") == "true",
		Port:                 getEnvOrDefault("PORT", "4242"),
		IsProduction:         os.Getenv("PRODUCTION") == "true",
		TLSCertPath:          getEnvOrDefault("TLS_CERT_PATH", "cert.pem"),
		TLSKeyPath:           getEnvOrDefault("TLS_KEY_PATH", "key.pem"),

		// RabbitMQ
		RabbitMQURL: getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),

		// Mail
		SMTPHost:     getEnvOrDefault("SMTP_HOST", "localhost"),
		SMTPPort:     smtpPort,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    getEnvOrDefault("FROM_EMAIL", "noreply@eshop.com"),
		FromName:     getEnvOrDefault("FROM_NAME", "eShop"),
	}

	// Validate required configuration
	if cfg.StripeSecretKey == "" {
		log.Fatal("You need to set STRIPE_SECRET_KEY environment variable")
	}

	return cfg
}

// getEnvOrDefault returns environment variable value or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
