# Deferred Work

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Dockerfile runs npm dev server instead of production build
  evidence: Pre-existing; not from this migration

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Dockerfile uses npm install instead of npm ci
  evidence: Pre-existing; not from this migration

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Checkout.vue pre-fills dummy shipping data before user input
  evidence: Pre-existing behavior

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Checkout.vue never cleans up Stripe script on unmount
  evidence: Pre-existing behavior

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Multiple views silently swallow errors with empty catch blocks
  evidence: Pre-existing pattern across codebase

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Router guard reads localStorage directly instead of using Pinia auth store
  evidence: Pre-existing design

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Products.vue hardcodes category list instead of fetching from API
  evidence: Pre-existing

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: App.vue mutates cart store properties directly on logout
  evidence: Pre-existing; cart store should expose reset() method

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Form validation absent across Login, Register, Checkout, and admin forms
  evidence: Pre-existing

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Multiple components duplicate formatDate instead of reusing lib/utils.js
  evidence: Pre-existing

- source_spec: `_bmad-output/implementation-artifacts/spec-vue-ant-design.md`
  summary: Home.vue newsletter subscription has no submit handler
  evidence: Pre-existing

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: WebP magic-byte detection in internal/storage/s3.go hardcodes a zero RIFF chunk-size, rejecting nearly all real WebP uploads
  evidence: internal/storage/s3.go:53 matches {0x52,0x49,0x46,0x46,0x00,0x00,0x00,0x00,0x57,0x45,0x42,0x50}; bytes 4-7 of RIFF encode file size and are essentially never zero. Pre-existing, not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: Cart Upsert overwrites customization_data instead of creating a separate line when the same product/variant is added again with different customization values
  evidence: internal/repositories/cart.go:109-119 ON CONFLICT ... DO UPDATE SET quantity = cart_items.quantity + $4, customization_data = $5 replaces the prior customization on repeat add. Pre-existing (story 5 territory), not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: Cart AddItem does not validate CustomizationData against product_customization_fields (required fields can be left empty, unknown keys accepted)
  evidence: internal/handlers/cart.go AddItem (verified ~line 114-154) marshals req.CustomizationData with no lookup of the product's customization field definitions. Pre-existing (story 5 territory), not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: compare_at_price is never validated on product create/update (can be negative or below price)
  evidence: internal/handlers/product.go only calls models.ValidateProductPrice on Price (lines 438, 516); CompareAtPrice has no equivalent check. Pre-existing (product-expansion story), not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: Admin UI has no way to set a product's status to "archived" even though the backend supports it
  evidence: frontend/src/views/admin/Products.vue and admin/ProductDetail.vue status selects only offer Published/Draft. Pre-existing (product-expansion story), not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: Negative option-value price modifiers render as positive in both storefront and admin product detail pages
  evidence: frontend ProductDetail.vue and admin/ProductDetail.vue formatPrice uses Math.abs(cents) and only prepends "+" for positive values, never "-" for negative. Pre-existing (story 2/4 frontend), not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: updateCategory function in frontend/src/views/admin/Products.vue is dead code from an incomplete refactor
  evidence: not referenced by any template element after the category column became a static tag list. Pre-existing, not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: ListProducts/AdminListProducts in internal/handlers/product.go redundantly re-fetch categories per product (N+1) even though List/AdminList already populate them, and swallow lookup errors silently
  evidence: verified pattern of `if err == nil` around a second h.productRepo.GetCategories call per product despite p.Categories already being populated via loadProductAssociations. Pre-existing, not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: CreateVariant and UpdateVariant in internal/repositories/product.go are not atomic — a mid-sequence failure leaves a variant row committed with an incomplete or partially-replaced set of option-value links
  evidence: verified CreateVariant (product.go:725-753) inserts the variant then loops un-transacted Exec calls for variant_option_values; UpdateVariant (product.go:759-830) commits the scalar-field UPDATE separately from the later transactional option-value replacement. Pre-existing (story 1 work), not touched by story 3 (story 3 only changed ValidateOptionValues/HasVariantCombination and their call sites).

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: CreateOptionGroup has a TOCTOU race on the duplicate-name check and no unique-constraint-to-409 fallback, unlike CreateOptionValue/UpdateOptionValue
  evidence: internal/handlers/product.go CreateOptionGroup fetches existing groups in application code before insert; DB has UNIQUE(product_id, name) but no violation-to-409 mapping here. Pre-existing (story 2 work), not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: ReorderOptionValues re-fetches the group via UpdateOptionGroup with an empty request just to read it back, instead of using a proper read method
  evidence: internal/handlers/product.go ReorderOptionValues calls h.productRepo.UpdateOptionGroup(ctx, groupID, &models.UpdateOptionGroupRequest{}); an unexported getOptionGroupByID already exists internally but isn't exposed on the interface. Pre-existing (story 2 work), not touched by story 3.

