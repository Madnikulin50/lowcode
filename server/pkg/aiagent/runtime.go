package aiagent

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

const defaultMaxSteps = 5

// ChatModel is the LLM surface the runtime needs. *chat.Client implements it.
type ChatModel interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error)
	Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)
	IsToolsSupported() bool
	Model() string
}

// ContinueHint tells the runtime how to feed tool results into the next turn.
type ContinueHint struct {
	UserMessage  string
	DisableTools bool
	StreamPrefix string
}

type ContinueFunc func(assistantContent, toolResult string, calls []Call) ContinueHint

type Options struct {
	Client       ChatModel
	Messages     []*schema.Message
	Tools        []chat.ToolDef
	MaxSteps     int
	ExtraParams  map[string]string
	Confirmed    bool
	NeedsConfirm func([]Call) bool
	Continue     ContinueFunc
	Validator    func(*AgentResult) error
	Stream       chat.StreamFunc
	HideToolXML  bool
	EmptyAnswer  string
	SkipWarm     bool
}

func AgentContinue(_ string, toolResult string, _ []Call) ContinueHint {
	return ContinueHint{
		UserMessage: "Tool results:\n" + toolResult + "\n\nContinue with next step or provide final answer.",
	}
}

func wantTools(opt Options) bool {
	return opt.Client != nil && opt.Client.IsToolsSupported() && len(opt.Tools) > 0
}

func Run(ctx context.Context, opt Options) *AgentResult {
	return run(ctx, opt, copyMessages(opt.Messages), wantTools(opt), "")
}

// ContinueFromTools executes already-decided calls (after user confirm) and
// resumes the loop from the continuation turn.
func ContinueFromTools(ctx context.Context, opt Options, calls []Call) *AgentResult {
	start := time.Now()
	result := &AgentResult{Steps: make([]AgentStep, 0)}
	toolResult := ExecCalls(ctx, calls, opt.Tools, opt.ExtraParams)
	step := AgentStep{Type: "execute", Tools: toChatToolCalls(calls), Output: toolResult, Duration: time.Since(start).String()}
	result.Steps = append(result.Steps, step)

	hint := continueHint(opt, "", toolResult, calls)
	if err := emitPrefix(opt, hint.StreamPrefix); err != nil {
		result.Err = err
		result.Error = err.Error()
		result.Duration = time.Since(start).String()
		return result
	}
	msgs := appendToolTurn(opt.Messages, "", hint)
	bind := wantTools(opt) && !hint.DisableTools
	out := run(ctx, opt, msgs, bind, hint.StreamPrefix)
	out.Steps = append(result.Steps, out.Steps...)
	if out.Duration == "" {
		out.Duration = time.Since(start).String()
	}
	return out
}

// ContinueAfterResult skips tool execution (fast-path already has a result)
// and runs one continuation turn.
func ContinueAfterResult(ctx context.Context, opt Options, toolResult string) *AgentResult {
	hint := continueHint(opt, "", toolResult, nil)
	if err := emitPrefix(opt, hint.StreamPrefix); err != nil {
		return &AgentResult{Error: err.Error(), Err: err}
	}
	msgs := appendToolTurn(opt.Messages, "", hint)
	bind := wantTools(opt) && !hint.DisableTools
	return run(ctx, opt, msgs, bind, hint.StreamPrefix)
}

