<template>
  <div class="container d-flex flex-column pt-2 pb-3 flex-fill">
    <c-content-header :title="$t('title')">
      <c-corredor-manual-buttons
        ui-page="dashboard"
        ui-slot="toolbar"
        resource-type="system'"
        @click="dispatchCortezaSystemEvent($event)"
      />
    </c-content-header>

    <div class="row flex-fill">
      <div class="col-12">
        <div class="card shadow-sm h-100" style="min-height: 500px;">
          <div class="card-header border-bottom">
            <h4 class="card-title">
              <router-link
                :to="{ name: 'system.user.list' }"
                :area-label="`${users.valid} ${$t('users.title')}`"
                class="display-3 text-decoration-none"
              >
                {{ users.valid }}
              </router-link>
            </h4>
            <h4>
              {{ $t('users.title') }}
            </h4>
          </div>

          <div class="card-body position-relative p-0">
            <c-chart
              v-if="userChart"
              :chart="userChart"
            />
          </div>

          <div class="card-footer border-top">
            <div class="row">
              <div class="col-12 col-sm-4 mb-2 mb-sm-0">
                <router-link
                  :to="{ name: 'system.user.list', query: { deleted: 1, suspended: 1 } }"
                  :aria-label="users.total + ' ' + $t('users.users') + ' ' + $t('users.total')"
                  class="text-decoration-none"
                >
                  {{ users.total }}
                </router-link>
                <span class="d-sm-block">
                  {{ $t('users.total') }}
                </span>
              </div>
              <div class="col-12 col-sm-4 mb-2 mb-sm-0">
                <router-link
                  :to="{ name: 'system.user.list', query: { deleted: 1, suspended: 2 } }"
                  :aria-label="users.suspended + ' ' + $t('users.users') + ' ' + $t('users.suspended')"
                  class="text-decoration-none"
                >
                  {{ users.suspended }}
                </router-link>
                <span class="d-sm-block">
                  {{ $t('users.suspended') }}
                </span>
              </div>
              <div class="col-12 col-sm-4 mb-2 mb-sm-0">
                <router-link
                  :to="{ name: 'system.user.list', query: { deleted: 2, suspended: 1 } }"
                  :aria-label="users.deleted + ' ' + $t('users.users') + ' ' + $t('users.deleted')"
                  class="text-decoration-none"
                >
                  {{ users.deleted }}
                </router-link>
                <span class="d-sm-block">
                  {{ $t('users.deleted') }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="row align-items-stretch">
      <div
        v-show="roles.total"
        class="col-12 col-md-6 mt-3"
      >
        <div class="card shadow-sm h-100">
          <div class="card-header border-bottom">
            <h4 class="card-title">
              <router-link
                :to="{ name: 'system.role.list' }"
                :aria-label="roles.valid + ' ' + $t('roles.title')"
                class="display-4 text-decoration-none"
              >
                {{ roles.valid }}
              </router-link>
            </h4>
            <h4>
              {{ $t('roles.title') }}
            </h4>
          </div>

          <div class="card-footer border-top">
            <div class="row">
              <div class="col-12 col-sm-4 mb-2 mb-sm-0">
                <router-link
                  :to="{ name: 'system.role.list', query: { deleted: 1, archived: 1 } }"
                  :aria-label="roles.total + ' ' + $t('roles.roles') + ' ' + $t('roles.total')"
                  class="text-decoration-none"
                >
                  {{ roles.total }}
                </router-link>
                <span class="d-sm-block">
                  {{ $t('roles.total') }}
                </span>
              </div>
              <div class="col-12 col-sm-4 mb-2 mb-sm-0">
                <router-link
                  :to="{ name: 'system.role.list', query: { deleted: 1, archived: 2 } }"
                  :aria-label="roles.archived + ' ' + $t('roles.roles') + ' ' + $t('roles.archived')"
                  class="text-decoration-none"
                >
                  {{ roles.archived }}
                </router-link>
                <span class="d-sm-block">
                  {{ $t('roles.archived') }}
                </span>
              </div>
              <div class="col-12 col-sm-4 mb-2 mb-sm-0">
                <router-link
                  :to="{ name: 'system.role.list', query: { deleted: 2, archived: 1 } }"
                  :aria-label="roles.deleted + ' ' + $t('roles.roles') + ' ' + $t('roles.deleted')"
                  class="text-decoration-none"
                >
                  {{ roles.deleted }}
                </router-link>
                <span class="d-sm-block">
                  {{ $t('roles.deleted') }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-show="applications.total"
        class="col-12 col-md-6 mt-3"
      >
        <div class="card shadow-sm h-100">
          <div class="card-header border-bottom">
            <h4 class="card-title">
              <router-link
                :to="{ name: 'system.application.list' }"
                :aria-label="applications.valid + ' ' + $t('applications.title')"
                class="display-4 text-decoration-none"
              >
                {{ applications.valid }}
              </router-link>
            </h4>
            <h4>
              {{ $t('applications.title') }}
            </h4>
          </div>

          <div class="card-footer border-top">
            <div class="row">
              <div class="col-12 col-sm-4 mb-2 mb-sm-0">
                <router-link
                  :to="{ name: 'system.application.list', query: { deleted: 1 } }"
                  :aria-label="applications.total + ' ' + $t('applications.applications') + ' ' + $t('applications.total')"
                  class="text-decoration-none"
                >
                  {{ applications.total }}
                </router-link>
                <span class="d-sm-block">
                  {{ $t('applications.total') }}
                </span>
              </div>
              <div class="col-12 col-sm-4 mb-2 mb-sm-0">
                <router-link
                  :to="{ name: 'system.application.list', query: { deleted: 2 } }"
                  :aria-label="applications.deleted + ' ' + $t('applications.applications') + ' ' + $t('applications.deleted')"
                  class="text-decoration-none"
                >
                  {{ applications.deleted }}
                </router-link>
                <span class="d-sm-block">
                  {{ $t('applications.deleted') }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
import moment from 'moment'

const { CChart } = components
const { t } = useI18n()

const userChart = ref(null)
const users = ref({ total: 0, valid: 0, deleted: 0, suspended: 0, dailyCreated: [], dailyUpdated: [], dailySuspended: [], dailyDeleted: [] })
const roles = ref({ total: 0, valid: 0, archived: 0, deleted: 0 })
const applications = ref({ total: 0, valid: 0, deleted: 0 })

onMounted(() => {
  window.__systemAPI.statsList().then(({ users: u, roles: r, applications: a }) => {
    if (u) users.value = u
    if (r) roles.value = r
    if (a) applications.value = a
    initUserChart()
  })
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
    series: [{ name: t('users.created'), type: 'line', data: values, smooth: 0.5, areaStyle: { opacity: 0.5 } }],
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
  const data = users.value.dailyCreated
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
