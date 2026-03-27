---
name: lpagent
description: |
  LP Agent CLI — query and manage Solana liquidity pool positions, discover pools,
  check token balances, and generate add/remove liquidity transactions via the
  LP Agent Open API.
triggers:
  - LP positions
  - liquidity pool
  - pool discovery
  - DeFi portfolio
  - Solana LP
  - meteora positions
  - zap-in
  - zap-out
  - add liquidity
  - remove liquidity
  - token balance
  - LP Agent
---

# LP Agent CLI

A command-line interface for the LP Agent Open API. Manages Solana LP positions,
discovers pools, and generates transactions for adding/removing liquidity.

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
lpagent pools discover --search "SOL" --min-liquidity 10000
```

### Get pool details
```bash
lpagent pools info <poolId>
lpagent pools onchain-stats <poolId>
lpagent pools top-lpers <poolId>
```

### Generate add liquidity transaction (Zap-In)
```bash
lpagent pools add-tx <poolId> --owner <wallet> --strategy Spot --input-sol 1
```

### Generate remove liquidity transaction (Zap-Out)
```bash
lpagent tx decrease-quotes --id <encrypted-position-id> --bps 10000
lpagent tx decrease-tx --position-id <id> --bps 10000 --owner <wallet> --slippage-bps 500
```

### Raw API access
```bash
lpagent api get /lp-positions/opening --query "owner=<wallet>"
lpagent api post /position/decrease-quotes --data '{"id":"...","bps":5000}'
```

## Output Formats

All commands support `--output` / `-o` flag:
- `json` (default) — full JSON response, best for piping and AI agents
- `table` — human-readable table format
- `quiet` — IDs only, one per line

## Environment Variables

- `LPAGENT_API_KEY` — API key (overrides config file)
- `LPAGENT_API_URL` — Base URL (overrides default)
- `LPAGENT_DEFAULT_OWNER` — Default wallet owner
