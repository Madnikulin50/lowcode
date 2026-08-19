<template>
  <div
    class="card shadow-sm h-100 ns-card"
    :class="{ 'ns-card-enabled': isEnabled, 'disabled': !isEnabled }"
    @mouseover="hovered = true"
    @mouseleave="hovered = false"
  >
    <div class="card-body d-flex flex-column align-items-center text-center pt-4 pb-3">
      <div
        class="ns-icon mb-3"
        :style="iconStyle"
      >
        <img
          v-if="imageSrc && !imageFailed"
          :src="imageSrc"
          :alt="namespace.name"
          class="ns-icon-img"
          @error="imageFailed = true"
        >
        <font-awesome-icon
          v-else
          :icon="['fas', faIcon]"
          class="ns-icon-fa"
        />
      </div>
      <h5
        :data-test-id="namespace.name"
        class="ns-title mb-1"
      >
        {{ namespace.name }}
      </h5>
      <p
        v-if="namespace.meta?.subtitle"
        class="d-inline-block my-0 text-secondary small"
      >
        {{ namespace.meta.subtitle }}
      </p>
      <p
        v-if="namespace.meta?.description"
        class="overflow-auto mt-2 mb-0 px-1"
      >
        <small>{{ namespace.meta.description }}</small>
      </p>
      <router-link
        v-if="isEnabled"
        :to="{ name: 'pages', params: { slug: (namespace.slug || namespace.namespaceID) } }"
        :data-test-id="`link-visit-namespace-${namespace.slug}`"
        :aria-label="$t('visit') + ' ' + namespace.name"
        class="stretched-link"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'namespace' } })
import { ref, computed, watch } from 'vue'
import { namespaceIconName, namespaceImageSrc, namespacePalette } from './namespaceIcon'

const props = defineProps({
  namespace: { type: Object, required: true },
})

const hovered = ref(false)
const imageFailed = ref(false)

const isEnabled = computed(() => !!props.namespace.enabled)
const imageSrc = computed(() => namespaceImageSrc(props.namespace))
const faIcon = computed(() => namespaceIconName(props.namespace))
const palette = computed(() => namespacePalette(props.namespace))

const iconStyle = computed(() => {
  if (imageSrc.value && !imageFailed.value) {
    return {
      background: '#f8fafc',
      boxShadow: 'inset 0 0 0 1px rgba(15, 23, 42, 0.06)',
    }
  }
  const [from, to] = palette.value
  return {
    background: `linear-gradient(145deg, ${from} 0%, ${to} 100%)`,
  }
})

watch(imageSrc, () => { imageFailed.value = false })
</script>

<style lang="scss" scoped>
.ns-card {
  min-height: 13rem;
  border: 0;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.ns-card-enabled:hover {
  transform: translateY(-3px);
  box-shadow: 0 0.75rem 1.5rem rgba(15, 23, 42, 0.12) !important;
}

.ns-icon {
  width: 88px;
  height: 88px;
  border-radius: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 10px 22px rgba(15, 23, 42, 0.18);
}

.ns-icon-fa {
  font-size: 2.15rem;
  color: #fff;
  filter: drop-shadow(0 1px 1px rgba(15, 23, 42, 0.25));
}

.ns-icon-img {
  max-width: 72%;
  max-height: 72%;
  object-fit: contain;
}

.ns-title {
  font-weight: 600;
  letter-spacing: 0.01em;
}

.disabled {
  opacity: 0.6;
}
</style>
