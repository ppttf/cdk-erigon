# syntax = docker/dockerfile:1.2
# Kubernetes-specific Dockerfile for CDK Erigon
# Based on root Dockerfile with K8s enhancements

FROM docker.io/library/golang:1.25-alpine AS builder

RUN apk --no-cache add build-base linux-headers git bash ca-certificates libstdc++

WORKDIR /app
ADD go.mod go.mod
ADD go.sum go.sum
ADD erigon-lib/go.mod erigon-lib/go.mod
ADD erigon-lib/go.sum erigon-lib/go.sum

RUN go mod download
ADD . .

RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/tmp/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    make BUILD_TAGS=nosqlite,noboltdb,nosilkworm cdk-erigon


FROM docker.io/library/alpine:latest

# Install required runtime libs, along with some helpers for debugging
RUN apk add --no-cache ca-certificates libstdc++ tzdata
RUN apk add --no-cache curl jq bind-tools

# Setup user and group
ARG UID=1000
ARG GID=1000
RUN adduser -D -u $UID -g $GID erigon
USER erigon
RUN mkdir -p ~/.local/share/erigon
WORKDIR /home/erigon

# Add pre-defined run configs
ADD hermezconfig-mainnet.yaml.example mainnet.yaml
ADD hermezconfig-cardona.yaml.example cardona.yaml
ADD hermezconfig-bali.yaml.example bali.yaml

# Copy compiled artifacts from builder
COPY --from=builder /app/build/bin/cdk-erigon /usr/local/bin/cdk-erigon

# Expose ports
# RPC and engine API
EXPOSE 8545 \
       8551 \
       8546 \
       30303 \
       30303/udp \
       42069 \
       42069/udp \
       8080 \
       9090 \
       6060 \
       7900

# NATS datastream ports (K8s-specific)
EXPOSE 4222 \
       8222 \
       6222

# Metadata labels
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

# OCI standard labels
LABEL org.opencontainers.image.title="CDK Erigon" \
      org.opencontainers.image.description="CDK Erigon with embedded NATS JetStream for Kubernetes" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/0xPolygonHermez/cdk-erigon" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.vendor="Gateway.fm" \
      org.opencontainers.image.licenses="LGPL-3.0" \
      org.opencontainers.image.url="https://github.com/0xPolygonHermez/cdk-erigon"

# Kubernetes-specific labels
LABEL io.kubernetes.component="blockchain-node" \
      io.kubernetes.part-of="cdk-erigon-chain" \
      io.kubernetes.managed-by="helm"

# Legacy labels (for compatibility)
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.description="CDK Erigon ZKEVM Client for Kubernetes" \
      org.label-schema.name="CDK Erigon" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/0xPolygonHermez/cdk-erigon.git" \
      org.label-schema.vendor="Gateway.fm" \
      org.label-schema.version=$VERSION

ENTRYPOINT ["cdk-erigon"]
