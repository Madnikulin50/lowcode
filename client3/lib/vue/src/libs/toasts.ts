import { reactive } from 'vue'

export interface ToastItem {
  id: number
  payload: {
    title?: string
    notes?: string
    link?: unknown
  }
  options?: Record<string, unknown>
  actions?: Record<string, unknown>
}

let seq = 0

export const toasts: ToastItem[] = reactive([])

export function pushToast (message: string, { title, variant = 'success', ...rest }: Record<string, unknown> = {}) {
  const id = ++seq
  const t: ToastItem = {
    id,
    payload: { title: title as string | undefined, notes: message },
    options: { toastClass: `text-bg-${variant}`, ...rest },
    actions: {
      hide: { cb: () => removeToast(id) },
    },
  }
  toasts.push(t)
  return t
}

export function removeToast (id: number) {
  const i = toasts.findIndex(t => t.id === id)
  if (i > -1) {
    toasts.splice(i, 1)
  }
}

export function toastSuccess (message: string, title?: string) {
  pushToast(message, { title, variant: 'success' })
}

export function toastDanger (message: string, title?: string) {
  pushToast(message, { title, variant: 'danger' })
}

function getToastMessage (err: Error & { message?: string }): string {
  if (err.message && err.message.startsWith('notification')) {
    return err.message.substring('notification.'.length)
  }
  return err.message || ''
}

export function toastErrorHandler (opt: string | { prefix?: string; title?: string } = {}) {
  if (typeof opt === 'string') {
    opt = { title: opt }
  }

  const { prefix, title } = opt

  return (err: Error & { message?: string } = {} as Error) => {
    let toastMsg = ''
    let toastTitle = title

    err.message = getToastMessage(err)

    if (err.message) {
      toastMsg = err.message
    } else {
      toastMsg = title || ''
      toastTitle = ''
    }

    if (prefix) {
      toastMsg = `${prefix}: ${toastMsg}`
    }

    toastMsg = toastTitle ? toastMsg.charAt(0).toUpperCase() + toastMsg.slice(1) : toastMsg
    toastTitle = toastTitle ? toastTitle.charAt(0).toUpperCase() + toastTitle.slice(1) : toastTitle

    toastDanger(toastMsg, toastTitle)

    return err.message
  }
}

// Legacy window helpers (used by admin/privacy components)
if (typeof window !== 'undefined') {
  ;(window as unknown as Record<string, unknown>).__toastError = toastErrorHandler
  ;(window as unknown as Record<string, unknown>).__toastErrorHandler = toastErrorHandler
  ;(window as unknown as Record<string, unknown>).__toastSuccess = toastSuccess
}
