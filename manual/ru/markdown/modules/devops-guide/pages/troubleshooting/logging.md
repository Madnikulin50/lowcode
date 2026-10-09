# Ведение журналов

Ведение журналов можно настроить в файле `.env` с помощью переменных `LOG_*`.

## Уровни журналирования

!!! note
    `LOG_DEBUG=true` действует как установка `LOG_LEVEL=ALL`


[cols="1s,5a"]
|===
| [#log-level-all]#[log-level-all,ALL](#log-level-all,ALL)#
|
Уровень `LOG_LEVEL=ALL` записывает все системные и пользовательские уровни журналирования.

| [#log-level-debug]#[log-level-debug,DEBUG](#log-level-debug,DEBUG)#
|
Уровень `LOG_LEVEL=DEBUG` записывает диагностическую информацию подробным, детальным образом.
Этот уровень следует использовать для целей тестирования, диагностики и устранения неполадок.

| [#log-level-info]#[log-level-info,INFO](#log-level-info,INFO)#
|
Уровень `LOG_LEVEL=INFO` записывает информативные выходные данные системы, такие как остановка или запуск сервисов.
Этот уровень можно использовать для описания того, что сделала конкретная задача, например, к каким данным был доступ и какие взаимодействия были выполнены.

| [#log-level-warn]#[log-level-warn,WARN](#log-level-warn,WARN)#
|
Уровень `LOG_LEVEL=WARN` записывает неожиданные проблемы, которые могут возникнуть, когда конкретному сервису неясно, что делать с входными данными, или предоставленные данные повреждены.
Этот уровень используется, когда система функционирует корректно, но на проблему следует обратить внимание.

| [#log-level-error]#[log-level-error,ERROR](#log-level-error,ERROR)#
|
Уровень `LOG_LEVEL=ERROR` записывает неожиданные проблемы, которые препятствуют правильному функционированию системы, например, невозможность доступа к файлам, другим сервисам LowCoooode или внешним сервисам.
Этот уровень используется, когда проблема блокирует правильное функционирование системы.
Её следует решать в приоритетном порядке.

| [#log-level-dpanic]#[log-level-dpanic,DPANIC](#log-level-dpanic,DPANIC)#
|
Зарезервирован; не используется.

| [#log-level-panic]#[log-level-panic,PANIC](#log-level-panic,PANIC)#
|
Зарезервирован; не используется.

| [#log-level-fatal]#[log-level-fatal,FATAL](#log-level-fatal,FATAL)#
|
Уровень `LOG_LEVEL=FATAL` записывает неожиданные проблемы, которые могут вызвать серьёзную проблему или повреждение.
Этот уровень используется для указания критического состояния системы, которое должно быть решено в приоритетном порядке.

| [#log-level-off]#[log-level-off,OFF](#log-level-off,OFF)#
|
Если обе переменные журналирования `.env` не установлены (`LOG*DEBUG` и `LOG*LEVEL`), журналы не записываются.
|===

## Переменные окружения:
[source,.env]
# Logging level we want to use (values: debug, info, warn, error, dpanic, panic, fatal)
# Minimise the logging level. If set to "warn", Levels warn, error, dpanic panic and fatal will be logged.
LOG_LEVEL=debug


# Disables JSON format for logging and enables more human-readable output with colors.
# Disable for production.
LOG_DEBUG=true/false

# Log filtering rules by level and name (log-level:log-namespace).
# Please note that level (LOG_LEVEL) is applied before filter and it affects the final output!
# Leave unset for production.
# Log warnings, errors, panic, fatals. Everything from workflow is logged.
# See more examples and documentation here: https://github.com/moul/zapfilter
LOG_FILTER=warn+:workflow.*

# Set to true to see where the logging was called from.
# Disable for production.
LOG*INCLUDE*CALLER=true/false

# Include stack-trace when logging at a specified level or below.
# Disable for production.
# Default value: "dpanic"
# Possible values: debug, info, warn, error, dpanic, panic, fatal
LOG*STACKTRACE*LEVEL


## Журналирование рабочих процессов

Шаги [функций рабочего процесса](modules/integrator-guide/pages/automation/workflows/index.md#functions) позволяют записывать сообщения изнутри рабочего процесса.

Убедитесь, что переменные `.env` установлены соответствующим образом.

!!! tip
    Вы можете использовать `LOG_FILTER`, чтобы показывать только те журналы, которые соответствуют заданному шаблону.
    
    [source,.env]
    ----
    # Log filtering rules by level and name (log-level:log-namespace).
    # Please note that level (LOG_LEVEL) is applied before filter and it affects the final output!
    # Leave unset for production.
    # Log warnings, errors, panic, fatals. Everything from workflow is logged.
    # LOG_FILTER={LOG_LEVEL}+:workflow.*
    LOG_FILTER=warn+:workflow.*
    
    # For colorful and human-readable output
    LOG_DEBUG=true
    ----


[cols="1s,5a"]
|===
| Функция | Конфигурация `.env`

| [#workflow-log-debug]#[workflow-log-debug,Записать отладочное сообщение](#workflow-log-debug,Записать отладочное сообщение)#
|
[source,.env]
LOG_DEBUG=true

# OR

LOG_LEVEL=debug

| [#workflow-log-info]#[workflow-log-info,Записать информационное сообщение](#workflow-log-info,Записать информационное сообщение)#
|
[source,.env]
LOG_DEBUG=true

# OR

LOG_LEVEL=debug
# OR
LOG_LEVEL=info

| [#workflow-log-warning]#[workflow-log-warning,Записать предупреждающее сообщение](#workflow-log-warning,Записать предупреждающее сообщение)#
|
[source,.env]
LOG_DEBUG=true

# OR

LOG_LEVEL=debug
# OR
LOG_LEVEL=info
# OR
LOG_LEVEL=warn

| [#workflow-log-error]#[workflow-log-error,Записать сообщение об ошибке](#workflow-log-error,Записать сообщение об ошибке)#
|
[source,.env]
LOG_DEBUG=true

# OR

LOG_LEVEL=debug
# OR
LOG_LEVEL=info
# OR
LOG_LEVEL=warn
# OR
LOG_LEVEL=error
|===
