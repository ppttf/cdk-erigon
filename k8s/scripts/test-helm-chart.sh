#!/usr/bin/env bash
set -euo pipefail

# Helm Chart Test Suite
# Enforces TDD and validates all templates

CHART_DIR="${1:-k8s/helm}"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Helm Chart Test Suite ===${NC}"
echo "Chart: $CHART_DIR"
echo ""

# Track failures
FAILURES=0

# 1. Unit Tests (helm-unittest)
echo -e "${YELLOW}1. Running unit tests...${NC}"
if helm unittest "$CHART_DIR" 2>&1; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    ((FAILURES++))
fi
echo ""

# 2. Chart Linting
echo -e "${YELLOW}2. Linting chart...${NC}"
if helm lint "$CHART_DIR" 2>&1; then
    echo -e "${GREEN}✓ Lint passed${NC}"
else
    echo -e "${RED}✗ Lint failed${NC}"
    ((FAILURES++))
fi
echo ""

# 3. Values Schema Validation
echo -e "${YELLOW}3. Validating values schema...${NC}"
if [ -f "$CHART_DIR/values.schema.json" ]; then
    if helm template test "$CHART_DIR" --validate > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Values schema valid${NC}"
    else
        echo -e "${RED}✗ Values schema validation failed${NC}"
        ((FAILURES++))
    fi
else
    echo -e "${YELLOW}⚠ No values.schema.json found (will create in Phase 2)${NC}"
fi
echo ""

# 4. Template Rendering Validation
echo -e "${YELLOW}4. Validating template rendering...${NC}"
if [ -d "k8s/test-values" ]; then
    TEST_COUNT=0
    TEST_FAILURES=0

    for values_file in k8s/test-values/*.yaml; do
        ((TEST_COUNT++))
        echo -n "  Testing with $(basename "$values_file")... "

        if helm template test "$CHART_DIR" \
            --values "$values_file" \
            > /tmp/rendered.yaml 2>&1; then

            # Validate with kubeconform
            if kubeconform -strict -summary /tmp/rendered.yaml > /dev/null 2>&1; then
                echo -e "${GREEN}✓${NC}"
            else
                echo -e "${RED}✗ (kubeconform)${NC}"
                ((TEST_FAILURES++))
            fi
        else
            echo -e "${RED}✗ (render)${NC}"
            ((TEST_FAILURES++))
        fi
    done

    if [ $TEST_FAILURES -eq 0 ]; then
        echo -e "${GREEN}✓ All $TEST_COUNT rendering tests passed${NC}"
    else
        echo -e "${RED}✗ $TEST_FAILURES/$TEST_COUNT rendering tests failed${NC}"
        ((FAILURES++))
    fi
else
    echo -e "${YELLOW}⚠ No k8s/test-values/ directory found${NC}"
fi
echo ""

# 5. Optional: Dry-run installation (if cluster available)
if kubectl cluster-info &>/dev/null; then
    echo -e "${YELLOW}5. Testing dry-run installation...${NC}"
    if helm install test-release "$CHART_DIR" \
        --values k8s/test-values/minimal.yaml \
        --dry-run --debug > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Dry-run installation successful${NC}"
    else
        echo -e "${RED}✗ Dry-run installation failed${NC}"
        ((FAILURES++))
    fi
else
    echo -e "${YELLOW}5. Skipping dry-run (no cluster available)${NC}"
fi
echo ""

# Summary
echo -e "${YELLOW}=== Test Summary ===${NC}"
if [ $FAILURES -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ $FAILURES test suite(s) failed${NC}"
    echo ""
    echo "Fix the failures and run again:"
    echo "  ./k8s/scripts/test-helm-chart.sh"
    exit 1
fi
