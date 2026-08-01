<template>
  <div>
    <div
      v-if="showModal"
      class="modal d-block"
      tabindex="-1"
      role="dialog"
      @click.self="onHide"
    >
      <div class="modal-dialog modal-xl modal-dialog-scrollable" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ title }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" @click="onHide"></button>
          </div>
          <div class="modal-body position-static p-0">
            <CTranslatorForm
              v-if="loaded"
              :primary-resource="resource"
              :translations="translations"
              :key-prettifier="keyPrettyfier"
              :languages="languages"
              :titles="titles"
              :highlight-key="highlightKey"
              @change="onFormChange"
            />
          </div>
          <div class="modal-footer">
            <button
              class="btn btn-primary"
              :tabindex="languages.length + 1"
              :disabled="disabled || changes.length === 0"
              @click="onSubmit()"
            >
              {{ $t('save-changes') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, unref } from 'vue'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'
import { useLanguagesStore } from 'corteza-webapp-compose/src/store/languages'
import CTranslatorForm from './CTranslatorForm.vue'

const { t: $t } = useI18n({ useScope: 'global' })
const { toastSuccess, toastErrorHandler } = composables.useToast()
const languagesStore = useLanguagesStore()

const resource = ref(undefined)
const updater = ref(undefined)
const loaded = ref(false)
const translations = ref([])
const changes = ref([])
const titles = ref({})
const highlightKey = ref(undefined)
const keyPrettyfier = ref(undefined)

const languages = computed(() => languagesStore.set || [])

const title = computed(() => titles.value[resource.value] || '')

const showModal = computed({
  get: () => loaded.value,
  set: (open) => { if (!open) clear() },
})

const disabled = computed(() => false)

function loadModal(payload) {
  if (!payload) {
    clear()
    return
  }

  resource.value = payload.resource
  titles.value = payload.titles
  highlightKey.value = payload.highlightKey
  updater.value = payload.updater
  keyPrettyfier.value = payload.keyPrettyfier
  changes.value = []

  payload.fetcher().then(tt => {
    translations.value = tt
    loaded.value = true
  })
}

function onFormChange(val) {
  changes.value = val
}

function onSubmit() {
  if (changes.value.length === 0) {
    clear()
    return
  }

  updater.value(changes.value)
    .then(() => {
      toastSuccess($t('notification.translations.saved'))
      clear()
    })
    .catch(toastErrorHandler($t('notification.translations.saveFailed')))
}

function onHide() {
  clear()
}

function clear() {
  titles.value = {}
  changes.value = []
  resource.value = undefined
  highlightKey.value = undefined
  updater.value = undefined
  keyPrettyfier.value = undefined
  loaded.value = false
}

onMounted(() => {
  languagesStore.load()
  window.addEventListener('c-translator', (e) => loadModal(e.detail))
})

onBeforeUnmount(() => {
  window.removeEventListener('c-translator', loadModal)
})
</script>
