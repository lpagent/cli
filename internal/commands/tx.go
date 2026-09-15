package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lpagent/cli/internal/appctx"
	"github.com/lpagent/cli/internal/output"
)

func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Position transaction operations (zap-out)",
	}

	cmd.AddCommand(
		newTxZapOutEstimatesCmd(),
		newTxDecreaseTxCmd(),
		newTxLandingDecreaseTxCmd(),
	)

	return cmd
}

func newTxZapOutEstimatesCmd() *cobra.Command {
	var (
		positionId string
		id         string
		bps        int
		payouts    []string
	)

	cmd := &cobra.Command{
		Use: "zap-out-estimates",
		// The endpoint this replaced; kept so scripts and agent skills that still
		// call it land on the new estimate instead of an unknown command.
		Aliases: []string{"decrease-quotes"},
		Short:   "Estimate what withdrawing liquidity pays out, per output option",
		Long: `Estimates each decrease-tx output option (allToken0, allToken1, both, allBaseToken):
what the wallet ends up with after the swap and fees, priced on the same route
decrease-tx uses, next to the position's value at the pool price. Run it before
decrease-tx to see the real payout. Meteora DLMM only.`,
		Example: `  lpagent tx zap-out-estimates --position-id <id> --bps 10000
  lpagent tx zap-out-estimates --position-id <id> --bps 10000 --payouts allToken1 -o table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if positionId == "" && id == "" {
				return fmt.Errorf("--position-id (or --id, the encrypted position ID) is required")
			}
			if bps < 1 || bps > 10000 {
				return fmt.Errorf("--bps must be between 1 and 10000")
			}

			app := appctx.FromContext(cmd.Context())

			body := map[string]any{"bps": bps}
			if positionId != "" {
				body["position_id"] = positionId
			} else {
				body["id"] = id
			}
			if len(payouts) > 0 {
				body["outputs"] = payouts
			}

			data, err := app.Client.Post("/position/zap-out-estimates", body)
			if err != nil {
				return err
			}

			if app.Format == "table" {
				printZapOutEstimates(data)
				return nil
			}
			output.Print(data, app.Format, nil)
			return nil
		},
	}

	cmd.Flags().StringVar(&positionId, "position-id", "", "Position ID, as passed to decrease-tx")
	cmd.Flags().StringVar(&id, "id", "", "Encrypted position ID (alternative to --position-id)")
	cmd.Flags().IntVar(&bps, "bps", 10000, "Basis points to withdraw (1-10000); 10000 closes the position")
	cmd.Flags().StringSliceVar(&payouts, "payouts", nil, "Only these output options: allToken0, allToken1, both, allBaseToken")

	return cmd
}

type zapOutEstimate struct {
	Output  string `json:"output"`
	Ok      bool   `json:"ok"`
	Error   string `json:"error"`
	Receive []struct {
		Address  string  `json:"address"`
		UIAmount float64 `json:"uiAmount"`
	} `json:"receive"`
	ValueUsd       float64  `json:"valueUsd"`
	MarketValueUsd float64  `json:"marketValueUsd"`
	PriceImpactPct *float64 `json:"priceImpactPct"`
}

func printZapOutEstimates(data []byte) {
	var envelope struct {
		Data struct {
			Estimates []zapOutEstimate `json:"estimates"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || len(envelope.Data.Estimates) == 0 {
		output.Print(data, "json", nil)
		return
	}

	columns := []output.Column{
		{Header: "OUTPUT"},
		{Header: "YOU GET"},
		{Header: "VALUE"},
		{Header: "MARKET"},
		{Header: "IMPACT"},
	}
	var rows [][]string
	for _, e := range envelope.Data.Estimates {
		if !e.Ok {
			rows = append(rows, []string{e.Output, "unavailable: " + e.Error})
			continue
		}
		var parts []string
		for _, r := range e.Receive {
			parts = append(parts, fmt.Sprintf("%.6g %s", r.UIAmount, shortAddr(r.Address)))
		}
		impact := "-"
		if e.PriceImpactPct != nil {
			impact = fmt.Sprintf("%.2f%%", *e.PriceImpactPct)
		}
		rows = append(rows, []string{
			e.Output,
			strings.Join(parts, " + "),
			fmt.Sprintf("$%.2f", e.ValueUsd),
			fmt.Sprintf("$%.2f", e.MarketValueUsd),
			impact,
		})
	}
	output.PrintRows(columns, rows)
}

func shortAddr(address string) string {
	if address == "So11111111111111111111111111111111111111112" {
		return "SOL"
	}
	if len(address) <= 10 {
		return address
	}
	return address[:4] + "..." + address[len(address)-4:]
}

func newTxDecreaseTxCmd() *cobra.Command {
	var (
		positionId  string
		bps         int
		owner       string
		slippageBps int
		provider    string
		outputType  string
		poolType    string
		fromBinId   float64
		toBinId     float64
	)

	cmd := &cobra.Command{
		Use:   "decrease-tx",
		Short: "Generate transaction to withdraw liquidity from a position",
		Long:  "Generates serialized zap-out transactions. Supports both Meteora DLMM and DAMM V2 positions.",
		Example: `  lpagent tx decrease-tx --position-id <id> --bps 10000 --owner <addr> --slippage-bps 500
  lpagent tx decrease-tx --position-id <id> --bps 5000 --owner <addr> --slippage-bps 300 --output allToken1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if positionId == "" {
				return fmt.Errorf("--position-id is required")
			}

			app := appctx.FromContext(cmd.Context())
			resolvedOwner, err := app.Config.ResolveOwner(owner)
			if err != nil {
				return err
			}

			body := map[string]any{
				"position_id":  positionId,
				"bps":          bps,
				"owner":        resolvedOwner,
				"slippage_bps": slippageBps,
				"type":         poolType,
			}
			if provider != "" {
				body["provider"] = provider
			}
			if outputType != "" {
				body["output"] = outputType
			}
			if cmd.Flags().Changed("from-bin-id") {
				body["fromBinId"] = fromBinId
			}
			if cmd.Flags().Changed("to-bin-id") {
				body["toBinId"] = toBinId
			}

			data, err := app.Client.Post("/position/decrease-tx", body)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, nil)
			return nil
		},
	}

	cmd.Flags().StringVar(&positionId, "position-id", "", "Position ID (required)")
	cmd.Flags().IntVar(&bps, "bps", 0, "Basis points to decrease (0-10000, required)")
	cmd.Flags().StringVar(&owner, "owner", "", "Owner wallet address")
	cmd.Flags().IntVar(&slippageBps, "slippage-bps", 500, "Slippage tolerance in basis points (0-10000)")
	cmd.Flags().StringVar(&provider, "provider", "", "Swap provider: OKX, JUPITER_ULTRA")
	cmd.Flags().StringVar(&outputType, "output-type", "", "Output token: allToken0, allToken1, both, allBaseToken")
	cmd.Flags().StringVar(&poolType, "type", "meteora", "Pool type: meteora, meteora_damm_v2")
	cmd.Flags().Float64Var(&fromBinId, "from-bin-id", 0, "Starting bin ID (DAMM V2)")
	cmd.Flags().Float64Var(&toBinId, "to-bin-id", 0, "Ending bin ID (DAMM V2)")

	_ = cmd.MarkFlagRequired("bps")

	return cmd
}

func newTxLandingDecreaseTxCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "landing-decrease-tx",
		Short: "Submit signed decrease position transactions via Jito",
		Long:  "Submit signed zap-out transactions via Jito bundles for on-chain execution.",
		Example: `  lpagent tx landing-decrease-tx --file signed-tx.json
  cat signed-tx.json | lpagent tx landing-decrease-tx --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())

			body, err := readJSONInput(file)
			if err != nil {
				return err
			}

			data, err := app.Client.Post("/position/landing-decrease-tx", body)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, nil)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file (use - for stdin)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
