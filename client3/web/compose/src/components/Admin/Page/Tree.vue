<template>
  <div>
    <sortable-tree
      v-if="list?.length"
      :draggable="namespace.canCreatePage"
      :data="{ children: list }"
      tag="ul"
      mixin-parent-key="parent"
      class="list-group"
      @changePosition="handleChangePosition"
    >
      <template
        #default="{item}"
      >
        <div
          v-if="item.pageID"
          no-gutters
          class="d-flex flex-wrap align-content-center justify-content-between"
        >
          <div
            class="px-2 flex-fill overflow-hidden text-truncate gap-1 rounded-lg"
            :class="{'grab': namespace.canCreatePage }"
          >
            {{ item.title }}
            <font-awesome-icon
              v-if="!item.visible && item.moduleID == '0'"
              class="text-danger"
              :icon="['fas', 'eye-slash']"
              :title="$t('notVisible')"
            />
            <span
              v-if="!isValid(item)"
              class="badge bg-danger ms-1"
            >
              {{ $t('invalid') }}
            </span>
          </div>

          <div class="actions px-2">
            <div
              v-if="item.canUpdatePage"
              class="btn-group btn-group-sm"
            >
              <router-link
                data-test-id="button-page-builder"
                class="btn btn-primary"
                :to="{name: 'admin.pages.builder', params: { pageID: item.pageID }}"
              >
                {{ $t('block.general.label.pageBuilder') }}
                <font-awesome-icon
                  :icon="['fas', 'tools']"
                  class="ms-2"
                />
              </router-link>

              <router-link
                data-test-id="button-page-view"
                class="btn btn-primary d-flex align-items-center"
                :title="$t('tooltip.view')"
                :to="pageViewer(item)"
                style="margin-left:2px;"
              >
                <font-awesome-icon
                  :icon="['far', 'eye']"
                />
              </router-link>

              <router-link
                data-test-id="button-page-edit"
                class="btn btn-primary d-flex align-items-center"
                :title="$t('tooltip.edit.page')"
                :to="{name: 'admin.pages.edit', params: { pageID: item.pageID }}"
                style="margin-left:2px;"
              >
                <font-awesome-icon
                  :icon="['far', 'edit']"
                />
              </router-link>
            </div>

            <div
              v-if="item.canGrant || namespace.canGrant"
              class="dropdown d-inline-block ms-2"
            >
              <button
                class="btn btn-extra-light btn-sm dropdown-toggle"
                type="button"
                data-bs-toggle="dropdown"
                :title="$t('permissions.resources.compose.page.tooltip')"
              >
                <font-awesome-icon :icon="['fas', 'lock']" />
              </button>
              <ul class="dropdown-menu">
                <li v-if="namespace.canGrant">
                  <c-permissions-button
                    :title="item.title || item.handle || item.pageID"
                    :target="item.title || item.handle || item.pageID"
                    :resource="`corteza::compose:page/${namespace.namespaceID}/${item.pageID}`"
                    :button-label="$t('label.page')"
                    :show-button-icon="false"
                    class="dropdown-item"
                  />
                </li>
                <li v-if="item.canGrant">
                  <c-permissions-button
                    :title="item.title || item.handle || item.pageID"
                    :target="item.title || item.handle || item.pageID"
                    :resource="`corteza::compose:page-layout/${namespace.namespaceID}/${item.pageID}/*`"
                    :button-label="$t('label.pageLayout')"
                    :show-button-icon="false"
                    all-specific
                    class="dropdown-item"
                  />
                </li>
              </ul>
            </div>

            <template v-if="item.canDeletePage">
              <div
                v-if="hasChildren(item)"
                class="dropdown d-inline-block ms-2"
              >
                <button
                  class="btn btn-outline-danger btn-sm border-0 dropdown-toggle"
                  type="button"
                  data-bs-toggle="dropdown"
                  :disabled="item.processingDelete"
                >
                  <font-awesome-icon
                    v-if="!item.processingDelete"
                    :icon="['far', 'trash-alt']"
                  />
                  <span
                    v-else
                    class="spinner-border spinner-border-sm"
                  />
                </button>
                <ul class="dropdown-menu">
                  <li>
                    <button
                      data-test-id="dropdown-item-delete-update-parent-of-sub-pages"
                      class="dropdown-item"
                      @click="handleDeletePage(item, 'rebase')"
                    >
                      {{ $t('delete.rebase') }}
                    </button>
                  </li>
                  <li>
                    <button
                      data-test-id="dropdown-item-delete-sub-pages"
                      class="dropdown-item"
                      @click="handleDeletePage(item, 'cascade')"
                    >
                      {{ $t('delete.cascade') }}
                    </button>
                  </li>
                </ul>
              </div>

              <c-input-confirm
                v-else
                show-icon
                size="md"
                button-class="px-2"
                :processing="item.processingDelete"
                class="ms-2"
                @confirmed="handleDeletePage(item, 'cascade')"
              />
            </template>
          </div>
        </div>
      </template>
    </sortable-tree>

    <div
      v-else
      class="text-center mt-5 mb-4 pb-1"
    >
      {{ $t('noPages') }}
    </div>
  </div>
</template>

<script setup lang="js">
import { computed } from 'vue'
import { useStore } from '../../../store'
import { useI18n } from 'vue-i18n'
import SortableTree from '../../vendor/SortableTree.vue'
import { compose, NoID } from 'corteza-lib/js/dist'

const { t } = useI18n()

defineOptions({
  i18nOptions: {
    namespaces: 'page',
  },
  name: 'PageTree',
})

