---
title: 'Auto-generate variants from option groups'
type: 'feature'
created: '2026-09-07'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Admins can only create product variants one option-value combination at a time, hand-typing a unique SKU for each (`CreateVariant` in `internal/handlers/product.go`). For a product with 2-3 option groups, building every combination by hand is tedious and error-prone, and there is no way to bound how many variants a product accumulates.

**Approach:** Add `POST /api/admin/products/{id}/variants/generate`: load the product's option groups/values, compute the full cartesian product of value combinations, skip combinations that already exist (reuse `HasVariantCombination`), and bulk-create the rest via the existing `CreateVariant` repository path with an auto-generated unique SKU (`{product-slug}-{value-slug-1}-{value-slug-2}...`, numeric suffix on collision) and stock 0. Enforce a 500-variant-per-product cap identically here and in the existing single-variant create path — reject the whole request with 400 if it would be exceeded, matching how the codebase already handles every other cap (max option groups, max tags) as a hard reject rather than a silent truncation. Return 400 if the product has no option groups. Response reports created count, skipped (already-existing) count, and the list of created variants.

</frozen-after-approval>

## Implementation Notes

- Refactored `generateSlug`'s inline character-stripping logic into a pure `slugify(s, fallback string) string` helper (`internal/handlers/product.go`) so it could be reused for building option-value slugs in generated SKUs, without changing `generateSlug`'s existing behavior (still uniqueness-checked against `FindBySlug`).
- Cap enforcement design: computed the raw cartesian-product size (`total`) from option-group value counts *before* materializing any combinations, and reject outright if `total > 500` — this avoids ever iterating/querying a combinatorial explosion, and since `existingCount <= 500` is already guaranteed by the same cap on manual creation, `total <= 500` is a safe proxy for "existingCount + missing" fitting under the cap in the normal case. Added a second, precise `existingCount + len(missing) > 500` check after filtering out already-existing combinations, as a safety net for the edge case where existing variants reference option values that no longer align with the current group/value set (e.g. after an option value was deleted).
- SKU generation retries on insert conflict (catches `strings.Contains(err.Error(), "unique")`, same pattern the existing `CreateVariant`/`UpdateVariant` handlers already use for SKU conflicts) rather than a separate existence pre-check, to avoid a check-then-act race and to reuse the DB's own unique constraint as the source of truth.
- Added the 500-variant cap check to the existing single-variant `CreateVariant` handler too (per SPEC.md constraint that it applies identically to both paths), returning 400 `"Maximum 500 variants per product"` — matches the existing style used for the max-6-option-groups check.
- Files touched: `internal/models/product.go` (+`GenerateVariantsResponse`), `internal/repositories/product.go` (+`CountVariants` interface method and implementation), `internal/handlers/product.go` (+`GenerateVariants` handler, `cartesianCombinations`, `createGeneratedVariant`, `slugify` extraction, cap check in `CreateVariant`), `cmd/server/main.go` (+route), `internal/handlers/product_test.go` (mock extensions + 8 new tests).
- No intent gaps surfaced during implementation; stayed within the oneshot plan.
- Blind-hunter review (8 findings) triaged and two patched: (1) mid-generation failures now return partial `created`/`skipped` progress plus an `error` field instead of a bare 500 with silently-discarded progress; (2) generated SKU base is truncated to 95 chars to stay inside the `VARCHAR(100)` `sku` column and leave room for a collision suffix. Six tests added covering `CountVariants`/`GetOptionGroupsByProduct`/`HasVariantCombination` errors, non-collision create errors, and SKU truncation.

## Review Triage Log

- **false** — "CAP-3 (bulk update) unimplemented": out of scope for this story. `stories.yaml` scopes Story 1 to CAP-1/CAP-2 only; CAP-3 is Story 2, not yet dispatched.
- **medium, patched** — mid-loop generation failure discarded already-created variants from the response with a bare 500. Fixed: `GenerateVariantsResponse` gained an `Error` field; on failure the handler now returns what was created/skipped plus the error, at 500.
- **medium, patched** — generated SKU had no length bound against the `sku VARCHAR(100)` column; a product with a long slug and long option-value labels could hit an opaque insert failure. Fixed: SKU base is truncated to 95 chars before use.
- **low, deferred** — no transaction/lock around the cap and duplicate-combination checks (`internal/repositories/product.go`, `internal/handlers/product.go`), so two concurrent requests against the same product could both pass checks. Pre-existing pattern from `spec-product-options` stories, not introduced here; logged in `deferred-work.md`.
- **low, deferred** — `GenerateVariants` does one `HasVariantCombination` DB round-trip per combination (N+1), bounded by the 500-variant cap. Logged in `deferred-work.md`.
- **low, patched** — new error branches (`CountVariants`, `GetOptionGroupsByProduct`, `HasVariantCombination` errors; non-unique-constraint create errors) lacked test coverage. Added `TestGenerateVariants_OptionGroupsLookupError`, `TestGenerateVariants_CountVariantsError`, `TestGenerateVariants_HasVariantCombinationError`, `TestGenerateVariants_NonCollisionCreateErrorReportsPartialProgress`, `TestCreateVariant_CountVariantsError`.
- **false** — "empty `product.Slug` feeds into SKU prefix": unreachable. `products.slug` is `NOT NULL UNIQUE` and always populated via `generateSlug` (falls back to `"product"` if the input strips to empty), so `FindByID` never returns an empty slug for a normally-created product.
- **false** — "skipped combinations aren't identified in the response, only counted": matches the spec exactly. `SPEC.md` CAP-1's success criterion defines the response contract as created/skipped counts only, not an itemized list of which combinations were skipped.

