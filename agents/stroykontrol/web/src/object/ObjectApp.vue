<script setup>
import { errText } from '../utils.js'
// «Обзор» объекта: какие документы есть (комплекты ПД и РД по версиям), насколько
// комплект полон (ведомость томов) и какая версия комплекта с какой сравнивается.
import { reactive, computed, onMounted } from 'vue'

const params = new URLSearchParams(location.search)
const objectID = params.get('objectID') || ''
const namespaceID = params.get('namespaceID') || ''
const qs = `objectID=${encodeURIComponent(objectID)}&namespaceID=${encodeURIComponent(namespaceID)}`

const state = reactive({ loading: true, busy: false, error: '', notice: '', data: null })
const picked = reactive({}) // proposed row ID → true (taken)

async function api (url, body) {
  const res = await fetch(url, body ? { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) } : undefined)
  const data = await res.json().catch(() => ({}))
  if (!res.ok || data.error) throw new Error(errText(data.error, res.status) || res.statusText)
  return data
}

async function load () {
  try {
    state.data = await api(`api/sets/overview?${qs}`)
    state.error = ''
  } catch (e) {
    state.error = 'Не удалось загрузить документы объекта: ' + e.message
  } finally {
    state.loading = false
  }
}
onMounted(load)

async function check () {
  state.busy = true
  state.notice = ''
  try {
    state.data = await api(`api/sets/check?${qs}`, {})
    state.notice = 'Комплектность пересчитана.'
  } catch (e) {
    state.error = e.message
  } finally {
    state.busy = false
  }
}

const UNIT = { section: 'раздел', appendix: 'приложение', reference: 'ссылочный' }
const UPLOAD = { new: 'новый том', edition: 'новая редакция', repeat: 'повтор файла', unparsed: 'шифр не прочитан' }
const SOURCE_LABEL = { contents: 'из содержания', set: 'из первого комплекта', manual: 'вручную' }
const proposals = computed(() => (state.data && state.data.proposals) || [])
const pickedIDs = computed(() => proposals.value.filter(p => picked[p.id]).map(p => p.id))

async function suggest () {
  state.busy = true
  state.notice = ''
  state.error = ''
  try {
    const r = await api(`api/volumes/suggest?${qs}`, {})
    state.data = r.overview
    proposals.value.forEach(p => { picked[p.id] = true })
    state.notice = r.created ? `Предложено строк ведомости: ${r.created}. Отметьте нужные и примите.` : 'Новых строк предложить нечего: загрузите комплект с содержанием или выпустите первый комплект стадии.'
  } catch (e) {
    state.error = e.message
  } finally {
    state.busy = false
  }
}

async function resolve (accept, reject, all = false) {
  state.busy = true
  state.error = ''
  try {
    const r = await api('api/volumes/resolve', { objectID, namespaceID, accept, reject, acceptAll: all })
    state.data = r.overview
    state.notice = `Принято: ${r.accepted}, отклонено: ${r.rejected}. Комплектность пересчитана.`
  } catch (e) {
    state.error = e.message
  } finally {
    state.busy = false
  }
}
const acceptPicked = () => resolve(pickedIDs.value, [])
const rejectPicked = () => resolve([], pickedIDs.value)

const STATUS = { draft: 'черновик', accepted: 'принят', superseded: 'заменён новой версией' }
const CELL = { present: 'есть', missing: 'нет', extra: 'лишний', no_set: '—', unparsed: 'не прочитан' }
const COMP_STATUS = { new: 'новое', processing: 'в обработке', done: 'готово', failed: 'ошибка' }

const stages = computed(() => {
  const d = state.data
  if (!d) return []
  return [
    { key: 'ref', name: 'ПД', sets: d.sets.filter(s => s.stage === 'ref'), expected: d.expected.ref },
    { key: 'target', name: 'РД', sets: d.sets.filter(s => s.stage === 'target'), expected: d.expected.target },
  ]
})

