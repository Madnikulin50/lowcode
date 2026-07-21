<template>
  <div class="card shadow-sm mt-3 overflow-hidden">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body p-0">
      <table class="table table-hover mb-0">
        <thead class="table-light">
          <tr>
            <th>{{ $t('resourceType') }}</th>
            <th>{{ $t('eventType') }}</th>
            <th>{{ $t('constraints') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="trigger in triggers" :key="trigger.ID">
            <td>{{ formatResourceType(trigger.resourceType) }}</td>
            <td>{{ trigger.eventType }}</td>
            <td>
              <samp v-for="(c, index) in trigger.constraints" :key="index">
                <template v-if="c.name">
                  {{ c.name[0].toUpperCase() + c.name.slice(1).toLowerCase() }} {{ c.op }} "{{ c.values.join(' or ') }}"
                </template>
                <code v-if="index < trigger.constraints.length - 1">{{ $t('and') }}</code>
              </samp>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps({
  triggers: { type: Array, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
})

function formatResourceType(rt) {
  return rt.split(':').map(s => {
    return s[0].toUpperCase() + s.slice(1).toLowerCase()
  }).join(' ')
}
</script>
