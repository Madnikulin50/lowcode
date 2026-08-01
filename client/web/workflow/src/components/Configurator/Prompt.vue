<script>
import Function from './Function'
import { components } from 'corteza-lib/vue/dist'

export default {
  extends: Function,

  methods: {
    async getFunctionTypes () {
      this.functions = components.promptDefinitions || []
    },
  },
}
</script>
