package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lpagent/cli/internal/appctx"
	"github.com/lpagent/cli/internal/output"
)

func NewAPICmd() *cobra.Command {
	var (
		query string
		data  string
	)

	cmd := &cobra.Command{
		Use:   "api <method> <path>",
		Short: "Raw API access to any endpoint",
		Long: `Send raw HTTP requests to the LP Agent Open API.
Supports GET, POST, PUT, DELETE methods.`,
		Args: cobra.ExactArgs(2),
		Example: `  lpagent api get /lp-positions/opening --query "owner=9WzDX..."
  lpagent api post /position/decrease-quotes --data '{"id":"...","bps":5000}'
  lpagent api get /pools/discover --query "chain=SOL&sortBy=tvl"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			path := args[1]

			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			app := appctx.FromContext(cmd.Context())

			// Parse query params
			params := map[string]string{}
			if query != "" {
				for _, pair := range strings.Split(query, "&") {
					parts := strings.SplitN(pair, "=", 2)
					if len(parts) == 2 {
						params[parts[0]] = parts[1]
					}
				}
			}

			var result []byte
			var err error

			switch method {
			case "GET":
				result, err = app.Client.Get(path, params)
			case "POST", "PUT", "DELETE":
				var body any
				if data != "" {
					body, err = parseJSONString(data)
					if err != nil {
						return fmt.Errorf("invalid --data JSON: %w", err)
					}
				}
				result, err = app.Client.Post(path, body)
			default:
				return fmt.Errorf("unsupported method: %s (use GET, POST, PUT, DELETE)", method)
			}

			if err != nil {
				return err
			}

			output.Print(result, app.Format, nil)
			return nil
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Query parameters (key=value&key2=value2)")
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")

	return cmd
}

func parseJSONString(s string) (any, error) {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	return v, nil
}
