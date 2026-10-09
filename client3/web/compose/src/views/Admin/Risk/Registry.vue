<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      Реестр рисков
    </Teleport>

    <div class="row g-3">
      <div class="col-lg-4">
        <div class="card shadow-sm">
          <div class="card-header d-flex justify-content-between align-items-center">
            <strong>Привязки модели к модулю</strong>
            <button
              class="btn btn-sm btn-outline-primary"
              @click="newBindingOpen = true"
            >
              + привязка
            </button>
          </div>
          <div class="list-group list-group-flush">
            <button
              v-for="b in bindings"
              :key="b.bindingID"
              class="list-group-item list-group-item-action"
              :class="{ active: selected?.bindingID === b.bindingID }"
              @click="select(b)"
            >
              <div class="d-flex justify-content-between">
                <span>{{ moduleByID(b.moduleID)?.name || `Модуль #${b.moduleID}` }}</span>
                <span
                  class="badge"
                  :class="b.autoRecalc ? 'bg-success' : 'bg-secondary'"
                >
                  {{ b.autoRecalc ? 'авто' : 'вручную' }}
                </span>
              </div>
              <small class="text-secondary">модель #{{ b.modelID }}</small>
            </button>
            <div
              v-if="!bindings.length"
              class="list-group-item text-secondary"
            >
              Привязок пока нет.
            </div>
          </div>

          <div
            v-if="newBindingOpen"
            class="card-body border-top"
          >
            <div class="mb-2">
              <label class="form-label">Модель</label>
              <select
                v-model="newBinding.modelID"
                class="form-select form-select-sm"
              >
                <option
                  v-for="m in models"
                  :key="m.modelID"
                  :value="m.modelID"
                >
                  {{ m.name }}
                </option>
              </select>
            </div>
            <div class="mb-2">
              <label class="form-label">Модуль</label>
              <c-input-select
                v-model="newBinding.moduleID"
                label="name"
                placeholder="Выберите модуль"
                :options="modules"
                :get-option-key="m => m.moduleID"
                :reduce="m => m.moduleID"
              />
            </div>
            <div class="form-check mb-2">
              <input
                id="autoRecalc"
                v-model="newBinding.autoRecalc"
                class="form-check-input"
                type="checkbox"
              >
              <label
                class="form-check-label"
                for="autoRecalc"
              >Автопересчёт при изменении записи</label>
            </div>
            <div class="d-flex justify-content-end gap-2">
              <button
                class="btn btn-sm btn-outline-secondary"
                @click="newBindingOpen = false"
              >
                Отмена
              </button>
              <button
                class="btn btn-sm btn-primary"
                @click="createBinding"
              >
                Создать
              </button>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="selected"
        class="col-lg-8 d-flex flex-column gap-3"
      >
        <!-- binding settings -->
        <div class="card shadow-sm">
          <div class="card-header d-flex justify-content-between align-items-center">
            <strong>Настройки привязки #{{ selected.bindingID }}</strong>
            <c-input-confirm
              :text="'Удалить привязку'"
              show-icon
              borderless
              variant="link"
              size="sm"
              button-class="text-danger p-0"
              @confirmed="removeBinding"
            />
          </div>
          <div class="card-body">
            <div class="row g-2 mb-2">
              <div class="col-md-6">
                <label class="form-label">Модель</label>
                <select
                  v-model="bindingForm.modelID"
                  class="form-select form-select-sm"
                >
                  <option
                    v-for="m in models"
                    :key="m.modelID"
                    :value="m.modelID"
                  >
                    {{ m.name }}
                  </option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">Модуль</label>
                <c-input-select
                  v-model="bindingForm.moduleID"
                  label="name"
                  placeholder="Выберите модуль"
                  :options="modules"
                  :get-option-key="m => m.moduleID"
                  :reduce="m => m.moduleID"
                />
              </div>
              <div class="col-md-2 d-flex align-items-end">
                <div class="form-check">
                  <input
                    id="bindingAutoRecalc"
                    v-model="bindingForm.autoRecalc"
                    class="form-check-input"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label small"
                    for="bindingAutoRecalc"
                  >Авто</label>
                </div>
              </div>
            </div>

            <label class="form-label d-flex justify-content-between align-items-center">
              <span>Соответствие факторов полям записи (fieldMap)</span>
              <button
                class="btn btn-sm btn-outline-primary"
                @click="addFieldMapRow"
              >
                + строка
              </button>
            </label>
            <table
              v-if="fieldMapRows.length"
              class="table table-sm"
            >
              <thead>
                <tr>
                  <th>Код фактора</th>
                  <th>Поле записи</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(row, i) in fieldMapRows"
                  :key="i"
                >
                  <td>
                    <input
                      v-model="row.handle"
                      class="form-control form-control-sm"
                      placeholder="incidents90d"
                    >
                  </td>
                  <td>
                    <input
                      v-model="row.field"
                      class="form-control form-control-sm"
                      placeholder="incidents_90d"
                    >
                  </td>
                  <td>
                    <button
                      class="btn btn-sm btn-outline-danger"
                      @click="fieldMapRows.splice(i, 1)"
                    >
                      &times;
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
            <p
              v-else
              class="text-secondary small"
            >
              Не задано — актуально только для автопересчёта (факторы, читаемые прямо из поля записи-субъекта).
            </p>

            <div class="d-flex justify-content-end">
              <button
                class="btn btn-primary btn-sm"
                @click="saveBinding"
              >
                Сохранить
              </button>
            </div>
          </div>
        </div>

        <!-- manual assess -->
        <div class="card shadow-sm">
          <div class="card-header"><strong>Пересчитать вручную</strong></div>
          <div class="card-body">
            <div class="row g-2">
              <div class="col-md-4">
                <label class="form-label">ID записи-субъекта</label>
                <input
                  v-model="assessForm.subjectRecordID"
                  class="form-control form-control-sm"
                >
              </div>
              <div class="col-md-6">
                <label class="form-label">Evidence (JSON)</label>
                <input
                  v-model="assessForm.evidenceJSON"
                  class="form-control form-control-sm font-monospace"
                  placeholder='{"shrinkPct":5.8}'
                >
              </div>
              <div class="col-md-2 d-flex align-items-end">
                <button
                  class="btn btn-primary btn-sm w-100"
                  @click="runAssess"
                >
                  Оценить
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- portfolio summary -->
        <div class="card shadow-sm">
          <div class="card-header">
            <strong>Портфельная сводка</strong>
            <span class="text-secondary ms-2">{{ summary?.subjects || 0 }} объектов</span>
          </div>
          <div class="card-body">
            <div
              v-for="f in summary?.factors || []"
              :key="f.factorHandle"
              class="mb-2"
            >
              <div class="d-flex justify-content-between small">
                <span>{{ f.factorHandle }}</span>
                <span>{{ round1(f.meanAbsValue) }}</span>
              </div>
              <div class="progress" style="height:6px">
                <div
                  class="progress-bar"
                  :style="{ width: Math.min(100, f.meanAbsValue * 2) + '%' }"
                />
              </div>
            </div>
            <p
              v-if="!(summary?.factors || []).length"
              class="text-secondary small mb-0"
            >
              Пока нет оценок с разложением — запустите расчёт выше.
            </p>
          </div>
        </div>

        <!-- assessments -->
        <div class="card shadow-sm">
          <div class="card-header"><strong>Последние оценки по объектам</strong></div>
          <table class="table table-sm mb-0">
            <thead>
              <tr>
                <th>Объект</th>
                <th>Inherent</th>
                <th>Residual</th>
                <th>Уровень</th>
                <th>Когда</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="a in assessments"
                :key="a.assessmentID"
                role="button"
                class="risk-record-row"
                title="Открыть запись"
                @click="openRecord(a)"
              >
                <td>#{{ a.subjectRecordID }}</td>
                <td>{{ round1(a.inherentScore) }}</td>
                <td>{{ round1(a.residualScore) }}</td>
                <td>
                  <span
                    class="badge"
                    :class="levelBadgeClass(a.level)"
                  >{{ a.level }}</span>
                </td>
                <td>{{ new Date(a.assessedAt).toLocaleString() }}</td>
              </tr>
              <tr v-if="!assessments.length">
                <td
                  colspan="5"
                  class="text-center text-secondary py-3"
                >
                  Оценок пока нет.
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- treatments -->
        <div class="card shadow-sm">
          <div class="card-header d-flex justify-content-between align-items-center">
            <strong>Мероприятия по снижению риска</strong>
            <button
              class="btn btn-sm btn-outline-primary"
              @click="newTreatmentOpen = true"
            >
              + мероприятие
            </button>
          </div>
          <table class="table table-sm mb-0">
            <thead>
              <tr>
                <th>Описание</th>
                <th>Статус</th>
                <th />
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="t in treatments"
                :key="t.treatmentID"
              >
                <td>{{ t.description }}</td>
                <td>
                  <select
                    :value="t.status"
                    class="form-select form-select-sm"
                    style="width:auto"
                    @change="updateTreatmentStatus(t, $event.target.value)"
                  >
                    <option value="open">
                      open
                    </option>
                    <option value="in_progress">
                      in_progress
                    </option>
                    <option value="done">
                      done
                    </option>
                    <option value="cancelled">
                      cancelled
                    </option>
                  </select>
                </td>
                <td class="text-end">
                  <button
                    class="btn btn-sm btn-outline-danger"
                    @click="removeTreatment(t)"
                  >
                    Удалить
                  </button>
                </td>
              </tr>
              <tr v-if="!treatments.length">
                <td
                  colspan="3"
                  class="text-center text-secondary py-3"
                >
                  Мероприятий пока нет.
                </td>
              </tr>
            </tbody>
          </table>
          <div
            v-if="newTreatmentOpen"
            class="card-body border-top"
          >
            <div class="input-group">
              <input
                v-model="newTreatmentText"
                class="form-control"
                placeholder="Провести внеплановый аудит"
              >
              <button
                class="btn btn-primary"
                @click="createTreatment"
              >
                Добавить
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { composables } from 'corteza-lib/vue/dist'
import { useStore } from '../../../store'

