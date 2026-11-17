# PRIM — Production Roadmap & Task Backlog

A prioritized tracking document for post-simulation production stages, infrastructure enhancements, and feature milestones.

---

## 🚀 High-Priority Production Tasks

### 1. Telegram Gateway Integration for Phone OTP Verification
- [ ] **Objective:** Enable official Telegram Gateway as the primary, ultra-cost-effective (~$0.01 / verification) phone authentication delivery channel.
- [ ] **Components to Implement:**
  - Create `TelegramGatewaySender` in [`backend/pkg/sms/sms.go`](file:///home/amr/tmp/prim/backend/pkg/sms/sms.go) implementing the existing `Sender` interface:
    ```go
    type Sender interface {
        Send(ctx context.Context, to string, message string) error
    }
    ```
  - Integrate with Telegram Gateway API: `POST https://gatewayapi.telegram.org/sendVerificationMessage`.
  - Add configuration variables to [`backend/pkg/config/config.go`](file:///home/amr/tmp/prim/backend/pkg/config/config.go):
    - `SMS_PROVIDER=telegram_gateway`
    - `TELEGRAM_GATEWAY_TOKEN=...`
  - Implement delivery status check and fallback strategy (e.g. if the recipient does not have a Telegram account, fall back to standard SMS).
- [ ] **Cost Impact:** Drops customer phone verification costs by ~80% compared to traditional SMS ($0.01 vs $0.05–$0.08 in Egypt & MENA).

---

### 2. Live Cloud Environment Transition
- [ ] **Managed Cloud Primitives:**
  - Create AWS RDS (or Neon) PostgreSQL database (`sslmode=require`).
  - Create AWS S3 bucket (or Cloudflare R2) and configure public media CDN domain (`media.prim.com`).
  - Configure managed Redis (AWS ElastiCache, Upstash, or Redis Cloud) with TLS (`rediss://`).
  - Configure managed Kafka (AWS MSK or Confluent Cloud) or Redpanda.
- [ ] **Transactional Email Deployment:**
  - Verify sending domain with DKIM/SPF/DMARC DNS records.
  - Switch `backend/.env` from local Mailpit to **AWS SES** (`email-smtp.us-east-1.amazonaws.com:587`) or **Brevo** (300 free emails/day).
- [ ] **Live Stripe Gateway Activation:**
  - Set production `STRIPE_SECRET_KEY` and `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY`.
  - Register production webhook listener in Stripe Dashboard (`https://api.prim.com/api/v1/checkout/webhook`) and set `STRIPE_WEBHOOK_SECRET`.

---

### 3. Observability, Security & Production Operations
- [ ] **Reverse Proxy & SSL Termination:**
  - Configure Caddy or Cloudflare in front of Next.js (`:3000`) and the Go API (`:8080`) with automatic TLS.
  - Enforce HSTS, CSP, and strict CORS whitelist (`CORS_ALLOWED_ORIGINS`).
- [ ] **Error Monitoring:**
  - Connect Sentry for Go backend (`sentry-go`) and Next.js frontend (`@sentry/nextjs`).
- [ ] **Uptime & Health Checks:**
  - Set up external health monitoring (Better Uptime or Pingdom) on `/health` and `/api/v1/health`.
