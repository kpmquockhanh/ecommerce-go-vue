---
name: E-Shop Checkout Redesign
status: final
sources:
  - {planning_artifacts}/ux-designs/ux-ecommerce-api-go-checkout-2026-09-05/DESIGN.md
updated: 2026-09-05
---

# E-Shop — Checkout Experience Spine

> Checkout page redesign focused on Stripe trust signals, step-based flow, and conversion optimization through trust building. Vue 3 + Ant Design Vue. `DESIGN.md` is the visual identity reference; this spine is the experience.

## Foundation

Desktop-first responsive web: mobile (320px+), tablet (768px+), desktop (1024px+). Desktop-first layout with mobile adaptation. Uses Ant Design Vue components inheriting from `DESIGN.md` tokens.

Vue 3 Composition API with Pinia stores. Vue Router for navigation. Axios for API calls. Stripe Elements for payment processing.

## Information Architecture

| Surface | Route | Purpose |
|---|---|---|
| Step 1: Shipping | `/checkout/shipping` | Collect shipping address |
| Step 2: Payment | `/checkout/payment` | Stripe payment element + trust signals |
| Step 3: Review | `/checkout/review` | Full order summary with edit capabilities |
| Confirmation | `/orders/:id` | Success message + order details + recommendations |

Navigation: Sticky header with logo and "Secure Checkout" text. Progress bar below header. Order summary sidebar on desktop.

## Voice and Tone

Microcopy. Brand voice lives in `DESIGN.md.Brand & Style`.

| Do | Don't |
|---|---|
| "Secure Checkout" | "Complete Your Purchase Transaction" |
| "Continue to Payment" | "Proceed to Payment Processing" |
| "Place Order — $570.00" | "Submit Your Order for Processing" |
| "Money-back guarantee" | "Our Comprehensive Return Policy" |
| "47 customers today" | "Join Thousands of Satisfied Customers" |

Short, direct, helpful. No corporate jargon, no unnecessary politeness.

## Component Patterns

Behavioral. Visual specs live in `DESIGN.md.Components`.

| Component | Use | Behavioral rules |
|---|---|---|
| ProgressBar | All checkout steps | Shows current step (1-3) with labels. Completed steps show checkmark. Clicking completed step navigates back. |
| StripeTrustBadge | Payment step header | "Secured by Stripe" text with Stripe purple branding. Always visible during payment. |
| TrustBar | Below order summary | Horizontal strip with 4 trust signals: guarantee, SSL, social proof, Stripe. Wraps on mobile. |
| PaymentCard | Payment step | Stripe-bordered card with payment element + inline trust badges + lock icon CTA. |
| OrderSummary | All steps (sidebar) | Sticky sidebar with items, shipping summary, total, and trust bar. |
| ReviewCard | Review step | Full summary with inline edit buttons for shipping and payment. |
| ConfirmationPage | Post-checkout | Success icon, order number, email notice, delivery estimate, recommendations. |
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
| Step 1 active | Progress bar | Circle 1 filled green, circles 2-3 gray |
| Step 2 active | Progress bar | Circle 1 checkmark, circle 2 filled green, circle 3 gray |
| Step 3 active | Progress bar | Circles 1-2 checkmarks, circle 3 filled green |
| Empty cart | Cart | Illustration + "Your cart is empty" + CTA to browse |
| No products found | Product Listing | "No products found" + clear filters button |
| Out of stock | Product Detail | Disabled "Add to Cart" + "Out of Stock" badge |
| Low stock | Product Detail | "Only X left" warning badge |
| Added to cart | Product Detail | Success toast + cart icon bounce animation |
| Processing payment | Checkout | Button spinner + disabled form |
| Order complete | Confirmation | Success message + order number + view order CTA |
| Form validation | Checkout, Login, Register | Inline error messages below inputs |
| Network error | All | Toast with retry option |
| Payment error | Checkout | Error message below payment element + retry option |

## Interaction Primitives

- **Step navigation** — Click "Continue" to advance. Click "Back" to return. Progress bar updates. URL changes.
- **Inline editing** — Review step shows edit links. Clicking edit navigates to relevant step. Changes reflected on return.
- **Stripe payment** — Elements mount on step 2. Card validation happens in real-time. Submit triggers `confirmPayment`.
- **Trust signal interaction** — Non-interactive, display only. Clicking guarantee/return policy opens modal or navigates to policy page.
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
- Progress steps have aria-label for screen readers (e.g., "Step 1 of 3: Shipping").
- Trust signals are decorative, not interactive, so no keyboard interaction required.

## Key Flows

### Flow 1 — Complete checkout with trust signals (Sarah, ready to buy)

