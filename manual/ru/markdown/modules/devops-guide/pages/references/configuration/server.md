# Конфигурация сервера

:leveloffset: +1


# Подключение к хранилищу данных

## *DB_DSN*

### Type

`string`


### Default

```
```
sqlite3://file::memory:?cache=shared&mode=memory
### Description

Строка подключения к базе данных.
# HTTP Client

## *HTTP_CLIENT_TIMEOUT*

### Type

`time.Duration`


### Default

```
```
30s
### Description

Тайм-аут по умолчанию для клиентов.

## *HTTP_CLIENT_TLS_INSECURE*

### Type

`bool`


### Description

Разрешить небезопасные (недействительные, с истекшим сроком действия TLS/SSL сертификаты) подключения.

!!! important
    Мы настоятельно рекомендуем оставлять это значение равным false, за исключением локальной разработки или демонстраций.

# HTTP Server

## *HTTP_ADDR*

### Type

`string`


### Default

```
```
:80
### Description

IP-адрес и порт для HTTP-сервера.

## *HTTP_API_BASE_URL*

### Type

`string`


### Default

```
```
/
### Description

Когда веб-приложения включены (HTTP*WEBAPP*ENABLED), этот путь перемещается на '/api', если не указано иное.
Базовый URL API внутренне дополняется baseUrl.

## *HTTP_API_ENABLED*

### Type

`bool`


## *HTTP_SERVER_ASSETS_PATH*

### Type

`string`


### Description

LowCoooode будет напрямую обслуживать эти ресурсы (статические файлы).
Если путь пуст (значение по умолчанию), используются встроенные файлы.

## *HTTP_BASE_URL*

### Type

`string`


### Default

```
```
/
### Description

Базовый URL (префикс) для всех маршрутов (<baseUrl>/auth, <baseUrl>/api, ...)

## *DOMAIN*

### Type

`string`


### Default

```
```
localhost
### Description

Домен для HTTP-сервера.

## *DOMAIN_WEBAPP*

### Type

`string`


### Default

```
```
localhost
### Description

Домен для HTTP веб-приложения.

## *HTTP_ENABLE_DEBUG_ROUTE*

### Type

`bool`


### Description

Включить маршрут `/debug`.

## *HTTP_ENABLE_HEALTHCHECK_ROUTE*

### Type

`bool`


## *HTTP_METRICS*

### Type

`bool`


### Description

Включить метрики (prometheus).

## *HTTP_REPORT_PANIC*

### Type

`bool`


### Description

Отправлять отчёты о панике HTTP в Sentry.

## *HTTP_ENABLE_VERSION_ROUTE*

### Type

`bool`


### Description

Включить маршрут `/version`.

## *HTTP_LOG_REQUEST*

### Type

`bool`


### Description

Логировать HTTP-запросы.

## *HTTP_LOG_RESPONSE*

### Type

`bool`


### Description

Логировать HTTP-ответы.

## *HTTP_METRICS_PASSWORD*

### Type

`string`


### Description

Пароль для эндпоинта метрик.

## *HTTP_METRICS_NAME*

### Type

`string`


### Default

```
```
lowcode
### Description

Имя эндпоинта метрик.

## *HTTP_METRICS_USERNAME*

### Type

`string`


### Default

```
```
metrics
### Description

Имя пользователя для эндпоинта метрик.

## *HTTP_SSL_TERMINATED*

### Type

`bool`


### Description

Включено ли прекращение SSL во входящем прокси или балансировщике нагрузки перед LowCoooode?
По умолчанию LowCoooode проверяет наличие переменной окружения LETSENCRYPT_HOST.
Это НЕ включает прекращение SSL в LowCoooode!

## *HTTP_ERROR_TRACING*

### Type

`bool`


## *HTTP_SERVER_WEB_CONSOLE_ENABLED*

### Type

`bool`


### Description

Включить веб-консоль. При запуске в среде разработки веб-консоль включена по умолчанию.

## *HTTP_SERVER_WEB_CONSOLE_PASSWORD*

### Type

`string`


### Description

Пароль для эндпоинта веб-консоли. При запуске в среде разработки пароль не требуется.

LowCoooode намеренно устанавливает пароль по умолчанию на случайные символы для предотвращения инцидентов безопасности.

## *HTTP_SERVER_WEB_CONSOLE_USERNAME*

