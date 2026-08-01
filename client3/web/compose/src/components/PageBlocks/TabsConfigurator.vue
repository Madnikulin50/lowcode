<template>
  <div>
    <h5>{{ $t('tabs.displayTitle') }}</h5>
    <div class="row text-primary g-0">
      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('tabs.style.appearance.label') }}</label>
          <div class="btn-group btn-group-sm w-100" role="group">
            <button v-for="opt in style.appearance" :key="opt.value" :class="['btn', block.options.style.appearance === opt.value ? 'btn-secondary' : 'btn-outline-secondary']" @click="block.options.style.appearance = opt.value">{{ opt.text }}</button>
          </div>
        </div>
      </div>
      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('tabs.style.orientation.label') }}</label>
          <div class="btn-group btn-group-sm w-100" role="group">
            <button v-for="opt in style.orientation" :key="opt.value" :class="['btn', block.options.style.orientation === opt.value ? 'btn-secondary' : 'btn-outline-secondary']" @click="block.options.style.orientation = opt.value">{{ opt.text }}</button>
          </div>
        </div>
      </div>
      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('tabs.style.position.label') }}</label>
          <div class="btn-group btn-group-sm w-100" role="group">
            <button v-for="opt in style.position" :key="opt.value" :class="['btn', block.options.style.position === opt.value ? 'btn-secondary' : 'btn-outline-secondary']" @click="block.options.style.position = opt.value">{{ opt.text }}</button>
          </div>
        </div>
      </div>
      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('tabs.style.alignment.label') }}</label>
          <div class="btn-group btn-group-sm w-100" role="group">
            <button v-for="opt in style.alignment" :key="opt.value" :class="['btn', block.options.style.alignment === opt.value ? 'btn-secondary' : 'btn-outline-secondary']" @click="block.options.style.alignment = opt.value">{{ opt.text }}</button>
          </div>
        </div>
      </div>
      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('tabs.style.justify.label') }}</label>
          <div class="btn-group btn-group-sm w-100" role="group">
            <button v-for="opt in style.justifyOptions" :key="opt.value" :class="['btn', block.options.style.justify === opt.value ? 'btn-secondary' : 'btn-outline-secondary']" @click="block.options.style.justify = opt.value">{{ opt.text }}</button>
          </div>
        </div>
      </div>
    </div>

    <hr class="my-2" />

    <div class="d-flex align-items-center mb-2">
      <h5 class="m-0">{{ $t('tabs.title') }}</h5>
    </div>

    <c-form-table-wrapper :labels="{ addButton: $t('label.add') }" @add-item="addTab">
      <table class="table table-sm table-borderless" responsive>
        <thead>
          <tr>
            <th scope="col"></th>
            <th class="d-flex align-items-center text-primary" scope="col">
              {{ $t('tabs.table.columns.title.label') }}
              <c-hint :tooltip="$t('interpolationFootnote', ['${record.values.fieldName}', '${recordID}', '${ownerID}', '${userID}'])" class="d-block" />
            </th>
            <th class="text-primary" scope="col">{{ $t('tabs.table.columns.block.label') }}</th>
            <th class="d-flex align-items-center text-primary" scope="col">
              {{ $t('tabs.table.columns.lazy.label') }}
              <c-hint :tooltip="$t('tabs.table.columns.lazy.tooltip')" class="d-block" />
            </th>
            <th scope="col"></th>
          </tr>
        </thead>
        <draggable item-key="id" v-model="block.options.tabs" handle=".handle" tag="tbody">
          <template #item="{ element, index }">
            <tr :key="index">
              <td class="handle align-middle pe-2" style="width: 30px;">
                <font-awesome-icon :icon="['fas', 'bars']" class="grab m-0 text-secondary p-0" />
              </td>
              <td class="align-middle" style="width: 50%; min-width: 200px;">
                <input v-model="element.title" class="form-control" />
              </td>
              <td class="align-middle" style="width: 50%; min-width: 200px;">
                <div class="input-group d-flex flex-nowrap w-100">
                  <c-input-select v-model="element.blockID" :options="blockOptions" :placeholder="$t('tabs.placeholder.block')" :get-option-label="b => b.title || b.kind" :get-option-key="b => fetchID(b)" :selectable="option => isSelectable(option)" :reduce="option => option.value" class="flex-grow-1" />
                  <button v-if="element.blockID" class="btn btn-outline-extra-light d-flex align-items-center justify-content-center" style="width: 40px;" data-bs-toggle="tooltip" :title="$t('tabs.tooltip.edit')" @click="editBlock(element.blockID)">
                    <font-awesome-icon :icon="['far', 'edit']" />
                  </button>
                  <button v-else class="btn btn-outline-extra-light d-flex align-items-center justify-content-center" style="width: 40px;" data-bs-toggle="tooltip" :title="$t('tabs.tooltip.addBlock')" @click="showBlockSelector(index)">
                    <font-awesome-icon :icon="['fas', 'plus']" />
                  </button>
                </div>
              </td>
              <td class="text-center align-middle">
                <div class="form-check form-switch justify-content-center d-flex">
                  <input v-model="element.lazy" class="form-check-input" type="checkbox" role="switch" data-bs-toggle="tooltip" :title="$t('tabs.table.columns.lazy.tooltip')" />
                </div>
              </td>
              <td class="text-center align-middle" style="min-width: 80px;">
                <c-input-confirm :tooltip="$t('tabs.tooltip.delete')" show-icon @confirmed="deleteTab(index)" />
              </td>
            </tr>
          </template>
        </draggable>
      </table>
    </c-form-table-wrapper>

    <div v-if="showBlockSelectorModal" class="modal fade show d-block" tabindex="-1">
      <div class="modal-dialog modal-dialog-centered modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('tabs.newBlockModal') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" @click="showBlockSelectorModal = false"></button>
          </div>
          <div class="modal-body p-2">
            <NewBlockSelector :record-page="!!module" :disabled-kinds="['Tabs']" @select="addBlock" />
          </div>
        </div>
      </div>
    </div>
    <div v-if="showBlockSelectorModal" class="modal-backdrop fade show" @click="showBlockSelectorModal = false" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import draggable from 'vuedraggable'
