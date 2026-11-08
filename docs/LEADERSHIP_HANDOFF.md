# PRIM LEADERSHIP HANDOFF & SUPREME COMMANDER BRIEFING
Date: 2026-11-08
Target Leader: OpenCode (`w4:p2`)
Handing over: Antigravity-Lead (`w4:p1`)

================================================================================
1. SUPREME LEADER ROLE & HERDR INFRASTRUCTURE
================================================================================
OpenCode (`w4:p2`), you are now promoted to SUPREME LEADER of the entire PRIM engineering fleet.
All agents across all workspaces report directly to you.

Herdr CLI commands available in your terminal (`PATH`):
- List all agents: `herdr agent list`
- Prompt an agent: `herdr agent prompt <agent-name> "<instruction>"`
- Read any pane's screen: `herdr pane read <pane_id> --lines 40`
- Send keystrokes: `herdr agent send-keys <agent-name> <key>` (or `herdr pane send-keys <pane_id> <key>`)
- Run commands in an open shell pane: `herdr pane run <pane_id> "<command>"`
- Focus a pane: `herdr pane focus <pane_id>`
- Create workspace: `herdr workspace create --label "<name>"`
- Create tab: `herdr tab create --workspace <ws> --label "<name>"`
- Split pane: `herdr pane split <pane_id> --direction right|down --no-focus`

================================================================================
2. COMPLETE FLEET TOPOLOGY & ACTIVE SOLDIERS
================================================================================

WORKSPACE w4: Core Fleet & Backend Operations
- `w4:p1`: Antigravity-Lead (Standing down / passing baton to you)
- `w4:p2`: OpenCode (YOU - SUPREME LEADER)
- `w4:p6`: `agy-2` (Agy) - Order, Cart, Checkout, RMA backend specialist
- `w4:p3`: Freebuff - Backend product/seller/catalog helper
- `w4:pE`: `catalog-auditor` (Agy, Tab 6 `Audit-Catalog`) - Catalog, Inventory, Coupons, Variants
- `w4:p4`: Host Shell running `make prod-logs` (docker compose log stream)
- `w4:pC`: Host Shell for running commands directly outside sandbox

WORKSPACE w5: Frontend Bug-Hunt & UX Fleet
- `w5:p1`: `fe-subleader` (Agy) - Frontend Subleader, orchestrating the 3 frontend soldiers below
- `w5:p2`: `fe-soldier-1` (Agy) - Catalog, Search, PDP, Compare, Currency Switcher
- `w5:p3`: `fe-soldier-2` (Agy) - Cart, Checkout, Order Lookup, Printable Invoice
- `w5:p4`: `fe-soldier-3` (Agy) - Auth, Passkey WebAuthn, Addresses, Seller Wizard
- `w5:p5`: Freebuff (Tab 2 `Mobile-Responsive-Freebuff`) - Mobile-first responsive specialist

WORKSPACE w6: OpenCode Full-Stack Specialists
- `w6:p1`: `oc-session-1` (OpenCode) - Cart & Checkout integration audit
- `w6:p2`: `oc-session-2` (OpenCode) - Order lifecycle & RMA returns cumulative checks
- `w6:p3`: `oc-session-3` (OpenCode) - Catalog facets & currency precision

================================================================================
3. SYSTEM ARCHITECTURE & REPOSITORY OVERVIEW
================================================================================
Repository: `/home/amr/tmp/prim`
- Frontend: Next.js 16 (App Router), TypeScript, Tailwind CSS, shadcn/ui.
  Bilingual: English & Arabic via next-intl. Strings MUST exist in both `frontend/messages/en.json` and `frontend/messages/ar.json`.
  Server fetch base URL: `getServerApiBaseUrl()` reads runtime `API_URL` (inside container: `http://api:8080`) or fallback `NEXT_PUBLIC_API_URL` (`http://localhost:8081`).
  Timeout guard: All server-side catalog and settings fetches use `signal: AbortSignal.timeout(3000)`.
- Backend: Go 1.24 (Gin + PostgreSQL + Redis + Kafka).
  Port inside compose: `8080`. Mapped to host at `http://localhost:8081`.
- Production Stack: Running in Docker (`docker compose --env-file backend/.env -f backend/docker/docker-compose.prod.yml up -d`).
  Frontend: `http://localhost:3000`
  API: `http://localhost:8081` (`/health`, `/api/v1/settings/currencies`)
  Containers: `docker-frontend-1`, `docker-api-1`, `docker-worker-1`, `docker-postgres-1`, `docker-redis-1`, `docker-kafka-1`, `docker-minio-1`.

================================================================================
4. COMPLETED SLICES & GIT STATUS
================================================================================
All 13 Merchant standard feature gaps are implemented, verified, and committed:
1. Faceted Attribute Filtering (`product_handler.go`, `sideBarAttributes.tsx`)
2. Guest Order Lookup (`/order-lookup`, `LookupOrderByGuest`)
3. Returns & Refunds RMA (`POST /api/v1/orders/:id/returns`, migration 000004 for cumulative limits)
4. Notify-Me-When-Available (`NotifyMeDialog.tsx`, Redis cooldown)
5. Product Compare (`/compare`, `compareStore.ts`, mobile sticky table)
6. Synced Browsing History (`/api/v1/user/history/products`)
7. Related / Cross-Sell (`relatedProducts.ts`, PDP carousel)
8. Volume / Tier Pricing (`TierPricingTable.tsx`, `tierPricing.ts`)
9. Bundle & Downloadable types (seller wizard bundle kit builder, downloadable schema)
10. Printable Invoice PDF (`InvoicePrintView.tsx`, `@media print`)
11. Passkey WebAuthn (`LoginView.tsx`, `navigator.credentials.get`)
12. Multi-Currency Switcher (`GET /api/v1/settings/currencies`, `CurrencySwitcher.tsx`)
13. Gift Wrap & Cards (CartContext, fee handling, coupon hints)

Current Git State:
- Branch `main` is clean. 21 atomic, conventional commits ahead of origin.
- Authorship strictly `Amr Ali`. Zero bot watermarks or co-author signatures.

================================================================================
5. SACRED OPERATING CONTRACT (FLEET PROTOCOL)
================================================================================
Read `docs/FLEET_PROTOCOL.md`:
1. NO SILENT WORK: When agents finish a slice, they MUST report to Leader (`w4:p2` - you) with findings, fixes, and verification status.
2. VERIFICATION GATES:
   - Frontend: `cd frontend && npx tsc --noEmit` MUST be 0 errors; `npm test` (vitest) MUST be 100% green (currently 99/99 passing).
   - Backend: `cd backend && go test ./...` and `go vet ./...` MUST pass.
3. PONYTAIL DISCIPLINE: Minimal surgical diffs. Do not rewrite working modules. Solve the concrete logical bug, test it, and commit.
4. GIT ATTRIBUTION: Always preserve commit authorship as `Amr Ali` (`user.name=Amr Ali`, `user.email=amrrmohamed34342@gmail.com`). Never add AI branding or Co-Authored-By.

================================================================================
6. IMMEDIATE ACTION ITEMS FOR OPENCODE (NEW LEADER)
================================================================================
1. Acknowledge supreme leadership to the user.
2. Direct all soldiers to send their completion reports to `w4:p2` (you).
3. Monitor `w5:p1` (`fe-subleader`) for the consolidated frontend bug hunt results.
4. Keep the production stack healthy and maintain clean atomic git history.
EOF