package agent

import (
	"github.com/madnikulin50/lowcode/agents/sdk"
)

func composeJobStatus(agentStatus string) string {
	return string(sdk.NormalizeStatus(agentStatus))
}

func (a *Agent) notifyCallback(target ScanTarget, id, kind string, items []Device) {
	st := a.GetStatus(id)
	var env sdk.Envelope
	if st != nil {
		e := st.Envelope(sdk.EventKind(kind), target.ScanRecordID, nsString(target.NamespaceID), items)
		if e != nil {
			env = *e
		}
	} else {
		env = sdk.Envelope{
			Service:     "cmdb",
			Operation:   "scan",
			ID:          id,
			JobID:       id,
			Kind:        sdk.EventKind(kind),
			Status:      sdk.NormalizeStatus(kind),
			RecordID:    target.ScanRecordID,
			NamespaceID: nsString(target.NamespaceID),
			Items:       items,
		}
	}
	if kind == "failed" {
		env.Kind = sdk.KindFailed
		env.Status = sdk.StatusFailed
	}
	if kind == "complete" {
		env.Kind = sdk.KindComplete
		env.Status = sdk.StatusCompleted
	}
	a.cb.NotifyEnv(id, target.CallbackURL, target.Token, env)
}

func nsString(id FlexUint64) string {
	if id == 0 {
		return ""
	}
	return sdk.FlexID(id).String()
}
