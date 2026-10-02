#!/usr/bin/env bash
# ==============================================================================
# Prim Notification Publisher Script
# ==============================================================================
# Sends simulated notification events directly into Apache Kafka.
# Does NOT touch Go source files or trigger Air hot-reloading.
# ==============================================================================

set -e

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml}"
USER_ID="${1:-67ce034a-47e2-4ada-9b46-dbc6f94cecd0}" # Default: test@gmail.com

generate_uuid() {
    if command -v uuidgen >/dev/null 2>&1; then
        uuidgen | tr '[:upper:]' '[:lower:]'
    else
        cat /proc/sys/kernel/random/uuid
    fi
}

send_notification() {
    local category="$1"
    local title_en="$2"
    local title_ar="$3"
    local msg_en="$4"
    local msg_ar="$5"
    local action_url="$6"
    local event_id
    event_id=$(generate_uuid)
    local now
    now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local inner_data
    inner_data="{\"userId\":\"$USER_ID\",\"category\":\"$category\",\"actionUrl\":\"$action_url\",\"title\":{\"en\":\"$title_en\",\"ar\":\"$title_ar\"},\"message\":{\"en\":\"$msg_en\",\"ar\":\"$msg_ar\"}}"

    local kafka_payload
    kafka_payload="{\"eventId\":\"$event_id\",\"type\":\"in_app\",\"category\":\"$category\",\"userId\":\"$USER_ID\",\"recipient\":\"$USER_ID\",\"locale\":\"ar\",\"timestamp\":\"$now\",\"data\":$inner_data}"

    echo "$kafka_payload" | docker compose -f "$COMPOSE_FILE" exec -T kafka \
        /opt/kafka/bin/kafka-console-producer.sh \
        --bootstrap-server localhost:9092 \
        --topic notifications.inapp >/dev/null 2>&1

    echo "✓ Published [$category] $title_en"
}

echo "=========================================================="
echo "Publishing test notifications for user: $USER_ID"
echo "=========================================================="

send_notification "order" \
    "Your package has shipped! 🚚" \
    "تم شحن شحنتك بنجاح! 🚚" \
    "Courier is out for delivery. Expected arrival within 2 hours." \
    "مندوب التوصيل في طريقه إليك. موعد الوصول المتوقع خلال ساعتين." \
    "/user/orders"

send_notification "system" \
    "Flash Weekend Sale! ⚡ 40% Off" \
    "عروض نهاية الأسبوع الخاطفة! ⚡ خصم 40%" \
    "Don't miss our weekend discounts across all top categories." \
    "لا تفوّت خصومات عطلة نهاية الأسبوع على جميع الأقسام المميزة." \
    "/catalog"

send_notification "auth" \
    "New Device Signed In 🔐" \
    "تسجيل دخول من جهاز جديد 🔐" \
    "A login was detected from Chrome on Linux at $(date -u +'%H:%M UTC')." \
    "تم تسجيل الدخول بنجاح عبر متصفح Chrome على نظام Linux." \
    "/user/security"

echo "=========================================================="
echo "Done! Check your browser bell icon 🔔 (auto-updates every 15s or on open)."
