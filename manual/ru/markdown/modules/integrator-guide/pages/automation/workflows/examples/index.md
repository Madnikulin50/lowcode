# Examples

:leveloffset: +1

# Добавление вложения в письмо
:attachment-path: ../../../_attachments/automation/workflows/
:page-noindex: true

В этом примере предположим, что мы хотим отправить клиенту коммерческое предложение (quote) на новый продукт, который он заказал.
Коммерческое предложение будет предоставлено нашей [системой шаблонов](modules/integrator-guide/pages/automation/workflows/examples/templates/index.md).

!!! important
    Для рендеринга шаблонов в виде PDF-документов необходимо [настроить рендеринг PDF](modules/devops-guide/pages/pdf-renderer.md).


.На скриншоте показан рабочий процесс, который можно использовать для отправки вложения по электронной почте.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/email-attachment/workflow.png",
    "alias": "automation-workflows-examples-email-attachment-workflow.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 214,
    "y": 178,
    "w": 791,
    "h": 575
  },
  "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}email*attachment*send.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) System; onManual**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Render quote**:
*** *type**: `Template render`
*** *arguments**:
**** *lookup**:
****** **type**: `Handle`
****** **value type**: constant
****** **value**: `quote`
**** *documentName**:
****** **value type**: constant
****** **value**: `quote`
**** *documentType**:
****** **value type**: constant
****** **value**: `application/pdf`
**** *options**:
****** **value type**: constant
****** **value**: `{ "documentSize": "A4", "contentScale": "1", "orientation": "portrait", "margin": "0.3" }`
*** *results**:
**** *document target**: `quote`
3. **(3) Render email**:
*** *type**: `Template render`
*** *arguments**:
**** *lookup**:
****** **type**: `Handle`
****** **value type**: constant
****** **value**: `content`
**** *documentType**:
****** **value type**: constant
****** **value**: `text/html`
*** *results**:
**** *document target**: `content`
4. **(4) Build base email**:
*** *type**: `Email builder`
*** *arguments**:
**** *subject**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `Quote for your product`
**** *to**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `example@mail.tld`
**** *html**:
****** **type**: `Reader`
****** **value type**: expression
****** **value**: `content.document`
*** *results**:
**** *message target**: `email`
5. **(5) Attach rendered template**:
*** *type**: `Email embedded attachment`
*** *arguments**:
**** *message**:
****** **type**: `EmailMessage`
****** **value type**: expression
****** **value**: `email`
**** *content**:
****** **type**: `Reader`
****** **value type**: expression
****** **value**: `quote.document`
**** *name**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `quote.pdf`
6. **(6) Send email**:
*** *type**: `Email sender`
*** *arguments**:
**** *message**:
****** **type**: `EmailMessage`
****** **value type**: expression
****** **value**: `email`
7. **(7) Done**
******

!!! caution
    Убедитесь, что параметр `name` вложений `email embedded attachment` использует правильное расширение, например `.txt` для обычного текста и `.pdf` для PDF.
    Если расширение опущено или указано неверно, некоторые почтовые клиенты могут отобразить вложение неправильно или полностью его проигнорировать.


Итоговое электронное письмо включает коммерческое предложение в формате PDF в качестве вложения.

.На скриншоте показано исходное полученное электронное письмо.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/email-attachment/email-mime-contents.png",
    "alias": "automation-workflows-examples-email-attachment-email-mime-contents.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

## Примеры шаблонов

В приведённом выше примере используются два шаблона: `quote` и `content`

.На скриншоте показан базовый шаблон содержимого письма, используемый в примере.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/email-attachment/template-base-content.png",
    "alias": "automation-workflows-examples-email-attachment-template-base-content.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

.Исходный код шаблона содержимого письма:
```html
```
<!DOCTYPE html>
<html>
<head>
  <meta charset='utf-8'>
  <meta http-equiv='X-UA-Compatible' content='IE=edge'>
  <title>Quote for our services</title>
  <meta name='viewport' content='width=device-width, initial-scale=1'>
</head>
<body>
  <h1>Hello!</h1>
  <p>
    Attached is the quote for our services.
  </p>
  <p>
    Best, Team ltd
  </p>
</body>
</html>

.На скриншоте показан базовый шаблон коммерческого предложения, используемый в примере.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/email-attachment/template-base-quote.png",
    "alias": "automation-workflows-examples-email-attachment-template-base-quote.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

.Исходный код шаблона коммерческого предложения:
```html
```
<!DOCTYPE html>
<html>
<head>
  <meta charset='utf-8'>
  <meta http-equiv='X-UA-Compatible' content='IE=edge'>
  <title>Quote#012356</title>
  <meta name='viewport' content='width=device-width, initial-scale=1'>
