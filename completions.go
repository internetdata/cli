package main

import (
	complete "github.com/mslmio/libgo-complete"
	"github.com/mslmio/libgo-complete/predict"
)

// formats is every --format value, offered wherever one is taken.
var formats = predict.Set([]string{"csvgz", "mmdb"})

// globalCompletions are the options every command accepts.
func globalCompletions() map[string]complete.Predictor {
	return map[string]complete.Predictor{
		"--key":      predict.Something,
		"-k":         predict.Something,
		"--session":  sessionNames(),
		"--base-url": predict.Something,
		"--retries":  predict.Something,
		"--help":     predict.Nothing,
		"-h":         predict.Nothing,
	}
}

// sessionNames offers the sessions this machine actually holds, which is the
// completion that saves the most typing and the one a static list cannot give.
func sessionNames() complete.Predictor {
	return predict.Func(func(string) []string { return gConfig.SessionNames() })
}

// jsonFlags are the ones every command with a JSON form takes.
var jsonFlags = map[string]complete.Predictor{"--json": predict.Nothing, "-j": predict.Nothing}

var completions = &complete.Command{
	Sub: map[string]*complete.Command{
		"database": databaseCompletions,
		"db":       databaseCompletions,
		"login":    loginCompletions,
		"init":     loginCompletions,
		"logout": {Flags: map[string]complete.Predictor{
			"--session": sessionNames(),
			"--all":     predict.Nothing,
		}},
		"signup":     {Flags: map[string]complete.Predictor{"--no-browser": predict.Nothing}},
		"session":    sessionCompletions,
		"sessions":   sessionCompletions,
		"whoami":     {Flags: jsonFlags},
		"config":     {Args: predict.Set([]string{"list", "retries="})},
		"completion": {Args: predict.Set([]string{"install", "uninstall", "bash", "zsh", "fish"})},
		"version":    {},
	},
	Flags: map[string]complete.Predictor{
		"--version": predict.Nothing,
		"--vsn":     predict.Nothing,
		"-v":        predict.Nothing,
	},
}

// databaseCompletions is shared by `database` and its `db` alias, so the two
// cannot drift into offering different subcommands.
var databaseCompletions = &complete.Command{
	Sub: map[string]*complete.Command{
		"list":     {Flags: jsonFlags},
		"metadata": {},
		"checksum": {Flags: map[string]complete.Predictor{
			"--format": formats, "--json": predict.Nothing, "-j": predict.Nothing,
		}},
		"url": {Flags: map[string]complete.Predictor{"--format": formats}},
		"downloads": {Flags: map[string]complete.Predictor{
			"--limit": predict.Something, "--json": predict.Nothing, "-j": predict.Nothing,
		}},
		"download": {Flags: map[string]complete.Predictor{
			"--format":    formats,
			"--stdout":    predict.Nothing,
			"--no-verify": predict.Nothing,
		}},
	},
}

// loginCompletions is shared by `login` and its `init` alias.
var loginCompletions = &complete.Command{
	Flags: map[string]complete.Predictor{
		"--session":    sessionNames(),
		"--base-url":   predict.Something,
		"--key":        predict.Something,
		"-k":           predict.Something,
		"--no-check":   predict.Nothing,
		"--paste":      predict.Nothing,
		"--no-browser": predict.Nothing,
	},
}

// sessionCompletions is shared by `session` and its `sessions` alias.
var sessionCompletions = &complete.Command{
	Sub: map[string]*complete.Command{
		"list":   {},
		"use":    {Args: sessionNames()},
		"show":   {Args: sessionNames()},
		"rename": {Args: sessionNames()},
		"rm":     {Args: sessionNames()},
	},
}

// handleCompletions answers a shell completion request, if this is one.
func handleCompletions() {
	// Also the top-level flag set, so completing at the very start offers the
	// options every command takes.
	for flag, p := range globalCompletions() {
		if _, taken := completions.Flags[flag]; !taken {
			completions.Flags[flag] = p
		}
	}
	completions.Complete(progBase)
}
