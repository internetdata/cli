# internetdata/cli/internetdata

The official CLI, a presentation layer over `github.com/internetdata/sdk-go/v2`.
Public repo: `github.com/internetdata/cli`, with gitea as a second push URL.

A copy of `vpndetection/cli/vpndetection` minus everything the per-IP lookup
API brought: no lookup, `bulk` or `myip`, no result cache, no field selection or
output formats, no entitlement. The flat dispatch, the FOUR places a command
lives, sessions, the device-flow login and the whole release plumbing are that
CLI's, and so are their gotchas: read its `AGENTS.md` first, and change both
when you change a shared mechanic.

## What differs, and why

- **The API is at the apex (`internetdata.io`)**, so `signupURL()` has no `api`
  label to swap: the apex maps to `app.`, a three-label `E.internetdata.io` to
  `app-E.internetdata.io`, and an `api`-labelled host keeps the sibling shape -
  the rule oauth_api's `consoleOriginForHost` implements for the device page.
  Anything else falls back to production. `cmd_signup_test.go` pins it.
- **`whoami` has no entitlement to show.** It prints the credential first, so a
  refused key still says which key it was, then the families the organization
  holds a license for (standing not `unlicensed`) from `Database.List()`.
  `checkKey` validates through the same call, so a key without `db.download` is
  refused at login, and it arrives as a 401, not a 403: keys are default-deny.
- **A bare invocation prints help, and an unknown word is named as a mistake.**
  There is no default command to fall into.
- **No color.** Nothing prints any once the lookup output is gone, so neither
  `--nocolor` nor `fatih/color` came across.

## Channel state (2026-09-21, v1.0.0)

GitHub release, ghcr (public, anonymous pull), the tap (`brew trust` then
`brew install`, Homebrew 7.0.4 on Linux), apt (`apt.internetdata.io`) and
`go install` were each verified installing 1.0.0. Chocolatey 1.0.0 sits in
moderation, re-pushed from `main` on 2026-09-22 with the `<copyright>`
vpndetection's 1.1.0 was sent back for, before any human had reviewed it; a
later version's push is refused until 1.0.0 is approved, then goes out with
`channel=chocolatey` (`docs/cli/channel-windows.md`). winget's first
manifest is microsoft/winget-pkgs#438455. The README names neither: add each
once it installs, per `docs/cli/releasing.md`.
Until that PR merges every release's `winget` job is red by design: the org's
`WINGET_TOKEN` already reaches this repo, and winget-releaser only bumps a
package that exists.
