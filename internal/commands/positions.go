package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lpagent/cli/internal/appctx"
	"github.com/lpagent/cli/internal/output"
)

func NewPositionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "positions",
		Aliases: []string{"pos"},
		Short:   "Manage LP positions",
	}

	cmd.AddCommand(
		newPositionsOpeningCmd(),
		newPositionsHistoricalCmd(),
		newPositionsOverviewCmd(),
		newPositionsLogsCmd(),
		newPositionsGetCmd(),
		newPositionsRevenueCmd(),
	)

	return cmd
}

func newPositionsOpeningCmd() *cobra.Command {
	var (
		owner  string
		native bool
	)

	cmd := &cobra.Command{
		Use:     "open",
		Aliases: []string{"opening"},
		Short:   "List open LP positions for an owner",
		Example: `  lpagent positions open --owner 9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM
  lpagent positions open -o table
  lpagent positions open -o table --native`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())
			resolved, err := app.Config.ResolveOwner(owner)
			if err != nil {
				return err
			}

			data, err := app.Client.Get("/lp-positions/opening", map[string]string{
				"owner": resolved,
			})
			if err != nil {
				return err
			}

			columns := []output.Column{
				{Header: "Pair", Width: 12},
				{Header: "Pool", Width: 14},
				{Header: "Age", Width: 8},
				{Header: "Value", Width: 14},
				{Header: "Claimed Fee", Width: 12},
				{Header: "Unclaim Fee", Width: 12},
				{Header: "uPnL", Width: 12},
				{Header: "DPR", Width: 8},
				{Header: "In Range", Width: 8},
				{Header: "Protocol", Width: 8},
			}

			output.PrintWithOpts(data, app.Format, &output.TableOptions{
				Columns: columns,
				RowFunc: func(row map[string]any) []string {
					return formatOpeningRow(row, native)
				},
				SummaryFunc: func(rows []map[string]any) string {
					return formatOpeningSummary(rows, native)
				},
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Wallet address of the position owner")
	cmd.Flags().BoolVar(&native, "native", false, "Show values in native token (SOL) instead of USD")
	return cmd
}

func formatOpeningRow(row map[string]any, native bool) []string {
	// Pair
	pairName := fmt.Sprintf("%v", row["pairName"])
	tokenName1 := "SOL"
	if t1, ok := row["token1Info"].(map[string]any); ok {
		if sym, ok := t1["token_symbol"].(string); ok {
			tokenName1 = sym
		}
	}
	displayPair := pairName
	if !strings.Contains(strings.ToUpper(pairName), "/"+strings.ToUpper(tokenName1)) {
		displayPair = pairName + "/" + tokenName1
	}

	// Pool (just truncated address)
	pool := output.TruncAddr(fmt.Sprintf("%v", row["pool"]), 5, 4)

	// Age
	age := "-"
	if a, ok := row["age"]; ok && a != nil {
		if ageF, ok := output.ToFloatPublic(a); ok {
			if ageF < 1 {
				age = fmt.Sprintf("%.0f min", ageF*60)
			} else if ageF < 24 {
				age = fmt.Sprintf("%.0f hours", ageF)
			} else {
				age = fmt.Sprintf("%.1f days", ageF/24)
			}
		}
	}

	// Value
	var value string
	if native {
		value = output.FormatSOL(row["valueNative"])
	} else {
		value = output.FormatUSD(row["value"])
	}

	// Claimed Fee
	var claimedFee string
	if native {
		claimedFee = output.FormatFloat(row["collectedFeeNative"], 2) + " SOL"
	} else {
		claimedFee = output.FormatUSD(row["collectedFee"])
	}

	// Unclaim Fee
	var unclaimFee string
	if native {
		unclaimFee = output.FormatFloat(row["unCollectedFeeNative"], 2) + " SOL"
	} else {
		unclaimFee = output.FormatUSD(row["unCollectedFee"])
	}

	// uPnL (colored, percent in parentheses)
	var pnlVal float64
	pnlMap, _ := row["pnl"].(map[string]any)
	var pnlValueStr, pnlPctStr string
	if native {
		pnlVal, _ = output.ToFloatPublic(safeGet(pnlMap, "valueNative"))
		pnlPct, _ := output.ToFloatPublic(safeGet(pnlMap, "percentNative"))
		pnlValueStr = fmt.Sprintf("%.2f SOL", pnlVal)
		pnlPctStr = fmt.Sprintf("(%.2f%%)", pnlPct)
	} else {
		pnlVal, _ = output.ToFloatPublic(safeGet(pnlMap, "value"))
		pnlPct, _ := output.ToFloatPublic(safeGet(pnlMap, "percent"))
		pnlValueStr = output.FormatUSD(pnlVal)
		pnlPctStr = fmt.Sprintf("(%.2f%%)", pnlPct)
	}
	upnl := output.ColorBySign(pnlValueStr+" "+pnlPctStr, pnlVal)

	// DPR (colored)
	var dprVal float64
	if native {
		dprVal, _ = output.ToFloatPublic(row["dprNative"])
	} else {
		dprVal, _ = output.ToFloatPublic(row["dpr"])
	}
	dpr := output.ColorBySign(fmt.Sprintf("%.2f%%", dprVal), dprVal)

	// In Range
	inRange := "no"
	if v, ok := row["inRange"].(bool); ok && v {
		inRange = "yes"
	}

	// Protocol
	protocol := fmt.Sprintf("%v", row["protocol"])

	return []string{displayPair, pool, age, value, claimedFee, unclaimFee, upnl, dpr, inRange, protocol}
}

func formatOpeningSummary(rows []map[string]any, native bool) string {
	var totalValue, totalUPnL, totalPnLPct, totalClaimed, totalUnclaimed float64
	for _, row := range rows {
		if native {
			f, ok := output.ToFloatPublic(row["valueNative"])
			if ok {
				totalValue += f
			}
			f, ok = output.ToFloatPublic(row["collectedFeeNative"])
			if ok {
				totalClaimed += f
			}
			f, ok = output.ToFloatPublic(row["unCollectedFeeNative"])
			if ok {
				totalUnclaimed += f
			}
		} else {
			f, ok := output.ToFloatPublic(row["value"])
			if ok {
				totalValue += f
			}
			f, ok = output.ToFloatPublic(row["collectedFee"])
			if ok {
				totalClaimed += f
			}
			f, ok = output.ToFloatPublic(row["unCollectedFee"])
			if ok {
				totalUnclaimed += f
			}
		}
		if pnl, ok := row["pnl"].(map[string]any); ok {
			if native {
				f, ok := output.ToFloatPublic(pnl["valueNative"])
				if ok {
					totalUPnL += f
				}
			} else {
				f, ok := output.ToFloatPublic(pnl["value"])
				if ok {
					totalUPnL += f
				}
			}
		}
	}

	if totalValue > 0 {
		totalPnLPct = totalUPnL / totalValue * 100
	}

	var valStr, upnlStr, claimedStr, unclaimedStr string
	if native {
		valStr = fmt.Sprintf("%.4f SOL", totalValue)
		upnlStr = fmt.Sprintf("%.2f SOL (%.2f%%)", totalUPnL, totalPnLPct)
		claimedStr = fmt.Sprintf("%.2f SOL", totalClaimed)
		unclaimedStr = fmt.Sprintf("%.2f SOL", totalUnclaimed)
	} else {
		valStr = output.FormatUSD(totalValue)
		upnlStr = fmt.Sprintf("%s (%.2f%%)", output.FormatUSD(totalUPnL), totalPnLPct)
		claimedStr = output.FormatUSD(totalClaimed)
		unclaimedStr = output.FormatUSD(totalUnclaimed)
	}

	upnlStr = output.ColorBySign(upnlStr, totalUPnL)
	return fmt.Sprintf("Total\t\t\t%s\t%s\t%s\t%s\t\t\t", valStr, claimedStr, unclaimedStr, upnlStr)
}

func safeGet(m map[string]any, key string) any {
	if m == nil {
		return nil
	}
	return m[key]
}

func newPositionsHistoricalCmd() *cobra.Command {
	var (
		owner    string
		fromDate string
		toDate   string
		page     int
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "historical",
		Short: "List historical (closed) LP positions",
		Example: `  lpagent positions historical --owner 9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM
  lpagent positions historical --owner <addr> --from 2025-01-01 --to 2025-09-01 --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())
			resolved, err := app.Config.ResolveOwner(owner)
			if err != nil {
				return err
			}

			params := map[string]string{
				"owner": resolved,
			}
			if fromDate != "" {
				params["from_date"] = fromDate
			}
			if toDate != "" {
				params["to_date"] = toDate
			}
			if page > 0 {
				params["page"] = strconv.Itoa(page)
			}
			if limit > 0 {
				params["limit"] = strconv.Itoa(limit)
			}

			data, err := app.Client.Get("/lp-positions/historical", params)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, []output.Column{
				{Header: "Position", Key: "tokenId", Width: 44},
				{Header: "Pair", Key: "pairName", Width: 12},
				{Header: "Input ($)", Key: "inputValue", Width: 14},
				{Header: "Output ($)", Key: "outputValue", Width: 14},
				{Header: "PnL %", Key: "pnl.percent", Width: 10},
				{Header: "Protocol", Key: "protocol", Width: 10},
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Wallet address")
	cmd.Flags().StringVar(&fromDate, "from", "", "Start date (ISO 8601)")
	cmd.Flags().StringVar(&toDate, "to", "", "End date (ISO 8601)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&limit, "limit", 0, "Items per page")
	return cmd
}

func newPositionsOverviewCmd() *cobra.Command {
	var (
		owner    string
		protocol string
		native   bool
	)

	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Get aggregated metrics for an owner's LP activity",
		Example: `  lpagent positions overview --owner 9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM
  lpagent positions overview --owner <addr> -o table --native`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())
			resolved, err := app.Config.ResolveOwner(owner)
			if err != nil {
				return err
			}

			data, err := app.Client.Get("/lp-positions/overview", map[string]string{
				"owner":    resolved,
				"protocol": protocol,
			})
			if err != nil {
				return err
			}

			if app.Format == "table" {
				printOverviewTable(data, native)
				return nil
			}

			output.Print(data, app.Format, nil)
			return nil
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Wallet address")
	cmd.Flags().StringVar(&protocol, "protocol", "meteora", "Protocol filter")
	cmd.Flags().BoolVar(&native, "native", false, "Show values in native token (SOL) instead of USD")
	return cmd
}

func printOverviewTable(data json.RawMessage, native bool) {
	var resp struct {
		Status string           `json:"status"`
		Data   []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil || len(resp.Data) == 0 {
		fmt.Fprintln(os.Stdout, "No overview data found.")
		return
	}

	d := resp.Data[0]

	tf := func(field string, timeframe string) float64 {
		if m, ok := d[field].(map[string]any); ok {
			if v, ok := output.ToFloatPublic(m[timeframe]); ok {
				return v
			}
		}
		if v, ok := output.ToFloatPublic(d[field]); ok {
			return v
		}
		return 0
	}

	fmtVal := func(v float64) string {
		if native {
			return fmt.Sprintf("%.2f SOL", v)
		}
		return output.FormatUSD(v)
	}

	var totalNetWorth, winRate, avgInvested, feeEarned, totalProfit, avgMonthly, expectedValue float64
	if native {
		totalNetWorth = tf("total_inflow_native", "") - tf("total_outflow_native", "") + tf("total_pnl_native", "ALL")
		winRate = tf("win_rate_native", "ALL") * 100
		avgInvested = tf("avg_inflow_native", "ALL")
		feeEarned = tf("total_fee_native", "ALL")
		totalProfit = tf("total_pnl_native", "ALL")
		avgMonthly = tf("avg_monthly_pnl_native", "")
		expectedValue = tf("expected_value_native", "ALL")
	} else {
		totalNetWorth = tf("total_inflow", "") - tf("total_outflow", "") + tf("total_pnl", "ALL")
		winRate = tf("win_rate", "ALL") * 100
		avgInvested = tf("avg_inflow", "ALL")
		feeEarned = tf("total_fee", "ALL")
		totalProfit = tf("total_pnl", "ALL")
		avgMonthly = tf("avg_monthly_pnl", "")
		expectedValue = tf("expected_value", "ALL")
	}
	closedLP := tf("closed_lp", "ALL")

	columns := []output.Column{
		{Header: "Metric", Width: 20},
		{Header: "Value", Width: 20},
	}

	rows := [][]string{
		{"Total Net Worth", fmtVal(totalNetWorth)},
		{"Total Closed", fmt.Sprintf("%.0f", closedLP)},
		{"Win Rate", output.ColorBySign(fmt.Sprintf("%.2f%%", winRate), winRate)},
		{"Avg Invested", fmtVal(avgInvested)},
		{"Fee Earned", fmtVal(feeEarned)},
		{"Total Profit", output.ColorBySign(fmtVal(totalProfit), totalProfit)},
		{"Avg Monthly Profit", output.ColorBySign(fmtVal(avgMonthly), avgMonthly)},
		{"Expected Value", output.ColorBySign(fmtVal(expectedValue), expectedValue)},
	}

	output.PrintRows(columns, rows)
}

func newPositionsLogsCmd() *cobra.Command {
	var (
		position string
		chain    string
		owner    string
	)

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get transaction logs for LP positions",
		Example: `  lpagent positions logs --position Ep22EwKegXis3bTC6P8JLgsHaT5J2beM2TncKe2Hmv24
  lpagent positions logs --owner 9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())

			if position == "" && owner == "" {
				// Try resolving default owner
				resolved, err := app.Config.ResolveOwner("")
				if err != nil {
					return fmt.Errorf("position or owner is required. Use --position or --owner")
				}
				owner = resolved
			}

			params := map[string]string{
				"position": position,
				"chain":    chain,
				"owner":    owner,
			}

			data, err := app.Client.Get("/lp-positions/logs", params)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, []output.Column{
				{Header: "Action", Key: "action", Width: 16},
				{Header: "Amount0", Key: "amount0", Width: 16},
				{Header: "Amount1", Key: "amount1", Width: 16},
				{Header: "Price0", Key: "price0", Width: 12},
				{Header: "Price1", Key: "price1", Width: 12},
				{Header: "Timestamp", Key: "timestamp", Width: 24},
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&position, "position", "", "Position ID (token ID)")
	cmd.Flags().StringVar(&chain, "chain", "SOL", "Blockchain chain")
	cmd.Flags().StringVar(&owner, "owner", "", "Wallet address")
	return cmd
}

func newPositionsGetCmd() *cobra.Command {
	var position string

	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get details for a specific position",
		Example: `  lpagent positions get --position Ep22EwKegXis3bTC6P8JLgsHaT5J2beM2TncKe2Hmv24`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if position == "" {
				return fmt.Errorf("--position is required")
			}

			app := appctx.FromContext(cmd.Context())

			data, err := app.Client.Get("/lp-positions/position", map[string]string{
				"position": position,
			})
			if err != nil {
				return err
			}

			output.Print(data, app.Format, nil)
			return nil
		},
	}

	cmd.Flags().StringVar(&position, "position", "", "Position ID (required)")
	return cmd
}

func newPositionsRevenueCmd() *cobra.Command {
	var (
		period   string
		rangeVal string
		protocol string
	)

	cmd := &cobra.Command{
		Use:   "revenue <owner>",
		Short: "Get revenue data for an owner",
		Args:  cobra.ExactArgs(1),
		Example: `  lpagent positions revenue 9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM
  lpagent positions revenue <addr> --range 7D`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())

			params := map[string]string{}
			if period != "" {
				params["period"] = period
			}
			if rangeVal != "" {
				params["range"] = rangeVal
			}
			if protocol != "" {
				params["protocol"] = protocol
			}

			data, err := app.Client.Get("/lp-positions/revenue/"+args[0], params)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, nil)
			return nil
		},
	}

	cmd.Flags().StringVar(&period, "period", "", "Period filter")
	cmd.Flags().StringVar(&rangeVal, "range", "", "Range filter: 7D, 1M")
	cmd.Flags().StringVar(&protocol, "protocol", "", "Protocol filter")
	return cmd
}
