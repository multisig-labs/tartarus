package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jxskiss/mcli"
	"github.com/multisig-labs/tartarus/models"
)

func main() {
	var args struct {
		Input    string `cli:"-i, --input, input JSON file containing nodes" default:"nodes.json"`
		Output   string `cli:"-o, --output, output directory for staking files" default:"staking-dirs"`
		Verbose  bool   `cli:"-v, --verbose, verbose output" default:"false"`
	}

	mcli.Parse(&args)

	// Read the JSON file
	jsonFile, err := os.ReadFile(args.Input)
	if err != nil {
		panic(fmt.Sprintf("Error reading input file: %v", err))
	}

	var nodeMap struct {
		Nodes []models.Node `json:"nodes"`
	}

	if err := json.Unmarshal(jsonFile, &nodeMap); err != nil {
		panic(fmt.Sprintf("Error parsing JSON: %v", err))
	}

	// Create the main output directory
	if err := os.MkdirAll(args.Output, 0755); err != nil {
		panic(fmt.Sprintf("Error creating output directory: %v", err))
	}

	// Process each node
	for _, node := range nodeMap.Nodes {
		nodeDir := filepath.Join(args.Output, node.NodeID)
		if args.Verbose {
			fmt.Printf("Creating staking directory for node: %s\n", node.NodeID)
		}

		// Create directory for this node
		if err := os.MkdirAll(nodeDir, 0755); err != nil {
			panic(fmt.Sprintf("Error creating node directory: %v", err))
		}

		// Write staker.crt
		if err := os.WriteFile(filepath.Join(nodeDir, "staker.crt"), []byte(node.Cert), 0644); err != nil {
			panic(fmt.Sprintf("Error writing staker.crt: %v", err))
		}

		// Write staker.key
		if err := os.WriteFile(filepath.Join(nodeDir, "staker.key"), []byte(node.Key), 0644); err != nil {
			panic(fmt.Sprintf("Error writing staker.key: %v", err))
		}

		// Decode and write signer.key
		blsPrivateBytes, err := hex.DecodeString(node.BLSPrivateKey)
		if err != nil {
			panic(fmt.Sprintf("Error decoding BLS private key: %v", err))
		}

		if err := os.WriteFile(filepath.Join(nodeDir, "signer.key"), blsPrivateBytes, 0644); err != nil {
			panic(fmt.Sprintf("Error writing signer.key: %v", err))
		}
	}

	fmt.Printf("Successfully created staking directories for %d nodes in: %s\n", len(nodeMap.Nodes), args.Output)
}
