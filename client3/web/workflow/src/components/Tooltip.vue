<template>
  <div
    :id="kind"
    class="d-none"
  >
    <div class="card text-left text-dark shadow-sm" style="width: 18rem;">
      <img v-if="img" :src="img" :alt="kind" height="100" class="card-img-top" />
      <div class="card-body py-2 px-3">
        <h6 class="card-title mb-1">{{ title }}</h6>
        <p class="card-text small text-muted mb-0">{{ text }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: true },
  kind: { type: String, required: true },
  img: { type: String, required: true },
  text: { type: String, required: true },
})
</script>
