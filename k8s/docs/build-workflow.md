# CDK Erigon K8s Build Workflow

Complete guide for building and deploying CDK Erigon container images for Kubernetes.

---

## Quick Start

```bash
# 1. Build image
./k8s/scripts/build-image.sh

# 2. Load into local cluster
./k8s/scripts/image-load.sh

# 3. Validate
./k8s/scripts/validate-build.sh
```

---

## Prerequisites

### Required Tools
- **Docker** - Container runtime
- **Git** - Version control
- **kubectl** - Kubernetes CLI (for validation)
- **Local K8s cluster** - One of:
  - [kind](https://kind.sigs.k8s.io/) - Recommended
  - [k3d](https://k3d.io/)
  - [minikube](https://minikube.sigs.k8s.io/)

### Verify Installation
```bash
docker --version
git --version
kubectl version --client
kind version  # or k3d version / minikube version
```

---

## Building the Image

### Script: `k8s/scripts/build-image.sh`

**Basic Build:**
```bash
cd /path/to/cdk-erigon
./k8s/scripts/build-image.sh
```

**Output:**
```
[INFO] Building cdk-erigon:k8s-dev-a1b2c3d
[INFO]   Build date: 2025-11-05T10:30:00Z
[INFO]   VCS ref: a1b2c3d...
[INFO]   Dockerfile: k8s/Dockerfile
[INFO] Starting Docker build...
[INFO] ✓ Built cdk-erigon:k8s-dev-a1b2c3d
[INFO]   Image size: 450MB
```

**With Version Tag:**
```bash
./k8s/scripts/build-image.sh v1.2.3
```
Creates image: `cdk-erigon:k8s-v1.2.3`

**With Custom Tag:**
```bash
./k8s/scripts/build-image.sh feature-nats
```
Creates image: `cdk-erigon:k8s-feature-nats`

### Build Arguments

The build script automatically injects:
- `BUILD_DATE` - ISO 8601 timestamp
- `VCS_REF` - Git commit SHA
- `VERSION` - Version tag or dev-<short-hash>

These become image labels for traceability.

### Dockerfile Location

**K8s-specific:** `k8s/Dockerfile`
**Root (DO NOT modify):** `/Dockerfile` - Used by Kurtosis/docker-compose

The k8s/Dockerfile includes:
- ✓ All root Dockerfile functionality
- ✓ NATS ports (4222, 8222, 6222)
- ✓ Kubernetes OCI labels
- ✓ Same build process

### Build Process

1. **Builder stage:** Compile Go binaries with cache mounts
2. **Tools-builder stage:** Compile database tools
3. **Final stage:** Alpine + binaries + configs
4. **Total time:** ~5-10 minutes (first build), ~30s (cached)

### Troubleshooting Builds

**Build fails with "No space left on device":**
```bash
docker system prune -a
docker builder prune
```

**Build is slow:**
- Check Docker resource limits
- Ensure BuildKit is enabled: `export DOCKER_BUILDKIT=1`
- Use cache mounts (already configured)

**Modified files not included:**
```bash
# Check .dockerignore isn't excluding your changes
cat .dockerignore

# Build without cache
docker build --no-cache -f k8s/Dockerfile -t cdk-erigon:k8s-test .
```

---

## Loading into Clusters

### Script: `k8s/scripts/image-load.sh`

**Auto-detect cluster:**
```bash
./k8s/scripts/image-load.sh
```

**Specify image:**
```bash
./k8s/scripts/image-load.sh cdk-erigon:k8s-v1.2.3
```

**Specify cluster name:**
```bash
./k8s/scripts/image-load.sh cdk-erigon:k8s-v1.2.3 my-cluster
```

### Cluster-Specific Commands

**kind:**
```bash
# Create cluster
kind create cluster --name cdk-erigon

# Load image
kind load docker-image cdk-erigon:k8s-dev-abc123 --name cdk-erigon

# Verify
docker exec cdk-erigon-control-plane crictl images | grep cdk-erigon
```

**k3d:**
```bash
# Create cluster
k3d cluster create cdk-erigon

# Load image
k3d image import cdk-erigon:k8s-dev-abc123 --cluster cdk-erigon

# Verify
k3d image list --cluster cdk-erigon | grep cdk-erigon
```

**minikube:**
```bash
# Start cluster
minikube start --profile cdk-erigon

# Load image
minikube image load cdk-erigon:k8s-dev-abc123 --profile cdk-erigon

# Verify
minikube image ls --profile cdk-erigon | grep cdk-erigon
```

### Important: imagePullPolicy

When using loaded images, set `imagePullPolicy: Never` in your manifests:

```yaml
spec:
  containers:
  - name: erigon
    image: cdk-erigon:k8s-dev-abc123
    imagePullPolicy: Never  # Critical for local images
```

---

## Validating the Image

### Script: `k8s/scripts/validate-build.sh`

**Full validation (with runtime tests):**
```bash
./k8s/scripts/validate-build.sh
```

**Quick validation (skip runtime):**
```bash
./k8s/scripts/validate-build.sh --skip-runtime
```

**Validate specific image:**
```bash
./k8s/scripts/validate-build.sh cdk-erigon:k8s-v1.2.3
```

### Tests Performed

**Image Structure:**
- ✓ Image exists locally
- ✓ Required ports exposed (4222, 8222, 8545, 6060, 9090)
- ✓ Kubernetes labels present
- ✓ Entrypoint configured
- ✓ Non-root user (UID 1000)
- ✓ Correct working directory

**Runtime (if not skipped):**
- ✓ Deploy test pod to cluster
- ✓ Wait for pod to be running
- ✓ Verify container started successfully
- ✓ Cleanup test resources

**Output:**
```
[INFO] Validating image: cdk-erigon:k8s-dev-abc123
[TEST] Checking if image exists locally
  ✓ Image found in local registry
[TEST] Checking exposed ports
  ✓ Port 4222/tcp exposed
  ✓ Port 8222/tcp exposed
  ✓ Port 8545/tcp exposed
...
[INFO] ✓ All validation tests passed!
```

---

## Complete Workflow Example

### Local Development Cycle

```bash
# 1. Make code changes
vim zk/datastream/natsstream/stream_server.go

# 2. Build new image
./k8s/scripts/build-image.sh dev-myfeature

# 3. Load into cluster
./k8s/scripts/image-load.sh cdk-erigon:k8s-dev-myfeature

# 4. Validate
./k8s/scripts/validate-build.sh cdk-erigon:k8s-dev-myfeature

# 5. Deploy (Phase 2+)
# helm upgrade cdk-erigon ./k8s/helm/cdk-erigon \
#   --set image.tag=k8s-dev-myfeature
```

### CI/CD Integration

**.github/workflows/k8s-test.yml:**
```yaml
jobs:
  k8s-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup kind
        uses: helm/kind-action@v1
        with:
          cluster_name: cdk-erigon

      - name: Build image
        run: ./k8s/scripts/build-image.sh ${GITHUB_SHA::7}

      - name: Load into kind
        run: ./k8s/scripts/image-load.sh cdk-erigon:k8s-${GITHUB_SHA::7}

      - name: Validate
        run: ./k8s/scripts/validate-build.sh cdk-erigon:k8s-${GITHUB_SHA::7}

      # Future: Deploy and test
```

---

## Version Tagging Strategy

### Development Builds
```bash
./k8s/scripts/build-image.sh
# → cdk-erigon:k8s-dev-abc1234
```
- Auto-generated from git short hash
- For local testing
- Not pushed to registry

### Feature Branches
```bash
./k8s/scripts/build-image.sh feature-nats
# → cdk-erigon:k8s-feature-nats
```
- Named after feature
- For feature testing
- Optionally pushed to dev registry

### Release Candidates
```bash
./k8s/scripts/build-image.sh v1.2.3-rc1
# → cdk-erigon:k8s-v1.2.3-rc1
```
- Semantic versioning with RC suffix
- For staging/QA testing
- Pushed to staging registry

### Production Releases
```bash
./k8s/scripts/build-image.sh v1.2.3
# → cdk-erigon:k8s-v1.2.3
```
- Semantic versioning
- Tagged in git
- Pushed to production registry
- Immutable

---

## Image Registry Integration

### Docker Hub (example)
```bash
# Build
./k8s/scripts/build-image.sh v1.2.3

# Tag for registry
docker tag cdk-erigon:k8s-v1.2.3 gatewayfm/cdk-erigon:k8s-v1.2.3

# Push
docker push gatewayfm/cdk-erigon:k8s-v1.2.3

# Update Helm values
# image:
#   repository: gatewayfm/cdk-erigon
#   tag: k8s-v1.2.3
#   pullPolicy: IfNotPresent
```

### GitHub Container Registry
```bash
# Login
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Tag
docker tag cdk-erigon:k8s-v1.2.3 ghcr.io/gateway-fm/cdk-erigon:k8s-v1.2.3

# Push
docker push ghcr.io/gateway-fm/cdk-erigon:k8s-v1.2.3
```

### Private Registry
```bash
# Tag
docker tag cdk-erigon:k8s-v1.2.3 registry.company.com/cdk-erigon:k8s-v1.2.3

# Push
docker push registry.company.com/cdk-erigon:k8s-v1.2.3

# Create pull secret in K8s
kubectl create secret docker-registry regcred \
  --docker-server=registry.company.com \
  --docker-username=user \
  --docker-password=pass

# Reference in Helm
# imagePullSecrets:
#   - name: regcred
```

---

## Advanced Usage

### Multi-Architecture Builds

```bash
# Enable buildx
docker buildx create --use

# Build for multiple platforms
docker buildx build \
  --file k8s/Dockerfile \
  --platform linux/amd64,linux/arm64 \
  --tag cdk-erigon:k8s-v1.2.3 \
  --push \
  .
```

### Build from Different Branch
```bash
# Checkout branch
git checkout feature/nats-jetstream

# Build
./k8s/scripts/build-image.sh nats-jetstream

# Load and test
./k8s/scripts/image-load.sh cdk-erigon:k8s-nats-jetstream
./k8s/scripts/validate-build.sh cdk-erigon:k8s-nats-jetstream
```

### Inspect Image Contents
```bash
# List files
docker run --rm cdk-erigon:k8s-dev-abc123 ls -la /usr/local/bin/

# Check labels
docker inspect cdk-erigon:k8s-dev-abc123 --format '{{json .Config.Labels}}' | jq

# Check user
docker inspect cdk-erigon:k8s-dev-abc123 --format '{{.Config.User}}'

# Interactive shell
docker run --rm -it --entrypoint sh cdk-erigon:k8s-dev-abc123
```

---

## Troubleshooting

### Image won't load into cluster
```bash
# Verify image exists
docker images | grep cdk-erigon

# Check cluster is running
kind get clusters  # or k3d/minikube equivalent

# Try with full image name
./k8s/scripts/image-load.sh cdk-erigon:k8s-dev-abc123 cdk-erigon
```

### Validation fails
```bash
# Run with detailed output
./k8s/scripts/validate-build.sh cdk-erigon:k8s-dev-abc123

# Skip runtime tests if cluster issues
./k8s/scripts/validate-build.sh --skip-runtime cdk-erigon:k8s-dev-abc123

# Check logs if pod fails
kubectl logs -n cdk-erigon-test validate-xxxxx
```

### Build cache issues
```bash
# Clear Docker cache
docker builder prune -a

# Force rebuild
docker build --no-cache -f k8s/Dockerfile -t cdk-erigon:k8s-test .
```

---

## Next Steps

**Phase 2:** Use this workflow with Helm charts:
```bash
# Build
./k8s/scripts/build-image.sh v1.0.0

# Load
./k8s/scripts/image-load.sh cdk-erigon:k8s-v1.0.0

# Deploy via Helm
helm install my-chain ./k8s/helm/cdk-erigon \
  --set image.tag=k8s-v1.0.0 \
  --set image.pullPolicy=Never
```

---

*Last Updated: 2025-11-05 | Phase 1: Container Foundation*
