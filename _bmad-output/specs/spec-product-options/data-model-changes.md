# Data Model Changes — Product Options & Dynamic Variants

## New Tables

### `product_option_groups`

```sql
CREATE TABLE product_option_groups (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(product_id, name)
);
CREATE INDEX idx_option_groups_product ON product_option_groups(product_id);
```

### `product_option_values`

```sql
CREATE TABLE product_option_values (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES product_option_groups(id) ON DELETE CASCADE,
    label VARCHAR(50) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    price_adjustment INTEGER DEFAULT 0,
    UNIQUE(group_id, label)
);
CREATE INDEX idx_option_values_group ON product_option_values(group_id);
```

### `variant_option_values` (junction)

```sql
CREATE TABLE variant_option_values (
    variant_id INTEGER NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    option_value_id INTEGER NOT NULL REFERENCES product_option_values(id) ON DELETE CASCADE,
    PRIMARY KEY (variant_id, option_value_id)
);
CREATE INDEX idx_variant_ov_variant ON variant_option_values(variant_id);
CREATE INDEX idx_variant_ov_value ON variant_option_values(option_value_id);
```

### `product_customization_fields`

```sql
CREATE TABLE product_customization_fields (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    label VARCHAR(100) NOT NULL,
    input_type VARCHAR(20) NOT NULL CHECK (input_type IN ('text', 'textarea', 'file')),
    max_length INTEGER,
    price_surge INTEGER DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_required BOOLEAN DEFAULT false
);
CREATE INDEX idx_customization_fields_product ON product_customization_fields(product_id);
```

## Altered Tables

### `product_variants` — drop size/color columns

```sql
ALTER TABLE product_variants DROP COLUMN IF EXISTS size;
ALTER TABLE product_variants DROP COLUMN IF EXISTS color;
```

**Note:** After migration, variants are defined purely by their linked option values via the `variant_option_values` junction table. The `sku`, `stock`, and dimension columns remain.

### `cart_items` — add customization_data

```sql
ALTER TABLE cart_items ADD COLUMN customization_data JSONB DEFAULT NULL;
```

**Note:** Stores customer-provided customization values as `{"field_id": 1, "value": "Happy Birthday"}`.

### `order_items` — add customization_data

```sql
ALTER TABLE order_items ADD COLUMN customization_data JSONB DEFAULT NULL;
```

**Note:** Snapshot of customization values at checkout time.

## Model Changes (Go structs)

### New Models

```go
type ProductOptionGroup struct {
    ID        int                  `json:"id"`
    ProductID int                  `json:"product_id"`
    Name      string               `json:"name"`
    SortOrder int                  `json:"sort_order"`
    Values    []ProductOptionValue `json:"values"`
}

type ProductOptionValue struct {
    ID               int    `json:"id"`
    GroupID          int    `json:"group_id"`
    Label            string `json:"label"`
    SortOrder        int    `json:"sort_order"`
    PriceAdjustment  int    `json:"price_adjustment"`
}

type ProductCustomizationField struct {
    ID          int    `json:"id"`
    ProductID   int    `json:"product_id"`
    Label       string `json:"label"`
    InputType   string `json:"input_type"`
    MaxLength   *int   `json:"max_length,omitempty"`
    PriceSurge  int    `json:"price_surge"`
    SortOrder   int    `json:"sort_order"`
    IsRequired  bool   `json:"is_required"`
}

type ProductWithOptions struct {
    Product
    OptionGroups    []ProductOptionGroup        `json:"option_groups"`
    Variants        []ProductVariant             `json:"variants"`
    Customizations  []ProductCustomizationField  `json:"customizations"`
    AvgRating       float64                      `json:"avg_rating"`
    ReviewCount     int                          `json:"review_count"`
}
```

### Updated Models

