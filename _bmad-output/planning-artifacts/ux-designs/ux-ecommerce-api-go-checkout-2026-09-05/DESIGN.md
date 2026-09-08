---
name: E-Shop Checkout Redesign
description: Checkout page redesign focused on Stripe trust signals, step-based flow, and conversion optimization through trust building.
status: final
updated: 2026-09-05
colors:
  surface-base: '#F8F6F3'
  surface-raised: '#FFFFFF'
  surface-sunken: '#F0EDE8'
  ink-primary: '#1A1A1A'
  ink-secondary: '#6B6B6B'
  ink-disabled: '#B0B0B0'
  accent: '#2D6A4F'
  accent-hover: '#1B4332'
  accent-light: '#D8F3DC'
  accent-muted: '#95D5B2'
  danger: '#C1121F'
  danger-light: '#FFE5E5'
  warning: '#E09F3E'
  warning-light: '#FFF3CD'
  success: '#2D6A4F'
  success-light: '#D8F3DC'
  border-hairline: '#E8E4DD'
  border-default: '#D4D0CB'
  overlay: 'rgba(0, 0, 0, 0.5)'
  stripe-purple: '#635BFF'
  trust-green: '#2D6A4F'
  trust-gray: '#6B6B6B'
typography:
  font-family: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif'
  heading-xl: '2.5rem / 1.2 font-bold'
  heading-lg: '2rem / 1.25 font-bold'
  heading-md: '1.5rem / 1.3 font-semibold'
  heading-sm: '1.125rem / 1.4 font-semibold'
  body-lg: '1.125rem / 1.6'
  body: '1rem / 1.6'
  body-sm: '0.875rem / 1.5'
  meta: '0.75rem / 1.4'
  price: '1.5rem / 1.2 font-bold'
  price-lg: '2rem / 1.2 font-bold'
rounded:
  sm: 6px
  md: 8px
  lg: 12px
  xl: 16px
  full: 9999px
spacing:
  '1': 4px
  '2': 8px
  '3': 12px
  '4': 16px
  '5': 24px
  '6': 32px
  '7': 48px
  '8': 64px
  '9': 96px
elevation:
  none: 'none'
  sm: '0 1px 2px rgba(0,0,0,0.06)'
  md: '0 4px 6px -1px rgba(0,0,0,0.07), 0 2px 4px -1px rgba(0,0,0,0.04)'
  lg: '0 10px 15px -3px rgba(0,0,0,0.08), 0 4px 6px -2px rgba(0,0,0,0.03)'
  xl: '0 20px 25px -5px rgba(0,0,0,0.08), 0 10px 10px -5px rgba(0,0,0,0.02)'
---

## Brand & Style

E-Shop is a sustainable clothing ecommerce platform. The checkout redesign maintains the clean, warm, product-focused design language while adding strategic trust signals through Stripe branding and social proof elements.

The aesthetic draws from modern direct-to-consumer brands: generous whitespace, warm neutrals that feel approachable rather than clinical, and a single forest green accent that signals eco-consciousness without being preachy. Photography does the heavy lifting; the UI steps back.

## Colors

The palette is warm and grounded — designed to feel trustworthy and let product imagery pop.

- **Surface Base (`#F8F6F3`)** — The primary canvas. A warm off-white that reduces eye strain and feels more inviting than pure white.
- **Surface Raised (`#FFFFFF`)** — Cards, modals, and interactive surfaces. Clean white for contrast against the warm base.
- **Surface Sunken (`#F0EDE8`)** — Subtle recessed areas, input backgrounds, and dividers.
- **Ink Primary (`#1A1A1A`)** — Primary text. Near-black but softer than true black for readability.
- **Ink Secondary (`#6B6B6B`)** — Supporting text, labels, placeholders.
- **Accent (`#2D6A4F`)** — Forest green. The only chromatic color in the primary UI. Signals action, trust, and sustainability.
- **Accent Hover (`#1B4332`)** — Darker green for hover/active states.
- **Danger (`#C1121F`)** — Error states, destructive actions, out-of-stock indicators.
- **Warning (`#E09F3E`)** — Caution states, low stock alerts.
- **Border Hairline (`#E8E4DD`)** — Lightest separation between elements.
- **Border Default (`#D4D0CB`)** — Input borders, card borders, table dividers.
- **Stripe Purple (`#635BFF`)** — Stripe branding accent. Used sparingly for "Powered by Stripe" text.
- **Trust Green (`#2D6A4F`)** — Primary trust signals (guarantee, security badges).
- **Trust Gray (`#6B6B6B`)** — Secondary trust signals (SSL, social proof).

Avoid: blue as a primary color (generic), gradients on UI elements (distracts from product imagery), neon or saturated fills.

