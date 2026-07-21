<template>
  <button
    :title="$t('title')"
    type="button"
    class="btn btn-outline-extra-light btn-lg nav-icon rounded-circle text-center border-0 d-flex align-items-center justify-content-center position-relative"
    @click="toggleNotifications"
  >
    <font-awesome-icon
      :icon="['far', 'bell']"
      class="text-dark"
    />
    <span
      v-if="unreadCount > 0 && !muted"
      class="badge rounded-pill bg-primary position-absolute notification-badge"
    >
      {{ unreadCount > 9 ? '9+' : unreadCount }}
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance } from 'vue'
import { useNotificationsStore } from '../../store/notifications'

const instance = getCurrentInstance()
const $t = instance!.appContext.config.globalProperties.$t

const notificationsStore = useNotificationsStore()

const unreadCount = computed(() => notificationsStore.unreadCount)
const muted = computed(() => notificationsStore.muted)

function toggleNotifications() {
  notificationsStore.toggleVisibility()
}
</script>

<style lang="scss" scoped>
.notification-badge {
  top: 0;
  right: 0;
  transform: translate(25%, -25%);
  font-size: 0.7rem;
}
</style>
