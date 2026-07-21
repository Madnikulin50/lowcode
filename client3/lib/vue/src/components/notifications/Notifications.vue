<template>
  <div class="h-100 d-flex flex-column">
    <ul class="nav nav-tabs" role="tablist">
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ active: activeTab === 0 }"
          @click="activeTab = 0"
        >
          {{ $t('unread') }}
        </button>
      </li>
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ active: activeTab === 1 }"
          @click="activeTab = 1"
        >
          {{ $t('all') }}
        </button>
      </li>
      <li
        class="nav-item ms-auto d-flex align-items-center"
        style="min-width: 6rem;"
      >
        <button
          v-if="hasUnread"
          :title="$t('markAllAsRead')"
          class="btn btn-outline-light p-2 border-0 d-flex align-items-center justify-content-center"
          style="width: 2rem; height: 2rem;"
          @click="handleMarkAllAsRead"
        >
          <font-awesome-icon
            :icon="['fas', 'check-double']"
            class="h6 mb-0 text-primary"
          />
        </button>

        <button
          :title="$t(muted ? 'unmute' : 'mute')"
          class="btn btn-outline-light p-2 border-0 d-flex align-items-center justify-content-center"
          style="width: 2rem; height: 2rem;"
          @click="toggleMuted"
        >
          <font-awesome-icon
            :icon="['fas', muted ? 'bell-slash' : 'bell']"
            class="h6 mb-0"
            :class="{ 'text-secondary': muted, 'text-primary': !muted }"
          />
        </button>
      </li>
    </ul>

    <div class="tab-content h-100 overflow-hidden d-flex flex-column">
      <div
        class="tab-pane d-flex flex-column h-100 p-0"
        :class="{ show: activeTab === 0, active: activeTab === 0 }"
      >
        <div class="overflow-auto flex-grow-1 h-100">
          <div
            v-if="loading"
            class="d-flex justify-content-center p-5"
          >
            <span class="spinner-border text-primary" />
          </div>

          <div
            v-else-if="notifications.length > 0"
            class="list-group"
          >
            <NotificationItem
              v-for="notification in notifications"
              :key="notification.notificationID"
              :notification="notification"
              @click="onNotificationClick(notification)"
              @mark-read="onMarkAsRead"
              @mark-unread="onMarkAsUnread"
              @delete="onDeleteNotification"
            />

            <div
              v-if="hasMorePages"
              class="text-center my-3"
            >
              <button
                class="btn btn-outline-primary btn-sm"
                :disabled="loadingMore"
                @click="loadMore()"
              >
                <span
                  v-if="loadingMore"
                  class="spinner-border spinner-border-sm"
                />
                <span v-else>{{ $t('loadMore') }}</span>
              </button>
            </div>
          </div>

          <div
            v-else
            class="text-center p-5"
          >
            <font-awesome-icon
              :icon="['far', 'bell']"
              class="text-secondary mb-3"
              size="3x"
            />
            <p class="text-secondary">
              {{ $t('empty') }}
            </p>
            <p class="text-muted small">
              {{ $t('emptyDescription') }}
            </p>
          </div>
        </div>
      </div>

      <div
        class="tab-pane d-flex flex-column h-100 p-0"
        :class="{ show: activeTab === 1, active: activeTab === 1 }"
      >
        <div class="overflow-auto flex-grow-1 h-100">
          <div
            v-if="loading"
            class="d-flex justify-content-center p-5"
          >
            <span class="spinner-border text-primary" />
          </div>

          <div
            v-else-if="notifications.length > 0"
            class="list-group"
          >
            <NotificationItem
              v-for="notification in notifications"
              :key="notification.notificationID"
              :notification="notification"
              @click="onNotificationClick(notification)"
              @mark-read="onMarkAsRead"
              @mark-unread="onMarkAsUnread"
              @delete="onDeleteNotification"
            />

            <div
              v-if="hasMorePages"
              class="text-center my-3"
            >
              <button
                class="btn btn-outline-primary btn-sm"
                :disabled="loadingMore"
                @click="loadMore()"
              >
                <span
                  v-if="loadingMore"
                  class="spinner-border spinner-border-sm"
                />
                <span v-else>{{ $t('loadMore') }}</span>
              </button>
            </div>
          </div>

          <div
            v-else
            class="text-center p-5"
          >
            <font-awesome-icon
              :icon="['far', 'bell']"
              class="text-secondary mb-3"
              size="3x"
            />
            <p class="text-secondary">
              {{ $t('empty') }}
            </p>
            <p class="text-muted small">
              {{ $t('emptyDescription') }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { getCurrentInstance } from 'vue'
import { useNotificationsStore } from '../../store/notifications'
import NotificationItem from './NotificationItem.vue'
import { useToast } from '../../composables/useToast'

const { $t } = getCurrentInstance()!.appContext.config.globalProperties as any
const notificationsStore = useNotificationsStore()
const { toastSuccess, toastError } = useToast()

const activeTab = ref(0)
const loading = ref(false)
const loadingMore = ref(false)

const notifications = computed(() => notificationsStore.notifications)
const hasUnread = computed(() => notificationsStore.hasUnread)
const hasMorePages = computed(() => notificationsStore.hasMorePages)
const muted = computed(() => notificationsStore.muted)

watch(activeTab, () => {
  loading.value = true
  notificationsStore.setPageCursor(null)
  loadNotifications()
    .finally(() => {
      setTimeout(() => {
        loading.value = false
      }, 300)
    })
})

function fetchNotifications(payload?: any) {
  return notificationsStore.fetchNotifications(payload)
}

function markAsRead(id: string) {
  return notificationsStore.markAsRead(id)
}

function markAsUnread(id: string) {
  return notificationsStore.markAsUnread(id)
}

function markAllAsRead() {
  return notificationsStore.markAllAsRead()
}

function deleteNotification(id: string) {
  return notificationsStore.deleteNotification(id)
}

function setPageCursor(cursor: any) {
  notificationsStore.setPageCursor(cursor)
}

function toggleMuted() {
  notificationsStore.toggleMuted()
}

function onNotificationClick({ notificationID, readAt }: { notificationID: string; readAt?: string }) {
  if (!readAt) {
    markAsRead(notificationID)
  }
}

function onMarkAsRead({ notificationID }: { notificationID: string }) {
  markAsRead(notificationID)
}

function onMarkAsUnread({ notificationID }: { notificationID: string }) {
  markAsUnread(notificationID)
}

function onDeleteNotification({ notificationID }: { notificationID: string }) {
  deleteNotification(notificationID)
    .then(() => {
      toastSuccess($t('notificationDeleted'))
    })
    .catch(() => {
      toastError($t('notificationDeletedError'))
    })
}

function handleMarkAllAsRead() {
  markAllAsRead()
    .then(() => {
      toastSuccess($t('allMarkedAsRead'))
    })
    .catch(() => {
      toastError($t('markAllAsReadError'))
    })
}

function loadNotifications() {
  return fetchNotifications({ unreadOnly: activeTab.value === 0 })
}

function loadMore() {
  loadingMore.value = true
  return loadNotifications()
    .finally(() => {
      setTimeout(() => {
        loadingMore.value = false
      }, 300)
    })
}
</script>