</head>
<body>
  <h1>Quote#012356</h1>
  <p>
    Service quote for our services
  </p>
</body>
</html>


# Разбор запроса шлюза интеграции (Integration Gateway)
:attachment-path: ../../../_attachments/automation/workflows/
:page-noindex: true

В этом примере мы разберём запрос, отправленный на [Api Gw](modules/integrator-guide/pages/automation/workflows/examples/api-gw/index.md)

.На скриншоте показан рабочий процесс, который можно использовать для обработки запроса.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parse-gateway-payload.png",
    "alias": "automation-workflows-examples-parse-gateway-payload.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 93,
    "y": 183,
    "w": 1006,
    "h": 144
  },
  "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}request_process.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) System; onManual**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Parse request payload**:
*** *type**: `Process arbitrary data in jsenv`
*** *arguments**:
**** *scope**:
****** **type**: `Any`
****** **value type**: expression
****** **value**: `payload`
**** *source**: refer below
*** *results**:
**** *resultAny**: `parsedPayload`
3. **(3) Debug state**
4. **(7) Done**
******

## Конфигурация шлюза интеграции

.Ниже приведены параметры конфигурации соответствующего шлюза интеграции:
[cols="1s,5a"]
|===
| Endpoint
| `/examples/payload`

| Prefilter
| Header: `Token == "SOME*SECRET*TOKEN"`

| Processing
| Workflow processer: `On request payload notify user`

| Postfilter
| `Default JSON Response`
|===

## Запрос cURL

.Ниже приведён пример запроса cURL, вызывающего шлюз интеграции и рабочий процесс:
```bash
```
curl -X POST "$LOWCODE_BASE/api/gateway/examples/payload" \
  -H "Content-Type: application/json" \
  -H 'Token: SOME*SECRET*TOKEN' \
  -d '{"name":"Peter","id":123}';

## Функция разбора JSEnv

.Ниже приведён извлечённый исходный код, который следует использовать в шаге функции JSEnv:
```js
```
var inputData;

try {
    inputData = JSON.parse(input)
} catch (e) {
    throw new Error('could not parse input payload: ' + e);
}

if (!inputData.name) {
    throw new Error('could not parse input payload, name missing');
}

return {
    id: inputData.id,
    name: inputData.name
}


# Интервал
:page-noindex: true

В этом примере предположим, что мы хотим отправить письмо всем пользователям каждое Рождество в полночь.
Для этого мы используем триггер типа `onInterval` с интервалом `0 0 25 12 *`

