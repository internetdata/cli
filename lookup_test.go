package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	iputil "github.com/mslmio/libgo-iputil"
	"github.com/spf13/pflag"

	"github.com/internetdata/cli/lib"
)

// Chunking to the batch endpoint's 1000 belongs to the SDK, so the bulk path
// must hand it everything it read rather than capping or splitting on its own:
// 2500 addresses are three requests to POST /batch and one answer each, with a
// key, which the batch needs.
func TestBulkIsNotCappedAtTheBatchEndpointsThousand(t *testing.T) {
	var batches atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/batch" {
			t.Errorf("requested %s %s, want only POST /batch", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		batches.Add(1)
		var in struct {
			IPs []string `json:"ips"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"unreadable body"}`, http.StatusBadRequest)
			return
		}
		results := make(map[string]any, len(in.IPs))
		for _, ip := range in.IPs {
			results[ip] = answerFor(ip)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"results": results, "errors": map[string]any{}})
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "test-key")

	addrs := make([]string, 2500)
	for i := range addrs {
		addrs[i] = fmt.Sprintf("9.1.%d.%d", i/256, i%256)
	}
	out, err := captureStdout(t, func() error {
		return runLookup(t.Context(), addrs, lookupOpts{format: lib.FormatJSONL, forceBulk: true})
	})
	if err != nil {
		t.Fatal(err)
	}

	if n := batches.Load(); n != 3 {
		t.Errorf("sent %d POST /batch request(s), want 3 for %d addresses", n, len(addrs))
	}
	answered := map[string]bool{}
	lines := bufio.NewScanner(strings.NewReader(out))
	for lines.Scan() {
		var answer struct {
			IP    string `json:"ip"`
			Error string `json:"error"`
		}
		if err := json.Unmarshal(lines.Bytes(), &answer); err != nil {
			t.Fatalf("unreadable output line %q: %v", lines.Text(), err)
		}
		if answer.Error != "" {
			t.Errorf("%s: %s", answer.IP, answer.Error)
		}
		if answered[answer.IP] {
			t.Errorf("%s was answered twice", answer.IP)
		}
		answered[answer.IP] = true
	}
	for _, ip := range addrs {
		if !answered[ip] {
			t.Fatalf("%s has no answer; %d of %d were answered", ip, len(answered), len(addrs))
		}
	}
	if len(answered) != len(addrs) {
		t.Errorf("answered %d address(es), want %d", len(answered), len(addrs))
	}
}

// POST /batch needs a key, so without one each address is its own GET, answered
// all the same.
func TestLookupWithoutAKeyIsOneGetPerAddress(t *testing.T) {
	var gets atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "" {
			t.Errorf("requested %s %s with %q, want GET without a key", r.Method, r.URL.Path,
				r.Header.Get("Authorization"))
			http.NotFound(w, r)
			return
		}
		gets.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(answerFor(strings.TrimPrefix(r.URL.Path, "/")))
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "")

	addrs := make([]string, 25)
	for i := range addrs {
		addrs[i] = fmt.Sprintf("9.1.0.%d", i+1)
	}
	out, err := captureStdout(t, func() error {
		return runLookup(t.Context(), addrs, lookupOpts{format: lib.FormatJSONL})
	})
	if err != nil {
		t.Fatal(err)
	}
	if n := int(gets.Load()); n != len(addrs) {
		t.Errorf("sent %d GET(s), want one for each of %d addresses", n, len(addrs))
	}
	lines := 0
	for scanner := bufio.NewScanner(strings.NewReader(out)); scanner.Scan(); lines++ {
		if strings.Contains(scanner.Text(), `"error"`) {
			t.Errorf("unanswered: %s", scanner.Text())
		}
	}
	if lines != len(addrs) {
		t.Errorf("wrote %d answer(s), want %d", lines, len(addrs))
	}
}

