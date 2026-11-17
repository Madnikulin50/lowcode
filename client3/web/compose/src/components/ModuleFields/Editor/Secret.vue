<template>
  <div class="mb-3" :class="formGroupStyleClasses">
    <div v-if="!valueOnly" :class="labelColClass">
      <div class="d-flex align-items-center text-primary p-0">
        <span :title="label" class="d-flex">{{ label }}</span>
        <c-hint :tooltip="hint" />
        <slot name="tools" />
      </div>
      <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>
    </div>
    <div :class="contentColClass">
      <div class="input-group">
        <input
          ref="input"
          v-model="draft"
          :type="revealed ? 'text' : 'password'"
          class="form-control"
          :placeholder="placeholder"
          autocomplete="new-password"
          :data-test-id="getFieldCypressId(label)"
          @change="commit"
        >
        <button
          class="btn btn-outline-secondary"
          type="button"
          :title="revealed ? t('kind.secret.hide') : t('kind.secret.show')"
          @click="revealed = !revealed"
        >
          <font-awesome-icon :icon="['fas', revealed ? 'eye-slash' : 'eye']" />
        </button>
        <button
          v-if="!revealedValue && record.recordID && canReveal"
          class="btn btn-outline-secondary"
          type="button"
          :title="t('kind.secret.reveal')"
          @click="reveal"
        >
          <font-awesome-icon :icon="['fas', 'key']" />
        </button>
      </div>
      <div v-if="isMaskedExisting" class="small text-muted mt-1">
        {{ t('kind.secret.unchangedHint') }}
      </div>
      <field-errors :errors="errors" />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed, inject, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorBase } from './base'
import FieldErrors from '../errors'

const props = defineProps({
  namespace: { type: Object, required: true },
  field: { type: Object, required: true },
  record: { type: Object, required: true },
  errors: { type: Object, required: true },
  valueOnly: { type: Boolean, default: false },
  horizontal: { type: Boolean, default: false },
  extraOptions: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['change', 'update:preventPopoverClose'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { value, formGroupStyleClasses, labelColClass, contentColClass, label, hint, description, getFieldCypressId } = useEditorBase(props, emit)
const $ComposeAPI = inject('$ComposeAPI')

const MASK = '••••••••'

const draft = ref('')
const revealed = ref(false)
// Полное значение, дозапрошенное с exposeSecrets=1 (нужно право обновления).
const revealedValue = ref('')

const isMaskedExisting = computed(() => value.value === MASK)

const placeholder = computed(() => {
  if (revealedValue.value) return ''
  if (isMaskedExisting.value) return MASK
  return t('kind.secret.placeholder')
})

const canReveal = computed(() => {
  return props.record.recordID && props.record.moduleID && props.namespace.namespaceID
})

function commit () {
  // Пустой ввод при существующем значении = «не менять» (сервер оставит
  // старое по маске). Отправляем маску — keep-old семантика сервера.
  if (draft.value === '' && (isMaskedExisting.value || revealedValue.value)) {
    value.value = revealedValue.value || MASK
    return
  }
  value.value = draft.value
}

async function reveal () {
  try {
    const { response } = await $ComposeAPI.api().request({
      method: 'get',
      url: `/namespace/${props.namespace.namespaceID}/module/${props.record.moduleID}/record/${props.record.recordID}`,
      params: { exposeSecrets: '1' },
    })
    const rv = response?.response?.values?.find(v => v.name === props.field.name)
    if (rv && rv.value && rv.value !== MASK) {
      revealedValue.value = rv.value
      draft.value = rv.value
    }
    revealed.value = true
  } catch (e) {
    // Нет права на expose — остаёмся в режиме замены.
    value.value = value.value
  }
}
</script>
