<template>
  <tr>
    <td />
    <td>
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('navigation.fieldLabel') }}</label>
        <input
          v-model="options.item.label"
          class="form-control"
          type="text"
        />
      </div>
    </td>
    <td>
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('navigation.url') }}</label>
        <input
          v-model="options.item.url"
          class="form-control"
          type="text"
        />
      </div>
    </td>
    <td>
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('navigation.openIn') }}</label>
        <select v-model="options.item.target" class="form-select form-control">
          <option value="sameTab">{{ $t('navigation.sameTab') }}</option>
          <option value="newTab">{{ $t('navigation.newTab') }}</option>
        </select>
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
</script>

<style lang="scss" scoped>
th, td {
  padding-left: 15px;
  padding-right: 15px;
}
</style>
