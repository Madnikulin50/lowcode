<template>
  <b-modal
    id="page-block-modal"
    v-model="showModal"
    scrollable
    body-class="p-0"
    :content-class="contentClass"
    :dialog-class="dialogClass"
    hide-header
    hide-footer
    size="xl"
    no-fade
    @hidden="onHidden"
  >
    <chat
      v-if="showModal"
      :start-prompt="startPrompt"
      :files="attachedFiles"
      :page="page"
      magnified
      v-bind="$props"
      v-on="$listeners"
    />
  </b-modal>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'
import Chat from './Chat'

export default {
  i18nOptions: {
    namespaces: 'chat',
  },

  name: 'ChatModal',

  components: {
    Chat,
  },

  props: {
    page: {
      type: String,
      required: false,
      default: '',
    },
    module: {
      type: String,
      required: false,
      default: '',
    },
    namespace: {
      type: String,
      required: false,
      default: '',
    },
  },

  data () {
    return {
      showModal: false,
      startPrompt: '',
      attachedFiles: [],
      record: undefined,
      // Used if you want to display a specific block in the modal
      // Otherwise its retrieved based on the page and blockID
      customBlock: undefined,
    }
  },

  computed: {
    ...mapGetters({
      getPageByID: 'page/getByID',
      getModuleByID: 'module/getByID',
    }),

    dialogClass () {
      return this.block && this.block.options.magnifyOption === 'fullscreen' ? 'h-100 mw-100 m-0 mh-100' : 'h-100 modal-max-width'
    },

    contentClass () {
      return `${this.block && this.block.options.magnifyOption === 'fullscreen' ? 'mh-100 rounded-0' : ''} position-initial`
    },
  },

  mounted () {
    this.$root.$on('show-chat-modal', this.startChatModal)
  },

  beforeDestroy () {
    this.destroyEvents()
    this.setDefaultValues()
  },

  methods: {
    startChatModal (data) {
      const { prompt = '', files = [] } = data
      this.startPrompt = prompt
      this.attachedFiles = files
      this.showModal = true
    },

    onHidden () {
      this.showModal = false
    },
    destroyEvents () {
      this.$root.$off('show-chat-modal', this.startChatModal)
    },

    setDefaultValues () {
      this.showModal = false
      this.block = undefined
      this.record = undefined
      this.page = undefined
      this.customBlock = undefined
    },
  },
}

</script>

<style lang="scss">
.position-initial {
  position: initial;
}

.modal-max-width {
  max-width: 80vw;
}
</style>
