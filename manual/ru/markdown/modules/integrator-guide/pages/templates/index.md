# Шаблоны

Шаблоны позволяют вам определить общую структуру документа (например, приветственное email-сообщение или PDF-цитату), которая затем преобразуется в актуальный документ на основе предоставленных данных.

Если вы хотите узнать больше о конкретной теме, обратитесь к подразделам в меню menu:Low-Code Platform Developer Guide[Templates].

## Создание шаблона

Шаблоны создаются и управляются в LowCoooode Admin.

Перейдите в menu:System[Templates] и нажмите на кнопку btn:[New] в правом верхнем углу.

Введите основную информацию: короткое имя, handle, описание и тип шаблона.
Тип шаблона определяет формат шаблона и неявно подразумевает, какие типы документов шаблон может рендерить.

Частичный шаблон используется как часть другого шаблона (например, общий заголовок или подвал) и не может быть отрендерен независимо.
Вы можете преобразовать шаблон из частичного и в частичный.

Нажмите на кнопку btn:[Submit], чтобы подготовить ваш шаблон.

После отправки базовых параметров появятся три новых раздела.

[cols="1s,5a"]
|===

| [#templates-toolbox]#[templates-toolbox,Toolbox](#templates-toolbox,Toolbox)#
|
Toolbox предоставляет фрагменты для наиболее распространенных операций, образец HTML-шаблона и образцы включения частичных шаблонов.

| [#templates-content]#[templates-content,Template content](#templates-content,Template content)#
|
Редактор содержимого шаблона предоставляет простой редактор кода для редактирования вашего шаблона.
HTML-редактор шаблонов реализует подсветку синтаксиса и некоторые другие полезные инструменты, такие как автозаполнение.

| [#templates-preview]#[templates-preview,Preview](#templates-preview,Preview)#
|
Раздел предварительного просмотра позволяет вам проверить, как ваши шаблоны будут выглядеть при рендеринге в актуальный документ.

!!! important
    Обратитесь к [Pdf Renderer](modules/devops-guide/pages/pdf-renderer.md) для настройки рендеринга PDF-документов.


|===

## Содержимое

!!! note
    Продвинутые пользователи могут ознакомиться с https://golang.org/pkg/text/template/[полной документацией по текстовым шаблонам], https://golang.org/pkg/html/template/[полной документацией по HTML-шаблонам] и https://masterminds.github.io/sprig/[документацией по расширенным функциям].


Давайте начнем с копирования стандартного HTML-образца из toolbox.

```html
```
<!DOCTYPE html>
<html>
<head>
  <meta charset='utf-8'>
  <meta http-equiv='X-UA-Compatible' content='IE=edge'>
  <title>Title</title>
  <meta name='viewport' content='width=device-width, initial-scale=1'>
</head>
<body>
  <h1>Hello, world!</h1>
</body>
</html>

Если вам не нужно никакое динамическое содержимое (разные имена для разных контактов), вы можете остановиться после этого шага.
Приведенный выше шаблон действителен и уже может быть использован.

Если вам нужно динамическое содержимое, необходимо рассмотреть дополнительные темы.

### Интерполяция значений

Интерполяция значений позволяет вам определить некоторый заполнитель, который затем заменяется актуальным значением при рендеринге шаблона.

В этом случае такой заполнитель выглядит так:

```
```
{{.name}}

Значение заменяет указанный выше заполнитель на свойство `name` из предоставленного объекта `value`.

Давайте рассмотрим несколько примеров.
Каждый пример сначала определяет объект `value`, а затем заполнитель.

.Пример с простым свойством:
```
```
{
  "name": "Jane"
}

{{.name}}

.Пример с вложенным свойством:
```
```
{
  "contact": {
    "details": {
      "name": "Jane"
    }
  }
}

{{.contact.details.name}}

Полный пример будет выглядеть так:

.Предоставленные данные:
```json
```
{
  "contact": {
    "details": {
      "firstName": "Jane",
      "lastName": "Doe"
    }
  }
}

.Шаблон:
```html
```
<!DOCTYPE html>
<html>
<head>
  <meta charset='utf-8'>
  <meta http-equiv='X-UA-Compatible' content='IE=edge'>
  <title>Title</title>
  <meta name='viewport' content='width=device-width, initial-scale=1'>
</head>
<body>
  <h1>Hello, {{.contact.details.firstName}} {{.contact.details.lastName}}!</h1>
</body>
</html>

### Условный рендеринг

Условный рендеринг позволяет показывать или скрывать разделы отрендеренного документа на основе входных параметров.

.Оператор `if`:
```
```
{{if condition}}
  The condition was true.
{{end}}

.Оператор `if-else`:
```
```
{{if condition}}
  The condition was true.
{{else}}
  The condition was false.
{{end}}

.Оператор `if-else if-else`:
```
```
{{if condition}}
  The condition was true.
{{else if condition}}
  The other condition was true.
{{else}}
  Neither conditions were true.
{{end}}

Часть `condition` — это выражение, возвращающее одно булево значение.
Пример выражения:

```
```
{{if .lead.cost > 1000}}
  The lead {{.lead.name}} was expensive!
{{end}}

.Логические операторы:
[cols="1s,5a"]
|===

| [#tpl-syntax-and]#[tpl-syntax-and,AND](#tpl-syntax-and,AND)#
|
- **Синтаксис**: `a && b`
- **Примечания**: Результат true, если и `a`, и `b` равны `true`.

| [#tpl-syntax-or]#[tpl-syntax-or,OR](#tpl-syntax-or,OR)#
|
- **Синтаксис**: `a \|\| b`
- **Примечания**: Результат true, если `a` или `b` равно `true`.

| [#tpl-syntax-not]#[tpl-syntax-not,NOT](#tpl-syntax-not,NOT)#
|
- **Синтаксис**: `!a`
- **Примечания**: Результат true, если `a` равно `false`, и наоборот.

| [#tpl-syntax-eq]#[tpl-syntax-eq,Equal](#tpl-syntax-eq,Equal)#
|
- **Синтаксис**: `a == b`
- **Примечания**: Результат true, если `a` равно `b`.

| [#tpl-syntax-neq]#[tpl-syntax-neq,Not equal](#tpl-syntax-neq,Not equal)#
|
- **Синтаксис**: `a != b`
- **Примечания**: Результат true, если `a` не равно `b`.

| [#tpl-syntax-lt]#[tpl-syntax-lt,Less than](#tpl-syntax-lt,Less than)#
|
- **Синтаксис**: `a < b`
- **Примечания**: Результат true, если `a` меньше `b`.

| [#tpl-syntax-gt]#[tpl-syntax-gt,Greater than](#tpl-syntax-gt,Greater than)#
|
- **Синтаксис**: `a > b`
- **Примечания**: Результат true, если `a` больше `b`.

| [#tpl-syntax-let]#[tpl-syntax-let,Less equal than](#tpl-syntax-let,Less equal than)#
|
- **Синтаксис**: `a \<= b`
- **Примечания**: Результат true, если `a` меньше или равно `b`.

| [#tpl-syntax-get]#[tpl-syntax-get,Greater equal than](#tpl-syntax-get,Greater equal than)#
|
- **Синтаксис**: `a >= b`
- **Примечания**: Результат true, если `a` больше или равно `b`.

|===

### Работа со списками

Наши шаблоны позволяют довольно легко работать со списками.
Например, вы можете захотеть сгенерировать цитату с множеством позиций.

Синтаксис для перебора списка выглядит так:

```
```
{{range .listOfItems}} 
  {{.itemName}}; {{.itemCost}}$
{{end}}

Если вы предпочитаете указать, в какую переменную сохраняется текущий элемент, используйте этот синтаксис:

```
```
{{range $index, $item := .ListOfItems}}
  {{$item.itemName}}; {{$item.itemCost}}$
{{end}}

### Использование функций

Иногда вам нужно дополнительно обработать данные перед их рендерингом в документ.

Некоторую легкую обработку можно выполнить непосредственно в шаблонизаторе.
Более сложную обработку следует выполнять в коде, который запрашивает рендеринг шаблона.

.Функция вызывается следующим образом:
```
```
{{functionName arg1 arg2 ... argN}}

Переданный аргумент может быть константой или свойством из предоставленных данных.

Вы также можете объединять функции в цепочки.
Когда две функции объединены в цепочку, вывод левой функции передается в качестве аргумента правой функции.

```
```
{{funcA | funcB | ... | funcN}}

.Справочник наиболее распространенных функций:
[cols="1s,5a"]
|===

| [#tpl-syntax-fref-len]#[tpl-syntax-fref-len,Length of](#tpl-syntax-fref-len,Length of)#
|
- **Синтаксис**: `{{len listOfThings}}`
- **Примечания**: Возвращает количество элементов в указанном `listOfThings`.

| [#tpl-syntax-fref-printf]#[tpl-syntax-fref-printf,Format string](#tpl-syntax-fref-printf,Format string)#
|
- **Синтаксис**: `{{printf "pattern goes here" arg1 arg2 ... argn}}`
- **Примечания**: Возвращает форматированную строку, следуя указанному шаблону, используя значения, предоставленные в качестве аргументов.

| [#tpl-syntax-fref-inlineRemote]#[tpl-syntax-fref-inlineRemote,Inline remote file](#tpl-syntax-fref-inlineRemote,Inline remote file)#
|
- **Синтаксис**: `{{inlineRemote "url goes here"}}`
- **Примечания**: Возвращает файл, закодированный в base64, по указанному URL.
Строка форматируется в виде `data:\{mime-type};base64,\{encoded remote file}`.
Полезно для вставки изображений в PDF-документы.

| [#tpl-syntax-fref-trim]#[tpl-syntax-fref-trim,Trim string](#tpl-syntax-fref-trim,Trim string)#
|
- **Синтаксис**: `{{trim "string goes here"}}`
- **Примечания**: Удаляет все пробелы в начале/конце указанной строки.

| [#tpl-syntax-fref-trimSuffix]#[tpl-syntax-fref-trimSuffix,Trim suffix from a string](#tpl-syntax-fref-trimSuffix,Trim suffix from a string)#
|
- **Синтаксис**: `{{trimSuffix "suffix to remove here" "string goes here"}}`
- **Примечания**: Удаляет суффикс из указанной строки.

| [#tpl-syntax-fref-trimPrefix]#[tpl-syntax-fref-trimPrefix,Trim prefix from a string](#tpl-syntax-fref-trimPrefix,Trim prefix from a string)#
|
- **Синтаксис**: `{{trimPrefix "prefix to remove here" "string goes here"}}`
- **Примечания**: Удаляет префикс из указанной строки.

| [#tpl-syntax-fref-upper]#[tpl-syntax-fref-upper,To uppercase](#tpl-syntax-fref-upper,To uppercase)#
|
- **Синтаксис**: `{{upper "string goes here"}}`
- **Примечания**: Преобразует строку в верхний регистр.

| [#tpl-syntax-fref-lower]#[tpl-syntax-fref-lower,To lowercase](#tpl-syntax-fref-lower,To lowercase)#
|
- **Синтаксис**: `{{lower "string goes here"}}`
- **Примечания**: Преобразует строку в нижний регистр.

|===

### Использование частичных шаблонов

Частичные шаблоны позволяют поддерживать согласованность документов, используя общие заголовки и подвалы.
Частичные шаблоны также могут быть полезны при отображении ресурсов LowCoooode, например, отображении записи в таблице.

.Частичные шаблоны включаются следующим образом:
```
```
{{template "partial_handle"}}

`partial_handle` — это handle, который вы использовали при определении частичного шаблона.
Например:

```
```
{{template "email*general*header"}}

Если вашему частичному шаблону нужен доступ к некоторым данным, которые вы предоставили текущему шаблону (тому, который использует частичный шаблон), вам нужно передать второй аргумент в процесс включения частичного шаблона.

```
```
{{template "email*general*header" .property.to.pass}}

Смотрите пример ниже:

```json
```
{
  "contact": {
    "values": {...}
  },
  "account": {
    "values": {...}
  }
}

Если вы хотите передать `contact` в частичный шаблон, вы должны включить ваш частичный шаблон, как показано ниже:

```
```
{{template "partial_handle" .contact}}

Если вы хотите передать все данные в ваш частичный шаблон, вы должны включить ваш частичный шаблон, как показано ниже:

```
```
{{template "partial_handle" .}}

Вы можете получить доступ к `contact` в вашем частичном шаблоне, как показано ниже:

```
```
{{/* In case of the first example */}}
Hello {{.values.FirstName}}

{{/* In case of the second example */}}
Hello {{.contact.values.LastName}} of the {{.account.values.Name}}


## Предварительный просмотр

Раздел предварительного просмотра в нижней части страницы позволяет вам проверить, как будут выглядеть ваши документы после рендеринга шаблона.

Поле ввода должно содержать валидный JSON-объект (полезная нагрузка рендеринга) с двумя корневыми свойствами: `variables` и `options`:

```json
```
{
  "variables": {},<1>
  "options": {}<2>
}
<1> Параметр `variables` определяет, какие данные будут доступны при рендеринге документа.
Структура не определена.
<2> Параметр `options` определяет параметры рендеринга и в настоящее время доступен только для PDF.

.Пример полезной нагрузки рендеринга:
```json
```
{
  "variables": {
    "param1": "value1",
    "param2": {
      "nestedParam1": "value2"
    }
  },
  "options": {
    "documentSize": "A4",
    "contentScale": "1",
    "orientation": "portrait",
    "margin": "0.3"
  }
}

A

!!! important
    Параметры, помеченные #PDF only#, могут использоваться только с PDF-документами и игнорируются в остальных случаях.


.Полный справочник полезной нагрузки рендеринга:
[cols="1s,5a"]
|===

| [#tpl-render-variables]#[tpl-render-variables,`variables`](#tpl-render-variables,`variables`)#
|
- **Тип**: `Object<any>`
- **Описание**: Переменные, которые вы хотите применить к шаблону.
Например, если вы хотите, чтобы `\{\{testing}}` работало, вы должны передать `{"variables": {"testing": "some value"}}`.

| [#tpl-render-options*marginBottom]#[tpl-render-options*marginBottom,`options.marginBottom`](#tpl-render-options_marginBottom,`options.marginBottom`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n; дюймы >`
- **Описание**: Управляет отступом снизу страницы.

| [#tpl-render-options*marginLeft]#[tpl-render-options*marginLeft,`options.marginLeft`](#tpl-render-options_marginLeft,`options.marginLeft`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n; дюймы >`
- **Описание**: Управляет отступом слева страницы.

| [#tpl-render-options*marginRight]#[tpl-render-options*marginRight,`options.marginRight`](#tpl-render-options_marginRight,`options.marginRight`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n; дюймы >`
- **Описание**: Управляет отступом справа страницы.

| [#tpl-render-options*marginTop]#[tpl-render-options*marginTop,`options.marginTop`](#tpl-render-options_marginTop,`options.marginTop`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n; дюймы >`
- **Описание**: Управляет отступом сверху страницы.

| [#tpl-render-options*marginY]#[tpl-render-options*marginY,`options.marginY`](#tpl-render-options_marginY,`options.marginY`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n; дюймы >`
- **Описание**: Управляет отступом сверху и снизу страницы.

| [#tpl-render-options*marginX]#[tpl-render-options*marginX,`options.marginX`](#tpl-render-options_marginX,`options.marginX`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n; дюймы >`
- **Описание**: Управляет отступом слева и справа страницы.

| [#tpl-render-options*margin]#[tpl-render-options*margin,`options.margin`](#tpl-render-options_margin,`options.margin`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n; дюймы >`
- **Описание**: Управляет отступом слева, справа, сверху и снизу страницы.

| [#tpl-render-options*documentSize]#[tpl-render-options*documentSize,`options.documentSize`](#tpl-render-options_documentSize,`options.documentSize`)#
|
#PDF only.#

- **Тип**: `string<A0...A10, B0...B10, C0...C10, ANSI A, ANSI B, ANSI C, ANSI D, ANSI E, junior legal, letter, legal, tabloid>`
- **Описание**: Размер документа в соответствии со стандартом ISO216 для серий A, B и C; стандартом ANSI для ANSI A, B, C и D; и стандартами NA для последних нескольких.

| [#tpl-render-options*documentWidth]#[tpl-render-options*documentWidth,`options.documentWidth`](#tpl-render-options_documentWidth,`options.documentWidth`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n; дюймы >`
- **Описание**: Задает ширину документа, если ни один из пресетов не подходит.

| [#tpl-render-options*documentHeight]#[tpl-render-options*documentHeight,`options.documentHeight`](#tpl-render-options_documentHeight,`options.documentHeight`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n; дюймы >`
- **Описание**: Задает высоту документа, если ни один из пресетов не подходит.

| [#tpl-render-options*contentScale]#[tpl-render-options*contentScale,`options.contentScale`](#tpl-render-options_contentScale,`options.contentScale`)#
|
#PDF only.#

- **Тип**: `string< float; 0 >= n >`
- **Описание**: В каком масштабе должен быть отрендерен документ; большее число => больший контент.

!!! note
    PDF-документы ограничены значением `0 >= n \<= 8`


| [#tpl-render-options*orientation]#[tpl-render-options*orientation,`options.orientation`](#tpl-render-options_orientation,`options.orientation`)#
|
#PDF only.#

- **Тип**: `string<landscape,portrait>`
- **Описание**: В какой ориентации рендерить документ.

|===
