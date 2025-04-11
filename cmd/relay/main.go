package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon/zk/datastream/grpcdatastreamservice"
	"github.com/gateway-fm/zkevm-data-streamer/datastreamer"
	"google.golang.org/grpc"
)

const (
	streamerSystemID    = 137
	streamerVersion     = 1
	streamTypeSequencer = 1
)

// RelayConfig holds all configuration options
type RelayConfig struct {
	ServerAddr        string
	RelayPort         uint
	GRPCPort          uint
	DataFile          string
	LogLevel          string
	WriteTimeoutMs    uint
	InactivityTimeout uint
}

var config *RelayConfig
var logger log.Logger

func parseFlags() {
	serverAddr := flag.String("server", "127.0.0.1:6900", "datastream server address to connect to")
	relayPort := flag.Uint("relay-port", 7900, "port to expose for clients to connect")
	grpcPort := flag.Uint("grpc-port", 7070, "port to expose for gRPC clients to connect")
	dataFile := flag.String("datafile", "test.dat", "relay data file name")
	logLevel := flag.String("log", "info", "log level (debug, info, warn, error)")
	writeTimeoutMs := flag.Uint("writetimeout", 3000, "timeout for write operations on client connections in ms (0=no timeout)")
	inactivityTimeout := flag.Uint("inactivitytimeout", 120, "timeout to kill an inactive client connection in seconds (0=no timeout)")
	flag.Parse()

	config = &RelayConfig{ServerAddr: *serverAddr, RelayPort: *relayPort, GRPCPort: *grpcPort, DataFile: *dataFile, LogLevel: *logLevel, WriteTimeoutMs: *writeTimeoutMs, InactivityTimeout: *inactivityTimeout}

}
func initLogger() {
	logger = log.New()
}

func createRelayServer() (*datastreamer.StreamRelay, error) {
	if config == nil {
		config = &RelayConfig{} // Initialize with defaults if not set
	}
	return datastreamer.NewRelay(
		config.ServerAddr,
		uint16(config.RelayPort),
		streamerVersion,
		streamerSystemID,
		streamTypeSequencer,
		config.DataFile,
		time.Duration(config.WriteTimeoutMs)*time.Millisecond,
		time.Duration(config.InactivityTimeout)*time.Second,
		5*time.Second,
		nil,
	)
}

func setupGRPCServer(factory grpcdatastreamservice.DatastreamFactory) (*grpc.Server, net.Listener, error) {
	if config == nil {
		config = &RelayConfig{} // Initialize with defaults if not set
	}
	server := grpc.NewServer()
	datastreamClient, err := factory.NewRelayDatastreamClient(context.Background(), config.RelayPort, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create datastream client: %v", err)
	}

	datastreamService, err := factory.NewGRPCDataStreamServer(logger, datastreamClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create datastream service: %v", err)
	}
	grpcdatastreamservice.RegisterWithGrpcServer(server, datastreamService)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GRPCPort))
	if err != nil {
		return nil, nil, err
	}

	return server, listener, nil
}

func handleSignals(sigChan chan os.Signal) {
	sig := <-sigChan
	logger.Info("Received signal %v, shutting down...", sig)
}

func main() {

	//create logger
	initLogger()

	// Parse command line flags
	parseFlags()

	logger.Info(">> Relay server starting: port[%d] grpc-port[%d] file[%s] server[%s] log[%s]",
		config.RelayPort, config.GRPCPort, config.DataFile, config.ServerAddr, config.LogLevel)

	// Create and start relay server
	relay, err := createRelayServer()
	if err != nil {
		logger.Error(">> Relay server: NewRelay error! (%v)", err)
		os.Exit(1)
	}

	if err := relay.Start(); err != nil {
		logger.Error(">> Relay server: Start error! (%v)", err)
		os.Exit(1)
	}

	defer func() {
		if err := relay.Stop(); err != nil {
			logger.Error("Error stopping relay: %v", err)
		}
		logger.Info(">> Relay server stopped")
	}()

	logger.Info(">> Relay server started successfully")

	// Create and start gRPC server
	grpcServer, listener, err := setupGRPCServer(grpcdatastreamservice.NewDefaultDatastreamFactory())
	if err != nil {
		logger.Error(">> Failed to setup gRPC server: %v", err)
		os.Exit(1)
	}
	defer listener.Close()

	go func() {
		logger.Info(">> gRPC server starting on port %d", config.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error(">> gRPC server failed to serve: %v", err)
			os.Exit(1)
		}
	}()

	defer func() {
		grpcServer.GracefulStop()
		logger.Info(">> gRPC server stopped")
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go handleSignals(sigChan)

	// Wait for shutdown signal
	<-sigChan

}
