import Vue from 'vue'

import resourceTranslations from './resource-translations'

import { mixins } from '../../../../lib/vue/dist'

Vue.mixin(mixins.toast)
Vue.mixin(resourceTranslations)
