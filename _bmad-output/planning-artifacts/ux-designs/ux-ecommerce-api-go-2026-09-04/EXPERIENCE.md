---
name: E-Shop
status: final
sources:
  - {planning_artifacts}/ux-designs/ux-ecommerce-api-go-2026-09-04/DESIGN.md
updated: 2026-09-04
---

# E-Shop — Experience Spine

> Sustainable clothing ecommerce platform. Desktop-first responsive web app. Vue 3 + Tailwind CSS. `DESIGN.md` is the visual identity reference; this spine is the experience.

## Foundation

Multi-surface responsive web: mobile (320px+), tablet (768px+), desktop (1024px+). Desktop-first layout with mobile adaptation. No UI system named — uses custom Tailwind components inheriting from `DESIGN.md` tokens.

Vue 3 Composition API with Pinia stores. Vue Router for navigation. Axios for API calls.

## Information Architecture

| Surface | Route | Purpose |
|---|---|---|
| Home | `/` | Hero, featured products, trust signals, categories |
| Product Listing | `/products` | Filterable product grid with search and sort |
| Product Detail | `/products/:slug` | Full product info, variants, reviews, add to cart |
| Cart | `/cart` | Review items, adjust quantities, proceed to checkout |
| Checkout | `/checkout` | Shipping info, payment, order confirmation |
| Login | `/login` | User authentication |
| Register | `/register` | Account creation |
| Order History | `/orders` | List of past orders (authenticated) |
| Order Detail | `/orders/:id` | Single order details (authenticated) |
| Admin Dashboard | `/admin` | Stats, recent orders (admin only) |
| Admin Products | `/admin/products` | Product CRUD (admin only) |
| Admin Orders | `/admin/orders` | Order management (admin only) |
| Admin Users | `/admin/users` | User list (admin only) |

Navigation: Sticky header with logo, search, nav links, cart icon, auth controls. Footer with multi-column links.

## Voice and Tone

Microcopy. Brand voice lives in `DESIGN.md.Brand & Style`.

| Do | Don't |
|---|---|
| "Free shipping on all orders" | "Enjoy our complimentary shipping service!" |
| "Add to Cart" | "Add This Item To Your Shopping Cart" |
| "No products found" | "Sorry, we couldn't find any products matching your criteria" |
| "In Stock" | "This product is currently available for purchase" |
| "Loading..." | "Please wait while we fetch the data" |

Short, direct, helpful. No corporate jargon, no unnecessary politeness.

## Component Patterns

Behavioral. Visual specs live in `DESIGN.md.Components`.

| Component | Use | Behavioral rules |
|---|---|---|
| ProductCard | Home, Product Listing | Shows image, name, category, price, rating. Hover reveals quick-add overlay. Click navigates to detail. |
| Button | All surfaces | Primary for main CTAs, Secondary for cancel/back, Ghost for icon actions. Loading state disables and shows spinner. |
| Badge | Product cards, filters | Status pills for "New", "Sale", "Out of Stock", "Low Stock". Category tags on filters. |
| PriceDisplay | All product prices | Consistent `$XX.XX` format. Strikethrough for original price on sale items. |
| StarRating | Product cards, detail | 5-star display with review count. Interactive mode for review submission. |
| Skeleton | All loading states | Shimmer animation matching element shape. Replaces "Loading..." text. |
| Toast | Feedback | Bottom-right stack. Auto-dismiss success (3s). Manual dismiss for errors. |
| Modal | Confirmations, quick view | Overlay with escape/click-outside dismiss. Focus trap inside. |
| Pagination | Product listing | Numbered pages with prev/next. Active state highlighted. |

## State Patterns

| State | Surface | Treatment |
|---|---|---|
| Loading | All | Skeleton placeholders matching content shape |
| Empty cart | Cart | Illustration + "Your cart is empty" + CTA to browse |
| No products found | Product Listing | "No products found" + clear filters button |
| Out of stock | Product Detail | Disabled "Add to Cart" + "Out of Stock" badge |
| Low stock | Product Detail | "Only X left" warning badge |
| Added to cart | Product Detail | Success toast + cart icon bounce animation |
| Processing payment | Checkout | Button spinner + disabled form |
| Order complete | Checkout | Success message + order number + view order CTA |
| Form validation | Checkout, Login, Register | Inline error messages below inputs |
| Network error | All | Toast with retry option |

