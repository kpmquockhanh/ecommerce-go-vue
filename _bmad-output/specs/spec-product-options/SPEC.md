---
id: SPEC-product-options
companions: [data-model-changes.md]
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Product Options & Dynamic Variants

## Why

The current variant system hardcodes `size` and `color` as the only variant dimensions. Every product — regardless of whether it's clothing, electronics, furniture, or food — is forced into the same two-attribute model. This breaks for products with no size (e.g., a phone case with only color options), products with three+ dimensions (e.g., a t-shirt with Size + Color + Material), and products that need custom user input (e.g., engraving text, monogram). The hardcoded model also prevents price differentiation per option (e.g., "Premium material +$10") and makes the variant label always render as "Size / Color" even when one attribute is empty. This blocks real-world product catalogs and forces workarounds (empty strings, unused columns).

## Capabilities

- **CAP-1**
  - **intent:** Admin can define named option groups (e.g., "Size", "Color", "Material") on a per-product basis, with each group containing an ordered list of option values.
  - **success:** A product can have 0–N option groups. Each option group has a name and 1–N values. Reordering option values within a group is supported. Deleting a product cascades to its option groups and values.

- **CAP-2**
  - **intent:** Admin creates product variants as specific combinations of option values, replacing the hardcoded size/color columns with a dynamic linking table.
  - **success:** A variant is defined by its set of selected option values (one per option group). Creating a variant without exactly one value per option group returns an error. Duplicate combination within the same product returns a conflict error. The old `size` and `color` columns are dropped from `product_variants`.

- **CAP-3**
  - **intent:** Storefront product detail pages render option selectors dynamically based on the product's option groups, with each group showing its values as selectable buttons or swatches.
  - **success:** When a product has option groups, the storefront renders a selector per group. Selecting a combination highlights the matching variant and shows its stock/SKU. Selecting a combination with no matching variant shows "Unavailable". Out-of-stock variant values are visually disabled.

- **CAP-4**
  - **intent:** Admin can set a price modifier (positive or negative integer cents) on individual option values, which adjusts the displayed and charged price when that option is selected.
  - **success:** When an option value has a `price_adjustment`, the displayed price becomes `base_price + sum(adjustments of selected values)`. Price is recalculated on each option selection change in the storefront. Cart and order snapshot the final adjusted price at time of add/checkout.

- **CAP-5**
  - **intent:** Admin can define per-product customization fields (free-text input, limited-length text, or file upload) that customers fill out when ordering, with optional price surcharge.
  - **success:** Customization fields are defined per-product with a label, input type (text/textarea/file), optional max_length, optional price_surge (integer cents), and sort_order. Cart items store the customer-provided customization values. Order items snapshot the customization values. Empty required fields block add-to-cart.

- **CAP-6**
  - **intent:** Variant label is computed dynamically from selected option values instead of hardcoded "Size / Color".
  - **success:** Variant label is built as "Value1 / Value2 / ..." in option group display order. Frontend renders the computed label. Cart and order items store the denormalized label at snapshot time.

## Constraints

- The `product_variants` table already exists; new variant creation uses the same table but drops `size` and `color` columns.
- Option groups are product-scoped, not global — each product defines its own option structure.
- Price modifiers are per-option-value, not per-variant. The variant's effective price is derived.
- Customization field values are stored as JSONB on `cart_items` and denormalized as JSONB on `order_items`.
- Max 6 option groups per product, max 100 option values per group, max 50 characters per value label.
- Max 5 customization fields per product, max 500 characters per text input, max 10MB per file upload.
- All new admin endpoints require the existing admin auth middleware.
- Public endpoints must not expose draft or archived products' option structures.

## Non-goals

- Predefined/global option catalogs (e.g., "Standard Size Chart") — options are per-product, not reusable templates.
- Conditional/dependent option groups (e.g., "Color options change based on selected Size") — all groups are independent.
- Option-level inventory tracking (stock tracked at variant level, not per-option-value).
- Customer-saved customization presets or profiles.
- Multi-language option value labels — labels are single-language strings.

## Success signal

- Admin creates a product with option groups "Size" (S, M, L) and "Color" (Red, Blue), then creates 6 variants (one per combination). Storefront displays two selectors, and choosing S + Red highlights the correct variant and shows its stock.
- Admin adds a "Material" option group with values "Cotton ($0)" and "Silk (+$10)" to a $50 product. Storefront shows $50 when Cotton is selected and $60 when Silk is selected.
- Admin defines a "Custom Engraving" text customization field with a $5 surcharge. Customer enters text, sees price increase, and the engraving text appears in their order details.
- A product with no option groups displays "Add to Cart" directly (no variant selector), preserving backward compatibility.
- Variant label in cart and order reads "M / Red / Silk" instead of "Size / Color".

## Assumptions

- Existing products with size/color variants will be migrated: size becomes an option group "Size", color becomes "Color", each existing variant row becomes a combination of option values.
- The `product_variants` table drops `size` and `color` columns after migration. No backward compatibility with the old column-based model.
- Frontend variant selector is rebuilt as a dynamic component driven by the option groups API response.
- Customization file uploads use the existing S3/MinIO storage infrastructure.

## Open Questions

<!-- All resolved. See memlog for decision history. -->