![](automation/workflows/examples/workflow-example-interval.png)

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onInterval`
*** *enabled**: checked
*** *constraints**: 
**** *interval**: `0 0 25 12 *`
2. **(2) Iterate over Users**:
*** *type**: `Users`
*** *results**:
**** *user target**: `user`
3. **(3) Send Email**:
*** *type**: `Email`
*** *arguments**:
**** *subject**:
****** **value type**: constant
****** **value**: `Merry christmas`
**** *to**:
****** **value type**: expression
****** **value**: `user.email`
**** *plain**:
****** **value type**: constant
****** **value**: `Merry christmas`
4. **(4) Done**:
******


# Рендеринг шаблонов
:attachment-path: ../../../_attachments/automation/workflows/
:page-noindex: true

В этом примере мы рассмотрим, как отображать различные шаблоны, созданные в нашей [системе шаблонов](modules/integrator-guide/pages/automation/workflows/examples/templates/index.md).

Для начала создайте шаблон типа HTML в нашей системе шаблонов, расположенной в админ-области.

## Рендеринг HTML-шаблона

Создайте рабочий процесс, похожий на приведённый ниже.

Обязательно посмотрите детали шагов (2) и (3).

.На скриншоте показан рабочий процесс, который отображает HTML-шаблон и отправляет его по электронной почте.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/template-rendering/html-workflow.png",
    "alias": "automation-workflows-examples-template-rendering-html-workflow.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 128,
    "y": 256,
    "w": 755,
    "h": 142
  },
  "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}template*rendering*html.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) System; onManual**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Render template**:
*** *type**: `Template render`
*** *arguments**:
**** *lookup**:
****** **type**: `Handle`
****** **value type**: constant
****** **value**: `email-template`
**** *documentName**:
****** **value type**: constant
****** **value**: `Email template`
**** *documentType**:
****** **value type**: constant
****** **value**: `text/html`
*** *results**:
**** *document target**: `renderedTemplate`
3. **(3) Send email**:
*** *type**: `Email`
*** *arguments**:
**** *subject**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `Email template`
**** *from**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `demo@mail.com`
**** *to**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `test@mail.com`
**** *html**:
****** **type**: `Reader`
****** **value type**: expression
****** **value**: `renderedTemplate.document`
*** *results**:
**** *document target**: `content`
******

!!! caution
    Убедитесь, что параметр `documentType` шага `(2) Render template` имеет значение `text/html`.


## Рендеринг PDF-шаблона
Разница между HTML- и PDF-рендерингом заключается в том, что параметр `documentType` меняется на `application/pdf`.

Для PDF-файлов вы также можете настроить [параметры рендеринга](modules/integrator-guide/pages/templates/index.md#tpl-render-options_marginbottom), чтобы изменить то, как будет отрендерен итоговый PDF.

!!! important
    Для рендеринга шаблонов в виде PDF-документов необходимо настроить [рендеринг PDF](modules/devops-guide/pages/pdf-renderer.md).


.На скриншоте показан рабочий процесс, который отображает PDF-шаблон и отправляет его как вложение в письмо.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/template-rendering/pdf-workflow.png",
    "alias": "automation-workflows-examples-template-rendering-pdf-workflow.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 128,
    "y": 255,
    "w": 755,
    "h": 324
  },
  "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}template*rendering*pdf.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) System; onManual**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Render template**:
*** *type**: `Template render`
*** *arguments**:
**** *lookup**:
****** **type**: `Handle`
****** **value type**: constant
****** **value**: `pdf-template`
**** *documentName**:
****** **value type**: constant
****** **value**: `PDF template`
**** *documentType**:
****** **value type**: constant
****** **value**: `application/pdf`
**** *options**:
****** **type**: `renderOptions`
****** **value type**: expression
****** **value**: `{
  "documentSize": "A4",
  "contentScale": "1",
  "orientation": "portrait",
  "margin": "0.3"
}`
*** *results**:
**** *document target**: `renderedTemplate`
3. **(3) Build email**:
*** *type**: `Email builder`
*** *arguments**:
**** *subject**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `PDF template`
**** *from**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `test@mail.com`
**** *to**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `example@mail.com`
**** *html**:
****** **type**: `Reader`
****** **value type**: expression
****** **value**: `content.document`
*** *results**:
**** *message target**: `email`
4. **(4) Attach PDF template**:
*** *type**: `Attach PDF template`
*** *arguments**:
**** *message**:
****** **type**: `EmailMessage`
****** **value type**: expression
****** **value**: `email`
**** *content**:
****** **type**: `Reader`
****** **value type**: expression
****** **value**: `renderedTemplate.document`
**** *name**:
****** **type**: `String`
****** **value type**: constant
****** **value**: `PDF template.pdf`
5. **(5) Send email**:
*** *type**: `Email sender`
*** *arguments**:
**** *message**:
****** **type**: `EmailMessage`
****** **value type**: expression
****** **value**: `email`
******

!!! tip
    Вы можете комбинировать оба типа шаблонов, чтобы отображать динамические письма с PDF-вложениями.
    
    Пример рабочего процесса можно найти [здесь](modules/integrator-guide/pages/automation/workflows/examples/email-attachment.md)


# Метка времени
:page-noindex: true

В этом примере предположим, что мы хотим поздравить всех пользователей с Новым годом.
Все письма должны быть отправлены ровно в полночь.
Для этого мы используем триггер типа `onTimestamp` с меткой времени `2021-01-01T00:00:00Z`

![](automation/workflows/examples/workflow-example-timestamp.png)

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onTimestamp`
*** *enabled**: checked
*** *constraints**: 
**** *timestamp**: `2021-01-01T00:00:00Z`
2. **(2) Iterate over Users**:
*** *type**: `Users`
*** *results**:
**** *user target**: `user`
3. **(3) Send Email**:
*** *type**: `Email`
*** *arguments**:
**** *subject**:
****** **value type**: constant
****** **value**: `Happy new year`
**** *to**:
****** **value type**: expression
****** **value**: `user.email`
**** *plain**:
****** **value type**: constant
****** **value**: `Happy new year`
4. **(4) Done**:
******


# Работа с записями
:attachment-path: ../../../_attachments/automation/workflows/examples/record/
:page-noindex: true

В этом разделе приведены некоторые советы и приёмы, которые можно использовать при работе с записями.

## Проверка существования

Если вы хотите выполнить какую-либо задачу в зависимости от наличия записей, вы можете использовать любой из следующих подходов.

Оба подхода допустимы, и нет никакой разницы в том, какой из них использовать.
Решайте исходя из своих предпочтений/контекста.

