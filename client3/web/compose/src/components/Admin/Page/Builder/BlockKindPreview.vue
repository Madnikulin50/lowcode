<template>
  <svg
    class="block-kind-preview"
    viewBox="0 0 320 180"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    :aria-label="kind"
  >
    <rect width="320" height="180" rx="10" fill="#f4f6fb" />
    <rect x="10" y="10" width="300" height="160" rx="8" fill="#fff" stroke="#e3e6f0" />
    <path d="M10 38h300" stroke="#e3e6f0" />
    <rect x="22" y="20" width="88" height="8" rx="2" fill="#3a3b45" opacity=".85" />
    <circle cx="288" cy="24" r="5" fill="none" stroke="#858796" stroke-width="1.4" />
    <path d="M292 28l4 4" stroke="#858796" stroke-width="1.4" fill="none" stroke-linecap="round" />

    <!-- Record list -->
    <g v-if="kind === 'RecordList'">
      <rect x="20" y="48" width="52" height="16" rx="3" fill="#4e73df" />
      <rect x="76" y="48" width="40" height="16" rx="3" fill="#eaecf4" />
      <rect x="214" y="48" width="86" height="16" rx="3" fill="#fff" stroke="#d1d3e2" />
      <rect x="20" y="72" width="280" height="12" rx="2" fill="#eaecf4" />
      <rect x="20" y="88" width="280" height="22" rx="2" fill="#fff" />
      <rect x="20" y="110" width="280" height="22" rx="2" fill="#f8f9fc" />
      <rect x="20" y="132" width="280" height="22" rx="2" fill="#fff" />
      <rect x="28" y="78" width="8" height="6" rx="1" fill="#b7b9cc" />
      <rect x="46" y="78" width="70" height="6" rx="1" fill="#b7b9cc" />
      <rect x="140" y="78" width="50" height="6" rx="1" fill="#b7b9cc" />
      <rect x="220" y="78" width="40" height="6" rx="1" fill="#b7b9cc" />
      <rect x="28" y="96" width="8" height="6" rx="1" fill="#d1d3e2" />
      <rect x="46" y="96" width="86" height="6" rx="1" fill="#5a5c69" />
      <rect x="140" y="94" width="44" height="10" rx="5" fill="#1cc88a" opacity=".25" />
      <rect x="148" y="97" width="28" height="4" rx="1" fill="#1cc88a" />
      <rect x="28" y="118" width="8" height="6" rx="1" fill="#d1d3e2" />
      <rect x="46" y="118" width="70" height="6" rx="1" fill="#5a5c69" />
      <rect x="140" y="116" width="44" height="10" rx="5" fill="#f6c23e" opacity=".3" />
      <rect x="148" y="119" width="28" height="4" rx="1" fill="#d4a017" />
    </g>

    <!-- Chart -->
    <g v-else-if="kind === 'Chart'">
      <rect x="28" y="142" width="14" height="18" rx="1" fill="#4e73df" />
      <rect x="48" y="124" width="14" height="36" rx="1" fill="#4e73df" />
      <rect x="68" y="110" width="14" height="50" rx="1" fill="#4e73df" />
      <rect x="88" y="128" width="14" height="32" rx="1" fill="#4e73df" />
      <rect x="108" y="96" width="14" height="64" rx="1" fill="#4e73df" />
      <rect x="128" y="116" width="14" height="44" rx="1" fill="#4e73df" />
      <path d="M28 118 C48 108, 68 92, 88 100 S128 86, 148 90" fill="none" stroke="#1cc88a" stroke-width="2.2" />
      <circle cx="228" cy="108" r="36" fill="none" stroke="#eaecf4" stroke-width="14" />
      <circle cx="228" cy="108" r="36" fill="none" stroke="#4e73df" stroke-width="14" stroke-dasharray="70 156" stroke-linecap="butt" transform="rotate(-90 228 108)" />
      <circle cx="228" cy="108" r="36" fill="none" stroke="#1cc88a" stroke-width="14" stroke-dasharray="40 186" stroke-dashoffset="-70" transform="rotate(-90 228 108)" />
      <circle cx="228" cy="108" r="36" fill="none" stroke="#f6c23e" stroke-width="14" stroke-dasharray="46 180" stroke-dashoffset="-110" transform="rotate(-90 228 108)" />
    </g>

    <!-- Metric -->
    <g v-else-if="kind === 'Metric'">
      <rect x="22" y="52" width="132" height="50" rx="6" fill="#f8f9fc" />
      <rect x="166" y="52" width="132" height="50" rx="6" fill="#f8f9fc" />
      <rect x="22" y="110" width="132" height="50" rx="6" fill="#f8f9fc" />
      <rect x="166" y="110" width="132" height="50" rx="6" fill="#f8f9fc" />
      <rect x="32" y="62" width="54" height="6" rx="1" fill="#858796" />
      <rect x="32" y="76" width="48" height="16" rx="2" fill="#4e73df" />
      <rect x="176" y="62" width="64" height="6" rx="1" fill="#858796" />
      <rect x="176" y="76" width="36" height="16" rx="2" fill="#1cc88a" />
      <rect x="32" y="120" width="48" height="6" rx="1" fill="#858796" />
      <rect x="32" y="134" width="72" height="16" rx="2" fill="#4e73df" />
      <rect x="176" y="120" width="70" height="6" rx="1" fill="#858796" />
      <rect x="176" y="134" width="56" height="16" rx="2" fill="#36b9cc" />
    </g>

    <!-- Calendar -->
    <g v-else-if="kind === 'Calendar'">
      <rect x="20" y="48" width="24" height="14" rx="3" fill="#eaecf4" />
      <rect x="148" y="50" width="24" height="8" rx="2" fill="#5a5c69" />
      <rect x="276" y="48" width="24" height="14" rx="3" fill="#eaecf4" />
      <g v-for="(row, r) in 4" :key="'cal-r'+r">
        <g v-for="c in 7" :key="'cal-'+r+c">
          <rect :x="20 + (c-1)*40" :y="70 + r*24" width="36" height="20" rx="3" fill="#f8f9fc" />
        </g>
      </g>
      <rect x="62" y="96" width="32" height="8" rx="2" fill="#4e73df" />
      <rect x="142" y="120" width="28" height="8" rx="2" fill="#1cc88a" />
      <rect x="222" y="72" width="24" height="8" rx="2" fill="#e74a3b" />
    </g>

    <!-- Map -->
    <g v-else-if="kind === 'Geometry'">
      <rect x="18" y="46" width="284" height="114" rx="4" fill="#dbe7f3" />
      <path d="M18 120 C70 100, 110 140, 170 112 S250 90, 302 130 V160 H18 Z" fill="#c5d9c8" />
      <path d="M40 90 C90 70, 140 95, 200 78 S270 70, 302 88" fill="none" stroke="#9bb8d4" stroke-width="6" />
      <g v-for="p in pins" :key="p.x" :transform="`translate(${p.x} ${p.y})`">
        <path d="M0 -14 C6 -14 10 -9 10 -4 C10 2 0 12 0 12 S-10 2 -10 -4 C-10 -9 -6 -14 0 -14Z" :fill="p.color" />
        <circle cy="-5" r="3.2" fill="#fff" />
      </g>
      <rect x="26" y="54" width="86" height="18" rx="4" fill="#fff" stroke="#d1d3e2" />
      <circle cx="38" cy="63" r="4" fill="none" stroke="#858796" stroke-width="1.3" />
    </g>

    <!-- Record form -->
    <g v-else-if="kind === 'Record'">
      <rect x="22" y="52" width="70" height="6" rx="1" fill="#858796" />
      <rect x="22" y="64" width="276" height="18" rx="3" fill="#fff" stroke="#d1d3e2" />
      <rect x="22" y="92" width="90" height="6" rx="1" fill="#858796" />
      <rect x="22" y="104" width="130" height="18" rx="3" fill="#fff" stroke="#d1d3e2" />
      <rect x="168" y="92" width="70" height="6" rx="1" fill="#858796" />
      <rect x="168" y="104" width="130" height="18" rx="3" fill="#fff" stroke="#d1d3e2" />
      <rect x="22" y="132" width="54" height="6" rx="1" fill="#858796" />
      <rect x="22" y="144" width="90" height="14" rx="7" fill="#eaecf4" />
      <circle cx="36" cy="151" r="5" fill="#4e73df" />
    </g>

    <!-- Content -->
    <g v-else-if="kind === 'Content'">
      <rect x="24" y="54" width="160" height="10" rx="2" fill="#3a3b45" />
      <rect x="24" y="74" width="272" height="6" rx="1" fill="#858796" />
      <rect x="24" y="88" width="250" height="6" rx="1" fill="#858796" />
      <rect x="24" y="102" width="264" height="6" rx="1" fill="#858796" />
      <rect x="24" y="122" width="200" height="6" rx="1" fill="#b7b9cc" />
      <rect x="24" y="136" width="220" height="6" rx="1" fill="#b7b9cc" />
    </g>

    <!-- File -->
    <g v-else-if="kind === 'File'">
      <g v-for="(f, i) in files" :key="i" :transform="`translate(${22 + (i%3)*94} ${54 + Math.floor(i/3)*54})`">
        <rect width="86" height="46" rx="6" fill="#f8f9fc" stroke="#e3e6f0" />
        <rect x="8" y="10" width="22" height="26" rx="3" :fill="f" />
        <rect x="36" y="14" width="40" height="6" rx="1" fill="#5a5c69" />
        <rect x="36" y="26" width="28" height="5" rx="1" fill="#b7b9cc" />
      </g>
    </g>

    <!-- IFrame -->
    <g v-else-if="kind === 'IFrame'">
      <rect x="20" y="48" width="280" height="112" rx="4" fill="#f8f9fc" stroke="#d1d3e2" />
      <rect x="20" y="48" width="280" height="18" rx="4" fill="#eaecf4" />
      <circle cx="32" cy="57" r="3" fill="#e74a3b" />
      <circle cx="42" cy="57" r="3" fill="#f6c23e" />
      <circle cx="52" cy="57" r="3" fill="#1cc88a" />
      <rect x="66" y="52" width="180" height="10" rx="3" fill="#fff" />
      <rect x="36" y="80" width="120" height="8" rx="2" fill="#4e73df" opacity=".35" />
      <rect x="36" y="96" width="248" height="6" rx="1" fill="#d1d3e2" />
      <rect x="36" y="110" width="220" height="6" rx="1" fill="#d1d3e2" />
      <rect x="36" y="124" width="180" height="6" rx="1" fill="#d1d3e2" />
    </g>

    <!-- Comments -->
    <g v-else-if="kind === 'Comment'">
      <circle cx="36" cy="68" r="10" fill="#4e73df" />
      <rect x="52" y="56" width="160" height="28" rx="8" fill="#f8f9fc" />
      <circle cx="284" cy="108" r="10" fill="#1cc88a" />
      <rect x="108" y="96" width="160" height="28" rx="8" fill="#eaf1ff" />
      <rect x="20" y="140" width="232" height="20" rx="10" fill="#fff" stroke="#d1d3e2" />
      <rect x="260" y="140" width="40" height="20" rx="6" fill="#4e73df" />
    </g>

    <!-- Automation -->
    <g v-else-if="kind === 'Automation'">
      <rect x="28" y="80" width="84" height="28" rx="6" fill="#4e73df" />
      <polygon points="48,88 48,100 60,94" fill="#fff" />
      <rect x="64" y="90" width="36" height="8" rx="2" fill="#fff" />
      <rect x="122" y="80" width="84" height="28" rx="6" fill="#eaecf4" />
      <rect x="140" y="90" width="48" height="8" rx="2" fill="#5a5c69" />
      <rect x="216" y="80" width="76" height="28" rx="6" fill="#1cc88a" />
      <rect x="232" y="90" width="44" height="8" rx="2" fill="#fff" />
    </g>

    <!-- Progress -->
    <g v-else-if="kind === 'Progress'">
      <rect x="28" y="70" width="264" height="6" rx="1" fill="#858796" />
      <rect x="28" y="92" width="264" height="16" rx="8" fill="#eaecf4" />
      <rect x="28" y="92" width="176" height="16" rx="8" fill="#4e73df" />
      <rect x="28" y="122" width="48" height="14" rx="2" fill="#4e73df" />
      <rect x="84" y="126" width="80" height="6" rx="1" fill="#858796" />
    </g>

    <!-- Organizer / kanban -->
    <g v-else-if="kind === 'RecordOrganizer'">
      <g v-for="(col, i) in 3" :key="'org'+i" :transform="`translate(${22 + i*94} 50)`">
        <rect width="86" height="110" rx="6" fill="#f8f9fc" />
        <rect x="8" y="10" width="48" height="6" rx="1" fill="#5a5c69" />
        <rect x="8" y="26" width="70" height="22" rx="4" fill="#fff" stroke="#e3e6f0" />
        <rect x="8" y="54" width="70" height="22" rx="4" fill="#fff" stroke="#e3e6f0" />
        <rect x="8" y="82" width="70" height="18" rx="4" :fill="i === 1 ? '#eaf1ff' : '#fff'" stroke="#e3e6f0" />
      </g>
    </g>

    <!-- Revisions -->
    <g v-else-if="kind === 'RecordRevisions'">
      <line x1="40" y1="56" x2="40" y2="156" stroke="#d1d3e2" stroke-width="2" />
      <g v-for="(item, i) in 3" :key="'rev'+i">
        <circle cx="40" :cy="68 + i*36" r="6" :fill="i === 0 ? '#4e73df' : '#b7b9cc'" />
        <rect x="56" :y="56 + i*36" width="220" height="24" rx="4" fill="#f8f9fc" />
        <rect x="66" :y="64 + i*36" width="90" height="6" rx="1" fill="#5a5c69" />
        <rect x="170" :y="64 + i*36" width="50" height="6" rx="1" fill="#b7b9cc" />
      </g>
    </g>

    <!-- Report -->
    <g v-else-if="kind === 'Report'">
      <rect x="22" y="52" width="160" height="108" rx="4" fill="#f8f9fc" />
      <rect x="30" y="62" width="144" height="10" rx="2" fill="#eaecf4" />
      <rect x="30" y="78" width="144" height="8" rx="1" fill="#fff" />
      <rect x="30" y="92" width="144" height="8" rx="1" fill="#fff" />
      <rect x="30" y="106" width="144" height="8" rx="1" fill="#fff" />
      <rect x="30" y="120" width="144" height="8" rx="1" fill="#fff" />
      <rect x="196" y="52" width="102" height="108" rx="4" fill="#f8f9fc" />
      <rect x="210" y="128" width="12" height="20" rx="1" fill="#4e73df" />
      <rect x="226" y="110" width="12" height="38" rx="1" fill="#4e73df" />
      <rect x="242" y="96" width="12" height="52" rx="1" fill="#4e73df" />
      <rect x="258" y="118" width="12" height="30" rx="1" fill="#4e73df" />
    </g>

    <!-- Tabs -->
    <g v-else-if="kind === 'Tabs'">
      <rect x="20" y="48" width="72" height="22" rx="4" fill="#fff" stroke="#4e73df" />
      <rect x="28" y="56" width="44" height="6" rx="1" fill="#4e73df" />
      <rect x="96" y="48" width="72" height="22" rx="4" fill="#eaecf4" />
      <rect x="108" y="56" width="40" height="6" rx="1" fill="#858796" />
      <rect x="172" y="48" width="72" height="22" rx="4" fill="#eaecf4" />
      <rect x="184" y="56" width="40" height="6" rx="1" fill="#858796" />
      <rect x="20" y="70" width="280" height="90" rx="4" fill="#f8f9fc" />
      <rect x="32" y="86" width="160" height="8" rx="2" fill="#d1d3e2" />
      <rect x="32" y="104" width="240" height="6" rx="1" fill="#e3e6f0" />
      <rect x="32" y="118" width="200" height="6" rx="1" fill="#e3e6f0" />
    </g>

    <!-- Navigation -->
    <g v-else-if="kind === 'Navigation'">
      <rect x="20" y="72" width="280" height="36" rx="8" fill="#eaecf4" />
      <rect x="20" y="72" width="70" height="36" rx="8" fill="#4e73df" />
      <rect x="38" y="86" width="34" height="8" rx="2" fill="#fff" />
      <rect x="106" y="86" width="40" height="8" rx="2" fill="#5a5c69" />
      <rect x="164" y="86" width="48" height="8" rx="2" fill="#5a5c69" />
      <rect x="230" y="86" width="44" height="8" rx="2" fill="#5a5c69" />
      <rect x="20" y="108" width="110" height="44" rx="4" fill="#fff" stroke="#e3e6f0" />
      <rect x="32" y="120" width="70" height="6" rx="1" fill="#5a5c69" />
      <rect x="32" y="134" width="54" height="6" rx="1" fill="#b7b9cc" />
    </g>

    <!-- Rule chain -->
    <g v-else-if="kind === 'RuleChain'">
      <rect x="28" y="86" width="64" height="28" rx="6" fill="#4e73df" />
      <path d="M92 100 H124" stroke="#858796" stroke-width="2" marker-end="url(#arrow)" />
      <rect x="124" y="70" width="72" height="28" rx="6" fill="#36b9cc" />
      <rect x="124" y="114" width="72" height="28" rx="6" fill="#f6c23e" />
      <path d="M196 84 H228" stroke="#858796" stroke-width="2" />
      <path d="M196 128 H228" stroke="#858796" stroke-width="2" />
      <rect x="228" y="86" width="64" height="28" rx="6" fill="#1cc88a" />
      <path d="M160 98 V114" stroke="#858796" stroke-width="2" />
    </g>

    <!-- AI chat -->
    <g v-else-if="kind === 'AiChat'">
      <rect x="36" y="56" width="180" height="36" rx="10" fill="#f8f9fc" />
      <rect x="104" y="100" width="180" height="36" rx="10" fill="#eaf1ff" />
      <rect x="20" y="148" width="236" height="18" rx="9" fill="#fff" stroke="#d1d3e2" />
      <circle cx="288" cy="157" r="10" fill="#4e73df" />
    </g>
  </svg>
</template>

<script setup>
defineProps({
  kind: { type: String, required: true },
})

const pins = [
  { x: 92, y: 92, color: '#e74a3b' },
  { x: 168, y: 78, color: '#4e73df' },
  { x: 236, y: 108, color: '#1cc88a' },
]
const files = ['#4e73df', '#e74a3b', '#1cc88a', '#f6c23e', '#36b9cc', '#858796']
</script>

<style scoped>
.block-kind-preview {
  display: block;
  width: 100%;
  height: auto;
}
</style>
