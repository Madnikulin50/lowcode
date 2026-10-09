# Примеры онлайн-развёртывания

:leveloffset: +1

# Многообразный MySQL
:page-noindex: true

!!! note
    *DevNote*: Опишите эту конфигурацию; сколько/какие сервисы она запускает и тому подобное.


.`docker-compose.yaml`
```yaml
```
version: '3.5'

services:
  server:
    image: lowcode/lowcode:${PAGE-VERSION}.x
    restart: always
    env_file: [ .env ]
    depends_on: [ db ]
    networks: [ proxy, internal ]
    # Uncomment to use local fs for data persistence
    volumes: [ "./data/server:/data" ]
    environment:
      # These two are needed only if you are using NginX Lets-Encrypt companion
      # (see docs.lowcode.org for details)
      # VIRTUAL_HOST helps NginX proxy route traffic for specific virtual host to this container
      VIRTUAL_HOST:     $lowcode.org
      # LETSENCRYPT_HOST helps NginX LE companion pull and configure SSL certificates for your domain
      LETSENCRYPT_HOST: $lowcode.org

  db:
    # MySQL Database
    # See https://hub.docker.com/r/percona/percona-server for details
    image: percona:8.0
    restart: always
    volumes: [ "./data/db:/var/lib/mysql" ]
    environment:
      MYSQL_DATABASE: dbname
      MYSQL_USER:     dbuser
      MYSQL_PASSWORD: dbpass
      # get the random generated password by running: docker-compose logs db | grep "GENERATED ROOT PASSWORD"
      MYSQL*RANDOM*ROOT_PASSWORD: random
    healthcheck: { test: ["CMD", "mysqladmin" ,"ping", "-h", "localhost"], timeout: 20s, retries: 10 }
    networks: [ internal ]

networks:
  internal: {}
  proxy: { external: true }


.`.env`
```env
```
########################################################################################################################
# docker-compose supports environment variable interpolation/substitution in compose configuration file
# (more info: https://docs.docker.com/compose/environment-variables)

########################################################################################################################
# General settings
DOMAIN=your-demo.example.tld
VERSION=2024.9

########################################################################################################################
# Database connection

DB*DSN=dbuser:dbpass@tcp(db:3306)/dbname?collation=utf8mb4*general_ci

########################################################################################################################
# Server settings

# Serve LowCoooode webapps alongside API
HTTP*WEBAPP*ENABLED=true

# Send action log to container logs as well
# ACTIONLOG_DEBUG=true

# Uncomment for extra debug info if something goes wrong
# LOG_LEVEL=debug

# Use nicer and colorful log instead of JSON
# LOG_DEBUG=true

########################################################################################################################
# Authentication

# Secret to use for JWT token
# Make sure you change it (>30 random characters) if
# you expose your deployment to outside traffic
# AUTH*JWT*SECRET=this-is-only-for-demo-purpose--make-sure-you-change-it-for-production

########################################################################################################################
# SMTP (mail sending) settings

# Point this to your local or external SMTP server if you want to send emails.
# In most cases, LowCoooode can detect that SMTP is disabled and skips over sending emails without an error
#SMTP_HOST=smtp-server.example.tld:587
#SMTP_USER=postmaster@smtp-server.example.tld
#SMTP_PASS=this-is-your-smtp-password
#SMTP_FROM='"Demo" <info@your-demo.example.tld>'


# Многообразный PostgreSQL
:page-noindex: true

!!! note
    *DevNote*: Опишите эту конфигурацию; сколько/какие сервисы она запускает и тому подобное.


.`docker-compose.yaml`
```yaml
```
version: '3.5'

services:
  server:
    image: lowcode/lowcode:${PAGE-VERSION}.x
    networks: [ proxy, internal ]
    restart: always
    env_file: [ .env ]
    depends_on: [ db ]
    volumes: [ "./data/server:/data" ]
    environment:
      # VIRTUAL_HOST helps NginX proxy route traffic for specific virtual host to
      # this container
      # This value is also picked up by initial boot autoconfiguration procedure
      # If this is changed, make sure you change settings accordingly
      VIRTUAL_HOST: $lowcode.org
      # This is needed only if you are using NginX Lets-Encrypt companion
      # (see docs.lowcode.org for details)
      LETSENCRYPT_HOST: $lowcode.org

  db:
    # PostgreSQL Database
    # See https://hub.docker.com/_/postgres for details
    # Support for postgres 13, 14 and 15 is available in the latest version of LowCoooode
    image: postgres:15
    networks: [ internal ]
    restart: always
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U lowcode"], interval: 10s, timeout: 5s, retries: 5 }
    environment:
      # Warning: these are values that are only used on 1st start
      #          if you want to change it later, you need to do that
      #          manually inside db container
      POSTGRES_USER:     lowcode
      POSTGRES_PASSWORD: lowcode

networks:
  internal: {}
  proxy: { external: true }


.`.env`
```env
```
########################################################################################################################
# docker-compose supports environment variable interpolation/substitution in compose configuration file
# (more info: https://docs.docker.com/compose/environment-variables)

########################################################################################################################
# General settings
DOMAIN=your-demo.example.tld
VERSION=2024.9

########################################################################################################################
# Database connection

DB_DSN=postgres://lowcode:lowcode@db:5432/lowcode?sslmode=disable

########################################################################################################################
# Server settings

# Serve LowCoooode webapps alongside API
HTTP*WEBAPP*ENABLED=true

# Send action log to container logs as well
# ACTIONLOG_DEBUG=true

# Uncomment for extra debug info if something goes wrong
# LOG_LEVEL=debug

# Use nicer and colorful log instead of JSON
# LOG_DEBUG=true

########################################################################################################################
# Authentication

# Secret to use for JWT token
# Make sure you change it (>30 random characters) if
# you expose your deployment to outside traffic
# AUTH*JWT*SECRET=this-is-only-for-demo-purpose--make-sure-you-change-it-for-production

########################################################################################################################
# SMTP (mail sending) settings

# Point this to your local or external SMTP server if you want to send emails.
# In most cases, LowCoooode can detect that SMTP is disabled and skips over sending emails without an error
#SMTP_HOST=smtp-server.example.tld:587
#SMTP_USER=postmaster@smtp-server.example.tld
#SMTP_PASS=this-is-your-smtp-password
#SMTP_FROM='"Demo" <info@your-demo.example.tld>'


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


:leveloffset: -1
