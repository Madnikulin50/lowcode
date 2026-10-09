<template>
  <module-records
    v-slot="{ value }"
    :modules="payloadValues"
  >
    <p
      v-for="(v, vi) in value.value"
      :key="vi"
      class="mb-0"
      :class="{ 'mt-1': vi > 0 }"
    >
      {{ v }}
    </p>
  </module-records>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'request', keyPrefix: 'view.correct' } })
import { computed } from 'vue'
import ModuleRecords from '../../Common/ModuleRecords.vue'

const props = defineProps({
  request: { type: Object, required: true },
})

const payload = computed(() => {
  const { payload = [] } = props.request || {}
  return payload[0]
})

const payloadValues = computed(() => {
  const { modules = {} } = payload.value || {}

  return Object.entries(modules).map(([moduleID, { module, namespace, records = {} }]) => {
    records = Object.entries(records).map(([recordID, { values = {} }]) => {
      values = Object.entries(values).map(([name, value = []]) => {
        return { name, value }
      })
      return { recordID, values }
    })
    return { module, namespace, moduleID, records }
  })
})
</script>