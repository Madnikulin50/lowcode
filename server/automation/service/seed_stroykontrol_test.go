package service

// AI workflows for the stroykontrol namespaces ("Стройконтроль ПТО" and
// "Проверка ИД (АОСР)"): a worked example of the whole AI stack on a real
// domain - flows tied to the modules of a namespace, prompts from the library
// with golden examples, a budget on every flow.
//
// TestStroykontrolFlows_* always runs: it compiles every flow and runs it on
// records taken from a real database, with a scripted model.
//
// TestSeedStroykontrol installs them into a database. It does nothing unless
// SEED_DSN is set:
//
//	SEED_DSN=postgres://... go test ./automation/service -run TestSeedStroykontrol -v
//
// and it can be run again: what already exists is left alone.

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/automation/types"
	composeEvent "github.com/madnikulin50/lowcode/server/compose/service/event"
	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/label"
	labelTypes "github.com/madnikulin50/lowcode/server/pkg/label/types"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/postgres"
	"github.com/stretchr/testify/require"
)

type skPrompt struct {
	handle, description, text string
	cases                     types.PromptCases
}

var skPrompts = []skPrompt{
	{
		handle:      "sk_id_finding_explain",
		description: "Разбор замечания проверки исполнительной документации (ИД)",
		text: `Ты помощник инженера ПТО. Тебе дано замечание автоматической проверки исполнительной документации (акты скрытых работ, АОСР).
Объясни простым языком, в чём проблема (2-3 предложения), назови одно конкретное действие, что проверить или исправить, и оцени срочность:
- high: замечание блокирует приёмку работ;
- medium: нужно исправить до сдачи пакета;
- low: можно оставить как есть.
Опирайся только на данные замечания, ничего не выдумывай. Отвечай по-русски.`,
		cases: types.PromptCases{
			{Name: "номер акта не соответствует позиции ВРЦ", Inputs: map[string]interface{}{
				"severity": "medium", "check_name": "Акт ↔ реестр (номер, дата, объёмы кол. 12/13)", "field": "doc_number",
				"expected": "позиция 2.14.2.6", "actual": "3/2.14.1.6/ОП№5/ИССО 2.2/СКТРАСТ",
				"description": "Реестр, строка 30: номер акта не соответствует позиции ВРЦ 2.14.2.6, под которой он указан",
			}, Expect: map[string]interface{}{"priority": "medium"}},
			{Name: "нет подписи ответственного лица", Inputs: map[string]interface{}{
				"severity": "high", "check_name": "Участники и полномочия", "field": "participants",
				"expected": "подпись представителя заказчика", "actual": "подпись отсутствует",
				"description": "В акте скрытых работ отсутствует подпись представителя заказчика - акт не может быть принят",
			}, Expect: map[string]interface{}{"priority": "high"}},
			{Name: "расхождение в формате даты", Inputs: map[string]interface{}{
				"severity": "low", "check_name": "Акт ↔ реестр (номер, дата, объёмы кол. 12/13)", "field": "doc_date",
				"expected": "12.03.2024", "actual": "12.3.24", "description": "Дата акта записана в другом формате, значение совпадает",
			}, Expect: map[string]interface{}{"priority": "low"}},
		},
	},
	{
		handle:      "sk_id_package_summary",
		description: "Сводка по результатам проверки пакета ИД",
		text: `Ты помощник руководителя группы ПТО. Кратко (не больше пяти предложений) подведи итог проверки пакета исполнительной документации:
сколько актов проверено, сколько из них без замечаний, сколько замечаний высокой, средней и низкой серьёзности, и что сделать в первую очередь.
Используй только переданные числа. Отвечай по-русски.`,
		cases: types.PromptCases{
			{Name: "пакет с замечаниями", Inputs: map[string]interface{}{
				"title": "Пакет КС-2 №7", "acts_total": "120", "acts_ok": "90", "acts_with_issues": "30",
				"findings_high": "4", "findings_medium": "21", "findings_low": "9",
			}, Contains: []string{"120", "30"}},
			{Name: "чистый пакет", Inputs: map[string]interface{}{
				"title": "Пакет КС-2 №8", "acts_total": "15", "acts_ok": "15", "acts_with_issues": "0",
				"findings_high": "0", "findings_medium": "0", "findings_low": "0",
			}, Contains: []string{"15"}},
		},
	},
	{
		handle:      "sk_pd_rd_review",
		description: "Оценка сравнения проектной и рабочей документации (ПД/РД)",
		text: `Ты помощник главного инженера проекта. Тебе даны итоги автоматического сравнения проектной документации (ПД) и рабочей документации (РД).
В трёх-четырёх предложениях скажи, насколько документы согласованы, на что обратить внимание (число страниц, расхождение страниц) и какой следующий шаг разумен.
Если данных для вывода мало (сравнение не завершено), прямо скажи об этом. Используй только переданные данные. Отвечай по-русски.`,
		cases: types.PromptCases{
			{Name: "высокая схожесть", Inputs: map[string]interface{}{
				"title": "Сравнение ПД/РД - Насосная станция №8", "status": "processing", "similarity_percent": "94.5",
				"total_pages_pd": "116", "total_pages_rd": "119", "matching_pages": "31", "differing_pages": "4",
			}, Contains: []string{"94"}},
			{Name: "сравнение не начато", Inputs: map[string]interface{}{
				"title": "Сравнение ПД/РД - Поликлиника №1", "status": "new", "total_pages_pd": "70", "total_pages_rd": "70",
			}, Contains: []string{"70"}},
		},
	},
	{
		handle:      "sk_smeta_check_explain",
		description: "Разбор расхождения объёма и стоимости ВОИСР и сметы",
		text: `Ты помощник сметчика. Тебе дан результат сверки позиции ведомости объёмов и стоимости (ВОИСР) со сметой.
Объясни расхождение простыми словами, скажи, что проверить в первую очередь, и нужен ли разбор специалистом (needs_review): true, если расхождение по сумме больше 5% по модулю или статус "не совпадает"; иначе false.
Опирайся только на переданные числа. Отвечай по-русски.`,
		cases: types.PromptCases{
			{Name: "расхождение -15%", Inputs: map[string]interface{}{
				"status": "mismatch", "smeta_volume": "431", "smeta_amount": "4884523", "volume_diff": "-64.65",
				"amount_diff": "-732678.45", "amount_diff_percent": "-15",
			}, Expect: map[string]interface{}{"needs_review": true}},
			{Name: "совпадает", Inputs: map[string]interface{}{
				"status": "match", "smeta_volume": "100", "smeta_amount": "250000", "volume_diff": "0",
				"amount_diff": "0", "amount_diff_percent": "0",
			}, Expect: map[string]interface{}{"needs_review": false}},
		},
	},
	{
		handle:      "sk_triage_classify",
		description: "Срочность замечания (для автоматической сортировки)",
		text: `Определи срочность замечания по исполнительной или проектной документации.
блокирующее - не позволяет принять работы или документы; существенное - надо исправить до сдачи; незначительное - формальность, не влияет на приёмку.`,
	},
}

