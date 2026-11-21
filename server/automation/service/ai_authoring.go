package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
)

// Authoring AI workflows from a description.
//
// A language model writing raw workflow JSON (numeric step IDs, typed
// argument expressions, path objects) gets it wrong most of the time, and the
// failure only shows when the workflow is saved or run. Instead the assistant
// writes a small flow description that names steps, and this file turns it
// into a real workflow - checking every function, argument and connection
// against the platform on the way, and reporting what is wrong in words the
// model can act on. The result is validated by the same converter that
// prepares a workflow for execution, so "valid" here means it will run.
//
// A flow description:
//
//	{
//	  "name": "Explain anomalies",
//	  "trigger": {"resourceType": "anomaly:finding", "eventType": "onCreate",
//	              "constraints": [{"name": "finding.severity", "op": "=", "values": ["high"]}]},
//	  "steps": [
//	    {"id": "explain", "function": "aiAsk",
//	     "args": {"prompt": "=\"Explain this finding: \" + toPlainJSON(finding)"},
//	     "results": {"explanation": "text"}},
//	    {"id": "log", "function": "logInfo", "args": {"message": "=explanation"}}
//	  ],
//	  "flow": [{"from": "explain", "to": "log"}]
//	}
//
// An argument value starting with "=" is an expression over the workflow's
// variables; anything else is a constant. "results" maps a variable name to
// the function result it receives.

type (
	AIFlowSpec struct {
		Name        string         `json:"name"`
		Handle      string         `json:"handle,omitempty"`
		Description string         `json:"description,omitempty"`
		Trigger     *AIFlowTrigger `json:"trigger,omitempty"`
		Steps       []AIFlowStep   `json:"steps"`
		Flow        []AIFlowLink   `json:"flow"`
	}

	AIFlowTrigger struct {
		ResourceType string                     `json:"resourceType"`
		EventType    string                     `json:"eventType"`
		Constraints  types.TriggerConstraintSet `json:"constraints,omitempty"`
	}

	// AIFlowStep is a function call, a prompt to a user, or a gateway -
	// exactly one of Function, Prompt and Gateway is set.
	AIFlowStep struct {
		ID   string `json:"id"`
		Name string `json:"name,omitempty"`

		Function string `json:"function,omitempty"` // e.g. aiAsk, logInfo
		Prompt   string `json:"prompt,omitempty"`   // client prompt, e.g. choice, alert
		Gateway  string `json:"gateway,omitempty"`  // excl, incl, fork or join

		Args    map[string]interface{} `json:"args,omitempty"`
		Results map[string]string      `json:"results,omitempty"`
	}

	// AIFlowLink connects two steps. When is a condition, used on the paths
	// that leave an exclusive or inclusive gateway.
	AIFlowLink struct {
		From string `json:"from"`
		To   string `json:"to"`
		When string `json:"when,omitempty"`
	}

	AIFlowCompiled struct {
		Workflow *types.Workflow
		Trigger  *types.Trigger
		// Issues is everything wrong with the description, each one saying
		// what to change. Empty means the workflow is valid.
		Issues []string
	}
)

// userPrompts are the client prompts a flow can use, with the type of each
// argument (the server has no registry of them; the editor defines them).
var userPrompts = map[string]map[string]string{
	"choice": {
		"owner": "ID", "title": "String", "message": "String",
		"confirmButtonLabel": "String", "rejectButtonLabel": "String",
	},
	"alert": {
		"owner": "ID", "title": "String", "message": "String",
	},
	"notification": {
		"owner": "ID", "title": "String", "message": "String", "variant": "String",
	},
}

var gatewayKinds = map[string]bool{"excl": true, "incl": true, "fork": true, "join": true}

func (a AIFlowStep) kinds() []string {
	var k []string
	if a.Function != "" {
		k = append(k, "function")
	}
	if a.Prompt != "" {
		k = append(k, "prompt")
	}
	if a.Gateway != "" {
		k = append(k, "gateway")
	}
	return k
}

