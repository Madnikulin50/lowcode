# Аудит безопасности — подготовка к сертификации ФСТЭК (КИИ)

Дата: 2026-01-15 (обновлено 2026-11-01) · Ветка: `2026.10.x` · Охват: `server/` (Go), `client3/lib/*` + `client3/web/*` (JS)

## 1. Выполненные исправления кода

| # | Файл | Уязвимость | CWE | Исправление |
|---|------|-----------|-----|-------------|
| 1 | `server/system/renderer/gotenbergPDF.go` | SSRF: `pl.Options["url"]` переопределял адрес рендера | CWE-918 | Убран user-controlled `url`; цель фиксирована на URL драйвера из конфигурации |
| 2 | `client3/lib/vue/src/components/prompts/kinds/CPrompt*.vue` (6 файлов) | Stored/DOM XSS: `v-html="message"` из payload | CWE-79 | `v-html` → `v-text` (экранирование) |
| 3 | `server/pkg/auth/signer.go` | Слабый примитив: HMAC-SHA1 | CWE-327 | HMAC-SHA1 → HMAC-SHA256 (длина подписи 40 → 64) |
| 4 | `server/system/service/auth_credentials.go` | Предсказуемый OTP: `math/rand` | CWE-330 | `crypto/rand` через `secureRandomOTP()` |
| 5 | `server/pkg/gonec/engine.go` | Обход песочницы: строковый блэклист, неиспользуемый import-allowlist | CWE-94 / CWE-749 | Введена проверка импортов (`checkImports`/`importAllowed`) на основе `Engine.imports` перед компиляцией |
| 6 | `server/go.mod` + `vendor` | Уязвимые/устаревшие Go-зависимости (реестр — раздел 5) | CWE-1104 | Обновлены `jwt` v3→v5 (через `go-oauth2` v4.6.0), `jwt/v4` 4.5.2, `chi/v5` 5.3.0, `grpc` 1.83.2, `amqp091-go` 1.14.0, `goxmldsig` 1.6.0, `fasthttp` 1.70.0, `x/*`, `gorilla/*`, `jsonparser`, `klauspost/compress`, `yaml.v3` (в `pkg/goqu`); удалён неиспользуемый `ollama` |
| 7 | `server/pkg/api/server/metrics.go` | Зависимость от заброшенного `766b/chi-prometheus` → `go-chi/chi` v3 | CWE-1104 | Заменён на собственный prometheus-middleware на `chi/v5` |
| 8 | `server/auth/oauth2/oauth2.go`, `server/auth/handlers/` | Несовместимость с `go-oauth2` v4.6.0 | CWE-1104 | Адаптированы сигнатуры: `GetErrorData` получил `ctx`, `BearerAuth` удалён из интерфейса; обновлены моки |
| 9 | `server/system/reporting` | Нерабочий (не компилируемый) устаревший пакет, ломал `go build ./...` и `govulncheck ./...` | — | Удалён (не импортировался, типы переехали в `datasources`) |
| 10 | `server/webconsole` (`package.json`, `yarn.lock`, `tsconfig.json`) | Устаревший lock (axios 0.28, vue 3.2.31, ~30 транзитивных пакетов с CVE) | CWE-1104 | Lock пересобран, `vue` ^3.5.42, `typescript` ~5.4, `vue-tsc` ^2, `resolutions` для транзитивных пакетов; `yarn build` проходит |
| 11 | `client3/lib/*`, `client3/web/*` (`yarn.lock`) | JS-зависимости: `axios` 1.8–1.19, `lodash`, `picomatch`, `minimatch`, `immutable`, `nanoid`, `postcss`, `handlebars`, `svgo` и др. | CWE-1104 | Lock-файлы пересобраны в рамках текущих semver-диапазонов (без смены major): OSV 507 → 182 вхождения. Побочные правки: `lib/js/src/api-clients/compose.ts` (`params … as KV` под типы axios 1.20) и явный `include` в `rollup.config.js` обоих библиотек (`picomatch` ≥2.3.2 перестал матчить дефолтный `**/*.ts+(|x)` у `rollup-plugin-typescript2`). Сборка `lib/js`, `lib/vue` и всех 7 webapp проходит |
| 12 | `server/go.mod` + `vendor` (повторный прогон `govulncheck` 2026-11-01) | GO-2026-6860 / GO-2026-6866 в `golang.org/x/net` 0.60.0 — достижимая уязвимость: `webapp.serveConfig` вызывает `html.Parse` (`pkg/webapp/serve.go:185`), квадратичная сложность при слиянии текстовых узлов | CWE-407 | `x/net` → 0.62.0; попутно `x/crypto` 0.58.0, `x/sync` 0.24.0, `x/sys` 0.49.0, `x/term` 0.47.0, `x/text` 0.43.0 |
| 13 | `server/go.mod` + `vendor` | Новая зависимость `github.com/xuri/excelize/v2` v2.11.0 (нужна экспорту в XLSX, `compose/envoy/xlsx_encode.go`) имела достижимые GO-2026-6688 и GO-2026-6684 и ещё 4 advisory на уровне пакета | CWE-1104 | Зафиксирована pseudo-версия `v2.11.1-0.20261029001137-c5fd73bde8f6` (master с исправлениями); после выхода релиза ≥ v2.11.1 заменить на него. Также добавлен `go-pdf/fpdf` v0.9.0 (экспорт в PDF) |
| 14 | `server/pkg/errors/http.go` | Текст ошибки (`err.Error()`, может содержать данные запроса) и разделители передавались в `fmt.Fprintf` как строка формата | CWE-134 | Все вызовы `write(...)` переведены на константный формат (`"%s"`); `go vet` по пакету чист |
| 15 | `server/auth/handlers/handle_logout_test.go`, `server/pkg/envoyx/csv/decoder_test.go` | Тесты не проходили/не компилировались: `Test_logoutProc` падал из-за пустого hash-ключа `gothic.Store` (`gorilla/securecookie` ≥ 1.1.2 его отвергает; в проде `gothic.Store` задаётся в `auth/external/external.go`), `testReader()` возвращал `io.Reader` вместо `io.ReadSeeker` | — | В тесте logout задан `sessions.NewCookieStore` с тестовым ключом; `testReader()` возвращает `io.ReadSeeker`. Код прода не менялся |

