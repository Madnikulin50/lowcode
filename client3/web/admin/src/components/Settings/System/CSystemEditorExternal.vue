<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ t('title') }}
      </h4>
    </div>

    <div class="card-body p-0">
      <div class="d-flex align-items-center flex-grow-1 flex-wrap flex-fill-child gap-1 p-3">
        <button
          class="btn btn-primary btn-lg"
          @click="newOIDC()"
        >
          {{ t('oidc.add') }}
        </button>

        <div class="mb-0 ms-auto">
          <label class="form-label text-primary">{{ t('enabled') }}</label>
          <c-input-checkbox
            v-model="external.enabled"
            :value="true"
            :unchecked-value="false"
            :labels="checkboxLabel"
            switch
          />
        </div>
      </div>

      <table
        class="table table-hover mb-0"
      >
        <thead class="table-secondary">
          <tr>
            <th style="width: 50px;">{{ t('table.header.enabled') }}</th>
            <th style="width: 200px;" class="text-capitalize">{{ t('table.header.provider') }}</th>
            <th>{{ t('table.header.info') }}</th>
            <th style="width: 200px;" class="text-end"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in providerItems"
            :key="item.provider + '-' + item.tag"
            :class="item.rowBackground"
          >
            <td>
              <div class="form-check">
                <input
                  class="form-check-input-v3"
                  type="checkbox"
                  :checked="item.enabled"
                  @change="item.enable($event.target.checked)"
                >
              </div>
            </td>
            <td>{{ item.provider || item.tag }}</td>
            <td>{{ item.info }}</td>
            <td class="text-end">
              <c-input-confirm
                v-if="item.delete"
                :icon="item.deleted ? ['fas', 'trash-restore'] : undefined"
                :variant="item.deleted ? 'outline-warning' : 'outline-danger'"
                :variant-ok="item.deleted ? 'warning' : 'danger'"
                @confirmed="item.delete()"
              />

              <button
                class="btn btn-link"
                @click="openEditor(item.editor)"
              >
                <font-awesome-icon
                  :icon="['fas', 'wrench']"
                />
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <Teleport to="body">
        <div
          v-if="modal.open"
          class="modal fade show d-block"
          tabindex="-1"
          style="background: rgba(0,0,0,0.5);"
        >
          <div class="modal-dialog modal-lg modal-dialog-scrollable">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title text-capitalize">{{ modal.title }}</h5>
                <button type="button" class="btn-close" @click="modal.open = false"></button>
              </div>
              <div class="modal-body">
                <component
                  :is="modal.component"
                  v-model="modal.data"
                />
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-outline-secondary" @click="modal.open = false">{{ t('cancel') }}</button>
                <button type="button" class="btn btn-primary" @click="modal.updater(modal.data)">{{ t('save') }}</button>
              </div>
            </div>
          </div>
        </div>
      </Teleport>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        :disabled="!dirty || !canManage"
        :processing="processing"
        :success="success"
        :text="t('admin.general.label.submit')"
        class="ms-auto"
        @submit="emit('submit', changes)"
      />
    </div>
  </div>
</template>

<script setup>
import { useNsI18n } from 'corteza-lib/vue/dist'
defineOptions({ i18nOptions: { namespaces: 'system.settings', keyPrefix: 'editor.external' } })
import _ from 'lodash'
import { ref, reactive, computed, watch } from 'vue'

import OidcExternal from 'corteza-webapp-admin/src/components/Settings/System/Auth/ExternalOIDC'
import StandardExternal from 'corteza-webapp-admin/src/components/Settings/System/Auth/ExternalStd'
import SamlExternal from 'corteza-webapp-admin/src/components/Settings/System/Auth/ExternalSAML'

const t = useNsI18n()

const idpStandard = [
  'google',
  'github',
  'facebook',
  'linkedin',
]

const idpSecurity = {
  permittedRoles: [],
  prohibitedRoles: [],
  forcedRoles: [],
}

