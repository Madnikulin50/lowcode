import { getCurrentInstance } from 'vue'
import { pushToast } from '../libs/toasts'

export function useToast() {
  const instance = getCurrentInstance()

  function toastSuccess(message: string, title?: string) {
    title = title || instance?.appContext.config.globalProperties.$t('success')
    toast(message, { title, variant: 'success' })
  }

  function toastWarning(message: string, title?: string) {
    title = title || instance?.appContext.config.globalProperties.$t('warning')
    toast(message, { title, variant: 'warning' })
  }

  function toastInfo(message: string, title?: string) {
    title = title || instance?.appContext.config.globalProperties.$t('info')
    toast(message, { title, variant: 'info' })
  }

  function toastDanger(message: string, title?: string) {
    title = title || instance?.appContext.config.globalProperties.$t('error')
    toast(message, { title, variant: 'danger' })
  }

  function toast(msg: string, opt: { title?: string; variant?: string } = { variant: 'success' }) {
    const toastFn = instance?.appContext.config.globalProperties.$toast
    if (toastFn) {
      toastFn(msg, opt)
    } else {
      pushToast(msg, opt)
    }
  }

  function getToastMessage(err: Error & { message?: string }): string {
    if (err.message && err.message.startsWith('notification')) {
      return instance?.appContext.config.globalProperties.$t(`${err.message.substring('notification.'.length)}`) || err.message
    }
    return err.message || ''
  }

  function toastErrorHandler(opt: string | { prefix?: string; title?: string } = {}) {
    if (typeof opt === 'string') {
      opt = { title: opt }
    }

    const { prefix, title } = opt as { prefix?: string; title?: string }

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

  return {
    toastSuccess,
    toastWarning,
    toastInfo,
    toastDanger,
    toast,
    getToastMessage,
    toastErrorHandler,
  }
}
