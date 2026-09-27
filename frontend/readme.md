# PRIM Frontend — Next.js 15 Storefront

The frontend application for PRIM, built with **Next.js 15 (App Router)**, **React 19**, **Tailwind CSS**, and **next-intl**.

---

## Features

- **App Router & Turbopack:** High-performance server and client rendering.
- **Bilingual & RTL:** Seamless English and Arabic support using `next-intl` (`/[locale]/...`).
- **State Management:** Modular React context architecture:
  - `CartContext`: Manages cart items, coupon validation, drawer visibility, and API sync.
  - `CatalogContext`: Manages product categories, catalog listing, search, and active filters.
  - `ThemeContext`: Handles dark/light theme switching with CSS variable mapping.
- **Interactive UI:** Smooth transitions, compact responsive design, and accessible navigation drawers.

---

## Directory Structure

```
frontend/
├── messages/               # Internationalization strings (en.json, ar.json)
├── public/                 # Static assets, fonts, icons, placeholders
└── src/
    ├── api/                # Axios API client configured with auth interceptors
    ├── app/                # Next.js App Router directory
    │   └── [locale]/
    │       ├── (shop)/     # Storefront pages (catalog, products, cart, checkout)
    │       ├── (auth)/     # Authentication flows
    │       └── (admin)/    # Store management dashboard
    ├── components/         # Shared global components (Header, Footer, Cards)
    ├── context/            # React context providers (Cart, Catalog, Theme)
    ├── features/           # Domain feature slices (home, products, cart, checkout)
    ├── hooks/              # Custom React hooks
    └── styles/             # Global Tailwind stylesheets
```

---

## Development Inside Docker (Recommended)

From the project root:
```bash
make up
```
The frontend will hot-reload at [http://localhost:3000](http://localhost:3000).

---

## Local Development (Without Docker)

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Configure environment:**
   Create `.env.local`:
   ```bash
   NEXT_PUBLIC_API_URL=http://localhost:8081
   ```

3. **Start development server:**
   ```bash
   npm run dev
   ```
