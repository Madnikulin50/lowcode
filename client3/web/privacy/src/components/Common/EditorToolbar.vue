<template>
  <div class="container-fluid bg-white shadow-sm border-top p-3">
    <div class="row align-items-center g-0">
      <div class="col">
        <router-link
          v-if="backLink"
          data-test-id="button-back"
          :to="backLink"
          class="btn btn-link btn-lg d-flex align-items-center text-dark text-decoration-none gap-1"
        >
          <font-awesome-icon :icon="['fas', 'chevron-left']" class="back-icon" />
          {{ t('label.back') }}
        </router-link>

        <slot name="left" />
      </div>

      <div class="col d-flex justify-content-center">
        <slot name="middle" />
      </div>

      <div class="col d-flex justify-content-end">
        <template v-if="deleteShow">
          <c-input-confirm
            v-if="deleteConfirm"
            :disabled="deleteDisabled || processing"
            :processing="processingDelete"
            :text="deleteLabel"
            variant="danger"
            size="lg"
            size-confirm="lg"
            class="ms-1"
            @confirmed="emit('delete')"
          />

          <button
            v-else
            :data-test-id="buttonLabelCypressId(deleteLabel)"
            :disabled="deleteDisabled || processing"
            class="btn btn-danger btn-lg ms-1"
            @click="emit('delete')"
          >
            {{ deleteLabel }}
          </button>
        </template>

        <c-button-submit
          v-if="submitShow"
          :data-test-id="buttonLabelCypressId(submitLabel)"
          :disabled="submitDisabled"
          :processing="processing"
          :text="submitLabel"
          size="lg"
          class="ms-1"
          @submit="emit('submit')"
        />

        <slot />
      </div>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

defineProps({
  processing: { type: Boolean, default: false },
  processingDelete: { type: Boolean, default: false },
  backLink: { type: Object, default: () => ({ name: 'root' }) },
  deleteShow: { type: Boolean, default: false },
  deleteDisabled: { type: Boolean, default: false },
  deleteConfirm: { type: Boolean, default: true },
  deleteLabel: { type: String, default: '' },
  submitShow: { type: Boolean, default: false },
  submitDisabled: { type: Boolean, default: false },
  submitLabel: { type: String, default: '' },
})

const emit = defineEmits(['delete', 'submit'])

const { t } = useI18n()

function buttonLabelCypressId (label) {
  return `button-${label.toLowerCase().split(' ').join('-')}`
}
</script>

<style lang="scss" scoped>
.back {
  &:hover {
    text-decoration: none;

    .back-icon {
      transition: transform 0.3s ease-out;
      transform: translateX(-4px);
    }
  }
}
</style>