# CDK-Erigon Development Guide

This document provides comprehensive guidance for developers who want to contribute to or build on CDK-Erigon.

## Development Environment Setup

### Prerequisites

Before starting development work with CDK-Erigon, ensure you have the following prerequisites installed:

- **Go** (version 1.21 or later)
- **Git**
- **Make**
- **GCC** (for CGO compilation)
- **Docker** (optional, for containerized development)

### Setting Up Your Development Environment

1. **Clone the Repository**

```bash
git clone https://github.com/0xPolygon/cdk-erigon.git
cd cdk-erigon
```

2. **Install Dependencies**

```bash
make deps
```

3. **Build the Project**

```bash
make build
```

This will create binary files in the `./build/bin` directory.

### Development Workflow

CDK-Erigon follows a typical Go development workflow:

```
+---------------------+       +---------------------+       +---------------------+
|                     |       |                     |       |                     |
| Make Code Changes   | ----> | Run Tests           | ----> | Build & Run Locally |
|                     |       |                     |       |                     |
+---------------------+       +---------------------+       +---------------------+
```

## Project Structure

The CDK-Erigon codebase is organized as follows:

```
cdk-erigon/
├── cmd/                    # Command-line applications
│   ├── erigon/             # Main CDK-Erigon node application
│   ├── rpctest/            # RPC testing utilities
│   └── hack/               # Development utility commands
├── core/                   # Core domain logic
│   ├── types/              # Core data types
│   ├── rawdb/              # Database access layer
│   └── state/              # State management
├── ethdb/                  # Database interfaces and implementations
├── zkevm/                  # zkEVM-specific code
│   ├── circuit/            # Circuit implementation
│   ├── encoding/           # Data encoding utilities
│   ├── pool/               # Transaction pool
│   ├── rpc/                # RPC interfaces
│   ├── sequencer/          # Sequencer component
│   ├── synchronizer/       # L1/L2 sync components
│   └── witness/            # Witness generation  
├── build/                  # Build scripts and artifacts
│   └── bin/                # Binary output directory
├── tests/                  # Test suites
│   ├── state/              # State tests
│   ├── rpc/                # RPC tests
│   └── integration/        # Integration tests
└── examples/               # Example configurations and scripts
```

## Building and Testing

### Building CDK-Erigon

CDK-Erigon provides several build targets:

```bash
# Build all binaries
make build

# Build only the main node binary
make erigon

# Build with race detector enabled
make build-race

# Cross-compile for different platforms
make build-linux
make build-windows
make build-darwin
```

### Running Tests

CDK-Erigon has several test categories:

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Run tests with race detector
make test-race

# Run specific test package
go test -v ./zkevm/synchronizer/...

# Run specific test function
go test -v ./zkevm/synchronizer/... -run TestSyncFromL1
```

### Benchmarks

To measure performance:

```bash
# Run all benchmarks
make bench

# Run specific benchmark
go test -bench=BenchmarkBatchExecution -run=^$ ./zkevm/...
```

## Development Best Practices

### Code Style

CDK-Erigon follows standard Go code style conventions:

1. Run `gofmt` and `golint` before committing
2. Keep line length reasonable (around 100 chars)
3. Use meaningful variable and function names
4. Document public functions with proper godocs

```bash
# Format code
make fmt

# Check lint issues
make lint
```

### Git Workflow

Follow these guidelines for Git:

1. Create feature branches from `main`
2. Use descriptive branch names: `feature/new-sync-method` or `fix/memory-leak`
3. Write meaningful commit messages
4. Create pull requests with clear descriptions
5. Ensure CI passes before requesting review

### Writing Tests

Tests are critical in CDK-Erigon:

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test full system behavior

Test patterns to follow:

```go
func TestMyFunction(t *testing.T) {
    // Setup
    ctx := context.Background()
    
    // Define test cases
    testCases := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"Happy path", "valid input", "expected output", false},
        {"Error case", "invalid input", "", true},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Execute test
            result, err := MyFunction(ctx, tc.input)
            
            // Verify results
            if tc.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                require.Equal(t, tc.expected, result)
            }
        })
    }
}
```

## Debugging Tips

### Logging

CDK-Erigon uses a structured logging system:

```go
import "github.com/ledgerwatch/log/v3"

