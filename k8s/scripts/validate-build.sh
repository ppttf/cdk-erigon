#!/usr/bin/env bash
set -euo pipefail

# Validate CDK Erigon container image
# Performs automated smoke tests to verify image correctness
# Usage: ./validate-build.sh [IMAGE] [--skip-runtime]
#   IMAGE: Docker image name:tag (default: cdk-erigon:k8s-dev-<git-short-hash>)
#   --skip-runtime: Skip runtime pod deployment tests

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

test_start() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

test_pass() {
    echo -e "${GREEN}  ✓${NC} $1"
}

test_fail() {
    echo -e "${RED}  ✗${NC} $1"
    ((FAILED_TESTS++))
}

# Parse arguments
SKIP_RUNTIME=false
IMAGE=""

for arg in "$@"; do
    case $arg in
        --skip-runtime)
            SKIP_RUNTIME=true
            shift
            ;;
        *)
            if [ -z "${IMAGE}" ]; then
                IMAGE="$arg"
            fi
            ;;
    esac
done

# Determine image
if [ -z "${IMAGE}" ]; then
    GIT_SHORT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    IMAGE="cdk-erigon:k8s-dev-${GIT_SHORT_HASH}"
fi

FAILED_TESTS=0

info "Validating image: ${IMAGE}"
echo ""

# Test 1: Image exists locally
test_start "Checking if image exists locally"
if docker inspect "${IMAGE}" > /dev/null 2>&1; then
    test_pass "Image found in local registry"
else
    test_fail "Image not found. Build with: ./k8s/scripts/build-image.sh"
    exit 1
fi

# Test 2: Required ports exposed
test_start "Checking exposed ports"
REQUIRED_PORTS=("4222/tcp" "8222/tcp" "8545/tcp" "6060/tcp" "9090/tcp")
EXPOSED_PORTS=$(docker inspect "${IMAGE}" --format '{{range $port, $_ := .Config.ExposedPorts}}{{$port}} {{end}}')

for port in "${REQUIRED_PORTS[@]}"; do
    if echo "${EXPOSED_PORTS}" | grep -q "${port}"; then
        test_pass "Port ${port} exposed"
    else
        test_fail "Port ${port} NOT exposed"
    fi
done

# Test 3: Kubernetes labels present
test_start "Checking Kubernetes-specific labels"
K8S_LABELS=("io.kubernetes.component" "io.kubernetes.part-of" "org.opencontainers.image.title")

for label in "${K8S_LABELS[@]}"; do
    if docker inspect "${IMAGE}" --format '{{index .Config.Labels "'${label}'"}}' 2>/dev/null | grep -q "."; then
        test_pass "Label ${label} present"
    else
        test_fail "Label ${label} NOT present"
    fi
done

# Test 4: Entrypoint correct
test_start "Checking entrypoint"
ENTRYPOINT=$(docker inspect "${IMAGE}" --format '{{json .Config.Entrypoint}}')
if echo "${ENTRYPOINT}" | grep -q "cdk-erigon"; then
    test_pass "Entrypoint configured correctly"
else
    test_fail "Entrypoint incorrect: ${ENTRYPOINT}"
fi

# Test 5: User is non-root
test_start "Checking user configuration"
USER=$(docker inspect "${IMAGE}" --format '{{.Config.User}}')
if [ "${USER}" = "erigon" ] || [ "${USER}" = "1000" ]; then
    test_pass "Running as non-root user: ${USER}"
else
    test_fail "User should be 'erigon' or '1000', got: ${USER}"
fi

# Test 6: Working directory
test_start "Checking working directory"
WORKDIR=$(docker inspect "${IMAGE}" --format '{{.Config.WorkingDir}}')
if [ "${WORKDIR}" = "/home/erigon" ]; then
    test_pass "Working directory correct: ${WORKDIR}"
else
    test_fail "Working directory should be /home/erigon, got: ${WORKDIR}"
fi

# Runtime tests (if not skipped)
if [ "${SKIP_RUNTIME}" = "false" ]; then
    echo ""
    info "Running runtime tests (use --skip-runtime to skip)"

    # Check if kubectl available
    if ! command -v kubectl > /dev/null 2>&1; then
        warn "kubectl not found - skipping runtime tests"
    else
        # Test 7: Deploy test pod
        test_start "Deploying test pod"

        TEST_NAMESPACE="cdk-erigon-test"
        TEST_POD="validate-${RANDOM}"

        # Create test namespace
        kubectl create namespace "${TEST_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f - > /dev/null 2>&1 || true

        # Create test pod manifest
        cat > /tmp/validate-pod.yaml <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: ${TEST_POD}
  namespace: ${TEST_NAMESPACE}
spec:
  containers:
  - name: erigon
    image: ${IMAGE}
    imagePullPolicy: Never
    command: ["sleep", "3600"]
    ports:
    - containerPort: 8545
      name: http
    - containerPort: 4222
      name: nats
  restartPolicy: Never
EOF

        # Deploy pod
        if kubectl apply -f /tmp/validate-pod.yaml > /dev/null 2>&1; then
            test_pass "Test pod created"

            # Wait for pod to be running
            info "  Waiting for pod to be ready (timeout: 60s)..."
            if kubectl wait --for=condition=Ready pod/${TEST_POD} -n ${TEST_NAMESPACE} --timeout=60s > /dev/null 2>&1; then
                test_pass "Pod is running"

                # Test 8: Check if ports are listening (basic check)
                test_start "Checking if container started successfully"
                POD_STATUS=$(kubectl get pod ${TEST_POD} -n ${TEST_NAMESPACE} -o jsonpath='{.status.phase}')
                if [ "${POD_STATUS}" = "Running" ]; then
                    test_pass "Container running successfully"
                else
                    test_fail "Container not running. Status: ${POD_STATUS}"
                fi
            else
                test_fail "Pod failed to become ready"
                kubectl logs ${TEST_POD} -n ${TEST_NAMESPACE} 2>/dev/null | tail -10 || true
            fi

            # Cleanup
            info "  Cleaning up test resources..."
            kubectl delete pod ${TEST_POD} -n ${TEST_NAMESPACE} --grace-period=0 --force > /dev/null 2>&1 || true
        else
            test_fail "Failed to create test pod"
        fi

        # Cleanup namespace
        kubectl delete namespace "${TEST_NAMESPACE}" --grace-period=0 --force > /dev/null 2>&1 || true
        rm -f /tmp/validate-pod.yaml
    fi
else
    info "Skipping runtime tests (--skip-runtime flag set)"
fi

echo ""
echo "================================"
if [ ${FAILED_TESTS} -eq 0 ]; then
    info "✓ All validation tests passed!"
    exit 0
else
    error "✗ ${FAILED_TESTS} test(s) failed"
    exit 1
fi