const { useToast } = composables
const { toastSuccess, toastErrorHandler } = useToast()
const $ComposeAPI = window.__composeAPI
const store = useStore()
const router = useRouter()

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const modules = computed(() => store.module.set)
const moduleByID = (id) => store.module.getByID(id)

const bindings = ref([])
const models = ref([])
const selected = ref(null)
const assessments = ref([])
const summary = ref(null)
const treatments = ref([])

const newBindingOpen = ref(false)
const newBinding = reactive({ modelID: '', moduleID: '', autoRecalc: true })
const newTreatmentOpen = ref(false)
const newTreatmentText = ref('')
const assessForm = reactive({ subjectRecordID: '', evidenceJSON: '{}' })

const bindingForm = reactive({ modelID: '', moduleID: '', autoRecalc: true })
const fieldMapRows = ref([])

function loadBindings () {
  return $ComposeAPI.riskBindingList({ namespaceID: props.namespace?.namespaceID })
    .then(({ bindings: set }) => { bindings.value = set || [] })
}
function loadModels () {
  return $ComposeAPI.riskModelList({ namespaceID: props.namespace?.namespaceID })
    .then(({ models: set }) => { models.value = set || [] })
}

function select (b) {
  selected.value = b
  bindingForm.modelID = b.modelID
  bindingForm.moduleID = b.moduleID
  bindingForm.autoRecalc = !!b.autoRecalc
  fieldMapRows.value = Object.entries(b.fieldMap || {}).map(([handle, field]) => ({ handle, field }))
  refresh()
}

