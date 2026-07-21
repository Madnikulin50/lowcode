<template>
  <div
    id="discovery-modal"
    ref="modal"
    class="modal fade"
    tabindex="-1"
  >
    <div class="modal-dialog modal-lg modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-header p-3 pb-0 border-bottom-0">
          <h5 class="modal-title">{{ discoveryModalTitle }}</h5>
          <button
            type="button"
            class="btn-close"
            data-bs-dismiss="modal"
          />
        </div>
        <div class="modal-body p-0 border-top-0">
          <field-picker
            :module="module"
            v-model:fields="currentFields"
            disable-system-fields
            style="height: 90vh;"
          />
        </div>
        <div class="modal-footer">
          <button
            class="btn btn-primary"
            @click="onSave()"
          >
            {{ $t('label.saveAndClose') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useStore } from '../../../store'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import FieldPicker from 'corteza-webapp-compose/src/components/Common/FieldPicker'
import { Modal } from 'bootstrap'

const { t } = useI18n()

defineOptions({
  i18nOptions: {
    namespaces: 'module',
  },
})

const props = defineProps({
  modal: {
    type: Boolean,
    required: false,
  },
  module: {
    type: compose.Module,
    required: true,
  },
})

const emit = defineEmits(['update:modal', 'save'])

const store = useStore()

const publicScope = ref({})
const privateScope = ref({})
const protectedScope = ref({})

const currentTabIndex = ref(0)
const currentLang = ref(undefined)

const languages = computed(() => store.getters['languages/set'])

const discoveryModalTitle = computed(() => {
  const { handle } = props.module
  return handle ? `${t('edit.discoverySettings.title')} (${handle})` : t('edit.discoverySettings.title')
})

const scopeOptions = computed(() => {
  return [
    { value: 'private', title: t('edit.discoverySettings.private') },
  ]
})

const currentScope = computed(() => {
  if (currentTabIndex.value < 0) return undefined
  const { value } = scopeOptions.value[currentTabIndex.value] || {}
  if (value === 'public') return publicScope
  if (value === 'private') return privateScope
  if (value === 'protected') return protectedScope
  return undefined
})

const moduleFieldsSet = computed(() => {
  return new Set(props.module.fields.map(({ name }) => name))
})

const currentLanguageIndex = computed(() => {
  if (!currentScope.value) return 0
  const { result = [] } = currentScope.value || {}
  if (!result.length) return 0
  const index = result.findIndex(({ lang }) => lang === currentLang.value)
  return index >= 0 ? index : 0
})

const currentFields = computed({
  get () {
    const { result = [] } = currentScope.value || {}
    if (result[currentLanguageIndex.value]) {
      return result[currentLanguageIndex.value].fields
    }
    return []
  },
  set (fields) {
    if (!currentScope.value) return
    const { result = [] } = currentScope.value || {}
    if (result[currentLanguageIndex.value]) {
      result[currentLanguageIndex.value].fields = [...fields]
    }
  },
})

const showModal = computed({
  get () {
    return props.modal
  },
  set (val) {
    emit('update.modal', val)
  },
})

onMounted(() => {
  currentLang.value = ''

  scopeOptions.value.forEach(({ value }) => {
    const scope = value === 'public' ? publicScope : value === 'private' ? privateScope : protectedScope
    scope.value = { result: [] }

    languages.value.forEach(({ tag: lang }) => {
      let existingFields = new Set()
      if (props.module.config.discovery && props.module.config.discovery[value]) {
        const indexOfLanguage = props.module.config.discovery[value].result.findIndex(r => r.lang === lang)
        if (indexOfLanguage >= 0) {
          existingFields = new Set(props.module.config.discovery[value].result[indexOfLanguage].fields.filter(name => moduleFieldsSet.value.has(name)))
        }
      }
      const fields = [...existingFields].map(name => props.module.fields.find(field => field.name === name))
      scope.value.result.push({ lang, fields })
      existingFields.clear()
    })
  })
})

onBeforeUnmount(() => {
  setDefaultValues()
})

function onSave () {
  const discovery = { public: {}, private: {}, protected: {} }

  scopeOptions.value.forEach(({ value }) => {
    const scope = value === 'public' ? publicScope : value === 'private' ? privateScope : protectedScope
    discovery[value].result = (scope.value?.result || []).map(({ lang, fields }) => {
      return { lang, fields: fields.map(({ name }) => name) }
    })
  })

  emit('save', {
    ...props.module.meta,
    discovery,
  })
}

function setDefaultValues () {
  publicScope.value = {}
  privateScope.value = {}
  protectedScope.value = {}
  currentTabIndex.value = 0
  currentLang.value = undefined
}
</script>

<style scoped>
.mh-tab {
  max-height: calc(100vh - 16rem);
}
</style>
