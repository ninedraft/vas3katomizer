package feedgen

import (
	"fmt"
	"strings"

	"github.com/gorilla/feeds"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninedraft/vas3katomizer/internal/models"
	"github.com/russross/blackfriday/v2"
)

func Generate(input *models.FeedPage) *feeds.Feed {
	feed := &feeds.Feed{
		Title: input.Title,
		Link:  href(input.HomePageURL),
	}

	items := make([]*feeds.Item, 0, len(input.Items))

	for _, item := range input.Items {

		items = append(items, &feeds.Item{
			Title:       item.Title,
			Id:          item.ID,
			Updated:     item.DateModified,
			Created:     item.DatePublished,
			Link:        href(item.URL),
			Description: generateDesc(item),
			Content:     generateContent(item),
			Author: &feeds.Author{
				Name: composeAuthors(item.Authors),
			},
		})
	}

	feed.Items = items

	return feed
}

func href(url string) *feeds.Link {
	return &feeds.Link{Href: url}
}

func composeAuthors(authors []models.ItemAuthor) string {
	names := make([]string, 0, len(authors))
	for _, author := range authors {
		names = append(names, author.Name)
	}

	return strings.Join(names, ", ")
}

var rendererDesc = blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
	Flags: blackfriday.CommonHTMLFlags | blackfriday.NofollowLinks,
})

func generateDesc(item models.FeedItem) string {
	text, nUpvotes, nComments := item.ContentText, item.Club.Upvotes, item.Club.CommentCount
	sep := " "
	head := generateHead(text)
	if head != "🔒" {
		sep = "\n\n"
	}

	md := fmt.Sprintf("💚%d 💬%d%s%s", nUpvotes, nComments, sep, head)

	rendered := blackfriday.Run([]byte(md), blackfriday.WithRenderer(rendererDesc))
	sanitized := string(bluemonday.UGCPolicy().SanitizeBytes(rendered))

	return strings.TrimSpace(sanitized)
}

func generateHead(content string) string {
	head := &strings.Builder{}

	for budget := 5; budget > 0; {
		line, rest, ok := strings.Cut(content, "\n")
		if !ok {
			head.WriteString(rest)
		}
		content = rest

		if strings.Contains(line, "![") {
			continue
		}

		head.WriteString(line)
		head.WriteRune('\n')
		budget--
	}

	return strings.TrimSpace(head.String())
}

var rendererContent = blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
	Flags: blackfriday.CommonHTMLFlags | blackfriday.CompletePage | blackfriday.TOC,
})

func generateContent(item models.FeedItem) string {
	content := []byte(item.ContentText)
	rendered := blackfriday.Run(content, blackfriday.WithRenderer(rendererContent))
	sanitized := string(bluemonday.UGCPolicy().SanitizeBytes(rendered))

	return strings.TrimSpace(sanitized)
}
