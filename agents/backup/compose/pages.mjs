import {
  block, recordList, recordBlock, metricBlock, metricItem, ruleChain,
  withBlockIDs, pageIcon,
} from './helpers.mjs'

const recID = '${recordID}'

export function buildPages ({ modules, charts }) {
  const m = modules
  return [
    dashboardPage(m, charts),
    listPage('Источники', 'sources', 10, 'fas folder', m.sources, ['name', 'type', 'enabled', 'path', 'host', 'agent']),
    sourceCard(m),
    listPage('Политики', 'policies', 20, 'fas clock', m.policies, ['name', 'source', 'cron', 'retention_days', 'incremental', 'enabled', 'last_run']),
    policyCard(m),
    listPage('Джобы', 'jobs', 30, 'fas play', m.jobs, ['source', 'status', 'kind', 'progress', 'bytes_written', 'started_at', 'error']),
    jobCard(m),
    listPage('Снапшоты', 'snapshots', 40, 'fas database', m.snapshots, ['source', 's3_key', 'size_bytes', 'kind', 'engine', 'expires_at', 'restorable']),
    snapshotCard(m),
    listPage('Восстановления', 'restores', 50, 'fas download', m.restores, ['snapshot', 'dest_type', 'dest_path', 'status', 'started_at']),
    restoreCard(m),
    listPage('Агенты', 'agents', 60, 'fas server', m.agents, ['name', 'hostname', 'url', 'status', 'last_seen', 'capabilities']),
    hiddenRecord('Агент', 'agent', m.agents, 61, ['name', 'url', 'hostname', 'capabilities', 'last_seen', 'status']),
    listPage('Секреты', 'credentials', 70, 'fas key', m.credentials, ['name', 'handle', 'kind', 'username']),
    hiddenRecord('Секрет', 'credential', m.credentials, 71, ['name', 'handle', 'kind', 'username', 'extra']),
  ]
}

function dashboardPage (m, charts) {
  return {
    title: 'Дашборд',
    handle: 'dashboard',
    visible: true,
    weight: 0,
    description: 'Резервное копирование во внутренний MinIO',
    config: pageIcon('fas chart-pie'),
    blocks: withBlockIDs([
      metricBlock('Сводка', [0, 0, 48, 14], [
        metricItem('Источники', m.sources, "enabled = '1'", { role: 'hero', color: '#2e59d9', fontSize: '28' }),
        metricItem('Идут сейчас', m.jobs, "status = 'running'", { role: 'balloon', color: '#36b9cc' }),
        metricItem('Ошибки', m.jobs, "status = 'failed'", { role: 'balloon', color: '#e74a3b' }),
        metricItem('Снапшоты', m.snapshots, "restorable = '1'", { role: 'meta', color: '#1cc88a' }),
      ]),
      ruleChain('Run due', [0, 14, 16, 10], {
        chainID: 'backup-run-due',
        label: 'Запустить due',
        variant: 'primary',
        icon: 'play',
      }),
      ruleChain('Prune', [16, 14, 16, 10], {
        chainID: 'backup-prune',
        label: 'Очистить по retention',
        variant: 'warning',
        icon: 'trash',
      }),
      block('Chart', 'Джобы по статусу', [32, 14, 16, 18], { chartID: String(charts.jobsByStatus) }),
      block('Chart', 'Источники по типу', [0, 24, 16, 18], { chartID: String(charts.sourcesByType) }),
      block('Chart', 'Снапшоты по движку', [16, 24, 16, 18], { chartID: String(charts.snapshotsByEngine) }),
      recordList('Текущие джобы', [0, 42, 24, 20], m.jobs, ['source', 'status', 'progress', 'started_at'], {
        prefilter: "status = 'running'",
        perPage: 8,
        hideAddButton: true,
      }),
      recordList('Последние ошибки', [24, 42, 24, 20], m.jobs, ['source', 'error', 'finished_at'], {
        prefilter: "status = 'failed'",
        perPage: 8,
        hideAddButton: true,
      }),
      recordList('Снапшоты', [0, 62, 48, 20], m.snapshots, ['source', 's3_key', 'size_bytes', 'kind', 'expires_at'], {
        perPage: 10,
        hideAddButton: true,
      }),
    ]),
  }
}

function listPage (title, handle, weight, icon, moduleID, fields) {
  return {
    title,
    handle,
    visible: true,
    weight,
    config: pageIcon(icon),
    blocks: withBlockIDs([
      recordList(title, [0, 0, 48, 48], moduleID, fields),
    ]),
  }
}

function hiddenRecord (title, handle, moduleID, weight, fields) {
  return {
    title,
    handle,
    moduleID: String(moduleID),
    visible: false,
    weight,
    config: pageIcon('fas file-alt'),
    blocks: withBlockIDs([
      recordBlock(title, [0, 0, 48, 28], fields),
    ]),
  }
}

