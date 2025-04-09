package grpcdatastreamservice

import (
	"context"
	"fmt"
	"sync"

	"github.com/erigontech/erigon/zk/datastream/proto/datastream"
	servicepb "github.com/erigontech/erigon/zk/datastream/proto/datastream"
	"github.com/erigontech/erigon/zk/datastream/types"
	"google.golang.org/grpc/metadata"
)

// MockTransactionStream implements the servicepb.DataStreamService_GetTransactionStreamServer interface
type MockTransactionStream struct {
	ctx           context.Context
	receivedTxs   []*datastream.TransactionResponse
	mockSendError error
	mu            sync.Mutex
}

func NewMockTransactionStream(ctx context.Context) *MockTransactionStream {
	return &MockTransactionStream{
		ctx:         ctx,
		receivedTxs: make([]*datastream.TransactionResponse, 0),
	}
}

func (m *MockTransactionStream) Send(tx *datastream.TransactionResponse) error {
	if m.mockSendError != nil {
		return m.mockSendError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.receivedTxs = append(m.receivedTxs, tx)
	return nil
}

func (m *MockTransactionStream) SetSendError(err error) {
	m.mockSendError = err
}

func (m *MockTransactionStream) GetReceivedTransactions() []*datastream.TransactionResponse {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.receivedTxs
}

func (m *MockTransactionStream) Context() context.Context {
	return m.ctx
}

func (m *MockTransactionStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *MockTransactionStream) SetHeader(metadata.MD) error {
	return nil
}

func (m *MockTransactionStream) SetTrailer(metadata.MD) {
}

// RecvMsg implements grpc.ServerStream interface
func (m *MockTransactionStream) RecvMsg(msg interface{}) error {
	return nil
}

// SendMsg implements grpc.ServerStream interface
func (m *MockTransactionStream) SendMsg(msg interface{}) error {
	return nil
}

// MockDatastreamBridgeClient implements the DatastreamClient interface for testing
type MockDatastreamBridgeClient struct {
	mu            sync.Mutex
	started       bool
	startError    error
	stopError     error
	readError     error
	entryChan     chan interface{}
	header        *types.HeaderEntry
	streamInfo    *servicepb.StreamInfoResponse
	streamInfoErr error
	l2Blocks      map[uint64]*types.FullL2Block
}

func NewMockDatastreamBridgeClient() *MockDatastreamBridgeClient {
	return &MockDatastreamBridgeClient{
		entryChan: make(chan interface{}, 100),
		header: &types.HeaderEntry{
			Version:    1,
			StreamType: 1,
		},
		streamInfo: &servicepb.StreamInfoResponse{
			DatastreamVersion:  "1.0",
			HighestBlockNumber: 1,
			HighestBatchNumber: 1,
			TotalEntries:       1,
			PendingTxCount:     0,
			QueuedTxCount:      0,
		},
		l2Blocks: make(map[uint64]*types.FullL2Block),
	}
}

func (m *MockDatastreamBridgeClient) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return nil
	}
	m.started = true
	return m.startError
}

func (m *MockDatastreamBridgeClient) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.started {
		return nil
	}
	m.started = false
	close(m.entryChan)
	return m.stopError
}

func (m *MockDatastreamBridgeClient) GetEntryChan() *chan interface{} {
	return &m.entryChan
}

func (m *MockDatastreamBridgeClient) ReadAllEntriesToChannel() error {
	if m.readError != nil {
		return m.readError
	}
	return nil
}

func (m *MockDatastreamBridgeClient) GetL2BlockByNumber(blockNum uint64) (*types.FullL2Block, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if block, exists := m.l2Blocks[blockNum]; exists {
		return block, nil
	}
	return nil, fmt.Errorf("block %d not found", blockNum)
}

func (m *MockDatastreamBridgeClient) GetLatestL2Block() (*types.FullL2Block, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var latestBlock *types.FullL2Block
	var latestNum uint64
	for num, block := range m.l2Blocks {
		if num > latestNum {
			latestNum = num
			latestBlock = block
		}
	}
	if latestBlock == nil {
		return nil, fmt.Errorf("no blocks found")
	}
	return latestBlock, nil
}

func (m *MockDatastreamBridgeClient) GetHeader() (*types.HeaderEntry, error) {
	return m.header, nil
}

func (m *MockDatastreamBridgeClient) GetStreamInfo(ctx context.Context) (*servicepb.StreamInfoResponse, error) {
	return m.streamInfo, m.streamInfoErr
}

// Test helper methods
func (m *MockDatastreamBridgeClient) SetStartError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startError = err
}

func (m *MockDatastreamBridgeClient) SetStopError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopError = err
}

func (m *MockDatastreamBridgeClient) SetReadError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readError = err
}

func (m *MockDatastreamBridgeClient) SetStreamInfo(info *servicepb.StreamInfoResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streamInfo = info
}

func (m *MockDatastreamBridgeClient) SetStreamInfoError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streamInfoErr = err
}

func (m *MockDatastreamBridgeClient) AddL2Block(blockNum uint64, block *types.FullL2Block) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.l2Blocks[blockNum] = block
}

func (m *MockDatastreamBridgeClient) AddEntry(entry interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		m.entryChan <- entry
	}
}
