// Package riskengine defines the domain model and calculation strategies for
// the platform's universal risk engine: a reusable, namespace-scoped factor
// library, versioned risk models (weighted / matrix / bayes / blend), and a
// SHAP-style attribution layer used to explain assessments.
//
// This package holds only types and pure calculation logic. Persistence,
// namespace/permission enforcement and REST wiring live in
// server/compose/rest/risk_*.go — the same split server/pkg/rulesgo and
// server/compose/rest/rulechain_admin.go already use for rule chains.
//
// Decisions this package encodes (see the architecture note for the "why"):
//   - RiskFactorDef and RiskModel are scoped to a namespace, the same
//     boundary types.Module.NamespaceID and rulesgo.Chain.NamespaceID use.
//   - A "states" factor's states are arbitrary per factor, but every state
//     carries a Severity so the UI can color/sort nodes with unrelated
//     vocabularies consistently (RiskFactorState.Severity).
//   - Attribution is a separate step after Evaluate, not part of it: Explain
//     reads an Assessment's own output, never raw Evidence directly, so an
//     AI explanation layer built on top of it narrates already-computed
//     numbers instead of reasoning about risk from scratch.
//   - RiskAssessment / RiskTreatment are the *shape* written into records of
//     an ordinary Compose module (via RiskSubjectBinding) — they are not
//     stored by this package itself, so history, access control and
//     dashboards come from Compose for free.
package riskengine

import (
	"encoding/json"
	"time"
)

// ---------------------------------------------------------------------------
// Factor library
// ---------------------------------------------------------------------------

// ScaleType is how a factor's raw value is interpreted.
type ScaleType string

const (
	ScaleNumeric ScaleType = "numeric" // continuous/discrete number, normalized against Max
	ScaleBool    ScaleType = "bool"    // true/false, coerced to 0/1
	ScaleStates  ScaleType = "states"  // named discrete states — required for the bayes strategy
)

// SourceKind is where a factor's value comes from at evaluation time.
type SourceKind string

const (
	SourceField    SourceKind = "field"    // a field on the record bound via RiskSubjectBinding
	SourceExpr     SourceKind = "expr"     // an expression over the record/evidence (reuses rulesgo template resolution)
	SourceManual   SourceKind = "manual"   // entered by a person at assessment time
	SourceExternal SourceKind = "external" // supplied by an external system/integration call
)

// Severity is the fixed, small vocabulary used to interpret arbitrary
// per-factor states in the UI — graph node color, CPT-grid chips, portfolio
// sort. It is display-only and never enters a calculation: two bayes nodes
// with completely different state names still render with the same four
// colors. It mirrors the bands rulesgo's risk.band executor already
// produces (server/pkg/rulesgo/score.go).
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// RiskFactorState is one named, orderable state of a "states"-scale factor.
// States are defined per factor — there is no platform-wide low/medium/high
// enum — but every state must carry a Severity. Example: a "daysSinceAudit"
// factor might use recent/aging/overdue -> low/medium/high, while
// "staffTurnover" uses stable/elevated/exodus -> low/medium/critical; both
// render with the same severity palette despite sharing no state codes.
type RiskFactorState struct {
	Code     string   `json:"code"`  // stable id used in CPTs and evidence, e.g. "overdue"
	Label    string   `json:"label"` // human label, e.g. "Просрочено"
	Severity Severity `json:"severity"`
	Order    int      `json:"order"` // display / CPT-column order, ascending severity by convention
}

