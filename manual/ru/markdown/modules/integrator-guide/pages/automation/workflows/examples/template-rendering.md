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
