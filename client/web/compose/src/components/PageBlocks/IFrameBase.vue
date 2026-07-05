<template>
  <wrap
    v-bind="$props"
    v-on="$listeners"
    @refreshBlock="refresh"
  >
    <img
      v-if="src && true"
      ref="iframe"
      class="h-100 w-100 border-0"
      :src="src | checkValidURL"
    >
    <iframe
      v-if="src&&false"
      ref="iframe"
      class="h-100 w-100 border-0"
      :src="src | checkValidURL"
    />
  </wrap>
</template>
<script>
import base from './base'
import { NoID } from 'corteza-lib/js/dist'
import { evaluatePrefilter, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'

export default {
  extends: base,

  computed: {
    src () {
      const { srcField, src } = this.options
      const blank = 'about:blank'
      let url = src

      if (this.options.srcField) {
        if (this.record) {
          url = this.record.values[srcField]
        }
      }

      let interpolatedURL = evaluatePrefilter(url, {
        record: this.record,
        user: this.$auth.user || {},
        recordID: (this.record || {}).recordID || NoID,
        ownerID: (this.record || {}).ownedBy || NoID,
        userID: (this.$auth.user || {}).userID || NoID,
      })
      if (interpolatedURL[0] !== 'h') {
        interpolatedURL = window.CortezaAPI + interpolatedURL
      }

      return interpolatedURL || blank
    },
  },

  mounted () {
    this.refreshBlock(this.refresh)
    this.createEvents()
  },

  beforeDestroy () {
    this.destroyEvents()
  },

  methods: {
    refresh () {
      this.$refs.iframe.src = this.src
    },

    createEvents () {
      this.$root.$on('record-field-change', this.refetchOnPrefilterValueChange)
    },

    refetchOnPrefilterValueChange ({ fieldName }) {
      const { src } = this.options

      if (isFieldInFilter(fieldName, src)) {
        this.refresh()
      }
    },

    destroyEvents () {
      this.$root.$off('record-field-change', this.refetchOnPrefilterValueChange)
    },
  },
}
</script>
<style scoped lang="scss">
img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
</style>
