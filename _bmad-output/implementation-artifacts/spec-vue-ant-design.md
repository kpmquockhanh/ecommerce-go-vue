---
title: 'Implement Vue Ant Design v4'
type: 'feature'
created: '2026-09-05'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: '190f3c11f0aa850721dcebcfa8d789b23dec75e6'
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The ecommerce frontend uses Tailwind CSS with custom inline styles, requiring manual component styling and lacking enterprise UI patterns needed for forms, tables, modals, and admin dashboards.

**Approach:** Replace Tailwind CSS with Ant Design Vue v4 (4.2.6) using CSS-in-JS theming via ConfigProvider tokens and on-demand auto-imports via unplugin-vue-components. Migrate all 26 Vue files to use Ant Design components.

## Boundaries & Constraints

**Always:** Pinia stores and Vue Router remain unchanged; brand color #2D6A4F applied via ConfigProvider token; all components use on-demand auto-import (no manual imports for regular components).

**Never:** No Less/PostCSS theme overrides; no migration to Ant Design Vue Pro template; no SSR; no changes to store logic or API layer.

</frozen-after-approval>

## Code Map

- `frontend/package.json` — dependencies to update (add ant-design-vue@4, unplugin-vue-components; remove tailwindcss, @tailwindcss/vite)
- `frontend/vite.config.js` — add AntDesignVueResolver({ importStyle: false }), remove tailwindcss plugin
- `frontend/src/main.js` — register Antd plugin, remove style.css import
- `frontend/src/style.css` — remove Tailwind import, keep minimal body styles
- `frontend/src/App.vue` — migrate header/nav/footer to a-layout, a-menu, a-input-search, a-badge; wrap in ConfigProvider with theme tokens
- `frontend/src/views/Home.vue` — hero, features, categories, featured products, newsletter
- `frontend/src/views/Products.vue` — filters sidebar, search, sort, product grid, pagination
- `frontend/src/views/ProductDetail.vue` — image gallery, variant selector, reviews, add-to-cart
- `frontend/src/views/Cart.vue` — item list with quantity controls, order summary
- `frontend/src/views/Checkout.vue` — shipping form, Stripe payment, order confirmation
- `frontend/src/views/Login.vue` — email/password form with error handling
- `frontend/src/views/Register.vue` — registration form with validation
- `frontend/src/views/Orders.vue` — order history table
- `frontend/src/views/OrderDetail.vue` — order detail with items table
- `frontend/src/views/admin/AdminLayout.vue` — admin sidebar nav shell
- `frontend/src/views/admin/Dashboard.vue` — stat cards, recent orders table
- `frontend/src/views/admin/Products.vue` — product CRUD table + modal
- `frontend/src/views/admin/Orders.vue` — order table with status filter
- `frontend/src/views/admin/Users.vue` — user table with search, role edit
- `frontend/src/views/admin/Categories.vue` — category CRUD table + modal
- `frontend/src/views/admin/DeadLetters.vue` — dead letter queue viewer
- `frontend/src/components/ProductCard.vue` — product card with image, badges, rating
- `frontend/src/components/PriceDisplay.vue` — price with sale/strikethrough
- `frontend/src/components/BaseModal.vue` — replace with a-modal (remove custom implementation)
- `frontend/src/components/HelloWorld.vue` — delete (unused scaffolding)
- `frontend/src/components/ToastNotification.vue` — replace with message/notification API
- `frontend/src/components/BaseBadge.vue` — replace with a-badge
- `frontend/src/components/BaseButton.vue` — replace with a-button
- `frontend/src/components/StarRating.vue` — keep as-is (custom, no Ant Design equivalent)
- `frontend/src/components/BasePagination.vue` — replace with a-pagination
- `frontend/src/components/SkeletonLoader.vue` — replace with a-skeleton
- `frontend/src/composables/useToast.js` — replace with Ant Design message/notification API
- `frontend/src/lib/utils.js` — update statusClass() to return Ant Design tag colors instead of Tailwind classes
- `frontend/src/stores/*` — no changes
- `frontend/src/router/index.js` — no changes

## Tasks & Acceptance

