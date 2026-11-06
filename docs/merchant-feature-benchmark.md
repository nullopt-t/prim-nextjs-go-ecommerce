# Feature Benchmark & Gap Analysis: Magento Open Source Reference

This document benchmarks PRIM against the standard features defined in the [Magento Open Source Merchant Guide](https://docs.magento-opensource.com/) to guide feature parity, roadmap priorities, and architectural decisions.

---

## 1. Catalog & Product Architecture

| Domain / Capability | Magento Open Source Standard | PRIM Current Status | Roadmap Gap / Next Steps |
|---|---|---|---|
| **Product Types** | Simple, Configurable, Grouped, Virtual, Bundle, Downloadable | Simple + Configurable (Matrix Variants with option combinations) | Digital/downloadable assets; bundle products. |
| **Catalog Attributes & Sets** | Dynamic EAV attribute sets, custom attributes per category/type | Normalized attribute schema (`attribute_definitions`, `attribute_values`, category mapping) | Dynamic attribute builders in merchant console. |
| **Inventory & Stock Management** | Multi-source inventory (MSI), backorders, low stock alerts, reservation locks | Single-warehouse inventory ledger, real-time stock reservations, low stock threshold flags | Multi-location/vendor warehouse routing; backorder policy per variant. |
| **Pricing Rules & Discounts** | Tiered/volume pricing, special pricing schedules, catalog price rules | Base price, compare-at price, promotion engine with coupon redemptions | Volume/quantity-tier discounts; scheduled sale windows. |
| **Media & Assets** | Role-based gallery (Base, Small, Thumbnail, Swatch) | S3/MinIO storage objects with CDN URL resolution, primary thumbnail + gallery objects | Swatch asset binding per variant attribute. |

---

## 2. Marketing, Promotions & Merchandising

| Domain / Capability | Magento Open Source Standard | PRIM Current Status | Roadmap Gap / Next Steps |
|---|---|---|---|
| **Promotions & Cart Rules** | Cart price rules (coupons, percentage/fixed discount, free shipping, conditions) | Fixed/percentage coupon discounts, free shipping thresholds | Complex conditional rules (e.g. buy X get Y, category-specific coupon restrictions). |
| **Cross-sell / Up-sell / Related** | Manual relations, rule-based related products, cart cross-sells | Category/brand-based recommendations | Explicit related & upsell product associations. |
| **Customer Engagement** | Wishlists, product reviews with ratings & voting, email alerts | Multi-item wishlists, verified purchase reviews with voting (`review_votes`) | Out-of-stock notification subscriptions ("Notify me when available"). |
| **Search & Navigation** | Layered navigation, faceted search, synonyms, search suggestions | Price/category/brand filtering, full-text search across titles/descriptions | Faceted attribute filtering (color, size, specs) in catalog search. |

---

## 3. Order Processing & Fulfillment

| Domain / Capability | Magento Open Source Standard | PRIM Current Status | Roadmap Gap / Next Steps |
|---|---|---|---|
| **Order Lifecycle State Machine** | Pending -> Processing -> Complete / Cancelled / Closed / On Hold | Clean state machine: Pending -> Processing -> Shipped -> Delivered / Cancelled | Return merchandise authorization (RMA) & refund workflows. |
| **Invoicing & Shipments** | Split invoices, partial shipments, tracking number attachments | Order item status tracking, merchant line-item allocation, frozen commission capture | Carrier tracking link generator; multi-shipment dispatch per vendor order. |
| **Order Status Audit History** | Append-only status change log with admin/automated comments | `order_status_history` table with timestamped status changes and notes | Expose status timeline directly in the customer tracking portal. |
| **Guest to Customer Lifecycle** | Seamless guest checkout, convert guest order to registered customer account | Guest cart merging upon OTP login; authenticated checkout | Guest order lookup portal (`/order-lookup`) with email + order number. |

---

## 4. Multi-Vendor & Merchant Operations

| Domain / Capability | Magento Open Source Standard *(Enterprise/Extension)* | PRIM Native Implementation | Roadmap Gap / Next Steps |
|---|---|---|---|
| **Vendor Onboarding & Status** | Extension-based (Marketplace modules) | Core lifecycle state machine: `pending` -> `active` / `rejected` / `suspended` | Document verification attachment uploads (tax ID, commercial register). |
| **Catalog Partitioning** | Vendor ownership of products & inventory | Strict merchant tenancy: vendors manage only their own products & stock movements | Vendor-specific shipping fee matrices. |
| **Settlement & Payouts** | Periodic ledger balance with commission deductions | Double-entry balance calculation, payout requests with idempotency keys, admin transition queue | Automated settlement schedule via payment gateway splits. |

---

## 5. Customer Account & Storefront UX

| Domain / Capability | Magento Open Source Standard | PRIM Current Status | Roadmap Gap / Next Steps |
|---|---|---|---|
| **Authentication & Security** | Passwords with reset links, 2FA extensions, reCAPTCHA | Passwordless OTP challenge/verify (email), brute-force rate limiting, HttpOnly sessions | Passkey / WebAuthn support. |
| **Customer Addresses** | Multiple billing and shipping addresses with default selectors | Full address book with default shipping/billing flags | Map coordinate integration for delivery geofencing. |
| **Browsing History** | Recently viewed products stored per customer account | LocalStorage based (in progress: database-backed sync) | Multi-device synced browsing history (`user_product_views`). |
| **Localization & Multi-store** | Multi-store / multi-currency / multi-language root hierarchy | Bilingual (English/Arabic) with full RTL layouts, Next.js App Router subpaths | Multi-currency switcher with live exchange rate tables. |

---

## 6. Architecture & Integration Matrix

| Architectural Principle | Magento Open Source | PRIM Architecture |
|---|---|---|
| **Stack** | PHP 8.x + MySQL + Elasticsearch | Go 1.24 (Gin) + PostgreSQL (pgx) + Redis + Kafka |
| **Client** | Luma / Hyvä (Server-rendered HTML / Alpine / Tailwind) | Next.js 16 (App Router) + TypeScript + Tailwind CSS |
| **Coupling** | Monolithic plugin system (plugins/observers/interceptors) | Decoupled REST API + Clean Architecture + Event-driven Kafka workers |
| **Modifications & Extensibility** | XML dependency injection & class rewrites | Explicit dependency injection via Go constructors |
