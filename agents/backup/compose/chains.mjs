export function buildRuleChains ({ nsID, modules, agentUrl }) {
  const ns = String(nsID)
  const jobs = String(modules.jobs)
  const snapshots = String(modules.snapshots)
  const restores = String(modules.restores)
  const api = agentUrl.replace(/\/$/, '')

  const httpJob = (id, name, description, url, body, createModule, createHandle, fields) => ({
    id,
    name,
    description,
    entryNode: createModule ? 'record' : 'http',
    namespaceID: ns,
    nodes: [
      ...(createModule ? [{
        id: 'record',
        type: 'crud',
        label: 'Create record',
        config: {
          operation: 'create',
          namespaceID: ns,
          moduleID: createModule,
          moduleHandle: createHandle,
          fields,
        },
      }] : []),
      {
        id: 'http',
        type: 'http',
        label: 'Agent',
        config: {
          url,
          method: 'POST',
          body,
          timeout: 30,
        },
      },
      {
        id: 'detach_poll',
        type: 'detach',
        label: 'Poll agent',
        config: {
          kind: 'poll',
          ingestChainID: 'backup-ingest-job',
          statusUrl: '{{agentUrl}}/jobs/{{scanID}}',
          interval: 2,
          timeout: 3600,
          until: 'completed,failed,done,error',
        },
      },
    ],
    edges: createModule
      ? [{ from: 'record', to: 'http' }, { from: 'http', to: 'detach_poll' }]
      : [{ from: 'http', to: 'detach_poll' }],
  })

  return [
    httpJob(
      'backup-run-source',
      'Backup: запуск с источника',
      'Создаёт job и POST /jobs с sourceID.',
      api + '/jobs',
      JSON.stringify({
        sourceID: '{{recordID}}',
        jobID: '{{createdRecordID}}',
        namespaceID: '{{namespaceID}}',
        token: '{{authToken}}',
        callbackUrl: '{{callbackUrl}}',
      }),
      jobs, 'jobs',
      { source: '{{recordID}}', status: 'running', progress: '0', kind: 'full' },
    ),
    httpJob(
      'backup-run-policy',
      'Backup: запуск по политике',
      'Создаёт job и POST /jobs с policyID.',
      api + '/jobs',
      JSON.stringify({
        policyID: '{{recordID}}',
        sourceID: '{{source}}',
        jobID: '{{createdRecordID}}',
        namespaceID: '{{namespaceID}}',
        token: '{{authToken}}',
        callbackUrl: '{{callbackUrl}}',
      }),
      jobs, 'jobs',
      { policy: '{{recordID}}', source: '{{source}}', status: 'running', progress: '0' },
    ),
    {
      id: 'backup-run-due',
      name: 'Backup: запустить due-политики',
      description: 'POST /jobs/due — агент сам выбирает политики по cron.',
      entryNode: 'http',
      namespaceID: ns,
      nodes: [{
        id: 'http',
        type: 'http',
        label: 'Run due',
        config: {
          url: api + '/jobs/due',
          method: 'POST',
          body: '{"namespaceID":"{{namespaceID}}","token":"{{authToken}}"}',
          timeout: 60,
        },
      }],
      edges: [],
    },
    httpJob(
      'backup-restore',
      'Backup: восстановить снапшот',
      'Создаёт restore и POST /restore.',
      api + '/restore',
      JSON.stringify({
        snapshotID: '{{recordID}}',
        restoreID: '{{createdRecordID}}',
        destType: '{{destType}}',
        destPath: '{{destPath}}',
        namespaceID: '{{namespaceID}}',
        token: '{{authToken}}',
        callbackUrl: '{{callbackUrl}}',
      }),
      restores, 'restores',
      { snapshot: '{{recordID}}', dest_type: 'path', dest_path: '{{destPath}}', status: 'running', progress: '0' },
    ),
    {
      id: 'backup-prune',
      name: 'Backup: prune по retention',
      description: 'POST /prune. Из политики передаёт policyID.',
      entryNode: 'http',
      namespaceID: ns,
      nodes: [{
        id: 'http',
        type: 'http',
        label: 'Prune',
        config: {
          url: api + '/prune',
          method: 'POST',
          body: '{"policyID":"{{recordID}}","namespaceID":"{{namespaceID}}","token":"{{authToken}}"}',
          timeout: 60,
        },
      }],
      edges: [],
    },
    {
      id: 'backup-ingest-job',
      name: 'Backup: ingest статуса джоба',
      description: 'Callback/poll → обновление jobs.',
      entryNode: 'update_job',
      namespaceID: ns,
      nodes: [{
        id: 'update_job',
        type: 'crud',
        label: 'Update job',
        config: {
          operation: 'update',
          namespaceID: ns,
          moduleID: jobs,
          moduleHandle: 'jobs',
          recordID: '{{createdRecordID}}',
          omitEmpty: true,
          continueOnError: true,
          fields: {
            status: '{{status}}',
            progress: '{{progress}}',
            bytes_read: '{{bytesRead}}',
            bytes_written: '{{bytesWritten}}',
            files_count: '{{files}}',
            error: '{{error}}',
            message: '{{message}}',
            engine: '{{engine}}',
          },
        },
      }],
      edges: [],
    },
    {
      id: 'backup-failed-alert',
      name: 'Backup: письмо при ошибке',
      description: 'Mail если job status = failed.',
      entryNode: 'check',
      namespaceID: ns,
      nodes: [
        { id: 'check', type: 'condition', label: 'Failed?', config: { field: 'status', operator: 'eq', value: 'failed' } },
        {
          id: 'mail',
          type: 'mail',
          label: 'Notify',
          config: {
            to: '{{to}}',
            subject: '[Backup] failed {{source}}',
            body: '<p>Job failed.</p><p>{{error}}</p><p>{{message}}</p>',
            contentType: 'html',
          },
        },
      ],
      edges: [{ from: 'check', to: 'mail', condition: 'check_result' }],
    },
    {
      id: 'backup-create-snapshot-meta',
      name: 'Backup: заглушка (снапшот пишет агент)',
      description: 'Снапшоты создаёт агент. Цепочка оставлена для документации.',
      entryNode: 'noop',
      namespaceID: ns,
      nodes: [{
        id: 'noop',
        type: 'crud',
        label: 'Search snapshots',
        config: {
          operation: 'search',
          namespaceID: ns,
          moduleID: snapshots,
          moduleHandle: 'snapshots',
          query: 'restorable = \'1\'',
          limit: 5,
        },
      }],
      edges: [],
    },
  ]
}
