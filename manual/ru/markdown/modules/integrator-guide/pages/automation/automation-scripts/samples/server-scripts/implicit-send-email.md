# Уведомление владельца об изменении

Скрипт-пример получает владельца лида и отправляет ему email.

!!! important
    Убедитесь, что ваша конфигурация SMTP работает.


.server-scripts/Lead/NotifyChange.js
```js
```
export default {
  label: "Script label",
  description: 'Script description',

  - triggers ({ after }) {
    yield after('update')
      .for('compose:record')
      .where('module', 'Lead')
      .where('namespace', 'crm')
  },

  async exec ({ $record }, { Compose, System }) {
    let emailContent
    let emailSubject

    if (!$record.ownedBy) {
      return false
    }

    const owner = await System.findUserByID($record.ownedBy)


    await Compose.sendMail(
      owner.email,
      emailSubject,
      { html: emailContent }
    )
  }
}
