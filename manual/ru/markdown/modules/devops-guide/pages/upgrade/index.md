# Обновление LowCoooode

Это руководство поможет вам без труда обновить LowCoooode до последней версии!

!!! important
    Хотя существуют внутренние механизмы защиты, способные предотвратить потерю данных или другие виды аварий, обязательно сделайте резервную копию базы данных перед обновлением.
    Обратитесь к [руководству администратора по процедуре резервного копирования](modules/devops-guide/pages/maintenance/backups.md) за подробностями.


!!! important
    Если вы обновляете несколько версий одновременно; например, с 2019.12 до 2020.9; вы должны следовать всем руководствам по обновлению между этими двумя версиями.


## Рекомендуемые шаги для обновления LowCoooode

.Особенно перед обновлением до любой крупной версии рекомендуется перед обновлением продакшена либо:
- обновить среду предварительного развёртывания (staging), или
- развернуть временную среду с копией производственной базы данных и выполнить тестовое обновление в ней.

.Шаги для обновления LowCoooode, развёрнутого с помощью **docker-compose**:
1. изменение версий образов в файле `docker-compose.yaml` (или `.env`),
1. загрузка новых образов из Docker Hub с помощью `docker-compose pull`,
1. пересоздание контейнеров с помощью `docker-compose up -d`.

:leveloffset: +1


# Обновление до 2024.9.6

Вновь добавленные группы пользователей требуют изменений схемы базы данных, которые могут быть несовместимы с предыдущими версиями.
Таблица `role*members` получает новую колонку `rel*resource`, которая генерируется на основе колонки `rel_user`.
После обновления вам необходимо либо удалить колонку `rel*user`, либо снять требование о заполнении колонки `rel*user`.

## Перед обновлением

!!! caution
    Поскольку схема базы данных будет изменена, мы рекомендуем создать резервную копию на всякий случай.


Выполните следующие команды в зависимости от вашей базы данных:

### MySQL

```sql
```
ALTER TABLE role_members
    ADD COLUMN rel_resource TEXT;

ALTER TABLE role_members
    DROP PRIMARY KEY;

ALTER TABLE role_members
    ADD PRIMARY KEY (rel*resource, rel*role);

ALTER TABLE role_members
    MODIFY rel_user BIGINT UNSIGNED NULL;


### PostgreSQL

```sql
```
ALTER TABLE role_members
    ADD COLUMN rel_resource text;

ALTER TABLE role_members
    DROP CONSTRAINT role*members*pkey;

ALTER TABLE role_members
    ADD CONSTRAINT role*members*pkey
        PRIMARY KEY (rel*resource, rel*role);

ALTER TABLE role_members
    ALTER COLUMN rel_user DROP NOT NULL;


### SQL Server

```sql
```
ALTER TABLE role_members
    ADD rel_resource NVARCHAR(MAX) NULL;

ALTER TABLE role_members
    DROP CONSTRAINT role*members*pkey;

ALTER TABLE role_members
    ADD CONSTRAINT role*members*pkey
        PRIMARY KEY (rel*resource, rel*role);

ALTER TABLE role_members
    ALTER COLUMN rel_user BIGINT NULL;


## После обновления

Если колонка вам больше не требуется, вы можете безопасно её удалить.

```sql
```
ALTER TABLE role*members DROP COLUMN rel*user;


:leveloffset: -1
