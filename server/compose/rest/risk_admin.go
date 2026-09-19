package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/riskengine"
	"github.com/madnikulin50/lowcode/server/pkg/riskstore"
)

// RiskAdmin exposes CRUD for the risk-factor library, risk models, subject
// bindings, assessments and treatments, plus the Test/Explain/Suggest
// endpoints the Risk Model Editor calls. Storage is riskstore's in-memory
// maps — the same starting point rulesgo.MemoryPersistence was for rule
// chains — real persistence (namespace-scoped tables, and writing
// RiskAssessment/RiskTreatment into the bound Compose module's own records)
// is the natural next step once the shape has been exercised by the editor.
type RiskAdmin struct{}

func MountRiskAdminRoutes(r chi.Router) {
	admin := RiskAdmin{}
	r.Route("/admin/risk/factors", func(r chi.Router) {
		r.Get("/", admin.ListFactors)
		r.Post("/", admin.CreateFactor)
		r.Get("/{factorID}", admin.GetFactor)
		r.Put("/{factorID}", admin.UpdateFactor)
		r.Delete("/{factorID}", admin.DeleteFactor)
	})
	r.Route("/admin/risk/models", func(r chi.Router) {
		r.Get("/", admin.ListModels)
		r.Post("/", admin.CreateModel)
		r.Get("/{modelID}", admin.GetModel)
		r.Put("/{modelID}", admin.UpdateModel)
		r.Delete("/{modelID}", admin.DeleteModel)
		r.Post("/{modelID}/test", admin.Test)
		r.Post("/{modelID}/explain", admin.ExplainNarrative)
	})
	r.Post("/admin/risk/suggest-factors", admin.SuggestFactors)
	r.Route("/admin/risk/bindings", func(r chi.Router) {
		r.Get("/", admin.ListBindings)
		r.Post("/", admin.CreateBinding)
		r.Get("/{bindingID}", admin.GetBinding)
		r.Put("/{bindingID}", admin.UpdateBinding)
		r.Delete("/{bindingID}", admin.DeleteBinding)
		r.Post("/{bindingID}/assess", admin.Assess)
		r.Get("/{bindingID}/assessments", admin.ListAssessments)
		r.Get("/{bindingID}/summary", admin.PortfolioSummary)
		r.Get("/{bindingID}/treatments", admin.ListTreatments)
		r.Post("/{bindingID}/treatments", admin.CreateTreatment)
	})
	r.Route("/admin/risk/treatments", func(r chi.Router) {
		r.Put("/{treatmentID}", admin.UpdateTreatment)
		r.Delete("/{treatmentID}", admin.DeleteTreatment)
	})
}

// ---------------------------------------------------------------------------
// factors
// ---------------------------------------------------------------------------

func (a RiskAdmin) ListFactors(w http.ResponseWriter, r *http.Request) {
	nsFilter := parseUint64String(r.URL.Query().Get("namespaceID"))
	all := riskstore.ListFactors(nsFilter)

	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)
	total := len(all)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	api.Send(w, r, map[string]interface{}{
		"factors": all[offset:end],
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (a RiskAdmin) GetFactor(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "factorID"))
	f, ok := riskstore.GetFactor(id)
	if !ok {
		api.Send(w, r, fmt.Errorf("factor not found"))
		return
	}
	api.Send(w, r, map[string]interface{}{"factor": f})
}

func (a RiskAdmin) CreateFactor(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	var f riskengine.RiskFactorDef
	if err := json.Unmarshal(body, &f); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	if f.Handle == "" {
		api.Send(w, r, fmt.Errorf("handle is required"))
		return
	}
	if f.NamespaceID == 0 {
		api.Send(w, r, fmt.Errorf("namespaceID is required"))
		return
	}
	if err := validateFactor(&f); err != nil {
		api.Send(w, r, err)
		return
	}

	f.ID = riskstore.NextID()
	f.CreatedAt = time.Now()
	riskstore.SaveFactor(&f)

	api.Send(w, r, map[string]interface{}{"created": true, "factor": f})
}

func (a RiskAdmin) UpdateFactor(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "factorID"))
	existing, ok := riskstore.GetFactor(id)
	if !ok {
		api.Send(w, r, fmt.Errorf("factor not found"))
		return
	}

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var patch riskengine.RiskFactorDef
	if err := json.Unmarshal(body, &patch); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	if err := validateFactor(&patch); err != nil {
		api.Send(w, r, err)
		return
	}

	patch.ID = id
	patch.NamespaceID = existing.NamespaceID // namespace is fixed at creation
	patch.CreatedAt = existing.CreatedAt
	now := time.Now()
	patch.UpdatedAt = &now
	riskstore.SaveFactor(&patch)

	api.Send(w, r, map[string]interface{}{"updated": true, "factor": patch})
}

