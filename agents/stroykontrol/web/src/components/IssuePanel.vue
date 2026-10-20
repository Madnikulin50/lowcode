<script setup>
// The list of remarks made from the comparison: one remark per sheet, from the
// findings the specialist confirmed. It goes into the registry of the object's
// remarks, or out as Excel (the format of an expertise export) or a Word letter.
import { computed } from 'vue'
import { CRIT_LABEL } from '../remarks/labels.js'

const props = defineProps({ issue: Object, result: Object, busy: Boolean, qs: String, namespaceId: String, discrepancies: Array })
const emit = defineEmits(['register', 'confirm-many'])

const active = computed(() => props.issue?.active || [])
const fixed = computed(() => props.issue?.fixed || [])
const confirmable = computed(() => props.discrepancies.filter(d => (d.status || 'new') === 'new' && d.severity !== 'low' && d.kind !== 'not_released' && d.kind !== 'scope_diff' && d.id).map(d => d.id))
const workbench = computed(() => props.issue?.objectID ? `?view=remarks&objectID=${encodeURIComponent(props.issue.objectID)}&namespaceID=${encodeURIComponent(props.namespaceId || '')}` : '')
const href = (format) => `api/issue/export?${props.qs}&format=${format}`
</script>

<template>
  <div v-if="issue" class="issue">
    <h2>
      Список замечаний по сравнению
      <span class="muted">подтверждено находок: {{ issue.confirmed }} → замечаний: {{ active.length }}</span>
    </h2>

    <div v-if="!active.length && !fixed.length" class="empty">
      Замечания формируются по <b>подтверждённым</b> находкам (✓ в списке расхождений выше): все находки одного листа собираются в одно замечание.
      <button v-if="confirmable.length" class="small" :disabled="busy" @click="emit('confirm-many', confirmable)">Подтвердить все выше «низкой» ({{ confirmable.length }})</button>
    </div>

    <template v-else>
      <div class="issue-actions">
        <button class="primary" :disabled="busy || !active.length || !issue.objectID" :title="issue.objectID ? 'Создать замечания в реестре объекта (повторное нажатие обновит их, не создавая дублей)' : 'У сравнения не указан объект'" @click="emit('register')">Внести в реестр замечаний объекта</button>
        <a v-if="active.length" class="btn" :href="href('xlsx')" download>Скачать Excel</a>
        <a v-if="active.length" class="btn" :href="href('docx')" download>Скачать Word</a>
        <a v-if="workbench" class="small-link" :href="workbench" target="_self">Рабочее место замечаний →</a>
      </div>
      <div v-if="!issue.objectID" class="exp-note">Чтобы вести замечания в реестре, укажите объект в карточке сравнения.</div>
      <div v-if="result" class="comment">
        Реестр обновлён: создано {{ result.created }}<template v-if="result.numbers && result.numbers.length"> ({{ result.numbers.join(', ') }})</template>,
        обновлено {{ result.updated }}<template v-if="result.closed">, закрыто {{ result.closed }} (расхождения устранены)</template><template v-if="result.reopened">, возвращено {{ result.reopened }}</template>.
      </div>

      <div v-for="is in active" :key="is.groupKey" class="iss">
        <div class="iss-head">
          <b class="iss-num">№ {{ is.number }}</b>
          <span class="tag" :class="'crit-' + is.criticality">{{ CRIT_LABEL[is.criticality] }}</span>
          <span class="tag tag-page">{{ is.pages }}</span>
          <span class="muted">находок: {{ is.count }}</span>
        </div>
        <div class="iss-refs">{{ is.refs }}</div>
        <div class="iss-text">{{ is.text }}</div>
      </div>
      <div v-if="fixed.length" class="exp-note">Устранено при повторной проверке (замечания закрываются): {{ fixed.length }}.</div>
      <div v-if="issue.excluded" class="exp-note">Не вошли в список: {{ issue.excluded }} находок (не подтверждены, отклонены или информационные).</div>
    </template>
  </div>
</template>
