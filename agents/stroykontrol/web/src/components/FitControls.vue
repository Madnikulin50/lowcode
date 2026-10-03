<script setup>
const props = defineProps({ viewer: Object })
const emit = defineEmits(['align'])

const FIT_MODES = [['contain', 'Вписать'], ['cover', 'Заполнить'], ['stretch', 'Растянуть']]
const H_ALIGN = [['left', 'Лево'], ['center', 'По центру'], ['right', 'Право']]
const V_ALIGN = [['top', 'Верх'], ['center', 'По центру'], ['bottom', 'Низ']]

function onDx (e) { props.viewer.manualDX = Number(e.target.value) || 0 }
function onDy (e) { props.viewer.manualDY = Number(e.target.value) || 0 }
function onScale (e) { props.viewer.manualScale = Number(e.target.value) || 100 }
function resetManual () {
  props.viewer.manualDX = 0
  props.viewer.manualDY = 0
  props.viewer.manualScale = 100
}
</script>

<template>
  <div class="fit-row">
    <span>Подгонка РД под ПД:</span>
    <div class="switch">
      <button v-for="[key, label] in FIT_MODES" :key="key" :class="{ active: viewer.fitMode === key }" @click="viewer.fitMode = key">{{ label }}</button>
    </div>
  </div>

  <div class="fit-row">
    <span>Выровнять края РД:</span>
    <div class="switch">
      <button v-for="[h, label] in H_ALIGN" :key="h" @click="emit('align', h, null)">{{ label }}</button>
    </div>
    <div class="switch">
      <button v-for="[v, label] in V_ALIGN" :key="v" @click="emit('align', null, v)">{{ label }}</button>
    </div>
  </div>

  <div class="fit-manual">
    <label>Сдвиг X, px:<input type="number" :value="viewer.manualDX" @change="onDx" /></label>
    <label>Сдвиг Y, px:<input type="number" :value="viewer.manualDY" @change="onDy" /></label>
    <label>Масштаб, %:<input type="number" :value="viewer.manualScale" @change="onScale" /></label>
    <button @click="resetManual">Сброс</button>
  </div>
</template>
