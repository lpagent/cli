package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lpagent/cli/internal/appctx"
	"github.com/lpagent/cli/internal/commands"
	"github.com/lpagent/cli/internal/config"
	"github.com/lpagent/cli/internal/version"
)

var (
	flagOutput  string
	flagVerbose bool
	flagAPIKey  string
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lpagent",
		Short: "LP Agent CLI — manage Solana LP positions from the terminal",
		Long: `LP Agent CLI is a command-line interface for the LP Agent Open API.
It provides access to LP position management, pool discovery, and
transaction generation for Solana DeFi protocols.

Get started:
  lpagent auth set-key              Set your API key
  lpagent positions opening --owner <wallet>  View open positions
  lpagent pools discover            Discover pools`,
		Version:                    fmt.Sprintf("%s (commit: %s, built: %s)", version.Version, version.Commit, version.Date),
		SilenceUsage:               true,
		SilenceErrors:              true,
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip auth setup for auth commands and version/help
			if isAuthCommand(cmd) || cmd.Name() == "version" || cmd.Name() == "help" {
				return nil
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			app, err := appctx.NewApp(cfg, flagAPIKey, flagVerbose, flagOutput)
			if err != nil {
				return err
			}

			cmd.SetContext(appctx.WithApp(cmd.Context(), app))
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: json, table, quiet (default: json)")
	cmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key (overrides config and env)")

	cmd.AddCommand(commands.NewAuthCmd())
	cmd.AddCommand(commands.NewPositionsCmd())
	cmd.AddCommand(commands.NewPoolsCmd())
	cmd.AddCommand(commands.NewTokenCmd())
	cmd.AddCommand(commands.NewTxCmd())
	cmd.AddCommand(commands.NewAPICmd())

	return cmd
}

func isAuthCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "auth" {
			return true
		}
	}
	return false
}

func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
