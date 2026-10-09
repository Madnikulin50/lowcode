package agent

import (
	"github.com/madnikulin50/lowcode/agents/sdk"
)

func (a *Agent) notifyCallback(st *JobStatus, kind string) {
	if st == nil {
		return
	}
	a.cb.NotifyEnv(st.ID, st.CallbackURL, st.Token, *st.Envelope(sdk.EventKind(kind)))
}
