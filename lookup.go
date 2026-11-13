package main

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/pflag"

	internetdata "github.com/internetdata/sdk-go/v2"
	iputil "github.com/mslmio/libgo-iputil"

	"github.com/internetdata/cli/lib"
)

// chunkSize is how many addresses are looked up per batch.
//
// The input is unbounded - a single `8.0.0.0/8` argument is 16.7 million
// addresses - so it is consumed in chunks rather than collected. Each chunk is
// one SDK batch, whose answers are emitted and dropped before the next chunk is
// read, which puts a fixed ceiling on memory whatever was typed.
const chunkSize = 10_000

// lookupOpts is what the lookup commands were asked for.
type lookupOpts struct {
	// format is the explicit --format, or "" to choose by answer count.
	format lib.Format
	// forceBulk skips the single-answer form even for one address, which is
	// what the explicit `bulk` subcommand wants so its output shape is
	// predictable in a script.
	forceBulk  bool
	fields     []string
	showAbsent bool
	// input controls where addresses are read from.
	input iputil.Opts
}

// target is an AS or a domain typed as an argument. Each is its own request,
// as the SDK answers them: never cached, never batched.
type target struct {
	kind lib.Kind
	// key names it in the output: `AS<n>`, or the domain as typed.
	key string
	asn uint32
}

// runLookup reads addresses from args and standard input and answers them, then
// the AS numbers and domains typed as arguments.
//
// One answer gets the readable block; more than one gets the machine format.
// That is decided by how many addresses the input RESOLVES to, not by how many
// arguments there were: a single `10.0.0.0/30` is four addresses and a single
// `10.0.0.1/32` is one. Each AS or domain is one answer more.
func runLookup(ctx context.Context, args []string, opts lookupOpts) error {
	targets, err := checkArgs(args)
	if err != nil {
		return err
	}
	if opts.format == lib.FormatCSV && len(typedKinds(args)) > 1 {
		return lib.ErrMixedCSV
	}
	client, err := NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	var (
		chunk      []string
		writer     lib.Writer
		flushed    bool
		unanswered bool
	)

	// Buffered until the shape is known, because the format depends on the
	// answer count and nothing may be written before that is settled. The wait
	// is bounded by chunkSize, not by the size of the input.
	openWriter := func(single bool) {
		format := opts.format
		if format == "" {
			format = lib.FormatJSON
			if single {
				format = lib.Format(gConfig.Format)
				if format == "" {
					format = lib.FormatPretty
				}
			}
		}
		writer = lib.NewWriter(os.Stdout, format, lib.Opts{
			Fields:     opts.fields,
			ShowAbsent: opts.showAbsent,
			Color:      !color.NoColor,
		})
	}

	flush := func() error {
		if len(chunk) == 0 {
			return nil
		}
		if writer == nil {
			openWriter(!flushed && len(chunk) == 1 && len(targets) == 0 && !opts.forceBulk)
		}
		flushed = true
		answers, err := client.LookupBatch(ctx, chunk)
		if err != nil {
			return explain(err)
		}
		for _, ip := range chunk {
			answer, ok := answers[ip]
			if !ok {
				continue
			}
			if answer.Err != nil {
				unanswered = true
				if err := writer.WriteError(lib.KindAddress, ip, answer.Err); err != nil {
					return err
				}
				continue
			}
			if err := writer.Write(lib.KindAddress, ip, answer.Result); err != nil {
				return err
			}
		}
		chunk = chunk[:0]
		return nil
	}

	walkErr := iputil.WalkAddrs(args, opts.input, func(addr netip.Addr) error {
		chunk = append(chunk, addr.String())
		if len(chunk) < chunkSize {
			return nil
		}
		return flush()
	})
	if walkErr != nil {
		return walkErr
	}
	if err := flush(); err != nil {
		return err
	}

	for _, t := range targets {
		if writer == nil {
			openWriter(len(targets) == 1 && !opts.forceBulk)
		}
		answer, err := lookupTarget(ctx, client, t)
		if err != nil {
			unanswered = true
			if err := writer.WriteError(t.kind, t.key, err); err != nil {
				return err
			}
			continue
		}
		if err := writer.Write(t.kind, t.key, answer); err != nil {
			return err
		}
	}

	if writer == nil {
		return errNoInput
	}
	if err := writer.Close(); err != nil {
		return err
	}
	if unanswered {
		return errUnanswered
	}
	return nil
}

// lookupTarget answers one AS or domain.
func lookupTarget(ctx context.Context, client *Client, t target) (any, error) {
	if t.kind == lib.KindASN {
		return client.api.LookupASN(ctx, t.asn)
	}
	return client.api.LookupDomain(ctx, t.key)
}

// checkArgs refuses an argument that is no address, range, CIDR, AS or domain
// and names no readable file, and returns the AS numbers and domains, in the
// order typed. iputil skips such a token, which is right for a list file's
// header line but not for something typed: `internetdata 1.1.1.1 adresses`
// would answer for one address and exit 0 over the misspelled file.
func checkArgs(args []string) ([]target, error) {
	var targets []target
	for _, arg := range args {
		kind, ok := classify(arg)
		switch {
		case !ok:
			return nil, fmt.Errorf("%q is not an IP, CIDR, range, AS, domain or readable file", arg)
		case kind == lib.KindASN:
			in := iputil.Classify(arg)
			targets = append(targets, target{kind: kind, key: fmt.Sprintf("AS%d", in.ASN), asn: in.ASN})
		case kind == lib.KindDomain:
			targets = append(targets, target{kind: kind, key: arg})
		}
	}
	return targets, nil
}

