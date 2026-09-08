---
title: 'Option Group + Option Value Admin CRUD'
type: 'feature'
created: '2026-09-06'
status: 'done'
route: 'dispatch'
baseline_commit: '3c57e9e7a9f263b363c6661f6a776360e836dde4'
context:
  - internal/models/product.go
  - internal/repositories/product.go
  - internal/handlers/product.go
  - cmd/server/main.go
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Option groups and values were created in the database (story 1) but have no API endpoints or repository methods to manage them.

**Approach:** Add DTOs, repository methods, handler methods, and routes for CRUD operations on option groups and option values per product.

## Code Map

- `internal/models/product.go:60-73` — `ProductOptionGroup`, `ProductOptionValue` structs (already exist)
- `internal/repositories/product.go:14-36` — `ProductRepository` interface (needs new methods)
- `internal/handlers/product.go:18-26` — `ProductHandler` struct
- `cmd/server/main.go:240-314` — admin product routes (add option-group routes)

## Tasks & Acceptance

**Execution:**
- [ ] `internal/models/product.go` — Add request DTOs for option group/value CRUD
- [ ] `internal/repositories/product.go` — Add option group/value repository methods to interface and implementation
- [ ] `internal/handlers/product.go` — Add handler methods for option group/value CRUD
- [ ] `cmd/server/main.go` — Add routes for option group/value endpoints
- [ ] Verify `go build ./...` and `go test ./...` pass

**Acceptance Criteria:**
- Given admin creates option group with values, when GET product options endpoint called, then option groups with values returned
- Given admin updates option group name, when checked, then name updated
- Given admin deletes option group, when checked, then group and its values deleted
- Given admin adds value to existing group, when checked, then value appears in group
- Given admin updates value label/price_modifier, when checked, then value updated
- Given admin deletes value, when checked, then value removed from group
- Given product has max 6 option groups, when admin tries to add 7th, then error returned
- Given group has max 100 values, when admin tries to add 101st, then error returned

## Verification

**Commands:**
- `go build ./...` -- expected: no compilation errors
- `go test ./...` -- expected: all pass
