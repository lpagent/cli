package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lpagent/cli/internal/config"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage API key authentication",
	}

	cmd.AddCommand(newAuthSetKeyCmd())
	cmd.AddCommand(newAuthStatusCmd())
	cmd.AddCommand(newAuthSetDefaultOwnerCmd())

	return cmd
}

func newAuthSetKeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-key",
		Short: "Set your LP Agent API key",
		Long:  "Prompts for your API key and saves it to ~/.lpagent/config.json",
		Example: `  lpagent auth set-key
  echo "your-key" | lpagent auth set-key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				cfg = &config.Config{}
			}

			fmt.Print("Enter your API key: ")
			reader := bufio.NewReader(os.Stdin)
			key, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read input: %w", err)
			}

			key = strings.TrimSpace(key)
			if key == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			cfg.APIKey = key
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Println("API key saved successfully.")
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			fmt.Println("LP Agent CLI Status")
			fmt.Println("───────────────────")

			if cfg.APIKey != "" {
				masked := maskKey(cfg.APIKey)
				fmt.Printf("API Key:        %s\n", masked)
			} else {
				fmt.Println("API Key:        (not set)")
			}

			fmt.Printf("Base URL:       %s\n", cfg.BaseURL)

			if cfg.DefaultOwner != "" {
				fmt.Printf("Default Owner:  %s\n", cfg.DefaultOwner)
			} else {
				fmt.Println("Default Owner:  (not set)")
			}

			fmt.Printf("Output Format:  %s\n", cfg.OutputFormat)

			return nil
		},
	}
}

func newAuthSetDefaultOwnerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default-owner <address>",
		Short: "Set the default wallet owner address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				cfg = &config.Config{}
			}

			cfg.DefaultOwner = args[0]
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("Default owner set to: %s\n", args[0])
			return nil
		},
	}
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
