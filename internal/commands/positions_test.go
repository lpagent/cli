package commands

import (
	"strings"
	"testing"
)

func TestNativeSymbol(t *testing.T) {
	for chain, want := range map[string]string{"SOL": "SOL", "": "SOL", "ROBINHOOD": "ETH", "robinhood": "ETH"} {
		if got := nativeSymbol(chain); got != want {
			t.Errorf("nativeSymbol(%q) = %q, want %q", chain, got, want)
		}
	}
}

func TestFormatOpeningRowNativeUnit(t *testing.T) {
	row := map[string]any{
		"pairName":             "WETH/USDG",
		"token1Info":           map[string]any{"token_symbol": "USDG"},
		"valueNative":          1.5,
		"collectedFeeNative":   0.02,
		"unCollectedFeeNative": 0.01,
		"pnl":                  map[string]any{"valueNative": 0.1, "percentNative": 7.0},
	}

	got := strings.Join(formatOpeningRow(row, true, "ETH"), " | ")
	for _, want := range []string{"1.5000 ETH", "0.02 ETH", "0.01 ETH", "0.10 ETH"} {
		if !strings.Contains(got, want) {
			t.Errorf("row is missing %q: %s", want, got)
		}
	}
	if strings.Contains(got, "SOL") {
		t.Errorf("Robinhood row still says SOL: %s", got)
	}
}
