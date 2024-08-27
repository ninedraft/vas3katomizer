package models

import (
	"path"
	"time"
)

type FeedPage struct {
	Version     string     `json:"version"`
	Title       string     `json:"title"`
	HomePageURL string     `json:"home_page_url"`
	FeedURL     string     `json:"feed_url"`
	NextURL     string     `json:"next_url"`
	Items       []FeedItem `json:"items"`
}

type FeedItem struct {
	ID            string       `json:"id"`
	URL           string       `json:"url"`
	Title         string       `json:"title"`
	ContentText   string       `json:"content_text"`
	DatePublished time.Time    `json:"date_published"`
	DateModified  time.Time    `json:"date_modified"`
	Authors       []ItemAuthor `json:"authors"`
	Club          ClubInfo     `json:"_club"`
}

func (item *FeedItem) Ref() string {
	return path.Join(item.Club.Type, item.Club.Slug) + ".json"
}

type ItemAuthor struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Avatar string `json:"avatar"`
}

type ClubInfo struct {
	Type          string `json:"type"`
	Slug          string `json:"slug"`
	CommentCount  int    `json:"comment_count"`
	ViewCount     int    `json:"view_count"`
	Upvotes       int    `json:"upvotes"`
	IsPublic      bool   `json:"is_public"`
	IsCommentable bool   `json:"is_commentable"`
}
