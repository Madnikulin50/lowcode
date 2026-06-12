<template>
  <div>
    <!-- Extra empty line is added thanks to white-space: pre-line (multivalue) if we write div in multiple lines  -->
    <!-- eslint-disable-next-line -->
    <div v-if="field.options.display === 'number'" :class="classes">{{ formatted }}</div>

    <template v-if="field.options.display === 'progress'">
      <c-progress
        v-for="(v, i) in formatted"
        :key="i"
        :value="parseFloat(v)"
        :min="parseFloat(field.options.min)"
        :max="parseFloat(field.options.max)"
        :labeled="field.options.showValue"
        :relative="field.options.showRelative"
        :progress="field.options.showProgress"
        :striped="field.options.striped"
        :animated="field.options.animated"
        :variant="field.options.variant"
        :thresholds="field.options.thresholds"
        :class="{ 'mt-2': i }"
        style="height: 1.5rem; min-width: 15rem;"
      />
    </template>
    <template v-if="field.options.display === 'colorGrade'">
      <c-color-grade-number
        v-for="(v, i) in formatted"
        :key="i"
        :value="parseFloat(v)"
        :variant="field.options.variant"
        :thresholds="field.options.thresholds"
        :class="{ 'mt-2': i }"
      />
    </template>
    <template v-if="field.options.display === 'trafficLight'">
      <c-traffic-light
        v-for="(v, i) in formatted"
        :key="i"
        :value="parseFloat(v)"
        :variant="field.options.variant"
        :thresholds="field.options.thresholds"
        :class="{ 'mt-2': i }"
        style="height: 2rem; min-width: 5rem;"
      />
    </template>
  </div>
</template>

<script>
import base from './base'
import { components } from 'corteza-lib/vue/dist'
const { CProgress, CColorGradeNumber, CTrafficLight } = components

export default {
  components: {
    CProgress,
    CColorGradeNumber,
    CTrafficLight,
  },

  extends: base,

  computed: {
    formatted () {
      if (this.value === undefined) {
        return this.field.options.display === 'number' ? undefined : [this.field.options.min]
      }

      const value = this.field.isMulti ? this.value : [this.value]

      if (this.field.options.display === 'number') {
        return value.map(v => this.field.formatValue(v)).join(this.field.options.multiDelimiter)
      } else {
        return value.length ? value : [this.field.options.min]
      }
    },
  },
}
</script>
