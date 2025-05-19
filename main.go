package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jxskiss/mcli"
	"github.com/multisig-labs/tartarus/models"
	"github.com/multisig-labs/tartarus/node"
	"golang.org/x/term"
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

// Arguments for the generate command
type GenerateArgs struct {
	Count          int    `cli:"-n, --count, number of nodes to generate" default:"1"`
	Prefix         string `cli:"-p, --prefix, prefix for the node ID" default:""`
	Suffix         string `cli:"-s, --suffix, suffix for the node ID" default:""`
	CaseSensitive  bool   `cli:"-c, --case-sensitive, case sensitive node ID" default:"false"`
	Output         string `cli:"-o, --output, output file for the nodes" default:"nodes.csv"`
	Verbose        bool   `cli:"-v, --verbose, verbose output" default:"false"`
	ActiveProvider string `cli:"-a, --active-provider, active provider for the node" default:""`
	Threads        int    `cli:"-t, --threads, number of concurrent threads" default:"-1"`
}

// runGenerateCommand contains the original logic of the main function.
func runGenerateCommand(args *GenerateArgs) {
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

// --- Upload Command Functionality ---

// UploadArgs defines the arguments for the 'upload' subcommand.
type UploadArgs struct {
	DataFile           string `cli:"-d, --data-file, Path to the JSON file containing nodes (e.g., nodes.json)"`
	Email              string `cli:"-e, --email, Email for authentication (will prompt if not provided and no cached token)"`
	Password           string `cli:"--password, Password for authentication (will prompt securely if not provided via this flag or email)"`
	ForceReauth        bool   `cli:"--force-reauth, Force re-authentication even if a cached token exists"`
	SupabaseURL        string `cli:"--supabase-url, Supabase API URL" default:"https://sstqretxgcehhfbdjwcz.supabase.co"`
	SupabaseAnonKey    string `cli:"--supabase-anon-key, Supabase Anon Key" default:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNzdHFyZXR4Z2NlaGhmYmRqd2N6Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3MTk0MjQ0MzksImV4cCI6MjAzNTAwMDQzOX0.xIab8CXMlXf7SzsoW1DieuAkDI5GOIAwD9uA1z7Zz9k"`
	HardwareProviderID int    `cli:"--hp-id, Hardware Provider ID (integer, required)"`
	L1ID               string `cli:"-L, --l1-id, L1 ID for the node (optional, defaults to empty string)" default:""`
	Network            string `cli:"--network, Network for the nodes (e.g., fuji, mainnet)" default:"fuji"`
	IncludeSecrets     bool   `cli:"--include-secrets, Include staker cert, staker key, and BLS private key in the upload"`
	BatchSize          int    `cli:"--batch-size, Number of nodes to upload in each batch" default:"25"`
}

const (
	jwtCacheFile     = "ggp_api.json" // In current working directory
	authEndpointPath = "/auth/v1/token?grant_type=password"
	nodesTablePath   = "/rest/v1/nodes" // New path for direct table insert
)

// NodeTableInsertPayload is the structure for nodes when inserting directly into the table.
// This replaces NodeForUpload and NodesUploadPayload for the new system.
type NodeTableInsertPayload struct {
	NodeID             string `json:"node_id"`
	BLSPublicKey       string `json:"bls_public_key"`
	BLSSignature       string `json:"bls_signature"`
	HardwareProviderID int    `json:"hardware_provider_id"`
	UserID             string `json:"user_id"`
	HWStatus           string `json:"hw_status,omitempty"`
	NodeState          string `json:"node_state,omitempty"`
	L1ID               string `json:"l1_id,omitempty"` // Make L1ID omitempty as well if it can be truly optional
	Network            string `json:"network"`
	StakerCert         string `json:"staker_cert,omitempty"`     // Corresponds to models.Node.Cert
	StakerKey          string `json:"staker_key,omitempty"`      // Corresponds to models.Node.Key
	BLSPrivateKey      string `json:"bls_private_key,omitempty"` // Corresponds to models.Node.BLSPrivateKey
}

func promptForEmail() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your email: ")
	email, err := reader.ReadString(byte('\n'))
	if err != nil {
		return "", fmt.Errorf("failed to read email: %w", err)
	}
	return strings.TrimSpace(email), nil
}

func promptForPassword(email string) (string, error) {
	fmt.Printf("Enter password for %s: ", email)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Newline after password input
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(bytePassword), nil
}