```go
// ProductVariant — size/color removed, linked via variant_option_values
type ProductVariant struct {
    ID        int    `json:"id"`
    ProductID int    `json:"product_id"`
    Stock     int    `json:"stock"`
    SKU       string `json:"sku"`
    Weight    *int   `json:"weight,omitempty"`
    Length    *int   `json:"length,omitempty"`
    Width     *int   `json:"width,omitempty"`
    Height    *int   `json:"height,omitempty"`
    Label     string `json:"label"`      // computed: "S / Red"
    OptionValues []int `json:"option_values"` // IDs of selected option values
}

// CreateVariantRequest — dynamic option value selection
type CreateVariantRequest struct {
    OptionValueIDs []int  `json:"option_value_ids"` // exactly one per option group
    Stock          int    `json:"stock"`
    SKU            string `json:"sku"`
    Weight         *int   `json:"weight,omitempty"`
    Length         *int   `json:"length,omitempty"`
    Width          *int   `json:"width,omitempty"`
    Height         *int   `json:"height,omitempty"`
}

// UpdateVariantRequest — can change stock, sku, dimensions, or reassign option values
type UpdateVariantRequest struct {
    OptionValueIDs *[]int  `json:"option_value_ids,omitempty"`
    Stock          *int    `json:"stock"`
    SKU            *string `json:"sku"`
    Weight         *int    `json:"weight,omitempty"`
    Length         *int    `json:"length,omitempty"`
    Width          *int    `json:"width,omitempty"`
    Height         *int    `json:"height,omitempty"`
}

// CreateOptionGroupRequest — admin creates an option group on a product
type CreateOptionGroupRequest struct {
    Name   string                `json:"name"`
    Values []CreateOptionValueRequest `json:"values"`
}

type CreateOptionValueRequest struct {
    Label           string `json:"label"`
    PriceAdjustment *int   `json:"price_adjustment,omitempty"`
}

// UpdateOptionGroupRequest
type UpdateOptionGroupRequest struct {
    Name      *string `json:"name,omitempty"`
    SortOrder *int    `json:"sort_order,omitempty"`
}

// UpdateOptionValueRequest
type UpdateOptionValueRequest struct {
    Label           *string `json:"label,omitempty"`
    PriceAdjustment *int    `json:"price_adjustment,omitempty"`
    SortOrder       *int    `json:"sort_order,omitempty"`
}

// CreateCustomizationFieldRequest — admin defines a customization field
type CreateCustomizationFieldRequest struct {
    Label      string `json:"label"`
    InputType  string `json:"input_type"`
    MaxLength  *int   `json:"max_length,omitempty"`
    PriceSurge *int   `json:"price_surge,omitempty"`
    SortOrder  *int   `json:"sort_order,omitempty"`
    IsRequired *bool  `json:"is_required,omitempty"`
}

// AddToCartRequest — variant selection + customization values
type AddToCartRequest struct {
    ProductID         int                        `json:"product_id"`
    VariantID         *int                       `json:"variant_id"`
    Quantity          int                        `json:"quantity"`
    CustomizationData map[int]string             `json:"customization_data,omitempty"` // field_id -> value
}
```

## API Changes

### New Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/products/{id}/options` | Public | Get option groups and values for a product |
| POST | `/api/admin/products/{id}/option-groups` | Admin | Create option group with values |
| PUT | `/api/admin/products/{id}/option-groups/{group_id}` | Admin | Update option group (name, sort order) |
| DELETE | `/api/admin/products/{id}/option-groups/{group_id}` | Admin | Delete option group and its values |
| PUT | `/api/admin/products/{id}/option-groups/{group_id}/reorder` | Admin | Reorder option values within a group |
| POST | `/api/admin/products/{id}/option-groups/{group_id}/values` | Admin | Add option value to a group |
| PUT | `/api/admin/products/{id}/option-groups/{group_id}/values/{value_id}` | Admin | Update option value |
| DELETE | `/api/admin/products/{id}/option-groups/{group_id}/values/{value_id}` | Admin | Delete option value |
| POST | `/api/admin/products/{id}/customizations` | Admin | Create customization field |
| PUT | `/api/admin/products/{id}/customizations/{field_id}` | Admin | Update customization field |
| DELETE | `/api/admin/products/{id}/customizations/{field_id}` | Admin | Delete customization field |
| PUT | `/api/admin/products/{id}/customizations/reorder` | Admin | Reorder customization fields |

### Modified Endpoints

| Endpoint | Change |
|----------|--------|
| POST `/api/admin/products/{id}/variants` | Accept `option_value_ids[]` instead of `size`/`color` |
| PUT `/api/admin/products/{id}/variants/{variant_id}` | Accept `option_value_ids[]` instead of `size`/`color` |
| GET `/api/products/{slug}` | Response uses `ProductWithOptions` — includes `option_groups`, `customizations`, computed variant `label` |
| POST `/api/cart/items` | Accept optional `customization_data` map |
| POST `/api/checkout/init` | Snapshot customization data into order items |

### Removed

| Field | Change |
|-------|--------|
| `product_variants.size` | Replaced by option group linkage |
| `product_variants.color` | Replaced by option group linkage |

## Migration Strategy

1. Create new tables (`product_option_groups`, `product_option_values`, `variant_option_values`, `product_customization_fields`).
2. Migrate existing data: for each product that has variants, create "Size" and "Color" option groups, insert distinct size/color values, create option_value rows, and link existing variants via `variant_option_values`.
3. Add `customization_data` column to `cart_items` and `order_items`.
4. Drop `size` and `color` columns from `product_variants`.
5. Update all repository queries and handlers to use new models.
