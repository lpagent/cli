package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/lpagent/cli/internal/version"
)

const (
	githubRepo = "lpagent/cli"
	installURL = "https://raw.githubusercontent.com/lpagent/cli/main/install.sh"
)

// Testable function vars
var latestVersionFetcher = fetchLatestVersion

func NewUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade to the latest version",
		RunE:  runUpgrade,
	}
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	w := cmd.OutOrStdout()

	current := version.Version
	if current == "dev" {
		fmt.Fprintln(w, "Development build — upgrade not applicable. Build from source instead.")
		return nil
	}

	fmt.Fprintf(w, "Current version: %s\n", current)
	fmt.Fprint(w, "Checking for updates... ")

	latest, err := latestVersionFetcher()
	if err != nil {
		fmt.Fprintln(w, "failed")
		return fmt.Errorf("could not check for updates: %w", err)
	}

	if !isUpdateAvailable(current, latest) {
		fmt.Fprintln(w, "already up to date.")
		return nil
	}

	fmt.Fprintf(w, "update available: %s\n\n", latest)

	// On macOS/Linux, run the install script directly
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		fmt.Fprintln(w, "Upgrading...")
		return runInstallScript(cmd.Context(), w)
	}

	// Windows or other: print download link
	fmt.Fprintf(w, "Download the latest release:\n")
	fmt.Fprintf(w, "  https://github.com/%s/releases/tag/v%s\n", githubRepo, latest)
	return nil
}

func runInstallScript(ctx context.Context, w io.Writer) error {
	cmd := exec.CommandContext(ctx, "bash", "-c",
		fmt.Sprintf("curl -fsSL %s | bash", installURL))
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fetchLatestVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&release); err != nil {
		return "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

func isUpdateAvailable(current, latest string) bool {
	current = strings.TrimSpace(strings.TrimPrefix(current, "v"))
	latest = strings.TrimSpace(strings.TrimPrefix(latest, "v"))
	if current == "" || latest == "" || current == "dev" {
		return false
	}
	return latest != current && compareSemver(latest, current) > 0
}

// compareSemver compares two semver strings. Returns >0 if a > b.
func compareSemver(a, b string) int {
	partsA := strings.SplitN(a, ".", 3)
	partsB := strings.SplitN(b, ".", 3)
	for i := 0; i < 3; i++ {
		var va, vb int
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &va)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &vb)
		}
		if va != vb {
			return va - vb
		}
	}
	return 0
}
