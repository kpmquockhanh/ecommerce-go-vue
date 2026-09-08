# 🛒 E-Commerce Go + Vue

[![Go Version](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/doc/go1.26)
[![Vue](https://img.shields.io/badge/Vue-3.5-42b883.svg)](https://vuejs.org/)
[![Stripe](https://img.shields.io/badge/Stripe-v82.5.1-purple.svg)](https://stripe.com/)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://docker.com/)

A full-stack e-commerce application: a Go REST API backend and a Vue 3 storefront + admin dashboard frontend. Supports product catalog with variants/options, cart, checkout via Stripe, order management, reviews, image uploads, and background email/queue processing.

## ✨ Features

- 🔐 **Auth** — JWT-based register/login, role-based admin access
- 🛍️ **Catalog** — products, categories, variants, option groups, customization fields
- 🛒 **Cart & Checkout** — guest cart with merge-on-login, Stripe payment intents, idempotent checkout
- 📦 **Orders** — order history, admin order management, Stripe webhook-driven status updates
- ⭐ **Reviews** — product ratings and reviews
- 🖼️ **Image Storage** — product image upload/serving via S3-compatible storage (MinIO)
- 📬 **Background Jobs** — RabbitMQ-backed queue and worker for async email (order confirmations) with dead-letter handling
- 🛠️ **Admin Dashboard** — Vue admin UI for products, orders, users, categories, and dead letters
- 🐳 **Docker Compose** — one command spins up Postgres, MinIO, RabbitMQ, pgAdmin, backend, and frontend

## 🏗️ Stack

| Layer    | Tech                                                             |
| -------- | ----------------------------------------------------------------- |
| Backend  | Go 1.26, `net/http`, [pgx](https://github.com/jackc/pgx) (Postgres), [stripe-go](https://github.com/stripe/stripe-go), MinIO client, RabbitMQ (amqp091) |
| Frontend | Vue 3, Vite, Pinia, Vue Router, Ant Design Vue, Axios              |
| Infra    | PostgreSQL, MinIO (S3-compatible), RabbitMQ, pgAdmin                |

## 🚀 Quick Start

### Prerequisites

- Go 1.26+
- Node.js 18+ (for the frontend)
- Docker & Docker Compose
- A Stripe account (test keys are fine)

### 🐳 Run everything with Docker Compose (recommended)

```bash
cp .env.example .env
# edit .env and set STRIPE_SECRET_KEY / STRIPE_PUBLISHABLE_KEY

docker-compose up --build
```

This starts:

| Service    | URL                     |
| ---------- | ----------------------- |
| Backend API | http://localhost:4242   |
| Frontend    | http://localhost:3000   |
| pgAdmin     | http://localhost:5050   |
| MinIO console | http://localhost:9001 |
| RabbitMQ management | http://localhost:15672 |

The backend auto-runs DB migrations and seeds sample data on startup.

### 💻 Run locally without Docker

1. **Start infra dependencies** (Postgres, MinIO, RabbitMQ) — easiest via Docker Compose, or point `DATABASE_URL` / `S3_*` / `RABBITMQ_URL` at your own instances.

2. **Backend**

   ```bash
   cp .env.example .env   # set STRIPE_SECRET_KEY and other values
   go mod download
   go run cmd/server/main.go
   # or: make run
   ```

   Server starts at http://localhost:4242 (see `Makefile` for `build`, `test`, `lint`, `docker-*` targets).

3. **Frontend**

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   Dev server starts at http://localhost:5173 (Vite default) and proxies API calls to the backend.

## ⚙️ Configuration

Environment variables read by the backend (see `internal/config/config.go`):

| Variable               | Description                          | Default                                                |
| ---------------------- | ------------------------------------- | ------------------------------------------------------- |
| `STRIPE_SECRET_KEY`     | Stripe secret API key (required)      | -                                                       |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret         | -                                                       |
| `STRIPE_PUBLISHABLE_KEY`| Stripe publishable key (exposed to frontend) | -                                                 |
| `DATABASE_URL`          | Postgres connection string            | `postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable` |
| `JWT_SECRET`            | Secret used to sign JWTs              | `your-secret-key-change-in-production`                  |
| `S3_ENDPOINT`           | MinIO/S3 endpoint                     | `localhost:9000`                                        |
| `S3_PUBLIC_ENDPOINT`    | Public-facing endpoint for image URLs | -                                                       |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | MinIO/S3 credentials        | `minioadmin` / `minioadmin`                             |
| `S3_BUCKET`             | Bucket for product images             | `ecommerce-images`                                      |
| `S3_USE_SSL` / `S3_PUBLIC_USE_SSL` | Use HTTPS for S3 endpoints  | `false`                                                 |
| `RABBITMQ_URL`          | RabbitMQ connection string            | `amqp://guest:guest@localhost:5672/`                    |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` | Outbound mail (order emails) | `localhost` / `587` / - / -                |
| `FROM_EMAIL` / `FROM_NAME` | Sender identity for outbound mail  | `noreply@eshop.com` / `eShop`                           |
| `PORT`                  | HTTP port                             | `4242`                                                  |
| `PRODUCTION`             | Enable production mode (HTTPS)        | `false`                                                 |
| `TLS_CERT_PATH` / `TLS_KEY_PATH` | TLS cert/key paths (production only) | `cert.pem` / `key.pem`                          |

> MinIO and RabbitMQ are optional at startup — if unavailable, image storage/background jobs are disabled with a warning rather than crashing the server.

## 📡 API Overview

All routes are prefixed `/api` unless noted. Full route wiring lives in `cmd/server/main.go`.

| Area       | Routes                                                                 |
| ---------- | ------------------------------------------------------------------------ |
| Auth       | `POST /api/auth/register`, `POST /api/auth/login`, `GET/POST /api/auth/me` |
| Products   | `GET /api/products`, `GET /api/products/{slug}` (+ options/customizations) |
| Categories | `GET /api/categories`                                                     |
| Cart       | `GET /api/cart`, `POST/PUT/DELETE /api/cart/items`, `POST /api/cart/merge` |
| Checkout   | `POST /api/checkout/init`, `POST /api/checkout/confirm`                   |
| Orders     | `GET /api/orders`, `GET /api/orders/{id}`                                 |
| Reviews    | `GET /api/reviews/product/{id}`, `POST /api/reviews`                      |
| Images     | `POST /api/images/upload`, `GET /api/images/product/{id}`                 |
| Admin      | `/api/admin/products`, `/api/admin/orders`, `/api/admin/users`, `/api/admin/categories`, `/api/admin/dead-letters`, `/api/admin/stats` |
| Payments   | `POST /create-payment-intent`, `POST /webhooks/stripe`                    |
| Misc       | `GET /health`, `GET /api/config` (publishable key for frontend)           |

Admin routes require a JWT with an admin role; cart/checkout support both guest and authenticated flows.

## 🏗️ Project Structure

```
.
├── cmd/server/              # Application entrypoint
├── internal/
│   ├── checkout/            # Checkout session cleanup job
│   ├── config/              # Environment configuration
│   ├── database/            # Postgres connection, migrations, seeding
│   ├── handlers/            # HTTP handlers (auth, products, cart, orders, admin, ...)
│   ├── mail/                # SMTP mailer
│   ├── middleware/          # Auth (JWT), CORS, logging, rate limiting
│   ├── models/              # Data structures
│   ├── queue/                # RabbitMQ queue client
│   ├── repositories/         # Postgres data access layer
│   ├── storage/              # S3/MinIO image storage
│   ├── validation/           # Request validation
│   └── worker/                # Background job worker (queue consumer)
├── frontend/                 # Vue 3 storefront + admin dashboard (Vite, Pinia, Ant Design Vue)
├── docker-compose.yml         # Postgres, MinIO, RabbitMQ, pgAdmin, backend, frontend
├── Dockerfile / Dockerfile.dev
└── Makefile                   # run, build, test, lint, docker-* targets
```

## 🧪 Testing

```bash
go test -v ./...
# or: make test
```

## 🚀 Production Deployment

```bash
docker build -t ecommerce-api-production .
docker run -d \
  -p 443:443 \
  -e PRODUCTION=true \
  -e STRIPE_SECRET_KEY=sk_live_xxxxx \
  -e DATABASE_URL=... \
  --name ecommerce-api \
  ecommerce-api-production
```

- Use live Stripe keys and configure `STRIPE_WEBHOOK_SECRET` for webhook signature verification
- Set a strong `JWT_SECRET`
- Provide TLS certs via `TLS_CERT_PATH` / `TLS_KEY_PATH` when `PRODUCTION=true`
- Point `DATABASE_URL`, `S3_*`, and `RABBITMQ_URL` at production infrastructure

## 📚 Resources

- [Stripe Go SDK](https://github.com/stripe/stripe-go)
- [Vue 3 Documentation](https://vuejs.org/)
- [Go Documentation](https://golang.org/doc/)
