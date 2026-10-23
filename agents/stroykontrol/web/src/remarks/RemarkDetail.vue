<script setup>
// One remark: the text, the designer's answer next to the previous one (words
// that changed highlighted), the check against the documents, the history of
// iterations and the specialist's own fields.
import { ref, watch, computed } from 'vue'
import { wordDiffSpans } from '../composables/useTextDiff.js'
import { me, can } from '../platform.js'
import { STATUS_LABEL, STATUS_ORDER, CRIT_LABEL, VERDICT_LABEL, AI_LABEL, OUTCOME_LABEL, OUTCOME_ORDER, STATUS_NEEDS_REASON, MAT_STATE_LABEL, MATERIALS_LABEL, MAT_KIND_LABEL, fmtDate } from './labels.js'

const props = defineProps({ remark: Object, board: Object, history: Array, busy: Boolean, assignees: Array })
const emit = defineEmits(['update', 'close', 'ai', 'requirements'])

const form = ref({ status: '', statusComment: '', assignee: '', dueDate: '', note: '', estimateH: '' })
watch(() => props.remark, (r) => {
  form.value = { status: r.status, statusComment: '', assignee: r.assignee || '', dueDate: (r.dueDate || '').slice(0, 10), note: r.note || '', estimateH: r.estimateH || '' }
}, { immediate: true })

const dirty = computed(() => {
  const r = props.remark
  return form.value.status !== r.status || form.value.assignee !== (r.assignee || '') || form.value.dueDate !== (r.dueDate || '').slice(0, 10) || form.value.note !== (r.note || '') ||
    form.value.estimateH !== (r.estimateH || '')
})
// «Устранено» и «Возвращено на доработку» нужно объяснить: без комментария не сохраняется
const needsReason = computed(() => form.value.status !== props.remark.status && STATUS_NEEDS_REASON.includes(form.value.status))
const reasonMissing = computed(() => needsReason.value && !form.value.statusComment.trim())
function save () {
  const r = props.remark
  const patch = {}
  if (form.value.status !== r.status) { patch.status = form.value.status; patch.statusComment = form.value.statusComment.trim() }
  if (form.value.estimateH !== (r.estimateH || '')) patch.estimateH = form.value.estimateH
  if (form.value.assignee !== (r.assignee || '')) patch.assignee = form.value.assignee
  if (form.value.dueDate !== (r.dueDate || '').slice(0, 10)) patch.dueDate = form.value.dueDate
  if (form.value.note !== (r.note || '')) patch.note = form.value.note
  emit('update', r.id, patch)
}

// Результат проверки (ТЗ 3.4.6): вывод системы и решение специалиста — раздельно
const outcome = ref({ final: '', comment: '' })
watch(() => props.remark, (r) => {
  outcome.value = { final: r.outcomeFinal || r.outcome, comment: r.outcomeComment || '' }
}, { immediate: true })
const outcomeChanged = computed(() => outcome.value.final !== props.remark.outcome)
const outcomeDirty = computed(() => outcome.value.final !== (props.remark.outcomeFinal || '') || outcome.value.comment !== (props.remark.outcomeComment || ''))
const outcomeInvalid = computed(() => outcomeChanged.value && !outcome.value.comment.trim())
function saveOutcome () {
  emit('update', props.remark.id, { outcomeFinal: outcome.value.final, outcomeComment: outcome.value.comment.trim() })
}
function clearOutcome () {
  emit('update', props.remark.id, { outcomeFinal: '', outcomeComment: '' })
}

// Requirements of the remark (ТЗ 3.4.3, 3.6.6): the list the program made, edited by the specialist; each requirement has its own
// result and its own decision (another result than the system's needs a comment).
const reqDraft = ref([])
const reqEditing = ref(false)
watch(() => props.remark, (r) => {
  reqDraft.value = (r.requirements || []).map(q => ({ ...q, text: q.text, outcomeFinal: q.outcomeFinal || '', outcomeComment: q.outcomeComment || '' }))
  reqEditing.value = false
}, { immediate: true })
const reqEffective = (q) => q.outcomeFinal || q.outcome || 'check'
const reqInvalid = computed(() => reqDraft.value.some(q => !q.text.trim() || (q.outcomeFinal && q.outcomeFinal !== (q.outcome || '') && !q.outcomeComment.trim())))
const reqDirty = computed(() => JSON.stringify(reqDraft.value.map(q => [q.text, q.outcomeFinal, q.outcomeComment])) !==
  JSON.stringify((props.remark.requirements || []).map(q => [q.text, q.outcomeFinal || '', q.outcomeComment || ''])))
