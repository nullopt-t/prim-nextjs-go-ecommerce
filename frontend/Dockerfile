# syntax=docker/dockerfile:1

# ─── Dependencies ───────────────────────────────────────
FROM node:22-alpine AS deps
WORKDIR /app

RUN apk add --no-cache libc6-compat

COPY package.json package-lock.json ./
RUN npm ci

# ─── Build ──────────────────────────────────────────────
FROM node:22-alpine AS builder
WORKDIR /app

ARG NEXT_PUBLIC_API_URL

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# NEXT_PUBLIC_* values are compiled into the client bundle, so this must be
# supplied at build time (docker build --build-arg NEXT_PUBLIC_API_URL=...).
RUN test -n "${NEXT_PUBLIC_API_URL}" \
    || (echo "NEXT_PUBLIC_API_URL build arg is required" >&2; exit 1)
RUN NEXT_TELEMETRY_DISABLED=1 NEXT_PUBLIC_API_URL="${NEXT_PUBLIC_API_URL}" npm run build

# ─── Runtime ────────────────────────────────────────────
FROM node:22-alpine AS runner
WORKDIR /app

RUN apk add --no-cache libc6-compat ca-certificates \
    && addgroup -S -g 1001 nodejs \
    && adduser -S -u 1001 -G nodejs nextjs

COPY --from=builder /app/public ./public
COPY --from=builder /app/messages ./messages
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD node -e "fetch('http://127.0.0.1:'+(process.env.PORT||3000)+'/').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"

# Runtime values (NODE_ENV, PORT, HOSTNAME, NEXT_PUBLIC_API_URL) come from
# the env file or compose environment.
CMD ["node", "server.js"]