Подтверждение сборки: `go build -mod=vendor ./...` проходит (2026-11-01), все тестовые пакеты компилируются (`go test -run XXX_NONE ./...`), `govulncheck ./...` — на уровне символов **No vulnerabilities found**; ранее — `go build -mod=vendor` ключевых пакетов (`./pkg/... ./system/service/... ./system/rest/... ./system/types/... ./compose/... ./automation/... ./federation/... ./store/...`) — без ошибок. `get_diagnostics` — чисто.

Pre-existing падения тестов (проверено на чистом `HEAD` в отдельном worktree — воспроизводятся без моих правок, к безопасности не относятся): `pkg/dal` (`TestAggregator*`, `TestStepAggregate*`), `compose/service` (`TestRecord_boolFieldPermissionIssueKBR`, nil pointer), `compose/service/values` (`Test_sanitizer_Run`).

## 2. Результаты аудита (изменений кода не потребовалось)

- **SQL-инъекции**: `custom_*.go` (`custom_roles`, `custom_labels`, `custom_auth_sessions`, `automation_prompt_versions`, `compose_record_drafts`, `custom_compose_pages` и др.) строят запросы через goqu — параметризация, инъекции нет. Единственный `fmt.Sprintf` с данными — миграция `upgrade_fixes.go:234` (JSON-конфиг модуля в `%s`); низкий риск (выполняется при апгрейде), рекомендуется перевести на плейсхолдеры.
- **SSRF-поверхность**: `pkg/chat/client.go` (Ollama URL из настроек/env), `pkg/aiagent/remote.go`, `federation/service` — целевые URL административно-конфигурируемые, не из пользовательского ввода. Единственная user-controlled точка была в `gotenbergPDF` — исправлена.
- **`gonec` (исполнение Go-кода)**: MCP-инструмент `gonec_run` гейтится `MCP_ENABLE_SCRIPTS=true` (по умолчанию выключен); rulesgo-нода `gonec` регистрируется без runner (no-op). Остаточный риск: при включении скриптов доступен широкий lowcode-SDK; в production `MCP_ENABLE_SCRIPTS` должен оставаться выключенным либо требуется OS-уровневая изоляция (gVisor/firejail/контейнер).

