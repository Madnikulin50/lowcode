<template>
  <div
    v-if="show"
    class="modal d-block"
    tabindex="-1"
    role="dialog"
    @click.self="close"
  >
    <div
      class="modal-dialog modal-dialog-scrollable"
      style="width: 65vw; min-width: 50vw; max-width: 80vw;"
      role="document"
    >
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            Рассылка страницы по почте
          </h5>
          <button
            type="button"
            class="btn-close"
            @click="close"
          />
        </div>

        <div class="modal-body">
          <div
            v-if="loading"
            class="text-center py-4"
          >
            <span class="spinner-border" />
          </div>

          <form
            v-else
            @submit.prevent="save"
          >
            <div class="form-check form-switch mb-3">
              <input
                id="export-schedule-enabled"
                v-model="form.enabled"
                class="form-check-input"
                type="checkbox"
              >
              <label
                class="form-check-label"
                for="export-schedule-enabled"
              >
                Отправлять по расписанию
              </label>
            </div>

            <div class="mb-3">
              <label class="form-label">Получатели</label>
              <textarea
                v-model="recipientsText"
                class="form-control"
                rows="2"
                placeholder="a@example.com, b@example.com"
              />
              <small class="text-muted">Адреса через запятую, точку с запятой или с новой строки</small>
            </div>

            <div class="mb-3">
              <label class="form-label">Тема</label>
              <input
                v-model="form.subject"
                class="form-control"
                type="text"
              >
            </div>

            <div class="mb-3">
              <label class="form-label">Текст письма</label>
              <textarea
                v-model="form.body"
                class="form-control"
                rows="4"
              />
            </div>

            <div class="mb-3">
              <label class="form-label">Формат вложения</label>
              <select
                v-model="form.format"
                class="form-select"
              >
                <option value="xlsx">
                  XLSX
                </option>
                <option value="pdf">
                  PDF
                </option>
                <option value="pptx">
                  PPTX
                </option>
              </select>
            </div>

            <div class="row g-2">
              <div class="col">
                <label class="form-label">Периодичность</label>
                <select
                  v-model="form.frequency"
                  class="form-select"
                >
                  <option value="daily">
                    Каждый день
                  </option>
                  <option value="weekly">
                    Раз в неделю
                  </option>
                  <option value="monthly">
                    Раз в месяц
                  </option>
                </select>
              </div>

              <div
                v-if="form.frequency === 'weekly'"
                class="col"
              >
                <label class="form-label">День недели</label>
                <select
                  v-model.number="form.weekday"
                  class="form-select"
                >
                  <option
                    v-for="(name, i) in weekdays"
                    :key="i"
                    :value="i + 1"
                  >
                    {{ name }}
                  </option>
                </select>
              </div>

              <div
                v-if="form.frequency === 'monthly'"
                class="col"
              >
                <label class="form-label">Число месяца</label>
                <input
                  v-model.number="form.monthDay"
                  class="form-control"
                  type="number"
                  min="1"
                  max="31"
                >
              </div>

              <div class="col">
                <label class="form-label">Время</label>
                <input
                  v-model="form.time"
                  class="form-control"
                  type="time"
                  required
                >
              </div>
            </div>
            <small class="text-muted d-block mt-1">Время указывается по часовому поясу сервера</small>

            <div
              v-if="form.lastRunAt"
              class="mt-3 small"
              :class="form.lastError ? 'text-danger' : 'text-muted'"
            >
              Последняя отправка: {{ new Date(form.lastRunAt).toLocaleString() }}
              <template v-if="form.lastError">
                , ошибка: {{ form.lastError }}
              </template>
            </div>

            <div
              v-if="error"
              class="alert alert-danger mt-3 mb-0"
            >
              {{ error }}
            </div>
          </form>
        </div>

        <div class="modal-footer">
          <button
            type="button"
            class="btn btn-outline-secondary me-auto"
            :disabled="loading || busy"
            title="Отправляет сохранённые настройки прямо сейчас"
            @click="sendNow"
          >
            Отправить сейчас
          </button>
          <button
            type="button"
            class="btn btn-light"
            @click="close"
          >
            Закрыть
          </button>
          <button
            type="button"
            class="btn btn-primary"
            :disabled="loading || busy"
            @click="save"
          >
            Сохранить
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { composables } from 'corteza-lib/vue/dist'

const { toastSuccess } = composables.useToast()

const props = defineProps({
  namespaceId: { type: String, required: true },
  pageId: { type: String, required: true },
})

const show = defineModel('show', { type: Boolean, default: false })

const weekdays = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота', 'Воскресенье']

const loading = ref(false)
const busy = ref(false)
const error = ref('')
const form = ref({})
const recipientsText = ref('')

const endpoint = computed(() => `${window.CortezaAPI}/compose/namespace/${props.namespaceId}/page/${props.pageId}/export-schedule`)

async function call (url, options = {}) {
  const rsp = await fetch(url, {
    ...options,
    headers: {
      Authorization: `Bearer ${window.__auth.accessTokenFn()}`,
      'Content-Type': 'application/json',
    },
  })
  if (!rsp.ok) {
    throw new Error((await rsp.text()).trim())
  }
  return (await rsp.json()).response
}

function apply (schedule) {
  form.value = schedule
  recipientsText.value = (schedule.recipients || []).join(', ')
}

async function load () {
  loading.value = true
  error.value = ''
  try {
    apply(await call(endpoint.value))
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function save () {
  busy.value = true
  error.value = ''
  try {
    const recipients = recipientsText.value.split(/[\s,;]+/).filter(Boolean)
    // the page address is stored to be rendered by the server on schedule
    const payload = { ...form.value, recipients, url: window.location.href }
    apply(await call(endpoint.value, { method: 'PUT', body: JSON.stringify(payload) }))
    toastSuccess('Настройки рассылки сохранены')
    return true
  } catch (e) {
    error.value = e.message
    return false
  } finally {
    busy.value = false
  }
}

async function sendNow () {
  // the server sends what is saved, so the form is saved first
  if (!(await save())) return

  busy.value = true
  try {
    apply(await call(`${endpoint.value}/send`, { method: 'POST' }))
    toastSuccess('Письмо отправлено')
  } catch (e) {
    error.value = e.message
    await load()
  } finally {
    busy.value = false
  }
}

function close () {
  show.value = false
}

watch(show, open => { if (open) load() })
</script>
