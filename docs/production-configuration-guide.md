# PRIM — Production Configuration & Cloud Transition Guide

This guide details how to transition PRIM from the **local Docker simulation** environment to **live cloud production**.

The architecture is **100% configuration-driven**: transitioning to production does **not** require modifying any Go backend or Next.js frontend code. You only swap and replace text values in `backend/.env`.

---

## 1. Two Deployment Modes

PRIM provides two separate deployment targets depending on your hosting model:

### Mode A: Self-Hosted / Full Docker Simulation (`make prod-up`)
- **Use case:** Local staging simulation or running the entire stack (including PostgreSQL, Redis, Kafka, and Floci S3 emulator) inside Docker on a single VPS (e.g. Hetzner, DigitalOcean, EC2).
- **Compose file:** [`backend/docker/docker-compose.prod.yml`](file:///home/amr/tmp/prim/backend/docker/docker-compose.prod.yml)
- **Command:**
  ```bash
  make prod-up
  ```

### Mode B: Cloud-Managed Infrastructure (`make cloud-up`)
- **Use case:** Production deployment connecting directly to managed cloud primitives (AWS RDS / Neon for PostgreSQL, AWS ElastiCache / Upstash for Redis, AWS MSK / Confluent for Kafka, AWS S3 / Cloudflare R2 for object storage).
- **Compose file:** [`backend/docker/docker-compose.cloud.yml`](file:///home/amr/tmp/prim/backend/docker/docker-compose.cloud.yml)
- **Command:**
  ```bash
  make cloud-up
  ```
  *(Boots only `migrate`, `api`, `worker`, and `frontend`, with zero local database/broker containers).*

---

## 2. Text Swap Dictionary (`backend/.env`)

To switch to production, copy the production template:
```bash
cp backend/.env.production.example backend/.env
```
Then swap the following key sections:

### 1. Domains & CORS
| Environment Variable | Simulation / Dev | Production Example | Description |
| :--- | :--- | :--- | :--- |
| `APP_ENV` | `development` | `production` | Enforces production security checks |
| `CLIENT_BASE_URL` | `http://localhost:3000` | `https://prim.com` | Base frontend URL used in email links |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | `https://prim.com,https://admin.prim.com` | Strict CORS whitelist |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8081` | `https://api.prim.com` | Client-facing browser API endpoint |
| `NEXT_PUBLIC_SITE_URL` | `http://localhost:3000` | `https://prim.com` | SEO & canonical URL generation |

---

### 2. PostgreSQL Database
| Environment Variable | Simulation / Dev | Production Example (AWS RDS / Neon) |
| :--- | :--- | :--- |
| `DATABASE_URL` | `postgres://prim:prim@postgres:5432/prim?sslmode=disable` | `postgresql://prim_user:SecretPass@prim-db.rds.amazonaws.com:5432/prim?sslmode=require` |

*Note: Migrations run automatically on startup via the `migrate` container (`make cloud-migrate` or `make prod-migrate`).*

---

### 3. Redis Cache & Sessions
| Environment Variable | Simulation / Dev | Production Example (ElastiCache / Upstash) |
| :--- | :--- | :--- |
| `REDIS_URL` | `redis://redis:6379` | `rediss://default:StrongAuthToken@prim-cache.redis.amazonaws.com:6379/0` |

*Note: For TLS-enabled Redis, always use the `rediss://` protocol prefix.*

---

### 4. Object Storage & Media CDN
You can use **AWS S3** or **Cloudflare R2**. Next.js image optimization dynamically allows whatever domain you set in `NEXT_PUBLIC_MEDIA_BASE_URL`.

#### Option A: AWS S3 + CloudFront CDN
```ini
S3_ENDPOINT=s3.us-east-1.amazonaws.com
S3_REGION=us-east-1
S3_BUCKET=prim-media-production
S3_ACCESS_KEY=AKIA_YOUR_IAM_ACCESS_KEY
S3_SECRET_KEY=YOUR_IAM_SECRET_KEY
S3_USE_SSL=true
S3_PUBLIC_URL=https://media.prim.com
NEXT_PUBLIC_MEDIA_BASE_URL=https://media.prim.com
```

#### Option B: Cloudflare R2
```ini
S3_ENDPOINT=<ACCOUNT_ID>.r2.cloudflarestorage.com
S3_REGION=auto
S3_BUCKET=prim-media-production
S3_ACCESS_KEY=<R2_ACCESS_KEY_ID>
S3_SECRET_KEY=<R2_SECRET_ACCESS_KEY>
S3_USE_SSL=true
S3_PUBLIC_URL=https://media.prim.com
NEXT_PUBLIC_MEDIA_BASE_URL=https://media.prim.com
```

---

### 5. Event Broker (Kafka)
| Environment Variable | Simulation / Dev | Production Example (AWS MSK / Confluent) |
| :--- | :--- | :--- |
| `KAFKA_BROKERS` | `kafka:9092` | `b-1.msk.amazonaws.com:9092,b-2.msk.amazonaws.com:9092` |
| `KAFKA_GROUP_ID` | `prim-notification-workers` | `prim-notification-workers` |

---

### 6. Stripe Payments
| Environment Variable | Simulation / Dev | Production Example |
| :--- | :--- | :--- |
| `STRIPE_SECRET_KEY` | `sk_test_...` | `sk_live_...` |
| `STRIPE_PUBLISHABLE_KEY` | `pk_test_...` | `pk_live_...` |
| `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY` | `pk_test_...` | `pk_live_...` |
| `STRIPE_WEBHOOK_SECRET` | `whsec_...` | `whsec_live_...` |

*Stripe webhook endpoint path: `https://api.prim.com/api/v1/checkout/webhook`*

---

### 7. Transactional Emails & SMS
| Environment Variable | Simulation / Dev | Production Example (SendGrid / Resend) |
| :--- | :--- | :--- |
| `SMTP_HOST` | `mailpit` or `smtp.gmail.com` | `smtp.sendgrid.net` or `smtp.resend.com` |
| `SMTP_PORT` | `1025` | `587` |
| `SMTP_USERNAME` | *(empty)* | `apikey` or `resend` |
| `SMTP_PASSWORD` | *(empty)* | `YOUR_API_KEY` |
| `SMTP_FROM_EMAIL` | `no-reply@localhost` | `no-reply@prim.com` |
| `SMS_PROVIDER` | `log` | `twilio` |
| `TWILIO_ACCOUNT_SID` | *(empty)* | `AC_YOUR_ACCOUNT_SID` |
| `TWILIO_AUTH_TOKEN` | *(empty)* | `YOUR_AUTH_TOKEN` |
| `TWILIO_FROM_PHONE` | *(empty)* | `+1234567890` |

---

### 8. Security & Cryptographic Keys
Generate a unique, cryptographically secure 48-byte pepper:
```bash
openssl rand -base64 48
```
Set in `backend/.env`:
```ini
OTP_HASH_PEPPER=<GENERATED_BASE64_STRING>
```

---

## 3. Automation for Opencode / CI/CD

When automating deployment with Opencode or a CI/CD pipeline (e.g. GitHub Actions), you can generate `backend/.env` dynamically from vault secrets:

```bash
# Example CI/CD deployment script:
envsubst < backend/.env.production.example > backend/.env

# Build and deploy managed cloud stack
make cloud-up
```

All configurations will be loaded automatically by both the Go backend (`pkg/config`) and Next.js frontend (`next.config.mjs`).
