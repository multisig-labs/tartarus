#!/bin/bash

# =====================================================
# Script: post_nodes.sh
# Description: 
#   1. Authenticates to DB using email and password.
#   2. Saves the JWT to ~/.ggp_api.json.
#   3. Uses the JWT to authenticate and call a 
#      Edge Function, posting data from a specified JSON file.
# =====================================================

# -----------------------------
# Define Default Parameters
# -----------------------------
DATA_FILE=""
CHAIN_ID="43114"  # default chain_id

# -----------------------------
# Function: Display Usage
# -----------------------------
usage() {
    echo "Usage: $0 -d /path/to/data.json [-C chain_id]"
    echo
    echo "Options:"
    echo "  -d, --data      Path to the JSON file containing data to send to the Edge Function. (Required)"
    echo "  -C, --chain-id  Chain id to append as query parameter (default: 43114)"
    echo "  -h, --help      Display this help message."
    echo
    echo "Example:"
    echo "  $0 -d ~/projects/nodes_data.json -C 43114"
    exit 1
}

# -----------------------------
# Parse Command-Line Arguments
# -----------------------------
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -d|--data)
            DATA_FILE="$2"
            shift 2
            ;;
        -C|--chain-id)
            CHAIN_ID="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown parameter passed: $1"
            usage
            ;;
    esac
done

# Check if DATA_FILE is provided
if [[ -z "$DATA_FILE" ]]; then
    echo "Error: Data file not specified."
    usage
fi

# Check if DATA_FILE exists and is a file
if [[ ! -f "$DATA_FILE" ]]; then
    echo "Error: Data file '$DATA_FILE' does not exist or is not a file."
    exit 1
fi

# Validate that DATA_FILE contains valid JSON
if ! jq empty "$DATA_FILE" 2>/dev/null; then
    echo "Error: Data file '$DATA_FILE' contains invalid JSON."
    exit 1
fi


SUPABASE_URL="https://glahotetihpffpvaxvul.supabase.co"
# SUPABASE_URL="http://127.0.0.1:54321"
# anon key, this is fine to distribute
SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImdsYWhvdGV0aWhwZmZwdmF4dnVsIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MDU5NjQzODEsImV4cCI6MjAyMTU0MDM4MX0.jFNd-pmq4U57vL8bYi3WCjuzzIWfq_Q3lyIIH4XGpRg"
# SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0"
# Append the chain_id as a query parameter to the EDGE_FUNCTION_URL.
EDGE_FUNCTION_URL="$SUPABASE_URL/functions/v1/add_node_keys?chain_id=$CHAIN_ID"

# -----------------------------
# Obtain Email and Password Securely
# -----------------------------

read -p "Enter your email: " EMAIL

# Validate that EMAIL is not empty
if [[ -z "$EMAIL" ]]; then
    echo "Error: Email cannot be empty."
    exit 1
fi

read -s -p "Enter password for $EMAIL: " PASSWORD
echo

# -----------------------------
# Define Output File
# -----------------------------

JWT_FILE="$HOME/.ggp_api.json"

# -----------------------------
# Check if jq is installed
# -----------------------------

if ! command -v jq &> /dev/null
then
    echo "Error: 'jq' is not installed. Please install jq to proceed."
    exit 1
fi

# -----------------------------
# Perform Authentication Request
# -----------------------------

echo "Authenticating with GGP DB..."

AUTH_RESPONSE=$(curl -s -X POST "$SUPABASE_URL/auth/v1/token?grant_type=password" \
  -H "Content-Type: application/json" \
  -H "apikey: $SUPABASE_ANON_KEY" \
  -d "{\"email\":\"$EMAIL\", \"password\":\"$PASSWORD\"}")

# -----------------------------
# Parse the Authentication Response
# -----------------------------

ACCESS_TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.access_token')

if [[ "$ACCESS_TOKEN" != "null" ]] && [[ -n "$ACCESS_TOKEN" ]]; then
    echo "{\"access_token\": \"$ACCESS_TOKEN\"}" > "$JWT_FILE"
    echo "JWT successfully saved to $JWT_FILE"
else
    ERROR_MESSAGE=$(echo "$AUTH_RESPONSE" | jq -r '.msg // "Unknown error occurred during authentication."')
    echo "Authentication failed: $ERROR_MESSAGE"
    exit 1
fi

# -----------------------------
# Transform and Batch Process Data
# -----------------------------

echo "Transforming data..."

# Get all nodes first
ALL_NODES=$(jq '{nodes: [.nodes[] | {
    node_id,
    bls_public_key: (if .bls_public | test("^0x") then .bls_public else "0x"+.bls_public end),
    bls_signature: (if .bls_signature | test("^0x") then .bls_signature else "0x"+.bls_signature end)
}]}' "$DATA_FILE")

TOTAL_NODES=$(echo "$ALL_NODES" | jq '.nodes | length')
BATCH_SIZE=25
TOTAL_BATCHES=$(( (TOTAL_NODES + BATCH_SIZE - 1) / BATCH_SIZE ))

echo "Total nodes to process: $TOTAL_NODES"
echo "Will process in batches of $BATCH_SIZE nodes"
echo "Total number of batches: $TOTAL_BATCHES"

for ((i=0; i<TOTAL_NODES; i+=BATCH_SIZE)); do
    BATCH_NUM=$(( (i / BATCH_SIZE) + 1 ))
    echo "Processing batch $BATCH_NUM of $TOTAL_BATCHES..."
    
    # Extract current batch of nodes
    BATCH_DATA=$(echo "$ALL_NODES" | jq "{nodes: .nodes[$i:$(($i+$BATCH_SIZE))]}")

    # Optional: Validate the batch data
    if ! echo "$BATCH_DATA" | jq empty 2>/dev/null; then
        echo "Error: Batch data is invalid JSON."
        exit 1
    fi

    echo "POSTing batch of nodes..."

    EDGE_RESPONSE=$(curl -s -X POST "$EDGE_FUNCTION_URL" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -d "$BATCH_DATA")

    echo "Batch $BATCH_NUM response:"
    echo "$EDGE_RESPONSE"

    # Save each batch response
    BATCH_RESPONSE_FILE="post_nodes_response_batch_${BATCH_NUM}.json"
    echo "$EDGE_RESPONSE" > "$BATCH_RESPONSE_FILE"
    echo "Batch $BATCH_NUM response saved to $BATCH_RESPONSE_FILE"
done

echo "All batches processed successfully!"
