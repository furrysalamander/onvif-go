package codegen

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/furrysalamander/onvif-go/internal/onvifgen/ir"
)

var serviceNames = map[string]string{
	"http://www.onvif.org/ver10/device/wsdl":                 "devicemgmt",
	"http://www.onvif.org/ver10/media/wsdl":                  "media",
	"http://www.onvif.org/ver20/media/wsdl":                  "media2",
	"http://www.onvif.org/ver20/ptz/wsdl":                    "ptz",
	"http://www.onvif.org/ver10/events/wsdl":                 "events",
	"http://www.onvif.org/ver20/imaging/wsdl":                "imaging",
	"http://www.onvif.org/ver10/accessrules/wsdl":            "accessrules",
	"http://www.onvif.org/ver10/actionengine/wsdl":           "actionengine",
	"http://www.onvif.org/ver10/advancedsecurity/wsdl":       "advancedsecurity",
	"http://www.onvif.org/ver10/appmgmt/wsdl":                "appmgmt",
	"http://www.onvif.org/ver10/authenticationbehavior/wsdl": "authenticationbehavior",
	"http://www.onvif.org/ver10/credential/wsdl":             "credential",
	"http://www.onvif.org/ver10/deviceIO/wsdl":               "deviceio",
	"http://www.onvif.org/ver10/display/wsdl":                "display",
	"http://www.onvif.org/ver10/accesscontrol/wsdl":          "accesscontrol",
	"http://www.onvif.org/ver10/doorcontrol/wsdl":            "doorcontrol",
	"http://www.onvif.org/ver10/provisioning/wsdl":           "provisioning",
	"http://www.onvif.org/ver10/receiver/wsdl":               "receiver",
	"http://www.onvif.org/ver10/recording/wsdl":              "recording",
	"http://www.onvif.org/ver10/replay/wsdl":                 "replay",
	"http://www.onvif.org/ver10/schedule/wsdl":               "schedule",
	"http://www.onvif.org/ver10/search/wsdl":                 "search",
	"http://www.onvif.org/ver10/thermal/wsdl":                "thermal",
	"http://www.onvif.org/ver10/uplink/wsdl":                 "uplink",
	"http://www.onvif.org/ver20/analytics/wsdl":              "analytics",
}

type facadeOp struct {
	Name   string
	Action string
	Res    string
	Req    string
	Fields []facadeField
}

type facadeField struct {
	Name    string
	Type    string
	XMLName string
}

func EmitFacades(tab *SymTab, modules []*ir.Module, outBase string) (map[string]string, error) {
	out := map[string]string{}
	for _, m := range modules {
		if m.Kind != ir.ModuleWSDL {
			continue
		}
		w := m.WSDL
		svcName, ok := serviceNames[w.TargetNS]
		if !ok {
			continue
		}
		schemaPkg, ok := NSPkg[w.TargetNS]
		if !ok {
			continue
		}
		ops, err := collectOps(w, schemaPkg, tab)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", svcName, err)
		}
		if len(ops) == 0 {
			continue
		}
		src := emitClientFile(svcName, schemaPkg, w.TargetNS, ops)
		path := filepath.Join(outBase, svcName+"/client_gen.go")
		path = filepath.ToSlash(path)
		out[path] = src
	}
	return out, nil
}

