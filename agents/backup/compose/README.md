# Backup Lowcode namespace

Пространство резервного копирования: SMB, локальные папки, PostgreSQL/MySQL и внешний S3 → внутренний MinIO.

## Apply

Сервер на `:3333`. JWT берётся из живой строки `auth_oa2tokens` (нужен хотя бы один логин в UI).

```bash
cd agents/backup/compose
node apply.mjs
```

| Variable | Default |
|---|---|
| `COMPOSE_API` | auto-detect `http://127.0.0.1:3333/compose` |
| `COMPOSE_DSN` | `test10` (GoLand Server RU test10), then `test9` |
| `TOKEN` | mint from `auth_oa2tokens` (override if set) |
| `BACKUP_AGENT_URL` | `http://localhost:8087/api` |

Скрипт идемпотентен (по handle). Пишет `applied.json`. При пустом модуле `destinations` создаёт дефолтное назначение из env агента (`MINIO_ENDPOINT`/`MINIO_BUCKET`/`MINIO_SECURE`).

Открыть: `/ns/backup`.

## Назначения (куда сохраняется бэкап)

Управляется записью модуля `destinations` (страница «Назначения»). Тип (`type`) определяет используемые поля:

| Тип | Поля | Учётки |
|---|---|---|
| `minio` (S3-совместимое) | endpoint, bucket, region, secure | credential: access/secret |
| `local` | path (корень на агенте) | — |
| `nfs` | endpoint=server:/export, path=mountpoint, options=mount opts | монтирование: mount(8), иначе ошибка с инструкцией |
| `sftp` | endpoint=хост, port (default 22), path=base dir | credential: username + BACKUP_SECRET (пароль или PEM-ключ; fallback — SSH-agent). Либо логин/пароль прямо в форме назначения (поля username/password — перекрывают credentials; хранятся открыто) |
| `smb` | endpoint=хост, share, path=подкаталог, options=`{"port","domain"}` | credential: username + BACKUP_SECRET (пароль). Либо логин/пароль прямо в форме (username/password) |
| `rclone` | remote (имя из `rclone config` агента), path=подкаталог | как настроено в rclone |
| `webdav` | endpoint=URL, path=подкаталог | credential: username + пароль. Либо прямо в форме (username/password) |

Одно назначение может быть отмечено `default` — его получают политики без явной связи. Кнопка «Проверить» вызывает `backup/dest-test` (агент пишет last_check* в запись).

Резолвинг для прогона: `destinationID` в вызове `backup/run` → `policies.destination` → legacy `policies.dest_bucket` → default-назначение → env агента (`MINIO_*`). `policies.dest_prefix` перекрывает префикс назначения.

Снапшот запоминает фактическое место: `destination`, `dest_type`, `s3_endpoint` (локатор: endpoint S3, `file:///…`, `sftp://…`, `rclone:…`, `smb://…`), `s3_bucket` (бакет/корень). Restore/prune/scan/verify/extract читают через стор снапшота: не-S3 назначения переподключаются через запись destinations (учётки в снапшот не денормализуются), S3 — по endpoint+bucket с учётками агента. Копия в сейф с не-S3 назначения идёт потоком через агента (`VaultStore.CopyRetainedFrom`).

Движки: `archive` работает на всех типах; `restic` — minio/local/nfs/sftp/rclone; `kopia` — minio/local/nfs/sftp (нативные backend'ы; для smb/webdav используйте archive или настройте rclone-remote). Бинари `restic`/`kopia`/`rclone` должны быть в PATH агента (проверка — при запуске джобы).

Миграция старых политик с `dest_bucket`:

```bash
node migrate-destinations.mjs   # MINIO_ENDPOINT — endpoint legacy-бакетов
```

## MinIO

Snap `minio` на этом хосте не стартует (сборка 2018 + glibc). Локальный destination — Docker:

```bash
cd agents/backup
docker compose up -d
```

API `http://127.0.0.1:9000`, консоль `http://127.0.0.1:9001`, ключи `minioadmin` / `minioadmin`, бакет `backups` (агент создаёт сам).

## Агент

SDK-контракт: все операции идут в `POST /api/jobs` с полем `operation` (`backup`, `restore`, `prune`, `due`). Rule chains вызывают типизированные ноды `backup/run`, `backup/restore`, `backup/prune`, `backup/due` — без ручных HTTP/detach.

```bash
cd agents/backup
go run . \
  --api=http://127.0.0.1:3333 \
  --namespace=<namespaceID> \
  --token="$TOKEN" \
  --listen=:8087 \
  --minio-endpoint=127.0.0.1:9000 \
  --minio-access=minioadmin \
  --minio-secret=minioadmin \
  --minio-bucket=backups
```

Секреты источников: `BACKUP_SECRET_<handle>` в окружении агента (пароль SMB/БД/S3). Username — в записи `credentials`.

| Метод | Назначение |
|---|---|
| `POST /api/jobs` | SDK: `operation=backup\|restore\|prune\|due` |
| `POST /api/call/dest-test` | Проверить назначение (sync, пишет `last_check*`) |
| `GET /api/jobs/{id}` | Статус |
| `POST /api/restore` | Alias → `operation=restore` |
| `POST /api/prune` | Alias → `operation=prune` |
| `POST /api/jobs/due` | Alias → `operation=due` |
| `GET /api/meta` | Дескрипторы компонентов |
| `POST /api/register` | Heartbeat в модуль agents |
| `GET /api/health` | Жив + MinIO |
| `GET /metrics` | Prometheus |

## Модель

| Модуль | Роль |
|---|---|
| `agents` | Хосты агента, heartbeat |
| `credentials` | Handle секрета, не пароль |
| `sources` | fs / smb / database / s3 |
| `destinations` | Куда сохранять: endpoint/бакет/префикс/учётки |
| `policies` | cron, retention, incremental, ссылка на назначение |
| `jobs` | Запуски |
| `snapshots` | Ключ MinIO, endpoint, checksum |
| `restores` | Операции восстановления |

Инкремент файлов — restic, если бинарь в `PATH` и на политике включён incremental. Иначе полный tar.gz / dump / s3 copy.

`data_model/` — envoy-описание (handles, не live ID).
