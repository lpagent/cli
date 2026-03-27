package output

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ANSI color helpers
const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

var ansiRegex = regexp.MustCompile(`\033\[[0-9;]*m`)

func ColorBySign(s string, val float64) string {
	if val >= 0 {
		return colorGreen + s + colorReset
	}
	return colorRed + s + colorReset
}

// visibleLen returns the length of a string excluding ANSI escape codes.
func visibleLen(s string) int {
	return len(ansiRegex.ReplaceAllString(s, ""))
}

// PadRight pads a string to the given visible width, accounting for ANSI codes.
func PadRight(s string, width int) string {
	return padRight(s, width)
}

func padRight(s string, width int) string {
	vl := visibleLen(s)
	if vl >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vl)
}

func printTable(data json.RawMessage, opts *TableOptions) {
	if opts == nil || (len(opts.Columns) == 0 && opts.RowFunc == nil) {
		printJSON(data)
		return
	}

	rows, count := parseRows(data)
	if len(rows) == 0 {
		fmt.Fprintln(os.Stdout, "No results found.")
		return
	}

	cols := opts.Columns
	gap := "  "

	// Collect all rendered rows first to compute max widths
	var renderedRows [][]string
	for _, row := range rows {
		var vals []string
		if opts.RowFunc != nil {
			vals = opts.RowFunc(row)
		} else {
			vals = make([]string, len(cols))
			for i, col := range cols {
				vals[i] = ExtractValue(row, col.Key)
			}
		}
		renderedRows = append(renderedRows, vals)
	}

	// Compute column widths: max of header, specified width, and actual data
	widths := make([]int, len(cols))
	for i, col := range cols {
		widths[i] = len(col.Header)
		if col.Width > widths[i] {
			widths[i] = col.Width
		}
	}
	for _, vals := range renderedRows {
		for i := range cols {
			if i < len(vals) {
				vl := visibleLen(vals[i])
				if vl > widths[i] {
					widths[i] = vl
				}
			}
		}
	}

	// Print header
	var headerParts []string
	for i, col := range cols {
		headerParts = append(headerParts, padRight(col.Header, widths[i]))
	}
	fmt.Fprintln(os.Stdout, strings.Join(headerParts, gap))

	// Print separator
	var sepParts []string
	for _, w := range widths {
		sepParts = append(sepParts, strings.Repeat("─", w))
	}
	fmt.Fprintln(os.Stdout, strings.Join(sepParts, gap))

	// Print rows
	for _, vals := range renderedRows {
		var parts []string
		for i := range cols {
			val := ""
			if i < len(vals) {
				val = vals[i]
			}
			parts = append(parts, padRight(val, widths[i]))
		}
		fmt.Fprintln(os.Stdout, strings.Join(parts, gap))
	}

	// Summary
	if opts.SummaryFunc != nil {
		fmt.Fprintln(os.Stdout, strings.Join(sepParts, gap))
		summaryLine := opts.SummaryFunc(rows)
		// Summary uses tab-separated values; split and pad
		summaryParts := strings.Split(summaryLine, "\t")
		var parts []string
		for i := range cols {
			val := ""
			if i < len(summaryParts) {
				val = summaryParts[i]
			}
			parts = append(parts, padRight(val, widths[i]))
		}
		fmt.Fprintln(os.Stdout, strings.Join(parts, gap))
	}

	if count > 0 && opts.SummaryFunc == nil {
		fmt.Fprintf(os.Stdout, "\nTotal: %d\n", count)
	}
}

func parseRows(data json.RawMessage) ([]map[string]any, int) {
	var envelope struct {
		Status string          `json:"status"`
		Count  int             `json:"count"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, 0
	}

	var rows []map[string]any
	rawData := envelope.Data
	if rawData == nil {
		rawData = data
	}

	if err := json.Unmarshal(rawData, &rows); err != nil {
		var nested struct {
			Data []map[string]any `json:"data"`
		}
		if err2 := json.Unmarshal(rawData, &nested); err2 != nil {
			var single map[string]any
			if err3 := json.Unmarshal(rawData, &single); err3 != nil {
				return nil, 0
			}
			rows = []map[string]any{single}
		} else {
			rows = nested.Data
		}
	}

	return rows, envelope.Count
}

func ExtractValue(row map[string]any, key string) string {
	parts := strings.Split(key, ".")
	var current any = row

	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return "-"
		}
		current = m[part]
	}

	if current == nil {
		return "-"
	}

	switch v := current.(type) {
	case bool:
		if v {
			return "yes"
		}
		return "no"
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%.4f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func FormatFloat(v any, decimals int) string {
	f, ok := toFloat(v)
	if !ok {
		return "-"
	}
	return fmt.Sprintf("%.*f", decimals, f)
}

func FormatPercent(v any) string {
	f, ok := toFloat(v)
	if !ok {
		return "-"
	}
	return fmt.Sprintf("%.2f%%", f)
}

func FormatUSD(v any) string {
	f, ok := toFloat(v)
	if !ok {
		return "-"
	}
	return fmt.Sprintf("$%.2f", f)
}

func FormatSOL(v any) string {
	f, ok := toFloat(v)
	if !ok {
		return "-"
	}
	return fmt.Sprintf("%.4f SOL", f)
}

func TruncAddr(s string, prefixLen, suffixLen int) string {
	if len(s) <= prefixLen+suffixLen+3 {
		return s
	}
	return s[:prefixLen] + "..." + s[len(s)-suffixLen:]
}

func ToFloatPublic(v any) (float64, bool) {
	return toFloat(v)
}

func toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err == nil
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