func (a RiskAdmin) DeleteFactor(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "factorID"))
	if !riskstore.DeleteFactor(id) {
		api.Send(w, r, fmt.Errorf("factor not found"))
		return
	}
	api.Send(w, r, map[string]interface{}{"deleted": true})
}

func validateFactor(f *riskengine.RiskFactorDef) error {
	switch f.Scale {
	case riskengine.ScaleNumeric:
		if f.Max <= 0 {
			return fmt.Errorf("numeric factor %q: max must be > 0", f.Handle)
		}
	case riskengine.ScaleBool:
		// no extra fields required
	case riskengine.ScaleStates:
		if len(f.States) == 0 {
			return fmt.Errorf("states factor %q: at least one state is required", f.Handle)
		}
		for _, s := range f.States {
			if s.Code == "" {
				return fmt.Errorf("states factor %q: every state needs a code", f.Handle)
			}
			switch s.Severity {
			case riskengine.SeverityLow, riskengine.SeverityMedium, riskengine.SeverityHigh, riskengine.SeverityCritical:
			default:
				return fmt.Errorf("states factor %q: state %q has invalid severity %q", f.Handle, s.Code, s.Severity)
			}
		}
	default:
		return fmt.Errorf("factor %q: unknown scale %q", f.Handle, f.Scale)
	}
	return nil
}

// ---------------------------------------------------------------------------
// models
// ---------------------------------------------------------------------------

func (a RiskAdmin) ListModels(w http.ResponseWriter, r *http.Request) {
	nsFilter := parseUint64String(r.URL.Query().Get("namespaceID"))
	all := riskstore.ListModels(nsFilter)

	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)
	total := len(all)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	result := make([]map[string]interface{}, 0, end-offset)
	for _, m := range all[offset:end] {
		result = append(result, map[string]interface{}{
			"modelID":     m.ID,
			"namespaceID": m.NamespaceID,
			"handle":      m.Handle,
			"name":        m.Name,
			"strategy":    m.Strategy,
			"status":      m.Status,
			"version":     m.Version,
			"nodeCount":   len(m.Nodes),
			"edgeCount":   len(m.Edges),
		})
	}

	api.Send(w, r, map[string]interface{}{
		"models": result,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (a RiskAdmin) GetModel(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "modelID"))
	m, ok := riskstore.GetModel(id)
	if !ok {
		api.Send(w, r, fmt.Errorf("model not found"))
		return
	}
	resp := map[string]interface{}{"model": m}
	if r.URL.Query().Get("withFactors") != "" {
		resp["factors"] = riskstore.FactorsForModel(m)
	}
	api.Send(w, r, resp)
}

func (a RiskAdmin) CreateModel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	var m riskengine.RiskModel
	if err := json.Unmarshal(body, &m); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	if m.Name == "" {
		api.Send(w, r, fmt.Errorf("name is required"))
		return
	}
	if m.NamespaceID == 0 {
		api.Send(w, r, fmt.Errorf("namespaceID is required"))
		return
	}
	if m.Status == "" {
		m.Status = riskengine.StatusDraft
	}
	if m.Version == 0 {
		m.Version = 1
	}

	m.ID = riskstore.NextID()
	m.CreatedAt = time.Now()
	riskstore.SaveModel(&m)

	api.Send(w, r, map[string]interface{}{"created": true, "model": m})
}

func (a RiskAdmin) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "modelID"))
	existing, ok := riskstore.GetModel(id)
	if !ok {
		api.Send(w, r, fmt.Errorf("model not found"))
		return
	}

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var patch riskengine.RiskModel
	if err := json.Unmarshal(body, &patch); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	patch.ID = id
	patch.NamespaceID = existing.NamespaceID
	patch.CreatedAt = existing.CreatedAt
	if patch.Status == riskengine.StatusPublished && existing.Status != riskengine.StatusPublished {
		patch.Version = existing.Version + 1
		now := time.Now()
		patch.PublishedAt = &now
	} else if patch.Version == 0 {
		patch.Version = existing.Version
	}
	now := time.Now()
	patch.UpdatedAt = &now
	riskstore.SaveModel(&patch)

	api.Send(w, r, map[string]interface{}{"updated": true, "model": patch})
}

func (a RiskAdmin) DeleteModel(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "modelID"))
	if !riskstore.DeleteModel(id) {
		api.Send(w, r, fmt.Errorf("model not found"))
		return
	}
	api.Send(w, r, map[string]interface{}{"deleted": true})
}

