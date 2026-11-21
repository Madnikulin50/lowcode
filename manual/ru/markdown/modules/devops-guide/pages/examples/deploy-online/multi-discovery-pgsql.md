# Многообразный Discovery с PostgreSQL
:page-noindex: true

!!! note
    В настоящее время LowCoooode Discovery тестируется на работающих производственных серверах в сочетании с базой данных PostgreSQL, но должен работать с последними версиями MySQL.
    Помимо контейнеров `server` и `db`, должны также работать индексатор (`opensearch-node`) и искатель (`discovery`).


!!! caution
    Если ваш Discovery не работает, попробуйте перезапустить сервис `discovery`.
    Существует известная проблема с порядком выполнения и отчётами о проверке работоспособности, из-за которой сервис может у вас не работать.


## Файлы конфигурации

.`docker-compose.yaml`
```yaml
```
version: '3.5'

services:
  server:
    image: lowcode/lowcode:${PAGE-VERSION}.x
    env_file: [ .env ]
    networks: [ proxy, internal ]
    environment:
      VIRTUAL_HOST:     $lowcode.org
      LETSENCRYPT_HOST: $lowcode.org
    volumes:
      - data:/data
    restart: on-failure

  db:
    image: postgres:13
    networks: [ internal ]
    restart: on-failure
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U lowcode"], interval: 10s, timeout: 5s, retries: 5 }
    environment:
      POSTGRES_USER:     lowcode
      POSTGRES_PASSWORD: lowcode

  discovery:
    image: lowcode/lowcode-server-discovery:${PAGE-VERSION}.x
    env_file: [ .env ]
    restart: always
    depends_on:
      - opensearch-node
      - server
    networks:
      - proxy
      - internal
    environment:
      VIRTUAL*HOST:     ${DOMAIN*DISCOVERY}
      LETSENCRYPT*HOST: ${DOMAIN*DISCOVERY}
      ES_ADDRESS: "https://opensearch-node:9200"
      ES_USERNAME: "admin"
      ES_PASSWORD:  "supersecurepassword75@!1A"
      ES_SECURE: "false"
      ES*INDEX*INTERVAL: 60
      LOWCODE*SERVER*BASE_URL: "https://$lowcode.org"
      LOWCODE*SERVER*AUTH_URL: "https://$lowcode.org/auth"
    ports:
      - "8888:80"

  opensearch-node:
    image: opensearchproject/opensearch:latest
    networks:
      - internal
    ports:
      - "9200:9200"
      - "9600:9600"
    environment:
      - discovery.type=single-node
      - plugins.security.ssl.http.enabled=true
      - plugins.security.ssl.transport.enabled=true
      - plugins.security.allow*default*init_securityindex=true
      - plugins.security.disabled=false
      - OPENSEARCH*INITIAL*ADMIN_PASSWORD=supersecurepassword75@!1A

volumes:
  data:

networks:
  internal:
  proxy:
    external: true


.`.env`
```env
```
########################################################################################################################
# General settings

DOMAIN=lowcode.mydomain.org
VERSION=2024.9.3

DB_DSN=postgres://lowcode:lowcode@db:5432/lowcode?sslmode=disable

########################################################################################################################
# Server settings

HTTP*WEBAPP*ENABLED=true
HTTP*WEBAPP*LIST="compose,admin,workflow,reporter,discovery"
AUTH*JWT*SECRET=supersecurejwtsecret
LOG_LEVEL=debug

########################################################################################################################
# Discovery settings

DISCOVERY_ENABLED="true"
DISCOVERY*BASE*URL="https://lowcode-discovery.mydomain.org"
DISCOVERY*INDEXER*ENABLED=true
DISCOVERY*SEARCHER*ENABLED=true
DISCOVERY*SEARCHER*JWT_SECRET=supersecurejwtsecret

DOMAIN_DISCOVERY=lowcode-discovery.mydomain.org
DISCOVERY*LOWCODE*DOMAIN=https://lowcode.mydomain.org

DISCOVERY*INDEXER*PRIVATE*INDEX*CLIENT_KEY="111111111111111111"
DISCOVERY*INDEXER*PRIVATE*INDEX*CLIENT_SECRET="supersecretsupersecretsupersecret"
DISCOVERY*INDEXER*PROTECTED*INDEX*CLIENT_KEY="111111111111111111"
DISCOVERY*INDEXER*PROTECTED*INDEX*CLIENT_SECRET="supersecretsupersecretsupersecret"
DISCOVERY*INDEXER*PUBLIC*INDEX*CLIENT_KEY="111111111111111111"
DISCOVERY*INDEXER*PUBLIC*INDEX*CLIENT_SECRET="supersecretsupersecretsupersecret"
DISCOVERY*SEARCHER*CLIENT_KEY="111111111111111111"
DISCOVERY*SEARCHER*CLIENT_SECRET="supersecretsupersecretsupersecret"


<a id="integration-configuration"></a>


## Подготовка LowCoooode

LowCoooode Discovery — это отдельное приложение, независимое от остальной части системы LowCoooode.
Чтобы сделать Discovery функциональным, вы должны **предоставить доступ**, создав **клиента аутентификации** вместе с **пользователем** и **ролью**.

!!! important
    Контроль доступа определяет, к каким данным имеет доступ индексатор Discovery.


### Роль индексатора

Сначала определите новую роль для использования индексатором Discovery.
Откройте веб-приложение LowCoooode Admin и перейдите в меню Система[Роли].
Нажмите кнопку **Создать**, заполните параметры и нажмите кнопку **Отправить**.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/role.png",
    "alias": "discovery-role",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 312,
    "y": 102,
    "w": 1313,
    "h": 544
  }
}

<a id="indexer-user"></a>
### Пользователь индексатора

Далее определите нового пользователя, за которого должен себя выдавать индексатор Discovery.
В веб-приложении LowCoooode Admin перейдите в меню Система[Пользователи].
Нажмите кнопку **Создать**, заполните параметры и нажмите кнопку **Отправить**.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/user.png",
    "alias": "discovery-user",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 360,
    "y": 105,
    "w": 1233,
    "h": 521
  },
  "annotations": []
}

После сохранения пользователя назначьте ему роль, которую вы создали ранее.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/user-membership.png",
    "alias": "discovery-user-membership",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 372,
    "y": 319,
    "w": 1156,
    "h": 245
  },
  "annotations": []
}

### Клиент аутентификации индексатора

Наконец, определите клиента аутентификации, который индексатор Discovery должен использовать для аутентификации в LowCoooode.

!!! note
    Поскольку аутентификация выполняется между двумя системами, вы должны использовать тип предоставления `client_credentials`.


В веб-приложении LowCoooode Admin перейдите в меню Система[Клиенты аутентификации].
Нажмите кнопку **Создать** и заполните основные параметры; убедитесь, что вы выбрали тип предоставления `client_credentials` и отметили «allow client access to LowCoooode Discovery API on behalf of user».
Выберите ранее созданного пользователя в поле ввода `impersonate user` и нажмите кнопку **Отправить**.

!!! note
    Рекомендуется определять новый клиент аутентификации для нового внешнего приложения вместо повторного использования существующих.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/auth-client.png",
    "alias": "discovery-auth-client",
    "w": 1916,
    "h": 985
  },
  "view": {
    "x": 527,
    "y": 104,
    "w": 1151,
    "h": 732
  },
  "annotations": []
}
