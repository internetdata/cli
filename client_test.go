package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	internetdata "github.com/internetdata/sdk-go/v2"
)

// Both refusals are a 403, and each needs a different thing done next: a key
// refused for its address is a valid key, which a license hint sends looking at
// the wrong setting.
func TestExplainDatabase(t *testing.T) {
	cases := []struct {
		kind   internetdata.ErrorKind
		status int
		code   string
		want   string
	}{
		{internetdata.KindForbidden, 403, "NOT_LICENSED", "license does not cover that"},
		{internetdata.KindForbidden, 403, "LICENSE_EXPIRED", "license does not cover that"},
		{internetdata.KindForbidden, 403, "IP_NOT_ALLOWED", "its allowed addresses, or its organization's"},
		{internetdata.KindUnauthorized, 401, "UNAUTHORIZED", "was not accepted"},
	}
	for _, c := range cases {
		err := explainDatabase(&internetdata.Error{Kind: c.kind, Message: c.code, StatusCode: c.status})
		if got := err.Error(); !strings.Contains(got, c.want) {
			t.Errorf("%s: got %q, want it to contain %q", c.code, got, c.want)
		}
		if c.code == "IP_NOT_ALLOWED" && strings.Contains(err.Error(), "license") {
			t.Errorf("%s names the license: %q", c.code, err.Error())
		}
	}
}

// A lookup's 403 refuses a valid key for how it was used, and the spec's three
// reasons are the API's messages here. None is a plan or a license, so explain
// adds nothing to them.
func TestExplainLeavesALookup403ToTheAPI(t *testing.T) {
	for _, msg := range []string{
		"source address not allowed for this API key",
		"API key lacks the ip.lookup scope; add Lookups to it in the console",
		"API key sent over plain HTTP; use https://",
	} {
		err := &internetdata.Error{Kind: internetdata.KindForbidden, Message: msg, StatusCode: 403}
		if got := explain(err).Error(); got != err.Error() {
			t.Errorf("explain(%q) = %q, want the SDK's error alone", msg, got)
		}
	}
}

// Both of the CLI's SDK clients, the keyed one and the keyless sign-in one, name
// the CLI first and the SDK after it, on a lookup, the API's other calls and the
// file transfer alike, so the API can tell this CLI's traffic, and its version,
// from any other program built on the SDK.
func TestEveryRequestNamesTheCLIThenTheSDK(t *testing.T) {
	var mu sync.Mutex
	agents := map[string]string{}
	var api *httptest.Server
	api = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		agents[r.URL.Path] = r.Header.Get("User-Agent")
		mu.Unlock()
		switch r.URL.Path {
		case "/api/v2/database/download":
			http.Redirect(w, r, api.URL+"/files/vpn_ip_v1.csvgz", http.StatusFound)
		case "/files/vpn_ip_v1.csvgz":
			_, _ = io.WriteString(w, "range\n")
		case "/9.1.0.1":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, answerJSON("9.1.0.1"))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":"slow_down"}`)
		}
	}))
	t.Cleanup(api.Close)

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("INTERNETDATA_API_KEY", "k-test")
	t.Setenv("INTERNETDATA_SESSION", "")
	savedConfig, savedBaseURL, savedRetries := gConfig, fBaseURL, fRetries
	t.Cleanup(func() { gConfig, fBaseURL, fRetries = savedConfig, savedBaseURL, savedRetries })
	gConfig, fBaseURL, fRetries = NewConfig(), api.URL, 0
	gConfig.CacheEnabled = false

	c, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_, err = c.Database().Download(t.Context(), "vpn_ip_v1", internetdata.FormatCSVGZ, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.api.Lookup(t.Context(), "9.1.0.1"); err != nil {
		t.Fatal(err)
	}
	if _, err := startDeviceAuth(t.Context()); err == nil {
		t.Fatal("the sign-in start succeeded against an origin that refuses it")
	}

	want := "internetdata-cli/" + version
	if len(agents) != 4 {
		t.Fatalf("saw requests to %v, want the download, its file, a lookup and a sign-in start", agents)
	}
	for path, agent := range agents {
		products := strings.Fields(agent)
		if len(products) != 2 || products[0] != want || !strings.HasPrefix(products[1], "internetdata-go/") {
			t.Errorf("%s was sent User-Agent %q, want %s internetdata-go/<the SDK's version>", path, agent, want)
		}
	}
}
