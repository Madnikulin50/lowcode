import { computed, getCurrentInstance } from 'vue'

export default {
  computed: {
    resourceTranslationLanguages() {
      const ll = this.$Settings.get('resourceTranslations.languages')
      if (!ll || !Array.isArray(ll) || ll.length === 0) {
        return ['en']
      }
      return ll
    },

    resourceTranslationsEnabled() {
      return this.resourceTranslationLanguages.length > 1
    },

    defaultTranslationLanguage() {
      return this.resourceTranslationLanguages[0]
    },

    currentLanguage() {
      return this.$i18n.i18next.language
    },
  },

  methods: {
    textDirectionality(language = this.currentLanguage) {
      switch (language) {
        case 'ar':
        case 'he':
        case 'pa':
        case 'fa':
        case 'ur':
        case 'sd':
          return 'rtl'
        default:
          return 'ltr'
      }
    },
  },
}

export function useResourceTranslations() {
  const instance = getCurrentInstance()

  const resourceTranslationLanguages = computed(() => {
    const ll = instance.proxy.$Settings.get('resourceTranslations.languages')
    if (!ll || !Array.isArray(ll) || ll.length === 0) {
      return ['en']
    }
    return ll
  })

  const resourceTranslationsEnabled = computed(() => resourceTranslationLanguages.value.length > 1)

  const defaultTranslationLanguage = computed(() => resourceTranslationLanguages.value[0])

  const currentLanguage = computed(() => {
  try {
    const i18nLocale = instance.proxy.$i18n?.locale
    if (typeof i18nLocale === 'string') return i18nLocale
    if (typeof i18nLocale?.value === 'string') return i18nLocale.value
  } catch {}
  return 'en'
})

  function textDirectionality(language) {
    if (language === undefined) {
      language = currentLanguage.value
    }
    switch (language) {
      case 'ar':
      case 'he':
      case 'pa':
      case 'fa':
      case 'ur':
      case 'sd':
        return 'rtl'
      default:
        return 'ltr'
    }
  }

  return {
    resourceTranslationLanguages,
    resourceTranslationsEnabled,
    defaultTranslationLanguage,
    currentLanguage,
    textDirectionality,
  }
}
