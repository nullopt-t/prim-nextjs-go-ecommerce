#!/usr/bin/env bash
set -euo pipefail
shopt -s nullglob

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT_DIR/backend/.env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "FAIL: backend/.env is missing" >&2
  exit 1
fi

# Read a key without sourcing the file (values may contain special characters).
env_value() {
  sed -n "s/^$1=//p" "$ENV_FILE" | head -n 1
}

fail=0
for key in DATABASE_URL REDIS_URL NEXT_PUBLIC_API_URL; do
  if [[ -z "$(env_value "$key")" ]]; then
    echo "FAIL: $key is not set in backend/.env" >&2
    fail=1
  else
    echo "OK:   $key is set"
  fi
done

migrations=( "$ROOT_DIR"/backend/migrations/*.up.sql )
if (( ${#migrations[@]} == 0 )); then
  echo "FAIL: no migrations found in backend/migrations" >&2
  fail=1
fi
for up in "${migrations[@]}"; do
  down="${up%.up.sql}.down.sql"
  if [[ ! -f "$down" ]]; then
    echo "FAIL: missing rollback: $down" >&2
    fail=1
  fi
done
if (( fail == 0 )); then
  echo "OK:   ${#migrations[@]} migrations have rollback files"
fi

mode="$(env_value APP_ENV)"; mode="${mode:-development}"
db_url="$(env_value DATABASE_URL)"
s3_secret="$(env_value S3_SECRET_KEY)"
[[ -n "$s3_secret" ]] || s3_secret="$(env_value MINIO_SECRET_KEY)"
pepper="$(env_value OTP_HASH_PEPPER)"
stripe="$(env_value STRIPE_SECRET_KEY)"

[[ "$db_url" != *"prim:prim@"* ]] || echo "WARN: DATABASE_URL still uses the default prim:prim credentials"
case "$s3_secret" in
  ""|test|supersecret|admin) echo "WARN: S3 secret key is a dev default or empty" ;;
esac
case "$pepper" in
  ""|dev-*|__REPLACE_*) echo "WARN: OTP_HASH_PEPPER is a dev placeholder or empty" ;;
esac
case "$stripe" in
  ""|sk_test_*) echo "WARN: STRIPE_SECRET_KEY is a test key or empty" ;;
esac
[[ "$mode" == "production" ]] || echo "WARN: APP_ENV='$mode', expected production"

if (( fail != 0 )); then
  echo "prod-check failed: fix the errors above" >&2
  exit 1
fi

echo "Running backend static analysis and tests..."
cd "$ROOT_DIR/backend"
go vet ./...
go test ./...
echo "prod-check passed"
