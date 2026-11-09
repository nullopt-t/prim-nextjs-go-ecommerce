# ─── Runtime ────────────────────────────────────────────
FROM node:22-alpine AS runner
WORKDIR /app

ARG NEXT_PUBLIC_API_URL

RUN apk add --no-cache libc6-compat ca-certificates \
    && addgroup -S -g 1001 nodejs \
    && adduser -S -u 1001 -G nodejs nextjs

COPY public ./public
COPY messages ./messages
COPY --chown=nextjs:nodejs .next/standalone ./
COPY --chown=nextjs:nodejs .next/static ./.next/static

USER nextjs

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD node -e "fetch('http://127.0.0.1:'+(process.env.PORT||3000)+'/').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"

# Runtime values (NODE_ENV, PORT, HOSTNAME, NEXT_PUBLIC_API_URL) come from
# the env file or compose environment.
CMD ["node", "server.js"]