// ParseAIFlow reads a flow description, rejecting unknown fields so that a
// misspelt "step" or "arg" is reported instead of silently ignored.
func ParseAIFlow(raw string) (*AIFlowSpec, error) {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	var spec AIFlowSpec
	if err := dec.Decode(&spec); err != nil {
		return nil, fmt.Errorf("the flow description is not valid: %w (expected fields: name, handle, description, trigger, steps, flow)", err)
	}
	return &spec, nil
}

// CompileAIFlow turns a flow description into a workflow and, when it has a
// trigger, a trigger - both disabled, to be reviewed and switched on by a
// person. installer is who the workflow will run as.
func CompileAIFlow(svc *workflow, spec *AIFlowSpec, installer uint64) *AIFlowCompiled {
	out := &AIFlowCompiled{}
	bad := func(format string, args ...interface{}) {
		out.Issues = append(out.Issues, fmt.Sprintf(format, args...))
	}

	if strings.TrimSpace(spec.Name) == "" {
		bad("name is required")
	}
	if len(spec.Steps) == 0 {
		bad("steps is empty: describe at least one step")
	}

	wf := &types.Workflow{
		Handle: spec.Handle,
		Meta:   &types.WorkflowMeta{Name: spec.Name, Description: spec.Description},
		RunAs:  installer,
	}
	if wf.Handle == "" {
		wf.Handle = handleFromName(spec.Name)
	}

	ids := map[string]uint64{}
	for i, s := range spec.Steps {
		stepID := uint64(i + 1)

		switch {
		case strings.TrimSpace(s.ID) == "":
			bad("step %d has no id", i+1)
			continue
		case ids[s.ID] != 0:
			bad("step id %q is used twice", s.ID)
			continue
		}
		ids[s.ID] = stepID

		kinds := s.kinds()
		if len(kinds) != 1 {
			bad("step %q must set exactly one of function, prompt or gateway (it sets %d)", s.ID, len(kinds))
			continue
		}

		step := &types.WorkflowStep{ID: stepID, Meta: types.WorkflowStepMeta{Name: firstNonEmpty(s.Name, s.ID)}}

		switch kinds[0] {
		case "function":
			compileFunctionStep(step, s, bad)
		case "prompt":
			compilePromptStep(step, s, bad)
		case "gateway":
			if !gatewayKinds[s.Gateway] {
				bad("step %q: gateway must be one of excl, incl, fork, join (got %q)", s.ID, s.Gateway)
			}
			step.Kind, step.Ref = types.WorkflowStepKindGateway, s.Gateway
			if len(s.Args) > 0 || len(s.Results) > 0 {
				bad("step %q: a gateway takes no args or results - put conditions on the flow links (\"when\")", s.ID)
			}
		}
		wf.Steps = append(wf.Steps, step)
	}

	for i, l := range spec.Flow {
		from, to := ids[l.From], ids[l.To]
		switch {
		case from == 0:
			bad("flow link %d: unknown step %q in \"from\" (steps: %s)", i+1, l.From, stepNames(ids))
		case to == 0:
			bad("flow link %d: unknown step %q in \"to\" (steps: %s)", i+1, l.To, stepNames(ids))
		case from == to:
			bad("flow link %d: step %q links to itself", i+1, l.From)
		default:
			wf.Paths = append(wf.Paths, &types.WorkflowPath{ParentID: from, ChildID: to, Expr: strings.TrimPrefix(strings.TrimSpace(l.When), "=")})
		}
	}

	if spec.Trigger != nil {
		if spec.Trigger.ResourceType == "" || spec.Trigger.EventType == "" {
			bad("trigger needs both resourceType and eventType")
		} else {
			out.Trigger = &types.Trigger{
				ResourceType: spec.Trigger.ResourceType,
				EventType:    spec.Trigger.EventType,
				Constraints:  spec.Trigger.Constraints,
			}
		}
	}

	out.Workflow = wf

	// only ask the converter about a description that is whole: its errors
	// about a half-built graph would just repeat the ones above
	if len(out.Issues) == 0 && svc != nil {
		if _, issues := Convert(svc, wf); len(issues) > 0 {
			for _, is := range issues {
				bad("%s", explainIssue(is, spec))
			}
		}
	}

	return out
}

