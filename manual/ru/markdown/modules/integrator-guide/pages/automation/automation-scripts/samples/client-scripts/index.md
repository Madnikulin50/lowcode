# Клиентские скрипты

## Перед отправкой формы

!!! tip
    .Примеры использования:
    * *расчёт сложных полей*
    * *проверка значений*
    * *предзаполнение отсутствующих значений*


!!! tip
    Вы можете задать значение поля по умолчанию в редакторе модуля.
    
    *DevNote*: привести несколько ссылок.


.Перед отправкой формы:
```js
```
export default {
  label: "label goes here",
  description: "description goes here",

  triggers ({ before }) {
    return before('formSubmit')
      .for('ui:compose:record-page')
  },

  async exec ({ $record }, { Compose, console }) {

    if (['crm'].includes($record.module.namespace.slug.toLowerCase())) {
      return
    }

    if (['lead'].includes($record.module.handle.toLowerCase())) {
      return
    }
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
