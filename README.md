# LP Agent CLI

`lpagent` is the command-line interface for [LP Agent](https://lpagent.io). Manage Solana LP positions, discover pools, and generate liquidity transactions from your terminal or through AI agents.

- Works standalone or as a Claude agent skill
- Single binary, no runtime dependencies
- JSON output for piping and automation
- Colored table output for humans

## Quick Start

```bash
curl -fsSL https://raw.githubusercontent.com/lpagent/cli/main/install.sh | bash
```

That's it. Now set up your API key and start querying:

```bash
lpagent auth set-key
lpagent auth set-default-owner <wallet-address>
lpagent positions open -o table --native
```

<details>
<summary>Other installation methods</summary>

**Go install:**

```bash
go install github.com/lpagent/cli/cmd/lpagent@latest
```

**Build from source:**

```bash
git clone https://github.com/lpagent/cli.git && cd cli
make build
./bin/lpagent --help
```

</details>

## Commands

```bash
# Positions
lpagent positions open --owner <addr>                          # Open positions
lpagent positions open -o table --native                       # Table view in SOL
lpagent positions historical --owner <addr> --from 2025-01-01  # Closed positions
lpagent positions overview -o table --native                   # Portfolio metrics
lpagent positions logs --position <id>                         # Transaction logs
lpagent positions get --position <id>                          # Position details
lpagent positions revenue <addr>                               # Revenue data

# Pools
lpagent pools discover --chain SOL --sort-by tvl               # Discover pools
lpagent pools info <poolId>                                    # Pool details
lpagent pools positions <poolId> --status Open                 # Positions in a pool
lpagent pools onchain-stats <poolId>                           # TVL, volume, fees
lpagent pools top-lpers <poolId>                               # Top LPs

# Zap-In (add liquidity)
lpagent pools add-tx <poolId> --owner <addr> --strategy Spot --input-sol 1
lpagent pools landing-add-tx --file signed-tx.json

# Zap-Out (remove liquidity)
lpagent tx decrease-quotes --id <id> --bps 10000
lpagent tx decrease-tx --position-id <id> --bps 10000 --owner <addr> --slippage-bps 500
lpagent tx landing-decrease-tx --file signed-tx.json

# Token
lpagent token balance --owner <addr>

# Raw API access
lpagent api get /lp-positions/opening --query "owner=<addr>"
lpagent api post /position/decrease-quotes --data '{"id":"...","bps":5000}'
```

## Output Formats

All commands support `--output` / `-o`: `json` (default), `table`, `quiet`.

Use `--native` on `positions open` and `positions overview` to show values in SOL instead of USD.

## Configuration

Config stored at `~/.lpagent/config.json`. Override with env vars or CLI flags:

| Variable               | Flag          | Description            |
|------------------------|---------------|------------------------|
| `LPAGENT_API_KEY`      | `--api-key`   | API key                |
| `LPAGENT_API_URL`      |               | Base URL               |
| `LPAGENT_DEFAULT_OWNER`| `--owner`     | Default wallet address |

Priority: CLI flags > env vars > config file.

## Development

```bash
make build      # Build to bin/lpagent
make test       # Run tests
make check      # fmt + vet + test
```

## Release

Automated via GitHub Actions + GoReleaser on version tags:

```bash
git tag v0.1.0 && git push origin v0.1.0
```

## License

MIT
