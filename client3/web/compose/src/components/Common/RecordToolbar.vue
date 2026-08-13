<template>
  <c-toolbar
    :class="{ 'shadow border-top': !inModal }"
    :style="{ padding: inModal ? '0 !important' : '' }"
  >
    <template #start>
      <button
        v-if="!(hideBack || settings.hideBack)"
        type="button"
        data-test-id="button-back"
        :disabled="isProcessing"
        class="btn btn-outline-light btn-lg border-0 text-dark back"
        @click.prevent="$emit('back')"
      >
        <span class="d-flex align-items-center gap-1">
          <font-awesome-icon
            :icon="['fas', hasBack ? 'chevron-left' : 'times']"
            :class="hasBack ? 'back-icon' : ''"
          />
          {{ backLabel }}
        </span>
      </button>

      <slot name="start-actions" />
    </template>

    <template #center>
      <div
        v-if="isCreated && (recordNavigation.prev || recordNavigation.next)"
        class="d-flex align-items-center fill-width gap-1"
      >
        <span :title="$t('recordNavigation.prev')">
          <button
            type="button"
            class="btn btn-outline-primary btn-lg rounded-pill w-100"
            :disabled="!record || isProcessing || !recordNavigation.prev"
            @click="navigateToRecord(recordNavigation.prev)"
          >
            <font-awesome-icon :icon="['fas', 'angle-left']" />
          </button>
        </span>

        <span :title="$t('recordNavigation.next')">
          <button
            type="button"
            class="btn btn-outline-primary btn-lg rounded-pill w-100"
            :disabled="!record || isProcessing || !recordNavigation.next"
            @click="navigateToRecord(recordNavigation.next)"
          >
            <font-awesome-icon :icon="['fas', 'angle-right']" />
          </button>
        </span>
      </div>

      <slot name="center-actions" />
    </template>

    <template
      v-if="module"
      #end
    >
      <slot name="end-actions" />

      <c-input-confirm
        v-if="(processingAction === 'delete' || isCreated || isDraft) && !(isDeleted || hideDelete || settings.hideDelete) && canDeleteRecord"
        :disabled="!record || isProcessing"
        :processing="processingAction === 'delete'"
        :text="labels.delete || $t('label.delete')"
        size="lg"
        size-confirm="lg"
        variant="danger"
        class="text-nowrap"
        @confirmed="$emit('delete')"
      />

      <c-input-confirm
        v-else-if="(processingAction === 'undelete' || isDeleted) && !(hideDelete || settings.hideDelete) && canUndeleteRecord"
        :disabled="!record || isProcessing"
        :processing="processingAction === 'undelete'"
        :text="$t('label.restore')"
        size="lg"
        size-confirm="lg"
        variant="warning"
        variant-ok="warning"
        class="text-nowrap"
        @confirmed="$emit('undelete')"
      />

      <button
        v-if="(isCreated || isDraft) && module.canCreateRecord && !(hideClone || settings.hideClone)"
        type="button"
        data-test-id="button-clone"
        class="btn btn-outline-secondary btn-lg text-nowrap"
        :disabled="!record || isProcessing"
        @click.prevent="$emit('clone')"
      >
        {{ labels.clone || (isDraft ? $t('label.saveAsNewDraft') : $t('label.saveAsCopy')) }}
      </button>

      <button
        v-if="!inEditing && (isCreated || isDraft) && !(hideEdit || settings.hideEdit) && canManageRecord"
        type="button"
        data-test-id="button-edit"
        :disabled="!record || isProcessing"
        class="btn btn-outline-secondary btn-lg text-nowrap"
        @click.prevent="$emit('edit')"
      >
        {{ labels.edit || $t('label.edit') }}
      </button>

      <button
        v-else-if="inEditing && (isCreated || (isDraft && !isNew)) && !(hideEdit || settings.hideEdit)"
        type="button"
        data-test-id="button-view"
        :disabled="!record || isProcessing"
        class="btn btn-outline-secondary btn-lg text-nowrap"
        @click.prevent="$emit('view')"
      >
        {{ labels.edit || $t('label.view') }}
      </button>

      <button
        v-if="!inEditing && module.canCreateRecord && !(hideNew || settings.hideNew)"
        type="button"
        data-test-id="button-add-new"
        class="btn btn-primary btn-lg text-nowrap"
        :disabled="!record || isProcessing"
        @click.prevent="$emit('add')"
      >
        {{ labels.new || $t('label.addNew') }}
      </button>

      <c-button-submit
        v-if="inEditing && !(hideSubmit || settings.hideSubmit) && canManageRecord"
        data-test-id="button-save"
        :disabled="!record || isProcessing"
        :processing="processingAction === 'submit'"
        :text="labels.submit || $t('label.save')"
        size="lg"
        class="text-nowrap"
        @submit="$emit('submit')"
      />
    </template>
  </c-toolbar>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'general' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
import { compose, NoID } from 'corteza-lib/js/dist'
import { throttle } from 'lodash'
const { CToolbar } = components

const { t } = useI18n()

const $Settings = window.__settings

const props = defineProps({
  module: {
    type: compose.Module,
    required: false,
    default: undefined,
  },
  record: {
    type: compose.Record,
    required: false,
    default: undefined,
  },
  labels: {
    type: Object,
    default: () => ({}),
  },
  processing: {
    type: Boolean,
    default: false,
  },
  processingAction: {
    type: String,
    default: '',
  },
  isCreated: {
    type: Boolean,
    required: true,
  },
  inEditing: {
    type: Boolean,
    required: true,
  },
  hideBack: {
    type: Boolean,
    default: () => true,
  },
  hideDelete: {
    type: Boolean,
    default: () => true,
  },
  hideNew: {
    type: Boolean,
    default: () => true,
  },
  hideClone: {
    type: Boolean,
    default: () => true,
  },
  hideEdit: {
    type: Boolean,
    default: () => true,
  },
  hideSubmit: {
    type: Boolean,
    default: () => true,
  },
  inModal: {
    type: Boolean,
    required: false,
  },
  recordNavigation: {
    type: Object,
    required: false,
    default: () => ({}),
  },
  hasBack: {
    type: Boolean,
    default: true,
  },
  isDraft: {
    type: Boolean,
    default: false,
  },
  isNew: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['back', 'delete', 'undelete', 'clone', 'edit', 'view', 'add', 'submit', 'update-navigation'])

const isDeleted = computed(() => props.record && props.record.deletedAt)
const isProcessing = computed(() => props.processing || !!props.processingAction)
const settings = computed(() => $Settings.get('compose.ui.record-toolbar', {}))

const canManageRecord = computed(() => {
  if (!props.module || !props.record) return false
  return props.record.recordID === NoID ? props.module.canCreateRecord : props.record.canUpdateRecord
})

const canDeleteRecord = computed(() => {
  if (!props.module || !props.record) return false
  return !isDeleted.value && (props.isDraft || (props.record.canDeleteRecord && props.record.recordID !== NoID))
})

const canUndeleteRecord = computed(() => {
  if (!props.module || !props.record) return false
  return isDeleted.value && props.record.canUndeleteRecord && props.record.recordID !== NoID
})

const backLabel = computed(() => {
  if (props.inModal) {
    return props.hasBack ? t('label.back') : t('label.close')
  }
  return props.hasBack ? props.labels.back || t('label.back') : t('label.home')
})

const navigateToRecord = throttle(function (recordID) {
  emit('update-navigation', recordID)
}, 500)
</script>
