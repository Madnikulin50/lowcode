<template>
  <div class="container-fluid p-0">
    <div class="row">
      <div class="col-3">
        <ul class="list-group list-group-flush">
          <draggable
            item-key="id"
            :list="items"
            :disabled="!draggable"
            group="items"
            handle=".grab"
          >
            <template #item="{ element, index }">
              <li
                :key="index"
                class="list-group-item item d-flex align-items-center justify-content-between"
                :class="{ 'active': currentIndex ? currentIndex === index : index === 0 }"
                style="cursor: pointer;"
                @click="$emit('select', index)"
              >
                <slot
                  name="label"
                  :item="element"
                />
              </li>
            </template>
          </draggable>

          <li
            class="list-group-item text-primary rounded-top"
            :class="{ 'border-top-0': items.length }"
            style="cursor: pointer;"
            @click="$emit('add')"
          >
            <font-awesome-icon
              :icon="['fas', 'plus']"
              size="sm"
              class="me-1"
            />
            {{ $t('label.add') }}
          </li>
        </ul>
      </div>
      <div
        v-if="currentIndex !== undefined"
        class="col-9"
      >
        <slot
          name="configurator"
        />

        <c-input-confirm
          variant="danger"
          size="lg"
          size-confirm="lg"
          :borderless="false"
          :text="$t('label.delete')"
          class="d-flex"
          @confirmed="$emit('delete')"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import Draggable from 'vuedraggable'

defineProps({
  items: {
    type: Array,
    required: true,
  },
  currentIndex: {
    type: Number,
    default: undefined,
  },
  draggable: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['select', 'add', 'delete'])
</script>

<style lang="scss">
.item {
  .grab {
    opacity: 0;
  }

  &:hover {
    .grab {
      opacity: 1;
    }
  }
}
</style>
