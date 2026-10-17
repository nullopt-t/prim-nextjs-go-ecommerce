# Connection-pool starvation: checking out a second connection inside a transaction

> Post-mortem and fix for the global API stalls caused by nested connection
> acquisition. Status: **fixed** (see [Verification](#verification)).

## TL;DR

`database.Executer.WithDB` hands back the pool, and `WithTx` checks a connection
out until the callback returns. Service methods that ran **inside a
transaction** called other service methods that were written for the pool, so
each of those calls asked the pool for a *second* connection while the first was
still held. With enough concurrent requests the 25-connection pool filled up
with transactions whose callers were waiting for another connection from that
same pool. Nothing could finish, so **every** database-backed endpoint stalled
for tens of seconds, even though nothing was slow or locked.

The fix has two layers:

1. **Transaction propagation (class-wide, in `pkg/database`).** `WithTx` binds
   the transaction to the context it hands the callback, and `WithDB` — plus any
   nested `WithTx` — reuses that transaction instead of acquiring a second
   connection. Every `WithTx` call site was moved to the new callback signature,
   so *any* database call made inside a transaction now joins it automatically.
2. **Explicit executor threading and shorter transactions (targeted).** The
   paths that were known to nest (`inventory`, `variant`, `product`, `user`,
   `address`, `cart`, `checkout`) also pass the caller's
   `database.QueryExecutor` into `...WithTx` methods, matching the convention
   already used by `ReserveStockWithTx` / `CreateOrderWithTx`, and the cart
   hydrator no longer runs inside transactions that do not need it.

Layer 1 removes the failure mode for the whole code base (including code written
later); layer 2 keeps the hot-path transactions short and does not depend on the
caller deriving its context.

## Impact

* Global stalls of 44–53 s across the whole API during cart-write bursts.
* 5xx responses on unrelated endpoints while the pool was saturated
  (`/products?limit=12`, `/categories`, `/brands`, `/user/notifications/unread-count`).
* `/cart` reads and checkouts (the money path) were affected too, because the
  same helper that hydrates cart items runs inside the checkout transaction.
* `/health` stayed at ~0.5 ms throughout, because it does not touch the
  database — which is what made this look like "the API randomly freezes".

## Root cause

### The pool

`backend/pkg/database/db.go`:

```go
poolConfig.MaxConns = 25
poolConfig.MinConns = 5
```

One 25-connection pool serves the whole process. `backend/pkg/database/runner.go`
as it was **before** the fix:

```go
// WithTx begins a transaction; the connection stays checked out until fn returns.
tx, err := t.db.Pool.Begin(ctx)
...
if err := fn(&Tx{tx}); err != nil { ... }

// WithDB runs fn against the pool, checking out a connection on demand.
func (t *Executer) WithDB(ctx context.Context, fn func(pool QueryExecutor) error) error {
	return fn(t.db)
}
```

Both are correct on their own. The bug is composing them.

### The pattern

```go
err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {   // connection #1
    ...
    stock, err := s.inventoryService.GetStockByVariantID(ctx, v.ID) // ← asks the pool again
    ...
})
```

and inside the inventory service:

```go
func (s *InventoryService) GetStockByVariantID(ctx context.Context, variantID uuid.UUID) (...) {
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {   // connection #2
		stock, repoErr = s.repo.GetStock(ctx, db, variantID)
		return repoErr
	})
	...
}
```

The caller holds connection #1 for the whole callback, then blocks waiting for
connection #2. Whenever the number of in-flight requests reaches `MaxConns`,
every connection is held by a transaction whose owner is waiting for another
connection from the same pool. That is a deadlock: it never resolves by itself,
it only clears when callers give up and their request contexts are cancelled
(which is why the server-side error surfaced as
`database operation canceled: timeout: context already done: context canceled` —
that is the *client* having hung up, not a query timeout).

Note that this is **not** lock contention and **not** a slow query: the blocked
connections sit in `idle in transaction` / `ClientRead` with no ungranted locks.
The database is idle; the application is the bottleneck.

### Where it happened

| In a transaction | Pool-backed call it made | Extra connections |
| --- | --- | --- |
| `CartService.AddItem` | `inventory.GetStockByVariantID` (×2) | 2 |
| `CartService.UpdateCartItemQuantity` | `inventory.GetStockByVariantID` | 1 |
| `CartService.MergeGuestCart` | `inventory.GetStockByVariantID` per item | N |
| `getOrCreateCartWithTx` (runs inside cart **and** checkout transactions) | `variant.GetVariantByID`, `product.GetByID`, `inventory.GetStockForVariants` per item | 2N + 1 |
| `CheckoutService.Checkout` transaction | `user.GetUserByID`, `address.GetAddress` | 2 |

The cart hydrator was the worst offender because it runs on **every** cart
response: add item, update quantity, merge guest cart, read cart, and inside
checkout. A cart with 3 lines needed 7 extra connections on top of the
transaction's own — while the transaction held one.

A naive scan of `WithTx` closure bodies misses most of these: the calls are one
or more service layers down. The reliable rule is *"a function that receives an
executor must pass it to every database-touching call it makes"*.

## How it was diagnosed

1. **Reproduce deterministically.** A single `POST /cart/items` returned `200` in
   ~40 ms; 30 concurrent adds with *distinct* guest sessions timed out, 29/30.
   That ruled out lock contention on a single cart (distinct sessions, distinct
   carts, so nothing was contending).
2. **Sample the database mid-burst.**

   ```sql
   select state, wait_event, count(*) from pg_stat_activity
   where datname = 'prim' group by 1, 2 order by 3 desc;
   ```

   All 25 connections reported `idle in transaction` / `ClientRead`, and
   `select * from pg_locks where not granted` was empty. The server was not busy
   and not blocked on locks — it was waiting for the application.
3. **Show the blast radius.** A single unrelated `GET /categories` issued during
   a cart burst took 4.65 s while `GET /health` stayed at 0.5 ms.
4. **Find the call sites statically.** Any function that receives a
   `database.QueryExecutor` but calls a service method without passing it is a
   suspect. That rule found exactly the five rows in the table above.

## The fix

**Rule:** *a function that has been given a `database.QueryExecutor` must pass it
to every database-touching call it makes. Never invoke a pool-backed
(`WithDB` / `ctx`-only) method from inside a transaction.*

### 1. Bind the transaction to the request context

The structural fix is in `pkg/database/runner.go`: `WithTx` publishes the
transaction on the context it passes to the callback, and both `WithDB` and a
nested `WithTx` reuse it.

```go
// txContextKey keys the in-flight transaction stored on a request context.
type txContextKey struct{}

func TxFromContext(ctx context.Context) (QueryExecutor, bool) {
	tx, ok := ctx.Value(txContextKey{}).(QueryExecutor)
	return tx, ok
}

func (t *Executer) WithTx(ctx context.Context, fn func(ctx context.Context, tx QueryExecutor) error) error {
	if t == nil || t.db == nil || t.db.Pool == nil {
		return ErrConnectionFailure
	}

	// Join the transaction already in flight for this context, if any.
	if currentTx, ok := TxFromContext(ctx); ok {
		return fn(ctx, currentTx)
	}

	tx, err := t.db.Pool.Begin(ctx)
	...
	wrappedTx := &Tx{tx}
	boundCtx := context.WithValue(ctx, txContextKey{}, QueryExecutor(wrappedTx))

	if err := fn(boundCtx, wrappedTx); err != nil {
		return MapError(err)
	}
	...
}

func (t *Executer) WithDB(ctx context.Context, fn func(pool QueryExecutor) error) error {
	if t == nil || t.db == nil {
		return ErrConnectionFailure
	}

	if currentTx, ok := TxFromContext(ctx); ok {
		return fn(currentTx)
	}

	return fn(t.db)
}
```

Consequences:

* A pool-backed service method called from inside a transaction now runs **on
  that transaction**, so the whole request is one connection from start to
  finish.
* The `Runner.WithTx` callback signature gained a `context.Context`, and all 50
  call sites were updated to receive it. The new parameter is named `ctx`, so it
  shadows the enclosing one: existing bodies needed no edits and downstream
  calls automatically carry the transaction.
* A *nested* `WithTx` joins the outer transaction instead of beginning a second
  one (**REQUIRED** semantics). A joined callback neither commits nor rolls back
  on its own — the outermost transaction decides.
* Code that must run outside the caller's transaction (for example audit writes
  that have to survive a rollback) has to ask for it explicitly with a detached
  `context.Background()`; `TxFromContext` makes that decision visible.

### 2. Executor-aware read methods

Each of these was split into a pool-backed public method, a `...WithTx` variant,
and one shared private core that does the real work on whichever executor it is
given:

| Package | Added |
| --- | --- |
| `catalog/inventory` | `GetStockByVariantIDWithTx`, `GetStockForVariantsWithTx` |
| `catalog/product/variant` | `GetVariantByIDWithTx`, `ListVariantMedia` |
| `catalog/product` | `GetByIDWithTx` |
| `user` | `GetUserByIDWithTx`, `GetAddressWithTx` |

For example:

```go
// Pool-backed entry point.
func (s *InventoryService) GetStockByVariantID(ctx context.Context, variantID uuid.UUID) (*model.InventoryStock, error) {
	var stock *model.InventoryStock
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		stock, repoErr = s.getStockByVariantID(ctx, db, variantID)
		return repoErr
	})
	if err != nil {
		return nil, err
	}
	return stock, nil
}

// In-transaction entry point: no second connection.
func (s *InventoryService) GetStockByVariantIDWithTx(ctx context.Context, tx database.QueryExecutor, variantID uuid.UUID) (*model.InventoryStock, error) {
	return s.getStockByVariantID(ctx, tx, variantID)
}
```

These explicit variants are redundant with propagation but not harmful, and they
match the convention this code base already used for its cross-transaction calls.
They also stay correct when a caller holds a transaction whose context it has
not derived (see [Known limitations](#known-limitations)).

### 3. Split the cart transaction

Hydration (variant, product, thumbnail, available stock) is presentation data.
The transaction now covers only the parts that need it, and hydration runs on an
explicit executor:

* `lockCart(ctx, tx, ...)` — resolve/create the cart and read its rows under
  `FOR UPDATE`. Cart rows only; no collaborator calls.
* `enrichCartWithTx(ctx, exec, cart)` — hydrate line items on the supplied
  executor.
* `enrichCart(ctx, cart)` — hydrate on the **pool**, after the transaction has
  committed.

Wiring:

```go
// Cart read paths: minimal transaction, then hydrate on the pool.
func (s *CartService) GetOrCreateCart(ctx, userID, sessionID) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		var txErr error
		cart, txErr = s.lockCart(ctx, tx, userID, sessionID)
		return txErr
	})
	if err != nil {
		return nil, err
	}
	return s.enrichCart(ctx, cart)
}

// Checkout: hydrate on the caller's transaction, because it prices the cart
// from live variant rows and must stay consistent with the locked cart.
func (s *CartService) GetOrCreateCartWithTx(ctx, tx, userID, sessionID) (*model.Cart, error) {
	cart, err := s.lockCart(ctx, tx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return s.enrichCartWithTx(ctx, tx, cart)
}
```

`ClearCartWithTx` no longer hydrates at all — it used to spend `2N + 1`
connections building item details that the caller immediately discarded.

### 4. Checkout uses the transactional lookups

`CheckoutService.Checkout` now reads the customer and shipping address on its own
transaction (`GetUserByIDWithTx` / `GetAddressWithTx`) instead of checking out
two more connections.

## Verification

**Mechanism.** A throwaway harness (`cmd/propcheck`, not committed) opens a pool
with `MaxConns = 1` — so a second concurrent checkout is guaranteed to hang —
and runs a nested `WithDB` through the transaction context:

```
PASS: nested WithDB reused the caller's transaction (MaxConns=1, no deadlock)
PASS: detached control starved on the single connection, as expected
```

The second line is the control: the same call made with a detached
`context.Background()` cannot find the transaction and does starve, which proves
the first line is not a false positive.

**Live load test** against the dev stack, same workload that previously wedged
the API:

| Scenario | Before | After |
| --- | --- | --- |
| 30 concurrent `POST /cart/items` | 29/30 timed out (20 s cap) | — |
| 120 concurrent `POST /cart/items` | — | **120/120 HTTP 200**, slowest 0.31 s |
| `GET /categories` during the burst | 4.65 s | **3.5 ms** |
| `GET /health` | 0.5 ms | 0.5 ms |

Integration suites: `cart`, `checkout`, `catalog/inventory`, `catalog/coupon`,
`catalog/review`, `wishlist` and `setting` all pass. `go build ./...` and
`go vet ./...` are clean.

> The `auth` suite currently fails on two assertions (`_prim_session` cookie
> name and the OTP resend limit). Those are pre-existing and unrelated: the
> session cookie is now `ps` (`middleware.SessionCookieName`), and
> `_prim_session` appears only in that stale test.

## Prevention

1. **Propagation is built in.** Prefer the ordinary pool-backed method inside a
   transaction; it joins automatically. Reach for a `...WithTx` variant only when
   you hold an executor that is not reachable from the context.
2. **Review rule.** Any call from inside a `WithTx` callback to a method that
   takes only a `ctx` is no longer a bug, *provided the call is given the
   callback's context*. Passing a stored or detached context is the thing to
   look for.
3. **Static check.** Locate functions that receive an executor but do not pass it
   to every service/repo call:

   ```bash
   # functions taking a QueryExecutor, and the service calls inside them
   grep -rn 'database.QueryExecutor' backend/internal --include=*.go | grep func
   ```

   A small AST/brace-matching script is more reliable and is worth wiring into
   CI; the two false positives to ignore are `object.GetObjectURL` (pure URL
   building, no database) and `*Repository` methods whose own executor parameter
   is named `db`.
4. **Do not hand a transaction-bound context to a goroutine.** `pgx.Tx` is not
   safe for concurrent use, so a goroutine that inherits the callback's context
   would now run queries on the parent's in-flight transaction. It must use a
   detached context (which then correctly takes its own connection).
5. **Degrade instead of hanging.** Add a pool acquire timeout
   (`poolConfig.ConnConfig.ConnectTimeout` / an acquire deadline) so saturation
   surfaces as fast `503`s rather than minute-long stalls, and size `MaxConns`
   with headroom over the peak number of *concurrent write paths*. Headroom only
   buys time — with propagation in place the nesting is bounded, but a bound is
   still worth having.

## Known limitations

* Propagation depends on the context being derived from the one `WithTx` hands
  the callback. A member of the call chain that switched to a fresh context (for
  example, one stored at start-up) falls back to the pool and can re-introduce
  the nesting. The explicit `...WithTx` methods on the hot paths are the safety
  net for exactly that case.
* A pool-backed read that used to run on the pool now participates in the
  caller's transaction and therefore sees its uncommitted rows. That is desired
  here (and how `getOrCreateCartWithTx` prices a cart), but it is a semantic
  change worth remembering when adding a query to an existing transaction.

## Follow-ups (optional)

* `CartService.AddItem` reads the same variant's stock three times (once before
  the transaction, twice inside it). Two of those are redundant.
* `MergeGuestCart` performs one inventory lookup per item inside a transaction;
  it could batch with `GetStockForVariantsWithTx` like the hydrator does.
