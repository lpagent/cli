package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lpagent/cli/internal/appctx"
	"github.com/lpagent/cli/internal/output"
)

func newPoolsAddTxCmd() *cobra.Command {
	var (
		strategy    string
		inputSOL    float64
		percentX    float64
		fromBinId   int
		toBinId     int
		amountX     float64
		amountY     float64
		owner       string
		slippageBps int
		provider    string
		mode        string
	)

	cmd := &cobra.Command{
		Use:   "add-tx <poolId>",
		Short: "Zap-In: Generate add liquidity transaction",
		Long:  "Generate serialized zap-in transactions to add liquidity to a pool. Returns transactions ready to be signed.",
		Args:  cobra.ExactArgs(1),
		Example: `  lpagent pools add-tx <poolId> --owner <addr> --strategy Spot --input-sol 1
  lpagent pools add-tx <poolId> --owner <addr> --strategy BidAsk --amount-x 100 --amount-y 0.5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())
			resolvedOwner, err := app.Config.ResolveOwner(owner)
			if err != nil {
				return err
			}

			if strategy == "" {
				return fmt.Errorf("--strategy is required (Spot, Curve, or BidAsk)")
			}

			body := map[string]any{
				"stratergy":    strategy, // matches API typo
				"owner":        resolvedOwner,
				"slippage_bps": slippageBps,
				"provider":     provider,
				"mode":         mode,
			}
			if inputSOL > 0 {
				body["inputSOL"] = inputSOL
			}
			if cmd.Flags().Changed("percent-x") {
				body["percentX"] = percentX
			}
			if cmd.Flags().Changed("from-bin-id") {
				body["fromBinId"] = fromBinId
			}
			if cmd.Flags().Changed("to-bin-id") {
				body["toBinId"] = toBinId
			}
			if cmd.Flags().Changed("amount-x") {
				body["amountX"] = amountX
			}
			if cmd.Flags().Changed("amount-y") {
				body["amountY"] = amountY
			}

			data, err := app.Client.Post("/pools/"+args[0]+"/add-tx", body)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, nil)
			return nil
		},
	}

	cmd.Flags().StringVar(&strategy, "strategy", "", "Distribution strategy: Spot, Curve, BidAsk (required)")
	cmd.Flags().Float64Var(&inputSOL, "input-sol", 0, "Amount of SOL to input")
	cmd.Flags().Float64Var(&percentX, "percent-x", 0, "Percentage of capital for token X (0-1)")
	cmd.Flags().IntVar(&fromBinId, "from-bin-id", 0, "Starting bin ID")
	cmd.Flags().IntVar(&toBinId, "to-bin-id", 0, "Ending bin ID")
	cmd.Flags().Float64Var(&amountX, "amount-x", 0, "Amount of token X")
	cmd.Flags().Float64Var(&amountY, "amount-y", 0, "Amount of token Y")
	cmd.Flags().StringVar(&owner, "owner", "", "Owner wallet address")
	cmd.Flags().IntVar(&slippageBps, "slippage-bps", 500, "Slippage tolerance in basis points (0-10000)")
	cmd.Flags().StringVar(&provider, "provider", "JUPITER_ULTRA", "Swap provider: OKX, JUPITER_ULTRA")
	cmd.Flags().StringVar(&mode, "mode", "zap-in", "Mode: normal, zap-in")

	return cmd
}

func newPoolsLandingAddTxCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "landing-add-tx",
		Short: "Zap-In: Submit signed add liquidity transactions via Jito",
		Long:  "Submit signed zap-in transactions via Jito bundles for on-chain execution.",
		Example: `  lpagent pools landing-add-tx --file signed-tx.json
  cat signed-tx.json | lpagent pools landing-add-tx --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())

			body, err := readJSONInput(file)
			if err != nil {
				return err
			}

			data, err := app.Client.Post("/pools/landing-add-tx", body)
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

func readJSONInput(file string) (any, error) {
	var raw []byte
	var err error

	if file == "-" {
		raw, err = os.ReadFile("/dev/stdin")
	} else {
		raw, err = os.ReadFile(file)
	}
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}

	var body any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	return body, nil
}
