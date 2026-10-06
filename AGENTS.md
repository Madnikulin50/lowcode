# Chat & MCP conventions

## Tech stack
- **Ollama** runs on `http://localhost:11434`
- **Default model**: `deepseek-v2` (set in `server/compose/service/chat.go`)
- **Model override**: wrap `<model>modelname</model>` in prompt
- **LLM client**: `github.com/cloudwego/eino-ext/components/model/ollama` in `server/pkg/chat/client.go`

## Chat tools (`server/compose/service/`)

### Static tools (`chat_tools.go` — `AllChatToolDefs()`)
18 tools: read/list/search/create/update/delete for modules, pages, charts.
Parameters are flat strings. Return JSON or user‑friendly text.
List functions return entity set; on empty set return `"No X found"` text.

### Dynamic per‑module tools (`chat.go` — `getTools()`)
Generated at runtime by querying `DefaultModule.Find`:
- `show_module_{handle}`, `module_search_{handle}`
- `module_{handle}_records`, `module_{handle}_create_record`, `module_{handle}_update_record`, `module_{handle}_delete_record`

`module_create_record` with empty `values` prints available fields.

### Confirmation
`needsConfirm()` triggers on `create_` / `delete_` prefixed tools.
Flow: LLM suggests action → stream asks for "да" → re‑entry with tool call.

## MCP handlers (`server/compose/mcp/handlers/`)

### Static tools (`modules.go`, `pages.go`, `charts.go`)
Same 18 CRUD tools as chat, with `namespaceID` as explicit param.

### Dynamic per‑module tools (`records.go` — `initModuleRecords()`)
Generated at startup, mirrors chat dynamic tools:
- `module_{handle}_records`, `module_{handle}_search`, `module_{handle}_create_record`, `module_{handle}_update_record`, `module_{handle}_delete_record`

### Transport
- STDIO: optional (`MCP_STDIO=true`)
- SSE: optional (`MCP_SSE_ADDR=:9090`)
- See `server/compose/mcp/index.go`

### Shared helpers (`ctrl.go`)
- `withAuth(ctx)` — injects auth identity
- `getString/parseUint64/argsMap` — extract typed args from `CallToolRequest`
- `jsonResult(textResult/errorResult)` — wrap response

## Adding a new entity
1. Create `chat_{entity}.go` handlers in `service/` with `chat{Action}{Entity}` funcs
2. Add tool defs in `service/chat_tools.go` → `AllChatToolDefs()` includes them
3. Create `handlers/{entity}.go` in `mcp/handlers/` with `init{Entity}()` + handler funcs
4. Add `init{Entity}(ctx, s)` call in `handlers/index.go`

## Vue 2 → Vue 3 Conversion Summary

### Completed (all files in `PageBlocks/` and subdirs)
All Vue 2 components converted to Vue 3 `<script setup>` + Composition API + Bootstrap 5 (no Bootstrap-Vue). Key conversions:

