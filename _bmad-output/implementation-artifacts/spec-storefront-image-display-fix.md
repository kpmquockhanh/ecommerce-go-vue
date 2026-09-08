---
title: 'Storefront product images resolve to raw S3 object keys instead of URLs'
type: 'bugfix'
created: '2026-09-08'
status: 'done'
baseline_commit: '1ce32d57a26338e12349900476dd7e49682a4954'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The public product endpoints (`GET /api/products`, `GET /api/products/{slug}`) return `images`/`primary_image` as raw S3 object keys (e.g. `products/abc.jpg`) instead of usable URLs. `internal/handlers/product.go`'s `ListProducts` and `GetProduct` copy `p.Images`/`p.PrimaryImage` straight from the repository, unlike `AddImageToProduct`/`UploadImage`/`GetProductImages` in `internal/handlers/image.go`, which all presign via `storage.GetPresignedURL`. The storefront (`frontend/src/components/ProductCard.vue`, `frontend/src/views/ProductDetail.vue`) binds `product.images[i]` directly as `<img :src>`, so it renders a broken/relative path instead of the image. Admin's `frontend/src/views/admin/ProductDetail.vue` doesn't hit this because it separately fetches `/api/images/product/{id}` for display URLs.

**Approach:** Presign `Images` and `PrimaryImage` on the response objects built by `ListProducts` and `GetProduct` (public routes only — `AdminListProducts`/`AdminGetProduct` keep raw keys, since admin already manages images by raw path for delete/order/primary operations). Mirror the existing guard/fallback pattern in `GetProductImages`: skip presigning when `storage.Client == nil`; on a per-image presign error, keep the raw key so array length/order (used for index-based image switching in `ProductDetail.vue`) isn't disturbed.

</frozen-after-approval>

## Implementation Notes

- Added `presignProductImages(ctx, images, primary)` in `internal/handlers/product.go`: no-op when `storage.Client == nil`; per-image presign errors keep the raw key (logged) instead of dropping the entry, preserving array order/length.
- Wired into `ListProducts` (mutates the `models.Product` value copy from `h.productRepo.List` before building each `ProductListItem`) and `GetProduct` (mutates the `*models.Product` from `FindBySlug` before embedding into `ProductWithOptions`).
- Deliberately left `AdminListProducts`/`AdminGetProduct` untouched — admin's `frontend/src/views/admin/ProductDetail.vue` already manages images by raw object key (delete/reorder/set-primary) and fetches presigned URLs separately via `/api/images/product/{id}` for display.
- `go build ./...` and `go test ./...` both pass; existing `TestListProducts_Handler` still passes since `storage.Client` is nil in tests, exercising the no-op path.
- User reported mid-implementation that shopping cart images are broken too. Investigated and found the identical root cause in `internal/handlers/cart.go`'s `GetCart`: `item.Product.Images` was set to the raw `row.ProductImages` key. Fixed by presigning through the same `presignProductImages` helper. `frontend/src/views/Checkout.vue` reads from the same `useCartStore`/`GetCart` data, so it's fixed for free.
- While tracing the image pipeline, found a second, independent root cause: `frontend/src/views/admin/Products.vue`'s quick "Add Product" modal uploaded via `/images/upload` and stored the response's presigned `url` (24h TTL) directly as the product's image, instead of the stable `path` (S3 object key) that the rest of the app stores and that `admin/ProductDetail.vue`'s per-product upload flow correctly uses. Any product created through that modal would show a working image for 24 hours, then break — plausibly the actual trigger for "storefront image ... now incorrect". Fixed `uploadImage`/`saveProduct` in that file to use `data.path` instead of `data.url`. `frontend/src/views/admin/Products.vue` build (`npm run build`) passes.
- Blind Hunter review ran against the full worktree diff (this fix plus pre-existing uncommitted work); findings unrelated to this change were verified and deferred to `deferred-work.md` rather than fixed here (see Review Triage Log). One finding (`AdminListProducts` not presigning) was verified false — the admin product list table has no image column, so nothing renders it.

## Review Triage Log

- `AdminListProducts`/`AdminGetProduct` don't presign images — **false**: verified `frontend/src/views/admin/Products.vue`'s table (`productColumns`: name/category/price/status/actions) never binds `record.images`/`record.primary_image`; admin's per-product detail page fetches presigned URLs separately via `/api/images/product/{id}` by design. No display impact.
- 10 findings unrelated to this change (cart nil-pointer panic on anonymous update/remove, missing panic-recovery middleware, CORS methods gap, insecure default JWT secret, cart variant/product mismatch validation, `RemoveTag` missing existence check, SKU length validation gap, `GenerateVariants` group-order edge case, missing `CompareAtPrice` validation, dead `is_new`/`on_sale` fields, `compare_at_price` not wired into storefront price display) — verified real via direct code inspection, but pre-existing and not caused or exposed by this fix. Logged to `deferred-work.md`.
