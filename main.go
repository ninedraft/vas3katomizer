package main

import (
	"cmp"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gorilla/feeds"
	"github.com/ninedraft/vas3katomizer/internal/feedgen"
	"github.com/ninedraft/vas3katomizer/internal/models"
	"github.com/ninedraft/vas3katomizer/internal/processor"
	"github.com/ninedraft/vas3katomizer/internal/vas3klient"
)

var (
	formatContentType = map[string]string{
		"atom": "application/atom+xml",
		"json": "application/json",
		"rss":  "application/rss+xml",
	}

	encoder = map[string]func(feed *feeds.Feed, dst io.Writer) error{
		"atom": (*feeds.Feed).WriteAtom,
		"json": (*feeds.Feed).WriteJSON,
		"rss":  (*feeds.Feed).WriteRss,
	}

	validFormats     = slices.Collect(maps.Keys(formatContentType))
	errInvalidFormat = fmt.Errorf("invalid format. Available formats: %s", strings.Join(validFormats, ", "))

	//go:embed index.html
	indexPage string
)

func main() {
	initLogger()

	serveAddr := cmp.Or(os.Getenv("SERVE_AT"), "localhost:8390")

	token := os.Getenv("VAS3KCLUB_TOKEN")
	if token == "" {
		panic("VAS3KCLUB_TOKEN env is not set")
	}

	feedEndpoint := cmp.Or(os.Getenv("VAS3KCLUB_ENDPOINT"), "https://vas3k.club/")

	blockedTypes := parseArray(cmp.Or(os.Getenv("BLOCKED_TYPES"), "intro"))
	blockedAuthors := parseArray(cmp.Or(os.Getenv("BLOCKED_AUTHORS")))

	// --- end of config ---

	client := vas3klient.New(feedEndpoint, vas3klient.AuthToken(token))

	ctx := context.Background()

	proc := processor.New(processor.Config{
		Client:         client,
		BlockedTypes:   blockedTypes,
		BlockedAuthors: blockedAuthors,
	})

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeContent(rw, r, "index.html", time.Time{}, strings.NewReader(indexPage))
	})

	fetchFeed := func(page int) (*models.FeedPage, error) {
		feed, err := client.FetchFeed(ctx, page)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch feed: %w", err)
		}

		if err := proc.Process(ctx, feed); err != nil {
			return nil, fmt.Errorf("processing feed: %w", err)
		}

		return feed, nil
	}

	writeFeed := func(rw http.ResponseWriter, feed *models.FeedPage, format string) {
		rw.Header().Set("Content-Type", formatContentType[format])
		encode := encoder[format]
		encode(feedgen.Generate(feed), rw)
	}

	mux.HandleFunc("GET /feed/{format}", func(rw http.ResponseWriter, r *http.Request) {
		format := r.PathValue("format")
		if !assertFormat(rw, format) {
			return
		}

		feed, err := fetchFeed(1)
		if err != nil {
			slog.Error("fetching feed feed",
				"error", err)
			http.Error(rw, "unable to fetch feed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeFeed(rw, feed, format)
	})

	mux.HandleFunc("GET /page/{page}/{format}", func(rw http.ResponseWriter, r *http.Request) {
		format := r.PathValue("format")
		if !assertFormat(rw, format) {
			return
		}

		page, err := strconv.Atoi(r.PathValue("page"))
		if err != nil {
			http.Error(rw, "invalid page number: "+err.Error(), http.StatusBadRequest)
		}

		page = max(page, 1)
		feed, err := fetchFeed(page)
		if err != nil {
			slog.Error("fetching feed",
				"error", err)
			http.Error(rw, "unable to fetch feed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeFeed(rw, feed, format)
	})

	slog.Info("serving",
		"address", serveAddr)

	err := (&http.Server{
		Addr:              serveAddr,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           logMW(mux),
	}).ListenAndServe()
	if err != nil {
		panic("serving: " + err.Error())
	}
}

func logMW(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			slog.Info("request",
				"handler", r.Pattern,
				"path", r.URL.Path,
				"remote-addr", r.RemoteAddr,
				"duration", time.Since(start))
		}(time.Now())

		next.ServeHTTP(rw, r)
	}
}

func assertFormat(rw http.ResponseWriter, format string) bool {
	if _, ok := formatContentType[format]; !ok {
		http.Error(rw, errInvalidFormat.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func initLogger() {
	var level slog.Level
	levelEnv := cmp.Or(os.Getenv("LOG_LEVEL"), "DEBUG")

	if err := level.UnmarshalText([]byte(levelEnv)); err != nil {
		slog.Error("invalid LOG_LEVEL env",
			"error", err)
	}

	handler := slog.NewTextHandler(os.Stderr,
		&slog.HandlerOptions{
			Level: level,
		})

	slog.SetDefault(slog.New(handler))
}

func parseArray(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(`,;|/`, r)
	})
}
