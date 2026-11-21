package automation

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
/// the code is regenerated from automation/automation/ai_handler.yaml

import (
	"context"
	atypes "github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/wfexec"
)

var _ wfexec.ExecResponse

type (
	aiHandlerRegistry interface {
		AddFunctions(ff ...*atypes.Function)
		Type(ref string) expr.Type
	}
)

func (h aiHandler) register() {
	h.reg.AddFunctions(
		h.Ask(),
		h.Extract(),
		h.Classify(),
		h.Tool(),
		h.RagSearch(),
		h.RunCalls(),
	)
}

type (
	aiAskArgs struct {
		hasPrompt bool
		Prompt    string

		hasAgent bool
		Agent    string

		hasModel bool
		Model    string

		hasSkill bool
		Skill    string

		hasInputs bool
		Inputs    map[string]string

		hasAllowMutating bool
		AllowMutating    bool

		hasTimeoutSec bool
		TimeoutSec    int64

		hasDeferConfirm bool
		DeferConfirm    bool
	}

	aiAskResults struct {
		Text          string
		NeedsApproval bool
		PendingCalls  string
		Summary       string
		Trace         *expr.Vars
	}
)

// Ask function AI - ask a question, get free text
//
// expects implementation of ask function:
//
//	func (h aiHandler) ask(ctx context.Context, args *aiAskArgs) (results *aiAskResults, err error) {
//	   return
//	}
func (h aiHandler) Ask() *atypes.Function {
	return &atypes.Function{
		Ref:    "aiAsk",
		Kind:   "function",
		Labels: map[string]string{"ai": "step"},
		Meta: &atypes.FunctionMeta{
			Short: "AI - ask a question, get free text",
		},

		Parameters: []*atypes.Param{
			{
				Name:  "prompt",
				Types: []string{"String"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Prompt",
					Description: "What to ask the model. Build it from earlier steps with an expression, e.g. \"Summarise: \" + ticket.description",
					Visual:      map[string]interface{}{"input": map[string]interface{}{"properties": map[string]interface{}{"rows": 6}, "type": "textarea"}},
				},
			},
			{
				Name:  "agent",
				Types: []string{"String"},
				Meta: &atypes.ParamMeta{
					Label:       "Agent",
					Description: "Which agent runs the step. Empty: a plain assistant with no tools",
				},
			},
			{
				Name:  "model",
				Types: []string{"String"},
				Meta: &atypes.ParamMeta{
					Label:       "Model",
					Description: "Model for this step only (a model name, or a role such as rulesgo.ai). Empty: the agent's own model",
				},
			},
			{
				Name:  "skill",
				Types: []string{"String"},
				Meta: &atypes.ParamMeta{
					Label:       "Skill",
					Description: "A skill the agent follows for this step: its handle, or handle@3 for a fixed version. Its instructions go into the prompt and the tools it needs become available",
				},
			},
			{
				Name:  "inputs",
				Types: []string{"KV"},
				Meta: &atypes.ParamMeta{
					Label:       "Inputs",
					Description: "Named values added to the prompt as JSON, e.g. {\"customer\": customer.name}",
				},
			},
			{
				Name:  "allowMutating",
				Types: []string{"Boolean"},
				Meta: &atypes.ParamMeta{
					Label:       "Allow data changes",
					Description: "Off by default: if the agent tries to create, update or delete something the step stops instead of doing it",
				},
			},
			{
				Name:  "timeoutSec",
				Types: []string{"Integer"},
				Meta: &atypes.ParamMeta{
					Label:       "Timeout (seconds)",
					Description: "Give up on the model after this long (default 300)",
				},
			},
			{
				Name:  "deferConfirm",
				Types: []string{"Boolean"},
				Meta: &atypes.ParamMeta{
					Label:       "Ask before data changes",
					Description: "Instead of failing, hand the proposed changes to the workflow (needsApproval, pendingCalls) so an approval step can decide, then run them with AI run calls",
				},
			},
		},

		Results: []*atypes.Param{

			{
				Name:  "text",
				Types: []string{"String"},
			},

			{
				Name:  "needsApproval",
				Types: []string{"Boolean"},
			},

			{
				Name:  "pendingCalls",
				Types: []string{"String"},
			},

			{
				Name:  "summary",
				Types: []string{"String"},
			},

			{
				Name:  "trace",
				Types: []string{"Vars"},
			},
		},

		Handler: func(ctx context.Context, in *expr.Vars) (out *expr.Vars, err error) {
			var (
				args = &aiAskArgs{
					hasPrompt:        in.Has("prompt"),
					hasAgent:         in.Has("agent"),
					hasModel:         in.Has("model"),
					hasSkill:         in.Has("skill"),
					hasInputs:        in.Has("inputs"),
					hasAllowMutating: in.Has("allowMutating"),
					hasTimeoutSec:    in.Has("timeoutSec"),
					hasDeferConfirm:  in.Has("deferConfirm"),
				}
			)

			if err = in.Decode(args); err != nil {
				return
			}

			var results *aiAskResults
			if results, err = h.ask(ctx, args); err != nil {
				return
			}

			out = &expr.Vars{}

			{
				// converting results.Text (string) to String
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("String").Cast(results.Text); err != nil {
					return
				} else if err = expr.Assign(out, "text", tval); err != nil {
					return
				}
			}

			{
				// converting results.NeedsApproval (bool) to Boolean
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("Boolean").Cast(results.NeedsApproval); err != nil {
					return
				} else if err = expr.Assign(out, "needsApproval", tval); err != nil {
					return
				}
			}

			{
				// converting results.PendingCalls (string) to String
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("String").Cast(results.PendingCalls); err != nil {
					return
				} else if err = expr.Assign(out, "pendingCalls", tval); err != nil {
					return
				}
			}

			{
				// converting results.Summary (string) to String
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("String").Cast(results.Summary); err != nil {
					return
				} else if err = expr.Assign(out, "summary", tval); err != nil {
					return
				}
			}

			{
				// converting results.Trace (*expr.Vars) to Vars
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("Vars").Cast(results.Trace); err != nil {
					return
				} else if err = expr.Assign(out, "trace", tval); err != nil {
					return
				}
			}

			return
		},
	}
}

