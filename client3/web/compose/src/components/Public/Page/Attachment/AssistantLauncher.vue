<template>
  <div
    v-if="current"
    class="assistant-launcher"
  >
    <div
      v-if="state.open"
      class="assistant-drawer shadow"
    >
      <div class="d-flex align-items-center justify-content-between px-2 py-1 border-bottom">
        <small class="text-truncate">
          ✦ Ассистент · {{ current.name || 'документ' }}
        </small>
        <div class="d-flex gap-2">
          <a
            :href="src"
            target="_blank"
            rel="noopener"
            title="Открыть в отдельной вкладке"
          >↗</a>
          <a
            href="#"
            title="Закрыть"
            @click.prevent="closeAssistant"
          >✕</a>
        </div>
      </div>
      <iframe
        :key="src"
        class="assistant-frame"
        :src="src"
        title="Ассистент по документу"
        sandbox="allow-scripts allow-same-origin allow-forms allow-popups allow-downloads"
      />
    </div>

    <button
      v-else
      type="button"
      class="btn btn-primary btn-sm shadow assistant-fab"
      title="Спросить ИИ-ассистента об открытом документе"
      @click="state.open = true"
    >
      ✦ Ассистент
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useAssistantContext, assistantURL, closeAssistant } from 'corteza-webapp-compose/src/composables/useAssistantContext'

const { state, current } = useAssistantContext()
const src = computed(() => (current.value ? assistantURL(current.value) : ''))
</script>

<style scoped lang="scss">
.assistant-launcher {
  position: fixed;
  right: 1rem;
  bottom: 1rem;
  z-index: 3000; // above the attachment lightbox
}

.assistant-drawer {
  width: min(440px, 92vw);
  height: min(72vh, 720px);
  display: flex;
  flex-direction: column;
  background: var(--bs-body-bg, #fff);
  border: 1px solid var(--bs-border-color, #dee2e6);
  border-radius: 0.5rem;
  overflow: hidden;
}

.assistant-frame {
  flex: 1;
  width: 100%;
  border: 0;
}
</style>
