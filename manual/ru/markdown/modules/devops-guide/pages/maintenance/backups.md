# Резервные копии

Мы рекомендуем регулярно делать резервные копии вашей базы данных и загруженных файлов.
Также настоятельно рекомендуется сделать резервную копию перед обновлением до более новой версии LowCoooode.

!!! tip
    Вы можете определить cron-задачу, которая будет сохранять ваши данные во внешнее хранилище.


!!! note
    *DevNote* добавить несколько примеров конфигураций, как сейчас с CRON-задачей?


<a id="reduce-size"></a>
## Уменьшение размера резервной копии

В некоторых случаях вы можете исключить определённые таблицы базы данных, что позволит уменьшить размер резервной копии.
Это уменьшение также пригодится, если вы хотите перенести производственную базу данных на ваш локальный инстанс.

[cols="2s,5a"]
|===
| [#backup-optional*table-auth*sessions]#[backup-optional*table-auth*sessions,auth*sessions](#backup-optional*table-auth*sessions,auth*sessions)#
| Таблица `auth_sessions` хранит сессии аутентификации пользователя.
Если вы исключите эту таблицу, пользователям придётся заново входить в систему после восстановления базы данных.
Таблицу `auth_sessions` можно безопасно исключать в любом случае.

| [#backup-optional*table-credentials]#[backup-optional*table-credentials,credentials](#backup-optional_table-credentials,credentials)#
| Таблица `credentials` хранит секреты, связанные с аутентификацией, такие как пароли, токены сброса и токены подтверждения электронной почты.
Если вы исключите таблицу `credentials`, все выданные учётные данные потребуется выпустить заново, что потребует от пользователей сброса паролей, повторной отправки писем для сброса пароля и повторной отправки писем подтверждения электронной почты.
Таблицу `credentials` можно безопасно исключать в любом случае, хотя в некоторых случаях её исключение может быть рекомендовано.

| [#backup-optional*table-auth*oa2tokens]#[backup-optional*table-auth*oa2tokens,auth*oa2tokens](#backup-optional*table-auth*oa2tokens,auth*oa2tokens)#
| Таблица `auth_oa2tokens` хранит токены доступа, выданные сервером LowCoooode веб-приложениям.
Если вы исключите таблицу `auth_oa2tokens`, все токены доступа будут аннулированы, и их потребуется выпустить заново.
Таблицу `auth_oa2tokens` можно безопасно исключать в любом случае, хотя в некоторых случаях её исключение может быть рекомендовано.

!!! caution
    Если вы аннулируете токены доступа, все аутентифицированные веб-приложения должны будут пройти аутентификацию заново.
    Это может вызвать проблемы в случаях, когда веб-приложение перегенерирует токены только при наступлении запланированного срока их истечения.


| [#backup-optional*table-actionlog]#[backup-optional*table-actionlog,actionlog](#backup-optional_table-actionlog,actionlog)#
| Таблица `actionlog` хранит события, которые сервер LowCoooode счёл значимыми, такие как создание пользователей, регистрация клиентов аутентификации и поиск записей.
Если вы исключите таблицу `actionlog`, история действий будет потеряна.
Таблицу `actionlog` можно безопасно исключать для случаев, связанных с разработкой.

| [#backup-optional*table-automation*sessions]#[backup-optional*table-automation*sessions,automation*sessions](#backup-optional*table-automation*sessions,automation*sessions)#
| Таблица `automation_sessions` хранит метаданные о выполнении рабочих процессов, такие как шаг, на котором находится рабочий процесс, какие параметры были переданы рабочему процессу и результат выполнения.
Если вы исключите таблицу `automation_sessions`, история выполнения рабочих процессов будет потеряна, а любые ожидающие ввода или приостановленные рабочие процессы не будут завершены.
Таблицу `automation_sessions` можно безопасно исключать для случаев, связанных с разработкой.

| [#backup-optional*table-compose*record]#[backup-optional*table-compose*record,compose*record](#backup-optional*table-compose*record,compose*record)#
| Таблица `compose_record` хранит метаданные записей, созданных в ваших приложениях LowCoooode Low Code.
Если вы исключите таблицу `compose_record`, все записи, созданные для ваших приложений Low Code, будут потеряны.
Таблицу `compose_record` можно исключить, если вы хотите сохранить структуру системы, но исключить все данные.

| [#backup-optional*table-compose*record*value]#[backup-optional*table-compose*record*value,compose*record*value](#backup-optional*table-compose*record*value,compose*record_value)#
| Таблица `compose*record*value` хранит значения записей, созданных в ваших приложениях LowCoooode Low Code.
Если вы исключите таблицу `compose*record*value`, все записи, созданные для ваших приложений Low Code, будут потеряны.
Таблицу `compose*record*value` можно исключить, если вы хотите сохранить структуру системы, но исключить все данные.

| [#backup-optional*table-queue*messages]#[backup-optional*table-queue*messages,queue*messages](#backup-optional*table-queue*messages,queue*messages)#
| Таблица `queue_messages` хранит сообщения, переданные в очередь сообщений.
Если вы исключите таблицу `queue_messages`, все сообщения, переданные в очередь сообщений, будут потеряны.
Таблицу `queue_messages` можно безопасно удалить в большинстве случаев.

| [#backup-optional*table-resource*activity*log]#[backup-optional*table-resource*activity*log,resource*activity*log](#backup-optional*table-resource*activity*log,resource*activity_log)#
| Таблица `resource*activity*log` хранит историю изменений ресурсов, таких как изменения значений записей.
Если вы исключите таблицу `resource*activity*log`, история всех ресурсов будет потеряна.
Таблицу `resource*activity*log` можно безопасно исключать для случаев, связанных с разработкой.

|===

## База данных MySQL

Если вы используете другой движок базы данных, обратитесь к его документации о том, как выполнять резервное копирование.

### Резервное копирование

!!! note
    Обратитесь к разделу <<reduce-size,уменьшение размера резервной копии>>, чтобы узнать, какие таблицы вы можете исключить для вашего сценария.
    Пример команды, исключающей определённые таблицы, приведён ниже.


Мы рекомендуем использовать инструмент `mysqldump`.
Он встроен в контейнер `db` (образ `percona:8.0`).

!!! caution
    Не пытайтесь копировать сырые файлы базы данных для выполнения резервного копирования.
    Это может привести к повреждению данных.


!!! warning
    По умолчанию `mysqldump` блокирует таблицы при выполнении экспорта.
    Блокировки таблиц могут вызвать проблемы при работе в продакшене, так что имейте это в виду.


.Команда дампа базы данных:
```bash
```
# This dumps the entire database and place it in the dump.sql file.
docker-compose exec -T \
    --env MYSQL_PWD=your-password db \
    mysqldump your-db-name --add-drop-database -u your-username > dump.sql

# This dumps the database without actionlog, automation sessions, and resource activity log
# These are generally the largest and can safely be omitted.
docker-compose exec -T \
    --env MYSQL_PWD=your-password db \
    mysqldump your-db-name --add-drop-database --ignore-table=dbname.actionlog --ignore-table=dbname.automation*sessions --ignore-table=dbname.resource*activity_log -u your-username > dump.sql

!!! caution
    Если вы изменили имя сервиса базы данных (`db`) в вашем `docker-compose.yaml`, обязательно измените его и в приведённой выше команде.


### Восстановление

!!! note
    Мы рекомендуем выключить сервер LowCoooode до завершения процедуры восстановления.


.Команда восстановления базы данных:
```bash
```
# This restores the database based on the dump.sql file.
docker-compose exec -T \
    --env MYSQL_PWD=your-password db \
    mysql your-db-name -u your-username < dump.sql

!!! caution
    Если вы изменили имя сервиса базы данных (`db`) в вашем `docker-compose.yaml`, обязательно измените его и в приведённой выше команде.


## База данных PostgreSQL

### Резервное копирование

!!! note
    Обратитесь к разделу <<reduce-size,уменьшение размера резервной копии>>, чтобы узнать, какие таблицы вы можете исключить для вашего сценария.
    Пример команды, исключающей определённые таблицы, приведён ниже.


Мы рекомендуем использовать инструмент `pg*dumpall` или `pg*dump`.
`pg_dumpall` — это утилита для записи («дампа») всех баз данных PostgreSQL кластера в один файл скрипта.
Файл скрипта содержит SQL-команды, которые можно использовать как входные данные для psql для восстановления баз данных.
Это делается путём вызова `pg_dump` для каждой базы данных в кластере.

!!! caution
    Не пытайтесь копировать сырые файлы базы данных для выполнения резервного копирования.
    Это может привести к повреждению данных.


!!! warning
    По умолчанию `pg_dump` блокирует таблицы при выполнении экспорта.
    Блокировки таблиц могут вызвать проблемы при работе в продакшене, так что имейте это в виду.


.Команда дампа базы данных:
```bash
```
# This dumps all databases and place them in the dump.sql file.
docker-compose exec db \
    pg_dumpall -c -U your-username > dump.sql

# This dumps the entire database and place it in the dump.sql file.
docker-compose exec db \
    pg_dump -d your-db-name -c -U your-username > dump.sql

# To reduce the size of the sql,
# This dumps all databases and place them in the dump.gz file.
docker-compose exec db \
    pg_dumpall -c -U your-username | \
    gzip > /var/data/postgres/backups/dump.gz


# This dumps the database without actionlog, automation sessions, and resource activity log
# These are generally the largest and can safely be omitted.
docker-compose exec db \
    pg*dump -T lowcode.actionlog -T lowcode.automation*sessions -T lowcode.resource*activity*log -c -U your-username lowcode > dump.sql

!!! caution
    Если вы изменили имя сервиса базы данных (`db`) в вашем `docker-compose.yaml`, обязательно измените его и в приведённой выше команде.


### Восстановление

!!! note
    Рекомендуется выключить сервер LowCoooode до завершения процедуры восстановления.


.Команда восстановления базы данных:
```bash
```
# This restores the database based on the dump.sql file.
cat dump.sql | \
    docker-compose exec db psql -U your-username

# This restores a specific database based on the dump.sql file.
cat dump.sql | \
    docker-compose exec db psql -U your-username -d your-db-name

# To restore a compressed sql,
# This restores the database based on the dump.gz file.
gzip < dump.gz | \
    docker-compose exec db psql -U your-username

!!! caution
    Если вы изменили имя сервиса базы данных (`db`) в вашем `docker-compose.yaml`, обязательно измените его и в приведённой выше команде.


## Файлы

### Резервное копирование

Без сервиса объектного хранения, такого как Min.io, загруженные файлы хранятся непосредственно в файловой системе.
Сервер LowCoooode хранит данные в каталоге `/data` (если не настроено иначе с помощью переменных окружения `**STORAGE*PATH`).

Вы можете использовать любые стандартные инструменты управления файлами, чтобы сделать резервную копию файлов.

.Сжатие файлов с помощью команды `tar`:
```bash
```
# This compresses all your uploaded files into the backup.tar.bz2 archive,
tar -cjf backup.tar.bz2 data/server/
### Восстановление

.Извлечение файлов из архива с помощью команды `tar`:
```bash
```
# This restores your backup.tar.bz2 archive
tar -xjf backup.tar.bz2
