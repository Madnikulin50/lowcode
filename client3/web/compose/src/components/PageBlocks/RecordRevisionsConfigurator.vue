<template>
  <div>
    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('sortDirection.label') }}</label>
          <small class="text-muted d-block mb-1">{{ $t('sortDirection.footnote') }}</small>
          <c-input-select v-model="options.sortDirection" :options="sortDirections" label="label" :clearable="false" :reduce="o => o.value" />
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('preload') }}</label>
          <div class="form-check form-switch">
            <input v-model="options.preload" class="form-check-input" type="checkbox" role="switch" />
          </div>
        </div>
      </div>
    </div>
    <div v-if="module" class="mb-3">
      <label class="form-label text-primary">{{ $t('fields.label') }}</label>
      <FieldPicker :module="module" v-model:fields="displayedFieldsArray" style="height: 50vh;" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import FieldPicker from 'corteza-webapp-compose/src/components/Common/FieldPicker'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  module: { type: Object, required: false },
})

const options = computed(() => props.block.options)

const sortDirections = ref([
  { label: $t('sortDirection.desc'), value: 'desc' },
  { label: $t('sortDirection.asc'), value: 'asc' },
])

const displayedFieldsArray = computed({
  get: () => props.module?.filterFields?.(options.value.displayedFields) || [],
  set: (fields) => { options.value.displayedFields = fields.map(f => f.name) },
})

if (!options.value.sortDirection) { options.value.sortDirection = 'desc' }

onBeforeUnmount(() => { sortDirections.value = [] })
</script>
