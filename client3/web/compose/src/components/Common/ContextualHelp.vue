<template>
  <c-help-trigger
    v-if="help.hasAny"
    class="ms-1"
    icon-class="text-dark"
    v-bind="help.triggerProps"
  />
</template>

<script setup>
import { useHelp } from '../../composables/useHelp'

const help = useHelp()
</script>