### Подход A

При поиске записей отметьте параметр `incTotal` и присвойте значение результата `total` переменной.

Внутри шлюза проверьте, больше ли значение `total` нуля.

.На скриншоте показан базовый пример проверки существования.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/existence-a.png",
    "alias": "automation-workflows-examples-record-existence-a",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 200,
    "y": 77,
    "w": 611,
    "h": 537
  },
    "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}existence-a.json).

### Подход B

Внутри шлюза проверьте, больше ли значение `count(items)` нуля.

.На скриншоте показан базовый пример проверки существования.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/existence-b.png",
    "alias": "automation-workflows-examples-record-existence-b",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 201,
    "y": 77,
    "w": 609,
    "h": 535
  },
    "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}existence-b.json).

## Создание или обновление

При создании записи нужно вызвать функцию `compose record create`, а при обновлении записи — функцию `compose record update`.

!!! note
    Только выделенная часть выполняет проверку создания/обновления; остальное — стандартный код для приведения в нужное состояние.


Если вам нужно вызвать ту или иную функцию на лету, можно использовать следующий подход.
Вы можете использовать `record.recordID != "0"`, чтобы определить, нужно ли обновлять запись — значением `recordID` по умолчанию является `"0"`.

.На скриншоте показан базовый пример проверки существования.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/create-update.png",
    "alias": "automation-workflows-examples-record-create-update",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 386,
    "y": 84,
    "w": 684,
    "h": 899
  },
  "annotations": [{
    "kind": "box-note",
    "x": 424,
    "y": 766,
    "w": 614,
    "h": 184
  }]
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}create-update.json).

## Удаление значения

Чтобы удалить какое-либо значение записи, используйте шаг выражения, чтобы задать соответствующему значению пустой `Any`.

.На скриншоте показан базовый пример удаления значений записи.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/remove-value.png",
    "alias": "automation-workflows-examples-record-remove-value",
    "w": 1920,
    "h": 1080
  },
 "view": {
    "x": 892,
    "y": 1,
    "w": 1027,
    "h": 405
  },
  "annotations": [{
    "kind": "box-note",
    "x": 1476,
    "y": 303,
    "w": 430,
    "h": 15
  }]
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}remove-value.json).

## Обработка отсутствующих значений

Чтобы использовать значение по умолчанию в случае отсутствия значения записи, нужно использовать оператор `??`.

Например, выражение `a ?? b` вернёт `a`, если оно существует, или `b`, если его нет.

!!! note
    В приведённом ниже примере в качестве значения по умолчанию используется переменная.
    Вы можете использовать константу, например `"something string"` или `42`.


.На скриншоте показан базовый пример использования значений по умолчанию.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/missing-value-default.png",
    "alias": "automation-workflows-examples-record-missing-value-default",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 892,
    "y": 1,
    "w": 1027,
    "h": 840
  },
  "annotations": [{
    "kind": "box-note",
    "x": 1476,
    "y": 304,
    "w": 431,
    "h": 494
  }]
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}missing-value-default.json).


# Параллелизм
:attachment-path: ../../../_attachments/automation/workflows/examples/
:page-noindex: true

В этом разделе приведены некоторые примеры того, как следует выполнять задачи параллельно.

## Безусловный параллелизм

Безусловный параллелизм следует использовать, когда две или более ветви выполнения должны выполняться параллельно независимо от состояния.

Для этого используйте **разветвляющий шлюз** (fork gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}, где каждая исходящая ветвь определяет одну ветвь параллельного выполнения.

!!! caution
    Если какая-либо из ветвей определяет *шаг завершения* {ICON_WORKFLOW_TERMINATION}, весь рабочий процесс будет завершён.


.На скриншоте показан базовый пример безусловного параллельного выполнения.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parallelism/unconditional.png",
    "alias": "automation-workflows-examples-parallelism-unconditional",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 244,
    "y": 84,
    "w": 683,
    "h": 539
  },
    "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}parallelism/unconditional.json).

## Безусловный параллельный сегмент

Параллельный сегмент — это место, где рабочий процесс переходит от последовательного выполнения к параллельному и обратно к последовательному.

Безусловный параллелизм следует использовать, когда две или более ветви выполнения должны выполняться параллельно независимо от состояния.

Для этого используйте **разветвляющий шлюз** (fork gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}, где каждая исходящая ветвь определяет одну ветвь параллельного выполнения.
Завершите параллельный сегмент с помощью **соединяющего шлюза** (join gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}.

!!! caution
    Если какая-либо из ветвей определяет *шаг завершения* {ICON_WORKFLOW_TERMINATION}, весь рабочий процесс будет завершён.


