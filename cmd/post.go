package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/zhangtianhua/rdt-cli-go/internal/cache"
	"github.com/zhangtianhua/rdt-cli-go/internal/client"
	"github.com/zhangtianhua/rdt-cli-go/internal/output"
	"github.com/zhangtianhua/rdt-cli-go/internal/parser"
)

var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Read posts and comments",
}

var readCmd = &cobra.Command{
	Use:   "read <post-id>",
	Short: "Read a post and its comments by Reddit post ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runRead,
}

var showCmd = &cobra.Command{
	Use:   "show <index>",
	Short: "Read a post by its index from the last listing",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	for _, cmd := range []*cobra.Command{readCmd, showCmd} {
		cmd.Flags().StringP("sort", "s", "best", "Comment sort: best|top|new|controversial|old|qa")
		cmd.Flags().Int("limit", 25, "Number of top-level comments")
		cmd.Flags().Bool("json", false, "Output as JSON")
		cmd.Flags().Bool("yaml", false, "Output as YAML")
	}
	postCmd.AddCommand(readCmd, showCmd)
}

func runRead(cmd *cobra.Command, args []string) error {
	return fetchAndPrintPost(cmd, args[0])
}

func runShow(cmd *cobra.Command, args []string) error {
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("index must be a number, got %q", args[0])
	}
	post, err := cache.GetByIndex(n)
	if err != nil {
		return err
	}
	return fetchAndPrintPost(cmd, post.ID)
}

func fetchAndPrintPost(cmd *cobra.Command, postID string) error {
	c := client.New()
	sort, _ := cmd.Flags().GetString("sort")
	limit, _ := cmd.Flags().GetInt("limit")
	asJSON, _ := cmd.Flags().GetBool("json")
	asYAML, _ := cmd.Flags().GetBool("yaml")

	params := map[string]string{
		"sort":  sort,
		"limit": fmt.Sprintf("%d", limit),
	}

	raw, err := c.Get(fmt.Sprintf("/comments/%s.json", postID), params)
	if err != nil {
		return handleErr(err, asJSON, asYAML)
	}

	detail := parser.ParsePostDetail(raw)

	if asJSON {
		output.PrintJSON(detail)
	} else if asYAML {
		output.PrintYAML(detail)
	} else {
		output.PrintPostDetail(detail)
	}
	return nil
}
