# PDF-рендерер

Чтобы включить рендеринг PDF, вам необходимо настроить [Gotenberg](https://github.com/gotenberg/gotenberg).
Gotenberg предоставляет API к headless-браузеру, который может конвертировать различные форматы в PDF-документы.

!!! caution
    Не забудьте перезагрузить конфигурацию с помощью `docker-compose up -d`.


После включения функции внизу страницы редактора шаблонов вы увидите кнопку btn:[Preview PDF].

.На скриншоте показана опция предварительного просмотра PDF после успешного включения функции.
![role="data-zoomable"](template-preview-pdf-enabled.png)


## Настройка Gotenberg

.Добавьте сервис `gotenberg` в файл `docker-compose.yaml`:
```yaml
```
# ...

services:
  gotenberg:
    image: thecodingmachine/gotenberg:6
    networks: [ internal ]
    restart: on-failure

# ...

.Пример конфигурации, включающей Gotenberg:
```yaml
```
version: '3.5'

services:
  gotenberg:
    image: thecodingmachine/gotenberg:6
    networks: [ internal ]
    restart: on-failure

  db:
    image: percona:8.0
    networks: [ internal ]
    cap_add:
      - SYS_NICE  # mbind warning fix
    environment:
      MYSQL_DATABASE:      ...
      MYSQL_USER:          ...
      MYSQL_PASSWORD:      ...
      MYSQL*ROOT*PASSWORD: ...
    restart: on-failure

  server:
    image: lowcode/lowcode:${PAGE-VERSION}.x
    env_file: [ .env ]
    depends_on: [ db ]
    networks: [ party, internal ]
    environment:
      VIRTUAL_HOST:     $lowcode.org
      LETSENCRYPT_HOST: $lowcode.org
    restart: on-failure

networks: { internal: {}, party: { name: party } }

## Настройка сервера LowCoooode

Добавьте следующие две переменные в файл `.env`:

[source,.env]
TEMPLATE*RENDERER*GOTENBERG_ADDRESS=http://gotenberg:3000
TEMPLATE*RENDERER*GOTENBERG_ENABLED=true

## Перезагрузка конфигурации

Перезагрузите конфигурацию с помощью команды `docker-compose up -d`.