## 3. Чеклист мер защиты (приказ ФСТЭК № 239, значимый объект КИИ)

- **ИАФ (идентификация/аутентификация)** — реализовано: bcrypt (DefaultCost), MFA TOTP (dgoogauth) и Email OTP, rate-limiting токенов (`checkTokenRate`), политика паролей, запрет повторного пароля, сброс access-токенов.
- **УПД (управление доступом)** — реализовано: RBAC (`custom_rbac.go`, проверки `svc.ac.*`).
- **АУД (аудит безопасности)** — реализовано: централизованный пакет `actionlog` + паттерн `recordAction` покрывает аутентификацию (`auth_actions.yaml`: входы, смена/сброс пароля, impersonate, TOTP/Email-OTP, ошибки `invalidCredentials`/`rateLimitExceeded`), пользователей (`user.go`: create/update/delete/suspend/unsuspend, отзыв токенов/сессий), роли (`role.go`: create/update/delete), вложения и data-privacy, federation и automation-сессии. AI-шаги фиксируются в стектрейсе сессии (`trace`).
- **ЦЛО / целостность, ОЦЛ, ЗНИ** — требуют эксплуатационной оценки и не покрываются статическим аудитом.

## 4. Замечания и рекомендации

1. **Сканеры зависимостей** (запуск 2026-01-15):
   - **Go** — `osv-scanner` отработал (база `api.osv.dev` доступна): 180 пакетов из `server/go.mod`, реестр в разделе 5. `govulncheck` в январе не отработал (база `vuln.go.dev` недоступна); повторный прогон 2026-10-11 (локальная копия базы, `govulncheck` v1.9.0, Go 1.27.1, `./...` и `-scan package`) — **No vulnerabilities found**. Прогон 2026-11-01 (база от 2026-10-27, Go 1.27.3, `GOFLAGS=-mod=vendor govulncheck ./...`) нашёл достижимую GO-2026-6860 (`x/net`) и уязвимости новой зависимости `excelize` — исправлено (раздел 1, п. 12–13); итог на уровне символов — **No vulnerabilities found**.
   - **JS** — `yarn audit` выполнен по всем пакетам `client3/lib/*` и `client3/web/*` (prod-зависимости); `osv-scanner` дополнительно выявил устаревший `server/webconsole/yarn.lock` (vite 2.9.13, rollup 2.67.2, axios 0.28.0) — раздел 5.
   - **`trivy`** (Docker-образы) — не запускался: инструмент не установлен, образы локально не собраны.
   - Команды: `osv-scanner scan -r server`, `govulncheck -mode=source ./...`, `trivy image <образ>`.
2. **Совместимость**: переход на HMAC-SHA256 инвалидирует ранее выпущенные подписанные URL вложений (`?sign=...`) — они короткоживущие, генерируются заново.
3. **Pre-existing `go vet`**: `system/service/auth_test.go` (unkeyed struct) и `setting_actions.gen.go` (`recordAction passes lock by value`) — не вносились в рамках аудита.
4. **НДВ-контроль**: оформить процедурно (репозиторий, сборочный конвейер, третьи стороны) — автоматически не проверяется.

## 5. Реестр CVE/БДУ (osv-scanner + yarn audit)

### Go (`osv-scanner`, `server/go.mod`, 180 пакетов)

Сводка OSV после обновления зависимостей: 36 пакетов, 110 вхождений (5 Critical / 51 High / 46 Medium / 5 Low / 3 Unknown), 2 экосистемы (Go + npm `webconsole`); 104 исправимо обновлением. (До обновления: 52 пакета, 167 вхождений — 9 Critical / 68 High / 33 Unknown.)

