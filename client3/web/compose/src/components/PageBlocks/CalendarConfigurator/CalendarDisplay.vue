<template>
  <div class="row">
    <div class="col-12 col-lg-6">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('calendar.view.enabled') }}</label>
        <div class="btn-group btn-group-sm" data-bs-toggle="buttons">
          <label
            v-for="view in views"
            :key="view.value"
            class="btn btn-outline-secondary"
            :class="{ active: options.header.views.includes(view.value) }"
          >
            <input
              :checked="options.header.views.includes(view.value)"
              type="checkbox"
              class="btn-check"
              :disabled="options.header.hide"
              @change="toggleView(view.value)"
            />
            {{ view.text }}
          </label>
        </div>
      </div>
    </div>

    <div class="col-12 col-lg-6">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('calendar.view.default') }}</label>
        <small class="form-text">{{ $t('calendar.view.footnote') }}</small>
        <div class="btn-group btn-group-sm" data-bs-toggle="buttons">
          <label
            v-for="view in views"
            :key="view.value"
            class="btn btn-outline-secondary"
            :class="{ active: options.defaultView === view.value }"
          >
            <input
              v-model="options.defaultView"
              type="radio"
              :value="view.value"
              class="btn-check"
            />
            {{ view.text }}
          </label>
        </div>
      </div>
    </div>

    <div class="col-12 col-lg-6">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('calendar.view.onEventClick') }}</label>
        <select
          v-model="options.eventDisplayOption"
          class="form-select"
        >
          <option
            v-for="opt in eventDisplayOptions"
            :key="opt.value"
            :value="opt.value"
          >
            {{ opt.text }}
          </option>
        </select>
      </div>
    </div>

    <div class="col-12 col-lg-6">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('calendar.calendarHeader') }}</label>
        <div class="form-check">
          <input
            v-model="options.header.hide"
            type="checkbox"
            class="form-check-input"
          />
          <label class="form-check-label">{{ $t('calendar.hideHeader') }}</label>
        </div>
        <div class="form-check">
          <input
            v-model="options.header.hidePrevNext"
            type="checkbox"
            class="form-check-input"
            :disabled="options.header.hide"
          />
          <label class="form-check-label">{{ $t('calendar.hideNavigation') }}</label>
        </div>
        <div class="form-check">
          <input
            v-model="options.header.hideToday"
            type="checkbox"
            class="form-check-input"
            :disabled="options.header.hide"
          />
          <label class="form-check-label">{{ $t('calendar.hideToday') }}</label>
        </div>
        <div class="form-check">
          <input
            v-model="options.header.hideTitle"
            type="checkbox"
            class="form-check-input"
            :disabled="options.header.hide"
          />
          <label class="form-check-label">{{ $t('calendar.hideTitle') }}</label>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePageBlockBase } from '../usePageBlockBase'
import { compose } from 'corteza-lib/js/dist'

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const emit = defineEmits(['errors'])

const { options } = usePageBlockBase(props, emit)

const views = computed(() =>
  compose.PageBlockCalendar.availableViews().map(view => ({
    value: view,
    text: view,
  }))
)

const eventDisplayOptions = computed(() => [
  { value: 'sameTab', text: 'Open in same tab' },
  { value: 'newTab', text: 'Open in new tab' },
  { value: 'modal', text: 'Open in modal' },
])

function toggleView (viewValue) {
  const idx = options.value.header.views.indexOf(viewValue)
  if (idx > -1) {
    options.value.header.views.splice(idx, 1)
  } else {
    options.value.header.views.push(viewValue)
  }
}
</script>
