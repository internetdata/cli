package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func printHelpDefault() {
	fmt.Printf(
		`Usage: %[1]s <cmd> [<opts>] [<args>]

Description:
  The databases your organization is licensed for: list them, see what is
  inside one before you fetch it, and download it, verified against its
  published checksum.

  Every database command needs an API key carrying the 'db.download' scope.
  '%[1]s login' signs this machine in and stores one.

Examples:
  $ %[1]s login
  $ %[1]s db list
  $ %[1]s db metadata vpn_ip_v1
  $ %[1]s db download vpn_ip_v1

Commands:
  database    list, inspect and download the licensed databases.
  signup      create an account.
  login       store an API key.
  logout      forget a stored API key.
  session     manage named credentials and switch between them.
  whoami      your key, and the databases it is licensed for.
  config      read or change stored settings.
  completion  install shell auto-completion.
  version     print the version.

Options:
  --key <key>, -k <key>
    API key for this run, instead of the stored one.
  --session <name>
    stored session to use for this run.
  --base-url <url>
    API to talk to, instead of the default.
  --retries <n>
    retries per request.
  --version, -v, --vsn
    print the version.
  --help, -h
    show help.
`, progBase)
}

// cmdDefault handles everything that is not a subcommand: a request for help,
// the version flag, or a mistake.
func cmdDefault() error {
	var fVersion bool
	globalFlags()
	apiFlags()
	pflag.BoolVarP(&fVersion, "version", "v", false, "print the version.")
	pflag.BoolVar(&fVersion, "vsn", false, "print the version.")
	args := parseFlags()

	if fVersion {
		return cmdVersion()
	}
	if len(args) > 0 {
		// The user typed SOMETHING, so say what was wrong with it rather than
		// printing the whole help text over it.
		return fmt.Errorf("%q is not a command; `%s --help` lists them", args[0], progBase)
	}
	printHelpDefault()
	return nil
}

// isTerminal reports whether a file is attached to a terminal rather than a
// pipe or a redirect.
func isTerminal(f *os.File) bool {
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}