## Typography

Inter is the primary typeface — clean, legible, and modern. The system font stack provides fallback.

- **Headings** — Bold weight, tight letter-spacing. Used sparingly: page titles, section headers, product names.
- **Body** — Regular weight, generous line-height (1.6). The workhorse for descriptions, content, and forms.
- **Meta** — Small caps or regular weight at small sizes for labels, timestamps, and secondary info.
- **Price** — Bold weight, slightly larger than surrounding text. Prices are first-class citizens.

Scale: 12px (meta) → 14px (body-sm) → 16px (body) → 18px (body-lg) → 20px (heading-sm) → 24px (heading-md) → 32px (heading-lg) → 40px (heading-xl).

## Layout & Spacing

Scale: 4 / 8 / 12 / 16 / 24 / 32 / 48 / 64 / 96 px. Consistent vertical rhythm throughout.

- **Page container** — `max-w-7xl mx-auto px-4 sm:px-6 lg:px-8`
- **Section spacing** — 48px-96px between major sections
- **Card padding** — 16px-24px
- **Component spacing** — 8px-16px between related elements

Responsive breakpoints: mobile-first. Single column on mobile, 2-column at md (768px), 3-column at lg (1024px), 4-column at xl (1280px).

## Elevation & Depth

A layered shadow system creates visual hierarchy without heavy borders.

- **None** — Flat elements, inline content
- **SM (`shadow-sm`)** — Subtle lift for cards at rest
- **MD (`shadow-md`)** — Cards on hover, dropdowns
- **LG (`shadow-lg`)** — Modals, popovers
- **XL (`shadow-xl`)** — Floating action elements

Shadows are warm-tinted (slightly brown) to match the warm palette. Never use colored shadows.

## Shapes

- **`rounded-sm` (6px)** — Small elements: badges, tags, small buttons
- **`rounded-md` (8px)** — Inputs, standard buttons
- **`rounded-lg` (12px)** — Cards, modals, dropdowns
- **`rounded-xl` (16px)** — Large cards, hero sections
- **`rounded-full`** — Avatars, circular buttons, pills

Imagery inherits container corners. Product images use `rounded-lg` on cards, `rounded-xl` on detail pages.

## Components

### Checkout Progress Bar

Step-based progress indicator showing current position in checkout flow.

- **Container** — Flexbox centered, horizontal layout
- **Step Circle** — 32px diameter circle with number/icon
  - Active: `{accent}` background, white text
  - Completed: `{accent}` background, white checkmark icon
  - Pending: `{border-hairline}` background, `{ink-secondary}` text
- **Step Label** — `body-sm` font, bold when active
- **Step Line** — 2px height, 80px width, `{border-hairline}` (pending) or `{accent}` (completed)

### Stripe Trust Badge

Subtle "Secured by Stripe" indicator near payment elements.

- **Layout** — Inline flex with gap, `body-sm` text
- **Text** — "Secured by" in `{ink-secondary}`, "Stripe" in `{stripe-purple}` bold
- **Placement** — Payment card header, payment section header
- **Size** — Small, non-intrusive

### Trust Signal Bar

Horizontal strip of trust indicators below order summary.

- **Container** — `{surface-sunken}` background, `rounded-md`, padding 16px
- **Layout** — Flexbox, space-between, wraps on mobile
- **Trust Item** — Flex column, centered, 12px gap
  - Icon: 16px font size
    - Primary (guarantee): `{trust-green}` color
    - Secondary (SSL, social proof): `{trust-gray}` color
  - Label: `meta` font, `{ink-secondary}` color
- **Spacing** — 16px gap between items, flex-wrap for mobile

### Payment Card

Enhanced card for Stripe payment element with trust indicators.

- **Container** — `{accent}` border (2px), `rounded-lg`, padding 20px
- **Header** — Flex, space-between, payment title + Stripe badge
- **Element** — `{surface-sunken}` background, min-height 200px, centered placeholder
- **Inline Trust** — Row of 3 trust badges (PCI, SSL, No data stored) below element
- **CTA** — `{accent}` fill button with lock icon, "Place Order — $XXX.XX"
- **Lock Text** — "Encrypted & secure payment" below CTA, centered, `meta` font

### Order Summary Card

Sticky sidebar with items and trust signals.

- **Position** — `sticky`, top 88px
- **Items** — Scrollable list, max-height 256px
- **Shipping Summary** — Collapsed view of shipping address with edit link
- **Divider** — `{border-hairline}` horizontal rule
- **Total** — `price-lg` font, bold
- **Trust Bar** — Trust signal strip below total

### Review Step Card

Full order summary with inline edit capabilities.

