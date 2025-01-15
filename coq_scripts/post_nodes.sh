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
# Function: Display Usage
# -----------------------------
usage() {
    echo "Usage: $0 -d /path/to/data.json"
    echo
    echo "Options:"
    echo "  -d, --data      Path to the JSON file containing data to send to the Edge Function. (Required)"
    echo "  -h, --help      Display this help message."
    echo
    echo "Example:"
    echo "  $0 -d ~/projects/nodes_data.json"
    exit 1
}

# -----------------------------
# Parse Command-Line Arguments
# -----------------------------
DATA_FILE=""

while [[ "$#" -gt 0 ]]; do
    case $1 in
        -d|--data)
            DATA_FILE="$2"
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
# anon key, this is fine to distribute
SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImdsYWhvdGV0aWhwZmZwdmF4dnVsIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MDU5NjQzODEsImV4cCI6MjAyMTU0MDM4MX0.jFNd-pmq4U57vL8bYi3WCjuzzIWfq_Q3lyIIH4XGpRg"
EDGE_FUNCTION_URL="$SUPABASE_URL/functions/v1/add_node_keys"

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
# Transform Data for Edge Function Call
# -----------------------------

echo "Transforming data..."

TRANSFORMED_DATA=$(jq '{nodes: [.nodes[] | {node_id, bls_public_key: .bls_public, bls_signature}]}' "$DATA_FILE")

# Optional: Validate the transformed data
if ! echo "$TRANSFORMED_DATA" | jq empty 2>/dev/null; then
    echo "Error: Transformed data is invalid JSON."
    exit 1
fi

# -----------------------------
# Call the Edge Function
# -----------------------------

echo "POSTing Nodes..."

EDGE_RESPONSE=$(curl -s -X POST "$EDGE_FUNCTION_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "$TRANSFORMED_DATA")

# -----------------------------
# Handle the Edge Function Response
# -----------------------------

echo "Response:"
echo "$EDGE_RESPONSE"

RESPONSE_FILE="$HOME/post_nodes_response.json"
echo "$EDGE_RESPONSE" > "$RESPONSE_FILE"
echo "Response saved to $RESPONSE_FILE"
