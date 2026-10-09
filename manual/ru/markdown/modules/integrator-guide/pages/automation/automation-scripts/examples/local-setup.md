# Локальная настройка Docker Compose

## Структура файлов

.Ваша структура файлов должна выглядеть так:
```
```
📁 my-lowcode
  📄 .env
  📄 docker-compose.yaml
  📁 custom-extensions

## Файлы конфигурации

.\.env:
```env
```
DOMAIN=localhost:18091
VERSION=2023.9.9

DB_DSN=postgres://lowcode:lowcode@db:5432/lowcode?sslmode=disable

HTTP*WEBAPP*ENABLED=true
ACTIONLOG_ENABLED=false

CORREDOR_ENABLED=true
CORREDOR*EXT*SEARCH_PATHS=/extensions
CORREDOR*EXEC*CSERVERS*API*HOST=server:80/api
CORREDOR*LOG*ENABLED=true
CORREDOR*LOG*PRETTY=true
CORREDOR*LOG*LEVEL=info
CORREDOR*EXEC*CSERVERS*API*BASEURL_TEMPLATE="server:80/api/{service}"
CORREDOR_ADDR=corredor:53051

.docker-compose.yaml
```yaml
```
version: '3.5'

services:
  server:
    image: lowcode/lowcode:${PAGE-VERSION}.x
    restart: always
    env_file: [ .env ]
    platform: linux/amd64
    depends_on: [ db ]
    ports: [ "127.0.0.1:18091:80" ]
    networks: [ internal ]
    volumes:
      - "./dd/server:/data"

  db:
    image: postgres:15
    restart: always
    platform: linux/amd64
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U lowcode"], interval: 10s, timeout: 5s, retries: 5 }
    volumes:
      - "dbdata:/var/lib/postgresql/data"
    networks: [ internal ]
    environment:
      POSTGRES_USER:     lowcode
      POSTGRES_PASSWORD: lowcode

  corredor:
    env_file: [ .env ]
    image: lowcode/lowcode-server-corredor:${PAGE-VERSION}.x
    restart: on-failure
    networks: [ internal ]
    volumes:
      - "./custom-extensions:/extensions"

volumes:
  dbdata:

networks: { internal: {} }