func compileFunctionStep(step *types.WorkflowStep, s AIFlowStep, bad func(string, ...interface{})) {
	step.Kind, step.Ref = types.WorkflowStepKindFunction, s.Function

	f := Registry().Function(s.Function)
	if f == nil {
		bad("step %q: unknown function %q. Functions you can use: %s", s.ID, s.Function, functionNames())
		return
	}

	params := map[string]*types.Param{}
	names := make([]string, 0, len(f.Parameters))
	for _, p := range f.Parameters {
		params[p.Name] = p
		names = append(names, p.Name)
	}

	for _, key := range sortedKeys(s.Args) {
		p := params[key]
		if p == nil {
			bad("step %q: %s has no argument %q. Its arguments: %s", s.ID, s.Function, key, strings.Join(names, ", "))
			continue
		}
		typ := "String"
		if len(p.Types) > 0 {
			typ = p.Types[0]
		}
		step.Arguments = append(step.Arguments, argExpr(key, typ, s.Args[key]))
	}

	for _, p := range f.Parameters {
		if _, given := s.Args[p.Name]; p.Required && !given {
			bad("step %q: %s needs the argument %q", s.ID, s.Function, p.Name)
		}
	}

	resultNames := map[string]bool{}
	var resNames []string
	for _, r := range f.Results {
		resultNames[r.Name] = true
		resNames = append(resNames, r.Name)
	}
	for _, variable := range sortedKeys2(s.Results) {
		src := s.Results[variable]
		if !resultNames[src] {
			bad("step %q: %s has no result %q. Its results: %s", s.ID, s.Function, src, strings.Join(resNames, ", "))
			continue
		}
		step.Results = append(step.Results, &types.Expr{Target: variable, Source: src})
	}
}

func compilePromptStep(step *types.WorkflowStep, s AIFlowStep, bad func(string, ...interface{})) {
	step.Kind, step.Ref = types.WorkflowStepKindPrompt, s.Prompt

	known, ok := userPrompts[s.Prompt]
	if !ok {
		bad("step %q: unknown prompt %q. Prompts you can use: %s", s.ID, s.Prompt, strings.Join(sortedKeys3(userPrompts), ", "))
		return
	}

	for _, key := range sortedKeys(s.Args) {
		typ, ok := known[key]
		if !ok {
			bad("step %q: the %q prompt has no argument %q. Its arguments: %s", s.ID, s.Prompt, key, strings.Join(sortedKeys4(known), ", "))
			continue
		}
		step.Arguments = append(step.Arguments, argExpr(key, typ, s.Args[key]))
	}
	if s.Prompt == "choice" || s.Prompt == "alert" || s.Prompt == "notification" {
		if _, ok := s.Args["message"]; !ok {
			bad("step %q: the %q prompt needs a message", s.ID, s.Prompt)
		}
	}

	for _, variable := range sortedKeys2(s.Results) {
		// the answer of a "choice" prompt is its "value"
		step.Results = append(step.Results, &types.Expr{Target: variable, Source: s.Results[variable]})
	}
}

// argExpr builds a typed argument: "=expr" is an expression, anything else a constant.
func argExpr(target, typ string, v interface{}) *types.Expr {
	if str, ok := v.(string); ok && strings.HasPrefix(strings.TrimSpace(str), "=") {
		return &types.Expr{Target: target, Type: typ, Expr: strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(str), "="))}
	}
	return &types.Expr{Target: target, Type: typ, Value: v}
}

// explainIssue says what the converter found in terms of the description:
// its issues point at step and path numbers, the author thinks in step names.
func explainIssue(is *types.WorkflowIssue, spec *AIFlowSpec) string {
	msg := is.Description
	where := ""

	if i, ok := is.Culprit["step"]; ok && i >= 0 && i < len(spec.Steps) {
		where = fmt.Sprintf("step %q: ", spec.Steps[i].ID)
	} else if i, ok := is.Culprit["path"]; ok && i >= 0 && i < len(spec.Flow) {
		where = fmt.Sprintf("flow link %s -> %s: ", spec.Flow[i].From, spec.Flow[i].To)
	}

	switch {
	case strings.Contains(msg, "incompatible argument type"):
		msg += " - give the argument as a constant of that type, or as an expression starting with \"=\""
	case strings.Contains(msg, "failed to parse"):
		msg += " - check the expression syntax"
	case strings.Contains(msg, "failed to resolve step"):
		msg += " - check the \"flow\" links"
	}
	return where + msg
}

