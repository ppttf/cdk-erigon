# CDK Erigon Kubernetes Deployment Project

## Overview

**Goal:** Replace Kurtosis-based testing infrastructure with native Kubernetes deployment using Helm charts.

**Why:** Kurtosis has proven difficult to maintain due to complex dependencies and fast-moving API changes. This project delivers a simpler, more robust "blockchain in a box" solution using standard K8s tooling.

**Timeline:** 12 weeks @ 16 hours/week = 192 hours total
**Start Date:** 2025-10-31
**Target Date:** 2026-01-23

---

## Core Principles & Rules

### 1. Isolation
- **All K8s work lives in `/k8s` directory**
- Do NOT modify files outside `/k8s` unless absolutely necessary
- Keep K8s infrastructure separate from existing deployment methods

### 2. Non-Breaking Changes
- **DO NOT modify existing `Dockerfile`** - Kurtosis and docker-compose use it
- Use `k8s/Dockerfile` for K8s-specific container image
- Existing CI/CD workflows must continue working
- Kurtosis tests must keep passing

### 3. Documentation First
- Document decisions in this file
- Each phase gets detailed design doc in `k8s/docs/phases/`
- Scripts must have clear comments and usage examples
- Migration guide for Kurtosis users

### 4. Standard K8s Patterns
- Use Helm charts (not raw manifests for deployment)
- Follow K8s best practices (health probes, graceful shutdown, etc.)
- Support kind, k3d, and minikube for local development
- Production-ready from day one

### 5. Zero Data Loss
- NATS JetStream persistence is critical
- Graceful shutdown must complete within grace period
- Test all lifecycle events (start, restart, upgrade, terminate)
- Validate message persistence across restarts

### 6. Sequential Documentation
- **Every change gets a numbered document** in `k8s/docs/phases/phase-N/`
- Document structure: `NN-descriptive-name.md` (e.g., `01-design.md`, `02-build-optimization.md`)
- Each document explains: What changed, Why we changed it, Design decisions, Trade-offs
- These documents become source material for:
  - Pull request descriptions
  - GitHub implementation documentation
  - Project handoff materials
- Maintain a `README.md` in each phase folder summarizing all changes

---

## Project Structure

```
k8s/
├── PROJECT.md                    # This file - project overview & rules
├── Dockerfile                    # K8s-specific container image (DO NOT use root /Dockerfile)
├── docs/
│   ├── phases/                   # Sequential phase documentation
│   │   ├── phase-1/              # Phase 1: Container Foundation
│   │   │   ├── README.md         # Phase overview & index
│   │   │   ├── 01-design.md      # Initial design decisions
│   │   │   ├── 02-build-optimization.md  # Build context optimization
│   │   │   └── ...               # Additional sequential changes
│   │   ├── phase-2/              # Phase 2: Helm Chart Structure
│   │   │   ├── README.md
│   │   │   └── ...
│   │   └── ...
│   ├── container-requirements.md # Container specs for Helm chart
│   ├── build-workflow.md         # How to build and load images
│   ├── shutdown-sequence.md      # Graceful shutdown design
│   ├── health-probes.md          # Readiness/liveness configuration
│   └── migration-from-kurtosis.md # Migration guide
├── scripts/
│   ├── build-image.sh            # Build k8s/Dockerfile
│   ├── image-load.sh             # Load into kind/k3d/minikube
│   ├── validate-build.sh         # Validate image meets requirements
│   ├── deploy.sh                 # Deploy via Helm
│   ├── undeploy.sh               # Clean teardown
│   ├── port-forward.sh           # Access services locally
│   └── logs.sh                   # View component logs
├── helm/
│   └── cdk-erigon/               # Helm chart (Phase 2)
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
└── base/                         # Existing kustomize manifests (reference only)
    ├── statefulset.yaml
    └── services/
```

---

## Phase Overview

### Phase 1: Container Foundation (Weeks 1-2) ✅ COMPLETE
**Status:** 100% Complete (13/13 tasks done)
**Started:** 2025-10-31
**Completed:** 2025-11-06

**Deliverables:**
- ✅ `k8s/Dockerfile` (K8s-specific, only cdk-erigon binary)
- ✅ Build scripts (build-image.sh, image-load.sh, validate-build.sh)
- ✅ Container requirements documentation
- ✅ Build workflow documentation
- ✅ Phase 1 design document
- ✅ Build optimization (99.6% context reduction, 94% faster, 88% smaller image)
- ✅ Cluster type auto-detection (OrbStack/kind/k3d/minikube)
- ✅ PreStop hook specification (RD-509)
- ✅ Shutdown sequence documentation (RD-512)
- ✅ Health probe configurations for Helm (RD-516)
- ✅ Security context configuration review (RD-505)

