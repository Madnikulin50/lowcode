<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      {{ isEdit ? 'Редактирование фактора риска' : 'Новый фактор риска' }}
    </Teleport>

    <div
      v-if="loading"
      class="d-flex justify-content-center py-5"
    >
      <span class="spinner-border" />
    </div>

    <div
      v-else
      class="row"
    >
      <div class="col-lg-8">
        <div class="card shadow-sm">
          <div class="card-body row g-3">
            <div class="col-md-4">
              <label class="form-label">Код (handle)</label>
              <input
                v-model="factor.handle"
                class="form-control"
                placeholder="daysSinceAudit"
                :disabled="isEdit"
              >
            </div>
            <div class="col-md-8">
              <label class="form-label">Название</label>
              <input
                v-model="factor.label"
                class="form-control"
                placeholder="Давность аудита"
              >
            </div>
            <div class="col-12">
              <label class="form-label">Описание</label>
              <textarea
                v-model="factor.description"
                class="form-control"
                rows="2"
              />
            </div>
            <div class="col-md-4">
              <label class="form-label">Шкала</label>
              <select
                v-model="factor.scale"
                class="form-select"
              >
                <option value="numeric">
                  Числовая
                </option>
                <option value="bool">
                  Да/нет
                </option>
                <option value="states">
                  Состояния (для байесовской сети)
                </option>
              </select>
            </div>

            <template v-if="factor.scale === 'numeric'">
              <div class="col-md-3">
                <label class="form-label">Максимум (нормализация)</label>
                <input
                  v-model.number="factor.max"
                  type="number"
                  class="form-control"
                >
              </div>
              <div class="col-md-3 d-flex align-items-end">
                <div class="form-check">
                  <input
                    id="factorInvert"
                    v-model="factor.invert"
                    class="form-check-input"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="factorInvert"
                  >
                    Чем больше — тем лучше (инвертировать)
                  </label>
                </div>
              </div>
              <div class="col-md-2">
                <label class="form-label">Baseline (для атрибуции)</label>
                <input
                  v-model="factor.baseline"
                  class="form-control"
                  placeholder="0"
                >
              </div>
            </template>

            <template v-if="factor.scale === 'states'">
              <div class="col-12">
                <label class="form-label d-flex justify-content-between">
                  <span>Состояния</span>
                  <button
                    class="btn btn-sm btn-outline-primary"
                    @click="addState"
                  >
                    + состояние
                  </button>
                </label>
                <table class="table table-sm">
                  <thead>
                    <tr>
                      <th>Код</th>
                      <th>Название</th>
                      <th>Severity</th>
                      <th />
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(s, i) in factor.states"
                      :key="i"
                    >
                      <td>
                        <input
                          v-model="s.code"
                          class="form-control form-control-sm"
                        >
                      </td>
                      <td>
                        <input
                          v-model="s.label"
                          class="form-control form-control-sm"
                        >
                      </td>
                      <td>
                        <select
                          v-model="s.severity"
                          class="form-select form-select-sm"
                        >
                          <option value="low">
                            low
                          </option>
                          <option value="medium">
                            medium
                          </option>
                          <option value="high">
                            high
                          </option>
                          <option value="critical">
                            critical
                          </option>
                        </select>
                      </td>
                      <td>
                        <button
                          class="btn btn-sm btn-outline-danger"
                          @click="factor.states.splice(i, 1)"
                        >
                          &times;
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
                <p class="text-secondary small mb-0">
                  Состояния произвольные для каждого фактора — severity нужна только чтобы
                  интерфейс мог единообразно раскрашивать/сортировать узлы разных факторов.
                </p>
              </div>
            </template>

            <div class="col-md-4">
              <label class="form-label">Источник значения</label>
              <select
                v-model="factor.source"
                class="form-select"
              >
                <option value="field">
                  Поле записи
                </option>
                <option value="expr">
                  Выражение
                </option>
                <option value="manual">
                  Ручной ввод
                </option>
                <option value="external">
                  Внешняя система / агрегат
                </option>
              </select>
            </div>
            <div
              v-if="factor.source === 'field'"
              class="col-md-4"
            >
              <label class="form-label">Имя поля</label>
              <input
                v-model="factor.fieldName"
                class="form-control"
              >
            </div>
            <div
              v-if="factor.source === 'expr'"
              class="col-md-4"
            >
              <label class="form-label">Выражение</label>
              <input
                v-model="factor.expr"
                class="form-control"
              >
            </div>
          </div>
          <div class="card-footer d-flex justify-content-end gap-2">
            <router-link
              :to="{ name: 'admin.risk.factors' }"
              class="btn btn-outline-secondary"
            >
              Отмена
            </router-link>
            <button
              class="btn btn-primary"
              @click="save"
            >
              Сохранить
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { composables } from 'corteza-lib/vue/dist'

const { useToast } = composables
const { toastSuccess, toastErrorHandler } = useToast()
const route = useRoute()
const router = useRouter()
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const isEdit = computed(() => !!route.params.factorID)
const loading = ref(false)

const factor = reactive({
  handle: '',
  label: '',
  description: '',
  scale: 'numeric',
  max: 10,
  invert: false,
  baseline: '',
  source: 'field',
  fieldName: '',
  expr: '',
  states: [],
})

function addState () {
  const n = (factor.states || []).length
  factor.states = [...(factor.states || []), { code: '', label: '', severity: 'low', order: n }]
}

function load () {
  if (!isEdit.value) return Promise.resolve()
  loading.value = true
  return $ComposeAPI.riskFactorRead({ factorID: route.params.factorID })
    .then(({ factor: f }) => {
      Object.assign(factor, f, { states: (f.states || []).map((s) => ({ ...s })) })
    })
    .catch(toastErrorHandler('Не удалось загрузить фактор'))
    .finally(() => { loading.value = false })
}

function save () {
  const payload = { ...factor, namespaceID: props.namespace?.namespaceID }
  const call = isEdit.value
    ? $ComposeAPI.riskFactorUpdate({ ...payload, factorID: route.params.factorID })
    : $ComposeAPI.riskFactorCreate(payload)
  call
    .then(() => {
      toastSuccess(isEdit.value ? 'Фактор обновлён' : 'Фактор создан')
      router.push({ name: 'admin.risk.factors' })
    })
    .catch(toastErrorHandler('Не удалось сохранить фактор'))
}

onMounted(load)
</script>
