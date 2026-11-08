# Frontend Feature Gap Analysis: PRIM vs Magento Merchant Standard

Frontend perspective, audited against `docs/merchant-feature-benchmark.md` and the
current Next.js storefront (`frontend/src/modules/`, `frontend/src/app/[locale]/`).
Ranked by shopper/seller impact per unit of work.

---

## P0 — missing now, shoppers feel it

### 1. Faceted attribute filtering (color, size, specs)
- **Status:** storefront filters today are price / category / brand only
  (`src/modules/catalog/components/ui/`, `model/catalogContext.tsx`); no
  attribute facets (`grep facet|attribute.*filter` → empty).
- **Magento standard:** layered navigation over EAV attributes.
- **Why needed:** without it, a shopper who knows "red, size M" must scroll
  every result. This is the single biggest catalog conversion gap.

### 2. Guest order lookup (`/order-lookup`)
- **Status:** no such route (`grep order-lookup` → empty; routes under
  `src/app/[locale]/` have no guest lookup page).
- **Magento standard:** email + order-number lookup for guests.
- **Why needed:** every guest-support ticket that today requires an agent and
  the admin console can be a 20-line self-serve page.

### 3. Returns & refunds (RMA flow)
- **Status:** no refund/return/RMA UI anywhere (`grep refund|rma` → only false
  positives in clipboard/notify utils).
- **Magento standard:** RMA with request, approve, refund states.
- **Why needed:** order state machine reaches Delivered/Cancelled and then
  stops; post-delivery every exception is a phone call + manual DB work.

## P1 — high value, moderate scope

### 4. Notify-me-when-available
- **Status:** out-of-stock shows a toast only (`toast.unitsAvailable`);
  no subscription surface (`grep notify.*avail|back.in.stock` → empty).
- **Why needed:** silent demand loss; cheapest win is an email-subscription
  form on the PDP plus a Kafka consumer reusing the existing notifier.

### 5. Product compare (side-by-side)
- **Status:** none (`grep compare` → only price-compare and `localeCompare`).
- **Why needed:** spec-heavy categories (phones, TVs — this catalog) are
  researched on PRIM and bought elsewhere; compare is the standard retention
  pattern.

### 6. Synced browsing history
- **Status:** `prim:recently-viewed` in localStorage only
  (`src/shared/lib/recentlyViewed.ts`); backend table `user_product_views`
  referenced in the benchmark is not reflected in the client.
- **Why negligible→real:** history that vanishes on a new device actively
  trains shoppers not to rely on it; syncing is one endpoint + one context
  change.

### 7. Explicit related / upsell / cross-sell merchandising
- **Status:** recommendations are category/brand heuristic only
  (`grep related.*product|upsell|cross.sell` → empty).
- **Why needed:** AOV lever; Magento treats relations as first-class seller
  assets, PRIM treats them as a fallback algorithm.

## P2 — monetization & completeness

### 8. Volume / tier pricing + scheduled sales
- **Status:** base + compare-at price + coupons only
  (`grep tier.*pric|quantity.*discount` → empty; wizard has
  `compareAt` fields but no tiers).
- **Why needed:** B2B / bulk buyers and seasonal campaigns have no surface;
  scheduled sale windows also unlock campaign pages.

### 9. Product types: bundle, downloadable, virtual
- **Status:** seller wizard knows Simple + Configurable only
  (`grep bundle|downloadable` in `sellerProductWizard.tsx` → empty).
- **Why needed:** blocks digital goods entirely and bundle/KPI offers that
  Magento merchants take for granted.

### 10. Invoice PDF download
- **Status:** no invoice/PDF surface (`grep invoice.*pdf` → empty).
- **Why needed:** VAT/business customers expense-claim every purchase; B2B
  buyers without invoices churn silently.

## P3 — nice-to-have parity

| # | Feature | Why |
|---|---|---|
| 11 | Passkey / WebAuthn | Passwordless OTP already exists; passkeys harden the same flow without SMS cost. |
| 12 | Multi-currency switcher | Storefront is bilingual (EN/AR) but single-currency; cross-border shoppers see foreign prices. |
| 13 | Gift cards & gift wrap | Giftable categories convert better; Magento gap-listed, cheap to model. |

---

## Recommended order
11 → no, start at: **facets first** (highest conversion impact, backend
attribute schema already exists per the benchmark), then **guest lookup**
(cheapest support win), then **RMA** (required for any real returns policy),
then **notify-me** (smallest stateful feature). Compare and synced history
ride along on existing endpoints.
