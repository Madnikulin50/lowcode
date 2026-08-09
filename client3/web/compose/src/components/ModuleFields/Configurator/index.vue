<template>
  <div>
    <ul class="nav nav-tabs" role="tablist">
      <li class="nav-item" role="presentation">
        <button class="nav-link active" data-bs-toggle="tab" data-bs-target="#tab-general" type="button" role="tab">{{ t('label.general') }}</button>
      </li>
      <li v-if="fieldComponent" class="nav-item" role="presentation">
        <button class="nav-link" data-bs-toggle="tab" data-bs-target="#tab-field" type="button" role="tab">{{ t(`general.fieldKinds.${field.kind}.label`) }}</button>
      </li>
      <li v-if="field.cap.multi" class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ disabled: !field.isMulti }"
          data-bs-toggle="tab"
          data-bs-target="#tab-multi"
          type="button"
          role="tab"
        >{{ t('label.multi') }}</button>
      </li>
      <li class="nav-item" role="presentation">
        <button class="nav-link" data-bs-toggle="tab" data-bs-target="#tab-validation" type="button" role="tab">{{ t('label.validation') }}</button>
      </li>
      <li class="nav-item" role="presentation">
        <button class="nav-link" data-bs-toggle="tab" data-bs-target="#tab-privacy" type="button" role="tab">{{ t('label.privacy') }}</button>
      </li>
    </ul>

    <div class="tab-content px-2 h-auto overflow-auto" style="max-height: 70vh;">
      <div id="tab-general" class="tab-pane active" role="tabpanel">
        <basic :namespace="namespace" :module="module" v-model:field="f" />
      </div>

      <div v-if="fieldComponent" id="tab-field" class="tab-pane" role="tabpanel">
        <component
          :is="fieldComponent"
          :namespace="namespace"
          :module="module"
          v-model:field="f"
          :has-records="hasRecords"
        />
      </div>

      <div v-if="field.cap.multi && field.isMulti" id="tab-multi" class="tab-pane" role="tabpanel">
        <multi :namespace="namespace" v-model:field="f" />
      </div>

      <div id="tab-validation" class="tab-pane" role="tabpanel">
        <validation :namespace="namespace" :module="module" v-model:field="f" />
      </div>

      <div id="tab-privacy" class="tab-pane" role="tabpanel">
        <DataPrivacySettings
          v-if="connection"
          :resource="field"
          :connection="connection"
          :sensitivity-levels="sensitivityLevels"
          :max-level="maxLevelID"
          :translations="{
            sensitivity: {
              label: t('privacy.sensitivity-level.label'),
              placeholder: t('privacy.sensitivity-level.placeholder'),
            },
            usage: {
              label: t('privacy.usage-disclosure.label'),
            },
          }"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { useConfiguratorBase } from './base'
import * as Configurators from './loader'
import multi from './multi'
import basic from './basic'
import validation from './validation'
import DataPrivacySettings from 'corteza-webapp-compose/src/components/Admin/Module/DataPrivacySettings'

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
  connection: { type: Object, required: true },
  sensitivityLevels: { type: Array, default: () => [] },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f, isNew, hasData } = useConfiguratorBase(props, emit)

const fieldComponent = computed(() => Configurators[props.field.kind])

const maxLevelID = computed(() => {
  const { sensitivityLevelID = NoID } = props.module.config.privacy || {}
  return sensitivityLevelID
})
</script>
