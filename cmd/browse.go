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

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse Reddit feeds and subreddits",
}

var popularCmd = &cobra.Command{
	Use:   "popular",
	Short: "Browse /r/popular",
	RunE:  runPopular,
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Browse /r/all",
	RunE:  runAll,
}

var subCmd = &cobra.Command{
	Use:   "sub <subreddit>",
	Short: "Browse a subreddit",
	Args:  cobra.ExactArgs(1),
	RunE:  runSub,
}

var subInfoCmd = &cobra.Command{
	Use:   "sub-info <subreddit>",
	Short: "Show subreddit information",
	Args:  cobra.ExactArgs(1),
	RunE:  runSubInfo,
}

var userCmd = &cobra.Command{
	Use:   "user <username>",
	Short: "Show user profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runUser,
}

var userPostsCmd = &cobra.Command{
	Use:   "user-posts <username>",
	Short: "Fetch a user's posts",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserPosts,
}

var userCommentsCmd = &cobra.Command{
	Use:   "user-comments <username>",
	Short: "Fetch a user's comments",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserComments,
}

func init() {
	addListingFlags(popularCmd)
	addListingFlags(allCmd)
	addListingFlags(userPostsCmd)
	addListingFlags(userCommentsCmd)

	subCmd.Flags().Int("limit", 25, "Number of posts per page")
	subCmd.Flags().String("after", "", "Pagination cursor")
	subCmd.Flags().StringP("sort", "s", "hot", "Sort: hot|new|top|rising|controversial|best")
	subCmd.Flags().StringP("time", "t", "", "Time filter (for top/controversial): hour|day|week|month|year|all")
	subCmd.Flags().Bool("json", false, "Output as JSON")
	subCmd.Flags().Bool("yaml", false, "Output as YAML")
	subCmd.Flags().Bool("all", false, "Fetch all pages (auto-paginate)")
	subCmd.Flags().String("output", "", "Save results to file (.json or .csv)")

	subInfoCmd.Flags().Bool("json", false, "Output as JSON")
	subInfoCmd.Flags().Bool("yaml", false, "Output as YAML")

	userCmd.Flags().Bool("json", false, "Output as JSON")
	userCmd.Flags().Bool("yaml", false, "Output as YAML")

	browseCmd.AddCommand(popularCmd, allCmd, subCmd, subInfoCmd, userCmd, userPostsCmd, userCommentsCmd)
}

func addListingFlags(cmd *cobra.Command) {
	cmd.Flags().Int("limit", 25, "Number of posts per page (max 100)")
	cmd.Flags().String("after", "", "Pagination cursor")
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("yaml", false, "Output as YAML")
	cmd.Flags().Bool("all", false, "Fetch all pages (auto-paginate)")
	cmd.Flags().String("output", "", "Save results to file (.json or .csv)")
}

// ── handlers ─────────────────────────────────────────────────────────

func runPopular(cmd *cobra.Command, _ []string) error {
	return runListing(cmd, "/r/popular.json", "r/popular", true)
}

func runAll(cmd *cobra.Command, _ []string) error {
	return runListing(cmd, "/r/all.json", "r/all", true)
}

func runSub(cmd *cobra.Command, args []string) error {
	sub := args[0]
	sort, _ := cmd.Flags().GetString("sort")
	timeFilter, _ := cmd.Flags().GetString("time")

	var path string
	if sort == "" || sort == "hot" {
		path = fmt.Sprintf("/r/%s.json", sub)
	} else {
		path = fmt.Sprintf("/r/%s/%s.json", sub, sort)
	}

	params := buildListingParams(cmd)
	if timeFilter != "" {
		params["t"] = timeFilter
	}

	return runListingWithPath(cmd, path, params, "r/"+sub, false)
}

func runSubInfo(cmd *cobra.Command, args []string) error {
	c := client.New()
	asJSON, _ := cmd.Flags().GetBool("json")
	asYAML, _ := cmd.Flags().GetBool("yaml")

	raw, err := c.Get(fmt.Sprintf("/r/%s/about.json", args[0]), nil)
	if err != nil {
		return handleErr(err, asJSON, asYAML)
	}
	info := parser.ParseSubredditInfo(raw)

	if asJSON {
		output.PrintJSON(info)
	} else if asYAML {
		output.PrintYAML(info)
	} else {
		output.PrintSubredditInfo(info)
	}
	return nil
}

func runUser(cmd *cobra.Command, args []string) error {
	c := client.New()
	asJSON, _ := cmd.Flags().GetBool("json")
	asYAML, _ := cmd.Flags().GetBool("yaml")

	raw, err := c.Get(fmt.Sprintf("/user/%s/about.json", args[0]), nil)
	if err != nil {
		return handleErr(err, asJSON, asYAML)
	}
	profile := parser.ParseUserProfile(raw)

	if asJSON {
		output.PrintJSON(profile)
	} else if asYAML {
		output.PrintYAML(profile)
	} else {
		output.PrintUserProfile(profile)
	}
	return nil
}

func runUserPosts(cmd *cobra.Command, args []string) error {
	return runListing(cmd, fmt.Sprintf("/user/%s/submitted.json", args[0]), args[0]+" posts", false)
}

