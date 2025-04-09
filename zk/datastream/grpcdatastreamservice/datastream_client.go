package grpcdatastreamservice

import (
	"context"
	"fmt"
	"time"

	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon/zk/datastream/client"
	servicepb "github.com/erigontech/erigon/zk/datastream/proto/datastream"
	"github.com/erigontech/erigon/zk/datastream/types"
)

// DatastreamClient defines the interface for a datastream client
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

// RelayDatastreamClient represents a client for the relay's datastream service
type RelayDatastreamClient struct {
	client *client.StreamClient
	logger log.Logger
}

// NewRelayDatastreamClient creates a new client connected to the relay server
func NewRelayDatastreamClient(ctx context.Context, relayPort uint, logger log.Logger) (*RelayDatastreamClient, error) {
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

// Start starts the client connection
func (c *RelayDatastreamClient) Start() error {
	return c.client.Start()
}

// Stop stops the client connection
func (c *RelayDatastreamClient) Stop() error {
	return c.client.Stop()
}

// GetEntryChan returns the channel for receiving entries
func (c *RelayDatastreamClient) GetEntryChan() *chan interface{} {
	return c.client.GetEntryChan()
}

// ReadAllEntriesToChannel starts reading all entries into the channel
func (c *RelayDatastreamClient) ReadAllEntriesToChannel() error {
	return c.client.ReadAllEntriesToChannel()
}

// GetL2BlockByNumber retrieves a specific L2 block by its number
func (c *RelayDatastreamClient) GetL2BlockByNumber(blockNum uint64) (*types.FullL2Block, error) {
	return c.client.GetL2BlockByNumber(blockNum)
}

// GetLatestL2Block retrieves the latest L2 block
func (c *RelayDatastreamClient) GetLatestL2Block() (*types.FullL2Block, error) {
	return c.client.GetLatestL2Block()
}

// GetHeader retrieves the stream header
func (c *RelayDatastreamClient) GetHeader() (*types.HeaderEntry, error) {
	return c.client.GetHeader()
}

// GetStreamInfo retrieves information about the current datastream
func (c *RelayDatastreamClient) GetStreamInfo(ctx context.Context) (*servicepb.StreamInfoResponse, error) {
	header, err := c.GetHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to get header: %v", err)
	}

	latestBlock, err := c.GetLatestL2Block()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %v", err)
	}

	return &servicepb.StreamInfoResponse{
		HighestBlockNumber: latestBlock.L2BlockNumber,
		HighestBatchNumber: latestBlock.BatchNumber,
		TotalEntries:       header.TotalEntries,
		PendingTxCount:     0, // TODO: Implement pending tx count
		QueuedTxCount:      0, // TODO: Implement queued tx count
		DatastreamVersion:  fmt.Sprintf("%d", header.Version),
	}, nil
}
