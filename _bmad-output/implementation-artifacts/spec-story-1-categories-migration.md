---
title: 'Database migration + multiple categories + public category endpoint'
type: 'feature'
created: '2026-09-05'
status: 'done'
route: 'dispatch'
baseline_commit: '14014991dc9fa18091feed3ad7c705b68110d987'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Products have a single string-based `category` column that cannot represent products in multiple categories. The `active` boolean is being replaced by a `status` field. No public category listing exists for storefront navigation.

**Approach:** Fresh schema migration — drop `products.category` and `products.active`, create `product_categories` junction table for many-to-many, add `status` column. Update all models, repositories, handlers, and routes to use the new schema. Add public `GET /api/categories` endpoint.

## Boundaries & Constraints

**Always:**
- Fresh migration: drop old columns, create new tables clean. No backward compatibility.
- All existing product queries must filter by `status = 'published'` instead of `active = true`.
- Category assignment uses category IDs (not names) in API requests.
- Public category endpoint returns only categories with at least one published product.

**Never:**
- Do not keep the `active` column or `category` VARCHAR column.
- Do not create a service layer — maintain existing handler-to-repo pattern.
- Do not modify cart, order, checkout, or payment logic in this story.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Create product with categories | `CreateProductRequest{CategoryIDs: [1,2,3]}` | Product created, linked to 3 categories in `product_categories` | 400 if any category ID doesn't exist |
| Update product categories | `UpdateProductRequest{CategoryIDs: &[1,4]}` | Old links replaced, product now in categories 1 and 4 | 400 if any category ID doesn't exist |
| List products filter by category | `GET /api/products?categories=pants,shirts` | Products in either pants or shirts returned | Empty array if no matches |
| List products (status filter) | `GET /api/products` (public) | Only `status='published'` products returned | N/A |
| Admin list products | `GET /api/admin/products?status=draft` | Products with that status returned | N/A |
| Public category listing | `GET /api/categories` | Categories with product counts, only those with >=1 published product | N/A |
| Soft delete → status | `DELETE /api/admin/products/1` | Product `status` set to `archived` | N/A |
| Category not found | Create product with non-existent category ID | 400 Bad Request | `"category not found"` |

</frozen-after-approval>

## Code Map

- `internal/database/database.go` -- Migration SQL: create junction tables, ALTER products, drop columns, update seed data
- `internal/models/product.go` -- Product struct: remove Category/Active, add Categories/Status; update Create/Update request DTOs
- `internal/models/category.go` -- Category struct unchanged
- `internal/repositories/product.go` -- ProductRepository interface + impl: update all SQL, add category-linking methods, update ProductFilters
- `internal/repositories/category.go` -- CategoryRepository: update List/CountProducts/RemoveCategoryFromProducts to use junction table
- `internal/handlers/product.go` -- ProductHandler: update create/update/list to handle CategoryIDs, status filtering
- `internal/handlers/category.go` -- CategoryHandler: add ListPublicCategories for storefront
- `cmd/server/main.go` -- Add `GET /api/categories` public route
- `internal/handlers/product_test.go` -- Update mockProductRepo and test data for new Product struct
- `internal/repositories/review.go:79` -- `active = true` reference must change to `status = 'published'`

## Tasks & Acceptance

**Execution:**
- [x] `internal/database/database.go` -- Rewrite migration: create `product_categories`, `product_tags`, `product_image_order` tables; ALTER products to drop `category`/`active`, add `status`; drop `idx_products_category`; update seed data to use new schema
- [x] `internal/models/product.go` -- Update Product struct (remove Category/Active, add Categories []Category, Status string); update CreateProductRequest (CategoryIDs []int), UpdateProductRequest (CategoryIDs *[]int, Status *string, remove Active)
- [x] `internal/repositories/product.go` -- Update all SQL queries to use `status` instead of `active`; remove `category` column refs; add `LinkCategories`/`GetCategories` methods; update `List` to JOIN `product_categories` and filter by multiple categories
- [x] `internal/repositories/category.go` -- Update `List` to JOIN `product_categories` and count published products; update `CountProducts` and `RemoveCategoryFromProducts` to use junction table
- [x] `internal/handlers/product.go` -- Update CreateProduct/UpdateProduct to accept CategoryIDs and insert into junction table; update ListProducts to accept `categories` comma-separated filter; update status handling
- [x] `internal/handlers/category.go` -- Add `ListPublicCategories` method returning categories with product counts (published only)
- [x] `cmd/server/main.go` -- Register `GET /api/categories` public route
- [x] `internal/handlers/product_test.go` -- Update mockProductRepo struct and all test cases for new Product/Request structs
- [x] `internal/repositories/review.go:79` -- Change `active = true` to `status = 'published'`

