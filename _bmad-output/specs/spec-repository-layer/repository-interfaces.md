# Repository Interfaces — Detailed Design

This companion defines every repository interface, its methods, and the query migration mapping.

## Package Structure

```
internal/
  repositories/
    user.go        — UserRepository interface + implementation
    product.go     — ProductRepository interface + implementation
    cart.go        — CartRepository interface + implementation
    order.go       — OrderRepository interface + implementation
    review.go      — ReviewRepository interface + implementation
    idempotency.go — IdempotencyRepository interface + implementation
    deadletter.go  — DeadLetterRepository interface + implementation
    stock.go       — StockRepository interface + implementation
```

## Interface Definitions

### UserRepository

```go
type UserRepository interface {
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    Create(ctx context.Context, email, passwordHash, firstName, lastName string) (*models.User, error)
    FindByEmail(ctx context.Context, email string) (*models.User, string, error) // returns user + password_hash
    FindByID(ctx context.Context, id int) (*models.User, error)
    UpdateProfile(ctx context.Context, id int, firstName, lastName string) error
    List(ctx context.Context, limit, offset int) ([]models.User, error)
    Count(ctx context.Context) (int, error)
}
```

**Queries migrated from:** `auth.go` (6 queries), `admin.go` (3 queries)

### ProductRepository

```go
type ProductRepository interface {
    FindBySlug(ctx context.Context, slug string) (*models.Product, error)
    FindByID(ctx context.Context, id int) (*models.Product, error)
    Create(ctx context.Context, p *models.CreateProductRequest, slug string) (*models.Product, error)
    Update(ctx context.Context, id int, p *models.UpdateProductRequest) (*models.Product, error)
    SoftDelete(ctx context.Context, id int) error
    List(ctx context.Context, filters ProductFilters) ([]models.Product, int, error)
    GetImages(ctx context.Context, productID int) ([]string, error)
    GetVariants(ctx context.Context, productID int) ([]models.ProductVariant, error)
    GetReviewStats(ctx context.Context, productID int) (float64, int, error)
}

type ProductFilters struct {
    Category string
    MinPrice *int
    MaxPrice *int
    Search   string
    Sort     string
    Limit    int
    Offset   int
}
```

**Queries migrated from:** `product.go` (8 queries), `image.go` (1 query)

### CartRepository

```go
type CartRepository interface {
    FindByUserID(ctx context.Context, userID int) ([]CartItemRow, error)
    FindBySessionID(ctx context.Context, sessionID string) ([]CartItemRow, error)
    Upsert(ctx context.Context, userID *int, sessionID *string, productID int, variantID *int, quantity int) error
    UpdateQuantity(ctx context.Context, itemID int, userID *int, sessionID *string, quantity int) error
    Delete(ctx context.Context, itemID int, userID *int, sessionID *string) error
    GetGuestItems(ctx context.Context, sessionID string) ([]GuestCartItem, error)
    MergeGuestCart(ctx context.Context, userID int, sessionID string) (int, error) // returns merged count
    ClearByUserID(ctx context.Context, userID int) error
    GetStock(ctx context.Context, variantID, productID int) (int, error)
}

type CartItemRow struct {
    CartItemID    int
    Quantity      int
    ProductID     int
    ProductName   string
    ProductSlug   string
    ProductPrice  int
    ProductImages []string
    VariantID     *int
    VariantSize   *string
    VariantColor  *string
    VariantStock  *int
}

type GuestCartItem struct {
    ProductID int
    VariantID *int
    Quantity  int
}
```

**Queries migrated from:** `cart.go` (13 queries)

### OrderRepository

```go
type OrderRepository interface {
    Checkout(ctx context.Context, userID int, req CheckoutParams) (*CheckoutResult, error)
    FindByID(ctx context.Context, orderID, userID int) (*models.Order, error)
    FindByPaymentIntent(ctx context.Context, piID string) (*models.Order, error)
    ListByUserID(ctx context.Context, userID, limit, offset int) ([]OrderSummary, int, error)
    ListAll(ctx context.Context, status string, limit, offset int) ([]AdminOrder, int, error)
    UpdateStatus(ctx context.Context, orderID int, status string) error
    MarkPaid(ctx context.Context, orderID int) (int64, error)
    MarkPaymentFailed(ctx context.Context, orderID int) error
    GetItems(ctx context.Context, orderID int) ([]models.OrderItem, error)
    GetStatus(ctx context.Context, orderID int) (string, error)
    GetUserIDAndTotal(ctx context.Context, orderID int) (int, int, error)
}

type CheckoutParams struct {
    Total         int
    ShippingAddress models.ShippingAddress
    PaymentIntentID string
    CartItems     []CheckoutCartItem
    IdempotencyKey string
}

type CheckoutResult struct {
    OrderID int
}

type CheckoutCartItem struct {
    ProductID    int
    VariantID    *int
    Quantity     int
    Price        int
    ProductName  string
    VariantLabel string
}

type OrderSummary struct {
    ID         int
    Status     string
    Total      int
    CreatedAt  string
    ItemsCount int
}

type AdminOrder struct {
    ID           int
    UserID       int
    CustomerName string
    Status       string
    Total        int
    CreatedAt    string
}
```

