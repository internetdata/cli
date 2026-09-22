package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/browser"
	"github.com/spf13/pflag"
)

// signupURL is where an account is created.
//
// Derived from the session's API host rather than fixed, so a session pointed
// at another deployment sends you to that deployment's console instead of to
// production - the one way this could quietly cost someone a real account they
// did not want. Anything we do not recognize falls back to production.
//
// The API is served at the APEX, so there is usually no `api` label to swap
// for `app`: the apex's console is `app.` on it, and a deployment leading the
// host moves behind `app-`. The same rule as the authorization server's own
// hand-off to the console.
func signupURL() string {
	const prod = "https://app.internetdata.io/auth/signup"
	base := gConfig.ResolveBaseURL()
	if base == "" {
		return prod
	}
	u, err := url.Parse(base)
	if err != nil {
		return prod
	}
	host := strings.ToLower(u.Hostname())
	if host != "internetdata.io" && !strings.HasSuffix(host, ".internetdata.io") {
		return prod
	}
	rest, isAPI := strings.CutPrefix(host, "api")
	if isAPI && (strings.HasPrefix(rest, ".") || strings.HasPrefix(rest, "-")) {
		return "https://app" + rest + "/auth/signup"
	}
	labels := strings.Split(host, ".")
	switch len(labels) {
	case 2:
		return "https://app." + host + "/auth/signup"
	case 3:
		return "https://app-" + labels[0] + "." + labels[1] + "." + labels[2] + "/auth/signup"
	}
	return prod
}

func printHelpSignup() {
	fmt.Printf(
		`Usage: %[1]s signup [<opts>]

Description:
  Create an account in the browser, then sign this machine in.

  The console opens so you can sign up. Come back and you are signed in the same
  way '%[1]s login' does it: confirm a short code in the browser, pick which of
  your API keys this machine should hold, and nothing is typed or pasted here.
  If no browser opens, both URLs are printed.

Options:
  --no-browser
    print the URL instead of opening it.
  --help, -h
    show help.
`, progBase)
}

func cmdSignup() error {
	var fNoBrowser bool
	globalFlags()
	pflag.BoolVar(&fNoBrowser, "no-browser", false, "print the URL instead of opening it.")
	parseSubFlags()

	if fHelp {
		printHelpSignup()
		return nil
	}

	url := signupURL()
	opened := false
	if !fNoBrowser {
		// Errors are ignored on purpose: a headless machine has no browser, and
		// the printed URL below is the answer in both cases.
		opened = browser.OpenURL(url) == nil
	}
	if opened {
		fmt.Fprintln(os.Stderr, "opened your browser to create an account.")
	}
	fmt.Fprintf(os.Stderr, "\n  %s\n\n", url)
	fmt.Fprintln(os.Stderr, "once you have signed up, come back here and we will finish signing you in.")

	// Falls through to the SAME device flow login uses, so a brand-new user
	// finishes in the browser they already have open rather than being asked to
	// find a key and paste it back. One path writes a credential.
	return browserLogin(fSession, fNoBrowser)
}
