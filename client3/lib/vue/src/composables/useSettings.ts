import { getCurrentInstance } from 'vue'

export function useSettings () {
  const instance = getCurrentInstance()
  const app = instance?.appContext?.app
  const $Settings = app?.config?.globalProperties?.$Settings
  return { $Settings }
}
