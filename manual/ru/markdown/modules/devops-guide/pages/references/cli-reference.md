# Справочник по CLI

!!! note
    *DevNote* подумать о добавлении дополнительных примеров/сценариев использования?
    Обновить/убедиться, что команды всё ещё актуальны и действительны.


!!! note
    *DevNote* добавить больше примечаний и описаний.
    Рассмотреть возможность генерации этого.


Инструмент командной строки LowCoooode позволяет быстро взаимодействовать с различными частями системы — от изменения настроек до назначения ролей пользователям.

## Запуск CLI через docker

Если вы запускаете развёртывание через docker или docker compose, вы можете легко выполнять команды уже запущенного сервера вне контейнеров.

### Docker Compose

!!! note
    Смотрите больше примеров docker compose в [полезных командах Docker](modules/devops-guide/pages/references/index.md#useful-commands).


```shell
```
$ docker-compose exec <service name> ./bin/lowcode-server <command>

### Docker

Сначала найдите ID контейнера, в котором запущен ваш сервер LowCoooode в текущем каталоге проекта.
Вы можете использовать полученный ID контейнера для выполнения команд CLI, как показано в следующем примере.

!!! note
    В примере предполагается, что имя сервиса — `server`.


```shell
```
$ docker ps | grep `basename $PWD | sed -e 's/\.//g'`_server
$ docker exec <container ID> ./bin/lowcode-server <command>

## Аутентификация

### Автоматическое обнаружение

Автоматически обнаруживает новый OIDC-клиент.

```shell
```
Usage:
  lowcode-server auth auto-discovery [name] [url] [flags]

Flags:
      --enable            Enable this provider and external auth
  -h, --help              help for auto-discovery
      --skip-validation   Skip validation

<a id="auth-jwt"></a>
### JWT

Генерирует новый JWT для пользователя.

```shell
```
Usage:
  lowcode-server auth jwt [email-or-id] [flags]

Flags:
  -h, --help   help for jwt

### Тестовое уведомление

Отправляет образцы всех уведомлений аутентификации получателю.

```shell
```
Usage:
  lowcode-server auth test-notifications [recipient] [flags]

Flags:
  -h, --help   help for test-notifications

<a id="import"></a>
## Import

Импортирует данные из yaml-источников.

!!! caution
    При импорте правил RBAC изменения не применяются автоматически.
    Правила RBAC перезагружаются с интервалом в 1 час.
    Если вам нужно немедленно отразить изменения, вам потребуется перезапустить сервер LowCoooode.


```shell
```
Usage:
  lowcode-server import [flags]

Flags:
  -h, --help                   help for import
      --merge-left-existing    Update any existing values; existing data takes priority. Default skips.
      --merge-right-existing   Update any existing values; new data takes priority. Default skips.
      --replace-existing       Replace any existing values. Default skips.

## Provision

Задачи provision

```shell
```
Usage:
  lowcode-server provision [flags]

Flags:
  -h, --help   help for provision

## RBAC

Проверяет и изменяет разрешения.

### Check

Проверяет применённые разрешения на соответствие заданному файлу (пока поддерживает только compose-разрешения).

```shell
```
Usage:
  lowcode-server rbac check [flags]

Flags:
  -h, --help   help for check

## Roles

Управление ролями

### Add user

Добавляет пользователя в роль.

```shell
```
Usage:
  lowcode-server roles useradd [role-ID-or-name-or-handle] [user-ID-or-email] [flags]

Flags:
  -h, --help   help for useradd

## Serve api

Запускает HTTP-сервер с REST API.

```shell
```
Usage:
  lowcode-server serve-api [flags]

Aliases:
  serve-api, serve

Flags:
  -h, --help   help for serve-api

<a id="sink-signature"></a>
## Sink signature

Создаёт сигнатуру для sink HTTP-эндпоинта.

```shell
```
Usage:
  lowcode-server sink signature [flags]

Flags:
      --content-type string   Content type (optional)
      --expires string        Date of expiration (YYYY-MM-DD, optional)
  -h, --help                  help for signature
      --max-body-size int     Max allowed body size
      --method string         HTTP method that will be used (optional)
      --origin string         Origin of the request (arbitrary string, optional)
      --path string           Full sink request path (do not include /sink prefix, add / for just root)
      --signature-in-path     Include signature in a path instead of query string

## Upgrading

```shell
```
Usage:
  lowcode-server upgrade [flags]

Flags:
  -h, --help   help for upgrade

## Users

Управление пользователями

### Add user

Добавить нового пользователя

```shell
```
Usage:
  lowcode-server users add [email] [flags]

Flags:
  -h, --help          help for add
      --no-password   Create user without password

### List

Список пользователей

```shell
```
Usage:
  lowcode-server users list [flags]

Flags:
  -h, --help           help for list
  -l, --limit int      How many entry to display (default 20)
  -q, --query string   Query and filter by handle, email, name

### Password

Смена пароля пользователя

```shell
```
Usage:
  lowcode-server users password [email] [flags]

Flags:
  -h, --help   help for password

## Federation Sync

.Есть две команды, с помощью которых вы можете управлять синхронизацией federation:
- синхронизация данных,
- синхронизация структуры.

Эти две команды (`lowcode-server sync data` и `lowcode-server sync structure`) заставляют фоновый наблюдатель запуститься.

```shell
```
Usage:
  lowcode-server sync [command]

Available Commands:
  data        Sync data
  structure   Sync structure

Flags:
  -h, --help   help for sync

## Version

```shell
```
Usage:
  lowcode-server version [flags]

Flags:
  -h, --help   help for version
