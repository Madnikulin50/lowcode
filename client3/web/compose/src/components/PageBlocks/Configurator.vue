<template>
  <div>
    <ul class="nav nav-tabs" role="tablist">
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ active: activeTab === 0 }"
          data-test-id="general-tab"
          @click="activeTabInternal = 0"
        >
          {{ $t('label.general') }}
        </button>
      </li>
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ active: activeTab === 1 }"
          @click="activeTabInternal = 1"
        >
          {{ $t('label.configurator') }}
        </button>
      </li>
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ active: activeTab === 2 }"
          data-test-id="ai-tab"
          @click="activeTabInternal = 2"
        >
          {{ $t('ai.tab.title') }}
        </button>
      </li>
    </ul>

    <div class="tab-content py-3">
      <div
        v-show="activeTab === 0"
        class="tab-pane active"
      >
        <div class="row">
          <div class="col-12">
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ $t('titleLabel') }}</label>
              <div class="input-group">
                <c-input-expression
                  id="title"
                  v-model="block.title"
                  auto-complete
                  :placeholder="$t('titlePlaceholder')"
                  :suggestion-params="recordAutoCompleteParams"
                  class="flex-grow-1"
                />
                <page-translator
                  v-if="page"
                  :page="page"
                  :block="block"
                  :disabled="isNew"
                  :highlight-key="`pageBlock.${block.blockID}.title`"
                />
              </div>
              <i18next
                path="interpolationFootnote"
                tag="small"
                class="text-muted"
              >
                <code>${record.values.fieldName}</code>
                <code>${recordID}</code>
                <code>${ownerID}</code>
                <span><code>${userID}</code>, <code>${user.name}</code></span>
              </i18next>
            </div>
          </div>

          <div class="col-12">
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ $t('descriptionLabel') }}</label>
              <div class="input-group">
                <c-input-expression
                  id="description"
                  v-model="block.description"
                  auto-complete
                  :placeholder="$t('descriptionPlaceholder')"
                  :suggestion-params="recordAutoCompleteParams"
                  class="flex-grow-1"
                />
                <page-translator
                  v-if="page"
                  :page="page"
                  :block="block"
                  :disabled="isNew"
                  :highlight-key="`pageBlock.${block.blockID}.description`"
                />
              </div>
              <i18next
                path="interpolationFootnote"
                tag="small"
                class="text-muted"
              >
                <code>${record.values.fieldName}</code>
                <code>${recordID}</code>
                <code>${ownerID}</code>
                <span><code>${userID}</code>, <code>${user.name}</code></span>
              </i18next>
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ $t('customID.label') }}</label>
              <input
                id="customID"
                v-model="block.meta.customID"
                class="form-control"
                :class="{ 'is-invalid': customIDState === false }"
                :placeholder="$t('customID.placeholder')"
              />
              <div
                v-if="customIDState === false"
                class="invalid-feedback"
              >
                {{ $t('customID.invalid-state') }}
              </div>
              <div
                v-else
                class="form-text text-muted"
              >
                {{ $t('customID.description') }}
              </div>
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ $t('customCSSClass.label') }}</label>
              <input
                id="customCSSClass"
                v-model="block.meta.customCSSClass"
                class="form-control"
                :class="{ 'is-invalid': customCSSClassState === false }"
                :placeholder="$t('customCSSClass.placeholder')"
              />
              <div
                v-if="customCSSClassState === false"
                class="invalid-feedback"
              >
                {{ $t('customCSSClass.invalid-state') }}
              </div>
              <div
                v-else
                class="form-text text-muted"
              >
                {{ $t('customCSSClass.description') }}
              </div>
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ $t('headerStyle') }}</label>
              <c-input-select
                id="color"
                v-model="block.style.variants.headerText"
                :options="textVariants"
                :reduce="o => o.value"
                :clearable="false"
                :placeholder="$t('label.none')"
                label="text"
                class="mb-1"
              />
              <div class="form-check">
                <input
                  v-model="block.style.wrap.kind"
                  :true-value="'card'"
                  :false-value="'plain'"
                  type="checkbox"
                  class="form-check-input"
                />
                <label class="form-check-label">{{ $t('wrap') }}</label>
              </div>
              <div class="form-check">
                <input
                  v-model="block.style.border.enabled"
                  type="checkbox"
                  class="form-check-input"
                />
                <label class="form-check-label">{{ $t('border.show') }}</label>
              </div>
            </div>
          </div>

          <div
            v-if="block.options.magnifyOption !== undefined"
            class="col-12 col-lg-6"
          >
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ $t('magnifyLabel') }}</label>
              <select
                v-model="block.options.magnifyOption"
                class="form-select form-control"
              >
                <option
                  v-for="opt in magnifyOptions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.text }}
                </option>
              </select>
            </div>
          </div>

          <div
            v-if="block.options.showRefresh !== undefined"
            class="col-12 col-lg-6"
          >
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ $t('refresh.auto') }}</label>
              <small class="form-text">{{ $t('refresh.description') }}</small>
              <div class="input-group mb-1">
                <input
                  v-model="block.options.refreshRate"
                  type="number"
                  class="form-control"
                  min="0"
                  @blur="updateRefresh"
                />
                <span class="input-group-text">s</span>
              </div>
              <div class="form-check">
                <input
                  v-model="block.options.showRefresh"
                  type="checkbox"
                  class="form-check-input"
                />
                <label class="form-check-label">{{ $t('refresh.show') }}</label>
              </div>
            </div>
          </div>
        </div>

        <hr />

        <h5 class="mb-3">
          {{ $t('visibility.label') }}
        </h5>

        <div class="row">
          <div class="col-12">
            <div class="mb-3">
              <label class="d-flex align-items-center text-primary mb-0 form-label">
                {{ $t('visibility.condition.label') }}
                <c-hint
                  :tooltip="$t('visibility.tooltip.performance.condition')"
                  icon-class="text-warning"
                />
                <a
                  :href="visibilityDocumentationURL"
                  target="_blank"
                  class="text-primary ms-auto p-0 btn btn-link"
                >
                  {{ $t('label.examples') }}
                </a>
              </label>
              <div class="input-group">
                <span class="input-group-text">ƒ</span>
                <c-input-expression
                  id="visibility-fields"
                  v-model="block.meta.visibility.expression"
                  auto-complete
                  :placeholder="$t('visibility.condition.placeholder')"
                  :suggestion-params="visibilityAutoCompleteParams"
                  class="flex-grow-1"
                />
              </div>
              <i18next
                v-if="isRecordPage"
                path="general.visibility.condition.description.record-page"
                tag="small"
                class="text-muted"
              >
                <code>record.values.fieldName</code>
                <code>user.(userID/email...)</code>
                <code>screen.(width/height)</code>
                <code>isView/isCreate/isEdit</code>
                <code>user.userID == record.createdBy</code>
                <code>screen.width &lt; 1024</code>
              </i18next>
              <i18next
                v-else
                path="general.visibility.condition.description.non-record-page"
                tag="small"
                class="text-muted"
              >
                <code>user.(userID/email...)</code>
                <code>screen.(width/height)</code>
                <code>user.email == "test@mail.com"</code>
                <code>screen.width &lt; 1024</code>
              </i18next>
            </div>
          </div>

          <div class="col-12">
            <div class="mb-3">
              <label class="form-label fw-semibold">{{ $t('visibility.roles.label') }}</label>
              <c-input-select
                v-model="currentRoles"
                :options="roles.options"
                :loading="roles.processing"
                :placeholder="$t('visibility.roles.placeholder')"
                :get-option-label="role => role.name"
                :reduce="role => role.roleID"
                :selectable="role => !currentRoles.includes(role.roleID)"
                multiple
              />
            </div>
          </div>
        </div>
      </div>

      <div
        v-show="activeTab === 1"
        class="tab-pane active"
      >
        <page-block
          v-bind="{ ...$attrs, ...$props }"
          mode="configurator"
          class="mh-tab overflow-auto"
        />
      </div>

      <div
        v-show="activeTab === 2"
        class="tab-pane active"
      >
        <div class="row">
          <div class="col-12">
            <div class="mb-3">
              <div class="form-check form-switch mb-2">
                <label class="form-check-label fw-semibold" for="hideBrainButton">{{ $t('ai.hideBrainButton.label', 'Hide AI button on block') }}</label>
                <input
                  id="hideBrainButton"
                  v-model="block.options.hideBrainButton"
                  class="form-check-input"
                  type="checkbox"
                />
              </div>

              <label class="form-label fw-semibold">{{ $t('ai.prompt.label') }}</label>
              <div class="input-group">
                <c-rich-text-input
                  v-model="block.prompt"
                  :placeholder="$t('ai.prompt.placehoder')"
                  body-class="form-control"
                  min-body-height="10rem"
                  output-format="markdown"
                  :to-markdown="htmlToMarkdown"
                  :to-html="markdownToHtml"
                  :labels="{
                    urlPlaceholder: $t('content.urlPlaceholder'),
                    ok: $t('content.ok'),
                  }"
                  class="flex-grow-1"
                />
                <page-translator
                  v-if="page"
                  :page="page"
                  :block="block"
                  :disabled="isNew"
                  :highlight-key="`pageBlock.${block.blockID}.prompt`"
                />
              </div>
              <i18next
                path="interpolationFootnote"
                tag="small"
                class="text-muted"
              >
                <code>${record.values.fieldName}</code>
                <code>${recordID}</code>
                <code>${ownerID}</code>
                <span><code>${userID}</code>, <code>${user.name}</code></span>
              </i18next>
            </div>
          </div>
        </div>
      </div>
    </div>

    <page-translator
      v-if="page"
      :page="page"
      :block="block"
      :disabled="isNew"
      button-variant="link"
      class="mt-2"
    />
  </div>