**Acceptance Criteria:**
- Given a product request with `category_ids: [1,2]`, when created, then `product_categories` has 2 rows linking the product to categories 1 and 2
- Given a product in categories 1 and 2, when `GET /api/products?categories=1`, then the product is returned
- Given a product in categories 1 and 2, when `GET /api/products?categories=1,3`, then the product is returned (OR logic)
- Given products exist in "pants" category, when `GET /api/categories`, then "pants" appears with accurate product count
- Given a category has 0 published products, when `GET /api/categories`, then that category is NOT returned
- Given a product with `active=true`, after migration, when queried, then `status = 'published'`
- Given `GET /api/products` (public), when queried, then only `status='published'` products returned
- Given `DELETE /api/admin/products/1`, when executed, then product `status` set to `archived`
- Given create product with non-existent category ID, when submitted, then 400 returned with error message

## Implementation Notes

<!-- Append-only during implementation. Leave empty at planning time. -->

## Spec Change Log

<!-- Append-only. Empty until first review loopback. -->

## Review Triage Log

| Finding | Source | Verdict | Evidence |
|---------|--------|---------|----------|
| CORS Allow-Methods missing PUT/DELETE/PATCH | blind-hunter | defer | Pre-existing middleware, not changed by this story |
| Production TLS port overwrite | blind-hunter | defer | Pre-existing config, not changed by this story |
| Route ambiguity /api/admin/products/ | blind-hunter | defer | Pre-existing routing, not changed by this story |
| database.Seed runs unconditionally | blind-hunter | defer | Pre-existing startup, not changed by this story |
| Health check nil dereference | blind-hunter, edge-case | defer | Pre-existing handler, not changed by this story |
| JWTSecret global mutable | blind-hunter | defer | Pre-existing pattern, not changed by this story |
| No config validation at startup | blind-hunter | defer | Pre-existing config, not changed by this story |
| Dockerfile.dev reference | blind-hunter | defer | Pre-existing docker config, not changed by this story |
| Missing Access-Control-Allow-Credentials | blind-hunter | defer | Pre-existing CORS, not changed by this story |
| Queue close before shutdown | blind-hunter, edge-case | defer | Pre-existing shutdown, not changed by this story |
| Nil queue passed to orderHandler | edge-case | defer | Pre-existing init, not changed by this story |
| SMTP port parse error discarded | edge-case | defer | Pre-existing config, not changed by this story |
| Payment error response format | verification-gap | defer | Pre-existing handler, not changed by this story |
| Health endpoint untested | verification-gap | defer | Pre-existing handler, not changed by this story |

## Design Notes

The junction table approach (`product_categories`) is standard for many-to-many relationships. Using integer category IDs in API requests (not names) ensures referential integrity. The `status` enum ('draft', 'published', 'archived') replaces the boolean `active` flag to support a proper product lifecycle.

## Verification

**Commands:**
- `go build ./...` -- expected: compiles without errors
- `go test ./internal/models/...` -- expected: PASS
- `go test ./internal/handlers/...` -- expected: PASS
- `go test ./...` -- expected: all tests pass

**Manual checks:**
- Start server, create product with multiple categories via admin API
- Verify `GET /api/categories` returns categories with correct counts
- Verify `GET /api/products?categories=1,2` filters correctly
- Verify public product list excludes draft/archived products
