---
id: SPEC-vue-ant-design
companions:
  - migration-guide.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Vue Ant Design v4 Integration

## Why

The ecommerce frontend currently uses Tailwind CSS with custom inline styles, requiring manual component styling and no built-in enterprise UI patterns. Adopting Ant Design Vue v4 replaces ad-hoc styling with a mature, enterprise-grade component library using CSS-in-JS theming (no Less pipeline needed), token-based design customization, and pre-built components for forms, tables, modals, notifications, and layout — directly accelerating the ecommerce features (product catalog, cart, checkout, admin dashboard) defined in the existing frontend spec. This is a **pain to solve**: inconsistent custom styling slows development and lacks enterprise UI patterns needed for the admin dashboard and complex forms.

## Capabilities

- **CAP-1: Vite Auto-Import Configuration**
  - **intent:** The build system automatically resolves and imports Ant Design Vue components on-demand without manual imports.
  - **success:** `vite.config.js` includes `unplugin-vue-components` with `AntDesignVueResolver({ importStyle: false })`; any `<a-button>` or `<a-table>` used in templates works without explicit import statements.

- **CAP-2: Plugin Registration**
  - **intent:** The application entry point registers Ant Design Vue globally.
  - **success:** `main.js` imports `ant-design-vue` and calls `app.use(Antd)`; all Ant Design components are available app-wide without per-component registration. No separate CSS import needed (CSS-in-JS handles styles).

- **CAP-3: Layout Migration**
  - **intent:** App shell (header, navigation, footer) uses Ant Design layout components instead of custom HTML/Tailwind.
  - **success:** `App.vue` uses `<a-config-provider>` wrapping `<a-layout>`, `<a-layout-header>`, `<a-layout-content>`, `<a-menu>`, `<a-input-search>`, `<a-badge>` for the navigation bar; responsive behavior preserved.

- **CAP-4: Component Migration**
  - **intent:** All existing views and components use Ant Design Vue components for forms, tables, cards, modals, and notifications.
  - **success:** Product listing uses `<a-table>` or `<a-card>` grid; forms use `<a-form>` with validation; cart uses `<a-list>`; admin uses `<a-modal>` for CRUD dialogs; toast notifications use `message`/`notification` API (no manual style import needed in v4).

- **CAP-5: Theme Customization**
  - **intent:** The brand color scheme is applied globally via Ant Design v4's token-based ConfigProvider.
  - **success:** `ConfigProvider` wraps the app with `theme.token.colorPrimary` set to brand color (#2D6A4F); all Ant Design components reflect the custom theme; dark/light algorithm switchable if needed.

## Constraints

- Tailwind CSS dependency will be removed; all styling moves to Ant Design Vue components and design tokens.
- v4 uses CSS-in-JS (emotion-based) — no Less/PostCSS variable overrides. Theme is purely token-driven via `ConfigProvider`.

## Non-goals

- Custom theme from Less/SCSS overrides — v4 uses token-based theming only.
- Migrating to Ant Design Vue Pro (admin template) — build on the base library.
- Replacing Pinia stores or Vue Router — state/routing layer unchanged.
- Server-side rendering (SSR) — SPA approach maintained.

## Success signal

The app renders with Ant Design v4 components throughout; `npm run build` succeeds with on-demand CSS-in-JS imports; brand colors are reflected in all components; no Tailwind classes remain in source files.

## Assumptions

- Ant Design Vue v4.x is compatible with Vue 3.5.x and Vite 8.x in the current project.
- `unplugin-vue-components` and `AntDesignVueResolver` from the same package work with v4.

## Resolved Questions

- v4 uses `importStyle: false` in the resolver for CSS-in-JS mode — eliminates manual style imports.
- Theme customization via `ConfigProvider` `theme` prop with token objects, no Less loader needed.