.На скриншоте показан базовый пример безусловного параллельного выполнения.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parallelism/unconditional-segment.png",
    "alias": "automation-workflows-examples-parallelism-unconditional-segment",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 77,
    "w": 683,
    "h": 897
  },
    "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}parallelism/unconditional-segment.json).

## Условный параллелизм

Условный параллелизм следует использовать, когда две или более ветви выполнения должны выполняться параллельно в зависимости от состояния.

Для этого используйте **разветвляющий шлюз** (fork gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}, где каждая исходящая ветвь определяет одну ветвь параллельного выполнения.

!!! caution
    Если какая-либо из ветвей определяет *шаг завершения* {ICON_WORKFLOW_TERMINATION}, весь рабочий процесс будет завершён.


.На скриншоте показан базовый пример условного параллельного выполнения.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parallelism/conditional.png",
    "alias": "automation-workflows-examples-parallelism-conditional",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 76,
    "w": 683,
    "h": 601
  },
  "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}parallelism/conditional.json).

## Условный параллельный сегмент

Параллельный сегмент — это место, где рабочий процесс переходит от последовательного выполнения к параллельному и обратно к последовательному.

Условный параллелизм следует использовать, когда две или более ветви выполнения должны выполняться параллельно в зависимости от состояния.

Для этого используйте **разветвляющий шлюз** (fork gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}, где каждая исходящая ветвь определяет одну ветвь параллельного выполнения.
Завершите параллельный сегмент с помощью **соединяющего шлюза** (join gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}.

!!! caution
    Если какая-либо из ветвей определяет *шаг завершения* {ICON_WORKFLOW_TERMINATION}, весь рабочий процесс будет завершён.


.На скриншоте показан базовый пример условного параллельного выполнения.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parallelism/conditional-segment.png",
    "alias": "automation-workflows-examples-parallelism-conditional-segment",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 214,
    "y": 57,
    "w": 898,
    "h": 870
  },
  "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}parallelism/conditional-segment.json).


# Динамическая конфигурация
:page-noindex: true
:attachment-path: ../../../_attachments/automation/workflows/examples/

При определении автоматизации, которая должна взаимодействовать с внешними системами, или когда вам нужно сделать выполнение рабочего процесса настраиваемым, статические рабочие процессы могут оказаться неудобными.

Вы можете определить модуль `settings`, в котором определите все настраиваемые параметры, необходимые вашей автоматизации.
Это может быть что угодно: от URL-адресов до учётных данных для входа и токенов доступа.

!!! caution
    При хранении токенов доступа и других учётных данных обязательно правильно настройте контроль доступа.


.На скриншоте показан базовый пример модуля `settings`.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/dynamic-config/settings-module.png",
    "alias": "automation-workflows-examples-dynamic-config-settings-module",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

Внутри рабочего процесса просто получите запись из модуля `settings` и настройте выполнение, используя её значения.

.На скриншоте показан базовый пример рабочего процесса, использующего модуль `settings`.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/dynamic-config/example-workflow.png",
    "alias": "automation-workflows-examples-dynamic-config-example-workflow",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 358,
    "y": 183,
    "w": 610,
    "h": 395
  },
  "annotations": [{
    "kind": "box-note",
    "x": 679,
    "y": 254,
    "w": 183,
    "h": 74
  }]
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}dynamic-config/example-workflow.json).


# Execution timeout
:page-noindex: true

LowCoooode workflows' execution does not timeout by default (the workflow can run indefinitely).

If your use-case requires you to define a timeout, you can implement this manually.

.The below example does two things:
- Schedule a timeout after 10sec of execution,
- repeat the iterator indefinitely.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/manual-timeout.png",
    "alias": "automation-workflows-examples-manual-timeout.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 92,
    "y": 219,
    "w": 1224,
    "h": 647
  },
  "annotations": []
}

******
.Workflow step details:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Fork**:
*** *gateway**: `Inclusive`
*** *conditions**:
*** `true`
*** `true`
3. **(3) Wait 10sec**:
*** *offset**: `10s`
4. **(4) Log context timeout**:
*** *type**: `Log debug message`
*** *arguments**:
**** *message**:
****** **value type**:  constant
****** **value**: `Context timeout`
5. **(5) Context timeout**: /
6. **(6) Repeat indefinitely**:
*** *type**: `Condition`
*** *arguments**:
**** *while**:
****** **value type**: expression
****** **value**: `true`
7. **(7) Log iteration**:
*** *type**: `Log debug message`
*** *arguments**:
**** *message**:
****** **value type**: constant
****** **value**: `Iterator loop`
8. **(8) Done**: /
9. **(9) Completed**: /
******


