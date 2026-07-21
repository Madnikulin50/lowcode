<template>
  <div v-if="options">
    <div class="mb-3">
      <h5 class="text-primary mb-2">{{ t('table.configurator.general') }}</h5>
      <div class="row g-0">
        <div class="col pe-3">
          <div class="mb-3">
            <label class="text-primary form-label">{{ t('table.configurator.table.variant') }}</label>
            <select v-model="options.tableVariant" class="form-select">
              <option v-for="tv in tableVariants" :key="tv.value" :value="tv.value">{{ tv.text }}</option>
            </select>
          </div>
        </div>
        <div class="col">
          <div class="mb-3">
            <label class="text-primary form-label">{{ t('table.configurator.head-variant') }}</label>
            <div class="mt-lg-2">
              <div class="form-check form-check-inline">
                <input v-model="options.headVariant" :value="null" class="form-check-input" type="radio" name="headVariant" id="hvNone" />
                <label class="form-check-label" for="hvNone">{{ t('table.configurator.none') }}</label>
              </div>
              <div class="form-check form-check-inline">
                <input v-model="options.headVariant" value="light" class="form-check-input" type="radio" name="headVariant" id="hvLight" />
                <label class="form-check-label" for="hvLight">{{ t('table.configurator.light') }}</label>
              </div>
              <div class="form-check form-check-inline">
                <input v-model="options.headVariant" value="dark" class="form-check-input" type="radio" name="headVariant" id="hvDark" />
                <label class="form-check-label" for="hvDark">{{ t('table.configurator.dark') }}</label>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="mb-3">
        <label class="text-primary form-label">{{ t('table.configurator.table.options.label') }}</label>
        <div>
          <div class="form-check form-check-inline"><input v-model="options.striped" class="form-check-input" type="checkbox" id="striped" /><label class="form-check-label" for="striped">{{ t('table.configurator.table.options.striped') }}</label></div>
          <div class="form-check form-check-inline"><input v-model="options.bordered" class="form-check-input" type="checkbox" id="bordered" /><label class="form-check-label" for="bordered">{{ t('table.configurator.table.options.bordered') }}</label></div>
          <div class="form-check form-check-inline"><input v-model="options.borderless" class="form-check-input" type="checkbox" id="borderless" /><label class="form-check-label" for="borderless">{{ t('table.configurator.table.options.borderless') }}</label></div>
          <div class="form-check form-check-inline"><input v-model="options.small" class="form-check-input" type="checkbox" id="small" /><label class="form-check-label" for="small">{{ t('table.configurator.table.options.small') }}</label></div>
          <div class="form-check form-check-inline"><input v-model="options.hover" class="form-check-input" type="checkbox" id="hover" /><label class="form-check-label" for="hover">{{ t('table.configurator.table.options.hover') }}</label></div>
          <div class="form-check form-check-inline"><input v-model="options.dark" class="form-check-input" type="checkbox" id="dark" /><label class="form-check-label" for="dark">{{ t('table.configurator.table.options.dark') }}</label></div>
          <div class="form-check form-check-inline"><input v-model="options.responsive" class="form-check-input" type="checkbox" id="responsive" /><label class="form-check-label" for="responsive">{{ t('table.configurator.table.options.responsive') }}</label></div>
          <div class="form-check form-check-inline"><input v-model="options.fixed" class="form-check-input" type="checkbox" id="fixed" /><label class="form-check-label" for="fixed">{{ t('table.configurator.table.options.fixed') }}</label></div>
          <div class="form-check form-check-inline"><input v-model="options.noCollapse" class="form-check-input" type="checkbox" id="noCollapse" /><label class="form-check-label" for="noCollapse">{{ t('table.configurator.table.options.no-collapse') }}</label></div>
        </div>
      </div>
    </div>
    <hr />
    <div class="mb-3">
      <h5 class="text-primary mb-2">{{ t('table.configurator.data') }}</h5>
      <div v-if="options.datasources?.length > 1" class="mb-3">
        <label class="text-primary form-label">{{ t('table.configurator.joined-datasource-handling') }}</label>
        <select v-model="currentConfigurableDatasourceName" class="form-select">
          <option v-for="ds in options.datasources" :key="ds.name" :value="ds.name">{{ ds.name }}</option>
        </select>
      </div>
      <div v-if="currentConfigurableDatasourceName && currentColumns.length" class="mb-3">
        <label class="text-primary form-label">{{ t('table.configurator.columns') }}</label>
        <ColumnPicker
          :all-items="currentColumns"
          :selected-items="currentSelectedColumns"
          class="d-flex flex-column"
          @update:selected-items="updateSelectedColumns"
        />
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import ColumnPicker from '../../../../Common/ColumnPicker.vue'

const props = defineProps({
  displayElementOptions: { type: Object, default: () => ({}) },
  columns: { type: Array, required: true },
  datasource: { type: Object, default: undefined },
})
const emit = defineEmits(['update:displayElementOptions'])

const { t } = useI18n()
const currentConfigurableDatasourceName = ref(undefined)

const options = computed({
  get: () => props.displayElementOptions,
  set: (val) => emit('update.displayElementOptions', val),
})

const tableVariants = [
  { value: '', text: t('table.configurator.none') },
  { value: 'primary', text: t('table.configurator.table.variants.primary') },
  { value: 'secondary', text: t('table.configurator.table.variants.secondary') },
  { value: 'info', text: t('table.configurator.table.variants.info') },
  { value: 'danger', text: t('table.configurator.table.variants.danger') },
  { value: 'warning', text: t('table.configurator.table.variants.warning') },
  { value: 'success', text: t('table.configurator.table.variants.success') },
  { value: 'light', text: t('table.configurator.table.variants.light') },
  { value: 'dark', text: t('table.configurator.table.variants.dark') },
]

const currentColumns = computed(() => {
  if (currentConfigurableDatasourceName.value && props.columns) {
    const dsIndex = options.value?.datasources?.findIndex(ds => ds.name === currentConfigurableDatasourceName.value)
    if (dsIndex >= 0) return props.columns[dsIndex] || []
  }
  return []
})

const currentSelectedColumns = computed({
  get: () => currentConfigurableDatasourceName.value ? options.value?.columns?.[currentConfigurableDatasourceName.value] : [],
  set: (columns) => {
    if (currentConfigurableDatasourceName.value) {
      options.value.columns[currentConfigurableDatasourceName.value] = columns || []
    }
  },
})

watch(() => options.value?.datasources, (datasources) => {
  if (!datasources) return
  datasources.forEach(({ name }) => {
    if (!options.value.columns) options.value.columns = {}
    if (!options.value.columns[name]) options.value.columns[name] = []
  })
  currentConfigurableDatasourceName.value = (datasources[0] || {}).name
}, { immediate: true })

function updateSelectedColumns(columns) {
  currentSelectedColumns.value = columns.map(c => currentColumns.value.find(({ name }) => name === c))
}
</script>