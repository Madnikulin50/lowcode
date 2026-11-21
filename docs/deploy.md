# Развёртывание со стендами и пространствами

Инструкция по подъёму платформы **и агентов** так, как это сделано в `docker-conf/`. Только сервер, пустой Compose, без `apply.mjs` — в [deploy-clean.md](deploy-clean.md). Архитектура стека — в [architecture.md](architecture.md).

Рабочие стенды — **test9** и **test11**. Это не «dev» и «prod», а два независимых инстанса на одной машине: разные порты, разные базы PostgreSQL, свои секреты. Их можно держать запущенными одновременно.

```bash
docker-conf/
  test9/     # сервер :28083 + агенты 8085–8089
  test11/    # сервер :28084 + агенты 8095–8099 + stroykontrol :8100
```

Корневой `docker-compose.yml` поднимает **только сервер** (без агентов). Отдельный `agents/docker-compose.yml` — агенты + MinIO, без сервера. Для полного контура используйте `docker-conf/`.

---

## 1. Что входит в стенд

| Сервис | Образ | test9 | test11 |
|--------|-------|-------|--------|
| Сервер (API + webapp) | `docker.io/madnikulin50/pnp-lowcode:${VERSION}` | `:28083` | `:28084` |
| cmdb | `…/pnp-cmdb` | `:8085` | `:8095` |
| invest | `…/pnp-invest` | `:8086` | `:8096` |
| backup | `…/pnp-backup` | `:8087` | `:8097` |
| calc-evm | `…/pnp-calc-evm` | `:8088` | `:8098` |
| scan-cidr | `…/pnp-scan-cidr` | `:8089` | `:8099` |
| stroykontrol | `pnp-stroykontrol:${VERSION}` (только локальная сборка) | нет | `:8100` |

PostgreSQL, Ollama и MinIO **в compose стенда нет**: они ожидаются на хосте (или в соседних контейнерах) и доступны контейнерам как `host.docker.internal`.

```mermaid
flowchart LR
  Browser["Браузер"] --> Server["pnp-lowcode\n:HTTP_PORT"]
  Server --> PG[("PostgreSQL на хосте\ntest9 / test11")]
  Server --> Data["том ./data/server"]
  Server -.-> Ollama["Ollama :11434"]
  Agents["агенты"] -->|"host.docker.internal:HTTP_PORT"| Server
  Backup["backup"] --> MinIO["MinIO :9000"]
```

---

## 2. Требования

На хосте:

