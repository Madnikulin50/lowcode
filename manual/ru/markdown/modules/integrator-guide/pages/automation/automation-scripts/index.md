# Скрипты автоматизации

Скрипт автоматизации (далее — скрипт) — это фрагмент кода, который позволяет реализовать пользовательскую бизнес-логику.
Скрипты автоматизации написаны на чистом JavaScript с поддержкой node-пакетов.

!!! important
    Начиная с [202103](modules/ROOT/pages/changelog/202103.md), с появлением рабочих процессов скрипты автоматизации отключены по умолчанию.


Посмотрите наши примеры (menu:Low-Code Platform Developer Guide[Automation scripts,Samples]) и [sample-project,пример проекта](#sample-project,пример проекта), чтобы начать.

!!! tip
    Проверьте, можно ли реализовать вашу логику с помощью [Workflows](modules/integrator-guide/pages/automation/automation-scripts/automation/workflows/index.md).
    Рабочий процесс способен [выполнять скрипты автоматизации](modules/integrator-guide/pages/automation/automation-scripts/automation/workflows/automation-scripts.md)


Существует две категории скриптов автоматизации: **серверные скрипты** и **клиентские скрипты**.

[cols="1a,5a"]
|===
| [#automation-scripts-server]#[automation-scripts-server,Серверные скрипты](#automation-scripts-server,Серверные скрипты)#
| Серверные скрипты выполняются на **сервере Corredor LowCoooode**.

.Используйте серверные скрипты, когда:
- вы работаете с **чувствительными данными**,
- вы взаимодействуете с **внешними API**,
- **прерывание скрипта пользователем должно быть невозможным**.

.Примеры использования:
- **создание дополнительных записей** на основе текущих данных,
- **отправка email-уведомлений**,
- выполнение **статистического анализа**.

| [#automation-scripts-client]#[automation-scripts-client,Клиентские скрипты](#automation-scripts-client,Клиентские скрипты)#
| Клиентские скрипты выполняются в браузере клиента (пользовательском агенте).

.Используйте клиентские скрипты, когда:
- вам нужно **взаимодействовать с пользователем**,
- вы выполняете **проверку данных**,
- вы **вставляете** значения по умолчанию.

.Примеры использования:
- **запросить у пользователя** подтверждение отправки формы,
- **проверить форму**, отправленную пользователем,
- **перенаправить пользователя** после отправки формы,
- открыть внешнюю веб-страницу.

!!! important
    Клиентские скрипты менее безопасны, так как выполняются в браузере пользователя.
    Любые встроенные учётные данные легко доступны, выполнение легко прервать.

|===

Какой из них использовать?
Если вам нужно взаимодействовать с пользователем (показать уведомление, запросить подтверждение), используйте **клиентские скрипты**, в остальных случаях — **серверные скрипты**.

## Структура файлов

Чтобы начать писать скрипты автоматизации, сначала необходимо определить соответствующую структуру файлов.

.Три основные части:
1. `package.json` определяет метаданные, а также ваши зависимости.
1. `/server-scripts` содержит набор скриптов автоматизации, которые выполняются на сервере Corredor.
1. `/client-scripts` содержит набор скриптов автоматизации, которые будут выполняться внутри веб-приложения.
** Каждый подкаталог внутри `/client-scripts` определяет бандл.
Когда веб-приложение загружается, оно получает бандл, назначенный ему (это делается автоматически).

Как `/server-scripts`, так и `/client-scripts` предполагают, что все файлы внутри — это скрипты автоматизации с корректными сигнатурами.
При определении файла со служебными функциями переместите его в `/shared` или определите `/util` (или аналогичный) в корне проекта.


.Пример структуры файлов, содержащей все доступные части:
```text
```
package.json
...
/server-scripts
  ...
/client-scripts
  /admin
      ...
  /compose 
      ...
  /shared
      ...

`...` означает, что вы можете структурировать файлы по своему усмотрению.
Мы рекомендуем группировать скрипты автоматизации по их содержанию; например, скрипты, работающие с лидами, должны находиться в каталоге `/Lead`.

`/admin` и `/compose` содержат скрипты, специфичные для каждого веб-приложения (как обсуждалось ранее).

`/shared` содержит код, который клиентские скрипты могут переиспользовать.

## Скрипт автоматизации

!!! caution
    В одном файле можно определить только один скрипт автоматизации.


.Так выглядит корректный скрипт автоматизации:
```js
```
{
  label: '...',

  description: '...',

  security: {...},

  triggers: (t) {...},

  exec: (args, ctx) {...};
}

.Вы можете использовать этот шаблон для большинства случаев использования:
```js
```
export default {
  label: "label goes here",
  description: "description goes here",

  triggers ({ before, after, on, at }) {
    return before('event goes here')
      .where('constraint goes here')
  },

  async exec(args, ctx) {
  },
}


!!! note
    См. [примеры серверных скриптов](modules/integrator-guide/pages/automation/automation-scripts/automation/automation-scripts/samples/server-scripts/index.md) и [примеры клиентских скриптов](modules/integrator-guide/pages/automation/automation-scripts/automation/automation-scripts/samples/client-scripts/index.md) для подробностей.


### Аргументы выполнения

Аргументы выполнения (параметр `args` функции `exec`) содержат основные данные, с которыми может работать ваша логика.
Данные различаются в зависимости от события, которое запустило скрипт автоматизации.
Например, события, связанные с пользователем, определяют, какой пользователь задействован, а события, связанные с записью, определяют задействованную запись и модуль.

!!! note
    Обратитесь к [Resource Events](modules/integrator-guide/pages/automation/automation-scripts/automation/automation-scripts/references/resource-events.md) за полным справочником данных, которые предоставляет каждое событие.


!!! important
    Аргументы клиентского скрипта передаются через *ссылки* на исходные объекты, то есть *любое изменение* параметра аргумента *отражается* на исходном объекте.
    
    Аргументы серверного скрипта передаются *копией* исходного объекта, то есть изменения *не отражаются* на исходном объекте.


<a id="execution-context"></a>
### Контекст выполнения

Контекст выполнения (параметр `ctx` функции `exec`) содержит контекстную информацию и утилиты, которые можно использовать во время выполнения скрипта.

.Параметры контекста выполнения:
[cols="1m,5a"]
|===
| [#context-console]#[context-console,ctx.console](#context-console,ctx.console)#
|
Параметр `console` определяет логгер.

!!! note
    При выполнении клиентских скриптов это встроенный `window.console`; при выполнении серверных скриптов это логгер `Pino`.


| [#context-log]#[context-log,ctx.log](#context-log,ctx.log)#
|
`ctx.log` — это псевдоним для [context-console,`ctx.console`](#context-console,`ctx.console`)

| [#context-authuser]#[context-authuser,ctx.$authUser](#context-authuser,ctx.$authUser)#
|
`$authUser` — это ссылка на [вызывающего пользователя](modules/integrator-guide/pages/automation/automation-scripts/automation/automation-scripts/index.md#involing-user).

| [#context-systemapi]#[context-systemapi,ctx.SystemAPI](#context-systemapi,ctx.SystemAPI)#
|
`SystemAPI` — клиент API LowCoooode.

| [#context-composeapi]#[context-composeapi,ctx.ComposeAPI](#context-composeapi,ctx.ComposeAPI)#
|
`ComposeAPI` — клиент API Low Code LowCoooode.

| [#context-system]#[context-system,ctx.System](#context-system,ctx.System)#
|
`System` определяет набор вспомогательных методов для работы с основной системой

| [#context-compose]#[context-compose,ctx.Compose](#context-compose,ctx.Compose)#
|
`Compose` определяет набор вспомогательных методов для работы с ресурсами Low Code LowCoooode.

| [#context-composeui]#[context-composeui,ctx.ComposeUI](#context-composeui,ctx.ComposeUI)#
|
`ComposeUI` определяет набор вспомогательных методов для работы с пользовательским интерфейсом Low Code LowCoooode.

| [#context-frontendbaseurl]#[context-frontendbaseurl,ctx.frontendBaseURL](#context-frontendbaseurl,ctx.frontendBaseURL)#
|
`frontendBaseURL` определяет базовый URL-адрес, по которому работают фронтенд-веб-приложения.
Это полезно при генерации URL-адресов, указывающих на приложения LowCoooode (например, ссылка на только что созданный лид).

|===

### Результат выполнения

Обратитесь к подразделу [детали потока выполнения](modules/integrator-guide/pages/automation/automation-scripts/automation/execution-flow-details.md) за сведениями о том, как результат выполнения влияет на систему.

```js
```
!!! note
    *DevNote* задокументируйте итераторы и их особую ошибку `Aborted`
    
    exec (args, ctx) {
      throw new Error('Aborted')
      // OR
      return false
    }
    ----


[cols="2s,5a"]
|===
| [#exec-result-terminate]#[exec-result-terminate,Выполнение прекращается](#exec-result-terminate,Выполнение прекращается)#
|
Скрипт автоматизации **прекращается** при возникновении ошибки.
Это обычно приводит к прекращению исходной операции.

.Пример:
```js
```
export default {
  trigger (t) {...}

  exec (args, ctx) {
    throw new Error('Oh no, something went wrong')
  }
}

| [#exec-result-unknown]#[exec-result-unknown,Выполнение успешно](#exec-result-unknown,Выполнение успешно)#
|
Во всех остальных случаях скрипт автоматизации **успешен**.
Возвращаемое значение является нулевым значением, если это `null` или `undefined`.

.Пример:
```js
```
export default {
  trigger (t) {...}

  exec ({ $record }, ctx) {
    return $record
  }
}
|===

## Триггеры автоматизации

Триггеры автоматизации (далее — триггеры) управляют временем выполнения конкретного скрипта автоматизации.
```js
```
!!! important
    Триггеры автоматизации вычисляются в изолированном контексте, который не допускает внешних данных (переменных или импортов).
    
    .Это не будет работать:
    const MOD_NAME = 'Contact'
    
    export default {
      triggers ({ on }) {
        return on('manual')
          .for(MOD_NAME) // 👈 we're referencing the constant here
      },
      exec (args, ctx) {...},
    }
    ----


.Три основные части триггера:
1. событие, которое определяет, на какие системные события реагирует триггер,
1. [trigger-define-resource,ресурс](#trigger-define-resource,ресурс), который определяет, для какого системного ресурса срабатывает триггер,
1. [trigger-define-constraint,ограничение](#trigger-define-constraint,ограничение), которое определяет, как должно выглядеть событие, чтобы триггер сработал.

.Доступные типы триггеров:
[cols="1a,5a"]
|===
| [#trigger-type-explicit]#[trigger-type-explicit,Явный](#trigger-type-explicit,Явный)#
| Они **явно запускаются** нажатием **кнопки**.

Используйте явные триггеры, когда хотите **вручную инициировать действие**, например поток аутентификации OAuth, перенаправление на внешний ресурс или экспорт данных.

| [#trigger-type-implicit]#[trigger-type-implicit,Неявный](#trigger-type-implicit,Неявный)#
| Они **запускаются неявно** на основе **системных событий**.

Используйте неявные триггеры, когда хотите, чтобы действие **выполнялось автоматически** при запуске другим действием или процессом; например, отправка email при регистрации нового пользователя или добавление записи в журнал изменений при изменении содержимого.

Обратитесь к [ресурсам и событиям](modules/integrator-guide/pages/automation/automation-scripts/automation/automation-scripts/references/resource-events.md) за полным списком событий, которые можно слушать.

| [#trigger-type-deferred]#[trigger-type-deferred,Отложенный](#trigger-type-deferred,Отложенный)#
|
!!! important
    Отложенные триггеры можно использовать только в серверных скриптах, и они требуют явного контекста безопасности.


Система запускает их **в определённый момент в будущем**; либо **периодически** (определяется **cron-выражениями**), либо **по метке времени** (используется формат **ISO 8601**, `YYYY-MM-DDThh:mm:ssZ`).


Используйте отложенные триггеры, когда хотите, чтобы действие **периодически повторялось** или **выполнялось в определённый момент в будущем**. Примерами такого использования являются повторяющиеся платежи или отправка праздничных рассылок вашим подписчикам.

!!! note
    Планировщик срабатывает *раз в минуту*, поэтому это максимальная точность, которую поддерживает LowCoooode.


| [#trigger-type-sink]#[trigger-type-sink,Приёмник (sink)](#trigger-type-sink,Приёмник (sink))#
|
!!! important
    Триггеры приёмника можно использовать только в серверных скриптах, и они требуют явного контекста безопасности.

Они запускаются системой **при получении запроса**; либо **HTTP**, либо **email**.

Используйте триггеры приёмника, когда хотите **отвечать на запросы**; например **вебхуки для внешних сервисов** или **пользовательские API-эндпоинты**, например **сбор данных из внешних форм**, **отслеживание изменений внешних документов** и **приём платежей**.

!!! note
    Рекомендуется по возможности использовать интерфейс REST API.

|===

<a id="trigger-define-resource"></a>
### Определение ресурса

Чтобы определить, для какого ресурса должен срабатывать триггер (например, модуль, пользователь, роль), мы используем метод `.for('resource:goes:here')`.

.Пример указания ресурса:
```js
```
triggers ({ before }) {
  return before('create', 'update')
    .for('compose:record')
},

Обратитесь к [ресурсам и событиям](modules/integrator-guide/pages/automation/automation-scripts/automation/automation-scripts/references/resource-events.md) за полным списком доступных ресурсов и поддерживаемых событий.

<a id="trigger-define-constraint"></a>
### Определение ограничения

!!! note
    Обратитесь к [ограничениям ресурсов](modules/integrator-guide/pages/automation/automation-scripts/automation/automation-scripts/references/resource-constraints/index.md) за списком доступных свойств ограничений для каждого ресурса.


Чтобы определить, как должно выглядеть событие (например, название модуля, email пользователя, handle роли), мы используем метод `.where(property, operator, value)` или его вариант.

.Подробнее:
- При передаче двух аргументов первый задаёт **свойство**, а второй — **значение**.
Используется оператор [constraint-operator-eq,равенства по умолчанию](#constraint-operator-eq,равенства по умолчанию).
- При передаче трёх аргументов первый задаёт **свойство**, второй — оператор сравнения [trigger-comparison-operators,оператор](#trigger-comparison-operators,оператор), а третий — **значение**.

.Пример объединения ограничений в цепочку:
```js
```
triggers ({ before }) {
  return before('create', 'update')
    .for('compose:record')
    .where('module', 'Lead')
    .where('namespace', 'crm')
},

<a id="trigger-comparison-operators"></a>
.Доступные операторы сравнения:
[cols="1s,5a"]
|===
| [#constraint-operator-eq]#[constraint-operator-eq,Равно (по умолчанию)](#constraint-operator-eq,Равно (по умолчанию))#
|
- `eq`
- `=`
- `==`
- `===`

| [#constraint-operator-neq]#[constraint-operator-neq,Не равно](#constraint-operator-neq,Не равно)#
|
- `not eq`
- `ne`
- `!=`
- `!==`

| [#constraint-operator-prt]#[constraint-operator-prt,Частичное совпадение](#constraint-operator-prt,Частичное совпадение)#
|
- `like`

.Поддерживаемые подстановочные знаки:
- **один или несколько символов**: `%`, `*`,
- **один символ**: `_`, `?`.

| [#constraint-operator-nprt]#[constraint-operator-nprt,Частичное несовпадение](#constraint-operator-nprt,Частичное несовпадение)#
|
- `not like`

.Поддерживаемые подстановочные знаки:
- **один или несколько символов**: `%`, `*`,
- **один символ**: `_`, `?`.

| [#constraint-operator-regex]#[constraint-operator-regex,Совпадение по регулярному выражению](#constraint-operator-regex,Совпадение по регулярному выражению)#
|
- `~`

| [#constraint-operator-nregex]#[constraint-operator-nregex,Несовпадение по регулярному выражению](#constraint-operator-nregex,Несовпадение по регулярному выражению)#
|
- `!~`

|===

### Соглашения

[cols="2s,5a"]
|===
| [#trigger-destructuring]#[trigger-destructuring,Используйте деструктуризацию объектов](#trigger-destructuring,Используйте деструктуризацию объектов)#
|
Деструктуризация объектов помогает сократить код.

.Пример:
```js
```
triggers (t) {
  return t.after('create')
    .for('compose:record')
    .where('module', 'super*secret*module')
},

triggers ({ after }) { // 👈 this thing
  return after('create')
    .for('compose:record')
    .where('module', 'super*secret*module')
},

| [#constrained-constraints]#[constrained-constraints,Используйте ограничения](#constrained-constraints,Используйте ограничения)#
|
Нестрогие ограничения могут привести к нежелательным побочным эффектам, например к запуску скрипта при создании записи в другом пространстве имён.

.Пример:
```js
```
triggers (t) {
  return t.after('create')
    .for('compose:record')
    .where('module', 'super*secret*module')
},

triggers ({ after }) {
  return after('create')
    .for('compose:record')
    .where('module', 'super*secret*module')
    .where('namespace', 'super*secret*namespace') // 👈 this thing
},


|===

## Контекст безопасности

<a id="involing-user"></a>
### Вызывающий пользователь

!!! important
    *Отложенные* и *sink*-скрипты требуют указать контекст безопасности, так как вызывающий пользователь неизвестен.


Вызывающий пользователь — это человек, который выполняет действие, запускающее выполнение скрипта.
Например, вы нажимаете кнопку, поэтому вы и являетесь вызывающим пользователем.

Указав вызывающего пользователя, скрипт автоматизации может получить доступ к некоторым ресурсам, к которым фактический вызывающий пользователь может не иметь доступа, например к личной информации клиентов.

!!! caution
    Вы можете установить вызывающего пользователя только для серверных скриптов.


.Пример определения вызывающего пользователя:
```js
```
export default {
  trigger (t) {...}

  security: 'some-user-identifier-here',

  exec (args, ctx) {...}
}

В качестве идентификатора пользователя можно использовать handle, email или ID пользователя.
Мы рекомендуем использовать email или handle.

!!! tip
    Хорошей практикой является создание нового системного пользователя, целью которого является выполнение скриптов, где бы это ни требовалось.


### Разрешение и запрет выполнения скриптов

Контекст безопасности позволяет предотвратить выполнение конкретных операций конкретными пользователями. Каждому пользователю назначается роль, определяющая степень контроля, с которой он может работать.
Например, вы можете запретить обычным пользователям подписывать документы или отправлять коммерческие предложения.

.Используйте эти свойства при определении контекста:
- `allow` определяет, каким ролям разрешён доступ к скрипту автоматизации,
- `deny` определяет, каким ролям запрещён доступ к скрипту автоматизации,

!!! important
    Это доступно только для явных скриптов.
    Для всех остальных типов скриптов это игнорируется


.Пример разрешения доступа:
```js
```
export default {
  trigger (t) {...}

  security: {
    allow: ['administrator', 'superuser'],
  },

  exec (args, ctx) {...}
}

.Пример запрета доступа:
```js
```
export default {
  trigger (t) {...}

  security: {
    deny: ['client', 'lead'],
  },

  exec (args, ctx) {...}
}

<a id="sample-project"></a>
## Пример настройки

.Структура файлов (исходные коды приведены ниже):
```text
```
/ .gitignore
/ .eslintrc.js
/ .mocharc.js
/ package.json

/ server-scripts
    / Sample.js
    / Sample.test.js
    / ...
/ client-scripts
    / ....

.gitignore:
```gitignore
```
.vscode
node_modules
.nyc_output
coverage
yarn-error.log


.eslintrc.js:
```js
```
module.exports = {
  root: false,
  env: {
    node: true,
    es6: true,
  },
  extends: [
    'standard',
  ],
}


.mocharc.js:
```js
```
module.exports = {
  require: [
    'esm',
  ],
  'full-trace': true,
  bail: true,
  recursive: true,
  extension: ['.test.js'],
  spec: [
    'client-scripts/***/**.test.js',
    'server-scripts/***/**.test.js',
  ],
  'watch-files': [ 'src/**' ],
}


.package.json:
```json
```
{
  "scripts": {
    "lint": "eslint {server-scripts,client-scripts}/***/** --ignore-pattern *.test.js",
    "test:unit": "mocha",
    "test:unit:cc": "nyc mocha"
  },
  "devDependencies": {
    "chai": "^4.2.0",
    "eslint": "^6.8.0",
    "eslint-config-standard": "^14.1.0",
    "eslint-plugin-import": "^2.18.2",
    "eslint-plugin-node": "^10.0.0",
    "eslint-plugin-promise": "^4.2.1",
    "eslint-plugin-standard": "^4.0.1",
    "esm": "^3.2.25",
    "mocha": "^7.0.1",
    "nyc": "^14.1.1",
    "sinon": "^8.1.1"
  },
  "nyc": {
    "all": true,
    "reporter": [
      "lcov",
      "text"
    ],
    "include": [
      "client-scripts/***/**.js",
      "server-scripts/***/**.js"
    ],
    "exclude": [
      "***/**.test.js"
    ],
    "check-coverage": true,
    "per-file": true,
    "branches": 0,
    "lines": 0,
    "functions": 0,
    "statements": 0
  }
}


.Sample.js
```js
```
export default {
  /* istanbul ignore next */
  trigger ({ before }) {
    return before('create')
  },

  exec () {
    return 'Hello World!'
  }
}


```js
```
!!! tip
    Обратите внимание на следующую часть:
    
    // vv this line here vv
    /* istanbul ignore next */
    trigger ({ before }) {
      return before('create')
    },
    ----
    
    `istanbul ignore next` исключает следующую функцию из отчёта о покрытии.


.Sample.test.js
```js
```
import { expect } from 'chai'
import Sample from './Sample'

describe(__filename, () => {
  describe('Sample exec result', () => {
    it('should return a string', () => {
      expect(Sample.exec()).to.eq("HelloWorld")
    })
  })
})


.Приведённый выше `package.json` определяет три скрипта:
- `lint` проверяет код в соответствии со стандартом ES6 по умолчанию (может быть настроен; см. [здесь](https://eslint.org/docs/rules/)),
- `test:unit` запускает модульные тесты, определённые в файлах `.test.js` (может быть настроен в файле `.mocharc.js`),
- `test:unit:cc` запускает модульные тесты с покрытием кода.

```bash
```
!!! note
    Отчёт о покрытии кода генерируется в каталог `coverage`.
    Для HTML-отчёта просмотрите каталог `coverage/lcov-report`.
    
    Обычно для этого используется пакет https://www.npmjs.com/package/http-server[http-server], но достаточно и простого «Открыть в <название браузера>».
    
    `http-server coverage/lcov-report`
    ----
