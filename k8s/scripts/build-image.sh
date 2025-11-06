#!/usr/bin/env bash
set -euo pipefail

# Build CDK Erigon container image for Kubernetes
# Usage: ./build-image.sh [VERSION]
#   VERSION: Optional version tag (default: dev-<git-short-hash>)

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

# Determine version
if [ $# -ge 1 ]; then
    VERSION="$1"
else
    GIT_SHORT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    VERSION="dev-${GIT_SHORT_HASH}"
fi

# Build metadata
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
VCS_REF=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

# Image name
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

info "✓ Built ${FULL_IMAGE}"

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
info "✓ Build complete!"
info ""
info "Next steps:"
info "  1. Load into cluster: ./k8s/scripts/image-load.sh ${FULL_IMAGE}"
info "  2. Validate build: ./k8s/scripts/validate-build.sh ${FULL_IMAGE}"
info ""
