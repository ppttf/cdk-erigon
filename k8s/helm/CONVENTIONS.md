# Template Naming Conventions

## File Naming

### Template Files
Format: `<resource-type>.yaml`

Examples:
- `statefulset.yaml`
- `service.yaml`
- `configmap.yaml`
- `pvc.yaml`

### Directory Structure
Organize by component:
```
templates/
├── _helpers.tpl           # Template helper functions
├── NOTES.txt              # Post-install notes
├── sequencer/             # Sequencer component
│   ├── statefulset.yaml
│   ├── service.yaml
│   └── configmap.yaml
├── rpc/                   # RPC component
│   ├── statefulset.yaml
│   ├── service.yaml
│   └── configmap.yaml
└── monitoring/            # Monitoring components
    ├── prometheus-statefulset.yaml
    └── grafana-statefulset.yaml
```

## Resource Naming

### Kubernetes Resources
Pattern: `{{ include "cdk-erigon.fullname" . }}-<component>`

Examples:
```yaml
# Sequencer StatefulSet
name: {{ include "cdk-erigon.fullname" . }}-sequencer

# RPC Service
name: {{ include "cdk-erigon.fullname" . }}-rpc

# NATS ConfigMap
name: {{ include "cdk-erigon.fullname" . }}-nats-config
```

### Multi-Instance Resources
For resources with multiple instances, add descriptive suffix:
```yaml
# Prometheus StatefulSet
name: {{ include "cdk-erigon.fullname" . }}-prometheus

# Grafana Service
name: {{ include "cdk-erigon.fullname" . }}-grafana
```

## Labels

### Required Labels
Always use the standard helpers:
```yaml
metadata:
  labels:
    {{- include "cdk-erigon.labels" . | nindent 4 }}
    component: sequencer
```

### Component Labels
Add component-specific label for all resources:
```yaml
labels:
  component: sequencer   # or rpc, monitoring, etc.
```

### Selector Labels
Use the selector helper consistently:
```yaml
selector:
  matchLabels:
    {{- include "cdk-erigon.selectorLabels" . | nindent 6 }}
    component: sequencer
```

## Values Structure

### Hierarchical Organization
Organize by component:
```yaml
sequencer:
  enabled: true
  replicas: 1
  resources: {}
  nats:
    embedded: true

rpc:
  enabled: false
  replicas: 2
  resources: {}
```

### Naming Conventions
- Use camelCase for value names
- Group related settings
- Add comments for all values

```yaml
# -- Enable sequencer component
sequencer:
  # -- Number of sequencer replicas
  replicas: 1

  # -- NATS configuration
  nats:
    # -- Enable embedded NATS server
    embedded: true
```

## Conditionals

### Feature Flags
Use `.enabled` pattern consistently:
```yaml
{{- if .Values.sequencer.enabled }}
# ... sequencer resources
{{- end }}

{{- if .Values.rpc.enabled }}
# ... RPC resources
{{- end }}
```

### Nested Conditionals
Keep conditionals clear and minimal:
```yaml
{{- if .Values.sequencer.enabled }}
{{- if .Values.sequencer.nats.embedded }}
# NATS configuration
{{- end }}
{{- end }}
```

## Helper Functions

### Naming Pattern
Pattern: `"cdk-erigon.<purpose>"`

Examples:
```yaml
{{- define "cdk-erigon.name" -}}
{{- define "cdk-erigon.fullname" -}}
{{- define "cdk-erigon.labels" -}}
{{- define "cdk-erigon.selectorLabels" -}}
{{- define "cdk-erigon.chart" -}}
```

### Component-Specific Helpers
For complex component logic:
```yaml
{{- define "cdk-erigon.sequencer.image" -}}
{{- define "cdk-erigon.rpc.probes" -}}
```

## Comments

### Template Comments
Use Go-style template comments:
```yaml
{{/*
Create the name of the service account to use
*/}}
{{- define "cdk-erigon.serviceAccountName" -}}
...
{{- end }}
```

### YAML Comments
Document complex sections:
```yaml
# Security context following best practices
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
```

## Anti-Patterns

❌ **Avoid:**
- Hardcoded values in templates
- Inconsistent naming schemes
- Deep nesting (>3 levels)
- Generic names like `service.yaml` in root templates/

✅ **Prefer:**
- Values-driven configuration
- Consistent naming patterns
- Component-organized structure
- Descriptive, specific names