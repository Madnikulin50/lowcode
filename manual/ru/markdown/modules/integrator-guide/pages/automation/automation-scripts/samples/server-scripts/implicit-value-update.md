# Расчёт стоимости лида

Скрипт-пример рассчитывает стоимость лида при его создании **или** обновлении.

.server-scripts/Lead/UpdateCost.js
```js
```
export default {
  label: "Script label",
  description: 'Script description',

  - triggers ({ before }) {
    yield before('create', 'update')
      .for('compose:record')
      .where('module', 'Lead')
      .where('namespace', 'crm')
  },

  async exec ({ $record }, { Compose }) {
    if (!$record.values.LeadSource) {
      return $record
    }

    switch ($record.values.LeadSource) {
      case 'source-a':
        $record.values.LeadCost = 10
        break

      case 'source-b':
        $record.values.LeadCost = 20
        break

      default:
        $record.values.LeadCost = 30
        break
    }

    return $record
  }
}
