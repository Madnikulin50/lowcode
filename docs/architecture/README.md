# Архитектура платформы low-code — диаграммы C4

Диаграммы по модели [C4](https://c4model.com/) для репозитория `lowcode`
(форк Corteza с агентами, rule chains, MCP и LLM-интеграцией).
Открываются в draw.io / diagrams.net: **File → Open** или drag-and-drop файла
на страницу [app.diagrams.net](https://app.diagrams.net).

Текстовое описание архитектуры — в [../architecture.md](../architecture.md),
репозиторий сервисов и сценарии — в [../agents.md](../agents.md).

## Файлы

| Файл | Уровень | Содержание |
|------|---------|------------|
| `c4-level1-system-context.drawio` | C1 — System Context | Платформа целиком как система, пользователи и внешние системы (Ollama, SMTP, 1С/REST, IdP, MCP-клиенты) |
| `c4-level2-containers.drawio` | C2 — Containers | Контейнеры внутри платформы: 7 SPA, Corteza Server, MCP, Corredor, агенты, server-discovery, RDBMS, объектное хранилище |
| `c4-level3-components.drawio` | C3 — Components | 4 страницы: C3.1 Corteza Server, C3.2 SPA Compose, C3.3 агент на SDK, C3.4 автоматизация и AI |

## Как читать

- **Синие прямоугольники** — система (C1) и контейнеры (C2).
- **Голубые прямоугольники** — компоненты (C3), сгруппированы пунктирными
  границами по слоям/пакетам.
- **Серые прямоугольники** — внешние системы и контейнеры вне границы.
- **Фигуры человека** — пользователи (акторы).
- **Цилиндры** — хранилища данных.
- Пунктирные стрелки — опциональные/асинхронные связи (gRPC к Corredor,
  колбэки агентов, enroll и т.п.).

## Уровень C1 — System Context

Платформа low-code как единая система: приложения Compose, workflow,
rule chains, вычислительные агенты, AI-чат и MCP-сервер. Взаимодействует с:

- пользователями (бизнес-пользователь, администратор, разработчик) — по HTTPS;
- внешними LLM-клиентами — по MCP (STDIO/SSE);
- Ollama — генерация и tool calling (HTTP, eino-ext);
- SMTP, внешними бизнес-системами (1С, REST), корпоративным IdP (SAML/OIDC).

## Уровень C2 — Containers

| Контейнер | Технология | Роль |
|-----------|-----------|------|
| compose / admin / one / workflow / reporter / discovery / privacy | Vue 3.5, Vite 6 | 7 независимых SPA поверх двух общих библиотек (`client3/lib`) |
| Corteza Server | Go 1.26, chi | монолит: REST `/api/{system,compose,automation,…}`, `/auth`, WebSocket; доменные сервисы + `server/pkg/*` |
| MCP Server | in-process | STDIO/SSE :9090, tools для внешних LLM-клиентов |
| Corredor | gRPC (опц.) | JS-рантайм скриптов automation |
| cmdb / invest / backup / stroykontrol | Go | продуктовые агенты (порты 8085–8092) |
| calc-evm / scan-cidr | Go | скиллы (порты 8088–8089), записи в Compose не пишут |
| server-discovery + Elasticsearch | Go + ES | глобальный поиск: индексатор + индекс |
| RDBMS | PostgreSQL / MySQL / SQLite | store (метаданные) + DAL (таблицы записей) |
| Объектное хранилище | FS / MinIO / S3 | вложения, экспорт, сканы |

Ключевые протоколы: SPA → Server (REST + WebSocket, JWT), Server → агенты
(HTTP, `lowcode.agent.v1`), агенты → Server (Compose REST, JWT через
`TOKEN` или `/agents/enroll`).

## Уровень C3 — Components

Файл содержит 4 страницы (переключатель внизу draw.io):

1. **C3.1 Corteza Server** — HTTP-слой (router, rest-пакеты), доменные
   сервисы (system / compose / automation) и платформенные пакеты
   `server/pkg/` (безопасность, данные, события/автоматизация, AI).
2. **C3.2 SPA Compose** — Router + Pinia stores, Views (Public/Admin),
   PageBlocks (24 вида), Public/Admin-компоненты, общие `lib/js` и `lib/vue`.
3. **C3.3 Агент** — `sdk.Service` (маршруты chi), Backend (sync/async),
   Job (Envelope, колбэк), домен, `sdk.Client` (Compose REST), SelfToken
   (enroll). Эталон: backup / cmdb / стройконтроль.
4. **C3.4 Автоматизация и AI** — контур Automation (Trigger → Workflow →
   Session → wfexec, AI-функции шагов), контур Rule chains (Engine, типы
   узлов, persistence, мосты) и общий AI-контур (chat, MCP handlers,
   реестр aiagent, LLM client → Ollama).

## Обновление диаграмм

Файлы — обычный XML draw.io (несжатый), удобно править руками или
редактором. После правок проверяйте XML (например, `xmllint --noout` или
открытием в draw.io). При изменении архитектуры (новые агенты, пакеты,
протоколы) обновляйте соответствующий уровень и эту таблицу.
