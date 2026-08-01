<template>
  <div
    v-if="namespace?.canManageNamespace"
    class="d-flex flex-column w-100 flex-grow-1"
    style="min-height: 0"
  >
    <router-view
      :namespace="namespace"
    />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const router = useRouter()

onMounted(() => {
  if (!props.namespace?.canManageNamespace) {
    router.push({ name: 'pages' })
  }
})
</script>