**Execution:**
- [x] `frontend/package.json` — install ant-design-vue@4, unplugin-vue-components; remove tailwindcss, @tailwindcss/vite
- [x] `frontend/vite.config.js` — add AntDesignVueResolver({ importStyle: false }), remove @tailwindcss/vite plugin
- [x] `frontend/src/main.js` — import Antd, call app.use(Antd), remove style.css import
- [x] `frontend/src/style.css` — remove `@import "tailwindcss"`, keep body/app styles
- [x] `frontend/src/App.vue` — wrap in ConfigProvider, migrate layout to a-layout components, migrate nav to a-menu, search to a-input-search, cart badge to a-badge, remove all Tailwind classes
- [x] `frontend/src/components/ToastNotification.vue` — delete, replace with message/notification imports
- [x] `frontend/src/components/BaseModal.vue` — delete, replaced by a-modal
- [x] `frontend/src/components/BaseBadge.vue` — delete, replaced by a-badge
- [x] `frontend/src/components/BaseButton.vue` — delete, replaced by a-button
- [x] `frontend/src/components/BasePagination.vue` — delete, replaced by a-pagination
- [x] `frontend/src/components/SkeletonLoader.vue` — delete, replaced by a-skeleton
- [x] `frontend/src/components/HelloWorld.vue` — delete (unused)
- [x] `frontend/src/components/ProductCard.vue` — migrate to a-card, a-badge, a-button
- [x] `frontend/src/components/PriceDisplay.vue` — migrate to a-typography-text
- [x] `frontend/src/views/Home.vue` — migrate hero, features, categories, products to Ant Design components
- [x] `frontend/src/views/Products.vue` — migrate filters, grid, pagination to a-layout-sider, a-card, a-pagination
- [x] `frontend/src/views/ProductDetail.vue` — migrate gallery, form, reviews to a-carousel, a-form, a-list
- [x] `frontend/src/views/Cart.vue` — migrate to a-list, a-input-number, a-button
- [x] `frontend/src/views/Checkout.vue` — migrate to a-form, a-steps, a-card
- [x] `frontend/src/views/Login.vue` — migrate to a-form, a-input, a-button
- [x] `frontend/src/views/Register.vue` — migrate to a-form, a-input, a-button
- [x] `frontend/src/views/Orders.vue` — migrate to a-table, a-tag
- [x] `frontend/src/views/OrderDetail.vue` — migrate to a-descriptions, a-table
- [x] `frontend/src/views/admin/AdminLayout.vue` — migrate to a-layout, a-layout-sider, a-menu
- [x] `frontend/src/views/admin/Dashboard.vue` — migrate to a-statistic, a-card, a-table
- [x] `frontend/src/views/admin/Products.vue` — migrate to a-table, a-modal, a-form, a-upload
- [x] `frontend/src/views/admin/Orders.vue` — migrate to a-table, a-select, a-tag
- [x] `frontend/src/views/admin/Users.vue` — migrate to a-table, a-input, a-modal, a-tag
- [x] `frontend/src/views/admin/Categories.vue` — migrate to a-table, a-modal, a-form
- [x] `frontend/src/views/admin/DeadLetters.vue` — migrate to a-table, a-button
- [x] `frontend/src/composables/useToast.js` — delete, replaced by message/notification API
- [x] `frontend/src/lib/utils.js` — update statusClass() to return Ant Design tag color strings

**Acceptance Criteria:**
- Given the frontend project, when `npm install` runs, then ant-design-vue@4 and unplugin-vue-components are installed without errors
- Given vite.config.js, when the build starts, then AntDesignVueResolver resolves components without manual imports
- Given the app renders, when any page loads, then Ant Design components are visible with brand color #2D6A4F
- Given all views, when inspected, then no Tailwind utility classes remain in templates
- Given `npm run build`, when executed, then it completes successfully with no errors

## Implementation Notes

- ant-design-vue@4 installed with CSS-in-JS (emotion-based). No Less/PostCSS pipeline needed.
- unplugin-vue-components@32 with AntDesignVueResolver({ importStyle: false }) for on-demand auto-imports.
- ConfigProvider wraps entire app in App.vue with full theme token mapping.
- 7 components deleted (BaseModal, BaseBadge, BaseButton, BasePagination, SkeletonLoader, HelloWorld, ToastNotification).
- useToast composable deleted; all toast calls replaced with message.success/error/info/warning API.
- StarRating.vue kept as custom component (no Ant Design equivalent).
- lib/utils.js statusClass() returns Ant Design tag color strings instead of Tailwind classes.
- Build warning: main chunk ~904KB. Consider lazy-loading admin routes for production optimization.

## Spec Change Log

## Review Triage Log

- **main.js double-imports reset.css** — `false`. `AntDesignVueResolver({ importStyle: false })` imports no CSS; `reset.css` provides base reset. Complementary, not duplicative.
- **Cart.vue success toast on failed delete** — `false`. `message.success` fires after `await cart.removeItem(itemId)` — correctly sequenced. The missing try/catch is a pre-existing pattern across the codebase, not from this migration. Relegated to defer.
- **Dockerfile runs npm run dev** — `defer`. Pre-existing, not from this migration.
- **Dockerfile uses npm install** — `defer`. Pre-existing, not from this migration.
- **Checkout.vue pre-fills dummy shipping data** — `defer`. Pre-existing behavior.
- **Checkout.vue never cleans up Stripe script** — `defer`. Pre-existing behavior.
- **Multiple views silently swallow errors** — `defer`. Pre-existing pattern.
- **Router guard reads localStorage directly** — `defer`. Pre-existing design.
- **Products.vue hardcodes category list** — `defer`. Pre-existing.
- **App.vue mutates cart store directly** — `defer`. Pre-existing.
- **Form validation absent** — `defer`. Pre-existing.
- **Duplicate formatDate functions** — `defer`. Pre-existing.
- **Home.vue newsletter no handler** — `defer`. Pre-existing.

## Design Notes

Theme token mapping from existing Tailwind arbitrary values:
```
colorPrimary: '#2D6A4F'      (was bg-[#2D6A4F], text-[#2D6A4F])
colorSuccess: '#52B788'       (was text-[#52B788])
colorError: '#C1121F'         (was text-[#C1121F])
colorWarning: '#E09F3E'       (was text-[#E09F3E])
colorBgLayout: '#F8F6F3'      (was bg-[#F8F6F3])
colorBgContainer: '#FFFFFF'   (was bg-white)
colorBorder: '#E8E4DD'        (was border-[#E8E4DD])
colorText: '#1A1A1A'          (was text-[#1A1A1A])
colorTextSecondary: '#6B6B6B' (was text-[#6B6B6B])
borderRadius: 8                (was rounded-lg)
```

## Verification

**Commands:**
- `cd frontend && npm run build` — expected: build succeeds with no errors
- `cd frontend && npm run dev` — expected: app starts, pages render with Ant Design components

**Manual checks:**
- Header shows Ant Design menu, search input, and cart badge
- Product listing page shows Ant Design cards/table with pagination
- Forms (login, register, checkout) use Ant Design form components with validation
- Admin tables use Ant Design table with proper columns and actions
- Brand color #2D6A4F is reflected across all components
