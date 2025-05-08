# Tartarus: Node ID Generator for Avalanche Nodes

Tartarus is a NodeID and BLS Key generator for the Avalanche Nodes. It generates node IDs with customizable prefixes and suffixes, and saves the generated nodes to a CSV file, a JSON file, or a directory that's compatible with the AvalancheGo Node.

## Pros

- Easy to use!
- Customizable prefixes and suffixes!
- Make one or a million nodes with a single command!
- Output as CSV, JSON, or an AvalancheGo-compatible directory!
- Multithreaded for faster generation!

## Cons

- No regex support for prefixes and suffixes.
- Requires a C compiler to build.

## Pre-requisites

- Go 1.23+
- A C compiler (GCC or Clang)

### Installing on Ubuntu

Here's a oneline command to install Go 1.23 on Ubuntu:

```sh
sudo rm -rf /usr/local/go && wget -qO- https://golang.org/dl/go1.23.2.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf - && echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a $HOME/.bashrc && source $HOME/.bashrc && rm -f go1.23.2.linux-amd64.tar.gz
```

You'll also need a C compiler, you can install GCC with the following command:

```sh
sudo apt-get update && sudo apt-get install -y build-essential
```

## Building

Once you have the prerequisites installed, you can build Tartarus with the following command:

```sh
go build -o tartarus main.go
```

## Usage

To use Tartarus, simply run the executable with the desired flags. For example:

```sh
./tartarus -n 10 -p abc -o output.csv
```

This will generate 10 node IDs with the prefix "abc" and save them to a CSV file called "output.csv".

NOTE: THIS WILL TAKE A LONG TIME if you choose long prefixes!

## Flags

- `-n, --count`: The number of nodes to generate. Defaults to 1.
- `-p, --prefix`: The prefix for the node IDs. Defaults to an empty string.
- `-s, --suffix`: The suffix for the node IDs. Defaults to an empty string.
- `-c, --case-sensitive`: Whether to make the node IDs case-sensitive. Defaults to false.
- `-o, --output`: The output file for the generated nodes. Defaults to "nodes.csv".
- `-v, --verbose`: Whether to print verbose output. Defaults to false.

## Upload Subcommand

The `upload` subcommand allows you to upload node data (Node ID, BLS public key, BLS signature) to a Supabase backend. This is typically used by hardware providers to register new nodes.

**Usage Example:**

```sh
./tartarus upload -d nodes_to_upload.json --hp-id 123 -e your_email@example.com
```

This command will attempt to upload nodes from `nodes_to_upload.json`, associating them with hardware provider ID `123`, and will prompt for your password for `your_email@example.com` to authenticate with Supabase.

**Input Data File Format (`--data-file`)**

The `--data-file` should be a JSON file containing an array of node objects under a top-level "nodes" key. Each node object should have at least the following fields (matching the output of the generation command when producing JSON):

```json
{
  "nodes": [
    {
      "node_id": "NodeID-xxxxxxxxxxxxxxxxxxxxxx",
      "cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
      "key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
      "bls_private": "aabbccdd...",
      "bls_public": "04aabbcc...", // Used for bls_public_key
      "bls_signature": "06aabbcc..."  // Used for bls_signature
    }
    // ... more nodes
  ]
}
```

**Upload Flags:**

- `-d, --data-file <file_path>`: Path to the JSON file containing the node data to upload. (Required)
- `--hp-id <id>`: Your integer Hardware Provider ID. This is used to associate the uploaded nodes with your provider account in Supabase. (Required)
- `-e, --email <email>`: Your email address for authenticating with Supabase. If not provided and no cached session exists, you will be prompted.
- `--password <password>`: Your password for Supabase authentication. If not provided with the email flag and no cached session exists, you will be prompted securely.
- `--force-reauth`: Forces re-authentication with Supabase, ignoring any cached session token.
- `-L, --l1-id <id>`: The L1 ID to associate with the uploaded nodes. (Optional, defaults to an empty string).
- `--network <network_name>`: The network these nodes belong to (e.g., `fuji`, `mainnet`, `local`). (Optional, defaults to `fuji`).
- `--include-secrets`: If specified, includes the staker certificate (`staker_cert`), staker key (`staker_key`), and BLS private key (`bls_private_key`) in the upload. By default, these are omitted.
- `--batch-size <count>`: The number of nodes to include in each batch upload to Supabase. (Optional, defaults to 25).
- `--supabase-url <url>`: The base URL for the Supabase API. (Optional, defaults to `https://sstqretxgcehhfbdjwcz.supabase.co`).
- `--supabase-anon-key <key>`: The public anonymous key for the Supabase project. (Optional, defaults to the key for the GGP dev instance).

**Authentication & Caching:**

When you first use the `upload` command, you'll be prompted for your email and password (if not provided via flags) to authenticate with Supabase. Upon successful authentication, an access token and your user ID are cached locally in a file named `ggp_api.json` in the current working directory. Subsequent commands will attempt to use this cached token unless `--force-reauth` is specified or the token is invalid/expired.
