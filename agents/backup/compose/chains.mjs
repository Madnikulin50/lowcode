export function buildRuleChains ({ nsID, modules, agentUrl }) {
  const ns = String(nsID)
  const jobs = String(modules.jobs)
  const snapshots = String(modules.snapshots)
  const restores = String(modules.restores)
  const verifyReports = String(modules.verify_reports)
  const sources = String(modules.sources)
  const policies = String(modules.policies)
  const requests = String(modules.backup_requests)
  const acts = String(modules.destruction_acts)
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
    {
      id: 'backup-dest-test',
      name: 'Backup: проверить назначение',
      description:
        'Со страницы назначения: backup/dest-test — endpoint/учётки/бакет, ' +
        'результат (last_check*) пишет сам агент.',
      entryNode: 'run',
      namespaceID: ns,
      nodes: [component('run', 'backup/dest-test', 'Проверить', { destinationID: '{{destinationID}}' })],
      edges: [],
    },
    {
      id: 'backup-restore',
      name: 'Backup: восстановить снапшот',
      description:
        'Создаёт restore и вызывает backup/restore. Восстановление в original — деструктивное: ' +
        'при заданном securityEmail уходит уведомление безопасника. Restore запускается в любом ' +
        'случае: notify идёт отдельной веткой после запуска и не может его заблокировать.',
      entryNode: 'record',
      namespaceID: ns,
      nodes: [
        {
          id: 'record',
          type: 'crud',
          label: 'Create restore',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: restores,
            moduleHandle: 'restores',
            fields: { snapshot: '{{recordID}}', dest_type: '{{destType}}', dest_path: '{{destPath}}', status: 'running', progress: '0' },
          },
        },
        { id: 'check', type: 'condition', label: 'Restore to original?', config: { field: 'destType', operator: 'eq', value: 'original' } },
        { id: 'email', type: 'condition', label: 'securityEmail задан?', config: { field: 'securityEmail', operator: 'notEmpty' } },
        {
          id: 'notify',
          type: 'mail',
          label: 'Notify security',
          config: {
            to: '{{securityEmail}}',
            subject: '[Backup] деструктивное восстановление snapshot {{snapshotID}}',
            body: '<p>Запрошено восстановление в исходное место.</p><p>snapshotID={{snapshotID}} destPath={{destPath}}</p>',
            contentType: 'html',
          },
        },
        component('run', 'backup/restore', 'Restore', {
          snapshotID: '{{snapshotID}}',
          restoreID: '{{createdRecordID}}',
          destType: '{{destType}}',
          destPath: '{{destPath}}',
        }),
      ],
      edges: [
        { from: 'record', to: 'run' },
        { from: 'record', to: 'check' },
        { from: 'check', to: 'email', condition: 'check_result' },
        { from: 'email', to: 'notify', condition: 'email_result' },
      ],
    },
    {
      id: 'backup-restore-run',
      name: 'Backup: запустить восстановление из записи',
      description:
        'Со страницы записи restores: берёт snapshot/dest_type/dest_path из самой записи ' +
        '(значения записи приходят в контекст триггера) и вызывает backup/restore.',
      entryNode: 'mark',
      namespaceID: ns,
      nodes: [
        {
          id: 'mark',
          type: 'crud',
          label: 'Mark running',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: restores,
            moduleHandle: 'restores',
            recordID: '{{recordID}}',
            omitEmpty: true,
            fields: { status: 'running' },
          },
        },
        component('run', 'backup/restore', 'Restore', {
          snapshotID: '{{snapshot}}',
          restoreID: '{{recordID}}',
          destType: '{{dest_type}}',
          destPath: '{{dest_path}}',
        }),
      ],
      edges: [{ from: 'mark', to: 'run' }],
    },
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
    {
      id: 'backup-vault',
      name: 'Backup: в сейф (скан + WORM-копия)',
      description:
        'Со страницы снапшота: сканирование и копия в vault. Агент сам обновляет ' +
        'scan_status/scanned_at/vault_key/vault_locked_until у снапшота.',
      entryNode: 'mark',
      namespaceID: ns,
      nodes: [
        {
          id: 'mark',
          type: 'crud',
          label: 'Mark queued',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: snapshots,
            moduleHandle: 'snapshots',
            recordID: '{{recordID}}',
            fields: { scan_status: 'queued' },
          },
        },
        component('run', 'backup/vault', 'Скан + сейф', { snapshotID: '{{recordID}}' }),
      ],
      edges: [{ from: 'mark', to: 'run' }],
    },
    {
      id: 'backup-vault-due',
      name: 'Backup: отправить новые снапшоты в сейф',
      description:
        'Периодически запускайте по расписанию. Ищет restorable-снапшоты без vault_key ' +
        'и для каждого вызывает backup/vault.',
      entryNode: 'search',
      namespaceID: ns,
      nodes: [
        {
          id: 'search',
          type: 'crud',
          label: 'Find unvaulted',
          config: {
            operation: 'search',
            namespaceID: ns,
            moduleID: snapshots,
            moduleHandle: 'snapshots',
            query: "restorable = '1' AND vault_key = ''",
            limit: 50,
          },
        },
        { id: 'loop', type: 'foreach', label: 'For each snapshot', config: { items: 'set', itemVar: 'snap' } },
        {
          id: 'mark',
          type: 'crud',
          label: 'Mark queued',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: snapshots,
            moduleHandle: 'snapshots',
            recordID: '{{snap.recordID}}',
            fields: { scan_status: 'queued' },
          },
        },
        component('run', 'backup/vault', 'Скан + сейф', { snapshotID: '{{snap.recordID}}' }),
      ],
      edges: [
        { from: 'search', to: 'loop' },
        { from: 'loop', to: 'mark' },
        { from: 'loop', to: 'run' },
      ],
    },
    {
      id: 'backup-verify',
      name: 'Backup: верификация восстановления',
      description:
        'Со страницы политики: создаёт verify_reports и вызывает backup/verify. ' +
        'Агент сам пишет verified_at/verify_duration_sec в снапшот.',
      entryNode: 'report',
      namespaceID: ns,
      nodes: [
        {
          id: 'report',
          type: 'crud',
          label: 'Create report',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: verifyReports,
            moduleHandle: 'verify_reports',
            fields: { policy: '{{recordID}}', status: 'running', level: '{{verifyLevel}}' },
          },
        },
        component('run', 'backup/verify', 'Verify', {
          policyID: '{{recordID}}',
          verifyLevel: '{{verifyLevel}}',
          jobID: '{{createdRecordID}}',
        }),
      ],
      edges: [{ from: 'report', to: 'run' }],
    },
    {
      id: 'backup-verify-due',
      name: 'Backup: верификация по расписанию',
      description:
        'Периодически запускайте (cron). backup/verify-due возвращает политики, у которых ' +
        'подошёл verify_cron; для каждой создаётся отчёт и запускается backup/verify.',
      entryNode: 'due',
      namespaceID: ns,
      nodes: [
        component('due', 'backup/verify-due', 'List verify-due'),
        { id: 'loop', type: 'foreach', label: 'For each', config: { items: 'due', itemVar: 'v' } },
        {
          id: 'report',
          type: 'crud',
          label: 'Create report',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: verifyReports,
            moduleHandle: 'verify_reports',
            fields: {
              snapshot: '{{v.snapshotID}}',
              source: '{{v.sourceID}}',
              level: '{{v.level}}',
              status: 'running',
            },
          },
        },
        component('run', 'backup/verify', 'Verify', {
          snapshotID: '{{v.snapshotID}}',
          verifyLevel: '{{v.level}}',
          jobID: '{{createdRecordID}}',
        }),
      ],
      edges: [
        { from: 'due', to: 'loop' },
        { from: 'loop', to: 'report' },
        { from: 'loop', to: 'run' },
      ],
    },
    {
      id: 'backup-request-approve',
      name: 'Backup: одобрить заявку на резервирование',
      description:
        'Со страницы заявки: создаёт источник (fs) и политику из параметров заявки, ' +
        'переводит заявку в approved. Конфиденциальные классы получают vault=1, ' +
        'ПДн спецкатегорий — engine=kopia (иначе бэкап откажет на enforce).',
      entryNode: 'source',
      namespaceID: ns,
      nodes: [
        {
          id: 'source',
          type: 'crud',
          label: 'Create source',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: sources,
            moduleHandle: 'sources',
            fields: {
              name: '{{name}}',
              type: 'fs',
              path: '{{path}}',
              data_class: '{{dataClass}}',
              enabled: '1',
              notes: 'Заявка {{recordID}} от {{owner}}',
            },
          },
        },
        {
          id: 'policy',
          type: 'crud',
          label: 'Create policy',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: policies,
            moduleHandle: 'policies',
            fields: {
              name: '{{name}}',
              source: '{{createdRecordID}}',
              cron: '{{cron}}',
              retention_days: '{{retentionDays}}',
              vault: '1',
              engine: '{{engine}}',
              enabled: '1',
            },
          },
        },
        {
          id: 'mark',
          type: 'crud',
          label: 'Approve request',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: requests,
            moduleHandle: 'backup_requests',
            recordID: '{{recordID}}',
            fields: { status: 'approved', comment: 'Источник+политика созданы' },
          },
        },
      ],
      edges: [{ from: 'source', to: 'policy' }, { from: 'policy', to: 'mark' }],
    },
    {
      id: 'backup-audit-due',
      name: 'Backup: аудит незабкапленных данных',
      description: 'Запускает backup/audit по всем включённым областям audit_scopes. Находки — в audit_findings.',
      entryNode: 'run',
      namespaceID: ns,
      nodes: [component('run', 'backup/audit', 'Аудит')],
      edges: [],
    },
    {
      id: 'backup-audit-scope',
      name: 'Backup: аудит одной области',
      description: 'С карточки области: запускает backup/audit только для этой области (scopeID из контекста). Явный запуск выполняется даже для выключенной области.',
      entryNode: 'run',
      namespaceID: ns,
      nodes: [component('run', 'backup/audit', 'Аудит области', { scopeID: '{{scopeID}}' })],
      edges: [],
    },
    {
      id: 'backup-manifest',
      name: 'Backup: манифест архива',
      description: 'Со страницы снапшота: подписанный перечень файлов с SHA-256.',
      entryNode: 'run',
      namespaceID: ns,
      nodes: [component('run', 'backup/archive-manifest', 'Манифест', { snapshotID: '{{recordID}}' })],
      edges: [],
    },
    {
      id: 'backup-destroy',
      name: 'Backup: уничтожение по сроку хранения',
      description:
        'Со страницы акта уничтожения: создаёт/обновляет акт и вызывает backup/destroy. ' +
        'Legal hold в сейфе заблокирует удаление — статус акта меняется на blocked.',
      entryNode: 'act',
      namespaceID: ns,
      nodes: [
        {
          id: 'act',
          type: 'crud',
          label: 'Mark act registered',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: acts,
            moduleHandle: 'destruction_acts',
            recordID: '{{recordID}}',
            omitEmpty: true,
            fields: { status: 'registered' },
          },
        },
        component('run', 'backup/destroy', 'Уничтожить', {
          snapshotID: '{{snapshot}}',
          basis: '{{basis}}',
          jobID: '{{recordID}}',
        }),
      ],
      edges: [{ from: 'act', to: 'run' }],
    },
  ]
}
