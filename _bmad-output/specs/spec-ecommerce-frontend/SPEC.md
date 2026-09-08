---
id: SPEC-ecommerce-frontend
companions:
  - api-endpoints.md
  - frontend-pages.md
  - data-models.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Full-Stack Ecommerce Platform

## Why

The existing Go backend is a Stripe payment demo with 2 endpoints, 3 hardcoded products, no database, and no auth. To ship a functional ecommerce site, the backend must be extended with persistent data, user management, product catalog, cart, and order APIs — then a Vue 3 frontend built to consume them. This is an **opportunity to capture**: turning a payment proof-of-concept into a working storefront.

## Capabilities

- **CAP-1: User Authentication**
  - **intent:** Users can register, log in, and maintain session state via JWT tokens.
  - **success:** A user can register with email/password, receive a JWT, use it to access protected endpoints, and log out (token invalidation or client-side discard).

- **CAP-2: Product Catalog**
  - **intent:** Visitors can browse, search, filter, and paginate through the product catalog.
  - **success:** Product listing page loads with sortable/filterable results; search returns relevant matches; pagination works correctly with configurable page size.

- **CAP-3: Product Detail**
  - **intent:** Users can view full product information including description, images, price, variants, and customer reviews.
  - **success:** Product detail page renders all product fields, displays images in a gallery/carousel, shows available variants, and lists existing reviews.

- **CAP-4: Shopping Cart**
  - **intent:** Users can add products to a cart, adjust quantities, remove items, and view the cart total.
  - **success:** Cart state persists across sessions (logged-in users server-side, guests via localStorage). Adding/removing items updates totals in real-time. Cart is accessible from any page.

- **CAP-5: Checkout & Payment**
  - **intent:** Users can enter shipping/billing address and complete payment via Stripe.
  - **success:** Checkout form validates address fields, creates a Stripe PaymentIntent via the backend, processes payment, and creates an order record on success.

- **CAP-6: Order History**
  - **intent:** Authenticated users can view their past orders and order details.
  - **success:** Order list page shows all user orders with status, date, and total. Order detail page shows line items, shipping address, and payment status.

- **CAP-7: Admin Dashboard**
  - **intent:** Admin users can manage products (CRUD), view/update orders, and overview user accounts.
  - **success:** Admin can create/edit/delete products, change order status, and view a user list. Non-admin users cannot access admin routes.

- **CAP-8: Responsive UI**
  - **intent:** The storefront works well on desktop and mobile browsers.
  - **success:** All pages render correctly on viewport widths from 320px to 1920px. Navigation, product grid, cart, and checkout are usable on mobile devices.

## Constraints

- Backend must use Go's standard `net/http` router (no migration to chi/gorilla/mux).
- Frontend is Vue 3 with Pinia for state management.
- Stripe SDK integration already exists; extend, don't replace.
- Database is PostgreSQL (new dependency for the backend).
- Authentication uses JWT (no session cookies).

## Non-goals

- Payment gateway beyond Stripe (no PayPal, Apple Pay, etc.).
- Real-time inventory management or stock tracking.
- Multi-vendor marketplace or seller accounts.
- Mobile native apps (iOS/Android).
- Email notification system (order confirmations, etc.).
- Internationalization (i18n) — English only for now.
- SEO server-side rendering for the Vue frontend (SPA is acceptable).

## Success signal

A user can register, browse products, add items to cart, complete a Stripe checkout, see their order in history, and an admin can manage products and orders — all through the Vue UI interacting with the Go backend API.

## Assumptions

- Existing Stripe API key and payment flow will be reused and extended.
- PostgreSQL will be the primary database (replacing the current no-database state).
- JWT tokens will be stored client-side in localStorage or httpOnly cookies.
- Product images will be served from the backend or a static file directory.
- The 3 existing hardcoded products become seed data in the database.

## Resolved Questions

- Product images stored on S3-compatible storage (MinIO or AWS S3).
- Admin auth uses same JWT with `role` field — no separate admin login.
- Guest checkout supported: cart works without login, checkout requires auth.
