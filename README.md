# emma CLI

A kubectl-quality command-line interface for the [emma cloud platform](https://emma.ms), written in Go.

## Installation

### Homebrew (macOS / Linux)

```sh
brew install emma-community/tap/emma
```

### Download binary

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases](https://github.com/emma-community/emma-cli/releases) page.

### Build from source

```sh
go install github.com/emma-community/emma-cli@latest
```

Or clone and build:

```sh
git clone https://github.com/emma-community/emma-cli.git
cd emma-cli
make build
```

## Getting started

### Authenticate

```sh
# Interactive login (secret prompted, never stored in shell history)
emma auth login --client-id <CLIENT_ID>

# Non-interactive / CI
EMMA_CLIENT_SECRET=<secret> emma auth login --client-id <CLIENT_ID>

# Check current context
emma auth whoami
```

Credentials are stored in `~/.emma/config.yaml` (mode `0600`). Multiple named contexts are supported, similar to `kubectl config`.

### Environment variables

| Variable | Description |
|---|---|
| `EMMA_CLIENT_ID` | Client ID (alternative to `--client-id`) |
| `EMMA_CLIENT_SECRET` | Client secret (avoids interactive prompt) |
| `NO_COLOR` | Disable colored output |

## Usage

```
emma [command] [subcommand] [flags]

Global flags:
  -o, --output string     Output format: table, json, yaml, wide (default "table")
      --project-id int    Project ID (overrides context default)
      --no-color          Disable colored output
      --debug             Enable debug HTTP logging
```

### Commands

| Command | Description |
|---|---|
| `emma auth` | Manage authentication (login, logout, whoami) |
| `emma vm` | Virtual machines (list, get, create, delete, actions, wait) |
| `emma spot` | Spot instances (list, get, create, delete, actions, wait) |
| `emma k8s` | Kubernetes clusters (list, get, create, edit, delete) |
| `emma volume` | Block volumes (list, get, create, delete, attach, detach) |
| `emma sg` | Security groups (list, get, create, update, delete, instance-add, instance-list) |
| `emma sshkey` | SSH keys (list, get, create, update, delete) |
| `emma subnet` | Subnetworks (list, get, create, edit, delete) |
| `emma mcn` | Multi-cloud networks (list, enable-network, disable-network, cloud-connect, cross-connect) |
| `emma workflow` | Workflow templates (list, create, delete) |
| `emma provider` | Providers, locations, datacenters, OS, and accelerator types |
| `emma config` | CLI configuration, contexts, and resource configuration listings |
| `emma completion` | Generate shell completion scripts |
| `emma version` | Print version information |

### VM actions

```sh
emma vm actions --id 123 --action start
emma vm actions --id 123 --action shutdown
emma vm actions --id 123 --action reboot
emma vm actions --id 123 --action clone --name my-clone
emma vm actions --id 123 --action rename --name new-name
emma vm actions --id 123 --action transfer --name <target-datacenter-id>
```

### Spot instance actions

```sh
emma spot actions --id 456 --action start
emma spot actions --id 456 --action shutdown
emma spot actions --id 456 --action reboot
emma spot actions --id 456 --action rename --name new-name
emma spot actions --id 456 --action change-price --price 0.50
```

### Examples

```sh
# List VMs in a specific project
emma vm list --project-id 42

# List only running VMs
emma vm list --status RUNNING

# Wide output with cost and provider columns
emma vm list -o wide

# Watch VMs refresh every 5 seconds
emma vm list --watch 5

# Wait for a VM to reach RUNNING state (timeout: 5 min)
emma vm wait --id 123 --state RUNNING --timeout 300

# Create a VM (shows estimated cost before creation)
emma vm create \
  --name my-vm \
  --datacenter-id dc-1 \
  --os-id 42 \
  --vcpu 4 \
  --ram 8 \
  --volume-size 50 \
  --volume-type ssd \
  --cloud-network-type multi-cloud \
  --ssh-key-id 99

# List available VM hardware configurations
emma config vm-configs --datacenter-id dc-1 --vcpu 4

# List available K8s node configurations
emma config k8s-configs --connection-type InternetConnect

# Create a Kubernetes cluster with worker nodes
emma k8s create \
  --name my-cluster \
  --deployment-location EU \
  --connection-type InternetConnect \
  --worker-name default \
  --worker-datacenter-id dc-1 \
  --worker-vcpu 4 \
  --worker-ram 8 \
  --worker-volume-size 50 \
  --worker-volume-type ssd

# View security group with rules
emma sg get --id 5

# Add an instance to a security group
emma sg instance-add --sg-id 5 --instance-id 123

# List multi-cloud network topology
emma mcn list

# Enable a direct network in a region
emma mcn enable-network --macro-region EMEA --connectivity-center-id 1

# List datacenters filtered by provider
emma provider datacenter-list --provider aws

# List GPU accelerator types
emma provider accelerator-list

# Output as JSON or YAML
emma sg list -o yaml
emma vm get --id 1 -o json

# Debug mode (logs HTTP requests)
emma vm list --debug

# Generate shell completions (bash)
emma completion bash >> ~/.bashrc
```

## Configuration

The CLI uses `~/.emma/config.yaml` for multi-context credential storage, following the kubectl pattern:

```yaml
current-context: production
contexts:
  - name: production
    client_id: abc123
    access_token: eyJ...
    token_expires_at: "2024-12-31T00:00:00Z"
  - name: staging
    client_id: xyz789
    access_token: eyJ...
```

Switch context:

```sh
emma config use-context staging
emma config get-contexts
```

### Resource configuration listings

Query available hardware configurations before creating resources:

```sh
emma config vm-configs                          # All VM configs
emma config vm-configs --provider-id 1          # Filter by provider
emma config spot-configs --datacenter-id dc-1   # Spot configs in a datacenter
emma config k8s-configs --connection-type InternetConnect
emma config volume-configs --volume-type ssd
```

## Shell completion

Static completions for commands and flags, plus dynamic completions for key fields (`--datacenter-id`, `--os-id`) that query the API and cache results in `~/.emma/cache/`.

```sh
# Bash
emma completion bash > /etc/bash_completion.d/emma

# Zsh
emma completion zsh > "${fpath[1]}/_emma"

# Fish
emma completion fish > ~/.config/fish/completions/emma.fish
```

## Building

```sh
make build    # Build binary to ./emma
make test     # Run tests
make lint     # Run go vet
make clean    # Remove binary
```

Cross-platform releases are produced by [goreleaser](https://goreleaser.com):

```sh
goreleaser release --snapshot --clean
```

## License

[MIT](LICENSE)
