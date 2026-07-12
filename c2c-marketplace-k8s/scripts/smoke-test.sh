#!/usr/bin/env bash
set -euo pipefail

LISTINGS_URL="${LISTINGS_URL:-http://localhost:8081}"
PAYMENTS_URL="${PAYMENTS_URL:-http://localhost:8084}"
START=$(date +%s)

wait_for() {
    local url="$1" name="$2" n=0 max=30
    printf "Waiting for %s" "$name"
    while ! curl -sf "$url/healthz" > /dev/null 2>&1; do
        n=$((n + 1))
        [ $n -ge $max ] && { echo " TIMEOUT"; echo "FAIL: $name did not become healthy after ${max}s"; exit 1; }
        printf "."
        sleep 1
    done
    echo " ready"
}

wait_for "$LISTINGS_URL" "listings-service"
wait_for "$PAYMENTS_URL" "payments-service"

echo ""
echo "=== Step 1: Create listing ==="
LISTING_RESP=$(curl -sf -X POST "$LISTINGS_URL/listings" \
    -H "Content-Type: application/json" \
    -d '{"sellerId":"smoke-seller","title":"Smoke Test Bike","priceCents":5000,"category":"sporting-goods","lat":47.6062,"lon":-122.3321}')
echo "Response: $LISTING_RESP"
LISTING_ID=$(echo "$LISTING_RESP" | jq -r '.id // empty')
[ -n "$LISTING_ID" ] || { echo "FAIL step 1: no id in response"; exit 1; }
echo "listingId=$LISTING_ID"

echo ""
echo "=== Step 2: Place order ==="
ORDER_RESP=$(curl -sf -X POST "$PAYMENTS_URL/orders" \
    -H "Content-Type: application/json" \
    -d "{\"listingId\":\"$LISTING_ID\",\"buyerId\":\"smoke-buyer\",\"sellerId\":\"smoke-seller\",\"amountCents\":5000}")
echo "Response: $ORDER_RESP"
ORDER_ID=$(echo "$ORDER_RESP" | jq -r '.id // empty')
ORDER_STATUS=$(echo "$ORDER_RESP" | jq -r '.status // empty')
[ -n "$ORDER_ID" ] || { echo "FAIL step 2: no id in response"; exit 1; }
[ "$ORDER_STATUS" = "HELD" ] || { echo "FAIL step 2: expected status=HELD, got '$ORDER_STATUS'"; exit 1; }
echo "orderId=$ORDER_ID status=$ORDER_STATUS"

echo ""
echo "=== Step 3: Confirm delivery ==="
CONFIRM_RESP=$(curl -sf -X POST "$PAYMENTS_URL/orders/$ORDER_ID/confirm-delivery")
echo "Response: $CONFIRM_RESP"
FINAL_STATUS=$(echo "$CONFIRM_RESP" | jq -r '.status // empty')
[ "$FINAL_STATUS" = "RELEASED" ] || { echo "FAIL step 3: expected status=RELEASED, got '$FINAL_STATUS'"; exit 1; }
echo "status=$FINAL_STATUS"

END=$(date +%s)
echo ""
echo "PASS ($((END - START))s)"
