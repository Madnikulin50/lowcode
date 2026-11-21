# Справочник функций

:leveloffset: +1


# `apigwBody`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-apigwbody-read]#[fnc-apigwbody-read,Read request body from integration gateway](#fnc-apigwbody-read,Read request body from integration gateway)#
| |
|
.Parameters:
- #*# `request`
(
   `HttpRequest`,)

.Results:
- body (`String`)

| [#fnc-apigwbody-readfile]#[fnc-apigwbody-readfile,Read file from integration gateway](#fnc-apigwbody-readfile,Read file from integration gateway)#
| |
|
.Parameters:
- #*# `request`
(
   `HttpRequest`,)
- #*# `name`
(
   `String`,)

.Results:
- file (`Reader`)
- fileName (`String`)
- exists (`Boolean`)

|===
# `corredor`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-corredor-exec]#[fnc-corredor-exec,Corredor automation script executor](#fnc-corredor-exec,Corredor automation script executor)#
| Выполняет скрипт на сервере автоматизации Corredor
|
.Parameters:
- #*# `script`
(
   `String`,)
- `args`
(
   `Vars`,)

.Results:
- results (`Vars`)

|===
# `email`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-email-send]#[fnc-email-send,Email](#fnc-email-send,Email)#
| Отправляет email напрямую
|
.Parameters:
- `Subject`
(
   `String`,
   `Reader`,)
- `Reply to`
(
   `String`,
   `User`,)
- `Sender`
(
   `String`,
   `User`,)
- `Recipients`
(
   `String`,
   `KV`,
   `User`,)
- `CC`
(
   `String`,
   `KV`,
   `User`,)
- `HTML message body`
(
   `String`,
   `Reader`,)
- `Plain text message body`
(
   `String`,
   `Reader`,)

| [#fnc-email-message]#[fnc-email-message,Email builder](#fnc-email-message,Email builder)#
| Создаёт новое email-сообщение из основных параметров без отправки
|
.Parameters:
- `Subject`
(
   `String`,
   `Reader`,)
- `Reply to`
(
   `String`,
   `User`,)
- `Sender`
(
   `String`,
   `User`,)
- `Recipients`
(
   `String`,
   `KV`,
   `User`,)
- `CC`
(
   `String`,
   `KV`,
   `User`,)
- `HTML message body`
(
   `String`,
   `Reader`,)
- `Plain text message body`
(
   `String`,
   `Reader`,)

.Results:
- message (`EmailMessage`)

| [#fnc-email-sendmessage]#[fnc-email-sendmessage,Email sender](#fnc-email-sendmessage,Email sender)#
| Отправляет email-сообщение
|
.Parameters:
- #*# `Message to be sent`
(
   `EmailMessage`,)

| [#fnc-email-setsubject]#[fnc-email-setsubject,Email subject](#fnc-email-setsubject,Email subject)#
| Устанавливает тему email-сообщения
|
.Parameters:
- #*# `Message to be sent`
(
   `EmailMessage`,)
- #*# `Subject`
(
   `String`,)

| [#fnc-email-setheaders]#[fnc-email-setheaders,Email headers](#fnc-email-setheaders,Email headers)#
| Устанавливает заголовки сообщения (переопределяет существующие заголовки, тему, получателей)
|
.Parameters:
- #*# `Message to be sent`
(
   `EmailMessage`,)
- #*# `Headers`
(
   `KVV`,)

| [#fnc-email-setheader]#[fnc-email-setheader,Email header](#fnc-email-setheader,Email header)#
| Устанавливает или удаляет определённый заголовок без изменения остальных
|
.Parameters:
- #*# `Message to be sent`
(
   `EmailMessage`,)
- #*# `Value`
(
   `String`,)
- `Value`
(
   `String`,)

| [#fnc-email-setaddress]#[fnc-email-setaddress,Email set address](#fnc-email-setaddress,Email set address)#
| Устанавливает адреса получателя, отправителя или ответа
|
.Parameters:
- #*# `Message to be sent`
(
   `EmailMessage`,)
- #*# `Type`
(
   `String`,)
- #*# `Address`
(
   `String`,)
- `Name`
(
   `String`,)

| [#fnc-email-addaddress]#[fnc-email-addaddress,Email add address](#fnc-email-addaddress,Email add address)#
| Добавляет нового получателя, отправителя или адрес ответа
|
.Parameters:
- #*# `Message to be sent`
(
   `EmailMessage`,)
- #*# `Type`
(
   `String`,)
- #*# `Address`
(
   `String`,)
- `Name`
(
   `String`,)

| [#fnc-email-attach]#[fnc-email-attach,Email attachment](#fnc-email-attach,Email attachment)#
| Прикрепляет содержимое к email-сообщению
|
.Parameters:
- #*# `Message to be sent`
(
   `EmailMessage`,)
- #*# `Content`
(
   `Reader`,
   `String`,)
- `Name`
(
   `String`,)

| [#fnc-email-embed]#[fnc-email-embed,Email embedded attachment](#fnc-email-embed,Email embedded attachment)#
| Встраивает файл (изображение) в email-сообщение
|
.Parameters:
- #*# `Message to be sent`
(
   `EmailMessage`,)
- #*# `Content`
(
   `Reader`,)
- `Name`
(
   `String`,)

|===
# `httpRequest`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-httprequest-send]#[fnc-httprequest-send,HTTP request](#fnc-httprequest-send,HTTP request)#
| Отправляет HTTP-запросы
|
.Parameters:
- #*# `url`
(
   `String`,)
- #*# `method`
(
   `String`,)
- `params`
(
   `KVV`,)
- `headers`
(
   `KVV`,)
- `headerAuthBearer`
(
   `String`,)
- `headerAuthUsername`
(
   `String`,)
- `headerAuthPassword`
(
   `String`,)
- `headerUserAgent`
(
   `String`,)
- `headerContentType`
(
   `String`,)
- `timeout`
(
   `Duration`,)
- `form`
(
   `KVV`,)
- `body`
(
   `String`,
   `Reader`,
   `Any`,)

.Results:
- status (`String`)
- statusCode (`Integer`)
- headers (`KVV`)
- contentLength (`Integer`)
- contentType (`String`)
- body (`Reader`)

|===
# `jsenv`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-jsenv-execute]#[fnc-jsenv-execute,Process arbitrary data in jsenv](#fnc-jsenv-execute,Process arbitrary data in jsenv)#
| |
|
.Parameters:
- #*# `scope`
(
   `Any`,
   `Reader`,)
- #*# `source`
(
   `String`,)

.Results:
- resultString (`String`)
- resultInt (`Integer`)
- resultBool (`Boolean`)
- resultAny (`Any`)

|===
# `jwt`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-jwt-generate]#[fnc-jwt-generate,Generate JWT](#fnc-jwt-generate,Generate JWT)#
| |
|
.Parameters:
- `scope`
(
   `String`,)
- #*# `header`
(
   `Vars`,
   `String`,)
- #*# `payload`
(
   `Vars`,
   `String`,)
- #*# `secret`
(
   `String`,
   `Reader`,)

.Results:
- token (`String`)

|===
# `log`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-log-debug]#[fnc-log-debug,Log debug message](#fnc-log-debug,Log debug message)#
| |
|
.Parameters:
- #*# `message`
(
   `String`,)
- `fields`
(
   `KV`,)

| [#fnc-log-info]#[fnc-log-info,Log info message](#fnc-log-info,Log info message)#
| |
|
.Parameters:
- #*# `message`
(
   `String`,)
- `fields`
(
   `KV`,)

| [#fnc-log-warn]#[fnc-log-warn,Log warning message](#fnc-log-warn,Log warning message)#
| |
|
.Parameters:
- #*# `message`
(
   `String`,)
- `fields`
(
   `KV`,)

| [#fnc-log-error]#[fnc-log-error,Log error message](#fnc-log-error,Log error message)#
| |
|
.Parameters:
- #*# `message`
(
   `String`,)
- `fields`
(
   `KV`,)

|===
# `loop`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

|===
# `oauth2`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-oauth2-authenticate]#[fnc-oauth2-authenticate,Authentication: OAUTH2](#fnc-oauth2-authenticate,Authentication: OAUTH2)#
| |
|
.Parameters:
- #*# `client`
(
   `String`,)
- #*# `secret`
(
   `String`,)
- #*# `scope`
(
   `String`,)
- #*# `tokenUrl`
(
   `String`,)

.Results:
- accessToken (`String`)
- refreshToken (`String`)
- token (`Any`)

|===
# `queue`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-queue-write]#[fnc-queue-write,Queue message send](#fnc-queue-write,Queue message send)#
| |
|
.Parameters:
- #*# `payload`
(
   `String`,
   `Reader`,)
- #*# `queue`
(
   `String`,)

|===
# `attachment`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-attachment-lookup]#[fnc-attachment-lookup,Attachment lookup](#fnc-attachment-lookup,Attachment lookup)#
| Найти конкретное вложение по ID
|
.Parameters:
- #*# `attachment`
(
   `ID`,)

.Results:
- attachment (`Attachment`)

| [#fnc-attachment-create]#[fnc-attachment-create,Create file and attach it to a resource](#fnc-attachment-create,Create file and attach it to a resource)#
| |
|
.Parameters:
- `name`
(
   `String`,)
- #*# `resource`
(
   `ComposeRecord`,)
- `fieldName`
(
   `String`,)
- #*# `content`
(
   `String`,
   `Reader`,
   `Bytes`,)

.Results:
- attachment (`Attachment`)

| [#fnc-attachment-delete]#[fnc-attachment-delete,Delete attachment](#fnc-attachment-delete,Delete attachment)#
| |
|
.Parameters:
- #*# `attachment`
(
   `ID`,)

| [#fnc-attachment-openoriginal]#[fnc-attachment-openoriginal,Open original attachment](#fnc-attachment-openoriginal,Open original attachment)#
| |
|
.Parameters:
- #*# `attachment`
(
   `ID`,
   `Attachment`,)

.Results:
- content (`Reader`)

| [#fnc-attachment-openpreview]#[fnc-attachment-openpreview,Open attachment preview](#fnc-attachment-openpreview,Open attachment preview)#
| |
|
.Parameters:
- #*# `attachment`
(
   `ID`,
   `Attachment`,)

.Results:
- content (`Reader`)

|===
# `modules`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-modules-lookup]#[fnc-modules-lookup,Compose module lookup](#fnc-modules-lookup,Compose module lookup)#
| Найти конкретный модуль по ID или handle
|
.Parameters:
- #*# `module`
(
   `ID`,
   `Handle`,
   `ComposeModule`,)
- #*# `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)

.Results:
- module (`ComposeModule`)

|===
# `namespaces`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-namespaces-lookup]#[fnc-namespaces-lookup,Compose namespace lookup](#fnc-namespaces-lookup,Compose namespace lookup)#
| Найти конкретное пространство имён по ID или handle
|
.Parameters:
- #*# `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)

.Results:
- namespace (`ComposeNamespace`)

|===
# `notification`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-notification-sendrecord]#[fnc-notification-sendrecord,Send record notification](#fnc-notification-sendrecord,Send record notification)#
| Отправляет уведомление со ссылкой на конкретную запись
|
.Parameters:
- #*# `recipient`
(
   `ID`,
   `Handle`,
   `String`,)
- #*# `title`
(
   `String`,)
- `description`
(
   `String`,)
- #*# `module`
(
   `ID`,
   `Handle`,)
- #*# `namespace`
(
   `ID`,
   `Handle`,)
- `record`
(
   `ID`,)
- `openMode`
(
   `String`,)
- `edit`
(
   `Boolean`,)

|===
# `records`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-records-lookup]#[fnc-records-lookup,Compose record lookup](#fnc-records-lookup,Compose record lookup)#
| Найти конкретную запись по ID
|
.Parameters:
- #*# `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)
- #*# `Module to set record type`
(
   `ID`,
   `Handle`,
   `ComposeModule`,)
- #*# `record`
(
   `ID`,
   `ComposeRecord`,)

.Results:
- record (`ComposeRecord`)

| [#fnc-records-search]#[fnc-records-search,Compose records search](#fnc-records-search,Compose records search)#
| |
|
.Parameters:
- #*# `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)
- #*# `Module to set record type`
(
   `ID`,
   `Handle`,
   `ComposeModule`,)
- `query`
(
   `String`,)
- `meta`
(
   `Meta`,)
- `deleted`
(
   `UnsignedInteger`,)
- `sort`
(
   `String`,)
- `limit`
(
   `UnsignedInteger`,)
- `incTotal`
(
   `Boolean`,)
- `incPageNavigation`
(
   `Boolean`,)
- `pageCursor`
(
   `String`,)

.Results:
- records (`ComposeRecord`)
- Total records found (`UnsignedInteger`)
- nextPage (`String`)
- prevPage (`String`)
- pageNavigation (`Array`)

| [#fnc-records-first]#[fnc-records-first,Compose record lookup (oldest)](#fnc-records-first,Compose record lookup (oldest))#
| |
|
.Parameters:
- #*# `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)
- #*# `Module to set record type`
(
   `ID`,
   `Handle`,
   `ComposeModule`,)

.Results:
- record (`ComposeRecord`)

| [#fnc-records-last]#[fnc-records-last,Compose record lookup (newest)](#fnc-records-last,Compose record lookup (newest))#
| |
|
.Parameters:
- #*# `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)
- #*# `Module to set record type`
(
   `ID`,
   `Handle`,
   `ComposeModule`,)

.Results:
- record (`ComposeRecord`)

| [#fnc-records-new]#[fnc-records-new,Compose record maker](#fnc-records-new,Compose record maker)#
| Создаёт новый экземпляр записи compose без сохранения
|
.Parameters:
- #*# `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)
- #*# `Module to set record type`
(
   `ID`,
   `Handle`,
   `ComposeModule`,)

.Results:
- record (`ComposeRecord`)

| [#fnc-records-validate]#[fnc-records-validate,Compose record validator](#fnc-records-validate,Compose record validator)#
| |
|
.Parameters:
- #*# `record`
(
   `ComposeRecord`,)

.Results:
- Set to true when record is valid (`Boolean`)

| [#fnc-records-create]#[fnc-records-create,Compose record create](#fnc-records-create,Compose record create)#
| |
|
.Parameters:
- #*# `record`
(
   `ComposeRecord`,)

.Results:
- record (`ComposeRecord`)

| [#fnc-records-update]#[fnc-records-update,Compose record update](#fnc-records-update,Compose record update)#
| |
|
.Parameters:
- #*# `record`
(
   `ComposeRecord`,)

.Results:
- record (`ComposeRecord`)

| [#fnc-records-delete]#[fnc-records-delete,Compose record delete](#fnc-records-delete,Compose record delete)#
| |
|
.Parameters:
- `Module to set record type`
(
   `ID`,
   `Handle`,
   `ComposeModule`,)
- `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)
- #*# `record`
(
   `ID`,
   `ComposeRecord`,)

| [#fnc-records-report]#[fnc-records-report,Report](#fnc-records-report,Report)#
| Отчёт по записям Compose
|
.Parameters:
- #*# `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)
- #*# `Module to set record type`
(
   `ID`,
   `Handle`,
   `ComposeModule`,)
- #*# `Metrics for records report`
(
   `String`,)
- #*# `Dimensons for records report`
(
   `String`,)
- #*# `Filter for records report`
(
   `String`,)

.Results:
- Complex structure holding complete records report (`Any`)

| [#fnc-records-clone]#[fnc-records-clone,Compose record cloner](#fnc-records-clone,Compose record cloner)#
| Создаёт копию существующей записи
|
.Parameters:
- #*# `namespace`
(
   `ID`,
   `Handle`,
   `ComposeNamespace`,)
- #*# `Module to set record type`
(
   `ID`,
   `Handle`,
   `ComposeModule`,)
- #*# `record`
(
   `ID`,
   `ComposeRecord`,)

.Results:
- record (`ComposeRecord`)

|===
# `actionlog`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-actionlog-search]#[fnc-actionlog-search,Action log search](#fnc-actionlog-search,Action log search)#
| |
|
.Parameters:
- `fromTimestamp`
(
   `DateTime`,)
- `toTimestamp`
(
   `DateTime`,)
- `beforeActionID`
(
   `ID`,)
- `actorID`
(
   `ID`,)
- `origin`
(
   `String`,)
- `resource`
(
   `String`,)
- `action`
(
   `String`,)
- `limit`
(
   `UnsignedInteger`,)

.Results:
- actions (`Action`)

| [#fnc-actionlog-record]#[fnc-actionlog-record,Record action into action log](#fnc-actionlog-record,Record action into action log)#
| |
|
.Parameters:
- `action`
(
   `String`,)
- `resource`
(
   `String`,)
- `error`
(
   `String`,)
- `severity`
(
   `String`,)
- `description`
(
   `String`,)
- `meta`
(
   `Vars`,)

|===
# `notification`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-notification-send]#[fnc-notification-send,Send simple notification](#fnc-notification-send,Send simple notification)#
| Отправляет простое уведомление с заголовком и описанием пользователю
|
.Parameters:
- #*# `recipient`
(
   `ID`,
   `Handle`,
   `String`,)
- #*# `title`
(
   `String`,)
- `description`
(
   `String`,)

|===
# `rbac`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-rbac-allow]#[fnc-rbac-allow,RBAC: Allow operation on resource to a role](#fnc-rbac-allow,RBAC: Allow operation on resource to a role)#
| |
|
.Parameters:
- #*# `resource`
(
   `RbacResource`,)
- #*# `role`
(
   `ID`,
   `Handle`,
   `Role`,)
- #*# `operation`
(
   `String`,)

| [#fnc-rbac-deny]#[fnc-rbac-deny,RBAC: Deny operation on resource to a role](#fnc-rbac-deny,RBAC: Deny operation on resource to a role)#
| |
|
.Parameters:
- #*# `resource`
(
   `RbacResource`,)
- #*# `role`
(
   `ID`,
   `Handle`,
   `Role`,)
- #*# `operation`
(
   `String`,)

| [#fnc-rbac-inherit]#[fnc-rbac-inherit,RBAC: Remove allow/deny operation of a role from resource](#fnc-rbac-inherit,RBAC: Remove allow/deny operation of a role from resource)#
| |
|
.Parameters:
- #*# `resource`
(
   `RbacResource`,)
- #*# `role`
(
   `ID`,
   `Handle`,
   `Role`,)
- #*# `operation`
(
   `String`,)

| [#fnc-rbac-check]#[fnc-rbac-check,RBAC: Can user perform an operation on a resource](#fnc-rbac-check,RBAC: Can user perform an operation on a resource)#
| |
|
.Parameters:
- #*# `resource`
(
   `RbacResource`,)
- #*# `operation`
(
   `String`,)
- `user`
(
   `User`,)

.Results:
- can (`Boolean`)

|===
# `roles`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-roles-lookup]#[fnc-roles-lookup,Role lookup](#fnc-roles-lookup,Role lookup)#
| Найти конкретную роль по ID или handle
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Role`,)

.Results:
- role (`Role`)

| [#fnc-roles-searchmembers]#[fnc-roles-searchmembers,Role members search](#fnc-roles-searchmembers,Role members search)#
| Найти участников для конкретной роли по ID или handle
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Role`,)

.Results:
- users (`User`)
- total (`UnsignedInteger`)

| [#fnc-roles-addmember]#[fnc-roles-addmember,Role membership add](#fnc-roles-addmember,Role membership add)#
| |
|
.Parameters:
- #*# `role`
(
   `ID`,
   `Handle`,
   `Role`,)
- #*# `user`
(
   `ID`,
   `Handle`,
   `String`,
   `User`,)

| [#fnc-roles-removemember]#[fnc-roles-removemember,Role membership remove](#fnc-roles-removemember,Role membership remove)#
| |
|
.Parameters:
- #*# `role`
(
   `ID`,
   `Handle`,
   `Role`,)
- #*# `user`
(
   `ID`,
   `Handle`,
   `String`,
   `User`,)

| [#fnc-roles-search]#[fnc-roles-search,Roles search](#fnc-roles-search,Roles search)#
| |
|
.Parameters:
- `query`
(
   `String`,)
- `memberID`
(
   `ID`,)
- `handle`
(
   `String`,)
- `name`
(
   `String`,)
- `labels`
(
   `KV`,)
- `deleted`
(
   `UnsignedInteger`,)
- `archived`
(
   `UnsignedInteger`,)
- `sort`
(
   `String`,)
- `limit`
(
   `UnsignedInteger`,)
- `incTotal`
(
   `Boolean`,)
- `incPageNavigation`
(
   `Boolean`,)
- `pageCursor`
(
   `String`,)

.Results:
- roles (`Role`)
- total (`UnsignedInteger`)

| [#fnc-roles-create]#[fnc-roles-create,Role creator](#fnc-roles-create,Role creator)#
| |
|
.Parameters:
- #*# `role`
(
   `Role`,)

.Results:
- role (`Role`)

| [#fnc-roles-update]#[fnc-roles-update,Role update](#fnc-roles-update,Role update)#
| |
|
.Parameters:
- #*# `role`
(
   `Role`,)

.Results:
- role (`Role`)

| [#fnc-roles-delete]#[fnc-roles-delete,Role delete](#fnc-roles-delete,Role delete)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Role`,)

| [#fnc-roles-recover]#[fnc-roles-recover,Role recover](#fnc-roles-recover,Role recover)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Role`,)

| [#fnc-roles-archive]#[fnc-roles-archive,Role archive](#fnc-roles-archive,Role archive)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Role`,)

| [#fnc-roles-unarchive]#[fnc-roles-unarchive,Role unarchive](#fnc-roles-unarchive,Role unarchive)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Role`,)

|===
# `templates`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-templates-lookup]#[fnc-templates-lookup,Template lookup](#fnc-templates-lookup,Template lookup)#
| Найти конкретный шаблон по ID или handle
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Template`,)

.Results:
- template (`Template`)

| [#fnc-templates-search]#[fnc-templates-search,Templates search](#fnc-templates-search,Templates search)#
| |
|
.Parameters:
- `handle`
(
   `String`,)
- `type`
(
   `String`,)
- `ownerID`
(
   `ID`,)
- `partial`
(
   `Boolean`,)
- `labels`
(
   `KV`,)
- `sort`
(
   `String`,)
- `limit`
(
   `UnsignedInteger`,)
- `incTotal`
(
   `Boolean`,)
- `incPageNavigation`
(
   `Boolean`,)
- `pageCursor`
(
   `String`,)

.Results:
- templates (`Template`)
- total (`UnsignedInteger`)

| [#fnc-templates-create]#[fnc-templates-create,Template create](#fnc-templates-create,Template create)#
| |
|
.Parameters:
- #*# `template`
(
   `Template`,)

.Results:
- template (`Template`)

| [#fnc-templates-update]#[fnc-templates-update,Template update](#fnc-templates-update,Template update)#
| |
|
.Parameters:
- #*# `template`
(
   `Template`,)

.Results:
- template (`Template`)

| [#fnc-templates-delete]#[fnc-templates-delete,Template delete](#fnc-templates-delete,Template delete)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Template`,)

| [#fnc-templates-recover]#[fnc-templates-recover,Template recover](#fnc-templates-recover,Template recover)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Template`,)

| [#fnc-templates-render]#[fnc-templates-render,Template render](#fnc-templates-render,Template render)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `Template`,)
- `documentName`
(
   `String`,)
- `documentType`
(
   `String`,)
- `variables`
(
   `Vars`,)
- `options`
(
   `RenderOptions`,)

.Results:
- document (`RenderedDocument`)

|===
# `users`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-users-lookup]#[fnc-users-lookup,User lookup](#fnc-users-lookup,User lookup)#
| Найти конкретного пользователя по ID, handle или строке
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `String`,
   `User`,)

.Results:
- user (`User`)

| [#fnc-users-searchmembership]#[fnc-users-searchmembership,User role search](#fnc-users-searchmembership,User role search)#
| Найти членство пользователя в роли по ID, handle или строке
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `String`,
   `User`,)

.Results:
- roles (`Role`)
- total (`UnsignedInteger`)

| [#fnc-users-checkmembership]#[fnc-users-checkmembership,User membership check](#fnc-users-checkmembership,User membership check)#
| Найти членство пользователя в роли по ID, handle или строке
|
.Parameters:
- #*# `user`
(
   `ID`,
   `Handle`,
   `String`,
   `User`,)
- #*# `role`
(
   `ID`,
   `Handle`,
   `Role`,)

.Results:
- member (`Boolean`)

| [#fnc-users-search]#[fnc-users-search,User search](#fnc-users-search,User search)#
| |
|
.Parameters:
- `query`
(
   `String`,)
- `email`
(
   `String`,)
- `handle`
(
   `String`,)
- `labels`
(
   `KV`,)
- `deleted`
(
   `UnsignedInteger`,)
- `suspended`
(
   `UnsignedInteger`,)
- `sort`
(
   `String`,)
- `limit`
(
   `UnsignedInteger`,)
- `incTotal`
(
   `Boolean`,)
- `incPageNavigation`
(
   `Boolean`,)
- `pageCursor`
(
   `String`,)

.Results:
- users (`User`)
- total (`UnsignedInteger`)

| [#fnc-users-create]#[fnc-users-create,User create](#fnc-users-create,User create)#
| |
|
.Parameters:
- #*# `user`
(
   `User`,)

.Results:
- user (`User`)

| [#fnc-users-update]#[fnc-users-update,User update](#fnc-users-update,User update)#
| |
|
.Parameters:
- #*# `user`
(
   `User`,)

.Results:
- user (`User`)

| [#fnc-users-delete]#[fnc-users-delete,User delete](#fnc-users-delete,User delete)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `String`,
   `User`,)

| [#fnc-users-recover]#[fnc-users-recover,User recover](#fnc-users-recover,User recover)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `String`,
   `User`,)

| [#fnc-users-suspend]#[fnc-users-suspend,User suspend](#fnc-users-suspend,User suspend)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `String`,
   `User`,)

| [#fnc-users-unsuspend]#[fnc-users-unsuspend,User unsuspend](#fnc-users-unsuspend,User unsuspend)#
| |
|
.Parameters:
- #*# `lookup`
(
   `ID`,
   `Handle`,
   `String`,
   `User`,)

|===
# `valuestore`

[cols="2m,4a,3a"]
|===
| Name | Description | I/O

| [#fnc-valuestore-env]#[fnc-valuestore-env,Get ENV variable](#fnc-valuestore-env,Get ENV variable)#
| Получает переменную окружения для указанного ключа. Если ключ не соответствует никакому значению, возвращается nil.

Чтобы избежать утечки конфиденциальной информации, эта функция разрешает доступ только к предопределённому набору переменных окружения.
Чтобы расширить предопределённый список переменных окружения, любая переменная с префиксом `LOWCODE_ENV` также доступна.
Эти ключи не чувствительны к регистру. Префикс пользовательских переменных окружения сохраняется, то есть `LOWCODE*ENV*TEST` остаётся `LOWCODE*ENV*TEST`.

.Список предопределённых переменных окружения:
- `NAME`: Имя окружения.
- `IS-DEVELOPMENT`: Указывает, установлено ли окружение в `dev`. Если нужен более точный контроль, используйте `NAME` напрямую.
- `IS-TEST`: Указывает, установлено ли окружение в `test`. Если нужен более точный контроль, используйте `NAME` напрямую.
- `IS-PRODUCTION`: Указывает, что окружение не является ни development, ни test. Если нужен более точный контроль, используйте `NAME` напрямую.
- `VERSION`: Указывает версию LowCoooode, работающую в данный момент.
- `BUILD-TIME`: Указывает время сборки работающего экземпляра LowCoooode.
- `AUTH.BASE-URL`: Указывает базовый URL сервиса аутентификации.
- `AUTH.DOMAIN`: Указывает домен сервиса аутентификации.
- `API.BASE-URL`: Указывает базовый URL экземпляра LowCoooode.
- `API.DOMAIN`: Указывает домен API экземпляра LowCoooode.
- `WEBAPP.BASE-URL`: Указывает базовый URL веб-приложений.
- `WEBAPP.DOMAIN`: Указывает домен веб-приложений.
- `WEBAPP.BASE-URL.COMPOSE`: Указывает базовый URL веб-приложения COMPOSE, если оно включено в переменной окружения `HTTP*WEBAPP*LIST`.
- `WEBAPP.BASE-URL.ADMIN`: Указывает базовый URL веб-приложения ADMIN, если оно включено в переменной окружения `HTTP*WEBAPP*LIST`.
- `WEBAPP.BASE-URL.WORKFLOW`: Указывает базовый URL веб-приложения WORKFLOW, если оно включено в переменной окружения `HTTP*WEBAPP*LIST`.
- `WEBAPP.BASE-URL.REPORTER`: Указывает базовый URL веб-приложения REPORTER, если оно включено в переменной окружения `HTTP*WEBAPP*LIST`.
- `WEBAPP.BASE-URL.DISCOVERY`: Указывает базовый URL веб-приложения DISCOVERY, если оно включено в переменной окружения `HTTP*WEBAPP*LIST`.

|
.Parameters:
- #*# `key`
(
   `String`,)

.Results:
- value (`Any`)

|===

:leveloffset: -1
