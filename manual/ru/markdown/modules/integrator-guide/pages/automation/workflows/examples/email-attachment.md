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
