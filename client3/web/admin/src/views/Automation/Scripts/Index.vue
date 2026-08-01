<template>
  <div class="container-fluid d-flex flex-column h-100 pt-2 pb-3">
    <c-content-header :title="$t('automation.scripts.list.title')" />
    <div class="card flex-fill shadow-sm">
      <div class="card-header">
        <form @submit.prevent="search">
          <div class="row mb-2">
            <div class="col-12 col-lg-6">
              <label class="text-primary mb-1">{{ $t('filter.searchQuery') }}</label>
              <input v-model="filter.query" class="form-control form-control-sm" />
            </div>
          </div>
          <div class="row">
            <div class="col-12 col-lg-6">
              <label class="text-primary mb-1">{{ $t('filter.incScriptsWithErrors', { count: totalScriptsWithErrors }) }}</label>
              <div class="form-check form-switch"><input v-model="filter.incScriptsWithErrors" type="checkbox" class="form-check-input-v3" role="switch" /><label class="form-check-label">{{ filter.incScriptsWithErrors ? $t('label.general.yes') : $t('label.general.no') }}</label></div>
            </div>
            <div class="col-12 col-lg-6">
              <label class="text-primary mb-1">{{ $t('filter.incScriptsWithTriggers', { count: totalScriptsWithTriggers }) }}</label>
              <div class="form-check form-switch"><input v-model="filter.incScriptsWithTriggers" type="checkbox" class="form-check-input-v3" role="switch" /><label class="form-check-label">{{ filter.incScriptsWithTriggers ? $t('label.general.yes') : $t('label.general.no') }}</label></div>
            </div>
            <div class="col-12 col-lg-6">
              <label class="text-primary mb-1">{{ $t('filter.incScriptsWithIterator', { count: totalScriptsWithIterator }) }}</label>
              <div class="form-check form-switch"><input v-model="filter.incScriptsWithIterator" type="checkbox" class="form-check-input-v3" role="switch" /><label class="form-check-label">{{ filter.incScriptsWithIterator ? $t('label.general.yes') : $t('label.general.no') }}</label></div>
            </div>
            <div class="col-12 col-lg-6">
              <label class="text-primary mb-1">{{ $t('filter.incScriptsWithSecurity', { count: totalScriptsWithSecurity }) }}</label>
              <div class="form-check form-switch"><input v-model="filter.incScriptsWithSecurity" type="checkbox" class="form-check-input-v3" role="switch" /><label class="form-check-label">{{ filter.incScriptsWithSecurity ? $t('label.general.yes') : $t('label.general.no') }}</label></div>
            </div>
            <div class="col-12 col-lg-6">
              <label class="text-primary mb-1">{{ $t('filter.absoluteTime') }}</label>
              <div class="form-check form-switch"><input v-model="filter.absoluteTime" type="checkbox" class="form-check-input-v3" role="switch" /><label class="form-check-label">{{ filter.absoluteTime ? $t('label.general.yes') : $t('label.general.no') }}</label></div>
            </div>
          </div>
        </form>
      </div>
      <div class="card-body p-0">
        <table class="table table-hover table-responsive mb-0">
          <thead class="table-light">
            <tr>
              <th v-for="f in fields" :key="f.key" :class="[f.sortable ? 'cursor-pointer' : '', f.tdClass]" @click="f.sortable && toggleSort(f.key)">{{ f.label }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in filtered" :key="r.name" @click="r._showDetails = !r._showDetails" class="cursor-pointer">
              <td>
                <div><span v-if="r.label">{{ r.label }}</span><span v-else class="text-secondary">{{ $t('labelMissing') }}</span>
                  <span v-if="r.security" class="badge rounded bg-primary m-1 py-1 px-2" @click.stop="r._showDetails = !r._showDetails">{{ $t('flags.security') }}</span>
                  <span v-if="r.triggers" class="badge rounded bg-primary m-1 py-1 px-2" @click.stop="r._showDetails = !r._showDetails">{{ $t('flags.triggers') }}</span>
                  <span v-if="r.iterator" class="badge rounded bg-primary m-1 py-1 px-2" @click.stop="r._showDetails = !r._showDetails">{{ $t('flags.iterator') }}</span>
                </div>
                <p v-if="r.description" class="text-secondary mb-0">{{ r.description }}</p>
                <div><small><code>{{ r.name }}</code></small></div>
                <div v-for="(error, i) in r.errors" :key="i" class="alert alert-warning py-1 my-1">{{ error }}</div>
                <div v-if="r._showDetails" class="card card-body mt-2 p-3"><pre>{{ r.triggers }}</pre><pre>{{ r.iterator }}</pre><pre>{{ r.security }}</pre></div>
              </td>
              <td class="text-end text-nowrap"><time :datetime="r.updatedAt?.toISOString()" :title="r.updatedAt">{{ filter.absoluteTime ? r.updatedAt : moment(r.updatedAt).fromNow() }}</time></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useListHelpers } from '../../../mixins/listHelpers'
import moment from 'moment'
const { t } = useI18n()
const lh = useListHelpers()
const { procListResults, encodeListParams } = lh
const id = 'automation'
const items = ref([])
const filter = reactive({ query: '', incScriptsWithErrors: false, incScriptsWithTriggers: false, incScriptsWithIterator: false, incScriptsWithSecurity: false, absoluteTime: false })
const sortKey = ref('name')
const sortDesc = ref(false)
const fields = computed(() => [{ key: 'name', sortable: true, tdClass: '' }, { key: 'updatedAt', sortable: true, tdClass: 'text-end text-nowrap' }].map(c => ({ label: t(`columns.${c.key}`), ...c })))
const filtered = computed(() => { const { query, incScriptsWithErrors, incScriptsWithTriggers, incScriptsWithIterator, incScriptsWithSecurity } = filter; const lcQuery = query.toLocaleLowerCase(); let f = items.value.filter(({ name, label }) => (lcQuery.length === 0 || (name + ' ' + label).toLocaleLowerCase().indexOf(lcQuery) > -1)).filter(({ errors }) => (incScriptsWithErrors === false || (errors && errors.length > 0))).filter(({ triggers }) => (incScriptsWithTriggers === false || !!triggers)).filter(({ iterator }) => (incScriptsWithIterator === false || !!iterator)).filter(({ security }) => (incScriptsWithSecurity === false || !!security)); const k = sortKey.value; if (k) { f.sort((a, b) => { const va = (a[k] || '').toString().toLowerCase(); const vb = (b[k] || '').toString().toLowerCase(); return sortDesc.value ? vb.localeCompare(va) : va.localeCompare(vb) }) } return f })
const totalScriptsWithErrors = computed(() => items.value.filter(({ errors }) => (errors && errors.length > 0)).length)
const totalScriptsWithSecurity = computed(() => items.value.filter(({ security }) => (security)).length)
const totalScriptsWithTriggers = computed(() => items.value.filter(({ triggers }) => (triggers)).length)
const totalScriptsWithIterator = computed(() => items.value.filter(({ iterator }) => (iterator)).length)
function toggleSort(key) { if (sortKey.value === key) sortDesc.value = !sortDesc.value; else { sortKey.value = key; sortDesc.value = false } }
onMounted(() => { procListResults(window.__SystemAPI.automationListCancellable(encodeListParams())).then(set => { items.value = set || [] }) })
</script>
<style>
.pointer, .cursor-pointer { cursor: pointer }
</style>
