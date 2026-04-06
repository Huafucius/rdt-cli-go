package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zhangtianhua/rdt-cli-go/internal/models"
)

var cacheFile = filepath.Join(os.Getenv("HOME"), ".config", "rdt-cli", "index_cache.json")

type indexCache struct {
	Source  string         `json:"source"`
	SavedAt float64        `json:"saved_at"`
	Count   int            `json:"count"`
	Items   []*models.Post `json:"items"`
}

// SaveIndex writes a listing to the index cache.
func SaveIndex(posts []*models.Post, source string) error {
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0o700); err != nil {
		return fmt.Errorf("mkdir cache dir: %w", err)
	}
	c := indexCache{
		Source:  source,
		SavedAt: float64(time.Now().Unix()),
		Count:   len(posts),
		Items:   posts,
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	if err := os.WriteFile(cacheFile, data, 0o600); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}

// GetByIndex returns a post by 1-based index from the cache.
func GetByIndex(n int) (*models.Post, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("no index cache (run a listing command first)")
	}
	var c indexCache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("corrupt cache: %w", err)
	}
	if n < 1 || n > len(c.Items) {
		return nil, fmt.Errorf("index %d out of range (cache has %d items)", n, len(c.Items))
	}
	return c.Items[n-1], nil
}
