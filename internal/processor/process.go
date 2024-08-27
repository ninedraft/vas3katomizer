package processor

import (
	"context"
	"slices"

	"github.com/ninedraft/vas3katomizer/internal/models"
	"github.com/ninedraft/vas3katomizer/internal/vas3klient"
)

type Config struct {
	Client *vas3klient.Client

	BlockedTypes   []string
	BlockedAuthors []string
}

type Processor struct {
	client          *vas3klient.Client
	isBlockedAuthor func(username string) bool
	isBlockedType   func(itemType string) bool
}

func New(cfg Config) *Processor {
	return &Processor{
		client:          cfg.Client,
		isBlockedAuthor: buildSet(cfg.BlockedAuthors),
		isBlockedType:   buildSet(cfg.BlockedTypes),
	}
}

func (processor *Processor) Process(ctx context.Context, feed *models.FeedPage) error {

	filter := func(item models.FeedItem) bool {
		return processor.isBlockedType(item.Club.Type) ||
			slices.ContainsFunc(item.AuthorsUsernames(), processor.isBlockedAuthor)
	}

	feed.Items = slices.DeleteFunc(feed.Items, filter)

	return nil
}

func buildSet[E comparable](items []E) func(E) bool {
	set := make(map[E]bool, len(items))
	for _, item := range items {
		set[item] = true
	}

	return func(item E) bool {
		return set[item]
	}
}