func run(ctx context.Context, opt Options, messages []*schema.Message, bindTools bool, prefix string) *AgentResult {
	start := time.Now()
	result := &AgentResult{Steps: make([]AgentStep, 0)}
	empty := opt.EmptyAnswer
	if empty == "" {
		empty = "Модель не сгенерировала ответ."
	}
	maxSteps := opt.MaxSteps
	if maxSteps <= 0 {
		maxSteps = defaultMaxSteps
	}
	if opt.Client == nil {
		result.Error = "LLM client is nil"
		result.Duration = time.Since(start).String()
		return result
	}
	if !opt.SkipWarm {
		if name := opt.Client.Model(); name != "" {
			if err := chat.EnsureWarm(ctx, name); err != nil {
				result.Error = err.Error()
				result.Err = err
				result.Duration = time.Since(start).String()
				return result
			}
		}
	}

	nativeOK := wantTools(opt)
	bind := bindTools && nativeOK
	xmlOK := bind
	shownPrefix := prefix

	for step := 0; step < maxSteps; step++ {
		stepStart := time.Now()
		agentStep := AgentStep{Type: "execute"}

		var opts []model.Option
		if bind {
			infos, err := chat.ToToolInfos(opt.Tools)
			if err != nil {
				result.Error = err.Error()
				result.Err = err
				result.Duration = time.Since(start).String()
				return result
			}
			opts = append(opts, model.WithTools(infos))
		}

		turn, err := generateTurn(ctx, opt, messages, opts)
		if err != nil {
			result.Error = err.Error()
			result.Err = err
			result.Duration = time.Since(start).String()
			return result
		}

		agentStep.Output = turn.raw
		calls := []Call(nil)
		if xmlOK {
			calls = MergeCalls(turn.native, turn.raw)
		}

		if len(calls) > 0 {
			if opt.NeedsConfirm != nil && opt.NeedsConfirm(calls) && !opt.Confirmed {
				result.ConfirmNeeded = true
				result.ConfirmCalls = calls
				result.Output = turn.raw
				agentStep.Tools = toChatToolCalls(calls)
				agentStep.Duration = time.Since(stepStart).String()
				result.Steps = append(result.Steps, agentStep)
				result.Duration = time.Since(start).String()
				return result
			}

			toolResult := ExecCalls(ctx, calls, opt.Tools, opt.ExtraParams)
			agentStep.Tools = toChatToolCalls(calls)
			agentStep.Duration = time.Since(stepStart).String()
			result.Steps = append(result.Steps, agentStep)

			hint := continueHint(opt, turn.raw, toolResult, calls)
			if err := emitPrefix(opt, hint.StreamPrefix); err != nil {
				result.Error = err.Error()
				result.Err = err
				result.Duration = time.Since(start).String()
				return result
			}
			if hint.StreamPrefix != "" {
				shownPrefix = hint.StreamPrefix
			}
			messages = appendToolTurn(messages, turn.raw, hint)
			bind = nativeOK && !hint.DisableTools
			xmlOK = bind
			continue
		}

		visible := strings.TrimSpace(turn.visible)
		if visible == "" {
			visible = strings.TrimSpace(turn.reasoning)
		}
		if visible == "" {
			visible = empty
			if opt.Stream != nil {
				_ = opt.Stream(empty, "", false)
			}
		} else if strings.TrimSpace(turn.visible) == "" && opt.Stream != nil {
			_ = opt.Stream(visible, "", false)
		}

		output := stripFinal(visible)
		output = mergePrefix(shownPrefix, output)
		agentStep.Type = "respond"
		agentStep.Duration = time.Since(stepStart).String()
		result.Steps = append(result.Steps, agentStep)
		result.Output = output
		result.Success = true

		if opt.Validator != nil {
			validateStep := AgentStep{Type: "validate", Input: result.Output}
			vStart := time.Now()
			if err := opt.Validator(result); err != nil {
				result.Success = false
				result.Error = err.Error()
				validateStep.Output = "FAILED: " + err.Error()
				messages = append(messages, schema.UserMessage("Validation failed: "+err.Error()+"\nPlease fix and try again."))
				validateStep.Duration = time.Since(vStart).String()
				result.Steps = append(result.Steps, validateStep)
				bind = nativeOK
				xmlOK = nativeOK
				continue
			}
			validateStep.Output = "OK"
			validateStep.Duration = time.Since(vStart).String()
			result.Steps = append(result.Steps, validateStep)
		}

		result.Duration = time.Since(start).String()
		return result
	}

	result.Error = "max steps exceeded"
	result.Duration = time.Since(start).String()
	return result
}

type llmTurn struct {
	raw       string
	visible   string
	reasoning string
	native    []schema.ToolCall
}

