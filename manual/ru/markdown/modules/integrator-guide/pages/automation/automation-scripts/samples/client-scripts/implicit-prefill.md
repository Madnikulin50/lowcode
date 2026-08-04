# Предзаполнение значений

Этот пример предзаполняет некоторые значения записи, если они не указаны.

!!! note
    Это также можно сделать с помощью настройки значения поля по умолчанию в модуле.


.client-scripts/compose/crm/Contact/Prefill.js
```js
```
export default {
  label: "Script label",
  description: 'Script description',

  - triggers ({ before }) {
    yield before('formSubmit')
      .for('ui:compose:record-page')
      .where('module', 'Request')
      .where('namespace', 'crm')
  },

  async exec ({ $record, $module }, { Compose }) {

    if (['crm'].includes($module.namespace.slug.toLowerCase())) {
      return
    }

    if (['lead'].includes($module.handle.toLowerCase())) {
      return
    }

    const defaults = await Compose.findFirstRecord('Defaults')

    for (const k in $record.values) {
      if (!$record.values[k]) {
        $record.values[k] = defaults.values[k]
      }
    }

    return $record
  }
}
