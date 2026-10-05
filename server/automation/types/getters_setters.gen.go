package types

// This file is auto-generated version 2.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from <no value>
//

import (
	"github.com/madnikulin50/lowcode/server/pkg/cast2"
)

func (r Workflow) GetID() uint64 { return r.ID }

func (r *Workflow) GetValue(name string, pos uint) (any, error) {
	if r == nil {
		return nil, nil
	}

	switch name {
	case "createdAt", "CreatedAt":
		return r.CreatedAt, nil
	case "createdBy", "CreatedBy":
		return r.CreatedBy, nil
	case "deletedAt", "DeletedAt":
		return r.DeletedAt, nil
	case "deletedBy", "DeletedBy":
		return r.DeletedBy, nil
	case "enabled", "Enabled":
		return r.Enabled, nil
	case "handle", "Handle":
		return r.Handle, nil
	case "id", "ID":
		return r.ID, nil
	case "keepSessions", "KeepSessions":
		return r.KeepSessions, nil
	case "ownedBy", "OwnedBy":
		return r.OwnedBy, nil
	case "runAs", "RunAs":
		return r.RunAs, nil
	case "trace", "Trace":
		return r.Trace, nil
	case "updatedAt", "UpdatedAt":
		return r.UpdatedAt, nil
	case "updatedBy", "UpdatedBy":
		return r.UpdatedBy, nil

	}
	return nil, nil
}

func (r *Workflow) SetValue(name string, pos uint, value any) (err error) {
	if r == nil {
		r = &Workflow{}
	}

	switch name {
	case "createdAt", "CreatedAt":
		return cast2.Time(value, &r.CreatedAt)
	case "createdBy", "CreatedBy":
		return cast2.Uint64(value, &r.CreatedBy)
	case "deletedAt", "DeletedAt":
		return cast2.TimePtr(value, &r.DeletedAt)
	case "deletedBy", "DeletedBy":
		return cast2.Uint64(value, &r.DeletedBy)
	case "enabled", "Enabled":
		return cast2.Bool(value, &r.Enabled)
	case "handle", "Handle":
		return cast2.String(value, &r.Handle)
	case "id", "ID":
		return cast2.Uint64(value, &r.ID)
	case "keepSessions", "KeepSessions":
		return cast2.Int(value, &r.KeepSessions)
	case "ownedBy", "OwnedBy":
		return cast2.Uint64(value, &r.OwnedBy)
	case "runAs", "RunAs":
		return cast2.Uint64(value, &r.RunAs)
	case "trace", "Trace":
		return cast2.Bool(value, &r.Trace)
	case "updatedAt", "UpdatedAt":
		return cast2.TimePtr(value, &r.UpdatedAt)
	case "updatedBy", "UpdatedBy":
		return cast2.Uint64(value, &r.UpdatedBy)

	}
	return nil
}

func (r Session) GetID() uint64 { return r.ID }

func (r *Session) GetValue(name string, pos uint) (any, error) {
	if r == nil {
		return nil, nil
	}

	switch name {
	case "completedAt", "CompletedAt":
		return r.CompletedAt, nil
	case "createdAt", "CreatedAt":
		return r.CreatedAt, nil
	case "createdBy", "CreatedBy":
		return r.CreatedBy, nil
	case "error", "Error":
		return r.Error, nil
	case "eventType", "EventType":
		return r.EventType, nil
	case "id", "ID":
		return r.ID, nil
	case "purgeAt", "PurgeAt":
		return r.PurgeAt, nil
	case "resourceType", "ResourceType":
		return r.ResourceType, nil
	case "suspendedAt", "SuspendedAt":
		return r.SuspendedAt, nil
	case "workflowID", "WorkflowID":
		return r.WorkflowID, nil

	}
	return nil, nil
}

func (r *Session) SetValue(name string, pos uint, value any) (err error) {
	if r == nil {
		r = &Session{}
	}

	switch name {
	case "completedAt", "CompletedAt":
		return cast2.TimePtr(value, &r.CompletedAt)
	case "createdAt", "CreatedAt":
		return cast2.Time(value, &r.CreatedAt)
	case "createdBy", "CreatedBy":
		return cast2.Uint64(value, &r.CreatedBy)
	case "error", "Error":
		return cast2.String(value, &r.Error)
	case "eventType", "EventType":
		return cast2.String(value, &r.EventType)
	case "id", "ID":
		return cast2.Uint64(value, &r.ID)
	case "purgeAt", "PurgeAt":
		return cast2.TimePtr(value, &r.PurgeAt)
	case "resourceType", "ResourceType":
		return cast2.String(value, &r.ResourceType)
	case "suspendedAt", "SuspendedAt":
		return cast2.TimePtr(value, &r.SuspendedAt)
	case "workflowID", "WorkflowID":
		return cast2.Uint64(value, &r.WorkflowID)

	}
	return nil
}

func (r State) GetID() uint64 { return r.ID }

func (r *State) GetValue(name string, pos uint) (any, error) {
	if r == nil {
		return nil, nil
	}

	switch name {
	case "createdAt", "CreatedAt":
		return r.CreatedAt, nil
	case "id", "ID":
		return r.ID, nil
	case "kind", "Kind":
		return r.Kind, nil
	case "resumeAt", "ResumeAt":
		return r.ResumeAt, nil
	case "sessionID", "SessionID":
		return r.SessionID, nil
	case "workflowID", "WorkflowID":
		return r.WorkflowID, nil

	}
	return nil, nil
}

