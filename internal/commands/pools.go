package commands

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/lpagent/cli/internal/appctx"
	"github.com/lpagent/cli/internal/output"
)

func NewPoolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Discover and inspect liquidity pools",
	}

	cmd.AddCommand(
		newPoolsDiscoverCmd(),
		newPoolsInfoCmd(),
		newPoolsPositionsCmd(),
		newPoolsOnchainStatsCmd(),
		newPoolsTopLPersCmd(),
		newPoolsAddTxCmd(),
		newPoolsLandingAddTxCmd(),
	)

	return cmd
}

func newPoolsDiscoverCmd() *cobra.Command {
	var (
		chain          string
		sortBy         string
		sortOrder      string
		page           int
		pageSize       int
		feeTVLInterval string
		quoteToken     string
		minMarketCap   string
		maxMarketCap   string
		minBinStep     string
		maxBinStep     string
		minBaseFee     string
		maxBaseFee     string
		minLiquidity   string
		maxLiquidity   string
		min24hFees     string
		max24hFees     string
		min24hVol      string
		max24hVol      string
		poolType       string
		search         string
	)

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover and filter liquidity pools",
		Example: `  lpagent pools discover
  lpagent pools discover --chain SOL --sort-by tvl --page-size 20 -o table
  lpagent pools discover --search "SOL" --min-liquidity 10000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())

			params := map[string]string{
				"chain":          chain,
				"sortBy":         sortBy,
				"sortOrder":      sortOrder,
				"feeTVLInterval": feeTVLInterval,
			}
			if page > 0 {
				params["page"] = strconv.Itoa(page)
			}
			if pageSize > 0 {
				params["pageSize"] = strconv.Itoa(pageSize)
			}
			if quoteToken != "" {
				params["quote_token"] = quoteToken
			}
			if minMarketCap != "" {
				params["min_market_cap"] = minMarketCap
			}
			if maxMarketCap != "" {
				params["max_market_cap"] = maxMarketCap
			}
			if minBinStep != "" {
				params["min_bin_step"] = minBinStep
			}
			if maxBinStep != "" {
				params["max_bin_step"] = maxBinStep
			}
			if minBaseFee != "" {
				params["min_base_fee"] = minBaseFee
			}
			if maxBaseFee != "" {
				params["max_base_fee"] = maxBaseFee
			}
			if minLiquidity != "" {
				params["min_liquidity"] = minLiquidity
			}
			if maxLiquidity != "" {
				params["max_liquidity"] = maxLiquidity
			}
			if min24hFees != "" {
				params["min_24h_fees"] = min24hFees
			}
			if max24hFees != "" {
				params["max_24h_fees"] = max24hFees
			}
			if min24hVol != "" {
				params["min_24h_vol"] = min24hVol
			}
			if max24hVol != "" {
				params["max_24h_vol"] = max24hVol
			}
			if poolType != "" {
				params["type"] = poolType
			}
			if search != "" {
				params["search"] = search
			}

			data, err := app.Client.Get("/pools/discover", params)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, []output.Column{
				{Header: "Pool", Key: "pool", Width: 44},
				{Header: "Pair", Key: "pairName", Width: 16},
				{Header: "TVL", Key: "tvl", Width: 14},
				{Header: "Vol 24h", Key: "vol_24h", Width: 14},
				{Header: "Fee/TVL", Key: "fee_tvl_ratio", Width: 10},
				{Header: "Protocol", Key: "protocol", Width: 12},
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&chain, "chain", "SOL", "Blockchain network")
	cmd.Flags().StringVar(&sortBy, "sort-by", "mcap", "Sort by: mcap, created_at, vol_24h, tvl, fee_tvl_ratio, volatility")
	cmd.Flags().StringVar(&sortOrder, "sort-order", "desc", "Sort order: asc, desc")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 10, "Items per page (max 100)")
	cmd.Flags().StringVar(&feeTVLInterval, "fee-tvl-interval", "24h", "Fee/TVL interval: 5m, 1h, 6h, 24h")
	cmd.Flags().StringVar(&quoteToken, "quote-token", "", "Quote token address filter")
	cmd.Flags().StringVar(&minMarketCap, "min-market-cap", "", "Minimum market cap")
	cmd.Flags().StringVar(&maxMarketCap, "max-market-cap", "", "Maximum market cap")
	cmd.Flags().StringVar(&minBinStep, "min-bin-step", "", "Minimum bin step")
	cmd.Flags().StringVar(&maxBinStep, "max-bin-step", "", "Maximum bin step")
	cmd.Flags().StringVar(&minBaseFee, "min-base-fee", "", "Minimum base fee")
	cmd.Flags().StringVar(&maxBaseFee, "max-base-fee", "", "Maximum base fee")
	cmd.Flags().StringVar(&minLiquidity, "min-liquidity", "", "Minimum liquidity")
	cmd.Flags().StringVar(&maxLiquidity, "max-liquidity", "", "Maximum liquidity")
	cmd.Flags().StringVar(&min24hFees, "min-24h-fees", "", "Minimum 24h fees")
	cmd.Flags().StringVar(&max24hFees, "max-24h-fees", "", "Maximum 24h fees")
	cmd.Flags().StringVar(&min24hVol, "min-24h-vol", "", "Minimum 24h volume")
	cmd.Flags().StringVar(&max24hVol, "max-24h-vol", "", "Maximum 24h volume")
	cmd.Flags().StringVar(&poolType, "type", "", "Pool type filter")
	cmd.Flags().StringVar(&search, "search", "", "Search term")

	return cmd
}

func newPoolsInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "info <poolId>",
		Short:   "Get detailed information for a pool",
		Args:    cobra.ExactArgs(1),
		Example: `  lpagent pools info 2DeF1QHAQMpNXCGjcsm2pWw1V4KknGtwd2wEh2fTriKC`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())

			data, err := app.Client.Get("/pools/"+args[0]+"/info", nil)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, nil)
			return nil
		},
	}
}

func newPoolsPositionsCmd() *cobra.Command {
	var (
		owner     string
		status    string
		page      int
		pageSize  int
		orderBy   string
		sortOrder string
		platform  string
	)

	cmd := &cobra.Command{
		Use:   "positions <poolId>",
		Short: "List positions in a specific pool",
		Args:  cobra.ExactArgs(1),
		Example: `  lpagent pools positions 2DeF1QHAQMpNXCGjcsm2pWw1V4KknGtwd2wEh2fTriKC
  lpagent pools positions <poolId> --owner <addr> --status Open -o table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())

			params := map[string]string{}
			if owner != "" {
				params["owner"] = owner
			}
			if status != "" {
				params["status"] = status
			}
			if page > 0 {
				params["page"] = strconv.Itoa(page)
			}
			if pageSize > 0 {
				params["pageSize"] = strconv.Itoa(pageSize)
			}
			if orderBy != "" {
				params["order_by"] = orderBy
			}
			if sortOrder != "" {
				params["sort_order"] = sortOrder
			}
			if platform != "" {
				params["platform"] = platform
			}

			data, err := app.Client.Get("/pools/"+args[0]+"/positions", params)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, []output.Column{
				{Header: "Position", Key: "tokenId", Width: 44},
				{Header: "Owner", Key: "owner", Width: 44},
				{Header: "Value ($)", Key: "currentValue", Width: 14},
				{Header: "PnL %", Key: "pnl.percent", Width: 10},
				{Header: "Status", Key: "status", Width: 8},
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Filter by wallet address")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: Open, Close")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Items per page (max 20)")
	cmd.Flags().StringVar(&orderBy, "order-by", "", "Order by field (default: inputNative)")
	cmd.Flags().StringVar(&sortOrder, "sort-order", "", "Sort order: asc, desc")
	cmd.Flags().StringVar(&platform, "platform", "", "Platform filter")

	return cmd
}

func newPoolsOnchainStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "onchain-stats <poolId>",
		Short:   "Get onchain statistics for a pool",
		Args:    cobra.ExactArgs(1),
		Example: `  lpagent pools onchain-stats 2DeF1QHAQMpNXCGjcsm2pWw1V4KknGtwd2wEh2fTriKC`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())

			data, err := app.Client.Get("/pools/"+args[0]+"/onchain-stats", nil)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, nil)
			return nil
		},
	}
}

func newPoolsTopLPersCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "top-lpers <poolId>",
		Short:   "Get top liquidity providers for a pool",
		Args:    cobra.ExactArgs(1),
		Example: `  lpagent pools top-lpers 2DeF1QHAQMpNXCGjcsm2pWw1V4KknGtwd2wEh2fTriKC -o table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appctx.FromContext(cmd.Context())

			data, err := app.Client.Get("/pools/"+args[0]+"/top-lpers", nil)
			if err != nil {
				return err
			}

			output.Print(data, app.Format, []output.Column{
				{Header: "Owner", Key: "owner", Width: 44},
				{Header: "Value ($)", Key: "totalValue", Width: 14},
				{Header: "Positions", Key: "positionCount", Width: 10},
			})
			return nil
		},
	}
}
