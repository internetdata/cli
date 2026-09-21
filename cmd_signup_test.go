package main

import "testing"

// The console URL is derived from the session's API host rather than listed,
// so a new deployment needs no code change here. A host we do not recognise
// must fall back to production rather than to a guess: sending someone to a
// made-up console is worse than sending them to the real one.
func TestSignupURL(t *testing.T) {
	t.Cleanup(func() { gConfig = NewConfig(); fBaseURL = "" })

	for _, tt := range []struct {
		base, want string
	}{
		{"", "https://app.internetdata.io/auth/signup"},
		// The API is served at the apex, so there is no `api` label to swap.
		{"https://internetdata.io", "https://app.internetdata.io/auth/signup"},
		{"https://internetdata.io/", "https://app.internetdata.io/auth/signup"},
		{"https://internetdata.io:8443", "https://app.internetdata.io/auth/signup"},
		{"https://eu.internetdata.io", "https://app-eu.internetdata.io/auth/signup"},
		// An api-labelled host keeps the sibling shape.
		{"https://api.internetdata.io", "https://app.internetdata.io/auth/signup"},
		{"https://api-eu.internetdata.io", "https://app-eu.internetdata.io/auth/signup"},
		// Nothing else is ours to rewrite.
		{"https://a.b.internetdata.io", "https://app.internetdata.io/auth/signup"},
		{"https://notinternetdata.io", "https://app.internetdata.io/auth/signup"},
		{"https://api.example.com", "https://app.internetdata.io/auth/signup"},
		{"https://example.com", "https://app.internetdata.io/auth/signup"},
		{"://broken", "https://app.internetdata.io/auth/signup"},
	} {
		gConfig = NewConfig()
		fBaseURL = tt.base
		if got := signupURL(); got != tt.want {
			t.Errorf("base %q: got %q, want %q", tt.base, got, tt.want)
		}
	}
}