</template>

<script setup>
import { ref, computed, onBeforeUnmount, onMounted, inject } from 'vue'
import { compose, NoID } from 'corteza-lib/js/dist'
import { handle, components, composables } from 'corteza-lib/vue/dist'
import PageTranslator from 'corteza-webapp-compose/src/components/Admin/Page/PageTranslator'
import PageBlock from './index'
import { useRoute } from 'vue-router'

const { CInputExpression, CRichTextInput } = components
import { htmlToMarkdown, markdownToHtml } from '../../lib/markdown'

const props = defineProps({
  block: { type: compose.PageBlock, required: true },
  module: { type: compose.Module, required: false, default: undefined },
  page: { type: compose.Page, required: true },
  record: { type: [Object, null], required: false, default: null },
  namespace: { type: compose.Namespace, required: true },
})

const route = useRoute()
const $auth = inject('$auth')
const $SystemAPI = inject('$SystemAPI')
const { toastErrorHandler } = composables.useToast()

const roles = ref({ processing: false, options: [] })
const abortableRequests = ref([])
const activeTabInternal = ref(0)

const activeTab = computed(() => activeTabInternal.value)

const textVariants = computed(() => [
  { value: 'dark', text: 'Default' },
  { value: 'primary', text: 'Primary' },
  { value: 'secondary', text: 'Secondary' },
  { value: 'success', text: 'Success' },
  { value: 'warning', text: 'Warning' },
  { value: 'danger', text: 'Danger' },
])

