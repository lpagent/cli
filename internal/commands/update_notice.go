package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lpagent/cli/internal/version"
)

var checkInterval = 24 * time.Hour

// UpdateCheck holds state for a non-blocking background version check.
type UpdateCheck struct {
	latest string
	done   chan struct{}
}

// StartUpdateCheck begins a background version check if the cache is stale.
// Returns nil if the check should be skipped.
func StartUpdateCheck() *UpdateCheck {
	if version.Version == "dev" {
		return nil
	}
	if os.Getenv("LPAGENT_NO_UPDATE_CHECK") == "1" {
		return nil
	}

	// Skip for non-interactive sessions
	fi, err := os.Stdout.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return nil
	}

	uc := &UpdateCheck{done: make(chan struct{})}
	cached := readUpdateCache()

	if cached != nil {
		age := time.Since(cached.CheckedAt)
		if age >= 0 && age < checkInterval {
			uc.latest = cached.LatestVersion
			close(uc.done)
			return uc
		}
	}

	// Cache stale or missing — fetch in background
	go func() {
		defer close(uc.done)
		latest, err := latestVersionFetcher()
		if err != nil || latest == "" {
			return
		}
		uc.latest = latest
		writeUpdateCache(latest)
	}()

	return uc
}

// Notice returns a formatted update notice, or "" if no update is available.
// Never blocks.
func (uc *UpdateCheck) Notice() string {
	if uc == nil {
		return ""
	}

	select {
	case <-uc.done:
	default:
		return ""
	}

	if !isUpdateAvailable(version.Version, uc.latest) {
		return ""
	}

	return fmt.Sprintf(
		"\nUpdate available: %s → %s — Run \"lpagent upgrade\" to update\n",
		version.Version, uc.latest,
	)
}

type updateCache struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

func updateCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".lpagent", ".update-check")
}

func readUpdateCache() *updateCache {
	p := updateCachePath()
	if p == "" {
		return nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var c updateCache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil
	}
	if c.LatestVersion == "" || c.CheckedAt.IsZero() {
		return nil
	}
	return &c
}

func writeUpdateCache(latestVersion string) {
	p := updateCachePath()
	if p == "" {
		return
	}
	c := updateCache{
		LatestVersion: latestVersion,
		CheckedAt:     time.Now().UTC(),
	}
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	dir := filepath.Dir(p)
	_ = os.MkdirAll(dir, 0700)
	_ = os.WriteFile(p, data, 0644)
}
