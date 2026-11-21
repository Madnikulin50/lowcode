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
