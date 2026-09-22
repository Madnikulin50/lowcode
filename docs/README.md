# Документация

Архитектурное описание платформы **lowcode** (форк Corteza с агентами, MCP и AI-расширениями).

| Документ | Содержание |
|----------|------------|
| [architecture.md](architecture.md) | Архитектура платформы: контекст, сервер, веб-приложения, агенты, данные, AI/MCP |
| [deploy-clean.md](deploy-clean.md) | Чистое развёртывание: только сервер и пустой Compose, без агентов и без `apply` пространств |
| [deploy.md](deploy.md) | Стенды `docker-conf/test9` и `test11`: агенты, apply namespace, MinIO, Ollama |
| [agents.md](agents.md) | Руководство разработчика агентов: SDK, sync/async, Compose apply, палитра rule chains |

Продуктовые руководства пользователя и интегратора живут в `manual/` (Antora). Этот каталог — инженерное описание текущего кода репозитория.
