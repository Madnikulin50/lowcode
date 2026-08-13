<template>
  <div>
    <div
      :title="popoverContent"
      :class="{ 'disabled': disabled, 'small': small }"
      class="h3 mx-auto rounded-circle circle border-primary text-primary text-center mb-5"
    >
      <font-awesome-icon
        v-if="done"
        :icon="['fas', 'check']"
      />
      <span v-else>{{ stepNumber }}</span>
    </div>
    <slot />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'onboarding' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  done: {
    type: Boolean,
    required: true,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  optional: {
    type: Boolean,
    default: false,
  },
  stepNumber: {
    type: String,
    default: '?',
  },
  small: {
    type: Boolean,
    default: false,
  },
})

const popoverContent = computed(() => {
  if (props.optional) {
    return t('step.optional')
  }
  return ''
})
</script>

<style lang="scss" scoped>
.circle {
  width: 75px;
  height: 75px;
  font-size: 35px;
  line-height: 75px;
  border: 2px solid;
}

.small {
  width: 50px;
  height: 50px;
  font-size: 25px;
  line-height: 45px;
}

.disabled {
  opacity: 0.65;
}
</style>
