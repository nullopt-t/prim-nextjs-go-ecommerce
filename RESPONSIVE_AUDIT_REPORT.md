# Responsive Issues Audit Report
**Platform:** PRIM Frontend (Next.js 16 + Tailwind CSS v4 + next-intl)
**Goal:** Professional support for ANY device on earth (mobile, tablet, foldable, desktop, ultrawide, TV)
**Status:** ⚠️ NOT yet device-agnostic — critical gaps exist

---

## Executive Summary

The frontend uses Tailwind CSS v4 with sensible breakpoint prefixes (`sm:`/`md:`/`lg:`) in many places, but it is **not production-ready for "any device on earth."** There are multiple critical issues: a missing viewport meta tag (breaks mobile rendering entirely), viewport unit misuse on mobile Safari, overlapping fixed-position UI, hidden-scroll affordance problems, and undersized touch targets.

Below is a severity-ranked inventory of every responsive issue found, with file locations and recommended fixes.

---

## CRITICAL — Blocks correct rendering on mobile devices

### C1. Missing viewport meta tag
**Severity:** Critical  
**Location:** `frontend/src/app/[locale]/layout.tsx` (root `<html>`)  
**Issue:** There is no `<meta name="viewport" content="width=device-width, initial-scale=1">` anywhere in the app (confirmed via repo-wide grep — zero matches). Without it, mobile browsers render the page at desktop viewport width and shrink-to-fit, defeating ALL `sm:`/`md:`/`lg:` Tailwind breakpoints. The site will look like a shrunken desktop site on every phone.  
**Fix:** Add to root layout:
```tsx
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
```
`viewport-fit=cover` also supports notch/dynamic-island devices.

---

### C2. `100vh` used instead of `100dvh` on `<body>`/`<html>` (mobile Safari)
**Severity:** Critical  
**Location:** `frontend/src/styles/index.css:8` — `min-height: 100vh;`  
**Issue:** On mobile Safari (iOS), `100vh` does not account for the collapsing/expanding browser chrome. Content at the bottom of the page can be permanently hidden behind the address bar or tab bar. Note: the cart sidebar correctly uses `100dvh` (lines 40, 47 of `cartSidebar.tsx`), but the global layout does not.  
**Fix:** Use `dvh` (dynamic viewport height) in `index.css`:
```css
html, body { min-height: 100dvh; }
```

---

### C3. Portal Switcher overlaps Mobile Bottom Navigation
**Severity:** Critical  
**Location:** `frontend/src/components/ui/portalSwitcher.tsx:49`  
**Issue:** `<PortalSwitcher />` is rendered globally (inside root layout context) with `fixed bottom-4 left-4`. The mobile bottom nav (`MobileNav`, `mobileNav.tsx:32`) is `fixed bottom-0 inset-x-0`. On mobile these two fixed elements collide — the portal button sits directly on top of the bottom nav bar, making both unreachable or unusable.  
**Fix:** Hide PortalSwitcher on mobile, or move it above the nav (`bottom-20` on mobile), or gate it with `md:` only:
```tsx
{/* only show on tablet+ */}
className="md:flex fixed bottom-4 left-4 ..."
```

---

### C4. Toast notification overlaps Mobile Bottom Navigation
**Severity:** Critical  
**Location:** `frontend/src/app/[locale]/layout.tsx:88` — `<Toaster position="bottom-right" />`  
**Issue:** Sonner toaster is pinned `bottom-right` globally. On mobile this sits exactly where the bottom navigation bar lives (`mobileNav.tsx:32` `fixed bottom-0`), so toasts will cover the entire mobile nav and block navigation.  
**Fix:** Make toast position responsive — `bottom-right` on desktop, `top-right` (or `top-center`) on mobile:
```tsx
<Toaster position="top-right" richColors />  {/* or use a responsive wrapper */}
```
Alternatively conditionally hide/offset toaster when `MobileNav` is present.

---

### C5. Touch targets below accessible minimum (44×44 CSS px)
**Severity:** Critical  
**Location:** Multiple files  
**Issue:** WCAG / Apple HIG recommend ≥44×44 px touch targets. Several interactive elements are smaller:
- Carousel pagination dots: `h-2 w-7` → 8×28 px (`hero.tsx:191`)
- Carousel playback/nav buttons: `size-7` → 28×28 px (`hero.tsx:207-225`)
- These are too small to reliably tap on phones, especially for users with motor impairments.

