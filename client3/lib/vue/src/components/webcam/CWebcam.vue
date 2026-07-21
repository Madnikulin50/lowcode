<template>
  <div>
    <button
      v-b-tooltip.noninteractive.hover="{ title: labels.tooltip, boundary: 'body' }"
      class="btn btn-light"
      :class="buttonClass"
      @click.prevent="openWebcamModal"
    >
      <slot />
    </button>

    <div
      ref="webcamModalRef"
      class="modal fade"
      tabindex="-1"
      data-bs-backdrop="static"
      data-bs-keyboard="false"
    >
      <div class="modal-dialog modal-lg modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ labels.modalTitle }}</h5>
          </div>
          <div class="modal-body p-0">
            <div
              v-if="showErrorMessage"
              class="p-3 text-danger"
            >
              {{ labels.cameraErrorMessage }}
            </div>
            <div
              v-else
              class="embed-responsive embed-responsive-4by3 d-flex justify-content-center align-items-center"
            >
              <div
                v-if="processingWebcam"
                class="spinner-border text-primary"
                role="status"
              >
                <span class="visually-hidden">Loading...</span>
              </div>

              <video
                v-show="!processingWebcam && !hasCapturedImage"
                ref="videoRef"
                autoplay
                playsinline
              />

              <img
                v-if="hasCapturedImage"
                :src="capturedImage"
                alt="Captured image"
                class="embed-responsive-item"
              >
            </div>
          </div>
          <div class="modal-footer">
            <div class="d-flex align-items-center gap-2">
              <button
                class="btn btn-light"
                @click="handleCloseClick"
              >
                {{ labels.cancelButtonLabel }}
              </button>

              <button
                :disabled="processingWebcam"
                class="btn btn-primary"
                @click="handleCaptureClick"
              >
                {{ hasCapturedImage ? labels.confirmButtonLabel : labels.captureButtonLabel }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { Modal } from 'bootstrap'

const props = withDefaults(defineProps<{
  buttonClass?: string
  labels?: Record<string, string>
}>(), {
  buttonClass: 'd-flex align-items-center h-100',
  labels: () => ({}),
})

const emit = defineEmits<{
  (e: 'upload', file: File): void
}>()

const videoRef = ref<HTMLVideoElement | null>(null)
const webcamModalRef = ref<HTMLDivElement | null>(null)

let stream: MediaStream | null = null
const capturedImage = ref<string | null>(null)
const processingWebcam = ref(true)
const showErrorMessage = ref(false)

let webcamModalInstance: Modal | null = null

const hasCapturedImage = computed(() => !!capturedImage.value)

function openWebcamModal(): void {
  if (webcamModalRef.value) {
    webcamModalInstance = new Modal(webcamModalRef.value)
    webcamModalInstance.show()
  }
}

async function initializeWebcam(): Promise<void> {
  await nextTick()
  startWebcam()
}

function startWebcam(): void {
  showErrorMessage.value = false

  navigator.mediaDevices.getUserMedia({
    video: {
      facingMode: 'user',
    },
  })
    .then(s => {
      stream = s
      if (videoRef.value) {
        videoRef.value.srcObject = s
      }
    })
    .catch(err => {
      console.error('Error accessing the camera:', err)
      showErrorMessage.value = true
    })
    .finally(() => {
      processingWebcam.value = false
    })
}

function capturePhoto(): void {
  const video = videoRef.value
  if (!video) return
  const canvas = document.createElement('canvas')
  canvas.width = video.videoWidth
  canvas.height = video.videoHeight
  canvas.getContext('2d')!.drawImage(video, 0, 0, canvas.width, canvas.height)
  capturedImage.value = canvas.toDataURL('image/jpeg')
}

function uploadCapturedImage(): void {
  if (!capturedImage.value) return

  fetch(capturedImage.value).then(res => res.blob()).then(blob => {
    const imageSuffix = new Date().toISOString().replace(/[:.]/g, '-')
    const file = new File([blob], `webcam-image-${imageSuffix}.jpg`, { type: 'image/jpeg' })
    emit('upload', file)
  })

  webcamModalInstance?.hide()
}

function stopWebcam(): void {
  if (!stream) return
  stream.getTracks().forEach(track => track.stop())
}

function handleCaptureClick(): void {
  if (hasCapturedImage.value) {
    uploadCapturedImage()
  } else {
    capturePhoto()
  }
}

function handleCloseClick(): void {
  if (hasCapturedImage.value) {
    discardCapturedImage()
  } else {
    webcamModalInstance?.hide()
  }
}

function closeCamera(): void {
  stopWebcam()
  capturedImage.value = null
  processingWebcam.value = true
}

function discardCapturedImage(): void {
  capturedImage.value = null
  startWebcam()
}

onMounted(() => {
  const el = webcamModalRef.value
  if (el) {
    el.addEventListener('show.bs.modal', initializeWebcam)
    el.addEventListener('hidden.bs.modal', closeCamera)
  }
})

onBeforeUnmount(() => {
  const el = webcamModalRef.value
  if (el) {
    el.removeEventListener('show.bs.modal', initializeWebcam)
    el.removeEventListener('hidden.bs.modal', closeCamera)
  }
  webcamModalInstance?.dispose()
})
</script>
