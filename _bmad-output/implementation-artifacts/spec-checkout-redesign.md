---
title: 'Checkout Page Stripe Trust Redesign'
type: 'feature'
created: '2026-09-05'
status: 'done'
baseline_commit: '190f3c11f0aa850721dcebcfa8d789b23dec75e6'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '_bmad-output/planning-artifacts/ux-designs/ux-ecommerce-api-go-checkout-2026-09-05/DESIGN.md'
  - '_bmad-output/planning-artifacts/ux-designs/ux-ecommerce-api-go-checkout-2026-09-05/EXPERIENCE.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The current checkout page lacks trust signals and uses a single-page flow that doesn't guide users through the purchase process, leading to potential cart abandonment and reduced conversion rates.

**Approach:** Redesign the checkout page with a step-based flow (Shipping → Payment → Review), add subtle Stripe branding and trust signals (money-back guarantee, SSL, social proof), and optimize the post-purchase experience with recommendations and trust reinforcement.

</frozen-after-approval>

## Code Map

- `frontend/src/views/Checkout.vue` — Main checkout page to be completely rewritten with step-based flow
- `frontend/src/views/Cart.vue` — Contains "Secure checkout powered by Stripe" badge pattern to reference
- `frontend/src/style.css` — Global styles; add new checkout-specific classes for steps, trust bar, payment card
- `frontend/src/App.vue` — Theme config; no changes needed
- `frontend/src/stores/cart.js` — Cart store; no changes needed
- `frontend/src/lib/api.js` — API client; no changes needed
- `frontend/src/lib/utils.js` — Utilities; no changes needed
- `internal/handlers/order.go` — Backend checkout handlers; DO NOT MODIFY
- `internal/repositories/checkout_session.go` — Backend session management; DO NOT MODIFY

## Tasks & Acceptance

**Execution:**
- [ ] `frontend/src/style.css` — Add new CSS classes for checkout steps, trust bar, payment card, review sections, confirmation page
- [ ] `frontend/src/views/Checkout.vue` — Rewrite with step-based flow, progress bar, Stripe trust badges, trust signals, review step, confirmation page

**Acceptance Criteria:**
- Given user is on checkout, when page loads, then progress bar shows step 1 (Shipping) active
- Given user completes shipping form, when clicks "Continue to Payment", then progress bar updates to step 2 and payment section loads
- Given user is on payment step, when payment element loads, then "Secured by Stripe" badge and security badges are visible
- Given user clicks "Place Order", when processing, then button shows lock icon and "Encrypted & secure payment" text
- Given order is confirmed, when redirected to review, then full summary shows with edit links for shipping and payment
- Given user clicks "Edit" on review step, when navigating back, then relevant step loads with pre-filled data
- Given order is placed successfully, when confirmation page loads, then success icon, order number, email notice, and recommendations are visible
- Given user is on mobile, when viewing checkout, then steps stack vertically and trust bar wraps appropriately
- Given all trust signals, when visible, then money-back guarantee, SSL, social proof, and Stripe badge are present

## Implementation Notes

### Step 1: Add CSS Classes to style.css

Add new classes for:
- `.checkout-progress` — Progress bar container
- `.progress-step` — Individual step with circle and label
- `.progress-circle` — Numbered circle (active/completed/pending states)
- `.progress-line` — Connecting line between steps
- `.payment-card` — Stripe-bordered payment section
- `.trust-bar` — Trust signals container
- `.trust-item` — Individual trust signal with icon and text
- `.review-section` — Review step sections with edit links
- `.confirmation-card` — Success page layout
- `.recommended-grid` — Product recommendations grid

### Step 2: Rewrite Checkout.vue

1. Add reactive state for current step (1-3)
2. Create progress bar component with numbered circles and labels
3. Implement step navigation (forward/back)
4. Create shipping step with form
5. Create payment step with Stripe Elements and trust badges
6. Create review step with full summary and edit links
7. Create confirmation page with success icon, order details, recommendations
8. Add trust bar component below order summary
9. Implement responsive design for mobile

### Key Implementation Details

- Keep existing Stripe Elements mounting logic
- Keep existing API calls to /checkout/init and /checkout/confirm
- Keep existing idempotency key generation
- Add progress bar above checkout grid
- Wrap payment element in styled card with Stripe badge
- Add trust bar below order summary
- Implement step transitions with state management
- Add confirmation page as separate view state
- Add recommended products section (static data for now)

## Verification

**Commands:**
- `cd frontend && npm run build` — Expected: Build succeeds without errors
- `cd frontend && npm run lint` — Expected: No linting errors

**Manual checks:**
- Open checkout page in browser
- Verify progress bar shows step 1 active
- Fill shipping form, click Continue
- Verify progress bar updates to step 2
- Verify Stripe payment element loads with "Secured by Stripe" badge
- Verify trust bar shows below order summary
- Click Place Order
- Verify confirmation page shows with order number and recommendations
- Test on mobile viewport
- Verify responsive layout works correctly

## Review Triage Log

| Finding | Verdict | Evidence |
|---------|---------|----------|
| Stripe.js script not removed on unmount | patch-applied | Added onUnmounted hook to remove script tag; prevents duplicate scripts on repeated navigation |
| No error handler on Stripe.js script load | patch-applied | Added stripeScript.onerror handler with user-facing error message |
| api.get('/config') error unhandled | patch-applied | Wrapped config fetch in try-catch with error message |
| cart.total equals 0 silent early return | patch-applied | Added explicit error message when cart is empty |
| confirmData.order_id undefined | patch-applied | Added validation for order_id before proceeding with Stripe confirmation |
| No format validation on shipping fields | defer | ZIP pattern, state maxlength not validated; low risk for initial release |
| CORS issues (Allow-Origin: *, missing PUT/DELETE) | defer | Pre-existing backend issues not caused by checkout redesign |
| Database seed on every startup | defer | Pre-existing backend issue not caused by checkout redesign |
| No rate limiting on auth routes | defer | Pre-existing backend issue not caused by checkout redesign |
| Hardcoded recommended products | defer | Acknowledged in spec as "static data for now"; needs real API integration later |
