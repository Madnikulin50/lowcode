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