1. Sarah clicks "Proceed to Checkout" from cart.
2. Checkout loads Step 1: Shipping with progress bar showing step 1 active.
3. She fills shipping form with name, address, city, state, zip, country.
4. She clicks "Continue to Payment".
5. Progress bar updates: step 1 checkmark, step 2 active.
6. Step 2: Payment loads with Stripe payment element.
7. **Trust signals visible:** Stripe badge in payment card header, PCI/SSL/No data badges below element, lock icon on CTA.
8. She enters card details into Stripe element.
9. She clicks "Place Order — $570.00" with lock icon.
10. Processing spinner appears, form disabled.
11. Progress bar updates: step 1-2 checkmarks, step 3 active.
12. Step 3: Review loads with full summary.
13. She reviews shipping, payment (last 4 digits), and items.
14. She sees trust bar below summary: guarantee, SSL, social proof, Stripe.
15. **Climax:** She clicks "Place Order" with confidence from visible trust signals.
16. Confirmation page loads with success icon, order number, email notice.
17. She sees recommended products below.
18. Trust bar at bottom reinforces post-purchase confidence.

### Flow 2 — Edit during review (Sarah, wants to change shipping)

1. Sarah is on Step 3: Review.
2. She notices shipping address has a typo.
3. She clicks "Edit" next to Shipping Address.
4. Progress bar updates: step 3 inactive, step 1 active.
5. She's taken to Step 1: Shipping with pre-filled form.
6. She corrects the address.
7. She clicks "Continue to Payment" (skips payment since it's already valid).
8. She's taken to Step 3: Review with updated address.
9. **Climax:** She confirms with correct address and places order.

### Flow 3 — Mobile checkout (Mike, on phone)

1. Mike opens checkout on mobile.
2. Progress bar shows at top, full-width single step.
3. Step 1: Shipping shows as full-width card.
4. He fills form, taps "Continue to Payment".
5. Step 2: Payment shows Stripe element full-width.
6. Trust signals stack vertically below payment element.
7. He enters card details, taps "Place Order".
8. Step 3: Review shows stacked layout with edit links.
9. **Climax:** He confirms order on mobile with all trust signals visible.

### Flow 4 — Post-purchase (Sarah, after checkout)

1. Sarah sees confirmation page with success icon.
2. Order number displayed prominently.
3. Email confirmation notice with her email address.
4. Shipping address and estimated delivery shown.
5. "View Order Details" and "Continue Shopping" buttons.
6. Recommended products section with 3 product cards.
7. Trust bar at bottom: guarantee, Stripe, returns, community size.
8. **Climax:** She feels confident about her purchase and browses recommendations.

## Responsive & Platform

| Breakpoint | Layout | Navigation | Progress Bar |
|---|---|---|---|
| Mobile (< 768px) | Single column | Hamburger menu | Full-width, smaller circles |
| Tablet (768px-1023px) | Two column where applicable | Inline nav | Horizontal, medium circles |
| Desktop (1024px+) | Full layout with sidebar | Inline nav | Horizontal, standard circles |

- Checkout switches from 2-column (form + summary) to single column on mobile.
- Progress bar remains visible at top on all screen sizes.
- Order summary collapses below form on mobile.
- Trust bar stacks vertically on mobile.
- Payment card remains full-width on all sizes.
- Review step stacks sections vertically on mobile.

## Checkout-Specific Patterns

### Step Transitions

- **Forward** — Click "Continue" → Validate current step → Animate to next step → Update progress bar
- **Backward** — Click "Back" → Animate to previous step → Update progress bar
- **From Review** — Click "Edit" → Navigate to relevant step → Update progress bar → Return to review on save

### Payment Security Flow

1. Stripe script loads async on checkout init
2. Stripe instance created with publishable key
3. Payment Intent created via API
4. Stripe Elements mount with client secret
5. User enters card details (validated by Stripe in real-time)
6. User clicks "Place Order" with lock icon
7. API confirms order → Stripe confirms payment
8. Success → redirect to confirmation
9. Error → display error message, allow retry

### Trust Signal Timing

- **Step 1 (Shipping)** — Trust bar visible in sidebar: guarantee, SSL, social proof
- **Step 2 (Payment)** — Trust bar + inline payment badges (PCI, SSL, No data) + Stripe badge
- **Step 3 (Review)** — Trust bar with all signals including Stripe
- **Confirmation** — Trust bar at bottom with all signals including community size

### Error Handling

- **Payment failed** — Show error below payment element, allow retry without losing data
- **Network error** — Toast with retry option, preserve form state
- **Validation error** — Inline error below specific input, highlight field
- **Session expired** — Redirect to cart with message, preserve items