type (
	aiExtractArgs struct {
		hasPrompt bool
		Prompt    string

		hasOutputSchema bool
		OutputSchema    map[string]string

		hasAgent bool
		Agent    string

		hasModel bool
		Model    string

		hasSkill bool
		Skill    string

		hasInputs bool
		Inputs    map[string]string

		hasAllowMutating bool
		AllowMutating    bool

		hasMaxRetries bool
		MaxRetries    int64

		hasTimeoutSec bool
		TimeoutSec    int64
	}

	aiExtractResults struct {
		Result *expr.Vars
		Trace  *expr.Vars
	}
)

// Extract function AI - extract structured data (validated JSON)
//
// expects implementation of extract function:
//
//	func (h aiHandler) extract(ctx context.Context, args *aiExtractArgs) (results *aiExtractResults, err error) {
//	   return
//	}
func (h aiHandler) Extract() *atypes.Function {
	return &atypes.Function{
		Ref:    "aiExtract",
		Kind:   "function",
		Labels: map[string]string{"ai": "step"},
		Meta: &atypes.FunctionMeta{
			Short: "AI - extract structured data (validated JSON)",
		},

		Parameters: []*atypes.Param{
			{
				Name:  "prompt",
				Types: []string{"String"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Prompt",
					Description: "What to ask the model. Build it from earlier steps with an expression, e.g. \"Summarise: \" + ticket.description",
					Visual:      map[string]interface{}{"input": map[string]interface{}{"properties": map[string]interface{}{"rows": 6}, "type": "textarea"}},
				},
			},
			{
				Name:  "outputSchema",
				Types: []string{"KV"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Output fields",
					Description: "Field name to type (string, number, boolean, array, object), e.g. {\"risk\": \"number\", \"reason\": \"string\"}",
				},
			},
			{
				Name:  "agent",
				Types: []string{"String"},
				Meta: &atypes.ParamMeta{
					Label:       "Agent",
					Description: "Which agent runs the step. Empty: a plain assistant with no tools",
				},
			},
			{
				Name:  "model",
				Types: []string{"String"},
				Meta: &atypes.ParamMeta{
					Label:       "Model",
					Description: "Model for this step only (a model name, or a role such as rulesgo.ai). Empty: the agent's own model",
				},
			},
			{
				Name:  "skill",
				Types: []string{"String"},
				Meta: &atypes.ParamMeta{
					Label:       "Skill",
					Description: "A skill the agent follows for this step: its handle, or handle@3 for a fixed version. Its instructions go into the prompt and the tools it needs become available",
				},
			},
			{
				Name:  "inputs",
				Types: []string{"KV"},
				Meta: &atypes.ParamMeta{
					Label:       "Inputs",
					Description: "Named values added to the prompt as JSON, e.g. {\"customer\": customer.name}",
				},
			},
			{
				Name:  "allowMutating",
				Types: []string{"Boolean"},
				Meta: &atypes.ParamMeta{
					Label:       "Allow data changes",
					Description: "Off by default: if the agent tries to create, update or delete something the step stops instead of doing it",
				},
			},
			{
				Name:  "maxRetries",
				Types: []string{"Integer"},
				Meta: &atypes.ParamMeta{
					Label:       "Max retries",
					Description: "Ask again this many times when the answer is not valid JSON of the right shape (default 1)",
				},
			},
			{
				Name:  "timeoutSec",
				Types: []string{"Integer"},
				Meta: &atypes.ParamMeta{
					Label:       "Timeout (seconds)",
					Description: "Give up on the model after this long (default 300)",
				},
			},
		},

		Results: []*atypes.Param{

			{
				Name:  "result",
				Types: []string{"Vars"},
			},

			{
				Name:  "trace",
				Types: []string{"Vars"},
			},
		},

		Handler: func(ctx context.Context, in *expr.Vars) (out *expr.Vars, err error) {
			var (
				args = &aiExtractArgs{
					hasPrompt:        in.Has("prompt"),
					hasOutputSchema:  in.Has("outputSchema"),
					hasAgent:         in.Has("agent"),
					hasModel:         in.Has("model"),
					hasSkill:         in.Has("skill"),
					hasInputs:        in.Has("inputs"),
					hasAllowMutating: in.Has("allowMutating"),
					hasMaxRetries:    in.Has("maxRetries"),
					hasTimeoutSec:    in.Has("timeoutSec"),
				}
			)

			if err = in.Decode(args); err != nil {
				return
			}

			var results *aiExtractResults
			if results, err = h.extract(ctx, args); err != nil {
				return
			}

			out = &expr.Vars{}

			{
				// converting results.Result (*expr.Vars) to Vars
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("Vars").Cast(results.Result); err != nil {
					return
				} else if err = expr.Assign(out, "result", tval); err != nil {
					return
				}
			}

			{
				// converting results.Trace (*expr.Vars) to Vars
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("Vars").Cast(results.Trace); err != nil {
					return
				} else if err = expr.Assign(out, "trace", tval); err != nil {
					return
				}
			}

			return
		},
	}
}

