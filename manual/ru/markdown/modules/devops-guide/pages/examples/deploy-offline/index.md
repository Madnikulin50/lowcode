# Примеры офлайн-развёртывания

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
    ports: [ "127.0.0.1:18080:80" ]

  db:
    # MySQL Database
    # See https://hub.docker.com/r/percona/percona-server for details
    image: percona:8.0
    command: --sort*buffer*size=512K
    restart: always
    environment:
      MYSQL_DATABASE: dbname
      MYSQL_USER:     dbuser
      MYSQL_PASSWORD: dbpass
      # get the random generated password by running: docker-compose logs db | grep "GENERATED ROOT PASSWORD"
      MYSQL*RANDOM*ROOT_PASSWORD: random
    healthcheck: { test: ["CMD", "mysqladmin" ,"ping", "-h", "localhost"], timeout: 20s, retries: 10 }


.`.env`
```yaml
```
########################################################################################################################
# docker-compose supports environment variable interpolation/substitution in compose configuration file
# (more info: https://docs.docker.com/compose/environment-variables)

########################################################################################################################
# General settings
DOMAIN=local.lowcode.org:18080
VERSION=2024.9

########################################################################################################################
# Database connection
DB*DSN=dbuser:dbpass@tcp(db:3306)/dbname?collation=utf8mb4*general_ci

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


:leveloffset: -1
