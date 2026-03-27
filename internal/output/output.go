package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Column struct {
	Header string
	Key    string
	Width  int
}

type TableOptions struct {
	Columns     []Column
	RowFunc     func(row map[string]any) []string
	SummaryFunc func(rows []map[string]any) string
}

func Print(data json.RawMessage, format string, columns []Column) {
	PrintWithOpts(data, format, &TableOptions{Columns: columns})
}

func PrintWithOpts(data json.RawMessage, format string, opts *TableOptions) {
	switch format {
	case "table":
		printTable(data, opts)
	case "quiet":
		printQuiet(data)
	default:
		printJSON(data)
	}
}

func printJSON(data json.RawMessage) {
	var pretty json.RawMessage
	if err := json.Unmarshal(data, &pretty); err != nil {
		fmt.Fprintln(os.Stdout, string(data))
		return
	}
	out, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stdout, string(data))
		return
	}
	fmt.Fprintln(os.Stdout, string(out))
}

func printQuiet(data json.RawMessage) {
	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		var arr []map[string]any
		if err2 := json.Unmarshal(data, &arr); err2 != nil {
			fmt.Fprintln(os.Stdout, string(data))
			return
		}
		envelope.Data = arr
	}

	for _, item := range envelope.Data {
		for _, key := range []string{"tokenId", "id", "position", "pool", "address"} {
			if v, ok := item[key]; ok {
				fmt.Fprintln(os.Stdout, v)
				break
			}
		}
	}
}

func PrintRows(columns []Column, rows [][]string) {
	// Compute widths
	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = len(col.Header)
		if col.Width > widths[i] {
			widths[i] = col.Width
		}
	}
	for _, row := range rows {
		for i := range columns {
			if i < len(row) {
				vl := visibleLen(row[i])
				if vl > widths[i] {
					widths[i] = vl
				}
			}
		}
	}

	gap := "  "

	// Header
	var hdr []string
	for i, col := range columns {
		hdr = append(hdr, padRight(col.Header, widths[i]))
	}
	fmt.Fprintln(os.Stdout, strings.Join(hdr, gap))

	// Separator
	var sep []string
	for _, w := range widths {
		sep = append(sep, strings.Repeat("─", w))
	}
	fmt.Fprintln(os.Stdout, strings.Join(sep, gap))

	// Rows
	for _, row := range rows {
		var parts []string
		for i := range columns {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			parts = append(parts, padRight(val, widths[i]))
		}
		fmt.Fprintln(os.Stdout, strings.Join(parts, gap))
	}
}

func Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
