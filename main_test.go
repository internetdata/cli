package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
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

// A bare command group prints its help, as a bare invocation does, rather than
// running whichever of its subcommands would be most useful.
func TestBareGroupPrintsHelp(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("INTERNETDATA_API_KEY", "")
	t.Setenv("INTERNETDATA_SESSION", "")
	savedArgs, savedFlags, savedConfig := os.Args, pflag.CommandLine, gConfig
	t.Cleanup(func() { os.Args, pflag.CommandLine, gConfig = savedArgs, savedFlags, savedConfig })

	cases := []struct {
		name string
		cmd  func() error
	}{
		{"session", cmdSession},
		{"sessions", cmdSession},
		{"database", cmdDatabase},
		{"db", cmdDatabase},
		{"config", cmdConfig},
		{"completion", cmdCompletion},
	}
	for _, c := range cases {
		os.Args = []string{"internetdata", c.name}
		pflag.CommandLine = pflag.NewFlagSet("internetdata", pflag.ContinueOnError)
		gConfig = NewConfig()
		out, err := captureStdout(t, c.cmd)
		if err != nil || !strings.HasPrefix(out, "Usage: ") {
			t.Errorf("bare %s: printed %q, err %v; want its help", c.name, out, err)
		}
	}
}

// captureStdout runs fn with standard output sent to a file, and returns what
// it printed.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })

	saved := os.Stdout
	os.Stdout = file
	fnErr := fn()
	os.Stdout = saved
	printed, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(printed), fnErr
}
