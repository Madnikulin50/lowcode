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
- **AiChat Modal.vue**: `fullscreen = ref(false)` по умолчанию; `.modal-dialog:not(.modal-fullscreen){max-width:70vw}`; `modal-xl`/`.modal-max-width` удалены
- **ETLSettings.vue создан**: списо-к/создание/редактирование/run/delete ETL-джобов модуля; API использует `etlID` (не `etlJobID`)
- **lib/vue node_modules → symlink'и (абсолютные)** на `web/compose/node_modules` (vue, @tiptap/*, vue-select, prosemirror-стек и др.) — фикс краша «Adding different instances of a keyed plugin» (дубли prosemirror-state); `autoprefixer` symlink нужен для rollup-сборки dist
- **vue-select@4.0.0-beta.6** в web/compose (`--legacy-peer-deps`, конфликт с vue-tweet-embed) — v3.20.4 был Vue 2 build (`this.$options.propsData` → undefined)
- **CRichTextInput**: проп `maxHeight` → `maxBodyHeight`; dist lib/vue пересобран
- **JWT-debug лог** удалён из `store/namespace.js`
- **faIcons.js**: добавлены faPlay, faTrash, faCertificate, faTimeline
- **ETL локализация**: полные блоки `etl:` в `locale/{en,ru}/corteza-webapp-compose/module.yaml`; синхронизировано в `server/pkg/locale/src/{en,ru}/` (module/block/chart/datasources); сервер пересобран (встраивание через `//go:embed src/*`)
- **vuedraggable 4.1.0 fix**: CItemPicker.vue и CInputPresort.vue переведены на `#item`-слоты (без него render() кидает «draggable element must have an item slot» → componentStructure undefined → краш в `updated()`); guard `if (this.error || !this.componentStructure) return` добавлен в `updated()` (web/compose/node_modules/vuedraggable); dist lib/vue пересобран
- **Pages RAG: исключение record-bound блоков**: `blockTiedToRecord()` в `server/compose/service/pages_rag.go` — блоки с `${recordID}` в options (prefilter/filter, рекурсивно по metrics[]) пропускаются краулом; сервер пересобран
- **Pages RAG: переводы**: `translatedBlockText()` — в текст чанка добавляется секция `Translations:` со страницей/блоком (title/description/content body) на всех языках, кроме дефолтного (ключи `types.LocaleKeyPage*`); `PagesRAGService` получил `locale ResourceTranslationsManagerService` (DefaultResourceTranslation); работает при `LOCALE_RESOURCE_TRANSLATIONS_ENABLED`
- **Admin Rules & Workflow (аналог Charts)**: раздел `admin.rulechains` в админке — список/создание/редактирование/удаление/тест цепочек правил (rulesgo):
  - lib/js `compose.ts`: `ruleChainList/Read/Create/Update/Delete/Test/NodeTypes/Stats` (REST `/api/compose/admin/rulechain/`); dist пересобран
  - сервер: `RuleChainAdmin.Update` теперь применяет `nodes`/`edges` (раньше игнорировались); `rulechain.yaml` (en/ru) встроен в сервер через `//go:embed`; `navigation.rulechains` ключ
  - webapp: роуты `admin.rulechains(.create/.edit)` в `views/routes.js`; пункт «Rule chains» в `NamespaceSidebar.vue` (`adminRoutes()`, `namespaceSelected`, фильтр поиска); `i18n.js` + ns `rulechain`
  - `views/Admin/RuleChains/List.vue` — c-resource-list (клиентские поиск/сортировка/пагинация по `route.query.page|limit`); `Edit.vue` — форма (name/description/entryNode), редактор узлов (label/type/config-JSON/description+configSchema из `/admin/rulechain/nodes`), редактор связей (from/to/label/condition), Toolbar (save/saveAndClose/delete), тест через Bootstrap Modal (`POST /admin/rulechain/{chainID}/test`, в create-режиме сначала сохраняет)
- **Chart legend toggle fix (Vue 3 регрессия)**: `CInputCheckbox` мигрирован на `modelValue`/`update:modelValue`, но 3 переключателя остались на Vue 2 паттерне `:value`+`@input` → `@input` падал в `$attrs` на нативный `<input>`, писал `Event`-объект в модель (всегда truthy) → «Отображать легенду» у пончиковой диаграммы не работал, легенда пропадала навсегда. Исправлено на `v-model`: `ReportEdit.vue` (`legendShown` computed: `!isHidden` — свитч ON = легенда видна), `GenericChart.vue` (`tooltip.labelsNextToPartition`), `views/Admin/Charts/Edit.vue` (`toolbox.saveAsImage`); важно: с `invert` первый клик глотается (модель меняется только со второго) — проверено эмпирически в headless Chrome; compose пересобран (21.98s)

