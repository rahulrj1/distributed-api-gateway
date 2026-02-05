#!/bin/bash
# Run all E2E tests for the distributed API gateway

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BOLD}======================================${NC}"
echo -e "${BOLD}  Distributed API Gateway E2E Tests  ${NC}"
echo -e "${BOLD}======================================${NC}"
echo ""

cd "$PROJECT_ROOT"

# Generate keys FIRST (gateway needs public key at startup)
if [ ! -f "keys/private.pem" ]; then
    echo "Generating JWT keys..."
    mkdir -p keys
    openssl genrsa -out keys/private.pem 2048 2>/dev/null
    openssl rsa -in keys/private.pem -pubout -out keys/public.pem 2>/dev/null
    echo "Keys generated"
fi

# Check if services are running
echo "Checking services..."
if ! docker ps | grep -q gateway-1; then
    echo -e "${YELLOW}Services not running. Starting docker compose...${NC}"
    docker compose up -d
    echo "Waiting for services to be healthy..."
    sleep 15
fi

# Set gateway URLs (each test generates its own token)
export GATEWAY_1="http://localhost:5000"
export GATEWAY_2="http://localhost:5001"
export GATEWAY_URL="$GATEWAY_1"

echo ""
TESTS_PASSED=0
TESTS_FAILED=0

# Run routing tests
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if bash "$SCRIPT_DIR/test_routing.sh"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""

# Run auth tests
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if bash "$SCRIPT_DIR/test_auth.sh"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""

# Run rate limit tests
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if bash "$SCRIPT_DIR/test_ratelimit.sh"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""

# Run circuit breaker tests
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if bash "$SCRIPT_DIR/test_circuitbreaker.sh"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""
echo -e "${BOLD}======================================${NC}"
echo -e "${BOLD}           TEST SUMMARY              ${NC}"
echo -e "${BOLD}======================================${NC}"
echo ""
echo -e "  Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "  Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ "$TESTS_FAILED" -eq 0 ]; then
    echo -e "${GREEN}${BOLD}All E2E tests passed!${NC}"
    exit 0
else
    echo -e "${RED}${BOLD}Some tests failed!${NC}"
    exit 1
fi
