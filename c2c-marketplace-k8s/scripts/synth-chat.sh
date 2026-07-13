#!/usr/bin/env bash
# Drive two WebSocket peers against messaging-service and assert delivery.
# Requires: websocat (brew install websocat)
set -euo pipefail

URL="ws://localhost:8083"
BUYER="synth-buyer-0"
SELLER="synth-seller-0"
MESSAGES=3
CONNECT_SETTLE_SECS=1
DELIVERY_WAIT_SECS=2
BUYER_TIMEOUT_SECS=5
SELLER_TIMEOUT_SECS=15

usage() {
    cat <<'EOF'
Usage: synth-chat.sh [--url ws://host:port] [--buyer ID] [--seller ID] [--messages N]

Connects a seller listener and a buyer sender to messaging-service WebSockets,
sends N JSON chat messages, and exits 0 iff the seller observes at least one
"Synth hello" payload.

Defaults:
  --url       ws://localhost:8083
  --buyer     synth-buyer-0
  --seller    synth-seller-0
  --messages  3
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --url)
            URL="${2:?--url requires a value}"
            shift 2
            ;;
        --buyer)
            BUYER="${2:?--buyer requires a value}"
            shift 2
            ;;
        --seller)
            SELLER="${2:?--seller requires a value}"
            shift 2
            ;;
        --messages)
            MESSAGES="${2:?--messages requires a value}"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown argument: $1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

if ! [[ "$MESSAGES" =~ ^[1-9][0-9]*$ ]]; then
    echo "ERROR: --messages must be a positive integer (got '$MESSAGES')" >&2
    exit 1
fi

if ! command -v websocat >/dev/null 2>&1; then
    echo "ERROR: websocat is required but was not found on PATH." >&2
    echo "Install it with: brew install websocat" >&2
    exit 1
fi

# Strip trailing slash so we can append /ws/{userId}.
URL="${URL%/}"
CONVERSATION_ID="${BUYER}:${SELLER}"
SELLER_WS="${URL}/ws/${SELLER}"
BUYER_WS="${URL}/ws/${BUYER}"

SELLER_LOG="$(mktemp -t synth-chat-seller.XXXXXX)"
SELLER_ERR="$(mktemp -t synth-chat-seller-err.XXXXXX)"
SELLER_PID=""
SELLER_WATCHDOG_PID=""

cleanup() {
    if [[ -n "${SELLER_WATCHDOG_PID}" ]]; then
        kill "${SELLER_WATCHDOG_PID}" 2>/dev/null || true
        wait "${SELLER_WATCHDOG_PID}" 2>/dev/null || true
    fi
    if [[ -n "${SELLER_PID}" ]] && kill -0 "${SELLER_PID}" 2>/dev/null; then
        kill "${SELLER_PID}" 2>/dev/null || true
        wait "${SELLER_PID}" 2>/dev/null || true
    fi
    rm -f "${SELLER_LOG}" "${SELLER_ERR}"
}
trap cleanup EXIT

# Send one JSON text frame from buyer; wall-clock bounded (no GNU timeout required).
# Background the whole pipeline so websocat keeps stdin from printf (a lone
# background job would get stdin from /dev/null under bash).
send_buyer_message() {
    local payload="$1"
    local status=0
    printf '%s\n' "${payload}" | websocat -t -1 -E "${BUYER_WS}" >/dev/null &
    local cmd_pid=$!
    (
        sleep "${BUYER_TIMEOUT_SECS}"
        if kill -0 "${cmd_pid}" 2>/dev/null; then
            kill "${cmd_pid}" 2>/dev/null || true
        fi
    ) &
    local watcher_pid=$!
    wait "${cmd_pid}" || status=$?
    kill "${watcher_pid}" 2>/dev/null || true
    wait "${watcher_pid}" 2>/dev/null || true
    return "${status}"
}

echo "Seller listening on ${SELLER_WS}"
# -t text frames; -U read-only (WS → stdout); keep stdin closed so we only listen.
websocat -t -U "${SELLER_WS}" < /dev/null >"${SELLER_LOG}" 2>"${SELLER_ERR}" &
SELLER_PID=$!

# Hard cap so a hung seller connection cannot stall forever.
(
    sleep "${SELLER_TIMEOUT_SECS}"
    if kill -0 "${SELLER_PID}" 2>/dev/null; then
        kill "${SELLER_PID}" 2>/dev/null || true
    fi
) &
SELLER_WATCHDOG_PID=$!

sleep "${CONNECT_SETTLE_SECS}"

if ! kill -0 "${SELLER_PID}" 2>/dev/null; then
    echo "FAIL: seller websocat exited before buyer connect (is messaging-service up at ${URL}?)" >&2
    if [[ -s "${SELLER_ERR}" ]]; then
        echo "--- seller stderr ---" >&2
        cat "${SELLER_ERR}" >&2
    fi
    exit 1
fi

echo "Buyer sending ${MESSAGES} message(s) on ${BUYER_WS} (conversationId=${CONVERSATION_ID})"
for ((i = 1; i <= MESSAGES; i++)); do
    body="Synth hello ${i}"
    payload=$(printf '{"conversationId":"%s","body":"%s"}' "${CONVERSATION_ID}" "${body}")
    echo "  → ${payload}"
    if ! send_buyer_message "${payload}"; then
        echo "WARN: buyer send ${i} failed or timed out after ${BUYER_TIMEOUT_SECS}s" >&2
    fi
done

# Wait briefly for fan-out / delivery into the seller log (bounded).
deadline=$((SECONDS + DELIVERY_WAIT_SECS))
while (( SECONDS < deadline )); do
    if grep -q "Synth hello" "${SELLER_LOG}" 2>/dev/null; then
        break
    fi
    if ! kill -0 "${SELLER_PID}" 2>/dev/null; then
        break
    fi
    sleep 0.2
done

# Stop watchdog before assert cleanup; ignore if it already exited.
kill "${SELLER_WATCHDOG_PID}" 2>/dev/null || true
wait "${SELLER_WATCHDOG_PID}" 2>/dev/null || true
SELLER_WATCHDOG_PID=""

if grep -q "Synth hello" "${SELLER_LOG}"; then
    echo "PASS: seller observed at least one 'Synth hello'"
    echo "--- seller observed ---"
    grep "Synth hello" "${SELLER_LOG}" || true
    exit 0
fi

echo "FAIL: seller did not observe 'Synth hello' within ${DELIVERY_WAIT_SECS}s" >&2
echo "--- seller stdout ---" >&2
cat "${SELLER_LOG}" >&2 || true
if [[ -s "${SELLER_ERR}" ]]; then
    echo "--- seller stderr ---" >&2
    cat "${SELLER_ERR}" >&2
fi
exit 1
