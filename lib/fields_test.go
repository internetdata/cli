package lib

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	internetdata "github.com/internetdata/sdk-go/v2"
)

// unlimitedAnswer is what a lookup without a key serves: the common set and
// InternetData's own flags, and every paid field ABSENT rather than false.
const unlimitedAnswer = `{
  "ip":"146.70.116.130","is_bogon":false,"is_tor":false,"is_relay":false,"is_anycast":false,
  "is_carrier":false,"is_satellite":false,"is_ssh":false,"is_smtp":false,"is_web":false,
  "is_dns_resolver":false,"is_bittorrent":false,"tor":{},"relay":{},
  "location":{"continent":"EU","continent_name":"Europe","country":"AT","country_name":"Austria",
    "timezone":"Europe/Vienna","currency_code":"EUR","currency_name":"Euro","currency_symbol":"€",
    "flag_emoji":"🇦🇹","flag_unicode":"U+1F1E6 U+1F1F9"},
  "asn":{"asn":9009,"name":"M247 Europe SRL","domain":"m247global.com","type":"hosting",
    "prefix":"146.70.116.0/24"},
  "company":{"name":"M247 LTD Vienna Infrastructure","domain":"m247.ro"},
  "hostname":"","abuse":{"email":"abuse@m247.ro"}
}`

// paidAnswer is the full shape a paid plan serves, with detail objects that are
// present but EMPTY - which means the flag above each is false, not that it is
// missing - and Web IP's list empty under its false flag.
const paidAnswer = `{
  "ip":"146.70.116.130","is_bogon":false,"is_vpn":true,"is_tor":false,"is_relay":false,
  "is_resproxy":false,"is_dcproxy":false,"is_mobproxy":false,"is_pubproxy":false,"is_hosting":true,
  "is_cdn":false,"is_anycast":false,"is_carrier":false,"is_satellite":false,"is_ssh":false,
  "is_smtp":false,"is_web":false,"is_dns_resolver":false,"is_bittorrent":false,
  "vpn":{"provider":"mullvad","provider_name":"Mullvad","confidence":"high","last_seen":"2026-11-13"},
  "tor":{},"relay":{},"resproxy":{},"dcproxy":{},"mobproxy":{},"pubproxy":{},
  "hosting":{"provider":"m247","provider_name":"M247","confidence":"high","last_seen":"2026-11-13"},
  "cdn":{},"carrier":{},"satellite":{},"ssh":{},"smtp":{},"web":[],"dns_resolver":{},"bittorrent":{},
  "location":{"continent":"EU","continent_name":"Europe","country":"AT","country_name":"Austria",
    "timezone":"Europe/Vienna","currency_code":"EUR","currency_name":"Euro","currency_symbol":"€",
    "flag_emoji":"🇦🇹","flag_unicode":"U+1F1E6 U+1F1F9","region":"Vienna","region_code":"AT-9",
    "city":"Vienna","postal_code":"1230","latitude":48.20849,"longitude":16.37208,"accuracy_radius":10,
    "geoname_id":"2761369","dma_code":"","last_changed":"2026-10-06"},
  "asn":{"asn":9009,"name":"M247 Europe SRL","domain":"m247global.com","type":"hosting",
    "prefix":"146.70.116.0/24","country":"RO","rpki_status":"not_found"},
  "company":{"name":"M247 LTD Vienna Infrastructure","domain":"m247.ro","type":"hosting","country":"AT"},
  "hostname":"",
  "abuse":{"email":"abuse@m247.ro","name":"M247 Europe","phone":"",
    "address":"Sos. Fabrica de Glucoza, Nr 11B, etaj 1, Sector 2, Bucuresti Romania","country":"AT",
    "network":"146.70.116.0-146.70.116.255","validated":false,"unvalidated_since":""},
  "whois":{"id":"M247-LTD-Vienna","name":"M247 Vienna","org":"ORG-MES59-RIPE","country":"AT",
    "status":"LEGACY","source":"ripe","domain":"m247.ro","abuse_email":"abuse@m247.ro","asn":9009},
  "rwhois":{"id":"","name":"","country":"","source":"","domain":"","abuse_email":""},
  "hosted_domains":{"total":0}
}`

