<template>
  <div
    v-if="namespace"
    class="py-3"
  >
    <Teleport to="#topbar-title">
      {{ $t('page.edit.edit') }}
    </Teleport>

    <Teleport to="#topbar-tools">
      <router-link
        v-if="page && isRecordPage && page.canUpdatePage"
        variant="primary"
        :disabled="!moduleEditor"
        :to="moduleEditor"
        class="btn btn-primary btn-sm d-flex align-items-center me-2"
      >
        {{ $t('navigation.editModule') }}
        <font-awesome-icon
          :icon="['far', 'edit']"
          class="ms-2"
        />
      </router-link>

      <div
        v-if="page && page.canUpdatePage"
        class="btn-group text-nowrap"
      >
        <router-link
          data-test-id="button-page-builder"
          :to="{ name: 'admin.pages.builder' }"
          class="btn btn-primary d-flex align-items-center"
        >
          {{ $t('label.pageBuilder') }}
          <font-awesome-icon
            :icon="['fas', 'tools']"
            class="ms-2"
          />
        </router-link>

        <router-link
          data-bs-toggle="tooltip"
          :title="$t('tooltip.view')"
          variant="primary"
          :to="pageViewer"
          class="btn btn-primary d-flex align-items-center"
          style="margin-left:2px;"
        >
          <font-awesome-icon
            :icon="['far', 'eye']"
          />
        </router-link>

        <page-translator
          v-model:page="page"
          v-model:page-layouts="layouts"
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
      class="container-fluid"
    >
      <div class="card shadow-sm">
        <div class="card-header d-flex py-3 align-items-center border-bottom">
          <div
            v-if="page && (page.canGrant || namespace.canGrant)"
            class="dropdown me-1"
          >
            <button
              data-test-id="dropdown-permissions"
              class="btn btn-outline-secondary btn-sm dropdown-toggle"
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
                  v-if="namespace.canGrant"
                  :title="page.title || page.handle || page.pageID"
                  :target="page.title || page.handle || page.pageID"
                  :resource="`corteza::compose:page/${namespace.namespaceID}/${page.pageID}`"
                  :button-label="$t('label.page')"
                  :show-button-icon="false"
                  class="dropdown-item"
                />
              </li>
              <li>
                <c-permissions-button
                  v-if="page.canGrant"
                  :title="page.title || page.handle || page.pageID"
                  :target="page.title || page.handle || page.pageID"
                  :resource="`corteza::compose:page-layout/${namespace.namespaceID}/${page.pageID}/*`"
                  :button-label="$t('label.pageLayout')"
                  :show-button-icon="false"
                  all-specific
                  class="dropdown-item"
                />
              </li>
            </ul>
          </div>
        </div>

        <div
          v-if="page"
          class="row px-4 py-3"
        >
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('label.title') }}
              </label>
              <input
                v-model="page.title"
                data-test-id="input-title"
                class="form-control mb-2"
                :class="{ 'is-invalid': titleState === false }"
                :placeholder="$t('placeholder.title')"
                required
              />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('label.handle') }}
              </label>
              <input
                v-model="page.handle"
                data-test-id="input-handle"
                class="form-control mb-2"
                :class="{ 'is-invalid': handleState === false }"
                :placeholder="$t('block.general.placeholder.handle')"
              />
              <div
                v-if="handleState === false"
                class="invalid-feedback d-block"
              >
                {{ $t('block.general.invalid-handle-characters') }}
              </div>
            </div>
          </div>

          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('label.description') }}
              </label>
              <textarea
                v-model="page.description"
                data-test-id="input-description"
                class="form-control"
                :placeholder="$t('edit.pageDescription')"
                rows="4"
              />
            </div>
          </div>
          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('prompt.label') }}
              </label>
              <div class="input-group">
                <c-rich-text-input
                  v-model="page.config.prompt"
                  :placeholder="$t('prompt.placeholder')"
                  body-class="form-control"
                  min-body-height="10rem"
                  output-format="markdown"
                  :to-markdown="htmlToMarkdown"
                  :to-html="markdownToHtml"
                  :labels="{
                    urlPlaceholder: $t('content.urlPlaceholder'),
                    ok: $t('content.ok'),
                  }"
                />
                <page-translator
                  v-model:page="page"
                  v-model:page-layouts="layouts"
                  button-variant="extra-light"
                  style="margin-left:2px;"
                />
              </div>
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="d-flex align-items-center form-label text-primary">
                {{ $t('icon.page') }}
                <button
                  data-bs-toggle="tooltip"
                  :title="$t('icon.configure')"
                  class="btn btn-outline-light d-flex align-items-center px-1 text-primary border-0 ms-1"
                  @click="openIconModal"
                >
                  <font-awesome-icon
                    :icon="['far', 'edit']"
                  />
                </button>
              </label>

              <font-awesome-icon
                v-if="icon.type === 'fontawesome'"
                :icon="pageIconFA"
                class="fs-3"
              />
              <img
                v-else-if="icon.src"
                :src="pageIcon"
                width="auto"
                height="50"
              >

              <span v-else>
                {{ $t('icon.noIcon') }}
              </span>
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('edit.otherOptions') }}
              </label>
              <div
                v-if="!isRecordPage"
                class="form-check"
              >
                <input
                  id="visible"
                  v-model="page.visible"
                  data-test-id="checkbox-page-visibility"
                  type="checkbox"
                  class="form-check-input"
                />
                <label
                  class="form-check-label"
                  for="visible"
                >
                  {{ $t('edit.visible') }}
                </label>
              </div>

              <div class="form-check">
                <input
                  id="navItemExpanded"
                  v-model="page.config.navItem.expanded"
                  data-test-id="checkbox-show-sub-pages-in-sidebar"
                  type="checkbox"
                  class="form-check-input"
                />
                <label
                  class="form-check-label"
                  for="navItemExpanded"
                >
                  {{ $t('showSubPages') }}
                </label>
              </div>

              <div
                v-if="isRecordPage"
                class="form-check"
              >
                <input
                  id="notificationsEnabled"
                  v-model="page.meta.notifications.enabled"
                  data-test-id="checkbox-page-notifications-enabled"
                  type="checkbox"
                  class="form-check-input"
                />
                <label
                  class="form-check-label"
                  for="notificationsEnabled"
                >
                  {{ $t('edit.notifications.enabled') }}
                </label>
              </div>
            </div>
          </div>

          <div class="col-12">
            <hr>

            <div class="mb-0">
              <label class="form-label text-primary">
                {{ $t('page-layout.layouts') }}
              </label>
              <c-form-table-wrapper
                :labels="{ addButton: $t('label.add') }"
                @add-item="addLayout"
              >
                <table
                  v-if="layouts.length > 0"
                  class="table table-sm table-borderless"
                >
                  <thead>
                    <tr>
                      <th
                        scope="col"
                        style="width: 40px;"
                      />
                      <th
                        class="text-primary"
                        scope="col"
                        style="min-width: 300px;"
                      >
                        {{ $t('page-layout.title') }}
                      </th>
                      <th
                        class="text-primary"
                        scope="col"
                        style="min-width: 300px;"
                      >
                        {{ $t('page-layout.handle') }}
                      </th>
                      <th
                        scope="col"
                        style="min-width: 100px;"
                      />
                    </tr>
                  </thead>

                  <draggable
            item-key="id"
                    v-model="layouts"
                    handle=".grab"
                    group="layouts"
                    tag="tbody"
                  >
                    <template #item="{ element, index }">
                      <tr
                        :key="index"
                      >
                        <td class="grab text-center align-middle">
                          <font-awesome-icon
                            :icon="['fas', 'bars']"
                            class="text-secondary"
                          />
                        </td>
                        <td class="align-middle">
                          <div class="input-group">
                            <c-input-expression
                              v-if="isRecordPage && element.config.useTitle"
                              v-model="element.meta.title"
                              auto-complete
                              :suggestion-params="recordAutoCompleteParams"
                              class="flex-grow-1"
                              @input="element.meta.updated = true"
                            />
                            <input
                              v-else
                              v-model="element.meta.title"
                              class="form-control"
                              :class="{ 'is-invalid': layoutTitleState(element.meta.title) === false }"
                              @input="element.meta.updated = true"
                            />
                            <page-layout-translator
                              :page-layout="element"
                              :disabled="element.pageLayoutID === '0'"
                              highlight-key="meta.title"
                            />
                          </div>
                        </td>
                        <td class="align-middle">
                          <div class="input-group">
                            <input
                              v-model="element.handle"
                              class="form-control"
                              :class="{ 'is-invalid': layoutHandleState(element.handle) === false }"
                              @input="element.meta.updated = true"
                            />
                            <button
                              data-bs-toggle="tooltip"
                              :title="$t('page-layout.tooltip.configure')"
                              class="btn btn-extra-light d-flex align-items-center px-3"
                              @click="configureLayout(index)"
                            >
                              <font-awesome-icon
                                :icon="['fas', 'wrench']"
                              />
                            </button>
                            <router-link
                              data-bs-toggle="tooltip"
                              :title="$t('page-layout.tooltip.builder')"
                              class="btn btn-primary d-flex align-items-center"
                              :class="{ disabled: element.pageLayoutID === '0' }"
                              :to="{ name: 'admin.pages.builder', query: { layoutID: element.pageLayoutID } }"
                            >
                              <font-awesome-icon
                                :icon="['fas', 'tools']"
                              />
                            </router-link>
                          </div>
                        </td>
                        <td class="text-end align-middle"
                          style="min-width: 100px;"
                        >
                          <c-permissions-button
                            v-if="page.canGrant && element.pageLayoutID !== '0'"
                            button-variant="outline-extra-light"
                            size="sm"
                            :title="element.meta.title || element.handle || element.pageLayoutID"
                            :target="element.meta.title || element.handle || element.pageLayoutID"
                            :tooltip="$t('permissions.resources.compose.page-layout.tooltip')"
                            :resource="`corteza::compose:page-layout/${element.namespaceID}/${element.pageID}/${element.pageLayoutID}`"
                            class="text-dark border-0 me-2"
                          />
                          <c-input-confirm
                            show-icon
                            @confirmed="removeLayout(index)"
                          />
                        </td>
                      </tr>
                    </template>
                  </draggable>
                </table>
              </c-form-table-wrapper>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="layoutEditor.layout"
      class="modal fade show d-block"
      tabindex="-1"
      style="background-color: rgba(0,0,0,0.5);"
    >
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header p-3 border-bottom-0">
            <h5 class="modal-title">
              {{ $t('page-layout.configure', { title: ((layoutEditor.layout || {}).meta || {}).title, interpolation: { escapeValue: false } }) }}
            </h5>
            <button
              type="button"
              class="btn-close"
              @click="layoutEditor.layout = undefined"
            />
          </div>
          <div class="modal-body">
            <h5 class="mb-3">
              {{ $t('page-layout.general') }}
            </h5>

            <div class="row">
              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('page-layout.title') }}
                  </label>
                  <div class="input-group">
                    <c-input-expression
                      v-if="isRecordPage && layoutEditor.layout.config.useTitle"
                      v-model="layoutEditor.layout.meta.title"
                      auto-complete
                      :state="layoutTitleState(layoutEditor.layout.meta.title)"
                      :suggestion-params="recordAutoCompleteParams"
                      :placeholder="$t('page-layout.title')"
                      class="flex-grow-1"
                      @input="layoutEditor.layout.meta.updated = true"
                    />
                    <input
                      v-else
                      v-model="layoutEditor.layout.meta.title"
                      class="form-control"
                      :class="{ 'is-invalid': layoutTitleState(layoutEditor.layout.meta.title) === false }"
                      @input="layoutEditor.layout.meta.updated = true"
                    />
                    <page-layout-translator
                      :page-layout="layoutEditor.layout"
                      :disabled="layoutEditor.layout.pageLayoutID === '0'"
                      highlight-key="meta.title"
                    />
                  </div>
                </div>
              </div>

              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">
                    {{ $t('page-layout.handle') }}
                  </label>
                  <input
                    v-model="layoutEditor.layout.handle"
                    class="form-control"
                    :class="{ 'is-invalid': layoutHandleState(layoutEditor.layout.handle) === false }"
                    @input="layoutEditor.layout.meta.updated = true"
                  />
                </div>
              </div>
            </div>

            <div
              v-if="isRecordPage"
              class="mb-3"
            >
              <label class="form-label text-primary ms-auto mt-2">
                {{ $t('page-layout.useTitle') }}
              </label>
              <c-input-checkbox
                v-model="layoutEditor.layout.config.useTitle"
                switch
                :labels="checkboxLabel"
              />

              <i18next
                path="page-layout.tooltip.title"
                tag="small"
                class="text-muted"
              >
                <code>${record.values.fieldName}</code>
                <code>${recordID}</code>
                <code>${ownerID}</code>
                <span><code>${userID}</code>, <code>${user.name}</code></span>
              </i18next>
            </div>

            <hr>

            <h5 class="mb-3">
              {{ $t('page-layout.visibility') }}
            </h5>

            <div class="mb-3">
              <label class="d-flex align-items-center form-label text-primary mb-0">
                {{ $t('page-layout.condition.label') }}
                <c-hint
                  :tooltip="$t('page-layout.tooltip.performance.condition')"
                  icon-class="text-warning"
                />

                <a
                  :href="visibilityDocumentationURL"
                  target="_blank"
                  class="btn btn-link ms-auto p-0"
                >
                  {{ $t('label.examples') }}
                </a>
              </label>
              <div class="input-group">
                <span class="input-group-text btn-extra-light">
                  ƒ
                </span>
                <c-input-expression
                  v-model="layoutEditor.layout.config.visibility.expression"
                  auto-complete
                  :placeholder="$t('page-layout.condition.placeholder')"
                  :suggestion-params="visibilityAutoCompleteParams"
                  class="flex-grow-1"
                />
              </div>

              <i18next
                v-if="isRecordPage"
                path="page-layout.condition.description.record-page"
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
                path="page-layout.condition.description.non-record-page"
                tag="small"
                class="text-muted"
              >
                <code>user.(userID/email...)</code>
                <code>screen.(width/height)</code>
                <code>user.email == "test@mail.com"</code>
                <code>screen.width &lt; 1024</code>
              </i18next>
            </div>

            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('page-layout.roles.label') }}
              </label>
              <span
                v-if="resolvingLayoutRoles"
                class="spinner-border"
              />

              <c-input-role
                v-else
                :value="getLayoutRoles()"
                :placeholder="$t('page-layout.roles.placeholder')"
                multiple
                @input="onLayoutRoleChange"
              />
            </div>

            <template v-if="isRecordPage">
              <hr>

              <h5 class="mb-3">
                {{ $t('page-layout.recordToolbar.label') }}
              </h5>

              <div class="mb-3">
                <label class="form-label text-primary">
                  {{ $t('page-layout.recordToolbar.buttons.label') }}
                </label>
                <div class="form-check">
                  <input
                    id="backEnabled"
                    v-model="layoutEditor.layout.config.buttons.back.enabled"
                    type="checkbox"
                    class="form-check-input"
                  />
                  <label
                    class="form-check-label"
                    for="backEnabled"
                  >
                    {{ $t('page-layout.recordToolbar.buttons.showBack') }}
                  </label>
                </div>

                <div class="form-check">
                  <input
                    id="deleteEnabled"
                    v-model="layoutEditor.layout.config.buttons.delete.enabled"
                    type="checkbox"
                    class="form-check-input"
                  />
                  <label
                    class="form-check-label"
                    for="deleteEnabled"
                  >
                    {{ $t('page-layout.recordToolbar.buttons.showDelete') }}
                  </label>
                </div>

                <div class="form-check">
                  <input
                    id="cloneEnabled"
                    v-model="layoutEditor.layout.config.buttons.clone.enabled"
                    type="checkbox"
                    class="form-check-input"
                  />
                  <label
                    class="form-check-label"
                    for="cloneEnabled"
                  >
                    {{ $t('page-layout.recordToolbar.buttons.showClone') }}
                  </label>
                </div>

                <div class="form-check">
                  <input
                    id="newEnabled"
                    v-model="layoutEditor.layout.config.buttons.new.enabled"
                    type="checkbox"
                    class="form-check-input"
                  />
                  <label
                    class="form-check-label"
                    for="newEnabled"
                  >
                    {{ $t('page-layout.recordToolbar.buttons.showNew') }}
                  </label>
                </div>

                <div class="form-check">
                  <input
                    id="editEnabled"
                    v-model="layoutEditor.layout.config.buttons.edit.enabled"
                    type="checkbox"
                    class="form-check-input"
                  />
                  <label
                    class="form-check-label"
                    for="editEnabled"
                  >
                    {{ $t('page-layout.recordToolbar.buttons.showEdit') }}
                  </label>
                </div>

                <div class="form-check">
                  <input
                    id="submitEnabled"
                    v-model="layoutEditor.layout.config.buttons.submit.enabled"
                    type="checkbox"
                    class="form-check-input"
                  />
                  <label
                    class="form-check-label"
                    for="submitEnabled"
                  >
                    {{ $t('page-layout.recordToolbar.buttons.showSave') }}
                  </label>
                </div>
              </div>

              <div class="mb-0">
                <label class="form-label text-primary">
                  {{ $t('page-layout.recordToolbar.actions.label') }}
                </label>
                <c-form-table-wrapper
                  :labels="{ addButton: $t('label.add') }"
                  @add-item="addLayoutAction"
                >
                  <table
                    v-if="layoutEditor.layout.config.actions.length > 0"
                    class="table table-sm table-borderless layout-actions"
                  >
                    <draggable
            item-key="id"
                      v-model="layoutEditor.layout.config.actions"
                      handle=".grab"
                      group="actions"
                      tag="tbody"
                    >
                      <template #item="{ element, index }">
                        <tr
                          :key="index"
                          :class="{ 'border-top border-light': index > 0 }"
                        >
                          <td style="width: 40px;">
                            <div
                              class="grab d-flex align-items-center justify-content-center"
                              style="height: calc(1.5em + 0.75rem + 45px);"
                            >
                              <font-awesome-icon
                                :icon="['fas', 'bars']"
                                class="text-secondary"
                              />
                            </div>
                          </td>
                          <td style="min-width: 250px;">
                            <div class="mb-1">
                              <label class="form-label text-primary">
                                {{ $t('page-layout.recordToolbar.actions.buttonLabel') }}
                              </label>
                              <input
                                v-model="element.meta.label"
                                class="form-control mb-1"
                              />
                            </div>
                            <div
                              v-if="element.kind === 'toLayout'"
                              class="mb-0"
                            >
                              <label class="form-label text-primary">
                                {{ $t('page-layout.recordToolbar.actions.toLayout.label') }}
                              </label>
                              <select
                                v-model="element.params.pageLayoutID"
                                class="form-select form-control"
                              >
                                <option
                                  v-for="opt in actionLayoutOptions"
                                  :key="opt.pageLayoutID"
                                  :value="opt.pageLayoutID"
                                >
                                  {{ opt.label }}
                                </option>
                              </select>
                            </div>
                            <div
                              v-if="element.kind === 'toURL'"
                              class="mb-0"
                            >
                              <label class="form-label text-primary">
                                {{ $t('page-layout.recordToolbar.actions.toURL.label') }}
                              </label>
                              <input
                                v-model="element.params.url"
                                type="url"
                                class="form-control"
                                :placeholder="$t('page-layout.recordToolbar.actions.toURL.placeholder')"
                              />
                            </div>
                          </td>
                          <td style="min-width: 250px;">
                            <div class="mb-1">
                              <label class="form-label text-primary">
                                {{ $t('page-layout.recordToolbar.actions.kind.label') }}
                              </label>
                              <select
                                v-model="element.kind"
                                class="form-select form-control mb-1"
                                @change="onActionKindChange(element)"
                              >
                                <option
                                  v-for="opt in actionKindOptions"
                                  :key="opt.value"
                                  :value="opt.value"
                                >
                                  {{ opt.text }}
                                </option>
                              </select>
                            </div>
                            <div
                              v-if="element.kind === 'toURL'"
                              class="mb-0"
                            >
                              <label class="form-label text-primary">
                                {{ $t('page-layout.recordToolbar.actions.openIn.label') }}
                              </label>
                              <select
                                v-model="element.params.openIn"
                                class="form-select form-control"
                              >
                                <option
                                  v-for="opt in actionOpenInOptions"
                                  :key="opt.value"
                                  :value="opt.value"
                                >
                                  {{ opt.text }}
                                </option>
                              </select>
                            </div>
                          </td>
                          <td style="min-width: 150px;">
                            <div class="mb-3">
                              <label class="form-label text-primary">
                                {{ $t('page-layout.recordToolbar.actions.variant') }}
                              </label>
                              <select
                                v-model="element.meta.style.variant"
                                class="form-select form-control"
                              >
                                <option
                                  v-for="opt in actionVariantOptions"
                                  :key="opt.value"
                                  :value="opt.value"
                                >
                                  {{ opt.text }}
                                </option>
                              </select>
                            </div>
                          </td>
                          <td style="min-width: 100px;">
                            <div class="mb-3">
                              <label class="form-label text-primary">
                                {{ $t('page-layout.recordToolbar.actions.placement.label') }}
                              </label>
                              <select
                                v-model="element.placement"
                                class="form-select form-control"
                              >
                                <option
                                  v-for="opt in actionPlacementOptions"
                                  :key="opt.value"
                                  :value="opt.value"
                                >
                                  {{ opt.text }}
                                </option>
                              </select>
                            </div>
                          </td>
                          <td style="min-width: 80px;">
                            <div class="mb-3">
                              <label class="form-label text-primary text-center">
                                {{ $t('page-layout.recordToolbar.actions.visible') }}
                              </label>
                              <div
                                class="d-flex align-items-center justify-content-center"
                                style="height: calc(1.5em + 0.75rem + 2px);"
                              >
                                <div class="form-check">
                                  <input
                                    :id="'action-' + index"
                                    v-model="element.enabled"
                                    type="checkbox"
                                    class="form-check-input ms-2"
                                  />
                                </div>
                              </div>
                            </div>
                          </td>
                          <td style="min-width: 80px;">
                            <div
                              class="d-flex align-items-center justify-content-end"
                              style="height: calc(1.5em + 0.75rem + 45px);"
                            >
                              <c-input-confirm
                                show-icon
                                class="ms-2"
                                @confirmed="removeLayoutAction(index)"
                              />
                            </div>
                          </td>
                        </tr>
                      </template>
                    </draggable>
                  </table>
                </c-form-table-wrapper>
              </div>

              <hr>

              <h5 class="d-flex align-items-center mb-3">
                {{ $t('page-layout.requiredFields.label') }}

                <c-hint
                  :tooltip="$t('page-layout.tooltip.performance.requiredFields')"
                  icon-class="text-warning"
                />

                <a
                  :href="visibilityDocumentationURL"
                  target="_blank"
                  class="btn btn-link ms-auto p-0"
                >
                  {{ $t('label.examples') }}
                </a>
              </h5>

              <p class="text-muted">
                {{ $t('page-layout.requiredFields.description') }}
              </p>

              <c-form-table-wrapper
                :labels="{ addButton: $t('label.add') }"
                :disable-add-button="addRequiredFieldDisabled"
                @add-item="addRequiredField"
              >
                <table
                  v-if="validationRequiredFields.length > 0"
                  class="table table-sm table-borderless"
                >
                  <thead>
                    <tr>
                      <th
                        class="text-primary"
                        style="min-width: 250px;"
                      >
                        {{ $t('page-layout.requiredFields.field') }}
                      </th>
                      <th
                        class="text-primary"
                        style="min-width: 300px;"
                      >
                        {{ $t('page-layout.requiredFields.condition') }}
                      </th>
                      <th />
                    </tr>
                  </thead>

                  <tbody>
                    <tr
                      v-for="(requiredField, index) in validationRequiredFields"
                      :key="index"
                    >
                      <td style="min-width: 250px;">
                        <c-input-select
                          v-model="requiredField.field"
                          :options="availableModuleFields"
                          :placeholder="$t('page-layout.requiredFields.selectPlaceholder')"
                          :selectable="option => isFieldSelectableForRequired(option, requiredField)"
                          :get-option-label="getFieldLabel"
                          :get-option-key="getFieldKey"
                          :clearable="false"
                          :reduce="option => option.name || option.fieldID"
                        />
                      </td>

                      <td
                        class="align-middle"
                        style="min-width: 300px;"
                      >
                        <div class="input-group">
                          <span class="input-group-text btn-extra-light">
                            ƒ
                          </span>
                          <c-input-expression
                            v-model="requiredField.condition"
                            auto-complete
                            :placeholder="$t('page-layout.requiredFields.conditionPlaceholder')"
                            :suggestion-params="visibilityAutoCompleteParams"
                            class="flex-grow-1"
                          />
                        </div>
                      </td>

                      <td
                        class="text-end"
                        style="width: 4rem;"
                      >
                        <c-input-confirm
                          show-icon
                          @confirmed="removeRequiredField(index)"
                        />
                      </td>
                    </tr>
                  </tbody>
                </table>
              </c-form-table-wrapper>

              <i18next
                path="page-layout.condition.description.record-page"
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
            </template>
          </div>
          <div class="modal-footer">
            <button
              class="btn btn-outline-secondary"
              @click="layoutEditor.layout = undefined"
            >
              {{ $t('label.cancel') }}
            </button>
            <button
              class="btn btn-primary"
              :disabled="!layoutEditor.layout.meta.title"
              @click="updateLayout(); layoutEditor.layout = undefined"
            >
              {{ $t('label.saveAndClose') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      class="modal fade"
      :class="{ show: showIconModal }"
      :style="{ display: showIconModal ? 'block' : 'none' }"
      tabindex="-1"
    >
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ $t('icon.configure') }}
            </h5>
            <button
              type="button"
              class="btn-close"
              @click="closeIconModal"
            />
          </div>
          <div class="modal-body">
            <div class="mb-0">
              <label class="form-label text-primary">
                {{ $t('icon.upload') }}
              </label>
              <c-uploader
                :endpoint="iconUploadEndpoint"
                :accepted-files="['image/*']"
                param-name="icon"
                :labels="{
                  uploading: $t('label.uploading'),
                  placeholder: $t('label.dropFiles'),
                  fileTypeNotAllowed: $t('label.fileTypeNotAllowed'),
                }"
                @upload="uploadAttachment"
              />

              <div class="mb-3 my-2">
                <label class="form-label text-primary">
                  {{ $t('url.label') }}
                </label>
                <div class="input-group">
                  <input
                    v-model="linkUrl"
                    class="form-control"
                    :disabled="isIconSet"
                  />
                  <a
                    :disabled="!linkUrl"
                    class="btn btn-outline-secondary d-flex align-items-center btn-light"
                    :href="linkUrl"
                    target="_blank"
                  >
                    <font-awesome-icon :icon="['fas', 'external-link-alt']" />
                  </a>
                </div>
              </div>
            </div>

            <hr>

            <div class="mb-3">
              <label class="form-label text-primary">
                Font Awesome icon
              </label>
              <input
                v-model="faIcon"
                class="form-control mb-2"
                placeholder="fas fa-file-alt"
              />
              <div class="d-flex flex-wrap gap-1">
                <button
                  v-for="fi in faIconList"
                  :key="fi"
                  type="button"
                  class="btn btn-outline-secondary btn-sm d-flex align-items-center"
                  :class="{ 'btn-primary text-white': faIcon === fi }"
                  @click="faIcon = fi"
                >
                  <font-awesome-icon
                    :icon="fi.split(' ')"
                    class="me-1"
                  />
                  {{ fi.split(' ').pop() }}
                </button>
              </div>
            </div>

            <template v-if="attachments.length > 0">
              <hr>

              <div class="mb-3">
                <label class="form-label text-primary">
                  {{ $t('icon.list') }}
                </label>
                <div
                  v-if="processingIcon"
                  class="d-flex align-items-center justify-content-center h-100"
                >
                  <span class="spinner-border" />
                </div>

                <div
                  v-else
                  class="d-flex flex-wrap"
                >
                  <img
                    v-for="a in attachments"
                    :key="a.attachmentID"
                    :src="a.src"
                    :alt="a.name"
                    width="auto"
                    height="50"
                    :class="{ 'selected-icon': selectedAttachmentID === a.attachmentID }"
                    class="rounded pointer m-2"
                    @click="toggleSelectedIcon(a.attachmentID)"
                  />
                </div>
              </div>
            </template>
          </div>
          <div class="modal-footer d-flex align-items-center">
            <c-input-confirm
              v-if="attachments && selectedAttachmentID"
              :disabled="(attachments && !selectedAttachmentID) || processingIcon"
              :text="$t('icon.delete')"
              size="md"
              variant="danger"
              @confirmed="deleteIcon"
            />

            <div class="ms-auto">
              <button
                class="btn btn-outline-secondary"
                @click="closeIconModal"
              >
                {{ $t('label.cancel') }}
              </button>

              <button
                class="btn btn-primary ms-2"
                @click="saveIconModal"
              >
                {{ $t('label.saveAndClose') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <Teleport to="#admin-toolbar">
      <editor-toolbar
        :hide-delete="hideDelete"
        :hide-clone="hideClone"
        :hide-save="hideSave"
        :disable-save="disableSave"
        :processing="processing"
        :processing-save="processingSave"
        :processing-save-and-close="processingSaveAndClose"
        :processing-delete="processingDelete"
        :processing-clone="processingClone"
        @clone="handleClone()"
        @delete="handleDeletePage()"
        @save="handleSave()"
        @saveAndClose="handleSave({ closeOnSuccess: true })"
        @back="router.push(previousPage || { name: 'admin.pages' })"
      >
        <template #delete>
          <div
            v-if="showDeleteDropdown"
            class="dropdown"
          >
            <button
              data-test-id="dropdown-delete"
              class="btn btn-danger btn-lg dropdown-toggle"
              type="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              {{ $t('label.delete') }}
            </button>
            <ul class="dropdown-menu m-0">
              <li>
                <button
                  data-test-id="dropdown-item-delete-update-parent-of-sub-pages"
                  class="dropdown-item"
                  @click="handleDeletePage('rebase')"
                >
                  {{ $t('delete.rebase') }}
                </button>
              </li>
              <li>
                <button
                  data-test-id="dropdown-item-delete-sub-pages"
                  class="dropdown-item"
                  @click="handleDeletePage('cascade')"
                >
                  {{ $t('delete.cascade') }}
                </button>
              </li>
            </ul>
          </div>
        </template>
      </editor-toolbar>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useStore } from '../../../store'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getCurrentInstance } from 'vue'
