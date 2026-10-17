# PRIM Frontend

The web app for PRIM: the customer storefront, the account area, the admin back office, and the vendor portal. Built with Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS 4, and next-intl.

## Tech stack

- **Framework:** Next.js 16 (App Router, Turbopack)
- **UI:** React 19, Tailwind CSS 4, `lucide-react`, `react-icons`, `motion`
- **Internationalization:** `next-intl` — English and Arabic with automatic RTL
- **State:** React Context (`Auth`, `Cart`, `Catalog`, `Wishlist`, `Theme`, `Notifications`, `PublicSettings`)
- **Data:** a small `fetch`-based client (`src/api/client.ts`) plus `useFetch` / `useMutation` hooks
- **Feedback:** Sonner toasts, Recharts for admin dashboards

## What's here

### Storefront — `/`
Home, catalog, product detail, and category pages; a slide-over cart and a full cart page; wishlist; checkout with saved addresses, coupons, and settings-driven shipping, tax, and payment options. Informational pages too: about, contact, FAQs, shipping, returns, careers, and press.

### Account — `/user`
Overview, orders and order details, addresses, reviews, and wishlist, plus settings for profile, preferences, active sessions, saved payment methods, and a danger zone.

### Auth — `/auth`
Passwordless sign-in: enter an email, receive a one-time code, and verify it.

### Admin — `/admin`
Dashboard; products and variants; categories, brands, and tags; inventory with ledgers; media library; orders, tracking, and payments; promotions and redemptions; customers; store settings; and system settings.

### Vendor — `/vendor`
Dashboard, products, inventory, orders, and payouts. This section is a prototype backed by mock data (`src/features/vendor/data/mockVendorData.ts`); it isn't wired to a backend yet.

`/portal` is a small landing page that links to each of the three areas.

## Directory structure

```
frontend/
├── messages/                # en.json / ar.json
├── public/                  # static assets
└── src/
    ├── api/                 # fetch client with auth/redirect handling
    ├── app/[locale]/        # App Router groups: (shop) (auth) (dashboard) (admin) (vendor)
    ├── components/          # shared layouts and UI primitives
    ├── context/             # Auth, Cart, Catalog, Wishlist, Theme, Notifications, PublicSettings
    ├── features/            # feature modules, each with its own routes/ and components/
    ├── hooks/               # useFetch, useMutation, useAuth, useCart, useCatalog, useOrders, ...
    ├── i18n/                # routing and request configuration
    ├── lib/                 # helpers: pagination, datetime, roles, toasts
    ├── services/            # typed API modules (admin, auth, cart, catalog, orders, settings, vendor, ...)
    ├── styles/              # global and theme stylesheets
    ├── types/               # shared TypeScript types
    └── utils/               # countries, media helpers
```

## Running

**With Docker (recommended).** From the repository root:

```bash
make up
```

The app is available at <http://localhost:3000>.

**Locally.** From `frontend/`:

```bash
npm install
npm run dev      # http://localhost:3000
```

Create `.env.local` pointing at the API:

```env
NEXT_PUBLIC_API_URL=http://localhost:8081
API_URL=http://localhost:8081
```

`NEXT_PUBLIC_API_URL` is used by the browser and the server-side rewrite proxy (`/api/*` → backend); `API_URL` is the server-side fallback. In the browser, API calls go through same-origin `/api/*` paths so the session cookie behaves correctly.

## Scripts

| Command | Description |
| --- | --- |
| `npm run dev` | Start the dev server (Turbopack) |
| `npm run build` | Production build |
| `npm run start` | Serve the production build |
| `npm run lint` | Run ESLint |

## Notes

- **Locales:** `en` (default) and `ar`, with `localePrefix: "as-needed"` — so `/products` and `/ar/products` are both valid. Direction flips to RTL for Arabic automatically.
- **Public settings:** store identity, contact details, the announcement banner, shipping rules, payment availability, and maintenance mode are fetched from `GET /api/v1/settings/public` and exposed through `PublicSettingsContext`. The announcement banner, storefront maintenance gate, and footer contact details all read from it.
- **Auth:** sessions live in an HTTP-only cookie managed by the backend; the client redirects to `/auth` on a `401`.
