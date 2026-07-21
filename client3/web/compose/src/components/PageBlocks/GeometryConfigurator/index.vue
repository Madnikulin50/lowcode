<template>
  <div>
    <ul class="nav nav-tabs" role="tablist">
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ active: activeTab === 0 }"
          @click="activeTab = 0"
        >
          {{ $t('geometry.viewLabel') }}
        </button>
      </li>
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ active: activeTab === 1 }"
          @click="activeTab = 1"
        >
          {{ $t('geometry.feedLabel') }}
        </button>
      </li>
    </ul>

    <div class="tab-content">
      <div
        v-show="activeTab === 0"
        class="tab-pane active"
      >
        <configurator
          v-bind="$attrs"
        />
      </div>

      <div
        v-show="activeTab === 1"
        class="tab-pane active"
      >
        <feed-source
          v-bind="$attrs"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import FeedSource from './FeedSource'
import Configurator from './Configurator'

defineOptions({ inheritAttrs: false })

const activeTab = ref(0)
</script>
