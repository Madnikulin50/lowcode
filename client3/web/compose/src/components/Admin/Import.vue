<template>
  <div class="d-flex">
    <button
      class="btn btn-outline-secondary btn-lg flex-fill"
      @click="showModal = true"
    >
      {{ $t('label.import') }}
    </button>

    <div
      id="import-modal"
      ref="modal"
      class="modal fade"
      tabindex="-1"
    >
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('label.import') }}</h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              @click="showModal = false"
            />
          </div>
          <div class="modal-body">
            <div class="input-group">
              <template v-if="!importObj">
                <input
                  type="file"
                  class="form-control"
                  :placeholder="$t('label.importPlaceholder')"
                  @change="loadFile"
                />

                <h6
                  v-if="processing"
                  class="my-auto ms-3"
                >
                  {{ $t('label.processing') }}
                </h6>
              </template>

              <template v-else>
                <div class="container-fluid p-0">
                  <div class="row g-0 mb-3">
                    <button
                      class="btn btn-outline-secondary"
                      @click="selectAll(true)"
                    >
                      {{ $t('field.selectAll') }}
                    </button>
                    <button
                      class="btn btn-outline-secondary ms-2"
                      @click="selectAll(false)"
                    >
                      {{ $t('field.unselectAll') }}
                    </button>
                  </div>
                  <div class="row g-0">
                    <div
                      v-for="(o, index) in importObj.list"
                      :key="index"
                      class="col-12 col-md-6 col-lg-4"
                    >
                      <div class="form-check">
                        <input
                          v-model="o.import"
                          class="form-check-input"
                          type="checkbox"
                        >
                        <label class="form-check-label">{{ o.name || o.title }}</label>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </div>
          <div class="modal-footer">
            <button
              :disabled="!importObj || !importObj.list.filter(i => i.import).length > 0"
              class="btn btn-primary"
              @click="jsonImport(importObj)"
            >
              {{ $t('label.import') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, computed } from 'vue'
import { useStore } from '../../store'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'

const { t } = useI18n()

defineOptions({
  i18nOptions: {
    namespaces: 'general',
  },
})

const props = defineProps({
  namespace: {
    type: Object,
    required: true,
  },
  type: {
    type: String,
    required: true,
  },
})

const emit = defineEmits(['importSuccessful'])

const store = useStore()

const showModal = ref(false)
const importObj = ref(null)
const processing = ref(false)
const classes = {
  module: compose.Module,
  chart: compose.Chart,
}

const modules = computed(() => store.getters['module/set'])

async function jsonImport ({ list, type }) {
  processing.value = true
  const { namespaceID } = props.namespace
  const ItemClass = classes[type]
  try {
    for (let item of list.filter(i => i.import)) {
      if (importObj.value) {
        item = new ItemClass(item).import(getModuleID)
        item.namespaceID = namespaceID
        await store.dispatch(`${props.type}/create`, item)
      } else {
        break
      }
    }
    emit('importSuccessful')
  } catch (e) {
    toastErrorHandler(t('notification.general.import.failed'))(e)
  }
  cancelImport()
}

function getModuleID (moduleName) {
  const matchedModules = modules.value.filter(m => m.name === moduleName)
  if (matchedModules.length > 0) {
    return matchedModules[0].moduleID
  }
  return null
}

function selectAll (selectAll) {
  importObj.value.list = importObj.value.list.map(i => {
    i.import = selectAll && true
    return i
  })
}

function cancelImport () {
  importObj.value = null
  processing.value = false
  showModal.value = false
}

function loadFile (e = {}) {
  const { files = [] } = (e.type === 'drop' ? e.dataTransfer : e.target) || {}

  if (files[0]) {
    processing.value = true
    const reader = new FileReader()
    reader.readAsText(files[0])
    reader.onload = (evt) => {
      try {
        importObj.value = JSON.parse(evt.target.result)
        if (!importObj.value.list) {
          throw new Error(t('notification.general.import.readingError'))
        } else {
          importObj.value.list = importObj.value.list.map(i => {
            return { import: true, ...i }
          })
        }
      } catch (err) {
        toastErrorHandler(t('notification.general.import.failed'))(err)
        importObj.value = null
      } finally {
        processing.value = false
      }
    }
    reader.onerror = () => {
      toastErrorHandler(t('notification.general.import.readingError'))
      processing.value = false
    }
  }
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>