- **Section** — Each section (shipping, payment, items) with header + edit link
- **Shipping Card** — `{surface-sunken}` background, name + address
- **Payment Card** — `{surface-sunken}` background, card icon + last 4 digits + expiry
- **Items List** — Stacked rows with image, name, variant, quantity, price
- **Edit Links** — `{accent}` color, `body-sm` font, aligned right

### Confirmation Page

Success page with order details and recommendations.

- **Success Icon** — 80px circle, `{accent-light}` background, checkmark icon
- **Title** — `heading-lg`, centered
- **Order Number** — `{surface-sunken}` background, inline-block
- **Email Notice** — `{accent-light}` background, flex with icon
- **Details** — Left-aligned sections for shipping, delivery estimate
- **CTA Group** — Centered flex, primary + secondary buttons
- **Recommended** — 3-column grid of product cards below confirmation
- **Trust Bar** — Full-width trust signals at bottom

### Button

Three variants: Primary (accent fill), Secondary (border only), Ghost (text only).

- **Primary** — `bg-accent text-white hover:bg-accent-hover`. For main CTAs: Add to Cart, Checkout, Submit.
- **Secondary** — `border border-default text-ink-primary hover:bg-surface-sunken`. For secondary actions: Cancel, Back.
- **Ghost** — `text-ink-secondary hover:text-ink-primary hover:bg-surface-sunken`. For tertiary actions: Close, icon buttons.
- **Sizes** — sm (32px height), md (40px), lg (48px).
- **Loading state** — Spinner replaces text, button disabled.

### Badge / Tag

Pill-shaped labels for status and categories.

- **Default** — `bg-surface-sunken text-ink-secondary`
- **Accent** — `bg-accent-light text-accent` (for "New", active filters)
- **Danger** — `bg-danger-light text-danger` (for "Out of Stock", errors)
- **Warning** — `bg-warning-light text-warning` (for "Low Stock")
- **Size** — sm (text-xs, px-2 py-0.5), md (text-sm, px-3 py-1)

### PriceDisplay

Formats and displays prices consistently.

- **Standard** — `$XX.XX` in bold, ink-primary
- **Strikethrough** — Original price in `line-through text-ink-disabled` before sale price
- **Sale** — Sale price in `text-danger` with "Save XX%" badge
- **Size** — sm (body), md (heading-sm), lg (price-lg)

### StarRating

Visual star display with optional interactive mode.

- **Display** — 5 stars, filled in `text-yellow-400`, empty in `text-gray-200`
- **With count** — Stars + "(XX reviews)" text in ink-secondary
- **Interactive** — Clickable stars for review form, hover preview

### Skeleton

Loading placeholder with shimmer animation.

- **Pulse animation** — Subtle gradient shimmer from surface-sunken to surface-base
- **Shapes** — Rectangle (text), circle (avatar), square (image), custom (cards)

### Toast / Notification

Non-blocking feedback messages.

- **Position** — Bottom-right, stack upward
- **Variants** — Success (green), Error (red), Info (gray), Warning (yellow)
- **Auto-dismiss** — 3-5 seconds for success, manual dismiss for errors
- **Animation** — Slide in from right, fade out

### Modal

Overlay dialog for focused tasks.

- **Overlay** — `bg-overlay` with backdrop blur
- **Container** — `bg-surface-raised rounded-xl shadow-xl` centered
- **Header** — Title + close button
- **Body** — Scrollable content area
- **Footer** — Action buttons aligned right
- **Animation** — Fade in overlay + scale up container

### Pagination

Page navigation for product listings.

- **Style** — Numbered buttons with active state
- **Active** — `bg-accent text-white`
- **Inactive** — `bg-surface-sunken text-ink-secondary hover:bg-surface-sunken`
- **Prev/Next** — Arrow buttons, disabled at boundaries

## Do's and Don'ts

| Do | Don't |
|---|---|
| Let product images be the visual hero | Use gradients, patterns, or busy backgrounds |
| Use the green accent sparingly for actions | Color everything green to signal "eco" |
| Show prices prominently and clearly | Hide prices or use inconsistent formatting |
| Provide clear loading and empty states | Show "Loading..." text without visual feedback |
| Use consistent card patterns across pages | Invent new layouts for each page |
| Build trust with badges, reviews, and guarantees | Use dark patterns or fake urgency |
| Keep Stripe branding subtle and trustworthy | Overwhelm with too many security badges |
| Use numbered progress steps for clarity | Hide progress or use ambiguous indicators |
| Place trust signals near CTAs and payment | Bury trust elements at bottom of page |
| Use green for primary trust, gray for secondary | Mix colors inconsistently in trust section |
