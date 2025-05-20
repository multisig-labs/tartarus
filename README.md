# Tartarus: Node ID Generator for Avalanche Nodes

Tartarus is a NodeID and BLS Key generator for Avalanche Nodes. It generates node IDs with customizable prefixes and suffixes, and saves the generated nodes to a CSV file, a JSON file, or a directory that's compatible with the AvalancheGo Node.

## Features

- Easy to use!
- Customizable prefixes and suffixes!
- Make one or a million nodes with a single command!
- Output as CSV, JSON, or an AvalancheGo-compatible directory!
- Multithreaded for faster generation!
- Upload nodes to Supabase backend for hardware providers

## Prerequisites

- Go 1.22 or later
- A C compiler (GCC or Clang)
- `curl` and `jq` command-line tools

### Installing on Ubuntu

Here's a oneline command to install Go 1.23 on Ubuntu:

```sh
sudo rm -rf /usr/local/go && wget -qO- https://golang.org/dl/go1.23.2.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf - && echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a $HOME/.bashrc && source $HOME/.bashrc && rm -f go1.23.2.linux-amd64.tar.gz
```

You'll also need a C compiler, you can install GCC with the following command:

```sh
sudo apt-get update && sudo apt-get install -y build-essential
```

## Getting Started

### Step 1: Sign Up

1. First, you'll need to sign up for an account using the signup script:

```bash
./scripts/signup.sh
```

2. The script will prompt you for:
   - Your email address
   - A password (make sure to use a unique password not used elsewhere)
   - Password confirmation

3. After signing up, you'll receive a verification email. You must verify your email address before proceeding.

### Step 2: Get Your Hardware Provider ID

Before you can generate and upload keys, you need a Hardware Provider ID. This ID will be assigned to you by the system administrators. Please contact the administrators to get your Hardware Provider ID.

### Step 3: Build Tartarus

Once you have the prerequisites installed, you can build Tartarus with the following command:

```sh
go build -o tartarus main.go
```

## Usage

### Generating Node Keys

You can customize the output format and number of nodes:

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

### Converting `nodes.json` to Staking Keys

You can also generate staking keys (avalanchego format) from `nodes.json` with:

```
go run cmd/convert/main.go -i nodes.json -o staking-dir
```

### Uploading Node Keys

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
- `-L, --l1-id`: The L1 ID to associate with the uploaded nodes
- `--supabase-url`: The base URL for the Supabase API
- `--supabase-anon-key`: The public anonymous key for the Supabase project

### Input Data File Format

The `--data-file` should be a JSON file containing an array of node objects under a top-level "nodes" key. Each node object should have at least the following fields:

```json
{
  "nodes": [
    {
      "node_id": "NodeID-xxxxxxxxxxxxxxxxxxxxxx",
      "cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
      "key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
      "bls_private": "aabbccdd...",
      "bls_public": "04aabbcc...",
      "bls_signature": "06aabbcc..."
    }
  ]
}
```

By default, Tartarus **will not** include the BLS private key, the staker cert, or the staker key in the output file. You can include these secrets by using the `--include-secrets` flag, however we do not recommend this and instead recommend you back up the keys yourself to a secure location.

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

## Authentication & Caching

When you first use the `upload` command, you'll be prompted for your email and password (if not provided via flags) to authenticate with Supabase. Upon successful authentication, an access token and your user ID are cached locally in a file named `ggp_api.json` in the current working directory. Subsequent commands will attempt to use this cached token unless `--force-reauth` is specified or the token is invalid/expired.

## Troubleshooting

If you encounter any issues:

1. Make sure you've verified your email address
2. Confirm you have the correct Hardware Provider ID
3. Check that your JSON file is properly formatted
4. Ensure you have the required permissions for the output directory

For additional help, please contact the system administrators.
