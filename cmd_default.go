package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func printHelpDefault() {
	fmt.Printf(
		`Usage: %[1]s <ip | cidr | range | file | as | domain>... [<opts>]
       %[1]s <cmd> [<opts>] [<args>]

Description:
  Look up what every InternetData database says about an address: where it is,
  the AS and company behind it, its abuse contact, whether it is a VPN, a proxy,
  a Tor exit, a relay, a hosting, CDN, carrier or satellite network, and what it
  runs. An AS (AS15169) or a domain (example.com) is looked up the same way.

  Addresses, CIDRs, ranges and file paths are all accepted, in any mix, from
  arguments and standard input at the same time. An AS or a domain is typed as
  an argument; a list, piped or in a file, is read for addresses. One answer
  prints a readable block; more than one prints JSON.

  Anything that could not be answered gets its error in the output, in place
  of an answer, and the command then exits 1.

  It works with no account at all: a lookup answers where the address is, its
  AS, company and abuse contact, and whether it is a Tor exit, a relay,
  anycast, a carrier or satellite network, or runs SSH, SMTP, a web server, a
  DNS resolver or BitTorrent, with no daily cap. Every paid plan answers every
  field and looks addresses up in batches, and a new account can try every
  field, 100 lookups a day.

Examples:
  # One address, one AS, one domain.
  $ %[1]s 146.70.116.130
  $ %[1]s AS9009
  $ %[1]s example.com

  # Several, from anywhere.
  $ %[1]s 1.1.1.1 8.8.8.0/24 9.9.9.1-9.9.9.9 addresses.txt
  $ cat addresses.txt | %[1]s
  $ %[1]s 10.0.0.0/16 --csv > answers.csv

  # One field, for a script.
  $ %[1]s 146.70.116.130 -f location.country

Commands:
  myip        look up the address you are coming from.
  bulk        look up many addresses, always in the machine format.
  database    list, inspect and download the databases.
  signup      create an account.
  login       store an API key.
  logout      forget a stored API key.
  session     manage named credentials and switch between them.
  whoami      your key, your plan, and what you have used.
  cache       inspect or clear the local answer cache.
  config      read or change stored settings.
  completion  install shell auto-completion.
  version     print the version.

Options:
  General:
    --key <key>, -k <key>
      API key for this run, instead of the stored one. Every account on this
      machine can read it while it runs; prefer INTERNETDATA_API_KEY.
    --session <name>
      stored session to use for this run.
    --base-url <url>
      API to talk to, instead of the default.
    --nocache
      neither read nor write the local cache.
    --concurrency <n>
      in-flight requests during a bulk lookup.
    --retries <n>
      retries per request.
    --version, -v, --vsn
      print the version.
    --help, -h
      show help.

  Output:
    --field <field>, -f <field>
      only these fields, comma separated. Dotted for nested ones, as
      'location.country'. Naming an object selects all of it.
    --show-absent
      also list the fields your plan does not include, each as '-', in the
      readable block.
    --nocolor
      disable colored output.

  Formats:
    --pretty, -p     the readable block. (default for one answer)
    --json, -j       one JSON object keyed by what was looked up. (default for many)
    --jsonl          one JSON object per line, for pipelines.
    --csv, -c        CSV with a fixed header.
    --yaml, -y       YAML. Collected in memory, so not for large runs.
`, progBase)
}

// cmdDefault handles everything that is not a subcommand: what to look up, a
// request for help, or a mistake.
func cmdDefault() error {
	var fVersion bool
	globalFlags()
	lookupFlags()
	pflag.BoolVarP(&fVersion, "version", "v", false, "print the version.")
	pflag.BoolVar(&fVersion, "vsn", false, "print the version.")
	resolve := formatFlags()
	args := parseFlags()

	if fVersion {
		return cmdVersion()
	}
	if fHelp {
		printHelpDefault()
		return nil
	}

	opts, err := resolve(answerKinds(args)...)
	if err != nil {
		return err
	}

	// No address named and nothing piped or redirected in is a request for
	// help. Standard input is then a device: a terminal, not to be left waiting
	// on, or the null device a service, a `docker run` without -i and nohup are
	// given.
	if len(args) == 0 && isDevice(os.Stdin) {
		printHelpDefault()
		return nil
	}

	err = runLookup(context.Background(), args, opts)
	if err == errNoInput {
		// The user typed SOMETHING, so say what was wrong with it rather than
		// printing the whole help text over it.
		return errNothingFound
	}
	return err
}

// isDevice reports whether a file is a character device, such as a terminal or
// the null device, rather than a pipe or a file: exactly what the default
// command's input reader leaves unread. The file mode, not term.IsTerminal,
// because the null device gets the help too.
func isDevice(f *os.File) bool {
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}
