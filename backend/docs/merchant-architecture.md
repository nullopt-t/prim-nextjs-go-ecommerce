# Merchant Architecture: Gateway, Onboarding & Lifecycle Specification

## 1. Executive Summary

In a modern e-commerce platform, merchants (vendors) require dedicated access patterns, operational isolation, and lifecycle governance separate from retail customers. 

This document defines the end-to-end merchant architecture for PRIM, detailing:
- The distinction between customer and merchant mental models and access gates.
- The merchant onboarding and registration flow (`POST /api/v1/vendor/register`).
- The role state machine (`customer` → `vendor`) and profile status lifecycle (`pending` → `active` ⇄ `suspended` / `rejected`).
- Frontend routing, authorization guards, and public storefront separation.
- Administration review, commission assignment, and payout settlement rails.

---

## 2. High-Level Architecture & Entity Topology

```
                  ┌───────────────────────────────────────────────┐
                  │              Public Storefront                │
                  │   - Customers browse & checkout (/products)   │
                  │   - "Sell on PRIM" public landing (/sell)     │
                  └───────────────────────┬───────────────────────┘
                                          │
                        CTA: "Open Store" │ /vendor
                                          ▼
                  ┌───────────────────────────────────────────────┐
                  │             Identity & Auth Gate              │
                  │    POST /api/v1/auth/challenge/start          │
                  │    POST /api/v1/auth/challenge/verify         │
                  └───────────────────────┬───────────────────────┘
                                          │
                               Authenticated Session (ps)
                                          │
            ┌─────────────────────────────┴─────────────────────────────┐
            ▼                                                           ▼
   Role: `customer` (No Vendor Profile)                   Role: `vendor` (Has Vendor Profile)
   ┌─────────────────────────────────┐                    ┌─────────────────────────────────┐
   │    Merchant Registration Gate   │                    │     Vendor Lifecycle Guard      │
   │      (/vendor/register)         │                    │    GET /api/v1/vendor/profile   │
   │                                 │                    └───────────────┬─────────────────┘
   │ Form:                           │                                    │
   │  - Store Name                   │           ┌────────────────────────┼────────────────────────┐
   │  - URL Slug                     │           ▼                        ▼                        ▼
   │  - Business Description         │      `pending`                  `active`               `suspended`
   │                                 │  ┌──────────────┐          ┌──────────────┐        ┌──────────────┐
   │ POST /api/v1/vendor/register    │  │ Waiting for  │          │ Full Console │        │ Catalog      │
   └────────────────┬────────────────┘  │ Admin Review │          │ Access       │        │ Unlisted     │
                    │                   └──────────────┘          └──────────────┘        │ Payouts Held │
                    ▼                                                                     └──────────────┘
   Status: `pending` (Awaiting Admin)
```

---

## 3. The 3 Pillars of Merchant Separation

### Pillar 1: Public Acquisition (`/sell` Landing Page)
Merchants are business partners, not buyers. Their public entry point is a business-to-business (B2B) landing page highlighting:
1. **Audience & Reach**: Customer traffic, geographic delivery footprint.
2. **Commercial Model**: Standard 10% commission rate (`1000 bps`), zero upfront listing fees.
3. **Settlement Frequency**: Weekly settlement via bank transfer or direct wire.
4. **Tooling**: Built-in inventory tracking, low-stock warnings, and performance analytics.

### Pillar 2: Dedicated Gateway & Routing
When a merchant accesses `/vendor`:
- **Unauthenticated**: Redirected to `/auth?redirect=/vendor` with merchant-focused context ("Sign in to your merchant console").
- **Authenticated Customer (No Store)**: Directed to `/vendor/register` (Store Onboarding Form).
- **Authenticated Vendor (`active`)**: Mounted into the enterprise Vendor Console (`/vendor/dashboard`).
- **Authenticated Vendor (`pending`)**: Rendered with an "Application in Review" banner with read-only profile access.
- **Authenticated Vendor (`suspended`)**: Rendered with an account suspension advisory directing to merchant compliance support.

### Pillar 3: Tenant & Data Isolation
- Vendor catalog queries are scoped in database transactions: `WHERE p.vendor_id = $1`.
- A vendor cannot view or modify listings, stock levels, or earnings of other sellers.
- Inactive or suspended vendors have their products automatically filtered from public storefront catalog and checkout validation pipelines.

---

## 4. End-to-End Registration & Onboarding Lifecycle

### Step 1: Merchant Identity Setup
The prospective seller authenticates using passwordless email challenge:
1. `POST /api/v1/auth/challenge/start` with `{ "email": "merchant@business.com" }`.
2. `POST /api/v1/auth/challenge/verify` with `{ "email": "merchant@business.com", "code": "123456" }`.
3. An active session cookie (`ps`) is issued. The user account role is initially `customer`.

### Step 2: Store Registration Application
The authenticated user submits store details to the registration endpoint:

```http
POST /api/v1/vendor/register
Content-Type: application/json
Cookie: ps=<session-token>

{
  "storeName": "SoundMaster Studio",
  "slug": "soundmaster-studio",
  "description": "Authorized retailer of audiophile hardware, studio monitors, and accessories.",
  "logoObjectId": "10000000-0000-0000-0000-00000000000f",
  "bannerObjectId": "10000000-0000-0000-0000-00000000000e"
}
```

#### Backend Validation & Transaction Rules:
1. **Store Name**: Required, 2 to 150 characters.
2. **Slug**: Required, lowercase alphanumeric characters and hyphens (`^[a-z0-9]+(-[a-z0-9]+)*$`). Must be unique across all vendors (HTTP 409 Conflict if taken).
3. **Account Limit**: Exactly one vendor profile per user ID (enforced via database unique constraint on `vendors.user_id`).
4. **Initial Status**: Defaults strictly to `pending`.
5. **Initial Commission**: Populated from system configuration defaults (default: 10% = 1000 bps).
6. **User Role Upgrade**: The user's role in the `users` table is updated from `customer` to `vendor`.

### Step 3: Admin Review & Moderation
Admins manage and review incoming vendor applications in the Admin Console (`/admin/vendors`):

```http
PUT /api/v1/admin/vendors/{vendorId}/status
Content-Type: application/json
Cookie: ps=<admin-session>

{
  "status": "active"
}
```

- When status is transitioned to `active`:
  - Products created by this vendor become publishable to the public catalog.
  - The vendor receives full access to `/vendor/products`, `/vendor/inventory`, and `/vendor/payouts`.
- If rejected (`rejected`):
  - Application is closed with reason logged in the audit trail.
- If suspended (`suspended`):
  - Catalog items are instantly suppressed from storefront search and search indexing.
  - Cart checkout prevents purchases of the vendor's items.

---

## 5. State Machine & Status Matrix

| Status | Public Catalog Visibility | Can Add/Edit Products | Can Adjust Stock | Can Request Payouts | Console UI State |
| :--- | :---: | :---: | :---: | :---: | :--- |
| **`pending`** | Hidden | No (Draft Only) | No | No | Onboarding banner: *"Application Under Review"* |
| **`active`** | Visible | Yes | Yes | Yes | Full operational access |
| **`suspended`** | Hidden | No | No | No | Red compliance banner: *"Account Suspended"* |
| **`rejected`** | Hidden | No | No | No | Application declined state |

---

## 6. Frontend Architectural Guidelines

### Client-Side Guard Architecture
Rather than hard-failing in a layout guard, the vendor root layout operates as a state-based router:

```tsx
export function VendorRootLayout({ children }: { children: React.ReactNode }) {
  const { user, loading: authLoading } = useAuthContext();
  const { vendor, loading: vendorLoading, isRegistered, isActive } = useVendorContext();

  if (authLoading || vendorLoading) return <LoadingSpinner />;

  // 1. Unauthenticated -> Prompt Sign-In
  if (!user) {
    return <Redirect to="/auth?redirect=/vendor" />;
  }

  // 2. Authenticated user without vendor profile -> Registration Flow
  if (!isRegistered) {
    return <VendorRegisterView />;
  }

  // 3. Registered but pending approval
  if (vendor?.status === "pending") {
    return <VendorPendingApprovalView profile={vendor} />;
  }

  // 4. Suspended account
  if (vendor?.status === "suspended") {
    return <VendorSuspendedView profile={vendor} />;
  }

  // 5. Active verified merchant -> Full enterprise console
  return <VendorLayout profile={vendor}>{children}</VendorLayout>;
}
```

---

## 7. API Reference Summary

| Endpoint | Method | Role Required | Description |
| :--- | :---: | :---: | :--- |
| `/api/v1/vendor/register` | `POST` | Authenticated | Create a new pending vendor application |
| `/api/v1/vendor/profile` | `GET` | Authenticated | Fetch current authenticated vendor profile & status |
| `/api/v1/vendor/profile` | `PUT` | `vendor` (Active) | Update store name, bio, logo, or banner |
| `/api/v1/vendor/dashboard` | `GET` | `vendor` (Active) | Aggregated financial and catalog metrics |
| `/api/v1/vendor/products` | `GET` | `vendor` (Active) | List products belonging to current vendor |
| `/api/v1/vendor/products` | `POST` | `vendor` (Active) | Create a new vendor product listing |
| `/api/v1/vendor/inventory/adjust` | `POST` | `vendor` (Active) | Adjust variant on-hand inventory levels |
| `/api/v1/vendor/payouts` | `GET` | `vendor` (Active) | View settlement history and pending balance |
| `/api/v1/vendor/payouts` | `POST` | `vendor` (Active) | Request fund withdrawal |
| `/api/v1/admin/vendors/{id}/status` | `PUT` | `admin` | Approve, suspend, or reject a vendor profile |