type (
	aiClassifyArgs struct {
		hasText bool
		Text    string

		hasLabels bool
		Labels    string

		hasInstruction bool
		Instruction    string

		hasModel bool
		Model    string

		hasSkill bool
		Skill    string

		hasMaxRetries bool
		MaxRetries    int64

		hasTimeoutSec bool
		TimeoutSec    int64
	}

	aiClassifyResults struct {
		Label      string
		Confidence float64
		Reason     string
		Trace      *expr.Vars
	}
)

// Classify function AI - classify text into one of the given labels
//
// expects implementation of classify function:
//
//	func (h aiHandler) classify(ctx context.Context, args *aiClassifyArgs) (results *aiClassifyResults, err error) {
//	   return
//	}
func (h aiHandler) Classify() *atypes.Function {
	return &atypes.Function{
		Ref:    "aiClassify",
		Kind:   "function",
		Labels: map[string]string{"ai": "step"},
		Meta: &atypes.FunctionMeta{
			Short: "AI - classify text into one of the given labels",
		},

		Parameters: []*atypes.Param{
			{
				Name:  "text",
				Types: []string{"String"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Text",
					Description: "The text to classify",
					Visual:      map[string]interface{}{"input": map[string]interface{}{"properties": map[string]interface{}{"rows": 4}, "type": "textarea"}},
				},
			},
			{
				Name:  "labels",
				Types: []string{"String"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Labels",
					Description: "The allowed answers, separated by commas or new lines",
				},
			},
			{
				Name:  "instruction",
				Types: []string{"String"},
				Meta: &atypes.ParamMeta{
					Label:       "Instruction",
					Description: "How to decide (optional), e.g. which label means what",
					Visual:      map[string]interface{}{"input": map[string]interface{}{"properties": map[string]interface{}{"rows": 3}, "type": "textarea"}},
				},
			},
			{
				Name:  "model",
				Types: []string{"String"},
				Meta: &atypes.ParamMeta{
					Label:       "Model",
					Description: "Model for this step only (a model name, or a role such as rulesgo.ai). Empty: the agent's own model",
				},
			},
			{
				Name:  "skill",
				Types: []string{"String"},
				Meta: &atypes.ParamMeta{
					Label:       "Skill",
					Description: "A skill the agent follows for this step: its handle, or handle@3 for a fixed version. Its instructions go into the prompt and the tools it needs become available",
				},
			},
			{
				Name:  "maxRetries",
				Types: []string{"Integer"},
				Meta: &atypes.ParamMeta{
					Label:       "Max retries",
					Description: "Ask again this many times when the answer is not valid JSON of the right shape (default 1)",
				},
			},
			{
				Name:  "timeoutSec",
				Types: []string{"Integer"},
				Meta: &atypes.ParamMeta{
					Label:       "Timeout (seconds)",
					Description: "Give up on the model after this long (default 300)",
				},
			},
		},

		Results: []*atypes.Param{

			{
				Name:  "label",
				Types: []string{"String"},
			},

			{
				Name:  "confidence",
				Types: []string{"Float"},
			},

			{
				Name:  "reason",
				Types: []string{"String"},
			},

			{
				Name:  "trace",
				Types: []string{"Vars"},
			},
		},

		Handler: func(ctx context.Context, in *expr.Vars) (out *expr.Vars, err error) {
			var (
				args = &aiClassifyArgs{
					hasText:        in.Has("text"),
					hasLabels:      in.Has("labels"),
					hasInstruction: in.Has("instruction"),
					hasModel:       in.Has("model"),
					hasSkill:       in.Has("skill"),
					hasMaxRetries:  in.Has("maxRetries"),
					hasTimeoutSec:  in.Has("timeoutSec"),
				}
			)

			if err = in.Decode(args); err != nil {
				return
			}

			var results *aiClassifyResults
			if results, err = h.classify(ctx, args); err != nil {
				return
			}

			out = &expr.Vars{}

			{
				// converting results.Label (string) to String
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("String").Cast(results.Label); err != nil {
					return
				} else if err = expr.Assign(out, "label", tval); err != nil {
					return
				}
			}

			{
				// converting results.Confidence (float64) to Float
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("Float").Cast(results.Confidence); err != nil {
					return
				} else if err = expr.Assign(out, "confidence", tval); err != nil {
					return
				}
			}

			{
				// converting results.Reason (string) to String
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("String").Cast(results.Reason); err != nil {
					return
				} else if err = expr.Assign(out, "reason", tval); err != nil {
					return
				}
			}

			{
				// converting results.Trace (*expr.Vars) to Vars
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("Vars").Cast(results.Trace); err != nil {
					return
				} else if err = expr.Assign(out, "trace", tval); err != nil {
					return
				}
			}

			return
		},
	}
}

