<template>
  <div class="d-inline-flex gap-1">
    <button
      v-if="!inConfirmation"
      :title="tooltip"
      :data-test-id="dataTestId"
      :disabled="disabled || processing"
      :class="[`btn btn-${variant}`, btnSize, buttonClass, borderless ? 'border-0' : '', 'flex-fill']"
      :style="buttonStyle"
      @click.stop.prevent="onPrompt"
    >
      <span
        v-if="processing"
        data-test-id="spinner"
        class="spinner-border spinner-border-sm align-middle"
      />

      <template v-else>
        <font-awesome-icon
          v-if="showIcon || !text"
          :icon="icon"
          :class="iconClass"
        />

        <span
          v-if="text"
          :class="textClass"
        >
          {{ text }}
        </span>
      </template>
    </button>

    <template v-else>
      <button
        :data-test-id="`${dataTestId}-confirm`"
        :class="[`btn btn-${variantOk}`, btnSizeConfirm, borderless ? 'border-0' : '', 'flex-fill']"
        :disabled="okDisabled"
        @blur.prevent="onCancel()"
        @click.prevent.stop="onConfirmation()"
      >
        <slot name="yes">
          <font-awesome-icon
            data-test-id="confirm"
            :icon="['fas', 'check']"
          />
        </slot>
      </button>

      <button
        :data-test-id="`${dataTestId}-cancel`"
        :class="[`btn btn-${variantCancel}`, btnSizeConfirm, borderless ? 'border-0' : '', 'flex-fill']"
        :disabled="cancelDisabled"
        @click.prevent.stop="onCancel()"
      >
        <slot name="no">
          <font-awesome-icon
            data-test-id="reject"
            :icon="['fas', 'times']"
          />
        </slot>
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  disabled?: boolean
  okDisabled?: boolean
  cancelDisabled?: boolean
  noPrompt?: boolean
  processing?: boolean
  showIcon?: boolean
  icon?: unknown[]
  buttonClass?: string
  buttonStyle?: Record<string, string>
  iconClass?: string
  textClass?: string
  borderless?: boolean
  variant?: string
  size?: string
  variantOk?: string
  variantCancel?: string
  sizeConfirm?: string
  tooltip?: string
  text?: string
  dataTestId?: string
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  okDisabled: false,
  cancelDisabled: false,
  noPrompt: false,
  processing: false,
  showIcon: false,
  icon: () => ['far', 'trash-alt'],
  buttonClass: '',
  buttonStyle: () => ({}),
  iconClass: '',
  textClass: '',
  borderless: true,
  variant: 'outline-danger',
  size: 'sm',
  variantOk: 'danger',
  variantCancel: 'light',
  sizeConfirm: 'sm',
  tooltip: '',
  text: '',
  dataTestId: 'button-delete',
})

const emit = defineEmits<{
  confirmed: []
  canceled: []
}>()

const btnSize = computed(() => props.size === 'sm' ? 'btn-sm' : props.size === 'lg' ? 'btn-lg' : '')
const btnSizeConfirm = computed(() => props.sizeConfirm === 'sm' ? 'btn-sm' : props.sizeConfirm === 'lg' ? 'btn-lg' : '')

const inConfirmation = ref(false)

function onPrompt () {
  if (props.noPrompt) {
    emit('confirmed')
  } else {
    inConfirmation.value = true
  }
}

function onConfirmation () {
  inConfirmation.value = false
  emit('confirmed')
}

function onCancel () {
  inConfirmation.value = false
  emit('canceled')
}
</script>
