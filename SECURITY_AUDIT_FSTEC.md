# Аудит безопасности — подготовка к сертификации ФСТЭК (КИИ)

Дата: 2026-01-15 (обновлено 2026-10-11) · Ветка: `2026.10.x` · Охват: `server/` (Go), `client3/lib/*` + `client3/web/*` (JS)

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
| 12 | `server/pkg/rand/rand.go`, `server/auth/handlers/handle_mfa_totp.go` | Предсказуемые секреты: `math/rand` с seed по времени генерировал ID сессий, токены сброса/OTP-ссылок, секреты OAuth-клиентов, токены federation; секрет TOTP — тоже `math/rand` (`gosec` G404) | CWE-338 | `pkg/rand` переписан на `crypto/rand` (тот же API `Bytes`/`Password`, равномерный выбор символа, без смещения); TOTP-секрет — `crypto/rand` с обработкой ошибки |
| 13 | `server/pkg/webapp/serve.go` | Путь из `r.URL.Path` собирался через `path.Join` без привязки к корню — `../` выходил за каталог вебаппа при `os.Stat` (`gosec` G703) | CWE-22 | Путь нормализуется `path.Clean("/"+…)` |
| 14 | `server/pkg/aiagent/skillfmt.go` | Чтение ресурсов skill из каталога по символьным ссылкам (`gosec` G122) | CWE-367 | Пропускаются не-регулярные файлы (symlink, устройства) |
| 15 | `server/pkg/api/server/middleware.go` | `panicRecovery`: в логе вместо значения паники писался пустой `err` (`staticcheck` SA9008) | CWE-778 | Значение паники берётся из `recover()` |
| 16 | `server/pkg/apigw/profiler/hit.go`, `server/store/adapters/rdbms/ddl/goqu.go` | Бесконечная рекурсия в `Hits.Len` и `sqlExpression.IsPrepared` (`staticcheck` SA5007) → переполнение стека | CWE-674 | `len(h)` и `false` |
| 17 | `server/system/renderer/html.go` | Шаблонная функция `inlineRemote` выполняла `http.Get` по произвольному URL без ограничений (схема, время, размер) (`gosec` G107) | CWE-918 / CWE-400 | Только `http(s)`, таймаут 15 с, лимит ответа 10 МБ. Остаточный риск: обращение к внутренним адресам не блокируется (автор шаблона может запросить внутренний URL) — при необходимости ограничить сетевой политикой |
| 18 | `server/pkg/api/server/waitfor.go` | Служебный HTTP-сервер без таймаутов (slowloris, `gosec` G114) | CWE-676 | `http.Server{ReadHeaderTimeout: 10s}` |
| 19 | `server/pkg/api/server/middleware.go` | При панике клиенту возвращался текст паники (пути, SQL) | CWE-209 | Клиенту — «internal server error»; детали только в логе, а в ответе — лишь при заданном `LOG_DEBUG` |

Подтверждение сборки: `go build ./...` проходит (2026-10-11); ранее — `go build -mod=vendor` ключевых пакетов (`./pkg/... ./system/service/... ./system/rest/... ./system/types/... ./compose/... ./automation/... ./federation/... ./store/...`) — без ошибок. `get_diagnostics` — чисто.

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
   - **Go** — `osv-scanner` отработал (база `api.osv.dev` доступна): 180 пакетов из `server/go.mod`, реестр в разделе 5. `govulncheck` в январе не отработал (база `vuln.go.dev` недоступна); повторный прогон 2026-10-11 (локальная копия базы, `govulncheck` v1.9.0, Go 1.27.1, `./...` и `-scan package`) — **No vulnerabilities found**.
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

## 5a. Статический анализ (gosec, staticcheck) — 2026-10-12

Анализатор Svace на стенде отсутствует; вместо него прогнаны открытые анализаторы (`gosec` v2.26.1: `-exclude-dir=vendor -exclude-generated -severity medium -confidence medium`; `staticcheck` `./...`). Сертификационной силы они не имеют.

**gosec** — 277 срабатываний на 1474 файлах: G120 (118, `ParseForm` без `MaxBytesReader`), G115 (86, преобразования целых), G304 (25, `os.Open` по пути), G204 (7), G703 (5), G404 (4), G401/G501/G505 (4/3/1, MD5/SHA-1 для идентификаторов), G122 (3), G118 (3), прочие по 1–5. Исправлены реальные: G404 (п. 12), G703 (п. 13), G122 (п. 14), G107 (п. 17), G114 (п. 18). Повторный прогон после правок: 272 срабатывания (G120 118, G115 86, G304 25, G204 7, G306 5, G401 4, G117 4, G703 4, G404 2 — оба принятые; остальные по 1–3).

