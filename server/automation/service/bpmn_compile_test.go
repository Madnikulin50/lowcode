package service

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/pkg/options"
	"go.uber.org/zap"
)

func init() {
	id.Init(context.Background())
}

func findStep(t *testing.T, steps types.WorkflowStepSet, name string) *types.WorkflowStep {
	t.Helper()
	for _, s := range steps {
		if s.Meta.Name == name {
			return s
		}
	}
	t.Fatalf("no step named %q in %d steps", name, len(steps))
	return nil
}

func argValue(step *types.WorkflowStep, target string) interface{} {
	for _, a := range step.Arguments {
		if a.Target == target {
			if a.Value != nil {
				return a.Value
			}
			return a.Expr
		}
	}
	return nil
}

const bpmnGatewayFixture = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                   xmlns:lc="http://lowcode/schema/bpmn">
  <bpmn:process id="Process_1">
    <bpmn:startEvent id="Start_1" name="Start" />
    <bpmn:sequenceFlow id="Flow_1" sourceRef="Start_1" targetRef="Task_Sync" />

    <bpmn:serviceTask id="Task_Sync" name="Sync 1C">
      <bpmn:extensionElements>
        <lc:properties>
          <lc:property name="chainID" value="sync_1c_customers" />
        </lc:properties>
      </bpmn:extensionElements>
    </bpmn:serviceTask>
    <bpmn:sequenceFlow id="Flow_2" sourceRef="Task_Sync" targetRef="Gateway_1" />

    <bpmn:exclusiveGateway id="Gateway_1" name="Risk?" default="Flow_Default" />
    <bpmn:sequenceFlow id="Flow_High" sourceRef="Gateway_1" targetRef="Task_Review">
      <bpmn:conditionExpression>${risk > 7}</bpmn:conditionExpression>
    </bpmn:sequenceFlow>
    <bpmn:sequenceFlow id="Flow_Default" sourceRef="Gateway_1" targetRef="Task_AutoApprove" />

    <bpmn:userTask id="Task_Review" name="Manual review">
      <bpmn:extensionElements>
        <lc:properties>
          <lc:property name="owner" value="=reviewerID" />
        </lc:properties>
      </bpmn:extensionElements>
    </bpmn:userTask>
    <bpmn:sequenceFlow id="Flow_3" sourceRef="Task_Review" targetRef="End_1" />

    <bpmn:serviceTask id="Task_AutoApprove" name="Auto-approve">
      <bpmn:extensionElements>
        <lc:properties>
          <lc:property name="chainID" value="auto_approve" />
        </lc:properties>
      </bpmn:extensionElements>
    </bpmn:serviceTask>
    <bpmn:sequenceFlow id="Flow_4" sourceRef="Task_AutoApprove" targetRef="End_1" />

    <bpmn:endEvent id="End_1" name="End" />
  </bpmn:process>
