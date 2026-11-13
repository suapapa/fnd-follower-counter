package sns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInstagramFetcher(t *testing.T) {
	tests := []struct {
		name       string
		handle     string
		wantFields string
		status     int
		body       string
		want       int
		wantErr    bool
	}{
		{
			name:       "success",
			handle:     "homin.crc",
			wantFields: "business_discovery.username(homin.crc){followers_count}",
			status:     http.StatusOK,
			body:       `{"business_discovery":{"followers_count":842,"id":"17841400000"},"id":"99999"}`,
			want:       842,
		},
		{
			name:       "strips leading at",
			handle:     "@homin.crc",
			wantFields: "business_discovery.username(homin.crc){followers_count}",
			status:     http.StatusOK,
			body:       `{"business_discovery":{"followers_count":12500}}`,
			want:       12500,
		},
		{
			name:       "graph api error payload",
			handle:     "homin.crc",
			wantFields: "business_discovery.username(homin.crc){followers_count}",
			status:     http.StatusOK,
			body:       `{"error":{"message":"Invalid user","type":"OAuthException","code":190}}`,
			wantErr:    true,
		},
		{
			name:       "missing business_discovery",
			handle:     "homin.crc",
			wantFields: "business_discovery.username(homin.crc){followers_count}",
			status:     http.StatusOK,
			body:       `{"id":"99999"}`,
			wantErr:    true,
		},
		{
			name:       "http error",
			handle:     "homin.crc",
			wantFields: "business_discovery.username(homin.crc){followers_count}",
			status:     http.StatusBadRequest,
			body:       `{"error":{"message":"bad request"}}`,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/17841405309211844" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				q := r.URL.Query()
				if q.Get("access_token") != "ig-token" {
					t.Fatalf("unexpected access_token: %s", q.Get("access_token"))
				}
				if got := q.Get("fields"); got != tc.wantFields {
					t.Fatalf("fields = %q, want %q", got, tc.wantFields)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			f := newInstagramFetcher(
				"ig-token",
				"17841405309211844",
				srv.URL,
				srv.Client(),
			)
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