# Очередь Messagebus
:attachment-path: ../../../_attachments/automation/workflows/
:page-noindex: true

Механизм очередей Messagebus дополняет рабочие процессы таким образом, что любые асинхронные вычисления могут быть вынесены в другой рабочий процесс или процесс, расширяемый рабочим процессом.
Больше информации по описанным темам можно найти в [Workflows](modules/integrator-guide/pages/automation/workflows/examples/automation/workflows/index.md), у шлюза интеграции — в [Profiler](modules/integrator-guide/pages/automation/workflows/examples/api-gw/profiler.md) и в подсистеме [Messagebus](modules/developer-guide/pages/lowcode-server/messagebus.md).

.В этом примере мы:
- создадим очередь
- создадим эндпоинт
- запишем в очередь через рабочий процесс
- прочитаем из очереди через рабочий процесс
** отправим полезную нагрузку на эндпоинт
- посмотрим полезную нагрузку эндпоинта в профилировщике

## Создание новой очереди

!!! important
    `Consumer` здесь должен быть `Eventbus`, поскольку именно так внутренний механизм предоставляет данные полезной нагрузки подсистемам LowCoooode, включая рабочие процессы.


.На скриншоте показано создание очереди в интерфейсе.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-create.png",
    "alias": "automation-workflows-examples-queue-create.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

## Запись в очередь

Это рабочий процесс, который мы будем использовать только для ручного запуска записи в очередь.

!!! note
    Вы можете использовать любые кнопки автоматизации, чтобы запустить этот рабочий процесс из интерфейса.
    Обычно это делается после создания записи, когда новый recordID отправляется в очередь.


Полезная нагрузка определяется как `Array`, который затем преобразуется в строку
```json
```
toJSON([
  {"key":"foo", "value": "bar"},
  {"key":"bar", "value": "baz"}
])

.Скриншот рабочего процесса добавления в очередь.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-wf-add.png",
    "alias": "automation-workflows-examples-queue-wf-add.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 164,
    "y": 111,
    "w": 251,
    "h": 287
  },
    "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}queue_add.json).

## Чтение из очереди

Здесь мы разбираем полезную нагрузку очереди и создаём пользовательский HTTP-запрос во внешнюю систему, т.е. Kafka (в нашем примере — эндпоинт, который мы создадим только для этой цели).

.Пример чтения данных из очереди.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-wf-read.png",
    "alias": "automation-workflows-examples-queue-wf-read.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 128,
    "y": 75,
    "w": 251,
    "h": 720
  },
  "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}queue_read.json).

## Создание эндпоинта имитации сервиса

Теперь нам нужно запустить рабочий процесс `Queue Write`, и запрос будет отправлен на наш локальный сервер на эндпоинт `/api/gateway/example_kafka`.

Результирующий запрос можно увидеть в профилировщике шлюза интеграции (если он включён, подробнее см. [Profiler](modules/integrator-guide/pages/automation/workflows/examples/api-gw/profiler.md)).

.Добавление нового эндпоинта в шлюз интеграции.
![role="data-zoomable"](automation/workflows/examples/queue-endpoint.png)

## Просмотр запроса в профилировщике

Этот просмотр приведён только для примера — это способ использования профилировщика для отладки отправляемых полезных нагрузок в производственные/внешние сервисы.

.Просмотр заголовков запроса и других метаданных.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-profiler-1.png",
    "alias": "automation-workflows-examples-queue-profiler-1.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

.Просмотр полезной нагрузки запроса.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-profiler-2.png",
    "alias": "automation-workflows-examples-queue-profiler-2.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}


# Уведомления
:attachment-path: ../../../_attachments/automation/workflows/examples/
:page-noindex: true


LowCoooode предоставляет гибкую систему уведомлений, позволяющую отправлять уведомления для любого события, поддерживаемого триггерами рабочих процессов.

Это даёт вам полный контроль над:

- Когда отправляются уведомления
- Кто их получает
- Какую информацию они содержат

