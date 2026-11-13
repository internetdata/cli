// Package lib holds the CLI's output machinery: how one answer becomes a table
// row, a CSV line, a JSON object or a colored block.
package lib

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Kind is what a lookup asked about, which decides the fields its answer can
// carry and the order they are listed in.
type Kind int

const (
	// KindAddress is an IP address, from a lookup, a batch or myip.
	KindAddress Kind = iota
	// KindASN is an autonomous system.
	KindASN
	// KindDomain is a domain.
	KindDomain
)

// Columns is every field an address's answer can carry, in CSV's order: the
// address, then the flags, then each flag's detail object, then where the
// address is and who holds it, then the fields added since.
//
// Fixed and explicit rather than derived by reflection, because this order IS
// the CSV contract. A caller's column indexes must not move because the API
// grew a field, and a new field appearing silently at the end is far easier to
// absorb than one appearing in the middle. The readable block lists each one
// with its object's other fields all the same.
var Columns = []string{
	"ip",
	"is_bogon",

	"is_vpn",
	"is_tor",
	"is_relay",
	"is_resproxy",
	"is_dcproxy",
	"is_mobproxy",
	"is_pubproxy",
	"is_hosting",
	"is_cdn",
	"is_anycast",
	"is_carrier",
	"is_satellite",
	"is_ssh",
	"is_smtp",
	"is_web",
	"is_dns_resolver",
	"is_bittorrent",

	"vpn.provider", "vpn.provider_name", "vpn.confidence", "vpn.last_seen",
	"tor.provider", "tor.provider_name", "tor.confidence", "tor.last_seen",
	"relay.provider", "relay.provider_name", "relay.confidence", "relay.last_seen",
	"resproxy.provider", "resproxy.provider_name", "resproxy.first_seen", "resproxy.last_seen",
	"resproxy.hits", "resproxy.hits_days_pct", "resproxy.providers_num",
	"dcproxy.provider", "dcproxy.provider_name", "dcproxy.first_seen", "dcproxy.last_seen",
	"dcproxy.hits", "dcproxy.hits_days_pct", "dcproxy.providers_num",
	"mobproxy.provider", "mobproxy.provider_name", "mobproxy.first_seen", "mobproxy.last_seen",
	"mobproxy.hits", "mobproxy.hits_days_pct", "mobproxy.providers_num",
	"pubproxy.entry", "pubproxy.exit", "pubproxy.protocols", "pubproxy.num_ports", "pubproxy.num_proxies",
	"pubproxy.last_seen", "pubproxy.last_usable",
	"hosting.provider", "hosting.provider_name", "hosting.confidence", "hosting.last_seen",
	"cdn.provider", "cdn.provider_name", "cdn.confidence", "cdn.last_seen",
	"carrier.name", "carrier.mcc", "carrier.mnc", "carrier.country", "carrier.confidence",
	"satellite.provider", "satellite.provider_name", "satellite.orbit", "satellite.segment",
	"satellite.country", "satellite.asn", "satellite.confidence",
	"ssh.port", "ssh.banner", "ssh.host_key_fp", "ssh.session_ok", "ssh.first_seen", "ssh.last_seen",
	"smtp.port", "smtp.banner", "smtp.session_ok", "smtp.first_seen", "smtp.last_seen",
	"web",
	"dns_resolver.port", "dns_resolver.recursion_available", "dns_resolver.rcode",
	"dns_resolver.answer_correct", "dns_resolver.first_seen", "dns_resolver.last_seen",
	"bittorrent.verified", "bittorrent.id_matches_ip", "bittorrent.first_seen", "bittorrent.last_seen",
	"bittorrent.last_reachable", "bittorrent.days_seen", "bittorrent.num_ports",
	"bittorrent.peak_num_ports", "bittorrent.indexer", "bittorrent.clients",

	"location.continent", "location.continent_name", "location.country", "location.country_name",
	"location.timezone", "location.currency_code", "location.currency_name", "location.currency_symbol",
	"location.flag_emoji", "location.flag_unicode", "location.region", "location.region_code",
	"location.city", "location.postal_code", "location.latitude", "location.longitude",
	"location.accuracy_radius", "location.geoname_id", "location.dma_code", "location.last_changed",
	"asn.asn", "asn.name", "asn.domain", "asn.type", "asn.prefix", "asn.country", "asn.rpki_status",
	"company.name", "company.domain", "company.type", "company.country",
	"hostname",
	"abuse.email", "abuse.name", "abuse.phone", "abuse.address", "abuse.country", "abuse.network",
	"abuse.validated", "abuse.unvalidated_since",
	"whois.id", "whois.name", "whois.org", "whois.country", "whois.status", "whois.source",
	"whois.domain", "whois.abuse_email", "whois.asn",
	"rwhois.id", "rwhois.name", "rwhois.country", "rwhois.source", "rwhois.domain", "rwhois.abuse_email",
	"hosted_domains.total",

	// On a free-plan key's answers alone: false for its daily trial's lookups with
	// every field, true once they are spent and the answer is Unlimited's.
	"trial_expired",
}

