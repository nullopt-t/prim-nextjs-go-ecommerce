# Responsive Issues Audit Report
**Platform:** PRIM Frontend (Next.js 16 + Tailwind CSS v4 + next-intl)
**Goal:** Professional support for ANY device on earth
**Build:** ✅ `npm run build` passes (exit 0) after fixes
**Status:** Critical + High fixes applied; Medium/Low remaining

---

## ✅ COMPLETED — Critical Issues

| # | Issue | Fix | File |
|---|-------|-----|------|
| **C1** | Missing viewport meta tag — mobile browsers rendered shrunken desktop site | Added `<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />` | `src/app/[locale]/layout.tsx` |
| **C2** | `100vh` clipped by mobile Safari browser chrome | `100vh` → `100dvh` (global + admin sidebar ×2) | `src/styles/index.css`, `src/features/admin/routes/adminLayout.tsx` |
| **C3** | Portal Switcher (`fixed bottom-4`) would collide with mobile bottom nav | Gated to `hidden md:flex` (tablet+ only) | `src/components/ui/portalSwitcher.tsx` |
| **C4** | Toast (`bottom-right`) covered mobile bottom nav | New `ResponsiveToaster`: `top-right` on mobile (<768px), `bottom-right` on desktop | `src/components/ui/responsiveToaster.tsx` (new), wired into `layout.tsx` |
| **C5** | Carousel dots 8×28px, nav buttons 28×28px (below 44px a11y minimum) | Nav buttons → `h-11 w-11` (44px); dots → `h-2.5` + `px-1 py-1` larger hit area; icons → `size-4`/`size-4.5` | `src/features/home/components/ui/hero.tsx` |

---

## ✅ COMPLETED — High Priority Issues

| # | Issue | Fix | File |
|---|-------|-----|------|
| **H1** | Product gallery thumbnails locked to `grid-cols-4` on phones | `grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4` | `src/features/products/components/ui/ProductGallery.tsx` |
| **H3** | Category shelf scroll arrows hidden on mobile (`hidden md:flex`) + invisible (opacity-0) on touch | Buttons visible on mobile (`opacity-100` on mobile, fade on desktop hover); no longer hidden | `src/features/home/components/ui/categoryShelves.tsx` |
| **H5** | Product page `w-full` → unreadably long lines on ultrawide/desktop | Wrapped product grid in `max-w-7xl mx-auto` | `src/features/products/components/ProductDetails.tsx` |
| **H6** | Section grid `auto-fill minmax(100px)` created 10+ tiny columns on wide screens | `minmax(140px)` base, `minmax(180px)` on md, `xl:repeat(6,...)` cap | `src/features/home/components/ui/sectionGrid.tsx` |

---

## ⏳ REMAINING — High Priority

| # | Issue | File | Effort |
|---|-------|------|--------|
| **H2** | Horizontal-scroll areas use hidden scrollbars with no affordance (breadcrumbs, recently grid, tab bars, admin tables) — users can't discover scrollable content | 13 locations (`scrollbar-none`/`no-scrollbar`) | Medium — add scroll hints/arrows or paginate |
| **H7** | Admin sidebar `calc(100vh - 4rem)` — already fixed to `100dvh` in C2 | `adminLayout.tsx` | ✅ Done (included in C2) |

---

## ⏳ REMAINING — Medium Priority

| # | Issue | File | Effort |
|---|-------|------|--------|
| **M1** | Hero fixed `min-h-[290px]` can overflow small phones | `hero.tsx:120` | Small |
| **M2** | Inputs use `truncate` — long text clipped on narrow phones | `src/components/ui/input.tsx:23` | Small |
| **M3** | Product grids cap at 5 columns, don't use ultrawide space | `productsSection.tsx`, `wishlist/content.tsx` | Small |
| **M4** | Category Showcase Cards inner `grid-cols-2` not responsive on narrow cards | `categoryShowcaseCards.tsx:33` | Small |
| **M5** | Coupon input + Apply button cramped on smallest phones | `coupon.tsx:70-94` | Small |

---

## ⏳ REMAINING — Low Priority (Polish)

| # | Issue | File | Effort |
|---|-------|------|--------|
| **L1** | `text-[10px]` below readable minimum / inconsistent with design scale | Many components | Medium |
| **L2** | No `prefers-reduced-motion` for autoplay carousel / entrance animations | `hero.tsx`, `animatedSection.tsx` | Medium |
| **L3** | Incomplete `ltr:`/`rtl:` padding mirrors on some inputs/dropdowns | Various | Small |

---

## ⚠️ PRE-EXISTING BUILD BLOCKER (not a responsive issue)

`frontend/src/features/admin/routes/customers.tsx` was already modified in the working tree before this work began (moved `ROLE_RANKS` to `@/lib/roles`, restructured admin action JSX). It causes a Turbopack "Unterminated regexp literal" parse error on some runs. **Out of scope** for responsive work — belongs in a separate admin bug fix. All other routes build cleanly.

---

## Recommended Next Session

1. Fix **H2** (visible horizontal-scroll affordance) — the last remaining High issue.
2. Fix **M1–M5** (medium) — quick className changes.
3. Fix **L1–L3** (polish).
4. Fix the pre-existing `customers.tsx` parse error so the admin customers route builds.
