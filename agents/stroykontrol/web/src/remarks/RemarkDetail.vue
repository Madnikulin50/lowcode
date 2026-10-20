<script setup>
// One remark: the text, the designer's answer next to the previous one (words
// that changed highlighted), the check against the documents, the history of
// iterations and the specialist's own fields.
import { ref, watch, computed } from 'vue'
import { wordDiffSpans } from '../composables/useTextDiff.js'
import { STATUS_LABEL, STATUS_ORDER, CRIT_LABEL, VERDICT_LABEL, AI_LABEL, fmtDate } from './labels.js'

const props = defineProps({ remark: Object, board: Object, history: Array, busy: Boolean, assignees: Array })
const emit = defineEmits(['update', 'close', 'ai'])

const form = ref({ status: '', assignee: '', dueDate: '', note: '' })
watch(() => props.remark, (r) => {
  form.value = { status: r.status, assignee: r.assignee || '', dueDate: (r.dueDate || '').slice(0, 10), note: r.note || '' }
}, { immediate: true })

const dirty = computed(() => {
  const r = props.remark
  return form.value.status !== r.status || form.value.assignee !== (r.assignee || '') || form.value.dueDate !== (r.dueDate || '').slice(0, 10) || form.value.note !== (r.note || '')
})
function save () {
  const r = props.remark
  const patch = {}
  if (form.value.status !== r.status) patch.status = form.value.status
  if (form.value.assignee !== (r.assignee || '')) patch.assignee = form.value.assignee
  if (form.value.dueDate !== (r.dueDate || '').slice(0, 10)) patch.dueDate = form.value.dueDate
  if (form.value.note !== (r.note || '')) patch.note = form.value.note
  emit('update', r.id, patch)
}

// answer now vs the previous answer
const answerDiff = computed(() => {
  const r = props.remark
  if (!r.answer || !r.prevAnswer || r.answer === r.prevAnswer) return null
  return wordDiffSpans(r.prevAnswer, r.answer)
})
const docName = (id) => props.board.documents[id] || id
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
    <div v-if="remark.basis" class="rm-block"><div class="rm-h">Основание</div><div class="rm-text muted">{{ remark.basis }}</div></div>

    <div v-if="remark.refs" class="rm-block">
      <div class="rm-h">Ссылка на материалы</div>
      <div class="rm-text muted">{{ remark.refs }}</div>
      <div v-if="remark.documents.length" class="rm-docs">
        Загружены: <span v-for="id in remark.documents" :key="id" class="tag doc">{{ docName(id) }}
          <a :href="`?view=doc&source=cmp_document&id=${encodeURIComponent(id)}&namespaceID=${encodeURIComponent(board.namespaceID || '')}`" target="_blank" title="Открыть том с ассистентом">✦</a></span>
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
        <div class="muted">Это подсказка по тексту томов: приведены только проверенные цитаты. Решение и статус — за специалистом.</div>
      </div>
      <div v-else class="muted">Не запускался. Нужны привязанные тома (кнопка «Проверить по документам») и модель (REMARK_MODEL).</div>
    </div>

    <div class="rm-block rm-work">
      <div class="rm-h">Работа</div>
      <div class="rm-form">
        <label>Статус
          <select v-model="form.status">
            <option v-for="s in STATUS_ORDER" :key="s" :value="s">{{ STATUS_LABEL[s] }}</option>
          </select>
        </label>
        <label>Ответственный
          <input v-model="form.assignee" list="assignees" placeholder="ФИО" />
          <datalist id="assignees"><option v-for="a in assignees" :key="a" :value="a"></option></datalist>
        </label>
        <label>Срок <input v-model="form.dueDate" type="date" /></label>
      </div>
      <textarea v-model="form.note" rows="2" placeholder="Заметка ПТО"></textarea>
      <button class="primary" :disabled="busy || !dirty" @click="save">Сохранить</button>
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
      </div>
    </div>
  </div>
</template>
