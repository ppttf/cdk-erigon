# CDK-Erigon Helm Chart

Deploy CDK-Erigon blockchain nodes with embedded NATS JetStream on Kubernetes.

## Quick Start

### Prerequisites

- Kubernetes 1.25+
- Helm 3.x
- `kubectl` configured for your cluster

### Deploy with Monitoring (Recommended)

```bash
# Add chart dependencies
helm dependency update

# Deploy with full observability stack
helm install cdk-erigon . \
  -f values-bali.yaml \
  --set monitoring.enabled=true \
  --set monitoring.prometheus.enabled=true
```

This automatically:
- Installs Prometheus Operator CRDs (idempotent, works on fresh or existing clusters)
- Deploys Prometheus + Grafana
- Configures ServiceMonitors for all components
- Provisions NATS and cdk-erigon Grafana dashboards

### Deploy without Monitoring

```bash
helm dependency update
helm install cdk-erigon . -f values-bali.yaml
```

### Access Services

```bash
# Grafana (admin/admin)
kubectl port-forward svc/cdk-erigon-prometheus-grafana 3000:80
# Open http://localhost:3000

# Sequencer RPC
kubectl port-forward svc/cdk-erigon-sequencer 8545:8545

# RPC Node
kubectl port-forward svc/cdk-erigon-rpc 8545:8545
```

## Architecture

```
                    ┌─────────────────────────────────┐
                    │         Grafana                 │
                    │  ┌───────────┬────────────────┐ │
                    │  │   NATS    │   cdk-erigon   │ │
                    │  │ JetStream │  Performance   │ │
                    │  └───────────┴────────────────┘ │
                    └────────────────┬────────────────┘
                                     │
                              ┌──────┴──────┐
                              │  Prometheus │
                              └──────┬──────┘
                                     │
            ┌────────────────────────┼────────────────────────┐
            │                        │                        │
     ┌──────▼──────┐          ┌──────▼──────┐          ┌──────▼──────┐
     │  Sequencer  │          │     RPC     │          │  L1 Proxy   │
     │ StatefulSet │          │ StatefulSet │          │ Deployment  │
     │             │          │             │          │             │
     │ + NATS      │◄─────────│  datastream │          │   cache     │
     │   :4222     │          │             │          │             │
     │   :6900     │          │             │          │             │
     └─────────────┘          └─────────────┘          └─────────────┘
           │                                                  │
           └────────────────────────┬─────────────────────────┘
                                    │
                              ┌─────▼─────┐
                              │  L1 RPC   │
                              │ (Sepolia) │
                              └───────────┘
```

### Components

| Component | Type | Description |
|-----------|------|-------------|
| Sequencer | StatefulSet | Block producer with embedded NATS JetStream |
| RPC | StatefulSet | Read-only nodes syncing via NATS datastream |
| L1 Proxy | Deployment | Shared L1 RPC cache (optional) |
| Prometheus | Deployment | Metrics collection (optional) |
| Grafana | Deployment | Dashboards and visualization (optional) |

## Configuration

### Values Files

| File | Purpose |
|------|---------|
| `values.yaml` | Default configuration |
| `values-bali.yaml` | Bali testnet chain config |
| `values-testing.yaml` | Testing/CI configuration |

### Key Configuration Options

```yaml
# Enable/disable components
sequencer:
  enabled: true
  replicas: 1  # Must be 1 for embedded NATS

rpc:
  enabled: true
  replicas: 2  # Scale horizontally

l1Proxy:
  enabled: false  # Enable for L1 caching

# Monitoring stack
monitoring:
  enabled: true
  prometheus:
    enabled: true  # Deploy Prometheus + Grafana
  grafana:
    dashboards:
      enabled: true  # Provision dashboards

# NATS configuration
sequencer:
  nats:
    embedded: true
    monitoring:
      enabled: true
      port: 8222
```

### L1 RPC Configuration

Create a local values file for sensitive configuration:

```yaml
# values-local.yaml (git-ignored)
sequencer:
  config:
    l1RpcUrl: "https://your-l1-rpc-endpoint.com"
    ethermanApiKey: "your-api-key"

rpc:
  config:
    l1RpcUrl: "https://your-l1-rpc-endpoint.com"
```

Deploy with:
```bash
helm install cdk-erigon . -f values-bali.yaml -f values-local.yaml
```

## Monitoring

### Integrated Stack

When `monitoring.prometheus.enabled: true`, the chart deploys:

- **Prometheus Operator** - CRD-based Prometheus management
- **Prometheus** - Metrics collection with 7-day retention
- **Grafana** - Visualization with auto-provisioned dashboards
- **ServiceMonitors** - Automatic scrape target discovery

CRDs are automatically installed via Helm pre-install hook (idempotent).

### External Prometheus

To use an existing Prometheus Operator installation:

```bash
helm install cdk-erigon . \
  -f values-bali.yaml \
  --set monitoring.enabled=true \
  --set monitoring.prometheus.enabled=false
```

This creates ServiceMonitors that your existing Prometheus will discover.

### Metrics Endpoints

| Component | Port | Path | Metrics |
|-----------|------|------|---------|
| Sequencer | 6060 | `/debug/metrics/prometheus` | Erigon performance |
| RPC | 6060 | `/debug/metrics/prometheus` | Erigon performance |
| NATS | 8222 | `/metrics` | JetStream stats |

### Grafana Dashboards

Pre-configured dashboards:

- **NATS JetStream** - Stream metrics, consumer lag, storage usage
- **cdk-erigon Performance** - Block sync, memory/CPU, RPC rates

Dashboards are auto-provisioned via Grafana sidecar.

### Manual Metrics Verification

```bash
# Sequencer erigon metrics
kubectl port-forward statefulset/cdk-erigon-sequencer 6060:6060
curl http://localhost:6060/debug/metrics/prometheus

# NATS metrics
kubectl port-forward statefulset/cdk-erigon-sequencer 8222:8222
curl http://localhost:8222/metrics
```

## Troubleshooting

### Pod Won't Start

```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/name=cdk-erigon

# Check events
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name> -f
```

### Sequencer OOMKilled

Increase memory limits:

```yaml
sequencer:
  resources:
    limits:
      memory: 32Gi
```

### RPC Can't Connect to NATS

Verify NATS is running and DNS is correct:

```bash
# Check NATS is listening
kubectl exec -it cdk-erigon-sequencer-0 -- nc -zv localhost 4222

# Check DNS resolution from RPC pod
kubectl exec -it cdk-erigon-rpc-0 -- nslookup cdk-erigon-sequencer
```

### ServiceMonitor Not Working

Verify Prometheus Operator CRDs are installed:

```bash
kubectl get crd servicemonitors.monitoring.coreos.com
```

If missing, ensure `prometheus.crds.upgradeJob.enabled: true` in values.

### Fresh Cluster Deployment Fails

Ensure dependencies are updated:

```bash
helm dependency update
helm install cdk-erigon . -f values-bali.yaml --set monitoring.prometheus.enabled=true
```

## Development

### Run Tests

```bash
# Install helm-unittest plugin
helm plugin install https://github.com/helm-unittest/helm-unittest

# Run all tests
helm unittest .

# Run specific test
helm unittest -f 'tests/sequencer_*_test.yaml' .
```

### Validate Templates

```bash
# Lint chart
helm lint .

# Render templates
helm template test . -f values-bali.yaml

# Validate with kubeconform
helm template test . | kubeconform -strict
```

## Dependencies

| Chart | Version | Purpose |
|-------|---------|---------|
| kube-prometheus-stack | ~80.0.0 | Prometheus + Grafana (optional) |

## License

See repository LICENSE file.
