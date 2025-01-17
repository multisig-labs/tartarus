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

- Go 1.22 or higher
- A C compiler (GCC or Clang)

### Installing on Ubuntu

Here's a oneline command to install Go 1.23 on Ubuntu:

```sh
sudo rm -rf /usr/local/go && wget -qO- https://golang.org/dl/go1.23.2.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf - && echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a $HOME/.bashrc && source $HOME/.bashrc && rm -f go1.22.0.linux-amd64.tar.gz
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
./tartarus -n 10 -p myprefix -s mysuffix -o output.csv
```

This will generate 10 node IDs with the prefix "myprefix" and suffix "mysuffix", and save them to a CSV file called "output.csv".

## Flags

- `-n, --count`: The number of nodes to generate. Defaults to 1.
- `-p, --prefix`: The prefix for the node IDs. Defaults to an empty string.
- `-s, --suffix`: The suffix for the node IDs. Defaults to an empty string.
- `-c, --case-sensitive`: Whether to make the node IDs case-sensitive. Defaults to false.
- `-o, --output`: The output file for the generated nodes. Defaults to "nodes.csv".
- `-v, --verbose`: Whether to print verbose output. Defaults to false.
