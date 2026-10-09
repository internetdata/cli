# internetdata/cli/internetdata

The official CLI, a presentation layer over `github.com/internetdata/sdk-go/v2`.
Public repo: `github.com/internetdata/cli`, with gitea as a second push URL.

Structure: a flat dispatch in `main.go`, one `cmd_<name>.go` per command, and
per-command completion predictors in `completions.go`. It lists, describes and
downloads the databases an organization is licensed for; there is no per-IP
lookup, result cache or field selection.

## The non-obvious parts

- **The API is at the apex (`internetdata.io`)**, so `signupURL()` has no `api`
  label to swap: the apex maps to `app.`, a three-label `E.internetdata.io` to
  `app-E.internetdata.io`, and an `api`-labeled host keeps the sibling shape -
  the same rule the API applies to its device page.
  Anything else falls back to production. `cmd_signup_test.go` pins it.
- **`whoami` has no entitlement to show.** It prints the credential first, so a
  refused key still says which key it was, then the families the organization
  holds a license for (standing not `unlicensed`) from `Database.List()`.
  `checkKey` validates through the same call, so a key without `db.download` is
  refused at login, and it arrives as a 401, not a 403: keys are default-deny.
- **A bare invocation prints help, and an unknown word is named as a mistake.**
  There is no default command to fall into.
- **No color.** Nothing prints any, so there is no `--nocolor` and no
  `fatih/color`.

## Channel state (2026-10-09, v1.3.0)

GitHub release (its notes are the version's CHANGELOG.md section), ghcr
(public, anonymous pull), apt (`apt.internetdata.io`, a clean bookworm install)
and `go install` were each verified installing 1.3.0. The tap's formula names
1.3.0 with the release's own checksums; 1.0.0 was installed from it (`brew
trust` then `brew install`, Homebrew 7.0.4 on Linux). Chocolatey approved every
version through 1.2.0, so the README names `choco install`; 1.3.0 is in
moderation, since every version waits for a human.
winget's first manifest is microsoft/winget-pkgs#441588, as `InternetData.CLI`
(the `Mslm.InternetData` PR was withdrawn unmerged, 2026-09-26), and the README
names winget once it installs. Until that PR merges
every release's `winget` job is red by design: the org's `WINGET_TOKEN` already
reaches this repo, and winget-releaser only bumps a package that exists.
