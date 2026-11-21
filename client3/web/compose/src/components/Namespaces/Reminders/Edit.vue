<template>
  <div class="d-flex flex-column h-100 p-3">
    <form
      v-if="reminder"
      class="flex-fill overflow-auto"
      @submit.prevent
    >
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('reminder.edit.titleLabel') }}</label>
        <input
          v-model="reminder.payload.title"
          data-test-id="input-title"
          required
          class="form-control"
          :placeholder="$t('reminder.edit.titlePlaceholder')"
        />
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('reminder.edit.notesLabel') }}</label>
        <textarea
          v-model="reminder.payload.notes"
          data-test-id="textarea-notes"
          class="form-control"
          :placeholder="$t('reminder.edit.notesPlaceholder')"
          rows="6"
        />
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('reminder.edit.remindAtLabel') }}</label>
        <c-input-date-time
          v-model="reminder.remindAt"
          data-test-id="select-remind-at"
          only-future
          :labels="{
            clear: $t('label.clear'),
            none: $t('label.none'),
            now: $t('label.now'),
            today: $t('label.today'),
          }"
        />
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('reminder.edit.assigneeLabel') }}</label>
        <c-input-user
          v-model="reminder.assignedTo"
          data-test-id="select-assignee"
          :placeholder="$t('field.kind.user.suggestionPlaceholder')"
        />
      </div>

      <div
        v-if="reminder.payload.link"
        class="mb-3"
      >
        <label class="form-label text-primary">{{ $t('reminder.routesTo') }}</label>
        <div class="input-group">
          <input
            v-model="reminder.payload.link.label"
            data-test-id="input-link"
            class="form-control"
          />
          <button
            :disabled="!recordViewer"
            class="btn btn-outline-primary d-flex align-items-center"
            @click="$router.push(recordViewer)"
          >
            <font-awesome-icon :icon="['far', 'file-alt']" />
          </button>
        </div>
      </div>

      <div
        v-if="reminder.reminderID !== '0'"
        class="mb-3"
      >
        <div class="form-check">
          <input
            :checked="!!reminder.dismissedAt"
            type="checkbox"
            class="form-check-input"
            @change="$emit('dismiss', reminder, $event)"
          />
          <label class="form-check-label">{{ $t('reminder.dismissed') }}</label>
        </div>
      </div>

      <div
        v-if="reminder.dismissedAt"
        class="mb-3"
      >
        <label class="form-label text-primary">{{ $t('reminder.dismissedAt') }}</label>
        <div>{{ $d(reminder.dismissedAt, 'long') }}</div>
      </div>

      <div
        v-if="reminder.snoozeCount"
        class="mb-3"
      >
        <label class="form-label text-primary">{{ $t('reminder.snooze.count') }}</label>
        <div>{{ reminder.snoozeCount }}</div>
      </div>
    </form>

    <div class="d-flex align-items-center justify-content-around">
      <button
        data-test-id="button-back"
        class="btn btn-outline-light d-flex align-items-center back text-primary border-0 gap-1"
        @click="$emit('back')"
      >
        <font-awesome-icon
          :icon="['fas', 'chevron-left']"
          class="back-icon"
        />
        {{ $t('label.back') }}
      </button>

      <c-button-submit
        data-test-id="button-save"
        :disabled="disableSave"
        :processing="processingSave"
        :text="$t('label.save')"
        @submit="$emit('save', reminder)"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'general' } })
import { ref, computed, watch } from 'vue'
import { system } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
const { CInputDateTime, CInputUser } = components

const emit = defineEmits(['back', 'dismiss', 'save'])

const props = defineProps({
  edit: { type: Object, required: false, default: () => ({}) },
  disableSave: { type: Boolean, default: false },
  processingSave: { type: Boolean, default: false },
})

const processingUsers = ref(false)
const reminder = ref(undefined)

const recordViewer = computed(() => {
  const { params } = (reminder.value || {}).payload?.link || {}
  return params ? { name: 'page.record', params } : undefined
})

watch(() => props.edit, (val) => {
    reminder.value = new system.Reminder(val)
  }, { immediate: true, deep: true })
</script>
