<template>
  <form @submit.prevent="onSubmit">
    <table class="table table-outline-secondary">
      <thead class="bg-light">
        <tr>
          <th class="key py-2 px-1">
            <div class="dropdown">
              <button
                class="btn btn-outline-secondary btn-sm dropdown-toggle"
                data-bs-toggle="dropdown"
              >
                {{ $t('add-language') }}
              </button>
              <ul class="dropdown-menu">
                <li v-for="lang in intLanguages" :key="lang.tag">
                  <button
                    class="dropdown-item"
                    :disabled="lang.default || lang.visible"
                    @click="lang.visible = true"
                  >
                    {{ lang.localizedName }}
                  </button>
                </li>
              </ul>
            </div>
          </th>
          <th
            v-for="lang in visibleLanguages"
            :key="lang.tag"
            class="text-truncate position-relative"
            :style="{ width: `${100 / visibleLanguages.length}%` }"
          >
            {{ lang.localizedName }}
            <button
              v-if="!lang.default"
              class="btn btn-link float-end p-0 m-0"
              @click="lang.visible=false"
            >
              <font-awesome-icon :icon="['fas', 'times']" />
            </button>
          </th>
        </tr>
      </thead>
      <template v-for="(r, i) in resources()" :key="i">
        <tbody v-if="!r.isPrimary" class="border-top">
          <tr class="bg-light">
            <th :colspan="visibleLanguages.length + 1">
              {{ r.title }}
            </th>
          </tr>
        </tbody>
        <tbody v-for="(key, k) in keys(r.resource)" :key="k" class="border-top">
          <tr :class="{ 'bg-light': key === highlightKey }">
            <td colspan="2" class="text-break small">
              <samp>{{ keyPrettifier(key) }}</samp>
            </td>
            <td
              v-for="(lang, langIndex) in visibleLanguages"
              :key="lang.tag"
              :class="{ 'm-0 p-0': true, 'bg-warning': isDirty(r.resource, key, lang.tag) }"
            >
              <button
                v-if="isDirty(r.resource, key, lang.tag)"
                class="btn btn-link float-end p-1 mt-2 me-2"
                @click="reset(r.resource, key, lang.tag)"
              >
                <font-awesome-icon :icon="['fas', 'times']" />
              </button>
              <Editable
                :value="msg(r.resource, key, lang.tag)"
                :placeholder="$t('missing-translation')"
                :tabindex="langIndex + 1"
                @input="onUpdate(r.resource, key, lang.tag, $event)"
              />
            </td>
          </tr>
        </tbody>
      </template>
    </table>
  </form>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Editable from './Editable.vue'

const { t: $t } = useI18n({ useScope: 'global' })
const emit = defineEmits(['change'])

const lsKey = 'resource-translator.languages'

const props = defineProps({
  languages: { type: Array, required: true },
  primaryResource: { type: String, required: true },
  translations: { type: Array, required: true },
  titles: { type: Object, default: () => ({}) },
  highlightKey: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  keyPrettifier: {
    type: Function,
    default: (key) => {
      return key
        .replace(/([A-Z])/, ' $1')
        .toLowerCase()
        .replace(/(\d+)/, '#$1')
        .split('.').map(s => s.substring(0, 1).toUpperCase() + s.substring(1))
        .join(' ')
    },
  },
})

const preselected = (window.localStorage.getItem(lsKey) || '').split(',')

const intLanguages = ref(props.languages.map((lang, i) => ({
  ...lang,
  default: i === 0,
  visible: i === 0 || preselected.includes(lang.tag),
})))

const intTranslations = ref(props.translations.map(t => ({ ...t, org: t.message, dirty: false })))

const visibleLanguages = computed(() => intLanguages.value.filter(({ visible }) => visible))

watch(intLanguages, () => {
  const selected = visibleLanguages.value.map(({ tag }) => tag).join(',')
  window.localStorage.setItem(lsKey, selected)
}, { deep: true })

function resources() {
  return intTranslations.value
    .map(r => r.resource)
    .filter((r, i, rr) => rr.indexOf(r) === i)
    .map(resource => ({
      resource,
      title: props.titles[resource],
      isPrimary: resource === props.primaryResource,
    }))
}

function keys(resource) {
  return intTranslations.value
    .filter(r => r.resource === resource)
    .map(r => r.key)
    .filter((r, i, rr) => rr.indexOf(r) === i)
}

function find(resource, key, lang) {
  return intTranslations.value.find(r => r.resource === resource && r.key === key && r.lang === lang) ||
    { dirty: false, message: '' }
}

function isDirty(resource, key, lang) {
  return find(resource, key, lang).dirty
}

function msg(resource, key, lang) {
  return find(resource, key, lang).message
}

function reset(resource, key, lang) {
  const t = find(resource, key, lang)
  if (t.dirty) {
    t.message = t.org
    t.dirty = false
  }
}

function stripHtml(v) {
  const el = document.createElement('div')
  el.innerHTML = v
  return el.textContent || el.innerText || ''
}

function onUpdate(resource, key, lang, message) {
  message = stripHtml(message)
  const v = intTranslations.value.find(r => r.resource === resource && r.key === key && r.lang === lang)
  if (v === undefined) {
    const fresh = { resource, key, lang, message, org: message, dirty: true }
    intTranslations.value.push(fresh)
  } else {
    v.dirty = v.org !== message
    v.message = message
  }
  const dirty = intTranslations.value
    .filter(({ dirty }) => dirty)
    .map(({ resource, key, lang, message }) => ({ resource, key, lang, message }))
  emit('change', dirty)
}
</script>

<style lang="scss" scoped>
.key {
  min-width: 200px;
}
</style>
