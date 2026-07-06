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
	Name      string
	Action    string
	SchemaPkg string
	ReqType   string
	ResType   string
}

func EmitFacades(modules []*ir.Module, outBase string) (map[string]string, error) {
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
		ops, err := collectOps(w, schemaPkg)
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

func collectOps(w *ir.WSDL, schemaPkg string) ([]facadeOp, error) {
	msgMap := map[string]*ir.Message{}
	for _, m := range w.Messages {
		key := m.Name
		if key == "" {
			continue
		}
		msgMap[key] = m
	}

	actionMap := map[string]string{}
	for _, b := range w.Bindings {
		for _, bo := range b.Operations {
			if bo.SOAPAction != "" {
				actionMap[bo.Name] = bo.SOAPAction
			}
		}
	}

	var ops []facadeOp
	for _, pt := range w.PortTypes {
		for _, op := range pt.Operations {
			fo := facadeOp{Name: op.Name}
			if fo.Name == "" {
				continue
			}
			fo.SchemaPkg = schemaPkg
			fo.Action = actionMap[op.Name]
			if fo.Action == "" {
				fo.Action = w.TargetNS + "/" + op.Name
			}

			if op.Input != nil {
				reqMsg, ok := resolveMessage(op.Input.Message, msgMap)
				if ok && len(reqMsg.Parts) > 0 {
					reqElem := reqMsg.Parts[0].Element
					if reqElem.Local == "" && reqMsg.Parts[0].Type.Local != "" {
						reqElem = reqMsg.Parts[0].Type
					}
					if reqElem.Local != "" {
						fo.ReqType = qualifyType(schemaPkg, reqElem)
					}
				}
			}
			if op.Output != nil {
				resMsg, ok := resolveMessage(op.Output.Message, msgMap)
				if ok && len(resMsg.Parts) > 0 {
					resElem := resMsg.Parts[0].Element
					if resElem.Local == "" && resMsg.Parts[0].Type.Local != "" {
						resElem = resMsg.Parts[0].Type
					}
					if resElem.Local != "" {
						fo.ResType = qualifyType(schemaPkg, resElem)
					}
				}
			}
			if fo.ReqType == "" || fo.ResType == "" {
				continue
			}
			ops = append(ops, fo)
		}
	}
	sort.SliceStable(ops, func(i, j int) bool { return ops[i].Name < ops[j].Name })
	return ops, nil
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
		for _, ty := range []string{op.ReqType, op.ResType} {
			for _, pkg := range NSPkg {
				if strings.HasPrefix(ty, pkg+".") && pkg != "" {
					imports[pkg] = true
				}
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
		if pkg == "soaphdr" {
			b.WriteString("\t\"github.com/furrysalamander/onvif-go/onvif/soaphdr\"\n")
		} else {
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
		req := ensurePkg(op.ReqType, schemaPkg)
		res := ensurePkg(op.ResType, schemaPkg)
		fmt.Fprintf(&b, "func (c *Client) %s(ctx context.Context) (*%s, error) {\n", op.Name, res)
		fmt.Fprintf(&b, "\tout := &%s{}\n", res)
		fmt.Fprintf(&b, "\terr := c.c.Do(ctx, %q, &%s{}, out)\n", op.Action, req)
		b.WriteString("\treturn out, err\n")
		b.WriteString("}\n\n")
	}
	return b.String()
}

func ensurePkg(ty, pkg string) string {
	if strings.Contains(ty, ".") {
		return ty
	}
	return pkg + "." + ty
}