// ---------------------------------------------------------------------------
// test / simulate — what the editor's simulator panel calls
// ---------------------------------------------------------------------------

type riskTestPayload struct {
	Evidence            riskengine.Evidence `json:"evidence"`
	ControlFactorHandle string              `json:"controlFactorHandle,omitempty"`
	Bands               []riskengine.Band   `json:"bands,omitempty"`
	WithAttribution     bool                `json:"withAttribution,omitempty"`
}

// Test runs a model against caller-supplied evidence without persisting
// anything — the Evidence -> Evaluate -> Assessment -> Attribution pipeline
// from the architecture note, in one call.
func (a RiskAdmin) Test(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "modelID"))
	model, ok := riskstore.GetModel(id)
	if !ok {
		api.Send(w, r, fmt.Errorf("model not found"))
		return
	}

	payload, err := decodeRiskTestPayload(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	factors := riskstore.FactorsForModel(model)
	assessment, err := riskengine.EvaluateWithOptions(model, factors, payload.Evidence, riskengine.EvalOptions{
		ControlFactorHandle: payload.ControlFactorHandle,
		Bands:               payload.Bands,
	})
	if err != nil {
		api.Send(w, r, err)
		return
	}
	if payload.WithAttribution {
		attr, err := riskengine.Explain(model, factors, payload.Evidence, assessment)
		if err != nil {
			api.Send(w, r, fmt.Errorf("attribution failed: %w", err))
			return
		}
		assessment.Attribution = attr
	}

	api.Send(w, r, map[string]interface{}{"assessment": assessment})
}

func decodeRiskTestPayload(r *http.Request) (*riskTestPayload, error) {
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var payload riskTestPayload
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	}
	return &payload, nil
}

// ---------------------------------------------------------------------------
// explain (Layer C, grounded in Attribution — see attribution.go) and
// suggest-factors (a structure draft a human edits, never auto-applied)
// ---------------------------------------------------------------------------

// ExplainNarrative runs Evaluate + Explain and turns the resulting
// AttributionResult into a short Russian narrative. It never reads raw
// Evidence for the wording — only the already-computed breakdown — so it
// narrates a computed number instead of reasoning about risk from scratch.
//
// This is a deterministic template, not a live LLM call: it is always
// correct and needs no model/API wiring to exercise the "Explain reads only
// breakdown" contract end to end. Swapping in a real call to the
// risk-analyst agent (server/pkg/aiagent, using AttributionResult as its
// prompt context instead of raw evidence) is the natural upgrade once this
// deployment's model wiring is confirmed live — see NarrateAttribution's doc
// comment for exactly where that swap goes.
func (a RiskAdmin) ExplainNarrative(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "modelID"))
	model, ok := riskstore.GetModel(id)
	if !ok {
		api.Send(w, r, fmt.Errorf("model not found"))
		return
	}
	payload, err := decodeRiskTestPayload(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}

	factors := riskstore.FactorsForModel(model)
	assessment, err := riskengine.EvaluateWithOptions(model, factors, payload.Evidence, riskengine.EvalOptions{
		ControlFactorHandle: payload.ControlFactorHandle,
		Bands:               payload.Bands,
	})
	if err != nil {
		api.Send(w, r, err)
		return
	}
	attr, err := riskengine.Explain(model, factors, payload.Evidence, assessment)
	if err != nil {
		api.Send(w, r, fmt.Errorf("attribution failed: %w", err))
		return
	}
	assessment.Attribution = attr
	assessment.Explanation = riskengine.NarrateAttribution(assessment, attr, factors)

	api.Send(w, r, map[string]interface{}{"assessment": assessment})
}

type suggestFactorsPayload struct {
	Description string `json:"description"`
}

// SuggestFactors returns a draft factor list for a human to edit in the
// library — never applied automatically. The v1 implementation is a
// deterministic keyword heuristic (see riskengine.SuggestFactors), not a
// live LLM call, for the same reason ExplainNarrative isn't one: it is
// exercisable and correct today without needing this deployment's model
// credentials, and it defines the exact contract (free text in, draft
// RiskFactorDef list out) a real risk-analyst agent call slots into later.
func (a RiskAdmin) SuggestFactors(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var payload suggestFactorsPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	if payload.Description == "" {
		api.Send(w, r, fmt.Errorf("description is required"))
		return
	}

	api.Send(w, r, map[string]interface{}{
		"factors": riskengine.SuggestFactors(payload.Description),
	})
}

// ---------------------------------------------------------------------------
// subject bindings
// ---------------------------------------------------------------------------

