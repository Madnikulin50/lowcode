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
			Type: "backup/run", Label: "Backup: run", Description: "Start a backup job from a source or policy",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "backup",
			ConfigFields: []sdk.Field{
				field("sourceID", "string", "Source ID", tmpl),
				field("policyID", "string", "Policy ID", tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/restore", Label: "Backup: restore", Description: "Restore a snapshot",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "restore",
			ConfigFields: []sdk.Field{
				field("snapshotID", "string", "Snapshot ID", req, tmpl),
				field("destType", "string", "Destination type", tmpl),
				field("destPath", "string", "Destination path", tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/prune", Label: "Backup: prune", Description: "Apply retention and delete expired snapshots",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "prune",
			ConfigFields: []sdk.Field{
				field("policyID", "string", "Policy ID", tmpl),
				field("sourceID", "string", "Source ID", tmpl),
				field("retentionDays", "number", "Retention days"),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/due", Label: "Backup: run due", Description: "Run policies whose cron matches now",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: false,
			Service: "backup", Operation: "due",
		}},
	}
}
