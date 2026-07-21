import { defineComponent, h } from 'vue'
import { capitalize } from 'lodash'
import { components } from 'corteza-lib/vue/dist'

const {
  CReportChart,
  CReportMetric,
  CReportTable,
  CReportText,
} = components

const Registry = {
  Chart: CReportChart,
  Metric: CReportMetric,
  Table: CReportTable,
  Text: CReportText,
}

function GetComponent ({ displayElement }) {
  if (!displayElement) {
    throw new Error('displayElement prop missing')
  }

  const { kind } = displayElement
  if (Object.hasOwnProperty.call(Registry, capitalize(kind))) {
    return Registry[kind]
  }

  throw new Error('unknown displayElement kind: ' + kind)
}

export default defineComponent({
  name: 'DisplayElement',
  props: {
    displayElement: {
      type: Object,
      required: true,
    },
  },
  render () {
    return h(GetComponent(this.$props), this.$attrs, this.$slots)
  },
})
