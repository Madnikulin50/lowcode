package envoyx

// Formerly generated from CUE; now maintained by hand.
//

import (
	"github.com/madnikulin50/lowcode/server/automation/types"
)

// AutomationWorkflowRbacReferences generates RBAC references
//
// Resources with "envoy: false" are skipped
//
// This function is auto-generated
func AutomationWorkflowRbacReferences(workflow string) (res *Ref, pp []*Ref, err error) {
	if workflow != "*" {
		res = &Ref{ResourceType: types.WorkflowResourceType, Identifiers: MakeIdentifiers(workflow)}
	}

	return
}
