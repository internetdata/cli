# [<img src="https://s3.internetdata.io/internetdata-public/brand/mark.svg" alt="InternetData" height="28"/>](https://internetdata.io/) InternetData CLI

[![release](https://img.shields.io/github/v/release/internetdata/cli)](https://github.com/internetdata/cli/releases)
[![license](https://img.shields.io/github/license/internetdata/cli)](LICENSE)

The official command line interface for the [InternetData](https://internetdata.io) API.

Look up what every InternetData database says about an IP address - where it is, the AS and company behind it, its abuse contact, whether it is a VPN, a proxy, a Tor exit, a relay, or a hosting, CDN, carrier or satellite network, and what it runs - one address at a time or a few million, look up an AS or a domain, and download the databases.

## Getting Started

No account is needed to start: Unlimited, the free plan, answers any address with no cap, with a key or without one: where it is, the AS and company behind it, its abuse contact, and whether it is a Tor exit, a relay, anycast, a carrier or satellite network, or runs SSH, SMTP, a web server, a DNS resolver or BitTorrent, one address per request. Every paid plan answers all our data fields and looks addresses up in batches of up to 1,000, and a free account tries every field, 100 lookups a day.

### macOS

```bash
brew trust internetdata/tap
brew tap internetdata/tap
brew install internetdata
```

Homebrew won't load a formula from a third-party tap until you trust it. Skip that first line and it fails with `invalid syntax in tap!`, which is misleading: the formula is fine. `brew trust` arrived in Homebrew 7, so run `brew update` first if it comes back as an unknown command.

The CLI needs macOS 13 Ventura or later.

### Debian / Ubuntu

Install from our apt repository, which keeps the CLI up to date with `apt upgrade`:

```bash
sudo install -d -m 0755 /etc/apt/keyrings
curl -fsSL https://apt.internetdata.io/public.key | sudo tee /etc/apt/keyrings/internetdata.asc > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/internetdata.asc] https://apt.internetdata.io/ /" | sudo tee /etc/apt/sources.list.d/internetdata.list
sudo apt update && sudo apt install internetdata
```

The repository is signed by the key with fingerprint `590AB7732517F95B056ED35317EA110D248C1B73`, and apt refuses an index it did not sign. If you added the repository with `[trusted=yes]` before, replace that line with the one above: apt still installs from it, but warns on every update that it cannot check the signature.

Or install a single `.deb` without the repository:

```bash
curl -Ls https://github.com/internetdata/cli/releases/latest/download/deb.sh | sh
```

### Windows

```powershell
winget install InternetData.CLI
```

Or with Chocolatey:

```powershell
choco install internetdata
```

Or install for the current user, which needs no admin rights:

```powershell
iwr -useb https://github.com/internetdata/cli/releases/latest/download/windows.ps1 | iex
```

### Docker

```bash
docker run --rm ghcr.io/internetdata/cli 146.70.116.130
```

### Using `go install`

```bash
go install github.com/internetdata/cli@latest
```

The binary is named after the module, so rename it to `internetdata` if you want the examples below to read the same.

### Using `curl` / `wget`

Binaries are published for 23 platform and architecture pairs on the [releases page](https://github.com/internetdata/cli/releases). Pick yours:

```bash
# Linux amd64; for Windows use ".zip" instead of ".tar.gz"
curl -LO https://github.com/internetdata/cli/releases/download/v1.4.14/internetdata_1.4.14_linux_amd64.tar.gz
tar -xzf internetdata_1.4.14_linux_amd64.tar.gz
sudo mv internetdata /usr/local/bin/
```

macOS has a one-line installer that picks the right architecture for you:

```bash
curl -Ls https://github.com/internetdata/cli/releases/latest/download/macos.sh | sh
```

Note that the binaries are not code-signed, so macOS Gatekeeper will ask before running one the first time, and Windows SmartScreen may warn.

### From source

```bash
git clone https://github.com/internetdata/cli
cd cli
./scripts/build.sh
```

The result lands in `build/`. Go 1.26 or newer.

## Quick Start

### Default help message

Invoking the CLI with nothing shows what it can do:

```console
$ internetdata
Usage: internetdata <ip | cidr | range | file | as | domain>... [<opts>]
       internetdata <cmd> [<opts>] [<args>]
...
```

### Login

You can use the CLI without an account, but a key widens the answer and raises your allowance. `internetdata login` opens your browser, you confirm a short code, and you pick which of your API keys this machine should hold:

```console
$ internetdata login
opened your browser to finish signing in.

  https://app.internetdata.io/device

and confirm this code:  MA8A-T2WQ

waiting for you to approve...
stored key mk_9************************************Etjx in session "default", now active
```

Nothing is typed or pasted here, so the key never reaches your shell history. If no browser opens, the URL is printed and you can open it anywhere, including on your phone - which is what makes this work over SSH.

`internetdata signup` is the same thing with the sign-up page first, so a new account and a working CLI are one step.

For a CI job, set `INTERNETDATA_API_KEY` in the job's environment from your CI's secret store instead. Every command reads it, so nothing is stored on the runner, and the key never sits on a command line, where every account on the machine can read it from the process list:

```console
$ export INTERNETDATA_API_KEY=...
$ internetdata 146.70.116.130
```

`internetdata login --paste` prompts for one without opening a browser, and `internetdata logout` signs this machine out, which ends the authorization rather than only deleting the local file.

### My IP

```console
$ internetdata myip -f ip,location.country
ip                146.70.116.130
location.country  AT
```

This is the address our edge observed, so behind a proxy or a VPN it reports the exit you left through.

### Any IP

```console
$ internetdata 146.70.116.130
ip                        146.70.116.130
is_bogon                  false
is_tor                    false
is_relay                  false
is_anycast                false
is_carrier                false
is_satellite              false
is_ssh                    false
is_smtp                   false
is_web                    false
is_dns_resolver           false
is_bittorrent             false
tor                       {}
relay                     {}
location.continent        EU
location.continent_name   Europe
location.country          AT
location.country_name     Austria
location.timezone         Europe/Vienna
location.currency_code    EUR
location.currency_name    Euro
location.currency_symbol  €
location.flag_emoji       🇦🇹
location.flag_unicode     U+1F1E6 U+1F1F9
asn.asn                   9009
asn.name                  M247 Europe SRL
asn.domain                m247global.com
asn.type                  hosting
asn.prefix                146.70.116.0/24
company.name              M247 LTD Vienna Infrastructure
company.domain            m247.ro
abuse.email               abuse@m247.ro
hostname                  ""
```

That is Unlimited's answer, without a key. A paid plan adds `is_vpn`, `is_hosting`, `is_cdn` and the four proxy flags, the provider and confidence behind each flag, what each service answered, the location down to the city, and the whois records.

Private, reserved and documentation addresses are answered locally without a request, so they cost nothing and work offline.

### An AS or a domain

```console
$ internetdata AS9009
asn        9009
name       M247 Europe SRL
country    RO
registry   ripe
allocated  2005-06-06
domain     m247global.com

$ internetdata example.com
domain         example.com
is_registered  true
has_a          true
has_aaaa       true
has_mx         true
has_ns         true
has_cname      false
```

Every plan answers an AS's registration, and whether a domain is registered and which record types it has. A paid plan adds the AS's whois record and the prefixes it announces, and the domain's records, its organization's IP ranges and `is_disposable`. Each is one request, never cached or batched, and is typed as an argument: a list, piped or in a file, is read for addresses. An AS or a domain beside addresses is answered after them.

### Piping

Addresses can come from standard input, and from arguments, at the same time:

```console
$ cat addresses.txt | internetdata
$ internetdata 1.1.1.1 8.8.8.0/24 9.9.9.1-9.9.9.9 addresses.txt
```

Anything resolving to more than one answer prints JSON keyed by what was looked up; a single answer prints the readable block. That follows the resolved **count**, so `10.0.0.1/32` is one and `10.0.0.0/30` is four.

### Field filter

```console
$ internetdata 146.70.116.130 -f location.country
location.country  AT

$ internetdata 146.70.116.130 -f location.country,asn.name
location.country  AT
asn.name          M247 Europe SRL

$ internetdata 146.70.116.130 -f location          # the whole object
```

It applies to every format. JSON, JSON Lines and YAML keep the answer's nesting, and leave out a field the answer does not carry:

```console
$ internetdata bulk 146.70.116.130 1.1.1.1 --jsonl -f ip,is_tor,asn.name
{"asn":{"name":"M247 Europe SRL"},"ip":"146.70.116.130","is_tor":false}
{"asn":{"name":"Cloudflare, Inc."},"ip":"1.1.1.1","is_tor":false}
```

An AS's and a domain's fields are their own, as `internetdata AS9009 -f name,registry`.

### Bulk

`bulk` always prints the machine format, whatever the count, which is what a script wants:

```console
$ internetdata bulk 8.8.8.0/24 --csv > answers.csv
$ internetdata bulk addresses.txt --jsonl | jq -r 'select(.is_tor) | .ip'
```

Bulk runs stream: nothing is held in memory beyond one batch, so a `/8` is a long command rather than an impossible one, and output starts immediately. With no arguments at a terminal it reads addresses you type, one per line, ending at a blank line.

Output formats are `--pretty`, `--json`, `--jsonl`, `--csv` and `--yaml`. A CSV holds one kind of answer, so look addresses, AS numbers and domains up in separate runs for it.

### Exit status

A lookup exits 1 when anything could not be answered, and 0 when everything was. Whatever went wrong - a key the API refused, a spent allowance, the API out of reach - that address, AS or domain gets its error in the output, in place of its answer, and everything else is answered as usual. A script that can live with gaps adds `|| true` and finds them in the output:

```console
$ internetdata bulk addresses.txt --jsonl > answers.jsonl || true
$ jq -r 'select(.error) | .ip' answers.jsonl
```

### Sessions

Credentials are stored in named sessions, so one machine can hold several organizations' keys and switch between them without logging in again:

```console
$ internetdata login --session work
$ internetdata login --session acme --base-url https://api.example.com
$ internetdata session list
   NAME     KEY                      API                      LAST USED
   acme     mk_4***************mnop  https://api.example.com  never
*  default  mk_1***************abcd  default                  never
   work     mk_9***************wxyz  default                  never

$ internetdata session use work
$ internetdata --session work 146.70.116.130    # one command, without switching
```

`--base-url` points a session at another deployment of the API; a session without one talks to `https://api.internetdata.io`, and signs in at `https://internetdata.io`.

A key can also come from `--key` or the `INTERNETDATA_API_KEY` environment variable, in that order of precedence. The config file is written `0600` inside a `0700` directory.

### Who am I

```console
$ internetdata whoami
key          mk_1************************abcd
from         session default
session      default
api          https://api.internetdata.io

DATABASE   STANDING  LICENSE   TERM
vpn_ip     licensed  standard  renews 2027-01-04
bogon_ip   licensed  standard
```

The databases listed are the ones your organization holds a license for, current, lapsed or yet to start; `--json` prints them as JSON. A license signed ahead of its start date reads `expired` until then, and its TERM says when it starts.

### Databases

The same data, published as files you host yourself. A download needs a key carrying the `db.download` scope; a key without it still looks addresses up. A license covers a database family, and a download names one of its versions, so the IDs below come from `db list`:

```console
$ internetdata db list
ID            NAME      LICENSE   STANDING    TERM
vpn_ip_v1     VPN IP    standard  licensed    renews 2027-01-04
bogon_ip_v1   Bogon IP  standard  licensed
cdn_ip_v1     CDN IP              unlicensed

$ internetdata db metadata vpn_ip_v1
$ internetdata db checksum vpn_ip_v1 --format mmdb
$ internetdata db download vpn_ip_v1
$ internetdata db download vpn_ip_v1 --format mmdb vpn.mmdb
$ curl -fL "$(internetdata db url vpn_ip_v1)" -o vpn_ip_v1.csv.gz
$ internetdata db downloads
```

Every organization holds a standard license to the Open databases, `bogon_ip_v1` and `tor_ip_v1` among them, which are free under [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/). `db list --json` marks each one `"open": true`.

`metadata` carries the columns, the row count, the build date and the size of each file without downloading anything. Downloads are verified against the published sha256 and land through a `.part` file, so an interrupted transfer never leaves a truncated file that reads as a whole database. `db url` hands back a time-limited link that carries no API key, for a downloader or another machine. `db downloads` lists your organization's recent download attempts, refusals included, and marks the download of an evaluation sample `(sample)` beside its database.

## Auto-Completion

Auto-completion is supported for at least the following shells:

```
bash
zsh
fish
```

It completes commands, flags, the values flags take, field names, and the session names this machine actually holds.

NOTE: it may work for other shells as well, because the implementation is in Go and is not shell-specific.

### Installation

Installing auto-completions is as simple as running one command, which covers `bash`, `zsh` and `fish`:

```bash
internetdata completion install
```

Start a new shell afterwards to pick it up.

If you want to customize the installation (for example when the auto-installation does not work as expected), you can request the completion line for each shell instead:

```bash
# get bash completion script
internetdata completion bash

# get zsh completion script
internetdata completion zsh

# get fish completion script
internetdata completion fish
```

`internetdata completion uninstall` removes it again.

### Shell not listed?

If your shell is not listed here, you can open an issue.

Note that as long as the `COMP_LINE` environment variable is provided to the binary itself, it will output completion results. So if your shell provides a way to pass `COMP_LINE` on auto-completion attempts to a binary, then have your shell do that with the `internetdata` binary itself.

## Data

How much you get back per lookup depends on the plan behind your key. See the [plans](https://internetdata.io/pricing) for what each includes. The examples in this document use no key.

**A field your plan does not include is missing from the answer, not `false`.** `is_vpn` absent means it is not in your plan; `is_vpn: false` means we checked and it is not a VPN. The readable output prints only what you were served; CSV writes an empty cell for absent and the literal `false` for false. `--show-absent` renders the lot.

On a key with no paid plan, each answer carries `trial_expired`: `false` for the day's 100 full lookups, `true` once they're spent and the answer is Unlimited's. `internetdata whoami` counts them.

## Caching

Address answers are cached on disk, so looking the same address up twice costs one request rather than two. An AS or a domain is never cached.

```bash
internetdata cache info
internetdata cache clear
internetdata --nocache 146.70.116.130
internetdata config list
internetdata config cache=disable cache_ttl=24h
```

Entries are partitioned by **credential**, because which fields an answer carries depends on the plan behind the key - one key's answers are never served to another's. Clearing empties every partition.

## Color Output

### Disabling color output

The CLI respects either the `--nocolor` flag or the [`NO_COLOR`](https://no-color.org/) environment variable to disable color output.

### Color on Windows

To enable color support for the Windows command prompt, run the following to enable [`Console Virtual Terminal Sequences`](https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences):

```cmd
REG ADD HKCU\CONSOLE /f /v VirtualTerminalLevel /t REG_DWORD /d 1
```

You can disable this by running the following:

```cmd
REG DELETE HKCU\CONSOLE /f /v VirtualTerminalLevel
```

## Other Libraries

There are official InternetData client libraries available for many languages including PHP, Python, Go, Java, and Ruby. See our GitHub at https://github.com/internetdata for more.

## About InternetData

Geolocation, anonymity, ownership and network databases for IP addresses and AS numbers, licensed as files you download and query yourself.

[<img src="https://s3.internetdata.io/internetdata-public/brand/mark.svg" alt="InternetData" height="64"/>](https://internetdata.io/)

## License

This project is licensed under the [MIT License](LICENSE).
