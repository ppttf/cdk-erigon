package grpcdatastreamservice

import (
	"context"
	"sync"

	libcommon "github.com/erigontech/erigon-lib/common"
	"github.com/erigontech/erigon-lib/log/v3"

	// Import from the local relative path based on the new go_package
	servicepb "github.com/erigontech/erigon/zk/datastream/proto/datastream"
	"google.golang.org/grpc"
)

// GRPCDataStreamServer implements the DataStreamService gRPC interface
type GRPCDataStreamServer struct {
	servicepb.UnimplementedDataStreamServiceServer
	logger log.Logger
	client DatastreamBridgeClient

	// For managing active stream connections
	streamsMu sync.RWMutex
	txStreams map[string]chan *servicepb.TransactionResponse
}

// GetTransactionStream implements the DataStreamService.GetTransactionStream method
func (s *GRPCDataStreamServer) GetTransactionStream(req *servicepb.TransactionStreamRequest, stream servicepb.DataStreamService_GetTransactionStreamServer) error {
	// Create a unique channel for this stream
	streamID := generateStreamID()

	// Create buffered channel for this stream
	txChan := make(chan *servicepb.TransactionResponse, 100)

	// Register the stream
	s.streamsMu.Lock()
	s.txStreams[streamID] = txChan
	s.streamsMu.Unlock()

	// Cleanup when the stream ends
	defer func() {
		s.streamsMu.Lock()
		delete(s.txStreams, streamID)
		close(txChan)
		s.streamsMu.Unlock()
	}()

	// Process transactions from the channel and send them to the client
	for {
		select {
		case txResp, ok := <-txChan:
			if !ok {
				// Channel closed
				return nil
			}
			if err := stream.Send(txResp); err != nil {
				s.logger.Warn("Failed to send transaction to stream", "err", err)
				return err
			}
		case <-stream.Context().Done():
			// Client disconnected
			return stream.Context().Err()
		}
	}
}

// GetStreamInfo implements the DataStreamService.GetStreamInfo method
func (s *GRPCDataStreamServer) GetStreamInfo(ctx context.Context, req *servicepb.StreamInfoRequest) (*servicepb.StreamInfoResponse, error) {
	return s.client.GetStreamInfo(ctx)
}

// BroadcastTransaction broadcasts a transaction to all active streams
func (s *GRPCDataStreamServer) BroadcastTransaction(txResp *servicepb.TransactionResponse) {
	s.streamsMu.RLock()
	defer s.streamsMu.RUnlock()

	// Send to all active streams
	for _, ch := range s.txStreams {
		select {
		case ch <- txResp:
			// Successfully sent
		default:
			// Channel full, transaction will be dropped for this client
			// This prevents slow clients from blocking the system
		}
	}
}

// RegisterWithGrpcServer registers the GRPCDataStreamServer with a gRPC server
func RegisterWithGrpcServer(grpcServer *grpc.Server, dataStreamServer servicepb.DataStreamServiceServer) {
	servicepb.RegisterDataStreamServiceServer(grpcServer, dataStreamServer)
}

// Helper function to generate a unique stream ID
// In a real implementation, you would use something more robust
func generateStreamID() string {
	return "stream-" + libcommon.Hash{}.String()
}

// NewGRPCDataStreamServer creates a new GRPCDataStreamServer instance
func NewGRPCDataStreamServer(logger log.Logger, client DatastreamBridgeClient) (*GRPCDataStreamServer, error) {
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