// classify says what one typed argument asks about: addresses, for an address,
// a range, a CIDR or a file listing them, an AS, or a domain. ok is false for
// anything else.
//
// A file wins over a domain, as an address wins over a file: `addresses.txt`
// reads like a domain, and a list file is what it names when one is there.
func classify(arg string) (kind lib.Kind, ok bool) {
	switch in := iputil.Classify(arg); {
	case in.Kind == iputil.KindASN:
		return lib.KindASN, true
	case in.Kind != iputil.KindUnknown, iputil.FileExists(arg):
		return lib.KindAddress, true
	case isDomain(arg):
		return lib.KindDomain, true
	}
	return 0, false
}

// isDomain reports whether s is shaped as a domain: two labels or more of
// letters, digits and inner hyphens, and a top label that is not all digits,
// which no domain has and a mistyped address does (`8.8.8.888`). A letter
// beyond ASCII is taken as is, for an internationalized name: the API puts it in
// punycode, and refuses what it cannot read.
func isDomain(s string) bool {
	labels := strings.Split(strings.TrimSuffix(s, "."), ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, r := range label {
			if r < 0x80 && !isLDH(r) {
				return false
			}
		}
	}
	return strings.Trim(labels[len(labels)-1], "0123456789") != ""
}

// isLDH reports whether an ASCII rune may stand in a hostname's label.
func isLDH(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-'
}

// typedKinds are the kinds the arguments name, one each, in the order first
// typed.
func typedKinds(args []string) []lib.Kind {
	var kinds []lib.Kind
	seen := make(map[lib.Kind]bool)
	for _, arg := range args {
		if kind, ok := classify(arg); ok && !seen[kind] {
			kinds = append(kinds, kind)
			seen[kind] = true
		}
	}
	return kinds
}

// answerKinds are the kinds a lookup of args can answer: addresses always, as
// standard input may list them, and an AS or a domain where one is typed.
func answerKinds(args []string) []lib.Kind {
	kinds := []lib.Kind{lib.KindAddress}
	for _, kind := range typedKinds(args) {
		if kind != lib.KindAddress {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}

// errNoInput is returned when the command found nothing to look up.
var errNoInput = errors.New("no addresses in input")

// errNothingFound is what a command reports for errNoInput, naming what it reads.
var errNothingFound = errors.New("no addresses found in the input; expected an IP, CIDR, range or file, " +
	"or an AS or a domain as an argument")

// errUnanswered is returned once every answer is written, when at least one
// address, AS or domain got its error in place of an answer. It only sets the
// exit status, so that a refused key or a spent allowance fails a script rather
// than reading as an answer.
var errUnanswered = errors.New("not every address was answered")

// writeOne renders one address's answer in the chosen format.
func writeOne(ip string, result *internetdata.Result, opts lookupOpts) error {
	format := opts.format
	if format == "" {
		format = lib.Format(gConfig.Format)
		if format == "" {
			format = lib.FormatPretty
		}
	}
	w := lib.NewWriter(os.Stdout, format, lib.Opts{
		Fields:     opts.fields,
		ShowAbsent: opts.showAbsent,
		Color:      !color.NoColor,
	})
	if err := w.Write(lib.KindAddress, ip, result); err != nil {
		return err
	}
	return w.Close()
}

// formatFlags registers the output options the lookup commands share, and
// returns a function resolving them into a lookupOpts, checking a --field
// selection against the kinds of answer the command can print.
func formatFlags() func(kinds ...lib.Kind) (lookupOpts, error) {
	var (
		fFields     []string
		fFormat     string
		fJSON       bool
		fJSONL      bool
		fCSV        bool
		fYAML       bool
		fPretty     bool
		fShowAbsent bool
	)
	pflag.StringSliceVarP(&fFields, "field", "f", nil, "only these fields, comma separated.")
	pflag.StringVar(&fFormat, "format", "", "output format: pretty, json, jsonl, csv, yaml.")
	pflag.BoolVarP(&fJSON, "json", "j", false, "output JSON.")
	pflag.BoolVar(&fJSONL, "jsonl", false, "output one JSON object per line.")
	pflag.BoolVarP(&fCSV, "csv", "c", false, "output CSV.")
	pflag.BoolVarP(&fYAML, "yaml", "y", false, "output YAML.")
	pflag.BoolVarP(&fPretty, "pretty", "p", false, "output the readable block.")
	pflag.BoolVar(&fShowAbsent, "show-absent", false, "show fields this plan does not include.")

	return func(kinds ...lib.Kind) (lookupOpts, error) {
		opts := lookupOpts{
			fields:     fFields,
			showAbsent: fShowAbsent,
			input:      iputil.ListOpts,
		}
		// The shorthands are conveniences for --format, and giving two of them
		// is a contradiction rather than a precedence question.
		chosen := map[string]bool{}
		if fJSON {
			chosen["json"] = true
		}
		if fJSONL {
			chosen["jsonl"] = true
		}
		if fCSV {
			chosen["csv"] = true
		}
		if fYAML {
			chosen["yaml"] = true
		}
		if fPretty {
			chosen["pretty"] = true
		}
		if fFormat != "" {
			chosen[fFormat] = true
		}
		if len(chosen) > 1 {
			return opts, fmt.Errorf("pick one output format, not %d", len(chosen))
		}
		for name := range chosen {
			format, err := lib.ParseFormat(name)
			if err != nil {
				return opts, err
			}
			opts.format = format
		}
		if err := lib.ValidateFields(opts.fields, kinds...); err != nil {
			return opts, err
		}
		return opts, nil
	}
}
