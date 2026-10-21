package main

import (
	"strings"
	"testing"

	internetdata "github.com/internetdata/sdk-go/v2"
)

// Both refusals are a 403, and each needs a different thing done next: a key
// refused for its address is a valid key, which a license hint sends looking at
// the wrong setting.
func TestExplain(t *testing.T) {
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
		err := explain(&internetdata.Error{Kind: c.kind, Message: c.code, StatusCode: c.status})
		if got := err.Error(); !strings.Contains(got, c.want) {
			t.Errorf("%s: got %q, want it to contain %q", c.code, got, c.want)
		}
		if c.code == "IP_NOT_ALLOWED" && strings.Contains(err.Error(), "license") {
			t.Errorf("%s names the license: %q", c.code, err.Error())
		}
	}
}
