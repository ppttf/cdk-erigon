# CDK Erigon Container Requirements

**Target:** Kubernetes deployment via Helm chart (Phase 2+)
**Image:** `cdk-erigon:k8s-<version>`
**Build:** `k8s/Dockerfile`

---

## Container Image

### Base Information
- **User:** `erigon` (UID: 1000, GID: 1000)
- **Working Directory:** `/home/erigon`
- **Entrypoint:** `cdk-erigon`
- **Base OS:** Alpine Linux

### Binaries Included
- `cdk-erigon` - Main blockchain client (includes embedded RPC server and NATS support)

---

## Port Mappings

### RPC & API Ports
| Port | Protocol | Purpose | Required | Notes |
|------|----------|---------|----------|-------|
| 8545 | TCP | HTTP RPC | Yes | Main JSON-RPC endpoint, health probe target |
| 8551 | TCP | Engine API | Yes | Consensus layer communication |
| 8546 | TCP | WebSocket RPC | Optional | WebSocket endpoint |
| 9090 | TCP | Private API | Yes | Internal API for components |
| 8080 | TCP | HTTP API | Optional | Alternative HTTP endpoint |

### P2P & Networking
| Port | Protocol | Purpose | Required | Notes |
|------|----------|---------|----------|-------|
| 30303 | TCP | ETH P2P | Yes | Ethereum protocol peer connections |
| 30303 | UDP | ETH Discovery | Yes | Node discovery |
| 42069 | TCP | Snap Sync | Yes | Snapshot synchronization |
| 42069 | UDP | Snap Sync | Yes | Snapshot synchronization |

### NATS Datastream (K8s-specific)
| Port | Protocol | Purpose | Required | Notes |
|------|----------|---------|----------|-------|
| 4222 | TCP | NATS Client | Yes | NATS client connections (embedded server) |
| 8222 | TCP | NATS Monitoring | Optional | NATS HTTP monitoring endpoint |
| 6222 | TCP | NATS Cluster | Future | For multi-node NATS clustering |

### Monitoring & Debugging
| Port | Protocol | Purpose | Required | Notes |
|------|----------|---------|----------|-------|
| 6060 | TCP | Metrics | Yes | Prometheus metrics endpoint |
| 6070 | TCP | Pprof | Optional | Go profiling endpoint |
| 7900 | TCP | Data Stream (TCP) | Optional | Legacy TCP datastream (NATS preferred) |

---

## Configuration

### Method
**Primary:** YAML configuration files mounted via ConfigMap
**Alternative:** Environment variables converted to CLI flags

### Configuration File Locations
```
/home/erigon/mainnet.yaml   - Mainnet config
/home/erigon/cardona.yaml   - Cardona testnet config
/home/erigon/bali.yaml      - Bali testnet config
/home/erigon/config.yaml    - Custom config (mounted via ConfigMap)
```

### Key Configuration Parameters

**Chain Configuration:**
```yaml
chain: hermez-bali              # Chain to run
datadir: ~/.local/share/erigon  # Data directory
```

**RPC Configuration:**
```yaml
http: true
http.addr: 0.0.0.0
http.port: 8545
http.api: [eth, debug, net, trace, web3, erigon, zkevm]
http.vhosts: any
http.corsdomain: any
private.api.addr: 0.0.0.0:9090
```

**NATS Configuration (Sequencer):**
```yaml
zkevm.data-stream-nats-host: 0.0.0.0    # Embedded NATS server bind address
zkevm.data-stream-nats-port: 4222        # NATS client port
```

**NATS Configuration (RPC Node):**
```yaml
zkevm.l2-nats-seed-host: sequencer-service  # NATS server hostname
zkevm.l2-nats-seed-port: 4222              # NATS server port
```

**L1 Integration:**
```yaml
zkevm.l1-chain-id: 11155111               # L1 chain ID (Sepolia)
zkevm.l1-rpc-url: https://...             # L1 RPC endpoint
zkevm.l2-chain-id: 2440                   # L2 chain ID
zkevm.l2-sequencer-rpc-url: https://...   # Sequencer RPC
```

### Environment Variable Overrides
For dynamic K8s configuration, key settings can be overridden via environment variables:

```bash
ERIGON_DATADIR=/home/erigon/.local/share/erigon
ERIGON_CHAIN=hermez-bali
NATS_HOST=0.0.0.0
NATS_PORT=4222
L1_RPC_URL=https://...
```

---

## Volume Requirements

### Data Volume
**Mount Path:** `/home/erigon/.local/share/erigon`
**Purpose:** Blockchain data, state, and NATS JetStream storage
**Type:** PersistentVolumeClaim
**Access Mode:** ReadWriteOnce
**Recommended Size:** 100Gi+ (varies by chain)

**Subdirectories:**
```
/home/erigon/.local/share/erigon/
├── chaindata/          # Blockchain database
├── nodes/              # Node data
├── nats/               # NATS JetStream storage (if embedded)
└── datastream/         # Datastream files
```

**Requirements:**
- Must be writable by UID 1000
- Should support fast I/O (SSD recommended)
- Must persist across pod restarts

