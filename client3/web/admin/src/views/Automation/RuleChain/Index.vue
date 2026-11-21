<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <div class="d-flex align-items-center justify-content-between mb-3">
      <h4 class="mb-0">{{ $t('ruleChain.title') }}</h4>
      <div class="d-flex gap-2">
        <button class="btn btn-sm btn-outline-secondary" @click="importChain">
          <font-awesome-icon :icon="['fas', 'file-import']" class="me-1" />
          {{ $t('ruleChain.import') }}
        </button>
        <button class="btn btn-sm btn-primary" @click="newChain">
          <font-awesome-icon :icon="['fas', 'plus']" class="me-1" />
          {{ $t('ruleChain.new') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status" />
    </div>

    <div v-else-if="chains.length === 0" class="text-center py-5 text-muted">
      <h5>{{ $t('ruleChain.empty') }}</h5>
      <p>{{ $t('ruleChain.emptyHint') }}</p>
      <button class="btn btn-primary" @click="newChain">
        {{ $t('ruleChain.createFirst') }}
      </button>
    </div>

    <div v-else class="row g-3">
      <div v-for="chain in chains" :key="chain.id" class="col-12">
        <div class="card shadow-sm">
          <div class="card-header d-flex align-items-center justify-content-between py-2">
            <div class="d-flex align-items-center gap-2">
              <font-awesome-icon :icon="['fas', 'project-diagram']" class="text-primary" />
              <strong>{{ chain.name }}</strong>
              <span class="badge bg-secondary">{{ chain.nodeCount }} nodes</span>
            </div>
            <div class="d-flex gap-1">
              <button class="btn btn-sm btn-outline-primary" @click="testChain(chain)" :disabled="chain._testing">
                <span v-if="chain._testing" class="spinner-border spinner-border-sm me-1" />
                <font-awesome-icon :icon="['fas', 'play']" />
              </button>
              <button class="btn btn-sm btn-outline-secondary" @click="exportChain(chain)">
                <font-awesome-icon :icon="['fas', 'file-export']" />
              </button>
              <button class="btn btn-sm btn-outline-danger" @click="deleteChain(chain)">
                <font-awesome-icon :icon="['fas', 'trash']" />
              </button>
            </div>
          </div>

          <div v-if="chain._result" class="card-body py-2" :class="chain._result.success ? 'bg-success-subtle' : 'bg-danger-subtle'">
            <div class="d-flex align-items-center gap-2 small">
              <font-awesome-icon :icon="['fas', chain._result.success ? 'check-circle' : 'exclamation-circle']"
                :class="chain._result.success ? 'text-success' : 'text-danger'" />
              <span v-if="chain._result.success">{{ chain._result.nodes?.length || 0 }} nodes executed</span>
              <span v-else class="text-danger">{{ chain._result.error }}</span>
            </div>
            <div v-if="chain._result.output" class="mt-2">
              <pre class="mb-0 small p-2 bg-light rounded">{{ chain._result.output }}</pre>
            </div>
          </div>

          <div class="card-body py-2">
            <div class="row small text-muted">
              <div class="col-auto">
                <strong>{{ $t('ruleChain.id') }}:</strong> {{ chain.id }}
              </div>
              <div class="col-auto">
                <strong>{{ $t('ruleChain.entry') }}:</strong> {{ chain.entryNode }}
              </div>
              <div class="col-auto" v-if="chain.description">
                {{ chain.description }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Test modal -->
    <div v-if="testChainData" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,.5)">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('ruleChain.test') }}: {{ testChainData.name }}</h5>
            <button class="btn-close" @click="testChainData = null" />
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">{{ $t('ruleChain.testInput') }}</label>
              <textarea v-model="testInput" class="form-control" rows="6" placeholder='{"key": "value"}' />
            </div>
            <div v-if="testResult" class="mt-3">
              <h6>{{ $t('ruleChain.testResult') }}</h6>
              <pre class="p-2 bg-light rounded small" style="max-height: 300px; overflow: auto">{{ JSON.stringify(testResult, null, 2) }}</pre>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="testChainData = null">
              {{ $t('ruleChain.close') }}
            </button>
            <button class="btn btn-primary" @click="runTest" :disabled="testRunning">
              <span v-if="testRunning" class="spinner-border spinner-border-sm me-1" />
              {{ $t('ruleChain.run') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- New chain modal -->
    <div v-if="showNewModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,.5)">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('ruleChain.newChain') }}</h5>
            <button class="btn-close" @click="showNewModal = false" />
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">{{ $t('ruleChain.name') }}</label>
              <input v-model="newChainData.name" class="form-control" />
            </div>
            <div class="mb-3">
              <label class="form-label">{{ $t('ruleChain.description') }}</label>
              <textarea v-model="newChainData.description" class="form-control" rows="2" />
            </div>
            <div class="mb-3">
              <label class="form-label">{{ $t('ruleChain.nodes') }} (JSON)</label>
              <textarea v-model="newChainData.nodes" class="form-control" rows="8" placeholder='[{"id":"n1","type":"condition","label":"Check","config":{"field":"name","operator":"notEmpty"}}]' />
            </div>
            <div class="mb-3">
              <label class="form-label">{{ $t('ruleChain.edges') }} (JSON)</label>
              <textarea v-model="newChainData.edges" class="form-control" rows="3" placeholder='[]' />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showNewModal = false">{{ $t('ruleChain.cancel') }}</button>
            <button class="btn btn-primary" @click="createChain">{{ $t('ruleChain.create') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Import modal -->
    <div v-if="showImportModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,.5)">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('ruleChain.importChain') }}</h5>
            <button class="btn-close" @click="showImportModal = false" />
          </div>
          <div class="modal-body">
            <textarea v-model="importData" class="form-control" rows="15" placeholder='{"id":"...","name":"...","nodes":[...],"edges":[...]}' />
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showImportModal = false">{{ $t('ruleChain.cancel') }}</button>
            <button class="btn btn-primary" @click="doImport">{{ $t('ruleChain.import') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const loading = ref(false)
const chains = ref([])
const testChainData = ref(null)
const testInput = ref('{}')
const testResult = ref(null)
const testRunning = ref(false)
const showNewModal = ref(false)
const showImportModal = ref(false)
const importData = ref('')
const newChainData = ref({ name: '', description: '', nodes: '[]', edges: '[]' })

onMounted(() => {
  fetchChains()
})

async function fetchChains() {
  loading.value = true
  try {
    const resp = await window.__systemAPI.api().request({
      method: 'get',
      url: window.__systemAPI.baseURL.replace('system', 'compose') + '/admin/rulechain/',
    })
    chains.value = (resp?.data?.response?.chains || resp?.data?.chains || []).map(c => ({
      ...c,
      _testing: false,
      _result: null,
    }))
  } catch (err) {
    chains.value = []
  } finally {
    loading.value = false
  }
}

function testChain(chain) {
  testChainData.value = chain
  testInput.value = '{}'
  testResult.value = null
}

async function runTest() {
  if (!testChainData.value) return
  testRunning.value = true
  let input = {}
  try { input = JSON.parse(testInput.value) } catch (e) { /* keep as string */ }

  try {
    const resp = await window.__systemAPI.api().request({
      method: 'post',
      url: window.__systemAPI.baseURL.replace('system', 'compose') + '/admin/rulechain/' + testChainData.value.id + '/test',
      data: input,
    })
    testResult.value = resp?.data?.response || resp?.data || {}
  } catch (err) {
    testResult.value = { error: err.message }
  } finally {
    testRunning.value = false
  }
}

function exportChain(chain) {
  const data = JSON.stringify(chain, null, 2)
  const blob = new Blob([data], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = Object.assign(document.createElement('a'), { href: url, download: `${chain.id}.json` })
  a.click()
  URL.revokeObjectURL(url)
}

async function deleteChain(chain) {
  if (!confirm('Delete chain "' + chain.name + '"?')) return
  try {
    await window.__systemAPI.api().request({
      method: 'delete',
      url: window.__systemAPI.baseURL.replace('system', 'compose') + '/admin/rulechain/' + chain.id,
    })
    await fetchChains()
  } catch (err) {
    alert(err.message || 'Delete failed')
  }
}

function newChain() {
  newChainData.value = { name: '', description: '', nodes: '[]', edges: '[]' }
  showNewModal.value = true
}

async function createChain() {
  let nodes = []
  let edges = []
  try { nodes = JSON.parse(newChainData.value.nodes) } catch (e) {
    alert('Invalid nodes JSON')
    return
  }
  try { edges = JSON.parse(newChainData.value.edges) } catch (e) {
    alert('Invalid edges JSON')
    return
  }

  try {
    await window.__systemAPI.api().request({
      method: 'post',
      url: window.__systemAPI.baseURL.replace('system', 'compose') + '/admin/rulechain/',
      data: {
        name: newChainData.value.name,
        description: newChainData.value.description,
        nodes,
        edges,
      },
    })
    showNewModal.value = false
    await fetchChains()
  } catch (err) {
    alert(err.message || 'Create failed')
  }
}

function importChain() {
  importData.value = ''
  showImportModal.value = true
}

async function doImport() {
  try {
    JSON.parse(importData.value)
  } catch (e) {
    alert('Invalid JSON')
    return
  }

  try {
    await window.__systemAPI.api().request({
      method: 'post',
      url: window.__systemAPI.baseURL.replace('system', 'compose') + '/rulechain/import',
      data: JSON.parse(importData.value),
    })
    showImportModal.value = false
    await fetchChains()
  } catch (err) {
    alert(err.message || 'Import failed')
  }
}
</script>
