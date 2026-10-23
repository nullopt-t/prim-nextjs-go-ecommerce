# TEST_INFRA.md: E2E and Decoupling Testing Framework

## 1. Test Philosophy & Invariants
This test suite provides a requirement-driven, opaque-box, offline-executable verification framework for the Go backend domains refactored under the `domain-service-decoupling` architectural standard.

### Core Architectural Invariants
1. **R1: Cross-Domain Decoupling & Consumer Repository Interfaces**
   - Domain services never inject foreign domain services (no `ObjectService`, `VariantService`, `InventoryService`, `ProductService` injected as dependencies).
   - All inter-domain data access is mediated through narrow, consumer-defined repository interfaces declared within the consumer's package.
   - Repositories return plain domain models and raw driver errors (no `apierr` in repositories).

2. **R2: Single-Connection Threading & Zero Connection Hopping**
   - Invariant: Exactly one database connection per request/use-case held from entry to exit.
   - Public read entry points acquire a single `WithDB`; public write entry points acquire a single `WithTx`.
   - Internal helper methods never call `WithDB` or `WithTx` while holding a connection — they accept `qe database.QueryExecutor`.
   - Service methods participating in a caller's transaction require explicit `tx *database.Tx`.

3. **R3: Direct Storage Provider & Asset Resolution**
   - Domains resolving asset URLs inject `storage.StorageProvider` directly.
   - URLs are generated in-memory via `storageProvider.GetPublicURL(bucket, key)` without database queries or cross-service network calls.
   - Mutations adhere to safe ordering: upload before commit (compensating delete on rollback); commit DB before blob deletion.

4. **R4: Elimination of Redundant `...WithTx` Twin Methods**
   - Uncalled twin methods (`ProductService.GetByIDWithTx`, `VariantService.GetVariantByIDWithTx`, `ObjectService.CreateObjectWithTx`, `ObjectService.MarkDeletingByKey`, etc.) are removed.
   - Callers requiring transactional access call the target domain's repository directly on their transaction.

5. **R5: Canonical Domain Error Mapping via `handleError`**
   - Domain services implement `handleError` translating driver errors (`pgx.ErrNoRows`, `database.ErrNotFound`, `database.ErrConflict`, `database.ErrForeignKeyViolation`, `context.Canceled`, `context.DeadlineExceeded`) into typed `apierr.APIError` instances.
   - Already-wrapped `*apierr.APIError` instances pass through without double wrapping.
   - Handlers never inspect raw database driver types or error codes.

---

## 2. Feature Inventory
The test suite covers features F1 through F15 defined in `PROJECT.md`:

| # | Feature | Scope / Requirement | Test Tier | Primary Verification Target |
|---|---------|---------------------|-----------|-----------------------------|
| F1 | Dead `...WithTx` Twin Removal | R4 | Tier 1, Arch | AST analysis & repository invocation validation |
| F2 | Product/Variant Decoupling | R1, R2 | Tier 1, Tier 3 | Removal of `variantService` in `ProductService`, `variantRepo` direct call |
| F3 | Supporting Domains `handleError` | R5 | Tier 1, Tier 2 | Error translation in `object`, `setting`, `audit`, `report` |
| F4 | Supporting Domains Single Connection | R2 | Tier 1, Tier 2 | `WithTx` usage for writes in `object`, `audit` |
| F5 | Order Consumer Repo Decoupling | R1 | Tier 1, Tier 3 | `order.InventoryRepository` interface compliance |
| F6 | Order Tx Signature Compliance | R2 | Tier 1, Arch | `CreateOrderWithTx` takes explicit `*database.Tx` |
| F7 | Order Canonical `handleError` | R5 | Tier 1, Tier 2 | Elimination of unmapped driver errors in `order` |
| F8 | Checkout Foreign Service Decoupling | R1 | Tier 1, Tier 3 | Consumer repositories for cart, order, user, address, coupon, inventory |
| F9 | Checkout Direct Storage Resolution | R3 | Tier 1, Tier 3 | Direct `StorageProvider.GetPublicURL` in checkout lines |
| F10 | Checkout Single Connection Threading | R2 | Tier 1, Tier 2 | Single `WithDB` in coupon validation; single `WithTx` in checkout |
| F11 | Checkout Canonical `handleError` | R5 | Tier 1, Tier 2 | Typed `APIError` mapping for checkout failure paths |
| F12 | User Domain `...WithTx` Twins Cleanup | R4 | Tier 1, Arch | Direct user/address repo calls on `tx` |
| F13 | Composition Root Wiring | R1, R2, R3 | Tier 3, Tier 4 | Application container wiring in `app.go` & `worker/main.go` |
| F14 | Architectural Invariant Verification | R1-R5 | Arch | Programmatic AST/regex checks (0 violations) |
| F15 | Comprehensive Scenario & Adversarial | Tiers 1-4 | Tiers 1-4 | End-to-end integration flows & stress edge cases |

---

## 3. Test Architecture

```
backend/tests/
├── arch/
│   └── arch_test.go              # Static AST & regex checks for R1-R5 architectural invariants
└── e2e/
    ├── harness.go                # Offline in-memory database runner, executor, storage, and repo fakes
    ├── tier1_feature_test.go     # Feature-level compliance tests (>=5 tests per feature group)
    ├── tier2_boundary_test.go    # Boundary conditions, error mappings, connection pool & tx invariants
    ├── tier3_cross_feature_test.go # Multi-domain interactions & composition workflows
    └── tier4_real_world_test.go  # Real-world end-to-end checkout, order lifecycle, media, and audit
```

