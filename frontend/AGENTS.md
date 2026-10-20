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
