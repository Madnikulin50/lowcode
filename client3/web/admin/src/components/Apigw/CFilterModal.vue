<template>
  <div
    class="modal fade"
    :class="{ show: visible }"
    :style="{ display: visible ? 'block' : 'none' }"
    tabindex="-1"
    @click.self="onHidden"
  >
    <div
      v-if="visible"
      class="modal-dialog modal-lg modal-dialog-centered"
    >
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            {{ (filter || {}).label }}
          </h5>
          <button
            type="button"
            class="btn-close"
            @click="onHidden"
          />
        </div>
        <div
          v-if="internalFilter"
          class="modal-body p-0"
        >
          <div class="card-body">
            <c-filter-params
              :filter="internalFilter"
            />
            <div class="form-check">
              <input
                :id="'filter-enabled-' + internalFilter.ref"
                v-model="internalFilter.enabled"
                type="checkbox"
                class="form-check-input"
                data-test-id="checkbox-filter-enable"
              >
              <label
                :for="'filter-enabled-' + internalFilter.ref"
                class="form-check-label"
              >{{ $t('filters.enabled') }}</label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button
            type="button"
            class="btn btn-outline-secondary"
            @click="onHidden"
          >
            {{ $t('label.cancel') }}
          </button>
          <button
            type="button"
            class="btn btn-primary"
            @click="onSave"
          >
            {{ $t('filters.modal.ok') }}
          </button>
        </div>
      </div>
    </div>
  </div>
  <div
    v-if="visible"
    class="modal-backdrop fade show"
  />
</template>

<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import CFilterParams from 'corteza-webapp-admin/src/components/Apigw/CFilterParams'

const { t: $t } = useI18n()

const props = defineProps({
  filter: {
    type: Object,
    default: () => ({}),
  },
  visible: {
    type: Boolean,
    required: false,
    default: false,
  },
})

const emit = defineEmits(['submit', 'reset'])

const internalFilter = ref(undefined)

watch(() => props.visible, (visible) => {
  if (visible) {
    internalFilter.value = {
      ...props.filter,
      params: props.filter.params.map(p => {
        let value = p.value || {}

        if (props.filter.ref === 'response') {
          if (p.type === 'header') {
            value = Object.entries(value).map(([name, v = []]) => ({ name, expr: v.join('') }))
          } else if (p.type === 'input') {
            value = { type: 'Any', expr: '', ...value }
          }
        }

        return { ...p, value }
      }),
    }
  } else {
    internalFilter.value = undefined
  }
})

function onSave () {
  const filter = {
    ...internalFilter.value,
    params: internalFilter.value.params.map(p => {
      if (props.filter.ref === 'response') {
        if (p.type === 'header') {
          p.value = p.value.reduce((obj, { name, expr = '' }) => {
            return { ...obj, [name]: [expr] }
          }, {})
        }
      }

      return p
      })
    }

  emit('submit', { ...filter, updated: true })
}

function onHidden () {
  internalFilter.value = undefined
  emit('reset')
}
</script>
