<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ t('title') }}
      </h4>
    </div>

    <form
      @submit.prevent="emit('submit', settings)"
    >
      <div class="card-body">
        <h5>{{ t('general') }}</h5>
        <div class="mb-3">
          <div class="form-check">
            <input
              id="hide-app-selector"
              v-model="topbarSettings.hideAppSelector"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-app-selector"
            >{{ t('app-selector.hide') }}</label>
          </div>

          <div class="form-check">
            <input
              id="hide-drafts"
              v-model="hideDrafts"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-drafts"
            >
              {{ t('drafts.hide') }}
              <span class="badge bg-warning ms-1">{{ t('label.experimental') }}</span>
            </label>
          </div>

          <div class="form-check">
            <input
              id="hide-notifications"
              v-model="topbarSettings.hideNotifications"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-notifications"
            >{{ t('notifications.hide') }}</label>
          </div>

          <div class="form-check">
            <input
              id="hide-help"
              v-model="topbarSettings.hideHelp"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-help"
            >{{ t('help.hide') }}</label>
          </div>

          <div
            v-if="$Settings.get('discovery.enabled', false)"
            class="form-check"
          >
            <input
              id="hide-search"
              v-model="hideSearch"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-search"
            >{{ t('search.hide') }}</label>
          </div>

          <div class="form-check">
            <input
              id="hide-profile"
              v-model="topbarSettings.hideProfile"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-profile"
            >{{ t('profile.hide') }}</label>
          </div>
        </div>

        <div>
          <hr>

          <div class="row">
            <div class="col-12 col-lg-3">
              <h5>{{ t('help.title') }}</h5>
              <div class="mb-3">
                <div class="form-check">
                  <input
                    id="hide-forum-link"
                    v-model="topbarSettings.hideForumLink"
                    class="form-check-input"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="hide-forum-link"
                  >{{ t('help.hide-forum-link') }}</label>
                </div>

                <div class="form-check">
                  <input
                    id="hide-documentation-link"
                    v-model="topbarSettings.hideDocumentationLink"
                    class="form-check-input"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="hide-documentation-link"
                  >{{ t('help.hide-documentation-link') }}</label>
                </div>

                <div class="form-check">
                  <input
                    id="hide-feedback-link"
                    v-model="topbarSettings.hideFeedbackLink"
                    class="form-check-input"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="hide-feedback-link"
                  >{{ t('help.hide-feedback-link') }}</label>
                </div>
              </div>
            </div>

            <div class="col">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('links.title') }}</label>
                <c-form-table-wrapper
                  :labels="{ addButton: t('label.add') }"
                  @add-item="topbarSettings.helpLinks.push({ handle: '', url: '', newTab: true })"
                >
                  <table
                    v-if="topbarSettings.helpLinks.length > 0"
                    class="table table-responsive table-sm mb-0"
                  >
                    <thead>
                      <tr class="text-primary">
                        <th style="width: 30%;">{{ t('links.handle') }}</th>
                        <th style="width: 55%;">{{ t('links.url') }}</th>
                        <th class="text-center" style="width: 10%;">{{ t('links.new-tab') }}</th>
                        <th style="width: 1%;"></th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(item, index) in topbarSettings.helpLinks" :key="index">
                        <td>
                          <input
                            v-model="item.handle"
                            class="form-control form-control-sm"
                          >
                        </td>
                        <td>
                          <input
                            v-model="item.url"
                            class="form-control form-control-sm"
                            type="url"
                          >
                        </td>
                        <td class="text-center align-middle">
                          <div class="form-check">
                            <input
                              :id="'help-newtab-' + index"
                              v-model="item.newTab"
                              class="form-check-input"
                              type="checkbox"
                            >
                            <label
                              class="form-check-label"
                              :for="'help-newtab-' + index"
                            >&nbsp;</label>
                          </div>
                        </td>
                        <td class="text-end align-middle">
                          <c-input-confirm
                            show-icon
                            @confirmed="topbarSettings.helpLinks.splice(index, 1)"
                          />
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </c-form-table-wrapper>
              </div>
            </div>
          </div>
        </div>

        <div>
          <hr>

          <div class="row">
            <div class="col-12 col-lg-3">
              <h5>{{ t('profile.title') }}</h5>

              <div class="mb-3">
                <div class="form-check">
                  <input
                    id="hide-profile-link"
                    v-model="topbarSettings.hideProfileLink"
                    class="form-check-input"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="hide-profile-link"
                  >{{ t('profile.hide-profile-link') }}</label>
                </div>

                <div class="form-check">
                  <input
                    id="hide-change-password-link"
                    v-model="topbarSettings.hideChangePasswordLink"
                    class="form-check-input"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="hide-change-password-link"
                  >{{ t('profile.hide-change-password-link') }}</label>
                </div>

                <div class="form-check">
                  <input
                    id="hide-theme-selector"
                    v-model="topbarSettings.hideThemeSelector"
                    class="form-check-input"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="hide-theme-selector"
                  >{{ t('profile.hide-theme-selector') }}</label>
                </div>
              </div>
            </div>

            <div class="col">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('links.title') }}</label>
                <c-form-table-wrapper
                  :labels="{ addButton: t('label.add') }"
                  @add-item="topbarSettings.profileLinks.push({ handle: '', url: '', newTab: true })"
                >
                  <table
                    v-if="topbarSettings.profileLinks.length > 0"
                    class="table table-responsive table-sm mb-0"
                  >
                    <thead>
                      <tr class="text-primary">
                        <th style="width: 30%;">{{ t('links.handle') }}</th>
                        <th style="width: 55%;">{{ t('links.url') }}</th>
                        <th class="text-center" style="width: 10%;">{{ t('links.new-tab') }}</th>
                        <th style="width: 1%;"></th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(item, index) in topbarSettings.profileLinks" :key="index">
                        <td>
                          <input
                            v-model="item.handle"
                            class="form-control form-control-sm"
                          >
                        </td>
                        <td>
                          <input
                            v-model="item.url"
                            class="form-control form-control-sm"
                            type="url"
                          >
                        </td>
                        <td class="text-center align-middle">
                          <div class="form-check">
                            <input
                              :id="'profile-newtab-' + index"
                              v-model="item.newTab"
                              class="form-check-input"
                              type="checkbox"
                            >
                            <label
                              class="form-check-label"
                              :for="'profile-newtab-' + index"
                            >&nbsp;</label>
                          </div>
                        </td>
                        <td class="text-end align-middle">
                          <c-input-confirm
                            show-icon
                            @confirmed="topbarSettings.profileLinks.splice(index, 1)"
                          />
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </c-form-table-wrapper>
              </div>
            </div>
          </div>
        </div>
      </div>
    </form>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        v-if="canManage"
        :processing="processing"
        :success="success"
        :text="t('admin.general.label.submit')"
        class="ms-auto"
        @submit="onSubmit"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, inject } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: _t } = useI18n()

