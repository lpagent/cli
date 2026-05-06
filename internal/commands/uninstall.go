package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func NewUninstallCmd() *cobra.Command {
	var flagYes bool
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the lpagent binary and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(cmd, flagYes)
		},
	}
	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func runUninstall(cmd *cobra.Command, skipConfirm bool) error {
	w := cmd.OutOrStdout()

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	binaryPath := findBinary(home)
	configDir := filepath.Join(home, ".lpagent")

	fmt.Fprintln(w, "This will remove:")
	if binaryPath != "" {
		fmt.Fprintf(w, "  - Binary:  %s\n", binaryPath)
	}
	if dirExists(configDir) {
		fmt.Fprintf(w, "  - Config:  %s/\n", configDir)
	}

	if binaryPath == "" && !dirExists(configDir) {
		fmt.Fprintln(w, "  (nothing found to remove)")
		return nil
	}

	if !skipConfirm {
		fmt.Fprint(w, "\nContinue? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(w, "Aborted.")
			return nil
		}
	}

	fmt.Fprintln(w)

	if binaryPath != "" {
		if err := os.Remove(binaryPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove binary: %w", err)
		}
		fmt.Fprintf(w, "Removed %s\n", binaryPath)
	}

	if dirExists(configDir) {
		if err := os.RemoveAll(configDir); err != nil {
			return fmt.Errorf("failed to remove config directory: %w", err)
		}
		fmt.Fprintf(w, "Removed %s/\n", configDir)
	}

	fmt.Fprintln(w, "\nlpagent has been uninstalled.")
	return nil
}

func findBinary(home string) string {
	candidates := []string{
		filepath.Join(home, ".local", "bin", "lpagent"),
	}

	if runtime.GOOS == "windows" {
		candidates[0] += ".exe"
	}

	if gopath := os.Getenv("GOPATH"); gopath != "" {
		bin := filepath.Join(gopath, "bin", "lpagent")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		candidates = append(candidates, bin)
	} else {
		bin := filepath.Join(home, "go", "bin", "lpagent")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		candidates = append(candidates, bin)
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
