package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	internetdata "github.com/internetdata/sdk-go/v2"
)

// The sign-in client is built on the API base URL and nothing else: the SDK
// derives the apex the authorization server is mounted on, so this program must
// not rewrite a host itself. An earlier shape derived an `auth.<brand>` hostname
// by string surgery, which works in production and silently points at the wrong
// deployment everywhere else - so these cases keep a derivation from creeping
// back.
func TestAuthBaseURL(t *testing.T) {
	tests := []struct {
		name string
		base string
		want string
	}{
		{"unset falls back to production", "", "https://api.internetdata.io"},
		// A hyphenated first label is the shape a non-production deployment
		// takes, and the old code rewrote exactly that away. Kept generic: this
		// repo is public.
		{"a session keeps its own deployment", "https://api-eu.example.com", "https://api-eu.example.com"},
		// The API is served at the apex here, so a deployment's host has no `api`
		// label at all, and is still used exactly as given.
		{"an apex-shaped deployment is not rewritten", "https://eu.example.com", "https://eu.example.com"},
		{"a trailing slash does not double up", "https://api.example.com/", "https://api.example.com"},
		{"an unrelated host is honored, not rewritten", "https://api.example.com", "https://api.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveFlag, saveCfg := fBaseURL, gConfig
			// ResolveBaseURL reads the flag first, which is the precedence a
			// `--base-url` on the command line relies on.
			fBaseURL = tt.base
			gConfig = Config{Sessions: map[string]*Session{}}
			defer func() { fBaseURL, gConfig = saveFlag, saveCfg }()

			if got := authBaseURL(); got != tt.want {
				t.Errorf("authBaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

// A session's calls go to its API host, all but the sign-in, which the SDK sends
// to that host's apex, where the authorization server is mounted.
func TestSignInGoesToTheApexAndTheRestToTheAPIHost(t *testing.T) {
	var mu sync.Mutex
	hosts := map[string]string{}
	savedTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = savedTransport })
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		hosts[r.URL.Path] = r.URL.Host
		mu.Unlock()
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"error":"slow_down"}`)),
			Request:    r,
		}, nil
	})

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("INTERNETDATA_API_KEY", "k-test")
	t.Setenv("INTERNETDATA_SESSION", "")
	savedConfig, savedBaseURL, savedRetries := gConfig, fBaseURL, fRetries
	t.Cleanup(func() { gConfig, fBaseURL, fRetries = savedConfig, savedBaseURL, savedRetries })
	gConfig, fBaseURL, fRetries = NewConfig(), "https://api-eu.example.test", 0

	if _, err := startDeviceAuth(t.Context()); err == nil {
		t.Fatal("the sign-in start succeeded against an origin that refuses it")
	}
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Database().List(t.Context()); err == nil {
		t.Fatal("the listing succeeded against an origin that refuses it")
	}

	if len(hosts) != 2 {
		t.Fatalf("saw requests to %v, want a sign-in start and a listing", hosts)
	}
	for path, host := range hosts {
		want := "api-eu.example.test"
		if strings.HasPrefix(path, "/oauth/") {
			want = "eu.example.test"
		}
		if host != want {
			t.Errorf("%s went to %s, want %s", path, host, want)
		}
	}
}

// roundTripFunc is a transport answering every request with a function.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// What a person reads when the server refuses a sign-in, and a revoke the
// server failed reported as failed rather than read as success.
func TestSignInRefusals(t *testing.T) {
	dev := &internetdata.DeviceAuthorization{DeviceCode: "mo_dc_x", ExpiresIn: 900, Interval: 1}
	start := func(ctx context.Context) error {
		_, err := startDeviceAuth(ctx)
		return err
	}
	poll := func(ctx context.Context) error {
		_, err := pollForToken(ctx, dev)
		return err
	}
	revoke := func(ctx context.Context) error {
		return revokeToken(ctx, "mo_rt_x")
	}
	tests := []struct {
		name   string
		status int
		answer string
		call   func(context.Context) error
		want   string
	}{
		{"slow_down refuses the start", 400, `{"error":"slow_down"}`, start,
			"too many sign-in attempts from this address; wait a minute and try again"},
		{"access_denied ends the poll", 400, `{"error":"access_denied"}`, poll,
			"the request was denied in the browser"},
		{"expired_token ends the poll", 400, `{"error":"expired_token"}`, poll,
			"the code expired before it was approved; run the command again"},
		{"a revoke answered 500 fails", 500, `{"rc":"ERROR"}`, revoke, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.answer)
			}))
			defer server.Close()
			saveFlag, saveRetries, saveCfg := fBaseURL, fRetries, gConfig
			fBaseURL, fRetries = server.URL, 0
			gConfig = Config{Sessions: map[string]*Session{}}
			defer func() { fBaseURL, fRetries, gConfig = saveFlag, saveRetries, saveCfg }()
			// A poll that mistook the answer for pending would otherwise never end.
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()

			err := tt.call(ctx)
			if err == nil {
				t.Fatal("succeeded, want an error")
			}
			if tt.want != "" && err.Error() != tt.want {
				t.Errorf("error was %q, want %q", err, tt.want)
			}
		})
	}
}
