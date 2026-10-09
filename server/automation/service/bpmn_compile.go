// Package service: BPMN 2.0 compiler.
//
// CompileBPMN turns a BPMN 2.0 XML diagram into the same Steps/Paths shape
// the existing hand-authored (JSON step-list) workflow editor already
// produces - see workflow_converter.go's Convert/makeGraph, which then runs
// entirely unchanged. No new execution engine: BPMN gets a compiler onto
// primitives that already exist (gateway fork/join/excl/incl, prompt,
// delay, function, termination).
//
// Supported BPMN elements:
//
//	startEvent                          -> not emitted as a step; its
//	                                        outgoing flow's target becomes
//	                                        the graph's natural entry point
//	                                        (Session.Start finds the orphan
//	                                        step with no incoming path)
//	endEvent                            -> termination
//	task/serviceTask/scriptTask/
//	businessRuleTask/sendTask/
//	receiveTask                          -> function (needs a chainID or
//	                                        functionRef extension property)
//	userTask/manualTask                  -> prompt (User Task - suspends the
//	                                        session; resumed via the
//	                                        existing prompt/Resume API, same
//	                                        one the interactive "pending
//	                                        prompts" modal already uses)
//	exclusiveGateway                     -> gateway:excl
//	inclusiveGateway                     -> gateway:incl
//	parallelGateway                      -> gateway:fork or gateway:join,
//	                                        picked by sequence-flow topology
//	                                        (>1 outgoing = fork, >1 incoming
//	                                        = join)
//	intermediateCatchEvent
//	  + timerEventDefinition             -> delay
//	  + messageEventDefinition           -> prompt, with a correlationKey
//	                                        argument - the entry point for
//	                                        message-correlated resume (see
//	                                        ResolveCorrelation)
//
// Unsupported (boundaryEvent, eventBasedGateway, subProcess, callActivity,
// ...) produce a clear error instead of a silently wrong graph.
//
// Extension properties: any BPMN tool's "properties" extension is read by
// local element name only (xml:"properties"/"property"), so both a custom
// <lowcode:properties> and Camunda's <camunda:properties> are accepted
// without extra configuration. A property value starting with "=" is
// treated as a Corteza expression (evaluated against the workflow scope);
// anything else is a literal value.
package service

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/id"
)

type bpmnDefinitions struct {
	XMLName xml.Name    `xml:"definitions"`
	Process bpmnProcess `xml:"process"`
}

type bpmnProcess struct {
	ID    string        `xml:"id,attr"`
	Name  string        `xml:"name,attr"`
	Flows []bpmnSeqFlow `xml:"sequenceFlow"`
	Nodes []bpmnNode    `xml:",any"`
}

type bpmnNode struct {
	XMLName xml.Name
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Default string `xml:"default,attr"` // gateways: ID of the default outgoing flow

	ExtensionElements struct {
		Properties struct {
			Property []bpmnProperty `xml:"property"`
		} `xml:"properties"`
	} `xml:"extensionElements"`

	TimerEventDefinition   *bpmnTimerEventDefinition   `xml:"timerEventDefinition"`
	MessageEventDefinition *bpmnMessageEventDefinition `xml:"messageEventDefinition"`
}

type bpmnProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type bpmnTimerEventDefinition struct {
	TimeDuration string `xml:"timeDuration"`
	TimeDate     string `xml:"timeDate"`
}

type bpmnMessageEventDefinition struct {
	MessageRef string `xml:"messageRef,attr"`
}

type bpmnSeqFlow struct {
	ID            string `xml:"id,attr"`
	Name          string `xml:"name,attr"`
	SourceRef     string `xml:"sourceRef,attr"`
	TargetRef     string `xml:"targetRef,attr"`
	ConditionExpr *struct {
		Body string `xml:",chardata"`
	} `xml:"conditionExpression"`
}

func (n bpmnNode) properties() map[string]string {
	out := make(map[string]string, len(n.ExtensionElements.Properties.Property))
	for _, p := range n.ExtensionElements.Properties.Property {
		out[p.Name] = p.Value
	}
	return out
}

