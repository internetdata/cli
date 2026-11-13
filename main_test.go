package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// TestMain runs main itself when a test re-runs this binary, so a test can read
// what a shell reads from it.
func TestMain(m *testing.M) {
	if os.Getenv("INTERNETDATA_TEST_RUN_MAIN") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// A flag may come before the subcommand, and a flag VALUE that happens to spell
// a subcommand must not be taken for one.
func TestDetectCommand(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"internetdata"}, ""},
		{[]string{"internetdata", "1.1.1.1"}, ""},
		{[]string{"internetdata", "1.1.1.1", "2.2.2.2"}, ""},
		{[]string{"internetdata", "AS15169"}, ""},
		{[]string{"internetdata", "example.com"}, ""},
		{[]string{"internetdata", "db", "list"}, "db"},
		{[]string{"internetdata", "database", "list"}, "database"},

		// A flag first.
		{[]string{"internetdata", "--nocache", "db", "list"}, "db"},
		{[]string{"internetdata", "--session", "work", "db", "list"}, "db"},
		{[]string{"internetdata", "-k", "secret", "whoami"}, "whoami"},
		{[]string{"internetdata", "--session=work", "db"}, "db"},

		// The value of a value-taking flag is never a command, however it is
		// spelled.
		{[]string{"internetdata", "--session", "database"}, ""},
		{[]string{"internetdata", "--session", "version", "1.1.1.1"}, ""},
		{[]string{"internetdata", "-f", "version", "1.1.1.1"}, ""},

		// An address cannot be followed by a command; everything after it is
		// more input.
		{[]string{"internetdata", "1.1.1.1", "db"}, ""},
		{[]string{"internetdata", "example.com", "version"}, ""},

		// Aliases.
		{[]string{"internetdata", "entitlement"}, "entitlement"},
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
		{"cache", cmdCache},
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
