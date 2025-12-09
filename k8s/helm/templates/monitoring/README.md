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
