package agent

import "github.com/madnikulin50/lowcode/agents/sdk"

func field(key, widget, label string, extra ...func(*sdk.Field)) sdk.Field {
	out := sdk.Field{Key: key, Widget: widget, Label: label}
	for _, fn := range extra {
		fn(&out)
	}
	return out
}

func req(f *sdk.Field)  { f.Required = true }
func tmpl(f *sdk.Field) { f.Template = true }

func Components() []sdk.Component {
	return []sdk.Component{
		sdk.Desc{D: sdk.Descriptor{
			Type: "cmdb/scan", Label: "CMDB: scan network",
			Description: "Scan a CIDR range and ingest discovered devices",
			Category:    sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "cmdb", Operation: "scan",
			ConfigFields: []sdk.Field{
				field("cidr", "string", "CIDR", req, tmpl),
				field("namespaceID", "string", "Namespace ID", tmpl),
			},
		}},
	}
}