import { NoID, compose } from 'corteza-lib/js/dist'
import { components, handle } from 'corteza-lib/vue/dist'
import { htmlToMarkdown, markdownToHtml } from '../../../lib/markdown'
import EditorToolbar from 'corteza-webapp-compose/src/components/Admin/EditorToolbar'
import PageTranslator from 'corteza-webapp-compose/src/components/Admin/Page/PageTranslator'
import PageLayoutTranslator from 'corteza-webapp-compose/src/components/Admin/PageLayout/PageLayoutTranslator'
import pages from 'corteza-webapp-compose/src/mixins/pages'
import { isEqual } from 'lodash'
import Draggable from 'vuedraggable'
import ModuleTranslator from 'corteza-webapp-compose/src/components/Admin/Module/ModuleTranslator.vue'

const { CInputRole, CInputExpression, CUploader, CInputSelect, CRichTextInput } = components
const { t } = useI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()

const { $ComposeAPI, $SystemAPI, $Settings } = getCurrentInstance().appContext.config.globalProperties

const props = defineProps({
  namespace: {
    type: compose.Namespace,
    required: true,
  },
  pageID: {
    type: String,
    required: true,
  },
})

const processing = ref(false)
const processingIcon = ref(false)
const processingSave = ref(false)
const processingSaveAndClose = ref(false)
const processingClone = ref(false)
const processingDelete = ref(false)
const loading = ref(false)
const page = ref(new compose.Page())
const initialPageState = ref(new compose.Page())
const showIconModal = ref(false)
const attachments = ref([])
const selectedAttachmentID = ref('')
const linkUrl = ref('')
const faIcon = ref('')
const layouts = ref([])
const layoutEditor = ref({ index: undefined, layout: undefined })
const resolvingLayoutRoles = ref(false)
const resolvedRoles = ref({})
const removedLayouts = ref(new Set())

