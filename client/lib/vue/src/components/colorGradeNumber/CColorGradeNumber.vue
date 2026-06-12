<template>
  <div
      :class="textClass">
    {{ value }}
  </div>
</template>

<script>
export default {
  props: {
    value: {
      type: Number,
      required: true,
    },

    variant: {
      type: String,
      default: 'success',
    },

    thresholds: {
      type: Array,
      default: () => [],
    },

    textStyle: {
      type: String,
      default: '',
    },
  },

  computed: {

    textClass () {
      const value = this.value

      let progressVariant = this.variant

      if (this.thresholds.length) {
        const { variant } = this.sortedVariants.find(t => value >= t.value) || {}
        progressVariant = variant || progressVariant
      }

      return "text-" + progressVariant
    },

    sortedVariants () {
      return [...this.thresholds].filter(t => t.value >= 0).sort((a, b) => b.value - a.value)
    },
  },
}
</script>
