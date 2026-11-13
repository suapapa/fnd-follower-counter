package sns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	defaultYouTubeBaseURL = "https://www.googleapis.com/youtube/v3"
	youtubeChannelIDLen   = 24
)

type youtubeFetcher struct {
	apiKey  string
	baseURL string
	client  jsonClient
}

func newYouTubeFetcherFromEnv() (*youtubeFetcher, error) {
	apiKey := strings.TrimSpace(os.Getenv("YOUTUBE_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("YOUTUBE_API_KEY is required for youtube")
	}
	return newYouTubeFetcher(apiKey, defaultYouTubeBaseURL, nil), nil
}

func newYouTubeFetcher(apiKey, baseURL string, client *http.Client) *youtubeFetcher {
	if baseURL == "" {
		baseURL = defaultYouTubeBaseURL
	}
	return &youtubeFetcher{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  newJSONClient(client),
	}
}

// FetchFollowerCount returns the subscriber count for a YouTube handle (e.g. "@vibetinker") or channel ID.
func (f *youtubeFetcher) FetchFollowerCount(ctx context.Context, handle string) (int, error) {
	target := strings.TrimSpace(handle)
	if target == "" {
		return 0, fmt.Errorf("youtube handle is empty")
	}

	paramKey := "forHandle"
	if isChannelID(target) {
		paramKey = "id"
	}

	params := url.Values{}
	params.Set("part", "statistics")
	params.Set("key", f.apiKey)
	params.Set(paramKey, target)

	endpoint := f.baseURL + "/channels?" + params.Encode()

	var payload struct {
		Items []struct {
			Statistics struct {
				SubscriberCount       string `json:"subscriberCount"`
				HiddenSubscriberCount bool   `json:"hiddenSubscriberCount"`
			} `json:"statistics"`
		} `json:"items"`
	}
	if err := f.client.get(ctx, getRequest{url: endpoint}, &payload); err != nil {
		return 0, fmt.Errorf("youtube api: %w", err)
	}
	if len(payload.Items) == 0 {
		return 0, fmt.Errorf("youtube channel not found for %q", target)
	}

	stats := payload.Items[0].Statistics
	if stats.HiddenSubscriberCount {
		return 0, fmt.Errorf("youtube subscriber count is hidden for %q", target)
	}
	cnt, err := strconv.Atoi(stats.SubscriberCount)
	if err != nil {
		return 0, fmt.Errorf("parse youtube subscriber count %q: %w", stats.SubscriberCount, err)
	}
	return cnt, nil
}

func isChannelID(target string) bool {
	return strings.HasPrefix(target, "UC") && len(target) == youtubeChannelIDLen
}