// Flags are an address's top-level booleans, which is what a summary line shows.
var Flags = []string{
	"is_vpn", "is_tor", "is_relay", "is_resproxy", "is_dcproxy", "is_mobproxy", "is_pubproxy",
	"is_hosting", "is_cdn", "is_anycast", "is_carrier", "is_satellite", "is_ssh", "is_smtp",
	"is_web", "is_dns_resolver", "is_bittorrent",
}

// Objects are an address's objects, in the order output presents them: each
// flag's detail object, then the location, the AS, the company, the abuse
// contact, the registry records and the hosted domains, which every keyed
// answer carries whole.
var Objects = []string{
	"vpn", "tor", "relay", "resproxy", "dcproxy", "mobproxy", "pubproxy", "hosting", "cdn",
	"carrier", "satellite", "ssh", "smtp", "dns_resolver", "bittorrent",
	"location", "asn", "company", "abuse", "whois", "rwhois", "hosted_domains",
}

// ASNColumns is every field an AS's answer can carry, in CSV's order, under the
// same rule as Columns. The whois record is keyed as the API documents it, and
// prefixes is one column holding the list.
var ASNColumns = []string{
	"asn", "name", "country", "registry", "allocated", "domain",
	"type", "num_ips", "num_prefixes4", "num_prefixes6", "rpki_valid", "rpki_invalid", "rpki_not_found",
	"whois.handle", "whois.name", "whois.org", "whois.country", "whois.source", "whois.status",
	"whois.created", "whois.updated", "whois.descr", "whois.remarks", "whois.admin", "whois.tech",
	"whois.abuse", "whois.maintainer", "whois.maintainer_email", "whois.abuse_email", "whois.abuse_name",
	"whois.abuse_phone", "whois.abuse_kind", "whois.abuse_country", "whois.abuse_city",
	"whois.admin_name", "whois.admin_email", "whois.admin_phone", "whois.admin_kind",
	"whois.admin_country", "whois.admin_city", "whois.tech_name", "whois.tech_email", "whois.tech_phone",
	"whois.tech_kind", "whois.tech_country", "whois.tech_city", "whois.org_name", "whois.org_address",
	"whois.org_country", "whois.org_city", "whois.abuse_validated", "whois.abuse_unvalidated_since",
	"prefixes",
	"trial_expired",
}

// DomainColumns is every field a domain's answer can carry, in CSV's order,
// under the same rule as Columns. Each record type is one column holding its
// list, as are the ranges.
var DomainColumns = []string{
	"domain", "is_registered", "is_disposable", "has_a", "has_aaaa", "has_mx", "has_ns", "has_cname",
	"dns.a", "dns.aaaa", "dns.mx", "dns.ns", "dns.cname",
	"ip_ranges.domain", "ip_ranges.redirects_to", "ip_ranges.total", "ip_ranges.ipv4", "ip_ranges.ranges",
	"disposable.confidence", "disposable.providers", "disposable.first_seen", "disposable.last_seen",
	"trial_expired",
}

