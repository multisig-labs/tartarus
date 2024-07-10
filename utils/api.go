package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func VerifyBLSViaAPI(nodeID, publicKey, proofOfPossession string) (bool, error) {
	url := "https://api.gogopool.com/bls/pop"

	// Create the request body
	requestBody := map[string]interface{}{
		"nodeID": nodeID,
		"nodePOP": map[string]string{
			"publicKey":         publicKey,
			"proofOfPossession": proofOfPossession,
		},
	}

	// Convert the request body to JSON
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return false, err
	}

	// Send the HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// Parse the response JSON
	var response map[string]interface{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return false, err
	}

	// Check the response status
	status, ok := response["valid"].(bool)
	if !ok {
		return false, fmt.Errorf("invalid response format")
	}

	return status, nil
}