// exprValue turns an extension property's raw text into an *types.Expr: a
// leading "=" marks it as a Corteza expression evaluated against the
// workflow scope, anything else is used as a literal value.
func exprValue(target, raw string) *types.Expr {
	if strings.HasPrefix(raw, "=") {
		return &types.Expr{Target: target, Expr: strings.TrimSpace(strings.TrimPrefix(raw, "="))}
	}
	return &types.Expr{Target: target, Value: raw}
}

// CompileBPMN parses a BPMN 2.0 XML document and returns a workflow with
// Steps/Paths populated - Handle/Meta/Enabled/RunAs are left for the caller
// to fill in before saving through the normal workflow create/update path
// (which already validates the result via Convert/makeGraph).
func CompileBPMN(raw []byte) (*types.Workflow, error) {
	var def bpmnDefinitions
	if err := xml.Unmarshal(raw, &def); err != nil {
		return nil, fmt.Errorf("invalid BPMN XML: %w", err)
	}
	if def.Process.ID == "" && len(def.Process.Nodes) == 0 {
		return nil, fmt.Errorf("no <process> element found (or it has no flow elements)")
	}

	c := &bpmnCompiler{
		ids:      make(map[string]uint64),
		nodeByID: make(map[string]bpmnNode),
		inCount:  make(map[string]int),
		outCount: make(map[string]int),
	}

	for _, n := range def.Process.Nodes {
		c.nodeByID[n.ID] = n
		c.ids[n.ID] = id.Next()
	}
	for _, f := range def.Process.Flows {
		c.inCount[f.TargetRef]++
		c.outCount[f.SourceRef]++
	}

	steps := make(types.WorkflowStepSet, 0, len(def.Process.Nodes))
	for _, n := range def.Process.Nodes {
		step, err := c.compileNode(n)
		if err != nil {
			return nil, fmt.Errorf("%s %q: %w", localName(n.XMLName), n.ID, err)
		}
		if step != nil {
			steps = append(steps, step)
		}
	}

	paths, err := c.compilePaths(def.Process.Flows)
	if err != nil {
		return nil, err
	}

	return &types.Workflow{
		Steps: steps,
		Paths: paths,
	}, nil
}

type bpmnCompiler struct {
	ids      map[string]uint64 // BPMN element ID -> Corteza step ID
	nodeByID map[string]bpmnNode
	inCount  map[string]int
	outCount map[string]int
}

func localName(n xml.Name) string { return n.Local }

// isStartEvent/isSkipped identify BPMN elements that don't become a step of
// their own (see the package doc comment for why startEvent is skipped).
func (c *bpmnCompiler) isStartEvent(bpmnID string) bool {
	n, ok := c.nodeByID[bpmnID]
	return ok && localName(n.XMLName) == "startEvent"
}

