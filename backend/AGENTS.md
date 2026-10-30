# Backend Agent Rules

Scope: everything under `backend/`. This file governs the Go API (`cmd/api`), the
notification worker (`cmd/worker`), shared packages, migrations, and Swagger
output. Do not make coordinated frontend changes from a backend task; the
frontend has its own `AGENTS.md`.

## Start here

1. Load `using-agent-skills` before a non-trivial task.
2. Load `test-driven-development` for behavior changes and
   `debugging-and-error-recovery` for failures before guessing at a fix.
3. Read the relevant package and its tests before editing. Follow the existing
   handler → service → repository shape; do not introduce a new pattern for one
   endpoint.

## Commands

Run from `backend/`:

```bash
go build ./...          # compile every package
go test ./...           # unit + integration tests
go test -race ./...     # race detector (CI gate)
go vet ./...            # static checks
golangci-lint run ./... # when available; config is .golangci.yml
make swagger            # regenerate public + admin specs
make build              # build ./bin/server and ./bin/worker
```

Run from the repository root:

```bash
make up-d               # full dev stack
make logs-api           # API logs
make logs-worker        # worker logs
make migrate-up         # apply migrations
make migrate-down       # roll back one migration
```

Integration tests use `testcontainers-go` and require a running Docker daemon.

## Architecture rules (non-negotiable)

The unit of work is the service method. One request/use case gets **one**
database connection from entry to exit.

- Handlers do HTTP only: bind/decode, validate shape, call the service, render.
  They must not run SQL or inspect `pgx`/`pgconn` errors.
- Services own business rules, the transaction boundary, and error mapping.
  They must not contain SQL.
- Repositories own SQL and row mapping. Every repository method takes
  `(ctx context.Context, ...)`; repositories never call `WithExecutor`/`WithTx`
  and never open their own connection. Instead, they extract the active executor
  at the execution boundary:
  - Operations that may use either pool or transaction:
    `executor, err := database.Executor(ctx)`
  - Operations that mandate an active transaction (locks, holds, `FOR UPDATE`):
    `tx, err := database.RequireTx(ctx)`
- Services own business rules, transaction boundaries, and error mapping.
  They use `s.runner.WithExecutor(ctx, func(ctx context.Context) error { ... })`
  for reads and `s.runner.WithTx(ctx, func(ctx context.Context) error { ... })`
  for writes. Services never extract the database connection or transaction
  directly and never pass explicit transaction or executor parameters.
- Context-Based Database Executor Propagation:
  - A single context key carries the active database executor (either the
    `*database.DB` connection pool or an active `*database.Tx`).
  - `WithExecutor(ctx, pool)` binds the pool if absent, or returns `ctx` untouched
    if an executor is already present.
  - `WithTx(ctx, tx)` binds the transaction to `ctx` under the same key.
  - Nested `WithTx` calls participate in the outer transaction without starting a
    new database transaction, creating savepoints, or committing/rolling back early.
    The outermost `WithTx` owns the transaction lifecycle.
- Writes inside `WithTx`: validate → hierarchy/constraint checks → mutate →
  reload, all on the derived transaction context. Do not validate on the pool
  and mutate on `tx`.
- Never run goroutines that share a held transaction or connection; a pgx
  connection is single-flight. Scan or close rows before issuing another query.
- Keep transactions short. No HTTP calls, uploads, email, Kafka publishes,
  sleeps, or other external I/O inside `WithTx`.
- Cross-domain data access: use a consumer-owned `*Reader` interface for
  ordinary reads, lookups, and projections (e.g. `ProductReader`, `VariantReader`,
  `InventoryReader`, `CartReader`, `UserReader`). Cross-domain mutations use
  a consumer-owned `*Writer` interface (`CartWriter`, `OrderWriter`, `StockWriter`,
  `TagWriter`) or inject another domain's `Service` only when the consumer must
  enforce that domain's business rules (validation, state transitions, limits,
  locking order, owned side effects). Declare the narrow interface in the consuming
  package and wire the concrete implementation in `internal/app`.
- All cross-domain reader and writer methods take `(ctx context.Context, ...)`.
  No explicit `QueryExecutor` or `*database.Tx` parameters.
- No dead `FooWithTx` twins: because transaction state propagates automatically
  via `context.Context`, methods do not have duplicated `WithTx` variants. A service
  method callable by both handlers and other services simply wraps its logic in
  `s.runner.WithTx(ctx, ...)`. When invoked from within an existing transaction, it
  transparently participates in the ambient transaction.
- Complete Construction and Fail-Fast Constructor Invariants: Every constructor
  (`NewService`, `NewHandler`, `NewRouter`, background workers, queues) must
  validate all required dependencies up-front and panic immediately if any
  required parameter is `nil` (e.g. `panic("domain: runner is required")`).
  Domain methods must never contain defensive per-call runtime nil checks
  (`if s.dep == nil`); objects are guaranteed to be completely constructed and
  valid from inception. Unit tests must use explicit fakes/doubles (`storage.NoopProvider`,
  `log.NewNoopLogger`, fake readers/runners) instead of passing `nil`.

## Storage and external effects

- Inject `storage.StorageProvider` directly. Resolve URLs with
  `GetPublicURL(bucket, key)`; store bucket + key, never the full URL.
