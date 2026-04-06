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
	Short: "Show a user's posts",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserPosts,
}

var userCommentsCmd = &cobra.Command{
	Use:   "user-comments <username>",
	Short: "Show a user's comments",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserComments,
}

func init() {
	addListingFlags(popularCmd)
	addListingFlags(allCmd)
	addListingFlags(userPostsCmd)
	addListingFlags(userCommentsCmd)

	subCmd.Flags().Int("limit", 25, "Number of posts")
	subCmd.Flags().String("after", "", "Pagination cursor")
	subCmd.Flags().StringP("sort", "s", "hot", "Sort: hot|new|top|rising|controversial|best")
	subCmd.Flags().StringP("time", "t", "", "Time filter (for top/controversial): hour|day|week|month|year|all")
	subCmd.Flags().Bool("json", false, "Output as JSON")
	subCmd.Flags().Bool("yaml", false, "Output as YAML")

	subInfoCmd.Flags().Bool("json", false, "Output as JSON")
	subInfoCmd.Flags().Bool("yaml", false, "Output as YAML")

	userCmd.Flags().Bool("json", false, "Output as JSON")
	userCmd.Flags().Bool("yaml", false, "Output as YAML")

	browseCmd.AddCommand(popularCmd, allCmd, subCmd, subInfoCmd, userCmd, userPostsCmd, userCommentsCmd)
}

func addListingFlags(cmd *cobra.Command) {
	cmd.Flags().Int("limit", 25, "Number of posts")
	cmd.Flags().String("after", "", "Pagination cursor")
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("yaml", false, "Output as YAML")
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
	limit, _ := cmd.Flags().GetInt("limit")
	after, _ := cmd.Flags().GetString("after")

	params := map[string]string{"limit": fmt.Sprintf("%d", limit)}
	if after != "" {
		params["after"] = after
	}

	raw, err := c.Get(fmt.Sprintf("/user/%s/comments.json", args[0]), params)
	if err != nil {
		return handleErr(err, asJSON, asYAML)
	}

	m := client.CastMap(raw)
	data := client.CastMap(m["data"])
	children := client.CastSlice(data["children"])

	type commentEntry struct {
		Author  string  `json:"author" yaml:"author"`
		Body    string  `json:"body" yaml:"body"`
		Score   int     `json:"score" yaml:"score"`
		Created float64 `json:"created_utc" yaml:"created_utc"`
		Sub     string  `json:"subreddit" yaml:"subreddit"`
	}

	var entries []commentEntry
	for _, child := range children {
		cm := client.CastMap(child)
		d := client.CastMap(cm["data"])
		if d == nil {
			continue
		}
		entries = append(entries, commentEntry{
			Author:  client.CastString(d["author"]),
			Body:    client.CastString(d["body"]),
			Score:   client.CastInt(d["score"]),
			Created: client.CastFloat(d["created_utc"]),
			Sub:     client.CastString(d["subreddit"]),
		})
	}

	if asJSON {
		output.PrintJSON(entries)
	} else if asYAML {
		output.PrintYAML(entries)
	} else {
		for i, e := range entries {
			body := e.Body
			if len(body) > 120 {
				body = body[:117] + "..."
			}
			fmt.Printf("%d. [r/%s] %s  (%s)\n   %s\n\n",
				i+1, e.Sub, output.FormatScore(e.Score), output.FormatTime(e.Created), body)
		}
	}
	return nil
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

	raw, err := c.Get(path, params)
	if err != nil {
		return handleErr(err, asJSON, asYAML)
	}

	page := parser.ParseListing(raw)

	if err := cache.SaveIndex(page.Items, title); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save index cache: %v\n", err)
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
