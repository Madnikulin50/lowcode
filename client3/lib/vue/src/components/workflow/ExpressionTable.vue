<template>
  <div class="table-responsive">
    <div class="row header ps-4">
      <div
        v-for="(field, index) in fields"
        :key="index"
        :class="`py-2 ${field.thClass}`"
      >
        <label>{{ field.label }}</label>
      </div>
    </div>

    <draggable
            item-key="id"
      :list="items"
      handle=".grab"
      @end="window.dispatchEvent(new CustomEvent('change-detected'))"
    >
      <template #item="{ element, index }">
        <div>
        <div
          class="row d-flex justify-content-between align-items-center pointer expr-item"
          @click="element._showDetails = !element._showDetails"
        >
          <div class="p-2 grab">
            <font-awesome-icon :icon="['fas', 'bars']" class="text-secondary" />
          </div>

          <div
            v-for="(field, i) in fields"
            :key="i"
            class="text-truncate p-2"
          >
            <div
              v-if="field.key === 'expr'"
              class="d-flex justify-content-between align-items-center"
            >
              <samp class="text-truncate">{{ element[field.key] }}</samp>

              <c-input-confirm
                v-if="element._showDetails"
                show-icon
                @confirmed="$emit('remove', index)"
              />
            </div>

            <var v-else class="">
              {{ field.formatter ? field.formatter(element) : element[field.key] }}
            </var>
          </div>
        </div>

        <transition name="fade">
          <div
            v-if="element._showDetails"
            class="mb-3 px-3"
          >
            <div class="arrow-up" />

            <div class="card bg-light">
              <div class="card-body px-4 pb-3">
                <div class="mb-3">
                  <label class="text-primary form-label">Target</label>
                  <input
                    class="form-control"
                    v-model="element.target"
                    placeholder="Target"
                    @input="window.dispatchEvent(new CustomEvent('change-detected'))"
                  />
                </div>

                <div class="mb-3">
                  <label class="text-primary form-label">{{ getTypeDescription(element.type) }}</label>
                  <c-input-select
                    v-model="element.type"
                    :options="types"
                    :get-option-key="getOptionKey"
                    :clearable="false"
                    :filter="varFilter"
                    append-to-body
                    @input="window.dispatchEvent(new CustomEvent('change-detected'))"
                  />
                </div>

                <div class="mb-0">
                  <expression-editor
                    v-model="element[valueField]"
                    show-line-numbers
                    @open="$emit('open-editor', index)"
                    @input="window.dispatchEvent(new CustomEvent('change-detected'))"
                  />
                </div>
              </div>
            </div>
          </div>
        </transition>
        </div>
      </template>
    </draggable>
  </div>
</template>

<script setup>
import { defineProps, defineEmits } from 'vue'
import ExpressionEditor from './ExpressionEditor.vue'
import { objectSearchMaker } from './lib/filter'
import draggable from 'vuedraggable'

const props = defineProps({
  valueField: { type: String, required: true },
  items: { type: Array, required: true },
  fields: { type: Array, required: true },
  types: { type: Array, required: true },
})

defineEmits(['remove', 'open-editor'])

const varFilter = objectSearchMaker('text')

function getTypeDescription(type) {
  const typeDescriptions = {
    ID: 'Make sure to provide the ID in double quotes if you\'re using a literal value. Example "123"',
  }
  return typeDescriptions[type]
}

function getOptionKey(type) {
  return type
}
</script>

<style lang="scss" scoped>
.header {
  background-color: var(--light);

  label {
    margin: 0;
  }
}

.table-responsive {
  overflow: hidden;
}

.expr-item:hover {
  background-color: var(--light);

  .grab > * {
    color: var(--secondary) !important;
  }
}
</style>
