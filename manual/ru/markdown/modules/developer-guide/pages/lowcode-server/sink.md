# Маршруты приёмника

Особый тип ресурса, поддерживаемый LowCoooode, — это [системный приёмник (`system:sink`)](modules/integrator-guide/pages/automation/automation-scripts/references/resource-events.md#_systemsink), который используется для **ответа на API-запросы**.
Вы можете использовать маршрут приёмника для **реализации вебхуков**; например, потока OAuth.

## Модель безопасности

### Аутентификация с подписью

Каждый URL приёмника **должен быть подписан** в целях безопасности.
Подпись генерируется на основе параметров (**путь** и **ограничения**) и снабжается солью.

Обратитесь к [команде CLI](modules/devops-guide/pages/references/cli-reference.md#sink-signature) за подробностями.

### Контекст безопасности скрипта автоматизации

Когда HTTP-запрос запускает скрипт, мы **не можем определить**, кто является **вызывающим пользователем**.
Из-за этого вам **нужно указать** вызывающего пользователя для контекста безопасности (параметр `runAs`).

.Диаграмма описывает полный жизненный цикл события от его вызова до выполнения скрипта.
[plantuml,diagram-name-here,svg,role=sink-request]
@startuml
skinparam ParticipantPadding 20
skinparam BoxPadding 200
skinparam SequenceArrowThickness 2
skinparam RoundCorner 10

actor "External service" as Ext
participant "HTTP request handler" as HttpReqProc
participant "Sink processor" as SinkProc
participant "Eventbus"
participant "EventHandler"
participant "Corredor"

activate Ext
Ext -> HttpReqProc: HTTP request

group Internal
    activate HttpReqProc
    activate SinkProc
    HttpReqProc -> SinkProc: Resolved Sink request

    activate Eventbus
    SinkProc -> Eventbus: Event raised

    activate EventHandler
    Eventbus -> EventHandler: Trigger

    group Automation script
        EventHandler <-> Corredor: Script execution
    end

    Eventbus <- EventHandler: Response
    deactivate EventHandler

    SinkProc <- Eventbus: Response
    deactivate Eventbus

    HttpReqProc <- SinkProc: Response
    deactivate SinkProc
end


HttpReqProc -> Ext: Response
deactivate HttpReqProc
deactivate Ext

@enduml

## Обработчик HTTP-запросов

Обработчик HTTP-запросов **проверяет запрос** и **преобразует** его в **запрос приёмника**.

.Схема потока:
1. проверьте, предоставлена ли подпись,
1. проверьте, действительна ли подпись,
1. проверьте, **соответствуют ли применяемые ограничения** параметрам запроса:
** HTTP-метод,
** content-type,
** время истечения,
** максимальный размер тела и так далее.

Если вышеуказанная **проверка пройдена**, запрос **становится запросом приёмника** и обрабатывается как любое другое событие.

## Обработчик приёмника

Обработчик приёмника берёт HTTP-запрос и преобразует его **в событие**, которое может запустить скрипт автоматизации на сервере Corredor.

Самое важное, что нужно **здесь отметить**, — это то, что есть **небольшие отклонения** в зависимости от **content-type запроса**.
Когда запрос **указывает на электронное письмо** (`message/rfc822`, `rfc822`, `email` или `mail` в качестве `Content-Type`), возникает событие OnReceive [системной почты (`system:mail`)](modules/integrator-guide/pages/automation/automation-scripts/references/resource-events.md#_systemmail).
В любом другом случае возникает событие OnRequest [системного приёмника (`system:sink`)](modules/integrator-guide/pages/automation/automation-scripts/references/resource-events.md#_systemsink).

Обработчик приёмника также **формирует правильный ответ** (заголовки и тело) **на основе запроса**.
