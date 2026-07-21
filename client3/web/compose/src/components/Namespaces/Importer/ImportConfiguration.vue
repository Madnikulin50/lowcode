<template>
  <div class="card">
    <div class="card-body">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('name.label') }}</label>
        <input
          v-model="name"
          data-test-id="input-name"
          class="form-control mt-1"
          :placeholder="$t('name.placeholder')"
        />
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('import.slug.label') }}</label>
        <input
          v-model="slug"
          data-test-id="input-handle"
          class="form-control mt-1"
          :class="{ 'is-invalid': slugState === false }"
          :placeholder="$t('slug.placeholder')"
        />
        <div
          v-if="slugState === false"
          class="invalid-feedback d-block"
        >
          {{ $t('slug.invalid-handle-characters') }}
        </div>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('import.connection.label') }}</label>
        <c-input-select
          v-model="connectionID"
          :options="connections"
          :clearable="false"
          :reduce="o => o.connectionID"
          :placeholder="$t('import.connection.placeholder')"
          :get-option-label="({ handle, meta }) => meta.name || handle"
          :get-option-key="getOptionKey"
        />
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('import.importData.label') }}</label>
        <c-input-checkbox
          v-model="importData"
          switch
        />
      </div>
    </div>
    <div class="card-footer border-top d-flex justify-content-between align-items-center">
      <button
        data-test-id="button-back"
        class="btn btn-link d-flex align-items-center text-dark back gap-1 text-decoration-none"
        @click="$emit('back')"
      >
        <font-awesome-icon
          :icon="['fas', 'chevron-left']"
          class="back-icon"
        />
        {{ $t('import.back') }}
      </button>

      <button
        data-test-id="button-import"
        class="btn btn-primary"
        :disabled="submitDisabled"
        @click="nextStep"
      >
        {{ $t('import.import') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { handle } from 'corteza-lib/vue/dist'

const emit = defineEmits(['back', 'configured'])

const props = defineProps({
  session: { type: Object, required: true, default: () => ({}) },
})

const $SystemAPI = inject('$SystemAPI')

const processing = ref({ connections: true, sensitiveData: true })
const name = ref('')
const slug = ref('')
const connectionID = ref(null)
const importData = ref(true)
const connections = ref([])

const submitDisabled = computed(() => [nameState.value, slugState.value, slug.value].includes(false))
const nameState = computed(() => name.value.length > 0 ? null : false)
const slugState = computed(() => handle.handleState(slug.value))

onMounted(() => {
  fetchConnections()
})

function fetchConnections () {
  processing.value.connections = true
  $SystemAPI.dataPrivacyConnectionList()
    .then(({ set = [] }) => {
      connections.value = set
      const { connectionID: cid } = set[0] || {}
      connectionID.value = cid
    })
    .catch((e) => {
      console.error(e)
    })
    .finally(() => {
      processing.value.connections = false
    })
}

function getOptionKey ({ connectionID: cid }) {
  return cid
}

function nextStep () {
  const rtr = {
    name: name.value,
    slug: slug.value,
    connectionID: connectionID.value,
    importData: importData.value,
  }
  emit('configured', rtr)
}
</script>
