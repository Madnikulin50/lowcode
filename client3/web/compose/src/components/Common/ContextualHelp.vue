<template>
  <c-help-trigger
    v-if="help.hasAny"
    class="ms-1 contextual-help-trigger"
    icon-class="text-dark"
    button-class="justify-content-center"
    v-bind="help.triggerProps"
  />
</template>

<script setup>
import { useHelp } from '../../composables/useHelp'

const help = useHelp()
</script>

<style lang="scss">
// CHelpTrigger's button ships unsized (btn-link, p-0) — fine inline next to
// text, but here it sits in the topbar's icon row alongside theme/
// notifications/help/avatar, all sized off --topbar-height (see CTopbar.vue
// and CSearchButton.vue, which each define this same $nav-icon-size locally
// rather than sharing CTopbar's scoped class — a scoped class can't reach
// here anyway: the `class` passed to <c-help-trigger> lands on an element
// from CHelpTrigger's own template, which never carries *this* component's
// data-v- scope attribute, so a `scoped` block here can never match it —
// hence a plain global block, kept safe by the specific class name.
$nav-icon-size: calc(var(--topbar-height, 64px) - 24px);

.contextual-help-trigger button {
  width: $nav-icon-size;
  height: $nav-icon-size;
}
</style>
