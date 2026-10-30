# Backend Refactor Log: Context-Based Executor & Transaction Architecture

master journal of the backend-wide database executor propagation and transaction-aware execution refactoring.

## 1. Goal & Objectives

Refactor the entire PRIM Go backend to implement centralized, context-based database executor propagation and transaction-aware execution:
1. **Remove explicit transaction parameters** (`*database.Tx`, `database.QueryExecutor`) from all service, repository, and reader/writer method signatures.
2. **Propagate active database executor through `context.Context`** under a single unified `executorKey`.
3. **Transaction-Aware `WithTx`**:
   - Outermost `WithTx`: begins a new PostgreSQL transaction, binds it to the context, and owns lifecycle (commit on success, rollback on error or panic).
   - Nested `WithTx`: detects the ambient transaction already in flight on the context, reuses it transparently, avoids nested transactions / savepoints, and neither commits nor rolls back early.
4. **Enforce transaction requirements at the database execution boundary**:
   - Unconstrained reads/writes: `executor, err := database.Executor(ctx)`
   - Transaction-mandated operations (locks, reservations, `FOR UPDATE`): `tx, err := database.RequireTx(ctx)`
5. **Decouple business services from PostgreSQL mechanics**: services own business invariants and transaction boundaries via `s.runner.WithTx` / `s.runner.WithExecutor`, without extracting connections or inspecting driver error types.

---

## 2. Architectural Blueprint

```text
Application Context (HTTP Request / Worker Job)
       |
       v
  executorKey (single context key)
       |
       +----> Pool Executor (*database.DB)
       |        |
       |        +----> database.Executor(ctx) -> Exec / Query / QueryRow
       |
       +----> Transaction Executor (*database.Tx)
                |
                +----> database.Executor(ctx)   -> Exec / Query / QueryRow
                +----> database.RequireTx(ctx)  -> Tx-only operations
```

### Context Helpers in `pkg/database/runner.go`
- `WithExecutor(ctx, executor)`: Binds executor if absent; preserves existing executor if one already exists.
- `WithTx(ctx, tx)`: Binds active transaction under the unified `executorKey`, replacing non-transaction pool executors.
- `Executor(ctx)`: Extracts active `QueryExecutor`; returns `ErrNoExecutor` if missing.
- `RequireTx(ctx)`: Enforces an active transaction; returns `ErrNoActiveTx` if missing or pool executor.
- `TxFromContext(ctx)`: Returns `(*Tx, bool)` for ambient transaction detection.

---

## 3. Timeline & Execution Phases

### Phase 1: Database Execution Primitives
- Refactored `pkg/database/runner.go` and `pkg/database/db.go`.
- Implemented `executorKey`, `WithExecutor`, `WithTx`, `Executor`, `RequireTx`, and `TxFromContext`.
- Updated `crunner.WithTx` and `crunner.WithExecutor` to join ambient transactions seamlessly.
- Added comprehensive unit tests in `pkg/database/runner_test.go` verifying ambient joins, key replacement, error propagation, and mismatched DB rejection.

### Phase 2: Application Entry Points
- Wrapped incoming HTTP requests in `internal/app/app.go` with `database.WithExecutor(ctx, app.db)`.
- Wrapped worker daemon context in `cmd/worker/main.go` with `database.WithExecutor(ctx, db)`.
- Ensured fail-fast constructor invariants are satisfied across all services (`storage.NoopProvider{}` for worker inventory service).

### Phase 3: Domain Services & Repositories Refactoring
Migrated all 11 backend domains from explicit `qe database.QueryExecutor` / `tx *database.Tx` parameters to `(ctx context.Context, ...)`:
- `internal/order`: `order_repository.go`, `order_service.go`, `order_handler.go`.
- `internal/catalog/inventory`: `inventory_repository.go`, `inventory_service.go`.
- `internal/cart`: `cart_repository.go`, `cart_service.go`.
- `internal/catalog/product` & `variant`: `product_repository.go`, `product_service.go`, `variant_repository.go`, `variant_service.go`.
- `internal/catalog/category` & `attribute`: `category_repository.go`, `category_service.go`, `attribute_repository.go`, `attribute_service.go`.
- `internal/catalog/brand`: `brand_repository.go`, `brand_service.go`.
- `internal/catalog/coupon`: `coupon_repository.go`, `coupon_service.go`.
- `internal/catalog/review`: `review_repository.go`, `review_service.go`.
- `internal/user`: `user_repository.go`, `user_service.go`, `address_repository.go`, `address_service.go`, `notification_repository.go`, `notification_service.go`.
- `internal/setting`: `setting_repository.go`, `setting_service.go`.
- `internal/object`: `object_repository.go`, `object_service.go`.
- `internal/audit`: `audit_repository.go`, `audit_service.go`.
- `internal/report`: `report_repository.go`, `report_service.go`.
- `internal/wishlist`: `wishlist_repository.go`, `wishlist_service.go`.
- `internal/checkout`: `checkout_service.go` (migrated all consumer reader interfaces).

### Phase 4: Test Doubles & Test Suites
- Updated `tests/arch/arch_test.go` (`TestArch_RepositorySignatures`) to enforce zero explicit `Tx` parameters in repository methods.
- Updated `tests/e2e/harness.go` test doubles (`MockRunner`, `MockInventoryRepository`, `MockCouponRepository`, `MockUserRepository`, `MockAddressRepository`, `MockCartRepository`, `MockVariantRepository`, `MockProductRepository`, `MockObjectRepository`, `MockOrderRepository`, `MockAuditRepository`).
- Refactored all tests across `tests/e2e/tier1_feature_test.go`, `tier2_boundary_test.go`, `tier3_cross_feature_test.go`, and `tier4_real_world_test.go`.

---

## 4. Verification & Quality Gates

| Gate | Command | Result |
| :--- | :--- | :--- |
| **Full Build** | `go build ./...` | **PASS** (0 errors) |
| **Static Analysis** | `go vet ./...` | **PASS** (0 warnings) |
| **Linter** | `golangci-lint run ./...` | **PASS** (0 issues) |
| **Architecture Invariants** | `go test -v ./tests/arch/...` | **PASS** (100%) |
| **Decoupling & E2E Suites** | `go test -v ./tests/e2e/...` | **PASS** (100%) |
| **Race Detector** | `go test -race ./tests/arch/... ./tests/e2e/... ./pkg/database/... ...` | **PASS** (0 data races) |
| **Production Compose** | `make prod-up` | **PASS** (All 17 containers healthy & live probes return 200) |