- **PageBlocks/base.vue** → uses `usePageBlockBase` composable
- **PageBlocks/Wrap/** → Card.vue, Plain.vue, index.js all converted
- **PageBlocks/index.js** → functional component, no `Vue.component`
- **PageBlocks/Navigation/** → Base, Configurator, NavTypes all converted
- **PageBlocks/Comment/Base.vue** → `<script setup>`, `$root` events → `window` events
- **PageBlocks/Shared/** → AutomationButtons, AutomationTab, AutomationTabButtonEditor
- **PageBlocks/RecordListBase.vue** (~3,268 lines) → largest, fully converted
- **PageBlocks/RecordListConfigurator.vue** (~1,387 lines) → fully converted
- **PageBlocks/RecordList/** → Prefilter, CustomFilterPreset, CustomSummary
- **PageBlocks/ProgressBase.vue, ProgressConfigurator.vue**
- **PageBlocks/RecordBase, RecordEditor, RecordConfigurator**
- **PageBlocks/RecordRevisionsBase, RecordRevisionsConfigurator**
- **PageBlocks/RecordOrganizerBase, RecordOrganizerConfigurator**
- **PageBlocks/Report/Base, Configurator**
- **PageBlocks/SocialFeedBase, SocialFeedConfigurator**
- **PageBlocks/TabsBase, TabsConfigurator**
- **PageBlocks/ChartBase, ContentBase, FileBase, GeometryBase, IFrameBase, MetricBase**
- **Public/** → AiChat, Attachment, Block/Modal, Grid, Scenarios
- **Public/Record/** → BulkEdit, Exporter, Importer, Modal
- **Drafts/** → CDraftButton, CDraftSidebar, DraftItem, Drafts
- **Translator/** → CTranslatorButton, CTranslatorForm, CTranslatorModal, Editable

### Conversion patterns applied
- `b-*` components → Bootstrap 5 native HTML/CSS classes
- `extends: base` → `usePageBlockBase()` composable
- `mixins` → imported utility functions
- `mapGetters/mapActions` → `useStore()` with store getters/actions
- `$root.$on/$emit/$off` → `window.addEventListener/dispatchEvent/removeEventListener`
- `$t()` → `useI18n()`
- `$auth`, `$ComposeAPI`, `$router` → `inject()`
- `data` → `ref()` / `reactive()`
- `beforeDestroy` → `onBeforeUnmount`
- `ml-*/mr-*/pl-*/pr-*` → `ms-*/me-*/ps-*/pe-*`
- `font-weight-bold` → `fw-bold`

## Objective
- Завершить сборку всех webapp (Vite) после Vue 2→3 миграции — исправить все ошибки разрешения зависимостей и импортов

## Important Details
- compose `vite.config.js` использует алиасы: `corteza-lib/vue/dist` → `../../../lib/vue/dist` (workspace root), `corteza-webapp-compose` → `.`, `corteza-webapp-compose/src/stores` → `src/store`
- rollup сборка lib/vue external'ит белый список; commonjs плагин с `defaultIsModuleExports: false`
- `tsconfig.json` в lib/vue: `"declaration": false`
- `src/stores/` → `src/store/` (без 's') — старые Vuex модули стали Pinia stores
- `@tiptap/*` версии должны совпадать между lib/vue и webapp (symlink или установка)
- `brace` и `ace-builds` нужны для vue3-ace-editor (CAceEditor)
- Bootstrap CSS импортирован в `main.js` — `.form-switch` стили теперь в бандле (bootstrap-vue-next.css их не содержит)
- `<b-form-checkbox switch>` не работает (switch prop как атрибут без значения); все заменены на нативный `div.form-check.form-switch > input[type=checkbox]`
- `BFormCheckbox` удалён из `main.js` — больше не регистрируется
- `form-check-input-v3` → `form-check-input` в 4 компонентах lib/vue (иначе `.form-switch .form-check-input` селектор Bootstrap не срабатывает)
- `#toolbar` Teleport warning — безвредная гонка при монтировании; фикс: `<div id="toolbar">` добавлен в `index.html` (всегда существует)

## Work State
### Completed
- **compose i18n.js**: создан
- **compose vite.config**: алиас `stores`→`store`; PostCSS `.cjs`
- **autoprefixer/rtlcss**: установлены
- **portal-vue@3.0.0**: установлен
- **base.vue → useViewerBase**: вынесен компосабл; 11 импортов обновлены
- **lib/vue stores**: 3 Vuex → Pinia
- **lib/vue rollup.config**: external белый список + `defaultIsModuleExports: false`
- **lib/vue index.ts**: re-exports `CortezaAPI`, `useSettings`, `mixins`, Pinia stores
- **CAceEditor.vue**: `import { VAceEditor }` fix; brace dynamic imports → template literal
- **useSettings composable**: создан
- **lib/vue node_modules**: symlinks для axios, i18next-pseudo, @vue-leaflet/vue-leaflet, vue-color, @popperjs/core, ace-builds, @tiptap/*
- **compose node_modules**: установлены brace, ace-builds, vue3-ace-editor, @tiptap/*, axios, i18next-pseudo, @vue-leaflet/vue-leaflet
- **✅ compose сборка**: успешно (422 модуля, 7.03s)
- **Switch fix**: Bootstrap CSS импортирован; все `<b-form-checkbox switch>` → нативный `.form-switch`; `BFormCheckbox` удалён; `form-check-input-v3` → `form-check-input`
- **CProgress `size="sm"`**: новый prop; Number.vue передаёт `size="sm"` вместо инлайн-стилей; label без `/100%` суффикса
- **Card.vue null-safety**: optional chaining на `block.style?.border?.enabled`
- **Index.vue #toolbar z-order**: перемещён перед `<router-view>`
- **Teleport warning устранён**: `<div id="toolbar">` добавлен в `index.html`
- **IFrameBase.vue**: добавлен `inject` в импорт из `vue`
- **Card.vue все computed**: `props.block` → `props.block?.` optional chaining везде (было `Cannot read properties of undefined (reading 'style')`)
- **MetricBase.vue guard filter**: добавлена проверка `!props.record` перед `evaluatePrefilter` для `${record`/`${ownerID}` ссылок
- **ProgressBase.vue guard filter**: то же самое
- **✅ Module Create fix**: copy-paste bug in `module_field.go` — `encodeTranslationsMetaPrefix` и `encodeTranslationsMetaSuffix` использовали `LocaleKeyModuleFieldMetaHintView.Path` вместо `MetaPrefix`/`MetaSuffix`, создавая дубликаты ключей `meta.hint.view` → `unique_violation` на `resource_translations_uniqueTranslation` constraint; первая ошибка глоталась `errorHandler`(`return nil`), транзакция абортилась, последующие запросы падали с `current transaction is aborted`
- **PostgreSQL errorHandler fix**: `unique_violation` теперь возвращает `store.ErrNotUnique.Wrap(implErr)` вместо `nil`, чтобы не оставлять транзакцию в абортированном состоянии незаметно
- **Builder.vue modals**: переписаны 3 модала с CSS show/hide на Bootstrap Modal API (`import { Modal }`) — watch на refs, hidden.bs.modal listeners, dispose/unmount
- **Edit.vue autocomplete fix**: `import autocomplete` удалён (Vue 2 миксин → мигрирован на inline-функции с `getCurrentInstance`)
- **FontAwesome icon picker**: добавлен в Edit.vue для страниц; sidebar (`NamespaceSidebar.vue`) отображает fontawesome иконки; `faIcons.js` — добавлены solid иконки (faEnvelope, faClock и др.)
- **✅ Все webapp собраны**:
  - **compose** — ✅ (422 модуля, 7.03s)
  - **admin** — ✅ (1341 модуль, 10.89s)
  - **discovery** — ✅ (693 модуля, 12.70s)
  - **one** — ✅ (606 модулей, 11.94s)
  - **privacy** — ✅ (617 модулей, 11.75s)
  - **reporter** — ✅ (1175 модулей, 15.73s)
  - **workflow** — ✅ (657 модулей, 17.57s)

### Active
- (none)

### Blocked
- (none)

## Next Move
- Протестировать собранные webapp в браузере — проверить функциональность модалов в Builder.vue, иконок страниц, и отсутствие vue-i18n `label.search` ошибок в консоли

## AI steps in automation workflows

Functions `aiAsk`, `aiExtract`, `aiClassify`, `aiTool`, `aiRagSearch`, `aiRunCalls` (`server/automation/automation/ai_handler.{yaml,go,gen.go}`; `.gen.go` is produced from the yaml by `make codegen-legacy`). They run on `aiagent.DefaultRegistry()`; no agent named = tool-less `workflow-llm`. Shared prompt/JSON/retry contract: `server/pkg/aiagent/structured.go` (also used by rulesgo `ai.operation`).

- Every LLM step has a deadline (`timeoutSec`, default 5 min). Mutating actions need `allowMutating`.
- **Approval flow**: `aiAsk` with `deferConfirm` (no `allowMutating`) returns `needsApproval`, `pendingCalls`, `summary` instead of failing → gateway → prompt step (owner = approver) → `aiRunCalls` executes the approved calls (only tools the agent owns).
- Rule chain node `workflow` runs a workflow (ID or handle) via `service.ExecWorkflowByRef`; the reverse bridge is the `compose.runRuleChain` function.

## Suspended sessions survive restarts

Sessions waiting on a delay/prompt are persisted to `automation_states` (`wfexec.Session.Snapshots`/`Restore`, `automation/service/session_persist.go`) and restored in `service.Activate`. Only the awaited states survive (not in-flight parallel branches); states inside loops cannot be restored; a session whose workflow is gone/changed is marked failed. Vars from JSON need `ResolveTypes(Registry().Type)`.
`store/adapters/rdbms/rdbms.gen.go` is hand-maintained: a new codegen'd resource also needs its `Store` methods (see `automation_states.go`).

## Triggering and running AI workflows (Phase 3)

- **Events** (`server/pkg/wfevent`, hand-written, not codegen): `anomaly:finding` (`onCreate`, `onReopen`), `risk:assessment` (`onAssessed`, `onEscalated`). Plain-data events; constraints are `<prop>.<field>` (e.g. `finding.severity`). Emitted from `anomaly/service/scanner.go` (`upsertFinding`) and `compose/service/risk_events.go` (`SaveRiskAssessment`, the single path for auto-recalc and the manual Assess endpoint). They are listed in the editor via `automation/rest/eventTypes.go` (`customEventTypeDefinitions`, outside `eventTypes.gen.go`) and require run-as on the workflow (`validateWorkflowTriggers`). Event variables are readable as `finding.severity`; use `toPlainJSON(finding)` (not `toJSON`, which keeps the `@type` envelope) to put them in a prompt.
- **Run from outside**: MCP `workflow_list|run|session|templates|install_template`; chat/agent toolkit `workflows` (`compose/service/chat_workflows.go`, in the `assistant` agent; `run_workflow` and `install_workflow_template` are Mutating → need "да"). Shared by both: `ExecWorkflowByRef`, `ListWorkflowBriefs`, `WorkflowSessionBrief`, `InstallAITemplate` in `automation/service`.
- **Templates** (`automation/service/ai_templates.go`): `anomaly-explain`, `risk-escalation-review`. Installed disabled, run as the installer; the final log steps are placeholders for the real action. Add a template = builder func + entry in `aiTemplates` + an end-to-end case in `ai_templates_test.go` (it converts the graph and runs it with a scripted LLM).
- **Remote agents** (cmdb, backup, invest) are steps via `aiTool` (`cmdb_scan`, `backup_run`, `invest_evm`, ...: same tools the LLM sees, from the catalog). Async ones return a job ID; poll `<kit>_job_status` in a loop with a delay step.
- **Webhook** trigger: the existing system sink (`SinkOnRequest`) already starts a workflow from an HTTP call.

## Editors and observability for AI steps (Phase 4)

- **AI call trace**: `aiagent.CallTrace` (`pkg/aiagent/trace.go`) — agent, model, attempts, LLM calls, prompt/completion tokens (`AgentResult` now carries `Model/LLMCalls/PromptTokens/CompletionTokens`, accumulated in `runtime.go`), duration, tools, rejected answers, prompt/response truncated to `TracePreviewLimit`. Workflow `aiAsk|aiExtract|aiClassify` return it as the `trace` result (so it lands in the session stacktrace frame results when tracing is on); rule chain `ai`/`ai.operation` report it through a context sink (`rulesgo/trace_sink.go`) into `NodeResult.Trace`, which — unlike `Output` — survives a failed node. `NodeResult.DurationMs` is recorded for every node.
- **Per-call model**: `aiagent.ContextWithModel(ctx, model)`; `Agent.clientForRun(ctx)` honours it. (A `<model>` tag in the prompt is only understood by compose chat, not by the agent runtime — don't use it for steps/nodes.) `ai.operation` now runs on `aiagent.RunStructured` like the workflow steps; the bridge is `compose/mcp/ai_call.go`.
- **Editors**: rule chain node catalog gets live agent/model choices and the `timeout`/`optional` fields (`compose/rest/rulechain_ai.go`, `chat.ModelChoices`); workflow function list gets agent/model pick-lists (`automation/rest/function.go`, copy-on-write) and param label/description/textarea hints from `ai_handler.yaml` (`ai_handler.gen.go` is generated from it by `make codegen-legacy`).
- **Test buttons** (real model, nothing saved, data changes forced off): rule chain `POST /admin/rulechain/node-test` (ai nodes only, `Engine.ExecuteNode`), workflow `POST /ai/functions/{ref}/test` (aiAsk/aiExtract/aiClassify/aiRagSearch only; needs permission to create workflows). Hand-written API client methods `ruleChainNodeTest` / `aiStepTest` live in `client3/lib/js/src/api-clients` outside the generated parts — re-add them if the client is regenerated.
- **Viewers**: compose rule chain editor (`RuleChainAiTrace.vue`, run history, test modal, node "Try this node"); admin session editor (`CSessionAiSteps.vue`, reads the stacktrace — only present when the workflow has tracing on).

## Prompt library, evaluation and AI-assisted authoring (Phase 5)

- **Prompt library**: named, versioned prompts in `automation_prompt_versions` (`automation/types/prompt.go`, service `automation/service/prompt.go`, store `store/adapters/rdbms/automation_prompt_versions.go`). Versions are immutable; one per handle is active; rollback = `Activate`. Steps and nodes use a prompt as `@prompt:<handle>` (active) or `@prompt:<handle>@<n>` (pinned) in their prompt argument — resolved by `aiagent.ResolvePrompt` in the workflow `ai*` functions and the rule chain `ai`/`ai.operation` nodes (library text first, then `{{variables}}`); an unknown reference fails the step, it is never sent to the model as text. The trace records `promptRef` ("handle@version"). Read = may search workflows, write = may create workflows. Surfaces: MCP `prompt_*`, chat/agent toolkit `prompts` (`save_prompt`/`activate_prompt` are Mutating).
- **Evaluation**: `aiagent.RunEval` runs golden cases (`inputs` + `expect` fields and/or `contains` phrases; strings compared ignoring case) and reports pass rate and tokens; `prompt_eval` / `evaluate_prompt` compare versions on the same cases. Cases belong to a version and carry forward unless replaced. Retries are off by default in an eval so it measures the prompt, not the retry loop.
- **AI authoring**: the workflow assistant writes a *flow description* (named steps, `=`-prefixed expressions, `flow` links) instead of raw step JSON; `CompileAIFlow` (`automation/service/ai_authoring.go`) checks functions, arguments, results and connections, then validates with the real converter, and reports issues by step name. Tools: `list_ai_functions`, `validate_ai_workflow`, `create_ai_workflow`, `list_ai_templates`, `create_workflow_from_template` (the `create_*` ones ask for "да"). Result is always created disabled and runs as the user.
- **Code generation — two generators**: `make codegen` (Go, `server/codegen`) does types/store/model for the registered resources; `make codegen-legacy` generates `*_handler.yaml` → `*_handler.gen.go`, `types.yaml` → `type_set.gen.go`, REST, events. The legacy one also rewrites unrelated generated files (and may create case-colliding duplicates such as `DalSchemaAlteration.go`): after running it, check `git status` and revert what you did not mean to change. `store/adapters/rdbms/rdbms.gen.go` is hand-maintained, so a new resource also needs its `Store` methods (copy the `automationSession` section, see `automation_states.go`).
- **Not done**: structured tool parameters are supported by `chat.ParamDef.Type` but most older tools still declare strings; a compose-chat assistant for page blocks beyond what the toolkits offer.

## Budget, chain authoring, typed tool parameters, streaming (Phase 5 follow-up)

- **AI budget** (`pkg/aiagent/budget.go`): caps LLM calls and tokens per *run* (a workflow session, a rule chain run); all AI steps of the run share one `Budget` on the context. Check happens before a call, charge after (a run can overshoot by its last call); a retry loop draws on the same budget. Workflow: `meta.aiBudget {maxTokens, maxLLMCalls}` (set in the workflow configurator), built per session in `session.Watch` (`workflow.aiBudgetFor`) and on restore; chain: `config.aiBudget` (chain editor). Platform default: env `AI_MAX_TOKENS_PER_RUN`, `AI_MAX_LLM_CALLS_PER_RUN`. Models that report no usage can only be held to the call count. A restored session starts with a fresh budget. `BudgetedRunner` wraps any `Runner`.
- **Rule chain authoring**: `rulesgo.ValidateChain` checks a chain against the node catalog (types, required/enum/number/bool settings, edges, reachability, `@prompt:` refs) — errors block saving, warnings do not. The catalog is injected (`handlers.SetRuleChainCatalog`, `service.SetRuleChainCatalog`, both set from `compose/rest`'s `MountRuleChainAdminRoutes`) because it lives in `compose/rest`. MCP: `rulechain_node_types`, `validate_rule_chain`, and `create_rule_chain` now validates first (with no catalog it saves as before and says it did not check). Assistant toolkit `rulechains`. NB: an edge `condition` is the *name* of a variable, followed when non-empty — not an expression.
- **Typed tool parameters**: `ParamDef.Type` may be `string` (default), `json` (text containing JSON, as before), `number`, `integer`, `boolean`, `array`, `object`, `objects`; handlers still receive text (`chat.ParamString`: composites as JSON, whole numbers without exponent). Keep IDs as strings — large numbers lose precision.
- **Streaming progress**: `aiagent.ContextWithProgress` puts a `ProgressFunc` on the context; `Agent.RunConfirmed` then generates as a stream and reports status/tokens/reasoning, `RunStructured` reports each attempt. `api.NewSSE` writes the events (same `data: {...}\n\n` framing as chat). Endpoints with `?stream=1`: workflow step test, rule chain node test, prompt eval (per case). Client: `postForEvents` in `lib/js/src/api-clients/sse.ts` (fetch-based; tests: `npx mocha --import=tsx src/api-clients/sse.test.ts`).
- **Prompt library UI**: admin app, Automation → Prompts (`views/Automation/Prompts`, `components/Prompts/PromptEval.vue`); REST `/prompts/...` in `automation/rest/prompts.go`. Evaluation goes through `service.SetPromptEvalRunner` (default: agent registry).

## Skills

A skill is a named, versioned set of instructions that an agent loads when a task calls for it (`server/pkg/aiagent/skills.go`, `skill_run.go`, `skillfmt.go`). Stored in the **prompt library** (`automation_prompt_versions`, `kind = "skill"`, plus `requires` and `resources` JSON columns), so versions, activation/rollback and permissions are the prompt library's (read = may search workflows, write = may create workflows). `promptLibrary.Skills()` is the view that sees only skills; a handle is a prompt or a skill, never both, and `@prompt:x` never resolves to a skill. Service: `automation/service/skill.go` (limits: 50 files, 100 KB a file, 512 KB in all; text only, no script execution). New columns on the existing table come from `fix_2026_10_00_extendAutomationPromptVersionsForSkills` (`upgrade_fixes.go`).

- **Format**: `SKILL.md` = YAML header (`name`, `description` = *when to use it*, optional `requires`) + markdown; with files a zip (SKILL.md at the root or one directory down). `ImportSkill` / `ExportSkill` (service), REST `POST /skills/import`, `GET /skills/{handle}/export`, MCP `skill_import|export`.
- **Sources** (`aiagent.SetSkillSource`, first wins): `db` (the library), `file` (`AI_SKILLS_DIR`: `<name>.md` or `<name>/SKILL.md` + files; read on every call), `remote` (agents publish `skills` in `/api/meta`, bodies from `GET /api/skills/{handle}`; `agents/sdk` `Service.Skills(...)`, `server/pkg/aiagent/remote_skills.go`). Files without a valid header are reported and skipped.
- **Universal agent**: `AgentSpec.Skills` (`["*"]` = all; the `assistant` has it). The system prompt carries only the catalog; the agent gets `list_skills`, `load_skill`, `read_skill_resource`. `load_skill` returns the instructions, the file list, and activates the skill's `requires` toolkits: `run()` adds them to the tool set from the next turn (`SkillRun` on `Options.Skills`/ctx, `withExtraTools`). Skills do not widen rights: tools they bring are confirmed like any other (`Mutating` still needs the yes). `ExecApproved` accepts the toolkits of skills the agent may load; compose chat (`AttachSkills`) stores loaded skills in `pendingToolCalls` and restores them on "да".
- **Chosen skill** (no catalog): workflow `aiAsk|aiExtract|aiClassify` param `skill`, rule chain `ai` / `ai.operation` field `skill` (`handle` or `handle@version`) → `aiagent.ContextWithSkill`; its text goes into the prompt and its tools are available from turn one. `ai_handler.gen.go` was edited by hand for the new param (mirrors the yaml).
- **Trace**: `AgentResult.Skills` / `CallTrace.Skills` ("handle@version").
- **Authoring**: MCP `skill_*`, assistant toolkit `skills` (`chat_skills.go`: `list_skill_library`, `get_skill`, `save_skill`, `import_skill`, `activate_skill`; the saving ones are Mutating — NB `list_skills` is the agent's catalog tool, a different thing). Admin UI: Automation → Skills (`views/Automation/Skills`, `components/Skills/CSkillEditorInfo.vue`), agent skills in Settings → AI agents; client methods `skill*` in the hand-written block of `api-clients/automation.ts`.
- **Not done**: eval of skills (prompt eval cases do not apply to a skill yet), binary resources, running scripts from a skill.
- **Stroykontrol skills**: `agents/stroykontrol/skills/<handle>/SKILL.md` (designer-analyst, estimator, reviser, process-engineer, procurement, read-drawings, estimator-kp with `price-lists.md`), loadable with `AI_SKILLS_DIR=agents/stroykontrol/skills` (test: `TestStroykontrolSkillsLoad`). The original documents are kept in `skills/_source/`; edit the SKILL.md files, not those.