const hasStatement = computed(() => !!state.data && (state.data.expected.ref + state.data.expected.target) > 0)
const hasAny = computed(() => state.data && (state.data.sets.length || state.data.pairs.length))
const barColor = (s) => s.total === 0 ? 'var(--muted)' : s.pct >= 100 ? 'var(--ok)' : s.pct >= 60 ? 'var(--warn)' : 'var(--bad)'
// Compose UI base ("…/ns/<slug>") taken from the page this view is embedded in
function composeBase () {
  const hrefs = []
  try { hrefs.push(window.top.location.href) } catch (e) { /* cross-origin */ }
  if (document.referrer) hrefs.push(document.referrer)
  for (const h of hrefs) {
    const m = h.match(/^(.*?\/ns\/[^/?#]+)/)
    if (m) return m[1]
  }
  return ''
}
const base = composeBase()
const pages = computed(() => (state.data && state.data.pages) || {})
// link to a record card (kind = page handle) or to the creation form (no id)
function recordURL (kind, id) {
  const pid = pages.value[kind]
  if (!base || !pid) return ''
  return `${base}/pages/${pid}/record${id ? '/' + encodeURIComponent(id) : ''}`
}
const cellLink = (c) => (c.docs && c.docs.length ? recordURL('cmp_document', c.docs[0]) : '')
const assistURL = (id) => `?view=doc&source=cmp_document&id=${encodeURIComponent(id)}&namespaceID=${encodeURIComponent(namespaceID)}`
const docLabel = (d) => ((d.codes || []).length ? d.codes.join(', ') : d.title || 'файл не прочитан') + (d.pages ? ` · ${d.pages} стр.` : '')
const viewerURL = (id) => `?recordID=${encodeURIComponent(id)}&namespaceID=${encodeURIComponent(namespaceID)}`
</script>

<template>
  <div class="wrap ov">
    <template v-if="state.loading">Загрузка…</template>
    <div v-else-if="state.error && !state.data" class="err">{{ state.error }}</div>
    <template v-else>
      <div class="head">
        <h1>Документы объекта — {{ state.data.objectName }}</h1>
        <span class="grow"></span>
        <button class="small" :disabled="state.busy" title="Прочитать новые файлы комплектов и сверить тома с ведомостью состава" @click="check">
          {{ state.busy ? 'Проверка…' : 'Проверить комплектность' }}
        </button>
        <button class="small" :disabled="state.busy" title="Агент прочитает содержание документов и первые комплекты и предложит ведомость томов — вы принимаете или отклоняете" @click="suggest">
          Предложить ведомость томов
        </button>
      </div>
      <div v-if="state.error" class="err action-err">{{ state.error }}</div>
      <div v-if="state.notice" class="comment">{{ state.notice }}</div>

      <div class="comment">Название комплекта, документа или сравнения — ссылка на его карточку. Новый комплект создаётся здесь или на вкладке «Комплекты»; файлы добавляются в карточке комплекта.</div>

      <div v-if="!hasAny && !proposals.length" class="empty">
        Комплектов документов ещё нет. Создайте комплект на вкладке «Комплекты» (стадия, версия) и добавьте в него PDF-файлы —
        здесь появятся комплектность и сравниваемые версии.
      </div>

      <div v-if="proposals.length" class="ov-prop">
        <h2>Предлагаемая ведомость томов <span class="muted">— пока не принята, на комплектность не влияет</span></h2>
        <table class="ov-table">
          <thead><tr><th></th><th>Стадия</th><th>Шифр</th><th>Наименование</th><th>Откуда</th></tr></thead>
          <tbody>
            <tr v-for="p in proposals" :key="p.id">
              <td><input v-model="picked[p.id]" type="checkbox"></td>
              <td>{{ p.stage === 'ref' ? 'ПД' : 'РД' }}</td>
              <td>{{ p.code }}</td>
              <td>{{ p.title }}</td>
              <td class="muted">{{ SOURCE_LABEL[p.source] || p.source }}<template v-if="p.evidence"> · {{ p.evidence }}</template></td>
            </tr>
          </tbody>
        </table>
        <div class="ov-prop-acts">
          <button class="small primary" :disabled="state.busy || !pickedIDs.length" @click="acceptPicked">Принять отмеченные ({{ pickedIDs.length }})</button>
          <button class="small" :disabled="state.busy || !pickedIDs.length" @click="rejectPicked">Отклонить отмеченные</button>
          <button class="small" :disabled="state.busy" @click="resolve([], [], true)">Принять все</button>
        </div>
      </div>

      <h2>Какая версия с какой сравнивается</h2>
      <div v-if="!state.data.pairs.length" class="empty">Сравнений пока нет.</div>
      <table v-else class="ov-table">
        <thead><tr><th>Сравнение</th><th>Версии комплектов</th><th>Эталон (ПД / до)</th><th>Проверяемый (РД / после)</th><th>Итерация</th><th>Статус</th><th>Схожесть</th><th></th></tr></thead>
        <tbody>
          <tr v-for="p in state.data.pairs" :key="p.id">
            <td><a v-if="recordURL('pd_rd_comparison', p.id)" :href="recordURL('pd_rd_comparison', p.id)" target="_top">{{ p.title }}</a><template v-else>{{ p.title }}</template></td>
            <td><b class="pair">{{ p.label }}</b></td>
            <td><span v-if="p.ref" :class="['tag', p.ref.latest ? 'ok' : 'warn']">{{ p.ref.label }}{{ p.ref.latest ? '' : ' · не последняя' }}</span><span v-else class="tag bad">не указан</span></td>
            <td><span v-if="p.target" :class="['tag', p.target.latest ? 'ok' : 'warn']">{{ p.target.label }}{{ p.target.latest ? '' : ' · не последняя' }}</span><span v-else class="tag bad">не указан</span></td>
            <td>{{ p.iteration || '' }}</td>
            <td>{{ COMP_STATUS[p.status] || p.status }}</td>
            <td>{{ p.similarity ? p.similarity + ' %' : '' }}</td>
            <td><a class="small-link" :href="viewerURL(p.id)" target="_blank">открыть просмотр</a></td>
          </tr>
          <template v-for="p in state.data.pairs" :key="p.id + 'w'">
            <tr v-if="p.warnings.length"><td></td><td colspan="7" class="warns"><span v-for="w in p.warnings" :key="w">⚠ {{ w }}</span></td></tr>
          </template>
        </tbody>
      </table>

      <h2>Комплекты и комплектность</h2>
      <div class="ov-stages">
        <section v-for="st in stages" :key="st.key" class="ov-stage">
          <h3>{{ st.name }} <small v-if="st.expected">в ведомости томов: {{ st.expected }}</small><small v-else>ведомость состава не задана</small></h3>
          <div v-if="!st.sets.length" class="empty">Комплектов {{ st.name }} нет.</div>
          <a v-if="recordURL('document_set', '')" class="small-link add" :href="recordURL('document_set', '')" target="_top">+ создать комплект</a>
          <div v-for="s in st.sets" :key="s.id" class="ov-set">
            <div class="ov-set-head">
              <b><a v-if="recordURL('document_set', s.id)" :href="recordURL('document_set', s.id)" target="_top" title="Открыть комплект: файлы, версия, статус">{{ st.name }} {{ s.version }}</a><template v-else>{{ st.name }} {{ s.version }}</template></b>
              <span v-if="s.newer" class="tag warn">есть более новая: {{ s.newer }}</span>
              <span class="muted">{{ STATUS[s.status] || s.status || '' }}</span>
              <span v-if="s.issuedAt" class="muted">от {{ s.issuedAt.slice(0, 10) }}</span>
            </div>
            <div v-if="s.total" class="bar"><div :style="{ width: s.pct + '%', background: barColor(s) }"></div></div>
            <div class="ov-set-line"><b v-if="s.total">{{ s.present }} из {{ s.total }} ({{ s.pct }} %)</b> {{ s.note }}</div>
            <div v-if="(s.units || []).length" class="ov-units">
              <span v-for="u in s.units" :key="u.kind + u.code" class="tag" :class="u.present ? 'ok' : (u.required ? 'bad' : '')" :title="u.present ? 'есть: ' + (u.where || '') : (u.required ? 'обязательна, нет в комплекте' : 'нет в комплекте; не обязательна — не нарушение')">
                {{ UNIT[u.kind] }} {{ u.code }}<template v-if="u.parent"> · {{ u.parent }}</template>
              </span>
            </div>
            <ul class="ov-docs">
              <li v-for="d in s.docs" :key="d.id">
                <a v-if="recordURL('cmp_document', d.id)" :href="recordURL('cmp_document', d.id)" target="_top" title="Открыть документ">{{ docLabel(d) }}</a>
                <template v-else>{{ docLabel(d) }}</template>
                <a class="as-link" :href="assistURL(d.id)" target="_blank" title="Открыть с ассистентом: вопросы по листам документа">✦ ассистент</a>
              </li>
              <li v-if="!s.docs.length" class="muted">в комплекте нет файлов — откройте комплект и добавьте PDF</li>
            </ul>
          </div>
        </section>
      </div>

      <h2>Документы по сооружениям и разделам</h2>
      <div v-if="!(state.data.tree || []).length" class="empty">Файлов в комплектах пока нет.</div>
      <div v-else class="ov-tree">
        <details v-for="st in state.data.tree" :key="st.id || '-'" open class="ov-tree-st">
          <summary><b>{{ st.name }}</b> <span class="muted">{{ st.sections.reduce((n, x) => n + x.volumes.length, 0) }} томов</span></summary>
          <div v-for="sec in st.sections" :key="sec.name" class="ov-tree-sec">
            <div class="ov-tree-sec-h">Раздел {{ sec.name }}</div>
            <div v-for="v in sec.volumes" :key="v.code" class="ov-tree-vol">
              <span class="code">{{ v.code }}</span> <span class="muted">{{ v.title }}</span>
              <span v-for="e in v.editions" :key="e.docID" class="tag" :class="{ ok: e.latest }" :title="e.latest ? 'Последний комплект стадии' : 'Более ранний комплект'">
                <a v-if="recordURL('cmp_document', e.docID)" :href="recordURL('cmp_document', e.docID)" target="_top">{{ e.set }}<template v-if="e.changeNo"> · изм. {{ e.changeNo }}</template></a>
                <template v-else>{{ e.set }}<template v-if="e.changeNo"> · изм. {{ e.changeNo }}</template></template>
              </span>
            </div>
          </div>
        </details>
        <div v-if="!(state.data.structures || []).length" class="muted">Сооружения объекта не заведены: добавьте их на вкладке «Сооружения» — документы привяжутся по наименованию или шифру при следующей проверке комплектности.</div>
      </div>

      <h2>Журнал загрузок <span class="muted">— последние {{ Math.min(10, (state.data.uploads || []).length) }}</span></h2>
      <div v-if="!(state.data.uploads || []).length" class="empty">Записей нет: журнал ведётся с первой проверки комплектности после обновления.</div>
      <table v-else class="ov-table">
        <thead><tr><th>Когда</th><th>Кто</th><th>Файл</th><th>Шифр</th><th>Что это</th></tr></thead>
        <tbody>
          <tr v-for="(u, i) in state.data.uploads.slice(0, 10)" :key="i">
            <td>{{ (u.at || '').replace('T', ' ').slice(0, 16) }}</td><td>{{ u.by || '—' }}</td><td>{{ u.file }}</td><td>{{ u.codes }}</td>
            <td><span class="tag" :class="{ warn: u.kind === 'repeat' || u.kind === 'unparsed' }">{{ UPLOAD[u.kind] || u.kind }}</span> <span class="muted">{{ u.note }}</span></td>
          </tr>
        </tbody>
      </table>

      <h2>Тома по стадиям</h2>
      <div v-if="!state.data.matrix.length" class="empty">Пока нет ни ведомости томов, ни прочитанных файлов.</div>
      <div v-else class="ov-matrix-wrap">
        <table class="ov-table matrix">
          <thead><tr><th>Шифр тома</th><th>Наименование</th><th>ПД</th><th>РД</th></tr></thead>
          <tbody>
            <tr v-for="r in state.data.matrix" :key="r.code" :class="{ unexpected: !r.expected && hasStatement }">
              <td>{{ r.code }}<span v-if="!r.expected && hasStatement" class="muted"> · не в ведомости</span></td>
              <td>{{ r.title }}</td>
              <td><component :is="cellLink(r.ref) ? 'a' : 'span'" :href="cellLink(r.ref)" target="_top" :class="['cell', r.ref.state]">{{ CELL[r.ref.state] }}<template v-if="r.ref.version && r.ref.state === 'present'"> · {{ r.ref.version }}</template></component></td>
              <td><component :is="cellLink(r.target) ? 'a' : 'span'" :href="cellLink(r.target)" target="_top" :class="['cell', r.target.state]">{{ CELL[r.target.state] }}<template v-if="r.target.version && r.target.state === 'present'"> · {{ r.target.version }}</template></component></td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<style scoped>
.ov h2 { font-size: 13px; margin: 14px 0 6px; }
.ov h3 { font-size: 12.5px; margin: 0 0 6px; }
.ov h3 small { font-weight: 400; color: var(--muted); margin-left: 6px; }
.ov .head .grow { flex: 1; }
.muted { color: var(--muted); font-size: 11.5px; }
.ov-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.ov-table th { text-align: left; padding: 5px 6px; background: var(--panel); border-bottom: 1px solid var(--border); white-space: nowrap; }
.ov-table td { padding: 4px 6px; border-bottom: 1px solid var(--border); vertical-align: top; }
.pair { white-space: nowrap; }
.tag { display: inline-block; padding: 1px 7px; border-radius: var(--radius-lg); font-size: 11px; border: 1px solid var(--border); background: var(--panel); }
.tag.ok { color: #fff; background: var(--ok); border-color: var(--ok); }
.tag.warn { color: #fff; background: var(--warn); border-color: var(--warn); }
.tag.bad { color: #fff; background: var(--bad); border-color: var(--bad); }
.ov-docs { margin: 4px 0 0; padding-left: 16px; font-size: 11.5px; }
.ov a { color: var(--accent); text-decoration: none; }
.ov a:hover { text-decoration: underline; }
.as-link { margin-left: 6px; font-size: 11px; }
.add { display: inline-block; margin: 2px 0 6px; }
.ov-prop { border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 6px 10px; background: var(--panel); margin: 10px 0; }
.ov :deep(button.small.primary) { background: var(--accent); border-color: var(--accent); color: #fff; }
.ov-prop-acts { margin-top: 6px; display: flex; flex-wrap: wrap; gap: 6px; }
.warns { color: var(--warn); font-size: 11.5px; }
.warns span { display: block; }
.ov-stages { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 10px; }
.ov-stage { border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 8px 10px; background: var(--panel); }
.ov-set { border-top: 1px solid var(--border); padding: 6px 0; }
.ov-set-head { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }
.bar { height: 7px; border-radius: 999px; background: var(--bg); border: 1px solid var(--border); overflow: hidden; margin: 4px 0; }
.bar > div { height: 100%; }
.ov-set-line { font-size: 12px; }
.ov-units { display: flex; gap: 4px; flex-wrap: wrap; margin: 3px 0; }
.ov-tree { display: flex; flex-direction: column; gap: 6px; font-size: 12px; }
.ov-tree-st { border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 6px 10px; }
.ov-tree-sec { margin: 4px 0 4px 12px; }
.ov-tree-sec-h { font-weight: 600; margin: 4px 0 2px; }
.ov-tree-vol { margin-left: 12px; display: flex; gap: 6px; align-items: baseline; flex-wrap: wrap; }
.ov-tree-vol .code { font-family: var(--mono, monospace); }
.ov-matrix-wrap { max-height: 60vh; overflow: auto; border: 1px solid var(--border); border-radius: var(--radius-lg); }
.matrix tr.unexpected td { background: color-mix(in srgb, var(--warn) 8%, transparent); }
.cell { padding: 1px 8px; border-radius: 4px; font-size: 11.5px; }
.cell.present { color: var(--ok); font-weight: 600; }
.cell.missing { color: var(--bad); font-weight: 600; background: color-mix(in srgb, var(--bad) 10%, transparent); }
.cell.extra { color: var(--warn); font-weight: 600; }
.cell.no_set, .cell.unparsed { color: var(--muted); }
</style>
