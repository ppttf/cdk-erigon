#!/usr/bin/env bash
set -euo pipefail

# Build L1 RPC Cache Proxy container image
# Usage: ./build-l1-proxy.sh [TAG]
#   TAG: Optional image tag (default: local)

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

cd "${PROJECT_ROOT}"

TAG="${1:-local}"
IMAGE_NAME="cdk-erigon-l1-proxy"
FULL_IMAGE="${IMAGE_NAME}:${TAG}"

info "Building ${FULL_IMAGE}"
info "  Dockerfile: k8s/l1-proxy/Dockerfile"

if [ ! -f "k8s/l1-proxy/Dockerfile" ]; then
    error "Dockerfile not found at k8s/l1-proxy/Dockerfile"
fi

info "Starting Docker build..."
docker build \
    --file "k8s/l1-proxy/Dockerfile" \
    --tag "${FULL_IMAGE}" \
    . || error "Docker build failed"

info "Built ${FULL_IMAGE}"

if ! docker inspect "${FULL_IMAGE}" > /dev/null 2>&1; then
    error "Image verification failed"
fi

IMAGE_SIZE=$(docker images "${FULL_IMAGE}" --format "{{.Size}}")
info "  Image size: ${IMAGE_SIZE}"

info ""
info "Build complete!"
info ""
info "Next steps:"
info "  1. Load into cluster: ./k8s/scripts/image-load.sh ${FULL_IMAGE}"
info ""
