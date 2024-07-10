package main

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jxskiss/mcli"
	"github.com/multisig-labs/tartarus/models"
	"github.com/multisig-labs/tartarus/node"
)

func main() {
	var args struct {
		Count         int    `cli:"-n, --count, number of nodes to generate" default:"1"`
		Prefix        string `cli:"-p, --prefix, prefix for the node ID" default:""`
		Suffix        string `cli:"-s, --suffix, suffix for the node ID" default:""`
		CaseSensitive bool   `cli:"-c, --case-sensitive, case sensitive node ID" default:"false"`
		Output        string `cli:"-o, --output, output file for the nodes" default:"nodes.csv"`
		Verbose       bool   `cli:"-v, --verbose, verbose output" default:"false"`
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

	for i := 0; i < args.Count; i++ {
		found := false
		attemptCounter := 0
		for !found {
			n, err := node.Generate()
			if err != nil {
				panic(err)
			}

			nodeID := strings.Replace(n.NodeID, "NodeID-", "", -1)

			if !args.CaseSensitive {
				nodeID = strings.ToLower(nodeID)
			}

			if strings.HasPrefix(nodeID, args.Prefix) && strings.HasSuffix(nodeID, args.Suffix) {
				nodes = append(nodes, n)
				found = true
				if attemptCounter > 0 {
					fmt.Println()
				}
				fmt.Println("Generated node:", n.NodeID)
			}
			attemptCounter += 1
			if attemptCounter > 1000 {
				fmt.Print(".")
				attemptCounter = 0
			}
		}
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
