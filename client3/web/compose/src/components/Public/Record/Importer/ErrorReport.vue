<template>
  <div class="card">
    <div class="card-body">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordList.import.report.title') }}</label>
        <div class="small ps-2">
          <b>{{ $t('recordList.import.report.startedAt') }}</b>: {{ datify(progress.startedAt) }}
        </div>
        <div class="small ps-2">
          <b>{{ $t('recordList.import.report.finishedAt') }}</b>: {{ datify(progress.finishedAt) }}
        </div>
        <div class="small ps-2">
          <b>{{ $t('recordList.import.report.totalRecords') }}</b>: {{ progress.entryCount }}
        </div>
        <div class="small ps-2">
          <b>{{ $t('recordList.import.report.importedRecords') }}</b>: <span class="text-success">{{ progress.completed }}</span>
        </div>
        <div class="small ps-2">
          <b>{{ $t('recordList.import.report.failedRecords') }}</b>: <span class="text-danger">{{ progress.failed }}</span>
        </div>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordList.import.report.detectedErrors') }}</label>
        <table id="error-list" class="table table-hover mb-0">
          <thead class="table-outline-secondary">
            <tr>
              <th>{{ $t('recordList.import.report.error') }}</th>
              <th>{{ $t('recordList.import.report.count') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, idx) in errorList" :key="idx">
              <td class="border-top">{{ item.k }}</td>
              <td class="border-top">{{ item.v }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordList.import.report.failedEntries') }}</label>
        <div v-for="(ee, ix) in progress.failLog.records" :key="ix" class="small ps-2">
          <template v-if="ee.length == 1 || ee[0] === ee[1]">
            <b>{{ $t('recordList.import.report.failedEntriesLine') }}</b>: {{ ee[0] }}
          </template>
          <template v-else>
            <span>
              <b>{{ $t('recordList.import.report.failedEntriesLines') }}</b>:
              {{ ee[0] }}
              <font-awesome-icon :icon="['fas', 'arrow-right']" size="sm" class="mx-1" />
              {{ ee[1] }}
            </span>
          </template>
        </div>
      </div>
    </div>
    <div class="card-footer">
      <button
        class="btn btn-outline-secondary float-end"
        @click="$emit('close')"
      >
        {{ $t('label.close') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { fmt } from 'corteza-lib/js/dist'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  session: { type: Object, required: true, default: () => ({}) },
  noPool: { type: Boolean, default: false },
})

defineEmits(['close'])

const expandedLogs = computed(() => ({}))
const completedLogs = computed(() => ({}))

const progress = computed(() => props.session.progress)

const errorList = computed(() => {
  return Object.entries(progress.value.failLog.errors)
    .map(([k, v]) => ({ k, v }))
})

function datify(dt) {
  return fmt.fullDateTime(dt)
}
</script>

<style lang="scss">
.progress-label {
  font-size: 15px;
}
.fit {
  white-space: nowrap;
  width: 15%;
}
.pointer {
  cursor: pointer;
}
</style>
