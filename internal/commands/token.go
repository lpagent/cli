package commands

import (
	"github.com/spf13/cobra"

	"github.com/lpagent/cli/internal/appctx"
	"github.com/lpagent/cli/internal/output"
)

func NewTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Query token information",
	}

	cmd.AddCommand(newTokenBalanceCmd())

	return cmd
}

func newTokenBalanceCmd() *cobra.Command {
	var (
		owner string
		ca    string
	)

	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Get token balances for a wallet",
		Example: `  lpagent token balance --owner 9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM
  lpagent token balance --owner <addr> --ca So11111111111111111111111111111111111111112 -o table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())
			resolved, err := app.Config.ResolveOwner(owner)
			if err != nil {
				return err
			}

			params := map[string]string{
				"owner": resolved,
			}
			if ca != "" {
				params["ca"] = ca
			}

			data, err := app.Client.Get("/token/balance", params)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, []output.Column{
				{Header: "Token", Key: "symbol", Width: 10},
				{Header: "Balance", Key: "balance", Width: 16},
				{Header: "USD Value", Key: "balanceInUsd", Width: 14},
				{Header: "Price", Key: "price", Width: 12},
				{Header: "Address", Key: "tokenAddress", Width: 44},
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Wallet address (required)")
	cmd.Flags().StringVar(&ca, "ca", "", "Comma-separated token addresses to filter")

	return cmd
}
