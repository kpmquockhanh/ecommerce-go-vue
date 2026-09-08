---
id: SPEC-product-expansion
companions: [data-model-changes.md]
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Product Expansion — Multi-Category & Enhanced Product Features

## Why

The current product model is limited: single string-based category, no variant management API, no sale pricing, no SEO metadata, and no product status lifecycle. These gaps block common e-commerce requirements — selling a product in multiple categories, managing inventory variants through the API, running sales, and optimizing for search engines. This expansion addresses the user's request for multiple categories and adds high-value features that a production e-commerce system needs.

## Capabilities

- **CAP-1**
  - **intent:** Admin can assign multiple categories to a single product, and products can be filtered by one or more categories in storefront queries.
  - **success:** A product can have 0–N categories. Filtering by multiple category names returns products in any of those categories. Existing single-category data migrates to the new join table without data loss.

- **CAP-2**
  - **intent:** Admin can create, update, and delete product variants (size, color, stock, SKU) through API endpoints instead of direct database manipulation.
  - **success:** POST/PUT/DELETE endpoints exist for variants under admin routes. Creating a variant with duplicate SKU within a product returns an error. Deleting a variant that is referenced by an active cart item or order returns a constraint error.

- **CAP-3**
  - **intent:** Storefront visitors can list all active categories with product counts for navigation UI (sidebar, filters).
  - **success:** A public GET endpoint returns categories with accurate product counts, sorted by name. Only categories with at least one published product are returned.

- **CAP-4**
  - **intent:** Admin can assign multiple string tags to products (max 12), and storefront users can filter products by one or more tags.
  - **success:** Products support a tags array (max 12). Filter endpoint accepts comma-separated tag values. Tag suggestions are returned from existing tags matching a prefix.

- **CAP-5**
  - **intent:** Admin can set a compare_at_price on products to show original price vs sale price on storefront.
  - **success:** Product response includes `compare_at_price` (nullable). When set, storefront displays "was $X, now $Y". Filtering by price range uses the effective (sale) price.

- **CAP-6**
  - **intent:** Products have a status field (draft/published/archived) that controls visibility, fully replacing the active boolean.
  - **success:** Draft products are not visible on storefront. Published products are visible. Archived products are not visible but retained. The `active` boolean is dropped. Existing `active=true` migrates to `published`, `active=false` to `archived`.

- **CAP-7**
  - **intent:** Admin can record product and variant weight/dimensions (length, width, height) for shipping calculation, with variant values overriding product defaults.
  - **success:** Product and variant models store weight (grams), length, width, height (cm). All fields optional. When a variant has its own weight/dimensions, those override the product's values for shipping calculation.

- **CAP-8**
  - **intent:** Admin can set SEO metadata (meta_title, meta_description, og_image) on products for search engine optimization.
  - **success:** Product supports meta_title (max 160 chars), meta_description (max 500 chars), og_image (URL). Defaults to product name/description/first image when not set.

- **CAP-9**
  - **intent:** Admin can designate a primary image and set display order for product images.
  - **success:** Product images have an `is_primary` flag and `sort_order` integer. Only one image can be primary. Reordering updates sort_order atomically. Default primary is the first uploaded image.

## Constraints

- Fresh migration: drop `products.category` and `products.active` columns, create new tables clean. No gradual transition.
- The `product_variants` table already exists; variant CRUD endpoints must use it, not create a new table.
- Price fields remain integer (cents) — no decimal migration.
- Image storage continues using S3/MinIO with the existing TEXT[] array; ordering is stored in a separate `product_image_order` table.
- All new admin endpoints require the existing admin auth middleware.
- Public endpoints must not expose draft or archived products.
- Max 12 tags per product, enforced at handler validation.

## Non-goals

- Bulk product import/export (CSV/Excel) — deferred to future expansion.
- Bulk product operations (bulk create/update/delete) — deferred to future expansion.
- Product collections or bundles — separate feature.
- Inventory management at product level (only variant-level stock).
- Multi-currency pricing — single currency per store.
- Product relationships (related products, upsells) — separate feature.
- Digital product / download management — physical products only.

## Success signal

- Admin can assign 3 categories to a product and verify storefront filtering returns it under all 3.
- Admin creates a variant via API, sees it in product detail, updates stock, and deletes it without errors.
- Storefront category sidebar shows accurate product counts from the public endpoint.
- A product with `compare_at_price` set displays "was $X, now $Y" on the storefront.
- A draft product is invisible on the storefront; publishing it makes it visible immediately.
- Product detail page shows SEO metadata in `<meta>` tags when rendered server-side.

## Assumptions

- Existing single-category products will be migrated to the new `product_categories` join table during deployment.
- Tags are simple strings, not a hierarchical entity with slugs or descriptions.
- Variant CRUD is admin-only, consistent with existing product management endpoints.
- Weight/dimensions at both product and variant level; variant values override product defaults when present.

## Open Questions

<!-- All resolved. See memlog for decision history. -->
