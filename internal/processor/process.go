package processor

import (
	"context"
	"slices"
	"strings"

	"github.com/ninedraft/vas3katomizer/internal/models"
	"github.com/ninedraft/vas3katomizer/internal/vas3klient"
)

type Processor struct {
	Client *vas3klient.Client

	AllowedTypes []string
	BlockedTypes []string
}

func (processor *Processor) Process(ctx context.Context, feed *models.FeedPage) error {
	feed.Items = slices.DeleteFunc(feed.Items, func(item models.FeedItem) bool {
		return slices.Contains(processor.BlockedTypes, item.Club.Type)
	})

	return nil
}

func isLocked(content string) bool {
	return strings.EqualFold(strings.TrimSpace(content), "🔒")
}
