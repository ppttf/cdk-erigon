package grpcdatastreamservice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon/zk/datastream/client"
	servicepb "github.com/erigontech/erigon/zk/datastream/proto/datastream"
	"github.com/erigontech/erigon/zk/datastream/types"
)

// RelayDatastreamClient represents a client for the relay's datastream service
type RelayDatastreamClient struct {
	client        *client.StreamClient
	logger        log.Logger
	keepaliveStop func()
	keepaliveMu   sync.Mutex
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
	// Start the underlying client connection
	err := c.client.HandleStart()
	if err != nil {
		return err
	}

	// Start the keepalive mechanism with a 30-second interval
	c.keepaliveMu.Lock()
	c.keepaliveStop = c.StartKeepAlive(30 * time.Second)
	c.keepaliveMu.Unlock()

	return nil
}

// Stop stops the client connection
func (c *RelayDatastreamClient) Stop() error {
	// Stop the keepalive mechanism if it's running
	c.keepaliveMu.Lock()
	if c.keepaliveStop != nil {
		c.keepaliveStop()
		c.keepaliveStop = nil
	}
	c.keepaliveMu.Unlock()

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

// StartKeepAlive starts a goroutine that periodically sends NOOP commands to keep
// the connection to the datastream server alive. This helps prevent timeouts during
// periods of inactivity.
//
// Parameters:
//   - interval: How often to send NOOP commands (e.g., 30s)
//
// Returns:
//   - A function that stops the keepalive mechanism when called
func (c *RelayDatastreamClient) StartKeepAlive(interval time.Duration) func() {
	stopCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		c.logger.Info("Started datastream keepalive mechanism", "interval", interval)

		for {
			select {
			case <-ticker.C:
				err := c.client.SendNoop()
				if err != nil {
					c.logger.Warn("Failed to send keepalive NOOP command", "err", err)
					// Don't exit the loop on failure; keep trying
				} else {
					c.logger.Debug("Sent keepalive NOOP command successfully")
				}
			case <-stopCh:
				c.logger.Info("Stopping datastream keepalive mechanism")
				return
			}
		}
	}()

	// Return function to stop the keepalive routine
	return func() {
		close(stopCh)
	}
}
