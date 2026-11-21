import { h } from 'vue'
import { capitalize } from 'lodash'
import { components } from 'corteza-lib/vue/dist'

const { CReportChart, CReportMetric, CReportTable, CReportText } = components

const Registry = {
  Chart: CReportChart,
  Metric: CReportMetric,
  Table: CReportTable,
  Text: CReportText,
}

function GetComponent({ displayElement }) {
  if (!displayElement) throw new Error('displayElement prop missing')
  const { kind } = displayElement
  if (Object.prototype.hasOwnProperty.call(Registry, capitalize(kind))) return Registry[kind]
  throw new Error('unknown displayElement kind: ' + kind)
}

const DisplayElement = (props, ctx) => {
  return h(GetComponent(props), ctx.attrs, ctx.slots)
}

export default DisplayElement