func (r *State) SetValue(name string, pos uint, value any) (err error) {
	if r == nil {
		r = &State{}
	}

	switch name {
	case "createdAt", "CreatedAt":
		return cast2.Time(value, &r.CreatedAt)
	case "id", "ID":
		return cast2.Uint64(value, &r.ID)
	case "kind", "Kind":
		return cast2.String(value, &r.Kind)
	case "resumeAt", "ResumeAt":
		return cast2.TimePtr(value, &r.ResumeAt)
	case "sessionID", "SessionID":
		return cast2.Uint64(value, &r.SessionID)
	case "workflowID", "WorkflowID":
		return cast2.Uint64(value, &r.WorkflowID)

	}
	return nil
}

func (r PromptVersion) GetID() uint64 { return r.ID }

func (r *PromptVersion) GetValue(name string, pos uint) (any, error) {
	if r == nil {
		return nil, nil
	}

	switch name {
	case "active", "Active":
		return r.Active, nil
	case "createdAt", "CreatedAt":
		return r.CreatedAt, nil
	case "createdBy", "CreatedBy":
		return r.CreatedBy, nil
	case "description", "Description":
		return r.Description, nil
	case "handle", "Handle":
		return r.Handle, nil
	case "id", "ID":
		return r.ID, nil
	case "kind", "Kind":
		return r.Kind, nil
	case "note", "Note":
		return r.Note, nil
	case "text", "Text":
		return r.Text, nil
	case "version", "Version":
		return r.Version, nil

	}
	return nil, nil
}

func (r *PromptVersion) SetValue(name string, pos uint, value any) (err error) {
	if r == nil {
		r = &PromptVersion{}
	}

	switch name {
	case "active", "Active":
		return cast2.Bool(value, &r.Active)
	case "createdAt", "CreatedAt":
		return cast2.Time(value, &r.CreatedAt)
	case "createdBy", "CreatedBy":
		return cast2.Uint64(value, &r.CreatedBy)
	case "description", "Description":
		return cast2.String(value, &r.Description)
	case "handle", "Handle":
		return cast2.String(value, &r.Handle)
	case "id", "ID":
		return cast2.Uint64(value, &r.ID)
	case "kind", "Kind":
		return cast2.String(value, &r.Kind)
	case "note", "Note":
		return cast2.String(value, &r.Note)
	case "text", "Text":
		return cast2.String(value, &r.Text)
	case "version", "Version":
		return cast2.Int(value, &r.Version)

	}
	return nil
}

func (r Trigger) GetID() uint64 { return r.ID }

func (r *Trigger) GetValue(name string, pos uint) (any, error) {
	if r == nil {
		return nil, nil
	}

	switch name {
	case "createdAt", "CreatedAt":
		return r.CreatedAt, nil
	case "createdBy", "CreatedBy":
		return r.CreatedBy, nil
	case "deletedAt", "DeletedAt":
		return r.DeletedAt, nil
	case "deletedBy", "DeletedBy":
		return r.DeletedBy, nil
	case "enabled", "Enabled":
		return r.Enabled, nil
	case "eventType", "EventType":
		return r.EventType, nil
	case "id", "ID":
		return r.ID, nil
	case "ownedBy", "OwnedBy":
		return r.OwnedBy, nil
	case "resourceType", "ResourceType":
		return r.ResourceType, nil
	case "stepID", "StepID":
		return r.StepID, nil
	case "updatedAt", "UpdatedAt":
		return r.UpdatedAt, nil
	case "updatedBy", "UpdatedBy":
		return r.UpdatedBy, nil
	case "workflowID", "WorkflowID":
		return r.WorkflowID, nil

	}
	return nil, nil
}

func (r *Trigger) SetValue(name string, pos uint, value any) (err error) {
	if r == nil {
		r = &Trigger{}
	}

	switch name {
	case "createdAt", "CreatedAt":
		return cast2.Time(value, &r.CreatedAt)
	case "createdBy", "CreatedBy":
		return cast2.Uint64(value, &r.CreatedBy)
	case "deletedAt", "DeletedAt":
		return cast2.TimePtr(value, &r.DeletedAt)
	case "deletedBy", "DeletedBy":
		return cast2.Uint64(value, &r.DeletedBy)
	case "enabled", "Enabled":
		return cast2.Bool(value, &r.Enabled)
	case "eventType", "EventType":
		return cast2.String(value, &r.EventType)
	case "id", "ID":
		return cast2.Uint64(value, &r.ID)
	case "ownedBy", "OwnedBy":
		return cast2.Uint64(value, &r.OwnedBy)
	case "resourceType", "ResourceType":
		return cast2.String(value, &r.ResourceType)
	case "stepID", "StepID":
		return cast2.Uint64(value, &r.StepID)
	case "updatedAt", "UpdatedAt":
		return cast2.TimePtr(value, &r.UpdatedAt)
	case "updatedBy", "UpdatedBy":
		return cast2.Uint64(value, &r.UpdatedBy)
	case "workflowID", "WorkflowID":
		return cast2.Uint64(value, &r.WorkflowID)

	}
	return nil
}
