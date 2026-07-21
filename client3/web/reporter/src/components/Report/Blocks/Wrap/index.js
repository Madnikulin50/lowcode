import { h, computed } from 'vue'
import { capitalize } from 'lodash'
import Card from './Card.vue'

const Registry = { Card }
const defaultWrap = 'Card'

export default {
  name: 'BlockWrap',
  props: { block: { type: Object, required: true }, wrap: { type: String, default: defaultWrap } },
  setup(props, { slots }) {
    const cmpName = computed(() => capitalize(props.wrap))
    return () => h(Registry[cmpName.value] || Registry.Card, null, slots)
  },
}