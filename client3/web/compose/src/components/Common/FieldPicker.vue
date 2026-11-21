<template>
  <div
    class="d-flex flex-column border border-light rounded p-2"
  >
    <c-item-picker
      v-model="selected"
      :options="options"
      :labels="{
        searchPlaceholder: $t('selector.search'),
        availableItems: $t('available-items'),
        selectAllItems: $t('selector.selectAll'),
        selectedItems: $t('selected-items'),
        unselectAllItems: $t('selector.unselectAll'),
        noItemsFound: $t('no-items-found'),
      }"
    >
      <template
        #default="{ field }"
      >
        <b class="cursor-default text-dark">
          <template
            v-if="field.label"
          >
            {{ field.label }} ({{ field.name }})
          </template>
          <template v-else>
            {{ field.name }}
          </template>
          <template v-if="field.isRequired">
            *
          </template>
        </b>
        <small
          v-if="field.isSystem"
          class="cursor-default ms-1 text-truncate"
        >
          {{ $t('selector.systemField') }}
        </small>
      </template>
    </c-item-picker>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
const { CItemPicker } = components

const { t } = useI18n()

const props = defineProps({
  module: {
    type: Object,
    required: true,
  },
  fields: {
    type: Array,
    required: true,
  },
  disabledTypes: {
    type: Array,
    default: () => [],
  },
  disableSystemFields: {
    type: Boolean,
    default: false,
  },
  disableSorting: {
    type: Boolean,
  },
  systemFields: {
    type: Array,
    default: null,
  },
  fieldSubset: {
    type: Array,
    default: null,
  },
})

const emit = defineEmits(['update:fields'])

const selected = computed({
  get () {
    return props.fields.map(({ name }) => name)
  },
  set (val) {
    const fields = val.map(s => {
      return (options.value.find(({ value }) => value === s) || {}).field
    }).filter(f => f)
    emit('update:fields', fields)
  },
})

const options = computed(() => {
  let mFields = [...(props.fieldSubset ? props.module.filterFields(props.fieldSubset) : props.module.fields)]

  if (props.disabledTypes.length > 0) {
    mFields = mFields.filter(({ kind }) => !props.disabledTypes.find(t => t === kind))
  }

  mFields = mFields.filter(({ canReadRecordValue }) => canReadRecordValue !== false)

  let sysFields = []

  if (props.disableSystemFields && mFields) {
    mFields = mFields.filter(({ isSystem }) => !isSystem)
  } else if (!props.fieldSubset) {
    sysFields = props.module.systemFields().map(sf => {
      sf.label = t(`system.${sf.name}`)
      return sf
    })

    if (props.systemFields) {
      sysFields = sysFields.filter(({ name }) => props.systemFields.find(sf => sf === name))
    }
  }

  if (!props.disableSorting && mFields) {
    mFields.sort((a, b) => a.label.localeCompare(b.label))
  }

  if (mFields && sysFields) {
    return [
      ...[...mFields],
      ...sysFields,
    ].map(field => ({
      value: field.name,
      text: [
        field.name,
        field.label,
        field.kind,
        field.isSystem ? 'system' : '',
        field.isRequired ? 'required' : '',
      ].join(' '),
      field,
    }))
  } else {
    return Object.keys(props.module).map(key => {
      return props.module[key]
    }).map(f => ({
      ...f,
      field: {
        name: f.text,
      },
    }))
  }
})

</script>