func collectOps(w *ir.WSDL, schemaPkg string, tab *SymTab) ([]facadeOp, error) {
	msgMap := map[string]*ir.Message{}
	for _, m := range w.Messages {
		if m.Name != "" {
			msgMap[m.Name] = m
		}
	}
	actionMap := map[string]string{}
	for _, b := range w.Bindings {
		for _, bo := range b.Operations {
			if bo.SOAPAction != "" {
				actionMap[bo.Name] = bo.SOAPAction
			}
		}
	}
	r := &resolver{tab: tab, curPkg: schemaPkg}

	var ops []facadeOp
	for _, pt := range w.PortTypes {
		for _, op := range pt.Operations {
			if op.Name == "" {
				continue
			}
			fo := facadeOp{Name: op.Name}
			fo.Action = actionMap[op.Name]
			if fo.Action == "" {
				fo.Action = w.TargetNS + "/" + op.Name
			}
			var reqQName ir.QName
			if op.Input != nil {
				reqMsg, ok := resolveMessage(op.Input.Message, msgMap)
				if ok && len(reqMsg.Parts) > 0 {
					reqQName = reqMsg.Parts[0].Element
					if reqQName.Local == "" {
						reqQName = reqMsg.Parts[0].Type
					}
				}
			}
			if reqQName.Local == "" {
				continue
			}
			fo.Req = qualifyType(schemaPkg, reqQName)
			if fo.Req == "" || fo.Req == "core.Extension" {
				continue
			}

			if op.Output != nil {
				resMsg, ok := resolveMessage(op.Output.Message, msgMap)
				if ok && len(resMsg.Parts) > 0 {
					resQName := resMsg.Parts[0].Element
					if resQName.Local == "" {
						resQName = resMsg.Parts[0].Type
					}
					if resQName.Local != "" {
						fo.Res = qualifyType(schemaPkg, resQName)
					}
				}
			}
			if fo.Res == "" {
				continue
			}

			fo.Fields = requestFields(r, w.TargetNS, reqQName)
			ops = append(ops, fo)
		}
	}
	sort.SliceStable(ops, func(i, j int) bool { return ops[i].Name < ops[j].Name })
	return ops, nil
}

func requestFields(r *resolver, targetNS string, q ir.QName) []facadeField {
	sym, ok := r.tab.Lookup(targetNS, q.Local)
	if !ok || sym.Element == nil {
		return nil
	}
	e := sym.Element

	var cmplx *ir.ComplexType
	if e.Type != (ir.QName{}) {
		s, ok := r.tab.Lookup(e.Type.NS, e.Type.Local)
		if ok && s.Complex != nil {
			cmplx = s.Complex
		}
	}
	if cmplx == nil && e.InlineComplex != nil {
		cmplx = e.InlineComplex
	}
	if cmplx == nil {
		return nil
	}

	var shapes []fieldShape
	anySeen := false
	_ = r.particleFields(cmplx.Content, &shapes, &anySeen)

	var params []facadeField
	for _, f := range shapes {
		if f.GoType == "" {
			continue
		}
		if f.IsExtension {
			continue
		}
		xmlName := f.XMLTag
		if comma := strings.IndexByte(xmlName, ','); comma >= 0 {
			xmlName = xmlName[:comma]
		}
		if xmlName == "" {
			continue
		}
		paramName := camel(xmlName)
		if paramName == "" {
			continue
		}
		// Qualify type with package if needed — the resolver omits
		// same-package prefixes, but facades always need them.
		goty := qualifyFieldType(f.GoType, targetNS)
		params = append(params, facadeField{
			Name:    paramName,
			Type:    goty,
			XMLName: xmlName,
		})
	}
	return params
}

func qualifyFieldType(goty, targetNS string) string {
	if goty == "" || strings.Contains(goty, ".") || isBuiltin(goty) {
		return goty
	}
	pkg, ok := NSPkg[targetNS]
	if !ok || pkg == "" {
		return goty
	}
	// Handle pointer and slice prefixes.
	prefix := ""
	for strings.HasPrefix(goty, "[]") {
		prefix += "[]"
		goty = goty[2:]
	}
	for strings.HasPrefix(goty, "*") {
		prefix += "*"
		goty = goty[1:]
	}
	return prefix + pkg + "." + goty
}

func isBuiltin(goty string) bool {
	base := strings.TrimLeft(goty, "*[]")
	switch base {
	case "string", "bool", "byte", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return true
	}
	return false
}

func camel(s string) string {
	if s == "" {
		return ""
	}
	n := strings.ToLower(s[:1]) + s[1:]
	if _, bad := goKeywords[n]; bad {
		n += "_"
	}
	return n
}

func resolveMessage(q ir.QName, msgMap map[string]*ir.Message) (*ir.Message, bool) {
	m, ok := msgMap[q.Local]
	if ok {
		return m, true
	}
	for _, m := range msgMap {
		if m.Name == q.Local {
			return m, true
		}
	}
	return nil, false
}

func qualifyType(schemaPkg string, q ir.QName) string {
	pkg, ok := NSPkg[q.NS]
	if !ok || pkg == "" {
		return ""
	}
	name := pascal(q.Local)
	if name == "" {
		return ""
	}
	if pkg == schemaPkg {
		return name
	}
	return pkg + "." + name
}

