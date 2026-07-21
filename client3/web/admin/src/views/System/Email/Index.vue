<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="$t('editor.title')" />
    <form @submit.prevent="update">
      <c-email-settings v-model="form" :settings="settings" class="mb-3" />
      <c-email-test v-if="areTestSettingsSet" v-model="test" :settings="settings" :processing="testing" :errors="testResultErrors" class="mb-3" @test="testSettings" />
      <button type="submit" class="btn btn-primary mt-2" :disabled="!changed">{{ $t('label.saveAndClose') }}</button>
    </form>
  </div>
</template>
<script setup>
import { ref, reactive, computed, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import CEmailSettings from '../../../components/Settings/Mail/CEmailSettings.vue'
import CEmailTest from '../../../components/Settings/Mail/CEmailTest.vue'
const { t } = useI18n()
const $Settings = inject('$Settings', {})
const settings = ref([])
const form = reactive({})
const test = reactive({})
const testing = ref(false)
const testResultErrors = ref([])
const areTestSettingsSet = computed(() => settings.value.length > 0)
const changed = computed(() => settings.value.some(s => form[s.name] !== undefined && form[s.name] !== s.value))
onMounted(() => { window.__systemAPI.settingsList({ prefix: 'mail' }).then(s => { settings.value = s; s.forEach(s => { form[s.name] = s.value }) }) })
function update() { const values = Object.entries(form).filter(([k, v]) => settings.value.some(s => s.name === k && s.value !== v)).map(([name, value]) => ({ name, value })); incLoader(); window.__systemAPI.settingsUpdate({ values }).then(() => { $Settings.fetch() }).finally(() => decLoader()) }
function incLoader() {} function decLoader() {}
function testSettings() { testing.value = true; testResultErrors.value = []; const values = Object.entries(test).map(([name, value]) => ({ name, value })); window.__systemAPI.settingsTest({ values }).then(p => { testResultErrors.value = []; if (p.errors) testResultErrors.value = Object.entries(p.errors).map(([key, value]) => `${key}: ${value}`) }).catch(e => { testResultErrors.value = [e.message || e] }).finally(() => { testing.value = false }) }
</script>
