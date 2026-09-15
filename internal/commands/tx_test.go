package commands

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = orig
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestPrintZapOutEstimatesTable(t *testing.T) {
	// Trimmed from a real /open-api/v1/position/zap-out-estimates answer.
	body := []byte(`{"status":"success","data":{"estimates":[
		{"output":"allToken1","ok":true,"receive":[{"address":"So11111111111111111111111111111111111111112","uiAmount":1.123593898}],
		 "valueUsd":113.03,"marketValueUsd":114.52,"commissionUsd":0.49,"swapCostUsd":1.0,"priceImpactPct":-0.05},
		{"output":"allToken0","ok":false,"error":"No swap route: insufficient liquidity"}
	]}}`)

	out := captureStdout(t, func() { printZapOutEstimates(body) })

	for _, want := range []string{"OUTPUT", "allToken1", "1.12359 SOL", "$113.03", "$114.52", "$0.49", "-0.05%", "unavailable: No swap route: insufficient liquidity"} {
		if !strings.Contains(out, want) {
			t.Errorf("table is missing %q:\n%s", want, out)
		}
	}
}

func TestShortAddr(t *testing.T) {
	if got := shortAddr("So11111111111111111111111111111111111111112"); got != "SOL" {
		t.Errorf("wSOL = %q, want SOL", got)
	}
	if got := shortAddr("9oXA1VWKNiYDYa6VKHBFFBBvZdvuhw1b2eD5Aka7YrVt"); got != "9oXA...YrVt" {
		t.Errorf("mint = %q", got)
	}
}
