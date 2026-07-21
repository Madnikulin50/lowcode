<template>
  <div>
    <div v-if="singleInput" class="mb-2">
      <slot name="single" />
    </div>

    <draggable
            item-key="id"
      v-if="showList"
      v-model="val"
      handle=".handle"
    >
      <template #item="{ element, index }">
        <div
          :key="index"
          class="d-flex w-100 align-items-center mb-2 px-1"
        >
          <font-awesome-icon
            :icon="['fas', 'bars']"
            :title="t('tooltip.dragAndDrop')"
            class="handle text-secondary me-3"
          />

          <div class="flex-grow-1">
            <slot :index="index" />
          </div>

          <font-awesome-icon
            v-if="removable"
            :icon="['fas', 'times']"
            class="pointer text-danger ms-3"
            @click="removeValue(index)"
          />
        </div>
      </template>
    </draggable>

    <errors :errors="errors" />

    <button
      v-if="!singleInput"
      class="btn btn-primary btn-sm"
      :class="{ 'mt-2': val.length }"
      @click="addValue()"
    >
      + {{ t('label.addValue') }}
    </button>
  </div>
</template>

<script setup>
import { computed, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import errors from '../errors'
import draggable from 'vuedraggable'
import { validator } from 'corteza-lib/js/dist'

const props = defineProps({
  value: { type: Array, required: true, default: () => [] },
  removable: { type: Boolean, default: true },
  singleInput: { type: Boolean, default: false },
  showList: { type: Boolean, default: true },
  errors: { type: validator.Validated, required: true },
  defaultValue: { type: undefined, default: undefined },
})

const emit = defineEmits(['update:value'])

const { t } = useI18n({ useScope: 'global', messages: {} })

const val = computed({
  get () { return props.value },
  set (v) { emit('update.value', v) },
})

function addValue () {
  const arr = [...props.value]
  arr.push(props.defaultValue)
  emit('update.value', arr)
}

function removeValue (index) {
  if (index > -1) {
    const arr = [...props.value]
    arr.splice(index, 1)
    emit('update.value', arr)
  }
}
</script>

<style lang="scss" scoped>
.handle {
  cursor: grab;
}
.pointer {
  cursor: pointer;
}
</style>
