# Одиночный образ SQLite в памяти
:page-noindex: true

!!! note
    *DevNote*: Опишите эту конфигурацию; сколько/какие сервисы она запускает и тому подобное.


.`docker-compose.yaml`
```yaml
```
version: '3.5'

services:
  server:
    image: lowcode/lowcode:2023.3
    restart: always
    # Direct your browser to http://localhost:18080
    ports: [ "127.0.0.1:18080:80" ]
    environment:
      # Used for cookies, links, ...
      DOMAIN: "local.lowcode.org:18080"

      # Serve web applications directly from the server container
      HTTP*WEBAPP*ENABLED: "true"

      # Disable action log to minimize db writes
      ACTIONLOG_ENABLED: "false"
