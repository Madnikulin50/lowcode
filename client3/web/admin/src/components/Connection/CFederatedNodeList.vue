<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <c-resource-list
      primary-key="federationID"
      :loading-text="$t('loading')"
      :pagination="pagination"
      :sorting="sorting"
      :items="items"
      :fields="fields"
    >
      <template #actions>
        <router-link
          class="btn btn-primary btn-lg"
          :to="{ name: 'system.connections.new' }"
        >
          {{ $t('add-button') }}
        </router-link>
      </template>

      <template #filter>
        <div class="input-group h-100">
          <input
            v-model.trim="filter.query"
            type="text"
            class="form-control text-truncate border-end-0 h-100"
            :placeholder="$t('query.placeholder')"
          >
          <span class="input-group-text text-primary bg-white">
            <font-awesome-icon :icon="['fas', 'search']" />
          </span>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.connections', keyPrefix: 'federation' } })
import { reactive } from 'vue'
import { useListHelpers } from 'corteza-webapp-admin/src/mixins/listHelpers'
import { useI18n } from 'vue-i18n'
import moment from 'moment'
import { fmt } from 'corteza-lib/js/dist'

const { t } = useI18n()

const {
  pagination,
  procListResults,
} = useListHelpers()

const sorting = reactive({
  sortBy: 'createdAt',
  sortDesc: true,
})

const filter = reactive({
  query: '',
})

const fields = [
  { key: 'name', sortable: true },
  { key: 'url', sortable: true },
  { key: 'location', sortable: true },
  { key: 'ownership', sortable: true },
  { key: 'createdBy', sortable: true },
  { key: 'createdAt', sortable: true, formatter: (v) => v ? fmt.fullDateTime(v) : v },
  { key: 'actions', class: 'text-end' },
].map(c => ({
  ...c,
  label: c.label || t(`columns.${c.key}`),
}))

function items() {
  const set = [
    { federationID: '1', name: 'ACME France', url: 'https://corteza.acme.fr', location: 'France', ownership: 'ACME SARL', createdBy: 'John Doe', createdAt: new Date() },
  ]

  const filterData = {
    count: set.length,
    limit: 10,
  }

  return procListResults(new Promise(resolve => setTimeout(() => resolve({ filter: filterData, set }), 200)), false)
}
</script>