---
id: SPEC-admin-features
companions:
  - api-endpoints.md
  - frontend-pages.md
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate.

# Admin Panel Missing Features

## Why

The admin panel has backend APIs implemented for Dead Letters management, image upload/delete, and partial user management, but the frontend only exposes Dashboard, Products, Orders, and Users (read-only). Four features are unreachable from the admin UI: failed job management, image handling in product forms, user role editing/deletion, and category management. This gap forces admins to use direct API calls or database access for routine operations.

## Capabilities

- **CAP-1: Dead Letters Management Page**
  - **intent:** Admin can view a list of failed background jobs, see error details, retry a failed job, or delete it from the queue.
  - **success:** Admin navigates to Dead Letters page, sees a table of failed jobs with queue name, error message, and date. Clicking Retry re-enqueues the job and removes it from the list. Clicking Delete removes it without retry. Page shows empty state when no dead letters exist.

- **CAP-2: Image Upload in Product Forms**
  - **intent:** Admin can upload product images when creating or editing a product, and delete previously uploaded images.
  - **success:** The Add/Edit Product modal includes an image upload area. Admin can select a file, see a preview, and the image uploads on form submit. Uploaded images display in the product table. Admin can delete an image from the product edit form.

- **CAP-3: User Management CRUD**
  - **intent:** Admin can search for users, change a user's role (user/admin), and delete a user account.
  - **success:** The Users page includes a search input that filters by name or email. The Users table includes an Edit button per row opening a modal with a role dropdown (user/admin). Saving updates the role via API. A Delete button removes the user after confirmation. The admin cannot delete their own account.

- **CAP-4: Category Management**
  - **intent:** Admin can create, edit, and delete product categories, and assign categories to products from a predefined list.
  - **success:** A Categories page lists all categories with name and product count. Admin can add a category via a form, edit its name, or delete it. If a category has products with only that one category, deletion is blocked. If products have multiple categories, the deleted category is removed from those products. The Products form category field becomes a dropdown populated from the categories list instead of free-text input.

## Constraints

- All new frontend components must follow existing patterns: Vue 3 Composition API with `<script setup>`, Tailwind CSS utility classes, no scoped styles.
- All new admin backend routes must use `middleware.AdminMiddleware` for authorization.
- Frontend API calls must use the existing axios instance from `src/lib/api.js`.
- Dead Letters and Image backend APIs already exist; no changes to those handlers.
- Category management requires a new `categories` table and CRUD endpoints.

## Non-goals

- Bulk operations on users (mass role change, mass delete).
- Image cropping, resizing, or multi-image reordering.
- Category hierarchy/nesting (flat categories only).
- Real-time dead letter count badges in the sidebar.
- User activity logs or audit trails.

## Success signal

- Admin can navigate to Dead Letters, retry or delete a failed job without leaving the UI.
- Admin can upload an image while creating a product and see it displayed.
- Admin can promote a user to admin or delete a user from the Users page.
- Admin can manage categories from a dedicated page, and product forms use a category dropdown.
- All new pages are accessible from the admin sidebar navigation.

## Assumptions

- The existing S3 storage client (`storage.Client`, `storage.UploadImage`, `storage.DeleteImage`) can be reused without modification for image management.
- Categories should become a database entity (`categories` table) rather than remaining free-text on the products table.
- The existing `dead_letter` handler functions (ListDeadLetters, RetryDeadLetter, DeleteDeadLetter) are sufficient and need no backend changes.
- Image uploads are constrained to max 5MB, formats jpg/png/webp.
