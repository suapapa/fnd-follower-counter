package sns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestYouTubeFetcher(t *testing.T) {
	tests := []struct {
		name      string
		handle    string
		wantParam string
		wantVal   string
		status    int
		body      string
		want      int
		wantErr   bool
	}{
		{
			name:      "by handle",
			handle:    "@vibetinker",
			wantParam: "forHandle",
			wantVal:   "@vibetinker",
			status:    http.StatusOK,
			body:      `{"items":[{"statistics":{"subscriberCount":"1230","hiddenSubscriberCount":false}}]}`,
			want:      1230,
		},
		{
			name:      "by channel id",
			handle:    "UC1234567890123456789012",
			wantParam: "id",
			wantVal:   "UC1234567890123456789012",
			status:    http.StatusOK,
			body:      `{"items":[{"statistics":{"subscriberCount":"45600","hiddenSubscriberCount":false}}]}`,
			want:      45600,
		},
		{
			name:      "hidden subscribers",
			handle:    "@hidden",
			wantParam: "forHandle",
			wantVal:   "@hidden",
			status:    http.StatusOK,
			body:      `{"items":[{"statistics":{"subscriberCount":"0","hiddenSubscriberCount":true}}]}`,
			wantErr:   true,
		},
		{
			name:      "channel not found",
			handle:    "@missing",
			wantParam: "forHandle",
			wantVal:   "@missing",
			status:    http.StatusOK,
			body:      `{"items":[]}`,
			wantErr:   true,
		},
		{
			name:      "http error",
			handle:    "@vibetinker",
			wantParam: "forHandle",
			wantVal:   "@vibetinker",
			status:    http.StatusForbidden,
			body:      `{"error":{"message":"quotaExceeded"}}`,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/channels" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				q := r.URL.Query()
				if q.Get("part") != "statistics" {
					t.Fatalf("unexpected part: %s", q.Get("part"))
				}
				if q.Get("key") != "test-key" {
					t.Fatalf("unexpected key: %s", q.Get("key"))
				}
				if got := q.Get(tc.wantParam); got != tc.wantVal {
					t.Fatalf("query %s = %q, want %q", tc.wantParam, got, tc.wantVal)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			f := newYouTubeFetcher("test-key", srv.URL, srv.Client())
			got, err := f.FetchFollowerCount(context.Background(), tc.handle)
			if (err != nil) != tc.wantErr {
				t.Fatalf("FetchFollowerCount() err = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Fatalf("FetchFollowerCount() = %d, want %d", got, tc.want)
			}
		})
	}
}