## Interaction Primitives

- **Add to Cart** — Click primary button. Success toast appears. Cart badge updates with count. No page redirect.
- **Quantity adjustment** — +/- buttons with min 1, max stock. Debounced API call on change.
- **Image gallery** — Thumbnail strip below main image. Click thumbnail to change main image. Keyboard arrow support.
- **Filter application** — Checkbox/price inputs with "Apply Filters" button. URL params update for shareability.
- **Search** — Input with Enter key trigger. Debounced autocomplete optional in future.
- **Mobile menu** — Hamburger icon toggles slide-down nav. Close on link click or outside tap.
- **Sticky elements** — Header sticks on scroll. Checkout order summary sticks on desktop.

## Accessibility Floor

Behavioral. Visual contrast lives in `DESIGN.md`.

- Keyboard navigation: all interactive elements focusable and operable via Tab/Enter/Space.
- Focus indicators visible on all focusable elements (minimum 2px outline).
- Alt text on all product images.
- Form inputs labeled with `<label>` elements (not just placeholders).
- Error messages associated with inputs via `aria-describedby`.
- Skip-to-content link for keyboard users.
- Color contrast: text meets WCAG AA (4.5:1 for normal text, 3:1 for large text).
- Reduce Motion: disable zoom/scale animations on product cards.

## Key Flows

### Flow 1 — Browse and discover (Sarah, Saturday morning, looking for a gift)

1. Sarah lands on homepage.
2. Hero section shows featured collection.
3. She scrolls to "Shop by Category" and clicks "Shirts".
4. Product listing loads with shirts filtered.
5. She adjusts price range filter to $50-$100.
6. She hovers over a product card — quick-add button appears.
7. She clicks the card to see full details.
8. **Climax:** She sees the product detail page with multiple images, size selector, and reviews — everything she needs to decide.

### Flow 2 — Add to cart and checkout (Sarah, ready to buy)

1. On product detail, Sarah selects size "M" from swatches.
2. She clicks "Add to Cart".
3. Toast appears: "Added to cart!" with "View Cart" link.
4. Cart badge on header updates to show "1".
5. She clicks the cart icon in header.
6. Cart page shows item with image, name, size, quantity controls, price.
7. She adjusts quantity to 2.
8. Order summary updates in real-time.
9. She clicks "Proceed to Checkout".
10. Checkout shows shipping form + order summary sidebar.
11. She fills shipping info, Stripe payment element loads.
12. She enters card details and clicks "Place Order".
13. Processing spinner appears, form disabled.
14. **Climax:** Success message appears with order number and "View Order" button — purchase complete.

### Flow 3 — Guest browsing (Mike, just looking, no account)

1. Mike browses products without logging in.
2. He adds items to cart (guest cart stored locally via session ID).
3. He clicks checkout — redirected to login/register.
4. He creates an account.
5. Guest cart merges with new account cart automatically.
6. He continues checkout seamlessly.

### Flow 4 — Review a purchase (Sarah, after receiving her order)

1. Sarah logs in and goes to Orders.
2. She sees her order history with status badges.
3. She clicks a delivered order.
4. Order detail shows items, shipping info, and "Write a Review" option.
5. She selects star rating and types her review.
6. She submits — review appears on the product page for others.

## Responsive & Platform

| Breakpoint | Layout | Navigation | Product Grid |
|---|---|---|---|
| Mobile (< 768px) | Single column | Hamburger menu | 1 column |
| Tablet (768px-1023px) | Two column where applicable | Inline nav | 2 columns |
| Desktop (1024px+) | Full layout with sidebar | Inline nav | 3-4 columns |

- Product listing sidebar collapses to horizontal filter bar on mobile.
- Cart table converts to stacked card layout on mobile.
- Checkout switches from 2-column to single column on mobile.
- Header search becomes expandable icon on mobile.
