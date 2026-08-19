<template>
  <div class="container-fluid">
    <div class="card">
      <div class="card-header d-flex align-items-center gap-2">
        <i class="bi bi-devices"></i>
        <strong>All Devices</strong>
        <button class="btn btn-sm btn-outline-primary ms-auto" @click="load">
          <i class="bi bi-arrow-clockwise"></i> Refresh
        </button>
        <input v-model="filter" class="form-control form-control-sm ms-2" style="max-width:200px" placeholder="Filter..." />
      </div>
      <div class="card-body p-0">
        <div v-if="loading" class="text-center p-4">
          <div class="spinner-border text-primary"></div>
        </div>
        <div v-else-if="filtered.length === 0" class="p-3 text-muted text-center">No devices found.</div>
        <div v-else class="table-responsive">
          <table class="table table-hover mb-0">
            <thead class="table-light">
              <tr>
                <th>IP</th>
                <th>MAC</th>
                <th>Hostname</th>
                <th>Type</th>
                <th>Vendor</th>
                <th>OS</th>
                <th>Ports</th>
                <th>Vulns</th>
                <th>Status</th>
                <th>Last Seen</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="d in filtered" :key="d.recordID">
                <td><code>{{ d.ip }}</code></td>
                <td class="small">{{ d.mac || '—' }}</td>
                <td>{{ d.hostname || '—' }}</td>
                <td><span class="badge bg-info">{{ d.deviceType || 'unknown' }}</span></td>
                <td>{{ d.vendor || '—' }}</td>
                <td class="small">{{ d.os || '—' }}</td>
                <td>
                  <span v-if="d.openPorts && d.openPorts.length" class="small">
                    {{ d.openPorts.length }} open
                  </span>
                  <span v-else class="text-muted">—</span>
                </td>
                <td>
                  <span v-if="d.vulnerabilities && d.vulnerabilities.length" class="badge"
                    :class="d.vulnerabilities.some(v => v.severity === 'CRITICAL') ? 'bg-danger' : d.vulnerabilities.some(v => v.severity === 'HIGH') ? 'bg-warning text-dark' : 'bg-secondary'"
                    style="cursor:pointer" @click="showVulns(d)">
                    {{ d.vulnerabilities.length }}
                  </span>
                  <span v-else class="text-muted">0</span>
                </td>
                <td>
                  <span :class="d.status === 'online' ? 'text-success' : 'text-muted'">
                    <i :class="d.status === 'online' ? 'bi bi-circle-fill' : 'bi bi-circle'"></i>
                    {{ d.status }}
                  </span>
                </td>
                <td class="small">{{ d.lastSeen ? formatTime(d.lastSeen) : '—' }}</td>
                <td>
                  <button class="btn btn-sm btn-outline-info" @click="showPorts(d)" title="View ports">
                    <i class="bi bi-plug"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-danger ms-1" @click="remove(d.recordID)">
                    <i class="bi bi-trash"></i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div v-if="selected" class="modal d-block" tabindex="-1" @click.self="selected=null">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ selected.ip }} — Ports</h5>
            <button type="button" class="btn-close" @click="selected=null"></button>
          </div>
          <div class="modal-body">
            <div v-if="selected.vulnerabilities && selected.vulnerabilities.length" class="mb-3">
              <h6>Vulnerabilities ({{ selected.vulnerabilities.length }})</h6>
              <div v-for="v in selected.vulnerabilities" :key="v.name" class="d-flex align-items-start gap-2 mb-1 small">
                <span class="badge rounded-pill"
                  :class="v.severity === 'CRITICAL' ? 'bg-danger' : v.severity === 'HIGH' ? 'bg-warning text-dark' : 'bg-secondary'">
                  {{ v.severity }}
                </span>
                <div>
                  <strong>{{ v.name }}</strong>
                  <span v-if="v.cve" class="ms-1 text-muted">({{ v.cve }})</span>
                  <p class="mb-0 text-muted">{{ v.description }}</p>
                </div>
              </div>
            </div>
            <div v-if="selected.shares && selected.shares.length" class="mb-3">
              <h6>Shared Folders</h6>
              <ul class="list-group list-group-flush small">
                <li v-for="s in selected.shares" :key="s" class="list-group-item py-1">{{ s }}</li>
              </ul>
            </div>
            <div v-if="!selected.openPorts || selected.openPorts.length === 0" class="text-muted">No open ports detected</div>
            <table v-else class="table table-sm">
              <thead>
                <tr><th>Port</th><th>Proto</th><th>Service</th><th>Version</th></tr>
              </thead>
              <tbody>
                <tr v-for="p in selected.openPorts" :key="p.port">
                  <td>{{ p.port }}</td>
                  <td>{{ p.proto }}</td>
                  <td>{{ p.service || '—' }}</td>
                  <td class="text-muted">{{ p.version || '—' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <div v-if="vulnDevice" class="modal d-block" tabindex="-1" @click.self="vulnDevice=null">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ vulnDevice.ip }} — Vulnerabilities</h5>
            <button type="button" class="btn-close" @click="vulnDevice=null"></button>
          </div>
          <div class="modal-body">
            <div v-if="!vulnDevice.vulnerabilities || vulnDevice.vulnerabilities.length === 0" class="text-muted">
              No vulnerabilities found
            </div>
            <div v-else v-for="v in vulnDevice.vulnerabilities" :key="v.name" class="mb-2 pb-2 border-bottom">
              <div class="d-flex align-items-start gap-2">
                <span class="badge rounded-pill"
                  :class="v.severity === 'CRITICAL' ? 'bg-danger' : v.severity === 'HIGH' ? 'bg-warning text-dark' : 'bg-secondary'">
                  {{ v.severity }}
                </span>
                <div>
                  <strong>{{ v.name }}</strong>
                  <span v-if="v.cve" class="ms-1 text-muted">({{ v.cve }})</span>
                  <p class="mb-0 small text-muted mt-1">{{ v.description }}</p>
                  <p v-if="v.remediation" class="mb-0 small text-success mt-1">
                    <i class="bi bi-tools"></i> {{ v.remediation }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { listDevices, deleteDevice } from '../api.js'

const devices = ref([])
const loading = ref(false)
const filter = ref('')
const selected = ref(null)
const vulnDevice = ref(null)

function showVulns(d) {
  vulnDevice.value = d
}

const filtered = computed(() => {
  if (!filter.value) return devices.value
  const q = filter.value.toLowerCase()
  return devices.value.filter(d =>
    (d.ip && d.ip.includes(q)) ||
    (d.hostname && d.hostname.toLowerCase().includes(q)) ||
    (d.vendor && d.vendor.toLowerCase().includes(q)) ||
    (d.deviceType && d.deviceType.includes(q)) ||
    (d.mac && d.mac.includes(q))
  )
})

function formatTime(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleString()
}

function showPorts(d) {
  selected.value = d
}

async function remove(recordID) {
  if (!confirm('Delete this device?')) return
  try {
    await deleteDevice(recordID)
    devices.value = devices.value.filter(d => d.recordID !== recordID)
  } catch (e) {
    alert('Delete failed: ' + e.message)
  }
}

async function load() {
  loading.value = true
  try {
    devices.value = await listDevices()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
