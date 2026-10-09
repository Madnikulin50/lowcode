<template>
  <tr>
    <td />
    <td colspan="5" class="p-0">
      <div class="d-flex">
        <div style="min-width: 200px;">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('navigation.fieldLabel') }}</label>
            <input v-model="options.item.dropdown.label" class="form-control" type="text" />
          </div>
        </div>
        <div style="min-width: 200px;">
          <div class="mb-3">
            <label class="form-label text-primary pr-2">{{ $t('navigation.drop') }}</label>
            <div class="btn-group" role="group" data-bs-toggle="buttons">
              <label
                v-for="al in aligns"
                :key="al.value"
                class="btn btn-outline-primary btn-sm"
                :class="{ active: options.item.align === al.value }"
              >
                <input
                  type="radio"
                  :value="al.value"
                  v-model="options.item.align"
                />
                {{ al.text }}
              </label>
            </div>
          </div>
        </div>
      </div>

      <table v-if="options.item.dropdown.items.length > 0" class="table borderless responsive border-top pt-2">
        <thead>
          <tr class="text-primary">
            <th scope="col" style="min-width: 200px;">{{ $t('navigation.text') }}</th>
            <th scope="col" style="min-width: 200px;">{{ $t('navigation.url') }}</th>
            <th scope="col" style="min-width: 200px;">{{ $t('navigation.openIn') }}</th>
            <th class="text-center" scope="col" style="width: 50px; min-width: 50px;">{{ $t('navigation.delimiter') }}</th>
          </tr>
        </thead>
        <tr v-for="(item, dropIndex) in options.item.dropdown.items" :key="`drop-${dropIndex}`">
          <td>
            <input v-model="item.label" class="form-control form-control-sm" type="text" />
          </td>
          <td>
            <input v-model="item.url" class="form-control form-control-sm" type="text" />
          </td>
          <td class="align-middle text-center">
            <select v-model="item.target" class="form-select form-control form-select-sm">
              <option value="sameTab">{{ $t('navigation.sameTab') }}</option>
              <option value="newTab">{{ $t('navigation.newTab') }}</option>
            </select>
          </td>
          <td class="align-middle text-center">
            <div class="form-check form-switch d-flex align-items-center justify-content-center">
              <input v-model="item.delimiter" class="form-check-input" type="checkbox" role="switch" />
            </div>
          </td>
          <td class="align-middle text-center">
            <c-input-confirm show-icon @confirmed="options.item.dropdown.items.splice(dropIndex, 1)" />
          </td>
        </tr>
      </table>

      <div class="mb-4 mb-3 px-3">
        <button
          class="btn btn-primary btn-sm text-decoration-none"
          @click="options.item.dropdown.items.push({ text: '', url: '', target: 'sameTab', delimiter: false })"
        >
          <font-awesome-icon :icon="['fas', 'plus']" size="sm" class="me-1" />
          {{ $t('navigation.addDropdown') }}
        </button>
      </div>
    </td>
  </tr>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  item: { type: Object, required: true },
})

const options = computed({
  get: () => props.item.options,
  set: (opts) => { props.item.options = opts },
})

const aligns = [
  { value: 'right', text: $t('navigation.right') },
  { value: 'left', text: $t('navigation.left') },
  { value: 'bottom', text: $t('navigation.bottom') },
  { value: 'top', text: $t('navigation.top') },
]
</script>

<style lang="scss" scoped>
th, td {
  padding-left: 15px;
  padding-right: 15px;
}
</style>
