---
title: 'Variant CRUD rewrite — option value linkage + computed label'
type: 'bugfix'
created: '2026-09-07'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Story 3's variant option-value rewrite is already implemented in the working tree (models, repository joins, `Label` computation, routes, and the admin frontend all use `option_value_ids`/`variant_option_values`), but three gaps remain against the spec's CAP-2 success criteria ("exactly one value per option group", "duplicate combination returns a conflict"): (1) `ValidateOptionValues` only rejects *extra* values per group, never checks that *every* option group is covered — an empty or partial `option_value_ids` list silently passes; (2) `UpdateVariant`'s handler never calls `ValidateOptionValues`/`HasVariantCombination` at all, so reassigning option values on update bypasses both checks; (3) `HasVariantCombination` has no way to exclude the variant being updated, so re-saving a variant with its own unchanged combination would false-positive as a duplicate once (2) is fixed.

**Approach:** Make `ValidateOptionValues` (internal/repositories/product.go) verify the selected values cover every one of the product's option groups exactly once, regardless of list length. Add an `excludeVariantID int` parameter to `HasVariantCombination` (pass `0` from `CreateVariant`, the target variant's ID from `UpdateVariant`) so it ignores the row being updated. In `CreateVariant` handler, always run both checks (drop the `len(req.OptionValues) > 0` guard now that empty lists are handled correctly). In `UpdateVariant` handler, run both checks only when `req.OptionValues != nil` (partial updates that don't touch options must not be forced to re-validate), extracting `productID` from the `/api/admin/products/{id}/variants/{variant_id}` URL via a small new path-parsing helper. Add focused unit tests for the fixed coverage check and the exclude-self duplicate check.

## Boundaries & Constraints

**Always:** Preserve existing behavior for products with zero option groups (empty `option_value_ids` stays valid). Keep `CreateVariant`'s existing SKU-uniqueness and product-existence checks unchanged.

**Never:** Do not touch storefront/customization work (stories 4–5) or add broad new test coverage beyond the fixed logic (story 6's job). Do not change the `product_variants`/`variant_option_values` schema.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Create, product has 2 option groups, empty `option_value_ids` | `[]` | 400 error naming the missing group(s) | Handled by fixed `ValidateOptionValues` |
| Create, product has 2 option groups, only 1 value given | `[id_for_group_A]` | 400 error — group B not covered | Handled by fixed `ValidateOptionValues` |
| Update, `option_value_ids` omitted from body | field absent (nil) | Stock/SKU/dimensions updated, options untouched, no validation run | N/A |
| Update, `option_value_ids` resubmits the variant's own current combination | same IDs as already linked | 200 OK, not treated as duplicate | Handled by `excludeVariantID` |
| Update, `option_value_ids` matches a *different* variant's combination | IDs used by another variant | 409 Conflict | Handled by `HasVariantCombination` |

</frozen-after-approval>

## Implementation Notes

- `internal/repositories/product.go` `ValidateOptionValues`: restructured to fetch the product's option groups unconditionally, then check both "no group gets two values" (existing) and "every group gets exactly one value" (new). Zero option groups + zero submitted IDs is still valid (backward compatible with option-less products).
- `internal/repositories/product.go` `HasVariantCombination`: added `excludeVariantID int` param, `AND pv.id != $4` in the SQL. `CreateVariant` passes `0` (no real variant has id 0); `UpdateVariant` passes the variant being updated.
- `internal/handlers/product.go` `CreateVariant`: removed the `len(req.OptionValues) > 0` guard — both checks now always run (repo handles the empty-list case correctly).
- `internal/handlers/product.go` `UpdateVariant`: added the same two checks, gated on `req.OptionValues != nil` (nil means the field was omitted from the JSON body — a genuine partial update that doesn't touch options). Added `extractIDBetween` helper to pull `productID` out of the `/products/{id}/variants/{variant_id}` URL, since the existing `extractID`/`extractProductID` helpers only handle a single trailing or single-prefixed segment.
- `internal/handlers/product_test.go`: added configurable `validateOptionValuesErr`/`hasVariantCombinationResult`/`hasVariantCombinationErr` fields plus call-tracking fields to `mockProductRepo`, and 6 new tests covering: validation-error propagation, duplicate-combination 409, always-validates-on-create, skip-validation-when-omitted-on-update, validate-and-exclude-self-on-update, duplicate-on-update. `gofmt -w` applied to the file for field alignment.
- No repository-level (SQL) tests were added — this project has no DB-backed test infrastructure (no testcontainers/sqlmock, no existing `internal/repositories/*_test.go`), consistent with story 6 being where broader test coverage lands.

## Review Triage Log

Blind-hunter review ran over the full uncommitted worktree (stories 1-2's option-groups work plus story 3's fixes, since the tree was kept dirty per the human's choice at Step 1). All 12 findings were verified real but trace to code story 3 did not touch — routed to `defer` and logged in `deferred-work.md`:

- `high` (data integrity, unverified severity not required — confirmed by reading the code): WebP magic-byte check hardcodes a zero RIFF chunk-size, rejecting real uploads. Pre-existing (`internal/storage/s3.go:53`), not from story 3.
- `medium`: Cart `Upsert` overwrites `customization_data` instead of splitting a line on repeat add with different customization. Pre-existing (story 5 territory).
- `medium`: Cart `AddItem` never validates `CustomizationData` against required customization fields. Pre-existing (story 5 territory).
- `medium`: `compare_at_price` has no validation (can be negative/below price), unlike `Price`. Pre-existing (product-expansion story).
- `medium`: No "archived" option in the admin status selects despite backend support. Pre-existing (product-expansion story).
- `medium`: Negative option-value price modifiers render as positive in both storefront and admin UI (`Math.abs` with no "-" sign). Pre-existing (story 2/4 frontend).
- `low`, rejected as not worth a separate fix here: dead `updateCategory` function in `admin/Products.vue`. Pre-existing.
- `medium`: N+1 category re-fetch in `ListProducts`/`AdminListProducts` with silently swallowed errors. Pre-existing.
- `high` (data integrity): `CreateVariant`/`UpdateVariant` are not atomic — a mid-sequence failure can leave a variant committed with incomplete option-value links. Verified by reading `internal/repositories/product.go:725-830`; pre-existing from story 1, story 3 only changed `ValidateOptionValues`/`HasVariantCombination` and their call sites.
- `medium`: `CreateOptionGroup` has a TOCTOU race on its duplicate-name check with no unique-constraint-to-409 fallback. Pre-existing (story 2 work).
- `low`: `ReorderOptionValues` re-fetches via `UpdateOptionGroup` with an empty request instead of a dedicated read method. Pre-existing (story 2 work).
- `low`/`medium` (architecture, not a bug): hand-rolled `strings.Contains`/`HasSuffix` route dispatch in `main.go`, mirrored by `extractID`/`extractProductID`/`extractIDBetween` helpers. Story 3 added `extractIDBetween` following the same established pattern and touched no routes — the underlying architecture predates this story and a real fix is a broad router refactor, out of scope.

No findings were traced to story 3's own diff (`ValidateOptionValues` coverage check, `HasVariantCombination`'s `excludeVariantID`, the `CreateVariant`/`UpdateVariant` handler wiring, or the new tests).

