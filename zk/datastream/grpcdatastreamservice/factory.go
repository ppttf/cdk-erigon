package grpcdatastreamservice

import (
	"context"
	"fmt"
	"time"

	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon/zk/datastream/client"
	servicepb "github.com/erigontech/erigon/zk/datastream/proto/datastream"
)

// DefaultDatastreamFactory is the default implementation of DatastreamFactory
type DefaultDatastreamFactory struct{}

// NewDefaultDatastreamFactory creates a new instance of DefaultDatastreamFactory
func NewDefaultDatastreamFactory() *DefaultDatastreamFactory {
	return &DefaultDatastreamFactory{}
}

// NewServer creates a new GRPCDataStreamServer instance
func (f *DefaultDatastreamFactory) NewGRPCDataStreamServer(logger log.Logger, client DatastreamBridgeClient) (servicepb.DataStreamServiceServer, error) {
	// Start the client
	if err := client.Start(); err != nil {
		return nil, err
	}

	return &GRPCDataStreamServer{
		txStreams: make(map[string]chan *servicepb.TransactionResponse),
		logger:    logger,
		client:    client,
	}, nil
}

// NewClient creates a new RelayDatastreamClient instance
func (f *DefaultDatastreamFactory) NewRelayDatastreamClient(ctx context.Context, relayPort uint, logger log.Logger) (*RelayDatastreamClient, error) {
	// Create stream client with default configuration
	serverAddr := fmt.Sprintf("localhost:%d", relayPort)
	dsClient := client.NewClient(
		ctx,
		serverAddr,
		false,         // useTLS
		5*time.Second, // checkTimeout
		0,             // latestDownloadedForkId
		client.DefaultEntryChannelSize,
	)

	return &RelayDatastreamClient{
		client: dsClient,
		logger: logger,
	}, nil
}
