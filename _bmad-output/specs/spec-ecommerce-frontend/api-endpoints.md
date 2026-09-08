# API Endpoints

Backend API reference for the ecommerce platform. All routes are prefixed with the base URL. Auth-required routes expect `Authorization: Bearer <jwt>` header.

## Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/auth/register` | No | Register new user |
| POST | `/api/auth/login` | No | Login, returns JWT |
| GET | `/api/auth/me` | Yes | Get current user profile |
| PUT | `/api/auth/me` | Yes | Update profile |

### POST `/api/auth/register`
**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepass123",
  "first_name": "John",
  "last_name": "Doe"
}
```
**Response (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { "id": 1, "email": "user@example.com", "role": "customer" }
}
```

### POST `/api/auth/login`
**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepass123"
}
```
**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { "id": 1, "email": "user@example.com", "role": "customer" }
}
```

## Products

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/products` | No | List products (paginated, filterable) |
| GET | `/api/products/:id` | No | Get product detail |
| POST | `/api/admin/products` | Admin | Create product |
| PUT | `/api/admin/products/:id` | Admin | Update product |
| DELETE | `/api/admin/products/:id` | Admin | Delete product |

### GET `/api/products`
**Query params:** `page`, `limit`, `category`, `min_price`, `max_price`, `sort` (price_asc, price_desc, newest, name), `search`

**Response (200):**
```json
{
  "products": [
    {
      "id": 1,
      "name": "Forever Pants",
      "slug": "forever-pants",
      "description": "Comfortable everyday pants",
      "price": 26000,
      "category": "pants",
      "images": ["/images/pants-1.jpg"],
      "variants": [
        { "id": 1, "size": "M", "color": "Black", "stock": 10 },
        { "id": 2, "size": "L", "color": "Black", "stock": 5 }
      ]
    }
  ],
  "pagination": { "page": 1, "limit": 20, "total": 45, "total_pages": 3 }
}
```

### GET `/api/products/:id`
**Response (200):**
```json
{
  "id": 1,
  "name": "Forever Pants",
  "slug": "forever-pants",
  "description": "Comfortable everyday pants made from sustainable materials.",
  "price": 26000,
  "category": "pants",
  "images": ["/images/pants-1.jpg", "/images/pants-2.jpg"],
  "variants": [
    { "id": 1, "size": "M", "color": "Black", "stock": 10 },
    { "id": 2, "size": "L", "color": "Black", "stock": 5 }
  ],
  "reviews": [
    { "id": 1, "user": "Jane D.", "rating": 5, "comment": "Great pants!", "created_at": "2026-01-15" }
  ],
  "avg_rating": 4.5,
  "review_count": 12
}
```

## Cart

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/cart` | Optional | Get cart (user or guest via session_id header) |
| POST | `/api/cart/items` | Optional | Add item to cart |
| PUT | `/api/cart/items/:id` | Optional | Update cart item quantity |
| DELETE | `/api/cart/items/:id` | Optional | Remove cart item |

Guest carts pass `X-Session-ID` header (UUID). Logged-in users use JWT. On login, frontend calls `POST /api/cart/merge` to combine guest cart.

### POST `/api/cart/items`
**Request:**
```json
{
  "product_id": 1,
  "variant_id": 1,
  "quantity": 2
}
```
**Response (201):**
```json
{
  "id": 1,
  "items": [
    { "id": 1, "product": { "id": 1, "name": "Forever Pants", "price": 26000 }, "variant": { "id": 1, "size": "M" }, "quantity": 2, "subtotal": 52000 }
  ],
  "total": 52000
}
```

## Checkout & Orders

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/checkout` | Yes | Create order from cart + process payment |
| GET | `/api/orders` | Yes | List user's orders |
| GET | `/api/orders/:id` | Yes | Get order detail |

### POST `/api/checkout`
**Request:**
```json
{
  "shipping_address": {
    "first_name": "John",
    "last_name": "Doe",
    "address_1": "123 Main St",
    "address_2": "",
    "city": "New York",
    "state": "NY",
    "zip": "10001",
    "country": "US"
  }
}
```
**Response (200):**
```json
{
  "clientSecret": "pi_xxx_secret_xxx",
  "order_id": 42
}
```

### GET `/api/orders`
**Response (200):**
```json
{
  "orders": [
    {
      "id": 42,
      "status": "paid",
      "total": 52000,
      "created_at": "2026-01-15T10:30:00Z",
      "items_count": 2
    }
  ]
}
```

### GET `/api/orders/:id`
**Response (200):**
```json
{
  "id": 42,
  "status": "paid",
  "total": 52000,
  "shipping_address": {
    "first_name": "John",
    "last_name": "Doe",
    "address_1": "123 Main St",
    "city": "New York",
    "state": "NY",
    "zip": "10001",
    "country": "US"
  },
  "items": [
    { "product_name": "Forever Pants", "variant": "M / Black", "quantity": 2, "price": 26000, "subtotal": 52000 }
  ],
  "payment_status": "succeeded",
  "stripe_payment_intent": "pi_xxx",
  "created_at": "2026-01-15T10:30:00Z"
}
```

## Reviews

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/products/:id/reviews` | Yes | Add review to product |
| GET | `/api/products/:id/reviews` | No | List reviews for product |

### POST `/api/products/:id/reviews`
**Request:**
```json
{
  "rating": 5,
  "comment": "Excellent product, very comfortable!"
}
```

## Admin

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/admin/orders` | Admin | List all orders |
| PUT | `/api/admin/orders/:id` | Admin | Update order status |
| GET | `/api/admin/users` | Admin | List all users |
| GET | `/api/admin/users/:id` | Admin | Get user detail |

### PUT `/api/admin/orders/:id`
**Request:**
```json
{
  "status": "shipped"
}
```
