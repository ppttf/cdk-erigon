# Monitoring Resources

This directory contains monitoring resources for observability integration.

## Prerequisites

### Prometheus Operator

ServiceMonitor resources require [Prometheus Operator](https://prometheus-operator.dev/) to be installed in the cluster:

```bash
# Using kube-prometheus-stack Helm chart
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack
```

## Resources

### ServiceMonitor - Sequencer

`servicemonitor-sequencer.yaml` - Prometheus scraping configuration for cdk-erigon sequencer metrics.

**Requires**:
- `monitoring.enabled: true`
- `monitoring.serviceMonitor.enabled: true`
- Prometheus Operator CRDs installed

**Metrics exposed**:
- Erigon diagnostics and performance metrics
- Block processing metrics
- Database metrics
- RPC statistics

**Endpoint**: `http://<sequencer-pod>:6060/debug/metrics/prometheus`

### ServiceMonitor - RPC

`servicemonitor-rpc.yaml` - Prometheus scraping configuration for cdk-erigon RPC node metrics.

**Requires**:
- `monitoring.enabled: true`
- `monitoring.serviceMonitor.enabled: true`
- `rpc.enabled: true`
- Prometheus Operator CRDs installed

**Metrics exposed**:
- Erigon diagnostics and performance metrics
- RPC request statistics
- Database metrics

**Endpoint**: `http://<rpc-pod>:6060/debug/metrics/prometheus`

### ServiceMonitor - NATS

`servicemonitor-nats.yaml` - Prometheus scraping configuration for NATS JetStream metrics.

**Requires**:
- `sequencer.nats.monitoring.enabled: true`
- Prometheus Operator CRDs installed

**Metrics exposed**:
- `gnatsd_varz_*` - Core NATS server metrics
- `jetstream_*` - JetStream specific metrics (streams, consumers, storage)

**Endpoint**: `http://<sequencer-pod>:8222/metrics`

### Grafana Dashboards

`grafana-dashboards.yaml` - ConfigMap containing pre-built Grafana dashboards for NATS and cdk-erigon.

**Requires**:
- `monitoring.enabled: true`
- `monitoring.grafana.dashboards.enabled: true`
- Grafana installed with dashboard provisioning configured

**Dashboards included**:
- **NATS JetStream**: Stream metrics, storage usage, consumer lag, message throughput
- **cdk-erigon Performance**: Block sync progress, memory/CPU usage, RPC request rate, database size

**Source files**:
- `dashboards/nats-jetstream.json`
- `dashboards/cdk-erigon-performance.json`

## Usage

### Global Monitoring

```yaml
# values.yaml or values-bali.yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s
```

### NATS Monitoring

```yaml
sequencer:
  nats:
    monitoring:
      enabled: true
      port: 8222
```

### Grafana Dashboards

```yaml
monitoring:
  enabled: true
  grafana:
    dashboards:
      enabled: true
```

**Grafana Configuration**:

The dashboards ConfigMap uses the label `grafana_dashboard: "1"` for automatic discovery by Grafana's sidecar. Configure Grafana to watch for this label:

```yaml
# Grafana Helm values
sidecar:
  dashboards:
    enabled: true
    label: grafana_dashboard
    labelValue: "1"
```

**Manual Import**:

If not using automatic provisioning, dashboards can be manually imported:

```bash
# Export dashboard JSON
kubectl get configmap <release>-cdk-erigon-grafana-dashboards -o jsonpath='{.data.nats-jetstream\.json}' > nats-jetstream.json
kubectl get configmap <release>-cdk-erigon-grafana-dashboards -o jsonpath='{.data.cdk-erigon-performance\.json}' > cdk-erigon-performance.json

# Import via Grafana UI: Create > Import > Upload JSON file
```

## Testing

Without Prometheus Operator installed, you can still verify metrics are exposed:

```bash
# Sequencer erigon metrics
kubectl port-forward statefulset/cdk-erigon-sequencer 6060:6060
curl http://localhost:6060/debug/metrics/prometheus

# RPC erigon metrics
kubectl port-forward statefulset/cdk-erigon-rpc 6060:6060
curl http://localhost:6060/debug/metrics/prometheus

# NATS metrics
kubectl port-forward statefulset/cdk-erigon-sequencer 8222:8222
curl http://localhost:8222/metrics
```
