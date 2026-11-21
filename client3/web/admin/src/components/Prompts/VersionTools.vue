<template>
  <div
    v-if="history.length"
    class="card mb-3"
  >
    <div class="card-body">
      <div class="d-flex flex-wrap align-items-center gap-2 mb-2">
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary"
          :disabled="!previousVersion || loading"
          data-test-id="button-compare-versions"
          @click="toggleCompare"
        >
          {{ comparing ? $t(`${ns}.versionTools.hideCompare`) : $t(`${ns}.versionTools.compare`, { version: viewing }) }}
        </button>

        <span class="flex-grow-1" />

        <label class="form-label mb-0 small">{{ $t(`${ns}.versionTools.keep`) }}</label>
        <input
          v-model.number="keep"
          type="number"
          min="1"
          class="form-control form-control-sm"
          style="width: 5rem;"
        >
        <button
          type="button"
          class="btn btn-sm btn-outline-danger"
          :disabled="pruning || history.length <= keep"
          data-test-id="button-prune-versions"
          @click="prune"
        >
          {{ $t(`${ns}.versionTools.prune`) }}
        </button>
      </div>

      <div
        v-if="message"
        class="small text-muted"
      >
        {{ message }}
      </div>

      <template v-if="comparing">
        <p
          v-if="!lines.length || !changed"
          class="text-muted mb-0"
        >
          {{ $t(`${ns}.versionTools.noChanges`, { version: viewing, previous: previousVersion }) }}
        </p>
        <pre
          v-else
          class="diff mb-0"
        ><span
          v-for="(l, i) in lines"
          :key="i"
          :class="'diff-' + l.type"
        >{{ prefix[l.type] }}{{ l.text }}</span></pre>
      </template>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { lineDiff, hasChanges } from '../../lib/line-diff'

const props = defineProps({
  // 'automation.prompts' or 'automation.skills' (locale namespace)
  ns: { type: String, required: true },
  kind: { type: String, required: true, validator: v => ['prompt', 'skill'].includes(v) },
  handle: { type: String, required: true },
  history: { type: Array, default: () => [] },
  viewing: { type: Number, default: 0 },
})

const emit = defineEmits(['pruned'])
const { t: $t } = useI18n()

const prefix = { same: '  ', add: '+ ', del: '− ' }
const keep = ref(10)
const comparing = ref(false)
const loading = ref(false)
const pruning = ref(false)
const lines = ref([])
const message = ref('')

const api = () => window.__AutomationAPI

// the version right before the one on screen
const previousVersion = computed(() => {
  const older = props.history.map(v => v.version).filter(v => v < props.viewing)
  return older.length ? Math.max(...older) : 0
})

const changed = computed(() => hasChanges(lines.value))

function textOf (res) {
  const v = res[props.kind] || {}
  return v.text || ''
}

async function loadVersion (version) {
  const get = props.kind === 'skill' ? api().skillVersion : api().promptVersion
  return textOf(await get.call(api(), { handle: props.handle, version }))
}

async function compare () {
  loading.value = true
  message.value = ''
  try {
    const [before, after] = await Promise.all([loadVersion(previousVersion.value), loadVersion(props.viewing)])
    lines.value = lineDiff(before, after)
  } catch (e) {
    message.value = (e && e.message) || String(e)
  } finally {
    loading.value = false
  }
}

async function toggleCompare () {
  comparing.value = !comparing.value
  if (comparing.value) await compare()
}

// follow the version being viewed while the comparison is open
watch(() => props.viewing, () => {
  if (!previousVersion.value) comparing.value = false
  else if (comparing.value) compare()
})

async function prune () {
  if (!window.confirm($t(`${props.ns}.versionTools.pruneConfirm`, { keep: keep.value }))) return

  pruning.value = true
  message.value = ''
  try {
    const call = props.kind === 'skill' ? api().skillPrune : api().promptPrune
    const { removed } = await call.call(api(), { handle: props.handle, keep: keep.value })
    message.value = $t(`${props.ns}.versionTools.pruned`, { count: removed })
    emit('pruned')
  } catch (e) {
    message.value = (e && e.response && e.response.data && e.response.data.error && e.response.data.error.message) || (e && e.message) || String(e)
  } finally {
    pruning.value = false
  }
}
</script>

<style scoped>
.diff {
  max-height: 24rem;
  overflow: auto;
  font-size: 0.85rem;
}
.diff span { display: block; white-space: pre-wrap; }
.diff-add { background: rgba(25, 135, 84, 0.15); }
.diff-del { background: rgba(220, 53, 69, 0.15); }
</style>
