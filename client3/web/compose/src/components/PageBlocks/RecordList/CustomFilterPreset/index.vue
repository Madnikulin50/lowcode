<template>
  <div v-if="showModal" class="modal fade show d-block" tabindex="-1">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ $t('recordList.filterPresets.saveFilterAsPreset') }}</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" @click="onModalHide"></button>
        </div>
        <div class="modal-body p-0">
          <div class="card pt-0 border-0">
            <div class="card-body">
              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('recordList.filterPresets.filterName') }}</label>
                <input v-model="filterName" class="form-control" />
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer d-flex w-100 align-items-center justify-content-between">
          <button class="btn btn-outline-secondary" @click="onModalHide">{{ $t('label.cancel') }}</button>
          <div>
            <button class="btn btn-primary" :disabled="!filterName" @click="onSave">{{ $t('label.save') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div v-if="showModal" class="modal-backdrop fade show" @click="onModalHide" />
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  visible: { type: Boolean, default: false },
})

const emit = defineEmits(['close', 'save'])

const showModal = ref(false)
const filterName = ref('')

watch(() => props.visible, (val) => { showModal.value = val }, { immediate: true })

function onModalHide() {
  showModal.value = false
  filterName.value = ''
  emit('close')
}

function onSave() {
  emit('save', { name: filterName.value })
  onModalHide()
}
</script>
