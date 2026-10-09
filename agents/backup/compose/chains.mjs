export function buildRuleChains ({ nsID, modules, agentUrl }) {
  const ns = String(nsID)
  const jobs = String(modules.jobs)
  const snapshots = String(modules.snapshots)
  const restores = String(modules.restores)
  const api = (agentUrl || '').replace(/\/$/, '')

  const component = (id, type, label, fields = {}) => ({
    id,
    type,
    label,
    config: {
      ...fields,
      ...(api ? { url: api } : {}),
    },
  })

  const createThen = (id, name, description, type, fields, createModule, createHandle, createFields) => ({
    id,
    name,
    description,
    entryNode: 'record',
    namespaceID: ns,
    nodes: [
      {
        id: 'record',
        type: 'crud',
        label: 'Create record',
        config: {
          operation: 'create',
          namespaceID: ns,
          moduleID: createModule,
          moduleHandle: createHandle,
          fields: createFields,
        },
      },
      component('run', type, 'Agent', fields),
    ],
    edges: [{ from: 'record', to: 'run' }],
  })

  return [
    createThen(
      'backup-run-source',
      'Backup: запуск с источника',
      'Создаёт job и вызывает backup/run (SDK POST /jobs).',
      'backup/run',
      { sourceID: '{{sourceID}}', source: '{{recordID}}', jobID: '{{createdRecordID}}' },
      jobs, 'jobs',
      { source: '{{recordID}}', status: 'running', progress: '0', kind: 'full' },
    ),
    createThen(
      'backup-run-policy',
      'Backup: запуск по политике',
      'Создаёт job и вызывает backup/run с policyID.',
      'backup/run',
      { policyID: '{{policyID}}', sourceID: '{{source}}', jobID: '{{createdRecordID}}' },
      jobs, 'jobs',
      { policy: '{{recordID}}', source: '{{source}}', status: 'running', progress: '0' },
    ),
    {
      id: 'backup-run-due',
      name: 'Backup: запустить due-политики',
      description:
        'Периодически запускайте (cron/systemd timer/Corteza Automation) с {"token":"..."} в теле. ' +
        'backup/due только читает у агента due-политики (cron подошёл, нет активного running-джоба) — сам агент ' +
        'ничего не пишет в Compose. Для каждой due-политики цепочка сама создаёт jobs-запись (как backup-run-policy) ' +
        'и запускает backup/run; дальше статус ведёт тот же поллер/ingest, что и у ручного запуска.',
      entryNode: 'due',
      namespaceID: ns,
      nodes: [
        component('due', 'backup/due', 'List due policies'),
        {
          id: 'loop',
          type: 'foreach',
          label: 'For each due policy',
          config: { items: 'due', itemVar: 'policy' },
        },
        {
          id: 'record',
          type: 'crud',
          label: 'Create job',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: jobs,
            moduleHandle: 'jobs',
            fields: { policy: '{{policy.id}}', source: '{{policy.sourceID}}', status: 'running', progress: '0' },
          },
        },
        component('run', 'backup/run', 'Start backup', {
          policyID: '{{policy.id}}',
          sourceID: '{{policy.sourceID}}',
          jobID: '{{createdRecordID}}',
        }),
      ],
      edges: [
        { from: 'due', to: 'loop' },
        { from: 'loop', to: 'record' },
        { from: 'loop', to: 'run' },
      ],
    },
    createThen(
      'backup-restore',
      'Backup: восстановить снапшот',
      'Создаёт restore и вызывает backup/restore.',
      'backup/restore',
      {
        snapshotID: '{{snapshotID}}',
        restoreID: '{{createdRecordID}}',
        destType: '{{destType}}',
        destPath: '{{destPath}}',
      },
      restores, 'restores',
      { snapshot: '{{recordID}}', dest_type: 'path', dest_path: '{{destPath}}', status: 'running', progress: '0' },
    ),
    {
      id: 'backup-prune',
      name: 'Backup: prune по retention',
      description: 'backup/prune. Из политики передаёт policyID.',
      entryNode: 'run',
      namespaceID: ns,
      nodes: [component('run', 'backup/prune', 'Prune', { policyID: '{{policyID}}' })],
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
      id: 'backup-ingest-restore',
      name: 'Backup: ingest статуса restore',
      description: 'Callback/poll → обновление restores.',
      entryNode: 'update_restore',
      namespaceID: ns,
      nodes: [{
        id: 'update_restore',
        type: 'crud',
        label: 'Update restore',
        config: {
          operation: 'update',
          namespaceID: ns,
          moduleID: restores,
          moduleHandle: 'restores',
          recordID: '{{createdRecordID}}',
          omitEmpty: true,
          continueOnError: true,
          fields: {
            status: '{{status}}',
            progress: '{{progress}}',
            error: '{{error}}',
            message: '{{message}}',
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
          query: "restorable = '1'",
          limit: 5,
        },
      }],
      edges: [],
    },
  ]
}