const checkboxLabel = computed(() => ({
  on: t('label.yes'),
  off: t('label.no'),
}))

const pagesStore = computed(() => store.getters['page/set'])
const previousPage = computed(() => store.getters['ui/previousPage'])
const getModuleByID = computed(() => store.getters['module/getByID'])

const titleState = computed(() => page.value.title.length > 0 ? null : false)
const handleState = computed(() => handle.handleState(page.value.handle))

const pageViewer = computed(() => {
  const { pageID } = page.value
  const name = isRecordPage.value ? 'page.record.create' : 'page'
  return { name, params: { pageID } }
})

const moduleEditor = computed(() => {
  if (!module.value) return undefined
  return { name: 'admin.modules.edit', params: { moduleID: module.value.moduleID } }
})

const isRecordPage = computed(() => page.value && page.value.moduleID !== NoID)

const module = computed(() => {
  if (isRecordPage.value) {
    return getModuleByID.value(page.value.moduleID)
  }
  return undefined
})

const hasChildren = computed(() => page.value ? pagesStore.value.some(({ selfID }) => selfID === page.value.pageID) : false)

const disableSave = computed(() => {
  return !page.value || [titleState.value, handleState.value].includes(false) || layouts.value.some(l => !l.meta.title || handle.handleState(l.handle) === false)
})

