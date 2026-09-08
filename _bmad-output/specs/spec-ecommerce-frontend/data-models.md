# Data Models

PostgreSQL database schema for the ecommerce platform.

## Tables

### users
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL PRIMARY KEY | |
| email | VARCHAR(255) UNIQUE NOT NULL | |
| password_hash | VARCHAR(255) NOT NULL | bcrypt hash |
| first_name | VARCHAR(100) NOT NULL | |
| last_name | VARCHAR(100) NOT NULL | |
| role | VARCHAR(20) DEFAULT 'customer' | 'customer' or 'admin' |
| created_at | TIMESTAMP DEFAULT NOW() | |
| updated_at | TIMESTAMP DEFAULT NOW() | |

### products
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL PRIMARY KEY | |
| name | VARCHAR(255) NOT NULL | |
| slug | VARCHAR(255) UNIQUE NOT NULL | URL-friendly name |
| description | TEXT | |
| price | INTEGER NOT NULL | Price in cents |
| category | VARCHAR(100) | |
| images | TEXT[] | Array of image paths/URLs |
| active | BOOLEAN DEFAULT true | Soft delete flag |
| created_at | TIMESTAMP DEFAULT NOW() | |
| updated_at | TIMESTAMP DEFAULT NOW() | |

### product_variants
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL PRIMARY KEY | |
| product_id | INTEGER REFERENCES products(id) ON DELETE CASCADE | |
| size | VARCHAR(50) | |
| color | VARCHAR(50) | |
| stock | INTEGER DEFAULT 0 | |
| sku | VARCHAR(100) UNIQUE | |

### cart_items
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL PRIMARY KEY | |
| user_id | INTEGER REFERENCES users(id) ON DELETE CASCADE | NULL for guest carts |
| session_id | VARCHAR(255) | Guest cart identifier (UUID stored in localStorage) |
| product_id | INTEGER REFERENCES products(id) ON DELETE CASCADE | |
| variant_id | INTEGER REFERENCES product_variants(id) ON DELETE SET NULL | |
| quantity | INTEGER NOT NULL DEFAULT 1 | |
| created_at | TIMESTAMP DEFAULT NOW() | |
| UNIQUE(user_id, product_id, variant_id) | | One row per user-product-variant (NULLs ignored) |
| UNIQUE(session_id, product_id, variant_id) | | Guest cart constraint |

### orders
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL PRIMARY KEY | |
| user_id | INTEGER REFERENCES users(id) | |
| status | VARCHAR(30) DEFAULT 'pending' | pending, paid, shipped, delivered, cancelled |
| total | INTEGER NOT NULL | Total in cents |
| shipping_first_name | VARCHAR(100) | |
| shipping_last_name | VARCHAR(100) | |
| shipping_address_1 | VARCHAR(255) | |
| shipping_address_2 | VARCHAR(255) | |
| shipping_city | VARCHAR(100) | |
| shipping_state | VARCHAR(50) | |
| shipping_zip | VARCHAR(20) | |
| shipping_country | VARCHAR(10) | |
| stripe_payment_intent | VARCHAR(255) | Stripe PI ID |
| payment_status | VARCHAR(30) DEFAULT 'pending' | pending, succeeded, failed |
| created_at | TIMESTAMP DEFAULT NOW() | |
| updated_at | TIMESTAMP DEFAULT NOW() | |

### order_items
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL PRIMARY KEY | |
| order_id | INTEGER REFERENCES orders(id) ON DELETE CASCADE | |
| product_id | INTEGER REFERENCES products(id) | |
| variant_id | INTEGER REFERENCES product_variants(id) | |
| product_name | VARCHAR(255) | Denormalized snapshot |
| variant_label | VARCHAR(100) | e.g. "M / Black" |
| quantity | INTEGER NOT NULL | |
| price | INTEGER NOT NULL | Price at time of order (cents) |

### reviews
| Column | Type | Notes |
|--------|------|-------|
| id | SERIAL PRIMARY KEY | |
| product_id | INTEGER REFERENCES products(id) ON DELETE CASCADE | |
| user_id | INTEGER REFERENCES users(id) ON DELETE CASCADE | |
| rating | INTEGER NOT NULL | 1-5 |
| comment | TEXT | |
| created_at | TIMESTAMP DEFAULT NOW() | |
| UNIQUE(product_id, user_id) | | One review per user per product |

## Indexes
- `idx_products_category` on products(category)
- `idx_products_price` on products(price)
- `idx_products_slug` on products(slug)
- `idx_cart_items_user` on cart_items(user_id)
- `idx_orders_user` on orders(user_id)
- `idx_orders_status` on orders(status)
- `idx_order_items_order` on order_items(order_id)
- `idx_reviews_product` on reviews(product_id)

## Seed Data
The 3 existing hardcoded products seed the products table:
- Forever Pants ($260.00)
- Forever Shirt ($155.00)
- Forever Shorts ($300.00)
