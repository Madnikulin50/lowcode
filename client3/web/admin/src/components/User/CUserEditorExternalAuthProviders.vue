<template>
  <div
    v-if="value && value.length > 0"
    class="card shadow-sm"
    data-test-id="card-external-auth-providers"
  >
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body p-0">
      <table
        class="table table-hover mb-0"
        style="min-height: 200px;"
      >
        <thead class="table-light">
          <tr>
            <th style="width: 350px">{{ $t('label') }}</th>
            <th style="width: 250px">{{ $t('type') }}</th>
            <th class="text-end"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in value" :key="item.credentialsID">
            <td>{{ item.label }}</td>
            <td>{{ item.type }}</td>
            <td class="text-end">
              <c-input-confirm
                data-test-id="button-remove-provider"
                show-icon
                @confirmed="$emit('delete', item.credentialsID)"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps({
  value: { type: Array, required: true },
})

defineEmits(['delete'])
</script>