### Type

`string`


### Default

```
```
admin
### Description

Имя пользователя для эндпоинта веб-консоли.

## *HTTP_WEBAPP_BASE_DIR*

### Type

`string`


### Default

```
```
./webapp/public


## *HTTP_WEBAPP_BASE_URL*

### Type

`string`


### Default

```
```
/
### Description

Базовый URL веб-приложения внутренне дополняется baseUrl.

## *HTTP_WEBAPP_ENABLED*

### Type

`bool`


## *HTTP_WEBAPP_LIST*

### Type

`string`


### Default

```
```
admin,compose,workflow,reporter
# RBAC options

## *RBAC_ANONYMOUS_ROLES*

### Type

`string`


### Default

```
```
anonymous
### Description

Список идентификаторов ролей, разделённых пробелами.
Эти роли автоматически назначаются анонимному пользователю.
Членство в этих ролях не может управляться.

## *RBAC_AUTHENTICATED_ROLES*

### Type

`string`


### Default

```
```
authenticated
### Description

Список идентификаторов ролей, разделённых пробелами.
Эти роли автоматически назначаются аутентифицированному пользователю.
Членство в этих ролях не может управляться.
Система откажется запускаться, если роли, перечисленные здесь, также указаны в анонимных ролях.

## *RBAC_BYPASS_ROLES*

### Type

`string`


### Default

```
```
super-admin
### Description

Список идентификаторов ролей, разделённых пробелами.
Эти роли вызывают сокращённую проверку контроля доступа, разрешая все операции.
Система откажется запускаться, если роли, обходящие проверку, также указаны как аутентифицированные или анонимные автоматически назначаемые роли.

## *RBAC_LOG*

### Type

`bool`


### Description

Логировать события и действия, связанные с RBAC.

## *RBAC_SERVICE_USER*

### Type

`string`


# SCIM Server

## *SCIM_BASE_URL*

### Type

`string`


### Default

```
```
/scim
### Description

Префикс для эндпоинтов API SCIM.

## *SCIM_ENABLED*

### Type

`bool`


### Description

Включить подсистему SCIM.

## *SCIM_EXTERNAL_ID_AS_PRIMARY*

### Type

`bool`


### Description

Использовать внешние ID в эндпоинтах API SCIM.

## *SCIM_EXTERNAL_ID_VALIDATION*

### Type

`string`


### Default

```
```
^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$
### Description

Проверяет формат внешних ID. По умолчанию UUID.

## *SCIM_SECRET*

### Type

`string`


### Description

Секрет для проверки запросов на эндпоинтах API SCIM.
# Email sending

## *SMTP_FROM*

### Type

`string`


### Description

Параметр email `from` SMTP.

## *SMTP_HOST*

### Type

`string`


### Description

Имя хоста SMTP-сервера.

## *SMTP_PASS*

### Type

`string`


### Description

Пароль SMTP.

## *SMTP_PORT*

### Type

`int`


### Description

Порт SMTP.

## *SMTP_TLS_INSECURE*

### Type

`bool`


### Description

Разрешить небезопасные (недействительные, с истекшим сроком действия TLS сертификаты) подключения.

## *SMTP_TLS_SERVER_NAME*

### Type

`string`


## *SMTP_USER*

### Type

`string`


### Description

Имя пользователя SMTP.
# Actionlog

## *ACTIONLOG_DEBUG*

### Type

`bool`


## *ACTIONLOG_ENABLED*

### Type

`bool`


## *ACTIONLOG_WORKFLOW_FUNCTIONS_ENABLED*

### Type

`bool`


# API Gateway

## *APIGW_DEBUG*

### Type

`bool`


### Description

Включить отладочную информацию API Gateway.

## *APIGW_ENABLED*

### Type

`bool`


### Description

Включить API Gateway.

## *APIGW_LOG_ENABLED*

### Type

`bool`


### Description

Включить дополнительное логирование.

## *APIGW_LOG_REQUEST_BODY*

### Type

`bool`


### Description

Включить вывод тела входящего запроса в логах.

## *APIGW_PROFILER_ENABLED*

### Type

`bool`


### Description

Включить профилировщик.

## *APIGW_PROFILER_GLOBAL*

### Type

`bool`


### Description

Профилировщик включён для всех маршрутов.