**Queries migrated from:** `order.go` (19 queries), `webhook.go` (5 queries)

### ReviewRepository

```go
type ReviewRepository interface {
    Create(ctx context.Context, productID, userID, rating int, comment string) (*models.Review, error)
    ListByProduct(ctx context.Context, productID, limit, offset int) ([]models.Review, error)
    GetStats(ctx context.Context, productID int) (float64, int, error)
    ExistsProduct(ctx context.Context, productID int) (bool, error)
}
```

**Queries migrated from:** `review.go` (5 queries)

### IdempotencyRepository

```go
type IdempotencyRepository interface {
    FindByKey(ctx context.Context, key string, userID int) (*IdempotencyResult, error)
    Store(ctx context.Context, key string, userID, orderID int) error
}

type IdempotencyResult struct {
    OrderID           int
    PaymentIntentID   string
}
```

**Queries migrated from:** `order.go` (2 queries)

### DeadLetterRepository

```go
type DeadLetterRepository interface {
    List(ctx context.Context, limit, offset int) ([]DeadLetter, error)
    Count(ctx context.Context) (int, error)
    FindByID(ctx context.Context, id int) (*DeadLetter, error)
    MarkRetried(ctx context.Context, id int) error
    Delete(ctx context.Context, id int) error
    Persist(ctx context.Context, queueName string, jobType string, payload []byte, errMsg string, retryCount int) error
}

type DeadLetter struct {
    ID           int
    JobType      string
    QueueName    string
    Payload      json.RawMessage
    ErrorMessage string
    RetryCount   int
    Status       string
    CreatedAt    string
    RetriedAt    *string
}
```

**Queries migrated from:** `dead_letter.go` (5 queries), `queue.go` (1 query)

### StockRepository

```go
type StockRepository interface {
    LockForUpdate(ctx context.Context, tx pgx.Tx, variantID int) (int, error)
    Decrement(ctx context.Context, tx pgx.Tx, variantID, quantity int) error
    Increment(ctx context.Context, variantID, quantity int) error
    GetAvailable(ctx context.Context, variantID, productID int) (int, error)
}
```

**Queries migrated from:** `order.go` (2 queries), `webhook.go` (2 queries), `cart.go` (1 query)

## Wiring

In `cmd/server/main.go`, repositories are instantiated with `database.DB` and passed to handlers:

```go
db := database.DB

userRepo := repositories.NewUserRepository(db)
productRepo := repositories.NewProductRepository(db)
cartRepo := repositories.NewCartRepository(db)
orderRepo := repositories.NewOrderRepository(db, stockRepo, idempotencyRepo)
reviewRepo := repositories.NewReviewRepository(db)
idempotencyRepo := repositories.NewIdempotencyRepo(db)
deadLetterRepo := repositories.NewDeadLetterRepository(db)
stockRepo := repositories.NewStockRepository(db)

authHandler := handlers.NewAuthHandler(userRepo)
productHandler := handlers.NewProductHandler(productRepo, reviewRepo)
cartHandler := handlers.NewCartHandler(cartRepo, productRepo)
orderHandler := handlers.NewOrderHandler(orderRepo, queue)
reviewHandler := handlers.NewReviewHandler(reviewRepo, productRepo)
adminHandler := handlers.NewAdminHandler(userRepo)
webhookHandler := handlers.NewWebhookHandler(webhookSecret, orderRepo, stockRepo, queue)
deadLetterHandler := handlers.NewDeadLetterHandler(deadLetterRepo, queue)
imageHandler := handlers.NewImageHandler(productRepo)
```

## Migration Order

1. **Repositories without dependencies:** `UserRepository`, `DeadLetterRepository`, `StockRepository`
2. **Repositories with simple dependencies:** `ProductRepository`, `ReviewRepository`, `IdempotencyRepository`
3. **Repositories with transaction dependencies:** `CartRepository`, `OrderRepository`
4. **Wire handlers:** Update constructors, remove SQL from handlers
5. **Verify:** `go build ./...` && `go vet ./...`