**Fix:** Increase to ≥44px and/or add generous padding:
```tsx
// dots
className="h-2.5 w-7 ... min-[touch]:w-10"
// buttons
className="size-10 ..." // 40px, or size-11 (44px)
```

---

## HIGH — Breaks usability on tablets, small laptops, or narrow viewports

### H1. Product Gallery thumbnails locked to 4 columns on all screen sizes
**Severity:** High  
**Location:** `frontend/src/features/products/components/ui/ProductGallery.tsx:39`  
**Issue:** `<div className="grid grid-cols-4 gap-4">` forces exactly 4 thumbnail columns even on narrow phones, squeezing thumbnails to ~60px each with no responsive reduction.  
**Fix:** `grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4`

---

### H2. Horizontal-scroll areas use hidden scrollbars with no affordance
**Severity:** High  
**Location:**  
- `categoryShelves.tsx:77` — `overflow-x-auto no-scrollbar scroll-smooth`
- `recentlyGrid.tsx:19` — `overflow-x-auto ... scrollbar-none`
- `ProductDetails.tsx:366` — breadcrumbs `overflow-x-auto whitespace-nowrap scrollbar-none`
- `userTopTabs.tsx:26` — `overflow-x-auto scrollbar-none`
- `settingsTopTabs.tsx:19` — `overflow-x-auto scrollbar-none`
- `orderDetailsTopNav.tsx:23` — `overflow-x-auto scrollbar-none`
- `settingsShell.tsx:66` — settings sub-nav `overflow-x-auto scrollbar-none`

**Issue:** `scrollbar-none`/`no-scrollbar` hides the scrollbar entirely, and there is no visual hint (arrow, gradient fade, "scroll" text) that more content exists horizontally. Users on touch devices may discover it by accident; on mouse devices they cannot scroll at all without a visible scrollbar.  
**Fix:** Keep a thin visible scrollbar, or add scroll indicators / "scroll left/right" buttons, or paginate instead of infinite horizontal scroll.

---

### H3. Category Shelf carousel navigation arrows hidden on mobile
**Severity:** High  
**Location:** `frontend/src/features/home/components/ui/categoryShelves.tsx:56-72`  
**Issue:** Scroll buttons are `hidden md:flex` — completely absent on phones. Combined with hidden scrollbars (H2), mobile users have no way to know more products exist beyond the visible row.  
**Fix:** Show buttons on mobile too (`flex` instead of `hidden md:flex`), and add a visible scroll affordance.

---

### H4. Breadcrumbs overflow horizontally on mobile with no scroll hint
**Severity:** High  
**Location:** `frontend/src/features/products/components/ui/ProductDetails.tsx:366`  
**Issue:** Breadcrumb nav uses `overflow-x-auto whitespace-nowrap scrollbar-none`. Long product/category names on narrow phones overflow and are unreadable, with no indication more content is scrollable.  
**Fix:** Allow wrapping (`whitespace-normal` / `break-words`) or truncate the last segment with a visible scroll indicator.

---

### H5. Product detail page has no max-width on ultrawide / desktop
**Severity:** High  
**Location:** `frontend/src/features/products/components/ProductDetails.tsx:364` — `<AnimatedSection className="w-full">`  
**Issue:** The product page stretches `w-full` with no `max-w-*` constraint. On 27"+ monitors / ultrawides the 2-column grid (`grid-cols-1 md:grid-cols-2`) becomes too wide to read comfortably (line lengths exceed ~80–100 chars).  
**Fix:** Wrap content in `max-w-7xl mx-auto` (as used in `wishlist/content.tsx:84`) or constrain the grid gap/columns at `xl:`/`2xl:`.

---

### H6. Section Grid uses `auto-fill` with 100px minimum — too many tiny columns on desktop
**Severity:** High  
**Location:** `frontend/src/features/home/components/ui/sectionGrid.tsx:23` — `grid-cols-[repeat(auto-fill,minmax(100px,1fr))]`  
**Issue:** On wide screens `auto-fill` can create 10+ narrow columns, each only ~100px wide. Category names wrap awkwardly and cards become cramped. No `xl:`/`2xl:` cap.  
**Fix:** Cap columns:
```tsx
grid-cols-[repeat(auto-fill,minmax(140px,1fr))]
md:grid-cols-[repeat(auto-fill,minmax(180px,1fr))]
xl:grid-cols-[repeat(6,minmax(0,1fr))]
```