</bpmn:definitions>`

func TestCompileBPMN_GatewayFixture(t *testing.T) {
	wf, err := CompileBPMN([]byte(bpmnGatewayFixture))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// startEvent must not become a step
	for _, s := range wf.Steps {
		if s.Meta.Name == "Start" {
			t.Fatalf("startEvent should not be compiled into a step")
		}
	}

	sync := findStep(t, wf.Steps, "Sync 1C")
	if sync.Kind != types.WorkflowStepKindFunction || sync.Ref != "compose.runRuleChain" {
		t.Fatalf("unexpected service task compile: kind=%s ref=%s", sync.Kind, sync.Ref)
	}
	if argValue(sync, "chainID") != "sync_1c_customers" {
		t.Fatalf("unexpected chainID argument: %#v", argValue(sync, "chainID"))
	}

	gw := findStep(t, wf.Steps, "Risk?")
	if gw.Kind != types.WorkflowStepKindGateway || gw.Ref != "excl" {
		t.Fatalf("unexpected gateway compile: kind=%s ref=%s", gw.Kind, gw.Ref)
	}

	review := findStep(t, wf.Steps, "Manual review")
	if review.Kind != types.WorkflowStepKindPrompt {
		t.Fatalf("expected userTask to compile to prompt, got %s", review.Kind)
	}
	if argValue(review, "owner") != "reviewerID" {
		t.Fatalf("expected owner argument to be an expression 'reviewerID', got %#v", review.Arguments)
	}
	for _, a := range review.Arguments {
		if a.Target == "owner" && a.Expr == "" {
			t.Fatalf("owner (prefixed with '=') should compile to an Expr, not a literal Value: %#v", a)
		}
	}

	end := findStep(t, wf.Steps, "End")
	if end.Kind != types.WorkflowStepKindTermination {
		t.Fatalf("expected endEvent to compile to termination, got %s", end.Kind)
	}

	// find the two paths out of the gateway; the default one must be last
	var gwPaths []*types.WorkflowPath
	for _, p := range wf.Paths {
		if p.ParentID == gw.ID {
			gwPaths = append(gwPaths, p)
		}
	}
	if len(gwPaths) != 2 {
		t.Fatalf("expected 2 outgoing paths from the gateway, got %d", len(gwPaths))
	}
	if gwPaths[0].Expr == "" {
		t.Fatalf("expected the conditional path first, default path last; got %#v", gwPaths)
	}
	if gwPaths[0].Expr != "risk > 7" {
		t.Fatalf("expected ${...} wrapper stripped, got %q", gwPaths[0].Expr)
	}
	if gwPaths[1].Expr != "" {
		t.Fatalf("expected the default path to have no condition, got %q", gwPaths[1].Expr)
	}

	// startEvent's own flow must not appear as a path (Task_Sync has no parent -> graph entry point)
	for _, p := range wf.Paths {
		if p.ChildID == sync.ID && p.ParentID == 0 {
			t.Fatalf("did not expect a path with ParentID=0")
		}
	}
	hasIncomingToSync := false
	for _, p := range wf.Paths {
		if p.ChildID == sync.ID {
			hasIncomingToSync = true
		}
	}
	if hasIncomingToSync {
		t.Fatalf("Sync 1C should have no incoming path (startEvent's flow is dropped, making it the graph's entry point)")
	}
}

const bpmnParallelFixture = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="Process_1">
    <bpmn:startEvent id="Start_1" />
    <bpmn:sequenceFlow id="f1" sourceRef="Start_1" targetRef="Fork_1" />

    <bpmn:parallelGateway id="Fork_1" name="Fork" />
    <bpmn:sequenceFlow id="f2" sourceRef="Fork_1" targetRef="Task_A" />
    <bpmn:sequenceFlow id="f3" sourceRef="Fork_1" targetRef="Task_B" />

    <bpmn:serviceTask id="Task_A" name="Task A">
      <bpmn:extensionElements><properties><property name="chainID" value="a" /></properties></bpmn:extensionElements>
    </bpmn:serviceTask>
    <bpmn:serviceTask id="Task_B" name="Task B">
      <bpmn:extensionElements><properties><property name="chainID" value="b" /></properties></bpmn:extensionElements>
    </bpmn:serviceTask>
    <bpmn:sequenceFlow id="f4" sourceRef="Task_A" targetRef="Join_1" />
    <bpmn:sequenceFlow id="f5" sourceRef="Task_B" targetRef="Join_1" />

    <bpmn:parallelGateway id="Join_1" name="Join" />
    <bpmn:sequenceFlow id="f6" sourceRef="Join_1" targetRef="End_1" />
    <bpmn:endEvent id="End_1" name="End" />
  </bpmn:process>
</bpmn:definitions>`

func TestCompileBPMN_ParallelGatewayForkJoin(t *testing.T) {
	wf, err := CompileBPMN([]byte(bpmnParallelFixture))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fork := findStep(t, wf.Steps, "Fork")
	if fork.Ref != "fork" {
		t.Fatalf("expected fork, got %s", fork.Ref)
	}
	join := findStep(t, wf.Steps, "Join")
	if join.Ref != "join" {
		t.Fatalf("expected join, got %s", join.Ref)
	}
}

