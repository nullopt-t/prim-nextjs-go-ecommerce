# Project: Go Backend Domain Service Decoupling & Architecture Refactoring

## Architecture
This project refactors and audits all Go backend domains (`order`, `checkout`, `object`, `setting`, `audit`, `report`, and leftover `product`/`variant` methods) to strictly adhere to the `domain-service-decoupling` architecture rules:
- **Layering**: Handlers decode/render -> Services orchestrate use-cases & map errors -> Repositories execute SQL on the held `qe database.QueryExecutor`.
- **Bounded Contexts**: Domains never inject foreign domain services. Cross-domain data access occurs via narrow consumer-defined repository interfaces declared in the consuming domain's package.
- **Connection Invariant**: Exactly one database connection per request/use-case. Public reads acquire a single `WithDB`; public writes acquire a single `WithTx`. Helpers accept `qe database.QueryExecutor`. Methods participating in a caller's transaction require explicit `tx *database.Tx`.
- **Direct Storage Resolution**: Assets are resolved in-memory via `storage.StorageProvider.GetPublicURL(bucket, key)`. No DB queries or cross-service calls for URLs.
- **Canonical Error Mapping**: All database driver errors are mapped at the service boundary to typed `apierr.APIError` via `handleError(err error, fallbackMsg string) error`. Handlers never inspect raw DB errors.
- **Composition Root**: Dependencies and concrete repositories are wired at the top level in `backend/internal/app/app.go` and `backend/cmd/worker/main.go`.

---

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| F1 | Dead `...WithTx` twin removal | Delete uncalled twin methods: `ProductService.GetByIDWithTx`, `VariantService.GetVariantByIDWithTx`, `ObjectService.CreateObjectWithTx`, `ObjectService.MarkDeletingByKey` | M1 | R4, survey |
| F2 | Product/Variant service decoupling | Remove `variantService` injection from `ProductService`, replace `ListVariantsByProductID` in `PublishProduct` with `variantRepo.CountByProductID`, clean leftover proxy methods | M1 | R1, survey |
| F3 | Supporting domains canonical `handleError` | Implement `handleError` translating `database.MapError` and pgx/pgconn errors to `apierr.APIError` in `object`, `setting`, `audit`, `report` services | M1 | R5, survey |
| F4 | Supporting domains single-connection threading | Use `WithTx` for mutating operations in `object` (`CreateObject`, `DeleteObject`) and `audit` (`Record`), avoid connection hopping | M1 | R2, survey |
| F5 | Order domain consumer repository decoupling | Replace `InventoryService` injection in `OrderService` with consumer-defined `InventoryRepository` interface in `package order` | M2 | R1, survey |
| F6 | Order transaction signature compliance | Change `OrderService.CreateOrderWithTx` signature from `database.QueryExecutor` to explicit `*database.Tx` | M2 | R2, survey |
| F7 | Order domain canonical `handleError` | Implement `handleError` in `OrderService` and eliminate driver error leakage in `UpdateOrderTracking`, `GetDashboardStats`, `ListPayments` | M2 | R5, survey |
| F8 | Checkout domain foreign service decoupling | Replace `OrderService`, `InventoryService`, `UserService`, `AddressService`, and `SettingsProvider` injections with consumer repository interfaces in `package checkout` | M3 | R1, survey |
| F9 | Checkout direct storage resolution | Inject `storage.StorageProvider` directly into `checkout.Service` and derive public URLs via `GetPublicURL(bucket, key)` (fixing broken `so.public_url` SQL query) | M3 | R3, survey |
| F10 | Checkout transaction threading & coupon fix | Fix connection hopping in `checkout.ValidateCoupon` (single `WithDB`), thread single `WithTx` in `Checkout` | M3 | R2, survey |
| F11 | Checkout domain canonical `handleError` | Implement canonical `handleError` in `CheckoutService` mapping all DB errors to `apierr.APIError` | M3 | R5, survey |
| F12 | User domain `...WithTx` twins cleanup | Remove `UserService.GetUserByIDWithTx` and `AddressService.GetAddressWithTx` once checkout calls consumer repositories | M3 | R4, survey |
| F13 | Composition root update | Rewire constructors and dependencies in `backend/internal/app/app.go` and `backend/cmd/worker/main.go` | M4 | R1, R2, survey |
| F14 | Architectural grep verification | Programmatic verification producing 0 violations for foreign service injections, raw DB errors in handlers, `...WithTx` twins, and repository signatures | M4 | Verification, survey |
| F15 | Comprehensive test suite & adversarial hardening | Pass 100% of E2E / domain decoupling tests across all tiers, followed by Tier 5 adversarial coverage hardening | M5 | Acceptance criteria |

