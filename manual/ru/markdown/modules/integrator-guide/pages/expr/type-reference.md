# Type reference

!!! note
    *DevNote* сгенерировать остальное.


## Type reference

:leveloffset: +2


<a id="datatype-any"></a>
# `Any`

**Is primitive**: `n/a`

Переменная типа `Any` не выполняет никаких проверок типа и может хранить что угодно.
Использовать это не рекомендуется, но возможно.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| null
| Инициализирует переменную значением `null`.

| v = "test"
| "test"
| Инициализирует переменную строковым значением `test`.

| v = 42
| 42
| Инициализирует переменную целочисленным значением `42`.

|===

<a id="datatype-vars"></a>
# `Vars`

**Is primitive**: `no`

Идентичен [datatype-kv,`KV`](#datatype-kv,`KV`), но используется для сложных структур.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| {}
| Инициализирует переменную vars без значения, по умолчанию пусто (`{}`).

| v = { "key": "value" }
| { "key": "value" }
| Инициализирует переменную vars со свойством `key` и значением `value`.

| v = { "key": { "key": "value" } }
| { "key": { "key": "value" } }
| Инициализирует переменную vars со свойством `key` и значением вложенной переменной vars со свойством `key` и значением `value`.

|===

<a id="datatype-boolean"></a>
# `Boolean`

**Is primitive**: `yes`

`Boolean` указывает, является ли переменная истинной (`true`) или ложной (`false`).

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| false
| Инициализирует логическую переменную без значения, по умолчанию false.

| v = true
| true
| Инициализирует логическую переменную логическим значением.

| v = "true"
| true
|
.Разбор строковых значений:
- `"1"``, `"t"``, `"T"``, `"true"``, `"TRUE"``, `"True"`` дают результат `true`
- `"0"``, `"f"``, `"F"``, `"false"``, `"FALSE"``, `"False"`` дают результат `false`

| v = 10 == 0
| true
| Инициализирует логическую переменную как результат выражения.
|===

<a id="datatype-datetime"></a>
# `DateTime`

**Is primitive**: `yes`

`DateTime` содержит абсолютное временное значение.
`DateTime` можно легко изменять с помощью [функций даты и времени](modules/integrator-guide/partials/expr/expr/fnc-reference.md#datetime).

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| ERROR
| Инициализация `DateTime` требует не-null значения.

| v = "2022-01-17T07:32:30Z"
| "2022-01-17T07:32:30Z"
| Инициализирует переменную DateTime заданной меткой времени ISO 8601.

| v = 1642414876
| "2022-01-17T11:21:16"
| Инициализирует переменную DateTime заданной unix-меткой времени.

!!! caution
    При инициализации переменных DateTime с помощью unix-меток времени внутри рабочих процессов вам потребуется промежуточная переменная для значения unix-метки времени.


|===

<a id="datatype-duration"></a>
# `Duration`

**Is primitive**: `yes`

Переменная `Duration` содержит относительное временное значение.

!!! note
    Вы можете использовать любое значение, разбираемое функцией https://pkg.go.dev/time#ParseDuration[ParseDuration] языка Go.


Обычно используется в сочетании со значениями [datatype-datetime,DateTime](#datatype-datetime,DateTime) для вычисления смещений.

<a id="duration-units"></a>
.Доступные единицы:
[cols="2a,2m,4m"]
|===
| Unit | Expression unit | Example

| наносекунды
| ns
| 10ns

| микросекунды
| us
| 10us

| миллисекунды
| ms
| 10ms

| секунды
| s
| 10s

| минуты
| m
| 10m

| часы
| h
| 10h

|===

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| ERROR
| Инициализация `Duration` требует не-null значения.

| v = "1m"
| 1m0s
| Инициализирует переменную Duration значением 1 минута.

| v = "3h"
| 3h0m0s
| Инициализирует переменную Duration значением 3 часа.

|===

!!! note
    Вы можете использовать несколько <<duration-units,единиц длительности>> при условии их расположения в порядке убывания.
    
    * Корректный пример: `10m5s2ms`,
    * Некорректный пример: `2ms5s10m`.


<a id="datatype-float"></a>
# `Float`

**Is primitive**: `yes`

`Float` — это 64-битное число с плавающей точкой, соответствующее стандарту [IEEE 754](https://en.wikipedia.org/wiki/IEEE_754).

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| 0
| Инициализирует переменную float без значения, по умолчанию 0.

| v = 4.2
| 4.2
| Инициализирует переменную float значением с плавающей точкой.

| v = "4.2"
| "4.2"
| Инициализирует переменную float строковым значением.

| v = "four.two"
| ERROR
| Инициализирует переменную float некорректным строковым значением.

|===

<a id="datatype-integer"></a>
# `Integer`

**Is primitive**: `yes`

`Integer` — это 64-битное знаковое целое; диапазон от `-9223372036854775808` до `9223372036854775807`.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v (Integer)
| 0
| Инициализирует целочисленную переменную без значения, по умолчанию 0.

| v (Integer) = 42
| 42
| Инициализирует целочисленную переменную целочисленным значением.

| v (Integer) = 3.14
| 3
| Инициализация целочисленной переменной числом с плавающей точкой отбрасывает десятичную часть.

| v (Integer) = "42"
| 42
| Инициализирует целочисленную переменную строкой, содержащей корректное целое число.

| v (Integer) = "forty two"
| ERROR
| Инициализация целочисленной переменной строкой, содержащей некорректное целое число, приводит к ошибке.
|===

<a id="datatype-unsignedinteger"></a>
# `UnsignedInteger`

**Is primitive**: `yes`

`UnsignedInteger` — это 64-битное беззнаковое целое; диапазон от `0` до `18446744073709551615`.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v (Integer)
| 0
| Инициализирует беззнаковую целочисленную переменную без значения, по умолчанию 0.

| v (Integer) = 42
| 42
| Инициализирует беззнаковую целочисленную переменную целочисленным значением.

| v (Integer) = 3.14
| 3
| Инициализация беззнаковой целочисленной переменной числом с плавающей точкой отбрасывает десятичную часть.

| v (Integer) = "42"
| 42
| Инициализирует беззнаковую целочисленную переменную строкой, содержащей корректное целое число.

| v (Integer) = "forty two"
| ERROR
| Инициализация целочисленной переменной строкой, содержащей некорректное целое число, приводит к ошибке.
|===

<a id="datatype-string"></a>
# `String`

**Is primitive**: `yes`

Переменная `String` содержит текст (строку).

!!! tip
    Если значение инициализации не-null и не строка; при наличии такой возможности; значение приводится к строке.
    
    *DevNote* предоставить полный список для этого.


.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| ""
| Инициализирует строковую переменную без значения, по умолчанию пустая строка (`""`).

| v = "Hi"
| "Hi"
| Инициализирует строковую переменную строковым значением.

| v = 10
| "10"
| Инициализирует строковую переменную целочисленным значением.

|===

<a id="datatype-id"></a>
# `ID`

**Is primitive**: `yes`

`ID` — это уникальный идентификатор, назначаемый системой для всех системных ресурсов.
Хотя его можно назначить или изменить вручную, система обычно либо игнорирует изменения, либо вызывает ошибку.

`ID` — это 64-битное беззнаковое целое; диапазон от `0` до `18446744073709551615`.

<a id="datatype-handle"></a>
# `Handle`

**Is primitive**: `yes`

`Handle` — это удобочитаемый идентификатор ресурса.
Большинство ресурсов LowCoooode (такие как пространства имён и модули) позволяют задать уникальный идентификатор, который можно использовать вместо назначенного системой [datatype-id,ID](#datatype-id,ID).

Значение должно быть пустой строкой или строкой, соответствующей следующему регулярному выражению: `/^[A-Za-z][0-9A-Za-z_\-.]*[A-Za-z0-9]$/`.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| 0
| Инициализирует переменную handle без значения, по умолчанию пусто (`""`).

| v = "transaction"
| "transaction"
| Инициализирует переменную handle строковым значением.

| v = "invalid handle"
| ERROR
| Инициализирует переменную handle некорректным строковым значением.

|===

<a id="datatype-array"></a>
# `Array`

**Is primitive**: `no`

`Array` содержит список элементов, например, список пользователей, записей или модулей.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| []
| Инициализирует переменную массива без значения, по умолчанию пусто (`[]`).

| v = [1, 2, 3]
| [1, 2, 3]
| Инициализирует переменную массива списком целых чисел.

|===

<a id="datatype-reader"></a>
# `Reader`

**Is primitive**: `no`

`Reader` представляет поток данных (например, файл или blob), который можно читать или итерировать с помощью итератора потока.

!!! important
    `Reader` можно прочитать только один раз.
    Если вам нужно использовать его несколько раз, потребуется кэшировать первоначальный вывод.


.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| ERROR
| Инициализация `Reader` требует не-nil значения.

| v = "Test"
| Reader("Test")
| Инициализирует переменную reader из строки.

|===

<a id="datatype-kv"></a>
# `KV`

**Is primitive**: `no`

`KV` — это хеш-карта, которая сопоставляет значения `String` с ключами `String`.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| {}
| Инициализирует переменную kv без значения, по умолчанию пусто (`{}`).

| v = { "key": "value" }
| { "key": "value" }
| Инициализирует переменную kv заданным набором ключей и значений.

|===

<a id="datatype-kvv"></a>
# `KVV`

**Is primitive**: `no`

`KVV` — это хеш-карта, которая сопоставляет набор значений `String` с ключами `String`.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| {}
| Инициализирует переменную kvv без значения, по умолчанию пусто (`{}`).

| v = { "key": [ "value 1", "value 2" ] }
| { "key": [ "value 1", "value 2" ] }
| Инициализирует переменную kvv заданным набором ключей и значений.

|===


<a id="datatype-template"></a>
# `Template`

**Is primitive**: `no`

`Template` содержит шаблон, который может использоваться движком рендеринга.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| { ... }
| Инициализирует переменную шаблона без значения, по умолчанию пусто.
Обратитесь к [objref-template,ссылке на объект](#objref-template,ссылке на объект) за подробностями.

| v = { "handle": "test", "type": "text/html" }
| { ... "handle": "test", "type": "text/html" ... }
| Инициализирует переменную шаблона заданным handle и типом шаблона.
Обратитесь к [objref-template,ссылке на объект](#objref-template,ссылке на объект) за подробностями.

|===

<a id="datatype-document"></a>
# `Document`

**Is primitive**: `no`

`Document` — это отрендеренный [datatype-template,Template](#datatype-template,Template).
Он использует [datatype-renderoptions,RenderOptions](#datatype-renderoptions,RenderOptions) и [datatype-rendervariables,RenderVariables](#datatype-rendervariables,RenderVariables) для создания документа.

!!! note
    `Document` — это результат функции рендеринга шаблона и не должен создаваться вручную.


<a id="datatype-renderoptions"></a>
# `RenderOptions`

**Is primitive**: `no`

`RenderOptions` определяет, как [datatype-template,Template](#datatype-template,Template) должен быть отрендерен в [datatype-document,Document](#datatype-document,Document).

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| {}
| Инициализирует переменную параметров рендеринга без значения, по умолчанию пусто (`{}`).

| v = { "marginY": "1" }
| { "marginY": "1" }
| Инициализирует переменную параметров рендеринга заданным объектом.

|===

<a id="datatype-emailmessage"></a>
# `EmailMessage`

**Is primitive**: `no`

`EmailMessage` представляет электронное письмо, которое должно быть отправлено получателям.
Тип `EmailMessage` не следует создавать или взаимодействовать с ним напрямую.
Рекомендуется использовать предопределённые функции для управления его содержимым.

!!! caution
    Новый `EmailMessage` должен быть инициализирован с помощью функции `Email builder`.


<a id="datatype-role"></a>
# `Role`

**Is primitive**: `no`

`Role` содержит системную роль.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| {...}
| Инициализирует переменную роли без значения, по умолчанию пусто.
Обратитесь к [objref-role,ссылке на объект](#objref-role,ссылке на объект) за подробностями.

| v = { "handle": "test_role" }
| {... "handle": "test_role" ...}
| Инициализирует переменную роли заданным handle.
Обратитесь к [objref-role,ссылке на объект](#objref-role,ссылке на объект) за подробностями.

|===

<a id="datatype-user"></a>
# `User`

**Is primitive**: `no`

`User` содержит системного пользователя.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| {...}
| Инициализирует переменную пользователя без значения, по умолчанию пусто.
Обратитесь к [objref-user,ссылке на объект](#objref-user,ссылке на объект) за подробностями.

| v = { "handle": "test_user" }
| {... "handle": "test_user" ...}
| Инициализирует переменную пользователя заданным handle.
Обратитесь к [objref-user,ссылке на объект](#objref-user,ссылке на объект) за подробностями.

|===

<a id="datatype-composemodule"></a>
# `ComposeModule`

**Is primitive**: `no`

`ComposeModule` содержит модуль Low Code.
Тип `ComposeModule` в основном используется при обновлении полей модуля или создании новых записей.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| {...}
| Инициализирует переменную модуля без значения, по умолчанию пусто.
Обратитесь к [objref-module,ссылке на объект](#objref-module,ссылке на объект) за подробностями.

| v = { "handle": "test_module" }
| {... "handle": "test_module" ...}
| Инициализирует переменную модуля заданным handle.
Обратитесь к [objref-module,ссылке на объект](#objref-module,ссылке на объект) за подробностями.

|===

<a id="datatype-composenamespace"></a>
# `ComposeNamespace`

**Is primitive**: `no`

`ComposeNamespace` содержит пространство имён Low Code.
Тип `ComposeNamespace` в основном используется при взаимодействии с ресурсами Low Code, специфичными для пространства имён.

.Шаблоны инициализации:
[cols="2m,2m,4a"]
|===
| Expression | Value/result | Notes

| v
| {...}
| Инициализирует переменную пространства имён без значения, по умолчанию пусто.
Обратитесь к [objref-namespace,ссылке на объект](#objref-namespace,ссылке на объект) за подробностями.

| v = { "slug": "test_namespace" }
| {... "slug": "test_namespace" ...}
| Инициализирует переменную пространства имён заданным handle.
Обратитесь к [objref-namespace,ссылке на объект](#objref-namespace,ссылке на объект) за подробностями.

|===

<a id="datatype-composerecord"></a>
# `ComposeRecord`

**Is primitive**: `no`


`ComposeRecords` содержит запись Low Code.
Тип `ComposeRecords` в основном используется при взаимодействии с записями Low Code, например, при изменении их значений или преобразовании в email-уведомления.

!!! caution
    Новый `ComposeRecord` должен быть инициализирован с помощью функции `Compose record maker`.


<a id="datatype-composerecordvalues"></a>
# `ComposeRecordValues`

**Is primitive**: `no`

`ComposeRecordValues` содержит набор значений записей LowCoooode.
Этот тип обычно не используется сам по себе, а в сочетании с [datatype-composerecord,ComposeRecord](#datatype-composerecord,ComposeRecord).

<a id="datatype-httprequest"></a>
# `HttpRequest`
**Is primitive**: `no`

`HttpRequest` содержит HTTP-запрос (обратитесь к [документации Go](https://pkg.go.dev/net/http#Request) за подробностями о сигнатуре).

!!! important
    Единственное различие между HttpRequest и http.Request в Go — это возможность буферизации тела запроса.
    
    Как только тело будет прочитано в первый раз, оно будет буферизовано и в дальнейшем использоваться, если тело пусто (поскольку это ReadCloser).


<a id="object-reference"></a>
# Object reference


:leveloffset: -2

## Object reference

:leveloffset: +2


[cols="2m,3a"]
|===
| Type | Structure

| [#objref-attachment]#[objref-attachment,Attachment](#objref-attachment,Attachment)#
|
```
```
{
   ID (ID)
   kind (String)
   url (Handle)
   previewUrl (Handle)
   name (Handle)
   createdAt (DateTime)
   updatedAt (DateTime)
   deletedAt (DateTime)
}

| [#objref-composemodule]#[objref-composemodule,ComposeModule](#objref-composemodule,ComposeModule)#
|
```
```
{
   ID (ID)
   namespaceID (ID)
   name (String)
   handle (Handle)
   labels (KV)
   createdAt (DateTime)
   updatedAt (DateTime)
   deletedAt (DateTime)
}

| [#objref-composenamespace]#[objref-composenamespace,ComposeNamespace](#objref-composenamespace,ComposeNamespace)#
|
```
```
{
   ID (ID)
   name (String)
   slug (Handle)
   labels (KV)
   createdAt (DateTime)
   updatedAt (DateTime)
   deletedAt (DateTime)
}

| [#objref-composerecord]#[objref-composerecord,ComposeRecord](#objref-composerecord,ComposeRecord)#
|
```
```
{
   ID (ID)
   moduleID (ID)
   namespaceID (ID)
   values (ComposeRecordValues)
   meta (Meta)
   ownedBy (ID)
   createdAt (DateTime)
   createdBy (ID)
   updatedAt (DateTime)
   updatedBy (ID)
   deletedAt (DateTime)
   deletedBy (ID)
}


| [#objref-httprequest]#[objref-httprequest,HttpRequest](#objref-httprequest,HttpRequest)#
|
```
```
{
   Method (String)
   URL (Url)
   Header (KVV)
   Body (Reader)
   Form (KVV)
   PostForm (KVV)
}


| [#objref-queuemessage]#[objref-queuemessage,QueueMessage](#objref-queuemessage,QueueMessage)#
|
```
```
{
   Queue (String)
   Payload (Bytes)
}

| [#objref-rendereddocument]#[objref-rendereddocument,RenderedDocument](#objref-rendereddocument,RenderedDocument)#
|
```
```
{
   document (Reader)
   name (string)
   type (string)
}

| [#objref-role]#[objref-role,Role](#objref-role,Role)#
|
```
```
{
   ID (ID)
   name (String)
   handle (Handle)
   labels (KV)
   createdAt (DateTime)
   updatedAt (DateTime)
   archivedAt (DateTime)
   deletedAt (DateTime)
}

| [#objref-template]#[objref-template,Template](#objref-template,Template)#
|
```
```
{
   ID (ID)
   handle (Handle)
   language (String)
   type (DocumentType)
   partial (Boolean)
   meta (TemplateMeta)
   template (String)
   labels (KV)
   ownerID (ID)
   createdAt (DateTime)
   updatedAt (DateTime)
   deletedAt (DateTime)
   lastUsedAt (DateTime)
}

| [#objref-templatemeta]#[objref-templatemeta,TemplateMeta](#objref-templatemeta,TemplateMeta)#
|
```
```
{
   short (String)
   description (String)
}

| [#objref-user]#[objref-user,User](#objref-user,User)#
|
```
```
{
   ID (ID)
   username (String)
   email (String)
   name (String)
   handle (Handle)
   emailConfirmed (Boolean)
   labels (KV)
   createdAt (DateTime)
   updatedAt (DateTime)
   suspendedAt (DateTime)
   deletedAt (DateTime)
}


|===


:leveloffset: -2
