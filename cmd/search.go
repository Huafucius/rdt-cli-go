package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zhangtianhua/rdt-cli-go/internal/cache"
	"github.com/zhangtianhua/rdt-cli-go/internal/client"
	"github.com/zhangtianhua/rdt-cli-go/internal/models"
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
	searchCmd.Flags().Int("limit", 25, "Number of results per page")
	searchCmd.Flags().String("after", "", "Pagination cursor")
	searchCmd.Flags().String("sub", "", "Restrict search to a subreddit")
	searchCmd.Flags().Bool("json", false, "Output as JSON")
	searchCmd.Flags().Bool("yaml", false, "Output as YAML")
	searchCmd.Flags().Bool("all", false, "Fetch all pages (auto-paginate)")
	searchCmd.Flags().String("output", "", "Save results to file (.json or .csv)")
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
	fetchAll, _ := cmd.Flags().GetBool("all")
	outFile, _ := cmd.Flags().GetString("output")

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

	title := fmt.Sprintf("Search: %q", query)
	if sub != "" {
		title += fmt.Sprintf(" in r/%s", sub)
	}

	c := client.New()

	if fetchAll {
		fmt.Fprintf(os.Stderr, "fetching all results for %s...\n", title)
		posts, err := c.FetchAllPosts(path, params, 100, func(n int) {
			fmt.Fprintf(os.Stderr, "\r  %d results fetched...", n)
		})
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return handleErr(err, asJSON, asYAML)
		}

		if outFile != "" {
			if err := output.WritePostsToFile(posts, outFile); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
			fmt.Fprintf(os.Stderr, "saved %d results to %s\n", len(posts), outFile)
			return nil
		}

		if asJSON {
			output.PrintJSON(posts)
		} else if asYAML {
			output.PrintYAML(posts)
		} else {
			output.PrintPostTable(&models.ListingPage{Items: posts}, title, true)
		}
		return nil
	}

	// Single page
	raw, err := c.Get(path, params)
	if err != nil {
		return handleErr(err, asJSON, asYAML)
	}

	page := parser.ParseListing(raw)

	if err := cache.SaveIndex(page.Items, "search:"+query); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save index cache: %v\n", err)
	}

	if outFile != "" {
		if err := output.WritePostsToFile(page.Items, outFile); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "saved %d results to %s\n", len(page.Items), outFile)
		return nil
	}

	if asJSON {
		output.PrintJSON(page)
	} else if asYAML {
		output.PrintYAML(page)
	} else {
		output.PrintPostTable(page, title, true)
	}
	return nil
}
