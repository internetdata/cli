# internetdata/cli/internetdata

The official CLI, a presentation layer over `github.com/internetdata/sdk-go/v2`.
Public repo: `github.com/internetdata/cli`, with gitea as a second push URL.

Structure: a flat dispatch in `main.go`, one `cmd_<name>.go` per command,
per-command completion predictors in `completions.go`, and output machinery in
`lib/`. It looks up an address, an AS or a domain (since 1.5.0), and lists,
describes and downloads the databases.

## The non-obvious parts

- **An AS (`AS15169`) or a domain (`example.com`) is a lookup too**, by the
  SDK's `LookupASN` and `LookupDomain`: one request each, never cached, never
  batched, answered after the addresses in the order typed. Only a TYPED
  argument is one: a list, piped or in a file, is read for addresses, so a log
  full of hostnames costs no domain lookups. `classify` decides: an address,
  range or CIDR, then `AS<n>`, then a file that exists (a list file reads like a
  domain), then a domain's shape (`isDomain`: LDH labels, a top label not all
  digits, so `8.8.8.888` is refused here rather than sent).
- **Each kind has its own columns**: `lib.Columns`, `ASNColumns` and
  `DomainColumns`, each append-only, as `lib.Columns` says, and the writers take
  the kind with every answer. A CSV's header is its first answer's kind, so a
  CSV of two kinds is refused (`lib.ErrMixedCSV`), before any request when the
  arguments alone mix them. A `--field` is checked against the kinds the
  arguments ask for (`answerKinds`): addresses always, as stdin may add them.
  An AS's error line names it by its number, `{"asn":15169,...}`, and its JSON
  key is `AS<n>` however it was typed.
- **The AS whois record is a map in the SDK**, so its 39 columns come from the
  record the spec's paid example documents, and
  `TestASNWhoisColumnsAreTheDocumentedKeys` holds them to that answer's keys,
  where `TestColumnsAreTheSDKsFields` holds every other column to the SDK's
  types.
- **Keys are default-deny**: a key without the `db.download` scope looks
  addresses up and is refused by every database call, as a 401. So database
  errors go through `explainDatabase`, which names the scope, an address the
  allowed addresses leave out and the license apart, and `checkKey` validates
  a key at login with a lookup, never the listing first. Since 2026-11-13 a key
  its owner took `ip.lookup` from is refused every lookup and `whoami` as a 403,
  so login asks the listing after a 403 and stores a key that answers it, with a
  warning. Never on an `http:` base URL: the listing's host redirects plain HTTP
  to HTTPS and Go keeps the key across it, so a build without that guard stored
  a key the lookup refused for plain HTTP (2026-11-14). A lookup's 403 is never
  a plan, and the API's message names its reason, so `explain` adds no hint to
  one.
- **Calls go to `https://api.internetdata.io`, sign-in to the apex**, which the
  SDK's `client.Oauth` derives from the base URL (`api.` dropped, `api-<env>.`
  becoming `<env>.`); `TestSignInGoesToTheApexAndTheRestToTheAPIHost` pins
  both. The apex answers the database paths too, so a session may hold it:
  `signupURL()` maps the apex to `app.`, a three-label `E.internetdata.io` to
  `app-E.internetdata.io`, and an `api`-labeled host keeps the sibling shape.
  `cmd_signup_test.go` pins it.

## Channel state (2026-11-14, v1.5.1)

GitHub release (its notes are the version's CHANGELOG.md section), ghcr
(public, anonymous pull, `1.5.1` and `latest`), apt and `go install` were each
verified serving 1.5.1. The image carries OCI labels for the site, docs, source
and license, and its index's description annotation is what GitHub's package
page shows. apt is signed since 2026-11-10: `InRelease`, `Release.gpg` and
`public.key` (fingerprint `590AB7732517F95B056ED35317EA110D248C1B73`), and the
README installs it with `signed-by`; a fresh bookworm installed 1.5.1 that way.
Its index lists all 23 versions, 1.0.0 to 1.5.1. The tap's formula names 1.5.1
with the release's own checksums. The tap's SECURITY.md and README are kept by
hand: the release job stages only `Formula/`. Chocolatey approved every version
through 1.4.14, each after a human's review; 1.5.0 and 1.5.1 are in moderation.
winget took its first version, 1.4.7 as `InternetData.CLI`, on 2026-11-03
(microsoft/winget-pkgs#441588), and every later one is the release's `winget`
job, which must go green; 1.5.0 merged, with its locale's `Description`
rewritten on the PR's branch for the lookups, since the job copies the last
version's forward, and 1.5.1's PR is #480032. The job posts the CLA agree when
a PR reads `Needs-CLA`, which none since 1.4.12 has.
