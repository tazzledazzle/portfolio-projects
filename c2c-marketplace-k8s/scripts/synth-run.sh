#!/usr/bin/env bash
# Run a synth profile: health-check services → HTTP harness → WebSocket chat → unified PASS/FAIL.
set -euo pipefail

PROFILE_NAME="${1:-demo}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

export LISTINGS_URL="${LISTINGS_URL:-http://localhost:8081}"
export SEARCH_URL="${SEARCH_URL:-http://localhost:8082}"
export MESSAGING_WS_URL="${MESSAGING_WS_URL:-ws://localhost:8083}"
export PAYMENTS_URL="${PAYMENTS_URL:-http://localhost:8084}"

if ! command -v jq >/dev/null 2>&1; then
    echo "ERROR: jq is required but was not found on PATH." >&2
    echo "Install it with: brew install jq" >&2
    exit 1
fi

PROFILE_FILE="$ROOT/synth/profiles/${PROFILE_NAME}.json"
if [[ ! -f "$PROFILE_FILE" ]]; then
    echo "ERROR: profile not found: $PROFILE_FILE" >&2
    exit 1
fi

wait_for() {
    local url="$1" name="$2" n=0 max=30
    printf "Waiting for %s" "$name"
    while ! curl -sf "$url/healthz" > /dev/null 2>&1; do
        n=$((n + 1))
        [ "$n" -ge "$max" ] && { echo " TIMEOUT"; echo "FAIL: $name did not become healthy after ${max}s"; exit 1; }
        printf "."
        sleep 1
    done
    echo " ready"
}

wait_for "$LISTINGS_URL" "listings-service"
wait_for "$SEARCH_URL" "search-service"
wait_for "$PAYMENTS_URL" "payments-service"

echo ""
echo "=== Harness: profile=${PROFILE_NAME} ==="
set +e
HARNESS_OUT=$(cd "$ROOT" && ./gradlew :synth-harness:run --args="--profile ${PROFILE_FILE}" 2>&1)
HARNESS_EXIT=$?
set -e

# Prefer last compact JSON line; fall back to last pretty-printed object ({ ... }).
extract_summary() {
    local out="$1"
    local line
    line=$(printf '%s\n' "$out" | grep -E '^\{.*\}$' | tail -n1 || true)
    if [[ -n "$line" ]] && printf '%s\n' "$line" | jq -e . >/dev/null 2>&1; then
        printf '%s\n' "$line"
        return 0
    fi
    # Pretty-printed Summary from kotlinx.serialization (opening { through closing }).
    printf '%s\n' "$out" | awk '
        /^\{/ { capturing=1; buf=$0; next }
        capturing {
            buf = buf ORS $0
            if (/^\}/) { last=buf; capturing=0 }
        }
        END { if (last != "") print last }
    '
}

SUMMARY=$(extract_summary "$HARNESS_OUT" || true)
if [[ -z "${SUMMARY}" ]] || ! printf '%s\n' "$SUMMARY" | jq -e . >/dev/null 2>&1; then
    echo "FAIL: could not parse JSON summary from synth-harness stdout" >&2
    printf '%s\n' "$HARNESS_OUT" >&2
    exit 1
fi

CHAT_PAIRS=$(jq -r '.chatPairs // 0' "$PROFILE_FILE")
MESSAGES_PER_PAIR=$(jq -r '.messagesPerPair // 3' "$PROFILE_FILE")

CHAT_OK=true
if [[ "$CHAT_PAIRS" -gt 0 ]]; then
    echo ""
    echo "=== Chat: ${CHAT_PAIRS} pair(s), ${MESSAGES_PER_PAIR} message(s) each ==="
    for ((i = 0; i < CHAT_PAIRS; i++)); do
        echo "--- chat pair ${i}: synth-buyer-${i} / synth-seller-${i} ---"
        if ! "$ROOT/scripts/synth-chat.sh" \
            --url "$MESSAGING_WS_URL" \
            --buyer "synth-buyer-${i}" \
            --seller "synth-seller-${i}" \
            --messages "$MESSAGES_PER_PAIR"; then
            CHAT_OK=false
        fi
    done
fi

if [[ "$CHAT_OK" == true ]]; then
    SUMMARY=$(printf '%s\n' "$SUMMARY" | jq -c '.chatOk = true')
else
    SUMMARY=$(printf '%s\n' "$SUMMARY" | jq -c '.chatOk = false')
fi

echo ""
echo "=== Summary ==="
printf '%s\n' "$SUMMARY" | jq .

# summary.ok() semantics: errors empty && created > 0; plus harness exit and chatOk.
SUMMARY_OK=$(printf '%s\n' "$SUMMARY" | jq -r '((.errors // []) | length) == 0 and (.created // 0) > 0')
CHAT_FIELD_OK=$(printf '%s\n' "$SUMMARY" | jq -r '.chatOk == true')

if [[ "$HARNESS_EXIT" -eq 0 && "$SUMMARY_OK" == true && "$CHAT_FIELD_OK" == true ]]; then
    echo "PASS"
    exit 0
fi

echo "FAIL (harness_exit=${HARNESS_EXIT} summary_ok=${SUMMARY_OK} chatOk=${CHAT_FIELD_OK})"
exit 1