// A key on the free plan may not batch more than one address: the API answers 402 for
// each, and the CLI looks each up on its own instead, answering every one.
func TestAFreePlanKeysRefusedBatchIsLookedUpOneByOne(t *testing.T) {
	var batches, gets atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			batches.Add(1)
			w.WriteHeader(http.StatusPaymentRequired)
			_, _ = fmt.Fprint(w, `{"error":"POST /batch needs a paid plan"}`)
			return
		}
		gets.Add(1)
		answer := answerFor(strings.TrimPrefix(r.URL.Path, "/"))
		answer["trial_expired"] = true
		_ = json.NewEncoder(w).Encode(answer)
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "free-key")

	addrs := []string{"9.1.0.1", "9.1.0.2", "9.1.0.3"}
	out, err := captureStdout(t, func() error {
		return runLookup(t.Context(), addrs, lookupOpts{format: lib.FormatJSONL})
	})
	if err != nil {
		t.Fatal(err)
	}
	if b, g := batches.Load(), gets.Load(); b != 1 || int(g) != len(addrs) {
		t.Errorf("sent %d batch(es) and %d GET(s), want 1 refused batch, then one GET for each of %d",
			b, g, len(addrs))
	}
	lines := 0
	for scanner := bufio.NewScanner(strings.NewReader(out)); scanner.Scan(); lines++ {
		if strings.Contains(scanner.Text(), `"error"`) {
			t.Errorf("unanswered: %s", scanner.Text())
		}
	}
	if lines != len(addrs) {
		t.Errorf("wrote %d answer(s), want %d", lines, len(addrs))
	}
}