const hideDelete = computed(() => !page.value || hasChildren.value || !page.value.canDeletePage || !!page.value.deletedAt)
const hideSave = computed(() => !page.value || !page.value.canUpdatePage)
const hideClone = computed(() => !page.value || page.value.moduleID !== NoID)

const showDeleteDropdown = computed(() => hasChildren.value && page.value.canDeletePage && !page.value.deletedAt)

const iconUploadEndpoint = computed(() => $ComposeAPI.baseURL + $ComposeAPI.iconUploadEndpoint({ namespaceID: props.namespace?.namespaceID }))

const icon = computed({
  get () { return page.value.config.navItem.icon || {} },
  set (v) { page.value.config.navItem.icon = v },
})

const isIconSet = computed(() => !!selectedAttachmentID.value)

const pageIcon = computed(() => {
  if (!icon.value.src) return undefined
  return icon.value.type === 'link' ? icon.value.src : makeAttachmentUrl(icon.value.src)
})

const pageIconFA = computed(() => {
  if (icon.value.type !== 'fontawesome' || !icon.value.src) return ['fas', 'file-alt']
  const parts = icon.value.src.split(' ')
  return parts.length >= 2 ? [parts[0], parts.slice(1).join(' ')] : ['fas', icon.value.src]
})

const faIconList = [
  'fas fa-file-alt',
  'fas fa-chart-bar',
  'fas fa-chart-pie',
  'fas fa-chart-line',
  'fas fa-database',
  'fas fa-cube',
  'fas fa-file',
  'fas fa-folder',
  'fas fa-home',
  'fas fa-cog',
  'fas fa-users',
  'fas fa-user',
  'fas fa-envelope',
  'fas fa-calendar-alt',
  'fas fa-comments',
  'fas fa-rss',
  'fas fa-map-marked-alt',
  'fas fa-image',
  'fas fa-tasks',
  'fas fa-search',
  'fas fa-star',
  'fas fa-heart',
  'fas fa-bell',
  'fas fa-clock',
  'fas fa-globe',
  'fas fa-book',
  'fas fa-code',
  'fas fa-table',
  'fas fa-filter',
  'fas fa-sitemap',
  'fas fa-wrench',
  'fas fa-tools',
  'fas fa-landmark',
  'fas fa-vault',
  'fas fa-wallet',
  'fas fa-industry',
  'fas fa-globe',
  'fas fa-certificate'  ,
  'fas fa-city',
  'fas fa-compass',
  'fas fa-copyright',
  'fas fa-fax',
  'fas fa-network-wired',
  'fas fa-percent',
  'fas fa-person-chalkboard',
  'fas fa-scale-balanced',
  'fas fa-scale-unbalanced',
  'fas fa-sitemap',
  'fas fa-table',
  'fas fa-tag',
  'fas fa-timeline',
  'fas fa-arrows-spin',
  'fas fa-arrows-to-dot',
  'fas fa-bars-progress',
  'fas fa-box-archive',
  'fas fa-building',
  'fas fa-briefcase',
  'fas fa-bullhorn',
  'fas fa-bullseye',
  'fas fa-calendar-days',
  'fas fa-certificate',
  'fas fa-bell',
  'fas fa-circle-exclamation',
  'fas fa-circle-radiation',
  'fas fa-skull-crossbones',
  'fas fa-triangle-exclamation',
  'fas fa-handshake',
  'fas fa-credit-card',
  'fas fa-file-invoice',
  'fas fa-money-bill-1-wave',
  'fas fa-award',
  'fas fa-cubes-stacked',
  'fas fa-award',
  'fas fa-display',
  'fas fa-house-chimney',
  'fas fa-house-medical',
  'fas fa-kitchen-set',
  'fas fa-people-group',
  'fas fa-bicycle',
  'fas fa-dumbbell',
  'fas fa-heart-pulse',
  'fas fa-person-walking',
  'fas fa-shirt',
  'fas fa-tree',
  'fas fa-fish',
  'fas fa-fire',
  'fas fa-water',
  'fas fa-car',
  'fas fa-cart-shopping',
  'fas fa-shop',
  'fas fa-dolly',
  'fas fa-hands-praying',
  'fas fa-boxes-stacked'
]

