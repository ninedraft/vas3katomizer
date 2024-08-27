package vas3klient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ninedraft/vas3katomizer/internal/models"
)

type AuthToken string

func New(endpoint string, token AuthToken) *Client {
	return &Client{
		endpoint:  endpoint,
		authtoken: token,
		c: &http.Client{
			Transport: newTransport(),
		},
	}
}

type Client struct {
	endpoint  string
	authtoken AuthToken
	c         *http.Client
}

var ErrUnexpectedStatus = errors.New("unexpected http status")

func errUnexpectedStatus(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("%w %d %q: %q", ErrUnexpectedStatus, resp.StatusCode, resp.Status, body)
}

func (client *Client) FetchArticle(ctx context.Context, ref string) (*models.Article, error) {
	endpoint, err := url.JoinPath(client.endpoint, ref)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	metaReq, err := client.getReq(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("preparing a http request: %w", err)
	}

	var article models.Article
	if err := client.doJSONReq(metaReq, &article); err != nil {
		return nil, fmt.Errorf("http: %v", err)
	}

	if article.Post.ContentText == "🔒" {
		target := strings.TrimSuffix(article.Post.URL, "/") + ".md"
		req, err := client.getReq(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("preparing article body request: %w", err)
		}

		body := &strings.Builder{}
		err = client.doReq(req, body)
		if err != nil {
			return nil, fmt.Errorf("fetching article body %s: %w", target, err)
		}
		article.Post.ContentText = body.String()
	}

	return &article, nil
}

func (client *Client) FetchFeed(ctx context.Context, page int) (*models.FeedPage, error) {
	endpoint, err := url.Parse(client.endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	endpoint.Path = "/all/new/feed.json"

	if page > 0 {
		q := endpoint.Query()
		q.Set("page", strconv.Itoa(page))
		endpoint.RawQuery = q.Encode()
	}

	req, err := client.getReq(ctx, endpoint.String())
	if err != nil {
		return nil, fmt.Errorf("preparing a http request: %w", err)
	}

	var feed models.FeedPage
	if err := client.doJSONReq(req, &feed); err != nil {
		return nil, fmt.Errorf("http: %v", err)
	}

	return &feed, nil
}

func (client *Client) getReq(ctx context.Context, target string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Service-Token", string(client.authtoken))

	return req, nil
}

func (client *Client) doJSONReq(req *http.Request, dst any) error {
	data := &bytes.Buffer{}
	err := client.doReq(req, data)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if err := json.Unmarshal(data.Bytes(), dst); err != nil {
		return fmt.Errorf("decoding json response: %w", err)
	}

	return nil
}

const maxBodySize = 1 << 20

func (client *Client) doReq(req *http.Request, dst io.Writer) error {
	resp, err := client.c.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("performing http request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return errUnexpectedStatus(resp)
	}

	_, err = io.Copy(dst, io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return fmt.Errorf("body: %w", err)
	}

	return nil
}

// "optimized" for single host usage
func newTransport() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	const maxIdleConns, maxConns = 1, 1

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConns,
		MaxConnsPerHost:       maxConns,
		IdleConnTimeout:       180 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
}
