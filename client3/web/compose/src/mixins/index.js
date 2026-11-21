import { mixins } from 'corteza-lib/vue/dist'

import resourceTranslations from './resource-translations'
import uiHelpers from './uiHelpers'

export { useResourceTranslations } from './resource-translations'
export { useUiHelpers } from './uiHelpers'

export default {
  install(app) {
    app.mixin(mixins.toast)
    app.mixin(resourceTranslations)
    app.mixin(uiHelpers)
  },
}
