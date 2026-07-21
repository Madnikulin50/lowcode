<template>
  <div v-if="filter.params.length">
    <div
      v-for="(param, index) in filter.params"
      :key="index"
      class="mb-3"
    >
      <label class="form-label text-primary">
        {{ $t(`filters.labels.${param.label}`) }}

        <template v-if="param.label === 'expr'">
          <a
            v-if="param.label === 'expr'"
            :href="documentationURL"
            target="_blank"
          >
            <font-awesome-icon
              :icon="['far', 'question-circle']"
            />
          </a>
        </template>
      </label>

      <div
        v-if="param.type === 'bool'"
        class="form-check"
      >
        <input
          :id="'param-' + index"
          v-model="param.value"
          type="checkbox"
          class="form-check-input"
        >
        <label
          :for="'param-' + index"
          class="form-check-label"
        />
      </div>

      <c-input-select
        v-else-if="param.label === 'workflow'"
        v-model="param.value"
        :options="workflows"
        :get-option-key="getOptionKey"
        :reduce="wf => wf.workflowID"
        :placeholder="$t('filters.placeholders.workflow')"
        :loading="loadingWorkflows"
      />

      <select
        v-else-if="param.label === 'status'"
        v-model="param.value"
        class="form-select"
      >
        <option :value="undefined">
          {{ $t('filters.httpStatus.none') }}
        </option>
        <option
          v-for="opt in httpStatusOptions"
          :key="opt.value"
          :value="opt.value"
        >
          {{ opt.text }}
        </option>
      </select>

      <template v-else-if="filter.ref === 'response'">
        <template v-if="param.type === 'input'">
          <select
            v-model="param.value.type"
            class="form-select mb-2"
          >
            <option
              v-for="opt in inputTypeOptions"
              :key="opt"
              :value="opt"
            >
              {{ opt }}
            </option>
          </select>

          <div class="input-group mb-2">
            <span class="input-group-text">ƒ</span>
            <input
              v-model="param.value.expr"
              type="text"
              class="form-control"
              :placeholder="$t('filters.help.expression.example')"
            >
          </div>
        </template>

        <template v-else>
          <c-form-table-wrapper
            :labels="{ addButton: $t('filters.addHeader') }"
            @add-item="param.value.push({ name: '', expr: '' })"
          >
            <div
              v-for="(header, hIndex) in param.value"
              :key="`header-${hIndex}`"
              class="input-group mb-2"
            >
              <input
                v-model="header.name"
                type="text"
                class="form-control"
                :placeholder="$t('filters.labels.name')"
              >
              <input
                v-model="header.expr"
                type="text"
                class="form-control"
                :placeholder="$t('filters.labels.value')"
              >

              <c-input-confirm
                show-icon
                @confirmed="param.value.splice(hIndex, 1)"
              />
            </div>
          </c-form-table-wrapper>
        </template>
      </template>

      <template v-else>
        <textarea
          v-if="param.label === 'jsfunc'"
          v-model="param.value"
          class="form-control"
          rows="3"
        />

        <div
          v-else
          class="input-group"
        >
          <span
            v-if="param.label === 'expr'"
            class="input-group-text"
          >ƒ</span>

          <input
            v-if="param.label === 'expr'"
            v-model="param.value"
            type="text"
            class="form-control"
            :placeholder="$t('filters.help.expression.example')"
          >
          <input
            v-else
            v-model="param.value"
            type="text"
            class="form-control"
          >
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const { proxy } = getCurrentInstance()
const $SystemAPI = proxy.$SystemAPI

const props = defineProps({
  filter: {
    type: Object,
    default: () => ({}),
  },
})

const loadingWorkflows = ref(false)
const workflows = ref([])

const httpStatusOptions = ref([
  { value: 300, text: '' },
  { value: 301, text: '' },
  { value: 302, text: '' },
  { value: 303, text: '' },
  { value: 304, text: '' },
  { value: 307, text: '' },
  { value: 308, text: '' },
])

const inputTypeOptions = ref([
  'String',
  'Any',
  'Array',
  'KV',
  'DateTime',
  'Float',
  'Integer',
  'Reader',
  'Vars',
])

const documentationURL = computed(() => {
  const [year, month] = VERSION.split('.')
  return `https://docs.cortezaproject.org/corteza-docs/${year}.${month}/integrator-guide/expr/index.html`
})

function getOptionKey ({ workflowID }) {
  return workflowID
}

onMounted(() => {
  httpStatusOptions.value.forEach(o => { o.text = t(`filters.httpStatus.${o.value}`) })

  if (props.filter.params.some(({ label = '' }) => label === 'workflow')) {
    loadingWorkflows.value = true

    $SystemAPI.workflowList()
      .then(({ set: workflowsList = [] }) => {
        workflows.value = workflowsList.map(({ workflowID, handle, meta }) => {
          return { label: meta.name || handle, workflowID }
        })
      })
      .finally(() => {
        loadingWorkflows.value = false
      })
  }
})
</script>