const bpmnTimerFixture = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="Process_1">
    <bpmn:startEvent id="Start_1" />
    <bpmn:sequenceFlow id="f1" sourceRef="Start_1" targetRef="Wait_1" />
    <bpmn:intermediateCatchEvent id="Wait_1" name="Wait an hour">
      <bpmn:timerEventDefinition>
        <bpmn:timeDuration>PT1H</bpmn:timeDuration>
      </bpmn:timerEventDefinition>
    </bpmn:intermediateCatchEvent>
    <bpmn:sequenceFlow id="f2" sourceRef="Wait_1" targetRef="End_1" />
    <bpmn:endEvent id="End_1" />
  </bpmn:process>
</bpmn:definitions>`

func TestCompileBPMN_TimerIntermediateEvent(t *testing.T) {
	wf, err := CompileBPMN([]byte(bpmnTimerFixture))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wait := findStep(t, wf.Steps, "Wait an hour")
	if wait.Kind != types.WorkflowStepKindDelay {
		t.Fatalf("expected delay step, got %s", wait.Kind)
	}
	if argValue(wait, "duration") != int64(3600) {
		t.Fatalf("expected 3600 seconds for PT1H, got %#v", argValue(wait, "duration"))
	}
}

const bpmnMessageFixture = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="Process_1">
    <bpmn:startEvent id="Start_1" />
    <bpmn:sequenceFlow id="f1" sourceRef="Start_1" targetRef="WaitMsg_1" />
    <bpmn:intermediateCatchEvent id="WaitMsg_1" name="Wait for contractor reply">
      <bpmn:extensionElements>
        <properties>
          <property name="correlationKey" value="=orderNumber" />
        </properties>
      </bpmn:extensionElements>
      <bpmn:messageEventDefinition messageRef="ContractorReply" />
    </bpmn:intermediateCatchEvent>
    <bpmn:sequenceFlow id="f2" sourceRef="WaitMsg_1" targetRef="End_1" />
    <bpmn:endEvent id="End_1" />
  </bpmn:process>
</bpmn:definitions>`

func TestCompileBPMN_MessageIntermediateEvent(t *testing.T) {
	wf, err := CompileBPMN([]byte(bpmnMessageFixture))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wait := findStep(t, wf.Steps, "Wait for contractor reply")
	if wait.Kind != types.WorkflowStepKindPrompt {
		t.Fatalf("expected message catch event to compile to prompt, got %s", wait.Kind)
	}
	if wait.Ref != "message:ContractorReply" {
		t.Fatalf("unexpected ref: %s", wait.Ref)
	}
	if argValue(wait, "correlationKey") != "orderNumber" {
		t.Fatalf("expected correlationKey expression 'orderNumber', got %#v", argValue(wait, "correlationKey"))
	}
}

func TestCompileBPMN_MessageEventRequiresCorrelationKey(t *testing.T) {
	bad := `<?xml version="1.0"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="P">
    <bpmn:startEvent id="S" />
    <bpmn:sequenceFlow id="f1" sourceRef="S" targetRef="W" />
    <bpmn:intermediateCatchEvent id="W">
      <bpmn:messageEventDefinition messageRef="X" />
    </bpmn:intermediateCatchEvent>
    <bpmn:sequenceFlow id="f2" sourceRef="W" targetRef="E" />
    <bpmn:endEvent id="E" />
  </bpmn:process>
</bpmn:definitions>`
	if _, err := CompileBPMN([]byte(bad)); err == nil {
		t.Fatal("expected error for missing correlationKey")
	}
}

func TestCompileBPMN_ServiceTaskRequiresChainIDOrFunctionRef(t *testing.T) {
	bad := `<?xml version="1.0"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="P">
    <bpmn:startEvent id="S" />
    <bpmn:sequenceFlow id="f1" sourceRef="S" targetRef="T" />
    <bpmn:serviceTask id="T" name="Nothing configured" />
    <bpmn:sequenceFlow id="f2" sourceRef="T" targetRef="E" />
    <bpmn:endEvent id="E" />
  </bpmn:process>
</bpmn:definitions>`
	_, err := CompileBPMN([]byte(bad))
	if err == nil {
		t.Fatal("expected error for service task with no chainID/functionRef")
	}
}

