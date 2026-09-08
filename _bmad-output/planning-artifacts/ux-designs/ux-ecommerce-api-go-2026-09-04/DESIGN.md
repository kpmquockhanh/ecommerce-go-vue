---
name: E-Shop
description: Modern sustainable clothing ecommerce. Clean, trustworthy, warm. Product-first design that lets merchandise speak.
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
  surface-base-dark: '#121212'
  surface-raised-dark: '#1E1E1E'
  ink-primary-dark: '#F0F0F0'
  ink-secondary-dark: '#A0A0A0'
  accent-dark: '#52B788'
  border-hairline-dark: '#333333'
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

E-Shop is a sustainable clothing ecommerce platform. The design language is clean, warm, and product-focused — letting the merchandise be the hero while building trust through consistent, professional patterns.

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

### Product Card

The centerpiece of the shopping experience. Designed to be visually rich while staying compact.

- **Image** — 1:1 aspect ratio, `object-cover`. Hover: subtle zoom (1.05 scale) with overflow hidden.
- **Quick-add overlay** — Slides up from bottom on hover. Semi-transparent dark background with white "Add to Cart" button.
- **Content** — Product name (heading-sm, ink-primary), category (meta, ink-secondary), price (price, ink-primary).
- **Rating** — Star rating inline below category, only if reviews exist.
- **Badge** — Optional top-left: "New", "Sale", or "Low Stock" pill badges.

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
