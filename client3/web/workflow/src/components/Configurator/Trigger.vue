<template>
  <div>
    <div class="card flex-grow-1 border-bottom border-light rounded-0">
      <div class="card-header p-0 mb-3">
        <h5 class="mb-0">{{ t('configurator.configuration') }}</h5>
      </div>
      <div class="card-body p-0">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('steps.trigger.configurator.resource*') }}</label>
          <c-input-select
            v-model="item.triggers.resourceType"
            :options="resourceTypeOptions"
            :get-option-key="getOptionTypeKey"
            label="text"
            :reduce="r => r.value"
            :filter="resTypeFilter"
            :placeholder="t('steps.trigger.configurator.select-resource-type')"
            :clearable="false"
            @input="resourceChanged"
          />
        </div>

        <div v-if="item.triggers.resourceType" class="mb-3">
          <label class="text-primary form-label">{{ t('steps.trigger.configurator.event*') }}</label>
          <c-input-select
            v-model="item.triggers.eventType"
            :options="eventTypeOptions"
            :get-option-key="getOptionEventTypeKey"
            :get-option-label="getEventTypeLabel"
            :reduce="e => e.eventType"
            :filter="evtTypeFilter"
            :placeholder="t('steps.trigger.configurator.select-event-type')"
            :clearable="false"
            @input="eventChanged"
          />
        </div>

        <div class="mb-0">
          <div class="form-check">
            <input class="form-check-input-v3" type="checkbox" v-model="item.triggers.enabled"
              :disabled="isSubworkflow && !item.triggers.enabled" id="trigger-enabled" @change="enabledChanged()" />
            <label class="form-check-label text-primary" for="trigger-enabled">{{ t('enabled') }}</label>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showConstraints" class="card flex-grow-1 border-bottom border-light rounded-0">
      <div class="card-header d-flex align-items-center">
        <h5 class="mb-0">{{ t('steps.trigger.configurator.constraints') }}</h5>
        <button v-if="constraintNameTypes.length" class="btn btn-primary ms-3" @click="addConstraint()">
          {{ t('steps.trigger.configurator.add-constraints') }}
        </button>
      </div>
      <div class="card-body p-0">
        <table v-if="constraintNameTypes.length" class="table table-borderless table-hover mb-0">
          <thead class="table-secondary">
            <tr>
              <th class="pointer">{{ t('steps.trigger.configurator.resource') }}</th>
              <th class="text-center">{{ t('steps.trigger.configurator.operator') }}</th>
              <th class="text-center"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(c, index) in item.triggers.constraints" :key="index"
              :class="c._showDetails ? 'border-thick' : 'border-thick-transparent'"
              @click="c._showDetails = !c._showDetails">
              <td class="pointer">{{ getConstraintNameLabel(c.name) }}</td>
              <td class="text-center">{{ getConstraintOperatorLabel(c.op) }}</td>
              <td class="d-flex align-items-start gap-2 pointer">
                {{ c.values.join(' or ') }}
                <c-input-confirm show-icon class="delete-btn ms-auto" @confirmed="removeConstraint(index)" />
              </td>
            </tr>
            <tr v-if="!item.triggers.constraints.length">
              <td colspan="3" class="text-center text-muted p-4">{{ t('steps.trigger.configurator.no-constraints') }}</td>
            </tr>
          </tbody>
        </table>

        <div v-if="c._showDetails" v-for="(c, index) in item.triggers.constraints" :key="'detail-' + index" class="mb-3 px-3">
          <div class="arrow-up"></div>
          <div class="card bg-light">
            <div class="card-body px-4 pb-3">
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('steps.trigger.configurator.resource') }}</label>
                <c-input-select v-model="c.name" :options="constraintNameTypes" :get-option-key="getOptionTypeKey"
                  label="text" :reduce="c => c.value" :filter="constrFilter"
                  :placeholder="t('steps.trigger.configurator.select-constraint-type')" :clearable="false"
                  @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
              </div>
              <div class="mb-3">
                <label class="text-primary form-label">{{ t('steps.trigger.configurator.operator') }}</label>
                <c-input-select v-model="c.op" :options="constraintOperatorTypes" :get-option-key="getOptionTypeKey"
                  label="text" :reduce="c => c.value"
                  :placeholder="t('steps.trigger.configurator.select-operator')" :clearable="false"
                  @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
              </div>
              <div class="mb-0">
                <label class="text-primary form-label">{{ t('steps.trigger.configurator.value') }}</label>
                <div v-for="(value, vi) in c.values" :key="vi" class="mb-2">
                  <p v-if="vi > 0" class="text-center text-uppercase text-muted mb-2">{{ t('label.or') }}</p>
                  <div class="d-flex align-items-center gap-1">
                    <input class="form-control" v-model="c.values[vi]"
                      @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
                    <c-input-confirm show-icon @confirmed="c.values.splice(vi, 1)" />
                  </div>
                </div>
                <button class="btn btn-primary btn-sm" @click="c.values.push('')">{{ t('steps.trigger.configurator.add') }}</button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="!constraintNameTypes.length && item.triggers.constraints[0]" class="mt-0 mb-4 mx-4">
          <label class="d-flex align-items-center text-primary form-label">
            {{ item.triggers.eventType.replace('on', '') }}
            <a v-if="intervalDocumentationURL" :href="intervalDocumentationURL" target="_blank" class="d-flex align-items-center h6 mb-0 ms-1 pointer">
              <font-awesome-icon :icon="['far', 'question-circle']" />
            </a>
          </label>

          <c-input-date-time
            v-if="item.triggers.eventType === 'onTimestamp'"
            v-model="item.triggers.constraints[0].values[0]"
            :labels="{ clear: t('clear'), none: t('none'), now: t('now'), today: t('today') }"
            @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
          <input v-else class="form-control" v-model="item.triggers.constraints[0].values[0]"
            @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
        </div>
      </div>
    </div>

    <div v-if="(eventType.properties || []).length" class="card flex-grow-1 rounded-0">
      <div class="card-header">
        <h5 class="mb-0">{{ t('steps.trigger.configurator.initial-scope') }}</h5>
      </div>
      <div class="card-body p-0">
        <table class="table table-borderless mb-4">
          <thead class="table-secondary">
            <tr>
              <th class="ps-3 py-2">{{ t('steps.function.configurator.name') }}</th>
              <th class="pe-3 py-2">{{ t('steps.function.configurator.type') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="v in eventType.properties" :key="v.name">
              <td class="text-truncate">{{ v.name }}</td>
              <td class="text-truncate"><var>{{ v.type || t('label.any') }}</var></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, inject, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
import { objectSearchMaker } from '../../lib/filter'
import { getConstraintNameLabel } from '../../lib/constraint'
import { getDocumentationURL } from '../../lib/version'
import { camelToTitle } from '../../lib/string'

const { CInputDateTime } = components

const { t } = useI18n()
const $AutomationAPI = inject('automationAPI', {})

const props = defineProps({
  item: { type: Object, default: () => ({}) },
  edges: { type: Object, default: () => ({}) },
  outEdges: { type: Number, default: 0 },
  isSubworkflow: { type: Boolean, default: false },
})

const emit = defineEmits(['update-value', 'update-default-value'])

const eventTypes = ref([])
const resourceTypes = ref([])

const resourceTypeOptions = computed(() => resourceTypes.value)

const eventTypeOptions = computed(() =>
  eventTypes.value.filter(({ resourceType }) => resourceType === props.item.triggers.resourceType)
)

const eventType = computed(() =>
  eventTypes.value.find(({ resourceType, eventType }) =>
    resourceType === props.item.triggers.resourceType && eventType === props.item.triggers.eventType
  ) || {}
)

const showConstraints = computed(() => {
  if (props.item.triggers.resourceType && props.item.triggers.eventType) {
    return constraintNameTypes.value.length ? true : props.item.triggers.eventType !== 'onManual'
  }
  return false
})

const constraintNameTypes = computed(() => {
  const constraints = eventType.value.constraints || []
  return constraints.reduce((acc, { name }) => {
    if (!name.includes('*')) {
      acc.push({ value: name, text: getConstraintNameLabel(name) })
    }
    return acc
  }, [])
})

const constraintOperatorTypes = [
  { value: '=', text: t('steps.trigger.configurator.equal') },
  { value: '!=', text: t('steps.trigger.configurator.not-equal') },
  { value: 'like', text: t('steps.trigger.configurator.like') },
  { value: 'not like', text: t('steps.trigger.configurator.not-like') },
]

const intervalDocumentationURL = computed(() =>
  getDocumentationURL('integrator-guide/automation/workflows/index.html#deferred-interval')
)

const resTypeFilter = objectSearchMaker('text')
const evtTypeFilter = objectSearchMaker('eventType')
const constrFilter = objectSearchMaker('text')

onMounted(async () => {
  if (!props.item.triggers) {
    props.item.triggers = { resourceType: null, eventType: null, constraints: [], enabled: true }
  }
  await getEventTypes()
})

async function getEventTypes() {
  return $AutomationAPI.eventTypesList()
    .then(({ set }) => {
      eventTypes.value = set
      const rts = new Set(set.map(({ resourceType }) => resourceType))
      resourceTypes.value = [...rts].map(resourceType => ({
        value: resourceType,
        text: getResourceTypeLabel(resourceType),
      }))
    })
}

function addConstraint() {
  props.item.triggers.constraints.push({ name: '', op: '=', values: [''], _showDetails: true })
  window.dispatchEvent(new CustomEvent('change-detected'))
}

function removeConstraint(index) {
  props.item.triggers.constraints.splice(index, 1)
  window.dispatchEvent(new CustomEvent('change-detected'))
}

function resourceChanged() {
  props.item.triggers.eventType = null
  props.item.triggers.constraints = []
  window.dispatchEvent(new CustomEvent('change-detected'))
  updateDefaultName()
}

function eventChanged() {
  if (['onTimestamp', 'onInterval'].includes(props.item.triggers.eventType)) {
    props.item.triggers.constraints = []
    addConstraint()
  }
  window.dispatchEvent(new CustomEvent('change-detected'))
  updateDefaultName()
}

function enabledChanged() {
  window.dispatchEvent(new CustomEvent('trigger-updated', { detail: props.item.node }))
  window.dispatchEvent(new CustomEvent('change-detected'))
}

function updateDefaultName() {
  const { resourceType, eventType } = props.item.triggers
  if (resourceType) {
    let value = [getResourceTypeLabel(resourceType), getEventTypeLabel({ eventType })].filter(v => v).join(' - ')
    value = value.charAt(0).toUpperCase() + value.slice(1)
    emit('update-default-value', { value, force: !props.item.node.value })
  }
}

function getOptionTypeKey({ value }) { return value }
function getOptionEventTypeKey({ eventType }) { return eventType }

function getResourceTypeLabel(resourceType) {
  if (!resourceType) return ''
  return resourceType.split(':').map(part => part.split('-').map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase()).join(' ')).join(' - ')
}

function getEventTypeLabel({ eventType = '' } = {}) {
  if (!eventType) return ''
  return camelToTitle(eventType.replace('on', ''))
}

function getConstraintOperatorLabel(op) {
  const operator = constraintOperatorTypes.find(type => type.value === op)
  return operator ? operator.text : op
}
</script>

<style scoped>
.delete-btn {
  display: none !important;
}
tr:hover .delete-btn {
  display: inline-flex !important;
}
</style>
