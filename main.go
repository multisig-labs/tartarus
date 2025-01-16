package main

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jxskiss/mcli"
	"github.com/multisig-labs/tartarus/models"
	"github.com/multisig-labs/tartarus/node"
)

// Worker function to generate nodes
func generateNode(prefix, suffix string, caseSensitive bool, activeProvider string, resultChan chan models.Node, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for {
		select {
		case <-done:
			return
		default:
			n, err := node.Generate()
			if err != nil {
				continue
			}

			nodeID := strings.Replace(n.NodeID, "NodeID-", "", -1)
			if !caseSensitive {
				nodeID = strings.ToLower(nodeID)
			}

			if strings.HasPrefix(nodeID, prefix) && strings.HasSuffix(nodeID, suffix) {
				if activeProvider != "" {
					n.ActiveProvider = activeProvider
				}
				select {
				case resultChan <- n:
					return
				case <-done:
					return
				}
			}
		}
	}
}

func main() {
	var args struct {
		Count          int    `cli:"-n, --count, number of nodes to generate" default:"1"`
		Prefix         string `cli:"-p, --prefix, prefix for the node ID" default:""`
		Suffix         string `cli:"-s, --suffix, suffix for the node ID" default:""`
		CaseSensitive  bool   `cli:"-c, --case-sensitive, case sensitive node ID" default:"false"`
		Output         string `cli:"-o, --output, output file for the nodes" default:"nodes.csv"`
		Verbose        bool   `cli:"-v, --verbose, verbose output" default:"false"`
		ActiveProvider string `cli:"-a, --active-provider, active provider for the node" default:""`
		Threads        int    `cli:"-t, --threads, number of concurrent threads" default:"-1"`
	}

	mcli.Parse(&args)

	if args.Verbose {
		fmt.Println("Generating", args.Count, "nodes with prefix:", args.Prefix, "and suffix:", args.Suffix)
	}

	if !args.CaseSensitive {
		args.Prefix = strings.ToLower(args.Prefix)
		args.Suffix = strings.ToLower(args.Suffix)
	}

	nodes := []models.Node{}
	numWorkers := runtime.NumCPU() // Default to number of CPUs
	if args.Threads > 0 {
		numWorkers = args.Threads
	}

	if args.Verbose {
		fmt.Printf("Using %d worker threads\n", numWorkers)
	}

	resultChan := make(chan models.Node)
	progressTicker := time.NewTicker(1 * time.Second)
	defer progressTicker.Stop()

	for i := 0; i < args.Count; i++ {
		var wg sync.WaitGroup
		done := make(chan struct{})
		found := false

		// Start workers
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go generateNode(args.Prefix, args.Suffix, args.CaseSensitive, args.ActiveProvider, resultChan, done, &wg)
		}

		// Wait for result or print progress
		for !found {
			select {
			case n := <-resultChan:
				found = true
				nodes = append(nodes, n)
				fmt.Println("\nGenerated node:", n.NodeID)
				close(done) // Signal all workers to stop
			case <-progressTicker.C:
				fmt.Print(".")
			}
		}

		// Wait for all workers to finish
		wg.Wait()
	}

	if strings.HasSuffix(args.Output, ".csv") {
		// save the nodes to a csv file
		f, err := os.Create(args.Output)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		w := csv.NewWriter(f)

		header := []string{"nodeID", "cert", "key", "bls_private", "bls_public", "bls_signature", "active_provider"}
		if err := w.Write(header); err != nil {
			panic(err)
		}

		for _, n := range nodes {
			record := []string{n.NodeID, n.Cert, n.Key, n.BLSPrivateKey, n.BLSPublicKey, n.BLSSignature, n.ActiveProvider}
			if err := w.Write(record); err != nil {
				panic(err)
			}
		}

		w.Flush()

		fmt.Println("Nodes saved to:", args.Output)
	} else if strings.HasSuffix(args.Output, ".json") {
		// save the nodes to a json file
		f, err := os.Create(args.Output)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		// write a json file
		nodeMap := map[string][]models.Node{
			"nodes": nodes,
		}

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(nodeMap); err != nil {
			panic(err)
		}

		fmt.Println("Nodes saved to:", args.Output)
	} else if !strings.Contains(args.Output, ".") && args.Count == 1 {
		// make the staking directory
		if err := os.MkdirAll(args.Output, 0755); err != nil {
			panic(err)
		}

		// write the cert and key files. Dump the strings to the files
		certFile, err := os.Create(filepath.Join(args.Output, "staker.crt"))
		if err != nil {
			panic(err)
		}
		defer certFile.Close()

		if _, err := certFile.WriteString(nodes[0].Cert); err != nil {
			panic(err)
		}

		keyFile, err := os.Create(filepath.Join(args.Output, "staker.key"))
		if err != nil {
			panic(err)
		}

		defer keyFile.Close()

		if _, err := keyFile.WriteString(nodes[0].Key); err != nil {
			panic(err)
		}

		// for the bls private key, encode the hex to bytes
		blsPrivateBytes, err := hex.DecodeString(nodes[0].BLSPrivateKey)
		if err != nil {
			panic(err)
		}

		// write the raw bytes to signer.key
		blsKeyFile, err := os.Create(filepath.Join(args.Output, "signer.key"))
		if err != nil {
			panic(err)
		}

		defer blsKeyFile.Close()

		if _, err := blsKeyFile.Write(blsPrivateBytes); err != nil {
			panic(err)
		}

		fmt.Println("Staking files saved to:", args.Output)
	} else {
		fmt.Println("Unsupported output format:", args.Output)
	}
}
