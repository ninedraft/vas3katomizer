package models

import "time"

type Article struct {
	Post Post `json:"post"`
}

type Post struct {
	ID            string       `json:"id"`
	URL           string       `json:"url"`
	Title         string       `json:"title"`
	ContentText   string       `json:"content_text"`
	DatePublished time.Time    `json:"date_published"`
	DateModified  time.Time    `json:"date_modified"`
	Authors       []ItemAuthor `json:"authors"`
	Club          ClubInfo
}
