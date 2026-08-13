<template>
  <div class="mb-3 p-0 m-0">
    <div
      v-for="(expr, ei) in value"
      :key="ei"
      class="d-flex align-items-center gap-1 mb-2"
    >
      <div class="input-group input-group-sm">
        <span
          class="input-group-text"
          title="$t('validators.expression.tooltip')"
        >
          ƒ
        </span>
        <slot :value="value[ei]">
          <input
            :value="value[ei]"
            class="form-control form-control-sm"
            :placeholder="placeholder"
            @input="onInput(ei, $event)"
          />
        </slot>
      </div>

      <c-input-confirm
        :no-prompt="noPrompt(value[ei])"
        show-icon
        @confirmed="emit('remove', ei)"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
const props = defineProps({
  value: {
    type: Array,
    default: () => ([]),
  },
  placeholder: {
    type: String,
    default: '',
  },
  noPrompt: {
    type: Function,
    default: v => v.length === 0,
  },
})

const emit = defineEmits(['input', 'remove'])

function onInput (ei, e) {
  const newVal = [...props.value]
  newVal[ei] = e.target.value
  emit('input', newVal)
}
</script>
