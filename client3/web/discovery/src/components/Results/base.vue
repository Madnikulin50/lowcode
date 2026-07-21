<script setup>
import { computed } from 'vue'
import { filters } from 'corteza-lib/vue/dist'

const props = defineProps({
  index: {
    type: Number,
    required: true,
  },
  hit: {
    type: Object,
    required: true,
  },
  showMap: {
    type: Boolean,
    required: true,
  },
})

const defaultBlacklistedFields = ['deleted', 'created', 'updated', 'security', 'catch_all']

const createdBy = computed(() => {
  const { by } = props.hit.value?.created || {}
  return by
})

const createdAt = computed(() => {
  const { at } = props.hit.value?.created || {}
  return at ? filters.locFullDateTime(at) : at
})

const updatedAt = computed(() => {
  const { at } = props.hit.value?.updated || {}
  return at ? filters.locFullDateTime(at) : at
})

const blacklistedFields = computed(() => defaultBlacklistedFields)

function limitData() {
  const out = {}
  if (props.hit.value) {
    for (const key in props.hit.value) {
      const value = props.hit.value[key]
      if (!!value && blacklistedFields.value.indexOf(key) < 0) {
        out[key] = value
      }
    }
  }
  if (createdBy.value) out.createdBy = createdBy.value
  if (createdAt.value) out.createdAt = createdAt.value
  if (updatedAt.value) out.updatedAt = updatedAt.value
  return out
}
</script>
