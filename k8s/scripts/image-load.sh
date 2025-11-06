#!/usr/bin/env bash
set -euo pipefail

# Load CDK Erigon image into local Kubernetes cluster
# Supports: kind, k3d, minikube
# Usage: ./image-load.sh [IMAGE] [CLUSTER_NAME]
#   IMAGE: Docker image name:tag (default: cdk-erigon:k8s-dev-<git-short-hash>)
#   CLUSTER_NAME: Cluster name (default: auto-detect or "cdk-erigon")

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Determine image
if [ $# -ge 1 ]; then
    IMAGE="$1"
else
    GIT_SHORT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    IMAGE="cdk-erigon:k8s-dev-${GIT_SHORT_HASH}"
fi

# Determine cluster name
DEFAULT_CLUSTER="cdk-erigon"
CLUSTER_NAME="${2:-${DEFAULT_CLUSTER}}"

info "Loading image: ${IMAGE}"

# Verify image exists locally
if ! docker inspect "${IMAGE}" > /dev/null 2>&1; then
    error "Image ${IMAGE} not found locally. Build it first with: ./k8s/scripts/build-image.sh"
fi

# Detect cluster type and load image
LOADED=false

# Check for OrbStack (images are automatically available)
if kubectl config current-context 2>/dev/null | grep -q "orbstack"; then
    info "Detected OrbStack cluster"
    info "✓ OrbStack automatically makes Docker images available to Kubernetes"
    info "  No loading step required - image is ready to use"
    LOADED=true
fi

# Try kind
if command -v kind > /dev/null 2>&1; then
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        info "Detected kind cluster: ${CLUSTER_NAME}"
        info "Loading image into kind..."
        kind load docker-image "${IMAGE}" --name "${CLUSTER_NAME}" || error "Failed to load image into kind"
        LOADED=true
    fi
fi

# Try k3d if not loaded
if [ "${LOADED}" = "false" ] && command -v k3d > /dev/null 2>&1; then
    if k3d cluster list 2>/dev/null | grep -q "^${CLUSTER_NAME}"; then
        info "Detected k3d cluster: ${CLUSTER_NAME}"
        info "Loading image into k3d..."
        k3d image import "${IMAGE}" --cluster "${CLUSTER_NAME}" || error "Failed to load image into k3d"
        LOADED=true
    fi
fi

# Try minikube if not loaded
if [ "${LOADED}" = "false" ] && command -v minikube > /dev/null 2>&1; then
    if minikube status --profile "${CLUSTER_NAME}" 2>/dev/null | grep -q "Running"; then
        info "Detected minikube cluster: ${CLUSTER_NAME}"
        info "Loading image into minikube..."
        minikube image load "${IMAGE}" --profile "${CLUSTER_NAME}" || error "Failed to load image into minikube"
        LOADED=true
    fi
fi

# Check if image was loaded
if [ "${LOADED}" = "false" ]; then
    error "No running cluster found named '${CLUSTER_NAME}'. Supported: kind, k3d, minikube"
fi

info "✓ Image loaded successfully"
info ""
info "Verify with:"
info "  kind:     docker exec ${CLUSTER_NAME}-control-plane crictl images | grep cdk-erigon"
info "  k3d:      k3d image list --cluster ${CLUSTER_NAME}"
info "  minikube: minikube image ls --profile ${CLUSTER_NAME} | grep cdk-erigon"
info ""
info "Next step: ./k8s/scripts/validate-build.sh ${IMAGE}"
info ""
