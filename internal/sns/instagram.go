package sns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultInstagramBaseURL = "https://graph.facebook.com/v21.0"

type instagramFetcher struct {
	accessToken string
	igUserID    string
	baseURL     string
	client      jsonClient
}

func newInstagramFetcherFromEnv() (*instagramFetcher, error) {
	token := strings.TrimSpace(os.Getenv("INSTAGRAM_ACCESS_TOKEN"))
	if token == "" {
		return nil, fmt.Errorf("INSTAGRAM_ACCESS_TOKEN is required for instagram")
	}
	igUserID := strings.TrimSpace(os.Getenv("INSTAGRAM_IG_USER_ID"))
	if igUserID == "" {
		return nil, fmt.Errorf("INSTAGRAM_IG_USER_ID is required for instagram")
	}
	return newInstagramFetcher(
		token,
		igUserID,
		defaultInstagramBaseURL,
		nil,
	), nil
}

func newInstagramFetcher(
	accessToken string,
	igUserID string,
	baseURL string,
	client *http.Client,
) *instagramFetcher {
	if baseURL == "" {
		baseURL = defaultInstagramBaseURL
	}
	return &instagramFetcher{
		accessToken: accessToken,
		igUserID:    igUserID,
		baseURL:     strings.TrimRight(baseURL, "/"),
		client:      newJSONClient(client),
	}
}

// FetchFollowerCount returns the follower count for an Instagram Business/Creator username.
func (f *instagramFetcher) FetchFollowerCount(ctx context.Context, handle string) (int, error) {
	username := strings.TrimPrefix(strings.TrimSpace(handle), "@")
	if username == "" {
		return 0, fmt.Errorf("instagram username is empty")
	}

	params := url.Values{}
	params.Set("fields", fmt.Sprintf("business_discovery.username(%s){followers_count}", username))
	params.Set("access_token", f.accessToken)

	endpoint := fmt.Sprintf("%s/%s?%s", f.baseURL, url.PathEscape(f.igUserID), params.Encode())

	var payload struct {
		BusinessDiscovery *struct {
			FollowersCount int `json:"followers_count"`
		} `json:"business_discovery"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if err := f.client.get(ctx, getRequest{url: endpoint}, &payload); err != nil {
		return 0, fmt.Errorf("instagram graph api: %w", err)
	}
	if payload.Error != nil {
		return 0, fmt.Errorf(
			"instagram graph api error (%d %s): %s",
			payload.Error.Code,
			payload.Error.Type,
			payload.Error.Message,
		)
	}
	if payload.BusinessDiscovery == nil {
		return 0, fmt.Errorf("instagram user %q not found in business_discovery", username)
	}
	return payload.BusinessDiscovery.FollowersCount, nil
}