| Пакет | Версия | Исправление | CVSS | Примечание |
|---|---|---|---|---|
| `github.com/rabbitmq/amqp091-go` | 1.10.0 | 1.13.0 | до 9.5 | Critical, множественные advisory |
| `google.golang.org/grpc` | 1.61.1 | 1.83.1 | до 9.1 | Critical, множественные |
| `github.com/golang-jwt/jwt` | 3.2.2 | — | 8.7 | устаревший (v3); заменить на v5 |
| `github.com/golang-jwt/jwt/v4` | 4.5.1 | 4.5.2 | 8.7 | обход проверки подписи |
| `github.com/russellhaering/goxmldsig` | 1.4.0 | 1.6.0 | 7.5 | XML-signature (SAML) |
| `github.com/buger/jsonparser` | 1.1.1 | 1.1.2 | 7.5 | DoS |
| `golang.org/x/oauth2` | 0.16.0 | 0.27.0 | 7.5 | — |
| `golang.org/x/image` | 0.22.0 | 0.45.0 | 7.5 | множественные |
| `gopkg.in/yaml.v3` | 3.0.0-2020… | 3.0.0-2022… | 7.5 | DoS |
| `github.com/ollama/ollama` | 0.30.10 | — | — | устаревший, 10+ advisory без патча |
| `github.com/go-chi/chi` + `/v5` | 3.3.4 / 5.0.7 | 5.3.0 | 5.1–8.7 | — |
| `github.com/gorilla/websocket` | 1.5.0 | 1.5.3 | 6.9 | — |
| `github.com/gorilla/csrf` | 1.7.1 | 1.7.3 | 5.4 | — |
| `golang.org/x/crypto` | 0.53.0 | 0.56.0 | — | множественные |
| `golang.org/x/net` | 0.56.0 | 0.60.0 | — | множественные |
| `github.com/valyala/fasthttp` | 1.35.0 | 1.70.0 | — | устаревший |
| `gopkg.in/square/go-jose.v2` | 2.3.1 | — | — | устаревший (перейти на v4) |

Повторный прогон 2026-10-11: 10 пакетов / 42 вхождения (3 Critical / 9 High / 24 Medium / 4 Low / 2 Unknown), из них Go — только `imaging` и `chi` v3 (без исправления), остальное — dev-инструменты `webconsole` (`vite` 2, `vitest`, `esbuild`, `braces`). `go-jose.v2`, `gorilla/csrf`, `x/crypto` и `fasthttp` из отчёта ушли.

Статус 2026-01-15: все пакеты таблицы, кроме трёх остаточных, обновлены до исправленных версий (раздел 1, п. 6–7). Остаточные: `go-chi/chi` v3 (только тестовая зависимость `go-chi/jwtauth`), `valyala/fasthttp` (тестовая зависимость `go-oauth2`), `gopkg.in/square/go-jose.v2` (из устаревшего `crusttech/go-oidc`, требует замены OIDC-клиента).

### JS (`yarn audit`, npm advisories)

Dev-зависимости сборки исключены (`--groups dependencies`).

| Область | Пакетов | Уязвимостей (prod) | Ключевые |
|---|---|---|---|
| `client3/lib/vue` | 427 | 55 (2 Low / 21 Mod / 32 High) | `prosemirror-view` XSS при вставке (High, ≥1.42.3, через `@tiptap/starter-kit`); `xlsx` prototype pollution + ReDoS (High, без патча) |
| `client3/lib/js` | 44 | 50 (2 Low / 25 Mod / 22 High / 1 Crit) | `form-data`→`axios`: PRNG boundary (Crit, ≥4.0.4); `lodash` prototype pollution (Mod, ≥4.17.23); `follow-redirects` утечка auth-заголовков (Mod, ≥1.16.0) |
| `client3/web/admin` | 90 | 21 (8 Mod / 13 High) | tiptap/prosemirror, xlsx, axios |
| `client3/web/compose` | 362 | 56 (2 Low / 20 Mod / 34 High) | tiptap/prosemirror, xlsx, axios |
| `client3/web/discovery` | 92 | 56 (1 Low / 28 Mod / 26 High / 1 Crit) | `form-data`→`axios` (Crit, ≥4.0.4); tiptap/prosemirror; xlsx |
| `client3/web/one` | 52 | 7 (1 Mod / 6 High) | xlsx, axios |
| `client3/web/privacy` | 89 | 16 (6 Mod / 10 High) | tiptap/prosemirror, xlsx |
| `client3/web/reporter` | 107 | 26 (8 Mod / 18 High) | tiptap/prosemirror, xlsx, axios |
| `client3/web/workflow` | 84 | 10 (3 Mod / 7 High) | xlsx, axios |
| **Итого (вхождения по пакетам)** | — | **297** (7 Low / 120 Mod / 168 High / 2 Crit) | — |

