# Frontend Pages

## Admin Menu Update

The sidebar in `AdminLayout.vue` needs two new items added to `menuItems`:

```js
// Add after UsersIcon definition
const DeadLettersIcon = { /* clipboard with alert/exclamation SVG */ }
const CategoriesIcon = { /* tag/label SVG */ }

// Updated menuItems array
const menuItems = [
  { path: '/admin', label: 'Dashboard', icon: DashboardIcon },
  { path: '/admin/products', label: 'Products', icon: ProductsIcon },
  { path: '/admin/orders', label: 'Orders', icon: OrdersIcon },
  { path: '/admin/users', label: 'Users', icon: UsersIcon },
  { path: '/admin/categories', label: 'Categories', icon: CategoriesIcon },
  { path: '/admin/dead-letters', label: 'Dead Letters', icon: DeadLettersIcon },
]
```

## Router Updates

Add to `router/index.js` admin children:

```js
{
  path: 'categories',
  name: 'AdminCategories',
  component: () => import('../views/admin/Categories.vue'),
},
{
  path: 'dead-letters',
  name: 'AdminDeadLetters',
  component: () => import('../views/admin/DeadLetters.vue'),
},
```

## DeadLetters.vue

**Purpose:** List failed background jobs with retry/delete actions.

**Layout:**
- Page header: "Dead Letters" title
- Table columns: Queue, Error, Created, Actions
- Each row has Retry (blue) and Delete (red) buttons
- Empty state: "No failed jobs" message

**API calls:**
- `GET /admin/dead-letters` on mount
- `POST /admin/dead-letters/:id` for retry
- `DELETE /admin/dead-letters/:id` for delete

**Patterns to follow:**
- Same table style as Orders.vue (white bg, rounded shadow, gray header row)
- Confirmation dialog on delete (use `confirm()`)
- After retry/delete, refresh the list

## Categories.vue

**Purpose:** CRUD management for product categories.

**Layout:**
- Page header: "Categories" title + "Add Category" button
- Table columns: Name, Product Count, Actions
- Add/Edit modal with name input field
- Delete confirmation dialog

**API calls:**
- `GET /admin/categories` on mount
- `POST /admin/categories` to create
- `PUT /admin/categories/:id` to update
- `DELETE /admin/categories/:id` to delete

**Patterns to follow:**
- Same modal pattern as Products.vue (fixed overlay, form with cancel/save)
- Same table pattern as existing pages

## Products.vue Updates

**Changes to Add/Edit modal:**
- Add image upload section below existing fields
- File input with accept="image/*"
- Preview thumbnail of selected image
- Display existing product images (if any) with delete button
- On form submit: upload image first via `POST /images/upload`, then include image URL in product create/update

**New fields in form:**
- Category: change from `<input>` to `<select>` dropdown
- Options populated from `GET /admin/categories`
- Image upload area with preview

**API calls to add:**
- `GET /admin/categories` on mount (for category dropdown)
- `POST /images/upload` (multipart) when saving with new image
- `DELETE /images/delete` when removing an image

## Users.vue Updates

**Changes to table:**
- Add search input above table (filters by name or email, client-side)
- Add Actions column with Edit and Delete buttons
- Edit opens modal with role dropdown (user/admin)
- Delete shows confirmation dialog

**New modal:**
- Role dropdown with current role pre-selected
- Save button calls `PUT /admin/users/:id`

**API calls to add:**
- `PUT /admin/users/:id` to update role
- `DELETE /admin/users/:id` to delete user

**Validation:**
- Hide Delete button if row is the current admin user
- Show admin badge (purple) more prominently

## Products.vue Updates - Image Constraints

- File input accepts: `image/jpeg,image/png,image/webp`
- Client-side validation: max 5MB before upload
- Show error toast if file exceeds limit or wrong format