function sourceCard (m) {
  return {
    title: 'Источник',
    handle: 'source',
    moduleID: String(m.sources),
    visible: false,
    weight: 11,
    config: pageIcon('fas folder'),
    blocks: withBlockIDs([
      recordBlock('Источник', [0, 0, 32, 28], [
        'name', 'type', 'enabled', 'agent', 'credential',
        'path', 'host', 'share', 'smb_path',
        'db_engine', 'db_name', 'db_port',
        's3_bucket', 's3_prefix', 's3_region', 's3_secure', 'notes',
      ], {
        fieldRoles: {
          name: 'title',
          type: 'badge',
          enabled: 'badge',
          path: 'subtitle',
          notes: 'body',
        },
      }),
      ruleChain('Backup', [32, 0, 16, 10], {
        chainID: 'backup-run-source',
        label: 'Запустить бэкап',
        icon: 'play',
        context: { sourceID: recID },
      }),
      recordList('Политики', [32, 10, 16, 18], m.policies, ['name', 'cron', 'enabled'], {
        prefilter: `source = ${recID}`,
        refField: 'source',
        perPage: 8,
      }),
      recordList('Джобы', [0, 28, 24, 20], m.jobs, ['status', 'kind', 'progress', 'started_at'], {
        prefilter: `source = ${recID}`,
        refField: 'source',
      }),
      recordList('Снапшоты', [24, 28, 24, 20], m.snapshots, ['s3_key', 'size_bytes', 'kind', 'restorable'], {
        prefilter: `source = ${recID}`,
        refField: 'source',
      }),
    ]),
  }
}

function policyCard (m) {
  return {
    title: 'Политика',
    handle: 'policy',
    moduleID: String(m.policies),
    visible: false,
    weight: 21,
    config: pageIcon('fas clock'),
    blocks: withBlockIDs([
      recordBlock('Политика', [0, 0, 32, 22], [
        'name', 'source', 'cron', 'retention_days', 'incremental', 'compression', 'enabled', 'last_run', 'dest_prefix', 'notes',
      ], {
        fieldRoles: { name: 'title', enabled: 'badge', cron: 'subtitle' },
      }),
      ruleChain('Run', [32, 0, 16, 10], {
        chainID: 'backup-run-policy',
        label: 'Запустить',
        icon: 'play',
        context: { policyID: recID },
      }),
      ruleChain('Prune', [32, 10, 16, 12], {
        chainID: 'backup-prune',
        label: 'Prune',
        variant: 'warning',
        icon: 'trash',
        context: { policyID: recID },
      }),
      recordList('Джобы', [0, 22, 48, 22], m.jobs, ['status', 'kind', 'progress', 'started_at', 'error'], {
        prefilter: `policy = ${recID}`,
        refField: 'policy',
      }),
    ]),
  }
}

function jobCard (m) {
  return {
    title: 'Джоб',
    handle: 'job',
    moduleID: String(m.jobs),
    visible: false,
    weight: 31,
    config: pageIcon('fas play'),
    blocks: withBlockIDs([
      recordBlock('Джоб', [0, 0, 48, 24], [
        'source', 'policy', 'status', 'kind', 'engine', 'progress',
        'bytes_read', 'bytes_written', 'files_count', 'started_at', 'finished_at', 'error', 'message',
      ], {
        fieldRoles: { status: 'badge', kind: 'badge', error: 'body' },
      }),
    ]),
  }
}

function snapshotCard (m) {
  return {
    title: 'Снапшот',
    handle: 'snapshot',
    moduleID: String(m.snapshots),
    visible: false,
    weight: 41,
    config: pageIcon('fas database'),
    blocks: withBlockIDs([
      recordBlock('Снапшот', [0, 0, 32, 26], [
        'source', 'policy', 'job', 's3_bucket', 's3_key', 'size_bytes', 'checksum',
        'files_count', 'kind', 'engine', 'restic_id', 'expires_at', 'restorable', 'verified',
      ], {
        fieldRoles: { s3_key: 'title', kind: 'badge', restorable: 'badge', checksum: 'subtitle' },
      }),
      ruleChain('Restore', [32, 0, 16, 12], {
        chainID: 'backup-restore',
        label: 'Восстановить',
        icon: 'download',
        context: { snapshotID: recID, destType: 'path' },
      }),
      recordList('Восстановления', [32, 12, 16, 14], m.restores, ['status', 'dest_type', 'started_at'], {
        prefilter: `snapshot = ${recID}`,
        refField: 'snapshot',
      }),
    ]),
  }
}

function restoreCard (m) {
  return {
    title: 'Восстановление',
    handle: 'restore',
    moduleID: String(m.restores),
    visible: false,
    weight: 51,
    config: pageIcon('fas download'),
    blocks: withBlockIDs([
      recordBlock('Восстановление', [0, 0, 48, 22], [
        'snapshot', 'dest_type', 'dest_path', 'status', 'progress', 'error', 'started_at', 'finished_at',
      ], {
        fieldRoles: { status: 'badge', dest_type: 'badge', error: 'body' },
      }),
    ]),
  }
}