## *APIGW_PROXY_ENABLE_DEBUG_LOG*

### Type

`bool`


### Description

Включить полный отладочный лог запросов/ответов — предупреждение: содержит конфиденциальные данные.

## *APIGW_PROXY_FOLLOW_REDIRECTS*

### Type

`bool`


### Description

Следовать перенаправлениям при прокси-запросах.

## *APIGW_PROXY_OUTBOUND_TIMEOUT*

### Type

`time.Duration`


### Default

```
```
30s
### Description

Тайм-аут исходящего запроса.
# Authentication

## *AUTH_OAUTH2_ACCESS_TOKEN_LIFETIME*

### Type

`time.Duration`


### Default

```
```
2h
### Description

Время жизни токена доступа. Должно быть короче времени жизни токена обновления.

## *AUTH_ASSETS_PATH*

### Type

`string`


### Description

Путь к исходным файлам js, css, изображений и шаблонов.

При запуске LowCoooode, если путь существует, он пытается загрузить файлы шаблонов из него.

Если путь пуст (значение по умолчанию), используются встроенные файлы.

## *AUTH_BASE_URL*

### Type

`string`


### Description

Базовый URL фронтенда. Должен быть абсолютным URL с доменом.
Используется для некоторых перенаправлений и ссылок в письмах аутентификации.

## *AUTH_CSRF_COOKIE_NAME*

### Type

`string`


### Default

```
```
same-site-authenticity-token
### Description

Имя cookie, используемое для защиты CSRF.

## *AUTH_CSRF_ENABLED*

### Type

`bool`


### Description

Включить защиту CSRF.

## *AUTH_CSRF_FIELD_NAME*

### Type

`string`


### Default

```
```
same-site-authenticity-token
### Description

Имя поля формы, используемое для защиты CSRF.

## *AUTH_CSRF_SECRET*

### Type

`string`


### Description

Секрет, используемый для защиты CSRF.

!!! important
    Если секрет не установлен, система автоматически генерирует его из переменных окружения DB_DSN и HOSTNAME.
    Сгенерированный секрет изменится, если вы измените любую из этих переменных.


## *AUTH_DEFAULT_CLIENT*

### Type

`string`


### Default

```
```
lowcode-webapp
### Description

Идентификатор OAuth2-клиента, используемый для автоматического перенаправления с эндпоинта /auth/oauth2/go.

Это упрощает настройку OAuth2-потока для веб-приложений LowCoooode, устраняя необходимость
указывать URL перенаправления и ID клиента (endpoint oauth2/go делает это внутренне).


## *AUTH_DEVELOPMENT_MODE*

### Type

`bool`


### Description

При включении LowCoooode перезагружает шаблоны перед каждым выполнением.
Включите это для отладки или при разработке шаблонов аутентификации.

Должно быть отключено в производственной среде, где шаблоны не меняются между перезапусками сервера.

## *AUTH_EXTERNAL_COOKIE_SECRET*

### Type

`string`


### Description

Секрет, используемый для защиты cookie.

!!! important
    Если секрет не установлен, система автоматически генерирует его из переменных окружения DB_DSN и HOSTNAME.
    Сгенерированный секрет изменится, если вы измените любую из этих переменных.


## *AUTH_EXTERNAL_REDIRECT_URL*

### Type

`string`


### Description

URL перенаправления, отправляемый с запросом аутентификации OAuth2 к провайдеру.

Плейсхолдер `provider` заменяется на фактическое значение при использовании.

## *AUTH_GARBAGE_COLLECTOR_INTERVAL*

### Type

`time.Duration`


### Default

```
```
15min
### Description

Как часто просроченные сессии и токены удаляются из базы данных.

## *AUTH_JWT_ALGORITHM*

### Type

`string`


### Default

```
```
HS512
### Description

Алгоритм, используемый для подписи JWT.

Поддерживаемые значения:
 - HS256, HS384, HS512
 - PS256, PS384, PS512,
 - RS256, RS384, RS512

Укажите общий секретный ключ для алгоритмов HS256, HS384, HS512 и полный закрытый ключ или путь к файлу для алгоритмов PS* и RS*.

## *AUTH_JWT_KEY*

### Type

`string`


### Description

Сырой закрытый ключ или абсолютный или относительный путь к файлу, содержащему его.

