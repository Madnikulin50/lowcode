<template>
  <div class="card shadow-sm h-100" data-test-id="card-template-toolbox">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body">
      <div v-for="sec in sections" :key="sec.key">
        <button
          type="button"
          class="btn btn-outline-secondary mb-2 w-100"
          :data-test-id="toolboxSectionLabelCypressId(sec.key)"
          @click="openSection(sec.key)"
        >
          {{ $t(sec.key) }}
        </button>
        <div v-if="expandedSections[sec.key]" class="pb-2 px-0">
          <div class="list-group list-group-flush">
            <button
              v-for="(opt, i) in sec.options"
              :key="opt.label + i"
              type="button"
              class="list-group-item list-group-item-action px-0 text-wrap"
              :data-test-id="toolboxOptionLabelCypressId(opt.label)"
              @click="opt.onClick || (() => {})"
            >
              {{ opt.label }}
              <button
                v-if="opt.copyValue"
                type="button"
                class="btn btn-link pe-0 float-end"
                data-test-id="button-copy"
                @click="copyToCb(opt.copyValue())"
              >
                <font-awesome-icon :icon="['far', 'copy']" />
              </button>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import copy from 'copy-to-clipboard'

const { t } = useI18n()

const props = defineProps({
  template: { type: Object, required: true, default: () => ({}) },
  partials: { type: Array, required: false, default: () => [] },
})

const expandedSections = ref({})

const sections = computed(() => {
  const partialsItems = props.partials.map(p => ({
    label: p.meta.short || p.handle,
    copyValue: () => `{{template "${p.handle}" }}`,
  }))

  const result = []
  if (partialsItems.length) {
    result.push({ key: 'partials', options: partialsItems })
  }

  result.push({
    key: 'snippets.label',
    options: [
      { label: t('snippets.interpolate'), copyValue: () => '{{.parameter}}' },
      { label: t('snippets.iterator'), copyValue: () => '{{range $index, $element := .ListOfItems}}\n\n{{end}}' },
      { label: t('snippets.funcCall'), copyValue: () => '{{funcName param1 param2 paramN}}' },
    ],
  },
  {
    key: 'samples.label',
    options: [
      {
        label: t('samples.defaultHTML'),
        copyValue: () => `<!DOCTYPE html>
<html>
<head>
  <meta charset='utf-8'>
  <meta http-equiv='X-UA-Compatible' content='IE=edge'>
  <title>Title</title>
  <meta name='viewport' content='width=device-width, initial-scale=1'>
</head>
<body>
  <h1>Hello, world!</h1>
</body>
</html>`,
      },
    ],
  })

  return result
})

function openSection(sec) {
  expandedSections.value[sec] = !expandedSections.value[sec]
}

function copyToCb(value) {
  copy(value)
}

function toolboxSectionLabelCypressId(section) {
  return `button-${t(section).toLowerCase()}`
}

function toolboxOptionLabelCypressId(label) {
  return label.toLowerCase().split(' ').join('-')
}
</script>