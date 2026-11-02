# ADR: Multi-Vendor Domain (Design B — Vendor-Owned Products)

Date: 2026-11-02
Status: Accepted

## Context

Phase 4 of the v1 Beta plan introduces a full marketplace vendor domain,
replacing frontend mock data. During design, two marketplace shapes were
evaluated:

- **A — Offer model**: platform-owned catalog; vendors attach priced
  `vendor_offers` to variants; a buyer picks the seller on the PDP.
- **B — Vendor-owned products**: vendors own their own product listings;
  `products.vendor_id` is the single ownership column.

Option A was rejected for v1: it requires `cart_items.offer_id`, checkout
pricing/stock-holds per offer, and a buyer seller-picker UI — roughly 2-3x
the backend scope plus cart migration for existing carts.

## Decision

Design B:

1. Products are **vendor-owned**: new `vendors` domain, `products.vendor_id`
   (nullable; NULL = platform-owned legacy/admin data, no backfill).
2. A product is created by a vendor only through the existing product
   service, with an ownership check against a consumer-defined
   `VendorReader`.
3. Checkout attribution: each `order_item` is stamped with `vendor_id` and
   `vendor_earning` (item total minus commission) **at checkout time** —
   frozen snapshot semantics, identical in kind to `ProductSnapshot`. A later
   commission-rate change never rewrites past items.
4. `duplicate products across vendors` are allowed (two vendors may list
   the same item as separate products). If grouping becomes a real problem,
   a `canonical_product_id` grouping is added in a later additive migration.
5. Payouts are a ledger + state machine in `vendor_payouts`:
   `requested -> approved -> processing -> paid` (terminal), `rejected` only
   from `requested`/`approved` (post-`processing` rejection requires a
   confirmed-failed transfer note). All transitions are conditional updates
   (`WHERE status = expected`) — never blind retries.
6. Settlement: an item's earning is payout-eligible when its **order is
   delivered/completed**. The available balance is derived on read
   (SUM order_items.vendor_earning over delivered orders minus outstanding
   payouts); no separate ledger table. `RequestPayout` runs under
   `FOR UPDATE` on the vendor row inside one transaction so two concurrent
   requests cannot both claim the same balance.

## Consequences

- Cart, checkout flow, and stock holds are untouched beyond the new
  attribution stamps; existing carts and seed data remain valid.
- Pre-existing products are `vendor_id = NULL` and behave as platform items
  (no commission, no vendor lock).
- Same-product competition is not enforced/normalized by the platform;
  effective "buy box" pricing is the vendors' own responsibility in v1.
- Real bank payouts are out of scope: `MarkPayoutPaid` records a completed
  external transfer (reference number) replaced by an admin action; the
  state machine is shaped so a real provider idempotency key can slot in
  behind `processing` later.
