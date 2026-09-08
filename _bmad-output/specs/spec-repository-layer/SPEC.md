---
id: SPEC-repository-layer
companions: [repository-interfaces.md]
sources: []
---

> **Canonical contract.** This SPEC and the files in `companions:` are the complete, preservation-validated contract for what to build, test, and validate. Source documents listed in frontmatter are for traceability — consult them only if you need narrative rationale or prose color this contract intentionally omits.

# Repository Layer Refactoring

## Why

All 68 SQL queries are embedded directly in HTTP handlers via the global `database.DB` variable. This couples data access to transport, makes unit testing impossible without a live database, and violates separation of concerns. The pain is specific: every handler mixes request parsing, business logic, query execution, and response formatting in a single method. A repository layer extracts the data access into testable, injectable interfaces.

## Capabilities

- **CAP-1** — UserRepository
  - **intent:** Handlers can register users, authenticate, and manage profiles through a `UserRepository` interface without direct SQL access.
  - **success:** `UserRepository` interface exists with `ExistsByEmail`, `Create`, `FindByEmail`, `FindByID`, `UpdateProfile`, `List`, `Count` methods. All user SQL queries removed from `auth.go` and `admin.go`.

- **CAP-2** — ProductRepository
  - **intent:** Product CRUD, filtered listing, variant queries, and review aggregation go through a `ProductRepository` interface.
  - **success:** `ProductRepository` interface exists with `FindBySlug`, `FindByID`, `Create`, `Update`, `SoftDelete`, `List` (with filter params), `Count`, `GetImages`, `GetVariants`, `GetReviewStats` methods. All product SQL removed from `product.go` and `image.go`.

- **CAP-3** — CartRepository
  - **intent:** Cart operations for users and guests, including the merge transaction, go through a `CartRepository` interface.
  - **success:** `CartRepository` interface exists with `FindByUserID`, `FindBySessionID`, `Upsert`, `UpdateQuantity`, `Delete`, `GetGuestItems`, `MergeGuestCart`, `ClearByUserID` methods. All cart SQL removed from `cart.go`.

- **CAP-4** — OrderRepository
  - **intent:** Checkout transaction, order listing, and admin order management go through an `OrderRepository` interface.
  - **success:** `OrderRepository` interface exists with `Create` (transaction), `FindByID`, `FindByUserID`, `ListByUserID`, `ListAll`, `UpdateStatus`, `GetItems` methods. Checkout transaction encapsulated inside `OrderRepository.Checkout`. All order SQL removed from `order.go`.

- **CAP-5** — ReviewRepository
  - **intent:** Review creation and listing with pagination go through a `ReviewRepository` interface.
  - **success:** `ReviewRepository` interface exists with `Create` (with ON CONFLICT), `ListByProduct`, `GetStats` methods. All review SQL removed from `review.go`.

- **CAP-6** — IdempotencyRepository
  - **intent:** Checkout idempotency key lookups and storage go through an `IdempotencyRepository` interface.
  - **success:** `IdempotencyRepository` interface exists with `FindByKey`, `Store` methods. Idempotency SQL removed from `order.go`.

- **CAP-7** — DeadLetterRepository
  - **intent:** Dead letter queue management goes through a `DeadLetterRepository` interface.
  - **success:** `DeadLetterRepository` interface exists with `List`, `Count`, `FindByID`, `MarkRetried`, `Delete`, `Persist` methods. All DLQ SQL removed from `dead_letter.go` and `queue.go`.

- **CAP-8** — WebhookRepository
  - **intent:** Stripe webhook order lookups and status updates go through a repository interface (part of `OrderRepository` or dedicated).
  - - **success:** Methods `FindByPaymentIntent`, `MarkPaid`, `MarkPaymentFailed` exist on `OrderRepository`. All webhook SQL removed from `webhook.go`.

- **CAP-9** — StockRepository
  - **intent:** Stock locking and decrement/increment for checkout and restore go through a `StockRepository` interface.
  - **success:** `StockRepository` interface exists with `LockForUpdate`, `Decrement`, `Increment` methods. Stock SQL removed from `order.go` and `webhook.go`.

## Constraints

- Repositories receive `*pgxpool.Pool` via constructor, not the global `database.DB`.
- All repository interfaces are defined in `internal/repositories/` alongside their implementations.
- Handlers depend on interfaces, not concrete types — constructors accept interfaces.
- Transactions are managed inside repository methods, not handlers. The `Checkout` method on `OrderRepository` encapsulates the entire checkout transaction.
- Dynamic query building (product filters, order status filters) is encapsulated in repository implementations using parameterized queries.
- No ORM or query builder library — raw SQL with pgx continues.

## Non-goals

- No introduction of an ORM (GORM, Ent, etc.).
- No query builder library (squirrel, goqu, etc.).
- No migration tooling changes.
- No new database schema changes.
- No handler logic changes beyond removing SQL and calling repository methods.

## Success signal

- All 68 SQL queries moved from handlers to repository implementations.
- Handlers contain zero `database.DB` references (except `health.go` ping).
- `go build ./...` and `go vet ./...` pass.
- Repository interfaces can be mocked for handler unit tests.

## Assumptions

- The global `database.DB` variable remains as the connection pool source; repositories receive it via constructor.
- `health.go` ping stays as-is (not a repository concern).
- The `queue.go` `persistDeadLetter` method uses its own injected `*pgxpool.Pool`, which becomes `DeadLetterRepository.Persist`.

## Open Questions

- Should `OrderRepository.Checkout` accept a callback for the Stripe PaymentIntent creation, or should the handler pass the PI details after repository creates the order?
- Should `WebhookRepository` be a separate interface or part of `OrderRepository`?
