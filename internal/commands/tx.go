package commands

import (
	"fmt"

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
		newTxDecreaseQuotesCmd(),
		newTxDecreaseTxCmd(),
		newTxLandingDecreaseTxCmd(),
	)

	return cmd
}

func newTxDecreaseQuotesCmd() *cobra.Command {
	var (
		id  string
		bps int
	)

	cmd := &cobra.Command{
		Use:   "decrease-quotes",
		Short: "Get quotes for withdrawing liquidity from a position",
		Example: `  lpagent tx decrease-quotes --id <encrypted-position-id> --bps 5000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return fmt.Errorf("--id is required (encrypted position ID)")
			}

			app := appctx.FromContext(cmd.Context())

			body := map[string]any{
				"id":  id,
				"bps": bps,
			}

			data, err := app.Client.Post("/position/decrease-quotes", body)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, nil)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Encrypted position ID (required)")
	cmd.Flags().IntVar(&bps, "bps", 0, "Basis points to decrease (0-10000, required)")
	_ = cmd.MarkFlagRequired("bps")

	return cmd
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
