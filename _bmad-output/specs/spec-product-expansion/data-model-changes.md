# Data Model Changes — Product Expansion

## New Tables

### `product_categories` (many-to-many)

```sql
CREATE TABLE product_categories (
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);
CREATE INDEX idx_product_categories_category ON product_categories(category_id);
```

**Migration:** Drop `products.category` column. Insert rows from old category string into `product_categories` join table by matching `categories.name`.

### `product_variants` (altered — add weight/dimensions)

```sql
ALTER TABLE product_variants ADD COLUMN weight INTEGER;   -- grams
ALTER TABLE product_variants ADD COLUMN length INTEGER;   -- cm
ALTER TABLE product_variants ADD COLUMN width INTEGER;    -- cm
ALTER TABLE product_variants ADD COLUMN height INTEGER;   -- cm
```

**Note:** Variant weight/dimensions override product-level values when present (used for shipping calc).

### `product_tags`

```sql
CREATE TABLE product_tags (
    id SERIAL PRIMARY KEY,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL,
    UNIQUE(product_id, tag)
);
CREATE INDEX idx_product_tags_tag ON product_tags(tag);
CREATE INDEX idx_product_tags_product ON product_tags(product_id);
```

**Constraint:** Max 12 tags per product, enforced at handler validation.

### `product_image_order`

```sql
CREATE TABLE product_image_order (
    id SERIAL PRIMARY KEY,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    image_path TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_primary BOOLEAN DEFAULT false,
    UNIQUE(product_id, image_path)
);
CREATE INDEX idx_product_image_order_product ON product_image_order(product_id);
```

## Altered Tables

### `products`

```sql
-- Drop old columns
ALTER TABLE products DROP COLUMN IF EXISTS category;
ALTER TABLE products DROP COLUMN IF EXISTS active;

-- Add new columns
ALTER TABLE products ADD COLUMN compare_at_price INTEGER;
ALTER TABLE products ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'published';
ALTER TABLE products ADD COLUMN weight INTEGER;           -- grams
ALTER TABLE products ADD COLUMN length INTEGER;           -- cm
ALTER TABLE products ADD COLUMN width INTEGER;            -- cm
ALTER TABLE products ADD COLUMN height INTEGER;           -- cm
ALTER TABLE products ADD COLUMN meta_title VARCHAR(160);
ALTER TABLE products ADD COLUMN meta_description VARCHAR(500);
ALTER TABLE products ADD COLUMN og_image TEXT;

-- Add index on status
CREATE INDEX idx_products_status ON products(status);
```

**Note:** Clean migration. Old `category` and `active` columns dropped. New columns added fresh.

## Model Changes (Go structs)

### Product (updated)

```go
type Product struct {
    ID             int        `json:"id"`
    Name           string     `json:"name"`
    Slug           string     `json:"slug"`
    Description    string     `json:"description"`
    Price          int        `json:"price"`
    CompareAtPrice *int       `json:"compare_at_price,omitempty"`
    Status         string     `json:"status"`           // draft, published, archived
    Categories     []Category `json:"categories"`
    Tags           []string   `json:"tags"`
    Images         []string   `json:"images"`
    PrimaryImage   string     `json:"primary_image"`
    Weight         *int       `json:"weight,omitempty"`  // grams
    Length         *int       `json:"length,omitempty"`  // cm
    Width          *int       `json:"width,omitempty"`   // cm
    Height         *int       `json:"height,omitempty"`  // cm
    MetaTitle      *string    `json:"meta_title,omitempty"`
    MetaDescription *string   `json:"meta_description,omitempty"`
    OgImage        *string    `json:"og_image,omitempty"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}
```

### ProductFilters (updated)

```go
type ProductFilters struct {
    Page       int
    Limit      int
    Categories []string   // changed from single Category string
    Tags       []string   // new
    MinPrice   int
    MaxPrice   int
    Search     string
    Sort       string
    Status     string     // new: filter by status (admin only)
}
```

## API Changes

### New Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/categories` | Public | List categories with product counts |
| POST | `/api/admin/products/{id}/variants` | Admin | Create variant |
| PUT | `/api/admin/products/{id}/variants/{variant_id}` | Admin | Update variant |
| DELETE | `/api/admin/products/{id}/variants/{variant_id}` | Admin | Delete variant |
| POST | `/api/admin/products/{id}/tags` | Admin | Add tags to product |
| DELETE | `/api/admin/products/{id}/tags/{tag}` | Admin | Remove tag from product |
| PUT | `/api/admin/products/{id}/images/order` | Admin | Reorder images / set primary |

### Modified Endpoints

| Endpoint | Change |
|----------|--------|
| GET `/api/products` | Accept `categories` (comma-separated) and `tags` (comma-separated) query params |
| POST `/api/admin/products` | Accept `categories[]`, `tags[]`, `compare_at_price`, `status`, `weight`, `length`, `width`, `height`, `meta_title`, `meta_description`, `og_image` |
| PUT `/api/admin/products/{id}` | Accept same new fields |
| GET `/api/products/{slug}` | Return new fields (categories, tags, compare_at_price, status, dimensions, SEO) |

### Removed

| Field | Change |
|-------|--------|
| `products.category` | Replaced by `product_categories` join table |
| `products.active` | Replaced by `products.status` (column dropped) |