function prepareExternal (external) {
  const extractKey = (name, t = 'string') => {
    const v = external.find(v => v.name === `auth.external.${name}`)

    switch (t) {
      case 'string':
        return (v || { value: null }).value || ''
      case 'boolean':
        return !!(v || { value: null }).value
      case 'array':
        return (v || { value: [] }).value || []
      case undefined:
        return v.value
    }
  }

  const extractKeys = (provider, base = {}) => {
    const out = { ...base }

    for (const k in base) {
      out[k] = extractKey(`providers.${provider}.${k}`, Array.isArray(out[k]) ? 'array' : typeof out[k])
    }

    return out
  }

  const extractSec = (prefix) => {
    return { ...idpSecurity, ...(extractKey(`${prefix}.security`, undefined) || {}) }
  }

  const data = {
    enabled: !!(external.find(v => v.name === 'auth.external.enabled') || {}).value,

    oidc: [],
    standard: [],

    saml: {
      enabled: extractKey('saml.enabled'),
      cert: extractKey('saml.cert'),
      name: extractKey('saml.name'),
      key: extractKey('saml.key'),
      'sign-method': extractKey('saml.sign-method'),
      'sign-requests': extractKey('saml.sign-requests', 'boolean'),
      binding: extractKey('saml.binding'),
      idp: {
        url: extractKey('saml.idp.url'),
        'ident-name': extractKey('saml.idp.ident-name'),
        'ident-handle': extractKey('saml.idp.ident-handle'),
        'ident-identifier': extractKey('saml.idp.ident-identifier'),
      },
      security: extractSec('saml'),
    },
  }

  data.standard = idpStandard.map(handle => ({
    handle,
    ...extractKeys(handle, {
      enabled: false,
      secret: '',
      key: '',
      security: {},
      usage: [],
    }),
    security: extractSec(`providers.${handle}`),
  }))

  const prefix = 'auth.external.providers.openid-connect.'

  data.oidc =
    [...new Set(external
      .filter(v => v.name.indexOf(prefix) === 0)
      .map(({ name }) => name.substring(prefix.length).split('.', 2)[0]))]
      .map(handle => ({
        ...extractKeys('openid-connect.' + handle, {
          enabled: false,
          issuer: '',
          key: '',
          secret: '',
          scope: '',
          security: {},
        }),
        handle,
        security: extractSec('providers.openid-connect.' + handle),
        deleted: false,
      }))

  return data
}

