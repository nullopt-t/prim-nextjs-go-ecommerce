# Comprehensive Implementation Plan: Professional Merchant Gateway & Onboarding

## 1. Goal & Objectives

Transform the current disjointed vendor access into a dedicated, enterprise-grade merchant entry point, onboarding funnel, and lifecycle control system matching the caliber of platforms like Shopify and Amazon Seller Central.

### Key Deliverables:
1. **Public Merchant Acquisition Landing (`/sell`)**: A B2B acquisition page with value props, commission transparency (10%), payout schedules, and direct onboarding CTAs.
2. **Context-Aware Merchant Auth Gateway (`/vendor/auth` or `/vendor` gate)**: Isolated sign-in experience that guides merchants cleanly without customer cart/shopping clutter.
3. **Storefront-to-Merchant Onboarding (`/vendor/register`)**: Unblocking the `RequireRole` barrier so any authenticated user can apply to become a merchant with real-time slug verification.
4. **Lifecycle State Handlers**: Dedicated UI states for `pending` (under review with ETA), `suspended` (compliance freeze banner), and `rejected` (appeal/reapply).
5. **Admin Moderation & Approval Pipeline (`/admin/vendors`)**: Seamless review, approval, suspension, and commission rate configuration in the Admin Console.
6. **Storefront Linkage**: Navigation header and footer links to "Sell on PRIM / ابدأ البيع على بريم".

---

## 2. Architecture & Route Decomposition

```
  PUBLIC ENTRY                   AUTHENTICATED ONBOARDING           MERCHANT CONSOLE
  ────────────                   ────────────────────────           ────────────────
  
  [ Storefront / ]
         │
         ▼
  [ /sell (Landing) ] ──────────► [ /auth?redirect=/vendor ]
  "Sell with PRIM"                         │
                                           ▼
                                 [ /vendor/register ]
                                 (Store Name, Slug, Bio)
                                           │
                                           ▼ (POST /vendor/register)
                                 [ /vendor/pending ]
                                 "Application Under Review"
                                           │
                                           ▼ (Admin Approves via /admin/vendors)
                                 [ /vendor/dashboard ]
                                 (Active Merchant Console)
```

---

## 3. Step-by-Step Task Breakdown

### Phase 1: Routing & Guard Decoupling (Unblocking Registration)
- [ ] **Task 1.1**: Update [`frontend/src/shared/ui/requireRole.tsx`](file:///home/amr/tmp/prim/frontend/src/shared/ui/requireRole.tsx) or create a specialized `RequireVendorAccess` component.
  - *Problem*: `RequireRole minimumRole="vendor"` immediately redirects users with `role="customer"` to `/auth`, making self-serve registration impossible.
  - *Fix*: If authenticated as `customer`, allow access to `/vendor/register`. If unauthenticated, redirect to `/auth?redirect=/vendor/register`.
- [ ] **Task 1.2**: Update [`frontend/src/app/[locale]/(vendor)/vendor/layout.tsx`](file:///home/amr/tmp/prim/frontend/src/app/[locale]/(vendor)/vendor/layout.tsx) to mount the multi-state router (Registration -> Pending State -> Suspended State -> Active Console).

### Phase 2: Public "Sell on PRIM" Acquisition Page (`/sell`)
- [ ] **Task 2.1**: Create `frontend/src/app/[locale]/(shop)/sell/page.tsx` and module view `frontend/src/modules/shop/routes/sellPage.tsx`.
- [ ] **Task 2.2**: Include sections:
  - **Hero**: "Grow Your Brand Across Egypt & MENA" with primary CTA: *"Open Your Store"*.
  - **Why Sell on PRIM**: Low 10% flat commission, zero listing fees, automated weekly bank payouts.
  - **How it Works**: 3-step guide (1. Register Store -> 2. List Products -> 3. Ship & Earn).
  - **Merchant FAQ**: Payout timing, delivery logistics, required documents.
- [ ] **Task 2.3**: Add "Sell on PRIM" / "ابدأ البيع" CTA in the storefront main header and footer.

### Phase 3: Registration Funnel (`/vendor/register`)
- [ ] **Task 3.1**: Enhance [`frontend/src/modules/vendor/routes/vendorRegister.tsx`](file:///home/amr/tmp/prim/frontend/src/modules/vendor/routes/vendorRegister.tsx):
  - Form fields: Store Name, Auto-generated URL slug (with real-time slug preview: `prim.com/store/[slug]`), and Business Description.
  - Client-side validation: Lowercase alphanumeric regex check for slug, min/max length constraints.
  - Live API integration with `vendorService.register()`.
- [ ] **Task 3.2**: Add application success state transitioning to the Pending View.

### Phase 4: Lifecycle State Views
- [ ] **Task 4.1**: Build `VendorPendingView` component:
  - Display submission summary, store name, slug, and status badge (`قيد المراجعة / Under Review`).
  - Expected review timeframe (usually 24–48 hours).
  - Contact channel for merchant support.
- [ ] **Task 4.2**: Build `VendorSuspendedView` component:
  - Security/compliance freeze alert with instructions to submit compliance documents.

### Phase 5: Verification & End-to-End Validation
- [ ] **Task 5.1**: Test customer account registration (`test.merchant@example.com`):
  - Sign up as new user -> Navigate to `/sell` -> Click "Open Your Store" -> Complete `/vendor/register`.
  - Verify database creates vendor record in `pending` state with `commission_rate_bps = 1000`.
- [ ] **Task 5.2**: Test admin approval flow:
  - Sign in as `superadmin@prim.com` -> Open `/admin/vendors` -> Locate new pending vendor -> Click Approve.
  - Verify vendor state switches to `active`.
  - Verify vendor page refreshes to the full active Merchant Dashboard.
- [ ] **Task 5.3**: Build container and verify with zero TypeScript / lint errors.

---

## 4. Technical File Impact Matrix

| File Path | Action | Purpose |
| :--- | :--- | :--- |
| `frontend/src/app/[locale]/(shop)/sell/page.tsx` | **Create** | App router page for public seller acquisition |
| `frontend/src/modules/shop/routes/sellPage.tsx` | **Create** | Marketing & commercial terms landing view |
| `frontend/src/app/[locale]/(vendor)/vendor/layout.tsx` | **Modify** | Smart gateway router (Auth -> Register -> Pending -> Console) |
| `frontend/src/modules/vendor/routes/vendorRegister.tsx` | **Enhance** | Store application wizard with slug validation |
| `frontend/src/modules/vendor/routes/vendorPending.tsx` | **Create** | Status view while waiting for admin review |
| `frontend/src/modules/vendor/routes/vendorSuspended.tsx` | **Create** | Compliance freeze advisory view |
| `frontend/messages/en.json` & `ar.json` | **Update** | Bilingual copy for `/sell`, onboarding steps, and pending alerts |
| `frontend/src/shared/ui/header.tsx` | **Modify** | Add "Sell on PRIM" link in storefront header |
