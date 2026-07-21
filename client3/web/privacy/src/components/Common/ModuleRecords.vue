<template>
  <div class="overflow-auto">
    <div
      v-for="m in modules"
      :key="m.moduleID"
      class="card border shadow-sm mb-3"
    >
      <div class="card-header d-flex justify-content-between border-bottom">
        <div class="mb-0">
          <label class="text-primary form-label">{{ t('module') }}</label>
          <div>{{ m.module }}</div>
        </div>
        <div class="mb-0">
          <label class="text-primary form-label">{{ t('namespace') }}</label>
          <div>{{ m.namespace }}</div>
        </div>
      </div>

      <div class="card-body">
        <h6 v-if="!m.records.length" class="text-center">{{ t('no-records') }}</h6>

        <div
          v-for="(r, ri) in m.records"
          :key="r.recordID"
          class="mb-0"
        >
          <div class="mb-2">
            <label class="text-primary form-label">RecordID</label>
            <div>{{ r.recordID }}</div>
          </div>

          <div class="row">
            <div
              v-for="value in r.values"
              :key="value.name"
              class="col-12 col-lg-6"
            >
              <div class="mb-2">
                <label class="text-primary form-label">{{ value.name }}</label>
                <slot
                  :namespace="{ namespaceID: m.namespaceID, name: m.namespace }"
                  :module="{ moduleID: m.moduleID, name: m.module }"
                  :record="r"
                  :value="value"
                />
              </div>
            </div>
          </div>

          <hr v-if="ri < m.records.length - 1">
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

defineProps({
  modules: { type: Array, required: true },
})

const { t } = useI18n()
</script>