**Design Doc:** [Phase 1 Documentation](docs/phases/phase-1/)

---

### Phase 2: Helm Chart Structure (Weeks 3-4)
**Status:** Not Started
**Hours:** 32 hours
**Depends On:** Phase 1

**Key Deliverables:**
- Helm chart skeleton
- Values schema
- Basic templates (StatefulSet, Service, ConfigMap)
- Configuration management

---

### Phase 3: Core Components (Weeks 5-8)
**Status:** Not Started
**Hours:** 64 hours (CRITICAL PATH)
**Depends On:** Phase 1 (container), Phase 2 (Helm structure)

**Key Deliverables:**
- Sequencer StatefulSet with embedded NATS
- RPC Node StatefulSets
- Service definitions
- Volume management
- NATS connectivity between components
- Health probes validation (readiness, liveness, startup)
- Graceful shutdown testing (SIGTERM, NATS flush, timeout validation)
- Resilience testing (reconnection, recovery)

---

### Phase 4: Orchestration & Automation (Weeks 9-10)
**Status:** Not Started
**Hours:** 32 hours
**Depends On:** Phase 3

**Key Deliverables:**
- Deployment scripts
- Local cluster setup automation
- Port forwarding utilities
- Log aggregation
- Common operations playbook

---

### Phase 5: Testing & Validation (Week 11)
**Status:** Not Started
**Hours:** 16 hours
**Depends On:** Phase 4

**Key Deliverables:**
- Integration test suite
- Graceful shutdown tests
- NATS persistence validation
- Performance benchmarks
- Failure scenario testing

---

### Phase 6: CI/CD & Documentation (Week 12)
**Status:** Not Started
**Hours:** 16 hours
**Depends On:** Phase 5

**Key Deliverables:**
- GitHub Actions workflow
- Migration guide from Kurtosis
- Complete user documentation
- Troubleshooting guide
- Handoff to ops team

---

## Key Technical Decisions

### Container Image
**Decision:** Use separate `k8s/Dockerfile` instead of modifying root `Dockerfile`
**Rationale:**
- Kurtosis uses root `Dockerfile` (see `.github/actions/setup-kurtosis/action.yml:61`)
- Docker Compose uses root `Dockerfile` (see `docker-compose.yml:28-32`)
- Breaking these would disrupt existing workflows
- Allows parallel development and testing
- Can merge later when K8s proven

**Build Command:**
```bash
docker build -t cdk-erigon:k8s-${VERSION} -f k8s/Dockerfile .
```

---

### NATS Architecture
**Decision:** Embedded NATS JetStream in sequencer pod
**Rationale:**
- Simplifies deployment (no separate NATS cluster)
- Matches existing TCP datastream architecture
- RPC nodes connect to sequencer's NATS port (4222)
- JetStream persistence ensures zero data loss

**Ports:**
- 4222: NATS client connections
- 8222: NATS monitoring (optional)
- 6222: NATS cluster port (future multi-node)

---

### Graceful Shutdown
**Decision:** 60s termination grace period with PreStop hook
**Rationale:**
- PreStop hook: 5s connection draining
- SIGTERM: Allow 10-15s for NATS flush
- Total budget: 60s (K8s default: 30s insufficient)
- Critical for zero data loss guarantee

**Sequence:**
1. PreStop hook (t=0s): 5s connection drain
2. SIGTERM (t=5s): Signal to shutdown
3. NATS flush (t=5-15s): Persist all messages
4. Process exit (t=15-20s): Clean termination
5. SIGKILL (t=60s): Only if process hung

---

### Health Probes
**Decision:** HTTP probes on port 8545
**Rationale:**
- Existing RPC endpoint, no new code needed
- Readiness: Checks if node ready for traffic
- Liveness: Checks if process alive (not hung)
- Startup: Allows 5min for initial sync

**Configuration:**
```yaml
readinessProbe:
  httpGet: { path: /, port: 8545 }
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5

livenessProbe:
  httpGet: { path: /, port: 8545 }
  initialDelaySeconds: 60
  periodSeconds: 30
  timeoutSeconds: 10

startupProbe:
  httpGet: { path: /, port: 8545 }
  initialDelaySeconds: 10
  periodSeconds: 10
  failureThreshold: 30  # 5 min total
```

