#!/bin/bash
# E2E Test: Rate Limiting - Verifies cross-instance rate limiting works

set -e

GATEWAY_1="${GATEWAY_1:-http://localhost:5000}"
GATEWAY_2="${GATEWAY_2:-http://localhost:5001}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; exit 1; }

echo "=== E2E Test: Rate Limiting ==="
echo "Gateway 1: $GATEWAY_1"
echo "Gateway 2: $GATEWAY_2"
echo ""

cd "$PROJECT_ROOT"

# Ensure keys exist
if [ ! -f "keys/private.pem" ]; then
    mkdir -p keys
    openssl genrsa -out keys/private.pem 2048 2>/dev/null
    openssl rsa -in keys/private.pem -pubout -out keys/public.pem 2>/dev/null
fi

RATE_LIMIT=100

# Test 1: Under limit - requests should succeed (use unique client)
echo "Test 1: Under limit requests..."
TOKEN=$(go run scripts/generate_jwt.go -sub test1 -client_id "under-$(date +%s)" -exp 1h)
SUCCESS_COUNT=0
for i in $(seq 1 5); do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_1/service-a/hello")
    if [ "$STATUS" = "200" ]; then
        ((SUCCESS_COUNT++))
    fi
done
if [ "$SUCCESS_COUNT" -eq 5 ]; then
    pass "5 requests under limit all succeeded"
else
    fail "Under limit requests should succeed, only $SUCCESS_COUNT/5 succeeded"
fi

# Test 2: Cross-instance rate limiting (120 requests across 2 gateways, limit=100)
echo "Test 2: Cross-instance rate limiting..."
TOKEN=$(go run scripts/generate_jwt.go -sub test2 -client_id "cross-$(date +%s)" -exp 1h)

G1_SUCCESS=0
G2_SUCCESS=0
G1_BLOCKED=0
G2_BLOCKED=0

# Send 60 requests to gateway-1
for i in $(seq 1 60); do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_1/service-a/hello")
    if [ "$STATUS" = "200" ]; then ((G1_SUCCESS++)); elif [ "$STATUS" = "429" ]; then ((G1_BLOCKED++)); fi
done

# Send 60 requests to gateway-2 (should share rate limit counter via Redis)
for i in $(seq 1 60); do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_2/service-a/hello")
    if [ "$STATUS" = "200" ]; then ((G2_SUCCESS++)); elif [ "$STATUS" = "429" ]; then ((G2_BLOCKED++)); fi
done

TOTAL_SUCCESS=$((G1_SUCCESS + G2_SUCCESS))
TOTAL_BLOCKED=$((G1_BLOCKED + G2_BLOCKED))

echo "  Gateway 1: $G1_SUCCESS succeeded, $G1_BLOCKED blocked"
echo "  Gateway 2: $G2_SUCCESS succeeded, $G2_BLOCKED blocked"
echo "  Total: $TOTAL_SUCCESS succeeded, $TOTAL_BLOCKED blocked"

if [ "$TOTAL_SUCCESS" -le "$RATE_LIMIT" ] && [ "$TOTAL_BLOCKED" -gt 0 ]; then
    pass "Cross-instance rate limiting working (shared counter)"
else
    fail "Expected ~100 success and ~20 blocked, got $TOTAL_SUCCESS/$TOTAL_BLOCKED"
fi

# Test 3: 429 response includes Retry-After header
echo "Test 3: Retry-After header on 429..."
TOKEN=$(go run scripts/generate_jwt.go -sub test3 -client_id "retry-$(date +%s)" -exp 1h)

# Exhaust the limit
for i in $(seq 1 $((RATE_LIMIT + 5))); do
    curl -s -o /dev/null -H "Authorization: Bearer $TOKEN" "$GATEWAY_1/service-a/hello" &
done
wait

# Check for 429 with Retry-After
RESPONSE=$(curl -s -D - -H "Authorization: Bearer $TOKEN" "$GATEWAY_1/service-a/hello")
if echo "$RESPONSE" | grep -q "429" && echo "$RESPONSE" | grep -qi "Retry-After"; then
    pass "429 response includes Retry-After header"
else
    fail "Expected 429 with Retry-After header"
fi

echo ""
echo -e "${GREEN}All rate limiting tests passed!${NC}"