type (
	aiToolArgs struct {
		hasTool bool
		Tool    string

		hasParams bool
		Params    map[string]string

		hasAllowMutating bool
		AllowMutating    bool
	}

	aiToolResults struct {
		Result string
	}
)

// Tool function AI - call a platform tool directly (no LLM)
//
// expects implementation of tool function:
//
//	func (h aiHandler) tool(ctx context.Context, args *aiToolArgs) (results *aiToolResults, err error) {
//	   return
//	}
func (h aiHandler) Tool() *atypes.Function {
	return &atypes.Function{
		Ref:    "aiTool",
		Kind:   "function",
		Labels: map[string]string{"ai": "step"},
		Meta: &atypes.FunctionMeta{
			Short: "AI - call a platform tool directly (no LLM)",
		},

		Parameters: []*atypes.Param{
			{
				Name:  "tool",
				Types: []string{"String"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Tool",
					Description: "Name of the platform tool to call, e.g. cmdb_scan, backup_run, search_records",
				},
			},
			{
				Name:  "params",
				Types: []string{"KV"},
				Meta: &atypes.ParamMeta{
					Label:       "Tool parameters",
					Description: "Parameter name to value, as the tool expects them",
				},
			},
			{
				Name:  "allowMutating",
				Types: []string{"Boolean"},
				Meta: &atypes.ParamMeta{
					Label:       "Allow data changes",
					Description: "Off by default: if the agent tries to create, update or delete something the step stops instead of doing it",
				},
			},
		},

		Results: []*atypes.Param{

			{
				Name:  "result",
				Types: []string{"String"},
			},
		},

		Handler: func(ctx context.Context, in *expr.Vars) (out *expr.Vars, err error) {
			var (
				args = &aiToolArgs{
					hasTool:          in.Has("tool"),
					hasParams:        in.Has("params"),
					hasAllowMutating: in.Has("allowMutating"),
				}
			)

			if err = in.Decode(args); err != nil {
				return
			}

			var results *aiToolResults
			if results, err = h.tool(ctx, args); err != nil {
				return
			}

			out = &expr.Vars{}

			{
				// converting results.Result (string) to String
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("String").Cast(results.Result); err != nil {
					return
				} else if err = expr.Assign(out, "result", tval); err != nil {
					return
				}
			}

			return
		},
	}
}

