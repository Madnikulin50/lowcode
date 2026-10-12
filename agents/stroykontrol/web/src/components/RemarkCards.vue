<script setup>
// Remark cards (карточки замечаний): drafted from discrepancies, edited by
// the specialist. The model only phrases a draft on request; the final
// text is always the specialist's.
import { reactive, watch } from 'vue'
import { REMARK_STATUS_LABEL } from '../constants.js'

const props = defineProps({ remarks: Array, busy: Boolean })
const emit = defineEmits(['save', 'suggest'])

// local edit buffers, keyed by card id
const edit = reactive({})
watch(() => props.remarks, (list) => {
  for (const r of list) {
    if (!edit[r.id]) edit[r.id] = { finalText: r.finalText || '', status: r.status || 'draft' }
  }
}, { immediate: true, deep: true })

const dirty = (r) => edit[r.id] && (edit[r.id].finalText !== (r.finalText || '') || edit[r.id].status !== (r.status || 'draft'))
function save (r) { emit('save', { ...r, finalText: edit[r.id].finalText, status: edit[r.id].status }) }
function useDraft (r) { edit[r.id].finalText = r.draftText }
</script>

<template>
  <div v-if="remarks.length" class="remarks">
    <h2>Карточки замечаний ({{ remarks.length }})</h2>
    <div v-for="r in remarks" :key="r.id" class="remark">
      <div class="r-head">
        <b>{{ r.documentRef || 'Карточка' }}</b>
        <span v-if="r.sheet" class="r-sheet">{{ r.sheet }}</span>
        <span class="badge">{{ REMARK_STATUS_LABEL[r.status] || r.status }}</span>
      </div>
      <div class="r-what">{{ r.whatDiverges }}</div>
      <div class="r-must"><b>Требуется:</b> {{ r.mustAlign }}</div>
      <div v-if="r.draftText" class="r-draft">
        <div class="r-draft-h">Черновик формулировки (ИИ) — проверьте перед использованием</div>
        <div>{{ r.draftText }}</div>
        <button @click="useDraft(r)">Взять в текст замечания</button>
      </div>
      <textarea v-if="edit[r.id]" v-model="edit[r.id].finalText" rows="3" placeholder="Текст замечания"></textarea>
      <div v-if="edit[r.id]" class="r-actions">
        <select v-model="edit[r.id].status">
          <option v-for="(label, k) in REMARK_STATUS_LABEL" :key="k" :value="k">{{ label }}</option>
        </select>
        <button class="primary" :disabled="busy || !dirty(r)" @click="save(r)">Сохранить</button>
        <button :disabled="busy" title="Модель только предложит формулировку, решение за вами" @click="emit('suggest', r.id)">Предложить формулировку (ИИ)</button>
      </div>
    </div>
  </div>
</template>
