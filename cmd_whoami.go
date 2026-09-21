package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/pflag"

	internetdata "github.com/internetdata/sdk-go/v2"
)

func printHelpWhoami() {
	fmt.Printf(
		`Usage: %[1]s whoami [<opts>]

Description:
  Which credential this machine uses, and the databases its organization holds
  a license for, current or lapsed. '%[1]s db list' shows the rest of the
  catalog too.

Examples:
  $ %[1]s whoami
  $ %[1]s whoami --json
  $ %[1]s --session work whoami

Options:
  --json, -j
    output the licensed databases as JSON instead of the readable block.
  --help, -h
    show help.
`, progBase)
}

func cmdWhoami() error {
	var fJSON bool
	globalFlags()
	apiFlags()
	pflag.BoolVarP(&fJSON, "json", "j", false, "output JSON.")
	parseSubFlags()

	if fHelp {
		printHelpWhoami()
		return nil
	}

	key, source := gConfig.ResolveKey()
	api := gConfig.ResolveBaseURL()
	if api == "" {
		api = internetdata.DefaultBaseURL
	}

	if key == "" {
		fmt.Println("not authenticated")
		fmt.Printf("\nRun `%s login` to use a key, or `%s signup` to create an account.\n", progBase, progBase)
		return nil
	}

	client, err := NewClient()
	if err != nil {
		return err
	}

	// The credential first, because "which key am I even using" is the question
	// that brings most people here - and it is answered even when the API then
	// refuses that key.
	if !fJSON {
		fmt.Printf("key          %s\n", maskKey(key))
		fmt.Printf("from         %s\n", source)
		if name := gConfig.ActiveSessionName(); name != "" && gConfig.Sessions[name] != nil {
			fmt.Printf("session      %s\n", name)
		}
		fmt.Printf("api          %s\n", api)
	}

	// There is no entitlement endpoint here: what a key is good for is the
	// databases its organization holds, which the catalog says beside each one.
	databases, err := client.Database().List(context.Background())
	if err != nil {
		return explain(err)
	}
	held := make([]internetdata.Database, 0, len(databases))
	for _, d := range databases {
		if d.Standing != internetdata.StandingUnlicensed {
			held = append(held, d)
		}
	}

	if fJSON {
		return emitJSON(held)
	}
	if len(held) == 0 {
		fmt.Println("\nno databases licensed to this organization")
		return nil
	}
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DATABASE\tSTANDING\tLICENSE\tTERM")
	for _, d := range held {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", d.Base, d.Standing, licenseType(d), licenseTerm(d))
	}
	return w.Flush()
}
