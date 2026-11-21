package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	labelTypes "github.com/madnikulin50/lowcode/server/pkg/label/types"
	"github.com/madnikulin50/lowcode/server/pkg/sql"
)

type (
	Trigger struct {
		ID      uint64 `json:"triggerID,string" schema:"col=id,dal=id,unique"`
		Enabled bool   `json:"enabled" schema:"col=enabled,dal=bool:true,sortable"`

		WorkflowID uint64 `json:"workflowID,string" schema:"col=workflow_id,store=rel_workflow,dal=ref:corteza::automation:workflow,sortable"`
		// Start workflow on this step. If 0, find first (only) orphan
		StepID uint64 `json:"stepID,string" schema:"col=step_id,store=rel_step,dal=id"`

		// Resource type that can trigger the workflow
		ResourceType string `json:"resourceType" schema:"col=resource_type,dal=text:64,sortable"`

		// Event type that can trigger the workflow
		EventType string `json:"eventType" schema:"col=event_type,dal,sortable"`

		// Trigger constraints
		Constraints TriggerConstraintSet `json:"constraints" schema:"col=constraints,dal=json:empty,omit"`

		// Initial input scope,
		// will be merged merged with workflow variables
		Input *expr.Vars `json:"input" schema:"col=input,dal=json:empty,omit"`

		Labels map[string]labelTypes.LabelValue `json:"labels,omitempty"`
		Meta   *TriggerMeta                     `json:"meta,omitempty" schema:"col=meta,dal=json:empty,omit"`

		OwnedBy   uint64     `json:"ownedBy,string" schema:"col=owned_by,dal=userref"`
		CreatedAt time.Time  `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		CreatedBy uint64     `json:"createdBy,string" schema:"col=created_by,dal=userref"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty" schema:"col=updated_at,dal=timestamp:nil,sortable"`
		UpdatedBy uint64     `json:"updatedBy,string,omitempty" schema:"col=updated_by,dal=userref"`
		DeletedAt *time.Time `json:"deletedAt,omitempty" schema:"col=deleted_at,dal=timestamp:nil,sortable"`
		DeletedBy uint64     `json:"deletedBy,string,omitempty" schema:"col=deleted_by,dal=userref"`
	}

	TriggerConstraint struct {
		Name   string   `json:"name"`
		Op     string   `json:"op,omitempty"`
		Values []string `json:"values,omitempty"`
	}

	TriggerMeta struct {
		Description string                 `json:"description"`
		Visual      map[string]interface{} `json:"visual"`
	}

	TriggerFilter struct {
		TriggerID  []string `json:"triggerID"`
		WorkflowID []string `json:"workflowID"`

		EventType    string `json:"eventType"`
		ResourceType string `json:"resourceType"`

		Deleted  filter.State `json:"deleted"`
		Disabled filter.State `json:"disabled"`

		LabeledIDs []uint64                         `json:"-"`
		Labels     map[string]labelTypes.LabelValue `json:"labels,omitempty"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(*Trigger) (bool, error) `json:"-"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}
)

func ParseTriggerMeta(ss []string) (p *TriggerMeta, err error) {
	p = &TriggerMeta{}
	return p, parseStringsInput(ss, p)
}

func ParseTriggerConstraintSet(ss []string) (p TriggerConstraintSet, err error) {
	p = TriggerConstraintSet{}
	return p, parseStringsInput(ss, &p)
}

func (vv *TriggerConstraintSet) Scan(src any) error          { return sql.ParseJSON(src, vv) }
func (vv TriggerConstraintSet) Value() (driver.Value, error) { return json.Marshal(vv) }

func (vv *TriggerMeta) Scan(src any) error           { return sql.ParseJSON(src, vv) }
func (vv *TriggerMeta) Value() (driver.Value, error) { return json.Marshal(vv) }

func (set TriggerSet) FilterByWorkflowID(workflowID uint64) (vv TriggerSet) {
	// Make sure we never return nil
	vv = TriggerSet{}

	for i := range set {
		if set[i].WorkflowID == workflowID {
			vv = append(vv, set[i])
		}
	}

	return
}