// schema is how one kind of answer is listed: its columns, the fields the
// readable block lists first, in order, and then its objects, each by its own
// fields. A top-level field neither names, as trial_expired, comes after them.
type schema struct {
	columns []string
	lead    []string
	objects []string
	// key is the field naming what was looked up, which an error's line carries.
	key string
}

var schemas = map[Kind]schema{
	KindAddress: {
		columns: Columns,
		lead:    append([]string{"ip", "is_bogon"}, Flags...),
		objects: Objects,
		key:     "ip",
	},
	KindASN: {
		columns: ASNColumns,
		lead: []string{
			"asn", "name", "country", "registry", "allocated", "domain", "type", "num_ips",
			"num_prefixes4", "num_prefixes6", "rpki_valid", "rpki_invalid", "rpki_not_found",
		},
		objects: []string{"whois"},
		key:     "asn",
	},
	KindDomain: {
		columns: DomainColumns,
		lead: []string{
			"domain", "is_registered", "is_disposable", "has_a", "has_aaaa", "has_mx", "has_ns", "has_cname",
		},
		objects: []string{"dns", "ip_ranges", "disposable"},
		key:     "domain",
	},
}

// Cell is one field of one answer.
//
// Present is the whole point and is not the same as an empty Text. A field the
// plan does not include is ABSENT; a field it does include and that is false is
// PRESENT and false. Collapsing the two is the mistake this product's API
// design exists to prevent, and it would be a shame to reintroduce it in the
// client that prints the answer.
type Cell struct {
	Text    string
	Present bool
	// EmptyObject is a detail object that arrived as {}, which says the flag
	// above it is false. Its Text is "", as a present empty string's is.
	EmptyObject bool
}

// Record is one answer flattened to dotted field names.
type Record map[string]Cell

// Flatten converts an answer into dotted fields.
//
// It goes through JSON rather than reflecting over the struct, so a field the
// SDK gains appears here with no change, and absence falls out of the wire
// shape exactly as the API sent it: an optional field the plan excludes is a
// missing key, not a zero value.
func Flatten(v any) (Record, error) {
	tree, err := asTree(v)
	if err != nil {
		return nil, err
	}
	rec := make(Record, len(Columns))
	walk(rec, "", tree)
	return rec, nil
}

// asTree is an answer as the API sent it: its JSON, decoded into maps. Every
// format reads the answer through it, so each one carries the API's keys and
// nesting whatever Go type the answer arrived in.
func asTree(v any) (map[string]any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var tree map[string]any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return nil, err
	}
	return tree, nil
}

// walk flattens nested objects into dotted keys.
func walk(rec Record, prefix string, tree map[string]any) {
	for k, v := range tree {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if sub, ok := v.(map[string]any); ok {
			// An EMPTY detail object is meaningful on its own: it says the flag
			// above it is false. Recorded as a present-but-empty cell so it can
			// be told apart from a detail object the plan does not include.
			if len(sub) == 0 {
				rec[key] = Cell{Present: true, EmptyObject: true}
				continue
			}
			walk(rec, key, sub)
			continue
		}
		rec[key] = Cell{Text: scalar(v), Present: true}
	}
}

// scalar renders a JSON value as the text a table or CSV shows.
func scalar(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		// JSON has one number type; render whole numbers without a decimal
		// point so a hit count reads as 42 rather than 42.000000.
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(b)
	}
}

// Get reads one dotted field.
func (r Record) Get(field string) Cell { return r[field] }

// Present lists the fields this answer actually carries, in display order: for
// an address, the address, the flags, then each object.
//
// A detail object that arrived EMPTY is listed by its own name rather than by
// its fields, because that is what it is - `"vpn": {}` says the flag above it is
// false, and expanding it into four blank rows would say something else.
func (r Record) Present(kind Kind) []string {
	return r.listed(schemas[kind], false)
}

// All lists what Present does and, each in its place among them, the fields
// this answer does not carry, which is what --show-absent prints.
func (r Record) All(kind Kind) []string {
	return r.listed(schemas[kind], true)
}

