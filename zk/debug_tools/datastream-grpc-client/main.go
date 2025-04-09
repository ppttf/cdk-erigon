package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	servicepb "github.com/erigontech/erigon/zk/datastream/proto/datastream"
)

var (
	serverAddr = flag.String("addr", "localhost:7070", "The server address in the format of host:port")
)

func main() {
	flag.Parse()

	// Set up a connection to the server
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client
	client := servicepb.NewDataStreamServiceClient(conn)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Get stream info
	info, err := client.GetStreamInfo(ctx, &servicepb.StreamInfoRequest{})
	if err != nil {
		log.Fatalf("Failed to get stream info: %v", err)
	}

	fmt.Printf("Stream Info:\n")
	fmt.Printf("  Highest Block: %d\n", info.HighestBlockNumber)
	fmt.Printf("  Highest Batch: %d\n", info.HighestBatchNumber)
	fmt.Printf("  Total Entries: %d\n", info.TotalEntries)
	fmt.Printf("  Pending TXs: %d\n", info.PendingTxCount)
	fmt.Printf("  Queued TXs: %d\n", info.QueuedTxCount)
	fmt.Printf("  Version: %s\n", info.DatastreamVersion)

	// Get transaction stream
	//stream, err := client.GetTransactionStream(ctx, &servicepb.TransactionStreamRequest{})
	//if err != nil {
	//	log.Fatalf("Failed to get transaction stream: %v", err)
	//}
	//
	//// Receive transactions
	//for {
	//	tx, err := stream.Recv()
	//	if err != nil {
	//		log.Printf("Error receiving transaction: %v", err)
	//		break
	//	}
	//	fmt.Printf("Received transaction:\n")
	//	fmt.Printf("  Hash: %x\n", tx.TxHash)
	//	fmt.Printf("  Sender: %x\n", tx.Sender)
	//	fmt.Printf("  Recipient: %x\n", tx.Recipient)
	//	fmt.Printf("  Pending: %v\n", tx.IsPending)
	//	fmt.Printf("  Gas Price: %d\n", tx.GasPrice)
	//	fmt.Printf("  Timestamp: %d\n", tx.Timestamp)
	//}
}
