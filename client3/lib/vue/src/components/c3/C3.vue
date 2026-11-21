<template>
  <div class="layout">
    <aside class="sidebar p-2">
      <h5 class="border-bottom">
        C3: Component Catalogue
      </h5>
      <ComponentList
        :catalogue="catalogue"
        @select="setCurrent($event)"
      />
    </aside>

    <main class="p-5">
      <component
        :is="current.component"
        v-if="current"
        v-bind="current.props"
      />
      <p
        v-else
        class="text-center"
      >
        Select a component from the C3 Catalogue and start hacking :)
      </p>
    </main>

    <div
      v-if="current"
      class="controls px-5 py-2 mt-2"
    >
      <div
        v-for="(cg, g) in controlGroups"
        :key="`control-group-${g}`"
        class="control-group me-2"
      >
        <h3>
          Controls
        </h3>
        <component
          :is="c.component"
          v-for="(c, i) in cg"
          :key="i"
          :model-value="c.value(current.props)"
          v-bind="c.props"
          @update:model-value="c.update(current.props, $event)"
        />
      </div>

      <div
        v-if="current.scenarios"
        class="control-group float-end"
      >
        <h3>
          Pre-set controls
        </h3>
        <ul class="ps-0">
          <li
            v-for="(s, i) in current.scenarios"
            :key="i"
            class="list-unstyled scenario"
            @click="setScenario(s)"
          >
            {{ s.label }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import ComponentList from './ComponentList.vue'

defineProps<{
  catalogue: Record<string, any>
}>()

const current = ref<any>(undefined)

const controlGroups = computed(() => {
  if (!current.value?.controls?.length) return []
  if (Array.isArray(current.value.controls[0])) {
    return current.value.controls
  }
  return [current.value.controls]
})

function setCurrent(component: any) {
  current.value = { props: {}, ...component }
  setScenario(current.value)
}

function setScenario({ props = {}, controls = [] }: any) {
  const apply = (c: any, p: any) => c.update(p, c.value(p) || null)
  controls.forEach((c: any) => {
    if (Array.isArray(c)) {
      c.forEach((cc: any) => apply(cc, props))
    } else {
      apply(c, props)
    }
  })
  current.value.props = props
}
</script>

<style lang="scss">
.layout {
  height: 100vh;
  width: 100vw;

  display: grid;
  grid-template-rows: auto 400px;
  grid-template-columns: 300px auto;
  align-content: stretch;
  grid-template-areas:
    "side main"
    "side controls"
  ;

  aside {
    grid-area: side;
    overflow: auto;
  }

  main {
    grid-area: main;
    background-image: linear-gradient(
      135deg,
      #F9FAFB 21.43%,
      #FFFFFF 21.43%,
      #FFFFFF 50%,
      #F9FAFB 50%,
      #F9FAFB 71.43%,
      #FFFFFF 71.43%,
      #FFFFFF 100%
    );
    background-size: 35.00px 35.00px;
    overflow: auto;
  }

  .controls {
    grid-area: controls;
    overflow: auto;

    .control-group {
      float: left;
    }
  }

  .scenario {
    cursor: pointer;
  }
}
</style>
