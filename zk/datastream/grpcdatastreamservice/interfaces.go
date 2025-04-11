package grpcdatastreamservice

import (
	"context"

	"github.com/erigontech/erigon-lib/log/v3"
	servicepb "github.com/erigontech/erigon/zk/datastream/proto/datastream"
	"github.com/erigontech/erigon/zk/datastream/types"
)

// DatastreamBridgeClient defines the interface for a datastream client
type DatastreamBridgeClient interface {
	// Start starts the client connection
	Start() error

	// Stop stops the client connection
	Stop() error

	// GetEntryChan returns the channel for receiving entries
	GetEntryChan() *chan interface{}

	// ReadAllEntriesToChannel reads all entries into the channel
	ReadAllEntriesToChannel() error

	// GetL2BlockByNumber retrieves an L2 block by its number
	GetL2BlockByNumber(blockNum uint64) (*types.FullL2Block, error)

	// GetLatestL2Block retrieves the latest L2 block
	GetLatestL2Block() (*types.FullL2Block, error)

	// GetHeader retrieves the stream header
	GetHeader() (*types.HeaderEntry, error)

	// GetStreamInfo retrieves information about the stream
	GetStreamInfo(ctx context.Context) (*servicepb.StreamInfoResponse, error)
}

// DatastreamFactory defines the interface for creating datastream server and client instances
type DatastreamFactory interface {
	// NewServer creates a new GRPCDataStreamServer instance
	NewGRPCDataStreamServer(logger log.Logger, client DatastreamBridgeClient) (servicepb.DataStreamServiceServer, error)

	// NewClient creates a new RelayDatastreamClient instance
	NewRelayDatastreamClient(ctx context.Context, relayPort uint, logger log.Logger) (*RelayDatastreamClient, error)
}
