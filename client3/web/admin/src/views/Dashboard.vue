<template>
  <div class="container-fluid d-flex flex-column pt-2 pb-3 flex-fill">
    <c-content-header :title="$t('dashboard.title')">
      <c-corredor-manual-buttons
        ui-page="dashboard"
        ui-slot="toolbar"
        resource-type="system'"
        @click="dispatchCortezaSystemEvent($event)"
      />
    </c-content-header>

    <div class="row g-3">
      <div class="col-12 col-md-4">
        <router-link
          :to="{ name: 'system.user.list' }"
          class="ae-kpi card shadow-sm text-decoration-none"
        >
          <div class="ae-kpi-value">{{ users.valid }}</div>
          <div class="ae-kpi-label">{{ tr('dashboard.kpi.users', 'Users') }}</div>
          <div class="ae-kpi-meta">
            <span><strong>{{ users.total }}</strong> {{ $t('dashboard.users.total') }}</span>
            <span><strong>{{ users.suspended }}</strong> {{ $t('dashboard.users.suspended') }}</span>
            <span><strong>{{ users.deleted }}</strong> {{ $t('dashboard.users.deleted') }}</span>
          </div>
        </router-link>
      </div>

      <div class="col-12 col-md-4">
        <router-link
          :to="{ name: 'system.role.list' }"
          class="ae-kpi card shadow-sm text-decoration-none"
        >
          <div class="ae-kpi-value">{{ roles.valid }}</div>
          <div class="ae-kpi-label">{{ tr('dashboard.kpi.roles', 'Roles') }}</div>
          <div class="ae-kpi-meta">
            <span><strong>{{ roles.total }}</strong> {{ $t('dashboard.roles.total') }}</span>
            <span><strong>{{ roles.archived }}</strong> {{ $t('dashboard.roles.archived') }}</span>
            <span><strong>{{ roles.deleted }}</strong> {{ $t('dashboard.roles.deleted') }}</span>
          </div>
        </router-link>
      </div>

      <div class="col-12 col-md-4">
        <router-link
          :to="{ name: 'system.application.list' }"
          class="ae-kpi card shadow-sm text-decoration-none"
        >
          <div class="ae-kpi-value">{{ applications.valid }}</div>
          <div class="ae-kpi-label">{{ tr('dashboard.kpi.applications', 'Applications') }}</div>
          <div class="ae-kpi-meta">
            <span><strong>{{ applications.total }}</strong> {{ $t('dashboard.applications.total') }}</span>
            <span><strong>{{ applications.deleted }}</strong> {{ $t('dashboard.applications.deleted') }}</span>
          </div>
        </router-link>
      </div>
    </div>

    <div
      v-if="userChart"
      class="card shadow-sm mt-3"
      style="min-height: 280px;"
    >
      <div class="card-header border-bottom">
        <div class="ae-section-title">{{ $t('dashboard.users.created') }}</div>
      </div>
      <div class="card-body position-relative p-0" style="min-height: 240px;">
        <c-chart :chart="userChart" />
      </div>
    </div>

    <div
      v-else-if="loaded && users.total === 0"
      class="card shadow-sm mt-3"
    >
      <div class="ae-empty">
        <div class="ae-empty-icon">
          <font-awesome-icon :icon="['fas', 'inbox']" />
        </div>
        <p class="ae-empty-title mb-1">{{ tr('dashboard.empty.title', 'Nothing to show yet') }}</p>
        <p class="ae-empty-text mb-3">{{ tr('dashboard.empty.text', 'Create a user to see activity here.') }}</p>
        <router-link
          :to="{ name: 'system.user.new' }"
          class="btn btn-primary btn-sm"
        >
          {{ tr('dashboard.empty.create', 'New user') }}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'dashboard' } })
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
import moment from 'moment'

const { CChart } = components
const { t } = useI18n()

const loaded = ref(false)
const userChart = ref(null)
const users = ref({ total: 0, valid: 0, deleted: 0, suspended: 0, dailyCreated: [], dailyUpdated: [], dailySuspended: [], dailyDeleted: [] })
const roles = ref({ total: 0, valid: 0, archived: 0, deleted: 0 })
const applications = ref({ total: 0, valid: 0, deleted: 0 })

function tr (key, fallback) {
  const v = t(key)
  return (!v || v === key || v.endsWith(`.${key}`)) ? fallback : v
}

onMounted(() => {
  window.__systemAPI.statsList().then(({ users: u, roles: r, applications: a }) => {
    if (u) users.value = u
    if (r) roles.value = r
    if (a) applications.value = a
    initUserChart()
  }).finally(() => { loaded.value = true })
})

function initUserChart() {
  if (users.value.total === 0) return

  const themeVariables = getThemeVariables()
  const { dates, values } = getUserTimeline()

  userChart.value = {
    tooltip: { trigger: 'axis' },
    textStyle: { fontFamily: themeVariables['font-regular'], color: themeVariables.black },
    xAxis: { type: 'category', data: dates, boundaryGap: false, axisTick: { show: false }, axisLine: { show: false } },
    yAxis: { type: 'value', axisLine: { show: false, onZero: false }, splitLine: { lineStyle: { color: [themeVariables['extra-light']] } } },
    grid: { top: 20, right: 50, bottom: 20, left: 40, containLabel: true },
    series: [{ name: t('dashboard.users.created'), type: 'line', data: values, smooth: 0.5, areaStyle: { opacity: 0.5 } }],
  }
}

function getThemeVariables() {
  const getCssVariable = (variableName) => getComputedStyle(document.documentElement).getPropertyValue(variableName).trim()
  return ['white', 'black', 'primary', 'secondary', 'success', 'warning', 'danger', 'light', 'extra-light', 'dark', 'font-regular'].reduce((acc, variable) => {
    acc[variable] = getCssVariable(`--${variable}`)
    return acc
  }, {})
}

function getUserTimeline() {
  const data = users.value.dailyCreated || []
  const unit = getComfortableTimeUnit(data)
  const aux = {}
  for (let i = 0; i < data.length; i += 2) {
    const ts = moment.unix(data[i]).startOf(unit).format(unit === 'month' ? 'MMM YYYY' : 'D MMM YYYY')
    aux[ts] = (aux[ts] || 0) + data[i + 1]
  }
  const dates = []
  const values = []
  for (const date in aux) {
    dates.push(date)
    values.push(aux[date])
  }
  return { dates, values }
}

function getComfortableTimeUnit(range) {
  if (range.length === 0) return undefined
  if (range.length === 2) return 'day'
  const ts = range.filter((v, i) => i % 2 === 0).sort()
  const min = ts[0]
  const max = ts[ts.length - 1]
  const diffInDays = (max - min) / (60 * 60 * 24)
  return diffInDays > 120 ? 'month' : 'day'
}

function dispatchCortezaSystemEvent($event) {
  // Placeholder for event bus dispatch
}
</script>