Повторяющиеся в большинстве пакетов первопричины: `form-data`→`axios` (Critical, CWE-330), `prosemirror-view`/tiptap XSS (High, CWE-79), `xlsx` prototype pollution (High, CWE-1321, без патча), `lodash` prototype pollution (Mod, CWE-1321), `follow-redirects` утечка заголовков (Mod, CWE-200), `echarts` XSS (Mod).

Полный скан `lib/vue` с devDependencies: 204 уязвимости (2 Critical) — цепочки сборки (`rollup-plugin-styles`, `sass`, `cssnano`), в рантайм не попадают.

Рекомендации: обновить `@tiptap/*`/`prosemirror-*` до ≥1.42.3; заменить/обновить `xlsx` (SheetJS — официальных патчей нет); обновить `axios`/`form-data` (≥4.0.4) и `lodash` (≥4.17.23).

## 6. Осталось (вне рамок текущей итерации)

Статус 2026-11-01 (повторный прогон):

- **`govulncheck`** (`govulncheck` v1.9.0, Go 1.27.3, база `vuln.go.dev` от 2026-10-27, `GOFLAGS=-mod=vendor`, `./...`) — на уровне символов **No vulnerabilities found** (достижимых уязвимостей нет). Не вызываются кодом, исправления у upstream нет — оставлены под наблюдение: `GO-2025-3884` (`gorilla/csrf` 1.7.3), `GO-2024-2631` (`gopkg.in/square/go-jose.v2` 2.3.1, тянется из `crusttech/go-oidc`; решение — заменить OIDC-клиент на `go-oidc/v3` + `go-jose/v4`), `GO-2026-5932` (`x/crypto` 0.58.0, уровень модуля). Перепроверять после каждого обновления зависимостей и при выходе новых версий.
- **`excelize`** — зафиксирована pseudo-версия master (`v2.11.1-0.20261029…`), так как релиз с исправлениями GO-2026-6684/6685/6687/6688/6689/6695 на 2026-11-01 не опубликован; после публикации релиза перейти на него (`go get github.com/xuri/excelize/v2@vX.Y.Z && go mod tidy && go mod vendor`).
- **`system/reporting`** — удалён (пакет не импортировался, типы переехали в `datasources`); `go build ./...` проходит. Закрыто.
- **Go (`osv-scanner`)** — остались 2 записи без исправления: `disintegration/imaging` 1.6.2 (GHSA-q7pp-wcgr-pffx / CVE-2023-36308, LOW — panic на crafted TIFF; используется в `system/service/attachment.go` и `compose/service/attachment.go`, рекомендуется ограничить типы изображений или заменить библиотеку) и `go-chi/chi` v3 (GO-2026-4316, только тестовая зависимость `go-chi/jwtauth`, в runtime не линкуется). `go-jose.v2`, `gorilla/csrf`, `x/crypto` в отчёте больше не фигурируют.
- **`server/webconsole`** — `yarn.lock` обновлён (был рассинхронизирован с `package.json`: axios 0.28 → 1.x), `vue` ^3.5.42, `typescript` ~5.4, `vue-tsc` ^2, `resolutions` для транзитивных пакетов; `yarn build` проходит. Остаток — только dev-инструменты, фиксящиеся лишь сменой major: `vite` 2.9.x, `vitest` 0.2.x, `esbuild` 0.14, `tinypool`, `braces`, `postcss-selector-parser` (в `dist` не попадают).
- JS `client3/*`: безопасные обновления в рамках semver выполнены (п. 11, OSV: 96 пакетов / 182 вхождения). Остаток требует смены major или замены библиотек: `xlsx` (SheetJS, патча нет — рассмотреть `exceljs`), `echarts` 5→6, `vitest` 3→4 / `tinypool` / `uuid` 10→11 / `moment`, `mxgraph` (патча нет), `braces` (патча нет), `@vue/server-renderer`.
- Не сделано: запуск `trivy` по собранным Docker-образам.
- Не сделано: уточнить модель угроз/категорию значимости КИИ для финального чеклиста.
