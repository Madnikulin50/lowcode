package aiagent

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type fakeModel struct {
	tools   bool
	gen     []genTurn
	i       int
	lastMsg []*schema.Message
}

type genTurn struct {
	content string
	calls   []schema.ToolCall
}

func (f *fakeModel) Generate(_ context.Context, messages []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	f.lastMsg = messages
	if f.i >= len(f.gen) {
		return &schema.Message{Role: schema.Assistant, Content: "FINAL: done"}, nil
	}
	t := f.gen[f.i]
	f.i++
	return &schema.Message{Role: schema.Assistant, Content: t.content, ToolCalls: t.calls}, nil
}

func (f *fakeModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := f.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}
	sr, sw := schema.Pipe[*schema.Message](2)
	go func() {
		sw.Send(msg, nil)
		sw.Close()
	}()
	return sr, nil
}

func (f *fakeModel) IsToolsSupported() bool { return f.tools }
func (f *fakeModel) Model() string          { return "" }

func echoTool(name string) chat.ToolDef {
	return chat.ToolDef{
		Name:        name,
		Description: "echo",
		Params:      []chat.ParamDef{{Name: "q", Type: "string"}},
		Handler: func(_ context.Context, params map[string]string) string {
			return "echo:" + params["q"]
		},
	}
}

func TestRunPlainAnswer(t *testing.T) {
	m := &fakeModel{gen: []genTurn{{content: "hello"}}}
	out := Run(context.Background(), Options{Client: m, Messages: []*schema.Message{schema.UserMessage("hi")}, SkipWarm: true})
	if !out.Success || out.Output != "hello" {
		t.Fatalf("got %+v", out)
	}
}

func TestRunStripsFinalPrefix(t *testing.T) {
	m := &fakeModel{gen: []genTurn{{content: "FINAL: 42"}}}
	out := Run(context.Background(), Options{Client: m, Messages: []*schema.Message{schema.UserMessage("q")}, SkipWarm: true})
	if out.Output != "42" {
		t.Fatalf("output=%q", out.Output)
	}
}

func TestRunXMLToolThenAnswer(t *testing.T) {
	m := &fakeModel{
		tools: true,
		gen: []genTurn{
			{content: `<tool name="search"><param name="q">abc</param></tool>`},
			{content: "FINAL: found"},
		},
	}
	out := Run(context.Background(), Options{
		Client:   m,
		Messages: []*schema.Message{schema.UserMessage("find")},
		Tools:    []chat.ToolDef{echoTool("search")},
		MaxSteps: 4,
		SkipWarm: true,
	})
	if !out.Success || out.Output != "found" {
		t.Fatalf("got %+v", out)
	}
	if len(out.Steps) < 2 {
		t.Fatalf("steps=%d", len(out.Steps))
	}
	var fed string
	for _, msg := range m.lastMsg {
		fed += msg.Content
	}
	if !strings.Contains(fed, "echo:abc") {
		t.Fatalf("tool result not fed back: %q", fed)
	}
}

func TestRunNativeToolThenAnswer(t *testing.T) {
	m := &fakeModel{
		tools: true,
		gen: []genTurn{
			{calls: []schema.ToolCall{{Function: schema.FunctionCall{Name: "search", Arguments: `{"q":"xyz"}`}}}},
			{content: "ok"},
		},
	}
	out := Run(context.Background(), Options{
		Client:   m,
		Messages: []*schema.Message{schema.UserMessage("find")},
		Tools:    []chat.ToolDef{echoTool("search")},
		MaxSteps: 4,
		SkipWarm: true,
	})
	if !out.Success || out.Output != "ok" {
		t.Fatalf("got %+v", out)
	}
}

func TestRunPausesForConfirm(t *testing.T) {
	m := &fakeModel{
		tools: true,
		gen: []genTurn{
			{content: `<tool name="create_module"><param name="name">X</param></tool>`},
		},
	}
	called := false
	out := Run(context.Background(), Options{
		Client:       m,
		Messages:     []*schema.Message{schema.UserMessage("make it")},
		Tools:        []chat.ToolDef{{Name: "create_module", Handler: func(context.Context, map[string]string) string { called = true; return "created" }}},
		NeedsConfirm: DefaultNeedsConfirm,
		SkipWarm:     true,
	})
	if !out.ConfirmNeeded || called {
		t.Fatalf("confirm=%v called=%v err=%s", out.ConfirmNeeded, called, out.Error)
	}
	if len(out.ConfirmCalls) != 1 || out.ConfirmCalls[0].Name != "create_module" {
		t.Fatalf("calls=%+v", out.ConfirmCalls)
	}
}

func TestContinueFromTools(t *testing.T) {
	m := &fakeModel{
		tools: true,
		gen:   []genTurn{{content: "done"}},
	}
	called := false
	out := ContinueFromTools(context.Background(), Options{
		Client:   m,
		Messages: []*schema.Message{schema.UserMessage("да")},
		Tools:    []chat.ToolDef{{Name: "create_module", Handler: func(context.Context, map[string]string) string { called = true; return "created X" }}},
		Continue: func(_ string, result string, _ []Call) ContinueHint {
			return ContinueHint{UserMessage: "results:\n" + result, DisableTools: true}
		},
		SkipWarm: true,
	}, []Call{{Name: "create_module", Params: `{"name":"X"}`}})
	if !called {
		t.Fatal("tool not executed")
	}
	if !out.Success || out.Output != "done" {
		t.Fatalf("got %+v", out)
	}
}

func TestRunStreamHidesToolXML(t *testing.T) {
	m := &fakeModel{
		tools: true,
		gen: []genTurn{
			{content: `Sure <tool name="search"><param name="q">z</param></tool>`},
			{content: "all good"},
		},
	}
	var buf strings.Builder
	out := Run(context.Background(), Options{
		Client:   m,
		Messages: []*schema.Message{schema.UserMessage("q")},
		Tools:    []chat.ToolDef{echoTool("search")},
		MaxSteps: 4,
		Stream: func(token, reason string, done bool) error {
			buf.WriteString(token)
			return nil
		},
		HideToolXML: true,
		SkipWarm:    true,
	})
	if !out.Success {
		t.Fatalf("err=%s", out.Error)
	}
	got := buf.String()
	if strings.Contains(got, "<tool") {
		t.Fatalf("xml leaked: %q", got)
	}
	if !strings.Contains(got, "Sure") || !strings.Contains(got, "all good") {
		t.Fatalf("stream=%q", got)
	}
}

func TestCallsFromXML(t *testing.T) {
	calls := CallsFromXML(`<tool name="search"><param name="q">abc</param></tool>`)
	if len(calls) != 1 || calls[0].Name != "search" {
		t.Fatalf("%+v", calls)
	}
	if !strings.Contains(calls[0].Params, "abc") {
		t.Fatalf("params=%s", calls[0].Params)
	}
}

func TestDefaultNeedsConfirm(t *testing.T) {
	if !DefaultNeedsConfirm([]Call{{Name: "create_page"}}) {
		t.Fatal("create_ should confirm")
	}
	if DefaultNeedsConfirm([]Call{{Name: "list_modules"}}) {
		t.Fatal("list should not confirm")
	}
}

func TestUserConfirmed(t *testing.T) {
	if !UserConfirmed("да") || !UserConfirmed("yes") {
		t.Fatal("expected confirm")
	}
	if UserCancelled("да") || !UserCancelled("отмена") {
		t.Fatal("cancel mismatch")
	}
}