const addReq = () => { reqDraft.value.push({ text: '', outcomeFinal: '', outcomeComment: '', outcome: '' }); reqEditing.value = true }
const removeReq = (i) => { reqDraft.value.splice(i, 1); reqEditing.value = true }
const saveReqs = () => emit('requirements', props.remark.id, reqDraft.value.map(q => ({ text: q.text.trim(), outcomeFinal: q.outcomeFinal, outcomeComment: q.outcomeComment.trim() })))

// answer now vs the previous answer
const answerDiff = computed(() => {
  const r = props.remark
  if (!r.answer || !r.prevAnswer || r.answer === r.prevAnswer) return null
  return wordDiffSpans(r.prevAnswer, r.answer)
})
const docName = (id) => props.board.documents[id] || id

// Where the remark leads (ТЗ 3.5): the places its references name and the regions of the findings it was
// matched to. Source = the original edition (ПД), corrected = the corrected one (РД).
const ns = () => encodeURIComponent(props.board.namespaceID || '')
const places = computed(() => (props.remark.links || []).filter(l => l.kind === 'place'))
const found = computed(() => (props.remark.links || []).filter(l => l.kind === 'finding'))
// the fragment the text of the remark names, found in the linked volumes (ТЗ 3.4.2, 3.5.1)
const fragments = computed(() => (props.remark.links || []).filter(l => l.kind === 'fragment'))
const boxParam = (b) => b ? '&box=' + encodeURIComponent(String(b).replace(/[[\]]/g, '')) : ''
const fragmentHref = (l) => placeHref(l) + boxParam(l.box)
// the AI quotations open the volume on their page with the quoted lines framed (ТЗ 3.4.7)
const citeHref = (c) => c.docID ? `?view=doc&source=cmp_document&id=${encodeURIComponent(c.docID)}&namespaceID=${ns()}&page=${c.page || 1}` + boxParam(c.box) : ''
const SIDE_LABEL = { ref: 'исходная редакция', target: 'исправленная редакция' }
const placeHref = (l) => l.comparison
  ? `?view=doc&source=comparison_doc&recordID=${encodeURIComponent(l.comparison)}&namespaceID=${ns()}&id=${encodeURIComponent(l.docID)}&page=${l.page || 1}`
  : `?view=doc&source=cmp_document&id=${encodeURIComponent(l.docID)}&namespaceID=${ns()}&page=${l.page || 1}`
// a declared material found in a loaded volume opens it on that page; a supporting file has no viewer link
const materialHref = (m) => m.docID && !m.docID.startsWith('att:')
  ? `?view=doc&source=cmp_document&id=${encodeURIComponent(m.docID)}&namespaceID=${ns()}&page=${m.at || 1}`
  : ''
const findingHref = (l) => `?recordID=${encodeURIComponent(l.comparison)}&namespaceID=${ns()}&findingID=${encodeURIComponent(l.findingID)}`
const sideHref = (l, side) => {
  const doc = side === 'ref' ? l.refDoc : l.targetDoc
  const page = side === 'ref' ? l.refPage : l.targetPage
  return doc ? `?view=doc&source=comparison_doc&recordID=${encodeURIComponent(l.comparison)}&namespaceID=${ns()}&id=${encodeURIComponent(doc)}&page=${page || 1}` : ''
}
// what the person may do (ТЗ 3.8): the agent refuses the rest anyway, the form only does not offer it
const mayConfirm = computed(() => can('confirm'))
const mayAssign = computed(() => can('assign'))
const mayWork = computed(() => can('work'))
const confirmHint = computed(() => mayConfirm.value ? '' : 'Подтверждать устранение и возвращать на доработку может специалист по контролю качества или руководитель проекта')
const linkedSet = computed(() => new Set(props.remark.documents))
</script>