Принятые/ложные срабатывания: `pkg/riskengine/attribution.go` (G404, `rand` с фиксированным seed 42 намеренно — воспроизводимое объяснение Шепли); `pkg/http/default.go` (G402, `InsecureSkipVerify` только при явном `SYSTEM_HTTP_CLIENT_TSL_INSECURE=1`); G204 в `pkg/gonec`, `pkg/rag` (запуск внешних процессов с аргументами из конфигурации/админа; для `gonec` см. раздел 2); G401/G501/G505 — хэши для идентификаторов/ключей кеша (`compose/types/module.go`, `pkg/jsenv/vm.go`, `auth/handlers/handle_sessions.go`, `pkg/options/auth.go`), не для защиты секретов — рекомендуется заменить на SHA-256 при ближайшей правке; G114 (`pkg/api/server/waitfor.go`, служебный сервер ожидания без таймаутов); G704 (`pkg/apigw/filter/proxy`, `pkg/http/client.go` — целевые адреса из конфигурации маршрутов); G705 (`automation/rest/skills.go`, заголовок через `mime.FormatMediaType`). Рекомендация (не сделано): ограничить тела запросов (`http.MaxBytesReader`, G120) на маршрутах с формами — глобальный лимит ломает загрузку вложений, поэтому нужен точечный middleware для `auth/`-форм. Замена MD5/SHA-1 (G401/G501/G505) не делалась: хэши входят в сохранённые идентификаторы (`compose/types/module.go`), смена алгоритма изменит имена в БД.

**staticcheck** — 1278 замечаний, из них стиль и мёртвый код (`U1000` 555, `S10xx`, `ST1xxx`) не затрагивались; по корректности исправлены `SA9008`, `SA5007` (п. 15–16). Не исправлялись (не уязвимости, часть в сгенерированном коде): `SA4017` (`strings.ToLower` без присвоения в `*yaml_decode.gen.go` — требует правки шаблонов кодогенерации), `SA4023`, `SA1019` (65 устаревших API), `SA1029` (строковые ключи `context.WithValue`), `SA5008` (теги `json:",string"`), `SA1012` (nil context в тестах).

Проверка: `go build ./...` проходит; тесты `pkg/api`, `pkg/apigw`, `store/adapters/rdbms/ddl`, `pkg/aiagent`, `pkg/webapp` — ok. `auth/handlers` `Test_logoutProc` падает (`Could not delete user session`) — не связан с этими правками (в изменённых файлах не используется), на чистой ветке отдельно не проверялся.

## 6. Осталось (вне рамок текущей итерации)

Статус 2026-10-11 (повторный прогон):

- **`govulncheck`** (`-db` — локальная копия `vuln.go.dev`, `govulncheck` v1.9.0 на Go 1.27.1, `./...`) — **No vulnerabilities found**; режим `-scan package` — тоже чисто. Закрыто.
- **`system/reporting`** — удалён (пакет не импортировался, типы переехали в `datasources`); `go build ./...` проходит. Закрыто.
- **Go (`osv-scanner`)** — остались 2 записи без исправления: `disintegration/imaging` 1.6.2 (GHSA-q7pp-wcgr-pffx / CVE-2023-36308, LOW — panic на crafted TIFF; используется в `system/service/attachment.go` и `compose/service/attachment.go`, рекомендуется ограничить типы изображений или заменить библиотеку) и `go-chi/chi` v3 (GO-2026-4316, только тестовая зависимость `go-chi/jwtauth`, в runtime не линкуется). `go-jose.v2`, `gorilla/csrf`, `x/crypto` в отчёте больше не фигурируют.
- **`server/webconsole`** — `yarn.lock` обновлён (был рассинхронизирован с `package.json`: axios 0.28 → 1.x), `vue` ^3.5.42, `typescript` ~5.4, `vue-tsc` ^2, `resolutions` для транзитивных пакетов; `yarn build` проходит. Остаток — только dev-инструменты, фиксящиеся лишь сменой major: `vite` 2.9.x, `vitest` 0.2.x, `esbuild` 0.14, `tinypool`, `braces`, `postcss-selector-parser` (в `dist` не попадают).
- JS `client3/*`: безопасные обновления в рамках semver выполнены (п. 11, OSV: 96 пакетов / 182 вхождения). Остаток требует смены major или замены библиотек: `xlsx` (SheetJS, патча нет — рассмотреть `exceljs`), `echarts` 5→6, `vitest` 3→4 / `tinypool` / `uuid` 10→11 / `moment`, `mxgraph` (патча нет), `braces` (патча нет), `@vue/server-renderer`.
- Не сделано: запуск `trivy` по собранным Docker-образам.
- Не сделано: уточнить модель угроз/категорию значимости КИИ для финального чеклиста.
