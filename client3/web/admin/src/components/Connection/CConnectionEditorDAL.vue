<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div v-if="issues.length" class="row">
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('system.connections.editor.dal.connectivity-issues') }}</label>
          <div v-for="issue in issues" :key="issue.issue" class="alert alert-danger" role="alert">
            {{ issue.issue }}
          </div>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('system.connections.editor.dal.form.model-ident.label') }}</label>
          <small class="form-text text-muted">{{ modelIdentDescription }}</small>
          <input
            v-model="dal.modelIdent"
            type="text"
            class="form-control"
            :disabled="disabled"
            :placeholder="$t('system.connections.editor.dal.form.model-ident.placeholder')"
          >
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('system.connections.editor.dal.form.type.label') }}</label>
          <small class="form-text text-muted">{{ $t('system.connections.editor.dal.form.type.description') }}</small>
          <input
            v-model="dal.type"
            type="text"
            class="form-control"
            :disabled="disabled"
            :placeholder="$t('system.connections.editor.dal.form.type.placeholder')"
          >
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('system.connections.editor.dal.form.params.label') }}</label>
          <small class="form-text text-muted">{{ $t('system.connections.editor.dal.form.params.description') }}</small>
          <textarea
            v-model="paramsJson"
            class="form-control"
            :class="paramsJsonEditorClass"
            :disabled="disabled"
            :placeholder="$t('system.connections.editor.dal.form.params.placeholder')"
            rows="5"
            @blur="processParamsJSON"
          />
        </div>
      </div>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        :disabled="disabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit')"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.connections', keyPrefix: 'editor.dal' } })
import { ref, watch, computed, inject } from 'vue'

defineProps({
  disabled: { type: Boolean, default: false },
  dal: { type: Object, required: true },
  issues: { type: Array, default: () => ([]) },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
})

defineEmits(['submit'])

const $t = inject('$t') || ((k) => k)
const modelIdentDescription = computed(() => $t('form.model-ident.description', { interpolation: { prefix: '{{{' , suffix: '}}}' } }))

const paramsJson = ref('')
const paramsJsonEditorClass = ref('')

watch(() => props.dal, (dal) => {
  paramsJson.value = JSON.stringify(dal.params || { dsn: '' }, null, 2)
}, { deep: true, immediate: true })

function processParamsJSON() {
  paramsJsonEditorClass.value = ''

  try {
    const json = JSON.parse(paramsJson.value)
    if (typeof json !== 'object') {
      throw new Error('JSON is not an object')
    }

    if (!props.dal.params) {
      props.dal.params = {}
    }
    for (const key in json) {
      if (Object.prototype.hasOwnProperty.call(json, key)) {
        props.dal.params[key] = json[key]
      }
    }
  } catch (e) {
    paramsJsonEditorClass.value = 'border-danger'
  }
}
</script>