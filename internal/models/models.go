package models

// Post represents a Reddit post.
type Post struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Subreddit   string  `json:"subreddit"`
	Author      string  `json:"author"`
	Score       int     `json:"score"`
	NumComments int     `json:"num_comments"`
	CreatedUTC  float64 `json:"created_utc"`
	Permalink   string  `json:"permalink"`
	URL         string  `json:"url"`
	Selftext    string  `json:"selftext"`
	IsSelf      bool    `json:"is_self"`
	Over18      bool    `json:"over_18"`
	IsVideo     bool    `json:"is_video"`
	Stickied    bool    `json:"stickied"`
}

// Comment represents a Reddit comment, potentially with nested replies.
type Comment struct {
	ID             string     `json:"id"`
	Fullname       string     `json:"fullname"`
	Author         string     `json:"author"`
	Body           string     `json:"body"`
	ParentFullname string     `json:"parent_fullname"`
	Score          int        `json:"score"`
	CreatedUTC     float64    `json:"created_utc"`
	Replies        []*Comment `json:"replies,omitempty"`
	MoreCount      int        `json:"more_count,omitempty"`
	MoreChildren   []string   `json:"more_children,omitempty"`
}

// ListingPage holds a page of posts and pagination cursors.
type ListingPage struct {
	Items  []*Post `json:"items"`
	After  string  `json:"after,omitempty"`
	Before string  `json:"before,omitempty"`
}

// PostDetail holds a post and its comment tree.
type PostDetail struct {
	Post         *Post      `json:"post"`
	Comments     []*Comment `json:"comments"`
	MoreCount    int        `json:"more_count,omitempty"`
	MoreChildren []string   `json:"more_children,omitempty"`
}

// UserProfile holds basic Reddit user info.
type UserProfile struct {
	Name         string  `json:"name"`
	LinkKarma    int     `json:"link_karma"`
	CommentKarma int     `json:"comment_karma"`
	CreatedUTC   float64 `json:"created_utc"`
	IsGold       bool    `json:"is_gold"`
	IsMod        bool    `json:"is_mod"`
}

// SubredditInfo holds subreddit metadata.
type SubredditInfo struct {
	DisplayName           string  `json:"display_name"`
	DisplayNamePrefixed   string  `json:"display_name_prefixed"`
	PublicDescription     string  `json:"public_description"`
	Description           string  `json:"description"`
	Subscribers           int     `json:"subscribers"`
	AccountsActive        int     `json:"accounts_active"`
	CreatedUTC            float64 `json:"created_utc"`
	Over18                bool    `json:"over18"`
}
