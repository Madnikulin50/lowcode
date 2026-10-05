package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/sql"
)

type (
	// PromptVersion is one saved version of a named prompt in the prompt
	// library. Versions are immutable: editing a prompt saves a new version,
	// and exactly one version per handle is Active - the one a step or node
	// gets when it refers to the prompt as "@prompt:<handle>". Rolling back
	// is activating an older version.
	PromptVersion struct {
		ID uint64 `json:"promptVersionID,string" schema:"col=id,dal=id,unique"`

		Handle  string `json:"handle" schema:"col=handle,dal=text:64,sortable"`
		Version int    `json:"version" schema:"col=version,dal=number:default0,sortable"`

		Description string `json:"description,omitempty" schema:"col=description,dal"`
		Text        string `json:"text" schema:"col=text,dal"`
		Note        string `json:"note,omitempty" schema:"col=note,dal"`

		Active bool `json:"active" schema:"col=active,dal=bool:false,sortable"`

		// Kind tells a plain prompt from a skill; empty means a prompt. A skill
		// is a prompt that an agent loads on demand: Description says when to
		// use it, Text is the instructions, Requires names the toolkits it
		// needs and Resources are the reference files it ships with.
		Kind      string         `json:"kind,omitempty" schema:"col=kind,dal=text:16,sortable"`
		Requires  PromptRequires `json:"requires,omitempty" schema:"col=requires,dal=json:empty,omit"`
		Resources PromptFiles    `json:"resources,omitempty" schema:"col=resources,dal=json:empty,omit"`

		// Golden examples this prompt is checked against (see aiagent.RunEval)
		Cases PromptCases `json:"cases,omitempty" schema:"col=cases,dal=json:empty,omit"`

		CreatedAt time.Time `json:"createdAt,omitempty" schema:"col=created_at,dal=timestamp:now,sortable"`
		CreatedBy uint64    `json:"createdBy,string" schema:"col=created_by,dal=userref"`
	}

	// PromptCase is one golden example: the variables a prompt is run with
	// and what a good answer looks like.
	PromptCase struct {
		Name   string                 `json:"name"`
		Inputs map[string]interface{} `json:"inputs,omitempty"`

		// Expect: output field -> required value (compared as JSON)
		Expect map[string]interface{} `json:"expect,omitempty"`

		// Contains: phrases the answer must include (case-insensitive); for
		// prompts that answer in prose rather than JSON
		Contains []string `json:"contains,omitempty"`
	}

	PromptCases []PromptCase

	// PromptFile is a text file that ships with a skill (a price list, a
	// checklist, a template)
	PromptFile struct {
		Path    string `json:"path"`
		Mime    string `json:"mime,omitempty"`
		Content string `json:"content"`
	}

	PromptFiles []PromptFile

	// PromptRequires is the toolkits a skill needs
	PromptRequires []string

	PromptVersionFilter struct {
		PromptVersionID []string `json:"promptVersionID"`
		Handle          string   `json:"handle"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(*PromptVersion) (bool, error) `json:"-"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}
)

const (
	PromptKindPrompt = "prompt"
	PromptKindSkill  = "skill"
)

// IsSkill reports whether the version belongs to a skill; an empty kind is a
// plain prompt (rows saved before skills existed have none).
func (v *PromptVersion) IsSkill() bool { return v.Kind == PromptKindSkill }

func (vv *PromptRequires) Scan(src any) error          { return sql.ParseJSON(src, vv) }
func (vv PromptRequires) Value() (driver.Value, error) { return json.Marshal(vv) }
func (vv *PromptFiles) Scan(src any) error             { return sql.ParseJSON(src, vv) }
func (vv PromptFiles) Value() (driver.Value, error)    { return json.Marshal(vv) }

func (vv *PromptCases) Scan(src any) error          { return sql.ParseJSON(src, vv) }
func (vv PromptCases) Value() (driver.Value, error) { return json.Marshal(vv) }
