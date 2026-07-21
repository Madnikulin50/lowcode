import { useUiStore } from '../store/ui'

export function useEditorHelpers() {
  const ui = useUiStore()

  function incLoader() {
    ui.incLoader()
  }

  function decLoader() {
    ui.decLoader()
  }

  function animateSuccess(key) {
    key.success = true
    setTimeout(() => {
      key.success = false
    }, 2000)
  }

  return {
    incLoader,
    decLoader,
    animateSuccess,
  }
}
