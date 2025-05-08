package main

import (
	"context"

	"github.com/erigontech/erigon-lib/log/v3" // Needed for RelayDatastreamClient constructor
	// Needed for RelayDatastreamClient constructor
	"github.com/erigontech/erigon/zk/datastream/grpcdatastreamservice"
	servicepb "github.com/erigontech/erigon/zk/datastream/proto/datastream" // Added import alias
)

// MockDatastreamFactory is a simple implementation of DatastreamFactory for testing
type MockDatastreamFactory struct{}

// Ensure MockDatastreamFactory implements the interface at compile time
var _ grpcdatastreamservice.DatastreamFactory = (*MockDatastreamFactory)(nil)

func NewMockDatastreamFactory() *MockDatastreamFactory { // Return concrete type for easier test access if needed
	return &MockDatastreamFactory{}
}

// NewGRPCDataStreamServer returns a real server instance for basic test compilation/execution.
// It now returns the servicepb.DataStreamServiceServer interface type.
func (m *MockDatastreamFactory) NewGRPCDataStreamServer(logger log.Logger, bridgeClient grpcdatastreamservice.DatastreamBridgeClient) (servicepb.DataStreamServiceServer, error) {
	return nil, nil // *GRPCDataStreamServer implements the interface
}

// NewRelayDatastreamClient returns a real client instance for basic test compilation/execution.
func (m *MockDatastreamFactory) NewRelayDatastreamClient(ctx context.Context, relayPort uint, logger log.Logger) (*grpcdatastreamservice.RelayDatastreamClient, error) {

	return nil, nil
}
