<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('federation.nodes.editor.info.title') }}</h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="$emit('submit', node)">
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('federation.nodes.editor.info.name') }}</label>
              <input
                v-model="node.name"
                type="text"
                class="form-control"
                :class="{ 'is-invalid': nameState === false }"
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('federation.nodes.editor.info.url') }}</label>
              <input
                v-model="node.baseURL"
                type="url"
                class="form-control"
                :class="{ 'is-invalid': urlState === false }"
                placeholder="https://example.com/federation"
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('federation.nodes.editor.info.email') }}</label>
              <input
                v-model="node.contact"
                type="email"
                class="form-control"
                placeholder="email@example.com"
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div v-if="node.status" class="mb-3">
              <label class="form-label text-primary">{{ $t('federation.nodes.editor.info.status') }}</label>
              <p class="form-control-plaintext">{{ node.status }}</p>
            </div>
          </div>
        </div>
  
        <c-system-fields :resource="node" />
  
        <input
          type="submit"
          class="d-none"
          :disabled="saveDisabled"
        >
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="node && node.nodeID"
        variant="danger"
        size="md"
        :text="getDeleteStatus"
        @confirmed="$emit('delete')"
      />

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('admin.general.label.submit')"
        class="ms-auto"
        @submit="$emit('submit', node)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'

const { t } = useI18n()

const props = defineProps({
  node: { type: Object, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, value: false },
})

defineEmits(['submit', 'delete'])

const fresh = computed(() => {
  return !props.node.nodeID || props.node.nodeID === NoID
})

const editable = computed(() => {
  return fresh.value ? props.canCreate : props.node.canManageNode
})

const saveDisabled = computed(() => {
  return !editable.value || [nameState.value, urlState.value].includes(false)
})

const nameState = computed(() => {
  const { name } = props.node
  return name ? null : false
})

const urlState = computed(() => {
  const { baseURL = '' } = props.node
  return /https?:\/\/*.*\/federation/gm.test(baseURL) ? null : false
})

const getDeleteStatus = computed(() => {
  return props.node.deletedAt ? t('undelete') : t('delete')
})
</script>