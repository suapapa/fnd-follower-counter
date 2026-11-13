package sns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
)

const defaultTwitchBaseURL = "https://api.twitch.tv/helix"

type twitchFetcher struct {
	clientID string
	baseURL  string
	client   jsonClient
}

func newTwitchFetcherFromEnv() (*twitchFetcher, error) {
	clientID := strings.TrimSpace(os.Getenv("TWITCH_CLIENT_ID"))
	if clientID == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID is required for twitch")
	}
	clientSecret := strings.TrimSpace(os.Getenv("TWITCH_CLIENT_SECRET"))
	if clientSecret == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_SECRET is required for twitch")
	}

	oauthCfg := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     twitch.Endpoint.TokenURL,
	}
	oauthCtx := context.WithValue(context.Background(), oauth2.HTTPClient, newJSONClient(nil).http)
	return newTwitchFetcher(clientID, defaultTwitchBaseURL, oauthCfg.Client(oauthCtx)), nil
}

func newTwitchFetcher(clientID, baseURL string, client *http.Client) *twitchFetcher {
	if baseURL == "" {
		baseURL = defaultTwitchBaseURL
	}
	return &twitchFetcher{
		clientID: clientID,
		baseURL:  strings.TrimRight(baseURL, "/"),
		client:   newJSONClient(client),
	}
}

// FetchFollowerCount returns the total follower count for a Twitch login name.
func (f *twitchFetcher) FetchFollowerCount(ctx context.Context, handle string) (int, error) {
	login := strings.TrimPrefix(strings.TrimSpace(handle), "@")
	if login == "" {
		return 0, fmt.Errorf("twitch login is empty")
	}

	headers := map[string]string{
		"Client-Id": f.clientID,
	}

	userParams := url.Values{}
	userParams.Set("login", login)
	var users struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	userReq := getRequest{
		url:     f.baseURL + "/users?" + userParams.Encode(),
		headers: headers,
	}
	if err := f.client.get(ctx, userReq, &users); err != nil {
		return 0, fmt.Errorf("twitch users api: %w", err)
	}
	if len(users.Data) == 0 {
		return 0, fmt.Errorf("twitch user %q not found", login)
	}

	followParams := url.Values{}
	followParams.Set("broadcaster_id", users.Data[0].ID)
	var followers struct {
		Total int `json:"total"`
	}
	followReq := getRequest{
		url:     f.baseURL + "/channels/followers?" + followParams.Encode(),
		headers: headers,
	}
	if err := f.client.get(ctx, followReq, &followers); err != nil {
		return 0, fmt.Errorf("twitch followers api: %w", err)
	}
	return followers.Total, nil
}