function t(key, ...args) {
  if (key.startsWith('label.') || key.startsWith('admin.')) return _t(key, ...args)
  return _t('editor.topbar.' + key, ...args)
}
const $Settings = inject('$Settings', {})

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
  canManage: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['submit'])

const topbarSettings = ref({})

const links = {
  fields: [
    {
      key: 'handle',
      label: t('links.handle'),
      thStyle: { width: '30%' },
    },
    {
      key: 'url',
      label: t('links.url'),
      thStyle: { width: '55%' },
    },
    {
      key: 'newTab',
      label: t('links.new-tab'),
      thClass: 'text-center',
      tdClass: 'text-center align-middle',
      thStyle: { width: '10%' },
    },
    {
      key: 'actions',
      label: '',
      tdClass: 'text-end align-middle',
      thStyle: { width: '1%' },
    },
  ],
}

const hideDrafts = computed({
  get () {
    return topbarSettings.value.showDrafts !== true
  },
  set (value) {
    topbarSettings.value.showDrafts = !value
  },
})

const hideSearch = computed({
  get () {
    return topbarSettings.value.showSearch !== true
  },
  set (value) {
    topbarSettings.value.showSearch = !value
  },
})

watch(() => props.settings, (settings) => {
  topbarSettings.value = settings['ui.topbar'] || {}

  if (!topbarSettings.value.helpLinks) {
    topbarSettings.value.helpLinks = []
  }

  if (!topbarSettings.value.profileLinks) {
    topbarSettings.value.profileLinks = []
  }
}, { immediate: true })

function onSubmit () {
  emit('submit', { 'ui.topbar': topbarSettings.value })
}
</script>
