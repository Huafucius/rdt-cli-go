package client

import (
	"fmt"

	"github.com/zhangtianhua/rdt-cli-go/internal/models"
)

const maxPagesDefault = 200 // safety cap (~20,000 items at 100/page)

// FetchAllPosts fetches every page of a listing endpoint until exhausted.
func (c *Client) FetchAllPosts(path string, baseParams map[string]string, pageSize int, onProgress func(n int)) ([]*models.Post, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 100
	}

	params := make(map[string]string, len(baseParams)+2)
	for k, v := range baseParams {
		params[k] = v
	}
	params["limit"] = fmt.Sprintf("%d", pageSize)

	var all []*models.Post
	after := ""

	for page := 0; page < maxPagesDefault; page++ {
		if after != "" {
			params["after"] = after
		} else {
			delete(params, "after")
		}

		raw, err := c.Get(path, params)
		if err != nil {
			return all, fmt.Errorf("page %d: %w", page+1, err)
		}

		// Inline listing parse to avoid circular dependency with parser
		m := CastMap(raw)
		data := CastMap(m["data"])
		children := CastSlice(data["children"])

		if len(children) == 0 {
			break
		}

		for _, child := range children {
			cm := CastMap(child)
			if CastString(cm["kind"]) != "t3" {
				continue
			}
			d := CastMap(cm["data"])
			if d == nil {
				continue
			}
			all = append(all, &models.Post{
				ID:          CastString(d["id"]),
				Name:        CastString(d["name"]),
				Title:       CastString(d["title"]),
				Subreddit:   CastString(d["subreddit"]),
				Author:      CastString(d["author"]),
				Score:       CastInt(d["score"]),
				NumComments: CastInt(d["num_comments"]),
				CreatedUTC:  CastFloat(d["created_utc"]),
				Permalink:   CastString(d["permalink"]),
				URL:         CastString(d["url"]),
				Selftext:    CastString(d["selftext"]),
				IsSelf:      CastBool(d["is_self"]),
				Over18:      CastBool(d["over_18"]),
				IsVideo:     CastBool(d["is_video"]),
				Stickied:    CastBool(d["stickied"]),
			})
		}

		if onProgress != nil {
			onProgress(len(all))
		}

		after = CastString(data["after"])
		if after == "" {
			break
		}
	}

	return all, nil
}

// FetchAllComments fetches every page of a user comments endpoint.
func (c *Client) FetchAllComments(path string, baseParams map[string]string, pageSize int, onProgress func(n int)) ([]map[string]any, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 100
	}

	params := make(map[string]string, len(baseParams)+2)
	for k, v := range baseParams {
		params[k] = v
	}
	params["limit"] = fmt.Sprintf("%d", pageSize)

	var all []map[string]any
	after := ""

	for page := 0; page < maxPagesDefault; page++ {
		if after != "" {
			params["after"] = after
		} else {
			delete(params, "after")
		}

		raw, err := c.Get(path, params)
		if err != nil {
			return all, fmt.Errorf("page %d: %w", page+1, err)
		}

		m := CastMap(raw)
		data := CastMap(m["data"])
		children := CastSlice(data["children"])

		if len(children) == 0 {
			break
		}

		for _, child := range children {
			cm := CastMap(child)
			d := CastMap(cm["data"])
			if d != nil {
				all = append(all, d)
			}
		}

		if onProgress != nil {
			onProgress(len(all))
		}

		after = CastString(data["after"])
		if after == "" {
			break
		}
	}

	return all, nil
}
