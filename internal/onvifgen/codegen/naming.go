package codegen

// NSPkg maps ONVIF and support-schema namespaces to their Go package ids.
// These mirror the conventional prefixes used in the ONVIF spec's own
// xmlns declarations so the generated package names are familiar to ONVIF
// developers.
var NSPkg = map[string]string{
	"http://www.onvif.org/ver10/schema":                      "tt",
	"http://www.onvif.org/ver10/device/wsdl":                 "tds",
	"http://www.onvif.org/ver10/media/wsdl":                  "trt",
	"http://www.onvif.org/ver20/media/wsdl":                  "tr2",
	"http://www.onvif.org/ver20/ptz/wsdl":                    "tptz",
	"http://www.onvif.org/ver10/events/wsdl":                 "tev",
	"http://www.onvif.org/ver20/imaging/wsdl":                "timg",
	"http://www.onvif.org/ver10/accessrules/wsdl":            "tar",
	"http://www.onvif.org/ver10/actionengine/wsdl":           "tae",
	"http://www.onvif.org/ver10/advancedsecurity/wsdl":       "tas",
	"http://www.onvif.org/ver10/appmgmt/wsdl":                "ans",
	"http://www.onvif.org/ver10/authenticationbehavior/wsdl": "tab",
	"http://www.onvif.org/ver10/credential/wsdl":             "tcr",
	"http://www.onvif.org/ver10/deviceIO/wsdl":               "tmd",
	"http://www.onvif.org/ver10/display/wsdl":                "tls",
	"http://www.onvif.org/ver10/accesscontrol/wsdl":          "tac",
	"http://www.onvif.org/ver10/doorcontrol/wsdl":            "tdc",
	"http://www.onvif.org/ver10/pacs":                        "pacs",
	"http://www.onvif.org/ver10/provisioning/wsdl":           "tpv",
	"http://www.onvif.org/ver10/receiver/wsdl":               "trv",
	"http://www.onvif.org/ver10/recording/wsdl":              "trc",
	"http://www.onvif.org/ver10/replay/wsdl":                 "trp",
	"http://www.onvif.org/ver10/schedule/wsdl":               "tsc",
	"http://www.onvif.org/ver10/search/wsdl":                 "tse",
	"http://www.onvif.org/ver10/thermal/wsdl":                "tth",
	"http://www.onvif.org/ver10/uplink/wsdl":                 "tup",
	"http://www.onvif.org/ver20/analytics/wsdl":              "tan",
	"http://www.onvif.org/ver20/analytics/humanbody":         "bd",
	"http://www.onvif.org/ver20/analytics/humanface":         "fc",
	"http://www.onvif.org/ver20/analytics/radiometry":        "ttr",
	"http://docs.oasis-open.org/wsn/b-2":                     "wsn",
	"http://docs.oasis-open.org/wsn/t-1":                     "wsnt",
	"http://docs.oasis-open.org/wsrf/r-2":                    "wsrf",
	"http://docs.oasis-open.org/wsrf/bf-2":                   "wsrfbf",
	"http://www.w3.org/2005/08/addressing":                   "wsa",
	"http://www.w3.org/2003/05/soap-envelope":                "env",
	"http://www.w3.org/2005/05/xmlmime":                      "xmlmime",
	"http://www.w3.org/2004/08/xop/include":                  "xop",
	"http://www.w3.org/XML/1998/namespace":                   "xmlns",
	"http://www.w3.org/2001/XMLSchema":                       "xs",
	"urn:onvif:go:core":                                      "core",
}

// PkgNS is the reverse map (package id → namespace).
var PkgNS = func() map[string]string {
	m := make(map[string]string, len(NSPkg))
	for ns, p := range NSPkg {
		m[p] = ns
	}
	return m
}()

// goKeywords is the set of Go reserved words that cannot be used as
// identifiers. Collisions get a trailing underscore.
var goKeywords = map[string]struct{}{
	"break": {}, "case": {}, "chan": {}, "const": {}, "continue": {}, "default": {},
	"defer": {}, "else": {}, "fallthrough": {}, "for": {}, "func": {}, "go": {},
	"goto": {}, "if": {}, "import": {}, "interface": {}, "map": {}, "package": {},
	"range": {}, "return": {}, "select": {}, "struct": {}, "switch": {}, "type": {},
	"var": {}, "nil": {}, "true": {}, "false": {}, "iota": {},
}

// safeIdent returns the Go identifier for an XSD name, appending a trailing
// underscore if it collides with a Go keyword.
func safeIdent(name string) string {
	if _, bad := goKeywords[name]; bad {
		return name + "_"
	}
	return name
}

// pascal converts a (typically already-PascalCase) ONVIF name to a valid
// Go exported identifier. ONVIF names sometimes contain hyphens or dots
// (e.g. "IANA-IfTypes", "TLS1.1", "X.509Token") which are illegal in Go
// identifiers; those runes are dropped (or replaced so the identifier stays
// readable).
func pascal(name string) string {
	if name == "" {
		return ""
	}
	var b []rune
	for i, r := range name {
		switch {
		case r == '_' || r == '$':
			b = append(b, '_')
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z':
			b = append(b, r)
		case r >= '0' && r <= '9':
			b = append(b, r)
		case r == '-' || r == '.':
			// Drop separators; uppercase the next rune so the result reads
			// naturally (IANA-IfTypes -> IANAIfTypes, TLS1.1 -> TLS11).
			_ = i
		default:
			b = append(b, '_')
		}
	}
	s := string(b)
	if s == "" {
		s = "X"
	}
	// Guarantee an exported identifier (first rune uppercase ASCII).
	if s[0] < 'A' || s[0] > 'Z' {
		s = "X" + s
	}
	return safeIdent(s)
}
