package output

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/zhangtianhua/rdt-cli-go/internal/models"
)

// Envelope wraps all output (schema v1, compatible with Python rdt-cli).
type Envelope struct {
	OK            bool   `json:"ok" yaml:"ok"`
	SchemaVersion string `json:"schema_version" yaml:"schema_version"`
	Data          any    `json:"data,omitempty" yaml:"data,omitempty"`
	Error         *ErrPayload `json:"error,omitempty" yaml:"error,omitempty"`
}

type ErrPayload struct {
	Code    string `json:"code" yaml:"code"`
	Message string `json:"message" yaml:"message"`
}

func successEnvelope(data any) Envelope {
	return Envelope{OK: true, SchemaVersion: "1", Data: data}
}

func errorEnvelope(code, message string) Envelope {
	return Envelope{OK: false, SchemaVersion: "1", Error: &ErrPayload{Code: code, Message: message}}
}

// PrintJSON prints a success envelope as JSON.
func PrintJSON(data any) {
	printEnvelopeJSON(successEnvelope(data))
}

// PrintYAML prints a success envelope as YAML.
func PrintYAML(data any) {
	printEnvelopeYAML(successEnvelope(data))
}

// PrintErrorJSON prints an error envelope as JSON.
func PrintErrorJSON(code, message string) {
	printEnvelopeJSON(errorEnvelope(code, message))
}

// PrintErrorYAML prints an error envelope as YAML.
func PrintErrorYAML(code, message string) {
	printEnvelopeYAML(errorEnvelope(code, message))
}

func printEnvelopeJSON(e Envelope) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(e) //nolint
}

func printEnvelopeYAML(e Envelope) {
	yaml.NewEncoder(os.Stdout).Encode(e) //nolint
}

// FormatScore converts an integer score to a human-readable string.
func FormatScore(score int) string {
	if score >= 1000 || score <= -1000 {
		return fmt.Sprintf("%.1fk", float64(score)/1000)
	}
	return fmt.Sprintf("%d", score)
}

// FormatTime converts a Unix UTC timestamp to a relative time string.
func FormatTime(ts float64) string {
	t := time.Unix(int64(ts), 0)
	dur := time.Since(t)

	switch {
	case dur < time.Minute:
		return "just now"
	case dur < time.Hour:
		return fmt.Sprintf("%dm ago", int(dur.Minutes()))
	case dur < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(dur.Hours()))
	case dur < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(dur.Hours()/24))
	case dur < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(dur.Hours()/(24*7)))
	default:
		months := int(math.Round(dur.Hours() / (24 * 30)))
		if months < 12 {
			return fmt.Sprintf("%dmo ago", months)
		}
		return fmt.Sprintf("%dy ago", int(dur.Hours()/(24*365)))
	}
}

// PrintPostTable prints a listing of posts as a text table with 1-based index.
func PrintPostTable(page *models.ListingPage, title string, showSubreddit bool) {
	if title != "" {
		fmt.Println(title)
		fmt.Println(strings.Repeat("─", 60))
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "#\tScore\tComments\tAge\tSubreddit\tTitle")
	fmt.Fprintln(w, "─\t─────\t────────\t───\t─────────\t─────")

	for i, p := range page.Items {
		sub := ""
		if showSubreddit {
			sub = p.Subreddit
		}
		title := p.Title
		if len(title) > 80 {
			title = title[:77] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\t%s\n",
			i+1,
			FormatScore(p.Score),
			p.NumComments,
			FormatTime(p.CreatedUTC),
			sub,
			title,
		)
	}
	w.Flush()

	if page.After != "" {
		fmt.Printf("\n(more available — use --after %s)\n", page.After)
	}
}

// PrintPostDetail prints a post and its comment tree.
func PrintPostDetail(detail *models.PostDetail) {
	if detail.Post == nil {
		fmt.Println("(no post data)")
		return
	}
	p := detail.Post

	fmt.Println(strings.Repeat("═", 70))
	fmt.Printf("  %s\n", p.Title)
	fmt.Printf("  r/%s  •  u/%s  •  %s  •  %s  •  %d comments\n",
		p.Subreddit, p.Author, FormatScore(p.Score), FormatTime(p.CreatedUTC), p.NumComments)
	fmt.Println(strings.Repeat("═", 70))

	if p.IsSelf && p.Selftext != "" {
		body := p.Selftext
		if len(body) > 2000 {
			body = body[:2000] + "\n[truncated]"
		}
		fmt.Println(body)
		fmt.Println()
	} else if !p.IsSelf {
		fmt.Printf("  Link: %s\n\n", p.URL)
	}

	if len(detail.Comments) == 0 {
		fmt.Println("(no comments)")
		return
	}

	fmt.Printf("── Comments (%d", len(detail.Comments))
	if detail.MoreCount > 0 {
		fmt.Printf(" + %d more", detail.MoreCount)
	}
	fmt.Println(") ──")
	printComments(detail.Comments, 0, 3)
}

func printComments(comments []*models.Comment, depth, maxDepth int) {
	if depth >= maxDepth {
		return
	}
	indent := strings.Repeat("  │", depth) + "  "
	for _, c := range comments {
		if c.Author == "" {
			continue
		}
		body := c.Body
		if len(body) > 500 {
			body = body[:497] + "..."
		}
		// Replace newlines with inline breaks
		body = strings.ReplaceAll(body, "\n", " ")

		fmt.Printf("%s▸ u/%s  %s  %s\n", indent, c.Author, FormatScore(c.Score), FormatTime(c.CreatedUTC))
		fmt.Printf("%s  %s\n", indent, body)

		if len(c.Replies) > 0 {
			printComments(c.Replies, depth+1, maxDepth)
		}
	}
}

// PrintUserProfile prints a user profile.
func PrintUserProfile(u *models.UserProfile) {
	fmt.Printf("u/%s\n", u.Name)
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("Post karma:    %s\n", FormatScore(u.LinkKarma))
	fmt.Printf("Comment karma: %s\n", FormatScore(u.CommentKarma))
	fmt.Printf("Account age:   %s\n", FormatTime(u.CreatedUTC))
	if u.IsGold {
		fmt.Println("Gold:          ✓")
	}
}

// PrintSubredditInfo prints subreddit metadata.
func PrintSubredditInfo(s *models.SubredditInfo) {
	fmt.Printf("%s\n", s.DisplayNamePrefixed)
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("Subscribers:  %s\n", FormatScore(s.Subscribers))
	if s.AccountsActive > 0 {
		fmt.Printf("Online:       %s\n", FormatScore(s.AccountsActive))
	}
	fmt.Printf("Created:      %s\n", FormatTime(s.CreatedUTC))
	if s.Over18 {
		fmt.Println("NSFW:         yes")
	}
	if s.PublicDescription != "" {
		fmt.Printf("\n%s\n", s.PublicDescription)
	}
}
