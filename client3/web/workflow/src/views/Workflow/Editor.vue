<template>
  <workflow-editor
    v-if="!processing"
    id="workflow-editor"
    :workflow-object="workflow"
    :workflow-triggers="triggers"
    :change-detected="changeDetected"
    :can-create="canCreate"
    :processing-save="processingSave"
    :processing-delete="processingDelete"
    class="overflow-hidden"
    @save="saveWorkflow"
    @delete="deleteWorkflow"
    @undelete="undeleteWorkflow"
  />
</template>

<script setup>
import { ref, computed, inject, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter, useRoute, onBeforeRouteLeave } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useToast, useAuth, useRBACStore } from 'corteza-lib/vue/dist'
import { automation } from 'corteza-lib/js/dist'
import { throttle } from 'lodash'
import WorkflowEditor from '../../components/WorkflowEditor.vue'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const toast = useToast()
const { auth } = useAuth()

const $ComposeAPI = inject('composeAPI', {})
const $AutomationAPI = inject('automationAPI', {})

const processing = ref(true)
const processingSave = ref(false)
const processingDelete = ref(false)
const workflow = ref({})
const triggers = ref([])
const changeDetected = ref(false)

const _rbac = useRBACStore()
const can = (resource, action) => _rbac.can(resource, action)

const canCreate = computed(() => can('automation/', 'workflow.create'))

const workflowID = computed(() => {
  return route.params.workflowID || (workflow.value.workflowID !== '0' ? workflow.value.workflowID : undefined)
})

const userID = computed(() => {
  if (auth.user) {
    return auth.user.userID
  }
  return undefined
})

onBeforeRouteLeave((to, from, next) => {
  if (changeDetected.value && !workflow.value.deletedAt) {
    next(window.confirm(t('notification.confirm-unsaved-changes')))
  } else {
    window.onbeforeunload = null
    next()
  }
})

onMounted(async () => {
  window.onbeforeunload = null

  window.addEventListener('change-detected', () => {
    if (!changeDetected.value) {
      window.onbeforeunload = () => {
        return true
      }
    }
    changeDetected.value = true
  })

  if (workflowID.value) {
    await fetchTriggers()
    await fetchWorkflow()
  } else {
    workflow.value = new automation.Workflow({
      ownedBy: userID.value,
      runAs: '0',
      enabled: true,
      handle: '',
    })
  }

  processing.value = false
})

watch(workflowID, async (id) => {
  if (id) {
    processing.value = true
    await fetchTriggers()
    await fetchWorkflow()
    processing.value = false
  }
})

onBeforeUnmount(() => {
  window.onbeforeunload = null
})

async function fetchWorkflow() {
  return $AutomationAPI.workflowRead({ workflowID: workflowID.value })
    .then(wf => {
      workflow.value = new automation.Workflow(wf)
    })
    .catch((e) => toast.error(t('notification.failed-fetch-workflow'), e))
}

async function fetchTriggers(wfID = workflowID.value) {
  return $AutomationAPI.triggerList({ workflowID: wfID, disabled: 1 })
    .then(({ set = [] }) => {
      triggers.value = set
    })
    .catch((e) => toast.error(t('notification.failed-fetch-triggers'), e))
}

const saveWorkflow = throttle(async function (wf) {
  try {
    processingSave.value = true

    const isNew = wf.workflowID === '0'

    const { triggers: wfTriggers = [] } = wf

    await Promise.all(triggers.value.filter(({ triggerID }) => {
      return !wfTriggers.find(t => triggerID === t.triggerID)
    }).map(({ triggerID }) => {
      return $AutomationAPI.triggerDelete({ triggerID })
    }),
    ).then(async () => {
      await Promise.all(wfTriggers.map(t => {
        if (t.triggerID) {
          return $AutomationAPI.triggerUpdate({
            ...t,
            workflowStepID: t.stepID,
          })
        } else {
          return $AutomationAPI.triggerCreate({
            ...t,
            workflowID: wf.workflowID,
            workflowStepID: t.stepID,
            ownedBy: userID.value,
          })
        }
      })).catch(() => {
        throw new Error(t('notification.configure-triggers'))
      })
    })

    if (isNew) {
      wf = await $AutomationAPI.workflowCreate(wf)
    } else {
      wf = await $AutomationAPI.workflowUpdate(wf)
    }

    await fetchTriggers(wf.workflowID)

    changeDetected.value = false
    window.onbeforeunload = null

    workflow.value = new automation.Workflow(wf)
    toast.success(t('notification.update.success'))

    if (isNew) {
      router.push({ name: 'workflow.edit', params: { workflowID: workflow.value.workflowID } })
    }
  } catch (e) {
    toast.error(t('notification.failed-save'), e)
  }

  processingSave.value = false
}, 500)

function deleteWorkflow() {
  if (workflow.value.workflowID) {
    processingDelete.value = true

    $AutomationAPI.workflowDelete(workflow.value)
      .then(() => {
        workflow.value = {}
        workflow.value.deletedAt = new Date()
        router.push({ name: 'workflow.list' })
        toast.success(t('notification.delete.success'))
      })
      .catch((e) => toast.error(t('notification.delete.failed'), e))
      .finally(() => {
        processingDelete.value = false
      })
  }
}

function undeleteWorkflow() {
  if (workflow.value.workflowID) {
    processingDelete.value = true

    $AutomationAPI.workflowUndelete(workflow.value)
      .then(() => {
        workflow.value.deletedAt = undefined
        workflow.value.deletedBy = undefined
        toast.success(t('notification.undelete.success'))
      })
      .catch((e) => toast.error(t('notification.undelete.failed'), e))
      .finally(() => {
        processingDelete.value = false
      })
  }
}
</script>

<style lang="scss">
#workflow-editor {
  tr.b-table-details > td {
    padding-top: 0;
  }

  .arrow-up {
    width: 0;
    height: 0;
    margin: 0 auto;
    border-left: 10px solid transparent;
    border-right: 10px solid transparent;
    border-bottom: 10px solid var(--light);
  }
}
</style>
