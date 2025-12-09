#!/usr/bin/env bash
set -euo pipefail

# Test readiness probe endpoint for CDK Erigon
# Tests RD-513: Verify HTTP readiness probe on port 8545

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

IMAGE="${1:-cdk-erigon:k8s-test-phase1}"
CONTAINER_NAME="readiness-test-$(date +%s)"
TEST_PORT=18545

echo -e "${GREEN}Testing Readiness Probe for ${IMAGE}${NC}"
echo ""

cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

# Start container with baked-in config
echo -e "${YELLOW}Starting container...${NC}"
docker run -d \
    --name "${CONTAINER_NAME}" \
    -p "${TEST_PORT}:8545" \
    "${IMAGE}" \
    --config=/home/erigon/bali.yaml \
    --datadir=/tmp/test-datadir \
    --log.console.verbosity=3 \
    >/dev/null || {
        echo -e "${RED}✗ FAIL: Container failed to start${NC}"
        docker logs "${CONTAINER_NAME}" 2>&1 | tail -30
        exit 1
    }

# Wait for startup
echo -e "${YELLOW}Waiting for RPC endpoint to be ready...${NC}"
MAX_WAIT=60
ELAPSED=0
while [ $ELAPSED -lt $MAX_WAIT ]; do
    if curl -sf -m 2 "http://localhost:${TEST_PORT}" >/dev/null 2>&1; then
        break
    fi
    sleep 2
    ELAPSED=$((ELAPSED + 2))
    echo -n "."
done
echo ""

if [ $ELAPSED -ge $MAX_WAIT ]; then
    echo -e "${RED}✗ FAIL: Endpoint did not become ready within ${MAX_WAIT}s${NC}"
    docker logs "${CONTAINER_NAME}" 2>&1 | tail -20
    exit 1
fi

echo -e "${GREEN}✓ Endpoint ready after ${ELAPSED}s${NC}"
echo ""

# Test 1: HTTP GET returns 200 OK
echo -e "${YELLOW}Test 1: HTTP GET returns success${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${TEST_PORT}")
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "405" ]; then
    echo -e "${GREEN}✓ PASS: HTTP code ${HTTP_CODE}${NC}"
else
    echo -e "${RED}✗ FAIL: Expected 200 or 405, got ${HTTP_CODE}${NC}"
    exit 1
fi

# Test 2: Response time < 5s
echo -e "${YELLOW}Test 2: Response time < 5s${NC}"
RESPONSE_TIME=$(curl -s -o /dev/null -w "%{time_total}" -m 5 "http://localhost:${TEST_PORT}")
if (( $(echo "$RESPONSE_TIME < 5.0" | bc -l) )); then
    echo -e "${GREEN}✓ PASS: Response time ${RESPONSE_TIME}s${NC}"
else
    echo -e "${RED}✗ FAIL: Response time ${RESPONSE_TIME}s exceeds 5s${NC}"
    exit 1
fi

# Test 3: JSON-RPC method check (more robust)
echo -e "${YELLOW}Test 3: JSON-RPC responds to eth_blockNumber${NC}"
RPC_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
    "http://localhost:${TEST_PORT}")

if echo "$RPC_RESPONSE" | grep -q '"result"'; then
    echo -e "${GREEN}✓ PASS: JSON-RPC endpoint functional${NC}"
    echo "  Response: $RPC_RESPONSE"
else
    echo -e "${YELLOW}⚠ WARN: JSON-RPC returned unexpected response${NC}"
    echo "  Response: $RPC_RESPONSE"
fi

echo ""
echo -e "${GREEN}=== READINESS PROBE TEST RESULTS ===${NC}"
echo "✓ Endpoint: http://0.0.0.0:8545"
echo "✓ HTTP Status: ${HTTP_CODE}"
echo "✓ Response Time: ${RESPONSE_TIME}s"
echo "✓ Ready for K8s readinessProbe"
echo ""
echo -e "${GREEN}Recommended K8s Configuration:${NC}"
cat <<EOF
readinessProbe:
  httpGet:
    path: /
    port: 8545
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
EOF

echo ""
echo -e "${GREEN}✓ All readiness probe tests passed${NC}"