---

### H7. Admin sidebar height uses `calc(100vh - 4rem)` — mobile Safari clipping
**Severity:** High  
**Location:** `frontend/src/features/admin/routes/adminLayout.tsx:222` — `h-[calc(100vh-4rem)]` and line 226 `max-h-[calc(100vh-8rem)]`  
**Issue:** Same `100vh` mobile Safari problem as C2 — admin sidebar can be clipped by browser chrome on phones/tablets.  
**Fix:** Use `dvh`: `h-[calc(100dvh-4rem)]`.

---

### H8. Cart sticky summary on desktop only — no sticky on tablets (md/lg gap)
**Severity:** High  
**Location:** `frontend/src/features/cart/components/layout/cartLayout.tsx:17-24`  
**Issue:** `lg:flex-row` means between `md` (768px) and `lg` (1024px) the cart stacks full-width; the payment summary (`sticky top-24`) is only sticky on `lg+`. On iPad/tablet in portrait (~768–1024px), the summary scrolls away. Acceptable but suboptimal for "any device."  
**Fix:** Consider `md:flex-row` for tablet or keep stacked but ensure summary stays accessible.

---

## MEDIUM — Degrades experience but doesn't block core tasks

### M1. Fixed hero min-heights can overflow on small phones
**Severity:** Medium  
**Location:** `frontend/src/features/home/components/ui/hero.tsx:120` — `min-h-[290px] sm:min-h-[320px]`  
**Issue:** Fixed `min-h` on the hero carousel. On very small phones (e.g. iPhone SE ~375px wide, short viewports) plus large typography (`lg:text-[42px]`), content can overflow the container or require excessive scrolling.  
**Fix:** Use `min-h-[200px] sm:min-h-[260px] md:min-h-[320px]` and let content drive height (`min-h-auto`) rather than fixed minimums.

---

### M2. Inputs use `truncate` class — text cut off instead of scrolling
**Severity:** Medium  
**Location:** `frontend/src/components/ui/input.tsx:23` — `className="... truncate w-full ..."`  
**Issue:** `truncate` applies `overflow: hidden; text-overflow: ellipsis; white-space: nowrap;` to `<input>`. On narrow mobile screens, long user input is visually clipped with no way to see the end (no horizontal scroll inside input). Bad for email/address entry.  
**Fix:** Remove `truncate`; let input text wrap or scroll naturally. Use `max-w-full` instead.

---

### M3. Missing `xl:`/`2xl:` breakpoints on several product grids
**Severity:** Medium  
**Location:**  
- `productsSection.tsx:27` — `grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5` (caps at 5 cols)
- `wishlist/content.tsx:176` — same cap at 5
- `content.tsx:104` — `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5` (better)

**Issue:** On ultrawide/desktop (1440px+) product cards remain small (5 columns) instead of utilizing available screen real estate. Cards get tiny on very wide screens.  
**Fix:** Add `xl:`/`2xl:` variants or constrain with `max-w-7xl mx-auto`.

---

### M4. Category Showcase Cards internal `grid-cols-2` not responsive
**Severity:** Medium  
**Location:** `frontend/src/features/home/components/ui/categoryShowcaseCards.tsx:33` — `<div className="grid grid-cols-2 gap-2.5">`  
**Issue:** When `top4.length >= 4`, it always renders a 2-column product grid inside each card, even when the card itself is narrow (e.g. on `sm` breakpoint where the outer card is full-width but still constrained). Images become very small.  
**Fix:** Make inner grid responsive: `grid-cols-1 sm:grid-cols-2`.

---

### M5. Coupon input + apply button cramped on smallest phones
**Severity:** Medium  
**Location:** `frontend/src/features/cart/components/ui/coupon.tsx:70-94`  
**Issue:** `flex gap-4` row with `flex-1` input and `w-28 md:w-36` button. On phones narrower than ~320px the Apply button (`w-28` = 112px) plus input leaves almost no room; text may wrap/clip.  
**Fix:** Stack on smallest breakpoint: `flex flex-col sm:flex-row gap-4 sm:gap-0` and `w-full sm:w-28` on button.

