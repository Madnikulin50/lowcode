<script setup>
import { ICONS, MODE_META } from '../constants.js'
import Icon from './Icon.vue'

defineProps({ viewer: Object })
const emit = defineEmits(['assistant'])

const PRIMARY_MODES = ['side', 'overlay', 'diff']
</script>

<template>
  <div class="modes">
    <button v-for="key in PRIMARY_MODES" :key="key" :class="{ active: viewer.mode === key }" @click="viewer.mode = key">
      <Icon :svg="ICONS[key]" />{{ MODE_META[key].label }}
    </button>
    <span class="mode-sep"></span>
    <button :class="{ active: viewer.mode === 'text' }" @click="viewer.mode = 'text'">
      <Icon :svg="ICONS.text" />{{ MODE_META.text.label }}
    </button>
    <span class="mode-sep"></span>
    <button class="assistant-btn" title="Спросить ассистента об открытых листах" @click="emit('assistant')">✦ Ассистент</button>
  </div>
</template>