func runUserComments(cmd *cobra.Command, args []string) error {
	c := client.New()
	asJSON, _ := cmd.Flags().GetBool("json")
	asYAML, _ := cmd.Flags().GetBool("yaml")
	fetchAll, _ := cmd.Flags().GetBool("all")
	outFile, _ := cmd.Flags().GetString("output")
	limit, _ := cmd.Flags().GetInt("limit")
	after, _ := cmd.Flags().GetString("after")

	path := fmt.Sprintf("/user/%s/comments.json", args[0])
	baseParams := map[string]string{}
	if after != "" {
		baseParams["after"] = after
	}

	if fetchAll {
		fmt.Fprintf(os.Stderr, "fetching all comments for u/%s...\n", args[0])
		rawEntries, err := c.FetchAllComments(path, baseParams, 100, func(n int) {
			fmt.Fprintf(os.Stderr, "\r  %d comments fetched...", n)
		})
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return handleErr(err, asJSON, asYAML)
		}

		entries := make([]output.CommentEntry, 0, len(rawEntries))
		for _, d := range rawEntries {
			entries = append(entries, output.CommentEntry{
				Author:  client.CastString(d["author"]),
				Body:    client.CastString(d["body"]),
				Score:   client.CastInt(d["score"]),
				Created: client.CastFloat(d["created_utc"]),
				Sub:     client.CastString(d["subreddit"]),
				PostID:  client.CastString(d["link_id"]),
			})
		}

		if outFile != "" {
			if err := output.WriteCommentsToFile(entries, outFile); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
			fmt.Fprintf(os.Stderr, "saved %d comments to %s\n", len(entries), outFile)
			return nil
		}

		if asJSON {
			output.PrintJSON(entries)
		} else if asYAML {
			output.PrintYAML(entries)
		} else {
			printCommentEntries(entries)
		}
		return nil
	}

	// Single page
	params := map[string]string{"limit": fmt.Sprintf("%d", limit)}
	if after != "" {
		params["after"] = after
	}
	raw, err := c.Get(path, params)
	if err != nil {
		return handleErr(err, asJSON, asYAML)
	}

	m := client.CastMap(raw)
	data := client.CastMap(m["data"])
	children := client.CastSlice(data["children"])

	var entries []output.CommentEntry
	for _, child := range children {
		cm := client.CastMap(child)
		d := client.CastMap(cm["data"])
		if d == nil {
			continue
		}
		entries = append(entries, output.CommentEntry{
			Author:  client.CastString(d["author"]),
			Body:    client.CastString(d["body"]),
			Score:   client.CastInt(d["score"]),
			Created: client.CastFloat(d["created_utc"]),
			Sub:     client.CastString(d["subreddit"]),
			PostID:  client.CastString(d["link_id"]),
		})
	}

	if outFile != "" {
		if err := output.WriteCommentsToFile(entries, outFile); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "saved %d comments to %s\n", len(entries), outFile)
		return nil
	}

	if asJSON {
		output.PrintJSON(entries)
	} else if asYAML {
		output.PrintYAML(entries)
	} else {
		printCommentEntries(entries)
	}
	return nil
}

func printCommentEntries(entries []output.CommentEntry) {
	for i, e := range entries {
		body := e.Body
		if len(body) > 120 {
			body = body[:117] + "..."
		}
		fmt.Printf("%d. [r/%s] %s  (%s)\n   %s\n\n",
			i+1, e.Sub, output.FormatScore(e.Score), output.FormatTime(e.Created), body)
	}
}

// ── helpers ───────────────────────────────────────────────────────────

func buildListingParams(cmd *cobra.Command) map[string]string {
	limit, _ := cmd.Flags().GetInt("limit")
	after, _ := cmd.Flags().GetString("after")
	params := map[string]string{"limit": fmt.Sprintf("%d", limit)}
	if after != "" {
		params["after"] = after
	}
	return params
}

func runListing(cmd *cobra.Command, path, title string, showSub bool) error {
	return runListingWithPath(cmd, path, buildListingParams(cmd), title, showSub)
}

func runListingWithPath(cmd *cobra.Command, path string, params map[string]string, title string, showSub bool) error {
	c := client.New()
	asJSON, _ := cmd.Flags().GetBool("json")
	asYAML, _ := cmd.Flags().GetBool("yaml")
	fetchAll, _ := cmd.Flags().GetBool("all")
	outFile, _ := cmd.Flags().GetString("output")

	if fetchAll {
		fmt.Fprintf(os.Stderr, "fetching all posts from %s...\n", title)
		posts, err := c.FetchAllPosts(path, params, 100, func(n int) {
			fmt.Fprintf(os.Stderr, "\r  %d posts fetched...", n)
		})
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return handleErr(err, asJSON, asYAML)
		}

		if outFile != "" {
			if err := output.WritePostsToFile(posts, outFile); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
			fmt.Fprintf(os.Stderr, "saved %d posts to %s\n", len(posts), outFile)
			return nil
		}

		if asJSON {
			output.PrintJSON(posts)
		} else if asYAML {
			output.PrintYAML(posts)
		} else {
			output.PrintPostTable(&models.ListingPage{Items: posts}, title, showSub)
		}
		return nil
	}

	// Single page
	raw, err := c.Get(path, params)
	if err != nil {
		return handleErr(err, asJSON, asYAML)
	}

	page := parser.ParseListing(raw)

	if err := cache.SaveIndex(page.Items, title); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save index cache: %v\n", err)
	}

	if outFile != "" {
		if err := output.WritePostsToFile(page.Items, outFile); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "saved %d posts to %s\n", len(page.Items), outFile)
		return nil
	}

	if asJSON {
		output.PrintJSON(page)
	} else if asYAML {
		output.PrintYAML(page)
	} else {
		output.PrintPostTable(page, title, showSub)
	}
	return nil
}

func handleErr(err error, asJSON, asYAML bool) error {
	if asJSON {
		output.PrintErrorJSON("api_error", err.Error())
		return nil
	}
	if asYAML {
		output.PrintErrorYAML("api_error", err.Error())
		return nil
	}
	return err
}
