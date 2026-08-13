<template>
  <div>
    <div v-if="field.isMulti" class="mb-3">
      <label class="form-label text-primary">{{ t('options.multiDelimiter.label') }}</label>
      <div class="btn-group" data-bs-toggle="buttons">
        <label
          v-for="opt in selectOptions"
          :key="opt.value"
          class="btn btn-outline-primary btn-sm"
          :class="{ active: field.options.multiDelimiter === opt.value }"
        >
          <input
            type="radio"
            class="btn-check"
            :value="opt.value"
            :checked="field.options.multiDelimiter === opt.value"
            :disabled="opt.disabled"
            autocomplete="off"
            @change="field.options.multiDelimiter = opt.value"
          />
          {{ t('options.multiDelimiter.' + (opt.value === '\n' ? 'newline' : 'comma')) }}
        </label>
      </div>

      <div class="mb-3 mt-2">
        <label class="form-label text-primary">{{ t('options.multiDelimiter.customLabel') }}</label>
        <input
          v-model="field.options.multiDelimiter"
          type="text"
          class="form-control form-control-sm"
          :disabled="isNotConfigurable"
          :placeholder="t('options.multiDelimiter.customPlaceholder')"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  field: { type: Object, required: true },
})

const { t } = useI18n({ useScope: 'global', messages: {} })

const isNotConfigurable = computed(() => ['File'].includes(props.field.kind))

const selectOptions = computed(() => [
  { text: t('options.multiDelimiter.newline'), value: '\n' },
  { text: t('options.multiDelimiter.comma'), value: ', ', disabled: isNotConfigurable.value },
])
</script>