<template>
  <div class="rm-detail">
    <div class="rm-d-head">
      <b class="rm-num">№ {{ remark.number || remark.key }}</b>
      <span v-if="remark.criticality" class="tag" :class="'crit-' + remark.criticality">{{ CRIT_LABEL[remark.criticality] }}</span>
      <span class="tag">итерация {{ remark.iterations || 1 }}</span>
      <span v-if="remark.sourceStatus" class="tag" title="Статус в последней выгрузке">в выгрузке: {{ remark.sourceStatus }}</span>
      <span v-if="remark.archived" class="tag warn">нет в последней выгрузке</span>
      <span class="grow"></span>
      <button class="small" @click="emit('close')">Закрыть ✕</button>
    </div>

    <div class="rm-meta">
      <span v-if="remark.discipline"><b>НД:</b> {{ remark.discipline }}</span>
      <span v-if="remark.expert"><b>Эксперт:</b> {{ remark.expert }}</span>
      <span v-if="remark.section"><b>Раздел:</b> {{ remark.section }}</span>
      <span v-if="remark.changedAt"><b>Изменено:</b> {{ fmtDate(remark.changedAt) }}</span>
    </div>

    <div class="rm-block">
      <div class="rm-h">Вывод о несоответствии</div>
      <div class="rm-text">{{ remark.text }}</div>
    </div>
    <div class="rm-block rm-reqs">
      <div class="rm-h">
        Требования ({{ reqDraft.length }})
        <button class="small" :disabled="busy" title="Разбить замечание на требования вручную: каждое проверяется отдельно" @click="addReq">+ требование</button>
      </div>
      <div v-if="remark.requirementsStale" class="rm-warn">Текст замечания изменился после того, как вы правили список: проверьте, что требования ему соответствуют.</div>
      <div v-for="(q, i) in reqDraft" :key="i" class="rm-req">
        <div class="rm-req-head">
          <b>{{ i + 1 }}.</b>
          <span class="tag" :class="'oc-' + reqEffective(q)" :title="q.outcomeFinal ? 'Решение специалиста' : 'Вывод системы'">{{ OUTCOME_LABEL[reqEffective(q)] }}</span>
          <span v-if="q.manual" class="tag" title="Требование введено или правлено вручную">вручную</span>
          <span v-if="q.verdict" class="muted" :title="q.note">проверка: {{ VERDICT_LABEL[q.verdict] || q.verdict }}</span>
          <span v-if="q.aiVerdict" class="tag" :class="'ai-' + q.aiVerdict" :title="q.aiNote">{{ AI_LABEL[q.aiVerdict] }}</span>
          <button class="small" :disabled="busy" @click="removeReq(i)">убрать</button>
        </div>
        <textarea v-model="q.text" rows="2" :placeholder="'Текст требования'"></textarea>
        <div class="rm-req-decision">
          <label>Решение
            <select v-model="q.outcomeFinal">
              <option value="">— не принято —</option>
              <option v-for="k in OUTCOME_ORDER" :key="k" :value="k">{{ OUTCOME_LABEL[k] }}</option>
            </select>
          </label>
          <input v-if="q.outcomeFinal" v-model="q.outcomeComment" type="text" :placeholder="q.outcomeFinal !== (q.outcome || '') ? 'Комментарий обязателен: результат отличается от вывода системы' : 'Комментарий'" />
        </div>
        <div v-if="q.note" class="muted rm-req-note">{{ q.note }}</div>
      </div>
      <button class="primary" :disabled="busy || !reqDirty || reqInvalid || !mayWork" @click="saveReqs">Сохранить требования</button>
      <div class="muted">Результат замечания складывается из результатов требований: все устранены — «Устранено», часть — «Частично устранено». Список можно править; поправленный агент не перезаписывает.</div>
    </div>

    <div v-if="remark.basis" class="rm-block"><div class="rm-h">Основание</div><div class="rm-text muted">{{ remark.basis }}</div></div>

    <div v-if="remark.refs" class="rm-block">
      <div class="rm-h">Ссылка на материалы</div>
      <div class="rm-text muted">{{ remark.refs }}</div>
      <div v-if="remark.documents.length" class="rm-docs">
        Загружены: <span v-for="id in remark.documents" :key="id" class="tag doc">{{ docName(id) }}
          <a :href="`?view=doc&source=cmp_document&id=${encodeURIComponent(id)}&namespaceID=${encodeURIComponent(board.namespaceID || '')}`" target="_blank" title="Открыть том с ассистентом">✦</a></span>
      </div>
    </div>

    <div v-if="places.length || found.length || fragments.length" class="rm-block rm-links">
      <div class="rm-h">Где в документах</div>
      <div v-for="(l, i) in places" :key="'p' + i" class="rm-link">
        <span class="tag">{{ SIDE_LABEL[l.side] || 'документ' }}</span>
        <a :href="placeHref(l)" target="_blank" title="Открыть том на этой странице">{{ l.doc || docName(l.docID) }}</a>
        <span v-if="l.page">, стр. {{ l.page }}</span><span v-if="l.sheet"> (лист {{ l.sheet }})</span>
        <span v-if="l.note" class="muted"> — {{ l.note }}</span>
      </div>
      <div v-for="(l, i) in fragments" :key="'g' + i" class="rm-link">
        <span class="tag">фрагмент · {{ SIDE_LABEL[l.side] || 'документ' }}</span>
        <a :href="fragmentHref(l)" target="_blank" title="Открыть том на этой странице с выделенным фрагментом">{{ l.doc || docName(l.docID) }}, стр. {{ l.page }}<span v-if="l.sheet"> (лист {{ l.sheet }})</span></a>
        <span class="muted"> — «{{ l.text }}»</span>
        <span v-if="l.note" class="muted"> ({{ l.note }})</span>
      </div>
      <div v-for="(l, i) in found" :key="'f' + i" class="rm-link">
        <span class="tag">находка сравнения</span>
        <a :href="findingHref(l)" target="_blank" title="Открыть сравнение на этой находке с выделенной областью">{{ l.text || 'расхождение' }}</a>
        <span class="rm-link-sides">
          <a v-if="sideHref(l, 'ref')" :href="sideHref(l, 'ref')" target="_blank">исходный фрагмент (стр. {{ l.refPage || 1 }})</a>
          <a v-if="sideHref(l, 'target')" :href="sideHref(l, 'target')" target="_blank">исправленный фрагмент (стр. {{ l.targetPage || 1 }})</a>
        </span>
      </div>
    </div>

    <div class="rm-block">
      <div class="rm-h">Ответ на замечание</div>
      <div v-if="!remark.answer" class="rm-text muted">ответ не предоставлен</div>
      <div v-else-if="answerDiff" class="rm-text">
        <template v-for="(s, i) in answerDiff.addSpans" :key="i"><span :class="{ 'w-add': s.hi }">{{ s.t }}</span>{{ ' ' }}</template>
      </div>
      <div v-else class="rm-text">{{ remark.answer }}</div>
      <div v-if="remark.rejectReason" class="rm-reject"><b>Причина отклонения:</b> {{ remark.rejectReason }}</div>
      <details v-if="remark.prevAnswer || remark.prevRejectReason" class="rm-prev">
        <summary>Предыдущая итерация</summary>
        <div v-if="remark.prevAnswer" class="rm-text muted">{{ remark.prevAnswer }}</div>
        <div v-if="remark.prevRejectReason" class="rm-reject"><b>Причина отклонения:</b> {{ remark.prevRejectReason }}</div>
      </details>
    </div>

    <div v-if="(remark.materials || []).length" class="rm-block rm-materials">
      <div class="rm-h">Заявленные в ответе материалы
        <span class="tag" :class="'mat-' + remark.materialsState">{{ MATERIALS_LABEL[remark.materialsState] || '' }}</span>
      </div>
      <div class="muted">Проверяется только наличие тома, листа, приложения или документа; правильность содержания оценивает специалист.</div>
      <div v-for="(m, i) in remark.materials" :key="i" class="rm-link">
        <span class="tag">{{ MAT_KIND_LABEL[m.kind] || m.kind }}</span>
        <span>{{ m.label }}</span>
        <span class="tag" :class="'mat-' + m.state">{{ MAT_STATE_LABEL[m.state] || m.state }}</span>
        <template v-if="m.state === 'found' && m.in">
          — <a v-if="materialHref(m)" :href="materialHref(m)" target="_blank" title="Открыть документ на этой странице">{{ m.in }}<span v-if="m.at">, стр. {{ m.at }}</span></a>
          <span v-else>{{ m.in }}<span v-if="m.at">, стр. {{ m.at }}</span></span>
        </template>
        <span v-if="m.note" class="muted"> — {{ m.note }}</span>
      </div>
    </div>

    <div v-if="remark.proposals || remark.expected" class="rm-block">
      <div class="rm-h">Предложения по отработке</div>
      <div class="rm-text">{{ remark.proposals }}</div>
      <div v-if="remark.expected && remark.expected !== remark.proposals" class="rm-text muted"><b>Ожидаемый результат:</b> {{ remark.expected }}</div>
    </div>

    <div class="rm-block rm-check" :class="'v-' + remark.verdict">
      <div class="rm-h">Проверка по документам</div>
      <div>
        <span v-if="remark.verdict" class="tag" :class="'ev-' + remark.verdict">{{ VERDICT_LABEL[remark.verdict] || remark.verdict }}</span>
        <span v-else class="muted">не проверялось</span>
        <span v-if="remark.claimMismatch" class="tag ev-claim" title="Ответ говорит, что правка внесена, но связанное расхождение в документах осталось">ответ ≠ документы</span>
      </div>
      <div v-if="remark.matchNote" class="rm-text muted">{{ remark.matchNote }}</div>
      <div v-if="remark.findings.length" class="rm-docs">
        Связанных находок: {{ remark.findings.length }}.
        <template v-for="c in board.comparisons" :key="c.id">
          <a :href="`?recordID=${encodeURIComponent(c.id)}&namespaceID=${encodeURIComponent(board.namespaceID || '')}`" target="_self">{{ c.title }}</a>{{ ' ' }}
        </template>
      </div>
    </div>

    <div class="rm-block rm-ai">
      <div class="rm-h">
        Разбор ИИ по фрагментам томов
        <button class="small" :disabled="busy" title="Модель прочитает замечание, ответ и найденные в связанных томах фрагменты. Статус замечания она не меняет." @click="emit('ai', remark.id)">
          {{ remark.aiVerdict ? 'Разобрать заново' : 'Разобрать ИИ' }}
        </button>
      </div>
      <div v-if="remark.aiVerdict">
        <span class="tag" :class="'ai-' + remark.aiVerdict">{{ AI_LABEL[remark.aiVerdict] || remark.aiVerdict }}</span>
        <span v-if="remark.aiAt" class="muted"> {{ fmtDate(remark.aiAt) }}</span>
        <div class="rm-text ai-note">{{ remark.aiNote }}</div>
        <div v-if="(remark.aiCites || []).length" class="rm-cites">
          <div v-for="(c, i) in remark.aiCites" :key="i" class="rm-link">
            <a v-if="citeHref(c)" :href="citeHref(c)" target="_blank" title="Открыть том на странице цитаты с выделенными строками">{{ c.doc }}, стр. {{ c.page }}</a>
            <span v-else>{{ c.doc }}, стр. {{ c.page }}</span>
            <span class="muted"> — «{{ c.quote }}»</span>
          </div>
        </div>
        <div class="muted">Это подсказка по тексту томов: приведены только проверенные цитаты. Решение и статус — за специалистом.</div>
      </div>
      <div v-else class="muted">Не запускался. Нужны привязанные тома (кнопка «Проверить по документам») и модель (REMARK_MODEL).</div>
    </div>

    <div class="rm-block rm-outcome">
      <div class="rm-h">Результат проверки устранения</div>
      <div class="rm-oc-grid">
        <div><div class="muted">Статус экспертизы (выгрузка)</div><div>{{ remark.sourceStatus || '—' }}</div></div>
        <div><div class="muted">Статус выполнения</div><div>{{ STATUS_LABEL[remark.status] || remark.status }}</div></div>
        <div><div class="muted">Вывод системы</div><span class="tag" :class="'oc-' + remark.outcome">{{ OUTCOME_LABEL[remark.outcome] }}</span></div>
        <div>
          <div class="muted">Решение специалиста</div>
          <span v-if="remark.outcomeFinal" class="tag" :class="'oc-' + remark.outcomeFinal">{{ OUTCOME_LABEL[remark.outcomeFinal] }}</span>
          <span v-else class="muted">не принято</span>
          <div v-if="remark.outcomeFinal" class="muted">{{ remark.outcomeDecision === 'confirmed' ? 'подтверждён вывод системы' : 'изменено специалистом' }} · {{ remark.outcomeBy }} · {{ fmtDate(remark.outcomeAt) }}</div>
        </div>
      </div>
      <div class="rm-form">
        <label>Результат
          <select v-model="outcome.final"><option v-for="k in OUTCOME_ORDER" :key="k" :value="k">{{ OUTCOME_LABEL[k] }}</option></select>
        </label>
      </div>
      <textarea v-model="outcome.comment" rows="2" :placeholder="outcomeChanged ? 'Комментарий обязателен: почему результат отличается от вывода системы' : 'Комментарий (необязательно)'"></textarea>
      <button class="primary" :disabled="busy || !outcomeDirty || outcomeInvalid || !mayConfirm" :title="confirmHint" @click="saveOutcome">{{ outcomeChanged ? 'Изменить результат' : 'Подтвердить вывод системы' }}</button>
      <button v-if="remark.outcomeFinal" class="small" :disabled="busy || !mayConfirm" @click="clearOutcome">Снять решение</button>
      <div class="muted">Вывод системы получается из проверки по документам и разбора ИИ; статус экспертизы и статус выполнения он не меняет.</div>
    </div>

    <div class="rm-block rm-work">
      <div class="rm-h">Работа</div>
      <div class="rm-form">
        <label>Статус
          <select v-model="form.status" :disabled="!mayWork" :title="confirmHint">
            <option v-for="s in STATUS_ORDER" :key="s" :value="s" :disabled="!mayConfirm && STATUS_NEEDS_REASON.includes(s) && s !== remark.status">{{ STATUS_LABEL[s] }}</option>
          </select>
        </label>
        <label>Ответственный
          <input v-model="form.assignee" list="assignees" placeholder="ФИО" :disabled="!mayAssign" :title="mayAssign ? '' : 'Назначает руководитель проекта'" />
          <datalist id="assignees"><option v-for="a in assignees" :key="a" :value="a"></option></datalist>
        </label>
        <label>Срок <input v-model="form.dueDate" type="date" :disabled="!mayAssign" /></label>
        <label>Оценка, ч <input v-model="form.estimateH" type="number" min="0" step="0.5" style="width:80px" /></label>
      </div>
      <textarea v-if="needsReason" v-model="form.statusComment" rows="2" :placeholder="form.status === 'rework' ? 'Что не так и что исправить (обязательно)' : 'Чем подтверждено устранение (обязательно)'"></textarea>
      <textarea v-model="form.note" rows="2" placeholder="Заметка ПТО"></textarea>
      <button class="primary" :disabled="busy || !dirty || reasonMissing || !mayWork" @click="save">Сохранить</button>
      <div v-if="me.checked && !mayWork" class="muted">Только просмотр: роль «{{ me.roleLabels.join(', ') || 'нет роли' }}» не позволяет менять замечания.</div>
    </div>

    <div class="rm-block">
      <div class="rm-h">История ({{ history ? history.length : '…' }})</div>
      <div v-if="history && !history.length" class="muted">записей нет</div>
      <div v-for="(h, i) in history || []" :key="i" class="rm-hist">
        <div class="rm-hist-h">
          <b>итерация {{ h.iteration }}</b>
          <span class="muted">{{ fmtDate(h.at) }} · {{ h.source }}<template v-if="h.author"> · {{ h.author }}</template></span>
          <span v-if="h.status" class="tag">{{ STATUS_LABEL[h.status] || h.status }}</span>
          <span v-else-if="h.sourceStatus" class="tag">{{ h.sourceStatus }}</span>
        </div>
        <div v-if="h.answer" class="rm-text">{{ h.answer }}</div>
        <div v-if="h.rejectReason" class="rm-reject"><b>Причина отклонения:</b> {{ h.rejectReason }}</div>
        <div v-if="h.note" class="rm-text muted"><b>Комментарий:</b> {{ h.note }}</div>
      </div>
    </div>
  </div>
</template>
