# PRIM Backend — Capacity Plan for 60,000 Users

**Generated:** 2026-10-23T04:19:52Z (UTC)
**Basis:** live k6 runs on the dev stack (`stress-k6-2026-10-23_040600Z.md`, `stress-k6-curve-2026-10-23_041555Z.md`)
**Measured baseline:** ~1,250 req/s ceiling with `GOMAXPROCS=4`, `MaxConns=25`, one shared customer cart; API avg 1.23 / peak 2.77 cores, Postgres avg 0.58 / peak 1.80 cores, Redis negligible.

> **The most important number is not "60,000 users".** Sizing is driven by **peak concurrent sessions** and **requests per second**. 60,000 registered accounts, 60,000 daily actives, and 60,000 simultaneous users are three completely different systems.

---

## 1. Three interpretations of "60,000 users"

| Interpretation | Assumed peak concurrency | Requests/user/s | RPS range | Design peak (2×) |
|---|---|---|---|---|
| 60k registered, ~1% online at peak | 600 | 0.1–0.2 | 60–120 | ~240 |
| 60k DAU, ~10% online at peak | 6,000 | 0.1–0.2 | 600–1,200 | ~2,400 |
| 60k **concurrent** users | 60,000 | 0.1–0.2 | 6,000–12,000 | ~24,000 |

Request rates assume a browse-heavy SPA/mobile client: one request every 5–10 seconds per active user. Chatty clients, live search, polling, or websockets can multiply this.

---

## 2. Measured cost per 1,000 req/s

These are **lower bounds**, because the benchmark ceiling was set by the DB pool and a deliberately shared cart, not raw CPU:

| Resource | Average per 1,000 rps | Peak per 1,000 rps |
|---|---|---|
| API CPU | ~1.0 core | ~2.2 cores |
| Postgres CPU | ~0.46 core | ~1.44 cores |
| Redis CPU | ~0.025 core | ~0.10 core |

A conservative planning figure is **~1,200 rps per 4-vCPU API node** at ~65% CPU. After pool tuning and per-user carts, 2,000–3,000 rps per node is plausible, but must be re-benchmarked before relying on it.

---

## 3. Recommended specs

### Option A — 60k registered users (≈600 concurrent, ~240 rps design)

| Component | Spec |
|---|---|
| API | 2 × 4 vCPU / 8 GB (active + HA standby), `GOMAXPROCS=4` |
| PostgreSQL | 1 × 8 vCPU / 32 GB, 500 GB NVMe |
| Read replica | optional 1 × 8 vCPU / 32 GB |
| PgBouncer | 2 × 2 vCPU / 4 GB |
| Redis | 1 × 4 GB + replica (sessions, cart, rate limits) |
| Load balancer | 2 × 2 vCPU / 4 GB (or managed LB) |
| Object storage | 500 GB–1 TB (MinIO/S3) |

This is close to the current workstation, but with HA, backups, and a managed database instead of a dev container.

### Option B — 60k DAU (≈6,000 concurrent, ~2,400 rps design)

| Component | Spec |
|---|---|
| API | 3 × 4 vCPU / 8 GB (or 2 × 8 vCPU / 16 GB), autoscale 3–5 |
| PostgreSQL primary | 16 vCPU / 64 GB, 1 TB NVMe, 10k+ IOPS |
| Read replicas | 2 × 8 vCPU / 32 GB |
| PgBouncer | 2 × 4 vCPU / 8 GB |
| Redis | 8 GB primary + replica (16 GB if caching catalog) |
| Kafka | 3 × 4 vCPU / 8 GB (only if events/notifications are async) |
| Object storage | 2 TB |
| Network | 5–10 Gbps |

This is the realistic target for most "60k users" products.

### Option C — 60k truly concurrent users (~24,000 rps design)

| Component | Spec |
|---|---|
| API | 10 × 8 vCPU / 16 GB (or 20 × 4 vCPU / 8 GB), autoscale 8–14 |
| PostgreSQL primary | 32–64 vCPU / 128–256 GB, 2 TB NVMe, 20k+ IOPS |
| Read replicas | 3–4 × 16–32 vCPU / 64–128 GB |
| PgBouncer | 2 × 4 vCPU / 8 GB (200–400 server connections) |
| Redis | 16–32 GB cluster (sessions + hot catalog + rate limits) |
| Kafka | 3–5 brokers × 8 vCPU / 16 GB |
| Object storage | 4–5 TB |
| Network | 10 Gbps |
| Observability | Prometheus/Grafana/Loki, 8 vCPU / 16 GB, 1 TB |

At this level a **single Postgres primary is the risk**, not the API tier. Plan for read replicas, Redis caching of hot catalog reads, partitioning orders/events by time, and async processing through Kafka. If write throughput exceeds ~8–10k transactions/s, evaluate Citus/sharding or a distributed SQL store.

---

## 4. Architecture notes

1. **Scale the API horizontally.** It is stateless; sessions live in Redis. Put 3+ nodes across availability zones behind a load balancer.
2. **Use PgBouncer.** Per-node pools of 20–50 connections multiplied by 10–20 nodes will exhaust Postgres without multiplexing. Target 200–400 server connections, not 1,000.
3. **Cache hot reads.** Catalog, product detail, and category/brand lists dominate the measured slow paths. A 70–80% Redis hit rate removes most read traffic from Postgres.
4. **Remove the benchmark artefact.** The stress test shared one cart across all VUs, which serialized cart/checkout writes. Production carts are per-user, so real write contention will be much lower than measured.
5. **Async side effects.** Email, notifications, thumbnails, and analytics should go through Kafka/workers, not the request path.
6. **Protect the edge.** CDN for media, WAF, per-IP and per-user rate limits, request size limits.
7. **Backups and DR.** PITR, daily snapshots, cross-AZ replicas; test restore time.

---

## 5. Tuning before buying hardware

Run these on a staging box (not the workstation) before committing to Option B/C:

1. `MaxConns` 25 → 50–100 per API instance, with PgBouncer in front.
2. `GOMAXPROCS` = vCPU count (the dev stack used 4 on a 12-core host).
3. Per-VU users/carts so the benchmark stops serializing on one cart row.
4. Enable `pg_stat_statements`; capture pool wait time, lock waits, and Go GC/pprof.
5. Re-run k6 at 100/200/400 VUs and measure **rps per node** and **p99**.
6. Soak at 70% capacity for 1–2 hours to catch leaks and lock buildup.

If per-node throughput doubles after tuning, Option C's API tier drops from ~10 nodes to ~5–6; the database remains the limiting factor.

---

## 6. Bottom line

- **If "60,000 users" means registered/MAU:** 2 × 4 vCPU API nodes, 8–16 vCPU Postgres with 32–64 GB RAM, 4–8 GB Redis. The current hardware is already in the right ballpark; production needs HA and backups.
- **If it means 60,000 DAU:** 3 × 4 vCPU API nodes, 16 vCPU Postgres + 2 read replicas, 8 GB Redis — roughly 5× the current measured ceiling at peak.
- **If it means 60,000 concurrent:** budget for ~10 × 8 vCPU API nodes, a 32–64 vCPU Postgres primary with 3–4 replicas, 16–32 GB Redis, and Kafka. At ~24k rps, the database and write path — not the API — determine the design.