// Install creates the compiled workflow (and its trigger), disabled.
func (c *AIFlowCompiled) Install(ctx context.Context) (*types.Workflow, error) {
	if len(c.Issues) > 0 {
		return nil, fmt.Errorf("the flow has problems: %s", strings.Join(c.Issues, "; "))
	}
	if DefaultWorkflow == nil || DefaultTrigger == nil {
		return nil, fmt.Errorf("workflow service not available")
	}
	if auth.GetIdentityFromContext(ctx) == nil {
		return nil, fmt.Errorf("a user is required")
	}

	if existing, _, err := DefaultWorkflow.Search(ctx, types.WorkflowFilter{Handle: c.Workflow.Handle}); err == nil && len(existing) > 0 {
		return nil, fmt.Errorf("a workflow with the handle %q already exists - choose another handle", c.Workflow.Handle)
	}

	created, err := DefaultWorkflow.Create(ctx, c.Workflow)
	if err != nil {
		return nil, err
	}
	if c.Trigger != nil {
		c.Trigger.WorkflowID = created.ID
		if _, err = DefaultTrigger.Create(ctx, c.Trigger); err != nil {
			return created, fmt.Errorf("workflow created but its trigger was not: %w", err)
		}
	}
	return created, nil
}

// AIFunctionCatalog describes the functions an AI flow can call, for the
// assistant that writes flows: name, what it does, and each argument with its
// type and whether it is required.
func AIFunctionCatalog() string {
	var b strings.Builder
	for _, f := range Registry().Functions() {
		if !isAuthoringFunction(f) {
			continue
		}
		short := ""
		if f.Meta != nil {
			short = f.Meta.Short
		}
		fmt.Fprintf(&b, "%s - %s\n", f.Ref, short)

		for _, p := range f.Parameters {
			req := ""
			if p.Required {
				req = ", required"
			}
			desc := ""
			if p.Meta != nil && p.Meta.Description != "" {
				desc = ": " + p.Meta.Description
			}
			fmt.Fprintf(&b, "    arg %s (%s%s)%s\n", p.Name, strings.Join(p.Types, "|"), req, desc)
		}
		if len(f.Results) > 0 {
			rr := make([]string, 0, len(f.Results))
			for _, r := range f.Results {
				rr = append(rr, r.Name)
			}
			fmt.Fprintf(&b, "    results: %s\n", strings.Join(rr, ", "))
		}
	}

	b.WriteString("\nUser prompts (step field \"prompt\"):\n")
	for _, name := range sortedKeys3(userPrompts) {
		fmt.Fprintf(&b, "    %s - args: %s; result: value\n", name, strings.Join(sortedKeys4(userPrompts[name]), ", "))
	}
	b.WriteString("\nGateways (step field \"gateway\"): excl (first matching link), incl (all matching links), fork (all links), join.\n")
	return b.String()
}

// isAuthoringFunction picks what an assistant should offer: the AI steps, and
// the plain helpers that make a flow useful without being a risk.
func isAuthoringFunction(f *types.Function) bool {
	if f.Labels["ai"] != "" {
		return true
	}
	switch f.Ref {
	case "logDebug", "logInfo", "logWarn", "logError":
		return true
	}
	return false
}

func functionNames() string {
	var names []string
	for _, f := range Registry().Functions() {
		if isAuthoringFunction(f) {
			names = append(names, f.Ref)
		}
	}
	return strings.Join(names, ", ")
}

func stepNames(ids map[string]uint64) string {
	names := make([]string, 0, len(ids))
	for n := range ids {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func handleFromName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "_"):
			b.WriteRune('_')
		}
	}
	h := strings.Trim(b.String(), "_")
	if h == "" || (h[0] >= '0' && h[0] <= '9') {
		h = "flow_" + h
	}
	if len(h) > 64 {
		h = h[:64]
	}
	return h
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func sortedKeys(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedKeys2(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedKeys3(m map[string]map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedKeys4(m map[string]string) []string { return sortedKeys2(m) }