func TestCompileBPMN_UnsupportedElementErrors(t *testing.T) {
	bad := `<?xml version="1.0"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="P">
    <bpmn:startEvent id="S" />
    <bpmn:sequenceFlow id="f1" sourceRef="S" targetRef="SP" />
    <bpmn:subProcess id="SP" name="Nested" />
    <bpmn:sequenceFlow id="f2" sourceRef="SP" targetRef="E" />
    <bpmn:endEvent id="E" />
  </bpmn:process>
</bpmn:definitions>`
	_, err := CompileBPMN([]byte(bad))
	if err == nil {
		t.Fatal("expected error for unsupported subProcess element")
	}
}

func TestCompileBPMN_InvalidXML(t *testing.T) {
	if _, err := CompileBPMN([]byte("not xml at all")); err == nil {
		t.Fatal("expected error for invalid XML")
	}
}

func TestParseISO8601Duration(t *testing.T) {
	cases := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{in: "PT1H", want: 3600},
		{in: "PT30M", want: 1800},
		{in: "PT1H30M", want: 5400},
		{in: "P1D", want: 86400},
		{in: "P1DT1H", want: 90000},
		{in: "PT10S", want: 10},
		{in: "garbage", wantErr: true},
		{in: "1H", wantErr: true},
	}
	for _, c := range cases {
		got, err := parseISO8601Duration(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseISO8601Duration(%q): expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseISO8601Duration(%q): unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseISO8601Duration(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// --- integration: prove the compiled output is actually a valid, executable
// wfexec graph, not just plausible-looking JSON - registers a fake function
// under the same ref the compiler emits for chainID-based service tasks,
// then runs the exact same makeGraph the hand-authored JSON editor uses.

func TestCompileBPMN_ProducesAValidExecutableGraph(t *testing.T) {
	if DefaultWorkflow == nil {
		// makeGraph's step verification reaches into this package-level
		// singleton (normally set up by full app bootstrap, see
		// automation/service/service.go) - construct a minimal stand-in so
		// the test can run standalone.
		DefaultWorkflow = Workflow(zap.NewNop(), options.CorredorOpt{}, options.WorkflowOpt{})
	}
	// Primitive expr types are likewise registered during full app bootstrap
	// (automation/service/service.go) rather than built into the registry -
	// register the subset every Expr in this fixture actually resolves to.
	Registry().AddTypes(&expr.Any{}, &expr.String{}, &expr.Vars{})

	Registry().AddFunctions(&types.Function{
		Ref:  "compose.runRuleChain",
		Kind: "function",
		Parameters: []*types.Param{
			{Name: "chainID", Types: []string{"Any"}},
			{Name: "input", Types: []string{"Any"}},
		},
		Results: []*types.Param{
			{Name: "success", Types: []string{"Any"}},
			{Name: "output", Types: []string{"Any"}},
			{Name: "error", Types: []string{"Any"}},
		},
		Handler: func(ctx context.Context, in *expr.Vars) (*expr.Vars, error) {
			return &expr.Vars{}, nil
		},
	})

	wf, err := CompileBPMN([]byte(bpmnGatewayFixture))
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	conv := workflowConverter{reg: Registry(), parser: expr.NewParser(), log: zap.NewNop()}
	g, issues := conv.makeGraph(wf)
	if len(issues) > 0 {
		for _, i := range issues {
			t.Logf("issue: %v", i)
		}
		t.Fatalf("expected a valid graph, got %d issue(s)", len(issues))
	}
	if g == nil {
		t.Fatal("expected a non-nil graph")
	}
	if g.Len() != len(wf.Steps) {
		t.Fatalf("expected all %d steps in the graph, got %d", len(wf.Steps), g.Len())
	}
}