---

### M6. Admin top bar user info hidden below `sm`, but avatar initials may wrap
**Severity:** Medium  
**Location:** `frontend/src/features/admin/routes/adminLayout.tsx:116` — `hidden sm:flex flex-col`  
**Issue:** The user name/email block is hidden until `sm`. Between `xs` and `sm` only the avatar initials show. Fine, but on foldables/small tablets at exactly the breakpoint boundary it can feel abrupt.  
**Fix:** Acceptable; no change strictly needed.

---

### M7. `px-3 sm:px-6 lg:px-8` page padding may be excessive on small phones
**Severity:** Medium  
**Location:** `mainLayout.tsx:13`, `userLayout.tsx:56`, etc.  
**Issue:** `px-3` (12px) on mobile is quite tight; content nearly touches screen edge on some sections (e.g. product cards `rounded-2xl` + shadow). Apple/Google recommend 16–20px horizontal padding on phones.  
**Fix:** Consider `px-4 sm:px-6 lg:px-8` for main layout.

---

## LOW — Polish / consistency

### L1. Inconsistent font-size scale as low as 10px
**Severity:** Low  
**Location:** theme.css defines `--font-size-txt-sm: 14px`; many components use `text-[10px]`, `text-[11px]`, `text-xs` directly.  
**Issue:** 10px text is below readable minimums for some users and below the design system's own scale. Appears in badges, timestamps, helper text across many components.  
**Fix:** Replace `text-[10px]` with `text-txt-sm` (14px) or at minimum `text-[12px]`.

---

### L2. No `prefers-reduced-motion` handling for animations
**Severity:** Low  
**Location:** `animatedSection.tsx` (motion), `hero.tsx` (carousel autoplay), various `animate-*` classes.  
**Issue:** Motion/reduced-motion media query not respected; autoplay carousels and entrance animations can trigger vestibular disorders.  
**Fix:** Wrap motion in `useReducedMotion` or respect `prefers-reduced-motion`.

---

### L3. RTL layout adjustments incomplete in a few spots
**Severity:** Low  
**Location:** Various — e.g. `notificationsDropdown.tsx:110` handles RTL `left-0`/`right-0`, but `categoryShelves.tsx:60-71` uses `ltr:left-2 rtl:right-2` correctly; mostly consistent. Some components lack explicit `ltr:`/`rtl:` padding overrides (e.g. `input.tsx` uses `ltr:`/`rtl:` but `coupon.tsx` input placeholder padding not mirrored).  
**Fix:** Audit for `ltr:`/`rtl:` completeness, especially padding/margin on inputs and dropdowns.

---

### L4. Portrait-only assumptions in admin dashboard
**Severity:** Low  
**Location:** `adminLayout.tsx` — sidebar `fixed lg:sticky` assumes desktop-first; on tablets in landscape the sidebar may consume too much vertical room.  
**Fix:** Test iPad landscape; consider collapsible sidebar on `md` for tablets.

---

### L5. Search bar hidden below `md` (deliberate, but verify discoverability)
**Location:** `header.tsx:23` — `<SearchBar />` wrapped in `hidden md:block`  
**Issue:** This is a deliberate pattern (mobile uses `MobileSearchDialog` via `mobileSearchDialog.tsx:83` `md:hidden` trigger). It's acceptable UX, but the mobile dialog trigger is an unlabeled icon button; ensure `aria-label` is present (it is: `aria-label={t("search")}`).  
**Fix:** No change needed; verified accessible.

---

## Recommended Remediation Priority Order

