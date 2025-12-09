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

```yaml
# values.yaml or values-bali.yaml
sequencer:
  nats:
    monitoring:
      enabled: true
      port: 8222
```

## Testing

Without Prometheus Operator installed, you can still verify metrics are exposed:

```bash
# Port-forward to sequencer
kubectl port-forward statefulset/cdk-erigon-sequencer 8222:8222

# Query metrics
curl http://localhost:8222/metrics
```