---

### Local Development
**Decision:** Support kind, k3d, and minikube
**Rationale:**
- Different teams use different tools
- Auto-detect cluster type in scripts
- Consistent experience across platforms

**Scripts:**
```bash
./k8s/scripts/build-image.sh v1.0.0
./k8s/scripts/image-load.sh cdk-erigon:k8s-v1.0.0
./k8s/scripts/deploy.sh
```

---

## Success Criteria

### Phase 1 Exit Criteria
- ✓ Container builds successfully
- ✓ Image loads into kind/k3d/minikube
- ✓ All required ports exposed
- ✓ Build scripts working and documented
- ✓ Container requirements documented
- ✓ Graceful shutdown designed
- ✓ Security review complete

### Project Exit Criteria
- One-command deployment on local K8s
- Zero message loss during pod restarts
- CI/CD pipeline functional
- Documentation complete
- Migration guide published
- Kurtosis users can migrate

---

## Dependencies & Blockers

### External Dependencies
- CDK Erigon embedded NATS support (exists, needs validation)
- NATS JetStream file storage (exists)
- HTTP health endpoint on 8545 (exists)

### Phase Dependencies
```
Phase 1 (Container) ──┬──> Phase 2 (Helm Structure) ──> Phase 3 (Components)
                      │
                      └──> Phase 3 (Components) ──> Phase 4 (Automation)

Phase 4 ──> Phase 5 (Testing) ──> Phase 6 (CI/CD)
```

**Critical Path:** Phase 1 → Phase 3 → Phase 4

---

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| NATS embedded not fully functional | HIGH | MEDIUM | Validate early in Phase 1, create fallback |
| Shutdown timing insufficient | HIGH | MEDIUM | Test with production workloads, make configurable |
| Kurtosis users resist migration | MEDIUM | MEDIUM | Clear migration guide, side-by-side support |
| Volume permissions issues | MEDIUM | MEDIUM | Test with various storage classes |
| Build scripts platform-specific | LOW | MEDIUM | Test on macOS, Linux, WSL early |

---

## Open Questions

1. **NATS Cluster:** Future multi-sequencer support needed?
2. **Monitoring:** Include Prometheus/Grafana in Helm chart or separate?
3. **Secrets Management:** How to handle keys in K8s (Sealed Secrets, External Secrets)?
4. **Network Policies:** Enforce at K8s level or application level?
5. **Multi-tenancy:** Single chart or separate charts per deployment?

---

## Resources

### Reference Documentation
- [Existing k8s/base manifests](base/) - Kustomize reference
- [Kurtosis setup action](.github/actions/setup-kurtosis/action.yml) - Current deployment
- [Docker Compose example](../docker-compose.yml) - Service architecture
- [NATS datastream tests](../zk/datastream/natsstream/) - NATS implementation

### External Links
- [NATS JetStream Docs](https://docs.nats.io/nats-concepts/jetstream)
- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)
- [K8s StatefulSet Basics](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
- [Graceful Shutdown](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/)

---

## Change Log

| Date       | Phase | Change | Rationale |
|------------|-------|--------|-----------|
| 2025-11-01 | All | **Sequential documentation structure** - Added Rule #7 | Every change documented sequentially in phase folders for PR descriptions and implementation docs |
| 2025-11-01 | Phase 1 | **Build optimization** - Reduced build context 76GB → 4GB | Excluded datadir (68GB), .git (2.6GB), local configs; 62% faster builds |
| 2025-11-01 | Phase 1 | **Binary build scope** - Build only cdk-erigon | Changed from 17 binaries to 1; 94% faster builds, 88% smaller image |
| 2025-11-01 | Phase 1 | **Phase 1 Complete** - Container foundation delivered | All deliverables complete: Dockerfile, build scripts, documentation, validation |
| 2025-11-01 | Phase 1 | Created k8s/Dockerfile instead of modifying root | Preserve Kurtosis/docker-compose compatibility |
| 2025-11-01 | Phase 1 | All K8s work in /k8s directory | Isolation and organization |
| 2025-11-01 | Phase 1 | YAML ConfigMaps chosen as primary config method | Familiar to users, K8s-native pattern |
| 2025-11-01 | Phase 1 | Automated validation with smoke tests | Ensure quality, enable CI/CD integration |
| 2025-11-01 | All | Created PROJECT.md | Central documentation and rules |
| 2025-11-01 | All | Added privacy rule: No personal contact info in docs | Keep contributions anonymous |

---

*Last Updated: 2025-11-01*