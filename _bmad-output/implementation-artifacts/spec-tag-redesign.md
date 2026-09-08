---
title: 'Restyle Tag component to match brand tokens'
type: 'refactor'
created: '2026-09-08'
status: 'done'
baseline_commit: '3dd8d37c3a60bf1d553c461a59cee39960671326'
route: 'oneshot'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Every `<a-tag>` in the Vue frontend (`frontend/src`) uses raw Ant Design preset colors (`green`/`red`/`orange`/`purple`/`error`/`warning`/default) and antd's default ~4px radius, which clashes with the app's own brand palette (`#2D6A4F` primary green, warm neutrals) and its 8/12/16px rounding scale used everywhere else.

**Approach:** Add a shared set of semantic tag classes to `frontend/src/style.css` built from the app's real tokens (`.tag-success`, `.tag-warning`, `.tag-error`, `.tag-neutral`, `.tag-plum`, `.tag-solid-error`, `.tag-photo-badge` + a `.tag-uppercase` modifier — 8px radius, 4px 12px padding, 13px/600 weight, no border), then swap every `<a-tag color="...">`/`:color="..."` usage across the app to the matching class, per the previously agreed redesign:

- `frontend/src/lib/utils.js` `statusClass()` — return `tag-*` class names instead of antd color strings; its 4 call sites (`Orders.vue:33`, `OrderDetail.vue:8`, `admin/Orders.vue:53`, `admin/Dashboard.vue:35`) switch from `:color=` to `:class=`.
- `ProductDetail.vue:60` category tag → `class="tag-success tag-uppercase"`, drop `color="green"`; remove the now-redundant `.product-category-tag` uppercase-only rule in `style.css:1031`.
- `admin/Products.vue:24` category tag → add `class="tag-neutral"`.
- `ProductCard.vue` badge (`badgeColor` computed, lines 83-87, used at line 18) → return `tag-photo-badge` + a per-variant modifier (`--new`/`--sale`/`--low-stock`) driving text color only; bind via `:class` instead of `:color`.
- `PriceDisplay.vue:9-15` "Save X%" tag → `class="tag-solid-error ml-8"`, drop `color="error"`.
- `Products.vue:84-100` closable filter chips (both) → `class="tag-success"`, drop `color="green"`.
- `Home.vue:16` hero tag → `class="tag-success mb-24"`, drop `color="green"`.
- `admin/Users.vue:29` role tag → `:class="record.role === 'admin' ? 'tag-plum' : 'tag-neutral'"`, drop `:color`.
- `admin/ProductDetail.vue`: status tag (10-12) → `tag-success`/`tag-neutral`; primary-image tag (77) → `tag-success`; values-count (105), SKU (222), option-value chips (235-237, keep inline `style`), field-type (310) → `tag-neutral`; stock tag (232) → `tag-success`/`tag-error`; required tag (313) → `tag-warning`/`tag-neutral`.

Keep every existing conditional (which state maps to which semantic meaning) exactly as it is today — only the color/radius mechanism changes, not the status→meaning mapping. No prop or template restructuring beyond swapping `color`/`:color` for `class`/`:class`. Do not touch unrelated antd components (buttons, badges, alerts) even if visually similar.

</frozen-after-approval>

## Implementation Notes

- Added `.tag-success`/`.tag-warning`/`.tag-error`/`.tag-neutral`/`.tag-plum`/`.tag-solid-error`/`.tag-photo-badge`/`.tag-uppercase` to `frontend/src/style.css`, backed by new `:root` tokens (`--color-warning-bg`, `--color-warning-text`, `--color-error-bg`, `--color-plum-bg`, `--color-plum-text`) alongside the existing ones, rather than hardcoded hex — matches the file's existing convention.
- `--color-warning` (#E09F3E) itself isn't used directly as tag text: at normal tag font-weight/size it fails contrast on a light background, so `--color-warning-text` (#8A5A1E) is a purpose-picked darker shade paired with a light tint background, same pattern as the other semantic tags.
- All ~20 `<a-tag>` sites swapped from `color`/`:color` to `class`/`:class` with no other template/prop changes; `statusClass()` in `frontend/src/lib/utils.js` now returns class names.
- `ProductCard.vue`'s `badgeColor` → `badgeClass`; kept the exact same conditional shape as before (no explicit "none" branch — still relies on the sibling `badge` computed's `v-if` to suppress rendering), per this spec's "keep every existing conditional exactly" boundary.
- `npm run build` passes with no errors. No browser/screenshot tool was available in this session, so the result was not visually verified in a running browser — only via source review and the earlier design-canvas reference it was built from.

## Review Triage Log

- medium — new tag tint colors (`.tag-warning`, `.tag-error` bg, `.tag-plum`) were hardcoded hex instead of `:root` tokens, unlike every other color in the file → patched: promoted to named CSS custom properties.
- low, rejected — flagged that `--color-warning` goes unused by `.tag-warning`: intentional, it fails text contrast on a light tag background; a dedicated `--color-warning-text` shade was added instead (see Implementation Notes).
- false — `.tag-uppercase`'s 11px font-size/letter-spacing called out as scope creep beyond "only swap color for class": this exactly reproduces the tag redesign already shown to and approved by the user in the design canvas that prompted this spec, not new styling invented during implementation.
- false — spec still showed `status: in-progress` with empty Implementation Notes: expected at that point in the workflow; resolved by this Finalize Spec pass.
- low, deferred — `ProductCard.vue` `badgeClass` has no explicit "no badge" branch: pre-existing coupling carried over unchanged from before this refactor; restructuring it was out of the frozen boundary ("keep conditionals exactly"). Logged in `deferred-work.md`.
- maybe-false, rejected — no in-browser visual check was performed (no screenshot/browser tool available this session); if it turned out to be wrong it would only be a low-severity cosmetic mismatch (e.g. closable-tag close-icon spacing), so not worth blocking on.