const props = defineProps({
  modelValue: {
    type: Array,
    required: true,
  },
  processing: {
    type: Boolean,
    default: false,
  },
  success: {
    type: Boolean,
    default: false,
  },
  canManage: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['submit'])

const modal = reactive({
  open: false,
  editor: null,
  title: null,
  data: null,
})

const external = ref(prepareExternal(props.modelValue))
const checkboxLabel = {
  on: t('label.general.yes'),
  off: t('label.general.no'),
}

const dirty = computed(() => changes.value.length > 0)

const original = computed(() => Object.freeze(prepareExternal(props.modelValue)))

const providerFields = computed(() => [
  { key: 'enabled', label: t('table.header.enabled'), thStyle: { width: '50px' } },
  { key: 'provider', label: t('table.header.provider'), thStyle: { width: '200px' }, tdClass: 'text-capitalize' },
  { key: 'info', label: t('table.header.info') },
  { key: 'editor', label: '', thStyle: { width: '200px' }, tdClass: 'text-end' },
])

const providerItems = computed(() => {
  return [
    {
      rowBackground: _.isEqual(original.value.saml, external.value.saml) ? '' : 'bg-extra-light',
      provider: external.value.saml.name,
      info: external.value.saml.idp.url,
      tag: 'SAML',

      enabled: external.value.saml.enabled,
      enable: (val) => { external.value.saml.enabled = val },

      editor: {
        component: 'saml-external',
        data: external.value.saml,
        title: t('saml.title'),
        updater: (changed) => updater('saml', changed),
      },
    },
    ...external.value.oidc
      .map((p, i) => ({
        rowBackground: (() => {
          if (_.isEqual(original.value.oidc[i], p)) {
            return ''
          }

          if (p.deleted) {
            return 'text-extra-light deleted'
          }

          return 'bg-extra-light'
        })(),
        provider: p.handle,
        tag: 'OIDC',
        info: p.issuer,

        enabled: p.enabled,
        deleted: p.deleted,
        enable: (val) => { external.value.oidc[i].enabled = val },
        delete: () => {
          external.value.oidc[i].deleted = !p.deleted
        },

        editor: {
          component: 'oidc-external',
          data: p,
          title: p.handle,
          updater: (changed) => updater('oidc', changed, i),
        },
      })),
    ...external.value.standard.map((p, i) => ({
      rowBackground: _.isEqual(original.value.standard[i], p) ? '' : 'bg-extra-light',
      provider: p.handle,
      info: p.key,

      enabled: p.enabled,
      enable: (val) => { external.value.standard[i].enabled = val },

      editor: {
        component: 'standard-external',
        data: p,
        title: p.handle,
        updater: (changed) => updater('standard', changed, i),
      },
    })),
  ]
})

const changes = computed(() => {
  let name, value
  const c = []

  const prefix = 'auth.external.providers'
  const o = original.value
  const e = external.value

  if (!_.isEqual(o.enabled, e.enabled)) {
    c.push({ name: 'auth.external.enabled', value: e.enabled })
  }

  const mapKeys = (prefix, wc, org, keys) => {
    for (const k of keys) {
      if (wc[k] === undefined) {
        console.error(`potential issue - unknown key "${prefix}.${k}" used`, wc)
      }

      if (_.isEqual(wc[k], org[k])) {
        continue
      }

      name = `${prefix}.${k}`
      value = wc[k]
      c.push({ name, value })
    }
  }

  e.standard.forEach((p, i) => {
    mapKeys(
      `${prefix}.${p.handle}`,
      p,
      o.standard[i],
      ['key', 'secret', 'enabled', 'security', 'usage'],
    )
  })

  const oidcKeys = ['key', 'secret', 'enabled', 'issuer', 'scope', 'security']
  e.oidc.forEach((p, i) => {
    if (p.deleted) {
      [...oidcKeys, 'weight', 'redirect', 'label']
        .forEach(name => c.push({ name: `${prefix}.openid-connect.${p.handle}.${name}`, value: null }))
    } else {
      mapKeys(
        `${prefix}.openid-connect.${p.handle}`,
        p,
        o.oidc[i] || {},
        oidcKeys,
      )
    }
  })

  mapKeys(
    'auth.external.saml',
    e.saml,
    o.saml,
    ['enabled', 'name', 'key', 'cert', 'sign-method', 'sign-requests', 'binding', 'security'],
  )

  mapKeys(
    'auth.external.saml.idp',
    e.saml.idp,
    o.saml.idp,
    ['url', 'ident-name', 'ident-handle', 'ident-identifier'],
  )

  return c
})

watch(() => props.modelValue, () => {
  external.value = prepareExternal(props.modelValue)
}, { immediate: true })

function openEditor ({ component, title, data, updater }) {
  modal.open = true
  modal.component = component
  modal.title = title
  modal.updater = updater
  modal.data = JSON.parse(JSON.stringify(data))
}

function newOIDC () {
  openEditor({
    component: 'oidc-external',
    title: t('oidc.add'),
    data: {
      handle: '',
      enabled: true,
      issuer: '',
      key: '',
      secret: '',
      scope: '',
      fresh: true,
      security: { ...idpSecurity },
    },
    updater: (changed) => {
      updater('oidc', changed, -1)
    },
  })
}

function updater (key, val, i = undefined) {
  if (i === undefined) {
    external.value[key] = val
  } else if (i < 0) {
    external.value[key].push(val)
  } else {
    external.value[key][i] = val
  }
}
</script>

<style lang="scss">
.deleted {
  text-decoration: line-through;
}
</style>
