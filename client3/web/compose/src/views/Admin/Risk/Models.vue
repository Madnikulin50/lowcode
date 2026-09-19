<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      Модели риска
    </Teleport>

    <div class="d-flex justify-content-between align-items-center mb-3">
      <p class="text-secondary mb-0">
        Граф факторов и стратегия расчёта для одного класса объектов.
      </p>
      <router-link
        :to="{ name: 'admin.risk.models.create' }"
        class="btn btn-primary"
      >
        <font-awesome-icon :icon="['fas', 'plus']" />
        <span class="ms-1">Новая модель</span>
      </router-link>
    </div>

    <div class="card shadow-sm">
      <table class="table mb-0 align-middle">
        <thead>
          <tr>
            <th>Название</th>
            <th>Стратегия</th>
            <th>Статус</th>
            <th class="text-end">
              Факторов
            </th>
            <th class="text-end">
              Связей
            </th>
            <th class="text-end">
              &nbsp;
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="m in models"
            :key="m.modelID"
            role="button"
            @click="edit(m)"
          >
            <td>
              <font-awesome-icon
                :icon="['fas', 'shield-halved']"
                class="me-2 text-secondary"
              />
              {{ m.name }}
            </td>
            <td><span class="badge bg-info-subtle text-info-emphasis">{{ m.strategy }}</span></td>
            <td>
              <span
                class="badge"
                :class="m.status === 'published' ? 'bg-success' : 'bg-secondary'"
              >
                {{ m.status }} · v{{ m.version }}
              </span>
            </td>
            <td class="text-end">
              {{ m.nodeCount }}
            </td>
            <td class="text-end">
              {{ m.edgeCount }}
            </td>
            <td
              class="text-end"
              @click.stop
            >
              <button
                class="btn btn-sm btn-outline-danger"
                @click="remove(m)"
              >
                Удалить
              </button>
            </td>
          </tr>
          <tr v-if="!loading && models.length === 0">
            <td
              colspan="6"
              class="text-center text-secondary py-4"
            >
              Пока нет ни одной модели риска.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { composables } from 'corteza-lib/vue/dist'

const { useToast } = composables
const { toastSuccess, toastErrorHandler } = useToast()
const router = useRouter()
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const models = ref([])
const loading = ref(false)

function load () {
  loading.value = true
  return $ComposeAPI.riskModelList({ namespaceID: props.namespace?.namespaceID })
    .then(({ models: set }) => { models.value = set || [] })
    .catch(toastErrorHandler('Не удалось загрузить модели риска'))
    .finally(() => { loading.value = false })
}

function edit (m) {
  router.push({ name: 'admin.risk.models.edit', params: { modelID: m.modelID } })
}

function remove (m) {
  // eslint-disable-next-line no-alert
  if (!window.confirm(`Удалить модель «${m.name}»?`)) return
  $ComposeAPI.riskModelDelete({ modelID: m.modelID })
    .then(() => {
      toastSuccess('Модель удалена')
      return load()
    })
    .catch(toastErrorHandler('Не удалось удалить модель'))
}

onMounted(load)
</script>
