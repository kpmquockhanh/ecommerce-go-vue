# API Endpoints

## Existing Endpoints (No Changes)

### Dead Letters
| Method | Endpoint | Handler | Notes |
|--------|----------|---------|-------|
| GET | `/api/admin/dead-letters` | `deadLetterHandler.ListDeadLetters` | Returns paginated list |
| POST | `/api/admin/dead-letters/{id}/retry` | `deadLetterHandler.RetryDeadLetter` | Re-enqueues job |
| DELETE | `/api/admin/dead-letters/{id}` | `deadLetterHandler.DeleteDeadLetter` | Removes entry |

### Images
| Method | Endpoint | Handler | Notes |
|--------|----------|---------|-------|
| POST | `/api/images/upload` | `imageHandler.UploadImage` | Multipart, returns presigned URL. Max 5MB, formats: jpg, png, webp |
| DELETE | `/api/images/delete` | `imageHandler.DeleteImage` | JSON body with key |

## New Endpoints Required

### User Management
| Method | Endpoint | Handler | Request Body | Response |
|--------|----------|---------|--------------|----------|
| PUT | `/api/admin/users/{id}` | `adminHandler.UpdateUserRole` | `{ "role": "admin" \| "user" }` | Updated user object |
| DELETE | `/api/admin/users/{id}` | `adminHandler.DeleteUser` | None | `{ "message": "User deleted" }` |

**Constraints:**
- Admin cannot delete their own account (return 400).
- Admin cannot demote themselves (return 400).
- Validate role is one of: `admin`, `user`.

### Category Management
| Method | Endpoint | Handler | Request Body | Response |
|--------|----------|---------|--------------|----------|
| GET | `/api/admin/categories` | `categoryHandler.ListCategories` | None | `{ "categories": [...] }` |
| POST | `/api/admin/categories` | `categoryHandler.CreateCategory` | `{ "name": "Electronics" }` | Created category |
| PUT | `/api/admin/categories/{id}` | `categoryHandler.UpdateCategory` | `{ "name": "New Name" }` | Updated category |
| DELETE | `/api/admin/categories/{id}` | `categoryHandler.DeleteCategory` | None | `{ "message": "Category deleted" }` |

**Constraints:**
- Category name must be unique (return 409 on conflict).
- GET endpoint includes `product_count` per category.
- DELETE behavior: If any product has only this category, return 409 with `{ "error": "Category is sole category for N products", "product_count": N }`. Otherwise, remove the category from all products that have it, then delete the category.

### Product Image Association
| Method | Endpoint | Handler | Request Body | Response |
|--------|----------|---------|--------------|----------|
| POST | `/api/admin/products/{id}/images` | `productHandler.AddProductImage` | `{ "image_url": "...", "alt_text": "..." }` | Updated product |
| DELETE | `/api/admin/products/{id}/images/{image_id}` | `productHandler.RemoveProductImage` | None | `{ "message": "Image removed" }` |

**Note:** If products already have an `image_url` field, this may simplify to just updating that field. Check existing schema.
