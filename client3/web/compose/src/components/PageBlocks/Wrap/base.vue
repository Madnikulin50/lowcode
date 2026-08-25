<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { compose, NoID } from 'corteza-lib/js/dist'
import { evaluatePrefilter } from 'corteza-webapp-compose/src/lib/record-filter'

const $auth = window.__auth
const route = useRoute()

const props = defineProps({
  block: { type: compose.PageBlock, required: true },
  record: { type: compose.Record, required: false, default: undefined },
  scrollableBody: { type: Boolean, required: false, default: true },
  cardClass: { type: String, required: false, default: '' },
  bodyClass: { type: String, required: false, default: '' },
  headerClass: { type: String, required: false, default: '' },
  magnified: { type: Boolean, required: false, default: false },
})

const blockID = computed(() => {
  const { blockID, meta } = props.block || {}
  return meta?.customID || blockID
})

const customCSSClass = computed(() => props.block?.meta?.customCSSClass)

const blockClass = computed(() => {
  return [
    'block',
    { border: props.block.style.border?.enabled },
    props.block.kind,
  ]
})

const isBlockMagnified = computed(() => {
  const { magnifiedBlockID } = route.query
  return props.magnified && magnifiedBlockID === props.block.blockID
})

const isAnotherBlockMagnified = computed(() => {
  const { magnifiedBlockID } = route.query
  return magnifiedBlockID && magnifiedBlockID !== props.block.blockID
})

const showMagnifyButton = computed(() => {
  return (props.block.options.magnifyOption || isBlockMagnified.value) && !isAnotherBlockMagnified.value
})

const showHeader = computed(() => {
  return [props.block.title, props.block.description, props.block.options.showRefresh, showMagnifyButton.value].some(c => !!c)
})

const showOptions = computed(() => {
  return [props.block.options.magnifyOption, props.block.options.showRefresh, showMagnifyButton.value].some(c => !!c)
})

const magnifyParams = computed(() => {
  const params = props.block.blockID === NoID ? { block: props.block } : { blockID: props.block.blockID }
  return isBlockMagnified.value ? undefined : params
})

const blockTitle = computed(() => {
  try {
    return evaluatePrefilter(props.block.title, {
      record: props.record,
      user: $auth.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth.user || {}).userID || NoID,
    })
  } catch (e) {
    return e
  }
})

const blockDescription = computed(() => {
  try {
    return evaluatePrefilter(props.block.description, {
      record: props.record,
      user: $auth.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth.user || {}).userID || NoID,
    })
  } catch (e) {
    return e
  }
})
</script>