// RiskFactorDef is a reusable factor definition, scoped to a namespace.
type RiskFactorDef struct {
	ID          uint64 `json:"factorID,string"`
	NamespaceID uint64 `json:"namespaceID,string"`
	Handle      string `json:"handle"` // e.g. "daysSinceAudit" — the evidence key
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`

	Scale ScaleType `json:"scale"`

	// Numeric/bool scale only.
	Max    float64 `json:"max,omitempty"`
	Invert bool    `json:"invert,omitempty"` // true when a higher raw value means *lower* risk

	// States scale only — required when this factor is placed as a bayes node.
	States []RiskFactorState `json:"states,omitempty"`

	Source SourceKind `json:"source"`
	// FieldName / Expr apply only to Source == SourceField / SourceExpr.
	FieldName string `json:"fieldName,omitempty"`
	Expr      string `json:"expr,omitempty"`

	// Baseline is the reference value attribution centers against
	// (score - baseline): a raw number for "numeric"/"bool", a state Code
	// for "states". Fixed per published version so historical assessments
	// stay comparable across the portfolio.
	Baseline string `json:"baseline,omitempty"`

	CreatedAt time.Time  `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// ---------------------------------------------------------------------------
// Risk model graph
// ---------------------------------------------------------------------------

// Strategy selects how a RiskModel turns evidence into a score.
type Strategy string

const (
	StrategyWeighted Strategy = "weighted"
	StrategyMatrix   Strategy = "matrix"
	StrategyBayes    Strategy = "bayes"
	StrategyBlend    Strategy = "blend"
)

// ModelStatus tracks the draft/published lifecycle, same shape as rulechain
// admin's chain versioning.
type ModelStatus string

const (
	StatusDraft     ModelStatus = "draft"
	StatusPublished ModelStatus = "published"
	StatusArchived  ModelStatus = "archived"
)

// RiskModel is a versioned graph of factors for a class of objects.
type RiskModel struct {
	ID          uint64      `json:"modelID,string"`
	NamespaceID uint64      `json:"namespaceID,string"`
	Handle      string      `json:"handle"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Version     int         `json:"version"`
	Status      ModelStatus `json:"status"`

	Strategy Strategy `json:"strategy"`
	// Config holds one of WeightedConfig / MatrixConfig / BayesConfig /
	// BlendConfig below, chosen by Strategy — same json.RawMessage +
	// per-type-registry pattern rulesgo.ChainNode.Config already uses.
	Config json.RawMessage `json:"config"`

	Nodes []RiskNode `json:"nodes"`
	Edges []RiskEdge `json:"edges"`

	// AttributionMaxExact caps exact 2^n Shapley enumeration for the bayes
	// strategy before Explain falls back to permutation sampling. Default 15
	// when zero.
	AttributionMaxExact int `json:"attributionMaxExact,omitempty"`

	// Bands is this model's own score -> Severity calibration (see Band in
	// engine.go). EvaluateWithOptions uses it whenever a call doesn't pass
	// its own EvalOptions.Bands, so the same evidence always classifies to
	// the same level regardless of which caller (REST, UI, trigger) is
	// asking — a threshold calibrated once when the model is published,
	// not re-guessed per request. Falls back to DefaultBands() when both
	// are empty.
	Bands []Band `json:"bands,omitempty"`

	CreatedAt   time.Time  `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
}

// NodeRole distinguishes leaf evidence from computed nodes on the graph.
type NodeRole string

const (
	RoleEvidence     NodeRole = "evidence"     // a leaf factor fed from RiskSubjectBinding
	RoleIntermediate NodeRole = "intermediate" // a bayes node with parents, itself feeding further nodes
	RoleOutput       NodeRole = "output"       // the model's single final risk node
)

// RiskNode places a RiskFactorDef on the model graph. X/Y are an editor
// concern (canvas position) the engine itself ignores.
//
// CPT belongs to the node, not to an individual edge: a bayes node's
// probability table is a single joint function of *all* its parents
// together (see CPT.ParentStates), so it lives where "all parents" is a
// well-defined set — the child node. Only nodes with Role
// RoleIntermediate/RoleOutput under the "bayes"/"blend" strategy set it; a
// node with no incoming edges (Role RoleEvidence, or a root of the network)
// leaves it nil.
type RiskNode struct {
	ID       string   `json:"id"` // graph-local id, stable across versions for diffing
	FactorID uint64   `json:"factorID,string"`
	Label    string   `json:"label,omitempty"` // overrides RiskFactorDef.Label for this model
	Role     NodeRole `json:"role"`
	CPT      *CPT     `json:"cpt,omitempty"`
	X        float64  `json:"x"`
	Y        float64  `json:"y"`
}

// RiskEdge is a directed influence between two nodes. Weight is used by the
// weighted strategy; for the bayes strategy an edge only declares "From is a
// parent of To" — the joint CPT itself lives on the To node (see
// RiskNode.CPT), and parent order within it follows the order edges into
// that node appear in RiskModel.Edges.
type RiskEdge struct {
	From string `json:"from"` // RiskNode.ID

	To string `json:"to"` // RiskNode.ID

	Weight float64 `json:"weight,omitempty"` // weighted strategy
}

// CPT is a conditional probability table for one bayes node, indexed by the
// joint state of its parents (parent order = the order incoming RiskEdges
// into the owning node appear in RiskModel.Edges). Every row must sum to 1
// across NodeStates.
type CPT struct {
	// ParentStates[i] lists the state codes of parent i, in edge order.
	ParentStates [][]string `json:"parentStates"`
	// NodeStates lists this node's own state codes, matching its
	// RiskFactorDef.States order.
	NodeStates []string `json:"nodeStates"`
	// Rows has len(cartesian product of ParentStates) entries, each with
	// len(NodeStates) probabilities. Row order is the standard odometer
	// order over ParentStates (the last parent varies fastest).
	Rows [][]float64 `json:"rows"`
}

// ---------------------------------------------------------------------------
// Strategy configs (RiskModel.Config, one per Strategy)
// ---------------------------------------------------------------------------

// WeightedConfig configures the "weighted" strategy. Per-factor weights live
// on the RiskEdge into the output node; this only carries the normalization
// shared across all of them — mirrors rulesgo's score.weighted node
// (server/pkg/rulesgo/score.go).
type WeightedConfig struct {
	Normalize *bool   `json:"normalize,omitempty"` // default true (nil means true, same convention as rulesgo's score.weighted)
	ScaleMax  float64 `json:"scaleMax,omitempty"`  // default 100 when normalizing
}

// MatrixConfig configures the "matrix" strategy — likelihood x impact, or a
// custom cell lookup — mirrors rulesgo's score.matrix node.
type MatrixConfig struct {
	LikelihoodNode string      `json:"likelihoodNode"` // RiskNode.ID with Role == RoleEvidence
	ImpactNode     string      `json:"impactNode"`
	ScaleSize      int         `json:"scaleSize,omitempty"` // default 5
	Formula        string      `json:"formula,omitempty"`   // "product" (default) | "sum"
	Matrix         [][]float64 `json:"matrix,omitempty"`    // optional custom cells [L-1][I-1]
}

// BayesConfig configures the "bayes" strategy.
type BayesConfig struct {
	OutputNode string `json:"outputNode"` // RiskNode.ID with Role == RoleOutput
	// ScoreByState maps the output node's state Code to a score-scale
	// number, so a posterior distribution can still produce a point score
	// (expected value) alongside the full distribution.
	ScoreByState map[string]float64 `json:"scoreByState"`
}

// BlendConfig combines two or more component strategies evaluated over the
// *same* Nodes/Edges. Each component is a full, independent view of that
// graph — its own Strategy and Config (a WeightedConfig or BayesConfig,
// exactly as if it were a whole RiskModel) — so Evaluate/Explain simply
// re-run themselves per component against a virtual model that borrows the
// blend's Nodes/Edges. Because Shapley values are additive, Explain never
// recomputes attribution for the blend itself — it combines the components'
// own AttributionResults with the same weights.
type BlendConfig struct {
	Components []BlendComponent `json:"components"`
}

type BlendComponent struct {
	Strategy Strategy `json:"strategy"` // "weighted" | "bayes" (not "blend" — no nesting)
	Weight   float64  `json:"weight"`   // combination weight; normalized at evaluation time
	// Config is this component's own WeightedConfig or BayesConfig.
	Config json.RawMessage `json:"config"`
}

// ---------------------------------------------------------------------------
// Binding to Compose data
// ---------------------------------------------------------------------------

// RiskSubjectBinding attaches a RiskModel to the records of one Compose
// module — the only place domain specifics ("this is a store", "this is a
// vendor") enter the otherwise domain-agnostic engine.
type RiskSubjectBinding struct {
	ID          uint64 `json:"bindingID,string"`
	NamespaceID uint64 `json:"namespaceID,string"`
	ModelID     uint64 `json:"modelID,string"`
	ModuleID    uint64 `json:"moduleID,string"` // the subject module, e.g. "Stores"

	// FieldMap maps RiskFactorDef.Handle -> module field name, for factors
	// whose Source is SourceField. Factors with other Source kinds don't
	// appear here.
	FieldMap map[string]string `json:"fieldMap"`

	// AssessmentModuleID is the module (usually auto-provisioned on first
	// publish) that stores RiskAssessment records for this binding.
	AssessmentModuleID uint64 `json:"assessmentModuleID,string"`
	// TreatmentModuleID similarly stores RiskTreatment records; optional —
	// a binding can skip mitigation tracking.
	TreatmentModuleID uint64 `json:"treatmentModuleID,string,omitempty"`

	// AutoRecalc, when true, reuses the existing automation trigger
	// infrastructure to reassess on every create/update of the subject
	// record instead of requiring a manual recalculation.
	AutoRecalc bool `json:"autoRecalc"`

	CreatedAt time.Time  `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// ---------------------------------------------------------------------------
// Evaluation results
// ---------------------------------------------------------------------------

// Evidence is the resolved input to Evaluate: RiskFactorDef.Handle -> raw
// value (float64 for "numeric"/"bool", a state Code string for "states").
type Evidence map[string]interface{}

// RiskFactorObservation is the snapshot of one factor's value for one
// subject record at assessment time.
type RiskFactorObservation struct {
	FactorHandle string      `json:"factorHandle"`
	NodeID       string      `json:"nodeID"`
	Raw          interface{} `json:"raw"`                  // the value as read or entered
	Normalized   float64     `json:"normalized,omitempty"` // 0..1, numeric/bool scales
	StateCode    string      `json:"stateCode,omitempty"`  // states scale
}

// FactorContribution is one line of an Attribution breakdown, in score
// units — see server/pkg/riskengine (Explain) for how it's computed per
// Strategy.
type FactorContribution struct {
	NodeID   string  `json:"nodeID"`
	Handle   string  `json:"factorHandle"`
	Value    float64 `json:"value"`    // this factor's Shapley value
	Baseline float64 `json:"baseline"` // the reference value used for it
}

// AttributionResult is the SHAP-style decomposition of one Assessment.
// Contributions must sum to InherentScore - Baseline.
type AttributionResult struct {
	Method        string               `json:"method"` // "closed-form" | "exact" | "sampled"
	Baseline      float64              `json:"baseline"`
	Contributions []FactorContribution `json:"contributions"`
	SampleCount   int                  `json:"sampleCount,omitempty"` // set when Method == "sampled"
}

// RiskAssessment is one calculation result. It is the shape written into and
// read from records of RiskSubjectBinding.AssessmentModuleID — this package
// does not persist it itself.
type RiskAssessment struct {
	// ID and BindingID are only meaningful once an assessment is persisted
	// (in Compose records once RiskSubjectBinding.AssessmentModuleID is
	// wired up; in the interim in-memory store — see server/pkg/riskstore —
	// they identify it the same way). Evaluate itself leaves them zero.
	ID              uint64 `json:"assessmentID,string,omitempty"`
	BindingID       uint64 `json:"bindingID,string,omitempty"`
	SubjectRecordID uint64 `json:"subjectRecordID,string"`
	ModelID         uint64 `json:"modelID,string"`
	ModelVersion    int    `json:"modelVersion"`

	InherentScore        float64  `json:"inherentScore"`
	ControlEffectiveness float64  `json:"controlEffectiveness,omitempty"`
	ResidualScore        float64  `json:"residualScore"`
	Level                Severity `json:"level"`

	Observations []RiskFactorObservation `json:"observations"`
	Attribution  *AttributionResult      `json:"attribution,omitempty"`
	// Explanation is the Layer-C narrative, generated from Attribution —
	// never from Observations directly.
	Explanation string `json:"explanation,omitempty"`

	OverriddenBy    string   `json:"overriddenBy,omitempty"`
	OverrideComment string   `json:"overrideComment,omitempty"`
	OverrideScore   *float64 `json:"overrideScore,omitempty"`

	AssessedAt time.Time `json:"assessedAt"`
}

// TreatmentStatus tracks a mitigation task's lifecycle.
type TreatmentStatus string

const (
	TreatmentOpen       TreatmentStatus = "open"
	TreatmentInProgress TreatmentStatus = "in_progress"
	TreatmentDone       TreatmentStatus = "done"
	TreatmentCancelled  TreatmentStatus = "cancelled"
)

// RiskTreatment is a mitigation task, persisted in
// RiskSubjectBinding.TreatmentModuleID.
type RiskTreatment struct {
	ID              uint64          `json:"treatmentID,string,omitempty"`
	BindingID       uint64          `json:"bindingID,string,omitempty"`
	SubjectRecordID uint64          `json:"subjectRecordID,string"`
	AssessedAt      time.Time       `json:"assessedAt"` // which assessment triggered it
	OwnerUserID     uint64          `json:"ownerUserID,string"`
	Description     string          `json:"description"`
	DueAt           *time.Time      `json:"dueAt,omitempty"`
	Status          TreatmentStatus `json:"status"`
}
