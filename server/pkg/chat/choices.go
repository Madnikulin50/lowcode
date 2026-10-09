package chat

import (
	"sync"
	"time"
)

var modelChoices struct {
	mux     sync.Mutex
	names   []string
	fetched time.Time
}

const (
	modelChoicesTTL     = time.Minute
	modelChoicesTimeout = 400 * time.Millisecond
)

// ModelChoices is what an editor offers for a "model" field: the roles
// (rulesgo.ai, mcp.agent, ...) followed by the models Ollama has installed
// and the admin allows.
//
// It is meant for building UI and so never makes it wait: the Ollama list is
// cached for a minute and fetched with a short timeout, so an Ollama that is
// slow or down only means the list shows the roles (and whatever was fetched
// last).
func ModelChoices() []string {
	modelChoices.mux.Lock()
	fresh := time.Since(modelChoices.fetched) < modelChoicesTTL
	names := modelChoices.names
	modelChoices.mux.Unlock()

	if !fresh {
		done := make(chan []string, 1)
		go func() {
			all, err := AvailableModels()
			if err != nil {
				done <- nil
				return
			}
			done <- all
		}()

		select {
		case all := <-done:
			if all != nil {
				names = all
				modelChoices.mux.Lock()
				modelChoices.names, modelChoices.fetched = all, time.Now()
				modelChoices.mux.Unlock()
			}
		case <-time.After(modelChoicesTimeout):
		}
	}

	out := []string{RoleRulesgoAI, RoleMCPAgent, RoleComposeChat}
	return append(out, names...)
}