const visibilityDocumentationURL = computed(() => {
  const [year, month] = VERSION.split('.')
  return `https://docs.cortezaproject.org/corteza-docs/${year}.${month}/integrator-guide/compose-configuration/page-layouts.html#visibility-condition`
})

const actionKindOptions = computed(() => [
  { value: 'toLayout', text: t('page-layout.recordToolbar.actions.kind.toLayout') },
  { value: 'toURL', text: t('page-layout.recordToolbar.actions.kind.toURL') },
])

const actionLayoutOptions = computed(() => [
  { pageLayoutID: '', label: t('page-layout.recordToolbar.actions.toLayout.placeholder') },
  ...layouts.value.filter(({ pageLayoutID }) => pageLayoutID !== NoID)
    .map(({ pageLayoutID, handle, meta }) => ({ pageLayoutID, label: meta.title || handle || pageLayoutID })),
])

const actionOpenInOptions = computed(() => [
  { value: 'sameTab', text: t('page-layout.recordToolbar.actions.openIn.sameTab') },
  { value: 'newTab', text: t('page-layout.recordToolbar.actions.openIn.newTab') },
])

const actionVariantOptions = computed(() => [
  { value: 'primary', text: t('variants.primary') },
  { value: 'secondary', text: t('variants.secondary') },
  { value: 'success', text: t('variants.success') },
  { value: 'warning', text: t('variants.warning') },
  { value: 'danger', text: t('variants.danger') },
  { value: 'info', text: t('variants.info') },
  { value: 'light', text: t('variants.light') },
  { value: 'dark', text: t('variants.dark') },
])

