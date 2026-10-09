import Vue from 'vue'
import { filters } from 'corteza-lib/vue/dist'

for (const n in filters) {
  Vue.filter(n, filters[n])
}