// Initialize logger
logger := log.New("component", "synchronizer")

// Log with context
logger.Info("Processing batch", "batchNum", batchNum, "txCount", txCount)

// Log errors
logger.Error("Failed to connect", "err", err, "endpoint", endpoint)
```

### Setting Log Levels

You can control log verbosity at runtime:

```bash
# Run with debug logging
./build/bin/erigon --verbosity=5

# Run with minimal logging
./build/bin/erigon --verbosity=2
```

Available log levels:
- 0: Silent
- 1: Error
- 2: Warn
- 3: Info
- 4: Debug
- 5: Trace

### Debugging Memory Issues

For memory-intensive operations:

1. Use the Go profiler:

```bash
# Run with profiling enabled
./build/bin/erigon --pprof --pprof.addr=localhost:6060

# Capture a heap profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

2. Enable debug mode for zkEVM operations:

```bash
./build/bin/erigon --zkEVM.debug-save-witness --zkEVM.witness-db-path=/tmp/witness
```

3. Tune memory parameters:

```bash
./build/bin/erigon --zkEVM.witness-memdb-size=256 --batch-size=64
```

## Contributing to CDK-Erigon

### Contribution Process

1. **Find an Issue**: Check the GitHub issue tracker for open issues
2. **Discuss**: Comment on the issue to express interest
3. **Fork & Branch**: Fork the repository and create a feature branch
4. **Implement**: Make your changes with appropriate tests
5. **Test Locally**: Ensure all tests pass
6. **Create PR**: Submit a pull request with detailed description
7. **Code Review**: Address reviewer feedback
8. **Merge**: Maintainers will merge approved PRs

### Documentation

When contributing, update relevant documentation:

1. Update code comments for functions you modify
2. Update README or documentation if behavior changes
3. Add examples for new features
4. Document configuration parameters

### Commit Message Format

Use this format for commit messages:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Where:
- `<type>`: feat, fix, docs, style, refactor, test, chore
- `<scope>`: component affected (e.g., sync, rpc, witness)
- `<subject>`: short description (imperative mood)
- `<body>`: detailed description
- `<footer>`: breaking changes, issue references

Example:
```
fix(sync): address l1 connection retry logic

Implements exponential backoff for L1 connection retries
to prevent excessive reconnection attempts and potential
API rate limiting.

Fixes #1234
```

## Extending CDK-Erigon

### Adding New RPC Methods

To add a new RPC method:

1. Locate the appropriate API package in `zkevm/rpc/`
2. Define your new method function:

```go
// NewZkRollupAPI returns a new zkRollup API instance.
func NewZkRollupAPI(backend *eth.EthAPIBackend) *ZkRollupAPI {
    return &ZkRollupAPI{
        backend: backend,
    }
}

// API returns the RPC descriptor as an array of methods
func (api *ZkRollupAPI) API() []rpc.API {
    return []rpc.API{
        {
            Namespace: "zkevm",
            Version:   "1.0",
            Service:   api,
            Public:    true,
        },
    }
}

// MyNewMethod implements an example RPC method
func (api *ZkRollupAPI) MyNewMethod(ctx context.Context, param string) (string, error) {
    // Implementation
    return "result", nil
}
```

3. Register the API with the node in `cmd/erigon/main.go`

### Custom Extensions

To build custom extensions:

1. Create a plugin package
2. Use the available hooks and interfaces
3. Register your extension during node startup

