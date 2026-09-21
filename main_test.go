package main

import (
	"os"
	"testing"
)

// A flag may come before the subcommand, and a flag VALUE that happens to spell
// a subcommand must not be taken for one.
func TestDetectCommand(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"internetdata"}, ""},
		{[]string{"internetdata", "db", "list"}, "db"},
		{[]string{"internetdata", "database", "list"}, "database"},

		// A flag first.
		{[]string{"internetdata", "--retries", "3", "db", "list"}, "db"},
		{[]string{"internetdata", "--session", "work", "db", "list"}, "db"},
		{[]string{"internetdata", "-k", "secret", "whoami"}, "whoami"},
		{[]string{"internetdata", "--session=work", "db"}, "db"},

		// The value of a value-taking flag is never a command, however it is
		// spelled.
		{[]string{"internetdata", "--session", "database"}, ""},
		{[]string{"internetdata", "--session", "version", "whoami"}, "whoami"},
		{[]string{"internetdata", "--format", "version", "db"}, "db"},

		// A word that is not a command cannot be followed by one: it is the
		// mistake the default command names.
		{[]string{"internetdata", "bogus", "db"}, ""},

		// Aliases.
		{[]string{"internetdata", "init"}, "init"},
		{[]string{"internetdata", "vsn"}, "vsn"},
		{[]string{"internetdata", "v"}, "v"},
	}
	for _, c := range cases {
		saved := os.Args
		os.Args = c.args
		got := detectCommand()
		os.Args = saved
		if got != c.want {
			t.Errorf("detectCommand(%v) = %q, want %q", c.args, got, c.want)
		}
	}
}
