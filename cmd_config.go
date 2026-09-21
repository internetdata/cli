package main

import (
	"fmt"
	"strconv"
	"strings"
)

func printHelpConfig() {
	fmt.Printf(
		`Usage: %[1]s config [list | <key>=<value>...]

Description:
  Read or change stored settings.

Examples:
  $ %[1]s config list
  $ %[1]s config retries=4

Settings:
  retries=<n>
    retries per request.

Options:
  --help, -h
    show help.

Credentials are not settings; see '%[1]s session'.
`, progBase)
}

func cmdConfig() error {
	globalFlags()
	args := parseSubFlags()

	// A bare `config` prints help rather than the settings. It is the shape
	// every other subcommand here has, and `config list` is one word away.
	if fHelp || len(args) == 0 {
		printHelpConfig()
		return nil
	}
	if len(args) == 1 && (args[0] == "list" || args[0] == "ls") {
		return configShow()
	}

	// Applied to a COPY, so a batch of settings with one bad value changes
	// nothing rather than applying up to the failure.
	next := gConfig
	for _, arg := range args {
		k, v, found := strings.Cut(arg, "=")
		if !found {
			return fmt.Errorf("expected <key>=<value>, got %q", arg)
		}
		if err := applySetting(&next, strings.ToLower(strings.TrimSpace(k)), strings.TrimSpace(v)); err != nil {
			return err
		}
	}
	if err := SaveConfig(next); err != nil {
		return err
	}
	gConfig = next
	return configShow()
}

// applySetting validates and applies one key=value.
func applySetting(cfg *Config, key, value string) error {
	switch key {
	case "retries":
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return fmt.Errorf("retries must be zero or more, got %q", value)
		}
		cfg.Retries = n
	default:
		return fmt.Errorf("%q is not a setting; `%s config --help` lists them", key, progBase)
	}
	return nil
}

func configShow() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	fmt.Printf("retries      %d\n", resolveRetries())
	fmt.Printf("\nstored in    %s\n", path)
	return nil
}
