<template>
  <div>
    <div class="mb-3">
      <div class="form-check">
        <input id="presetWithAuthenticated" v-model="f.options.presetWithAuthenticated" type="checkbox" class="form-check-input" />
        <label class="form-check-label" for="presetWithAuthenticated">{{ t('kind.user.presetWithCurrentUser') }}</label>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('kind.user.roles.label') }}</label>
      <span v-if="preloadingRoles" class="spinner-border spinner-border-sm"></span>
      <CInputRole
        v-else
        v-model="currentRoles"
        :placeholder="t('kind.user.roles.placeholder')"
        multiple
        @input="onRolesInput"
      />
    </div>

    <template v-if="f.isMulti">
      <div class="mb-3">
        <label class="form-label text-primary">{{ t('kind.select.optionType.label') }}</label>
        <div class="btn-group" data-bs-toggle="buttons">
          <label
            v-for="opt in selectOptions"
            :key="opt.value"
            class="btn btn-outline-primary btn-sm"
            :class="{ active: f.options.selectType === opt.value }"
          >
            <input
              type="radio"
              class="btn-check"
              :value="opt.value"
              :checked="f.options.selectType === opt.value"
              autocomplete="off"
              @change="onSelectTypeChange(opt.value)"
            />
            {{ opt.text }}
          </label>
        </div>
      </div>

      <div v-if="shouldAllowDuplicates" class="form-check mb-3">
        <input
          id="allowDuplicates"
          v-model="f.options.isUniqueMultiValue"
          :true-value="false"
          :false-value="true"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label" for="allowDuplicates">{{ t('kind.select.allow-duplicates') }}</label>
      </div>
    </template>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { ref, computed, inject, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'
import { components } from 'corteza-lib/vue/dist'
const { CInputRole } = components

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f, isNew, hasData } = useConfiguratorBase(props, emit)

const $SystemAPI = inject('$SystemAPI')

const selectOptions = ref([
  { text: t('kind.select.optionType.default'), value: 'default', allowDuplicates: true },
  { text: t('kind.select.optionType.multiple'), value: 'multiple' },
  { text: t('kind.select.optionType.each'), value: 'each', allowDuplicates: true },
])

const preloadingRoles = ref(false)
const currentRoles = ref([])

const shouldAllowDuplicates = computed(() => {
  if (!f.value.isMulti) return false
  const { allowDuplicates } = selectOptions.value.find(({ value }) => value === f.value.options.selectType) || {}
  return !!allowDuplicates
})

onMounted(() => {
  if (f.value.options.roles && f.value.options.roles.length) {
    preloadingRoles.value = true
    Promise.all(f.value.options.roles.map(roleID => {
      return ($SystemAPI.value || {}).roleRead ? $SystemAPI.value.roleRead({ roleID }).then(role => {
        currentRoles.value.push(role)
      }) : Promise.resolve()
    })).finally(() => {
      preloadingRoles.value = false
    })
  }
})

onBeforeUnmount(() => {
  setDefaultValues()
})

function onRolesInput ($event) {
  f.value.options.roles = $event.map(r => r.roleID)
}

function onSelectTypeChange (value) {
  f.value.options.selectType = value
  updateIsUniqueMultiValue(value)
}

function updateIsUniqueMultiValue (value) {
  const { allowDuplicates = false } = selectOptions.value.find(({ value: v }) => v === value) || {}
  if (!allowDuplicates) {
    f.value.options.isUniqueMultiValue = true
  }
}

function setDefaultValues () {
  selectOptions.value = []
}
</script>