// skFlow is one flow and where it belongs
type skFlow struct {
	namespace string // slug
	module    string // handle
	manual    bool   // a button on the module's records; otherwise reacts to new records (and starts switched off)
	spec      AIFlowSpec
}

// fieldsExpr builds {"f": coalesce(record.values.f, ""), ...}: the record's
// fields as the named inputs of an AI step. A field the record does not have
// (empty values are absent) is passed as an empty string, not as an error.
func fieldsExpr(fields ...string) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = fmt.Sprintf(`"%s": coalesce(record.values.%s, "")`, f, f)
	}
	return "={" + strings.Join(parts, ", ") + "}"
}

func skFlows() []skFlow {
	idFindingFields := []string{"severity", "check_name", "field", "expected", "actual", "description"}

	return []skFlow{
		{
			namespace: "stroykontrol-id", module: "findings", manual: true,
			spec: AIFlowSpec{
				Name:        "ИИ: разобрать замечание",
				Handle:      "ai_id_finding_explain",
				Description: "Кнопка на записи замечания: объясняет, в чём проблема, что исправить и насколько это срочно.",
				Steps: []AIFlowStep{
					{ID: "explain", Name: "Разбор замечания", Function: "aiExtract", Args: map[string]interface{}{
						"prompt":       "@prompt:sk_id_finding_explain",
						"inputs":       fieldsExpr(idFindingFields...),
						"outputSchema": map[string]interface{}{"explanation": "string", "action": "string", "priority": "string"},
						"timeoutSec":   120,
					}, Results: map[string]string{"analysis": "result"}},
					{ID: "show", Name: "Показать результат", Prompt: "alert", Args: map[string]interface{}{
						"title":   "ИИ-разбор замечания",
						"message": `=format("%v\n\nЧто сделать: %v\nСрочность: %v", analysis.explanation, analysis.action, analysis.priority)`,
					}},
				},
				Flow: []AIFlowLink{{From: "explain", To: "show"}},
			},
		},
		{
			namespace: "stroykontrol-id", module: "packages", manual: true,
			spec: AIFlowSpec{
				Name:        "ИИ: сводка по пакету ИД",
				Handle:      "ai_id_package_summary",
				Description: "Кнопка на записи пакета: краткий итог проверки и с чего начать.",
				Steps: []AIFlowStep{
					{ID: "summary", Name: "Сводка", Function: "aiAsk", Args: map[string]interface{}{
						"prompt": "@prompt:sk_id_package_summary",
						"inputs": fieldsExpr("title", "status", "acts_total", "acts_ok", "acts_with_issues", "findings_high", "findings_medium", "findings_low"),
					}, Results: map[string]string{"text": "text"}},
					{ID: "show", Name: "Показать результат", Prompt: "alert", Args: map[string]interface{}{
						"title": "ИИ-сводка по пакету", "message": "=text",
					}},
				},
				Flow: []AIFlowLink{{From: "summary", To: "show"}},
			},
		},
		{
			namespace: "stroykontrol-id", module: "findings", manual: false,
			spec: AIFlowSpec{
				Name:        "ИИ: сортировать новые замечания",
				Handle:      "ai_id_finding_triage",
				Description: "Реагирует на новое замечание и определяет его срочность. Выключен: включите, когда будете готовы к автоматической обработке.",
				Steps: []AIFlowStep{
					{ID: "triage", Name: "Срочность", Function: "aiClassify", Args: map[string]interface{}{
						"text":        `=coalesce(record.values.description, "")`,
						"labels":      "блокирующее, существенное, незначительное",
						"instruction": "@prompt:sk_triage_classify",
					}, Results: map[string]string{"urgency": "label", "reason": "reason"}},
					{ID: "log", Name: "В журнал", Function: "logInfo", Args: map[string]interface{}{
						"message": `=format("Замечание ИД: %v - %v", urgency, reason)`,
					}},
				},
				Flow: []AIFlowLink{{From: "triage", To: "log"}},
			},
		},
		{
			namespace: "stroykontrol", module: "pd_rd_comparisons", manual: true,
			spec: AIFlowSpec{
				Name:        "ИИ: оценить сравнение ПД/РД",
				Handle:      "ai_pd_rd_review",
				Description: "Кнопка на записи сравнения: насколько документы согласованы и что делать дальше.",
				Steps: []AIFlowStep{
					{ID: "review", Name: "Оценка", Function: "aiAsk", Args: map[string]interface{}{
						"prompt": "@prompt:sk_pd_rd_review",
						"inputs": fieldsExpr("title", "status", "similarity_percent", "total_pages_pd", "total_pages_rd", "matching_pages", "differing_pages", "comment"),
					}, Results: map[string]string{"text": "text"}},
					{ID: "show", Name: "Показать результат", Prompt: "alert", Args: map[string]interface{}{
						"title": "ИИ-оценка сравнения ПД/РД", "message": "=text",
					}},
				},
				Flow: []AIFlowLink{{From: "review", To: "show"}},
			},
		},
		{
			namespace: "stroykontrol", module: "voisr_check_results", manual: true,
			spec: AIFlowSpec{
				Name:        "ИИ: разобрать расхождение со сметой",
				Handle:      "ai_smeta_check_explain",
				Description: "Кнопка на результате сверки ВОИСР со сметой: объясняет расхождение и говорит, нужен ли разбор специалистом.",
				Steps: []AIFlowStep{
					{ID: "explain", Name: "Разбор расхождения", Function: "aiExtract", Args: map[string]interface{}{
						"prompt":       "@prompt:sk_smeta_check_explain",
						"inputs":       fieldsExpr("status", "smeta_volume", "smeta_amount", "volume_diff", "amount_diff", "amount_diff_percent", "note"),
						"outputSchema": map[string]interface{}{"explanation": "string", "action": "string", "needs_review": "boolean"},
						"timeoutSec":   120,
					}, Results: map[string]string{"analysis": "result"}},
					{ID: "show", Name: "Показать результат", Prompt: "alert", Args: map[string]interface{}{
						"title":   "ИИ-разбор расхождения со сметой",
						"message": `=format("%v\n\nЧто проверить: %v\nНужен разбор специалистом: %v", analysis.explanation, analysis.action, analysis.needs_review)`,
					}},
				},
				Flow: []AIFlowLink{{From: "explain", To: "show"}},
			},
		},
		{
			namespace: "stroykontrol", module: "pd_rd_discrepancies", manual: false,
			spec: AIFlowSpec{
				Name:        "ИИ: сортировать новые расхождения ПД/РД",
				Handle:      "ai_pd_rd_discrepancy_triage",
				Description: "Реагирует на новое расхождение ПД/РД и определяет его срочность. Выключен: включите, когда будете готовы к автоматической обработке.",
				Steps: []AIFlowStep{
					{ID: "triage", Name: "Срочность", Function: "aiClassify", Args: map[string]interface{}{
						"text":        `=coalesce(record.values.description, "")`,
						"labels":      "блокирующее, существенное, незначительное",
						"instruction": "@prompt:sk_triage_classify",
					}, Results: map[string]string{"urgency": "label", "reason": "reason"}},
					{ID: "log", Name: "В журнал", Function: "logInfo", Args: map[string]interface{}{
						"message": `=format("Расхождение ПД/РД: %v - %v", urgency, reason)`,
					}},
				},
				Flow: []AIFlowLink{{From: "triage", To: "log"}},
			},
		},
	}
}

