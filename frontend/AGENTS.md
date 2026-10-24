<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

# Frontend Architecture Rules

## Current phase

- The app runs on the Next.js App Router with a server-rendered shell, but most data fetching is still client-side.
- Client-side rendering is acceptable during active development for interactive and authenticated areas.
- The storefront catalog and SEO-critical pages must move to server rendering before launch.
- Migrate incrementally, page by page. Do not schedule a single big-bang rewrite.

## Rendering matrix

| Area | Target rendering | May stay client-rendered |
| --- | --- | --- |
| Admin dashboard | Client | Yes |
| Auth and account settings | Client shell + server auth | Yes |
| Cart, wishlist, checkout interactions | Client | Yes |
| Header, footer, announcement | Server shell + client islands | Only interactive pieces |
| Home page | Server-rendered with cached data | Interactive widgets |
| Product listing | Server-rendered / ISR | Filters and pagination controls |
| Product detail | Server-rendered / ISR | Add-to-cart and media controls |
| Category and brand pages | Server-rendered / ISR | Interactive filters |
| Search results | Server-rendered | Search input and filter controls |
| Static informational pages | Server-rendered / static | None required |

## Server and client boundaries

- Pages and layouts are Server Components by default. Keep them that way unless a browser API or event handler is required.
- Add `"use client"` to the smallest possible leaf component. Do not mark whole route trees as client components.
- Never import `window`, `document`, `localStorage`, or browser-only libraries into shared modules used by Server Components.
- Pass serializable props from server to client. Do not pass secrets, database handles, or server-only functions.
- Keep shared API contracts in `src/services`. Put server-only fetchers in `src/lib/server`.
- Client Components may hydrate from server-fetched data; they should not refetch the same data immediately on mount when initial data is available.

## Data fetching

- Prefer fetching data in Server Components close to where it is rendered.
- Use `process.env.API_URL` for server-to-API calls and `process.env.NEXT_PUBLIC_API_URL` only for browser calls.
- Never expose API keys, webhook secrets, or internal service URLs to the client bundle.
- Keep search, filters, sorting, and pagination in the URL so server rendering and navigation stay deterministic.
- Do not put SEO-critical content behind `useEffect` or a client-only loading state.
- When a client surface needs server data, pass initial data from the server and keep client refetching for freshness only.
- Use `<Suspense>` or route-level `loading.tsx` for slow data instead of blocking the whole page.

## Caching and revalidation

- Do not default every request to `no-store`.
- Use `no-store` only for session-specific data such as user state and for client-side freshness checks.
- Cache catalog reads (products, categories, brands) and revalidate them with tags or path revalidation after admin mutations.
- When adding admin mutations, invalidate the corresponding storefront cache entry or tag.
- Public settings are server-fetched with the `public-settings` tag. Invalidate that tag after an admin save; the client-side refresh can still use `no-store`.

## Internationalization

- Supported locales are only `en` and `ar`.
- Keep locale in the URL and resolve it on the server whenever possible.
- Do not add new client-only redirects for locale selection. Existing client fallback should be replaced by server-side locale resolution during migration.
- Server Components must receive locale from route params or request configuration, not from `localStorage`.
- Client-only preferences such as an explicitly chosen locale may stay in a cookie or local storage, but must not be required for the first server render.

## Authentication and authorization

- Authentication cookies are HTTP-only. Server Components may read them for protected server-rendered pages.
- Never trust role or identity data from client storage.
- Keep admin mutation checks on the backend. Client-side hiding is UX only, not security.
- Preserve `credentials: "include"` for browser API calls that depend on session cookies.

## SEO and metadata

- SEO-critical pages must expose their primary content in the initial HTML response.
- Move title, description, canonical URL, Open Graph, and product/category metadata to server-side `generateMetadata`.
- Generate sitemaps and structured data from server data, not from client fetches.
- Avoid hardcoded brand names in metadata and shared UI when the value is available from public settings.

## Migration plan

1. Fetch public settings on the server in the locale layout and pass them as initial data. **Applied.**
2. Move the home page's initial catalog data to the server.
3. Move product listing and product detail pages to server rendering with metadata and cache tags.
4. Move category, brand, and search pages to server rendering.
5. Replace the client-side default-language redirect with server-side locale resolution.
6. Add cache tags and on-demand revalidation for catalog mutations.
7. Audit client bundles and split remaining interactive islands.
8. Verify hydration, SEO output, and performance before launch.

## Definition of done for a migrated route

