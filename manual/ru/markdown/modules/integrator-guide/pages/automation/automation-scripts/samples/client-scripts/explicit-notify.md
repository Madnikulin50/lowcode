# Уведомление по запросу

Этот пример запрашивает у пользователя ввод значения, а затем отображает его в виде уведомления.

.client-scripts/compose/crm/Contact/CollectValue.js
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

  async exec ({ $record }, { Compose, ComposeUI }) {

    const value = window.prompt('Please insert a value')
    if (!value) {
      ComposeUI.warning('No value provided')
      return false
    }


    ComposeUI.success(`Hi! You've entered ${value}!`)
  }
}
