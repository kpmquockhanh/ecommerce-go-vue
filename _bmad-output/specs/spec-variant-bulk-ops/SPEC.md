---
id: SPEC-variant-bulk-ops
companions: []
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Variant Auto-Generation & Bulk Update

## Why

Admins can define option groups (Size, Color, Material) and create variants, but only one combination at a time, each requiring a hand-typed unique SKU (`CreateVariant`/`UpdateVariant` in `internal/handlers/product.go`). For a product with 2-3 option groups this means manually building every combination and inventing every SKU by hand, and editing stock or dimensions across many variants means one API call per variant. `spec-product-expansion` explicitly deferred "bulk product operations." This spec closes that gap for variants specifically: generate all combinations from existing option groups in one action, and edit many existing variants in one request.

## Capabilities

- **CAP-1**
  - **intent:** Admin triggers generation of the full cartesian product of a product's option group values into variants in one action, instead of creating each combination by hand.
  - **success:** Calling generate on a product with option groups Size(S,M,L) x Color(Red,Blue) with no existing variants creates all 6 variants, each with a system-generated unique SKU and stock 0. A product with 0 option groups returns a 400 error (nothing to generate). Total resulting variants for the product must not exceed the per-product cap (see Constraints).

- **CAP-2**
  - **intent:** Generation is safe to re-run after option groups change, only creating what's missing, so admins can add an option value later without redoing prior work.
  - **success:** Re-running generate with no option changes creates 0 new variants and reports all combinations as already-existing. After adding a new Color value "Green" to the Size/Color example, re-running generate creates exactly the 3 new S/M/L x Green variants and leaves the original 6 (their stock, SKU, and dimensions) untouched.

- **CAP-3**
  - **intent:** Admin updates stock, SKU, and dimensions on many existing variants of a product in a single request instead of one call per variant.
  - **success:** A request carrying a list of `{variant_id, fields}` applies each item's changes (stock/SKU/weight/length/width/height) independently and returns a per-item success/failure result. A batch of 6 items with all valid variant IDs returns 6 successes in one call. A batch where one `variant_id` belongs to a different product returns that item as failed while the other items still succeed.

## Constraints

- Reuses existing tables (`product_variants`, `product_option_groups`, `product_option_values`, `variant_option_values`) — no schema changes.
- SKU stays required and unique per variant (existing DB constraint). Generation must produce a unique SKU per created variant without admin input: `{product-slug}-{option-value-slug-1}-{option-value-slug-2}-...`, lowercase and hyphenated, with a numeric suffix appended on collision.
- New hard cap: a product may not exceed 500 total variants. Enforced identically by generation (rejects/truncates before exceeding it) and by the existing single-variant create endpoint.
- Bulk update does not accept `option_values` — changing a variant's option-value combination stays on the existing single-variant `PUT` endpoint. Bulk scope is limited to stock, SKU, and dimensions.
- Every `variant_id` in a bulk-update request must belong to the `product_id` in the path; a mismatched ID is rejected as a per-item failure, not a fatal error for the whole batch.
- Duplicate-combination and option-value-validation rules from `spec-product-options` CAP-2 apply to generated variants exactly as they do to manually created ones.
- Both new endpoints require the existing admin auth middleware, consistent with all other admin product/variant endpoints.

## Non-goals

- Filter/query-based mass update (e.g. "set stock=0 for every Red variant") — only explicit ID-list bulk update is in scope; selection-by-filter is a UI concern that can be layered on top later without changing this contract.
- CSV/spreadsheet import or export of variants.
- Generating variants across multiple products in one call.
- Undo/rollback of a completed bulk generate or bulk update operation.
- A dry-run/preview mode that shows what generation would create without committing it.

## Success signal

- Admin defines Size(S,M,L) x Color(Red,Blue) on a product with zero variants, calls generate, and gets 6 new variants with unique SKUs and stock 0 in one call — no manual per-combination entry.
- Admin adds a "Green" color value and re-runs generate: exactly 3 new variants appear; the original 6 are unchanged.
- Admin submits one bulk-update request setting stock for all 6 variants; all 6 update in a single call, and a response item for an invalid variant ID reports failure without blocking the other 5.
- Generating on a product with no option groups returns a clear 400 instead of silently doing nothing.

## Assumptions

- Auto-generated variants default to stock 0 and null weight/dimensions, matching the defaults an admin would use for manual creation — filled in afterward via CAP-3.

## Open Questions

<!-- All resolved. See memlog for decision history. -->
