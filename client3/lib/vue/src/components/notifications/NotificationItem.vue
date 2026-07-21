<template>
  <div class="notification-item-container px-3 pt-3 pb-2 border-bottom">
    <div
      class="notification-item border rounded bg-white p-3 position-relative list-group-item"
      :class="{ 'read': notification.readAt }"
      @click="emit('click', notification)"
    >
      <div
        class="action-menu bg-white pb-2 ps-2"
        style="margin-left: -1rem;"
      >
        <div class="dropdown">
          <button
            class="btn text-decoration-none border-0 d-flex align-items-center justify-content-center"
            style="width: 2rem; height: 2rem;"
            type="button"
            data-bs-toggle="dropdown"
          >
            <font-awesome-icon
              :icon="['fas', 'ellipsis-v']"
              class="text-secondary"
              style="margin-top: 0.3rem;"
            />
          </button>

          <ul class="dropdown-menu dropdown-menu-end">
            <li>
              <button
                v-if="!notification.readAt"
                class="dropdown-item"
                @click.stop="emit('mark-read', notification)"
              >
                <font-awesome-icon
                  :icon="['far', 'envelope-open']"
                  class="text-primary"
                />
                {{ $t('markAsRead') }}
              </button>

              <button
                v-else
                class="dropdown-item"
                @click.stop="emit('mark-unread', notification)"
              >
                <font-awesome-icon
                  :icon="['far', 'envelope']"
                  class="text-primary"
                />
                {{ $t('markAsUnread') }}
              </button>
            </li>

            <li>
              <hr class="dropdown-divider">
            </li>

            <li>
              <CInputConfirm
                :text="$t('delete')"
                show-icon
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="emit('delete', notification)"
              />
            </li>
          </ul>
        </div>
      </div>

      <component
        :is="notificationComponent"
        :notification="notification"
        class="notification-item-content"
      />
    </div>

    <div class="d-flex justify-content-end mt-2">
      <div
        :title="notification.createdAt"
        class="text-muted small cursor-pointer"
        @click="emit('click', notification)"
      >
        {{ $locFullDateTime(notification.createdAt) }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance } from 'vue'
import NotificationTypes from './types/index.ts'
import CInputConfirm from '../input/CInputConfirm.vue'

const { $t } = getCurrentInstance()!.appContext.config.globalProperties as any

const props = defineProps<{
  notification: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'click', notification: Record<string, any>): void
  (e: 'mark-read', notification: Record<string, any>): void
  (e: 'mark-unread', notification: Record<string, any>): void
  (e: 'delete', notification: Record<string, any>): void
}>()

const notificationComponent = computed(() => {
  let { kind } = props.notification
  kind = kind.charAt(0).toUpperCase() + kind.slice(1)

  return NotificationTypes[`Notification${kind}` as keyof typeof NotificationTypes]
})
</script>

<style lang="scss" scoped>
.notification-item-container {
  &:hover {
    background-color: var(--light) !important;
  }

  .notification-item {
    transition: background-color 0.2s ease;
    cursor: pointer;

    &.read {
      .notification-item-content {
        opacity: 0.5 !important;
      }
    }
  }

  &:hover {
    .action-menu {
      opacity: 1 !important;
      pointer-events: auto;
    }
  }
  .action-menu {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    z-index: 2;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.2s ease;
  }
}
</style>
