<template>
  <div class="d-flex align-items-center">
    <font-awesome-icon
      v-if="!disabled && !disabledSorting && !hideIcons"
      :icon="['fas', 'grip-vertical']"
      :class="{
        'text-muted': disabledDragging,
      }"
      class="align-baseline me-3 text-primary"
    />
    <b class="text-truncate">
      <slot v-bind="item">
        {{ item[textField] }}
      </slot>
    </b>
    <button
      v-if="_hideIcons"
      title="Toggle select"
      :data-test-id="`button-${selected ? 'unselect' : 'select'}`"
      class="btn btn-outline-light d-flex align-items-center ms-auto p-2 border-0"
      @click.prevent.stop="$emit(selected ? 'unselect' : 'select')"
    >
      <font-awesome-icon
        :icon="[selected ? 'far' : 'fas', selected ? 'eye' : 'eye-slash']"
        class="text-muted"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  item: any
  textField?: string
  selected?: boolean
  disabled?: boolean
  disabledDragging?: boolean
  disabledSorting?: boolean
  hideIcons?: boolean
}>(), {
  textField: 'text',
  selected: false,
  disabled: false,
  disabledDragging: false,
  disabledSorting: false,
  hideIcons: false,
})

defineEmits<{
  select: []
  unselect: []
}>()

const _hideIcons = computed(() => !props.disabled && !props.hideIcons)
</script>