func (a RiskAdmin) ListBindings(w http.ResponseWriter, r *http.Request) {
	nsFilter := parseUint64String(r.URL.Query().Get("namespaceID"))
	api.Send(w, r, map[string]interface{}{"bindings": riskstore.ListBindings(nsFilter)})
}

func (a RiskAdmin) GetBinding(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "bindingID"))
	b, ok := riskstore.GetBinding(id)
	if !ok {
		api.Send(w, r, fmt.Errorf("binding not found"))
		return
	}
	api.Send(w, r, map[string]interface{}{"binding": b})
}

func (a RiskAdmin) CreateBinding(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var b riskengine.RiskSubjectBinding
	if err := json.Unmarshal(body, &b); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	if b.NamespaceID == 0 || b.ModelID == 0 || b.ModuleID == 0 {
		api.Send(w, r, fmt.Errorf("namespaceID, modelID and moduleID are required"))
		return
	}
	if _, ok := riskstore.GetModel(b.ModelID); !ok {
		api.Send(w, r, fmt.Errorf("model %d not found", b.ModelID))
		return
	}

	b.ID = riskstore.NextID()
	b.CreatedAt = time.Now()
	riskstore.SaveBinding(&b)

	api.Send(w, r, map[string]interface{}{"created": true, "binding": b})
}

func (a RiskAdmin) UpdateBinding(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "bindingID"))
	existing, ok := riskstore.GetBinding(id)
	if !ok {
		api.Send(w, r, fmt.Errorf("binding not found"))
		return
	}
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var patch riskengine.RiskSubjectBinding
	if err := json.Unmarshal(body, &patch); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	patch.ID = id
	patch.NamespaceID = existing.NamespaceID
	patch.CreatedAt = existing.CreatedAt
	now := time.Now()
	patch.UpdatedAt = &now
	riskstore.SaveBinding(&patch)

	api.Send(w, r, map[string]interface{}{"updated": true, "binding": patch})
}

func (a RiskAdmin) DeleteBinding(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "bindingID"))
	if !riskstore.DeleteBinding(id) {
		api.Send(w, r, fmt.Errorf("binding not found"))
		return
	}
	api.Send(w, r, map[string]interface{}{"deleted": true})
}

// ---------------------------------------------------------------------------
// assessments
// ---------------------------------------------------------------------------

type assessPayload struct {
	SubjectRecordID     uint64              `json:"subjectRecordID,string"`
	Evidence            riskengine.Evidence `json:"evidence"`
	ControlFactorHandle string              `json:"controlFactorHandle,omitempty"`
	Bands               []riskengine.Band   `json:"bands,omitempty"`
}

// Assess runs and *persists* an assessment for one subject record under a
// binding — the manual-recalculation path from the roadmap's phase 1
// ("ручной пересчёт"); AutoRecalc (phase 4) reaches the same computation
// through compose/service.StartRiskRecordTriggers instead of this endpoint.
func (a RiskAdmin) Assess(w http.ResponseWriter, r *http.Request) {
	bindingID := parseUint64String(chi.URLParam(r, "bindingID"))
	binding, ok := riskstore.GetBinding(bindingID)
	if !ok {
		api.Send(w, r, fmt.Errorf("binding not found"))
		return
	}
	model, ok := riskstore.GetModel(binding.ModelID)
	if !ok {
		api.Send(w, r, fmt.Errorf("model %d not found", binding.ModelID))
		return
	}

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var payload assessPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	factors := riskstore.FactorsForModel(model)
	assessment, err := riskengine.EvaluateWithOptions(model, factors, payload.Evidence, riskengine.EvalOptions{
		ControlFactorHandle: payload.ControlFactorHandle,
		Bands:               payload.Bands,
	})
	if err != nil {
		api.Send(w, r, err)
		return
	}
	attr, err := riskengine.Explain(model, factors, payload.Evidence, assessment)
	if err == nil { // attribution is best-effort here; a bad evidence shape shouldn't block persisting the score
		assessment.Attribution = attr
		assessment.Explanation = riskengine.NarrateAttribution(assessment, attr, factors)
	}
	assessment.BindingID = bindingID
	assessment.SubjectRecordID = payload.SubjectRecordID

	riskstore.SaveAssessment(assessment)
	api.Send(w, r, map[string]interface{}{"created": true, "assessment": assessment})
}

