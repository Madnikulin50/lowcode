<template>
  <div
    class="dropdown"
    :class="{ dropup: options?.dropup }"
  >
    <button
      class="btn dropdown-toggle"
      :class="[`btn-${options?.variant || 'secondary'}`, options?.size ? `btn-${options.size}` : '']"
      type="button"
      data-bs-toggle="dropdown"
    >
      <span v-html="label" />
    </button>
    <ul class="dropdown-menu">
      <li
        v-for="(opt, i) in options?.items || []"
        :key="i"
      >
        <template v-if="opt.kind === 'divider'">
          <hr class="dropdown-divider">
        </template>
        <h6
          v-else-if="opt.kind === 'header'"
          class="dropdown-header"
          v-html="opt.label"
        />
        <button
          v-else
          class="dropdown-item"
          :disabled="opt.disabled"
          @click="emit('action', opt.value)"
        >
          <span v-html="opt.label" />
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  options?: Record<string, any>
  label?: string
}>(), {
  options: () => ({}),
  label: undefined,
})

const emit = defineEmits<{
  (e: 'action', value?: any): void
}>()
</script>
