<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ $t('settings.title') }}
      </h4>
    </div>

    <form @submit.prevent="$emit('submit', settings)">
      <div class="row g-3 p-3">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('settings.profiler.label') }}</label>
            <div class="btn-group" role="group">
              <template v-for="opt in profilerOptions" :key="opt.value">
                <input
                  :id="'profiler-' + opt.value"
                  v-model="profilerSetting"
                  type="radio"
                  class="btn-check"
                  name="profiler-options"
                  :value="opt.value"
                  autocomplete="off"
                >
                <label
                  :for="'profiler-' + opt.value"
                  class="btn btn-outline-primary btn-sm"
                >
                  {{ opt.text }}
                </label>
              </template>
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('settings.proxy.label') }}</label>
            <div class="form-check">
              <input
                id="proxy-follow"
                v-model="settings['apigw.proxy.follow-redirects']"
                type="checkbox"
                class="form-check-input"
              >
              <label
                class="form-check-label"
                for="proxy-follow"
              >
                {{ $t('settings.proxy.follow') }}
              </label>
            </div>
          </div>
        </div>
      </div>
    </form>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit', settings)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  settings: {
    type: Object,
    required: true,
  },
  processing: {
    type: Boolean,
    default: false,
  },
  success: {
    type: Boolean,
    default: false,
  },
})

const profilerOptions = [
  { value: 'disabled', text: 'Disabled' },
  { value: 'filter', text: 'Enabled as filter' },
  { value: 'global', text: 'Enabled for all routes' },
]

const profilerSetting = computed({
  get () {
    if (props.settings['apigw.profiler.enabled']) {
      return props.settings['apigw.profiler.global'] ? 'global' : 'filter'
    }
    return 'disabled'
  },
  set (setting) {
    props.settings['apigw.profiler.enabled'] = ['filter', 'global'].includes(setting)
    props.settings['apigw.profiler.global'] = setting === 'global'
  },
})
</script>
