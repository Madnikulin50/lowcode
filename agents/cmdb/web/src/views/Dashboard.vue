<template>
  <div class="container-fluid">
    <div class="row g-3">
      <div class="col-12">
        <div class="alert alert-info small py-1 mb-2">
          <strong>Debug:</strong> {{ debug || 'no debug info yet' }} |
          Devices: {{ devices ? devices.length : 0 }} |
          ModuleID: {{ moduleID }}
        </div>
      </div>
      <div class="col-md-5">
        <div class="card">
          <div class="card-header d-flex align-items-center gap-2">
            <i class="bi bi-wifi"></i>
            <strong>Network Scan</strong>
          </div>
          <div class="card-body">
            <form @submit.prevent="doScan">
              <div class="mb-3">
                <label class="form-label">Target (CIDR or IP)</label>
                <input v-model="cidr" class="form-control" placeholder="192.168.1.0/24" required />
              </div>
              <div class="mb-3">
                <label class="form-label">Namespace ID</label>
                <input v-model.number="namespaceID" type="number" class="form-control" placeholder="Leave empty for default" />
              </div>
              <div class="d-flex gap-2">
                <button type="submit" class="btn btn-primary" :disabled="scanning">
                  <span v-if="scanning" class="spinner-border spinner-border-sm me-1"></span>
                  <i v-else class="bi bi-search"></i>
                  {{ scanning ? 'Scanning...' : 'Start Scan' }}
                </button>
                <button type="button" class="btn btn-outline-secondary" @click="setupModule">
                  <i class="bi bi-gear"></i> Setup Module
                </button>
              </div>
            </form>
            <div v-if="moduleID" class="mt-2 text-success small">
              <i class="bi bi-check-circle"></i> Module ID: {{ moduleID }}
            </div>
          </div>
        </div>

        <div class="card mt-3">
          <div class="card-header d-flex align-items-center gap-2">
            <i class="bi bi-activity"></i>
            <strong>Active Scans</strong>
          </div>
          <div class="card-body p-0">
            <div v-if="scans.length === 0" class="p-3 text-muted text-center">No scans yet</div>
            <div v-for="s in (scans || [])" :key="s.id" class="border-bottom p-3">
              <div class="d-flex justify-content-between">
                <span class="fw-bold">{{ s.target }}</span>
                <span :class="statusBadge(s.status)" class="badge">{{ s.status }}</span>
              </div>
              <div v-if="s.status === 'running'" class="progress mt-2" style="height:6px">
                <div class="progress-bar progress-bar-striped progress-bar-animated" :style="{ width: s.progress + '%' }"></div>
              </div>
              <div class="small text-muted mt-1">
                <span v-if="s.status === 'running' && s.scanningIP">
                  Scanning: <code>{{ s.scanningIP }}</code> ({{ s.scannedIPs }}/{{ s.totalIPs }})
                </span>
                <span v-else>Found: {{ s.found }} devices</span>
                <span v-if="s.error" class="text-danger"> — {{ s.error }}</span>
              </div>
              <div class="small text-muted">{{ formatTime(s.startedAt) }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-md-7">
        <div class="card">
          <div class="card-header d-flex align-items-center gap-2">
            <i class="bi bi-hdd-stack"></i>
            <strong>Discovered Devices</strong>
            <button class="btn btn-sm btn-outline-primary ms-auto" @click="refreshDevices">
              <i class="bi bi-arrow-clockwise"></i> Refresh
            </button>
          </div>
          <div class="card-body p-0">
            <div v-if="loading" class="text-center p-4">
              <div class="spinner-border text-primary"></div>
            </div>
            <div v-else-if="devices.length === 0" class="p-3 text-muted text-center">
              No devices found. Run a scan first.
            </div>
            <div v-else class="table-responsive">
              <table class="table table-hover mb-0">
                <thead class="table-light">
                  <tr>
                    <th>IP</th>
                    <th>Hostname</th>
                    <th>Type</th>
                    <th>Vendor</th>
                    <th>Vulns</th>
                    <th>Status</th>
                    <th>Last Seen</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="d in devices" :key="d.recordID">
                    <td><code>{{ d.ip }}</code></td>
                    <td>{{ d.hostname || '—' }}</td>
                    <td><span class="badge bg-info">{{ d.deviceType || 'unknown' }}</span></td>
                    <td>{{ d.vendor || '—' }}</td>
                    <td>
                      <span v-if="d.vulnerabilities && d.vulnerabilities.length" class="badge"
                        :class="severityClass(d.vulnerabilities)"
                        style="cursor:pointer" @click="showVulns(d)">
                        {{ d.vulnerabilities.length }}
                      </span>
                      <span v-else class="text-muted small">0</span>
                    </td>
                    <td>
                      <span :class="d.status === 'online' ? 'text-success' : 'text-muted'">
                        <i :class="d.status === 'online' ? 'bi bi-circle-fill' : 'bi bi-circle'"></i>
                        {{ d.status }}
                      </span>
                    </td>
                    <td class="small">{{ d.lastSeen ? formatTime(d.lastSeen) : '—' }}</td>
                    <td>
                      <button class="btn btn-sm btn-outline-danger" @click="removeDevice(d.recordID)">
                        <i class="bi bi-trash"></i>
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
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
import { ref, onMounted, onUnmounted } from 'vue'
import { startScan, listScans, listDevices, deleteDevice, ensureModule } from '../api.js'

const cidr = ref('')
const namespaceID = ref(0)
const scanning = ref(false)
const scans = ref([])
const devices = ref([])
const loading = ref(false)
const moduleID = ref(0)
const debug = ref('')
const vulnDevice = ref(null)
let pollTimer = null

function showVulns(d) {
  vulnDevice.value = d
}

function statusBadge(status) {
  return {
    running: 'bg-primary',
    done: 'bg-success',
    error: 'bg-danger',
  }[status] || 'bg-secondary'
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleString()
}

function severityClass(vulns) {
  if (!vulns || !vulns.length) return ''
  const hasCritical = vulns.some(v => v.severity === 'CRITICAL')
  const hasHigh = vulns.some(v => v.severity === 'HIGH')
  if (hasCritical) return 'bg-danger'
  if (hasHigh) return 'bg-warning text-dark'
  return 'bg-secondary'
}

async function doScan() {
  scanning.value = true
  try {
    const s = await startScan(cidr.value, namespaceID.value || undefined)
    scans.value.unshift(s)
    cidr.value = ''
  } catch (e) {
    alert('Scan failed: ' + e.message)
  } finally {
    scanning.value = false
  }
}

async function checkDB() {
  try {
    const resp = await fetch('/api/debug/db')
    const data = await resp.json()
    debug.value = `DB check: count=${data.count}`
  } catch (e) {
    debug.value = `DB check error: ${e.message}`
  }
}

async function refreshDevices() {
  loading.value = true
  try {
    devices.value = await listDevices(moduleID.value || undefined)
    debug.value = `devices count: ${devices.value.length}, moduleID: ${moduleID.value}`
  } catch (e) {
    console.error(e)
    debug.value = `error: ${e.message}`
  } finally {
    loading.value = false
  }
}

async function removeDevice(recordID) {
  if (!confirm('Delete this device?')) return
  try {
    await deleteDevice(recordID, moduleID.value || undefined)
    devices.value = devices.value.filter(d => d.recordID !== recordID)
  } catch (e) {
    alert('Delete failed: ' + e.message)
  }
}

async function setupModule() {
  try {
    const r = await ensureModule()
    moduleID.value = r.moduleID
    alert('Module ready! ID: ' + r.moduleID)
  } catch (e) {
    alert('Setup failed: ' + e.message)
  }
}

async function pollScans() {
  try {
    scans.value = await listScans()
    for (const s of scans.value) {
      if (s.moduleID) {
        moduleID.value = s.moduleID
      }
    }
    refreshDevices()
    checkDB()
  } catch (e) {
    // ignore poll errors
  }
}

onMounted(() => {
  pollScans()
  refreshDevices()
  pollTimer = setInterval(pollScans, 2000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>
