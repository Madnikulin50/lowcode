<template>
  <div>
    <h5
      v-if="!root"
      class="mb-0"
    >
      {{ group }}
    </h5>

    <div
      v-if="subgroups.length > 0 || components.length > 0"
    >
      <div
        v-for="(cmp, i) in components"
        :key="i"
        class="component ms-2"
        @click="emit('select', cmp)"
      >
        {{ cmp.name || cmp.component.name || 'Untitled' }}
        <span
          v-if="cmp.wip"
          class="badge bg-warning float-end"
        >
          wip
        </span>
      </div>
      <ComponentList
        v-for="(g) in subgroups"
        :key="g"
        :catalogue="catalogue"
        :path="[...path, g]"
        class="my-3"
        @select="emit('select', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ExtractComponents, ExtractSubgroups } from './helpers.ts'

const props = defineProps<{
  catalogue: Record<string, any>
  path?: string[]
}>()

const emit = defineEmits<{
  (e: 'select', value: any): void
}>()

const root = computed(() => props.path?.length === 0)

const group = computed(() => root.value ? undefined : props.path?.[props.path.length - 1])

const subgroups = computed(() => ExtractSubgroups(props.catalogue, ...(props.path || [])))

const components = computed(() => ExtractComponents(props.catalogue, ...(props.path || [])))
</script>

<style lang="scss" scoped>
.component {
  cursor: pointer;
}
</style>
