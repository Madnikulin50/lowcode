<template>
  <div>
    <div>
      <div class="mb-3">
        <h5>{{ $t('navigation.displayOptions') }}</h5>
        <div class="row justify-content-between text-primary">
          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary pr-2">{{ $t('navigation.appearance') }}</label>
              <div class="btn-group" role="group" data-bs-toggle="buttons">
                <label
                  v-for="opt in appearanceOptions"
                  :key="opt.value"
                  class="btn btn-outline-secondary btn-sm"
                  :class="{ active: options.display.appearance === opt.value }"
                >
                  <input type="radio" :value="opt.value" v-model="options.display.appearance" />
                  {{ opt.text }}
                </label>
              </div>
            </div>
          </div>
          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary pr-2">{{ $t('navigation.justify') }}</label>
              <div class="btn-group" role="group" data-bs-toggle="buttons">
                <label
                  v-for="opt in justifyOptions"
                  :key="opt.value"
                  class="btn btn-outline-secondary btn-sm"
                  :class="{ active: options.display.justify === opt.value }"
                >
                  <input type="radio" :value="opt.value" v-model="options.display.justify" />
                  {{ opt.text }}
                </label>
              </div>
            </div>
          </div>
          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary pr-2">{{ $t('navigation.alignment') }}</label>
              <div class="btn-group" role="group" data-bs-toggle="buttons">
                <label
                  v-for="opt in alignmentOptions"
                  :key="opt.value"
                  class="btn btn-outline-secondary btn-sm"
                  :class="{ active: options.display.alignment === opt.value }"
                >
                  <input type="radio" :value="opt.value" v-model="options.display.alignment" />
                  {{ opt.text }}
                </label>
              </div>
            </div>
          </div>
        </div>
      </div>

      <hr class="my-2">

      <div class="mb-3 mt-2">
        <div class="d-flex align-items-center mb-4">
          <h5 class="mb-0">{{ $t('navigation.navigationItems') }}</h5>
        </div>
        <div class="mt-3">
          <c-form-table-wrapper
            :labels="{ addButton: $t('label.add') }"
            @add-item="addNavigationItem"
          >
            <draggable
            item-key="id"
              v-model="block.options.navigationItems"
              group="sort"
              handle=".grab"
            >
              <template #item="{ element, index }">
                <div
                  :key="index"
                >
                  <hr v-if="index">
                  <table class="table borderless responsive small">
                    <thead class="text-primary">
                      <tr>
                        <th scope="col" class="grab" style="vertical-align: middle; width: 1rem;">
                          <font-awesome-icon :icon="['fas', 'bars']" class="text-secondary" />
                        </th>
                        <th scope="col" style="vertical-align: middle; min-width: 200px;">{{ $t('navigation.type') }}</th>
                        <th scope="col" style="vertical-align: middle; min-width: 200px;">{{ $t('navigation.color') }}</th>
                        <th scope="col" style="vertical-align: middle; min-width: 200px;">{{ $t('navigation.background') }}</th>
                        <th scope="col" style="vertical-align: middle; width: 50px; min-width: 50px;">{{ $t('navigation.enabled') }}</th>
                        <th scope="col" style="vertical-align: middle; width: auto; min-width: 100px;" />
                        <th scope="col" class="text-end">
                          <c-input-confirm
                            show-icon
                            button-class="px-2"
                            size="md"
                            @confirmed="options.navigationItems.splice(index, 1)"
                          />
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr>
                        <td />
                        <td class="align-middle">
                          <select v-model="element.type" class="form-select form-control form-select-sm">
                            <option v-for="nt in navigationItemTypes" :key="nt.value" :value="nt.value">{{ nt.text }}</option>
                          </select>
                        </td>
                        <td class="align-middle">
                          <c-input-color-picker
                            v-model="element.options.textColor"
                            :translations="colorPickerLabels"
                            :theme-settings="themeSettings"
                            class="w-100"
                          />
                        </td>
                        <td class="align-middle">
                          <c-input-color-picker
                            v-model="element.options.backgroundColor"
                            :translations="colorPickerLabels"
                            :theme-settings="themeSettings"
                            class="w-100"
                          />
                        </td>
                        <td class="d-flex align-items-center">
                          <div class="form-check form-switch">
                            <input v-model="element.options.enabled" class="form-check-input" type="checkbox" role="switch" />
                          </div>
                        </td>
                      </tr>
                      <component
                        :is="element.type"
                        :item="element"
                        :namespace="namespace"
                      />
                    </tbody>
                  </table>
                </div>
              </template>
            </draggable>
            <div v-if="!block.options.navigationItems.length" class="text-center my-4">
              <p>{{ $t('navigation.noNavigationItems') }}</p>
            </div>
          </c-form-table-wrapper>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Draggable from 'vuedraggable'
import { compose } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import TextSection from './NavTypes/Text.vue'
import Url from './NavTypes/Url.vue'
import Compose from './NavTypes/ComposePage.vue'
import Dropdown from './NavTypes/Dropdown.vue'

const { CInputColorPicker } = components
const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  namespace: { type: Object, required: true },
})

const options = computed(() => props.block.options)

const colorPickerLabels = computed(() => ({
  modalTitle: $t('navigation.colorPicker'),
  light: $t('themes.labels.light'),
  dark: $t('themes.labels.dark'),
  cancelBtnLabel: $t('label.cancel'),
  saveBtnLabel: $t('label.saveAndClose'),
}))

const themeSettings = computed(() => window.__settings?.get?.('ui.studio.themes', []) || [])

const appearanceOptions = [
  { value: 'tabs', text: $t('navigation.tabs') },
  { value: 'pills', text: $t('navigation.pills') },
  { value: 'small', text: $t('navigation.small') },
]

const alignmentOptions = [
  { value: 'left', text: $t('navigation.left') },
  { value: 'center', text: $t('navigation.center') },
  { value: 'right', text: $t('navigation.right') },
]

const justifyOptions = [
  { value: 'justify', text: $t('navigation.justify') },
  { value: 'none', text: $t('navigation.none') },
]

const navigationItemTypes = [
  { value: 'url', text: $t('navigation.url') },
  { value: 'compose', text: $t('navigation.composePage') },
  { value: 'dropdown', text: $t('navigation.dropdown') },
  { value: 'text-section', text: $t('navigation.text') },
]

function addNavigationItem() {
  props.block.options.navigationItems.push(
    compose.PageBlockNavigation.makeNavigationItem({
      type: 'compose',
      options: {
        backgroundColor: '#FFFFFF00',
        item: {
          label: '',
          url: '',
          align: 'bottom',
          target: 'sameTab',
          displaySubPages: false,
          dropdown: {
            label: '',
            items: [],
          },
        },
      },
    }),
  )
}
</script>

<style lang="scss" scoped>
th {
  width: 25%;
}
th, td {
  padding-left: 15px;
  padding-right: 15px;
}
</style>