func (c *bpmnCompiler) compileNode(n bpmnNode) (*types.WorkflowStep, error) {
	kind := localName(n.XMLName)
	props := n.properties()

	step := &types.WorkflowStep{
		ID: c.ids[n.ID],
		Meta: types.WorkflowStepMeta{
			Name: n.Name,
		},
	}

	switch kind {
	case "startEvent":
		// Not emitted: the graph's natural entry point is whichever step
		// has no incoming path once the startEvent's own flow is dropped
		// (see compilePaths) - Session.Start finds it via g.Orphans().
		return nil, nil

	case "endEvent":
		step.Kind = types.WorkflowStepKindTermination
		return step, nil

	case "task", "serviceTask", "scriptTask", "businessRuleTask", "sendTask", "receiveTask":
		return c.compileServiceTask(step, props)

	case "userTask", "manualTask":
		return c.compilePromptTask(step, props, n.Name)

	case "exclusiveGateway":
		step.Kind = types.WorkflowStepKindGateway
		step.Ref = "excl"
		return step, nil

	case "inclusiveGateway":
		step.Kind = types.WorkflowStepKindGateway
		step.Ref = "incl"
		return step, nil

	case "parallelGateway":
		step.Kind = types.WorkflowStepKindGateway
		switch {
		case c.outCount[n.ID] > 1:
			step.Ref = "fork"
		case c.inCount[n.ID] > 1:
			step.Ref = "join"
		default:
			return nil, fmt.Errorf("parallel gateway needs either >1 incoming (join) or >1 outgoing (fork) sequence flow")
		}
		return step, nil

	case "intermediateCatchEvent":
		return c.compileIntermediateCatchEvent(step, props, n)

	case "sequenceFlow":
		// handled separately in compilePaths
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported BPMN element %q - not compiled (boundary events, event-based gateways and sub-processes aren't supported yet)", kind)
	}
}

func (c *bpmnCompiler) compileServiceTask(step *types.WorkflowStep, props map[string]string) (*types.WorkflowStep, error) {
	step.Kind = types.WorkflowStepKindFunction

	if functionRef, ok := props["functionRef"]; ok && functionRef != "" {
		step.Ref = functionRef
		delete(props, "functionRef")
		step.Arguments = propsToArguments(props)
		return step, nil
	}

	chainID, ok := props["chainID"]
	if !ok || chainID == "" {
		return nil, fmt.Errorf("needs an extension property \"chainID\" (which rule chain to run) or \"functionRef\" (which automation function to call)")
	}

	step.Ref = "compose.runRuleChain"
	step.Arguments = []*types.Expr{exprValue("chainID", chainID)}
	if input, ok := props["input"]; ok && input != "" {
		step.Arguments = append(step.Arguments, exprValue("input", input))
	}
	return step, nil
}

func (c *bpmnCompiler) compilePromptTask(step *types.WorkflowStep, props map[string]string, ref string) (*types.WorkflowStep, error) {
	step.Kind = types.WorkflowStepKindPrompt
	step.Ref = ref
	step.Arguments = propsToArguments(props)
	return step, nil
}

func (c *bpmnCompiler) compileIntermediateCatchEvent(step *types.WorkflowStep, props map[string]string, n bpmnNode) (*types.WorkflowStep, error) {
	switch {
	case n.TimerEventDefinition != nil:
		step.Kind = types.WorkflowStepKindDelay
		dur := n.TimerEventDefinition.TimeDuration
		if dur == "" {
			return nil, fmt.Errorf("timer intermediate event needs a timeDuration (ISO-8601, e.g. PT1H)")
		}
		seconds, err := parseISO8601Duration(dur)
		if err != nil {
			return nil, fmt.Errorf("timeDuration %q: %w", dur, err)
		}
		step.Arguments = []*types.Expr{{Target: "duration", Value: seconds}}
		return step, nil

	case n.MessageEventDefinition != nil:
		// Message intermediate catch events suspend the session exactly
		// like a user task (prompt) - the difference is who resumes it:
		// a person via the pending-prompts modal, or automatically via
		// ResolveCorrelation when a matching message arrives on a
		// kafka.subscribe/rabbitmq.subscribe-fed topic/queue (see
		// correlation.go). correlationKey is required so there's
		// something to match the incoming message against.
		key, ok := props["correlationKey"]
		if !ok || key == "" {
			return nil, fmt.Errorf("message intermediate catch event needs an extension property \"correlationKey\"")
		}
		step.Kind = types.WorkflowStepKindPrompt
		step.Ref = "message:" + n.MessageEventDefinition.MessageRef
		step.Arguments = propsToArguments(props)
		return step, nil

	default:
		return nil, fmt.Errorf("intermediate catch event needs a timerEventDefinition or messageEventDefinition (other event types aren't supported yet)")
	}
}

// propsToArguments turns extension properties into workflow step Arguments,
// sorted by name for a deterministic, diffable compile output.
func propsToArguments(props map[string]string) []*types.Expr {
	if len(props) == 0 {
		return nil
	}
	names := make([]string, 0, len(props))
	for k := range props {
		names = append(names, k)
	}
	sort.Strings(names)

	out := make([]*types.Expr, 0, len(names))
	for _, k := range names {
		out = append(out, exprValue(k, props[k]))
	}
	return out
}

func (c *bpmnCompiler) compilePaths(flows []bpmnSeqFlow) (types.WorkflowPathSet, error) {
	// default flows are evaluated last so an unconditional "default" branch
	// only wins when nothing else matched.
	defaultOf := make(map[string]string) // gateway (source) ID -> its default flow ID
	for id, n := range c.nodeByID {
		if n.Default != "" {
			defaultOf[id] = n.Default
		}
	}

	ordered := make([]bpmnSeqFlow, 0, len(flows))
	var defaults []bpmnSeqFlow
	for _, f := range flows {
		if c.isStartEvent(f.SourceRef) {
			// dropped: the target becomes the graph's natural entry point
			continue
		}
		if defaultOf[f.SourceRef] == f.ID {
			defaults = append(defaults, f)
			continue
		}
		ordered = append(ordered, f)
	}
	ordered = append(ordered, defaults...)

	paths := make(types.WorkflowPathSet, 0, len(ordered))
	for _, f := range ordered {
		parentID, ok := c.ids[f.SourceRef]
		if !ok {
			return nil, fmt.Errorf("sequenceFlow %q: unknown sourceRef %q", f.ID, f.SourceRef)
		}
		childID, ok := c.ids[f.TargetRef]
		if !ok {
			return nil, fmt.Errorf("sequenceFlow %q: unknown targetRef %q", f.ID, f.TargetRef)
		}

		p := &types.WorkflowPath{
			ParentID: parentID,
			ChildID:  childID,
			Meta:     types.WorkflowPathMeta{Name: f.Name},
		}
		if f.ConditionExpr != nil {
			p.Expr = normalizeConditionExpr(f.ConditionExpr.Body)
		}
		paths = append(paths, p)
	}

	return paths, nil
}

// normalizeConditionExpr strips the "${...}" wrapper some BPMN tools put
// around condition expressions (a JUEL/FEEL convention); the inner text is
// otherwise passed through as-is; for the common case (comparisons over
// scope variables using +-*/<>=, string literals) Corteza's own gval-based
// expression syntax already matches, so no real translation is needed.
func normalizeConditionExpr(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		s = strings.TrimSpace(s[2 : len(s)-1])
	}
	return s
}