- source_spec: `_bmad-output/specs/spec-product-options/stories/3-variant-crud-rewrite-option-value-linkage-computed-label.md`
  summary: The option-group/variant/customization-field routes in cmd/server/main.go are dispatched via order-dependent strings.Contains/HasSuffix chains rather than a path-param router, mirrored by string-slicing ID-extraction helpers in the handler
  evidence: verified cmd/server/main.go:306-388 dispatch chain and internal/handlers/product.go extractID/extractProductID/extractIDBetween helpers. Pre-existing architecture (stories 1-2); story 3 added one more helper (extractIDBetween) following the same established pattern rather than introducing a new risk, and touched no routes in main.go. A full fix is a broad routing refactor, out of scope for a small fix.

- source_spec: `_bmad-output/specs/spec-variant-bulk-ops/stories/1-auto-generate-variants-from-option-groups.md`
  summary: GenerateVariants (and the existing CreateVariant) enforce the 500-variant cap and duplicate-combination checks via separate un-transacted DB round-trips, so two concurrent requests against the same product can both pass the checks and together push the product over the cap or create duplicate combinations
  evidence: internal/handlers/product.go GenerateVariants (CountVariants, then a HasVariantCombination loop, then CreateVariant per missing combo) and CreateVariant (CountVariants then HasVariantCombination) have no transaction, advisory lock, or DB-level uniqueness constraint on the option-value combination backing them (confirmed via _bmad-output/specs/spec-product-options/data-model-changes.md: variant_option_values has no such constraint). Pre-existing pattern from story-1/2/3 of spec-product-options, not introduced by this story; low likelihood since these are sequential admin-only actions.

- source_spec: `_bmad-output/specs/spec-variant-bulk-ops/stories/1-auto-generate-variants-from-option-groups.md`
  summary: GenerateVariants checks each cartesian-product combination for existence with its own HasVariantCombination DB round-trip (up to 500 per call) instead of one batched query
  evidence: internal/handlers/product.go GenerateVariants loops over cartesianCombinations(groups) calling h.productRepo.HasVariantCombination once per combination. Bounded by the 500-variant cap so not unbounded, but a real N+1 pattern; fixing it needs a new batched repository method, more than a simple correction for this story's scope.

