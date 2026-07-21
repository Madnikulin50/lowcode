import { h } from 'vue'
import { capitalize } from 'lodash'

import Card from './Card.vue'
import Plain from './Plain.vue'

const Registry = {
  Card,
  Plain,
}

const defaultWrap = 'Card'

function GetWrapComponent({ block }) {
  const { kind = defaultWrap } = block?.style?.wrap || {}
  const cmpName = capitalize(kind)
  if (Object.hasOwnProperty.call(Registry, cmpName)) {
    return Registry[capitalize(cmpName)]
  }
  throw new Error('unknown wrap: ' + kind)
}

function Wrap(props, { slots, attrs }) {
  const component = GetWrapComponent(props)
  return h(component, { ...props, ...attrs }, slots)
}

export { GetWrapComponent, Registry }
export default Wrap