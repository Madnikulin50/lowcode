# Отправка email контакту

Скрипт-пример отправляет email контакту, для которого он был вызван.

!!! important
    Убедитесь, что ваша конфигурация SMTP работает.


.server-scripts/Contact/SendMail.js
```js
```
export default {
  label: "Script label",
  description: 'Script description',

  - triggers ({ on }) {
    yield on('manual')
      .for('compose:record')
      .where('module', 'Contact')
      .where('namespace', 'crm')
      .uiProp('app', 'compose')
  },

  async exec ({ $record }, { Compose }) {
    let emailContent
    let emailSubject


    if (!$record.values.Email) {
      return false
    }

    await Compose.sendMail(
      $record.values.Email,
      emailSubject,
      { html: emailContent }
    )
  }
}