## *AUTH_LOG_ENABLED*

### Type

`bool`


### Description

Включить дополнительное логирование для потоков аутентификации.

## *AUTH_PASSWORD_SECURITY*

### Type

`bool`


### Description

Безопасность паролей позволяет отключить ограничения, которым должны соответствовать пароли.

!!! caution
    Отключение безопасности паролей может быть полезно для сред разработки, поскольку устраняет необходимость в сложных паролях.
    Безопасность паролей *должна быть включена* в производственных средах для предотвращения инцидентов безопасности.


## *AUTH_PROVISION_SUPER_USER*

### Type

`string`


### Description

При установке LowCoooode создаёт одного или нескольких пользователей с заданными значениями, используя указанный email в качестве пароля.
Пропускает существующих (email, handle). Все новые пользователи назначаются во все обходные роли.

При установке в производственной среде LowCoooode останавливается и сообщает об ошибке.

## *AUTH_OAUTH2_REFRESH_TOKEN_LIFETIME*

### Type

`time.Duration`


### Default

```
```
72h
### Description

Время жизни токена обновления. Должно быть намного больше времени жизни токена доступа.

Токены обновления используются для обмена истёкших токенов доступа на новые.

## *AUTH_REQUEST_RATE_LIMIT*

### Type

`int`


### Default

```
```
60
### Description

Сколько запросов с определённого IP-адреса разрешено во временном окне.
Установите 0 для отключения.

## *AUTH_REQUEST_RATE_WINDOW_LENGTH*

### Type

`time.Duration`


### Default

```
```
1m
### Description

Сколько запросов с определённого IP-адреса разрешено во временном окне.

## *AUTH_JWT_SECRET*

### Type

`string`


### Description

Секрет, используемый для подписи JWT-токенов.
Значение используется только при использовании алгоритмов HS256, HS384 или HS512.

!!! important
    Если секрет не установлен, система автоматически генерирует его из переменных окружения DB_DSN и HOSTNAME.
    Сгенерированный секрет изменится, если вы измените любую из этих переменных.


## *AUTH_SESSION_COOKIE_DOMAIN*

### Type

`string`


### Description

Домен cookie сессии.

## *AUTH_SESSION_COOKIE_NAME*

### Type

`string`


### Default

```
```
session
### Description

Имя cookie сессии.

## *AUTH_SESSION_COOKIE_PATH*

### Type

`string`


### Description

Путь cookie сессии.

## *AUTH_SESSION_COOKIE_SECURE*

### Type

`bool`


### Description

По умолчанию true при использовании HTTPS. LowCoooode попытается угадать эту настройку.

## *AUTH_SESSION_LIFETIME*

### Type

`time.Duration`


### Default

```
```
24h
### Description

Максимальное время бездействия пользователя при входе без опции «запомнить меня» до истечения сессии.

Рекомендуемое значение — от часа до суток.

!!! important
    Это влияет только на страницы профиля (/auth). Использование приложений (admin, compose, ...) не продлевает сессию.


## *AUTH_SESSION_PERM_LIFETIME*

### Type

`time.Duration`


### Default

```
```
8640h
### Description

Длительность сессии в /auth, когда пользователь входит с опцией «запомнить меня».

Если установлено 0, опция «запомнить меня» удаляется.
# Connection to Corredor

## *CORREDOR_ADDR*

### Type

`string`


### Default

```
```
localhost:50051
### Description

Имя хоста и порт gRPC-сервера Corredor.

## *CORREDOR_DEFAULT_EXEC_TIMEOUT*

### Type

`time.Duration`


## *CORREDOR_ENABLED*

### Type

`bool`


### Description

Включить/отключить интеграцию Corredor.

## *CORREDOR_LIST_REFRESH*

### Type

`time.Duration`


## *CORREDOR_LIST_TIMEOUT*

### Type

`time.Duration`


### Default

```
```
2s

## *CORREDOR_MAX_BACKOFF_DELAY*

### Type

`time.Duration`


### Default

```
```
1m
### Description

Максимальная задержка для повторной попытки подключения.

## *CORREDOR_MAX_RECEIVE_MESSAGE_SIZE*

### Type

`int`


### Description

Максимальный размер принимаемого сообщения.

## *CORREDOR_RUN_AS_ENABLED*

### Type

`bool`