const isNew = computed(() => props.block.blockID === NoID)

const magnifyOptions = computed(() => [
  { value: '', text: 'Disabled' },
  { value: 'modal', text: 'Modal' },
  { value: 'fullscreen', text: 'Fullscreen' },
])

const customIDState = computed(() => handle.handleState(props.block.meta.customID))
const customCSSClassState = computed(() => handle.classState(props.block.meta.customCSSClass))

const isRecordPage = computed(() => props.page && props.page.moduleID !== NoID)

const visibilityDocumentationURL = computed(() => {
  const [year, month] = VERSION.split('.')
  return `https://docs.cortezaproject.org/corteza-docs/${year}.${month}/integrator-guide/compose-configuration/page-layouts.html#visibility-condition`
})

const recordAutoCompleteParams = computed(() => processRecordAutoCompleteParams({ module: props.module, operators: true }))
const visibilityAutoCompleteParams = computed(() => processVisibilityAutoCompleteParams({ module: props.module }))

function processRecordAutoCompleteParams ({ module: mod, operators = false } = {}) {
  const { fields = [] } = mod || {}
  const moduleFields = fields.map(({ name }) => name)
  const userProperties = ($auth.user?.properties?.()) || []

  const recordSuggestions = isRecordPage.value && props.record
    ? [
        ...(['ownerID', 'recordID'].map(value => ({ interpolate: true, value }))),
        {
          interpolate: true,
          value: 'record',
          properties: [
            { value: 'values', properties: Object.keys(props.record.values) || [] },
            ...(props.record.properties || []),
          ],
        },
      ]
    : []

  return [
    ...recordSuggestions,
    ...(operators ? ['AND', 'OR'] : []),
    { interpolate: true, value: 'userID' },
    { interpolate: true, value: 'user', properties: userProperties },
    ...moduleFields,
  ]
}

