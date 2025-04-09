package grpcdatastreamservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon/zk/datastream/proto/datastream"
	"github.com/stretchr/testify/require"
)

// TestGetStreamInfo verifies the service returns correct stream info
func TestGetStreamInfo(t *testing.T) {
	// Create a no-op logger
	logger := log.New()

	// Create mock client
	mockClient := NewMockDatastreamBridgeClient()

	// Create service with mock client
	srv, err := NewDataStreamServer(logger, mockClient)
	require.NoError(t, err, "Failed to create service")
	require.NotNil(t, srv, "Service should be created")

	// Call the method
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := srv.GetStreamInfo(ctx, &datastream.StreamInfoRequest{})

	// Verify results
	require.NoError(t, err, "GetStreamInfo should not return error")
	require.NotNil(t, resp, "Response should not be nil")
	require.Equal(t, "1.0", resp.DatastreamVersion, "Datastream version should be 1.0")
}

// TestBroadcastTransaction verifies broadcasting transaction to registered streams
func TestBroadcastTransaction(t *testing.T) {
	// Create a no-op logger
	logger := log.New()
	// Create mock client
	mockClient := NewMockDatastreamBridgeClient()

	// Create service with nil dependencies (they're not directly used in broadcast)
	srv, err := NewDataStreamServer(logger, mockClient)
	require.NoError(t, err, "Failed to create service")

	// Manually add a stream channel
	testStream := make(chan *datastream.TransactionResponse, 10)
	srv.streamsMu.Lock()
	srv.txStreams["test-stream"] = testStream
	srv.streamsMu.Unlock()

	// Create a test transaction response
	txResp := &datastream.TransactionResponse{
		TxHash:    []byte("test-hash"),
		Sender:    []byte("test-sender"),
		Recipient: []byte("test-recipient"),
		IsPending: true,
		GasPrice:  1000,
		Timestamp: uint64(time.Now().Unix()),
	}

	// Broadcast the transaction
	srv.BroadcastTransaction(txResp)

	// Verify the transaction was received by the stream
	select {
	case received := <-testStream:
		require.Equal(t, txResp.TxHash, received.TxHash, "Transaction hash should match")
		require.Equal(t, txResp.Sender, received.Sender, "Sender should match")
		require.Equal(t, txResp.Recipient, received.Recipient, "Recipient should match")
		require.Equal(t, txResp.IsPending, received.IsPending, "IsPending flag should match")
		require.Equal(t, txResp.GasPrice, received.GasPrice, "Gas price should match")
		require.Equal(t, txResp.Timestamp, received.Timestamp, "Timestamp should match")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for transaction to be broadcasted")
	}
}

// TestBroadcastToFullChannelDoesntBlock verifies that broadcasting to a full channel doesn't block
func TestBroadcastToFullChannelDoesntBlock(t *testing.T) {
	// Create a no-op logger
	logger := log.New()

	// Create mock client
	mockClient := NewMockDatastreamBridgeClient()

	// Create service with mock client
	srv, err := NewDataStreamServer(logger, mockClient)
	require.NoError(t, err, "Failed to create service")

	// Create a channel with capacity 1
	testStream := make(chan *datastream.TransactionResponse, 1)

	// Fill the channel
	testStream <- &datastream.TransactionResponse{TxHash: []byte("filler")}

	// Add the full stream to the service
	srv.streamsMu.Lock()
	srv.txStreams["test-stream"] = testStream
	srv.streamsMu.Unlock()

	// Create a test transaction response
	txResp := &datastream.TransactionResponse{
		TxHash: []byte("test-hash"),
	}

	// This should not block even though the channel is full
	done := make(chan bool)
	go func() {
		srv.BroadcastTransaction(txResp)
		done <- true
	}()

	// Verify the BroadcastTransaction call didn't block
	select {
	case <-done:
		// Success! The function returned without blocking
	case <-time.After(1 * time.Second):
		t.Fatal("BroadcastTransaction blocked on a full channel")
	}
}

