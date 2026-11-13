package sns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTwitchFetcher(t *testing.T) {
	tests := []struct {
		name         string
		handle       string
		usersStatus  int
		usersBody    string
		followStatus int
		followBody   string
		want         int
		wantErr      bool
	}{
		{
			name:         "success",
			handle:       "suapapa",
			usersStatus:  http.StatusOK,
			usersBody:    `{"data":[{"id":"12345","login":"suapapa"}]}`,
			followStatus: http.StatusOK,
			followBody:   `{"total":57,"data":[]}`,
			want:         57,
		},
		{
			name:        "user not found",
			handle:      "nobody",
			usersStatus: http.StatusOK,
			usersBody:   `{"data":[]}`,
			wantErr:     true,
		},
		{
			name:         "followers api error",
			handle:       "suapapa",
			usersStatus:  http.StatusOK,
			usersBody:    `{"data":[{"id":"12345","login":"suapapa"}]}`,
			followStatus: http.StatusUnauthorized,
			followBody:   `{"error":"Unauthorized"}`,
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Client-Id") != "tw-client" {
					t.Fatalf("missing Client-Id header")
				}
				switch r.URL.Path {
				case "/users":
					if r.URL.Query().Get("login") != tc.handle {
						t.Fatalf("unexpected login: %s", r.URL.Query().Get("login"))
					}
					w.WriteHeader(tc.usersStatus)
					_, _ = w.Write([]byte(tc.usersBody))
				case "/channels/followers":
					if r.URL.Query().Get("broadcaster_id") != "12345" {
						t.Fatalf("unexpected broadcaster_id: %s", r.URL.Query().Get("broadcaster_id"))
					}
					w.WriteHeader(tc.followStatus)
					_, _ = w.Write([]byte(tc.followBody))
				default:
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
			}))
			defer srv.Close()

			f := newTwitchFetcher("tw-client", srv.URL, srv.Client())
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