// A lookup exits 1 when any address went unanswered, so a refused key or a
// spent allowance fails a script, and the output stays what it was: every
// answer, and each failure in place of its address's answer.
func TestLookupReportsAnUnansweredAddress(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"results": {"1.1.1.1": `+answerJSON("1.1.1.1")+`},`+
			` "errors": {"2.2.2.2": {"status": 429, "error": "quota exceeded"}}}`)
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "test-key")

	out, err := captureStdout(t, func() error {
		return runLookup(t.Context(), []string{"1.1.1.1", "2.2.2.2"}, lookupOpts{format: lib.FormatJSONL})
	})
	if err != errUnanswered {
		t.Errorf("runLookup = %v, want errUnanswered", err)
	}
	want := answerJSON("1.1.1.1") + "\n" +
		`{"error":"internetdata: quota_exceeded (HTTP 429): quota exceeded","ip":"2.2.2.2"}` + "\n"
	if out != want {
		t.Errorf("wrote\n%s\nwant\n%s", out, want)
	}
}

// A list file that starts with a byte order mark, as Excel's "CSV UTF-8" export
// writes one, loses none of its addresses: the first is not the mark glued to it.
func TestLookupReadsAListAfterAByteOrderMark(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			IPs []string `json:"ips"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"unreadable body"}`, http.StatusBadRequest)
			return
		}
		results := make(map[string]any, len(in.IPs))
		for _, ip := range in.IPs {
			results[ip] = answerFor(ip)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"results": results, "errors": map[string]any{}})
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "test-key")

	file := filepath.Join(t.TempDir(), "ips.csv")
	if err := os.WriteFile(file, []byte("\xef\xbb\xbf1.1.1.1\r\n8.8.8.8\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := captureStdout(t, func() error {
		return runLookup(t.Context(), []string{file},
			lookupOpts{format: lib.FormatCSV, fields: []string{"ip"}, input: iputil.Opts{Files: true}})
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := "ip\n1.1.1.1\n8.8.8.8\n"; out != want {
		t.Errorf("wrote\n%s\nwant\n%s", out, want)
	}
}

// An AS and a domain typed as arguments are each one request of their own to
// the SDK's paths, never cached, never batched, after the addresses, in the
// order typed; an address beside them keeps its cache.
func TestAnASAndADomainAreEachTheirOwnRequest(t *testing.T) {
	var mu sync.Mutex
	var requests []string
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.Method+" "+r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, "/asn/"):
			_, _ = fmt.Fprintf(w, `{"asn":%s,"name":"N","country":"US","registry":"arin",`+
				`"allocated":"2000-03-30","domain":"n.example"}`, strings.TrimPrefix(r.URL.Path, "/asn/"))
		case strings.HasPrefix(r.URL.Path, "/domain/"):
			_, _ = fmt.Fprintf(w, `{"domain":%q,"is_registered":true,"has_a":true,"has_aaaa":false,`+
				`"has_mx":false,"has_ns":true,"has_cname":false}`, strings.TrimPrefix(r.URL.Path, "/domain/"))
		default:
			_ = json.NewEncoder(w).Encode(answerFor(strings.TrimPrefix(r.URL.Path, "/")))
		}
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "")
	gConfig.CacheEnabled = true

	args := []string{"as15169", "Example.com", "1.1.1.1", "AS9009"}
	var outs []string
	for range 2 {
		out, err := captureStdout(t, func() error {
			return runLookup(t.Context(), args, lookupOpts{format: lib.FormatJSONL})
		})
		if err != nil {
			t.Fatal(err)
		}
		outs = append(outs, out)
	}

	want := []string{
		"GET /1.1.1.1", "GET /asn/15169", "GET /domain/Example.com", "GET /asn/9009",
		"GET /asn/15169", "GET /domain/Example.com", "GET /asn/9009",
	}
	if strings.Join(requests, "\n") != strings.Join(want, "\n") {
		t.Errorf("requested\n%s\nwant\n%s", strings.Join(requests, "\n"), strings.Join(want, "\n"))
	}
	for _, out := range outs {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) != 4 || !strings.Contains(lines[0], `"ip":"1.1.1.1"`) ||
			!strings.HasPrefix(lines[1], `{"allocated":"2000-03-30","asn":15169,`) ||
			!strings.HasPrefix(lines[2], `{"domain":"Example.com",`) ||
			!strings.HasPrefix(lines[3], `{"allocated":"2000-03-30","asn":9009,`) {
			t.Errorf("wrote\n%s\nwant the address, then the two AS numbers and the domain as typed", out)
		}
	}
}

// One AS or one domain is one answer, so it prints the readable block, as one
// address does; beside an address it is a second answer, so JSON, keyed by
// what was looked up. bulk is always the machine format.
func TestAnASOrADomainCountsAsAnAnswer(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, "/asn/"):
			_, _ = fmt.Fprint(w, `{"asn":15169,"name":"Google LLC","country":"US","registry":"arin",`+
				`"allocated":"2000-03-30","domain":"google.com"}`)
		default:
			_ = json.NewEncoder(w).Encode(answerFor(strings.TrimPrefix(r.URL.Path, "/")))
		}
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "")

	alone, err := captureStdout(t, func() error {
		return runLookup(t.Context(), []string{"AS15169"}, lookupOpts{})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(alone, "asn        15169\nname       Google LLC\n") {
		t.Errorf("one AS printed\n%s\nwant the readable block", alone)
	}

	both, err := captureStdout(t, func() error {
		return runLookup(t.Context(), []string{"9.1.0.1", "as15169"}, lookupOpts{})
	})
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]map[string]any
	if err := json.Unmarshal([]byte(both), &parsed); err != nil || parsed["AS15169"]["name"] != "Google LLC" ||
		parsed["9.1.0.1"]["ip"] != "9.1.0.1" {
		t.Errorf("an address and an AS printed\n%s\nwant JSON keyed by each, the AS as AS<n> (%v)", both, err)
	}

	bulk, err := captureStdout(t, func() error {
		return runLookup(t.Context(), []string{"AS15169"}, lookupOpts{forceBulk: true})
	})
	if err != nil || !strings.HasPrefix(bulk, "{\n  \"AS15169\": {") {
		t.Errorf("bulk of one AS printed\n%s\nerr %v; want JSON", bulk, err)
	}
}

// An AS or a domain the API refuses gets its error in its place, and the run
// exits 1, as an address's does.
func TestARefusedDomainIsUnanswered(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, `{"error":"not a valid IP address, AS number or domain"}`)
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "")

	out, err := captureStdout(t, func() error {
		return runLookup(t.Context(), []string{"xn--a.example", "AS1"}, lookupOpts{format: lib.FormatJSONL})
	})
	if err != errUnanswered {
		t.Errorf("runLookup = %v, want errUnanswered", err)
	}
	want := `{"domain":"xn--a.example","error":"internetdata: bad_request (HTTP 400): ` +
		`not a valid IP address, AS number or domain"}` + "\n" +
		`{"asn":1,"error":"internetdata: bad_request (HTTP 400): ` +
		`not a valid IP address, AS number or domain"}` + "\n"
	if out != want {
		t.Errorf("wrote\n%s\nwant\n%s", out, want)
	}
}

// A list, piped or in a file, is read for addresses alone: an AS or a domain on
// one of its lines is skipped, as a header line is, and never looked up.
func TestAListIsReadForAddressesAlone(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(answerFor(strings.TrimPrefix(r.URL.Path, "/")))
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "")

	file := filepath.Join(t.TempDir(), "list.txt")
	if err := os.WriteFile(file, []byte("target\nAS15169\nexample.com\n9.1.0.1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := captureStdout(t, func() error {
		return runLookup(t.Context(), []string{file},
			lookupOpts{format: lib.FormatJSONL, input: iputil.Opts{Files: true}})
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(paths, " ") != "/9.1.0.1" {
		t.Errorf("requested %v, want the list's address alone", paths)
	}
}

// One CSV holds one kind of answer, so typing an address beside an AS asks for
// two, and is refused before anything is looked up.
func TestCSVOfTwoKindsIsRefusedFirst(t *testing.T) {
	var requests atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.NotFound(w, r)
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "")

	for _, args := range [][]string{{"9.1.0.1", "AS15169"}, {"AS15169", "example.com"}} {
		_, err := captureStdout(t, func() error {
			return runLookup(t.Context(), args, lookupOpts{format: lib.FormatCSV})
		})
		if !errors.Is(err, lib.ErrMixedCSV) {
			t.Errorf("%v as CSV: %v, want ErrMixedCSV", args, err)
		}
	}
	if n := requests.Load(); n != 0 {
		t.Errorf("sent %d request(s) before refusing", n)
	}
}

// A typed argument is an address, a range, a CIDR, an AS, a domain or a file that exists: anything
// else is named back rather than skipped, as a list file's unparseable lines are.
func TestCheckArgs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "addresses.txt")
	if err := os.WriteFile(file, []byte("ip\n8.8.8.8\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		arg  string
		want string
		kind lib.Kind
	}{
		{"8.8.8.8", "", lib.KindAddress},
		{"2001:db8::1", "", lib.KindAddress},
		{"10.0.0.0/30", "", lib.KindAddress},
		{"1.1.1.1-1.1.1.3", "", lib.KindAddress},
		{file, "", lib.KindAddress},
		{"AS15169", "", lib.KindASN},
		{"as15169", "", lib.KindASN},
		{"example.com", "", lib.KindDomain},
		{"Bücher.example", "", lib.KindDomain},
		{"example.com.", "", lib.KindDomain},
		{"adresses", "not an IP, CIDR, range, AS, domain or readable file", 0},
		{"not-an-ip", "not an IP, CIDR, range, AS, domain or readable file", 0},
		{"8.8.8.888", "not an IP, CIDR, range, AS, domain or readable file", 0},
		{"https://example.com", "not an IP, CIDR, range, AS, domain or readable file", 0},
		{"-example.com", "not an IP, CIDR, range, AS, domain or readable file", 0},
		{"example..com", "not an IP, CIDR, range, AS, domain or readable file", 0},
		{"AS-1", "not an IP, CIDR, range, AS, domain or readable file", 0},
	}
	for _, c := range cases {
		targets, err := checkArgs([]string{"8.8.8.8", c.arg})
		if c.want == "" && err != nil {
			t.Errorf("%q: %v, want no error", c.arg, err)
		}
		if c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)) {
			t.Errorf("%q: %v, want an error containing %q", c.arg, err, c.want)
		}
		if c.want != "" {
			continue
		}
		if kind, _ := classify(c.arg); kind != c.kind {
			t.Errorf("%q is kind %d, want %d", c.arg, kind, c.kind)
		}
		if wantTargets := c.kind != lib.KindAddress; wantTargets != (len(targets) == 1) {
			t.Errorf("%q gave targets %v", c.arg, targets)
		}
	}

	// A file named as a domain is the file: a list is what it names when one is there.
	t.Chdir(dir)
	if err := os.WriteFile("list.example", []byte("8.8.8.8\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if kind, ok := classify("list.example"); !ok || kind != lib.KindAddress {
		t.Errorf("a file named list.example is kind %d, want a list of addresses", kind)
	}
}

// Without a key each address is its own lookup, and as many are in flight as
// --concurrency allows: no more, and no fewer, or a bulk run is as slow as one
// address at a time.
func TestLookupsWithoutAKeyRunAtTheConcurrency(t *testing.T) {
	const width = 4
	var inFlight, peak atomic.Int32
	var opened sync.Once
	release := make(chan struct{})
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := inFlight.Add(1)
		defer inFlight.Add(-1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		if n == width {
			opened.Do(func() { close(release) })
		}
		select {
		case <-release:
		case <-time.After(2 * time.Second):
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(answerFor(strings.TrimPrefix(r.URL.Path, "/")))
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "")
	savedConcurrency := fConcurrency
	t.Cleanup(func() { fConcurrency = savedConcurrency })
	fConcurrency = width

	addrs := make([]string, 3*width)
	for i := range addrs {
		addrs[i] = fmt.Sprintf("9.1.0.%d", i+1)
	}
	if _, err := captureStdout(t, func() error {
		return runLookup(t.Context(), addrs, lookupOpts{format: lib.FormatJSONL})
	}); err != nil {
		t.Fatal(err)
	}
	if p := peak.Load(); p != width {
		t.Errorf("at most %d lookup(s) were in flight, want %d", p, width)
	}
}

// A --field is checked against what the arguments ask about: an AS's field is
// no typo beside an AS, and is one beside addresses alone, which standard input
// may add to either.
func TestFieldsAreCheckedAgainstWhatIsTyped(t *testing.T) {
	for _, c := range []struct {
		args []string
		want []lib.Kind
	}{
		{nil, []lib.Kind{lib.KindAddress}},
		{[]string{"1.1.1.1", "10.0.0.0/30"}, []lib.Kind{lib.KindAddress}},
		{[]string{"AS15169"}, []lib.Kind{lib.KindAddress, lib.KindASN}},
		{[]string{"example.com", "1.1.1.1", "AS1"}, []lib.Kind{lib.KindAddress, lib.KindDomain, lib.KindASN}},
	} {
		if got := answerKinds(c.args); !reflect.DeepEqual(got, c.want) {
			t.Errorf("answerKinds(%v) = %v, want %v", c.args, got, c.want)
		}
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	savedArgs, savedFlags, savedConfig := os.Args, pflag.CommandLine, gConfig
	t.Cleanup(func() { os.Args, pflag.CommandLine, gConfig = savedArgs, savedFlags, savedConfig })
	for _, c := range []struct {
		cmd  func() error
		args []string
	}{
		{cmdDefault, []string{"internetdata", "1.1.1.1", "-f", "registry"}},
		{cmdBulk, []string{"internetdata", "bulk", "1.1.1.1", "-f", "registry"}},
		{cmdMyIP, []string{"internetdata", "myip", "-f", "registry"}},
	} {
		os.Args = c.args
		pflag.CommandLine = pflag.NewFlagSet("internetdata", pflag.ContinueOnError)
		gConfig = NewConfig()
		_, err := captureStdout(t, c.cmd)
		if err == nil || !strings.Contains(err.Error(), "unknown field(s): registry") {
			t.Errorf("%v: %v, want registry refused as no address's field", c.args, err)
		}
	}
}

// Login checks a key with a lookup, which any key may make, so a key without
// the db.download scope is stored, and a key the API refuses is not.
func TestLoginChecksAKeyWithALookup(t *testing.T) {
	var mu sync.Mutex
	var seen []string
	var refuse atomic.Bool
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.URL.Path+" "+r.Header.Get("Authorization"))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if refuse.Load() {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprint(w, `{"error":"invalid API key"}`)
			return
		}
		_ = json.NewEncoder(w).Encode(answerFor(strings.TrimPrefix(r.URL.Path, "/")))
	}))
	t.Cleanup(api.Close)
	useAPI(t, api.URL, "")

	if err := checkKey("k_lookups_only", api.URL); err != nil {
		t.Errorf("a key the lookup took was refused: %v", err)
	}
	refuse.Store(true)
	if err := checkKey("k_refused", api.URL); err == nil || !strings.Contains(err.Error(), "was not accepted") {
		t.Errorf("a key the lookup refused: %v, want it refused", err)
	}
	want := []string{"/1.1.1.1 Bearer k_lookups_only", "/1.1.1.1 Bearer k_refused"}
	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(seen, want) {
		t.Errorf("requested %v, want %v", seen, want)
	}
}

