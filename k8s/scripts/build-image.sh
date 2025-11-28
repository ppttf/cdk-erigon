#!/usr/bin/env bash
set -euo pipefail

# Build CDK Erigon container image for Kubernetes
# Usage: ./build-image.sh [OPTIONS] [VERSION]
#   VERSION: Optional version tag (default: dev-<git-short-hash>)
#   --with-l1-proxy: Also build the L1 RPC cache proxy image
#   --l1-proxy-only: Only build the L1 RPC cache proxy image

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

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
K8S_DIR="${PROJECT_ROOT}/k8s"

# Change to project root
cd "${PROJECT_ROOT}"

# Parse arguments
BUILD_L1_PROXY=false
L1_PROXY_ONLY=false
VERSION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --with-l1-proxy)
            BUILD_L1_PROXY=true
            shift
            ;;
        --l1-proxy-only)
            L1_PROXY_ONLY=true
            BUILD_L1_PROXY=true
            shift
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

# Determine version
if [ -z "$VERSION" ]; then
    GIT_SHORT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    VERSION="dev-${GIT_SHORT_HASH}"
fi

# Build metadata
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
VCS_REF=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

# Build main image unless --l1-proxy-only
if [ "$L1_PROXY_ONLY" = false ]; then
    IMAGE_NAME="cdk-erigon"
    IMAGE_TAG="k8s-${VERSION}"
    FULL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"

    info "Building ${FULL_IMAGE}"
    info "  Build date: ${BUILD_DATE}"
    info "  VCS ref: ${VCS_REF}"
    info "  Dockerfile: k8s/Dockerfile"

    # Check if Dockerfile exists
    if [ ! -f "${K8S_DIR}/Dockerfile" ]; then
        error "Dockerfile not found at ${K8S_DIR}/Dockerfile"
    fi

    # Build the image
    info "Starting Docker build..."
    docker build \
        --file "${K8S_DIR}/Dockerfile" \
        --tag "${FULL_IMAGE}" \
        --build-arg BUILD_DATE="${BUILD_DATE}" \
        --build-arg VCS_REF="${VCS_REF}" \
        --build-arg VERSION="${VERSION}" \
        . || error "Docker build failed"

    info "Built ${FULL_IMAGE}"

    # Verify the image
    info "Verifying image..."
    if ! docker inspect "${FULL_IMAGE}" > /dev/null 2>&1; then
        error "Image verification failed - image not found in local registry"
    fi

    # Show image size
    IMAGE_SIZE=$(docker images "${FULL_IMAGE}" --format "{{.Size}}")
    info "  Image size: ${IMAGE_SIZE}"

    # Show exposed ports
    info "  Exposed ports:"
    docker inspect "${FULL_IMAGE}" --format '{{range $port, $_ := .Config.ExposedPorts}}  - {{$port}}{{println}}{{end}}' | head -15

    info ""
    info "Build complete: ${FULL_IMAGE}"
fi

# Build L1 proxy if requested
if [ "$BUILD_L1_PROXY" = true ]; then
    L1_PROXY_IMAGE="cdk-erigon-l1-proxy:${VERSION}"

    info ""
    info "Building L1 proxy: ${L1_PROXY_IMAGE}"
    info "  Dockerfile: k8s/l1-proxy/Dockerfile"

    if [ ! -f "${K8S_DIR}/l1-proxy/Dockerfile" ]; then
        error "L1 proxy Dockerfile not found at ${K8S_DIR}/l1-proxy/Dockerfile"
    fi

    docker build \
        --file "${K8S_DIR}/l1-proxy/Dockerfile" \
        --tag "${L1_PROXY_IMAGE}" \
        . || error "L1 proxy Docker build failed"

    L1_PROXY_SIZE=$(docker images "${L1_PROXY_IMAGE}" --format "{{.Size}}")
    info "Built ${L1_PROXY_IMAGE} (${L1_PROXY_SIZE})"
fi

info ""
info "Build complete!"
info ""
info "Next steps:"
if [ "$L1_PROXY_ONLY" = false ]; then
    info "  1. Load into cluster: ./k8s/scripts/image-load.sh ${FULL_IMAGE}"
    info "  2. Validate build: ./k8s/scripts/validate-build.sh ${FULL_IMAGE}"
fi
if [ "$BUILD_L1_PROXY" = true ]; then
    info "  Load L1 proxy: ./k8s/scripts/image-load.sh ${L1_PROXY_IMAGE}"
fi
info ""
