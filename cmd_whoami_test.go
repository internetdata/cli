package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// With no key, JSON is an error with nothing on standard output, because a
// script parses what is there and reads an exit 0 as an answer. The readable
// block still says so in prose and exits 0.
func TestWhoamiWithNoKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("INTERNETDATA_API_KEY", "")
	t.Setenv("INTERNETDATA_SESSION", "")
	savedArgs, savedFlags, savedConfig := os.Args, pflag.CommandLine, gConfig
	t.Cleanup(func() { os.Args, pflag.CommandLine, gConfig = savedArgs, savedFlags, savedConfig })

	for _, c := range []struct {
		args    []string
		wantErr bool
	}{
		{[]string{"--json"}, true},
		{[]string{"-j"}, true},
		{nil, false},
	} {
		os.Args = append([]string{"internetdata", "whoami"}, c.args...)
		pflag.CommandLine = pflag.NewFlagSet("internetdata", pflag.ContinueOnError)
		gConfig = NewConfig()
		out, err := captureStdout(t, cmdWhoami)
		if c.wantErr && (err == nil || out != "") {
			t.Errorf("whoami %v: printed %q, err %v; want an error and nothing printed", c.args, out, err)
		}
		if !c.wantErr && (err != nil || !strings.HasPrefix(out, "not authenticated")) {
			t.Errorf("whoami %v: printed %q, err %v; want the notice", c.args, out, err)
		}
	}
}
