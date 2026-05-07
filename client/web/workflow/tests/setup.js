/* global jest */

jest.mock('corteza-lib/js/dist', () => ({}), { virtual: true })
jest.mock('../../../../lib/vue/dist', () => ({
  components: {
    CToaster: jest.fn(),
    CPrompts: {
      name: 'c-prompts',
      render: () => {},
    },
    CPermissionsModal: {
      name: 'c-permissions-modal',
      render: () => {},
    },
    CTopbar: {
      name: 'c-topbar',
      render: () => {},
    },
    CSidebar: {
      name: 'c-sidebar',
      render: () => {},
    },
  },
  mixins: {
    corredor: {
      methods: {
        triggerScript: jest.fn(),
      },
    },
  },
}), { virtual: true })
