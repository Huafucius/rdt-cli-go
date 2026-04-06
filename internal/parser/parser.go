package parser

import (
	"github.com/zhangtianhua/rdt-cli-go/internal/client"
	"github.com/zhangtianhua/rdt-cli-go/internal/models"
)

// ParsePost builds a Post from a raw Reddit post payload.
func ParsePost(data map[string]any) *models.Post {
	return &models.Post{
		ID:          client.CastString(data["id"]),
		Name:        client.CastString(data["name"]),
		Title:       client.CastString(data["title"]),
		Subreddit:   client.CastString(data["subreddit"]),
		Author:      client.CastString(data["author"]),
		Score:       client.CastInt(data["score"]),
		NumComments: client.CastInt(data["num_comments"]),
		CreatedUTC:  client.CastFloat(data["created_utc"]),
		Permalink:   client.CastString(data["permalink"]),
		URL:         client.CastString(data["url"]),
		Selftext:    client.CastString(data["selftext"]),
		IsSelf:      client.CastBool(data["is_self"]),
		Over18:      client.CastBool(data["over_18"]),
		IsVideo:     client.CastBool(data["is_video"]),
		Stickied:    client.CastBool(data["stickied"]),
	}
}

// ParseListing parses a Reddit listing response into a ListingPage.
func ParseListing(raw any) *models.ListingPage {
	m := client.CastMap(raw)
	data := client.CastMap(m["data"])
	children := client.CastSlice(data["children"])

	page := &models.ListingPage{
		After:  client.CastString(data["after"]),
		Before: client.CastString(data["before"]),
	}

	for _, child := range children {
		cm := client.CastMap(child)
		if client.CastString(cm["kind"]) == "t3" {
			post := ParsePost(client.CastMap(cm["data"]))
			page.Items = append(page.Items, post)
		}
	}
	return page
}

// ParseCommentNode recursively parses a comment node.
func ParseCommentNode(node map[string]any) *models.Comment {
	kind := client.CastString(node["kind"])
	data := client.CastMap(node["data"])

	if kind == "more" {
		children := client.CastSlice(data["children"])
		ids := make([]string, 0, len(children))
		for _, c := range children {
			if s := client.CastString(c); s != "" {
				ids = append(ids, s)
			}
		}
		return &models.Comment{
			MoreCount:    client.CastInt(data["count"]),
			MoreChildren: ids,
		}
	}

	if kind != "t1" {
		return nil
	}

	comment := &models.Comment{
		ID:             client.CastString(data["id"]),
		Fullname:       client.CastString(data["name"]),
		Author:         client.CastString(data["author"]),
		Body:           client.CastString(data["body"]),
		ParentFullname: client.CastString(data["parent_id"]),
		Score:          client.CastInt(data["score"]),
		CreatedUTC:     client.CastFloat(data["created_utc"]),
	}

	// Parse nested replies
	repliesRaw := data["replies"]
	if repliesMap := client.CastMap(repliesRaw); repliesMap != nil {
		repliesData := client.CastMap(repliesMap["data"])
		replyChildren := client.CastSlice(repliesData["children"])
		for _, rc := range replyChildren {
			rcm := client.CastMap(rc)
			if child := ParseCommentNode(rcm); child != nil {
				if child.MoreCount > 0 {
					comment.MoreCount += child.MoreCount
					comment.MoreChildren = append(comment.MoreChildren, child.MoreChildren...)
				} else {
					comment.Replies = append(comment.Replies, child)
				}
			}
		}
	}

	return comment
}

// ParsePostDetail parses a Reddit [post_listing, comments_listing] response.
func ParsePostDetail(raw any) *models.PostDetail {
	arr := client.CastSlice(raw)
	if len(arr) < 2 {
		return &models.PostDetail{}
	}

	// First element: post listing
	postListing := ParseListing(arr[0])
	var post *models.Post
	if len(postListing.Items) > 0 {
		post = postListing.Items[0]
	}

	// Second element: comments listing
	commentsRaw := client.CastMap(arr[1])
	commentsData := client.CastMap(commentsRaw["data"])
	commentChildren := client.CastSlice(commentsData["children"])

	detail := &models.PostDetail{Post: post}
	for _, child := range commentChildren {
		cm := client.CastMap(child)
		c := ParseCommentNode(cm)
		if c == nil {
			continue
		}
		if c.MoreCount > 0 {
			detail.MoreCount += c.MoreCount
			detail.MoreChildren = append(detail.MoreChildren, c.MoreChildren...)
		} else {
			detail.Comments = append(detail.Comments, c)
		}
	}
	return detail
}

// ParseUserProfile parses a Reddit user about response.
func ParseUserProfile(raw any) *models.UserProfile {
	m := client.CastMap(raw)
	data := client.CastMap(m["data"])
	if data == nil {
		data = m
	}
	return &models.UserProfile{
		Name:         client.CastString(data["name"]),
		LinkKarma:    client.CastInt(data["link_karma"]),
		CommentKarma: client.CastInt(data["comment_karma"]),
		CreatedUTC:   client.CastFloat(data["created_utc"]),
		IsGold:       client.CastBool(data["is_gold"]),
		IsMod:        client.CastBool(data["is_mod"]),
	}
}

// ParseSubredditInfo parses a subreddit about response.
func ParseSubredditInfo(raw any) *models.SubredditInfo {
	m := client.CastMap(raw)
	data := client.CastMap(m["data"])
	if data == nil {
		data = m
	}
	return &models.SubredditInfo{
		DisplayName:         client.CastString(data["display_name"]),
		DisplayNamePrefixed: client.CastString(data["display_name_prefixed"]),
		PublicDescription:   client.CastString(data["public_description"]),
		Description:         client.CastString(data["description"]),
		Subscribers:         client.CastInt(data["subscribers"]),
		AccountsActive:      client.CastInt(data["accounts_active"]),
		CreatedUTC:          client.CastFloat(data["created_utc"]),
		Over18:              client.CastBool(data["over18"]),
	}
}
