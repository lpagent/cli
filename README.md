# LP Agent CLI

Command-line interface for the [LP Agent Open API](https://docs.lpagent.io). Manage Solana LP positions, discover pools, and generate add/remove liquidity transactions from the terminal.

Single binary, no runtime dependencies. Works standalone or as a Claude agent skill.

## Install

### From release binaries

Download the latest binary from [GitHub Releases](https://github.com/lpagent/cli/releases).

```bash
# macOS (Apple Silicon)
curl -sL https://github.com/lpagent/cli/releases/latest/download/lp-agent-cli_darwin_arm64.tar.gz | tar xz
sudo mv lpagent /usr/local/bin/

# macOS (Intel)
curl -sL https://github.com/lpagent/cli/releases/latest/download/lp-agent-cli_darwin_amd64.tar.gz | tar xz
sudo mv lpagent /usr/local/bin/

# Linux (amd64)
curl -sL https://github.com/lpagent/cli/releases/latest/download/lp-agent-cli_linux_amd64.tar.gz | tar xz
sudo mv lpagent /usr/local/bin/
```

### From source

```bash
go install github.com/lpagent/cli/cmd/lpagent@latest
```

### Build locally

```bash
git clone https://github.com/lpagent/cli.git
cd lp-agent-cli
make build
./bin/lpagent --help
```

## Quick start

```bash
# 1. Set your API key (get one at https://app.lpagent.io)
lpagent auth set-key

# 2. Set a default wallet so you don't need --owner every time
lpagent auth set-default-owner <wallet-address>

# 3. View open positions
lpagent positions open -o table

# 4. View open positions in SOL
lpagent positions open -o table --native

# 5. Check portfolio overview
lpagent positions overview -o table --native
```

## Commands

### Auth

```bash
lpagent auth set-key                    # Set API key (saved to ~/.lpagent/config.json)
lpagent auth status                     # Show current config
lpagent auth set-default-owner <addr>   # Set default wallet
```

### Positions

```bash
lpagent positions open --owner <addr>                          # Open positions
lpagent positions historical --owner <addr> --from 2025-01-01  # Closed positions
lpagent positions overview --owner <addr>                      # Portfolio metrics
lpagent positions logs --position <id>                         # Transaction logs
lpagent positions get --position <id>                          # Position details
lpagent positions revenue <addr>                               # Revenue data
```

### Pools

```bash
lpagent pools discover --chain SOL --sort-by tvl       # Discover pools
lpagent pools info <poolId>                            # Pool details
lpagent pools positions <poolId> --status Open         # Positions in a pool
lpagent pools onchain-stats <poolId>                   # TVL, volume, fees
lpagent pools top-lpers <poolId>                       # Top liquidity providers
lpagent pools add-tx <poolId> --owner <addr> --strategy Spot --input-sol 1   # Zap-In tx
lpagent pools landing-add-tx --file signed-tx.json     # Submit signed tx
```

### Token

```bash
lpagent token balance --owner <addr>                   # All token balances
lpagent token balance --owner <addr> --ca <mint>       # Specific tokens
```

### Transactions (Zap-Out)

```bash
lpagent tx decrease-quotes --id <id> --bps 10000      # Get withdrawal quotes
lpagent tx decrease-tx --position-id <id> --bps 10000 --owner <addr> --slippage-bps 500
lpagent tx landing-decrease-tx --file signed-tx.json   # Submit signed tx
```

### Raw API

```bash
lpagent api get /lp-positions/opening --query "owner=<addr>"
lpagent api post /position/decrease-quotes --data '{"id":"...","bps":5000}'
```

## Output formats

All commands support `--output` / `-o`:

| Format  | Description                        |
|---------|------------------------------------|
| `json`  | Full JSON response (default)       |
| `table` | Human-readable table with colors   |
| `quiet` | IDs only, one per line             |

Use `--native` on `positions open` and `positions overview` to show values in SOL instead of USD.

## Configuration

Config stored at `~/.lpagent/config.json`:

```json
{
  "api_key": "your-api-key",
  "api_base_url": "https://api.lpagent.io/open-api/v1",
  "default_owner": "your-wallet-address",
  "output_format": "json"
}
```

### Environment variables

| Variable               | Description                    |
|------------------------|--------------------------------|
| `LPAGENT_API_KEY`      | API key (overrides config)     |
| `LPAGENT_API_URL`      | Base URL (overrides config)    |
| `LPAGENT_DEFAULT_OWNER`| Default wallet (overrides config) |

CLI flags take highest priority, then env vars, then config file.

## Development

```bash
make build      # Build binary to bin/lpagent
make test       # Run tests
make fmt        # Format code
make vet        # Run go vet
make lint       # Run golangci-lint
make check      # All of the above
make install    # Install to $GOPATH/bin
```

## Release

Releases are automated via GitHub Actions + GoReleaser. To create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

This triggers the release workflow which builds cross-platform binaries (darwin/linux/windows, amd64/arm64), creates a GitHub Release, and publishes checksums.

## License

MIT