func (a RiskAdmin) ListAssessments(w http.ResponseWriter, r *http.Request) {
	bindingID := parseUint64String(chi.URLParam(r, "bindingID"))
	subjectRecordID := parseUint64String(r.URL.Query().Get("subjectRecordID"))

	// latestPerSubject means "one row per subject" — only meaningful across
	// *every* subject (the portfolio view). When a caller also passes
	// subjectRecordID (a record-page block asking "the latest assessment
	// for this one record"), that request must win: AssessmentsForBinding
	// is already newest-first, so its own [0] is that subject's latest.
	// Falling through to LatestPerSubject(bindingID) here used to silently
	// ignore subjectRecordID and hand back one row per *every* subject
	// instead — every record's risk block ended up showing whichever
	// subject happened to sort first, regardless of which record it was on.
	if subjectRecordID == 0 && r.URL.Query().Get("latestPerSubject") != "" {
		api.Send(w, r, map[string]interface{}{"assessments": riskstore.LatestPerSubject(bindingID)})
		return
	}
	api.Send(w, r, map[string]interface{}{"assessments": riskstore.AssessmentsForBinding(bindingID, subjectRecordID)})
}

// PortfolioSummary aggregates mean(|contribution|) per factor across a
// binding's most recent assessment for every subject — the "which factors
// actually move risk in this portfolio" view from the architecture note's
// dashboard section, not just a single object's breakdown.
func (a RiskAdmin) PortfolioSummary(w http.ResponseWriter, r *http.Request) {
	bindingID := parseUint64String(chi.URLParam(r, "bindingID"))
	latest := riskstore.LatestPerSubject(bindingID)

	type agg struct {
		Handle       string  `json:"factorHandle"`
		MeanAbsValue float64 `json:"meanAbsValue"`
		Count        int     `json:"count"`
	}
	sums := map[string]*agg{}
	var order []string
	for _, a := range latest {
		if a.Attribution == nil {
			continue
		}
		for _, c := range a.Attribution.Contributions {
			e, ok := sums[c.Handle]
			if !ok {
				e = &agg{Handle: c.Handle}
				sums[c.Handle] = e
				order = append(order, c.Handle)
			}
			v := c.Value
			if v < 0 {
				v = -v
			}
			e.MeanAbsValue += v
			e.Count++
		}
	}
	result := make([]*agg, 0, len(order))
	for _, h := range order {
		e := sums[h]
		if e.Count > 0 {
			e.MeanAbsValue /= float64(e.Count)
		}
		result = append(result, e)
	}

	api.Send(w, r, map[string]interface{}{
		"subjects": len(latest),
		"factors":  result,
	})
}

// ---------------------------------------------------------------------------
// treatments
// ---------------------------------------------------------------------------

func (a RiskAdmin) ListTreatments(w http.ResponseWriter, r *http.Request) {
	bindingID := parseUint64String(chi.URLParam(r, "bindingID"))
	api.Send(w, r, map[string]interface{}{"treatments": riskstore.ListTreatments(bindingID)})
}

func (a RiskAdmin) CreateTreatment(w http.ResponseWriter, r *http.Request) {
	bindingID := parseUint64String(chi.URLParam(r, "bindingID"))
	if _, ok := riskstore.GetBinding(bindingID); !ok {
		api.Send(w, r, fmt.Errorf("binding not found"))
		return
	}
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var t riskengine.RiskTreatment
	if err := json.Unmarshal(body, &t); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	if t.Description == "" {
		api.Send(w, r, fmt.Errorf("description is required"))
		return
	}
	if t.Status == "" {
		t.Status = riskengine.TreatmentOpen
	}
	t.BindingID = bindingID
	if t.AssessedAt.IsZero() {
		t.AssessedAt = time.Now()
	}
	riskstore.SaveTreatment(&t)

	api.Send(w, r, map[string]interface{}{"created": true, "treatment": t})
}

func (a RiskAdmin) UpdateTreatment(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "treatmentID"))
	existing, ok := riskstore.GetTreatment(id)
	if !ok {
		api.Send(w, r, fmt.Errorf("treatment not found"))
		return
	}
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var patch riskengine.RiskTreatment
	if err := json.Unmarshal(body, &patch); err != nil {
		api.Send(w, r, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	patch.ID = id
	patch.BindingID = existing.BindingID
	patch.AssessedAt = existing.AssessedAt
	riskstore.SaveTreatment(&patch)

	api.Send(w, r, map[string]interface{}{"updated": true, "treatment": patch})
}

func (a RiskAdmin) DeleteTreatment(w http.ResponseWriter, r *http.Request) {
	id := parseUint64String(chi.URLParam(r, "treatmentID"))
	if !riskstore.DeleteTreatment(id) {
		api.Send(w, r, fmt.Errorf("treatment not found"))
		return
	}
	api.Send(w, r, map[string]interface{}{"deleted": true})
}