func getOrRequestAccessToken(args *UploadArgs) (accessToken string, userID string, err error) {
	// Attempt to load from cache first, unless forced re-auth
	if !args.ForceReauth {
		if _, err := os.Stat(jwtCacheFile); err == nil {
			cachedData, errFile := os.ReadFile(jwtCacheFile)
			if errFile == nil {
				var tokenData struct { // Temp struct for parsing to get both token and user_id if stored
					AccessToken string `json:"access_token"`
					UserID      string `json:"user_id,omitempty"` // If we decide to cache user_id too
				}
				if json.Unmarshal(cachedData, &tokenData) == nil && tokenData.AccessToken != "" && tokenData.UserID != "" {
					fmt.Println("Using cached access token and user ID.")
					return tokenData.AccessToken, tokenData.UserID, nil
				}
			}
		}
	}

	fmt.Println("Attempting to authenticate...")
	email := args.Email
	var localErr error // Renamed to avoid conflict with return 'err'
	if email == "" {
		email, localErr = promptForEmail()
		if localErr != nil {
			return "", "", localErr
		}
	}

	password := args.Password
	if password == "" {
		password, localErr = promptForPassword(email)
		if localErr != nil {
			return "", "", localErr
		}
	}

	authURL := args.SupabaseURL + authEndpointPath
	requestBodyMap := map[string]string{"email": email, "password": password}
	requestBodyBytes, localErr := json.Marshal(requestBodyMap)
	if localErr != nil {
		return "", "", fmt.Errorf("failed to marshal auth request body: %w", localErr)
	}

	req, localErr := http.NewRequest("POST", authURL, bytes.NewBuffer(requestBodyBytes))
	if localErr != nil {
		return "", "", fmt.Errorf("failed to create auth request: %w", localErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", args.SupabaseAnonKey)

	client := &http.Client{Timeout: time.Second * 10}
	resp, localErr := client.Do(req)
	if localErr != nil {
		return "", "", fmt.Errorf("authentication request failed: %w", localErr)
	}
	defer resp.Body.Close()

	respBody, localErr := io.ReadAll(resp.Body)
	if localErr != nil {
		return "", "", fmt.Errorf("failed to read auth response body: %w", localErr)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("authentication failed. Status: %s, Body: %s", resp.Status, string(respBody))
	}

	var authResponse struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if localErr = json.Unmarshal(respBody, &authResponse); localErr != nil {
		return "", "", fmt.Errorf("failed to parse auth response JSON: %w. Body: %s", localErr, string(respBody))
	}

	if authResponse.AccessToken == "" || authResponse.User.ID == "" {
		return "", "", fmt.Errorf("access token or user ID not found in auth response. Body: %s", string(respBody))
	}

	// Cache the new token and user ID
	tokenToCache := struct {
		AccessToken string `json:"access_token"`
		UserID      string `json:"user_id"`
	}{
		AccessToken: authResponse.AccessToken,
		UserID:      authResponse.User.ID,
	}
	cachedTokenBytes, localErr := json.MarshalIndent(tokenToCache, "", "  ")
	if localErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to marshal token for caching: %v\n", localErr)
	} else {
		if err := os.WriteFile(jwtCacheFile, cachedTokenBytes, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to cache access token and user ID to %s: %v\n", jwtCacheFile, err)
		} else {
			fmt.Printf("Access token and user ID cached successfully to %s.\n", jwtCacheFile)
		}
	}
	return authResponse.AccessToken, authResponse.User.ID, nil
}

// uploadNodesToTable handles processing nodes from a file and uploading them to the Supabase table.
func uploadNodesToTable(args *UploadArgs, accessToken string, authUserID string) error {
	jsonFile, err := os.ReadFile(args.DataFile)
	if err != nil {
		return fmt.Errorf("failed to read data file %s: %w", args.DataFile, err)
	}

	var nodeContainer struct {
		Nodes []models.Node `json:"nodes"`
	}
	if err := json.Unmarshal(jsonFile, &nodeContainer); err != nil {
		return fmt.Errorf("failed to parse JSON from data file %s: %w", args.DataFile, err)
	}

	if len(nodeContainer.Nodes) == 0 {
		fmt.Println("No nodes found in the data file to upload.")
		return nil
	}
	fmt.Printf("Found %d nodes to process from %s.\n", len(nodeContainer.Nodes), args.DataFile)

	var payloadsToUpload []NodeTableInsertPayload
	for _, node := range nodeContainer.Nodes {
		blsPublicKey := node.BLSPublicKey
		if !strings.HasPrefix(blsPublicKey, "0x") {
			blsPublicKey = "0x" + blsPublicKey
		}
		blsSignature := node.BLSSignature
		if !strings.HasPrefix(blsSignature, "0x") {
			blsSignature = "0x" + blsSignature
		}

		payload := NodeTableInsertPayload{
			NodeID:             node.NodeID,
			BLSPublicKey:       blsPublicKey,
			BLSSignature:       blsSignature,
			HardwareProviderID: args.HardwareProviderID,
			UserID:             authUserID,
			HWStatus:           "inactive",
			NodeState:          "available",
			L1ID:               args.L1ID,
			Network:            args.Network,
		}

		if args.IncludeSecrets {
			payload.StakerCert = node.Cert
			payload.StakerKey = node.Key
			payload.BLSPrivateKey = node.BLSPrivateKey
		}

		payloadsToUpload = append(payloadsToUpload, payload)
	}

	totalNodes := len(payloadsToUpload)
	uploadURL := args.SupabaseURL + nodesTablePath
	client := &http.Client{Timeout: time.Second * 30}

	fmt.Printf("Total nodes to upload: %d. Will process in batches of %d.\n", totalNodes, args.BatchSize)

	for i := 0; i < totalNodes; i += args.BatchSize {
		batchNum := (i / args.BatchSize) + 1
		end := i + args.BatchSize
		if end > totalNodes {
			end = totalNodes
		}
		currentBatchPayloads := payloadsToUpload[i:end]

		payloadBytes, err := json.Marshal(currentBatchPayloads)
		if err != nil {
			return fmt.Errorf("failed to marshal batch %d payload: %w", batchNum, err)
		}

		fmt.Printf("POSTing batch %d of %d (%d nodes) to %s...\n", batchNum, (totalNodes+args.BatchSize-1)/args.BatchSize, len(currentBatchPayloads), uploadURL)

		req, err := http.NewRequest("POST", uploadURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return fmt.Errorf("failed to create request for batch %d: %w", batchNum, err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("apikey", args.SupabaseAnonKey)
		req.Header.Set("Prefer", "return=representation") // Optional: get inserted data back

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send batch %d: %w", batchNum, err)
		}

		respBodyBytes, ioErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if ioErr != nil {
			fmt.Fprintf(os.Stderr, "Error reading response body for batch %d: %v\n", batchNum, ioErr)
		}

		fmt.Printf("Batch %d response status: %s\n", batchNum, resp.Status)
		if len(respBodyBytes) > 0 {
			// Attempt to pretty-print if JSON, otherwise print as string
			var prettyJSON bytes.Buffer
			if json.Indent(&prettyJSON, respBodyBytes, "", "  ") == nil {
				fmt.Printf("Batch %d response body:\n%s\n", batchNum, prettyJSON.String())
			} else {
				fmt.Printf("Batch %d response body: %s\n", batchNum, string(respBodyBytes))
			}
		}

		batchResponseFile := fmt.Sprintf("upload_nodes_response_batch_%d.json", batchNum)
		err = os.WriteFile(batchResponseFile, respBodyBytes, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save batch %d response to %s: %v\n", batchNum, batchResponseFile, err)
		} else {
			fmt.Printf("Batch %d response saved to %s\n", batchNum, batchResponseFile)
		}

		// For PostgREST, 201 Created is typical for successful inserts.
		if resp.StatusCode != http.StatusCreated {
			fmt.Fprintf(os.Stderr, "Error processing batch %d: Status %s. See %s for details.\n", batchNum, resp.Status, batchResponseFile)
			// Decide if to continue or bail out. For now, continue.
		}
	}

	fmt.Println("All batches processed.")
	return nil
}

// runUploadCommand is the handler for the "upload" subcommand.
func runUploadCommand() {
	var uploadArgs UploadArgs
	_, parseErr := mcli.Parse(&uploadArgs)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Error parsing upload command arguments: %v\n", parseErr)
		os.Exit(1)
	}

	if uploadArgs.DataFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --data-file flag is required.")
		mcli.PrintHelp()
		os.Exit(1)
	}
	if uploadArgs.HardwareProviderID == 0 { // Integer default is 0, check if it was explicitly set.
		fmt.Fprintln(os.Stderr, "Error: --hp-id (Hardware Provider ID) flag is required and must be non-zero.")
		mcli.PrintHelp()
		os.Exit(1)
	}

	accessToken, authUserID, err := getOrRequestAccessToken(&uploadArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting access token: %v\n", err)
		os.Exit(1)
	}

	err = uploadNodesToTable(&uploadArgs, accessToken, authUserID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing or uploading nodes: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Upload process completed.")
}

func main() {
	// Define a variable to hold arguments for the generate command (root command)
	var generateArgs GenerateArgs

	// Add the root command (current functionality)
	mcli.AddRoot(func() {
		mcli.Parse(&generateArgs) // Parse arguments for the root command
		runGenerateCommand(&generateArgs)
	})

	// Add the 'upload' subcommand
	mcli.Add("upload", runUploadCommand, "Uploads node information using cached or prompted credentials.")

	// Run the CLI application
	mcli.Run()
}
