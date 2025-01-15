#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# =============================
#       Configuration
# =============================
SUPABASE_URL="https://glahotetihpffpvaxvul.supabase.co"
# anon key, this is fine to distribute
SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImdsYWhvdGV0aWhwZmZwdmF4dnVsIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MDU5NjQzODEsImV4cCI6MjAyMTU0MDM4MX0.jFNd-pmq4U57vL8bYi3WCjuzzIWfq_Q3lyIIH4XGpRg"

# =============================
#      Helper Functions
# =============================

# Function to display error messages
error_exit() {
    echo "❌ Error: $1" >&2
    exit 1
}

# Function to read user input securely for passwords
read_password() {
    local prompt="$1"
    local var
    read -s -p "$prompt: " var
    echo "$var"
}

# =============================
#        User Input
# =============================

echo "GGP Node Manager User Signup"
echo "======================"

# Read user email
read -p "Enter your email: " USER_EMAIL

echo "WARNING: Password should be unique and not used anywhere else."

# Read user password
PASSWORD1=$(read_password "Enter your password")
echo

# Read password confirmation
PASSWORD2=$(read_password "Confirm your password")
echo

# Validate that passwords match
if [[ "$PASSWORD1" != "$PASSWORD2" ]]; then
    error_exit "Passwords do not match. Please try again."
fi

# =============================
#      Signup Process
# =============================

# Define the signup endpoint
SIGNUP_ENDPOINT="${SUPABASE_URL}/auth/v1/signup"

# Create JSON payload
read -r -d '' PAYLOAD <<EOF
{
    "email": "$USER_EMAIL",
    "password": "$PASSWORD1"
}
EOF

# Make the POST request to sign up the user
response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$SIGNUP_ENDPOINT" \
    -H "apikey: $SUPABASE_ANON_KEY" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD")

# Extract body and status
body=$(echo "$response" | sed -e 's/HTTP_STATUS\:.*//g')
status=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTP_STATUS://')

# Check the HTTP status code
if [[ "$status" -ge 200 && "$status" -lt 300 ]]; then
    echo "User signed up successfully. Please check your email to verify your account."
    echo "Response:"
    echo "$body" | jq
else
    echo "❌ Failed to sign up user. HTTP Status: $status"
    echo "Response:"
    echo "$body" | jq
    exit 1
fi
