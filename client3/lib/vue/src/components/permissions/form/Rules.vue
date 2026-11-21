<template>
  <div
    data-test-id="role-permissions-list"
  >
    <Rule
      v-for="(p, i) in rules"
      :key="p.resource + p.operation"
      v-bind="p"
      :class="{ 'mt-4': i > 0 }"
      @update="onUpdate"
    />
  </div>
</template>

<script setup lang="ts">
import Rule from './Rule.vue'

const props = defineProps<{
  rules: any[]
}>()

const emit = defineEmits<{
  (e: 'update:rules', value: any[]): void
}>()

function onUpdate({ resource, operation, access }: { resource: string; operation: string; access: string }) {
  const rr = props.rules
  const ri = rr.findIndex((r: any) => r.resource === resource && r.operation === operation)
  if (ri > -1) {
    rr[ri].access = access
    emit('update:rules', rr)
  }
}
</script>
