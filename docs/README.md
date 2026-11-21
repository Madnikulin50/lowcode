# Документация

Архитектурное описание платформы **lowcode** (форк Corteza с агентами, MCP и AI-расширениями).

| Документ | Содержание |
|----------|------------|
| [architecture.md](architecture.md) | Архитектура платформы: контекст, сервер, веб-приложения, агенты, данные, AI/MCP |
| [rest-api.md](rest-api.md) | Схема работы REST API: регистрация команд, старт сервера, путь запроса, взаимодействие с БД |
| [deploy-clean.md](deploy-clean.md) | Чистое развёртывание: только сервер и пустой Compose, без агентов и без `apply` пространств |
| [deploy.md](deploy.md) | Стенды `docker-conf/test9` и `test11`: агенты, apply namespace, MinIO, Ollama |
| [agents.md](agents.md) | Руководство разработчика агентов: SDK, sync/async, Compose apply, палитра rule chains |
| [dal.md](dal.md) | Новые объекты: модули Compose, таблицы DAL, кодоген, чтение и запись |
| [ai-guard.md](ai-guard.md) | AI guard: таргетированные проверки после правок ИИ — режимы, baseline, периодический CI |
| [security/threat-model.md](security/threat-model.md) | Модель угроз: активы, границы доверия, нарушители, каталог угроз STRIDE/AI, план снижения риска |
| [security/attack-surface.md](security/attack-surface.md) | Поверхность атаки: порты, маршруты, каналы ввода, выполнение кода, цепочка поставки |
| [security/kii-categorization.md](security/kii-categorization.md) | Категорирование КИИ: методика, рабочий лист оператора, профиль безопасной конфигурации, финальный чеклист |
| [stroykontrol-user.md](stroykontrol-user.md) | Руководство пользователя агента Стройконтроль: объекты, сравнение ПД/РД, замечания, ИД, отчёты |

Продуктовые руководства пользователя и интегратора живут в `manual/` (Antora). Исключение — [stroykontrol-user.md](stroykontrol-user.md). Остальное в этом каталоге — инженерное описание текущего кода репозитория.