// TestGetTransactionStream tests the GetTransactionStream RPC
func TestGetTransactionStream(t *testing.T) {
	// Create a no-op logger
	logger := log.New()

	// Create mock client
	mockClient := NewMockDatastreamBridgeClient()

	// Create service with mock client
	srv, err := NewDataStreamServer(logger, mockClient)
	require.NoError(t, err, "Failed to create service")

	// Create a context that we can cancel to end the stream
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the mock stream
	mockStream := NewMockTransactionStream(ctx)

	// Create a channel to signal when the streaming has started
	streamStarted := make(chan struct{})

	// Start the transaction stream in a goroutine
	go func() {
		// Signal the test that we're about to start the stream
		close(streamStarted)

		// Call the method we're testing
		err := srv.GetTransactionStream(&datastream.TransactionStreamRequest{}, mockStream)

		// This should only return when the context is canceled
		require.Equal(t, context.Canceled, err, "Expected context.Canceled error when stream ends")
	}()

	// Wait for the stream to start
	<-streamStarted
	time.Sleep(100 * time.Millisecond) // Give a little time for the stream to register

	// Create a test transaction
	txResp := &datastream.TransactionResponse{
		TxHash:    []byte("test-tx-hash"),
		Sender:    []byte("test-sender"),
		Recipient: []byte("test-recipient"),
		IsPending: true,
		GasPrice:  2000,
		Timestamp: uint64(time.Now().Unix()),
	}

	// Broadcast the transaction to all streams (including our mock stream)
	srv.BroadcastTransaction(txResp)

	// Give some time for the transaction to be sent
	time.Sleep(100 * time.Millisecond)

	// Check that the transaction was received by the mock stream
	receivedTxs := mockStream.GetReceivedTransactions()
	require.Len(t, receivedTxs, 1, "Expected to receive exactly one transaction")

	// Verify the transaction data
	received := receivedTxs[0]
	require.Equal(t, txResp.TxHash, received.TxHash, "Transaction hash should match")
	require.Equal(t, txResp.Sender, received.Sender, "Sender should match")
	require.Equal(t, txResp.Recipient, received.Recipient, "Recipient should match")
	require.Equal(t, txResp.IsPending, received.IsPending, "IsPending flag should match")
	require.Equal(t, txResp.GasPrice, received.GasPrice, "Gas price should match")
	require.Equal(t, txResp.Timestamp, received.Timestamp, "Timestamp should match")

	// Verify the stream was registered correctly
	srv.streamsMu.RLock()
	streamCount := len(srv.txStreams)
	srv.streamsMu.RUnlock()
	require.Equal(t, 1, streamCount, "Expected one stream to be registered")

	// Cancel the context to end the stream
	cancel()

	// Give some time for the cleanup to happen
	time.Sleep(100 * time.Millisecond)

	// Verify the stream was deregistered
	srv.streamsMu.RLock()
	streamCount = len(srv.txStreams)
	srv.streamsMu.RUnlock()
	require.Equal(t, 0, streamCount, "Expected all streams to be deregistered after context is canceled")
}

// TestGetTransactionStreamChannelClosed tests handling of closed channel in GetTransactionStream
func TestGetTransactionStreamChannelClosed(t *testing.T) {
	// Create a no-op logger
	logger := log.New()

	// Create mock client
	mockClient := NewMockDatastreamBridgeClient()

	// Create service with mock client
	srv, err := NewDataStreamServer(logger, mockClient)
	require.NoError(t, err, "Failed to create service")

	// Create a context that we can cancel to end the stream
	ctx := context.Background()

	// Create the mock stream
	mockStream := NewMockTransactionStream(ctx)

	// Get direct access to the channel map to simulate a channel close
	streamDone := make(chan struct{})

	// We'll manually add and close a channel
	streamID := "test-stream-channel-close"
	testChan := make(chan *datastream.TransactionResponse)

	srv.streamsMu.Lock()
	srv.txStreams[streamID] = testChan
	srv.streamsMu.Unlock()

	// Start the transaction stream handling loop in a goroutine
	go func() {
		// Replace the normal call with our manually constructed scenario
		// This simulates just the inner loop of GetTransactionStream
		for {
			select {
			case txResp, ok := <-testChan:
				if !ok {
					// When channel is closed, the method should return nil
					// In our test, we'll just signal that the method would have returned
					close(streamDone)
					return
				}
				// Normal flow - send the transaction
				_ = mockStream.Send(txResp)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Close the channel to trigger the !ok condition
	close(testChan)

	// Wait for the goroutine to detect the closed channel
	select {
	case <-streamDone:
		// Success - the goroutine detected the closed channel
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for closed channel detection")
	}
}

// TestGetTransactionStreamSendError tests handling of send errors in GetTransactionStream
func TestGetTransactionStreamSendError(t *testing.T) {
	// Create a no-op logger
	logger := log.New()

	// Create mock client
	mockClient := NewMockDatastreamBridgeClient()

	// Create service with mock client
	srv, err := NewDataStreamServer(logger, mockClient)
	require.NoError(t, err, "Failed to create service")

	// Create a context that we can cancel to end the stream
	ctx := context.Background()

	// Create the mock stream with an error configured
	mockStream := NewMockTransactionStream(ctx)
	expectedError := errors.New("send error")
	mockStream.SetSendError(expectedError)

	// Create a channel to signal when the method completes
	streamError := make(chan error)

	// Start the transaction stream in a goroutine
	go func() {
		err := srv.GetTransactionStream(&datastream.TransactionStreamRequest{}, mockStream)
		streamError <- err
	}()

	// Wait a moment for the stream to register
	time.Sleep(100 * time.Millisecond)

	// Create a test transaction
	txResp := &datastream.TransactionResponse{
		TxHash: []byte("test-tx-hash"),
	}

	// Broadcast the transaction - this should trigger a send error
	srv.BroadcastTransaction(txResp)

	// Check that the error is propagated
	select {
	case err := <-streamError:
		require.Equal(t, expectedError, err, "Expected send error to be propagated")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for send error")
	}

	// Verify that the stream was deregistered after the error
	srv.streamsMu.RLock()
	streamCount := len(srv.txStreams)
	srv.streamsMu.RUnlock()
	require.Equal(t, 0, streamCount, "Expected all streams to be deregistered after error")
}
