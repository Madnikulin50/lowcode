<template>
  <div
    class="card shadow-sm pt-3 h-100"
    :class="{ 'shadow': hovered && isEnabled, 'namespace-item': isEnabled, 'disabled': !isEnabled }"
    @mouseover="hovered = true"
    @mouseleave="hovered = false"
  >
    <div
      class="circled-avatar d-flex align-items-center justify-content-center m-auto"
      :class="[namespace.meta.logoEnabled ? 'p-2' : 'bg-light p-3']"
    >
      <img
        v-if="namespace.meta.logoEnabled"
        :src="logo"
        :alt="namespace.name"
        class="mw-100 mh-100 img-fluid"
      />
      <h1
        v-else
        class="ns-initial m-auto text-uppercase text-secondary"
      >
        {{ namespace.initials }}
      </h1>
    </div>
    <div class="card-body mw-100 text-center pb-3">
      <div
        class="d-flex align-items-baseline"
        :class="{ 'h-100': !namespace.meta.description }"
      >
        <div class="d-flex flex-column justify-content-center w-100">
          <h5
            :data-test-id="namespace.name"
            class="mt-2"
          >
            {{ namespace.name }}
          </h5>
          <p
            v-if="namespace.meta.subtitle"
            class="d-inline-block my-1 text-secondary"
          >
            {{ namespace.meta.subtitle }}
          </p>
        </div>
      </div>
      <p
        v-if="namespace.meta.description"
        class="overflow-auto"
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
import { ref, computed } from 'vue'
import { useSettings } from 'corteza-lib/vue/dist'

const props = defineProps({
  namespace: { type: Object, required: true },
})

const { $Settings } = useSettings()

const processing = ref(false)
const hovered = ref(undefined)
const logoAttachment = ref(undefined)

const isEnabled = computed(() => !!props.namespace.enabled)
const canEdit = computed(() => !!props.namespace.canUpdateNamespace)
const showFooter = computed(() => isEnabled.value || canEdit.value)
const logo = computed(() => props.namespace.meta.logo || $Settings.attachment('ui.mainLogo'))
</script>

<style lang="scss" scoped>
$avatar-size: 120px;
$disabled-opacity: 0.6;

.namespace-item {
  min-height: 13rem;

  &:hover {
    transition: all 0.2s ease;
    top: -1px;
  }
}

.circled-avatar {
  width: $avatar-size;
  height: $avatar-size;
  border-radius: 50%;
}

.disabled {
  opacity: $disabled-opacity;
}
</style>