// skRecords are records as they are in the database (values as the
// compose_record table keeps them), one for each module the flows use.
var skRecords = map[string]map[string][]string{
	"findings": {
		"severity": {"medium"}, "check_no": {"1"}, "check_name": {"Акт ↔ реестр (номер, дата, объёмы кол. 12/13)"},
		"field": {"doc_number"}, "expected": {"позиция 2.14.2.6"}, "actual": {"3/2.14.1.6/ОП№5/ИССО 2.2/СКТРАСТ"},
		"description": {`Реестр, строка 30: номер акта "3/2.14.1.6/ОП№5/ИССО 2.2/СКТРАСТ" не соответствует позиции ВРЦ 2.14.2.6, под которой он указан`},
	},
	"packages": {
		"title": {"Пакет КС-2 №7"}, "status": {"done"}, "acts_total": {"120"}, "acts_ok": {"90"}, "acts_with_issues": {"30"},
		"findings_high": {"4"}, "findings_medium": {"21"}, "findings_low": {"9"},
	},
	"pd_rd_comparisons": {
		"title": {"Сравнение ПД/РД - Насосная станция №8"}, "status": {"processing"}, "similarity_percent": {"94.5"},
		"total_pages_pd": {"116"}, "total_pages_rd": {"119"}, "matching_pages": {"31"}, "differing_pages": {"4"},
		"comment": {"Агент: 94.45% схожести."},
	},
	"voisr_check_results": {
		"status": {"mismatch"}, "smeta_volume": {"431"}, "smeta_amount": {"4884523"}, "volume_diff": {"-64.65"},
		"amount_diff": {"-732678.45"}, "amount_diff_percent": {"-15"},
	},
	"pd_rd_discrepancies": {
		"description": {"На листе 12 в РД добавлена дополнительная строка спецификации, которой нет в ПД"}, "severity": {"medium"},
	},
}