Example plugin structure:
```go
type MyExtension struct {
    node *node.Node
}

func NewMyExtension(node *node.Node) *MyExtension {
    return &MyExtension{
        node: node,
    }
}

func (e *MyExtension) Start() error {
    // Start extension
    return nil
}

func (e *MyExtension) Stop() error {
    // Clean up
    return nil
}
```

Registration:
```go
myExt := NewMyExtension(stack)
err := stack.RegisterLifecycle(myExt)
if err != nil {
    return err
}
```

## Performance Optimization

### Database Optimization

- Use appropriate compaction strategies
- Adjust cache sizes based on available memory
- Consider separating hot and cold data

```yaml
chaindata:
  inbound:
    cache-size: 1024
  outbound:
    cache-size: 1024
```

### Synchronization Performance

- Use dedicated hardware for zkEVM components
- Adjust batch sizes for optimal processing
- Consider using SSD storage for database

### Memory Management

- Use appropriate buffer sizes
- Implement pooling for frequently allocated objects
- Schedule periodic garbage collection

## Troubleshooting Common Development Issues

### Build Failures

**Issue**: Build fails with dependency errors

**Solution**:
- Run `go mod tidy`
- Check Go version compatibility
- Verify required C libraries are installed

### Test Failures

**Issue**: Tests fail inconsistently

**Solution**:
- Look for race conditions with `go test -race`
- Check for timeouts and extend if needed
- Verify test dependencies are properly mocked

### Runtime Crashes

**Issue**: Application crashes during execution

**Solution**:
- Check logs for panic traces
- Run with `GODEBUG=gctrace=1` to identify memory issues
- Use delve for step-by-step debugging:
  ```bash
  dlv debug ./cmd/erigon/main.go
  ```

## Frequently Asked Development Questions

### Q: How do I debug a state root mismatch?

A: State root mismatches typically occur when transaction execution produces different results than expected. To debug:

1. Enable detailed logging: `--verbosity=5 --zkEVM.debug-save-witness=true`
2. Compare batch execution results:
   ```bash
   go run ./cmd/hack/compare_batches.go --batch=123 --node1=http://localhost:8545 --node2=http://reference:8545
   ```
3. Trace the problematic transaction:
   ```bash
   curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"debug_traceTransaction","params":["0x..."],"id":1}' localhost:8545
   ```

### Q: How do I profile the synchronization performance?

A: To profile sync performance:

1. Enable profiling: `--pprof --pprof.addr=localhost:6060`
2. Capture CPU profile during sync:
   ```bash
   go tool pprof -seconds 30 http://localhost:6060/debug/pprof/profile
   ```
3. Analyze results:
   ```bash
   go tool pprof -http=:8080 profile.pb.gz
   ```

### Q: How do I add support for a new L1 chain?

A: To add support for a new L1 chain:

1. Define chain configuration in `params/config.go`
2. Add contract deployments for the new chain
3. Update the chain detection logic in `zkevm/chain.go`
4. Test with the new chain configuration

## Advanced Development Topics

### Circuit Development

If modifying the zkEVM circuits:

1. Understand the circuit architecture in `zkevm/circuit/`
2. Make small, incremental changes
3. Test thoroughly with `make test-circuit`
4. Benchmark before and after changes

### State Management Optimization

For state management improvements:

1. Profile state operations with large accounts
2. Consider optimizing for specific access patterns
3. Test with the official state tests suite
4. Benchmark with representative workloads

### Custom RPC Extensions

For advanced RPC customization:

1. Use the extension pattern in `rpc/extension.go`
2. Implement custom authentication if needed
3. Consider websocket subscriptions for real-time data
4. Document your extensions thoroughly

## References

- [Go Documentation](https://golang.org/doc/)
- [Ethereum JSON-RPC Spec](https://ethereum.github.io/execution-apis/api-documentation/)
- [zkEVM Documentation](https://docs.polygon.technology/docs/zkEVM/)
- [Erigon Architecture](https://github.com/ledgerwatch/erigon) 