type (
	aiRagSearchArgs struct {
		hasNamespace bool
		Namespace    string

		hasQuery bool
		Query    string

		hasTopK bool
		TopK    int64
	}

	aiRagSearchResults struct {
		Text  string
		Count int64
	}
)

// RagSearch function AI - search the knowledge base (RAG)
//
// expects implementation of ragSearch function:
//
//	func (h aiHandler) ragSearch(ctx context.Context, args *aiRagSearchArgs) (results *aiRagSearchResults, err error) {
//	   return
//	}
func (h aiHandler) RagSearch() *atypes.Function {
	return &atypes.Function{
		Ref:    "aiRagSearch",
		Kind:   "function",
		Labels: map[string]string{"ai": "step"},
		Meta: &atypes.FunctionMeta{
			Short: "AI - search the knowledge base (RAG)",
		},

		Parameters: []*atypes.Param{
			{
				Name:  "namespace",
				Types: []string{"String"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Namespace",
					Description: "ID of the namespace whose knowledge base to search",
				},
			},
			{
				Name:  "query",
				Types: []string{"String"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Query",
					Description: "What to look for",
				},
			},
			{
				Name:  "topK",
				Types: []string{"Integer"},
				Meta: &atypes.ParamMeta{
					Label:       "How many passages",
					Description: "Number of passages to return (default 5)",
				},
			},
		},

		Results: []*atypes.Param{

			{
				Name:  "text",
				Types: []string{"String"},
			},

			{
				Name:  "count",
				Types: []string{"Integer"},
			},
		},

		Handler: func(ctx context.Context, in *expr.Vars) (out *expr.Vars, err error) {
			var (
				args = &aiRagSearchArgs{
					hasNamespace: in.Has("namespace"),
					hasQuery:     in.Has("query"),
					hasTopK:      in.Has("topK"),
				}
			)

			if err = in.Decode(args); err != nil {
				return
			}

			var results *aiRagSearchResults
			if results, err = h.ragSearch(ctx, args); err != nil {
				return
			}

			out = &expr.Vars{}

			{
				// converting results.Text (string) to String
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("String").Cast(results.Text); err != nil {
					return
				} else if err = expr.Assign(out, "text", tval); err != nil {
					return
				}
			}

			{
				// converting results.Count (int64) to Integer
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("Integer").Cast(results.Count); err != nil {
					return
				} else if err = expr.Assign(out, "count", tval); err != nil {
					return
				}
			}

			return
		},
	}
}

type (
	aiRunCallsArgs struct {
		hasAgent bool
		Agent    string

		hasCalls bool
		Calls    string
	}

	aiRunCallsResults struct {
		Result string
	}
)

// RunCalls function AI - run tool calls a human approved
//
// expects implementation of runCalls function:
//
//	func (h aiHandler) runCalls(ctx context.Context, args *aiRunCallsArgs) (results *aiRunCallsResults, err error) {
//	   return
//	}
func (h aiHandler) RunCalls() *atypes.Function {
	return &atypes.Function{
		Ref:    "aiRunCalls",
		Kind:   "function",
		Labels: map[string]string{"ai": "step"},
		Meta: &atypes.FunctionMeta{
			Short: "AI - run tool calls a human approved",
		},

		Parameters: []*atypes.Param{
			{
				Name:  "agent",
				Types: []string{"String"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Agent",
					Description: "Which agent runs the step. Empty: a plain assistant with no tools",
				},
			},
			{
				Name:  "calls",
				Types: []string{"String"}, Required: true,
				Meta: &atypes.ParamMeta{
					Label:       "Calls",
					Description: "The pendingCalls value produced by AI ask, after a human approved it",
				},
			},
		},

		Results: []*atypes.Param{

			{
				Name:  "result",
				Types: []string{"String"},
			},
		},

		Handler: func(ctx context.Context, in *expr.Vars) (out *expr.Vars, err error) {
			var (
				args = &aiRunCallsArgs{
					hasAgent: in.Has("agent"),
					hasCalls: in.Has("calls"),
				}
			)

			if err = in.Decode(args); err != nil {
				return
			}

			var results *aiRunCallsResults
			if results, err = h.runCalls(ctx, args); err != nil {
				return
			}

			out = &expr.Vars{}

			{
				// converting results.Result (string) to String
				var (
					tval expr.TypedValue
				)

				if tval, err = h.reg.Type("String").Cast(results.Result); err != nil {
					return
				} else if err = expr.Assign(out, "result", tval); err != nil {
					return
				}
			}

			return
		},
	}
}
