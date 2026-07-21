<template>
  <div class="card shadow-sm" data-test-id="card-sens-lvl-info">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <form @submit.prevent="$emit('submit', sensitivityLevel)">
      <div class="row g-3 p-3">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('name') }}</label>
            <input
              v-model="sensitivityLevel.meta.name"
              type="text"
              class="form-control"
              :class="{ 'is-invalid': nameState === false }"
              data-test-id="input-name"
              required
            >
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('handle.label') }}</label>
            <input
              v-model="sensitivityLevel.handle"
              type="text"
              class="form-control"
              :class="{ 'is-invalid': handleState === false }"
              :placeholder="$t('handle.placeholder')"
            >
            <div v-if="handleState === false" class="invalid-feedback">{{ $t('handle.invalid-characters') }}</div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('level', sensitivityLevel) }}</label>
            <input
              v-model="sensitivityLevel.level"
              type="range"
              class="form-range"
              min="1"
              max="10"
            >
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('description') }}</label>
            <textarea v-model="sensitivityLevel.meta.description" class="form-control" />
          </div>
        </div>
      </div>

      <c-system-fields :resource="sensitivityLevel" />

      <input
        type="submit"
        class="d-none"
        :disabled="saveDisabled"
      >
    </form>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="sensitivityLevel && sensitivityLevel.sensitivityLevelID"
        :text="getDeleteStatus"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit', sensitivityLevel)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { handle } from 'corteza-lib/vue/dist'

const { t } = useI18n()

const props = defineProps({
  sensitivityLevel: { type: Object, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, required: true },
})

defineEmits(['submit', 'delete'])

const fresh = computed(() => !props.sensitivityLevel.sensitivityLevelID || props.sensitivityLevel.sensitivityLevelID === NoID)
const editable = computed(() => fresh.value ? props.canCreate : true)
const nameState = computed(() => props.sensitivityLevel.meta.name ? null : false)
const handleState = computed(() => handle.handleState(props.sensitivityLevel.handle))
const saveDisabled = computed(() => !editable.value || [nameState.value, handleState.value].includes(false))
const getDeleteStatus = computed(() => props.sensitivityLevel.deletedAt ? t('undelete') : t('delete'))
</script>