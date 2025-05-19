# Tartarus Usage Guide

This guide will walk you through the process of signing up and generating/uploading keys for your Avalanche nodes.

## Prerequisites

- Go 1.22 or later
- A C compiler (GCC or Clang)
- `curl` and `jq` command-line tools

## Step 1: Sign Up

1. First, you'll need to sign up for an account using the signup script:

```bash
./scripts/signup.sh
```

2. The script will prompt you for:
   - Your email address
   - A password (make sure to use a unique password not used elsewhere)
   - Password confirmation

3. After signing up, you'll receive a verification email. You must verify your email address before proceeding.

## Step 2: Get Your Hardware Provider ID

Before you can generate and upload keys, you need a Hardware Provider ID. This ID will be assigned to you by the system administrators. Please contact the administrators to get your Hardware Provider ID.

## Step 3: Generate Node Keys

1. Build the Tartarus program:

```bash
go build -o tartarus main.go
```

2. Generate your node keys using the Tartarus program. You can customize the output format and number of nodes:

```bash
# Generate a single node and save as JSON
./tartarus -n 1 -o nodes.json

# Generate multiple nodes with a specific prefix
./tartarus -n 10 -p "yourprefix" -o nodes.json

# Generate nodes and save as AvalancheGo-compatible directory
./tartarus -n 1 -o staking-dir
```

Available flags:
- `-n, --count`: Number of nodes to generate (default: 1)
- `-p, --prefix`: Prefix for node IDs
- `-s, --suffix`: Suffix for node IDs
- `-c, --case-sensitive`: Make node IDs case-sensitive
- `-o, --output`: Output file/directory (default: "nodes.csv")
- `-v, --verbose`: Enable verbose output

## Step 4: Upload Node Keys

Once you have generated your node keys, you can upload them to the system:

```bash
./tartarus upload -d nodes.json --hp-id YOUR_HARDWARE_PROVIDER_ID
```

Required flags for upload:
- `-d, --data-file`: Path to your JSON file containing node data
- `--hp-id`: Your Hardware Provider ID (assigned by administrators)

Optional flags:
- `-e, --email`: Your email address (will prompt if not provided)
- `--password`: Your password (will prompt securely if not provided)
- `--network`: Network for the nodes (default: "fuji")
- `--include-secrets`: Include staker cert, staker key, and BLS private key in the upload
- `--batch-size`: Number of nodes to upload in each batch (default: 25)

## Output Formats

The program supports three output formats:

1. CSV file (default):
   - Contains node ID, certificate, key, and BLS information
   - Suitable for spreadsheet analysis

2. JSON file:
   - Contains the same information in JSON format
   - Required for uploading to the system

3. AvalancheGo-compatible directory:
   - Creates a directory with `staker.crt`, `staker.key`, and `signer.key` files
   - Ready to use with AvalancheGo nodes

## Security Notes

- Keep your generated keys secure and never share them
- Use a unique password for your account
- The BLS private key is particularly sensitive and should be protected
- Consider using the `--include-secrets` flag only when necessary

## Troubleshooting

If you encounter any issues:

1. Make sure you've verified your email address
2. Confirm you have the correct Hardware Provider ID
3. Check that your JSON file is properly formatted
4. Ensure you have the required permissions for the output directory

For additional help, please contact the system administrators.
