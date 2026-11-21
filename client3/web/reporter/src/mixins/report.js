import { system } from 'corteza-lib/js/dist'

export function useReportHelpers() {
  function fetchReport(reportID, toastErrorHandler, t) {
    return window.__systemAPI.reportRead({ reportID })
      .then(report => {
        const r = new system.Report(report)
        r.initialReportState = r.clone()
        return r
      })
      .catch(toastErrorHandler(t('notification.report.fetchFailed')))
  }

  function handleSave(report, isNew, toastSuccess, toastErrorHandler, t, router) {
    const { blocks } = report
    const cleaned = {
      ...report,
      blocks: blocks.map(block => {
        block.elements = block.elements.map(element => {
          delete element.dataframes
          return element
        })
        return block
      }),
    }

    if (isNew) {
      return window.__systemAPI.reportCreate(cleaned)
        .then(r => {
          toastSuccess(t('notification.report.created'))
          router.push({ name: 'report.builder', params: { reportID: r.reportID } })
          return new system.Report(r)
        })
        .catch(toastErrorHandler(t('notification.report.createFailed')))
    } else {
      return window.__systemAPI.reportUpdate(cleaned)
        .then(r => {
          toastSuccess(t('notification.report.updated'))
          return new system.Report(r)
        })
        .catch(toastErrorHandler(t('notification.report.updateFailed')))
    }
  }

  function handleDelete(report, toastSuccess, toastErrorHandler, t, router) {
    return window.__systemAPI.reportDelete(report)
      .then(() => {
        report.deletedAt = new Date()
        toastSuccess(t('notification.report.delete'))
        router.push({ name: 'report.list' })
      })
      .catch(toastErrorHandler(t('notification.report.deleteFailed')))
  }

  function handleClone(report, toastSuccess, toastErrorHandler, t) {
    const { meta, sources, blocks, scenarios, labels } = report
    const cloneSuffix = `${meta.name} (${t('cloneSuffix')})`
    let name = meta.name

    if (!report.initialReportState || report.initialReportState.meta.name === name) {
      name = cloneSuffix
    }

    return window.__systemAPI.reportCreate({ handle: '', meta: { ...meta, name }, sources, blocks, scenarios, labels })
      .then(r => {
        toastSuccess(t('notification.report.created'))
        return new system.Report(r)
      })
      .catch(toastErrorHandler(t('notification.report.createFailed')))
  }

  return { fetchReport, handleSave, handleDelete, handleClone }
}
