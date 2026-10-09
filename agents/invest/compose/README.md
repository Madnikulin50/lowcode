# Инвестпроекты — Compose namespace

Пространство сопровождения инвестиционных проектов по ТЗ v0.6: единый источник правды (документы, WBS, договоры, бюджет/EVM, RFC, риски) плюс ИИ-советчики Юрист и Финконтролёр.

Паттерн тот же, что у CMDB: `apply.mjs` идемпотентно создаёт namespace / модули / страницы / чарты / rule chains.

## Apply

Сервер должен быть на `:3333`. JWT минтуется из живого refresh-токена БД, либо `TOKEN=…`.

```bash
cd agents/invest/compose
TOKEN=$(node /tmp/opencode/mint-token.mjs | head -1) \
  node apply.mjs
```

| Variable | Default |
|---|---|
| `COMPOSE_API` | auto-detect `http://localhost:3333/api/compose` |
| `TOKEN` | обмен refresh из Postgres |
| `INVEST_ENGINE_URL` | `http://localhost:8086/api` |
| `CALC_EVM_URL` | `http://localhost:8088/api` |
| `COMPOSE_DSN` | `postgres://postgres:Zse45rdx@127.0.0.1:5432/test9?sslmode=disable` |

Скрипт идемпотентен (обновляет по handle). Пишет `applied.json` с ID.

Демо-данные (11 проектов на всех фазах и статусах) досеиваются без дублей по коду:

```bash
cd agents/invest/compose
node seed.mjs
```

Открыть: `/ns/invest` (страница **Дашборд**).

## Модель

| Модуль (`handle`) | Роль |
|---|---|
| **Проекты** (`projects`) | Карточка проекта, фаза 1–6, SPI/CPI/EAC |
| **Участники** (`project_members`) | User + роль в проекте |
| **WBS** (`wbs_items`) | Иерархия стадия/подстадия/работа, предшественник, EVM-поля |
| **Документы** (`documents`) | Реестр + статус согласования + summary файла (`summary`, `extracted_text`, `extract_status`) |
| **Версии** (`document_versions`) | Обход отсутствия file-versioning |
| **Согласования** (`approvals`) | Маршрут: кто / решение / срок |
| **Комментарии** (`document_comments`) | Блок Comment на карточке документа |
| **Договоры** (`contracts`) | Реестр договоров, пакет документов |
| **Риски** (`risks`) | Вероятность / влияние / балл (score) / митигация |
| **RFC** (`change_requests`) | Запросы на изменение + симуляция EAC |
| **Журнал** (`change_log`) | Аудиторский след RFC |
| **Статьи бюджета** (`budget_lines`) | План / факт / резерв; ETL-заглушка 1С (выключена) |
| **Денежный поток** (`cashflow_items`) | Приход / расход |
| **Факты прогресса** (`progress_facts`) | Веб-фиксация объёма + фото + гео |
| **Типы конструкций** (`construction_types`) | НСИ для шаблонов WBS |
| **Шаблоны WBS** (`wbs_templates`) | Типовая иерархия работ |
| **Документы фазы** (`phase_requirements`) | Обязательные типы документов на фазу |
| **Журнал** (`change_log`) | Аудиторский след RFC |
| **Статьи бюджета** (`budget_lines`) | План / факт / резерв; ETL-заглушка 1С (выключена) |
| **Денежный поток** (`cashflow_items`) | Приход / расход |
| **Факты прогресса** (`progress_facts`) | Веб-фиксация объёма + фото + гео |
| **НСИ** | `document_types`, `counterparties`, `materials`, `labor_norms`, `construction_types`, `wbs_templates`, `phase_requirements` |
| **ИИ-советчики** (`ai_advisors`) | Юрист, Финконтролёр (промпты) |

Системные роли (apply): `invest-investor`, `invest-bank`, `invest-contractor`, `invest-designer`, `invest-government`, `invest-pmo`.

Record revisions включены на `projects`, `wbs_items`, `documents`, `contracts`, `change_requests`, `budget_lines`.

## Страницы

- **Дашборд** — метрики, кнопки EVM/CPM/алерты, портфель SPI/CPI, doughnut, Gantt, «Документы на мне» (`assignee = текущий пользователь`)
- **Документы** — канбан + реестр; карточка: файл, summary ИИ, автоверсия, маршрут по шагам, комментарии, «Согласовать мой шаг»
- **Проекты / WBS / Договоры / Риски / Изменения / Бюджет / Прогресс / НСИ**
- **ИИ-советчики** — два блока AiChat (Юрист, Финконтролёр), human-in-the-loop

Карточки записей скрыты из сайдбара (`visible: false`).

## Rule chains

В памяти сервера (после рестарта — повторный `apply.mjs`):