const props = defineProps({
  namespace: {
    type: compose.Namespace,
    required: true,
  },
  value: {
    type: Array,
    required: true,
  },
  parentID: {
    type: String,
    default: NoID,
  },
  level: {
    type: Number,
    default: 0,
  },
})

const emit = defineEmits(['input', 'reorder'])

const store = useStore()
const $ComposeAPI = window.__composeAPI

const getModuleByID = computed(() => store.getters['module/getByID'])

const list = computed({
  get () {
    return props.value
  },
  set (pages) {
    emit('input', pages.filter(p => p))
  },
})

function moduleName ({ moduleID }) {
  if (moduleID === NoID) return ''
  return (getModuleByID.value(moduleID) || {}).name
}

function pageViewer ({ pageID = NoID, moduleID = NoID }) {
  const name = moduleID !== NoID ? 'page.record.create' : 'page'
  return { name, params: { pageID } }
}

function handleChangePosition ({ beforeParent, data, afterParent }) {
  const { namespaceID } = props.namespace
  const beforeID = beforeParent.parent ? beforeParent.pageID : NoID
  const afterID = afterParent.parent ? afterParent.pageID : NoID

  const reorder = () => {
    const pageIDs = afterParent.children.map(p => p.pageID)
    if (pageIDs.length) {
      $ComposeAPI.pageReorder({ namespaceID, selfID: afterID, pageIDs })
        .then(() => {
          emit('reorder')
          setTimeout(() => {
            store.dispatch('page/load', { namespaceID, clear: true, force: true })
          }, 1000)
        })
        .catch(toastErrorHandler(t('pageMoveFailed')))
    }
  }

  if (beforeID !== afterID) {
    data.weight = 1
    data.selfID = afterID
    data.namespaceID = namespaceID
    store.dispatch('page/update', data)
      .then(() => { reorder() })
      .catch(toastErrorHandler(t('pageMoveFailed')))
  } else {
    reorder()
  }
}

function handleDeletePage (page, strategy = 'abort') {
  page.processingDelete = true
  store.dispatch('page/delete', { ...page, strategy })
    .then(() => {
      setTimeout(() => {
        emit('reorder')
        toastSuccess(t('notification.page.deleted'))
      }, 300)
    })
    .catch(toastErrorHandler(t('notification.page.deleteFailed')))
    .finally(() => {
      setTimeout(() => {
        page.processingDelete = false
      }, 300)
    })
}

function isValid (page) {
  if (typeof page.validate === 'function') {
    return page.validate().length === 0
  }
  return true
}

function hasChildren (page) {
  return page.children && page.children.length > 0
}

function toastSuccess (...args) {
  console.log('toastSuccess', ...args)
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>

<style lang="scss" scoped>
.grab {
  cursor: grab;
  z-index: 1;
}
</style>

<style lang="scss">
$input-height: 2.4rem;
$content-height: 2.4rem;
$blank-li-height: 1rem;
$padding: 0.5rem;
$border-color: var(--light);
$hover-color: var(--extra-light);
$dropping-color: var(--primary);

.page-name-input {
  height: $input-height;
}

  .list-group {
    margin-right: $padding;

    .droper {
    background-color: $dropping-color !important;
    border-radius: 0.3rem !important;
    margin-left: 0 !important;

    .blank-li.first-child {
      background-color: $dropping-color !important;
    }
  }

  .draging {
    background-color: var(--white) !important;
  }

  ul {
    .content {
      height: 100% !important;
      min-height: $content-height !important;
      line-height: $content-height !important;
      background-color: var(--light);
      border-radius: 0.3rem !important;
      margin-left: 12px;

      .actions {
        display: none;

        .dropdown-menu {
          min-height: initial !important;
          line-height: initial !important;
        }
      }

      &:hover {
        background-color: $hover-color;

        .actions {
          display: block;
        }
      }
    }
  }

  li {
    white-space: nowrap;
    background: var(--white);

    &.blank-li {
      height: $blank-li-height !important;

      .sortable-tree {
        max-height: 100%;
      }

      &:nth-last-of-type(1)::before {
        border-left-color: var(--white) !important;
        height: 0;
      }
    }

    &::before {
      top: calc($content-height / -2) !important;
      border-left-color: var(--white) !important;
    }

    &::after {
      height: $content-height !important;
      top: calc($content-height / 2) !important;
      border-color: var(--white) !important;
    }

    &.parent-li:nth-last-child(2)::before {
      height: $content-height !important;
      top: calc($content-height / -2) !important;
    }
  }

  .parent-li {
    border-top: 1px solid $border-color;
    padding-bottom: $padding !important;
    padding-top: $padding !important;

    .exist-li,
    .blank-li {
      border-top: none;
      padding-top: 0 !important;
      padding-bottom: 0 !important;

      &::after {
        border-top: 2px solid $border-color !important;
        margin-left: 0;
      }

      &::before {
        border-left: 2px solid $border-color !important;
      }
    }

    &.blank-li {
      &::before {
        border-left: 2px solid $border-color !important;
      }
    }

    &.exist-li {
      &::before {
        border-color: var(--white) !important;
      }

      .parent-li {
        &.exist-li {
          &::before {
            border-color: $border-color !important;
          }
        }
      }
    }
  }
}

.pages-list-header {
  min-height: $content-height;
  background-color: var(--gray-200);
  margin-bottom: -1.8rem !important;
  border-bottom: 2px solid var(--light);
  z-index: 1;
}
</style>
