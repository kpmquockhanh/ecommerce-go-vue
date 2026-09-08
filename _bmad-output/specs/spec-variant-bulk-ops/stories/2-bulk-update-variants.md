---
title: 'Bulk-update variants'
type: 'feature'
created: '2026-09-07'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Admins can only update one variant at a time via `PUT /api/admin/products/{id}/variants/{variant_id}` (`UpdateVariant` in `internal/handlers/product.go`). After generating or otherwise accumulating many variants, editing stock, SKU, or dimensions across several of them means one HTTP call per variant.

**Approach:** Add `PATCH /api/admin/products/{id}/variants/bulk` accepting `{items: [{variant_id, stock?, sku?, weight?, length?, width?, height?}]}`. Reject the whole request with 400 if any item includes `option_values` (recombination stays on the existing single-variant `PUT` endpoint). For each item: verify the variant exists and belongs to the path's `product_id` (new repository lookup `GetVariantProductID`, since none currently exists), then apply the update via the existing `UpdateVariant` repository path — same optional-field/SKU-trim/SKU-uniqueness handling `UpdateVariant`'s handler already does. Process items independently (not one DB transaction): a bad `variant_id` in the batch fails only that item; the rest still apply. Return 200 with a per-item `{variant_id, success, error?}` result array plus overall succeeded/failed counts — no item-level error aborts the batch.

</frozen-after-approval>

## Implementation Notes

- `option_values` presence is detected via `json.RawMessage` (non-nil after decode means the key was present, regardless of value), rather than checking the decoded field's zero-ness, so `"option_values": null` and `"option_values": []` are both correctly treated as "included" per the spec's literal wording.
- `GetVariantProductID` is a new lightweight repository method (`SELECT product_id FROM product_variants WHERE id = $1`) — no such single-variant lookup existed before; needed to enforce the "must belong to this product" ownership check without changing `UpdateVariant`'s signature.
- Reused `UpdateVariant`'s existing SKU-trim/empty-check and unique-constraint-to-"SKU already exists" mapping conventions from the single-variant `UpdateVariant` handler, applied per-item instead of per-request.
- Per-item failures (missing `variant_id`, not found, wrong product, blank SKU, SKU conflict, or an unexpected repository error) are all captured in that item's result rather than aborting the request — matches CAP-3's independent-per-item processing requirement. Unexpected repository errors are still logged server-side (`log.Printf`) even though the HTTP response stays 200.
- Files touched: `internal/models/product.go` (+`BulkUpdateVariantItem`/`Request`/`Result`/`Response`, +`encoding/json` import), `internal/repositories/product.go` (+`GetVariantProductID`), `internal/handlers/product.go` (+`BulkUpdateVariants`, `applyBulkVariantUpdate`), `cmd/server/main.go` (+route), `internal/handlers/product_test.go` (mock extensions + 9 new tests).
- No intent gaps surfaced during implementation; stayed within the oneshot plan.
- Blind-hunter review (8 findings, N=7) triaged: 1 patched, 6 addressed by adding test coverage, 1 deferred. Patched: added a batch-size cap (`len(req.Items) > maxVariantsPerProduct` → 400), matching the codebase's existing pattern of hard-rejecting over-limit input (max option groups, max tags, max variants). 7 new tests added covering method-not-allowed, invalid product ID, invalid JSON, the oversized-batch rejection, the `GetVariantProductID` internal-error path, weight/length/width/height forwarding, and duplicate `variant_id` entries processed independently.

## Review Triage Log

- **medium, patched** — `BulkUpdateVariantsRequest.Items` had no upper bound, unlike `CreateVariant`/`GenerateVariants` which both enforce the 500-variant cap. Fixed: requests with more than `maxVariantsPerProduct` (500) items are now rejected with 400 before any processing.
- **low, patched (coverage)** — `GetVariantProductID` returning an unexpected (non-`ErrNotFound`) error was unhandled by any test. Added `TestBulkUpdateVariants_VariantLookupInternalError`.
- **low, patched (coverage)** — malformed JSON body untested. Added `TestBulkUpdateVariants_InvalidJSON`.
- **low, patched (coverage)** — invalid/non-numeric product ID in the path untested. Added `TestBulkUpdateVariants_InvalidProductID`.
- **low, patched (coverage)** — disallowed HTTP method (405) untested. Added `TestBulkUpdateVariants_RejectsMethodNotAllowed`.
- **low, patched (coverage)** — `weight`/`length`/`width`/`height` pass-through was never asserted. Added `TestBulkUpdateVariants_ForwardsDimensions`.
- **low, patched (coverage)** — duplicate `variant_id` within one batch was implemented (processed independently, no dedup) but unverified. Added `TestBulkUpdateVariants_DuplicateVariantIDProcessedIndependently`, confirming both entries apply in order.
- **low, deferred** — `applyBulkVariantUpdate` doesn't pre-validate SKU length against the `VARCHAR(100)` column before calling `UpdateVariant`; an overlong SKU falls through to a generic 500. Inherited from the pre-existing single-variant `UpdateVariant` handler's identical gap, not introduced by this story. Logged in `deferred-work.md`.

