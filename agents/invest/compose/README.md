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
| **Документы** (`documents`) | Реестр + статус согласования + заглушка УКЭП (`sign_status`) |
| **Версии** (`document_versions`) | Обход отсутствия file-versioning |
| **Согласования** (`approvals`) | Маршрут: кто / решение / срок |
| **Комментарии** (`document_comments`) | Блок Comment на карточке документа |
| **Договоры** (`contracts`) | Реестр договоров, пакет документов |
| **Риски** (`risks`) | Вероятность / влияние / митигация |
| **RFC** (`change_requests`) | Запросы на изменение |
| **Журнал** (`change_log`) | Аудиторский след RFC |
| **Статьи бюджета** (`budget_lines`) | План / факт / резерв; ETL-заглушка 1С (выключена) |
| **Денежный поток** (`cashflow_items`) | Приход / расход |
| **Факты прогресса** (`progress_facts`) | Веб-фиксация объёма + фото + гео |
| **НСИ** | `document_types`, `counterparties`, `materials`, `labor_norms` |
| **ИИ-советчики** (`ai_advisors`) | Юрист, Финконтролёр (промпты) |

Record revisions включены на `projects`, `wbs_items`, `documents`, `contracts`, `change_requests`, `budget_lines`.

## Страницы

- **Дашборд** — метрики, кнопки EVM/CPM/алерты, doughnut, Gantt, списки «на согласовании»
- **Документы** — канбан статусов + реестр; карточка: файл, версии, маршрут, комментарии, кнопки согласования
- **Проекты / WBS / Договоры / Риски / Изменения / Бюджет / Прогресс / НСИ**
- **ИИ-советчики** — два блока AiChat (Юрист, Финконтролёр), human-in-the-loop

Карточки записей скрыты из сайдбара (`visible: false`).

## Rule chains

В памяти сервера (после рестарта — повторный `apply.mjs`):

| ID | Где | Что делает |
|---|---|---|
| `invest-submit-approval` / `invest-approve-document` / `invest-reject-document` / `invest-escalate-approval` | Карточка документа | Смена статуса |
| `invest-submit-rfc` / `invest-approve-rfc` / `invest-reject-rfc` | Карточка RFC | Статус + строка `change_log` при утверждении |
| `invest-recalculate-evm` | Дашборд, проект | HTTP POST invest-engine `/recalculate-evm` |
| `invest-critical-path` | Дашборд | HTTP POST `/critical-path` |
| `invest-threshold-alert` | Дашборд | HTTP POST `/alerts` + поиск документов на согласовании |
| `invest-lawyer-review` | Карточка договора | AI-узел «Юрист» |
| `invest-fin-review` | Бюджет | AI-узел «Финконтролёр» |
| `invest-flag-low-cpi` | Карточка WBS | Риск, если CPI < 0.9 |

## invest-engine

```bash
cd agents/invest
# --api = origin without /compose (see applied.json engine.flags).
# Default is http://localhost:3333; the client also probes /api if you pass the old flag.
go run . \
  --api=http://localhost:3333 \
  --namespace=<namespaceID> \
  --token="$TOKEN" \
  --listen=:8086
```

| Метод | Назначение |
|---|---|
| `POST /api/recalculate-evm` | PV/EV/AC/SPI/CPI/EAC по WBS и фактам прогресса, агрегация на проект |
| `POST /api/critical-path` | CPM по предшественникам (или по датам), флаг `is_critical` |
| `POST /api/alerts` | Просроченные документы, CPI ниже порога → запись в `risks` |
| `GET /api/health` | Жив |

Тело POST: `{ "namespaceID", "projectID", "token", "cpiThreshold" }` — все поля опциональны, кроме живого токена на агенте или в теле.

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