| Priority | Issue | Effort | Impact |
|----------|-------|--------|--------|
| 1 | Add viewport meta tag (C1) | 5 min | Critical |
| 2 | `100vh` → `100dvh` globally & admin (C2, H7) | 10 min | Critical |
| 3 | Portal switcher vs mobile nav overlap (C3) | 15 min | Critical |
| 4 | Toast vs mobile nav overlap (C4) | 10 min | Critical |
| 5 | Touch targets ≥44px (C5) | 30 min | Critical |
| 6 | Product gallery thumbnails responsive (H1) | 10 min | High |
| 7 | Visible scroll affordance on horizontal areas (H2) | 45 min | High |
| 8 | Category shelf nav arrows on mobile (H3) | 10 min | High |
| 9 | Product page max-width on desktop (H5) | 10 min | High |
| 10 | Section grid column cap (H6) | 15 min | High |
| 11 | Remove `truncate` from inputs (M2) | 5 min | Medium |
| 12 | Hero min-heights responsive (M1) | 15 min | Medium |
| 13 | Coupon stack on mobile (M5) | 10 min | Medium |
| 14 | Font-size minimums (L1) | 30 min | Low |
| 15 | Prefers-reduced-motion (L2) | 30 min | Low |

---

## Files Audited (representative sample)

Layout & Shell: `layout.tsx`, `mainLayout.tsx`, `userLayout.tsx`, `settingsShell.tsx`, `adminLayout.tsx`, `cartLayout.tsx`, `homeLayout.tsx`, `aboutLayout.tsx`  
Shared UI: `header.tsx`, `headerActions.tsx`, `footer.tsx`, `button.tsx`, `input.tsx`, `Toggle.tsx`, `quantitySelector.tsx`, `text.tsx`, `title.tsx`, `lazyImage.tsx`, `CountrySelect.tsx`, `portalSwitcher.tsx`, `notificationsDropdown.tsx`, `mobileSearchDialog.tsx`, `cartSidebar.tsx`, `mobileNav.tsx`  
Home: `hero.tsx`, `dealsBanner.tsx`, `categoryShelves.tsx`, `categoryShowcaseCards.tsx`, `sectionGrid.tsx`, `sectionCard.tsx`, `productsSection.tsx`, `recently.tsx`, `recentlyGrid.tsx`, `productsGrid.tsx`, `productsCard.tsx`  
Products: `ProductDetails.tsx`, `ProductGallery.tsx`, `ProductActions.tsx`, `ProductSpecs.tsx`, `ProductReviews.tsx`, `content.tsx`, `productsLayout.tsx`, `sideBar.tsx`, `sideBarFilters.tsx`, `sideBarCategories.tsx`, `sideBarBrands.tsx`, `sideBarRatings.tsx`, `sideBarAvailability.tsx`, `sideBarDiscount.tsx`, `filterCheckbox.tsx`, `sideBarTitle.tsx`  
Cart: `cartTitle.tsx`, `ordersGrid.tsx`, `orderActions.tsx`, `paymentBox.tsx`, `paymentMethods.tsx`, `coupon.tsx`, `cartSidebar.tsx`  
Settings/User: `personalInfoForm.tsx`, `preferences.tsx`, `danger.tsx`, `settingsTopTabs.tsx`, `userTopTabs.tsx`, `orderDetailsTopNav.tsx`, `wishlist/content.tsx`  
Admin: `adminLayout.tsx`, dashboard/product/customer/order routes (uses `overflow-x-auto` tables)  
Pages: `about`, `contact`, `faqs`, `press`, `returns`, `careers`, `shipping`, `auth/login`, product `[id]`, products catalog, wishlist, cart, checkout  
Styles: `index.css` (global `100vh`), `theme.css` (design tokens incl. `--radius: 3px`)  
Config: `next.config.mjs` (no image `sizes`/`deviceSizes` responsive config — default Next.js), `package.json` (Tailwind v4 via `@tailwindcss/postcss`)

---

## Missing Configuration to Verify

- **No `tailwind.config.js` found** — Tailwind v4 is being used in CSS-first mode (`@import "tailwindcss"` + `@theme` in `theme.css`). Default breakpoints apply: `sm` 640px / `md` 768px / `lg` 1024px / `xl` 1280px / `2xl` 1536px. Confirm this matches intended device tiers.
- **Next.js image config** — `next.config.mjs` lists `remotePatterns` but no `deviceSizes`/`imageSizes`; responsive `next/image` `sizes` prop not audited. Product images may download full resolution on phones.

---

**Bottom line:** Fix C1–C5 first (viewport meta, dvh, portal/toast overlap, touch targets). Without those, the site cannot be considered to support mobile devices at all. Then address H1–H8 for tablet/professional desktop parity, then M1–M7 and L1–L5 for polish.
