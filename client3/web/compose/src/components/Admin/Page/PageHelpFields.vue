<template>
  <div class="page-help-fields">
    <h6
      v-if="showHeading"
      class="text-primary fw-semibold mb-1"
    >
      {{ $t('help.section') }}
    </h6>
    <p class="text-muted small mb-3">
      {{ $t('help.description') }}
    </p>

    <div
      v-if="showDescription"
      class="mb-3"
    >
      <label class="form-label text-primary">
        {{ $t('label.description') }}
      </label>
      <textarea
        :value="page.description"
        data-test-id="input-page-help-description"
        class="form-control"
        :placeholder="$t('edit.pageDescription')"
        rows="3"
        @input="page.description = $event.target.value"
      />
    </div>

    <div class="mb-0">
      <label class="form-label text-primary">
        {{ $t('help.label') }}
      </label>
      <div class="d-flex gap-2 align-items-start">
        <textarea
          :value="helpText"
          data-test-id="input-page-help"
          class="form-control"
          :placeholder="$t('help.placeholder')"
          rows="8"
          @input="setHelp($event.target.value)"
        />
        <slot name="append" />
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'page' } })
import { computed, watch } from 'vue'
import { hydratePageDocs } from '../../../help/appDocs'

const props = defineProps({
  page: { type: Object, required: true },
  namespace: { type: Object, default: null },
  showHeading: { type: Boolean, default: true },
  showDescription: { type: Boolean, default: false },
})

function ensureConfig () {
  if (!props.page.config) {
    props.page.config = { navItem: { expanded: false, icon: { type: '', src: '' } }, prompt: '', help: '' }
  }
  if (props.page.config.help == null) {
    props.page.config.help = ''
  }
  return props.page.config
}

watch(
  () => [props.page?.pageID, props.page?.handle, props.namespace?.slug],
  () => { hydratePageDocs(props.namespace, props.page) },
  { immediate: true },
)

const helpText = computed(() => props.page?.config?.help || '')

function setHelp (value) {
  ensureConfig().help = value
}
</script>
