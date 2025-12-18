# Chart Versioning Strategy

## Overview

This chart follows semantic versioning for the chart itself, while tracking the CDK Erigon application version separately.

## Version Components

### Chart Version (Chart.yaml: `version`)
Follows [Semantic Versioning 2.0.0](https://semver.org/):
- **MAJOR**: Breaking changes to chart API (values structure, template changes that require migration)
- **MINOR**: New features, backward-compatible additions
- **PATCH**: Bug fixes, documentation updates

### App Version (Chart.yaml: `appVersion`)
Tracks the CDK Erigon application version deployed by this chart.

## Version Examples

```yaml
# Initial release
version: 0.1.0
appVersion: "1.0.0"

# Added new component template (backward compatible)
version: 0.2.0
appVersion: "1.0.0"

# Updated app version (chart unchanged)
version: 0.2.1
appVersion: "1.0.1"

# Breaking change to values structure
version: 1.0.0
appVersion: "2.0.0"
```

## Breaking Changes

Breaking changes require MAJOR version bump and include:

- Renaming or removing values fields
- Changing default behavior that affects existing deployments
- Requiring new Kubernetes API versions
- Removing templates or resources

**Breaking changes must include:**
- Migration guide in CHANGELOG
- Deprecation warnings in previous minor versions when possible
- Clear documentation of changes

## Upgrade Policy

### Minor/Patch Upgrades
Should work with `helm upgrade`:
```bash
helm upgrade cdk-erigon k8s/helm
```

### Major Upgrades
May require manual intervention:
```bash
# Review breaking changes
helm get values cdk-erigon > current-values.yaml
# Migrate values to new structure
# Then upgrade
helm upgrade cdk-erigon k8s/helm -f migrated-values.yaml
```

## Pre-Release Versions

For development and testing:
```yaml
# Alpha release
version: 0.3.0-alpha.1

# Beta release
version: 0.3.0-beta.1

# Release candidate
version: 0.3.0-rc.1
```

## Release Process

1. Update Chart.yaml with new version
2. Update CHANGELOG.md with changes
3. Tag release: `git tag helm-v0.2.0`
4. Push: `git push --tags`

## Deprecation Policy

Features marked for removal:
1. Announce in CHANGELOG with deprecation warning
2. Keep deprecated feature for at least one MINOR version
3. Remove in next MAJOR version
4. Provide migration path in documentation