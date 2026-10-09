# Вычисление контекстных ролей

Обратитесь к [Expr](integrator-guide//expr/index.md) за подробностями о написании выражений контекстных ролей.

Если выражение возвращает `true`, контекстная роль применяется.

.Общие переменные:
[cols="1s,5a"]
|===
| [#ctx-eval-var-general-userID]#[ctx-eval-var-general-userID,userID](#ctx-eval-var-general-userID,userID)#
|
Идентификатор текущего пользователя.

|===

## Записи

.Переменные записей:
[cols="1s,5a"]
|===
| [#ctx-eval-var-record-ID]#[ctx-eval-var-record-ID,resource.ID](#ctx-eval-var-record-ID,resource.ID)#
|
Идентификатор ресурса.

| [#ctx-eval-var-record-recordID]#[ctx-eval-var-record-recordID,resource.recordID](#ctx-eval-var-record-recordID,resource.recordID)#
|
Идентификатор ресурса; то же, что и [ctx-eval-var-record-ID,resource.ID](#ctx-eval-var-record-ID,resource.ID).

| [#ctx-eval-var-record-moduleID]#[ctx-eval-var-record-moduleID,resource.moduleID](#ctx-eval-var-record-moduleID,resource.moduleID)#
|
Идентификатор связанного модуля.

| [#ctx-eval-var-record-labels]#[ctx-eval-var-record-labels,resource.labels](#ctx-eval-var-record-labels,resource.labels)#
|
Пара «ключ-значение» меток, связанных с этим ресурсом.

| [#ctx-eval-var-record-namespaceID]#[ctx-eval-var-record-namespaceID,resource.namespaceID](#ctx-eval-var-record-namespaceID,resource.namespaceID)#
|
Идентификатор связанного пространства имён.

| [#ctx-eval-var-record-ownedBy]#[ctx-eval-var-record-ownedBy,resource.ownedBy](#ctx-eval-var-record-ownedBy,resource.ownedBy)#
|
Идентификатор пользователя-владельца ресурса.

| [#ctx-eval-var-record-createdAt]#[ctx-eval-var-record-createdAt,resource.createdAt](#ctx-eval-var-record-createdAt,resource.createdAt)#
|
Временная метка создания ресурса.

| [#ctx-eval-var-record-createdBy]#[ctx-eval-var-record-createdBy,resource.createdBy](#ctx-eval-var-record-createdBy,resource.createdBy)#
|
Идентификатор пользователя, создавшего ресурс.

| [#ctx-eval-var-record-updatedAt]#[ctx-eval-var-record-updatedAt,resource.updatedAt](#ctx-eval-var-record-updatedAt,resource.updatedAt)#
|
Временная метка последнего обновления ресурса.
Параметр не определён, если ресурс ещё не обновлялся.

| [#ctx-eval-var-record-updatedBy]#[ctx-eval-var-record-updatedBy,resource.updatedBy](#ctx-eval-var-record-updatedBy,resource.updatedBy)#
|
Идентификатор пользователя, который последним обновил ресурс.
Параметр будет равен `0`, если ресурс ещё не обновлялся.

| [#ctx-eval-var-record-deletedAt]#[ctx-eval-var-record-deletedAt,resource.deletedAt](#ctx-eval-var-record-deletedAt,resource.deletedAt)#
|
Временная метка удаления ресурса.
Параметр не определён, если ресурс не удалялся.

| [#ctx-eval-var-record-deletedBy]#[ctx-eval-var-record-deletedBy,resource.deletedBy](#ctx-eval-var-record-deletedBy,resource.deletedBy)#
|
Идентификатор пользователя, удалившего ресурс.
Параметр будет равен `0`, если ресурс не обновлялся.

| [#ctx-eval-var-record-values]#[ctx-eval-var-record-values,resource.values](#ctx-eval-var-record-values,resource.values)#
|
Набор пар «ключ-значение» для значений записи, определённых полями модуля.

!!! note
    Если поле модуля является многозначным, соответствующая запись «ключ-значение» будет массивом.


|===

## Рабочие процессы

.Переменные рабочих процессов:
[cols="1s,5a"]
|===
| [#ctx-eval-var-workflow-ID]#[ctx-eval-var-workflow-ID,resource.ID](#ctx-eval-var-workflow-ID,resource.ID)#
|
Идентификатор ресурса.

| [#ctx-eval-var-workflow-workflowID]#[ctx-eval-var-workflow-workflowID,resource.workflowID](#ctx-eval-var-workflow-workflowID,resource.workflowID)#
|
Идентификатор ресурса; то же, что и [ctx-eval-var-workflow-ID,resource.ID](#ctx-eval-var-workflow-ID,resource.ID).

| [#ctx-eval-var-workflow-labels]#[ctx-eval-var-workflow-labels,resource.labels](#ctx-eval-var-workflow-labels,resource.labels)#
|
Пара «ключ-значение» меток, связанных с этим ресурсом.

| [#ctx-eval-var-workflow-ownedBy]#[ctx-eval-var-workflow-ownedBy,resource.ownedBy](#ctx-eval-var-workflow-ownedBy,resource.ownedBy)#
|
Идентификатор пользователя-владельца ресурса.

| [#ctx-eval-var-workflow-createdAt]#[ctx-eval-var-workflow-createdAt,resource.createdAt](#ctx-eval-var-workflow-createdAt,resource.createdAt)#
|
Временная метка создания ресурса.

| [#ctx-eval-var-workflow-createdBy]#[ctx-eval-var-workflow-createdBy,resource.createdBy](#ctx-eval-var-workflow-createdBy,resource.createdBy)#
|
Идентификатор пользователя, создавшего ресурс.

| [#ctx-eval-var-workflow-updatedAt]#[ctx-eval-var-workflow-updatedAt,resource.updatedAt](#ctx-eval-var-workflow-updatedAt,resource.updatedAt)#
|
Временная метка последнего обновления ресурса.
Параметр не определён, если ресурс ещё не обновлялся.

| [#ctx-eval-var-workflow-updatedBy]#[ctx-eval-var-workflow-updatedBy,resource.updatedBy](#ctx-eval-var-workflow-updatedBy,resource.updatedBy)#
|
Идентификатор пользователя, который последним обновил ресурс.
Параметр будет равен `0`, если ресурс ещё не обновлялся.

| [#ctx-eval-var-workflow-deletedAt]#[ctx-eval-var-workflow-deletedAt,resource.deletedAt](#ctx-eval-var-workflow-deletedAt,resource.deletedAt)#
|
Временная метка удаления ресурса.
Параметр не определён, если ресурс не удалялся.

| [#ctx-eval-var-workflow-deletedBy]#[ctx-eval-var-workflow-deletedBy,resource.deletedBy](#ctx-eval-var-workflow-deletedBy,resource.deletedBy)#
|
Идентификатор пользователя, удалившего ресурс.
Параметр будет равен `0`, если ресурс не обновлялся.

|===

## Клиенты аутентификации

.Переменные клиентов аутентификации:
[cols="1s,5a"]
|===
| [#ctx-eval-var-auth-client-ID]#[ctx-eval-var-auth-client-ID,resource.ID](#ctx-eval-var-auth-client-ID,resource.ID)#
|
Идентификатор ресурса.

| [#ctx-eval-var-auth-client-labels]#[ctx-eval-var-auth-client-labels,resource.labels](#ctx-eval-var-auth-client-labels,resource.labels)#
|
Пара «ключ-значение» меток, связанных с этим ресурсом.

| [#ctx-eval-var-auth-client-scope]#[ctx-eval-var-auth-client-scope,resource.scope](#ctx-eval-var-auth-client-scope,resource.scope)#
|
Область действия (scope), определённая клиентом аутентификации.

| [#ctx-eval-var-auth-client-validGrant]#[ctx-eval-var-auth-client-validGrant,resource.validGrant](#ctx-eval-var-auth-client-validGrant,resource.validGrant)#
|
Тип гранта, поддерживаемый клиентом аутентификации.

| [#ctx-eval-var-auth-client-redirectURI]#[ctx-eval-var-auth-client-redirectURI,resource.redirectURI](#ctx-eval-var-auth-client-redirectURI,resource.redirectURI)#
|
URI перенаправления для клиента аутентификации.

| [#ctx-eval-var-auth-client-trusted]#[ctx-eval-var-auth-client-trusted,resource.trusted](#ctx-eval-var-auth-client-trusted,resource.trusted)#
|
Настройка доверия клиента аутентификации.

| [#ctx-eval-var-auth-client-enabled]#[ctx-eval-var-auth-client-enabled,resource.enabled](#ctx-eval-var-auth-client-enabled,resource.enabled)#
|
Настройка включения клиента аутентификации.

| [#ctx-eval-var-auth-client-validFrom]#[ctx-eval-var-auth-client-validFrom,resource.validFrom](#ctx-eval-var-auth-client-validFrom,resource.validFrom)#
|
Временная метка, обозначающая, с какого момента клиент аутентификации действителен.

| [#ctx-eval-var-auth-client-expiresAt]#[ctx-eval-var-auth-client-expiresAt,resource.expiresAt](#ctx-eval-var-auth-client-expiresAt,resource.expiresAt)#
|
Временная метка, обозначающая, когда клиент аутентификации перестаёт быть действительным.

| [#ctx-eval-var-auth-client-ownedBy]#[ctx-eval-var-auth-client-ownedBy,resource.ownedBy](#ctx-eval-var-auth-client-ownedBy,resource.ownedBy)#
|
Идентификатор пользователя-владельца ресурса.

| [#ctx-eval-var-auth-client-createdAt]#[ctx-eval-var-auth-client-createdAt,resource.createdAt](#ctx-eval-var-auth-client-createdAt,resource.createdAt)#
|
Временная метка создания ресурса.

| [#ctx-eval-var-auth-client-createdBy]#[ctx-eval-var-auth-client-createdBy,resource.createdBy](#ctx-eval-var-auth-client-createdBy,resource.createdBy)#
|
Идентификатор пользователя, создавшего ресурс.

| [#ctx-eval-var-auth-client-updatedAt]#[ctx-eval-var-auth-client-updatedAt,resource.updatedAt](#ctx-eval-var-auth-client-updatedAt,resource.updatedAt)#
|
Временная метка последнего обновления ресурса.
Параметр не определён, если ресурс ещё не обновлялся.

| [#ctx-eval-var-auth-client-updatedBy]#[ctx-eval-var-auth-client-updatedBy,resource.updatedBy](#ctx-eval-var-auth-client-updatedBy,resource.updatedBy)#
|
Идентификатор пользователя, который последним обновил ресурс.
Параметр будет равен `0`, если ресурс ещё не обновлялся.

| [#ctx-eval-var-auth-client-deletedAt]#[ctx-eval-var-auth-client-deletedAt,resource.deletedAt](#ctx-eval-var-auth-client-deletedAt,resource.deletedAt)#
|
Временная метка удаления ресурса.
Параметр не определён, если ресурс не удалялся.

| [#ctx-eval-var-auth-client-deletedBy]#[ctx-eval-var-auth-client-deletedBy,resource.deletedBy](#ctx-eval-var-auth-client-deletedBy,resource.deletedBy)#
|
Идентификатор пользователя, удалившего ресурс.
Параметр будет равен `0`, если ресурс не обновлялся.

|===
