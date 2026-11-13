package sns

import "testing"

func TestNewFetcherMissingEnv(t *testing.T) {
	t.Setenv("YOUTUBE_API_KEY", "")
	if _, err := NewFetcher("youtube"); err == nil {
		t.Fatalf("expected error when YOUTUBE_API_KEY is empty")
	}

	t.Setenv("INSTAGRAM_ACCESS_TOKEN", "")
	t.Setenv("INSTAGRAM_IG_USER_ID", "")
	if _, err := NewFetcher("insta"); err == nil {
		t.Fatalf("expected error when instagram env vars are empty")
	}

	t.Setenv("TWITCH_CLIENT_ID", "")
	t.Setenv("TWITCH_CLIENT_SECRET", "")
	if _, err := NewFetcher("twitch"); err == nil {
		t.Fatalf("expected error when twitch env vars are empty")
	}

	if _, err := NewFetcher("unknown"); err == nil {
		t.Fatalf("expected error for unsupported provider")
	}
}