- source_spec: `_bmad-output/specs/spec-variant-bulk-ops/stories/2-bulk-update-variants.md`
  summary: applyBulkVariantUpdate does not pre-validate an admin-supplied SKU's length against the sku VARCHAR(100) column, so an overlong SKU falls through to a generic "internal server error" instead of a clear per-item validation message
  evidence: internal/handlers/product.go applyBulkVariantUpdate only trims/empty-checks item.SKU before calling UpdateVariant, same as the pre-existing single-variant UpdateVariant handler it reuses (internal/handlers/product.go UpdateVariant has the identical gap). Not introduced by story 2 -- inherited from UpdateVariant's existing validation, which bulk update calls through to unchanged.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: CartHandler.UpdateItem/RemoveItem don't reject an anonymous request with no userID and no X-Session-ID the way AddItem/GetCart do, so cartRepository.UpdateQuantity/Delete dereference a nil sessionID pointer
  evidence: internal/handlers/cart.go:167-220 (UpdateItem, RemoveItem) call h.getCartIdentifier then go straight to the repo call with no `userID == nil && sessionID == nil` guard; internal/repositories/cart.go:123-145 (UpdateQuantity, Delete) unconditionally dereference `*sessionID` in the else branch when userID is nil. Pre-existing, not caused by the image-presigning change.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: No panic-recovery middleware exists anywhere in the server, so any handler panic (e.g. the cart nil-pointer case above) surfaces as a raw connection reset instead of a clean error response
  evidence: `grep -rn "recover()" internal --include="*.go"` returns nothing. Pre-existing, unrelated to this fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: CORSMiddleware only allows POST, GET, OPTIONS, but the API defines many PUT/DELETE/PATCH endpoints (product/variant/option-group CRUD, cart update/remove, etc.), so cross-origin clients on those verbs fail preflight
  evidence: internal/middleware/middleware.go:52 sets `Access-Control-Allow-Methods: POST, GET, OPTIONS` verbatim; internal/handlers/product.go and cart.go register PUT/DELETE handlers. Pre-existing, unrelated to this fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: config.LoadConfig falls back to a hardcoded default JWT signing secret ("your-secret-key-change-in-production") with no fatal check, unlike StripeSecretKey which log.Fatals when unset
  evidence: internal/config/config.go:56 `getEnvOrDefault("JWT_SECRET", "your-secret-key-change-in-production")` vs config.go:82-84 `if cfg.StripeSecretKey == "" { log.Fatal(...) }` with no equivalent check for JWTSecret. Pre-existing, unrelated to this fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: CartHandler.AddItem validates the product exists but never validates that req.VariantID actually belongs to req.ProductID, unlike the option-value/variant validation used elsewhere
  evidence: internal/handlers/cart.go AddItem (114-163) has no equivalent of the ValidateOptionValues-style ownership check used in product.go's variant handlers. Pre-existing, unrelated to this fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: RemoveTag skips the FindByID existence check that AddTags performs, so removing a tag from a non-existent product silently "succeeds" (200) instead of 404
  evidence: internal/handlers/product.go RemoveTag (~1100-1132) vs AddTags (~1024-1063) — AddTags calls FindByID first, RemoveTag does not. Pre-existing, unrelated to this fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: CreateVariant/UpdateVariant don't validate a manually supplied SKU's length against the DB's VARCHAR(100) limit (only the auto-generated path truncates to 95 chars), so a long user-supplied SKU surfaces as an opaque 500 instead of a 400
  evidence: internal/handlers/product.go createGeneratedVariant truncates to 95 chars; CreateVariant/UpdateVariant have no equivalent check on req.SKU. Pre-existing, unrelated to this fix (same class as the already-logged applyBulkVariantUpdate SKU-length gap above).

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: GenerateVariants' cartesian-size loop breaks as soon as the running total exceeds maxVariantsPerProduct, so an option group appearing later in the list is never checked for having zero values -- the "every option group must have at least one value" error can be masked by the combinatorial-limit error depending on group order
  evidence: internal/handlers/product.go GenerateVariants (~792-806) multiplies group sizes into `total` and breaks on `total > maxVariantsPerProduct` before iterating remaining groups. Pre-existing, unrelated to this fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: No validation enforces CompareAtPrice is positive or greater than Price, so a product's "sale" price can be set to a nonsensical value (e.g. lower than the actual price, or negative)
  evidence: models.ValidateProductPrice only validates Price; CreateProduct/UpdateProduct in internal/handlers/product.go never validate CompareAtPrice against Price. Pre-existing, unrelated to this fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: product.is_new and product.on_sale are read by ProductCard.vue's badge logic but are never produced anywhere in the backend or frontend, so the "New"/"Sale" badges can never render
  evidence: frontend/src/components/ProductCard.vue:77-85 reads product.is_new/product.on_sale; grep across internal/ and frontend/src finds no producer for either field. Pre-existing, unrelated to this fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-storefront-image-display-fix.md`
  summary: compare_at_price is never wired into the customer-facing storefront -- PriceDisplay.vue already supports an originalPrice/strikethrough/"Save X%" display, but neither ProductCard.vue nor ProductDetail.vue passes compare_at_price into it
  evidence: frontend/src/components/PriceDisplay.vue supports originalPrice; frontend/src/components/ProductCard.vue and frontend/src/views/ProductDetail.vue's <PriceDisplay> usages don't pass compare_at_price. Pre-existing, unrelated to this fix.

- source_spec: `_bmad-output/implementation-artifacts/spec-tag-redesign.md`
  summary: ProductCard.vue's badgeClass computed has no explicit "no badge" branch, defaulting to the low-stock class; correctness today depends entirely on the separate badge computed's v-if guarding it
  evidence: Pre-existing coupling (the prior badgeColor computed had the same unconditional 'orange' fallback); spec required preserving conditionals exactly, so restructuring was out of scope for this change