func skRecord(module string) *composeTypes.Record {
	rec := &composeTypes.Record{ID: 1}
	for name, vals := range skRecords[module] {
		for i, v := range vals {
			rec.Values = append(rec.Values, &composeTypes.RecordValue{Name: name, Value: v, Place: uint(i)})
		}
	}
	return rec
}

// skModel plays the language model: it answers each kind of step the way a
// good model would, and remembers what it was asked.
func skModel(asked *[]string) aiagent.Runner {
	return func(_ context.Context, _, prompt string, _ bool) (*aiagent.AgentResult, error) {
		*asked = append(*asked, prompt)

		var out string
		switch {
		case strings.Contains(prompt, "Choose exactly one label"):
			out = `{"label":"существенное","confidence":0.8,"reason":"нужно исправить до сдачи"}`
		case strings.Contains(prompt, `"needs_review"`):
			out = `{"explanation":"Объём по ВОИСР на 15% меньше сметного.","action":"Сверить позицию со сметой.","needs_review":true}`
		case strings.Contains(prompt, `"priority"`):
			out = `{"explanation":"Номер акта не совпадает с позицией ВРЦ.","action":"Проверить номер акта в реестре.","priority":"medium"}`
		default:
			out = "Итог проверки получен."
		}
		return &aiagent.AgentResult{Success: true, Output: out, Model: "scripted", LLMCalls: 1}, nil
	}
}

