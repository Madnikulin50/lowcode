<template>
  <div class="card shadow-sm">
    <div class="card-body p-0">
      <div class="align-items-center gap-1 p-3">
        <button class="btn btn-primary btn-lg" @click="openEditor()">
          {{ $t('editor.code-snippets.add') }}
        </button>
      </div>

      <div class="table-responsive" style="min-height: 5rem; max-height: 82vh;">
        <table class="table table-hover mb-0">
          <thead class="table-outline-secondary">
            <tr>
              <th>{{ $t('editor.code-snippets.table-headers.name') }}</th>
              <th class="text-center">{{ $t('editor.code-snippets.table-headers.enabled') }}</th>
              <th style="min-width: 7rem" class="text-end"></th>
            </tr>
          </thead>
          <tbody>
            <template v-if="codeSnippets.length">
              <tr v-for="(s, i) in codeSnippets" :key="i" class="pointer" @click="openEditor(i)">
                <td>{{ s.name }}</td>
                <td class="text-center">
                  <font-awesome-icon v-if="s.enabled" :icon="['fas', 'check']" class="text-primary" />
                  <font-awesome-icon v-else :icon="['fas', 'times']" class="text-extra-light" />
                </td>
                <td class="text-end">
                  <button class="btn btn-link" @click.stop="openEditor(i)">
                    <font-awesome-icon :icon="['fas', 'wrench']" />
                  </button>
                  <c-input-confirm :disabled="codeSnippet.processing" @confirmed="deleteCodeSnippet(i)" />
                </td>
              </tr>
            </template>
            <tr v-else>
              <td colspan="3" class="text-center text-dark" style="padding-top: 1vh;">
                {{ $t('editor.code-snippets.empty') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="modal" ref="modalRef" tabindex="-1" :class="{ show: modal.open, 'd-block': modal.open }" :style="modal.open ? { backgroundColor: 'rgba(0,0,0,0.5)' } : {}">
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title text-capitalize">{{ modal.title }}</h5>
            <button type="button" class="btn-close" @click="closeModal"></button>
          </div>
          <div class="modal-body">
            <div class="form-check mb-3">
              <input id="snippet-enabled" v-model="modal.data.enabled" type="checkbox" class="form-check-input">
              <label class="form-check-label" for="snippet-enabled">{{ $t('editor.code-snippets.enabled') }}</label>
            </div>

            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('editor.code-snippets.form.name.label') }}</label>
              <input v-model="modal.data.name" type="text" class="form-control" required>
            </div>

            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('editor.code-snippets.form.code.label') }}</label>
              <small class="text-muted d-block mb-2">{{ $t('editor.code-snippets.form.code.description') }}</small>
              <c-ace-editor
                v-model="modal.data.script"
                lang="javascript"
                :min-height="500"
                :show-line-numbers="true"
                :border="false"
                :show-popout="false"
                :resizable="true"
              />
            </div>
          </div>
          <div class="modal-footer">
            <c-input-confirm
              v-if="modal.index >= 0"
              :text="$t('label.delete')"
              size="md"
              variant="danger"
              @confirmed="deleteCodeSnippet(modal.index)"
            />
            <button type="button" class="btn btn-outline-secondary ms-auto" @click="closeModal">
              {{ $t('label.cancel') }}
            </button>
            <button type="button" class="btn btn-primary" :disabled="saveDisabled" @click="saveSettings()">
              {{ $t('label.saveAndClose') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useToast } from 'corteza-lib/vue/dist'
import { components } from 'corteza-lib/vue/dist'
import { useUiStore } from '../../store/ui'

const { CAceEditor, CInputConfirm } = components

const $SystemAPI = window.__SystemAPI
const $Settings = inject('$Settings')

const { toastSuccess, toastErrorHandler } = useToast()
const uiStore = useUiStore()
const { incLoader, decLoader } = uiStore

const codeSnippets = ref([])

const modal = ref({
  open: false,
  index: null,
  title: null,
  data: {},
})

const codeSnippet = ref({
  processing: false,
  success: false,
})

const saveDisabled = computed(() => !modal.value.data.name || !modal.value.data.script)

function openEditor(index) {
  const item = index >= 0
    ? codeSnippets.value[index]
    : {
        name: '',
        script: '<script> </script>',
        enabled: true,
      }

  modal.value.index = index
  modal.value.title = item.name || 'Add Code Snippet'
  modal.value.data = { ...item }
  modal.value.open = true
}

function closeModal() {
  modal.value.open = false
}

function fetchSettings() {
  incLoader()
  $Settings.fetch()

  return $SystemAPI.settingsList({ prefix: 'code-snippets' })
    .then(settings => {
      if (settings && settings[0]) {
        codeSnippets.value = settings[0].value
      } else {
        codeSnippets.value = []
      }
    })
    .catch(toastErrorHandler({ title: 'notification.settings.code-snippet.fetch.error' }))
    .finally(() => {
      decLoader()
    })
}

function settingsUpdate(action) {
  codeSnippet.value.processing = true

  $SystemAPI.settingsUpdate({ values: [{ name: 'code-snippets', value: codeSnippets.value }] })
    .then(() => {
      $Settings.fetch()
      animateSuccess()
      if (action === 'delete') {
        toastSuccess('notification.settings.code-snippet.delete.success')
      } else {
        toastSuccess('notification.settings.code-snippet.update.success')
      }
    })
    .catch(toastErrorHandler({ title: 'notification.settings.code-snippet.update.error' }))
    .finally(() => {
      codeSnippet.value.processing = false
    })
}

function animateSuccess() {
  codeSnippet.value.success = true
  setTimeout(() => {
    codeSnippet.value.success = false
  }, 2000)
}

function saveSettings() {
  if (modal.value.index >= 0) {
    codeSnippets.value.splice(modal.value.index, 1, modal.value.data)
  } else {
    codeSnippets.value.push(modal.value.data)
  }

  settingsUpdate('update')
  closeModal()
}

function deleteCodeSnippet(i) {
  codeSnippets.value.splice(i, 1)
  settingsUpdate('delete')
  closeModal()
}

onMounted(fetchSettings)
</script>
