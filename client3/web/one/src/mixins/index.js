import Vue from 'vue'

import resourceTranslations from './resource-translations'

import { mixins } from 'corteza-lib/vue/dist'

Vue.mixin(mixins.toast)
Vue.mixin(resourceTranslations)