- Primary content appears in `view-source`, not only after hydration.
- Metadata is generated on the server from the same data as the page.
- No immediate client refetch of data already provided by the server.
- No hydration mismatch warnings.
- Loading and error states are meaningful.
- Auth and authorization remain enforced by the backend.
- `npx tsc --noEmit` passes.
- The page still works with JavaScript execution delayed.

## Verification commands

```bash
cd frontend
npx tsc --noEmit
npm run build
```

For SSR checks, inspect the raw HTML response (for example with `curl`) and confirm that primary content and metadata are present before client JavaScript runs.

# Working Agreement

Scope: everything under `frontend/`. This file governs the Next.js app only. Do
not change the Go backend, migrations, or API contracts from a frontend task;
call out the contract you need and keep the UI compatible with the current API.
The backend has its own `backend/AGENTS.md`.

## Start here

1. Load `using-agent-skills` before a non-trivial task.
2. For Next.js APIs, routing, caching, or metadata work, load
   `source-driven-development` and read the version-matched docs in
   `node_modules/next/dist/docs/` — this Next.js version differs from older
   training data.
3. Read the existing feature module and service before adding a parallel one.
   Reuse the established patterns instead of creating a second data layer.

## Commands

Run from `frontend/`:

```bash
npm run dev        # http://localhost:3000
npm run build      # production build; must pass before handoff
npm run lint       # ESLint
npx tsc --noEmit   # type check
```

The first two commands plus the `nextjs-agent-rules` block at the top of this
file are part of the normal workflow; do not delete that block.

## Code conventions

- Use tabs with width 4 per `.prettierrc`. Match the surrounding file's quote and
  semicolon style.
- Import through the `@/*` alias; do not introduce long `../../..` chains.
- Put typed API calls in `src/services/*` using `src/api/client.ts`. Components
  should call services/hooks, not raw `fetch`, except in server components or
  route handlers where server-side fetching is intentional.
- Keep browser calls credentialed (`credentials: "include"`); the session is an
  HTTP-only cookie. Never trust role/identity from local storage.
- Use `useFetch` / `useMutation` for client data and put URL state (filters,
  search, pagination) where it can be shared and restored.
- Keep feature code in `src/features/<feature>/` and route files in
  `src/app/[locale]/...` thin. Shared primitives belong in `src/components/ui`.
- Add every user-facing string to both `messages/en.json` and
  `messages/ar.json`; verify RTL layout, not just the English copy.
- Preserve the vendor portal's mock-data boundary until a backend exists.
- Do not hardcode brand, shipping, tax, payment, or contact values that already
  come from public settings.
- Avoid `any` in new public types. If the backend shape is genuinely unknown,
  define the narrow type you use and note the gap.

## Performance and accessibility

- Keep client bundles small; add `"use client"` only at interactive leaves.
- Use `next/image`, semantic HTML, labeled controls, and visible focus states.
- Do not ship a page whose primary content only appears after hydration if it is
  in the rendering matrix as server-rendered.
- Check the browser console and network panel for errors before calling UI work
  done; use `browser-testing-with-devtools` when a real browser is needed.

## Skill routing

Global skills live in `~/.agents/skills/` and are loaded with the `skill` tool.
Load the most specific one before starting.

| Task | Skill |
| --- | --- |
| Session setup / deciding which workflow applies | `using-agent-skills` |
| Next.js/React/library APIs and version-specific behavior | `source-driven-development` |
| Building or restyling UI, responsive layout, accessibility | `frontend-ui-engineering` |
| Real-browser verification, DOM/console/network checks | `browser-testing-with-devtools` |
| New behavior or bug fix that needs proof | `test-driven-development` |
| Broken build, hydration error, failing request | `debugging-and-error-recovery` |
| Client bundle size, Core Web Vitals, slow pages | `performance-optimization` |
| Auth flows, untrusted input, XSS, dependency risk | `security-and-hardening` |
| Reviewing a diff before merge | `code-review-and-quality` |
| Large UI change that must land in slices | `incremental-implementation` |
| High-stakes or irreversible change | `doubt-driven-development` |
| ADRs, component/API docs | `documentation-and-adrs` |
| Commits, branches, PRs, releases | `git-workflow-and-versioning` |
| Enforcing lint/type/performance thresholds | `constraint-driven-development` |

## Definition of done (frontend)

- `npx tsc --noEmit` and `npm run build` pass.
- `npm run lint` passes, or the remaining warnings are pre-existing and noted.
- No hydration warnings or console errors in the touched flow.
- Primary content and metadata are in the initial HTML for server-rendered
  routes.
- Both `en` and `ar` strings exist; RTL was checked.
- API calls still send the session cookie and handle `401` through the existing
  client behavior.
- No secrets or internal service URLs are exposed to the client bundle.