// parseISO8601Duration parses the (simple, non-recurring) subset of ISO-8601
// durations BPMN timer events use - PnYnMnDTnHnMnS - into whole seconds.
// Years/months are approximated (365/30 days) since a workflow delay is a
// wait, not a calendar computation.
func parseISO8601Duration(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "P") {
		return 0, fmt.Errorf("must start with P")
	}
	s = s[1:]

	datePart, timePart, hasTime := strings.Cut(s, "T")
	if !hasTime {
		datePart = s
		timePart = ""
	}

	var seconds int64
	readNum := func(part string, unit byte) (rest string, n float64, found bool, err error) {
		idx := strings.IndexByte(part, unit)
		if idx < 0 {
			return part, 0, false, nil
		}
		numStr := part[:idx]
		n, err = strconv.ParseFloat(numStr, 64)
		if err != nil {
			return part, 0, false, fmt.Errorf("invalid number before %q: %w", string(unit), err)
		}
		return part[idx+1:], n, true, nil
	}

	rest := datePart
	if r, n, found, err := readNum(rest, 'Y'); err != nil {
		return 0, err
	} else if found {
		seconds += int64(n * 365 * 24 * 3600)
		rest = r
	}
	if r, n, found, err := readNum(rest, 'M'); err != nil {
		return 0, err
	} else if found {
		seconds += int64(n * 30 * 24 * 3600)
		rest = r
	}
	if r, n, found, err := readNum(rest, 'D'); err != nil {
		return 0, err
	} else if found {
		seconds += int64(n * 24 * 3600)
		rest = r
	}
	if rest != "" {
		return 0, fmt.Errorf("unexpected trailing date component %q", rest)
	}

	rest = timePart
	if r, n, found, err := readNum(rest, 'H'); err != nil {
		return 0, err
	} else if found {
		seconds += int64(n * 3600)
		rest = r
	}
	if r, n, found, err := readNum(rest, 'M'); err != nil {
		return 0, err
	} else if found {
		seconds += int64(n * 60)
		rest = r
	}
	if r, n, found, err := readNum(rest, 'S'); err != nil {
		return 0, err
	} else if found {
		seconds += int64(n)
		rest = r
	}
	if rest != "" {
		return 0, fmt.Errorf("unexpected trailing time component %q", rest)
	}

	return seconds, nil
}