!!! important
    Убедитесь, что у пользователя, вызывающего рабочий процесс, есть разрешение на назначение уведомлений пользователям (меню:Admin Area[System > Permissions > `Allow notification assignment])


## Включение уведомлений

Сначала перейдите в админ-область (Admin Area) и откройте меню:User Interface[Settings], затем прокрутите вниз до раздела Topbar.

Убедитесь, что параметр `Hide notifications` не отмечен.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-settings.png",
    "alias": "automation-workflows-examples-notifications-notifications-settings.png",
    "w": 1258,
    "h": 657
  },
  "view": {},
  "annotations": []
}

Уведомления будут отображаться в верхней панели (topbar) приложения.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-topbar.png",
    "alias": "automation-workflows-examples-notifications-notifications-topbar.png",
    "w": 154,
    "h": 52
  },
  "view": {},
  "annotations": []
}

## Отправка простого уведомления

Чтобы отправить простое уведомление, вы можете использовать функцию `Send simple notification` внутри рабочего процесса.
Функция отправит уведомление получателю с указанными заголовком и описанием.

Она имеет следующие параметры:

- `recipient` — получатель уведомления.
- `title` — заголовок уведомления.
- `description` — описание уведомления.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-simple-workflow.png",
    "alias": "automation-workflows-examples-notifications-notifications-simple-workflow.png",
    "w": 1918,
    "h": 757
  },
  "view": {},
  "annotations": []
}

После выполнения рабочего процесса уведомление появится в `Notification Sidebar` (боковой панели уведомлений), доступ к которой осуществляется из верхней панели.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-simple.png",
    "alias": "automation-workflows-examples-notifications-notifications-simple.png",
    "w": 399,
    "h": 414
  },
  "view": {},
  "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}send-simple-notification.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Send simple notification**:
*** *recipient**:
****** **value type**: constant
****** **value**: `test-user`
*** *title**:
****** **value type**: constant
****** **value**: `Simple notification`
*** *description**:
****** **value type**: constant
****** **value**: `This is a simple notification`
3. **(3) Done**:
******

## Отправка уведомления о записи

Чтобы отправить уведомление о записи, вы можете использовать функцию `Send record notification` внутри рабочего процесса.
Функция отправит уведомление получателю с указанными заголовком и описанием.
Если на уведомление нажать, оно откроет запись в указанном режиме (модальное окно, новая вкладка, текущая вкладка).

Она имеет следующие параметры:

- `recipient` — получатель уведомления.
- `title` — заголовок уведомления.
- `description` — описание уведомления.
- `namespace` — пространство имён записи.
- `module` — модуль записи.
- `record` — запись.
- `openMode` — режим, в котором будет открыта запись (модальное окно, новая вкладка, текущая вкладка).
- `edit` — если true, запись будет открыта в режиме редактирования.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-record-workflow.png",
    "alias": "automation-workflows-examples-notifications-notifications-record-workflow.png",
    "w": 1918,
    "h": 873
  },
  "view": {},
  "annotations": []
}

После выполнения рабочего процесса уведомление появится в `Notification Sidebar` (боковой панели уведомлений), доступ к которой осуществляется из верхней панели.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-record.png",
    "alias": "automation-workflows-examples-notifications-notifications-record.png",
    "w": 397,
    "h": 373
  },
  "view": {},
  "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}send-record-notification.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Send record notification**:
*** *recipient**:
****** **value type**: constant
****** **value**: `test-user`
*** *title**:
****** **value type**: constant
****** **value**: `Record notification`
*** *description**:
****** **value type**: constant
****** **value**: `This is a record notification`
*** *namespace**:
****** **value type**: constant
****** **value**: `test-namespace`
*** *module**:
****** **value type**: constant
****** **value**: `test-module`
*** *record**:
****** **value type**: constant
****** **value**: `123`
*** *openMode**:
****** **value type**: constant
****** **value**: `modal`
*** *edit**:
****** **value type**: constant
****** **value**: `false`
3. **(3) Done**:
******


# Загрузка файлов с помощью шлюза интеграции (Integration Gateway)
:attachment-path: ../../../_attachments/automation/workflows/examples/
:page-noindex: true

Шлюзы интеграции в сочетании с Workflow позволяют пользователям обрабатывать загрузку файлов с помощью `multipart/form-data`.

.В этом примере мы:
- создадим рабочий процесс, который извлекает загруженный файл,
- создадим эндпоинт шлюза интеграции.

## Создание рабочего процесса

Чтобы создать рабочий процесс, сначала откройте Workflow и создайте новый рабочий процесс.
Основной рабочий процесс требует двух частей, а в примере добавлены несколько дополнительных шагов для прикрепления файла к записи.

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/workflow.png)

.Рабочий процесс требует этих двух шагов:
- **System / on manual** триггер: этот триггер позволяет шлюзу интеграции выполнять конкретный рабочий процесс.
- **Read file from integration gateway** функция:: эта функция извлекает файл из предоставленного HTTP-запроса:
** Аргумент `request` получает HTTP-запрос; при использовании со шлюзами интеграции запрос предоставляется в переменной `request`.
** Аргумент `name` задаёт имя поля `multipart/form-data`,
** Выходной аргумент `file` предоставляет `Reader` файла.

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/step-config.png)

`Reader` содержит содержимое файла и может использоваться по вашему усмотрению.

## Создание нового эндпоинта шлюза интеграции

Чтобы создать новый эндпоинт, перейдите в веб-приложение Admin и нажмите на пункт меню «Integration Gateway», затем нажмите на кнопку btn:[New Route].

Укажите эндпоинт (в примере для загрузки файлов используется `/fup`).
Выберите метод `POST` и обязательно отметьте шлюз как включённый.
Нажмите на кнопку btn:[Submit], чтобы инициализировать эндпоинт.

В новом разделе «filter list» используйте следующие параметры:
- **Prefilters** (предварительные фильтры):
** При желании выберите и включите «Profiler» (это помогает убедиться, что эндпоинт обрабатывается).

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/prefilter.png)

- **Processing** (обработка):
** Выберите и настройте «Workflow processor» (используйте рабочий процесс, подготовленный на предыдущем шаге).

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/processing.png)

- **Postfiltering** (постфильтрация):
** Добавьте «default JSON response», чтобы мы могли видеть статусы ответов и возможные ошибки

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/postfilter.png)

## Тестирование

Для тестирования вы можете отправить HTTP-запрос `POST` на эндпоинт your-lowcode-instance.tld/api/gateway/YOUR*ENDPOINT*HERE.
Вы можете использовать следующий шаблон cURL:

```shell
```
curl -v -X POST http://localhost:18080/api/gateway/fup \
  -F "file=@/path/to/your/file.txt"

В случае успеха ответ должен выглядеть следующим образом:

```txt
```
- Host localhost:18091 was resolved.
- IPv6: ::1
- IPv4: 127.0.0.1
-   Trying [::1]:18091...
- connect to ::1 port 18091 from ::1 port 58275 failed: Connection refused
-   Trying 127.0.0.1:18091...
- Connected to localhost (127.0.0.1) port 18091
> POST /api/gateway/fup HTTP/1.1
> Host: localhost:18091
> User-Agent: curl/8.7.1
> Accept: */*
> Content-Length: 206
> Content-Type: multipart/form-data; boundary=------------------------Caf19XAT8xI5uYCDVXRkRz
> 
- upload completely sent off: 206 bytes
< HTTP/1.1 202 Accepted
< Content-Type: application/json
< Vary: Origin
< Vary: Origin
< X-Request-Id: 4b9179be911f/FwhKqZfCsa-000622
< Date: Tue, 22 Jul 2025 12:47:20 GMT
< Content-Length: 2
< 
- Connection #0 to host localhost left intact
{}%


# Attaching Files to Records
:attachment-path: ../../../_attachments/automation/workflows/examples/
:page-noindex: true

Attaching files to records is straight forward but it requires a few steps.

!!! important
    This example receives files via integration gateway.


.In this example we will:
- Create a new record,
- upload an attachment,
- attach it to the newly created record.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/workflow-base.png",
    "alias": "automation-workflows-examples-attaching-files-workflow-base",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 131,
    "y": 160,
    "w": 1186,
    "h": 394
  },
  "focus": {
    "x": 165,
    "y": 338,
    "w": 1103,
    "h": 173
  },
  "annotations": []
}

Firstly, we need to prepare a new record using the "Compose Record Maker" function.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/make-record.png",
    "alias": "automation-workflows-examples-attaching-files-make-record",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "x": 259,
    "y": 388,
    "w": 180,
    "h": 73
  }]
}

Next, we need to upload the file and prepare an `Attachment`.
The attachment needs to specify the used module field name as well as the record we've prepared earlier.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/upload-attachment.png",
    "alias": "automation-workflows-examples-attaching-files-upload-attachment",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "x": 510,
    "y": 388,
    "w": 180,
    "h": 73
  }]
}

Next, we need to reference the Attachment in the newly prepared record.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/add-reference.png",
    "alias": "automation-workflows-examples-attaching-files-add-reference",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "x": 763,
    "y": 388,
    "w": 180,
    "h": 73
  }]
}

Lastly, we create the record.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/save-record.png",
    "alias": "automation-workflows-examples-attaching-files-save-record",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "x": 1014,
    "y": 388,
    "w": 180,
    "h": 73
  }]
}

The newly created record can be seen on the record list

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/record-list.png",
    "alias": "automation-workflows-examples-attaching-files-record-list",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 346,
    "y": 90,
    "w": 1538,
    "h": 410
  },
  "annotations": []
}


:leveloffset: -1