// useAPI points the next client at api, with key, or none, and nothing cached.
func useAPI(t *testing.T, api, key string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("INTERNETDATA_API_KEY", key)
	t.Setenv("INTERNETDATA_SESSION", "")
	savedConfig, savedBaseURL, savedNoCache := gConfig, fBaseURL, fNoCache
	t.Cleanup(func() { gConfig, fBaseURL, fNoCache = savedConfig, savedBaseURL, savedNoCache })
	gConfig, fBaseURL, fNoCache = NewConfig(), api, false
	gConfig.CacheEnabled = false
}

// answerFor is Unlimited's answer about ip: every member the API requires, which
// the SDK refuses an answer without.
func answerFor(ip string) map[string]any {
	var answer map[string]any
	if err := json.Unmarshal([]byte(answerJSON(ip)), &answer); err != nil {
		panic(err)
	}
	return answer
}

// answerJSON is answerFor as the API serves it, and as the CLI writes it back.
func answerJSON(ip string) string {
	return `{"abuse":{"email":""},"asn":{"asn":0,"domain":"","name":"","prefix":"","type":""},` +
		`"company":{"domain":"","name":""},"hostname":"","ip":"` + ip + `","is_anycast":false,` +
		`"is_bittorrent":false,"is_bogon":false,"is_carrier":false,"is_dns_resolver":false,"is_relay":false,` +
		`"is_satellite":false,"is_smtp":false,"is_ssh":false,"is_tor":false,"is_web":false,` +
		`"location":{"continent":"","continent_name":"","country":"","country_name":"","currency_code":"",` +
		`"currency_name":"","currency_symbol":"","flag_emoji":"","flag_unicode":"","timezone":""},` +
		`"relay":{},"tor":{}}`
}