| ID | Где | Что делает |
|---|---|---|
| `invest-submit-approval` | Карточка документа | Engine: in_review + автоверсия + шаги `approvals` |
| `invest-approve-document` / `invest-reject-document` | Карточка документа | Engine: закрыть шаг; документ approved только когда все шаги пройдены |
| `invest-escalate-approval` | Карточка документа | Шаг escalated + новый шаг PMO + mail |
| `invest-simulate-rfc` / `invest-approve-rfc` / `invest-reject-rfc` | Карточка RFC | Симуляция EAC; утверждение двигает baseline + журнал + EVM |
| `invest-clone-wbs` | Карточка проекта | Копия `wbs_templates` выбранного типа конструкции |
| `invest-recalculate-evm` | Дашборд, проект | Search WBS/факты → calc-evm → SPI/CPI/EAC |
| `invest-recalculate-evm-fact` | Карточка факта | То же умение; `projectID` из поля `project` |
| `invest-critical-path` | Дашборд | HTTP POST `/critical-path` |
| `invest-threshold-alert` | Дашборд | HTTP POST `/alerts` (документы, CPI, резерв, RFC) |
| `invest-lawyer-review` | Карточка договора | AI-узел «Юрист» |
| `invest-fin-review` | Бюджет | AI-узел «Финконтролёр» |
| `invest-flag-low-cpi` | Карточка WBS | Риск, если CPI < 0.9 |
| `invest-summarize-document` | Карточка документа (авто + кнопка) | Извлечь текст файла → AI summary → поля `summary` / `extracted_text` |

После загрузки файла в `documents` цепь стартует с `afterCreate`/`afterUpdate` (файл сменился). Узел `document.extract` читает вложение in-process (docx, xlsx, pdf, dxf, ifc; DWG/PLN — harvest или `dwg2dxf` на PATH). Агент `assistant` пишет `summary`. Повтор — кнопка «Обновить summary». Сканированный PDF без текстового слоя → `failed` / needs_ocr (OCR не в v1).

## calc-evm

Пересчёт EVM больше не идёт в invest-engine. Цепочка грузит `wbs_items` и `progress_facts`, POST на `CALC_EVM_URL/call/evm`, затем пишет метрики обратно. Engine остаётся для согласований, RFC и CPM.

```bash
cd agents/services/calc-evm
go run ./cmd/calc-evm --listen=:8088
```

## invest-engine
```bash
cd agents/invest
# --api = origin without /compose (see applied.json engine.flags).
# Default is http://localhost:3333; the client also probes /api if you pass the old flag.
go run . \
  --api=http://localhost:3333 \
  --namespace=<namespaceID> \
  --token="$TOKEN" \
  --listen=:8086 \
  --alerts-every=5m
```

| Метод | Назначение |
|---|---|
| `POST /api/recalculate-evm` | PV/EV/AC/SPI/CPI/EAC по WBS и фактам прогресса, агрегация на проект |
| `POST /api/critical-path` | CPM по предшественникам, флаг `is_critical` |
| `POST /api/alerts` | Просроченные документы/RFC, CPI, резерв ≤ 0 → `risks` |
| `POST /api/submit-approval` | Маршрут + автоверсия документа |
| `POST /api/decide-approval` | Шаг approved/rejected (`decision`) |
| `POST /api/escalate-approval` | Эскалация на PMO |
| `POST /api/simulate-rfc` | EAC до/после и прогноз финиша, без смены baseline |
| `POST /api/approve-rfc` | Применение RFC + журнал + EVM (нужна симуляция) |
| `POST /api/reject-rfc` | RFC → отклонён |
| `POST /api/clone-wbs` | WBS из шаблона типа конструкции проекта |
| `GET /api/health` | Жив |

Тело POST: `{ "namespaceID", "projectID", "recordID", "userID", "decision", "token", "cpiThreshold" }`. Планировщик `--alerts-every` сам гоняет алерты и EVM (нужен `--token`).

## ИИ-советчики

Страница `/ns/invest` → **ИИ-советчики**. Промпты задают режим советчика: читать `contracts`/`documents` (Юрист) и `budget_lines`/`wbs_items`/`change_requests` (Финконтролёр), не менять статусы. Решение — кнопки rule chain.

## Отдельное согласование (не в этом apply)

Эти требования ТЗ **намеренно не реализованы** — нужна отдельная проработка:

| Тема | FR | Почему отдельно |
|---|---|---|
| УКЭП / КриптоПро / DSS | FR-017 | В платформе нет квалифицированной подписи; поле `sign_status` — заглушка |
| Боевая двусторонняя 1С | FR-011 | На `budget_lines` только выключенный ETL REST-URL |
| Импорт/экспорт Primavera P6 / MS Project XML | FR-012 | Нет парсера XML расписания |
| Мобильный офлайн стройконтроль | FR-026 | Есть веб-форма `progress_facts` (фото + гео); офлайн-клиент — отдельное приложение |
| Пакет для банка XML 115-ФЗ | FR-008 | Future в ТЗ |
| ФСТЭК / КИИ | FR-033 | Сертификация под заказчика |

`data_model/` — envoy-описание той же модели (handles, не live ID).