## *CORREDOR_CLIENT_CERTIFICATES_CA*

### Type

`string`


### Default

```
```
ca.crt

## *CORREDOR_CLIENT_CERTIFICATES_ENABLED*

### Type

`bool`


## *CORREDOR_CLIENT_CERTIFICATES_PATH*

### Type

`string`


### Default

```
```
/certs/corredor/client

## *CORREDOR_CLIENT_CERTIFICATES_PRIVATE*

### Type

`string`


### Default

```
```
private.key

## *CORREDOR_CLIENT_CERTIFICATES_PUBLIC*

### Type

`string`


### Default

```
```
public.crt

## *CORREDOR_CLIENT_CERTIFICATES_SERVER_NAME*

### Type

`string`


# Environment

## *ENVIRONMENT*

### Type

`string`


### Default

```
```
production
# Events and scheduler

## *EVENTBUS_SCHEDULER_ENABLED*

### Type

`bool`


### Description

Включить планировщик eventbus.

## *EVENTBUS_SCHEDULER_INTERVAL*

### Type

`time.Duration`


### Default

```
```
60s
### Description

Установить интервал времени для планировщика `eventbus`.
# federation

## *FEDERATION_SYNC_DATA_MONITOR_INTERVAL*

### Type

`time.Duration`


### Default

```
```
1m
### Description

Задержка в секундах для синхронизации данных.

## *FEDERATION_SYNC_DATA_PAGE_SIZE*

### Type

`int`


### Default

```
```
100
### Description

Размер пакета при выборке для синхронизации данных.

## *FEDERATION_ENABLED*

### Type

`bool`


### Description

Федерация включена в системе; это включает эндпоинты REST API, возможность сопоставлять модули в Compose и синхронизироваться.

## *FEDERATION_HOST*

### Type

`string`


### Default

```
```
local.lowcode.org
### Description

Хост, используемый при сопряжении узлов, также включается в приглашение.

## *FEDERATION_LABEL*

### Type

`string`


### Default

```
```
federated
### Description

Метка федерации.

## *FEDERATION_SYNC_STRUCTURE_MONITOR_INTERVAL*

### Type

`time.Duration`


### Default

```
```
2m
### Description

Задержка в секундах для синхронизации структуры.

## *FEDERATION_SYNC_STRUCTURE_PAGE_SIZE*

### Type

`int`


### Default

```
```
1
### Description

Размер пакета при выборке для синхронизации структуры.
# Limits

## *LIMIT_SYSTEM_USERS*

### Type

`int`


### Description

Максимальное количество действующих (не удалённых, не заблокированных) пользователей.
# locale

## *LOCALE_DEVELOPMENT_MODE*

### Type

`bool`


### Description

При включении LowCoooode перезагружает языковые файлы при каждом запросе.
Включите для отладки или разработки.

## *LOCALE_LANGUAGES*

### Type

`string`


### Default

```
```
en
### Description

Список языков (языковых тегов), разделённых запятыми, для включения.
Если включённый язык не может быть загружен, ошибка логируется.

При загрузке языковых конфигураций (config.xml) из указанных путей.


## *LOCALE_LOG*

### Type

`bool`


### Description

Логировать события и действия, связанные с локалью.

## *LOCALE_PATH*

### Type

`string`


### Description

Один или несколько путей к файлам конфигурации локали и переводам, разделённых двоеточием.

При LOCALE*DEVELOPMENT*MODE=true значение по умолчанию — ../../locale.

## *LOCALE_QUERY_STRING_PARAM*

### Type

`string`


### Default

```
```
lng
### Description

Имя параметра строки запроса, используемого для передачи языкового тега (переопределяет заголовок Accept-Language).
Установите пустую строку, чтобы отключить определение из строки запроса.
Этот параметр игнорируется, если включён только один язык.


## *LOCALE_RESOURCE_TRANSLATIONS_ENABLED*

### Type

`bool`


### Description

При включении редактор переводов ресурсов становится доступен в интерфейсе.
# log

## *LOG_DEBUG*

### Type

`bool`


### Description

Отключает формат JSON для логирования и включает более читаемый вывод с цветами.

Отключите для производства.


## *LOG_FILTER*

### Type

`string`


### Description

Правила фильтрации логирования по уровню и имени (log-level:log-namespace).
Обратите внимание, что уровень (LOG_LEVEL) применяется до фильтра и влияет на конечный вывод!