### Offline Execution Philosophy
Integration tests relying on external Docker daemons and `testcontainers/ryuk` can suffer timeouts or resource contention in CI/CD sandboxes. To guarantee fast, repeatable, and offline execution, the E2E test framework introduces `tests/e2e/harness.go`:
- **`FakeRunner`**: A deterministic in-memory `database.Runner` tracking `WithDB` and `WithTx` invocations, connection acquisition count, and transaction state (committed/rolled back).
- **`SingleConnectionGuard`**: Detects and asserts against connection hopping (simulates pool capacity = 1).
- **`FakeStorage`**: An in-memory `storage.StorageProvider` tracking uploads, deletions, and validating `GetPublicURL` formatting.
- **`MockRepositories`**: In-memory repository fakes satisfying consumer-defined interfaces across all domains.

---

## 4. Test Tiers & Minimum Coverage Matrix

### Tier 1: Feature Coverage (>=5 tests per feature)
- **F1 & F12 (Twin Removal)**: Verify direct repo invocation on `tx`, lack of twin methods, and clean signature verification.
- **F2 (Product/Variant Decoupling)**: Verify product publishing with variant count, variant price lookups, product stock attachment via repository.
- **F3 & F4 (Supporting Domains)**: Verify single-connection `CreateObject`, `DeleteObject`, `Record` (audit), and `UpdateSettings`.
- **F5, F6 & F7 (Order Domain)**: Verify order creation with `*database.Tx`, stock release via consumer `InventoryRepository`, and error mapping.
- **F8, F9, F10 & F11 (Checkout Domain)**: Verify single-connection coupon validation, direct storage resolution, consumer repository aggregation, and error handling.

### Tier 2: Boundary & Corner Cases (>=5 tests per category)
- **Error Mapping Boundaries**:
  1. `pgx.ErrNoRows` / `database.ErrNotFound` -> HTTP 404 (`apierr.NotFound`).
  2. `pgerrcode.UniqueViolation` (23505) / `database.ErrConflict` -> HTTP 409 (`apierr.Conflict`).
  3. `pgerrcode.ForeignKeyViolation` (23503) -> HTTP 400 (`apierr.BadRequest`).
  4. `pgerrcode.NotNullViolation` (23502) / `CheckViolation` (23514) -> HTTP 400 (`apierr.BadRequest`).
  5. `context.Canceled` & `context.DeadlineExceeded` -> HTTP 408 / 499 / 504 (`apierr.Timeout`).
  6. Idempotency of already-wrapped `*apierr.APIError` (no double-wrapping).
  7. Internal database error sanitization (no SQL queries or database table names in user-facing message).
- **Connection & Transaction Invariants**:
  8. `MaxConns=1` concurrency & connection hopping prevention.
  9. Rollback guarantee on error return within `WithTx`.
  10. Panic recovery inside `WithTx` ensuring rollback before re-panic.
  11. Handling optional lookups (treating `ErrNoRows` as empty value rather than aborting transaction).

### Tier 3: Cross-Feature Combinations
- **Catalog + Variant + Storage**: Product retrieval resolving images via storage provider and stock via inventory repository on a single connection.
- **Order + Inventory**: Transactional order placement reserving stock, payment confirmation committing reservations, and cancellation restoring stock via inventory repository.
- **Checkout + Cart + Coupon + Variant**: Checkout orchestration validating cart, checking coupon thresholds, resolving variant snapshots, and clearing cart atomically.
- **Checkout + User + Address**: Validating address ownership and customer status on transaction.

### Tier 4: Real-World Scenarios
1. **Full Checkout Flow**: Customer adds items to cart -> applies discount coupon -> executes checkout -> order created -> reservations converted to sale ledger -> cart cleared.
2. **Order Lifecycle Flow**: Order pending -> payment confirmed -> status updated to processing -> customer requests cancellation -> order cancelled -> inventory restored via ledger -> audit entry recorded.
3. **Asset Upload & Resolution Flow**: File uploaded to storage provider -> metadata persisted in object repo -> URL derived via `GetPublicURL` -> compensating cleanup on DB transaction failure.
4. **Administrative Audit & Configuration**: Store setting updated -> cache invalidated -> audit trail entry recorded within transaction.

---

## 5. Architectural Invariant Checks (`tests/arch`)
Programmatic static analysis tests assert:
1. `TestArch_ZeroForeignServiceInjections`: Verifies services do not import or declare fields of foreign `*Service` types.
2. `TestArch_ZeroRawDBErrorsInHandlers`: Verifies HTTP handlers do not reference `pgx`, `pgconn`, or `pgerrcode`.
3. `TestArch_RepositorySignatures`: Asserts all repository methods follow `(ctx context.Context, qe database.QueryExecutor, ...)`.
4. `TestArch_NoDeadTwinWithTxMethods`: Scans domain services for uncalled `...WithTx` twin methods.
5. `TestArch_DirectStorageResolution`: Asserts asset URLs are generated via `storage.StorageProvider.GetPublicURL`.

---

## 6. Execution Commands & Coverage Thresholds
- **Test Compilation Check**:
  ```bash
  cd backend && go test -run=^$ ./...
  ```
- **Execute E2E & Architectural Tests**:
  ```bash
  cd backend && go test -v ./tests/...
  ```
- **Execute Specific Tiers**:
  ```bash
  cd backend && go test -v ./tests/e2e -run TestTier1
  cd backend && go test -v ./tests/e2e -run TestTier2
  cd backend && go test -v ./tests/e2e -run TestTier3
  cd backend && go test -v ./tests/e2e -run TestTier4
  cd backend && go test -v ./tests/arch
  ```

### Coverage Thresholds
- **Architectural Conformance**: 0 violations across all 5 architectural invariant checks.
- **Test Execution**: 100% pass rate across all test cases in `backend/tests/e2e` and `backend/tests/arch`.
- **Offline Reliability**: 0 dependencies on external networks or live Docker daemons.
