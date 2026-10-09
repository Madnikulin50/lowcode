<script setup lang="ts">
defineProps<{
  labels?: Record<string, any>
  src?: string | Record<string, any>
  mime?: string
  meta?: Record<string, any>
  name?: string | null
  previewStyle?: Record<string, any>
  previewClass?: any[]
  inline?: boolean
}>()
</script>
