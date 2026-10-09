# Конфигурация сервера Corredor

.[#CORREDOR*ADDR]#[CORREDOR*ADDR,**CORREDOR*ADDR**](#CORREDOR*ADDR,**CORREDOR_ADDR**)#
- **type**: `string`
- **description**: Будьте осторожны при передаче ваших переменных окружения: `CORREDOR_ADDR` используется и Corredor, и API-сервером.

При использовании сервером Corredor задаёт IP-адрес и порт, на которых слушает сервер.
По умолчанию `0.0.0.0:80` в Docker и `localhost:50051` в режиме разработки.
Это указывает серверу слушать все интерфейсы, все IP-адреса на порту 80.

Он также используется API-сервером для указания того, где сервер может найти и подключиться к серверу Corredor.
Значения по умолчанию определяются контекстом. При сборке из исходников и запуске в режиме разработки значение равно `localhost:50051`,
в Docker-контейнере оно установлено как `corredor:80`.

Эти зависящие от контекста значения по умолчанию позволяют нам иметь рабочую конфигурацию без дополнительных изменений.

- **default**: `0.0.0.0:80`

.[#CORREDOR*ENABLED]#[CORREDOR*ENABLED,**CORREDOR*ENABLED**](#CORREDOR*ENABLED,**CORREDOR_ENABLED**)#
- **type**: `bool`
- **description**: Используется только API-сервером для включения или отключения использования сервера Corredor.
Сервер Corredor игнорирует эту конфигурацию.

- **default**: `false`

.[#CORREDOR*ENABLED]#[CORREDOR*ENVIRONMENT,**CORREDOR*ENVIRONMENT**](#CORREDOR*ENVIRONMENT,**CORREDOR_ENVIRONMENT**)#
- **type**: `string`
- **description**: Включает режим разработки, когда значение переменной начинается с `"dev"`.
В противном случае Corredor запускается в производственном режиме.

Когда `CORREDOR*ENVIRONMENT` не установлен, Corredor ищет `CORREDOR*ENV` и `NODE_ENV` и использует первое доступное значение.

- **default**: `false`

.[#CORREDOR*MAX*BACKOFF*DELAY]#[CORREDOR*MAX*BACKOFF*DELAY,**CORREDOR*MAX*BACKOFF*DELAY**](#CORREDOR*MAX*BACKOFF*DELAY,**CORREDOR*MAX*BACKOFF_DELAY**)#
- **type**: `duration`
- **description**: Максимальное время ожидания перед повторной попыткой неудачного подключения от API к серверу Corredor.
- **default**: `1m`

.[#CORREDOR*EXEC*CSERVERS*API*BASEURL*TEMPLATE]#[CORREDOR*EXEC*CSERVERS*API*BASEURL*TEMPLATE,**CORREDOR*EXEC*CSERVERS*API*BASEURL*TEMPLATE**](#CORREDOR*EXEC*CSERVERS*API*BASEURL*TEMPLATE,**CORREDOR*EXEC*CSERVERS*API*BASEURL_TEMPLATE**)#
- **type**: `string`
- **description**: Расположение API сервера LowCoooode.
\{host\} заменяется значением из переменных окружения (в этом порядке: `CORREDOR*EXEC*CSERVERS*API*HOST`, `DOMAIN`, `HOSTNAME`, `HOST`), \{service\} динамически заменяется внутри Corredor на `compose`, `system` или `messaging`.
- **default**: `https://api.\{host\}/\{service\}`

.[#CORREDOR*EXEC*CSERVERS*API*HOST]#[CORREDOR*EXEC*CSERVERS*API*HOST,**CORREDOR*EXEC*CSERVERS*API*HOST**](#CORREDOR*EXEC*CSERVERS*API*HOST,**CORREDOR*EXEC*CSERVERS*API*HOST**)#
- **type**: `string`
- **description**: Имя хоста используется для шаблона.
- **default**: ``

.[#CORREDOR*LOG*ENABLED]#[CORREDOR*LOG*ENABLED,**CORREDOR*LOG*ENABLED**](#CORREDOR*LOG*ENABLED,**CORREDOR*LOG*ENABLED**)#
- **type**: `boolean`
- **description**: Включить ведение журнала на сервере Corredor.
- **default**: `corredor`

.[#CORREDOR*LOG*LEVEL]#[CORREDOR*LOG*LEVEL,**CORREDOR*LOG*LEVEL**](#CORREDOR*LOG*LEVEL,**CORREDOR*LOG*LEVEL**)#
- **type**: `string`
- **description**: По умолчанию `trace`, когда `CORREDOR_ENVIRONMENT` установлен в `dev`.
Определяет объём информации, которую ведёт сервер.
- **default**: `info`

.[#CORREDOR*LOG*PRETTY]#[CORREDOR*LOG*PRETTY,**CORREDOR*LOG*PRETTY**](#CORREDOR*LOG*PRETTY,**CORREDOR*LOG*PRETTY**)#
- **type**: `boolean`
- **description**: Если установлено true, журналы форматируются для упрощения разработки.
- **default**: `false`

.[#CORREDOR*EXT*DEPENDENCIES*AUTO*UPDATE]#[CORREDOR*EXT*DEPENDENCIES*AUTO*UPDATE,**CORREDOR*EXT*DEPENDENCIES*AUTO*UPDATE**](#CORREDOR*EXT*DEPENDENCIES*AUTO*UPDATE,**CORREDOR*EXT*DEPENDENCIES*AUTO*UPDATE**)#
- **type**: `boolean`
- **description**: Corredor автоматически обновляет зависимости скриптов, найденные в файлах `package.json`.
- **default**: `true`

.[#CORREDOR*EXT*SERVER*SCRIPTS*ENABLED]#[CORREDOR*EXT*SERVER*SCRIPTS*ENABLED,**CORREDOR*EXT*SERVER*SCRIPTS*ENABLED**](#CORREDOR*EXT*SERVER*SCRIPTS*ENABLED,**CORREDOR*EXT*SERVER*SCRIPTS*ENABLED**)#
- **type**: `boolean`
- **description**: Серверные скрипты включены.
- **default**: `true`

.[#CORREDOR*EXT*SERVER*SCRIPTS*WATCH]#[CORREDOR*EXT*SERVER*SCRIPTS*WATCH,**CORREDOR*EXT*SERVER*SCRIPTS*WATCH**](#CORREDOR*EXT*SERVER*SCRIPTS*WATCH,**CORREDOR*EXT*SERVER*SCRIPTS*WATCH**)#
- **type**: `boolean`
- **description**: Corredor перезагружает серверные скрипты при изменении.
- **default**: `true`

.[#CORREDOR*EXT*CLIENT*SCRIPTS*ENABLED]#[CORREDOR*EXT*CLIENT*SCRIPTS*ENABLED,**CORREDOR*EXT*CLIENT*SCRIPTS*ENABLED**](#CORREDOR*EXT*CLIENT*SCRIPTS*ENABLED,**CORREDOR*EXT*CLIENT*SCRIPTS*ENABLED**)#
- **type**: `boolean`
- **description**: Клиентские скрипты включены.
- **default**: `true`

.[#CORREDOR*EXT*CLIENT*SCRIPTS*WATCH]#[CORREDOR*EXT*CLIENT*SCRIPTS*WATCH,**CORREDOR*EXT*CLIENT*SCRIPTS*WATCH**](#CORREDOR*EXT*CLIENT*SCRIPTS*WATCH,**CORREDOR*EXT*CLIENT*SCRIPTS*WATCH**)#
- **type**: `boolean`
- **description**: Corredor перезагружает клиентские скрипты при изменении.
- **default**: `true`

.[#CORREDOR*SERVER*CERTIFICATES*ENABLED]#[CORREDOR*SERVER*CERTIFICATES*ENABLED,**CORREDOR*SERVER*CERTIFICATES*ENABLED**](#CORREDOR*SERVER*CERTIFICATES*ENABLED,**CORREDOR*SERVER*CERTIFICATES_ENABLED**)#
- **type**: `boolean`
- **description**: Требуется действительный сертификат для подключения к серверу Corredor.
Установите false при запуске в режиме разработки.

Используется сервером Corredor.
- **default**: `true`

.[#CORREDOR*SERVER*CERTIFICATES*PATH]#[CORREDOR*SERVER*CERTIFICATES*PATH,**CORREDOR*SERVER*CERTIFICATES*PATH**](#CORREDOR*SERVER*CERTIFICATES*PATH,**CORREDOR*SERVER*CERTIFICATES_PATH**)#
- **type**: `string`
- **description**: Базовый путь для всех файлов сертификатов

Используется сервером Corredor.
- **default**: `/certs`

.[#CORREDOR*SERVER*CERTIFICATES*CA]#[CORREDOR*SERVER*CERTIFICATES*CA,**CORREDOR*SERVER*CERTIFICATES*CA**](#CORREDOR*SERVER*CERTIFICATES*CA,**CORREDOR*SERVER*CERTIFICATES_CA**)#
- **type**: `string`
- **description**: Путь к файлу центра сертификации

Используется сервером Corredor.
- **default**: `ca.crt`

.[#CORREDOR*SERVER*CERTIFICATES*PRIVATE]#[CORREDOR*SERVER*CERTIFICATES*PRIVATE,**CORREDOR*SERVER*CERTIFICATES*PRIVATE**](#CORREDOR*SERVER*CERTIFICATES*PRIVATE,**CORREDOR*SERVER*CERTIFICATES_PRIVATE**)#
- **type**: `string`
- **description**: Путь к файлу закрытого ключа для сервера Corredor.

Используется сервером Corredor.
- **default**: `private.key`

.[#CORREDOR*SERVER*CERTIFICATES*PUBLIC]#[CORREDOR*SERVER*CERTIFICATES*PUBLIC,**CORREDOR*SERVER*CERTIFICATES*PUBLIC**](#CORREDOR*SERVER*CERTIFICATES*PUBLIC,**CORREDOR*SERVER*CERTIFICATES_PUBLIC**)#
- **type**: `string`
- **description**: Путь к файлу сертификата для сервера Corredor.

Используется сервером Corredor.
- **default**: `public.crt`

.[#CORREDOR*CLIENT*CERTIFICATES*ENABLED]#[CORREDOR*CLIENT*CERTIFICATES*ENABLED,**CORREDOR*CLIENT*CERTIFICATES*ENABLED**](#CORREDOR*CLIENT*CERTIFICATES*ENABLED,**CORREDOR*CLIENT*CERTIFICATES_ENABLED**)#
- **type**: `boolean`
- **description**: Установить безопасное подключение к серверу Corredor.
Установите false при запуске в режиме разработки.

Используется API-сервером.
- **default**: `true`

.[#CORREDOR*CLIENT*CERTIFICATES*PATH]#[CORREDOR*CLIENT*CERTIFICATES*PATH,**CORREDOR*CLIENT*CERTIFICATES*PATH**](#CORREDOR*CLIENT*CERTIFICATES*PATH,**CORREDOR*CLIENT*CERTIFICATES_PATH**)#
- **type**: `string`
- **description**: Базовый путь для всех файлов сертификатов.

Используется API-сервером.
- **default**: `/certs/corredor/client`

.[#CORREDOR*CLIENT*CERTIFICATES*CA]#[CORREDOR*CLIENT*CERTIFICATES*CA,**CORREDOR*CLIENT*CERTIFICATES*CA**](#CORREDOR*CLIENT*CERTIFICATES*CA,**CORREDOR*CLIENT*CERTIFICATES_CA**)#
- **type**: `string`
- **description**: Путь к файлу центра сертификации.

Используется API-сервером.
- **default**: `ca.crt`

.[#CORREDOR*CLIENT*CERTIFICATES*PRIVATE]#[CORREDOR*CLIENT*CERTIFICATES*PRIVATE,**CORREDOR*CLIENT*CERTIFICATES*PRIVATE**](#CORREDOR*CLIENT*CERTIFICATES*PRIVATE,**CORREDOR*CLIENT*CERTIFICATES_PRIVATE**)#
- **type**: `string`
- **description**: Путь к файлу закрытого ключа для сервера Corredor.

Используется API-сервером.
- **default**: `private.key`

.[#CORREDOR*CLIENT*CERTIFICATES*PUBLIC]#[CORREDOR*CLIENT*CERTIFICATES*PUBLIC,**CORREDOR*CLIENT*CERTIFICATES*PUBLIC**](#CORREDOR*CLIENT*CERTIFICATES*PUBLIC,**CORREDOR*CLIENT*CERTIFICATES_PUBLIC**)#
- **type**: `string`
- **description**: Путь к файлу сертификата для сервера Corredor.

Используется API-сервером.
- **default**: `public.crt`
