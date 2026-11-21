<template>
  <div
    ref="modalRef"
    class="modal fade"
    tabindex="-1"
    data-bs-backdrop="static"
    data-bs-keyboard="false"
  >
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-body d-flex flex-column justify-content-center align-items-center gap-2 p-4">
          <h5>{{ labels.warning(countdownTime) }}</h5>

          <button
            class="btn btn-primary btn-lg"
            @click="extendSession"
          >
            {{ labels.extend }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import { throttle } from 'lodash'
import { Modal } from 'bootstrap'

interface ExtendLabels {
  extend: string
  warning: (countdownTime: number) => string
}

const props = withDefaults(defineProps<{
  timeout?: number
  labels?: ExtendLabels
}>(), {
  timeout: 60,
  labels: (): ExtendLabels => ({
    extend: 'Extend Session',
    warning: (countdownTime: number) => `You will be logged out in ${countdownTime} seconds`,
  }),
})

const countdownTime = ref<number | null>(null)
let countdownTimer: ReturnType<typeof setInterval> | null = null
let lastActivityTime = 0
let heartbeatInterval: ReturnType<typeof setInterval> | null = null

const modalRef = ref<HTMLDivElement | null>(null)
let modalInstance: Modal | null = null

const { proxy } = getCurrentInstance()!
const $auth = proxy!.$auth

function updateActivity(): void {
  lastActivityTime = Date.now()
}

const debouncedUpdateActivity = throttle(updateActivity, 5000, { leading: true, trailing: true })

function setupActivityListeners(): void {
  window.addEventListener('mousemove', debouncedUpdateActivity, { passive: true })
  window.addEventListener('keypress', debouncedUpdateActivity, { passive: true })
  window.addEventListener('click', debouncedUpdateActivity, { passive: true })
  window.addEventListener('scroll', debouncedUpdateActivity, { passive: true })
  window.addEventListener('touchstart', debouncedUpdateActivity, { passive: true })
  window.addEventListener('touchmove', debouncedUpdateActivity, { passive: true })
  window.addEventListener('touchend', debouncedUpdateActivity, { passive: true })
}

function removeActivityListeners(): void {
  window.removeEventListener('mousemove', debouncedUpdateActivity)
  window.removeEventListener('keypress', debouncedUpdateActivity)
  window.removeEventListener('click', debouncedUpdateActivity)
  window.removeEventListener('scroll', debouncedUpdateActivity)
  window.removeEventListener('touchstart', debouncedUpdateActivity)
  window.removeEventListener('touchmove', debouncedUpdateActivity)
  window.removeEventListener('touchend', debouncedUpdateActivity)
}

function setupHeartbeatMonitoring(): void {
  if (props.timeout === 0) return
  lastActivityTime = Date.now()
  heartbeatInterval = setInterval(() => {
    const secondsSinceLastActivity = Math.floor((Date.now() - lastActivityTime) / 1000)
    if (secondsSinceLastActivity > props.timeout) {
      promptExtendSession()
    }
  }, 5000)
}

function startCountdown(expiresIn: number): void {
  countdownTime.value = expiresIn
  countdownTimer = setInterval(() => {
    countdownTime.value!--
    if (countdownTime.value! <= 0) {
      clearInterval(countdownTimer!)
      $auth.logout()
    }
  }, 1000)
}

function stopCountdown(): void {
  if (!countdownTimer) return
  clearInterval(countdownTimer)
  countdownTimer = null
}

function extendSession(): void {
  if (modalInstance) {
    modalInstance.hide()
  }
  $auth.stopAutoLogout().then(() => {
    setupHeartbeatMonitoring()
  }).catch((err: Error) => {
    console.error('Failed to stop auto logout', err)
    $auth.logout()
  })
}

function promptExtendSession(): void {
  if (heartbeatInterval) {
    clearInterval(heartbeatInterval)
    heartbeatInterval = null
  }
  $auth.startAutoLogout().then((expiresIn: number) => {
    startCountdown(expiresIn)
    if (modalRef.value) {
      modalInstance = new Modal(modalRef.value)
      modalInstance.show()
    }
  }).catch((err: Error) => {
    console.error('Failed to start auto logout', err)
    $auth.logout()
  })
}

onMounted(() => {
  debouncedUpdateActivity()
  setupActivityListeners()
  setupHeartbeatMonitoring()
})

onBeforeUnmount(() => {
  if (heartbeatInterval) clearInterval(heartbeatInterval)
  stopCountdown()
  removeActivityListeners()
  debouncedUpdateActivity.cancel()
  if (modalInstance) {
    modalInstance.dispose()
    modalInstance = null
  }
})
</script>
