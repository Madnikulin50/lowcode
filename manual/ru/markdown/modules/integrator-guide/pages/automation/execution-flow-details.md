# Детали потока выполнения

В этом разделе подробно рассматривается, что запускает автоматизацию, какие шаги важны в потоке выполнения и что происходит с результатом автоматизации.
Упрощённое объяснение описано [здесь](modules/integrator-guide/pages/automation/automation/index.md#automation-flow-overview).

<a id="exec-flow-implicit"></a>
## Неявная автоматизация

Автоматизация считается **неявной**, если триггер использует событие `before` или `after`; например `beforeCreate` и `afterCreate`.

Неявная автоматизация выполняется автоматически при взаимодействии с конкретными ресурсами, такими как записи или пользователи.
Два варианта событий (`before` и `after`) следуют одному и тому же пути, с различием в том, когда событие отправляется и как используется результирующее значение.

События `before` отправляются первыми, и их результат выполнения **может** повлиять на остальное выполнение, тогда как события `after` отправляются после, и результат их выполнения **не влияет** на остальное выполнение.

.Диаграмма даёт абстрактный обзор потока выполнения неявной автоматизации. Поток подробно описан под диаграммой.
[plantuml,exec-flow-dig-implicit,svg,role="exec-flow-dig-implicit data-zoomable"]
@startuml
actor User as u
participant "LowCoooode component" as cc

u -> cc: Perform operation for resource

activate cc

cc -> cc: Access control check
cc -> cc: Content sanitization and validation
note over cc
Not all resources implement
sanitization and validation.
end note

box "Automation system"

participant "Event bus" as evtbs

cc -> evtbs: Dispatch the before event

activate evtbs
evtbs -> evtbs: Filter for registered triggers

participant "Automation" as aa

evtbs -> aa: Trigger execution

activate aa

aa -> aa: Execute automation
note over aa
The automation inherits the
invoking users' permissions
when trying to access LowCoooode.
end note

aa -> cc: Return the execution result
note over cc
If the automation is asynchronous
the component will not await
for the automation to complete.
end note

deactivate aa
deactivate evtbs

end box

cc -> cc: Process execution result
note over cc
Refer to the bellow description
for details on how the
result is handled.
end note

cc -> cc: Content sanitization and validation
note over cc
Not all resources implement
sanitization and validation.
===
The content is again sanitized
and validated as the automation
may change the values.
end note

cc -> cc: Complete the initial operation
note over cc
This normally changes the data
in the database, hence a
destructive operation.
end note

cc -> evtbs: Dispatch the after event

activate evtbs
evtbs -> evtbs: Filter for registered triggers

evtbs -> aa: Trigger execution

activate aa

aa -> aa: Execute automation
note over aa
The automation inherits the
invoking users' permissions
when trying to access LowCoooode.
end note

note over cc
The component does not
await for the after event to
complete.
end note

deactivate aa
deactivate evtbs

cc -> u: Response

deactivate cc
@enduml

.Детальное описание потока выполнения неявной автоматизации:
[cols="2s,4a"]
|===
| Компонент получает запрос на взаимодействие с ресурсом.
| Например; мы хотим создать новую запись или обновить email пользователя.

| Контроль доступа проверяет, разрешено ли вызывающему выполнить операцию.
| Если вызывающему не разрешено выполнять операцию, запрос отклоняется, и автоматизация не выполняется.

| Санация и валидация содержимого подготавливают и проверяют данные.
| Перед любой *разрушающей* операцией и перед отправкой события `before` значения проверяются и очищаются.
Если валидация не удаётся, запрос отклоняется, и автоматизация не выполняется.

| Событие `before` отправляется в `eventbus`.
| Если автоматизация **синхронная**, операция ожидает разрешения отправленного события, прежде чем продолжить исходную операцию.
Если автоматизация **асинхронная**, исходная операция не ожидает.

.Результирующее значение **синхронной** автоматизации влияет на то, как должна продолжаться исходная операция:
- Если выполнение завершилось успешно (без ошибок):
** Если результирующее значение задано (ненулевое), оно заменяет исходное значение, и исходная операция продолжается.
** Если результирующее значение не задано (нулевое), используется исходное значение, а любые изменения, внесённые рабочим процессом, игнорируются.
- Если выполнение завершилось с ошибкой (произошла ошибка), исходная операция отменяется, и ресурс остаётся нетронутым.

!!! caution
    Неудачное выполнение прекращает только текущую операцию и не откатывает изменения, которые были внесены в выполненной автоматизации.


| Происходит валидация и санация содержимого.
| Перед любой *разрушающей* операцией **изменённые значения** проверяются и очищаются.
Если валидация не удаётся, запрос отклоняется, и событие `after` **не** отправляется.

| Ресурс изменяется.
| Исходная операция завершается, и изменение окончательно записывается в хранилище.

| Событие `after` отправляется в `eventbus`.
| Операция **не ожидает** разрешения события, и результат автоматизации, таким образом, игнорируется.

!!! caution
    Результирующее значение автоматизации игнорируется и *не влияет* на исходную операцию.


|===

<a id="exec-flow-explicit"></a>
## Явная автоматизация

Автоматизация считается **явной**, если триггер использует событие `onManual` для любого ресурса.

!!! tip
    В некоторых случаях вы можете использовать явную автоматизацию, чтобы заменить необходимость в sink-маршрутах.
    
    *DevNote* добавить документацию по приведённому выше утверждению.


Явная автоматизация выполняется при ручном вызове.
Во фронтенде это обычно делается нажатием кнопки, но «под капотом» это просто HTTP-запрос к API-эндпоинту.

!!! caution
    Скрипты автоматизации и рабочие процессы определяют отдельные API-эндпоинты для вызова явной автоматизации.
    
    Для *скрипта автоматизации* это `POST $API/$COMPONENT/automation/trigger`, где `$API` — базовый URL вашего API LowCoooode, а `$COMPONENT` — имя компонента, который должен выполнить скрипт автоматизации (`compose` или `system`).
    
    Для *рабочих процессов* это `POST $API/automation/$WORKFLOW_ID/exec`, где `$API` — базовый URL вашего API LowCoooode, а `$WORKFLOW_ID` — ID рабочего процесса, который вы хотите выполнить.


.Диаграмма даёт абстрактный обзор потока выполнения явной автоматизации. Поток подробно описан под диаграммой.
[plantuml,exec-flow-dig-explicit,svg,role="exec-flow-dig-explicit data-zoomable"]
@startuml
actor User as u
participant "LowCoooode component" as cc

u -> cc: Request to execute an automation

activate cc

cc -> cc: Find the requested automation
cc -> cc: Access control
note over cc
Access control checks if the
user is allowed to execute
this automation.
===
This is workflow specific.
end note

box "Automation system"

participant "Automation" as aa

cc -> aa: Request to execute automation

activate aa

aa -> aa: Execute automation
note over aa
The automation inherits the
invoking users' permissions
when trying to access LowCoooode.
end note

aa -> cc: Return the execution result
note over cc
If the automation is asynchronous
the component will not await
for the automation to complete.
end note

deactivate aa

end box

cc -> u: Response

deactivate cc
@enduml

.Детальное описание потока выполнения явной автоматизации:
[cols="2s,4a"]
|===
| Компонент получает запрос на выполнение автоматизации.
| Например; пользователь CRM нажал кнопку отправки email-сообщения или инициации исходящего телефонного звонка.

| (#Специфично для рабочих процессов#) Механизм RBAC проверяет, разрешено ли вызывающему выполнить автоматизацию.
| Если вызывающему не разрешено выполнять рабочий процесс, запрос отклоняется.

!!! caution
    Пользовательский контекст безопасности (опция «выполнить от имени») *не влияет* на RBAC на этом этапе.
    Если рассматриваемая автоматизация была запущена другой автоматизацией, которая определила опцию «выполнить от имени», RBAC проверит указанного вызывающего.


| Автоматизация выполняется, и результаты возвращаются в виде стандартного HTTP-ответа.
| Результаты содержат вывод автоматизации, а также некоторые метаданные, такие как трассировки стека выполнения и сообщения об ошибках.
|===

<a id="exec-flow-deferred"></a>
## Отложенная автоматизация

Автоматизация считается **отложенной**, если триггер использует `onInterval` или `onTimestamp` событие.

Отложенная автоматизация вызывается системой на основе предоставленной информации о времени.
Выполнение не привязано к операции (например, к ручному вызову или событию).

«Под капотом» LowCoooode определяет тикер, который отправляет события `onInterval` и `onTimestamp` раз в минуту (может быть настроено через переменную `.env` [`EVENTBUS*SCHEDULER*INTERVAL`](modules/devops-guide/pages/references/configuration/server.md#*eventbus*scheduler_interval)).
Отправленные события затем используются для сопоставления и вызова любой автоматизации с соответствующими триггерами.

!!! note
    LowCoooode отправляет события интервала и метки времени для компонентов system и compose.
    Внутренне эти события одинаковы, но сохранены из соображений обратной совместимости.


!!! caution
    Отложенная автоматизация *требует* явно указать вызывающего пользователя, так как автоматизация выполняется системой, и мы не можем определить вызывающего пользователя.


.Диаграмма даёт абстрактный обзор потока выполнения отложенной автоматизации. Поток подробно описан под диаграммой.
[plantuml,exec-flow-dig-deferred,svg,role="exec-flow-dig-deferred data-zoomable"]
@startuml

participant "Scheduler" as sch

activate sch
note over sch
The scheduler is started at
boot time and terminated when
the server is shut down.
end note

box "Automation system"

participant "Event bus" as evtbs

sch -> evtbs: Dispatch event in intervals
note over sch
The default interval is 1min,
but can be configured.
end note

activate evtbs
evtbs -> evtbs: Filter for registered triggers

participant "Automation" as aa

evtbs -> aa: Trigger execution

activate aa

aa -> aa: Execute automation

deactivate aa
deactivate evtbs
deactivate sch

end box

@enduml

.Детальное описание потока выполнения отложенной автоматизации:
[cols="2s,4a"] 
|===
| Системный тикер отправляет событие `onInterval` и `onTimestamp`.
| События отправляются в eventbus.

| Автоматизация выполняется асинхронно, и результаты игнорируются.
| Значение выполнения автоматизации не влияет на другие автоматизации.

|===

<a id="exec-flow-sink"></a>
## Sink-автоматизация

Автоматизация считается **sink**, если триггер привязан к ресурсу `System Sink`.

Sink-автоматизация выполняется, когда системный компонент получает HTTP-запрос к API-эндпоинту `/sink`.
Выполнение не привязано к операции (например, к ручному вызову или событию).

!!! caution
    Sink-автоматизация *требует* явно указать вызывающего пользователя, так как автоматизация выполняется внешним пользователем, и мы не можем определить вызывающего пользователя.


.Диаграмма даёт абстрактный обзор потока выполнения sink-автоматизации. Поток подробно описан под диаграммой.
[plantuml,exec-flow-dig-sink,svg,role="exec-flow-dig-sink data-zoomable"]
@startuml

actor Unknown as u

box "LowCoooode system component"
participant "HTTP Router" as rr
participant "System service" as ss

activate rr
activate ss

u -> rr: HTTP request to /sink

note over u, rr #F5D380A5
Sink requests are validated with
a sink signature, not with the
standard auth token.
end note

rr -> ss: Request to handle

ss -> ss: Verify sink signature
note over ss
- existence,
- location,
- structure,
- validity.
end note

ss -> ss: Validate request with signature
note over ss
The signature claims are
used to validate the request
and the contents.
end note

end box

box "Automation system"

participant "Event bus" as evtbs

ss -> evtbs: Dispatch the onRequest event

activate evtbs
evtbs -> evtbs: Filter for registered triggers

participant "Automation" as aa

evtbs -> aa: Trigger execution

activate aa

aa -> aa: Execute automation
aa -> ss: Return the execution result
note over ss
If the automation is asynchronous
the component will not await
for the automation to complete.
end note

deactivate aa
deactivate evtbs

end box

ss -> ss: Prepare HTTP response
note over ss
Set headers, status, and payload.
end note

ss -> u: Response
note over u, ss
The system component bypasses the REST
layer when responding to sink requests.
end note

deactivate ss

@enduml

.Детальное описание потока выполнения sink-автоматизации:
[cols="2s,4a"]
|===
| Системный компонент получает HTTP-запрос к API-эндпоинту `/sink`.
| Например; `GET $API/system/sink/leads/?__sign=$SIGN`, где `$API` — базовый URL вашего API LowCoooode, а `$SIGN` — подпись sink-маршрута.

Обратитесь к [Руководству DevOps/sink-маршруты](modules/devops-guide/pages/sink-route.md) за сведениями о настройке.

| Подпись проверяется
| Сначала проверяется подпись sink, чтобы убедиться, что она не была изменена.

!!! note
    Подписи sink генерируются на лету на основе [JWT-секрета](modules/devops-guide/pages/references/configuration/server.md#_auth_jwt_secret) и *не хранятся* в базе данных.


| Подпись используется для проверки запроса.
| Система проверяет протокол, заголовки, происхождение и другие ограничения подписи.

| Событие `onRequest` отправляется в `eventbus`.
| Если автоматизация **синхронная**, операция ожидает разрешения отправленного события, прежде чем продолжить исходную операцию.
Если автоматизация **асинхронная**, исходная операция не ожидает.

.Результирующее значение **синхронной** автоматизации влияет на то, как сервер должен ответить:
- Если выполнение завершилось с ошибкой (произошла ошибка), сервер отвечает сообщением об ошибке и (опционально) трассировками стека ошибок и другими отладочными метаданными.
- Если выполнение завершилось успешно (без ошибок), сервер готовит ответ на основе результата выполнения:
** устанавливаются код состояния и заголовки ответа,
** тело ответа кодируется как байтовый поток и отправляется в качестве ответа.

!!! caution
    В настоящее время в качестве ответа можно использовать только строки и срезы байтов.

|===

<a id="exec-flow-email"></a>
## Email-автоматизация

Системный email; получение/отправка

!!! note
    *DevNote*: @todo