// proxyAnswer carries three proxy families' detail objects populated, as the API
// serves a populated one: with every key.
const proxyAnswer = `{
  "ip":"217.80.249.248","is_bogon":false,"is_resproxy":true,"is_dcproxy":false,"is_mobproxy":true,
  "is_pubproxy":true,
  "resproxy":{"provider":"","provider_name":"","first_seen":"2026-07-16","last_seen":"2026-10-12",
    "hits":120,"hits_days_pct":19,"providers_num":18},
  "dcproxy":{},
  "mobproxy":{"provider":"p","provider_name":"P","first_seen":"2026-09-30","last_seen":"2026-10-11",
    "hits":3,"hits_days_pct":2,"providers_num":1},
  "pubproxy":{"entry":false,"exit":true,"protocols":["socks5"],"num_ports":0,"num_proxies":1,
    "last_seen":"2026-10-17","last_usable":"2026-10-17"}
}`

// asnUnlimitedAnswer is an AS's answer without a key: the Open ASN row alone.
const asnUnlimitedAnswer = `{"asn":9009,"name":"M247 Europe SRL","country":"RO","registry":"ripe",` +
	`"allocated":"2005-06-06","domain":"m247global.com"}`

// asnPaidAnswer is an AS's answer on a paid plan: the counts, the whois record
// keyed as the API documents it, and the prefixes.
const asnPaidAnswer = `{
  "asn":9009,"name":"M247 Europe SRL","country":"RO","registry":"ripe","allocated":"2005-06-06",
  "domain":"m247global.com","type":"hosting","num_ips":1615872,"num_prefixes4":4510,"num_prefixes6":807,
  "rpki_valid":2972,"rpki_invalid":50,"rpki_not_found":2295,
  "whois":{"abuse":"","abuse_city":"","abuse_country":"","abuse_email":"abuse@m247.ro","abuse_kind":"role",
    "abuse_name":"M247 Europe","abuse_phone":"","abuse_unvalidated_since":"","abuse_validated":"",
    "admin":"","admin_city":"","admin_country":"","admin_email":"","admin_kind":"","admin_name":"",
    "admin_phone":"","country":"RO","created":"2011-04-11","descr":"","handle":"M247",
    "maintainer":"RIPE-NCC-END-MNT|M247-EU-MNT","maintainer_email":"","name":"M247 Europe SRL",
    "org":"ORG-MES6-RIPE",
    "org_address":"11B Sos. Fabrica de Glucoza, 1st floor District 2, 020331, Bucharest, ROMANIA",
    "org_city":"","org_country":"RO","org_name":"M247 Europe SRL","remarks":"","source":"ripe",
    "status":"ASSIGNED","tech":"","tech_city":"","tech_country":"","tech_email":"","tech_kind":"",
    "tech_name":"","tech_phone":"","updated":"2023-03-15"},
  "prefixes":[{"prefix":"2.56.16.0/22","peer_count":17,"rpki_status":"not_found","first_seen":"2026-10-02",
    "country":"VN","size":1024,"other_origins":[]}]
}`

// domainPaidAnswer is a domain's answer on a paid plan: its records, its
// organization's ranges, and an empty disposable object under a false flag.
const domainPaidAnswer = `{
  "domain":"example.com","is_registered":true,"is_disposable":false,"has_a":true,"has_aaaa":true,
  "has_mx":true,"has_ns":true,"has_cname":false,
  "dns":{"a":[{"ip":"192.0.2.10","last_seen":"2026-11-08"}],
    "aaaa":[{"ip":"2001:db8::10","last_seen":"2026-11-08"}],
    "mx":[{"priority":10,"host":"mail.example.com","last_seen":"2026-11-08"}],
    "ns":[{"host":"ns1.example.com","last_seen":"2026-11-08"}],"cname":[]},
  "ip_ranges":{"domain":"example.com","redirects_to":"","total":1,"ipv4":256,
    "ranges":[{"range":"192.0.2.0/24","name":"EXAMPLE-NET","country":"US"}]},
  "disposable":{}
}`

