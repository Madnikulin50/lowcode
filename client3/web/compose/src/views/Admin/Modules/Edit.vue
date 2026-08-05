<template>
  <div
    v-if="namespace"
    class="py-3 d-flex flex-column flex-grow-1"
    style="min-height: 0"
  >
    <Teleport to="#topbar-title">
      {{ title }}
    </Teleport>

    <Teleport to="#topbar-tools">
      <div
        v-if="isEdit"
        class="btn-group me-1"
      >
        <router-link
          data-test-id="button-all-records"
          variant="primary"
          :to="allRecords"
          class="btn btn-primary btn-sm d-flex align-items-center"
        >
          {{ $t('allRecords.label') }}
          <font-awesome-icon
            :icon="['fas', 'columns']"
            class="ms-2"
          />
        </router-link>

        <module-translator
          v-if="module"
          v-model:module="trModule"
          button-variant="primary"
          style="margin-left:2px;"
        />
      </div>
    </Teleport>

    <div
      v-if="loading"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <span class="spinner-border" />
    </div>

    <div
      v-else
      class="d-flex flex-column flex-grow-1"
      style="min-height: 0"
      @submit.prevent="handleSave"
    >
      <div class="container-fluid flex-grow-1 d-flex flex-column" style="min-height: 0">
      <div class="row flex-grow-1" style="min-height: 0">
        <div class="col d-flex flex-column" style="min-height: 0">
          <div class="card shadow-sm d-flex flex-column flex-grow-1" style="min-height: 0">
            <div
              v-if="isEdit"
              class="card-header py-3"
            >
              <div class="d-flex align-items-center flex-fill-child gap-1">
                <button
                  v-if="federationEnabled"
                  data-test-id="button-federation-settings"
                  class="btn btn-outline-secondary btn-lg me-1"
                  @click="federationSettingsState.modal = true"
                >
                  <font-awesome-icon
                    :icon="['fas', 'share-alt']"
                  />
                  {{ $t('edit.federationSettings.title') }}
                </button>

                <button
                  v-if="discoveryEnabled"
                  data-test-id="button-discovery-settings"
                  class="btn btn-outline-secondary me-1"
                  @click="discoverySettingsState.modal = true"
                >
                  <font-awesome-icon
                    :icon="['fas', 'search-location']"
                  />
                   {{ $t('edit.discoverySettings.title') }}
                </button>

                <export
                  v-if="namespace.canExportModules"
                  :list="[module]"
                  type="module"
                  class="me-1"
                />

                <div
                  v-if="module.canGrant"
                  class="dropdown me-1"
                >
                  <button
                    data-test-id="dropdown-permissions"
                    class="btn btn-outline-secondary dropdown-toggle"
                    type="button"
                    data-bs-toggle="dropdown"
                    aria-expanded="false"
                  >
                    <font-awesome-icon :icon="['fas', 'lock']" />
                    <span>
                      {{ $t('label.permissions') }}
                    </span>
                  </button>
                  <ul class="dropdown-menu m-0">
                    <li>
                      <c-permissions-button
                        :title="module.name || module.handle || module.moduleID"
                        :target="module.name || module.handle || module.moduleID"
                        :resource="`corteza::compose:module/${namespace.namespaceID}/${module.moduleID}`"
                        :button-label="$t('label.module.single')"
                        :show-button-icon="false"
                        class="dropdown-item"
                      />
                    </li>
                    <li>
                      <c-permissions-button
                        :title="module.name || module.handle || module.moduleID"
                        :target="module.name || module.handle || module.moduleID"
                        :resource="`corteza::compose:module-field/${namespace.namespaceID}/${module.moduleID}/*`"
                        :button-label="$t('label.field')"
                        :show-button-icon="false"
                        all-specific
                        class="dropdown-item"
                      />
                    </li>
                    <li>
                      <c-permissions-button
                        :title="module.name || module.handle || module.moduleID"
                        :target="module.name || module.handle || module.moduleID"
                        :resource="`corteza::compose:record/${namespace.namespaceID}/${module.moduleID}/*`"
                        :button-label="$t('label.record')"
                        :show-button-icon="false"
                        all-specific
                        class="dropdown-item"
                      />
                    </li>
                  </ul>
                </div>

                <related-pages
                  :namespace="namespace"
                  :module="module"
                  size="lg"
                  class="d-flex ms-auto"
                />
              </div>
            </div>

            <div class="d-flex flex-column flex-grow-1" style="min-height: 0">
              <ul class="nav nav-tabs card-header-tabs">
                <li class="nav-item">
                  <a
                    class="nav-link"
                    :class="{ active: activeTab === 0 }"
                    @click.prevent="activeTab = 0"
                  >
                    {{ $t('edit.fields.label') }}
                  </a>
                </li>
                <li
                  v-if="isBasic || isDbRef"
                  class="nav-item"
                >
                  <a
                    class="nav-link"
                    :class="{ active: activeTab === 1 }"
                    @click.prevent="activeTab = 1"
                  >
                    {{ $t('edit.config.dal.title') }}
                  </a>
                </li>
                <li
                  v-if="isDatasource"
                  class="nav-item"
                >
                  <a
                    class="nav-link"
                    :class="{ active: activeTab === 2 }"
                    @click.prevent="activeTab = 2"
                  >
                    {{ $t('edit.config.datasource.title') }}
                  </a>
                </li>
                <li
                  v-if="isBasic"
                  class="nav-item"
                >
                  <a
                    class="nav-link"
                    :class="{ active: activeTab === 3 }"
                    @click.prevent="activeTab = 3"
                  >
                    {{ $t('edit.config.uniqueValues.title') }}
                  </a>
                </li>
                <li
                  v-if="isBasic"
                  class="nav-item"
                >
                  <a
                    class="nav-link"
                    :class="{ active: activeTab === 4 }"
                    @click.prevent="activeTab = 4"
                  >
                    {{ $t('edit.config.record-revisions.title') }}
                  </a>
                </li>
                <li class="nav-item">
                  <a
                    class="nav-link"
                    :class="{ active: activeTab === 5 }"
                    @click.prevent="activeTab = 5"
                  >
                    {{ $t('edit.config.privacy.title') }}
                  </a>
                </li>
                <li
                  v-if="isEdit"
                  class="nav-item"
                >
                  <a
                    class="nav-link"
                    :class="{ active: activeTab === 6 }"
                    @click.prevent="activeTab = 6"
                  >
                    ETL
                  </a>
                </li>
                <li
                  v-if="isConnector"
                  class="nav-item"
                >
                  <a
                    class="nav-link"
                    :class="{ active: activeTab === 8 }"
                    @click.prevent="activeTab = 8"
                  >
                    {{ $t('connector.title') }}
                  </a>
                </li>
                <li
                  v-if="module && module.issues.length > 0"
                  class="nav-item"
                >
                  <a
                    class="nav-link text-danger"
                    :class="{ active: activeTab === 9 }"
                    @click.prevent="activeTab = 9; checkAlterations()"
                  >
                    {{ $t('edit.issues.label', { count: module.issues.length }) }}
                  </a>
                </li>
              </ul>

              <div class="tab-content p-3 flex-grow-1 overflow-auto" style="min-height: 0">
                <div
                  class="tab-pane"
                  :class="{ active: activeTab === 0 }"
                >
                  <h5 class="mb-3">
                    {{ $t('edit.moduleInfo') }}
                  </h5>

                  <div class="row">
                    <div class="col-12 col-lg-6">
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('label.name') }}
                        </label>
                        <input
                          v-model="module.name"
                          data-test-id="input-module-name"
                          required
                          class="form-control"
                          :class="{ 'is-invalid': nameState === false }"
                          :placeholder="$t('placeholder.name')"
                        />
                      </div>
                    </div>

                    <div class="col-12 col-lg-6">
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('label.handle') }}
                        </label>
                        <input
                          v-model="module.handle"
                          data-test-id="input-module-handle"
                          class="form-control mb-2"
                          :class="{ 'is-invalid': handleState === false }"
                          :placeholder="$t('placeholder.handle')"
                        />
                        <div
                          v-if="handleState === false"
                          class="invalid-feedback d-block"
                        >
                          {{ $t('placeholder.invalid-handle-characters') }}
                        </div>
                      </div>
                    </div>
                  </div>

                  <div class="row">
                    <div class="col-12 col-lg-6">
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('type.label') }}
                        </label>
                        <select
                          v-model="moduleType"
                          class="form-select form-control form-select-sm"
                        >
                          <option
                            v-for="opt in typeValueOptions"
                            :key="opt.value"
                            :value="opt.value"
                          >
                            {{ opt.text }}
                          </option>
                        </select>
                      </div>
                    </div>
                  </div>
                  <div class="row">
                    <div class="col-12">
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('prompt.label') }}
                        </label>
                        <div class="input-group">
                          <c-rich-text-input
                            v-model="module.meta.prompt"
                            :placeholder="$t('prompt.placeholder')"
                            body-class="form-control"
                            style="width: 100%"
                            min-body-height="10rem"
                            output-format="markdown"
                            :to-markdown="htmlToMarkdown"
                            :to-html="markdownToHtml"
                            :labels="{
                              urlPlaceholder: $t('content.urlPlaceholder'),
                              ok: $t('content.ok'),
                            }"
                          />
                          <module-translator
                            v-if="module"
                            v-model:module="trModule"
                            highlight-key="meta.prompt"
                            :disabled="isNew"
                          />
                        </div>
                      </div>
                    </div>
                  </div>


                  <hr>

                  <h5 class="mb-3">
                    {{ $t('edit.manageRecordFields') }}
                  </h5>

                  <div class="row">
                    <c-form-table-wrapper
                      :labels="{ addButton: $t('edit.newField') }"
                      class="mb-2"
                      @add-item="handleNewField"
                    >
                      <table
                        v-if="module.fields.length > 0"
                        data-test-id="table-module-fields"
                        class="table table-sm table-borderless"
                      >
                        <thead>
                          <tr>
                            <th scope="col" />
                            <th
                              class="text-primary"
                              scope="col"
                            >
                              <div class="d-flex align-items-center">
                                {{ $t('edit.fields.columns.name.label') }}
                                <c-hint
                                  :tooltip="$t('edit.tooltip.name')"
                                />
                              </div>
                            </th>
                            <th
                              class="text-primary"
                              scope="col"
                            >
                              <div class="d-flex align-items-center">
                                {{ $t('edit.fields.columns.title.label') }}
                                <c-hint
                                  :tooltip="$t('edit.tooltip.title')"
                                />
                              </div>
                            </th>
                            <th
                              class="text-primary"
                              scope="col"
                            >
                              {{ $t('edit.fields.columns.type.label') }}
                            </th>
                            <th scope="col" />
                            <th scope="col" />
                            <th
                              class="text-primary text-center pe-3"
                              scope="col"
                            >
                              <div data-v-248edfd0-s="" class="d-flex align-items-center">
                                {{ $t('edit.label.required') }}
                                <c-hint
                                  :tooltip="$t('edit.tooltip.required')"
                                />
                              </div>
                            </th>
                            <th
                              class="text-primary text-center ps-2"
                              scope="col"
                            >
                              <div data-v-248edfd0-s="" class="d-flex align-items-center">
                                {{ $t('edit.label.multi') }}
                                <c-hint
                                  :tooltip="$t('edit.tooltip.multi')"
                                />
                              </div>
                            </th>
                            <th scope="col" />
                          </tr>
                        </thead>

                        <draggable
            item-key="id"
                          v-model="module.fields"
                          handle=".handle"
                          tag="tbody"
                        >
                          <template #item="{ element, index }">
                            <field-row-edit
                              :key="index"
                              v-model="module.fields[index]"
                              :can-grant="namespace.canGrant"
                              :has-records="hasRecords"
                              :module="module"
                              :is-duplicate="!!duplicateFields[index]"
                              @edit="handleFieldEdit(module.fields[index])"
                              @delete="module.fields.splice(index, 1)"
                              @updateKind="handleFieldKindUpdate(index)"
                            />
                          </template>
                        </draggable>
                      </table>
                    </c-form-table-wrapper>
                  </div>

                  <hr>

                  <h5
                    v-if="systemFieldsEnabled"
                    class="mb-3"
                  >
                    {{ $t('edit.systemFields') }}
                  </h5>

                  <div
                    v-if="systemFieldsEnabled"
                    class="row"
                  >
                    <c-form-table-wrapper hide-add-button>
                      <table class="table table-sm table-borderless">
                        <thead>
                          <tr>
                            <th
                              v-if="module.fields.length > 0"
                              scope="col"
                            />
                            <th
                              class="text-primary"
                              scope="col"
                            >
                              {{ $t('label.name') }}
                            </th>
                            <th
                              class="text-primary"
                              scope="col"
                            >
                              {{ $t('label.title') }}
                            </th>
                            <th
                              colspan="5"
                              class="text-primary"
                              scope="col"
                            >
                              {{ $t('label.type') }}
                            </th>
                          </tr>
                        </thead>

                        <tbody>
                          <tr
                            v-for="(field, index) in systemFields"
                            :key="index"
                          >
                            <td
                              v-if="module.fields.length > 0"
                              class="pe-2"
                              style="width: 30px;"
                            />
                            <td style="width: 250px;">
                              {{ field.name }}
                              <span
                                v-if="field.omit"
                                class="badge bg-info ms-2 align-middle"
                              >
                                {{ $t('unavailable') }}
                              </span>
                            </td>
                            <td style="width: 250px;">
                              {{ field.label }}
                            </td>
                            <td>
                              {{ field.kind }}
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </c-form-table-wrapper>
                  </div>
                </div>

                <div
                  v-if="isBasic || isDbRef"
                  class="tab-pane"
                  :class="{ active: activeTab === 1 }"
                >
                  <dal-settings
                    :module="module"
                  />
                </div>
                <div
                  v-if="isDatasource"
                  class="tab-pane"
                  :class="{ active: activeTab === 2 }"
                >
                  <data-source-settings
                    :module="module"
                  />
                </div>
                <div
                  v-if="isBasic"
                  class="tab-pane"
                  :class="{ active: activeTab === 3 }"
                >
                  <unique-values
                    :module="module"
                  />
                </div>
                <div
                  v-if="isBasic"
                  class="tab-pane"
                  :class="{ active: activeTab === 4 }"
                >
                  <record-revisions-settings
                    :module="module"
                  />
                </div>
                <div
                  class="tab-pane"
                  :class="{ active: activeTab === 5 }"
                >
                  <data-privacy-settings
                    v-if="connection"
                    :resource="module"
                    :connection="connection"
                    :sensitivity-levels="sensitivityLevels"
                    :max-level="maxLevelID"
                    :translations="{
                      sensitivity: {
                        label: $t('edit.config.privacy.sensitivity-level.label'),
                        description: $t('edit.config.privacy.sensitivity-level.description'),
                        placeholder: $t('edit.config.privacy.sensitivity-level.placeholder'),
                      },
                      usage: {
                        label: $t('edit.config.privacy.usage-disclosure.label'),
                      },
                    }"
                  />
                </div>
                <div
                  v-if="isEdit"
                  class="tab-pane"
                  :class="{ active: activeTab === 6 }"
                >
                  <etl-settings
                    v-if="activeTab === 6"
                    :namespace="namespace"
                    :module="module"
                  />
                </div>
                <div
                  v-if="isConnector"
                  class="tab-pane"
                  :class="{ active: activeTab === 8 }"
                >
                  <connector-settings
                    v-if="activeTab === 8"
                    :module="module"
                  />
                </div>
                <div
                  v-if="module && module.issues.length > 0"
                  class="tab-pane"
                  :class="{ active: activeTab === 9 }"
                >
                  <div
                    v-for="(issue, index) in module.issues"
                    :key="index"
                    class="alert alert-danger"
                    role="alert"
                  >
                    {{ issue.issue }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="updateField"
        class="modal fade show d-block"
        tabindex="-1"
        style="background-color: rgba(0,0,0,0.5);"
      >
        <div class="modal-dialog modal-lg">
          <div class="modal-content">
            <div class="modal-header p-3 pb-0 border-bottom-0">
              <h5 class="modal-title">
                {{ editModalTitle }}
              </h5>
              <button
                type="button"
                class="btn-close"
                @click="updateField = null"
              />
            </div>
            <div class="modal-body p-0 border-top-0">
              <field-configurator
                v-model:field="updateField"
                :namespace="namespace"
                :module="module"
                :connection="connection"
                :sensitivity-levels="sensitivityLevels"
                :has-records="hasRecords"
              />
            </div>
            <div class="modal-footer">
              <button
                class="btn btn-primary"
                @click="handleFieldSave(updateField)"
              >
                {{ $t('label.saveAndClose') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <dal-schema-alterations
        v-if="dalAlterationsEnabled"
        :batch="dalAlterationsState.batchID"
        :module="module"
        @hide="onDalAlterationsHide"
      />

      <federation-settings
        v-if="federationEnabled"
        :modal="federationSettingsState.modal"
        :module="module"
        @change="setFederationModal($event)"
      />

      <discovery-settings
        v-if="discoveryEnabled"
        :modal="discoverySettingsState.modal"
        :module="module"
        @update:modal="setDiscoveryModal($event)"
        @save="onDiscoverySettingsSave"
      />
    </div>
    </div>

    <Teleport to="#admin-toolbar">
      <editor-toolbar
        :processing="processing"
        :processing-save="processingSave"
        :processing-clone="processingClone"
        :processing-save-and-close="processingSaveAndClose"
        :processing-delete="processingDelete"
        :hide-delete="hideDelete"
        :hide-clone="!isEdit"
        :hide-save="hideSave"
        :disable-save="disableSave"
        @delete="handleDelete"
        @save="handleSave()"
        @clone="handleClone"
        @saveAndClose="handleSave({ closeOnSuccess: true })"
        @back="router.push(previousPage || { name: 'admin.modules' })"
      />
    </Teleport>
  </div>
</template>

<script setup>
import { ref, shallowRef, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useStore } from '../../../store'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getCurrentInstance } from 'vue'
import axios from 'axios'
import { isEqual } from 'lodash'
import draggable from 'vuedraggable'
import FieldConfigurator from 'corteza-webapp-compose/src/components/ModuleFields/Configurator'
import FieldRowEdit from 'corteza-webapp-compose/src/components/Admin/Module/FieldRowEdit'
import FederationSettings from 'corteza-webapp-compose/src/components/Admin/Module/FederationSettings'
import DalSchemaAlterations from 'corteza-webapp-compose/src/components/Admin/Module/DalSchemaAlterations'
import DiscoverySettings from 'corteza-webapp-compose/src/components/Admin/Module/DiscoverySettings'
import DalSettings from 'corteza-webapp-compose/src/components/Admin/Module/DalSettings'
import DataSourceSettings from 'corteza-webapp-compose/src/components/Admin/Module/DataSourceSettings'
import RecordRevisionsSettings from 'corteza-webapp-compose/src/components/Admin/Module/RecordRevisionsSettings'
import DataPrivacySettings from 'corteza-webapp-compose/src/components/Admin/Module/DataPrivacySettings'
import ModuleTranslator from 'corteza-webapp-compose/src/components/Admin/Module/ModuleTranslator'
import UniqueValues from 'corteza-webapp-compose/src/components/Admin/Module/UniqueValues'
import RelatedPages from 'corteza-webapp-compose/src/components/Admin/Module/RelatedPages'
import EtlSettings from 'corteza-webapp-compose/src/components/Admin/Module/ETLSettings'
import ConnectorSettings from 'corteza-webapp-compose/src/components/Admin/Module/ConnectorSettings'
import { compose, NoID } from 'corteza-lib/js/dist'
import { handle, components, composables } from 'corteza-lib/vue/dist'
import EditorToolbar from 'corteza-webapp-compose/src/components/Admin/EditorToolbar'
import Export from 'corteza-webapp-compose/src/components/Admin/Export'
import { htmlToMarkdown, markdownToHtml } from '../../../lib/markdown'

const { CRichTextInput } = components
const { t } = useI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()

const { $ComposeAPI, $SystemAPI, $Settings, $auth } = getCurrentInstance().appContext.config.globalProperties

const props = defineProps({
  namespace: {
    type: compose.Namespace,
    required: true,
  },
  moduleID: {
    type: String,
    required: false,
    default: NoID,
  },
})

const loading = ref(false)
const activeTab = ref(0)
const connection = ref(undefined)
const sensitivityLevels = ref([])
const updateField = ref(null)
const module = ref(undefined)
const initialModuleState = ref(undefined)
const hasRecords = ref(true)
const processing = ref(false)
const processingSave = ref(false)
const processingClone = ref(false)
const processingSaveAndClose = ref(false)
const processingDelete = ref(false)
const federationSettingsState = shallowRef({ modal: false })
const dalAlterationsState = shallowRef({ modal: false, batchID: undefined })
const discoverySettingsState = shallowRef({ modal: false })
const abortableRequests = ref([])

const pages = computed(() => store.getters['page/set'])
const previousPage = computed(() => store.getters['ui/previousPage'])

const title = computed(() => route.name === 'admin.modules.edit' ? t('edit.edit') : t('edit.create'))

const isNew = computed(() => props.moduleID === NoID)

const typeValueOptions = computed(() => [
  { value: 'basic', text: t('edit.type.basic') },
  { value: 'datasource', text: t('edit.type.datasource') },
  { value: 'dbref', text: t('edit.type.dbref') },
  { value: 'connector', text: t('edit.type.connector') },
])

const moduleType = computed({
  get () {
    const ds = (module.value && module.value.config.type) ?? 'basic'
    return ds
  },
  set (value) {
    if (module.value) module.value.config.type = value
  },
})

const isDatasource = computed(() => moduleType.value === 'datasource')
const isBasic = computed(() => moduleType.value === 'basic')
const isDbRef = computed(() => moduleType.value === 'dbref')
const isConnector = computed(() => moduleType.value === 'connector')

const trModule = computed({
  get () {
    if (!module.value) return new compose.Module()
    return module.value
  },
  set (v) {
    module.value = v
    store.dispatch('module/updateSet', v)
  },
})

const nameState = computed(() => (module.value && module.value.name.length > 0) ? null : false)

const handleState = computed(() => module.value ? handle.handleState(module.value.handle) : null)

const duplicateFields = computed(() => {
  const rtr = {}
  const ix = new Set()
  const { fields = [] } = module.value || {}
  fields.forEach((f, i) => {
    if (ix.has(f.name)) rtr[i] = f
    ix.add(f.name)
  })
  return rtr
})

const fieldsValid = computed(() => {
  const { fields = [] } = module.value || {}
  const valid = !fields.some(f => f.fieldID === NoID && !f.isValid)
  const unique = Object.keys(duplicateFields.value).length === 0
  return valid && unique
})

const systemFieldsEnabled = computed(() => isBasic.value)

const systemFields = computed(() => {
  if (!isBasic.value || !module.value) return []
  const systemFieldEncoding = module.value.config.dal.systemFieldEncoding || {}
  return module.value.systemFields().map(sf => {
    if (!sf) return false
    sf.label = t(`system.${sf.name}`)
    return { ...sf, ...(systemFieldEncoding[sf.name] || {}) }
  }).filter(sf => sf)
})

const editModalTitle = computed(() => {
  if (!updateField.value) return undefined
  const { name } = updateField.value
  return name ? t('edit.specificFieldSettings', { name: updateField.value.name }) : t('edit.moduleFieldSettings')
})

const federationEnabled = computed(() => isEdit.value && $Settings.get('federation.enabled', false))

const discoveryEnabled = computed(() => $Settings.get('discovery.enabled', false))

const dalAlterationsEnabled = computed(() => !isDatasource.value)

const hideDelete = computed(() => !isEdit.value || !module.value || !module.value.canDeleteModule || !!module.value.deletedAt)

const disableSave = computed(() => !module.value || [fieldsValid.value, nameState.value, handleState.value].includes(false))

const hideSave = computed(() => isEdit.value && module.value && !module.value.canUpdateModule)

const isEdit = computed(() => module.value && module.value.moduleID !== NoID)

const allRecords = computed(() => ({ name: 'admin.modules.record.list', params: { moduleID: props.moduleID } }))

const maxLevelID = computed(() => {
  const { sensitivityLevelID = NoID } = (connection.value && connection.value.config.privacy) || {}
  return sensitivityLevelID
})

watch(() => props.moduleID, () => {
  fetchModule(props.moduleID).then(() => {
    checkAlterations()
  })
}, { immediate: true })

watch(() => module.value ? module.value.config.dal.connectionID : undefined, (connectionID) => {
  fetchConnection(connectionID)
})

onMounted(() => {
  fetchSensitivityLevels()
})

onBeforeUnmount(() => {
  abortRequests()
  setDefaultValues()
})

function handleNewField () {
  if (module.value) module.value.fields.push(new compose.ModuleFieldString())
}

function handleFieldEdit (field) {
  updateField.value = compose.ModuleFieldMaker({ ...field })
}

function handleFieldKindUpdate (index) {
  const field = module.value.fields[index]
  module.value.fields.splice(index, 1, compose.ModuleFieldMaker({ ...field }))
}

function handleFieldSave (field) {
  const i = module.value.fields.findIndex(f => f.name === field.name)
  if (i > -1) {
    module.value.fields.splice(i, 1, field)
  }
}

function onDiscoverySettingsSave (changes) {
  if (module.value) module.value.config = { ...module.value.config, ...changes }
}

const { toastErrorHandler, toastSuccess } = composables.useToast()

function handleSave ({ mod = module.value, closeOnSuccess = false, isClone = false } = {}) {
  const resourceTranslationLanguage = currentLanguage()

  const toggleProcessing = (value = true) => {
    if (closeOnSuccess) processingSaveAndClose.value = value
    else if (isClone) processingClone.value = value
    else processingSave.value = value
  }

  processing.value = true
  toggleProcessing()

  if (mod.moduleID === NoID) {
    let fields = []
    const toBeUpdatedFields = []
    mod.fields.forEach(f => {
      if (f.kind === 'Record' && f.options.moduleID === '-1') {
        toBeUpdatedFields.push(f)
      } else {
        fields.push(f)
      }
    })
    store.dispatch('module/create', { ...mod, fields, resourceTranslationLanguage }).then(async m => {
      if (toBeUpdatedFields.length) {
        fields = [
          ...m.fields,
          ...toBeUpdatedFields.map(f => {
            f.options.moduleID = m.moduleID
            return f
          }),
        ]
        m = await store.dispatch('module/update', { ...m, fields })
      }
      loading.value = true
      module.value = new compose.Module({ ...m }, props.namespace)
      initialModuleState.value = module.value.clone()
      document.title = t('label.app-name.module.edit', { label: module.value.name, interpolation: { escapeValue: false } })
      toastSuccess(t('notification.module.created'))
      toggleProcessing(false)
      if (closeOnSuccess) {
        router.push({ name: 'admin.modules' })
      } else {
        router.push({ name: 'admin.modules.edit', params: { moduleID: module.value.moduleID } })
      }
    }).catch(e => {
      toastErrorHandler(t('notification.module.createFailed'))(e)
      processing.value = false
      toggleProcessing(false)
    })
  } else {
    store.dispatch('module/update', { ...mod, resourceTranslationLanguage }).then(m => {
      module.value = new compose.Module({ ...m }, props.namespace)
      initialModuleState.value = module.value.clone()
      document.title = t('label.app-name.module.edit', { label: module.value.name, interpolation: { escapeValue: false } })
      toastSuccess(t('notification.module.saved'))
      if (closeOnSuccess) {
        router.push({ name: 'admin.modules' })
      }
    }).catch(toastErrorHandler(t('notification.module.saveFailed')))
      .finally(() => {
        setTimeout(() => {
          processing.value = false
          toggleProcessing(false)
        }, 300)
      })
  }
}

function currentLanguage () {}

function setFederationModal (value) {
  federationSettingsState.value = { modal: !!value }
}

function setDiscoveryModal (value) {
  discoverySettingsState.value = { modal: !!value }
}

function onDalAlterationsHide () {
  fetchModule(props.moduleID)
  dalAlterationsState.value = { modal: false, batchID: undefined }
}

async function fetchModule (moduleID = props.moduleID) {
  activeTab.value = 0

  if (moduleID === NoID) {
    module.value = new compose.Module(
      { fields: [new compose.ModuleFieldString({ fieldID: NoID, name: '' })] },
      props.namespace,
    )
    initialModuleState.value = module.value.clone()
    document.title = t('label.app-name.module.create')
  } else {
    loading.value = true
    processing.value = true

    const params = { force: true, namespace: props.namespace, moduleID }

    return store.dispatch('module/findByID', params).then((m) => {
      module.value = m.clone()
      initialModuleState.value = module.value.clone()
      document.title = t('label.app-name.module.edit', { label: module.value.name, interpolation: { escapeValue: false } })
      const { moduleID, namespaceID, issues = [] } = module.value
      if (issues.length > 0) return
      const { response, cancel } = $ComposeAPI.recordListCancellable({ moduleID, namespaceID, limit: 1 })
      abortableRequests.value.push(cancel)
      return response().then(({ set = [] }) => {
        hasRecords.value = set.length > 0
      }).catch(e => {
        if (!axios.isCancel(e)) console.error(e)
      })
    }).catch(e => {
      toastErrorHandler(t('notification.module.loadFailed'))(e)
      router.push({ name: 'admin.modules' })
    }).finally(() => {
      setTimeout(() => {
        loading.value = false
        processing.value = false
      }, 300)
    })
  }
}

function checkAlterations () {
  const { issues = [] } = module.value || {}
  if (!issues.length) return
  const aux = (module.value.issues || []).map(({ meta }) => meta.batchID).filter(b => b)
  if (aux.length > 0) {
    dalAlterationsState.value = { modal: false, batchID: aux }
  } else {
    dalAlterationsState.value = { modal: false, batchID: undefined }
  }
}

function handleDelete () {
  processing.value = true
  processingDelete.value = true

  store.dispatch('module/delete', module.value).then(() => {
    module.value.deletedAt = new Date()
    const moduleRecordPage = pages.value.find(p => p.moduleID === module.value.moduleID)
    if (moduleRecordPage) {
      return store.dispatch('page/delete', { ...moduleRecordPage, strategy: 'rebase' })
    }
  }).then(() => {
    initialModuleState.value = module.value.clone()
    toastSuccess(t('notification.module.deleted'))
    router.push({ name: 'admin.modules' })
  }).catch(toastErrorHandler(t('notification.module.deleteFailed')))
    .finally(() => {
      processing.value = false
      processingDelete.value = false
    })
}

function handleClone () {
  const mod = module.value.clone()
  mod.moduleID = NoID
  mod.name = `${module.value.name} (copy)`
  mod.handle = ''
  handleSave({ mod, isClone: true })
}

async function fetchConnection (connectionID) {
  if (connectionID && connectionID !== NoID) {
    $SystemAPI.dalConnectionRead({ connectionID })
      .then(conn => {
        connection.value = conn
        if (initialModuleState.value && initialModuleState.value.config && initialModuleState.value.config.dal) {
          initialModuleState.value.config.dal.connectionID = conn.connectionID
        }
      })
      .catch(toastErrorHandler(t('notification.connection.read-failed')))
      .finally(() => { processing.value = false })
  }
}

async function fetchSensitivityLevels () {
  processing.value = true
  return $SystemAPI.dalSensitivityLevelList()
    .then(({ set = [] }) => { sensitivityLevels.value = set })
    .catch(toastErrorHandler(t('notification.sensitivity-level.fetch-failed')))
    .finally(() => { processing.value = false })
}

function setDefaultValues () {
  activeTab.value = 0
  connection.value = undefined
  sensitivityLevels.value = []
  updateField.value = null
  module.value = undefined
  initialModuleState.value = undefined
  hasRecords.value = true
  processing.value = false
  processingSaveAndClose.value = false
  processingSave.value = false
  federationSettingsState.value = {}
  discoverySettingsState.value = {}
  abortableRequests.value = []
}

function abortRequests () {
  abortableRequests.value.forEach((cancel) => cancel())
}
</script>