const actionPlacementOptions = computed(() => [
  { value: 'start', text: t('page-layout.recordToolbar.actions.placement.start') },
  { value: 'center', text: t('page-layout.recordToolbar.actions.placement.center') },
  { value: 'end', text: t('page-layout.recordToolbar.actions.placement.end') },
])

// Mixins integration
const $auth = getCurrentInstance()?.appContext.config.globalProperties.$auth

function processVisibilityAutoCompleteParams ({ module: mod = module.value } = {}) {
  const { fields = [] } = mod || {}
  const moduleFields = fields.map(({ name }) => name)
  const userProperties = $auth?.user()?.properties() || []

  const recordSuggestions = isRecordPage.value && record.value
    ? [
        {
          value: 'record',
          properties: [
            { value: 'values', properties: Object.keys(record.value?.values || {}) },
            ...(record.value?.properties || []),
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

function processRecordAutoCompleteParams ({ module: mod = module.value, operators = false } = {}) {
  const { fields = [] } = mod || {}
  const moduleFields = fields.map(({ name }) => name)
  const userProperties = $auth?.user()?.properties() || []

  const recordSuggestions = isRecordPage.value && record.value
    ? [
        ...(['ownerID', 'recordID'].map(value => ({ interpolate: true, value }))),
        {
          interpolate: true,
          value: 'record',
          properties: [
            { value: 'values', properties: Object.keys(record.value?.values || {}) },
            ...(record.value?.properties || []),
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

const visibilityAutoCompleteParams = computed(() => processVisibilityAutoCompleteParams({ module: module.value }))
const recordAutoCompleteParams = computed(() => isRecordPage.value ? processRecordAutoCompleteParams({ module: module.value }) : [])

const availableModuleFields = computed(() => module.value ? module.value.fields : [])

const addRequiredFieldDisabled = computed(() => {
  if (!layoutEditor.value.layout || !module.value) return true
  return module.value.fields.length === 0
})

const validationRequiredFields = computed(() => {
  if (!layoutEditor.value.layout) return []
  const { config = {} } = layoutEditor.value.layout || {}
  const { validation = {} } = config || {}
  const { requiredFields = [] } = validation || {}
  return requiredFields
})

const currentLayoutRoles = computed({
  get () {
    if (!layoutEditor.value.layout) return []
    return layoutEditor.value.layout.config.visibility.roles
  },
  set (roles) {
    layoutEditor.value.layout.config.visibility.roles = roles
  },
})

watch(() => props.pageID, () => { fetchPage() }, { immediate: true })

onBeforeUnmount(() => { setDefaultValues() })

function toastErrorHandler (msg) { return (e) => {} }
function toastSuccess (msg) {}

function fetchPage (pageID = props.pageID) {
  loading.value = true
  processing.value = true

  if (pageID === NoID) {
    router.push({ name: 'admin.pages' })
    toastErrorHandler(t('notification.page.loadFailed'))({ message: t('notification.page.notFound') })
    return
  }

  const namespaceID = props.namespace?.namespaceID

  store.dispatch('page/findByID', { namespaceID, pageID, force: true }).then((p) => {
    page.value = p.clone()
    initialPageState.value = p.clone()
    document.title = t('label.app-name.page.edit', { label: p.title || p.handle, interpolation: { escapeValue: false } })
    return fetchAttachments()
  }).then(fetchLayouts)
    .catch(e => {
      toastErrorHandler(t('notification.page.loadFailed'))(e)
      router.push({ name: 'admin.pages' })
    })
    .finally(() => {
      setTimeout(() => {
        loading.value = false
        processing.value = false
      }, 300)
    })
}

async function fetchLayouts () {
  const namespaceID = props.namespace?.namespaceID
  return store.dispatch('pageLayout/findByPageID', { namespaceID, pageID: props.pageID, force: true }).then(l => {
    layouts.value = l.map((ly) => new compose.PageLayout(ly))
  })
}

async function resolveLayoutRoles () {
  if (currentLayoutRoles.value.length) {
    resolvingLayoutRoles.value = true
    Promise.all(currentLayoutRoles.value.map(roleID => {
      if (resolvedRoles.value[roleID]) return Promise.resolve()
      return $SystemAPI.roleRead({ roleID }).then(role => {
        resolvedRoles.value[roleID] = role
      })
    })).finally(() => { resolvingLayoutRoles.value = false })
  }
}

function getLayoutRoles () {
  return currentLayoutRoles.value.map(roleID => resolvedRoles.value[roleID]).filter(r => !!r)
}

function onLayoutRoleChange (roles) {
  roles.forEach(r => {
    if (!resolvedRoles.value[r.roleID]) resolvedRoles.value[r.roleID] = r
  })
  currentLayoutRoles.value = roles.map(r => r.roleID)
}

function addLayout () {
  layouts.value.push(new compose.PageLayout({ namespaceID: props.namespace?.namespaceID, pageID: props.pageID }))
}

function updateLayout () {
  layoutEditor.value.layout.meta.updated = true
  layoutEditor.value.layout.config.validation.requiredFields = validationRequiredFields.value.filter(rf => rf.field !== undefined)
  layouts.value.splice(layoutEditor.value.index, 1, layoutEditor.value.layout)
  layoutEditor.value.index = undefined
  layoutEditor.value.layout = undefined
}

function removeLayout (index) {
  const { pageLayoutID } = layouts.value[index] || {}
  if (pageLayoutID !== NoID) removedLayouts.value.add(layouts.value[index])
  layouts.value.splice(index, 1)
}

function configureLayout (index) {
  layoutEditor.value.index = index
  layoutEditor.value.layout = new compose.PageLayout(layouts.value[index])
  resolveLayoutRoles()
}

async function handleSaveLayouts () {
  return Promise.all([...removedLayouts.value].map(item => store.dispatch('pageLayout/delete', item))).then(() => {
    return Promise.all(layouts.value.map(layout => {
      if (layout.pageLayoutID === NoID) return store.dispatch('pageLayout/create', layout)
      else if (layout.meta.updated) return store.dispatch('pageLayout/update', layout)
      return Promise.resolve([])
    }))
  })
}

async function handlePageLayoutReorder () {
  const namespaceID = props.namespace?.namespaceID
  const pageIDs = layouts.value.map(({ pageLayoutID }) => pageLayoutID)
  return $ComposeAPI.pageLayoutReorder({ namespaceID, pageID: props.pageID, pageIDs }).then(() => {
    store.dispatch('pageLayout/load', { namespaceID, clear: true, force: true })
  })
}

function handleSave ({ closeOnSuccess = false } = {}) {
  const toggleProcessing = (value = true) => {
    if (closeOnSuccess) processingSaveAndClose.value = value
    else processingSave.value = value
  }

  processing.value = true
  toggleProcessing()

  const namespaceID = props.namespace?.namespaceID

  return saveIcon().then(iconData => {
    page.value.config.navItem.icon = iconData
    return store.dispatch('page/update', { namespaceID, ...page.value })
  }).then(p => {
    page.value = p.clone()
    initialPageState.value = p.clone()
    document.title = t('label.app-name.page.edit', { label: page.value.title || page.value.handle, interpolation: { escapeValue: false } })
    return handleSaveLayouts()
  }).then(handlePageLayoutReorder)
    .then(() => {
      fetchLayouts()
      removedLayouts.value = new Set()
      toastSuccess(t('notification.page.saved'))
      if (closeOnSuccess) router.push(previousPage.value || { name: 'admin.pages' })
    }).finally(() => {
      setTimeout(() => {
        processing.value = false
        toggleProcessing(false)
      }, 300)
    }).catch(toastErrorHandler(t('notification.page.saveFailed')))
}

function handleDeletePage (strategy = 'abort') {
  processingDelete.value = true
  store.dispatch('page/delete', { ...page.value, strategy }).then(() => {
    setTimeout(() => {
      toastSuccess(t('notification.page.deleted'))
      router.push({ name: 'admin.pages' })
    }, 300)
  })
    .catch(toastErrorHandler(t('notification.page.deleteFailed')))
    .finally(() => {
      setTimeout(() => { processingDelete.value = false }, 300)
    })
}

function handleClone () {
  const p = page.value.clone()
  const namespaceID = props.namespace?.namespaceID
  p.pageID = NoID
  p.title = `${page.value.title} (copy)`
  p.visible = false
  store.dispatch('page/create', { ...p, namespaceID }).then(({ pageID }) => {
    router.push({ name: 'admin.pages.edit', params: { pageID } })
  }).catch(toastErrorHandler(t('notification.page.saveFailed')))
}

function uploadAttachment ({ attachmentID }) {
  fetchAttachments()
  toggleSelectedIcon(attachmentID)
}

async function fetchAttachments () {
  processingIcon.value = true
  return $ComposeAPI.iconList({ sort: 'id DESC' })
    .then(({ set: atts = [] }) => {
      const baseURL = $ComposeAPI.baseURL
      attachments.value = []
      if (atts.length === 0) {
        icon.value = {}
        initialPageState.value.config.navItem.icon = {}
      } else {
        atts.forEach(a => {
          const src = !a.url.includes(baseURL) ? makeAttachmentUrl(a.url) : a.url
          attachments.value.push({ ...a, src })
        })
      }
    })
    .catch(toastErrorHandler(t('notification.page.iconFetchFailed')))
    .finally(() => { processingIcon.value = false })
}

function addLayoutAction () {
  layoutEditor.value.layout.addAction()
}

function removeLayoutAction (index) {
  layoutEditor.value.layout.config.actions.splice(index, 1)
}

function onActionKindChange (action) {
  if (action.kind === 'toURL' && !action.params.openIn) {
    action.params.openIn = 'sameTab'
  }
}

async function saveIcon () {
  if (icon.value.type === 'fontawesome') return icon.value
  return $ComposeAPI.pageUpdateIcon({
    namespaceID: props.namespace?.namespaceID,
    pageID: props.pageID,
    type: icon.value.type || 'link',
    source: icon.value.src,
  })
}

function toggleSelectedIcon (attachmentID = '') {
  selectedAttachmentID.value = selectedAttachmentID.value === attachmentID ? '' : attachmentID
}

function openIconModal () {
  linkUrl.value = icon.value.type === 'link' ? icon.value.src : ''
  faIcon.value = icon.value.type === 'fontawesome' ? icon.value.src : ''
  setCurrentIcon()
  showIconModal.value = true
}

function saveIconModal () {
  if (faIcon.value) {
    icon.value = { type: 'fontawesome', src: faIcon.value }
  } else if (selectedAttachmentID.value) {
    const src = (attachments.value.find(({ attachmentID }) => attachmentID === selectedAttachmentID.value) || {}).url
    icon.value = { type: 'attachment', src }
  } else if (linkUrl.value) {
    icon.value = { type: 'link', src: linkUrl.value }
  } else {
    icon.value = {}
  }
  showIconModal.value = false
}

function deleteIcon () {
  processingIcon.value = true
  return $ComposeAPI.iconDelete({ iconID: selectedAttachmentID.value }).then(() => {
    return fetchAttachments().then(() => {
      setCurrentIcon()
      toastSuccess(t('notification.page.iconDeleteSuccess'))
    })
  }).finally(() => { processingIcon.value = false })
    .catch(toastErrorHandler(t('notification.page.iconDeleteFailed')))
}

function closeIconModal () {
  linkUrl.value = icon.value.type === 'link' ? icon.value.src : ''
  faIcon.value = icon.value.type === 'fontawesome' ? icon.value.src : ''
  setCurrentIcon()
  showIconModal.value = false
}

function setCurrentIcon () {
  if (icon.value.type === 'fontawesome') return
  selectedAttachmentID.value = (attachments.value.find(a => a.url === icon.value.src) || {}).attachmentID
  if (!selectedAttachmentID.value) icon.value = {}
}

function makeAttachmentUrl (src) {
  return `${$ComposeAPI.baseURL}${src}`
}

function layoutTitleState (title) {
  return title ? null : false
}

function layoutHandleState (layoutHandle) {
  return handle.handleState(layoutHandle)
}

function addRequiredField () {
  layoutEditor.value.layout.config.validation.requiredFields.push({ field: undefined, condition: '' })
}

function removeRequiredField (index) {
  layoutEditor.value.layout.config.validation.requiredFields.splice(index, 1)
}

function isFieldSelectableForRequired (option, currentRequiredField) {
  const currentFieldId = option.isSystem ? option.name : option.fieldID
  const existingField = validationRequiredFields.value.find(rf => {
    const rfFieldId = rf.field
    return rfFieldId === currentFieldId && rf !== currentRequiredField
  })
  return !existingField
}

function getFieldLabel (field) {
  return field.label || field.name
}

function getFieldKey (field) {
  return field.isSystem ? field.name : field.fieldID
}

function setDefaultValues () {
  processing.value = false
  processingSaveAndClose.value = false
  processingSave.value = false
  processingClone.value = false
  processingDelete.value = false
  page.value = {}
  initialPageState.value = {}
  showIconModal.value = false
  attachments.value = []
  selectedAttachmentID.value = ''
  linkUrl.value = ''
  layouts.value = []
  layoutEditor.value = {}
  resolvedRoles.value = {}
  removedLayouts.value.clear()
  checkboxLabel.value = {}
}
</script>

<style lang="scss" scoped>
.selected-icon {
  outline: 2px solid var(--success);
}

.list-background {
  background-color: var(--body-bg);
}

.layout-actions {
  tr td {
    padding-bottom: 0.75rem;
  }
  tr:not(:first-child) td {
    padding-top: 0.75rem;
  }
}
</style>
