<template>
  <div class="d-flex flex-column w-100 py-2 overflow-hidden h-100">
    <Teleport to="#topbar-title">
      {{ $t('title') }}
    </Teleport>

    <Teleport to="#topbar-tools">
      <div>
        <b-button
          v-if="canManage"
          data-test-id="button-manage-namespaces"
          variant="primary"
          size="sm"
          :to="{ name: 'namespace.manage' }"
        >
          {{ $t('manage-view.label') }}
          <font-awesome-icon
            :icon="['far', 'edit']"
            size="sm"
            class="ms-2"
          />
        </b-button>
      </div>

    </Teleport>

    <div class="d-flex flex-column justify-content-center align-items-center mx-4 mb-2">
      <div class="search w-100 mx-auto my-4">
        <c-input-search
          v-model.trim="query"
          :placeholder="$t('searchPlaceholder')"
          :debounce="200"
        />
      </div>
    </div>

    <div class="flex-fill overflow-auto">
      <div class="ns-wrapper h-100 container-fluid">
        <transition-group
          v-if="filtered && filtered.length"
          name="namespace-list"
          tag="div"
          class="row d-flex flex-wrap align-items-stretch justify-content-center mx-2"
        >
          <div
            v-for="n in filtered"
            :key="n.namespaceID"
            class="col-12 col-md-6 col-lg-4 col-xl-3 p-2"
          >
            <namespace-item :namespace="n" />
          </div>
        </transition-group>

        <div
          v-else
          class="d-flex justify-content-center align-items-center mt-5 w-100"
        >
          <h3 data-test-id="no-namespaces-found">
            {{ $t('noResults') }}
          </h3>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'
import { useNamespaceStore } from '../../store/namespace'
import { useRBACStore } from 'corteza-lib/vue/dist'
import NamespaceItem from 'corteza-webapp-compose/src/components/Namespaces/NamespaceItem'
import { components } from 'corteza-lib/vue/dist'
import {BButton} from "bootstrap-vue-next";
import {Portal} from "portal-vue";

const { CInputSearch } = components

const { t } = useI18n()
const { toastSuccess, toastErrorHandler } = composables.useToast()
const router = useRouter()
const nsStore = useNamespaceStore()
const rbac = useRBACStore()
const $ComposeAPI = window.__composeAPI

const query = ref('')

const namespaces = computed(() => nsStore.set)

const can = (resource, action) => rbac.can(resource, action)

const canManage = computed(() => {
  if (can('compose/', 'namespace.create') || can('compose/', 'grant')) {
    return true
  }
  return namespaces.value.reduce((acc, ns) => {
    return acc || ns.canUpdateNamespace || ns.canDeleteNamespace
  }, false)
})

const importNamespaceEndpoint = computed(() => $ComposeAPI.namespaceImportEndpoint({}))

const filtered = computed(() => {
  const q = query.value.toLowerCase()
  return namespaces.value
    .filter(({ enabled }) => enabled)
    .filter(({ slug, name }) => (slug + name).toLowerCase().indexOf(q) > -1)
})

onMounted(() => {
  document.title = t('label.app-name.namespace.list')
})

function onImported() {
  nsStore.load({ force: true })
    .then(() => toastSuccess(t('notification.namespace.imported')))
    .catch(toastErrorHandler(t('notification.namespace.importFailed')))
}

function onFailed(err) {
  toastErrorHandler(t('notification.namespace.importFailed'))(err)
}

function handleRowClicked({ namespace }) {
  router.push({
    name: 'namespace.edit',
    params: { namespaceID: namespace.namespaceID },
  })
}
</script>

<style lang="scss" scoped>
.search {
  max-width: 600px;
}
</style>