- Docker Engine + Compose v2
- PostgreSQL 13+ на `:5432` (не в этом compose)
- свободные порты стенда (см. таблицу выше)
- для чата и LLM-узлов — [Ollama](https://ollama.com/) на `:11434`
- для backup — MinIO на `:9000` (см. [§8](#8-minio-для-backup))
- для stroykontrol — локально собранный образ (см. [§6](#6-сборка-образов))

Контейнеры ходят на хост через `extra_hosts: host.docker.internal:${LOCAL}`. `LOCAL` — IPv4 этой машины в LAN (не `127.0.0.1`: из контейнера loopback — это сам контейнер).

cmdb и scan-cidr получают `cap_add: NET_RAW, NET_ADMIN` — без них ICMP/ARP-скан не работает.

---

## 3. Подготовка PostgreSQL

Создайте пустую базу под стенд. Имя должно совпасть с `DB_NAME` / хвостом `DB_DSN`.

```sql
CREATE DATABASE test9;
CREATE DATABASE test11;
```

Пользователь из `DB_DSN` должен уметь создавать таблицы: сервер сам накатывает схему при старте (`UPGRADE_ALWAYS` по умолчанию включён в образе).

Проверка с хоста:

```bash
psql "postgres://postgres@127.0.0.1:5432/test9?sslmode=disable" -c 'SELECT 1'
```

---

## 4. Файл `.env`

В каждом каталоге стенда уже есть `.env`. Перед первым запуском на новой машине **обязательно** поправьте адреса, пароль БД и секреты. Не копируйте секреты из чужого стенда.

Минимальный набор:

```bash
# Тег образов (должен совпасть с тем, что в Docker Hub / локально)
VERSION=2026.09.20

# Как браузер и агенты видят сервер
DOMAIN=<LAN-IP>:<HTTP_PORT>
LOCAL=<LAN-IP>
HTTP_PORT=28083

# БД на хосте. Из контейнера хост = host.docker.internal
DB_DSN=postgres://<user>:<password>@host.docker.internal:5432/<db>?sslmode=disable
DB_USER=<user>
DB_PASSWORD=<password>
DB_NAME=<db>

HTTP_API_ENABLED=true
HTTP_BASE_URL=/
ENVIRONMENT=dev
LOCALE_LANGUAGES=ru
ACTIONLOG_ENABLED=false

OLLAMA_URL=http://host.docker.internal:11434

# Одна и та же строка у сервера и агентов. Сервер по ней выдаёт JWT (POST /agents/enroll).
AGENT_SHARED_SECRET=<случайная hex-строка>
TOKEN=

DOCKER_IMAGE_PREFIX=docker.io/madnikulin50/
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=backups
RESTIC_PASSWORD=
```

Сгенерировать секрет:

```bash
openssl rand -hex 24
```

Замечания:

| Переменная | Смысл |
|------------|--------|
| `DOMAIN` | Host:port, который попадёт в ссылки и cookie auth. Для доступа с других машин — LAN IP, не localhost. |
| `LOCAL` | Тот же IP: Docker прописывает `host.docker.internal` → этот адрес. |
| `HTTP_PORT` | Публикация `80` контейнера сервера. test9 = `28083`, test11 = `28084`. |
| `AGENT_SHARED_SECRET` | Если пусто, маршрут `/agents/enroll` не монтируется. Тогда агентам нужен ручной `TOKEN`. |
| `TOKEN` | Устаревший ручной JWT. Оставьте пустым, если задан shared secret. |
| `DOCKER_IMAGE_PREFIX` | Префикс образов агентов. Пустая строка — локальные теги после `make ddebug`. |

Образ сервера сам включает webapp (`HTTP_WEBAPP_ENABLED=true`, файлы в `/pnp/webapp`). Отдельно прописывать пути к SPA не нужно.

---

## 5. Запуск стенда

### test9

```bash
cd docker-conf/test9
docker compose pull
docker compose up -d
docker compose ps
```

UI: `http://<DOMAIN>/` (например `http://192.168.173.74:28083/`).  
Health: `http://<DOMAIN>/healthcheck`.

### test11

Образ stroykontrol **не публикуется** в Docker Hub. Сначала соберите его:

```bash
make -C agents/stroykontrol ddebug VERSION=2026.09.20
```

Получится локальный тег `pnp-stroykontrol:2026.09.20`. Затем:

```bash
cd docker-conf/test11
docker compose pull          # stroykontrol pull пропустит — это нормально
docker compose up -d
```

UI: порт `28084`.

Пока сервер накатывает схему, `/healthcheck` может не отвечать — в Dockerfile `start-period` = 1 минута. Смотрите логи:

```bash
docker compose logs -f server
```

Первый пользователь: в `ENVIRONMENT=dev` откройте UI и зарегистрируйтесь (signup). Либо задайте `AUTH_PROVISION_SUPER_USER` в `.env` до первого старта (см. `server/.env.example`).

---

## 6. Сборка образов

Если образы с Hub не подходят (свой патч, офлайн), соберите их в корне репозитория.

Сервер + встроенные webapp:

```bash
# Сборка бинаря и SPA, затем docker build + push
make drelease VERSION=2026.09.20

# Только локальный тег pnp-lowcode:VERSION (без push).
# Нужны уже собранные client3/web/*/dist и проверка cdist/webapp/compose/index.html
make -C client3 build
make ddebug VERSION=2026.09.20
```

`drelease` делает `server/make build`, `client3/make build`, затем `docker build -t madnikulin50/pnp-lowcode:$VERSION .` по корневому `Dockerfile`.

Агенты (backup, cmdb, invest, calc-evm, scan-cidr):

```bash
make drelease-agents VERSION=2026.09.20   # build + push docker.io/madnikulin50/pnp-*
make ddebug-agents VERSION=2026.09.20     # локальные теги pnp-backup:… и т.д.
```

stroykontrol в `agents/Makefile` не входит:

```bash
make -C agents/stroykontrol ddebug VERSION=2026.09.20
```

Для стенда на локальных тегах в `.env`:

```bash
DOCKER_IMAGE_PREFIX=
```

и в `docker-compose.yml` сервера замените image на `pnp-lowcode:${VERSION}` (так уже сделано в `docker-debug/`).

---

## 7. Пространства Compose (apply)

Контейнеры поднимают **платформу**. Модули, страницы и rule chains продуктов создаёт `apply.mjs` по REST. Скрипты по умолчанию ищут API на `:3333` (GoLand/dev). Для docker-conf укажите порт стенда явно.

Сначала один раз зайдите в UI (нужна живая сессия: mint JWT читает `auth_oa2tokens`). Дальше с хоста:

### test9 (`:28083`, агентские порты 8085–8089)

```bash
export COMPOSE_API=http://127.0.0.1:28083/api/compose
export COMPOSE_DSN='postgres://USER:PASS@127.0.0.1:5432/test9?sslmode=disable'
export AUTH_API=http://127.0.0.1:28083

cd agents/backup/compose
BACKUP_AGENT_URL=http://localhost:8087/api node apply.mjs

cd ../../cmdb/compose
CMDB_AGENT_URL=http://localhost:8089/api node apply.mjs

cd ../../invest/compose
INVEST_ENGINE_URL=http://localhost:8086/api \
CALC_EVM_URL=http://localhost:8088/api \
  node apply.mjs
```

Открыть: `/ns/backup`, `/ns/cmdb`, `/ns/invest`.

Демо-данные invest:

```bash
cd agents/invest/compose
COMPOSE_API=http://127.0.0.1:28083/api/compose node seed.mjs
```

### test11 (`:28084`, порты агентов сдвинуты)

```bash
export COMPOSE_API=http://127.0.0.1:28084/api/compose
export COMPOSE_DSN='postgres://USER:PASS@127.0.0.1:5432/test11?sslmode=disable'
export AUTH_API=http://127.0.0.1:28084

cd agents/stroykontrol/compose
STROYKONTROL_WEB_URL=http://localhost:8100 node apply.mjs
```

`apply.mjs` пишет `applied.json`. В `docker-conf/test11/docker-compose.yml` у stroykontrol зашит `--namespace=<ID>`. После apply подставьте ID из `applied.json` (поле namespace) и пересоздайте контейнер:

```bash
cd docker-conf/test11
docker compose up -d --force-recreate stroykontrol
```

Опционально те же backup/cmdb/invest, но URL агентов — `8097` / `8099` / `8096` / `8098`.

Seed стройконтроля (тяжёлый, тысячи записей и docx):

```bash
cd agents/stroykontrol/compose
COMPOSE_DSN='postgres://USER:PASS@127.0.0.1:5432/test11?sslmode=disable' \
COMPOSE_API=http://127.0.0.1:28084/api/compose \
  node seed.mjs
```

Скрипты идемпотентны по `handle`. Повторный apply не сбрасывает ручную вёрстку страниц, если не передать `--reset-pages` / `APPLY_RESET_PAGES=1` (cmdb).

Rule chains держатся в памяти процесса: после `docker compose restart server` снова выполните `apply.mjs`.

---

## 8. MinIO для backup

Стенды **не** поднимают MinIO: backup-агент ходит на `host.docker.internal:9000`. Отдельный compose:

```bash
cd agents/backup
docker compose up -d
```

- S3 API: `http://127.0.0.1:9000`
- консоль: `http://127.0.0.1:9001`
- ключи по умолчанию: `minioadmin` / `minioadmin`
- бакет `backups` агент создаёт сам

Ключи в `.env` стенда (`MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`) должны совпасть. Пароли источников бэкапа (SMB, БД) не кладутся в Compose: на агенте задайте `BACKUP_SECRET_<handle>`.

Альтернатива: MinIO внутри `agents/docker-compose.yml` (сервис `minio`, агент `--minio-endpoint=minio:9000`). Тогда backup из `docker-conf` нужно перенацелить — сейчас он жёстко смотрит на хост `:9000`.

---

## 9. Ollama

Сервер читает `OLLAMA_URL`. На хосте:

```bash
ollama serve          # если ещё не запущен как сервис
ollama pull qwen3:8b  # модель чата по умолчанию (CHAT_MODEL)
```

Из контейнера URL — `http://host.docker.internal:11434`. Проверка: в Compose открыть блок AiChat или `curl http://127.0.0.1:11434/api/tags`.

---

## 10. Как агенты находят сервер

В command агентов:

```
--api=http://host.docker.internal:${HTTP_PORT}
```

у stroykontrol — `…/${HTTP_PORT}/api/compose`.

Токен:

1. Предпочтительно `AGENT_SHARED_SECRET` (одинаковый в `.env` сервера и `environment` агента) → `POST /agents/enroll`.
2. Иначе `TOKEN` — длинный JWT (mint из `auth_oa2tokens`, см. `agents/stroykontrol/mint-token.sh`). Короткий токен из UI (~2 часа) для долгоживущего контейнера не годится.

calc-evm и scan-cidr к Compose сами не ходят: их вызывают rule chains, результат пишут цепочки.

Проверка агента:

```bash
curl -sS http://127.0.0.1:8087/api/health   # backup, test9
curl -sS http://127.0.0.1:8085/api/meta     # cmdb
curl -sS http://127.0.0.1:8088/api/meta     # calc-evm
```

---

## 11. Два стенда сразу

Порты test11 сдвинуты специально. Одновременно:

| | test9 | test11 |
|--|-------|--------|
| UI / API | 28083 | 28084 |
| БД | `test9` | `test11` |
| агенты | 8085–8089 | 8095–8099, 8100 |
| том вложений | `docker-conf/test9/data/server` | `docker-conf/test11/data/server` |

Не пересекаются: `HTTP_PORT`, `DB_NAME`, `AGENT_SHARED_SECRET`, опубликованные порты агентов. Общие на хосте: PostgreSQL `:5432`, Ollama `:11434`, MinIO `:9000` (оба backup пишут в один бакет, если не развести `MINIO_BUCKET`).

---

## 12. Данные и остановка

```bash
cd docker-conf/test9
docker compose stop          # остановить
docker compose down          # контейнеры; тома named (cmdb-data) сохраняются
docker compose down -v       # ещё и named volumes — SQLite CMDB пропадёт
```

Что где лежит:

| Что | Где |
|-----|-----|
| Метаданные и записи Compose | PostgreSQL (`test9` / `test11`) |
| Вложения, объектное хранилище сервера | `docker-conf/<стенд>/data/server` |
| SQLite CMDB (режим embedded) | volume `cmdb-data` |
| Снапшоты backup | MinIO |

Перенос БД на другой хост: дамп PostgreSQL + копирование `data/server`. Пример разрушающего restore — `scripts/migrate-test11-to-db_ru5.sh` (меняет **удалённую** базу, после restore нужен `docker compose restart server`). Вложения скрипт не копирует.

---

## 13. Только сервер

Без агентов:

```bash
# из корня репозитория, свой .env рядом
docker compose up -d
```

или отладочный образ без Docker Hub:

```bash
cd docker-debug
# image: pnp-lowcode:${VERSION} — сначала make ddebug
docker compose up -d
```

`.env` тот же набор переменных (`VERSION`, `LOCAL`, `HTTP_PORT`, `DB_DSN`, …).

---

## 14. Типичные проблемы

**Контейнер не видит Postgres / Ollama.** `LOCAL` должен быть LAN-IP хоста. `127.0.0.1` в `LOCAL` или в `DB_DSN` внутри контейнера указывает на сам контейнер. С хоста DSN — `127.0.0.1`, из compose — `host.docker.internal`.

**Агенты 401.** Пустой `AGENT_SHARED_SECRET` у сервера (маршрут enroll не смонтирован) или разные строки у сервера и агента. После смены секрета: `docker compose up -d --force-recreate`.

**stroykontrol: empty comparison / 404 namespace.** В command зашит чужой `--namespace`. Возьмите ID из `agents/stroykontrol/compose/applied.json` после apply.

**Не стартует stroykontrol.** Нет локального образа — `make -C agents/stroykontrol ddebug`. Внутри нужен `pdftoppm` (есть в Dockerfile).

**Чат не отвечает.** Ollama не слушает `:11434` или модель не скачана. `OLLAMA_URL` в `.env` сервера.

**backup health красный.** MinIO не поднят на `:9000` или другие ключи, чем в `.env`.

**UI открывается, API 404 на `/api`.** Образ отдаёт API и webapp с одного порта; `HTTP_BASE_URL=/`. Не ставьте одинаковый base URL у API и webapp (сервер тогда отключит webapp).

**Порты заняты.** Второй стенд — только test11 со сдвинутыми портами. Либо остановите test9.

**apply.mjs не находит API.** Без `COMPOSE_API` скрипт стучится на `:3333`. Для docker-conf всегда задавайте `COMPOSE_API` и `COMPOSE_DSN`. Нужен хотя бы один логин в UI, иначе mint JWT из `auth_oa2tokens` падает.

**После рестарта сервера кнопки цепочек «пустые».** Повторный `node apply.mjs`.
