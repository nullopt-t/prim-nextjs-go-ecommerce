# API Overview

PRIM exposes two documented REST surfaces under `/api/v1`:

- **Public** — storefront and customer account routes. No session required for catalog/cart; session required for account actions.
- **Admin** — back-office routes for catalog, orders, users, settings, media, and reports. Requires an authenticated admin session.

Both surfaces are documented with Swagger / OpenAPI and served as interactive Swagger UI by the API:

| Surface | Path | Access |
| --- | --- | --- |
| Public | `GET /api/v1/docs/public/index.html` | Open |
| Admin | `GET /api/v1/docs/admin/index.html` | Requires admin session cookie (`ps`) |

Regenerate both specs after changing handler doc annotations:

```bash
cd backend
make swagger
```

## Versioning

All routes are versioned under `/api/v1`. Breaking changes are shipped as a new version path rather than in-place modifications where practical. Non-breaking additive changes (new fields, new endpoints, optional query parameters) may land within `v1`.

## Requests & responses

- **Content-Type:** `application/json` for request bodies and responses.
- **Dates/times:** RFC 3339 strings.
- **Money:** minor units (e.g. piastres/cents) as integers.
- **IDs:** UUIDv4 strings.
- **Locale:** `Accept-Language` header (`en` / `ar`) drives bilingual responses; the admin can override per-resource where applicable.

## Errors

All errors return a consistent JSON envelope:

```json
{
  "error": {
    "code": "ERR_NOT_FOUND",
    "message": "product not found",
    "details": null
  }
}
```

| HTTP status | Meaning |
| --- | --- |
| 400 | Validation or bad request (`ERR_VALIDATION`, `ERR_BAD_REQUEST`) |
| 401 | Unauthenticated / invalid or expired OTP (`ERR_UNAUTHENTICATED`) |
| 403 | Forbidden / insufficient role (`ERR_FORBIDDEN`) |
| 404 | Resource not found (`ERR_NOT_FOUND`) |
| 409 | Conflict / duplicate (`ERR_CONFLICT`) |
| 422 | Unprocessable / business-rule violation (`ERR_UNPROCESSABLE`) |
| 429 | Rate limited (`ERR_RATE_LIMITED`) |
| 500 | Internal server error (`ERR_INTERNAL`) |

Error codes are stable strings; client code should branch on `code`, not on the human-readable `message`.

## Authentication

Public routes that act on behalf of a customer require a session cookie named `ps`, issued after successful OTP verification (`POST /auth/challenge/verify`). Admin routes require an admin session. See [`authentication.md`](authentication.md) for the full auth contract.

## Idempotency

State-changing endpoints that accept an `Idempotency-Key` header are safe to retry: the same key within the TTL returns the cached result without re-executing. Endpoints without an idempotency key are not guaranteed to be idempotent — check each endpoint's Swagger spec.

## Media uploads

`POST /admin/media` accepts raster images only (JPEG, PNG, GIF, WebP, AVIF) up to 10 MiB. The handler sniffs the real bytes rather than trusting the client's `Content-Type`, which rejects SVG and other script-bearing payloads that could cause stored XSS when served from the public bucket. Uploads are stored in MinIO under `MINIO_BUCKET` (`product-media` by default).

## Stock reservation lifecycle

Checkout never deducts stock by guessing: it creates a reservation row linked to the order, and that hold is resolved explicitly.

- **Cash on delivery** is confirmed when placed, so its reservation is committed immediately in the checkout transaction. It is never cancelled by the expiry reaper.
- **Online payment** holds stock for the configured reservation window. The hold is committed into a sale when an admin marks the payment `captured`; if the window lapses first, the reaper cancels the still-pending order and releases the hold.
- **Cancellation** releases a still-held reservation (unpaid order), or writes a `return` ledger entry when the sale was already committed (paid order).

Stock math: `available = max(on_hand - reserved, 0)`.