function fieldMapFromRows () {
  const out = {}
  for (const row of fieldMapRows.value) {
    if (row.handle) out[row.handle] = row.field
  }
  return out
}

function addFieldMapRow () {
  fieldMapRows.value = [...fieldMapRows.value, { handle: '', field: '' }]
}

function saveBinding () {
  $ComposeAPI.riskBindingUpdate({
    bindingID: selected.value.bindingID,
    modelID: bindingForm.modelID,
    moduleID: bindingForm.moduleID,
    autoRecalc: bindingForm.autoRecalc,
    fieldMap: fieldMapFromRows(),
  })
    .then(({ binding }) => {
      toastSuccess('Привязка обновлена')
      return loadBindings().then(() => select(binding))
    })
    .catch(toastErrorHandler('Не удалось сохранить привязку'))
}

function removeBinding () {
  $ComposeAPI.riskBindingDelete({ bindingID: selected.value.bindingID })
    .then(() => {
      toastSuccess('Привязка удалена')
      selected.value = null
      return loadBindings().then(() => {
        if (bindings.value.length) select(bindings.value[0])
      })
    })
    .catch(toastErrorHandler('Не удалось удалить привязку'))
}

function refresh () {
  if (!selected.value) return
  const bindingID = selected.value.bindingID
  $ComposeAPI.riskAssessmentList({ bindingID, latestPerSubject: 1 })
    .then(({ assessments: set }) => { assessments.value = set || [] })
    .catch(toastErrorHandler('Не удалось загрузить оценки'))
  $ComposeAPI.riskPortfolioSummary({ bindingID })
    .then((s) => { summary.value = s })
    .catch(toastErrorHandler('Не удалось загрузить сводку'))
  $ComposeAPI.riskTreatmentList({ bindingID })
    .then(({ treatments: set }) => { treatments.value = set || [] })
    .catch(toastErrorHandler('Не удалось загрузить мероприятия'))
}