### ConfigMap Volume (Optional)
**Mount Path:** `/home/erigon/config.yaml`
**Purpose:** Custom configuration
**Type:** ConfigMap
**Access Mode:** ReadOnly

---

## Security Context

### Pod Security Context
```yaml
securityContext:
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
  runAsNonRoot: true
```

### Container Security Context
```yaml
securityContext:
  runAsUser: 1000
  runAsGroup: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
      - NET_RAW
```

**Notes:**
- Root filesystem is read-only
- Data volume must be writable
- No privileged access required

---

## Health Probes

### Readiness Probe
Indicates when container is ready to serve traffic

```yaml
readinessProbe:
  httpGet:
    path: /
    port: 8545
    scheme: HTTP
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  successThreshold: 1
  failureThreshold: 3
```

### Liveness Probe
Detects if container is hung and needs restart

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 8545
    scheme: HTTP
  initialDelaySeconds: 60
  periodSeconds: 30
  timeoutSeconds: 10
  successThreshold: 1
  failureThreshold: 3
```

### Startup Probe
Allows extended time for initial sync

```yaml
startupProbe:
  httpGet:
    path: /
    port: 8545
    scheme: HTTP
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  successThreshold: 1
  failureThreshold: 30  # 5 minutes total
```

---

## Lifecycle Management

### Graceful Shutdown
**Grace Period:** 60 seconds (adjust via `terminationGracePeriodSeconds`)

**PreStop Hook:**
```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 5"]
```

**Shutdown Sequence:**
1. PreStop hook executes (5s drain)
2. SIGTERM sent to process
3. NATS flushes pending messages
4. Process exits gracefully
5. SIGKILL after grace period (if needed)

---

## Resource Requirements

### Recommended Requests/Limits

**Sequencer:**
```yaml
resources:
  requests:
    cpu: "4"
    memory: "16Gi"
  limits:
    cpu: "8"
    memory: "32Gi"
```

**RPC Node:**
```yaml
resources:
  requests:
    cpu: "2"
    memory: "8Gi"
  limits:
    cpu: "4"
    memory: "16Gi"
```

**Notes:**
- Actual requirements vary by chain and workload
- Monitor and adjust based on metrics
- Consider node size when setting limits

---

## Node Role Configuration

### Sequencer Mode
```yaml
# Config that enables sequencer mode
zkevm.sequencer: true
zkevm.data-stream-nats-host: 0.0.0.0
zkevm.data-stream-nats-port: 4222
```

### RPC Node Mode
```yaml
# Config that enables RPC node mode (connects to sequencer)
zkevm.l2-nats-seed-host: sequencer-0.sequencer.default.svc.cluster.local
zkevm.l2-nats-seed-port: 4222
```

---

## Labels & Annotations

### Required Labels
```yaml
app.kubernetes.io/name: cdk-erigon
app.kubernetes.io/instance: <release-name>
app.kubernetes.io/version: <version>
app.kubernetes.io/component: <sequencer|rpc-node>
app.kubernetes.io/part-of: cdk-erigon-chain
app.kubernetes.io/managed-by: Helm
```

### Recommended Annotations
```yaml
prometheus.io/scrape: "true"
prometheus.io/port: "6060"
prometheus.io/path: "/debug/metrics/prometheus"
```

---

## Networking

### Service Requirements

**Sequencer Service:**
- Type: ClusterIP (internal) or LoadBalancer (if external access needed)
- Ports: 8545 (RPC), 4222 (NATS), 6060 (metrics)
- Headless service for StatefulSet pod discovery

**RPC Node Service:**
- Type: ClusterIP or LoadBalancer
- Ports: 8545 (RPC), 6060 (metrics)

### Service Mesh Compatibility
- Compatible with Istio, Linkerd
- mTLS supported
- No special sidecar requirements

---

## Image Build Information

### Build Arguments
```dockerfile
BUILD_DATE    # ISO 8601 timestamp
VCS_REF       # Git commit SHA
VERSION       # Version tag
UID           # User ID (default: 1000)
GID           # Group ID (default: 1000)
```

### Labels Present
- OCI standard labels (`org.opencontainers.image.*`)
- Kubernetes labels (`io.kubernetes.*`)
- Legacy labels (`org.label-schema.*`)

### Build Command
```bash
cd /path/to/cdk-erigon
./k8s/scripts/build-image.sh v1.0.0
```

---

## Testing & Validation

### Validation Script
```bash
./k8s/scripts/validate-build.sh cdk-erigon:k8s-v1.0.0
```

**Checks:**
- All required ports exposed
- Labels present
- Non-root user
- Entrypoint correct
- Runtime smoke test (optional)

---

## Next Steps

This document provides specifications for **Phase 2: Helm Chart Development**.

Use these requirements to:
1. Design Helm chart `values.yaml` schema
2. Create StatefulSet and Service templates
3. Configure ConfigMap generation
4. Set up proper RBAC if needed

---

*Last Updated: 2025-11-05 | Phase 1: Container Foundation*
