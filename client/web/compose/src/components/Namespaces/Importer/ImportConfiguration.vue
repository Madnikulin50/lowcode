<template>
  <b-card footer-class="border-top d-flex justify-content-between align-items-center">
    <b-form-group
      :label="$t('name.label')"
      label-class="text-primary"
    >
      <b-form-input
        v-model="name"
        data-test-id="input-name"
        :placeholder="$t('name.placeholder')"
        class="mt-1"
      />
    </b-form-group>

    <b-form-group
      :label="$t('import.slug.label')"
      label-class="text-primary"
    >
      <b-form-input
        v-model="slug"
        data-test-id="input-handle"
        class="mt-1"
        :state="slugState"
        :placeholder="$t('slug.placeholder')"
      />
      <b-form-invalid-feedback :state="slugState">
        {{ $t('slug.invalid-handle-characters') }}
      </b-form-invalid-feedback>
    </b-form-group>

    <b-form-group
      :label="$t('import.connection.label')"
      label-class="text-primary"
      class="mb-0"
    >
      <c-input-select
        v-model="connectionID"
        :options="connections"
        :clearable="false"
        :reduce="o => o.connectionID"
        :placeholder="$t('import.connection.placeholder')"
        :get-option-label="({ handle, meta }) => meta.name || handle"
        :get-option-key="getOptionKey"
      />
    </b-form-group>

    <b-form-group
      :label="$t('import.importData.label')"
      label-class="text-primary"
      class="mb-0"
    >
      <c-input-checkbox
        v-model="importData"
        switch
      />
    </b-form-group>

    <template #footer>
      <b-button
        data-test-id="button-back"
        variant="link"
        class="d-flex align-items-center text-dark back gap-1 text-decoration-none"
        @click="$emit('back')"
      >
        <font-awesome-icon
          :icon="['fas', 'chevron-left']"
          class="back-icon"
        />
        {{ $t('import.back') }}
      </b-button>

      <b-button
        data-test-id="button-import"
        variant="primary"
        :disabled="submitDisabled"
        @click="nextStep"
      >
        {{ $t('import.import') }}
      </b-button>
    </template>
  </b-card>
</template>

<script>
import { handle } from 'corteza-lib/vue/dist'

export default {
  i18nOptions: {
    namespaces: 'namespace',
  },

  props: {
    session: {
      type: Object,
      required: true,
      default: () => ({}),
    },
  },

  data () {
    return {
      processing: {
        connections: true,
        sensitiveData: true,
      },
      name: '',
      slug: '',
      connectionID: null,
      importData: true,
      connections: [],
    }
  },

  computed: {
    submitDisabled () {
      return [this.nameState, this.slugState, this.slug].includes(false)
    },

    nameState () {
      return this.name.length > 0 ? null : false
    },

    slugState () {
      return handle.handleState(this.slug)
    },
  },

  created () {
    this.fetchConnections()
  },

  methods: {
    fetchConnections () {
      this.processing.connections = true

      this.$SystemAPI.dataPrivacyConnectionList()
        .then(({ set = [] }) => {
          this.connections = set
          const { connectionID } = set[0] || {}
          this.connectionID = connectionID
        })
        .catch(this.toastErrorHandler(this.$t('notification:connection-load-failed')))
        .finally(() => {
          this.processing.connections = false
        })
    },


    getOptionKey ({ connectionID }) {
      return connectionID
    },
    nextStep () {
      // convert to api's structure
      const rtr = {
        name: this.name,
        slug: this.slug,
        connectionID: this.connectionID,
        importData: this.importData,
      }

      this.$emit('configured', rtr)
    },
  },
}
</script>
