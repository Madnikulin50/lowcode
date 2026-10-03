package types

// This file is auto-generated version 2.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from <no value>
//

import (
	"github.com/madnikulin50/lowcode/server/pkg/cast2"
)

func (r Rule) GetID() uint64 { return r.ID }

func (r *Rule) GetValue(name string, pos uint) (any, error) {
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
	case "detector", "Detector":
		return r.Detector, nil
	case "enabled", "Enabled":
		return r.Enabled, nil
	case "field", "Field":
		return r.Field, nil
	case "id", "ID":
		return r.ID, nil
	case "lastScannedAt", "LastScannedAt":
		return r.LastScannedAt, nil
	case "moduleID", "ModuleID":
		return r.ModuleID, nil
	case "namespaceID", "NamespaceID":
		return r.NamespaceID, nil
	case "threshold", "Threshold":
		return r.Threshold, nil
	case "updatedAt", "UpdatedAt":
		return r.UpdatedAt, nil
	case "updatedBy", "UpdatedBy":
		return r.UpdatedBy, nil

	}
	return nil, nil
}

func (r *Rule) SetValue(name string, pos uint, value any) (err error) {
	if r == nil {
		r = &Rule{}
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
	case "detector", "Detector":
		return cast2.String(value, &r.Detector)
	case "enabled", "Enabled":
		return cast2.Bool(value, &r.Enabled)
	case "field", "Field":
		return cast2.String(value, &r.Field)
	case "id", "ID":
		return cast2.Uint64(value, &r.ID)
	case "lastScannedAt", "LastScannedAt":
		return cast2.TimePtr(value, &r.LastScannedAt)
	case "moduleID", "ModuleID":
		return cast2.Uint64(value, &r.ModuleID)
	case "namespaceID", "NamespaceID":
		return cast2.Uint64(value, &r.NamespaceID)
	case "threshold", "Threshold":
		return cast2.Float64(value, &r.Threshold)
	case "updatedAt", "UpdatedAt":
		return cast2.TimePtr(value, &r.UpdatedAt)
	case "updatedBy", "UpdatedBy":
		return cast2.Uint64(value, &r.UpdatedBy)

	}
	return nil
}

func (r Baseline) GetID() uint64 { return r.ID }

func (r *Baseline) GetValue(name string, pos uint) (any, error) {
	if r == nil {
		return nil, nil
	}

	switch name {
	case "count", "Count":
		return r.Count, nil
	case "field", "Field":
		return r.Field, nil
	case "id", "ID":
		return r.ID, nil
	case "moduleID", "ModuleID":
		return r.ModuleID, nil
	case "namespaceID", "NamespaceID":
		return r.NamespaceID, nil
	case "ruleID", "RuleID":
		return r.RuleID, nil
	case "updatedAt", "UpdatedAt":
		return r.UpdatedAt, nil

	}
	return nil, nil
}

func (r *Baseline) SetValue(name string, pos uint, value any) (err error) {
	if r == nil {
		r = &Baseline{}
	}

	switch name {
	case "count", "Count":
		return cast2.Uint64(value, &r.Count)
	case "field", "Field":
		return cast2.String(value, &r.Field)
	case "id", "ID":
		return cast2.Uint64(value, &r.ID)
	case "moduleID", "ModuleID":
		return cast2.Uint64(value, &r.ModuleID)
	case "namespaceID", "NamespaceID":
		return cast2.Uint64(value, &r.NamespaceID)
	case "ruleID", "RuleID":
		return cast2.Uint64(value, &r.RuleID)
	case "updatedAt", "UpdatedAt":
		return cast2.TimePtr(value, &r.UpdatedAt)

	}
	return nil
}

func (r Finding) GetID() uint64 { return r.ID }

func (r *Finding) GetValue(name string, pos uint) (any, error) {
	if r == nil {
		return nil, nil
	}

	switch name {
	case "createdAt", "CreatedAt":
		return r.CreatedAt, nil
	case "id", "ID":
		return r.ID, nil
	case "moduleID", "ModuleID":
		return r.ModuleID, nil
	case "namespaceID", "NamespaceID":
		return r.NamespaceID, nil
	case "recordID", "RecordID":
		return r.RecordID, nil
	case "ruleID", "RuleID":
		return r.RuleID, nil
	case "score", "Score":
		return r.Score, nil
	case "severity", "Severity":
		return r.Severity, nil
	case "status", "Status":
		return r.Status, nil
	case "updatedAt", "UpdatedAt":
		return r.UpdatedAt, nil

	}
	return nil, nil
}

func (r *Finding) SetValue(name string, pos uint, value any) (err error) {
	if r == nil {
		r = &Finding{}
	}

	switch name {
	case "createdAt", "CreatedAt":
		return cast2.Time(value, &r.CreatedAt)
	case "id", "ID":
		return cast2.Uint64(value, &r.ID)
	case "moduleID", "ModuleID":
		return cast2.Uint64(value, &r.ModuleID)
	case "namespaceID", "NamespaceID":
		return cast2.Uint64(value, &r.NamespaceID)
	case "recordID", "RecordID":
		return cast2.Uint64(value, &r.RecordID)
	case "ruleID", "RuleID":
		return cast2.Uint64(value, &r.RuleID)
	case "score", "Score":
		return cast2.Float64(value, &r.Score)
	case "severity", "Severity":
		return cast2.String(value, &r.Severity)
	case "status", "Status":
		return cast2.String(value, &r.Status)
	case "updatedAt", "UpdatedAt":
		return cast2.TimePtr(value, &r.UpdatedAt)

	}
	return nil
}
