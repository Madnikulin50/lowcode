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