Оставьте пустым для производства.

Пример:
`warn+:* **:auth,workflow.**`
Логировать предупреждения, ошибки, паники, фаталы. Всё из auth и workflow логируется.


Подробнее: https://github.com/moul/zapfilter


## *LOG_INCLUDE_CALLER*

### Type

`bool`


### Description

Установите true, чтобы видеть, откуда был вызван логгер.

Отключите для производства.


## *LOG_LEVEL*

### Type

`string`


### Default

```
```
warn
### Description

Минимальный уровень логирования. Если установлен "warn",
будут логироваться уровни warn, error, dpanic, panic и fatal.

Рекомендуемое значение для производства: warn.

Возможные значения: debug, info, warn, error, dpanic, panic, fatal.


## *LOG_STACKTRACE_LEVEL*

### Type

`string`


### Default

```
```
dpanic
### Description

Включать стек-трейс при логировании на указанном уровне или ниже.
Отключите для производства.

Возможные значения: debug, info, warn, error, dpanic, panic, fatal.

# Messaging queue

## *MESSAGEBUS_ENABLED*

### Type

`bool`


### Description

Включить messagebus.

## *MESSAGEBUS_LOG_ENABLED*

### Type

`bool`


### Description

Включить дополнительное логирование для наблюдателей messagebus.
# Monitoring

## *MONITOR_INTERVAL*

### Type

`time.Duration`


### Default

```
```
5m
### Description

Интервал вывода (логирования) для мониторинга.
# Object (file) storage

## *MINIO_ACCESS_KEY*

### Type

`string`


## *MINIO_BUCKET*

### Type

`string`


### Default

```
```
{component}
### Description

Плейсхолдер `component` заменяется на имя сервиса (например system).

## *MINIO_ENDPOINT*

### Type

`string`


## *MINIO_PATH_PREFIX*

### Type

`string`


### Description

Плейсхолдер `component` заменяется на имя сервиса (например system).

## *MINIO_SSEC_KEY*

### Type

`string`


## *MINIO_SECRET_KEY*

### Type

`string`


## *MINIO_SECURE*

### Type

`bool`


## *MINIO_STRICT*

### Type

`bool`


## *STORAGE_PATH*

### Type

`string`


### Default

```
```
var/store
### Description

Расположение, где хранятся загруженные файлы.
# Provisioning

## *PROVISION_ALWAYS*

### Type

`bool`


### Description

Управляет, должна ли выполняться настройка при запуске сервера.

## *PROVISION_PATH*

### Type

`string`


### Default

