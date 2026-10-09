# Руководство DevOps

Руководство DevOps описывает процесс настройки, конфигурирования и обслуживания вашего инстанса LowCoooode.

Эта страница содержит простые инструкции по настройке инстанса, которые подойдут для большинства случаев.
Более сложные сценарии использования и продвинутые читатели должны обращаться к дополнительным подстраницам руководства DevOps.

!!! important
    LowCoooode скомпилирован, разработан и протестирован для современных браузеров.
    Если вам нужно поддерживать старые браузеры, такие как InternetExplorer, вам необходимо собрать свои образы.
    Обратитесь к [System Requirements](modules/devops-guide/pages/system-requirements.md) за подробностями.


## Предварительные требования

<a id="docker"></a>
### Docker

Чтобы использовать наши предварительно собранные Docker-образы, на каждой системе, где вы планируете развернуть LowCoooode, должен быть установлен и запущен Docker.
Вы можете следовать [официальной документации](https://docs.docker.com/get-docker/), чтобы установить его.

Кроме того, вы можете скачать предварительно собранные исходники со [страницы релизов](https://releases.lowcode.org/files/) или собрать свои собственные.

<a id="docker-compose"></a>
### Docker Compose

Docker Compose облегчает жизнь при запуске нескольких Docker-образов, каждый из которых может быть произвольно настроен.
Вы можете следовать [официальной документации](https://docs.docker.com/compose/install/), чтобы установить его.

```yaml
```
!!! note
    При использовании файлов `.env` вам необходимо явно указать это внутри файла `docker-compose.yaml`, например:
    
    version: '3.5'
    
    services:
      server:
        image: lowcode/lowcode:${PAGE-VERSION}.x
        restart: always
        env_file: [ .env ] # <- see here
        depends_on: [ db ]
        ports: [ "127.0.0.1:18084:80" ]
    
      # ...
    ----


## Обзор архитектуры и репозиториев

### Архитектура LowCoooode

.Диаграмма описывает различные части LowCoooode, их взаимосвязь и взаимодействие.
[plantuml,data-sync-origin,svg,role=component data-zoomable]
@startuml
component "Public Network" #E5412215 {
  package "Web Applications" {
    [LowCoooode Compose] as webappCompose
    [LowCoooode Workflow] as webappWorkflow
    [LowCoooode Admin] as webappAdmin
    [LowCoooode Reporter] as webappReporter
    [LowCoooode Discovery] as webappDiscovery

    interface "REST API" as restW

    webappCompose -down- restW
    webappWorkflow -down- restW
    webappAdmin -down- restW
    webappReporter -down- restW
    webappDiscovery -down- restW

    interface "HTTP" as httpW

    webappCompose -down- httpW
    webappWorkflow -down- httpW
    webappAdmin -down- httpW
    webappReporter -down- httpW
    webappDiscovery -down- httpW
  }
}

component "Private Network" {
  package "LowCoooode Server" {
    interface "REST API" as rest
    restW -down- rest

    interface "HTTP" as authHTTP
    httpW -down- authHTTP

    [LowCoooode Services] as svc
    rest -down- svc

    [LowCoooode Auth] as auth
    auth -right- svc
    authHTTP -down- auth
  }

  package "LowCoooode Discovery Server" {
    interface "REST API" as restD
    svc -down- restD

    interface "REST API" as restS
    webappDiscovery -down- restS

    [Indexer] as Indexer
    restD -down- Indexer

    [Searcher] as Searcher
    restS -down- Searcher
  }

  package "LowCoooode Corredor Server" {
    interface "GRPC" as grpc
    [Automation Runner] as corredorRunner

    svc -left- grpc
    grpc -left- corredorRunner
  }


  package "Auxiliary Services" as aux {
    component db [
      Database
      ....
      MySQL, PostgreSQL,
      ElasticSearch,
      OpenSearch, ...
    ]
    component mta [
      MTA
      ....
      Local, remote
      SMTP server
    ]
    component objs [
      Object Storage
      ....
      Local disc,
      Min.io
    ]
    component logs [
      Log Storage
      ....
      STDOUT,
      ElasticSearch
    ]
    component errt [
      Error Tracking
      ....
      STDOUT,
      Sentry
    ]
  }

  svc -down- db
  svc -down- mta
  svc -down- objs
  svc -down- logs
  svc -down- errt

  Indexer -down- db
  Searcher -down- db
}
@enduml

Веб-приложения общаются с сервером через REST API и аутентифицируются через сервер аутентификации.

LowCoooode Discovery дополнительно общается с `lowcode-server-discovery`.

Сервер взаимодействует со всеми внутренними вспомогательными сервисами, такими как база данных, объектное хранилище, отслеживание ошибок и журналов, исполнители автоматизации, ...

.Некоторые вспомогательные сервисы включают:
- федерацию данных,
- хранение данных,
- журналирование,
- отправку электронной почты.

### Репозитории LowCoooode и их взаимосвязь

.Диаграмма описывает репозитории LowCoooode и их взаимосвязь в контексте пайплайна сборки.
![role="data-zoomable"](developer-guide:build-pipelines.png)

.Основные репозитории LowCoooode:
1. {GIT*REPO*LINK_PREFIX}[`lowcode`]: Монорепозиторий, содержащий кодовую базу основных функций LowCoooode.
1. [`lowcode-server-corredor`](https://github.com/{GIT*REPO*GROUP}/{GIT*REPO*PREFIX}-server-corredor): Исполнитель автоматизации Corredor.
1. [`lowcode-docs`](https://github.com/{GIT*REPO*GROUP}/{GIT*REPO*PREFIX}-docs): Документация LowCoooode.

## Файлы конфигурации системы

LowCoooode настраивается через файл окружения (`.env`).
Он позволяет быстро развернуть и настроить поведение LowCoooode на другой системе.

Наши примеры настроены так, чтобы работать как есть, но не стесняйтесь настраивать [сервер](modules/devops-guide/pages/references/configuration/server.md) и [сервер Corredor](modules/devops-guide/pages/references/configuration/corredor.md) по своему усмотрению.

!!! note
    Файл `.env` находится в корне папки проекта.
    В контексте Docker Compose он находится рядом с файлом `docker-compose.yaml`.


Файл `.env` выполняет **неявную конфигурацию Docker Compose**, **подстановку переменных для Docker-конфигураций** и **конфигурации сервисов**.

Вы можете использовать переменные, определённые в файле `.env`, внутри ваших файлов `docker-compose.yaml`, используя `$\{VARIABLE_HERE}`.

<a id="deploy-offline"></a>
## Офлайн-развёртывание

Офлайн-развёртывания запускают все сервисы в одной сети, где порты привязаны к сети хоста.

!!! important
    Офлайн-развёртывания подходят только для локальной разработки и демонстраций (сред, недоступных извне).


В этом разделе приведена минимальная конфигурация с PostgreSQL в качестве постоянного хранилища базы данных.
См. [Deploy Offline](modules/devops-guide/pages/examples/deploy-offline/index.md) для примеров конфигураций офлайн-развёртывания.

### Настройка структуры файлов

.Ваша структура файлов должна выглядеть так:
```
```
📁 my-lowcode
  📄 .env
  📄 docker-compose.yaml
  📁 data <1>
    📁 server <2>
    📁 db <3>
<1> Обязательно измените владельца на Docker-контейнер (вы можете использовать `chown 1001:1001 data/db` и `chown 4242:4242 data/server`).
Пропустите, если вы не будете использовать постоянное хранилище.
<2> Здесь хранятся все данные сервера, например, загруженные вложения.
<3> Здесь хранятся данные базы данных.

### Настройка docker-compose.yaml

.Ваша конфигурация должна выглядеть так:
```yaml
```
version: '3.5'

services:
  server:
    image: lowcode/lowcode:${PAGE-VERSION}.x
    restart: always
    env_file: [ .env ]
    depends_on: [ db ]
    ports: [ "127.0.0.1:18080:80" ]

  db:
    # PostgreSQL Database
    # See https://hub.docker.com/_/postgres for details
    # Support for postgres 13, 14 and 15 is available in the latest version of LowCoooode
    image: postgres:15
      #    networks: [ internal ]
    restart: always
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U lowcode"], interval: 10s, timeout: 5s, retries: 5 }
    volumes:
      - "dbdata:/var/lib/postgresql/data"
    environment:
      # Warning: these are values that are only used on 1st start
      #          if you want to change it later, you need to do that
      #          manually inside db container
      POSTGRES_USER:     lowcode
      POSTGRES_PASSWORD: lowcode

volumes:
  dbdata:


!!! note
    Поддерживаемые версии PostgreSQL 13, 14, 15


### Настройка .env

.Ваша конфигурация должна выглядеть так:
[source,.env]
########################################################################################################################
# docker-compose supports environment variable interpolation/substitution in compose configuration file
# (more info: https://docs.docker.com/compose/environment-variables)

########################################################################################################################
# General settings
DOMAIN=localhost:18080
VERSION=2024.9

########################################################################################################################
# Database connection
DB_DSN=postgres://lowcode:lowcode@db:5432/lowcode?sslmode=disable

########################################################################################################################
# Server settings

# Running all-in-one and serving web applications directly from server container
HTTP*WEBAPP*ENABLED=true

# Disabled, we do not need detailed persistent logging of actions in local env
ACTIONLOG_ENABLED=false

########################################################################################################################
# SMTP (mail sending) settings

# Point this to your local or external SMTP server if you want to send emails.
# In most cases, LowCoooode can detect that SMTP is disabled and skips over sending emails without an error
#SMTP_HOST=smtp-server.example.tld:587
#SMTP_USER=postmaster@smtp-server.example.tld
#SMTP_PASS=this-is-your-smtp-password
#SMTP_FROM='"Demo" <info@your-demo.example.tld>'


### Запуск сервисов

В корне вашего проекта (рядом с файлами `docker-compose.yaml` и `.env`) выполните docker compose (выполнение может занять несколько секунд).
Команда (загружает и) запускает все сервисы, настроенные в вашем файле `docker-compose.yaml`.

```bash
```
docker-compose up -d

Проверьте, всё ли запустилось корректно, выполнив `docker-compose ps`.
Вывод должен выглядеть так:

```
```
      Name                    Command                   State                      Ports
-------------------------------------------------------------------------------------------------------
demo*pgsql*db_1         /docker-entrypoint.sh psql       Up (healthy)     5432/tcp
demo*pgsql*server_1     bin/server serve-api             Up (healthy)     127.0.0.1:18080->80/tcp


См. [Troubleshooting](modules/devops-guide/pages/troubleshooting/index.md), если что-то пошло не так или не запустилось.

### Проверка развёртывания

1. Направьте ваш браузер на http://localhost:18080 (измените порт, если вы использовали другой порт).
Вы должны быть перенаправлены на страницу аутентификации (`/auth`).
1. Создайте свою учётную запись через форму регистрации (первая созданная учётная запись по умолчанию является администратором).
1. Проверьте версию сервера http://localhost:18080/version.
1. Проверьте работоспособность сервера http://localhost:18080/healthcheck.
1. Проверьте документацию API http://localhost:18080/api/docs/.

!!! note
    Если вы не настроили параметры SMTP, все регистрации помечаются как подтверждённые.


<a id="deploy-online"></a>
## Онлайн-развёртывание

!!! note
    Если вы используете Nginx и WebSocket-подключение не устанавливается, обратитесь к [menu:Обслуживание[Устранение неполадок](modules/devops-guide/pages/troubleshooting/index.md#ws-nginx-connection-fail)].


Онлайн-конфигурации разделяют ваши сервисы на две сети: внутреннюю и прокси.
Внутренняя сеть скрывает большую часть системы от интернета.
В этом разделе приведена минимальная конфигурация с PostgreSQL в качестве постоянного хранилища базы данных и включённым сервером Corredor.

!!! tip
    Вы можете использовать те же шаги для настройки нескольких онлайн-развёртываний, например, staging- и производственной среды.


См. [Deploy Online](modules/devops-guide/pages/examples/deploy-online/index.md) для примеров конфигураций онлайн-развёртывания.

### Настройка структуры файлов

.Ваша структура файлов должна выглядеть так:
```
```
📁 my-proxy <1>
  📄 docker-compose.yaml
  📄 custom.conf <2>
📁 my-lowcode
  📄 .env
  📄 docker-compose.yaml
  📁 data <3>
    📁 server <4>
    📁 db <5>
<1> Пропустите это, если вы не планируете использовать обратный прокси Nginx или если он уже настроен.
<2> `custom.conf` должен находиться рядом с файлом `docker-compose.yaml`.
<3> Обязательно измените владельца на Docker-контейнер (вы можете использовать `chown 1001:1001 data/db` и `chown 4242:4242 data/server`).
Пропустите, если вы не будете использовать постоянное хранилище.
<4> Здесь хранятся все данные сервера, например, загруженные вложения.
<5> Здесь база данных может хранить данные.

### Настройка вашего обратного прокси Nginx

⚠️ Мы находимся в каталоге `my-proxy`.

Эта часть автоматизирует создание и продление TLS-сертификатов Let's Encrypt, пересылку трафика в Docker-контейнеры и упрощает сложные конфигурации межсетевого экрана.

!!! caution
    Следующие инструкции предполагают, что в вашей текущей среде нет ничего похожего.
    
    Если вы используете другие средства пересылки трафика или обработки SSL-сертификатов, действуйте с осторожностью.


Мы будем использовать [Nginx Proxy](https://github.com/jwilder/nginx-proxy) и [LetsEncrypt Nginx Proxy Companion](https://github.com/JrCs/docker-letsencrypt-nginx-proxy-companion).
Если вы хотите использовать или уже используете что-то другое, можете смело пропустить этот раздел.


.Ваш `docker-compose.yaml` должен выглядеть так:
```yaml
```
version: '3.5'

services:
  nginx-proxy:
    image: nginxproxy/nginx-proxy
    container_name: nginx-proxy
    restart: always
    networks:
      - proxy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./certs:/etc/nginx/certs
      - ./htpasswd:/etc/nginx/htpasswd
      - ./vhost.d:/etc/nginx/vhost.d
      - ./html:/usr/share/nginx/html
      - ./custom.conf:/etc/nginx/conf.d/custom.conf:ro
      - /var/run/docker.sock:/tmp/docker.sock:ro

  nginx-letsencrypt:
    image: nginxproxy/acme-companion
    container_name: nginx-letsencrypt
    restart: always
    depends_on:
      - nginx-proxy
    volumes_from:
      - nginx-proxy:rw
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - acme:/etc/acme.sh

# Create network if it does not exist
networks:
  proxy:
    external: true

volumes:
  acme:

!!! note
    Поддерживаемые версии PostgreSQL 13, 14, 15


.Ваш `custom.conf` должен выглядеть так:
```yaml
```
client*max*body_size 1g;
# allows nginx to support larger file uploads

proxy*read*timeout 86400s;
# support longer running requests as opposed to the default 60s

client*header*buffer_size 64k;
# increase the size of the client header buffer to support larger headers, as sometimes cookie data can get large

large*client*header_buffers 4 64k;
# increase the number of buffers to support larger headers

proxy*ignore*client_abort  on;
# close connection immidiately when client aborts

proxy*buffer*size          128k;
# increase the size of the buffer used for reading the response from the proxied server

proxy_buffers              4 256k;
# size of the buffers for a single connection

proxy*busy*buffers_size    256k;
# size of buffers during sending to proxied server and buffering the response


Внутри вашего каталога `my-proxy` выполните `docker-compose up -d` (выполнение может занять несколько секунд), чтобы запустить обратный прокси.

Проверьте, всё ли запустилось корректно, выполнив `docker-compose ps`.
Вывод должен выглядеть так:

```
```
      Name                     Command               State                    Ports
-----------------------------------------------------------------------------------------------------
nginx-letsencrypt   /bin/bash /app/entrypoint. ...   Up
nginx-proxy         /app/docker-entrypoint.sh  ...   Up      0.0.0.0:443->443/tcp, 0.0.0.0:80->80/tcp


### Настройка docker-compose.yaml

⚠️ Мы находимся в каталоге `my-lowcode`.

!!! important
    Контейнеры *должны* находиться в той же сети, что и `nginx-proxy` (в примерах мы используем сеть с именем `proxy`).


.Ваша конфигурация должна выглядеть так:
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


### Настройка .env

⚠️ Мы находимся в каталоге `my-lowcode`.

.Ваша конфигурация должна выглядеть так:
[source,.env]
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


### Запуск сервисов

Внутри вашего каталога `my-lowcode` (рядом с файлами `docker-compose.yaml` и `.env`) выполните docker compose (выполнение может занять несколько секунд).
Команда (загружает и) запускает все сервисы, настроенные в вашем файле `docker-compose.yaml`.

```bash
```
docker-compose up -d

Проверьте, всё ли запустилось корректно, выполнив `docker-compose ps`.
Вывод должен выглядеть так:

```
```
        Name                                Command                  State              Ports
----------------------------------------------------------------------------------------------------
my*production*demo*db*1         docker-entrypoint.sh postgres    Up (healthy)   3306/tcp, 33060/tcp
my*production*demo*server*1     /bin/lowcode-server serve-api    Up (healthy)   80/tcp


См. [Troubleshooting](modules/devops-guide/pages/troubleshooting/index.md), если что-то пошло не так или не запустилось.

### Проверка развёртывания

1. Направьте ваш браузер на `http://your-demo.example.tld`.
Вы будете перенаправлены на страницу аутентификации (`/auth`).
1. Создайте свою учётную запись через форму регистрации (первая созданная учётная запись по умолчанию является администратором).
1. Проверьте версию сервера http://your-demo.example.tld/version
1. Проверьте работоспособность сервера http://your-demo.example.tld/healthcheck
1. Проверьте документацию API http://your-demo.example.tld/api/docs/

<a id="useful-commands"></a>
## Полезные команды

.Список полезных команд Docker:
[cols="2s,2m"]
|===
| [#docker-exec]#[docker-exec,Выполнение команд внутри контейнера (запущенного)](#docker-exec,Выполнение команд внутри контейнера (запущенного))#
| docker exec -it <container name> help

| [#docker-exec-stopped]#[docker-exec-stopped,Выполнение команд внутри контейнера (не запущенного)](#docker-exec-stopped,Выполнение команд внутри контейнера (не запущенного))#
| docker run -it --rm <container name> help
|===

.Список полезных команд Docker Compose:
[cols="2s,2m"]
|===
| [#docker-compose-stop-rm]#[docker-compose-stop-rm,Остановить и удалить контейнеры вместе с их томами без подтверждения](#docker-compose-stop-rm,Остановить и удалить контейнеры вместе с их томами без подтверждения)#
| docker-compose rm --force --stop -v

| [#docker-compose-logs-full]#[docker-compose-logs-full,Просмотр журналов всех запущенных контейнеров](#docker-compose-logs-full,Просмотр журналов всех запущенных контейнеров)#
| docker-compose logs --follow --tail 20

| [#docker-compose-logs-specific]#[docker-compose-logs-specific,Просмотр журналов конкретного контейнера](#docker-compose-logs-specific,Просмотр журналов конкретного контейнера)#
| docker-compose logs --follow --tail 20 <service name>

| [#docker-compose-exec]#[docker-compose-exec,Выполнение с Docker Compose](#docker-compose-exec,Выполнение с Docker Compose)#
| docker-compose exec <service name> help

| [#docker-compose-bash]#[docker-compose-bash,Доступ к bash в контейнере в `WORKDIR`](#docker-compose-bash,Доступ к bash в контейнере в `WORKDIR`)#
| docker-compose exec <service name> bash
|===

## Куда дальше

.Подразделы в разделе меню DevOps guide более подробно рассматривают конкретные темы, например:
- [выполнение резервного копирования](modules/devops-guide/pages/maintenance/backups.md),
- [устранение неполадок](modules/devops-guide/pages/troubleshooting/index.md),
- открытие дополнительных возможностей автоматизации с помощью [Email Relay](modules/devops-guide/pages/email-relay.md), [Sink Route](modules/devops-guide/pages/sink-route.md), [Pdf Renderer](modules/devops-guide/pages/pdf-renderer.md) и
- [I18N](modules/devops-guide/pages/i18n/index.md).

.Для разработки собственных приложений Low Code обратитесь к [Integrator Guide](modules/integrator-guide/pages/index.md):
- [Authentication](modules/integrator-guide/pages/authentication/index.md) и [Security Model](modules/integrator-guide/pages/security-model/index.md),
- [интернационализация](modules/integrator-guide/pages/i18n/index.md),
- [взаимодействие через REST API](modules/integrator-guide/pages/accessing-lowcode/index.md),
- [конфигурация Low Code](modules/integrator-guide/pages/compose-configuration/index.md) и
- [Automation](modules/integrator-guide/pages/automation/index.md).