---

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | Cleanups & Supporting Domains | Remove dead `...WithTx` methods in `product`, `variant`, `object`; decouple `ProductService` from `VariantService`; implement `handleError` and single-connection threading in `object`, `setting`, `audit`, `report` | none | IN_PROGRESS |
| M2 | Order Domain Decoupling | Decouple `order` from `InventoryService` via consumer `InventoryRepository`; change `CreateOrderWithTx` parameter to explicit `*database.Tx`; implement canonical `handleError` and eliminate DB error leakage in `order` | none | PLANNED |
| M3 | Checkout Domain Decoupling & Storage | Decouple `checkout` from foreign services (`Order`, `Inventory`, `User`, `Address`, `Settings`); inject `StorageProvider` for URL resolution; fix `ValidateCoupon` connection hopping; implement `handleError`; clean up `user` twins | M2 | PLANNED |
| M4 | Composition Root & Architectural Greps | Update wiring in `app.go` and `worker/main.go`; verify all programmatic architectural greps (0 foreign injections, 0 handler leaks, 0 dead twins, 100% repository signatures); verify clean build | M1, M2, M3 | PLANNED |
| M5 | Final E2E Test Suite & Adversarial Hardening | Pass 100% of requirement-driven E2E tests (Tiers 1-4); run Tier 5 adversarial stress testing and forensic audit sign-off | M4 | PLANNED |

---

## Interface Contracts

### `order` ↔ `inventory` (Consumer-defined in `package order`)
```go
type InventoryRepository interface {
	ReleaseReservationsForOrder(ctx context.Context, qe database.QueryExecutor, orderID uuid.UUID) (int64, error)
	ExpiredReservationOrderIDs(ctx context.Context, qe database.QueryExecutor) ([]uuid.UUID, error)
	CreateLedger(ctx context.Context, qe database.QueryExecutor, ledger *model.InventoryLedger) error
	GetUnreleasedReservationsForOrder(ctx context.Context, qe database.QueryExecutor, orderID uuid.UUID) ([]model.InventoryReservation, error)
	LockVariantForUpdate(ctx context.Context, qe database.QueryExecutor, variantID uuid.UUID) error
	GetStock(ctx context.Context, qe database.QueryExecutor, variantID uuid.UUID) (*model.InventoryStock, error)
	ReleaseReservation(ctx context.Context, qe database.QueryExecutor, reservationID uuid.UUID) error
}
```

### `checkout` ↔ `user` (Consumer-defined in `package checkout`)
```go
type UserRepository interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.User, error)
}
type AddressRepository interface {
	GetByIDAndUser(ctx context.Context, qe database.QueryExecutor, id uuid.UUID, userID uuid.UUID) (*model.UserAddress, error)
}
```

### `checkout` ↔ `coupon` (Consumer-defined in `package checkout`)
```go
type CouponRepository interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Coupon, error)
	GetByCode(ctx context.Context, qe database.QueryExecutor, code string) (*model.Coupon, error)
	GetRedemptionCount(ctx context.Context, qe database.QueryExecutor, couponID uuid.UUID) (int, error)
	GetUserRedemptionCount(ctx context.Context, qe database.QueryExecutor, couponID uuid.UUID, userID uuid.UUID) (int, error)
	RecordRedemption(ctx context.Context, qe database.QueryExecutor, red *model.CouponRedemption) error
}
```

### `checkout` ↔ `inventory` (Consumer-defined in `package checkout`)
```go
type InventoryRepository interface {
	LockVariantForUpdate(ctx context.Context, qe database.QueryExecutor, variantID uuid.UUID) error
	GetStock(ctx context.Context, qe database.QueryExecutor, variantID uuid.UUID) (*model.InventoryStock, error)
	CreateReservation(ctx context.Context, qe database.QueryExecutor, res *model.InventoryReservation) error
	GetUnreleasedReservationsForOrder(ctx context.Context, qe database.QueryExecutor, orderID uuid.UUID) ([]model.InventoryReservation, error)
	ReleaseReservation(ctx context.Context, qe database.QueryExecutor, reservationID uuid.UUID) error
	ReleaseReservationsForOrder(ctx context.Context, qe database.QueryExecutor, orderID uuid.UUID) (int64, error)
	CreateLedger(ctx context.Context, qe database.QueryExecutor, ledger *model.InventoryLedger) error
}
```

### `checkout` ↔ `variant` & `storage` (Consumer-defined in `package checkout`)
```go
type VariantRepository interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.ProductVariant, error)
}
// Storage resolution: direct storage.StorageProvider injection
// s.storageProvider.GetPublicURL(bucket, key)
```

### `checkout` ↔ `order`
```go
// Inside order package:
func (s *OrderService) CreateOrderWithTx(ctx context.Context, tx *database.Tx, in *CreateOrderInput) (*model.Order, error)
```

---

## Code Layout
- `backend/internal/catalog/product/`: `product_service.go`, `product_router.go`, `product_handler.go`
- `backend/internal/catalog/product/variant/`: `variant_service.go`, `variant_repository.go`
- `backend/internal/object/`: `object_service.go`, `object_repository.go`, `object_handler.go`
- `backend/internal/setting/`: `setting_service.go`, `setting_repository.go`, `setting_handler.go`
- `backend/internal/audit/`: `audit_service.go`, `audit_repository.go`, `audit_handler.go`
- `backend/internal/report/`: `report_service.go`, `report_repository.go`, `report_handler.go`
- `backend/internal/order/`: `order_service.go`, `order_repository.go`, `order_handler.go`
- `backend/internal/checkout/`: `checkout_service.go`, `checkout_handler.go`
- `backend/internal/user/`: `user_service.go`, `address_service.go`
- `backend/internal/app/`: `app.go`
- `backend/cmd/worker/`: `main.go`
- `backend/tests/e2e/`: E2E architectural tests
