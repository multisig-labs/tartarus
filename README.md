# Tartarus: Node ID Generator for Avalanche Nodes

Tartarus is a NodeID generator for the Avalanche Nodes. It generates node IDs with customizable prefixes and suffixes, and saves the generated nodes to a CSV file, a JSON file, or a directory that's compatible with the AvalancheGo Node.

## Usage

To use Tartarus, simply run the executable with the desired flags. For example:
```
./tartarus -n 10 -p myprefix -s mysuffix -o output.csv
```

This will generate 10 node IDs with the prefix "myprefix" and suffix "mysuffix", and save them to a CSV file called "output.csv".

## Flags

* `-n, --count`: The number of nodes to generate. Defaults to 1.
* `-p, --prefix`: The prefix for the node IDs. Defaults to an empty string.
* `-s, --suffix`: The suffix for the node IDs. Defaults to an empty string.
* `-c, --case-sensitive`: Whether to make the node IDs case-sensitive. Defaults to false.
* `-o, --output`: The output file for the generated nodes. Defaults to "nodes.csv".
* `-v, --verbose`: Whether to print verbose output. Defaults to false.

## Building 
Make sure you have a C Compiler installed on your system, as the AvalancheGo code used by Tartarus requires it. Then, build the executable using the following command:

```sh
go build -o tartarus main.go
```