```
```
provision/*
### Description

Пути к файлам конфигурации для настройки, разделённые двоеточием.
# Sentry monitoring

## *SENTRY_DSN*

### Type

`string`


### Description

Установите для включения клиента Sentry.

## *SENTRY_ATTACH_STACKTRACE*

### Type

`bool`


### Description

Прикреплять стек-трейсы.

## *SENTRY_DEBUG*

### Type

`bool`


### Description

Выводить отладочную информацию.

## *SENTRY_DIST*

### Type

`string`


### Description

Установить сообщаемое распространение.

## *SENTRY_ENVIRONMENT*

### Type

`string`


### Description

Установить сообщаемое окружение.

## *SENTRY_MAX_BREADCRUMBS*

### Type

`int`


### Description

Максимальное количество breadcrumbs.

## *SENTRY_RELEASE*

### Type

`string`


### Description

Установить сообщаемый релиз.

## *SENTRY_SAMPLE_RATE*

### Type

`float64`


### Description

Коэффициент выборки для отправки событий (0.0 - 1.0, по умолчанию 1.0).

## *SENTRY_SERVERNAME*

### Type

`string`


### Description

Установить сообщаемое имя сервера.

## *SENTRY_WEBAPP_DSN*

### Type

`string`


### Description

Установите для включения клиента Sentry для веб-приложения.
# Rendering engine

## *TEMPLATE_RENDERER_GOTENBERG_ADDRESS*

### Type

`string`


### Description

Адрес контейнера рендеринга Gotenberg.

## *TEMPLATE_RENDERER_GOTENBERG_ENABLED*

### Type

`bool`


### Description

Включён ли контейнер рендеринга Gotenberg.
# Data store (database) upgrade

## *UPGRADE_ALWAYS*

### Type

`bool`


### Description

Управляет, должны ли обновляемые системы выполняться при запуске сервера.

## *UPGRADE_DEBUG*

### Type

`bool`


### Description

Включить/отключить отладочное логирование.
Для включения отладочного логирования установите `UPGRADE_DEBUG=true`.
# Delay system startup

## *WAIT_FOR*

### Type

`time.Duration`


### Description

Задерживает запуск API на указанное время (10s, 2m...).
Эта задержка происходит до проверки сервисов (`WAIT*FOR*SERVICES`).

## *WAIT_FOR_SERVICES*

### Type

`string`


### Description

Список хостов и/или URL для проверки, разделённый пробелами.
Формат хоста: `host` или `host:443` (порт по умолчанию 80).

!!! note
    Сервисы проверяются параллельно.


## *WAIT_FOR_SERVICES_PROBE_INTERVAL*

### Type

`time.Duration`


### Default

```
```
5s
### Description

Интервал между проверками сервисов.

## *WAIT_FOR_SERVICES_PROBE_TIMEOUT*

### Type

`time.Duration`


### Default

```
```
30s
### Description

Тайм-аут для каждой проверки сервиса.

## *WAIT_FOR_SERVICES_TIMEOUT*

### Type

`time.Duration`


### Default

```
```
1m
### Description

Максимальное время для каждой проверки сервиса.

## *WAIT_FOR_STATUS_PAGE*

### Type

`bool`


### Description

Показать временную веб-страницу состояния.
# Websocket server

## *WEBSOCKET_LOG_ENABLED*

### Type

`bool`


### Description

Включить дополнительное логирование для потоков аутентификации.

## *WEBSOCKET_PING_PERIOD*

### Type

`time.Duration`


### Default

```
```
119s

## *WEBSOCKET_PING_TIMEOUT*

### Type

`time.Duration`


### Default

```
```
120s

## *WEBSOCKET_TIMEOUT*

### Type

`time.Duration`


### Default

```
```
15s
### Description

Время до истечения тайм-аута `WsServer`.
# Workflow

## *WORKFLOW_CALL_STACK_SIZE*

### Type

`int`


### Description

Определяет максимальный размер стека вызовов между рабочими процессами.

## *WORKFLOW_EXEC_DEBUG*

### Type

`bool`


### Description

Включает подробное логирование выполнения рабочего процесса.

## *WORKFLOW_REGISTER*

### Type

`bool`


### Description

Регистрирует включённые и действующие рабочие процессы и выполняет их при срабатывании.

## *WORKFLOW_STACK_TRACE_ENABLED*

### Type

`bool`


### Description

Включает построение стек-трейса выполнения.

## *WORKFLOW_STACK_TRACE_FULL*

### Type

`bool`


### Description

Заставляет стек-трейс записывать все шаги.
# Discovery

## *DISCOVERY_BASE_URL*

### Type

`string`


### Description

Указывает хост сервера обнаружения lowcode.

## *DISCOVERY_LOWCODE_DOMAIN*

### Type

`string`


### Description

Указывает хост веб-приложения compose lowcode.

## *DISCOVERY_DEBUG*

### Type

`bool`


### Description

Включить информацию об активности обнаружения.

## *DISCOVERY_ENABLED*

### Type

`bool`


### Description

Включить эндпоинты обнаружения.
# attachment

## *AVATAR_INITIALS_BACKGROUND_COLOR*

### Type

`string`


### Default

```
```
#F3F3F3
### Description

Цвет фона инициалов аватара.

## *AVATAR_INITIALS_COLOR*

### Type

`string`


### Default

```
```
#162425
### Description

Цвет текста инициалов аватара.

## *AVATAR_INITIALS_FONT_PATH*

### Type

`string`


### Default

```
```
fonts/Poppins-Regular.ttf
### Description

Путь к файлу шрифта инициалов аватара.

## *ATTACHMENT_AVATAR_MAX_FILE_SIZE*

### Type

`int64`


### Description

Максимальный размер загрузки изображения аватара, значение по умолчанию — 1 МБ.
# webapp

## *WEBAPP_SCSS_DIR_PATH*

### Type

`string`


### Description

Путь к каталогу исходных файлов пользовательского SCSS.

:leveloffset: -1