func emitClientFile(svcName, schemaPkg, targetNS string, ops []facadeOp) string {
	var b strings.Builder
	b.WriteString("// Code generated by onvifgen; DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", svcName)

	imports := map[string]bool{schemaPkg: true}
	for _, op := range ops {
		for _, ty := range []string{op.Req, op.Res} {
			cleanTy := strings.TrimLeft(ty, "*[]")
			for _, pkg := range NSPkg {
				if strings.HasPrefix(cleanTy, pkg+".") && pkg != "" {
					imports[pkg] = true
				}
			}
			if strings.HasPrefix(cleanTy, "time.") {
				imports["time"] = true
			}
		}
		for _, f := range op.Fields {
			baseType := strings.TrimLeft(f.Type, "*[]")
			for _, pkg := range NSPkg {
				if strings.HasPrefix(baseType, pkg+".") && pkg != "" {
					imports[pkg] = true
				}
			}
			if strings.HasPrefix(baseType, "time.") {
				imports["time"] = true
			}
		}
	}
	imports["soaphdr"] = true

	pkgList := make([]string, 0, len(imports))
	for pkg := range imports {
		pkgList = append(pkgList, pkg)
	}
	sort.Strings(pkgList)

	b.WriteString("import (\n")
	b.WriteString("\t\"context\"\n")
	b.WriteString("\t\"net/http\"\n\n")
	for _, pkg := range pkgList {
		switch pkg {
		case "soaphdr":
			b.WriteString("\t\"github.com/furrysalamander/onvif-go/onvif/soaphdr\"\n")
		case "time":
			b.WriteString("\t\"time\"\n")
		default:
			fmt.Fprintf(&b, "\t\"github.com/furrysalamander/onvif-go/onvif/schema/%s\"\n", pkg)
		}
	}
	b.WriteString(")\n\n")
	fmt.Fprintf(&b, "const actionBase = %q\n\n", targetNS)
	b.WriteString("type Client struct {\n\tc *soaphdr.Client\n}\n\n")
	b.WriteString("func NewClient(endpoint, username, password string) *Client {\n")
	b.WriteString("\treturn &Client{c: soaphdr.New(endpoint, username, password)}\n")
	b.WriteString("}\n\n")
	b.WriteString("func NewClientWithTransport(c *soaphdr.Client) *Client {\n")
	b.WriteString("\treturn &Client{c: c}\n")
	b.WriteString("}\n\n")
	b.WriteString("func (c *Client) HTTP() *http.Client { return c.c.HTTP }\n")
	b.WriteString("func (c *Client) SetHTTP(h *http.Client) { c.c.HTTP = h }\n\n")

	for _, op := range ops {
		req := ensurePkg(op.Req, schemaPkg)
		res := ensurePkg(op.Res, schemaPkg)
		emitMethod(&b, op.Name, req, res, op.Action, op.Fields)
	}
	return b.String()
}

func emitMethod(b *strings.Builder, name, reqType, resType, action string, fields []facadeField) {
	var params []string
	var reqFields []string
	for _, f := range fields {
		params = append(params, f.Name+" "+f.Type)
		reqFields = append(reqFields, fmt.Sprintf("%s: %s", pascal(f.XMLName), f.Name))
	}
	fmt.Fprintf(b, "func (c *Client) %s(ctx context.Context", name)
	for _, p := range params {
		fmt.Fprintf(b, ", %s", p)
	}
	fmt.Fprintf(b, ") (*%s, error) {\n", resType)
	fmt.Fprintf(b, "\treq := &%s{", reqType)
	if len(reqFields) > 0 {
		b.WriteString(strings.Join(reqFields, ", "))
	}
	b.WriteString("}\n")
	fmt.Fprintf(b, "\tout := &%s{}\n", resType)
	fmt.Fprintf(b, "\terr := c.c.Do(ctx, %q, req, out)\n", action)
	b.WriteString("\treturn out, err\n")
	b.WriteString("}\n\n")
}

func ensurePkg(ty, pkg string) string {
	if strings.Contains(ty, ".") {
		return ty
	}
	return pkg + "." + ty
}
