# Frontend Pages

Vue 3 + Pinia frontend specification. All pages use the Composition API with `<script setup>`.

## Tech Stack

- **Framework:** Vue 3 (Vite build)
- **State:** Pinia
- **Router:** Vue Router 4
- **HTTP:** Axios
- **CSS:** Tailwind CSS
- **Stripe:** @stripe/stripe-js + @stripe/vue-stripe-js

## Layout

### App Shell
- **Navbar:** Logo, search bar, cart icon (with count badge), user menu (login/register or profile/logout)
- **Mobile nav:** Hamburger menu with drawer
- **Footer:** Links, copyright

### Protected Route Guard
- Unauthenticated users redirected to `/login` on protected routes
- Admin routes check `user.role === 'admin'`

## Pages

### Home `/`
- Hero section with CTA
- Featured products grid (latest 8)
- Category links

### Product Catalog `/products`
- Filter sidebar: category checkboxes, price range slider
- Sort dropdown: Price Low-High, Price High-Low, Newest, Name A-Z
- Search input in header
- Product grid (responsive: 1 col mobile, 2 col tablet, 3-4 col desktop)
- Pagination controls

### Product Detail `/products/:id`
- Image gallery/carousel
- Product name, price, description
- Variant selector (size/color dropdowns or swatches)
- Quantity input + Add to Cart button
- Reviews section with average rating, individual reviews
- Review form (logged-in users only)

### Cart `/cart`
- Table: product image, name, variant, unit price, quantity adjuster (+/-), line total, remove button
- Cart summary: subtotal, shipping estimate, total
- Proceed to Checkout button
- Empty cart state with link to catalog

### Checkout `/checkout`
- Two-column layout: form left, order summary right
- Address form: first name, last name, address 1, address 2, city, state (dropdown), zip, country
- Stripe Payment Element embedded
- Place Order button (disabled until payment element confirms)
- Order confirmation page after success

### Login `/login`
- Email + password form
- Link to register
- Error display for invalid credentials

### Register `/register`
- Email, password, confirm password, first name, last name
- Client-side validation (email format, password strength, match)
- Link to login
- Auto-login after registration

### Order History `/orders`
- Table: order ID, date, status badge, total, items count, view link
- Empty state if no orders

### Order Detail `/orders/:id`
- Order info: ID, date, status badge, total
- Items list: product name, variant, quantity, price, subtotal
- Shipping address block
- Payment status

### Admin Dashboard `/admin`
- Stats cards: total orders, total revenue, total users, total products
- Recent orders table (last 10)

### Admin Products `/admin/products`
- Table: image thumbnail, name, category, price, stock, actions
- Create product button → modal or separate form
- Edit button → pre-filled form
- Delete button with confirmation

### Admin Orders `/admin/orders`
- Table: order ID, customer name, date, status (editable dropdown), total, view link
- Filter by status

### Admin Users `/admin/users`
- Table: name, email, role, join date, order count, view link

## Pinia Stores

### `authStore`
```
State: user (object|null), token (string|null), loading (boolean)
Actions: login(email, password), register(data), logout(), fetchProfile()
Getters: isAuthenticated, isAdmin
```

### `productStore`
```
State: products (array), currentProduct (object|null), pagination (object), filters (object), loading (boolean)
Actions: fetchProducts(params), fetchProduct(id), search(query)
```

### `cartStore`
```
State: items (array), loading (boolean), sessionId (string|null)
Actions: fetchCart(), addItem(productId, variantId, qty), updateItem(id, qty), removeItem(id), mergeGuestCart()
Getters: itemCount, total, isGuest
Notes: Guest carts use sessionId (UUID in localStorage). On login, guest cart merges into user cart.
```

### `orderStore`
```
State: orders (array), currentOrder (object|null), loading (boolean)
Actions: createOrder(shippingAddress), fetchOrders(), fetchOrder(id)
```

### `adminStore`
```
State: products (array), orders (array), users (array), stats (object)
Actions: fetchStats(), CRUD products, updateOrderStatus(), fetchUsers()
```
