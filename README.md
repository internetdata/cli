# [<img src="https://s3.internetdata.io/internetdata-public/brand/mark.svg" alt="InternetData" width="24"/>](https://internetdata.io/) InternetData CLI

[![release](https://img.shields.io/github/v/release/internetdata/cli)](https://github.com/internetdata/cli/releases)
[![license](https://img.shields.io/github/license/internetdata/cli)](LICENSE)

The official command line interface for the [InternetData](https://internetdata.io) API.

List the IP and network databases your organization is licensed for, see what is inside one before you fetch it, and download it in CSV.GZ or MMDB, verified against its published checksum.

## Getting Started

Every database command needs an API key carrying the `db.download` scope. `internetdata login` signs this machine in through your browser and stores one.

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
echo "deb [trusted=yes] https://apt.internetdata.io/ /" | sudo tee /etc/apt/sources.list.d/internetdata.list
sudo apt update && sudo apt install internetdata
```

Or install a single `.deb` without the repository:

```bash
curl -Ls https://github.com/internetdata/cli/releases/latest/download/deb.sh | sh
```

### Windows

Install for the current user, which needs no admin rights:

```powershell
iwr -useb https://github.com/internetdata/cli/releases/latest/download/windows.ps1 | iex
```

### Docker

```bash
docker run --rm -e INTERNETDATA_API_KEY ghcr.io/internetdata/cli db list
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
curl -LO https://github.com/internetdata/cli/releases/download/v1.0.1/internetdata_1.0.1_linux_amd64.tar.gz
tar -xzf internetdata_1.0.1_linux_amd64.tar.gz
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
Usage: internetdata <cmd> [<opts>] [<args>]
...
```

### Login

`internetdata login` opens your browser, you confirm a short code, and you pick which of your API keys this machine should hold:

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

For a CI job, pass the key directly instead:

```console
$ internetdata login --key "$INTERNETDATA_API_KEY"
```

`internetdata login --paste` prompts for one without opening a browser, and `internetdata logout` signs this machine out, which ends the authorization rather than only deleting the local file.

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
$ internetdata --session work db list    # one command, without switching
```

`--base-url` points a session at another deployment of the API; a session without one talks to `https://internetdata.io`.

A key can also come from `--key` or the `INTERNETDATA_API_KEY` environment variable, in that order of precedence. The config file is written `0600` inside a `0700` directory.

### Who am I

```console
$ internetdata whoami
key          mk_1************************abcd
from         session default
session      default
api          https://internetdata.io

DATABASE   STANDING  LICENSE   TERM
vpn_ip     licensed  standard  renews 2027-01-04
bogon_ip   licensed  standard
```

The databases listed are the ones your organization holds a license for, current or lapsed; `--json` prints them as JSON.

### Databases

A license covers a database family, and a download names one of its versions, so the ids below come from `db list`:

```console
$ internetdata db list
ID            NAME      LICENSE   STANDING    TERM
vpn_ip_v1     VPN IP    standard  licensed    renews 2027-01-04
bogon_ip_v1   Bogon IP  standard  licensed
tor_ip_v1     Tor IP              unlicensed

$ internetdata db metadata vpn_ip_v1
$ internetdata db checksum vpn_ip_v1 --format mmdb
$ internetdata db download vpn_ip_v1
$ internetdata db download vpn_ip_v1 --format mmdb vpn.mmdb
$ curl -fL "$(internetdata db url vpn_ip_v1)" -o vpn_ip_v1.csv.gz
$ internetdata db downloads
```

`metadata` carries the columns, the row count, the build date and the size of each file without downloading anything. Downloads are verified against the published sha256 and land through a `.part` file, so an interrupted transfer never leaves a truncated file that reads as a whole database. `db url` hands back a time-limited link that carries no API key, for a downloader or another machine.

## Auto-Completion

Auto-completion is supported for at least the following shells:

```
bash
zsh
fish
```

It completes commands, flags, the values flags take, and the session names this machine actually holds.

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

## Other Libraries

There are official InternetData client libraries available for many languages including PHP, Python, Go, Java, Ruby, and many popular frameworks such as Django, Rails, and Laravel. See our GitHub at https://github.com/internetdata for more.

## About InternetData

IP, ASN and Domain data to reveal unique insights about the internet. APIs, Databases and Live Feeds available.

[<img src="https://s3.internetdata.io/internetdata-public/brand/mark.svg" alt="InternetData" width="96"/>](https://internetdata.io/)

## License

This project is licensed under the [MIT License](LICENSE).
