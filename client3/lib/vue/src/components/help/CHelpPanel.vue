<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="offcanvas-backdrop fade show"
      @click="close"
    />
    <div
      class="offcanvas offcanvas-end"
      :class="{ show: open }"
      tabindex="-1"
      :style="open ? { visibility: 'visible' } : undefined"
      role="dialog"
      aria-modal="true"
    >
      <div class="offcanvas-header border-bottom">
        <h5 class="offcanvas-title mb-0">
          {{ title || labels.title }}
        </h5>
        <button
          type="button"
          class="btn-close"
          :aria-label="labels.close"
          @click="close"
        />
      </div>
      <div class="offcanvas-body">
        <section
          v-if="hasApp"
          class="mb-4"
        >
          <h6
            v-if="hasProduct"
            class="text-primary text-uppercase small fw-bold mb-2"
          >
            {{ labels.app }}
          </h6>
          <p
            v-if="description"
            class="text-muted"
          >
            {{ description }}
          </p>
          <div
            v-if="bodyHtml"
            class="c-help-md"
            v-html="bodyHtml"
          />
        </section>

        <section v-if="hasProduct">
          <h6
            v-if="hasApp"
            class="text-primary text-uppercase small fw-bold mb-2"
          >
            {{ labels.product }}
          </h6>
          <p
            v-if="productHint && !hasApp"
            class="text-muted"
          >
            {{ productHint }}
          </p>
          <div
            v-if="productHtml"
            class="c-help-md"
            v-html="productHtml"
          />
        </section>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'

defineOptions({ inheritAttrs: false })

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  description: { type: String, default: '' },
  bodyHtml: { type: String, default: '' },
  productHint: { type: String, default: '' },
  productHtml: { type: String, default: '' },
  labels: {
    type: Object,
    default: () => ({
      title: 'Help',
      app: 'About this',
      product: 'How to use',
      close: 'Close',
    }),
  },
})

const emit = defineEmits(['update:open', 'close'])

const hasApp = computed(() => !!(props.description || props.bodyHtml))
const hasProduct = computed(() => !!(props.productHint || props.productHtml))

function close () {
  emit('update:open', false)
  emit('close')
}
</script>

<style scoped>
.offcanvas-end {
  width: min(28rem, 100vw);
}
.c-help-md :deep(p:last-child) {
  margin-bottom: 0;
}
.c-help-md :deep(ul),
.c-help-md :deep(ol) {
  padding-left: 1.25rem;
}
.c-help-md :deep(code) {
  font-size: 0.875em;
}
</style>
