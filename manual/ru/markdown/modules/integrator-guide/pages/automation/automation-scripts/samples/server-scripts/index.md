# Серверные скрипты

## Универсальный неявный скрипт

!!! note
    Используйте этот универсальный шаблон, чтобы покрыть 90% случаев использования.


.Универсальный неявный скрипт:
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


<a id="ss-before-record"></a>
## Перед сохранением записи

!!! tip
    .Примеры использования:
    * *расчёт сложных полей*
    * *проверка значений*
    * *создание записи в журнале изменений*


!!! important
    Лучше всего использовать выражения полей для большинства проверок значений и расчётов значений.
    
    *DevNote*: привести несколько ссылок.


.Перед сохранением записи:
```js
```
export default {
  label: "label goes here",
  description: "description goes here",

  triggers ({ before }) {
    return before('create', 'update')
      .for('compose:record')
      .where('module', 'module goes here')
      .where('namespace', 'namespace goes here')
  },

  async exec ({ $record }, { Compose }) {

    return $record
  },
}


## После сохранения записи

!!! tip
    .Примеры использования:
    * *уведомить владельца об изменении*
    * *отправить изменение во внешнюю интеграцию*
    * *создать запись в журнале изменений*


!!! important
    Если вы хотите *изменить значение*, вам следует использовать <<ss-before-record>> вместо этого.


.После сохранения записи:
```js
```
export default {
  label: "label goes here",
  description: "description goes here",

  triggers ({ after }) {
    return after('create', 'update')
      .for('compose:record')
      .where('module', 'module goes here')
      .where('namespace', 'namespace goes here')
  },

  async exec ({ $record }, { Compose }) {
  },
}


## Универсальное действие по клику

!!! tip
    .Примеры использования:
    * *потоки OAuth*
    * *отправка уведомлений*
    * *рендеринг документов*


.Универсальное действие по клику:
```js
```
export default {
  label: "label goes here",
  description: "description goes here",

  triggers ({ on }) {
    return on('manual')
      .uiProp('app', 'compose')
  },

  async exec (args, { Compose }) {
  },
}


## Обработка записи по клику

!!! tip
    .Примеры использования:
    * *отправить email контакту записи*
    * *отправить email владельцу записи*
    * *сгенерировать отчёт по конкретной записи*


.Обработка записи по клику:
```js
```
export default {
  label: "label goes here",
  description: "description goes here",

  triggers ({ on }) {
    return on('manual')
      .for('compose:record')
      .where('module', 'module goes here')
      .where('namespace', 'namespace goes here')
      .uiProp('app', 'compose')
  },

  async exec ({ $record }, { Compose }) {
  },
}


## Ответ на HTTP-запрос

!!! tip
    .Примеры использования:
    * *создание записей из внешней формы*
    * *отслеживание утверждения/подписания документов*
    * *приём платежей по кредитным картам*


!!! important
    Вам нужно будет *добавить пакет `base-64`*.


.Ответ на HTTP-запрос:
```js
```
import base64 from 'base-64'

export default {
  label: "label goes here",
  description: "description goes here",

  security: {
    runAs: 'user goes here',
  },

  triggers ({ on }) {
    return on('request')
      .where('request.path', '/path')
      .where('request.method', 'POST')
      .for('system:sink')
  },

  async exec ({ $request, $response }) {
    const body = JSON.parse(base64.decode($request.rawBody))


    $response.status = 200
    $response.header = { 'Content-Type': ['application/json'] }
    $response.body = JSON.stringify({ result: 'example' })

    return $response
  }
}
