<template>
  <div class="card shadow-sm">
    <div class="card-header">
      <form class="d-flex flex-column w-100" @submit.prevent="search">
        <div class="row">
          <div class="col-12 col-lg-6 mb-2">
            <label class="text-primary small">{{ $t('filter.from') }}</label>
            <c-input-date-time v-model="filter.from" :labels="{ clear: $t('label.clear'), none: $t('label.none'), now: $t('label.now'), today: $t('label.today') }" />
          </div>
          <div class="col-12 col-lg-6 mb-2">
            <label class="text-primary small">{{ $t('filter.to') }}</label>
            <c-input-date-time v-model="filter.to" only-past :labels="{ clear: $t('label.clear'), none: $t('label.none'), now: $t('label.now'), today: $t('label.today') }" />
          </div>
        </div>
        <div class="row">
          <div class="col-12 col-lg-4 mb-2">
            <label class="text-primary small">{{ $t('filter.resource') }}</label>
            <input v-model="filter.resource" class="form-control form-control-sm" data-test-id="input-resource">
          </div>
          <div class="col-12 col-lg-4 mb-2">
            <label class="text-primary small">{{ $t('filter.action') }}</label>
            <input v-model="filter.action" class="form-control form-control-sm" data-test-id="input-action">
          </div>
          <div class="col-12 col-lg-4 mb-2">
            <label class="text-primary small">{{ $t('filter.actor') }}</label>
            <input v-model="filter.actorID" class="form-control form-control-sm" data-test-id="input-user-id">
          </div>
        </div>
        <div class="d-flex">
          <button type="submit" class="btn btn-primary ms-auto" :disabled="processing" data-test-id="button-submit">
            {{ $t('filter.search') }}
          </button>
        </div>
      </form>
    </div>

    <div class="table-responsive">
      <table class="table table-hover mb-0 small">
        <thead class="table-secondary">
          <tr>
            <th v-for="f in fields" :key="f.key">{{ f.label }}</th>
          </tr>
        </thead>
        <tbody v-if="!processing">
          <template v-for="a in items" :key="a.actionID">
            <tr class="pointer" @click="a._showDetails = !a._showDetails">
              <td class="text-nowrap">{{ $locFullDateTime(a.timestamp) }}</td>
              <td>
                <router-link v-if="a.actorID && a.actorID !== '0'" :to="drillDownLink({ actorID: a.actorID })" data-test-id="item-user-id">
                  {{ a.actor || a.actorID }}
                </router-link>
              </td>
              <td>{{ a.requestOrigin }}</td>
              <td>
                <router-link :to="drillDownLink({ resource: a.resource })" data-test-id="item-resource">
                  {{ a.resource }}
                </router-link>
              </td>
              <td>
                <router-link :to="drillDownLink({ action: a.action })" data-test-id="item-action">
                  {{ a.action }}
                </router-link>
              </td>
              <td>{{ a.description }}</td>
              <td class="text-end" :class="severity[a.severity]?.class">{{ severity[a.severity]?.label }}</td>
            </tr>
            <tr v-if="a._showDetails" :key="'detail-' + a.actionID">
            <td colspan="7">
              <div class="row">
                <div class="col-6">
                  <h6>{{ $t('details.header') }}</h6>
                  <div v-for="(val, key) in detailFields(a)" :key="key" class="row small mb-1">
                    <div class="col-4 text-muted">{{ key }}</div>
                    <div class="col-8">{{ val }}</div>
                  </div>
                </div>
                <div class="col-6">
                  <h6>{{ $t('details.headerAdditional') }}</h6>
                  <div class="row small mb-1">
                    <div class="col-4 text-muted">{{ $t('details.description') }}</div>
                    <div class="col-8">{{ a.description }}</div>
                  </div>
                  <div v-if="a.error" class="row small mb-1">
                    <div class="col-4 text-muted">{{ $t('details.error') }}</div>
                    <div class="col-8 text-danger">{{ a.error }}</div>
                  </div>
                  <template v-if="a.meta">
                    <hr>
                    <h6>{{ $t('details.meta') }}</h6>
                    <div v-for="(val, key) in a.meta" :key="key" class="row small mb-1">
                      <div class="col-6"><code>{{ key }}</code></div>
                      <div class="col-6"><code>{{ val }}</code></div>
                    </div>
                  </template>
                </div>
              </div>
            </td>
          </tr>
          </template>
        </tbody>
        <tbody v-else>
          <tr>
            <td colspan="7" class="text-center p-3">
              <div class="spinner-border spinner-border-sm me-2" />
              {{ $t('loading') }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="items.length" class="card-footer d-flex border-top">
      <button class="btn btn-outline-secondary mx-auto" :disabled="processing" @click="load()" data-test-id="button-load-older-actions">
        {{ $t('loadOlder') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'

const { CInputDateTime } = components
const { t } = useI18n()

const processing = ref(false)
const items = ref([])

const filter = reactive({
  from: undefined,
  to: undefined,
  beforeActionID: undefined,
  actorID: undefined,
  resource: undefined,
  action: undefined,
})

const pagination = reactive({ limit: 10 })

const severity = [
  { label: t('severity.emergency'), class: 'text-danger' },
  { label: t('severity.alert'), class: 'text-danger' },
  { label: t('severity.critical'), class: 'text-danger' },
  { label: t('severity.error'), class: 'text-danger' },
  { label: t('severity.warning'), class: 'text-warning' },
  { label: t('severity.notice'), class: 'text-success' },
  { label: t('severity.info'), class: 'text-success' },
  { label: t('severity.debug'), class: '' },
]

const fields = computed(() => [
  { key: 'timestamp', label: t('columns.timestamp') },
  { key: 'actor', label: t('columns.actor') },
  { key: 'requestOrigin', label: t('columns.requestOrigin') },
  { key: 'resource', label: t('columns.resource') },
  { key: 'action', label: t('columns.action') },
  { key: 'description', label: t('columns.description') },
  { key: 'severity', label: '' },
])

function search() {
  load(true)
}

function load(reset = false) {
  if (reset) {
    items.value = []
    pagination.beforeActionID = undefined
  } else {
    const len = items.value.length
    if (len > 0) {
      pagination.beforeActionID = items.value[len - 1]?.actionID
    }
  }

  const f = { ...filter, ...pagination }
  if (!f.actorID) delete f.actorID
  if (!f.action) delete f.action
  if (!f.resource) delete f.resource

  processing.value = true

  window.__systemAPI.actionlogList(f)
    .then(rr => {
      items.value.push(...(rr || []))
    })
    .finally(() => {
      processing.value = false
    })
}

function drillDownLink(query = {}) {
  return {
    name: 'system.actionlog',
    query: { ...filter, ...query, sort: undefined },
  }
}

function detailFields(a) {
  return {
    [t('details.id')]: a.actionID,
    [t('details.timestamp')]: a.timestamp,
    [t('details.requestOrigin')]: a.requestOrigin,
    [t('details.requestID')]: a.requestID,
    [t('details.actorIPAddr')]: a.actorIPAddr,
    [t('details.actor')]: a.actor,
    [t('details.actorID')]: a.actorID,
    [t('details.severity')]: severity[a.severity]?.label,
    [t('details.resource')]: a.resource,
    [t('details.action')]: a.action,
  }
}

onMounted(() => load())
</script>