// Selected lists what a --field selection names, an object as its fields,
// except that a detail object which arrived empty is listed once by its own
// name, as Present lists it.
func (r Record) Selected(kind Kind, fields []string) []string {
	out := make([]string, 0, len(fields))
	seen := make(map[string]bool)
	for _, f := range ExpandFields(kind, fields) {
		obj, _, nested := strings.Cut(f, ".")
		if !nested || !r[obj].EmptyObject {
			out = append(out, f)
			continue
		}
		if !seen[obj] {
			out = append(out, obj)
			seen[obj] = true
		}
	}
	return out
}

func (r Record) listed(s schema, absent bool) []string {
	out := make([]string, 0, len(s.columns))
	seen := make(map[string]bool, len(s.columns))
	emit := func(k string) {
		if (absent || r[k].Present) && !seen[k] {
			out = append(out, k)
			seen[k] = true
		}
	}

	for _, f := range s.lead {
		emit(f)
	}
	for _, obj := range s.objects {
		if r[obj].Present {
			emit(obj)
			continue
		}
		for _, col := range s.columns {
			if strings.HasPrefix(col, obj+".") {
				emit(col)
			}
		}
	}
	// A top-level field that is neither a flag nor an object, as trial_expired.
	for _, col := range s.columns {
		if !strings.Contains(col, ".") {
			emit(col)
		}
	}
	// A field the API grew since this build must still be printed, or the CLI
	// silently hides what the customer is paying for.
	extra := make([]string, 0)
	for k, c := range r {
		if !seen[k] && c.Present {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	return append(out, extra...)
}

// ValidateFields refuses a --field selection naming something no answer of the
// kinds asked for could carry, so a typo is an error rather than a column of
// blanks.
//
// A field that is merely ABSENT from this plan is accepted: it is a real field,
// and printing it empty is the honest answer.
func ValidateFields(fields []string, kinds ...Kind) error {
	known := make(map[string]bool)
	for _, kind := range kinds {
		for _, col := range schemas[kind].columns {
			known[col] = true
			// A bare object name selects the whole object.
			if i := strings.IndexByte(col, '.'); i > 0 {
				known[col[:i]] = true
			}
		}
	}
	var bad []string
	for _, f := range fields {
		if !known[f] {
			bad = append(bad, f)
		}
	}
	if len(bad) == 0 {
		return nil
	}
	return fmt.Errorf("unknown field(s): %s", strings.Join(bad, ", "))
}

// ExpandFields turns a selection into concrete columns, so naming an object
// selects every field inside it.
func ExpandFields(kind Kind, fields []string) []string {
	columns := schemas[kind].columns
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if strings.ContainsRune(f, '.') {
			out = append(out, f)
			continue
		}
		matched := false
		for _, col := range columns {
			if strings.HasPrefix(col, f+".") {
				out = append(out, col)
				matched = true
			}
		}
		if !matched {
			out = append(out, f)
		}
	}
	return out
}

// selectFields keeps what a --field selection names, nested as in the full
// answer: `is_vpn` keeps that flag, `vpn` the whole detail object as it was
// served, and `vpn.provider` that object holding its provider alone.
//
// A field the answer does not carry stays out rather than appearing as null,
// and a detail object that arrived empty stays `{}` whichever of its fields
// was named, because that is the answer saying its flag is false.
func selectFields(tree map[string]any, fields []string) map[string]any {
	out := make(map[string]any, len(fields))
	for _, f := range fields {
		name, sub, nested := strings.Cut(f, ".")
		v, ok := tree[name]
		if !ok {
			continue
		}
		if !nested {
			out[name] = v
			continue
		}
		obj, ok := v.(map[string]any)
		if !ok {
			continue
		}
		kept, ok := out[name].(map[string]any)
		if !ok {
			kept = map[string]any{}
			out[name] = kept
		}
		if sv, ok := obj[sub]; ok {
			kept[sub] = sv
		}
	}
	return out
}