func skLibrary() func(context.Context, string, int) (*aiagent.ResolvedPrompt, error) {
	return func(_ context.Context, handle string, _ int) (*aiagent.ResolvedPrompt, error) {
		for _, p := range skPrompts {
			if p.handle == handle {
				return &aiagent.ResolvedPrompt{Text: p.text, Ref: handle + "@1"}, nil
			}
		}
		return nil, fmt.Errorf("prompt %q does not exist", handle)
	}
}

func TestStroykontrolFlows_CompileAndRunOnRealRecords(t *testing.T) {
	aiagent.SetPromptResolver(skLibrary())
	defer aiagent.SetPromptResolver(nil)

	for _, f := range skFlows() {
		t.Run(f.spec.Handle, func(t *testing.T) {
			var asked []string
			svc := templateWorkflowService(t, skModel(&asked))
			req := require.New(t)

			c := CompileAIFlow(svc, &f.spec, 9)
			req.Empty(c.Issues, issuesText(c))
			c.Workflow.ID = 700

			g, issues := Convert(svc, c.Workflow)
			req.Empty(issues)

			// the variables the platform hands a workflow for this record
			module := &composeTypes.Module{ID: 2, Handle: f.module}
			ns := &composeTypes.Namespace{ID: 3, Slug: f.namespace}
			var input *expr.Vars
			var err error
			if f.manual {
				input, err = composeEvent.RecordOnManual(skRecord(f.module), nil, module, ns, nil, nil).EncodeVars()
			} else {
				input, err = composeEvent.RecordAfterCreate(skRecord(f.module), nil, module, ns, nil, nil).EncodeVars()
			}
			req.NoError(err)

			ses, stop := sessionService(t)
			defer stop()
			ctx := context.Background()

			wait, sessionID, err := ses.Start(ctx, g, types.SessionStartParams{Invoker: auth.Authenticated(9), WorkflowID: 700, Input: input})
			req.NoError(err)

			if f.manual {
				// ends at an alert shown to the person who pressed the button
				waitFor(t, "the alert", func() bool { return len(ses.states.ids(sessionID)) == 1 })
				pending := ses.PendingPrompts(auth.SetIdentityToContext(ctx, auth.Authenticated(9)))
				req.Len(pending, 1)
				req.Equal("alert", pending[0].Ref)
				msg, _ := expr.Must(expr.Select(pending[0].Payload, "message")).Get().(string)
				req.NotEmpty(strings.TrimSpace(msg), "the alert must say something")
				t.Logf("alert: %s", msg)
			} else {
				wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				_, _, _, _, err = wait(wctx)
				req.NoError(err)
			}

			req.NotEmpty(asked, "the model was asked")
			// the record's data reached the model (a flow may use only some fields)
			all := strings.Join(asked, "\n")
			found := false
			for _, vals := range skRecords[f.module] {
				found = found || strings.Contains(all, vals[0])
			}
			req.True(found, "none of the record's values reached the model: %s", all)
			// the library prompt, not the reference, was sent
			req.NotContains(strings.Join(asked, "\n"), "@prompt:")
		})
	}
}

func TestStroykontrolFlows_Definitions(t *testing.T) {
	seenHandles := map[string]bool{}
	for _, f := range skFlows() {
		require.False(t, seenHandles[f.spec.Handle], "duplicate handle %s", f.spec.Handle)
		seenHandles[f.spec.Handle] = true
		require.NotEmpty(t, f.spec.Description)
		require.Contains(t, []string{"stroykontrol", "stroykontrol-id"}, f.namespace)
		require.NotNil(t, skRecords[f.module], "no sample record for module %s", f.module)
	}

	prompts := map[string]bool{}
	for _, p := range skPrompts {
		require.True(t, validHandle(p.handle), p.handle)
		prompts[p.handle] = true
	}
	// every library prompt a flow uses exists
	for _, f := range skFlows() {
		for _, s := range f.spec.Steps {
			for _, v := range s.Args {
				if str, ok := v.(string); ok && aiagent.IsPromptRef(str) {
					require.True(t, prompts[strings.TrimPrefix(str, "@prompt:")], "%s uses a prompt that is not defined: %s", f.spec.Handle, str)
				}
			}
		}
	}
}

func validHandle(h string) bool { return aiagent.PromptHandlePattern.MatchString(h) }

// ---------------------------------------------------------------------------
// installing into a database

