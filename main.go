package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/oasislabs/oasis-core/go/common/logging"

	"github.com/oasislabs/oasis-core-rosetta-gateway/oasis-client"
	"github.com/oasislabs/oasis-core-rosetta-gateway/services"
)

// GatewayPortEnvVar is the name of the environment variable that specifies
// which port the Oasis Rosetta gateway should run on.
const GatewayPortEnvVar = "OASIS_ROSETTA_GATEWAY_PORT"

var logger = logging.GetLogger("oasis-rosetta-gateway")

// NewBlockchainRouter returns a Mux http.Handler from a collection of
// Rosetta service controllers.
func NewBlockchainRouter(oasisClient oasis_client.OasisClient) http.Handler {
	networkAPIController := server.NewNetworkAPIController(services.NewNetworkAPIService(oasisClient))
	accountAPIController := server.NewAccountAPIController(services.NewAccountAPIService(oasisClient))
	blockAPIController := server.NewBlockAPIController(services.NewBlockAPIService(oasisClient))
	constructionAPIController := server.NewConstructionAPIController(services.NewConstructionAPIService(oasisClient))

	return server.NewRouter(networkAPIController, accountAPIController, blockAPIController, constructionAPIController)
}

func main() {
	// Get server port from environment variable or use the default.
	port := os.Getenv(GatewayPortEnvVar)
	if port == "" {
		port = "8080"
	}
	nPort, err := strconv.Atoi(port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Malformed %s environment variable: %v\n", GatewayPortEnvVar, err)
		os.Exit(1)
	}

	// Prepare a new Oasis gRPC client.
	oasisClient, err := oasis_client.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to prepare Oasis gRPC client: %v\n", err)
		os.Exit(1)
	}

	// Make a test request using the client to see if the node works.
	cid, err := oasisClient.GetChainID(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Node connectivity error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging.
	err = logging.Initialize(os.Stdout, logging.FmtLogfmt, logging.LevelDebug, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to initialize logging: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Connected to Oasis node", "chain_context", cid)

	// Start the server.
	router := NewBlockchainRouter(oasisClient)
	logger.Info("Oasis Rosetta Gateway listening", "port", nPort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", nPort), router)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Oasis Rosetta Gateway server exited with error: %v\n", err)
		os.Exit(1)
	}
}