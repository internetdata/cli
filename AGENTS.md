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

## Channel state (2026-11-10, v1.4.12)

GitHub release (its notes are the version's CHANGELOG.md section), ghcr
(public, anonymous pull), apt and `go install` were each verified serving
1.4.12. apt is signed since 2026-11-10: `InRelease`, `Release.gpg` and
`public.key` (fingerprint `590AB7732517F95B056ED35317EA110D248C1B73`), and the
README installs it with `signed-by`, proved on bookworm and Ubuntu 24.04. The
tap's formula names 1.4.12 with the release's own checksums; 1.0.0 was
installed from it (`brew trust` then `brew install`, Homebrew 7.0.4 on Linux).
The tap's SECURITY.md and README are kept by hand: the release job stages only
`Formula/`. Chocolatey approved every version through 1.4.12, each after a
human's review. winget took its first version, 1.4.7 as `InternetData.CLI`, on
2026-11-03 (microsoft/winget-pkgs#441588), and every later one is the
release's `winget` job, which must go green; 1.4.12 (#475319) merged on
2026-11-09. The job posts the CLA agree when a PR reads `Needs-CLA`, which
1.4.12's never did.
