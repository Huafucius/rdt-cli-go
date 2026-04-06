package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zhangtianhua/rdt-cli-go/internal/cache"
	"github.com/zhangtianhua/rdt-cli-go/internal/client"
	"github.com/zhangtianhua/rdt-cli-go/internal/output"
	"github.com/zhangtianhua/rdt-cli-go/internal/parser"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Reddit posts",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().StringP("sort", "s", "relevance", "Sort: relevance|hot|top|new|comments")
	searchCmd.Flags().StringP("time", "t", "all", "Time filter: hour|day|week|month|year|all")
	searchCmd.Flags().Int("limit", 25, "Number of results")
	searchCmd.Flags().String("after", "", "Pagination cursor")
	searchCmd.Flags().String("sub", "", "Restrict search to a subreddit")
	searchCmd.Flags().Bool("json", false, "Output as JSON")
	searchCmd.Flags().Bool("yaml", false, "Output as YAML")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	sort, _ := cmd.Flags().GetString("sort")
	timeFilter, _ := cmd.Flags().GetString("time")
	limit, _ := cmd.Flags().GetInt("limit")
	after, _ := cmd.Flags().GetString("after")
	sub, _ := cmd.Flags().GetString("sub")
	asJSON, _ := cmd.Flags().GetBool("json")
	asYAML, _ := cmd.Flags().GetBool("yaml")

	var path string
	if sub != "" {
		path = fmt.Sprintf("/r/%s/search.json", sub)
	} else {
		path = "/search.json"
	}

	params := map[string]string{
		"q":           query,
		"sort":        sort,
		"t":           timeFilter,
		"limit":       fmt.Sprintf("%d", limit),
		"restrict_sr": "off",
	}
	if sub != "" {
		params["restrict_sr"] = "on"
	}
	if after != "" {
		params["after"] = after
	}

	c := client.New()
	raw, err := c.Get(path, params)
	if err != nil {
		return handleErr(err, asJSON, asYAML)
	}

	page := parser.ParseListing(raw)

	if err := cache.SaveIndex(page.Items, "search:"+query); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save index cache: %v\n", err)
	}

	if asJSON {
		output.PrintJSON(page)
	} else if asYAML {
		output.PrintYAML(page)
	} else {
		title := fmt.Sprintf("Search: %q", query)
		if sub != "" {
			title += fmt.Sprintf(" in r/%s", sub)
		}
		output.PrintPostTable(page, title, true)
	}
	return nil
}
