<template>

  <b-badge
      :variant="badgeVariant"
      class="d-inline-block mb-0 ml-2 d-inline-flex align-items-center"
  >
    <div class="align-content-centere mb-0 w-100">
      {{ value }}
    </div>
  </b-badge>


</template>

<script>
export default {
  props: {
    value: {
      type: Number,
      required: true,
    },

    thresholds: {
      type: Array,
      default: () => [],
    },

    variant: {
      type: String,
      default: 'primary',
    },

    textStyle: {
      type: String,
      default: '',
    },
  },

  computed: {
    badgeVariant () {
      const value = this.value

      let progressVariant = this.variant

      if (this.thresholds.length) {
        const { variant } = this.sortedVariants.find(t => value >= t.value) || {}
        progressVariant = variant || progressVariant
      }

      return progressVariant
    },

    sortedVariants () {
      return [...this.thresholds].filter(t => t.value >= 0).sort((a, b) => b.value - a.value)
    },

    textVariant () {
      return ['dark', 'primary'].includes(this.badgeVariant) ? 'text-white' : 'text-dark'
    },
  },
}
</script>
