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

Скрипт идемпотентен (по handle). Пишет `applied.json`.

Открыть: `/ns/backup`.

## MinIO

Snap `minio` на этом хосте не стартует (сборка 2018 + glibc). Локальный destination — Docker:

```bash
cd agents/backup
docker compose up -d
```

API `http://127.0.0.1:9000`, консоль `http://127.0.0.1:9001`, ключи `minioadmin` / `minioadmin`, бакет `backups` (агент создаёт сам).

## Агент

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
| `POST /api/jobs` | Старт бэкапа (`sourceID` или `policyID`) |
| `POST /api/jobs/due` | Политики, у которых cron совпал |
| `GET /api/jobs/{id}` | Статус |
| `POST /api/restore` | Восстановление снапшота |
| `POST /api/prune` | Retention |
| `POST /api/register` | Heartbeat в модуль agents |
| `GET /api/health` | Жив + MinIO |
| `GET /metrics` | Prometheus |

## Модель

| Модуль | Роль |
|---|---|
| `agents` | Хосты агента, heartbeat |
| `credentials` | Handle секрета, не пароль |
| `sources` | fs / smb / database / s3 |
| `policies` | cron, retention, incremental |
| `jobs` | Запуски |
| `snapshots` | Ключ MinIO, checksum |
| `restores` | Операции восстановления |

Инкремент файлов — restic, если бинарь в `PATH` и на политике включён incremental. Иначе полный tar.gz / dump / s3 copy.

`data_model/` — envoy-описание (handles, не live ID).