- Upload before the DB transaction commits; delete the blob only after a
  confirmed rollback. On an ambiguous commit/timeout, leave the blob for the
  orphan sweeper rather than deleting a row's asset.
- For deletes, commit the DB state first, then delete the blob, or use the
  durable `deleting` state machine with a recovery worker.
- A timeout is an **unknown outcome**, not a failure. Make retries idempotent or
  deduplicated, and never auto-revert an intermediate state whose external
  effect may have succeeded.
- Every workflow that crosses storage, payments, email, or Kafka needs a state
  transition and a recovery path. State the failure matrix in the PR/answer.

## Errors and API responses

- Map driver errors to `apierr.APIError` in the service via a local
  `handleError`, using `database.MapError` and `errors.Is`/`errors.As`.
- Handlers must never branch on `pgx.ErrNoRows`, `*pgconn.PgError`, or error
  strings; they only return the typed error.
- Return user-safe messages and register stable codes in
  `internal/shared/errcode` when the resource has them.
- Do not leak SQL, table names, constraint names, or driver messages to clients.
- Optional lookups (for example a missing logo) treat `ErrNotFound` as absent
  and continue.

## HTTP, auth, and docs

- All routes live under `/api/v1`. Keep the public and admin surfaces separate.
- Admin handlers use an `Admin <Resource>` Swagger tag. The backend `Makefile`
  derives the public/admin tag split from those annotations; a missing or
  shared tag leaks or drops docs. Run `make swagger` after changing annotations.
- Preserve the middleware order and security posture in `internal/app` and
  `internal/shared/middleware` (auth, RBAC, CORS, rate limits, security headers,
  error handler, audit). Do not weaken CORS or cookie settings per endpoint.
- Public settings, maintenance mode, and locale behavior are cross-cutting; check
  `internal/setting` and `internal/shared/middleware/locale.go` before changing
  request flow.
- Validation belongs at the handler boundary (`internal/shared/validation`) plus
  the service's invariants. Back invariants with DB constraints and map
  violations, not just pre-checks.

## Migrations and data

- Add sequential migrations under `migrations/` with both `.up.sql` and
  `.down.sql`. Never edit a migration that has already been applied.
- Keep stock math consistent with the reservation lifecycle in `README.md`:
  `on_hand` = ledger sum, `reserved` = live holds, `available =
  max(on_hand - reserved, 0)`.
- Do not change seed data casually; it is part of the demo/development contract.
- Schema changes need a matching rollback and a note about backfill/expand-
  contract when the table is already live.

## Configuration and secrets

- Configuration is environment-only. Never commit `.env` values, credentials,
  peppers, or SMTP secrets.
- Add new variables to `.env.example` and `pkg/config`, with sane local defaults
  and a production note where it matters.
- Never log secrets or full request bodies containing tokens/OTP values.

## Testing

- Table-driven tests with `testify`; repository tests against real PostgreSQL
  through testcontainers when SQL/constraints matter.
- Service tests use fake implementations of the consumer-owned repository
  interfaces.
- Add a `MaxConns=1` regression test for hot paths that compose several queries;
  it fails loudly if a hidden second connection appears.
- Cover `handleError` for `ErrNotFound`, unique/FK/check violations, wrapped
  errors, already-mapped `APIError`, and context cancellation.
- Failure-injection tests for multi-system workflows: external failure,
  timeout-with-success (unknown outcome), and success-then-crash before the next
  DB step.
- Tests must not depend on execution order or a shared database row unless they
  create and clean it themselves.

## Skill routing

Global skills live in `~/.agents/skills/` and are loaded with the `skill` tool.
Load the most specific one before starting; do not restate its content here.

| Task | Skill |
| --- | --- |
| Session setup / deciding which workflow applies | `using-agent-skills` |
| New feature or behavior change | `test-driven-development` |
| Failing tests, broken build, unexpected behavior | `debugging-and-error-recovery` |
| API contracts, DTOs, module boundaries, REST design | `api-and-interface-design` |
| Auth, untrusted input, uploads, secrets, dependency audit | `security-and-hardening` |
| Query plans, N+1, pool/throughput work | `performance-optimization` |
| Logging, metrics, tracing, alerting, worker visibility | `observability-and-instrumentation` |
| Reviewing a diff before merge | `code-review-and-quality` |
| Removing/renaming APIs, columns, or domains | `deprecation-and-migration` |
| ADRs, endpoint docs, operational runbooks | `documentation-and-adrs` |
| High-stakes or irreversible change | `doubt-driven-development` |
| Large change that must land in slices | `incremental-implementation` |
| Commits, branches, PRs, releases | `git-workflow-and-versioning` |
| Enforcing coverage/lint/quality thresholds | `constraint-driven-development` |

## Definition of done

- `go build ./...`, `go test ./...`, and `go vet ./...` pass.
- `go test -race ./...` passes for concurrency-sensitive changes.
- No new `WithExecutor`/`WithTx` in helpers; no cross-domain service injection.
- Swagger regenerated if annotations/routes changed.
- Migration has a working down path when schema changed.
- Errors are typed and handled; no driver types reach handlers.
- External-side-effect failure paths are stated and covered.