### Active
- (none)

### Blocked
- **Порты 18081/18082 — НЕ наш dev-стек**: это Docker-контейнеры `corteza-docs-server-1` (docs) и `crm-corteza-1` (corteza 2023.9.1, образ `cortezaproject/corteza`) — не трогать/не перезапускать; реальный сервер — GoLand debug `Server_RU_test_translations` на `:3333` (vite dev `:8080` проксирует туда)
- **GoLand debugger паузит сервер**: при срабатывании брейкпоинта весь процесс уходит в `tracing stop` (22 треда в `t`), все HTTP-запросы висят до Resume в GoLand; если API «висит» — сначала проверить `/proc/<pid>/status` → `State: t (tracing stop)`

### Verification done (2026-08-10)
- **Dev-сервер :3333 актуален** (бинарь собран 18:53, содержит все патчи chat/rulechain/RAG/локали): `GET /compose/chat/models` 200, `POST /compose/chat/warmup` 200
- **JWT для curl-тестов**: секрет = `md5("jwt secret"+DB_DSN+HOSTNAME)` (HOSTNAME не задан → "localhost"); токен минтуется скриптом `/tmp/opencode/mint-token.mjs` (jti = `auth_oa2tokens.access`, sub = `rel_user`, scope `profile api`, HS512); exp на 24h
- **RAG reindex**: `POST /compose/pages-rag/reindex` → 12/12 страниц, complete:true; чанки содержат секцию `Translations:` (RAG-переводы работают)
- **Rule chains**: list/create/update(с nodes+edges — фикс работает)/test/delete — все 200 (тестовая цепочка создана и удалена)
- **Chat stream e2e**: `POST /namespace/{id}/prompt/stream` — reasoning+token чанки, done:true; first byte ~30s на CPU (cold prompt-eval + большой tool-промпт), генерация ~10 tok/s
- **Cold prompt-eval (deepseek-r1, CPU)**: 1606 токенов = 46.5s = **35 tok/s**; cache-hit = 92ms = **17317 tok/s** — кэш KV работает; узкое место — cold eval и генерация, только GPU решит
- **Block/Modal.vue fix**: `Property "module" was accessed during render` — в шаблоне был `:module="module"`, но ref не объявлен (Vue 3 <script setup> не имеет this); добавлен `module = ref(undefined)` + установка в `loadModal`/сброс в `setDefaultValues` (по образцу Public/Record/Modal.vue)
- **Unhandled CanceledError fix (Vue 3)**: при unmount блоков с cancelToken-запросами промис отклонялся без catch → «Uncaught (in promise) CanceledError: abort-request-*»; добавлен `.catch(e => { if (axios.isCancel(e)) return; console.error(...) })`: GeometryBase.vue (RecordFeed), CalendarBase.vue (RecordFeed/ReminderFeed), AutomationBase.vue (automationListCancellable), Configurator.vue (roleListCancellable); RecordListBase.vue уже обрабатывал
- **9 задач по продвинутым диаграммам (todo-сессия, 2026-08-10)**: 8 из 9 уже были реализованы (проверено по коду): (1) сервер — 2 измерения в `recordReportToAggPipelineStep` (`;`-разделитель, `dimension_0/1`, Group+OrderBy); (2) `util.ts` — `ChartType` + 10 новых типов; (3) `base.ts` — `formatReporterParams` `.slice(0,2)`+`;`, `rows` в `processReporterResults`; (4) `chart.ts` — `makeAdvancedOptions` ветки для 10 типов; (5) `components/index.js` — регистрация ECharts-модулей (включая MapChart/Sunburst/Parallel/Calendar); (6) `lib/chart-maps.js` — `ensureMapRegistered` (fetch `public/maps/{world,china}.json` → `registerMap`), вызов в `Chart/index.vue`; (7) переводы — не хватало `edit.metric.calendarType.*` (label+3 options) → добавлены в `server/pkg/locale/src/{en,ru}/corteza-webapp-compose/chart.yaml` (все ключи GenericChart.vue теперь покрыты, сверено скриптом); (8) `GenericChart.vue` — `mapType` селект для `metric.type === 'map'`; (9) сборки: lib/js dist (2s), compose (23.29s), сервер `/tmp/opencode/corteza-server` через `/snap/bin/go build` — OK. **Примечание**: `/snap/bin/go` снова работает (snap-блокер в AGENTS.md устарел), standalone Go из `/tmp/opencode/go` удалён
- **Line chart: опция «Отображать точки»**: чекбокс `edit.metric.showSymbol` в GenericChart.vue (line-only, `v-model` на `metric.showSymbol`, дефолт `true` — нормализация в `watch(report)` и `chartTypeChanged`); в `chart.ts` — `showSymbol: m.showSymbol` в makeDataset + `showSymbol: false` в серии line (ECharts, точки скрыты но видны при ховере); переводы `edit.metric.showSymbol` en/ru в `server/pkg/locale/src/{en,ru}/corteza-webapp-compose/chart.yaml`; сборки lib/js (2s) + compose (26.16s) + сервер — OK
- **Градиент на заливке Line chart**: из `gradientItemStyle` вынесен `linearGradientColor()` (chart.ts) — линейный top→bottom градиент (darkToLight: цвет→светлее, lightToDark: светлее→цвет); при `gradient && type === 'line' && fill` цвет применяется к `areaStyle` заливки под линией; сборки lib/js (2s) + compose (22.37s) — OK
- **Line chart: Stacked (накопленный)**: чекбокс `edit.metric.stacked` в GenericChart.vue (line-only, v-model `metric.stacked`) + **текстовое поле `metric.stack` возвращено для line** (оно — источник истины; чекбокс — ярлык: `onStackedToggle` ставит `stack='total'` при включении, чистит при выключении; нормализация `stacked===true && !stack → 'total'` в watch); chart.ts makeDataset: `stack: m.stack` — все серии с одинаковым ключом стека накапливаются (как в примере ECharts area-stack-gradient, с градиентной заливкой из прошлой правки); проверено node-тестом: M1/M2 `stack=total` + gradient на areaStyle, M3 без стека; сборки lib/js (2.3s) + compose (21.60s) — OK
- **Серверная группировка по полю (stackBy)**: селект «Группировать по полю (стек)» в GenericChart.vue для line-метрик (`metric.stackBy`), @change включает «Накопленный» (`onStackByChange` → `stacked=true, stack='total'`); `formatReporterParams` (base.ts) вставляет `stackBy` как `dimension_1` в dimensions (через `;`; только при наличии основного измерения); сервер уже умеет группировать по 2 измерениям (`recordReportToAggPipelineStep` — dimension_0/1); `processReporterResults` режет строки на серии: уникальные `dimension_1` → отдельные линии с общим `stack='total'`; переводы `edit.metric.stackBy` en/ru; `makeDataset` в base.ts получил `: any` (был `never` из-за throw); сборки lib/js (2.2s) + compose (21.42s — 22.48s с guard) — OK. **Важно**: нужно хотя бы 1 основное измерение (ось X); стек-поле передаётся в `dimensions` (не отдельным query-параметром) — в Network вкладке браузера `/record/report?dimensions=DATE_FORMAT(...)%3Bfield_name`
- **Цвет баллонов Select-полей в RecordList (2026-08-11)**: цвета опций хранятся в `compose_module_field.options.options[].style.{textColor,backgroundColor}` именами темы (`danger`, `primary`, `light`), hex получается через `getColor()` из темы `ui.studio.themes`. Баг: `useViewerBase.js`, `Viewer/base.vue`, `Configurator/Select.vue`, `Configurator/File.vue`, `Editor/File.vue` обращались к `$Settings` как `$settings.value.get(...)` (реактивный объект из `plugins/settings.js` НЕ имеет `.value` — паттерн остался от Vue 2 миграции) → темы всегда пусты → `getColor('danger')` возвращал `'danger'` как есть → невалидный CSS → бейдж без цвета; плюс `Editor/Select.vue` не использовал `getColor` вовсе (рендерится при включённом inline-редактировании RecordList) — добавлен `getColor` в `Editor/base.js` (inject('$Settings')) и применён; `Editor/File.vue` также `$ComposeAPI.value` → `$ComposeAPI`; сборка compose (22.22s) — OK
- **Vue 3 фиксы в конфигураторах блоков (2026-08-11)**: (1) `ColumnPicker.vue:116` — `watch(fn, {handler})` (устаревшая Vue 3.4 сигнатура) → `watch(source, cb, {immediate})`; (2) баг `suggestionParams: Expected Array, got Object` + `(params || []).forEach is not a function` в CInputExpression: 6 конфигураторов обращались к `window.__composeAPI.processRecordAutoCompleteParams` (метод существовал только в Vue 2 миксине `mixins/autocomplete.js`, в lib/js `$ComposeAPI` его нет) с fallback `{}` (объект!) → заменены на локальные `processRecordAutoCompleteParams`/`processVisibilityAutoCompleteParams` с `inject('$auth')`: `RecordList/Prefilter.vue`, `RecordConfigurator.vue`, `ProgressConfigurator.vue`, `RecordOrganizerConfigurator.vue` (вызывал метод миксина без `this` — краш), `Shared/AutomationTabButtonEditor.vue` (inject с fallback `{}`, провайдера нет), `IFrameConfigurator.vue` + `MetricConfigurator/index.vue` + `CalendarConfigurator/FeedSource/configs/Record.vue` + `GeometryConfigurator/FeedSource/configs/Record.vue` (ref не был определён вовсе — подсказки не работали); `Comment/Configurator.vue` и `Configurator.vue` и `Pages/Edit.vue` уже имели локальные реализации; (3) `CInputExpression.vue` — guard `Array.isArray(props.suggestionParams) ? ... : []`; сборки lib/vue dist (13s) + compose (22.17s) — OK
- **Fix «Failed to resolve component: CFormTableWrapper/CInputCheckbox/CHint» (2026-08-12)**: Vue 3 `resolveComponent` НЕ делает kebab-fallback — компоненты, зарегистрированные глобально в kebab (`c-form-table-wrapper`, `c-input-checkbox`, `c-hint`, `c-input-select`, `c-input-confirm`, `c-webcam` в `web/compose/src/components/index.js` + `c-corredor-manual-buttons`), при использовании PascalCase (`<CFormTableWrapper>`) НЕ резолвятся → warn [Vue warn]: Failed to resolve component (в Vue 2 работало — регрессия миграции). Заменены PascalCase → kebab в 9 файлах: `ModuleFields/Configurator/{Record,Select,validation}.vue`, `ModuleFields/Editor/{Bool,File,Record,Select,User}.vue` (compose), `lib/vue/src/components/notifications/NotificationItem.vue` (CInputConfirm). Воспроизведено в headless Chrome (CDP + intercept `refresh_token` → подмена на свежий из БД через `psql`; маршрут `/ns/loop/admin/modules/495727984904372225/edit`; до фикса 6 warn'ов, после — 0, модал конфигуратора рендерится). Сборки: lib/vue dist (2.5s) + compose (1m8s) — OK; **важно**: правило — в шаблонах использовать ТОЛЬКО kebab-case для глобальных компонентов
- **Fix устаревшей сигнатуры `watch(fn, {handler})` (2026-08-12)**: Vue 3.4+ — `watch(fn, options?)` перенесён в `watchEffect`; все `watch(src, {handler () {...}})` → `watch(src, cb [, opts])`. Исправлены: `ModuleFields/Configurator/basic.vue:223,240` (дубликат проблемы из ColumnPicker), `Admin/Module/Datasources/{Join,Link}.vue` (3 watch), `Admin/Module/FederationSettings.vue` (4), `Namespaces/Reminders/Edit.vue`, `ModuleFields/Editor/{User,Geometry}.vue` (1+3); сборка compose (21.62s) — OK

- **Fix «Сохранить и закрыть» в модале Настройки поля (2026-08-12)**: `handleFieldSave` в `views/Admin/Modules/Edit.vue` (1) НЕ закрывал модал — `updateField` оставался set, модал `v-if="updateField"` не исчезал; (2) поиск поля по `f.name === field.name` ломал сохранение при переименовании поля в модале (findIndex -1 → поле молча терялось). Исправлено: поиск по `fieldID` (приоритет) с фолбэком на имя, при ненахождении — push, в конце `updateField.value = null`; подтверждено CDP-скриптом (SAVE CLICK → modalVisible:false); сборка compose (2m30s) — OK
- **«Не переведён» ключ — причина: dev-сервер читает НЕ embed, а внешний `../locale` (2026-08-12)**: в dev-режиме `LocaleOpt.Defaults()` (server/pkg/options/locale.go) ставит `Path = "../locale"` → сервер (`ReloadStatic` в `pkg/locale/service.go`) грузит переводы с диска `/home/maxim/work/lowcode/locale/{en,ru}/` (устарели 24.07), перекрывая embed `server/pkg/locale/src` (go:embed, актуальный). Фикс: rsync `server/pkg/locale/src/{en,ru}/` → `/home/maxim/work/lowcode/locale/{en,ru}/` (--delete, структура идентична); перезапуск НЕ нужен — `LOCALE_DEVELOPMENT_MODE=true` → `resolveAcceptLanguageHeaders` перечитывает переводы на каждый запрос. Заодно: ru/chart.yaml не имел `stackBy` (en имел) — добавлен `stackBy: label/placeholder` в `server/pkg/locale/src/ru/.../chart.yaml` + скопирован в locale/. **Правило: правки yaml-переводов = синхронизировать и src, и /lowcode/locale (и если бинарь используется вне dev — пересобрать сервер для embed)**
- **Кружок CInputColorPicker не показывал цвет имени темы (2026-08-12)**: в БД `compose_module_field.options.options[].style.{textColor,backgroundColor}` ХРАНИТ ИМЕНА ТЕМ (`light`, `danger`…) — CInputColorPicker рендерит кружок через `color: ${modelValue}` → для имени невалидный CSS → чёрный кружок. В `ModuleFields/Configurator/Select.vue` добавлен резолв `resolveColor()` (имя → hex из `ui.studio.themes`, как в Editor/base.js `getColor`); v-model заменён на `:model-value`+`@update:model-value` (при изменении пишется hex — имена тем сохраняются, пока цвет не трогают). Подтверждено CDP: кружки criticality — `light`→rgb(248,249,252), `danger`→rgb(231,74,59), `primary`→rgb(78,115,223); сборка compose (23.32s) — OK
- **Системный цвет — название рядом с кружком (2026-08-12)**: CInputColorPicker дополнен `systemColorKey` + `displayValue` computed — при клике на swatch темы рядом с кружком отображается имя переменной (`primary`, `danger`…), а не hex. `displayValue` также reverse-резолвит hex → имя темы при переоткрытии (чтобы при «Сохранить и закрыть» имя не сбрасывалось на hex). При `saveColor` / внешнем изменении модели `systemColorKey` сбрасывается; сборки lib/vue dist (3.2s) + compose (22.97s) OK
- **badgeGradient не сохранялся (2026-08-12)**: свойство `badgeGradient` не было объявлено в `SelectOptions`-интерфейсе и не копировалось в `applyOptions` (lib/js `compose/types/module-field/select.ts`) → при `ModuleFieldMaker` (клон поля для редактирования) терялось. Добавлен `badgeGradient: boolean` в интерфейс, в `defaults()` и в `Apply(..., Boolean)`. Проверено CDP: переключение → сохранить → переоткрыть → `badgeGradientChecked:true`; сборки lib/js (3.3s) + lib/vue (2.3s) + compose (23.10s) OK
- **Chrome Color Picker не отображался в модале CInputColorPicker (2026-08-12)**: vue-color@3.3.3 (Vue 3) — SFC-компоненты БЕЗ инжекта стилей: UMD-cjs рендерит DOM без CSS → `.saturation`/`.hue-wrap` имели height:0 → в модале были видны только инпуты rgba/hex (пикер «голый»). Стили лежат отдельно в `vue-color/dist/vue-color.css` — никто его не подключал. Фикс: в `lib/vue/src/components/input/CInputColorPicker.vue` добавлен блок `<style lang="css">@import 'vue-color/dist/vue-color.css';</style>` (в `<script>` `import` НЕ работал: rollup считал модуль external'ным; SFC-style + rollup-plugin-styles инлайнит в styleInject). Проверено CDP: пикер 544×428, saturation 299px, hue 10px, стили на странице; сборки lib/vue dist (3.5s) + compose (22.8s) — OK

## Next Move
- Server needs rebuild for stackBy locale keys (`edit.metric.stackBy.*` — бинарь собран до их добавления); GoLand debug restart on :3333 (бинарь заменён 23:02, перезапуск уже сделан — PID 800822 жив, `:3333` отвечает)