function processVisibilityAutoCompleteParams ({ module: mod } = {}) {
  const { fields = [] } = mod || {}
  const moduleFields = fields.map(({ name }) => name)
  const userProperties = ($auth.user?.properties?.()) || []

  const recordSuggestions = isRecordPage.value && props.record
    ? [
        {
          value: 'record',
          properties: [
            { value: 'values', properties: Object.keys(props.record.values) || [] },
            ...(props.record.properties || []),
          ],
        },
      ]
    : []

  return [
    ...recordSuggestions,
    { value: 'user', properties: userProperties },
    { value: 'screen', properties: ['width', 'height', 'userAgent', 'breakpoint'] },
    ...moduleFields,
  ]
}

const currentRoles = computed({
  get: () => {
    if (!props.block.meta.visibility.roles) return []
    return props.block.meta.visibility.roles
  },
  set: (v) => { props.block.meta.visibility.roles = v },
})

onBeforeUnmount(() => {
  abortRequests()
  setDefaultValues()
})

onMounted(() => {
  fetchRoles()
})

function updateRefresh (e) {
  props.block.options.refreshRate = e.target.value < 5 && e.target.value > 0 ? 5 : e.target.value
}

function fetchRoles () {
  roles.value.processing = true
  const { response, cancel } = $SystemAPI.roleListCancellable({})
  abortableRequests.value.push(cancel)
  response()
    .then(({ set: r = [] }) => {
      roles.value.options = r.filter(({ meta }) => !(meta.context && meta.context.resourceTypes))
    }).finally(() => {
      roles.value.processing = false
    })
}

function abortRequests () {
  abortableRequests.value.forEach((cancel) => cancel())
}

function setDefaultValues () {
  roles.value = { processing: false, options: [] }
  abortableRequests.value = []
}
</script>

<style scoped>
.mh-tab {
  max-height: calc(100vh - 16rem);
}

.nav-tabs {
  padding: 0 1rem;
  border-bottom: 2px solid var(--bs-border-color, #dee2e6);
}

.nav-tabs .nav-link {
  padding: 0.625rem 1.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--bs-secondary-color, #6c757d);
  border: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
}

.nav-tabs .nav-link.active {
  color: var(--bs-primary, #0d6efd);
  background: transparent;
  border-bottom: 2px solid var(--bs-primary, #0d6efd);
  font-weight: 600;
}

.nav-tabs .nav-link:hover:not(.active) {
  border-bottom-color: var(--bs-border-color, #dee2e6);
}
</style>
