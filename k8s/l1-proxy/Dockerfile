# Build stage
FROM docker.io/library/golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /build

COPY go.mod go.sum ./
COPY erigon-lib/go.mod erigon-lib/go.sum ./erigon-lib/
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /rpc-cache ./cmd/hack/rpc_cache/main.go

# Final stage
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /rpc-cache /rpc-cache

EXPOSE 6969

ENTRYPOINT ["/rpc-cache"]
