# Coqnet Scripts

This is the temporary home of some scripts to help our GGP node partners with the Coqnet!

## Prerequisites
All scripts will require the following command line tools to be installed.

- `curl`
- `jq`

If you are running Ubuntu (which is recommended by us) you can simply run this command.

```sh
sudo apt-get update && sudo apt-get install -y curl jq
```

## Scripts

### `signup.sh`

This script creates an account on our node manager DB. It will send you an email to confirm your account. You will need to confirm your account before you can use it.

### `post_nodes.sh`

You'll use tartarus (the program this repo is for) to generate all your Node IDs and keys. This script will then take that output and post it to the node manager DB.

### `update_status.sh`

This script will update the status of your nodes in the node manager DB. It should be downloaded and ran on your node(s) via cron every 10 to 15 minutes.
