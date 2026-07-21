<template>
  <div
    class="d-flex flex-column"
    @click="handleRecordNavigation"
  >
    <h5 class="fw-bold text-break">
      {{ title }}
    </h5>

    <div class="text-secondary mb-1 text-break">
      {{ description }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from '../../../composables/useToast'

const props = defineProps<{
  notification: Record<string, any>
}>()

const { $t, $ComposeAPI } = getCurrentInstance()!.appContext.config.globalProperties as any
const route = (() => { try { return useRoute() } catch(e) { return {} } })() || {}
const router = (() => { try { return useRouter() } catch(e) { return {} } })() || {}
const { toastDanger, toastErrorHandler } = useToast()

const title = computed(() => {
  if (!props.notification || !props.notification.config) {
    return ''
  }

  return props.notification.config.title
})

const description = computed(() => {
  if (!props.notification || !props.notification.config) {
    return ''
  }

  return props.notification.config.description || ''
})

const isOnPagesRouteOrChild = computed(() => {
  return route && (['pages', 'page', 'page.record', 'page.record.edit', 'page.record.create'].includes(route.name as string))
})

async function handleRecordNavigation() {
  const { namespaceID, recordID, moduleID, openMode, edit } = props.notification.config

  try {
    const namespace = await $ComposeAPI.namespaceRead({ namespaceID })

    if (!namespace) {
      toastDanger($t('namespaceNotFound'))
      return
    }

    const slug = namespace.slug || namespace.namespaceID

    const recordPages = await $ComposeAPI.pageList({ moduleID, namespaceID }).then(({ set = [] }: { set: any[] }) => set)

    if (!recordPages || recordPages.length === 0) {
      toastDanger($t('pageNotFound'))
      return
    }

    const record = await $ComposeAPI.recordRead({ recordID, moduleID, namespaceID })

    if (!record) {
      toastDanger($t('recordNotFound'))
      return
    }

    const { pageID } = recordPages[0]

    const instance = getCurrentInstance()
    const appName = instance?.appContext.app?.options?.name
    if (appName !== 'compose') {
      const u = new URL(window.location)
      const url = `${u.origin}/compose/ns/${slug}/pages/${pageID}/record/${recordID}/${edit ? 'edit' : ''}`

      if (openMode === 'newTab') {
        window.open(url, '_blank')
      } else {
        window.location.href = url
      }

      return
    }

    let routeName = 'page.record'

    if (!recordID || recordID === '0') {
      routeName += '.create'
    } else if (edit) {
      routeName += '.edit'
    }

    const routeParams = {
      name: routeName,
      params: {
        recordID,
        pageID,
        slug,
        edit,
      },
    }

    if (openMode === 'newTab') {
      window.open(router.resolve(routeParams).href, '_blank')
    } else if (isOnPagesRouteOrChild.value && openMode === 'modal' && slug === route.params.slug) {
      window.dispatchEvent(new CustomEvent('show-record-modal', {
        detail: {
          recordID: !recordID || recordID === '0' ? '0' : recordID,
          recordPageID: pageID,
          edit,
        },
      }))

      return
    } else {
      router.push(routeParams)
    }
  } catch (error) {
    toastErrorHandler($t('recordRedirectError'))(error as Error)
  }
}
</script>
