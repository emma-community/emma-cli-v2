# emma CLI

A kubectl-quality command-line interface for the [emma cloud platform](https://emma.ms), written in Go.

## Installation

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
  -o, --output string     Output format: table, json, yaml (default "table")
      --project-id int    Project ID (overrides context default)
      --no-color          Disable colored output
      --debug             Enable debug output
```

### Commands

| Command | Description |
|---|---|
| `emma auth` | Manage authentication (login, logout, whoami) |
| `emma vm` | Virtual machines (list, get, create, delete, start, stop, reboot, snapshot) |
| `emma spot` | Spot instances (list, get, create, delete, start, stop, reboot, snapshot) |
| `emma k8s` | Kubernetes clusters (list, get, create, delete) |
| `emma volume` | Block volumes (list, get, create, delete) |
| `emma sg` | Security groups (list, get, create, delete, add-rule) |
| `emma sshkey` | SSH keys (list, get, create, delete) |
| `emma subnet` | Subnetworks (list, get, create, delete) |
| `emma provider` | Providers and locations (list locations, data centers, providers) |
| `emma config` | Manage CLI configuration and contexts |
| `emma completion` | Generate shell completion scripts |
| `emma version` | Print version information |

### Examples

```sh
# List VMs in a specific project
emma vm list --project-id 42

# Create a VM (outputs JSON)
emma vm create \
  --name my-vm \
  --provider-id 1 \
  --location-id 6 \
  --dc-id 12 \
  --os-family UBUNTU \
  --os-version 22.04 \
  --vcpu 2 \
  --ram 4 \
  --volume-size 50 \
  --ssh-key-id 99 \
  -o json

# Watch a VM reach ACTIVE state
emma vm get 1234

# List security groups as YAML
emma sg list -o yaml

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

## Shell completion

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
