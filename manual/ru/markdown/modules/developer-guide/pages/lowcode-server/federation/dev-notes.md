# Локальная разработка

Используйте следующие инструкции для настройки двух федеративных инстансов LowCoooode.

## Создание каталога данных

.Команда CLI для создания каталога с правами доступа:
```bash
```
mkdir data && chmod 755 data/ -R

## Создание docker compose файла

.Файлы `docker-compose.yaml` для двух инстансов:
```yaml
```
version: "3"
services:
  db_origin:
    image: percona:8.0
    container*name: db*origin
    restart: always
    cap_add:
      - SYS_NICE  # mbind warning fix
    environment:
      MYSQL_DATABASE:      lowcode
      MYSQL_USER:          lowcode
      MYSQL_PASSWORD:      rootlowcode
      MYSQL*ROOT*PASSWORD: rootlowcode
    volumes:
      - "../../data/db_origin:/var/lib/mysql"
    ports:
      - 3306:3306

  db_destination:
    image: percona:8.0
    container*name: db*destination
    restart: always
    cap_add:
      - SYS_NICE  # mbind warning fix
    environment:
      MYSQL_DATABASE:      lowcode
      MYSQL_USER:          lowcode
      MYSQL_PASSWORD:      rootlowcode
      MYSQL*ROOT*PASSWORD: rootlowcode
    volumes:
      - "../../data/db_destination:/var/lib/mysql"
    ports:
      - 3307:3306

  node_origin:
    image: golang
    container*name: node*origin
    entrypoint: [ make, watch ]
    depends*on: [ db*origin ]
    volumes:
      - "../../:/app"
      - "./.env.orig:/app/.env"
    working_dir: /app
    restart: always
    ports:
      - 8084:8084

  node_destination:
    image: golang
    container*name: node*destination
    entrypoint: [ make, watch ]
    depends*on: [ db*destination ]
    volumes:
      - "../../:/app"
      - "./.env.dest:/app/.env"
    working_dir: /app
    restart: always
    ports:
      - 8085:8084

Этот docker compose файл создаёт два инстанса LowCoooode с соответствующими базами данных.
Загрузки бинарных данных хранятся в каталоге `data`, созданном выше.

## Создание файлов окружения

.\.env.orig:
```env
```
PROVISION_ALWAYS=false
HTTP_ADDR=:8084

LOG_LEVEL=info

DB*DSN=lowcode:rootlowcode@tcp(db*origin:3306)/lowcode?collation=utf8mb4*general*ci
LOG_DEBUG=true
CORREDOR_ADDR=localhost:50051
CORREDOR_ENABLED=false
CORREDOR*CLIENT*CERTIFICATES_ENABLED=false

GRPC*SERVER*ADDR=localhost:50052

LOWCODE*PROTOBUF*PATH=/home/wrk/Projects/lowcode/lowcode-protobuf

FEDERATION_ENABLED=true
FEDERATION*HOST=node*origin:8084
FEDERATION_LABEL=Federation origin host

# Sync settings
FEDERATION*SYNC*STRUCTURE*MONITOR*INTERVAL=60s
FEDERATION*SYNC*STRUCTURE*PAGE*SIZE=1
FEDERATION*SYNC*DATA*MONITOR*INTERVAL=20s
FEDERATION*SYNC*DATA*PAGE*SIZE=100

.\.env.dest:
```env
```
PROVISION_ALWAYS=false
HTTP_ADDR=:8084

LOG_LEVEL=info

DB*DSN=lowcode:rootlowcode@tcp(db*destination:3306)/lowcode?collation=utf8mb4*general*ci
LOG_DEBUG=true
CORREDOR_ADDR=localhost:50051
CORREDOR_ENABLED=false
CORREDOR*CLIENT*CERTIFICATES_ENABLED=false

GRPC*SERVER*ADDR=localhost:50052

LOWCODE*PROTOBUF*PATH=/home/wrk/Projects/lowcode/lowcode-protobuf

FEDERATION_ENABLED=true
FEDERATION*HOST=node*destination:8084
FEDERATION_LABEL=Federation destination host

# Sync settings
FEDERATION*SYNC*STRUCTURE*MONITOR*INTERVAL=60s
FEDERATION*SYNC*STRUCTURE*PAGE*SIZE=1
FEDERATION*SYNC*DATA*MONITOR*INTERVAL=20s
FEDERATION*SYNC*DATA*PAGE*SIZE=100

## Запуск инстансов

```bash
```
$ docker-compose up -d node_origin
$ docker-compose up -d node_destination

После запуска инстансов вы можете следить за логами docker-compose, выполнив:

.Логи для инстанса источника LowCoooode:
```bash
```
$ docker-compose logs -f node_origin

.Логи для инстанса назначения LowCoooode:
```bash
```
$ docker-compose logs -f node_destination

## Сопоставление портов

[cols="m,a,a"]
|===
|Инстанс |Локальное подключение |Подключение Docker

| node_origin
| `localhost:8084`
| `node*origin:8084` (из **node*destination**)

| node_destination
| `localhost:8085`
| `node*destination:8084` (из **node*origin**)

| db_origin
| `localhost:3306`
| `db_origin:3306` (из любого из узлов)

| db_destination
| `localhost:3307`
| `db_destination:3306` (из любого из узлов)
|===
