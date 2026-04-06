package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zhangtianhua/rdt-cli-go/internal/models"
)

// WritePostsToFile writes posts to a file, format inferred from extension.
// Supported: .json, .csv (default: json)
func WritePostsToFile(posts []*models.Post, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return writePostsCSV(posts, path)
	default:
		return writePostsJSON(posts, path)
	}
}

func writePostsJSON(posts []*models.Post, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(posts)
}

func writePostsCSV(posts []*models.Post, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	w.Write([]string{"id", "name", "title", "subreddit", "author", "score", "num_comments", "created_utc", "permalink", "url", "is_self", "over_18"}) //nolint

	for _, p := range posts {
		w.Write([]string{ //nolint
			p.ID,
			p.Name,
			p.Title,
			p.Subreddit,
			p.Author,
			strconv.Itoa(p.Score),
			strconv.Itoa(p.NumComments),
			strconv.FormatFloat(p.CreatedUTC, 'f', 0, 64),
			p.Permalink,
			p.URL,
			strconv.FormatBool(p.IsSelf),
			strconv.FormatBool(p.Over18),
		})
	}
	return w.Error()
}

// WriteCommentsToFile writes comment entries to a file.
type CommentEntry struct {
	Author  string  `json:"author"`
	Body    string  `json:"body"`
	Score   int     `json:"score"`
	Created float64 `json:"created_utc"`
	Sub     string  `json:"subreddit"`
	PostID  string  `json:"post_id,omitempty"`
}

func WriteCommentsToFile(entries []CommentEntry, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return writeCommentsCSV(entries, path)
	default:
		return writeCommentsJSON(entries, path)
	}
}

func writeCommentsJSON(entries []CommentEntry, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

func writeCommentsCSV(entries []CommentEntry, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"author", "subreddit", "score", "created_utc", "post_id", "body"}) //nolint
	for _, e := range entries {
		w.Write([]string{ //nolint
			e.Author,
			e.Sub,
			strconv.Itoa(e.Score),
			strconv.FormatFloat(e.Created, 'f', 0, 64),
			e.PostID,
			e.Body,
		})
	}
	return w.Error()
}
