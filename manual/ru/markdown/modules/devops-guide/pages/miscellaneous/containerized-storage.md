<a id="container-db"></a>
# Настройка контейнеризованной базы данных

## MySQL

Обычно мы используем форк [Percona](https://hub.docker.com/_/percona) базы данных MySQL.

.Используйте следующие команды для настройки переменных окружения; при необходимости замените значения:
```bash
```
export DOCKER_NAME=percona;
export ROOT_PWD=root;
export MYSQL_PORT=3306;

.Для настройки базы данных выполните следующие шаги:
1. `docker run --name $DOCKER*NAME -e MYSQL*ROOT*PASSWORD=$ROOT*PWD -d -p $MYSQL_PORT:3306 percona:8.0;`
1. `docker exec -it $DOCKER*NAME mysql -uroot -p$ROOT*PWD;`
1. `CREATE DATABASE lowcode;`
1. `CREATE USER 'lowcode'@'172.17.0.1' IDENTIFIED BY 'lowcode';`
1. `GRANT ALL PRIVILEGES ON lowcode.* TO 'lowcode'@'172.17.0.1';`
1. `FLUSH PRIVILEGES;`

.Используйте следующий шаблон для создания переменной [`DB*DSN`](modules/devops-guide/pages/references/configuration/server.md#*db_dsn) `.env`:
[source,.env]
DB*DSN="lowcode:lowcode@tcp(localhost:$MYSQL*PORT)/lowcode?collation=utf8mb4*general*ci"

## PostgreSQL

Обычно мы используем официальный образ [PostgreSQL](https://hub.docker.com/_/postgres).

.Используйте следующие команды для настройки переменных окружения; при необходимости замените значения:
```bash
```
export DOCKER_NAME=pgsql2;
export ROOT_PWD=root;
export PGSQL_PORT=5432;

.Для настройки базы данных выполните следующие шаги:
1. `docker run --name $DOCKER*NAME -e POSTGRES*PASSWORD=$ROOT*PWD -d -p $PGSQL*PORT:5432 postgres:13;`
1. `docker exec -it $DOCKER_NAME psql -U postgres;`
1. `CREATE DATABASE lowcode;`
1. `CREATE USER lowcode WITH PASSWORD 'lowcode';`
1. `GRANT ALL PRIVILEGES ON DATABASE lowcode TO lowcode;`

.Используйте следующий шаблон для создания переменной [`DB*DSN`](modules/devops-guide/pages/references/configuration/server.md#*db_dsn) `.env`:
[source,.env]
DB*DSN="postgres://lowcode:lowcode@localhost:$PGSQL*PORT/lowcode?sslmode=disable"

## SQLServer

!!! note
    Это черновик, чтобы эти инструкции не потерялись; будет доработано позже.


Используйте один из следующих образов:

- `SQLServer 2022`: mcr.microsoft.com/mssql/server:2022-latest
- `SQLServer 2017`: mcr.microsoft.com/mssql/server:2017-latest
- `SQLServer 2019`: mcr.microsoft.com/mssql/server:2019-CU10-ubuntu-20.04


.Пример docker-compose для всех трёх образов:
```yaml
```
  mssql_2022:
    image: mcr.microsoft.com/mssql/server:2022-latest
    restart: on-failure
    ports: [ "127.0.0.1:1433:1433" ]
    volumes:
      - "dbdata*mssql*2022:/var/opt/mssql"
    environment:
        - ACCEPT_EULA=Y
        - SA_PASSWORD=SUPERSECRET123

  mssql_2017:
    image: mcr.microsoft.com/mssql/server:2017-latest
    restart: on-failure
    ports: [ "127.0.0.1:1433:1433" ]
    volumes:
      - "dbdata*mssql*2017:/var/opt/mssql"
    environment:
        - ACCEPT_EULA=Y
        - SA_PASSWORD=SUPERSECRET123

  mssql_2019:
    image: mcr.microsoft.com/mssql/server:2019-CU10-ubuntu-20.04
    restart: on-failure
    ports: [ "127.0.0.1:1433:1433" ]
    volumes:
      - "dbdata*mssql*2019:/var/opt/mssql"
    environment:
        - ACCEPT_EULA=Y
        - SA_PASSWORD=SUPERSECRET123

.Для настройки базы данных выполните следующие шаги:
1. `docker compose exec -it mssql_2022 sh`
1. /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "SUPERSECRET123"
1. `CREATE DATABASE lowcode;`
1. `GO`

Пример DSN базы данных: `sqlserver://sa:SUPERSECRET123@localhost:1433?database=lowcode`
