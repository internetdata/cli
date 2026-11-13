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
- **Every SDK client is built with `namedHTTPClient()`**, which puts
  `internetdata-cli/<version>` first in the User-Agent, file transfer included,
  and keeps the SDK's 30 s timeout, which applies only when no client is passed.
  `NewClient` and `signInAPI` are the two today, and
  `TestEveryRequestNamesTheCLIFirst` covers only those: a third
  `internetdata.New` without it sends Go's default. sdk-go v2.9.1 sets no
  User-Agent of its own, so the CLI's name is the whole header until one does.

## Channel state (2026-11-13, v1.4.14)

GitHub release (its notes are the version's CHANGELOG.md section), ghcr
(public, anonymous pull), apt and `go install` were each verified serving
1.4.14. apt is signed since 2026-11-10: `InRelease`, `Release.gpg` and
`public.key` (fingerprint `590AB7732517F95B056ED35317EA110D248C1B73`), and the
README installs it with `signed-by`, proved on bookworm and Ubuntu 24.04. Its
index lists all 21 versions, 1.0.0 to 1.4.14. The
tap's formula names 1.4.14 with the release's own checksums; 1.0.0 was
installed from it (`brew trust` then `brew install`, Homebrew 7.0.4 on Linux).
The tap's SECURITY.md and README are kept by hand: the release job stages only
`Formula/`. Chocolatey approved every version through 1.4.13, each after a
human's review, and 1.4.14 is in moderation. winget took its first version,
1.4.7 as `InternetData.CLI`, on 2026-11-03 (microsoft/winget-pkgs#441588), and
every later one is the release's `winget` job, which must go green; 1.4.14's
PR is #479312. The job posts the CLA agree when a PR reads `Needs-CLA`, which
none since 1.4.12 has.