function createBinding () {
  $ComposeAPI.riskBindingCreate({
    namespaceID: props.namespace?.namespaceID,
    modelID: newBinding.modelID,
    moduleID: newBinding.moduleID,
    autoRecalc: newBinding.autoRecalc,
    fieldMap: {},
  })
    .then(({ binding }) => {
      toastSuccess('Привязка создана')
      newBindingOpen.value = false
      return loadBindings().then(() => select(binding))
    })
    .catch(toastErrorHandler('Не удалось создать привязку'))
}

function runAssess () {
  let evidence = {}
  try {
    evidence = JSON.parse(assessForm.evidenceJSON || '{}')
  } catch {
    toastErrorHandler('Не удалось разобрать JSON')(new Error('invalid JSON'))
    return
  }
  $ComposeAPI.riskAssess({
    bindingID: selected.value.bindingID,
    subjectRecordID: assessForm.subjectRecordID,
    evidence,
    controlFactorHandle: 'controlEffectiveness',
  })
    .then(() => {
      toastSuccess('Оценка сохранена')
      refresh()
    })
    .catch(toastErrorHandler('Расчёт не удался'))
}

function createTreatment () {
  if (!newTreatmentText.value) return
  $ComposeAPI.riskTreatmentCreate({ bindingID: selected.value.bindingID, description: newTreatmentText.value })
    .then(() => {
      newTreatmentText.value = ''
      newTreatmentOpen.value = false
      return refresh()
    })
    .catch(toastErrorHandler('Не удалось создать мероприятие'))
}
function updateTreatmentStatus (t, status) {
  $ComposeAPI.riskTreatmentUpdate({ treatmentID: t.treatmentID, description: t.description, status })
    .then(() => refresh())
    .catch(toastErrorHandler('Не удалось обновить статус'))
}
function removeTreatment (t) {
  $ComposeAPI.riskTreatmentDelete({ treatmentID: t.treatmentID })
    .then(() => refresh())
    .catch(toastErrorHandler('Не удалось удалить мероприятие'))
}

function round1 (v) {
  return Math.round((v || 0) * 10) / 10
}
function levelBadgeClass (level) {
  return { low: 'bg-success', medium: 'bg-warning text-dark', high: 'bg-orange text-white', critical: 'bg-danger' }[level] || 'bg-secondary'
}

function openRecord (a) {
  const moduleID = selected.value?.moduleID
  if (!moduleID || !a.subjectRecordID) return
  // Same lookup components/ModuleFields/Viewer/Record.vue uses for a
  // "Record" field's link: the object's real page (with its blocks —
  // including a Risk block, if one was added), not the raw admin editor.
  const page = store.page.set.find((p) => p.moduleID === moduleID)
  if (!page) {
    toastErrorHandler('Для этого модуля не настроена страница записи')(new Error('no record page'))
    return
  }
  router.push({ name: 'page.record', params: { pageID: page.pageID, recordID: a.subjectRecordID } })
}

onMounted(() => {
  loadBindings().then(() => {
    if (bindings.value.length) select(bindings.value[0])
  })
  loadModels()
  if (props.namespace) {
    store.module.load({ namespace: props.namespace })
    store.page.load({ namespaceID: props.namespace.namespaceID })
  }
})
</script>

<style scoped>
.risk-record-row:hover {
  background-color: var(--bs-tertiary-bg, #f8f9fa);
}
</style>
