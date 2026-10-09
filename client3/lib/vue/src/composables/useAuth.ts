import { getCurrentInstance } from 'vue'

export function useAuth() {
  const instance = getCurrentInstance()
  const app = instance?.appContext?.app
  const auth = app?.config?.globalProperties?.$auth
  return { auth }
}