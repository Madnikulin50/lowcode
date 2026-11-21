<template>
  <div v-if="showModal" class="modal fade show d-block" tabindex="-1">
    <div class="modal-dialog modal-dialog-centered modal-lg">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ $t('recordList.summaries.customSummaries.label') }}</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" @click="onModalHide"></button>
        </div>
        <div class="modal-body p-0">
          <div class="card pt-0 border-0">
            <div class="card-body">
              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('recordList.summaries.param.label.label') }}</label>
                <input v-model="summary.label" class="form-control" :placeholder="$t('recordList.summaries.param.label.placeholder')" />
              </div>
              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('recordList.summaries.param.field.label') }}</label>
                <c-input-select v-model="summary.field" :options="recordListModuleFields" :reduce="field => field.name" :placeholder="$t('recordList.summaries.param.field.placeholder')" />
              </div>
              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('recordList.summaries.param.metric.label') }}</label>
                <c-input-select v-model="summary.metric" :options="summaryMetrics" :reduce="m => m.value" :placeholder="$t('recordList.summaries.param.metric.placeholder')" />
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer d-flex w-100 align-items-center">
          <c-input-confirm v-if="summaryIndex !== -1" :text="$t('label.delete')" variant="danger" size="md" @confirmed="onDelete" />
          <button class="btn btn-outline-secondary ms-auto" @click="onModalHide">{{ $t('label.cancel') }}</button>
          <button class="btn btn-primary" :disabled="isSaveDisabled" @click="onSave">{{ $t('label.save') }}</button>
        </div>
      </div>
    </div>
  </div>
  <div v-if="showModal" class="modal-backdrop fade show" @click="onModalHide" />
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  visible: { type: Boolean, default: false },
  module: { type: Object, required: true },
  summaryIndex: { type: Number, default: -1 },
  summary: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['close', 'save', 'delete'])

const showModal = ref(props.visible)

watch(() => props.visible, (val) => { showModal.value = val }, { immediate: true })

const recordListModuleFields = computed(() => {
  if (!props.module) return []
  return [
    ...props.module.fields,
    ...props.module.systemFields().map(sf => ({ label: $t(`system.${sf.name}`), name: sf.name === 'recordID' ? 'ID' : sf.name })),
  ].map(({ name, label }) => ({ name, label }))
})

const summaryMetrics = computed(() => [
  { value: 'sum', label: $t('recordList.summaries.metrics.sum.label') },
  { value: 'min', label: $t('recordList.summaries.metrics.min.label') },
  { value: 'max', label: $t('recordList.summaries.metrics.max.label') },
  { value: 'avg', label: $t('recordList.summaries.metrics.avg.label') },
  { value: 'emptyCount', label: $t('recordList.summaries.metrics.emptyCount.label') },
  { value: 'notEmptyCount', label: $t('recordList.summaries.metrics.notEmptyCount.label') },
  { value: 'uniqueCount', label: $t('recordList.summaries.metrics.uniqueCount.label') },
])

const isSaveDisabled = computed(() => !props.summary?.label || !props.summary?.field || !props.summary?.metric)

function onModalHide() {
  showModal.value = false
  emit('close')
}

function onSave() {
  emit('save', props.summary)
}

function onDelete() {
  emit('delete')
}
</script>
