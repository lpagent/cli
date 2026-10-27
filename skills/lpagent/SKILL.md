---
name: lpagent
description: |
  LP Agent CLI - query and manage liquidity pool positions on Solana (Meteora, Orca)
  and the Robinhood chain (Uniswap v3/v4), discover pools, check token balances,
  and generate add/remove liquidity transactions via the LP Agent Open API.
triggers:
  - LP positions
  - liquidity pool
  - pool discovery
  - DeFi portfolio
  - Solana LP
  - Robinhood LP
  - Uniswap v4
  - EVM LP
  - meteora positions
  - zap-in
  - zap-out
  - add liquidity
  - remove liquidity
  - claim fees
  - token balance
  - LP Agent
---

# LP Agent

A command-line interface for the LP Agent Open API. Manages LP positions on Solana
(Meteora DLMM / DAMM v2, Orca) and on the Robinhood chain (Uniswap v3 / v4),
discovers pools, and generates transactions for adding/removing liquidity.

## Chains

| Chain | `--chain` value | Protocols | Native token | Address / position id |
| --- | --- | --- | --- | --- |
| Solana | `SOL` (default) | Meteora DLMM, Meteora DAMM v2, Orca | SOL | base58 pubkey / position pubkey |
| Robinhood | `ROBINHOOD` | Uniswap v3, Uniswap v4 | ETH | `0x…` hex / NFT `tokenId` |

Read commands take `--chain` (`SOL` by default): `positions open/historical/
overview/logs/revenue`, `token balance`, `pools discover` and `pools positions`.
Pass `--chain ROBINHOOD` with a `0x` wallet or pool, or the lookup runs on Solana.
`positions get`, `pools onchain-stats` and `pools top-lpers` read the chain from
the id itself. Native values are SOL on Solana and ETH on Robinhood.

`pools info` and every transaction command (`add-tx`, `zap-out-estimates`,
`decrease-tx`, `claim-fee-tx` and the `landing-` ones) are **Solana only**: do not
pass a Robinhood address to them.

```bash
lpagent pools discover --chain ROBINHOOD --sort-by tvl -o table
lpagent positions open --chain ROBINHOOD --owner 0xYourWallet -o table --native
lpagent positions logs --chain ROBINHOOD --owner 0xYourWallet
lpagent token balance --chain ROBINHOOD --owner 0xYourWallet
```

## Installation

### Quick install (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/lpagent/cli/main/install.sh | bash
```

This detects your OS and architecture, downloads the latest release from GitHub,
and places the `lpagent` binary at `~/.local/bin/lpagent`. If `~/.local/bin` is
not in your `PATH`, the script will print instructions to add it.

### Go install

```bash
go install github.com/lpagent/cli/cmd/lpagent@latest
```

### Build from source

```bash
git clone https://github.com/lpagent/cli.git && cd cli
make build
./bin/lpagent --help
```

### Windows

Download the latest `.zip` from [GitHub Releases](https://github.com/lpagent/cli/releases)
and add the binary to your `PATH`.

### Verify installation

```bash
lpagent --version
```

### Upgrade

```bash
lpagent upgrade
```

## Authentication

Before using any command, set your API key:

```bash
lpagent auth set-key
```

You can also set a default wallet owner to avoid passing `--owner` every time:

```bash
lpagent auth set-default-owner <wallet-address>
```

## Common Workflows

### View open positions
```bash
lpagent positions opening --owner <wallet>
lpagent positions opening --owner <wallet> -o table
```

### View historical (closed) positions
```bash
lpagent positions historical --owner <wallet> --from 2025-01-01 --to 2025-06-01
```

### Get position overview metrics (PnL, APR, win rate)
```bash
lpagent positions overview --owner <wallet>
```

### Check token balances
```bash
lpagent token balance --owner <wallet>
lpagent token balance --owner <wallet> --ca So11111111111111111111111111111111111111112
```

### Discover pools
```bash
lpagent pools discover --chain SOL --sort-by tvl -o table
lpagent pools discover --chain ROBINHOOD --sort-by tvl -o table
lpagent pools discover --search "SOL" --min-liquidity 10000
```

### Get pool details
```bash
lpagent pools info <poolId>
lpagent pools onchain-stats <poolId>
lpagent pools top-lpers <poolId>
```

### View position activity logs
```bash
lpagent positions logs --position <id>
lpagent positions logs --chain ROBINHOOD --owner 0xYourWallet
```

### Generate add liquidity transaction (Zap-In) - Solana only
```bash
lpagent pools add-tx <poolId> --owner <wallet> --strategy Spot --input-sol 1
```

### Generate remove liquidity transaction (Zap-Out) - Solana only
Estimate first: `zap-out-estimates` returns, per output option (allToken0, allToken1,
both, allBaseToken), what the wallet ends up with after the swap and LP Agent's fee,
priced on the same route `decrease-tx` uses. Pick the `--output-type` from it.
```bash
lpagent tx zap-out-estimates --position-id <id> --bps 10000
lpagent tx decrease-tx --position-id <id> --bps 10000 --owner <wallet> --slippage-bps 500 --output-type allToken1
```

### Claim fees without closing the position - Solana only
`claim-fee-tx` returns `claimTxsWithJito` (and `swapTxsWithJito` with `--swap-to-sol`,
which turns the claimed fees into SOL). Sign them, keep `lastValidBlockHeight`, and
submit the same JSON shape with `landing-claim-fee-tx`.
```bash
lpagent tx claim-fee-tx --position-id <id> --owner <wallet>
lpagent tx claim-fee-tx --position-id <id> --owner <wallet> --swap-to-sol --slippage-bps 300
lpagent tx landing-claim-fee-tx --file signed-tx.json
```

### Raw API access
```bash
lpagent api get /lp-positions/opening --query "owner=<wallet>"
lpagent api post /position/zap-out-estimates --data '{"position_id":"...","bps":5000}'
```

## Output Formats

All commands support `--output` / `-o` flag:
- `json` (default): full JSON response, best for piping and AI agents
- `table`: human-readable table format
- `quiet`: IDs only, one per line

## Environment Variables

- `LPAGENT_API_KEY`: API key (overrides config file)
- `LPAGENT_API_URL`: Base URL (overrides default)
- `LPAGENT_DEFAULT_OWNER`: Default wallet owner