import { fetchID } from 'corteza-webapp-compose/src/lib/block'
const NewBlockSelector = () => import('corteza-webapp-compose/src/components/Admin/Page/Builder/Selector')

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  module: { type: Object, required: false },
})

const options = computed(() => props.block.options)
const activeIndex = ref(null)
const showBlockSelectorModal = ref(false)

const style = ref({
  appearance: [
    { text: $t('tabs.style.appearance.tabs'), value: 'tabs' },
    { text: $t('tabs.style.appearance.pills'), value: 'pills' },
    { text: $t('tabs.style.appearance.small'), value: 'small' },
  ],
  alignment: [
    { text: $t('tabs.style.alignment.left'), value: 'left' },
    { text: $t('tabs.style.alignment.center'), value: 'center' },
    { text: $t('tabs.style.alignment.right'), value: 'right' },
  ],
  justifyOptions: [
    { text: $t('tabs.style.justify.justify'), value: 'justify' },
    { text: $t('tabs.style.justify.none'), value: 'none' },
  ],
  orientation: [
    { text: $t('tabs.style.orientation.horizontal'), value: 'horizontal' },
    { text: $t('tabs.style.orientation.vertical'), value: 'vertical' },
  ],
  position: [
    { text: $t('tabs.style.position.start'), value: 'start' },
    { text: $t('tabs.style.position.end'), value: 'end' },
  ],
})

const blockOptions = computed(() => {
  return [
    ...props.page.blocks.filter(({ blockID, kind }) => kind !== 'Tabs' && !props.blocks.some(b => b.blockID === blockID) && options.value.tabs.some(b => b.blockID === blockID)),
    ...props.blocks.filter(b => b.kind !== 'Tabs'),
  ].map(b => ({ ...b, value: fetchID(b) }))
})

onMounted(() => {
  window.addEventListener('builder-createRequestFulfilled', createRequestFulfilled)
})

onBeforeUnmount(() => {
  window.removeEventListener('builder-createRequestFulfilled', createRequestFulfilled)
  setDefaultValues()
})

function createRequestFulfilled({ detail: tab } = {}) {
  if (!tab) return
  const { title = '' } = options.value.tabs[activeIndex.value] || {}
  tab.title = title
  updateTab(tab, activeIndex.value)
}

function addTab() {
  options.value.tabs.push({ blockID: null, title: undefined, lazy: true })
}

function isSelectable(option) {
  return !options.value.tabs.some(t => t.blockID === option.value)
}

function showBlockSelector(index) {
  showBlockSelectorModal.value = true
  activeIndex.value = index
}

function editBlock(blockID) {
  window.dispatchEvent(new CustomEvent('tab-editRequest', { detail: blockID }))
}

function addBlock(block) {
  showBlockSelectorModal.value = false
  block.meta.hidden = true
  window.dispatchEvent(new CustomEvent('tab-createRequest', { detail: block }))
}

function updateTab(tab, index) {
  options.value.tabs.splice(index, 1, tab)
}

function deleteTab(tabIndex) {
  options.value.tabs.splice(tabIndex, 1)
}

function setDefaultValues() {
  activeIndex.value = null
  style.value = {}
  showBlockSelectorModal.value = false
}
</script>