func TestSeedStroykontrol(t *testing.T) {
	dsn := os.Getenv("SEED_DSN")
	if dsn == "" {
		t.Skip("SEED_DSN not set")
	}

	ctx := context.Background()
	req := require.New(t)

	st, err := connectForSeed(ctx, dsn)
	req.NoError(err)

	owner, err := store.LookupUserByHandle(ctx, st, os.Getenv("SEED_USER"))
	req.NoError(err, "SEED_USER must be the handle of an existing user: the flows are owned by it and run on its behalf")
	ctx = auth.SetIdentityToContext(ctx, auth.Authenticated(owner.ID))

	// prompts first: flows refer to them
	lib := PromptLibrary(st, allowEverything{})
	for _, p := range skPrompts {
		if _, err := lib.Get(ctx, p.handle, 0); err == nil {
			t.Logf("prompt %s: already there", p.handle)
			continue
		}
		cases := p.cases
		saved, err := lib.Save(ctx, SavePrompt{Handle: p.handle, Text: p.text, Description: p.description, Note: "стартовая версия", Cases: &cases, Activate: true})
		req.NoError(err, p.handle)
		t.Logf("prompt %s: saved v%d with %d test cases", p.handle, saved.Version, len(p.cases))
	}

	svc := templateWorkflowService(t, skModel(new([]string)))
	aiagent.SetPromptResolver(skLibrary()) // only so that the flows can be checked; nothing is sent
	defer aiagent.SetPromptResolver(nil)

	for _, f := range skFlows() {
		ns, err := store.LookupComposeNamespaceBySlug(ctx, st, f.namespace)
		req.NoError(err, "namespace %s", f.namespace)
		mod, err := store.LookupComposeModuleByNamespaceIDHandle(ctx, st, ns.ID, f.module)
		req.NoError(err, "module %s/%s", f.namespace, f.module)

		if existing, err := store.LookupAutomationWorkflowByHandle(ctx, st, f.spec.Handle); err == nil && existing != nil {
			t.Logf("workflow %s: already there", f.spec.Handle)
			continue
		}

		c := CompileAIFlow(svc, &f.spec, owner.ID)
		req.Empty(c.Issues, issuesText(c))

		wf := c.Workflow
		wf.ID = nextID()
		wf.OwnedBy, wf.CreatedBy, wf.CreatedAt = owner.ID, owner.ID, time.Now()
		wf.Enabled = f.manual // buttons are live; reactions to new records wait for a person
		wf.Meta.AIBudget = &types.AIBudget{MaxLLMCalls: 3, MaxTokens: 8000}
		wf.Labels = map[string]labelTypes.LabelValue{
			"ref_namespace": {Values: []string{fmt.Sprintf("corteza::compose:namespace/%d", ns.ID)}},
			"ref_module":    {Values: []string{fmt.Sprintf("corteza::compose:module/%d/%d", ns.ID, mod.ID)}},
		}
		req.NoError(store.CreateAutomationWorkflow(ctx, st, wf), f.spec.Handle)
		req.NoError(label.Create(ctx, st, wf), f.spec.Handle)

		event := "afterCreate"
		desc := "Реагирует на создание записи"
		if f.manual {
			event, desc = "onManual", "Кнопка на записях модуля"
		}
		trigger := &types.Trigger{
			ID: nextID(), WorkflowID: wf.ID, StepID: wf.Steps[0].ID,
			ResourceType: "compose:record", EventType: event, Enabled: wf.Enabled,
			Constraints: types.TriggerConstraintSet{
				{Name: "namespace", Op: "=", Values: []string{f.namespace}},
				{Name: "module", Op: "=", Values: []string{f.module}},
			},
			Meta:    &types.TriggerMeta{Description: desc},
			OwnedBy: owner.ID, CreatedBy: owner.ID, CreatedAt: time.Now(),
		}
		req.NoError(store.CreateAutomationTrigger(ctx, st, trigger), f.spec.Handle)

		t.Logf("workflow %s -> %s/%s (%s, enabled=%v)", f.spec.Handle, f.namespace, f.module, event, wf.Enabled)
	}

}

type allowEverything struct{}

func (allowEverything) CanSearchWorkflows(context.Context) bool { return true }
func (allowEverything) CanCreateWorkflow(context.Context) bool  { return true }

func connectForSeed(ctx context.Context, dsn string) (store.Storer, error) {
	return postgres.Connect(ctx, dsn)
}
