package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"testing"

	zkevmLog "github.com/0xPolygonHermez/zkevm-data-streamer/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Initialize logging for tests
	zkevmLog.Init(zkevmLog.Config{
		Environment: "test",
		Level:       "debug",
		Outputs:     []string{"stdout"},
	})

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestFlagParsing(t *testing.T) {
	// Save original args and config
	originalArgs := os.Args
	originalConfig := config
	defer func() {
		os.Args = originalArgs
		config = originalConfig
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset flags
	}()

	// Set test arguments
	os.Args = []string{
		"test",
		"-server=127.0.0.1:6900",
		"-relay-port=7900",
		"-grpc-port=7070",
		"-datafile=test.dat",
		"-log=debug",
		"-writetimeout=3000",
		"-inactivitytimeout=120",
	}

	// Parse flags
	parseFlags()

	// Verify parsed values
	assert.Equal(t, "127.0.0.1:6900", config.ServerAddr)
	assert.Equal(t, uint(7900), config.RelayPort)
	assert.Equal(t, uint(7070), config.GRPCPort)
	assert.Equal(t, "test.dat", config.DataFile)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, uint(3000), config.WriteTimeoutMs)
	assert.Equal(t, uint(120), config.InactivityTimeout)
}

func TestGRPCServerSetup(t *testing.T) {
	// Save and restore config
	originalConfig := config
	defer func() { config = originalConfig }()

	// Set test configuration
	testConfig := &RelayConfig{
		GRPCPort: 7070,
	}
	config = testConfig

	server, listener, err := setupGRPCServer()
	require.NoError(t, err)
	require.NotNil(t, server)
	require.NotNil(t, listener)
	defer listener.Close()

	serviceInfo := server.GetServiceInfo()
	require.NotEmpty(t, serviceInfo)
}

func TestSignalHandling(t *testing.T) {
	// Create a channel for signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start signal handler in a goroutine
	go handleSignals(sigChan)

	// Send a signal
	sigChan <- syscall.SIGINT

	// If we get here without hanging, the test passed
}
