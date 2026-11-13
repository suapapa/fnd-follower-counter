// Package sns provides follower and subscriber count fetchers for supported social platforms.
package sns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = 15 * time.Second
	maxResponseBytes   = 1 << 20
	maxErrorBodyLen    = 256
)

// Fetcher retrieves the current follower or subscriber count for a target handle.
type Fetcher interface {
	FetchFollowerCount(ctx context.Context, handle string) (int, error)
}

// NewFetcher constructs a Fetcher for the given SNS provider name using environment credentials.
func NewFetcher(provider string) (Fetcher, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "youtube":
		return newYouTubeFetcherFromEnv()
	case "insta", "instagram":
		return newInstagramFetcherFromEnv()
	case "twitch":
		return newTwitchFetcherFromEnv()
	default:
		return nil, fmt.Errorf("unsupported sns provider %q", provider)
	}
}

type getRequest struct {
	url     string
	headers map[string]string
}

type jsonClient struct {
	http *http.Client
}

func newJSONClient(client *http.Client) jsonClient {
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	return jsonClient{http: client}
}

func (c jsonClient) get(ctx context.Context, req getRequest, dst any) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	for k, v := range req.headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		msg := strings.TrimSpace(string(body))
		if len(msg) > maxErrorBodyLen {
			msg = msg[:maxErrorBodyLen] + "..."
		}
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, msg)
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
