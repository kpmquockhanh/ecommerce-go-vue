---
title: 'Database Migration + Option Group Models + Data Migration'
type: 'feature'
created: '2026-09-06'
status: 'done'
route: 'dispatch'
baseline_commit: '3c57e9e7a9f263b363c6661f6a776360e836dde4'
review_loop_iteration: 0
context:
  - internal/database/database.go
  - internal/models/product.go
  - internal/repositories/product.go
  - internal/repositories/cart.go
  - internal/repositories/order.go
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The variant system hardcodes `size` and `color` as fixed columns on `product_variants`. This forces every product into a two-attribute model and prevents dynamic option groups, price modifiers, or customizations.

**Approach:** Create 4 new tables (`product_option_groups`, `product_option_values`, `variant_option_values`, `product_customization_fields`), add Go model structs, migrate existing size/color data into option groups, add `customization_data` JSONB columns to cart/order items, and drop the old size/color columns.

## Boundaries & Constraints

**Always:** Migrations use the existing `DO $$ ... EXCEPTION WHEN duplicate_column` pattern for idempotency. New tables use `CREATE TABLE IF NOT EXISTS`. Data migration must preserve all existing variant relationships.

**Never:** Do not create separate migration files. Do not add new API endpoints (stories 2–5). Do not change business logic in handlers beyond adapting to new model shapes.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Migration runs on clean DB | No existing data | 4 new tables created, no data migration needed | N/A — IF NOT EXISTS guards |
| Migration runs with existing variants | Product has variants with size/color | Option groups created, values linked, columns dropped | Rollback on any failure |
| Product with no variants | Product exists, no variant rows | No option groups created for that product | N/A |
| Product with duplicate sizes | Variants: S/Red, S/Blue | "Size" group has one "S" value, linked to both variants | Deduplicated on DISTINCT |
| Go build after changes | All model changes applied | `go build ./...` succeeds | Compilation errors must be fixed |

</frozen-after-approval>

## Code Map

- `internal/database/database.go:34-218` — `Migrate()` function, flat `[]string` of SQL, all migrations inline
- `internal/database/database.go:57-64` — `product_variants` CREATE TABLE (has size/color columns)
- `internal/database/database.go:65-74` — `cart_items` CREATE TABLE (no customization_data yet)
- `internal/database/database.go:93-102` — `order_items` CREATE TABLE (no customization_data yet)
- `internal/database/database.go:184-188` — ALTER TABLE product_variants add weight/dimensions
- `internal/models/product.go:10-31` — `Product` struct
- `internal/models/product.go:33-44` — `ProductVariant` struct (has Size/Color fields to remove)
- `internal/models/product.go:46-51` — `ProductWithVariants` (to be replaced by `ProductWithOptions`)
- `internal/models/product.go:100-120` — `CreateVariantRequest`/`UpdateVariantRequest` (Size/Color to replace)
- `internal/repositories/product.go:14-36` — `ProductRepository` interface (22 methods)
- `internal/repositories/product.go:430-448` — `GetVariants` (queries size/color columns)
- `internal/repositories/product.go:513-526` — `CreateVariant` (inserts size/color)
- `internal/repositories/product.go:528-601` — `UpdateVariant` (updates size/color)
- `internal/repositories/cart.go:10-21` — `CartRepository` interface
- `internal/repositories/cart.go:23-35` — `CartItemRow` struct (has VariantSize/VariantColor)
- `internal/repositories/order.go:13-27` — `OrderRepository` interface
- `internal/repositories/order.go:49-56` — `CheckoutCartItem` (has VariantLabel computed from Size/Color)
- `internal/handlers/product.go:336-451` — Variant CRUD handlers (CreateVariant, UpdateVariant, DeleteVariant)
- `internal/handlers/order.go:84-86` — Checkout builds variantLabel from Size+" / "+Color
- `internal/handlers/cart.go:81-88` — Cart display builds ProductVariant from CartItemRow
- `internal/handlers/product_test.go:143-178` — Mock variant methods (need updating)

## Tasks & Acceptance

**Execution:**
- [ ] `internal/database/database.go` — Add 4 new CREATE TABLE statements + data migration SQL + customization_data columns + drop size/color columns
- [ ] `internal/models/product.go` — Add ProductOptionGroup, ProductOptionValue, ProductCustomizationField, ProductWithOptions structs; update ProductVariant, CreateVariantRequest, UpdateVariantRequest
- [ ] `internal/repositories/product.go` — Update ProductRepository interface (remove old variant methods, update signatures); update all implementations (GetVariants, CreateVariant, UpdateVariant, DeleteVariant queries)
- [ ] `internal/repositories/cart.go` — Update CartItemRow to remove VariantSize/VariantColor, add CustomizationData; update Upsert/FindByUserID queries for customization_data
- [ ] `internal/repositories/order.go` — Update CheckoutCartItem and CheckoutParams for customization_data; update Checkout query to snapshot customization_data
- [ ] `internal/handlers/product.go` — Update variant handlers (CreateVariant, UpdateVariant) to use new DTOs; remove size/color references
- [ ] `internal/handlers/order.go` — Update checkout to compute variant label from option values; snapshot customization_data
- [ ] `internal/handlers/cart.go` — Update cart display to use computed variant label
- [ ] `internal/handlers/product_test.go` — Update mockProductRepo to match new interface; fix broken tests
- [ ] Verify `go build ./...` succeeds

**Acceptance Criteria:**
- Given the migration runs, when DB is checked, then 4 new tables exist and product_variants has no size/color columns
- Given existing variants with size/color, when migration completes, then option groups and values exist with correct links
- Given Go models compile, when `go build ./...` runs, then no errors
- Given ProductVariant struct, when inspected, then Size/Color fields are gone and Label/OptionValues are present

## Verification

**Commands:**
- `go build ./...` -- expected: no compilation errors
- `go vet ./...` -- expected: no issues

**Manual checks:**
- Verify database.go has all 4 new CREATE TABLE statements
- Verify ProductVariant no longer has Size/Color fields
- Verify ProductWithOptions replaces ProductWithVariants