func flattenJSON(t *testing.T, body string) Record {
	t.Helper()
	var v any
	if err := jsonUnmarshal(body, &v); err != nil {
		t.Fatal(err)
	}
	rec, err := Flatten(v)
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

// Absent and present-false are different answers and must stay different all
// the way to the output. This is the single subtlest thing about this product.
func TestAbsentIsNotFalse(t *testing.T) {
	free := flattenJSON(t, unlimitedAnswer)
	paid := flattenJSON(t, paidAnswer)

	// Unlimited: is_vpn was never served.
	if c := free.Get("is_vpn"); c.Present {
		t.Error("unlimited: is_vpn should be absent")
	}
	// Paid: is_vpn was served, and is true.
	if c := paid.Get("is_vpn"); !c.Present || c.Text != "true" {
		t.Errorf("paid: is_vpn = %+v", c)
	}
	// Paid: is_cdn was served, and is false. Not the same as Unlimited's absence.
	if c := paid.Get("is_cdn"); !c.Present || c.Text != "false" {
		t.Errorf("paid: is_cdn = %+v", c)
	}
	// So inside a shared object: Unlimited's location has no city.
	if c := free.Get("location.city"); c.Present {
		t.Errorf("unlimited: location.city should be absent, got %+v", c)
	}
}

// An empty detail object is PRESENT and means its flag is false, which is not
// the answer a plan without that object gives.
func TestEmptyObjectIsPresentNotAbsent(t *testing.T) {
	paid := flattenJSON(t, paidAnswer)
	if c := paid.Get("resproxy"); !c.Present || c.Text != "" {
		t.Errorf("resproxy should be present and empty, got %+v", c)
	}
	// The Unlimited answer has no resproxy object at all.
	free := flattenJSON(t, unlimitedAnswer)
	if c := free.Get("resproxy"); c.Present {
		t.Errorf("unlimited: resproxy should be absent, got %+v", c)
	}
}

// Present() orders by meaning, not alphabetically: the address, then the flags,
// then each detail object with its own fields, then where the address is, and
// a top-level field after the objects.
func TestPresentOrder(t *testing.T) {
	got := flattenJSON(t, `{
  "hostname":"","web":[],"location":{"country":"AT"},
  "hosting":{"provider":"m247","provider_name":"M247","confidence":"high","last_seen":"2026-11-13"},
  "tor":{},
  "vpn":{"provider":"mullvad","provider_name":"Mullvad","confidence":"high","last_seen":"2026-11-13"},
  "is_hosting":true,"is_tor":false,"is_vpn":true,"is_bogon":false,"ip":"146.70.116.130"
}`).Present(KindAddress)
	want := []string{
		"ip", "is_bogon",
		"is_vpn", "is_tor", "is_hosting",
		"vpn.provider", "vpn.provider_name", "vpn.confidence", "vpn.last_seen",
		"tor",
		"hosting.provider", "hosting.provider_name", "hosting.confidence", "hosting.last_seen",
		"location.country",
		"web", "hostname",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

// The location, the AS, the company and the abuse contact print after the
// detail objects, each with its own fields in Columns' order, and the hostname,
// a top-level field, after them.
func TestPresentListsLocationAndAsnAfterTheDetails(t *testing.T) {
	got := flattenJSON(t, unlimitedAnswer).Present(KindAddress)
	want := []string{
		"ip", "is_bogon", "is_tor", "is_relay", "is_anycast", "is_carrier", "is_satellite",
		"is_ssh", "is_smtp", "is_web", "is_dns_resolver", "is_bittorrent", "tor", "relay",
		"location.continent", "location.continent_name", "location.country", "location.country_name",
		"location.timezone", "location.currency_code", "location.currency_name", "location.currency_symbol",
		"location.flag_emoji", "location.flag_unicode",
		"asn.asn", "asn.name", "asn.domain", "asn.type", "asn.prefix",
		"company.name", "company.domain", "abuse.email", "hostname",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

// --show-absent lists every field of the answer's kind, each in its place.
func TestAllListsEveryColumn(t *testing.T) {
	for _, c := range []struct {
		kind    Kind
		body    string
		columns []string
	}{
		{KindAddress, `{"ip":"1.1.1.1","is_bogon":false}`, Columns},
		{KindASN, `{"asn":15169}`, ASNColumns},
		{KindDomain, `{"domain":"example.com"}`, DomainColumns},
	} {
		got := flattenJSON(t, c.body).All(c.kind)
		if len(got) != len(c.columns) {
			t.Errorf("%s: All lists %d fields, want every column (%d): %v", c.body, len(got), len(c.columns), got)
		}
		for _, col := range c.columns {
			if !slices.Contains(got, col) {
				t.Errorf("%s: All leaves out %s", c.body, col)
			}
		}
	}
}

// An AS lists its registration, then its counts, then its whois record, then
// its prefixes; a domain its flags, then its records, ranges and disposable
// detail. Each in its own order, never in an address's.
func TestPresentListsAnASAndADomainInTheirOwnOrder(t *testing.T) {
	gotASN := flattenJSON(t, asnPaidAnswer).Present(KindASN)
	if !reflect.DeepEqual(gotASN, ASNColumns[:len(ASNColumns)-1]) {
		t.Errorf("AS: got %v\nwant %v", gotASN, ASNColumns[:len(ASNColumns)-1])
	}
	gotDomain := flattenJSON(t, domainPaidAnswer).Present(KindDomain)
	records := DomainColumns[:slices.Index(DomainColumns, "disposable.confidence")]
	wantDomain := append(append([]string{}, records...), "disposable")
	if !reflect.DeepEqual(gotDomain, wantDomain) {
		t.Errorf("domain: got %v\nwant %v", gotDomain, wantDomain)
	}
	free := flattenJSON(t, asnUnlimitedAnswer).Present(KindASN)
	if want := ASNColumns[:6]; !reflect.DeepEqual(free, want) {
		t.Errorf("AS without a key: got %v, want %v", free, want)
	}
}

// A field the API grows after this build must still be printed, or the CLI
// silently hides what the customer is paying for.
func TestPresentKeepsUnknownFields(t *testing.T) {
	for kind, body := range map[Kind]string{
		KindAddress: `{"ip":"1.1.1.1","is_vpn":false,"is_something_new":true}`,
		KindASN:     `{"asn":1,"whois":{"handle":"X","something_new":"y"}}`,
		KindDomain:  `{"domain":"example.com","something_new":true}`,
	} {
		found := false
		for _, f := range flattenJSON(t, body).Present(kind) {
			if strings.HasSuffix(f, "something_new") {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: an unrecognized field was dropped", body)
		}
	}
}

func TestScalarRendering(t *testing.T) {
	rec := flattenJSON(t, `{"ip":"1.1.1.1","is_vpn":false,"resproxy":{"provider":"p","hits":42}}`)
	if c := rec.Get("is_vpn"); c.Text != "false" {
		t.Errorf("bool: %q", c.Text)
	}
	// JSON has one number type; a hit count must not render as 42.000000.
	if c := rec.Get("resproxy.hits"); c.Text != "42" {
		t.Errorf("number: %q", c.Text)
	}
	// A list prints as its JSON, so an empty one is not an absent field's blank.
	rec = flattenJSON(t, `{"pubproxy":{"protocols":["socks5","http"]}}`)
	if c := rec.Get("pubproxy.protocols"); c.Text != `["socks5","http"]` {
		t.Errorf("list: %q", c.Text)
	}
	rec = flattenJSON(t, `{"pubproxy":{"protocols":[]},"web":[]}`)
	if c := rec.Get("pubproxy.protocols"); !c.Present || c.Text != "[]" {
		t.Errorf("empty list: %+v", c)
	}
	if c := rec.Get("web"); !c.Present || c.Text != "[]" {
		t.Errorf("Web IP's empty list: %+v", c)
	}
}

func TestValidateAndExpandFields(t *testing.T) {
	if err := ValidateFields([]string{"is_vpn", "vpn.provider", "vpn"}, KindAddress); err != nil {
		t.Errorf("valid fields rejected: %v", err)
	}
	// A typo is an error, not a column of blanks.
	err := ValidateFields([]string{"is_vpnn"}, KindAddress, KindASN, KindDomain)
	if err == nil || !strings.Contains(err.Error(), "is_vpnn") {
		t.Errorf("expected the typo to be named, got %v", err)
	}
	// A field is checked against the kinds asked for: an AS's name is no
	// address's field, and a domain's records are no AS's.
	if err := ValidateFields([]string{"name"}, KindAddress); err == nil {
		t.Error("an address-only lookup took an AS's field")
	}
	if err := ValidateFields([]string{"name", "registry", "whois.handle"}, KindAddress, KindASN); err != nil {
		t.Errorf("an AS's fields rejected: %v", err)
	}
	if err := ValidateFields([]string{"dns.mx"}, KindAddress, KindASN); err == nil {
		t.Error("a lookup with no domain took a domain's field")
	}
	if err := ValidateFields([]string{"dns", "is_registered"}, KindDomain); err != nil {
		t.Errorf("a domain's fields rejected: %v", err)
	}

	got := ExpandFields(KindAddress, []string{"is_vpn", "vpn"})
	want := []string{"is_vpn", "vpn.provider", "vpn.provider_name", "vpn.confidence", "vpn.last_seen"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("got %v, want %v", got, want)
	}
	got = ExpandFields(KindDomain, []string{"dns"})
	want = []string{"dns.a", "dns.aaaa", "dns.mx", "dns.ns", "dns.cname"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("domain: got %v, want %v", got, want)
	}
}

// Each kind's columns name every field the SDK's answer can carry, and nothing
// else, so neither CSV nor --field leaves one out. A field a new sdk-go brings
// fails here, on the bump that brings it.
func TestColumnsAreTheSDKsFields(t *testing.T) {
	for _, c := range []struct {
		kind    Kind
		answer  any
		columns []string
	}{
		{KindAddress, &internetdata.Result{}, Columns},
		{KindASN, &internetdata.AsnLookupResponse{}, ASNColumns},
		{KindDomain, &internetdata.DomainLookupResponse{}, DomainColumns},
	} {
		fill(reflect.ValueOf(c.answer).Elem())
		rec, err := Flatten(c.answer)
		if err != nil {
			t.Fatal(err)
		}
		named := make(map[string]bool, len(c.columns))
		for _, col := range c.columns {
			named[col] = true
			// An AS's whois record is a map in the SDK, keyed as the API
			// documents it; TestASNWhoisColumnsAreTheDocumentedKeys holds those.
			if c.kind == KindASN && strings.HasPrefix(col, "whois.") {
				continue
			}
			if !rec[col].Present {
				t.Errorf("%T: the columns name %s, which the SDK's answer does not carry", c.answer, col)
			}
		}
		for k := range rec {
			if c.kind == KindASN && k == "whois" {
				continue
			}
			if !named[k] {
				t.Errorf("%T: the SDK's answer carries %s, which its columns do not name: "+
					"append it, so no CSV column moves", c.answer, k)
			}
		}
	}
}

// The AS columns' whois fields are the keys the API documents for the record,
// every one of them.
func TestASNWhoisColumnsAreTheDocumentedKeys(t *testing.T) {
	var answer internetdata.AsnLookupResponse
	if err := json.Unmarshal([]byte(asnPaidAnswer), &answer); err != nil {
		t.Fatal(err)
	}
	documented := make([]string, 0, len(*answer.Whois))
	for k := range *answer.Whois {
		documented = append(documented, "whois."+k)
	}
	var named []string
	for _, col := range ASNColumns {
		if strings.HasPrefix(col, "whois.") {
			named = append(named, col)
		}
	}
	slices.Sort(documented)
	slices.Sort(named)
	if !reflect.DeepEqual(named, documented) {
		t.Errorf("ASNColumns' whois fields are %v\nthe documented record's keys are %v", named, documented)
	}
}

// Flags and Objects are what the readable block lists Columns by, so a flag or
// an object Columns gains and they do not name prints apart from its family.
func TestFlagsAndObjectsFollowColumns(t *testing.T) {
	var flags, objects []string
	seen := map[string]bool{}
	for _, col := range Columns {
		if strings.HasPrefix(col, "is_") && col != "is_bogon" {
			flags = append(flags, col)
		}
		if obj, _, nested := strings.Cut(col, "."); nested && !seen[obj] {
			objects = append(objects, obj)
			seen[obj] = true
		}
	}
	if !reflect.DeepEqual(Flags, flags) {
		t.Errorf("Flags = %v, want the flags Columns names, in its order: %v", Flags, flags)
	}
	if !reflect.DeepEqual(Objects, objects) {
		t.Errorf("Objects = %v, want the objects Columns names, in its order: %v", Objects, objects)
	}
	// Web IP's detail is a list, one column, and Anycast IP has none.
	for _, f := range Flags {
		detail := strings.TrimPrefix(f, "is_")
		if !seen[detail] && !slices.Contains(Columns, detail) && f != "is_anycast" {
			t.Errorf("%s has no detail in Columns", f)
		}
	}
}

// fill gives every member of v a value, so that its JSON carries every key.
func fill(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		v.Set(reflect.New(v.Type().Elem()))
		fill(v.Elem())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanSet() {
				fill(v.Field(i))
			}
		}
	}
}
