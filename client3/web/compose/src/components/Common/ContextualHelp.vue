<template>
  <c-help-trigger
    v-if="hasAny"
    class="ms-1"
    icon-class="text-dark"
    v-bind="triggerProps"
  />
</template>

<script setup>
import { useHelp } from '../../composables/useHelp'

const { hasAny, triggerProps } = useHelp()
</script>
