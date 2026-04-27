<template>
  <e-charts
    ref="chart"
    :option="chart"
    :theme="theme"
    autoresize
    class="position-absolute w-100 h-100 overflow-hidden"
    v-on="$listeners"
  />
</template>

<script>
import { shared } from '../../../../../lib/js/dist'

export default {
  props: {
    chart: {
      type: shared.Chart,
      required: true,
    },
  },

  computed: {
    theme () {
      const { darkMode } = this.chart || {}
      return darkMode ? 'dark' : 'light'
    },
  },

  beforeDestroy () {
    if (this.$refs.chart) {
      this.$refs.chart.dispose()
    }
  },
}
</script>