func generateTurn(ctx context.Context, opt Options, messages []*schema.Message, opts []model.Option) (llmTurn, error) {
	if opt.Stream != nil {
		sr, err := opt.Client.Stream(ctx, messages, opts...)
		if err != nil {
			return llmTurn{}, err
		}
		raw, reasoning, native, err := pumpStream(ctx, sr, opt.Stream, opt.HideToolXML)
		if err != nil {
			return llmTurn{}, err
		}
		visible := raw
		if opt.HideToolXML {
			visible = hideToolXML(raw)
		}
		return llmTurn{raw: raw, visible: visible, reasoning: reasoning, native: native}, nil
	}

	msg, err := opt.Client.Generate(ctx, messages, opts...)
	if err != nil {
		return llmTurn{}, err
	}
	if msg == nil {
		return llmTurn{}, nil
	}
	raw := msg.Content
	if raw == "" && msg.ReasoningContent != "" {
		raw = msg.ReasoningContent
	}
	return llmTurn{raw: raw, visible: raw, reasoning: msg.ReasoningContent, native: msg.ToolCalls}, nil
}

func pumpStream(ctx context.Context, streamReader *schema.StreamReader[*schema.Message], stream chat.StreamFunc, hideXML bool) (fullContent, fullReasoning string, toolCalls []schema.ToolCall, err error) {
	defer streamReader.Close()
	for {
		select {
		case <-ctx.Done():
			return fullContent, fullReasoning, toolCalls, ctx.Err()
		default:
		}
		chunk, recvErr := streamReader.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				return fullContent, fullReasoning, toolCalls, nil
			}
			if chat.IsTimeout(recvErr) {
				note := "Превышено время ожидания ответа модели."
				if strings.TrimSpace(fullContent) != "" || strings.TrimSpace(fullReasoning) != "" {
					_ = stream("\n\n"+note, "", false)
				} else {
					_ = stream(note, "", false)
					fullContent = note
				}
				return fullContent, fullReasoning, toolCalls, nil
			}
			_ = stream("⚠ Error: "+recvErr.Error(), "", false)
			return fullContent, fullReasoning, toolCalls, recvErr
		}
		if chunk.Content != "" {
			fullContent += chunk.Content
			emit := chunk.Content
			if hideXML {
				if idx := strings.Index(fullContent, "<tool "); idx >= 0 {
					already := len(fullContent) - len(chunk.Content)
					if already >= idx {
						emit = ""
					} else {
						emit = fullContent[already:idx]
					}
				}
			}
			if emit != "" {
				if err := stream(emit, "", false); err != nil {
					return fullContent, fullReasoning, toolCalls, err
				}
			}
		}
		if chunk.ReasoningContent != "" {
			fullReasoning += chunk.ReasoningContent
			if err := stream("", chunk.ReasoningContent, false); err != nil {
				return fullContent, fullReasoning, toolCalls, err
			}
		}
		if len(chunk.ToolCalls) > 0 {
			if len(toolCalls) == 0 {
				chat.EmitStatus(ctx, chat.StatusUsingTools)
			}
			toolCalls = append(toolCalls, chunk.ToolCalls...)
		}
	}
}

func hideToolXML(s string) string {
	if idx := strings.Index(s, "<tool "); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}

func continueHint(opt Options, assistant, toolResult string, calls []Call) ContinueHint {
	if opt.Continue != nil {
		return opt.Continue(assistant, toolResult, calls)
	}
	return AgentContinue(assistant, toolResult, calls)
}

func emitPrefix(opt Options, prefix string) error {
	if prefix == "" || opt.Stream == nil {
		return nil
	}
	return opt.Stream(prefix, "", false)
}

func appendToolTurn(msgs []*schema.Message, assistantContent string, hint ContinueHint) []*schema.Message {
	out := make([]*schema.Message, 0, len(msgs)+2)
	out = append(out, msgs...)
	if stripped := stripAssistantXML(assistantContent); stripped != "" {
		out = append(out, schema.AssistantMessage(stripped, nil))
	}
	if hint.UserMessage != "" {
		out = append(out, schema.UserMessage(hint.UserMessage))
	}
	return out
}

func stripAssistantXML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || chat.HasToolCallsStr(s) {
		return ""
	}
	return s
}

func mergePrefix(prefix, comment string) string {
	prefix = strings.TrimSpace(prefix)
	comment = strings.TrimSpace(comment)
	if prefix == "" {
		return comment
	}
	if comment == "" || strings.Contains(comment, "```") {
		return prefix
	}
	return prefix + "\n\n" + comment
}

func copyMessages(msgs []*schema.Message) []*schema.Message {
	if len(msgs) == 0 {
		return nil
	}
	out := make([]*schema.Message, len(msgs))
	copy(out, msgs)
	return out
}
