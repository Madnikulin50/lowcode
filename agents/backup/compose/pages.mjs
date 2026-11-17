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
    listPage('Назначения', 'destinations', 15, 'fas warehouse', m.destinations, ['name', 'type', 'endpoint', 'bucket', 'default', 'enabled', 'last_check_status']),
    destinationCard(m),
    listPage('Политики', 'policies', 20, 'fas clock', m.policies, ['name', 'source', 'destination', 'cron', 'retention_days', 'incremental', 'enabled', 'last_run']),
    policyCard(m),
    listPage('Джобы', 'jobs', 30, 'fas play', m.jobs, ['source', 'status', 'kind', 'progress', 'bytes_written', 'started_at', 'error']),
    jobCard(m),
    listPage('Снапшоты', 'snapshots', 40, 'fas database', m.snapshots, ['source', 'destination', 's3_key', 'size_bytes', 'kind', 'engine', 'scan_status', 'vault_locked_until', 'verified', 'expires_at', 'restorable']),
    snapshotCard(m),
    listPage('Восстановления', 'restores', 50, 'fas download', m.restores, ['snapshot', 'dest_type', 'dest_path', 'status', 'started_at']),
    restoreCard(m),
    listPage('Верификации', 'verify-reports', 55, 'fas vial', m.verify_reports, ['snapshot', 'source', 'level', 'status', 'rto_sec', 'created_at']),
    hiddenRecord('Верификация', 'verify-report', m.verify_reports, 56, ['snapshot', 'source', 'level', 'status', 'rto_sec', 'log', 'created_at'], { readOnly: true }),
    listPage('Заявки на резервирование', 'backup-requests', 80, 'fas inbox', m.backup_requests, ['name', 'owner', 'path', 'data_class', 'status', 'created_at']),
    requestCard(m),
    listPage('Аудит: области', 'audit-scopes', 90, 'fas compass', m.audit_scopes, ['name', 'type', 'path', 'masks', 'enabled']),
    auditScopeCard(m),
    listPage('Аудит: находки', 'audit-findings', 91, 'fas triangle-exclamation', m.audit_findings, ['scope', 'path', 'criticality', 'guessed_class', 'status', 'created_at']),
    auditFindingCard(m),
    listPage('Акты уничтожения', 'destruction-acts', 92, 'fas file-contract', m.destruction_acts, ['snapshot', 'basis', 'status', 'destroyed_by', 'created_at']),
    listPage('Агенты', 'agents', 60, 'fas server', m.agents, ['name', 'hostname', 'url', 'status', 'last_seen', 'capabilities']),
    hiddenRecord('Агент', 'agent', m.agents, 61, ['name', 'url', 'hostname', 'capabilities', 'last_seen', 'status'], { readOnlyFields: ['last_seen', 'status', 'capabilities', 'hostname', 'url'] }),
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
      ruleChain('Run due', [0, 14, 12, 10], {
        chainID: 'backup-run-due',
        label: 'Запустить due',
        variant: 'primary',
        icon: 'play',
      }),
      ruleChain('Prune', [12, 14, 12, 10], {
        chainID: 'backup-prune',
        label: 'Очистить по retention',
        variant: 'warning',
        icon: 'trash',
      }),
      ruleChain('Audit', [24, 14, 12, 10], {
        chainID: 'backup-audit-due',
        label: 'Аудит данных',
        variant: 'info',
        icon: 'compass',
      }),
      block('Chart', 'Джобы по статусу', [36, 14, 12, 18], { chartID: String(charts.jobsByStatus) }),
      block('Chart', 'Источники по типу', [0, 24, 16, 18], { chartID: String(charts.sourcesByType) }),
      block('Chart', 'Снапшоты по движку', [16, 24, 16, 18], { chartID: String(charts.snapshotsByEngine) }),
      recordList('Текущие джобы', [0, 42, 24, 20], m.jobs, ['source', 'status', 'progress', 'started_at'], {
        prefilter: "status = 'running'",
        perPage: 8,
        hideAddButton: true,
        magnifyOption: 'modal',
      }),
      recordList('Последние ошибки', [24, 42, 24, 20], m.jobs, ['source', 'error', 'finished_at'], {
        prefilter: "status = 'failed'",
        perPage: 8,
        hideAddButton: true,
        magnifyOption: 'modal',
      }),
      recordList('Снапшоты', [0, 62, 48, 20], m.snapshots, ['source', 's3_key', 'size_bytes', 'kind', 'expires_at'], {
        perPage: 10,
        hideAddButton: true,
        magnifyOption: 'modal',
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

function hiddenRecord (title, handle, moduleID, weight, fields, extra = {}) {
  return {
    title,
    handle,
    moduleID: String(moduleID),
    visible: false,
    weight,
    config: pageIcon('fas file-alt'),
    blocks: withBlockIDs([
      recordBlock(title, [0, 0, 48, 28], fields, { hideEmptyFields: true, ...extra }),
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
        's3_bucket', 's3_prefix', 's3_region', 's3_secure',
        'provider', 'base_url', 'account', 'namespaces', 'include_pv', 'last_sync', 'notes',
      ], {
        hideEmptyFields: true,
        readOnlyFields: ['last_sync'],
        fieldRoles: {
          name: 'title',
          type: 'badge',
          enabled: 'badge',
          path: 'subtitle',
          notes: 'body',
        },
        fieldConditions: [
          { field: 'path', condition: 'record.values.type == "fs"', clearOnHide: true },
          { field: 'share', condition: 'record.values.type == "smb"', clearOnHide: true },
          { field: 'smb_path', condition: 'record.values.type == "smb"', clearOnHide: true },
          { field: 'host', condition: 'record.values.type in ["smb", "database", "s3"]', clearOnHide: true },
          { field: 'db_engine', condition: 'record.values.type == "database"', clearOnHide: true },
          { field: 'db_name', condition: 'record.values.type == "database"', clearOnHide: true },
          { field: 'db_port', condition: 'record.values.type == "database"', clearOnHide: true },
          { field: 's3_bucket', condition: 'record.values.type == "s3"', clearOnHide: true },
          { field: 's3_prefix', condition: 'record.values.type == "s3"', clearOnHide: true },
          { field: 's3_region', condition: 'record.values.type == "s3"', clearOnHide: true },
          { field: 's3_secure', condition: 'record.values.type == "s3"', clearOnHide: true },
          { field: 'credential', condition: 'record.values.type in ["smb", "database", "s3", "saas", "kubernetes"]', clearOnHide: true },
          { field: 'provider', condition: 'record.values.type == "saas"', clearOnHide: true },
          { field: 'base_url', condition: 'record.values.type == "saas"', clearOnHide: true },
          { field: 'account', condition: 'record.values.type == "saas"', clearOnHide: true },
          { field: 'last_sync', condition: 'record.values.type == "saas"', clearOnHide: true },
          { field: 'namespaces', condition: 'record.values.type == "kubernetes"', clearOnHide: true },
          { field: 'include_pv', condition: 'record.values.type == "kubernetes"', clearOnHide: true },
        ],
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
        magnifyOption: 'modal',
      }),
      recordList('Джобы', [0, 28, 24, 20], m.jobs, ['status', 'kind', 'progress', 'started_at'], {
        prefilter: `source = ${recID}`,
        refField: 'source',
        magnifyOption: 'modal',
      }),
      recordList('Снапшоты', [24, 28, 24, 20], m.snapshots, ['s3_key', 'size_bytes', 'kind', 'restorable'], {
        prefilter: `source = ${recID}`,
        refField: 'source',
        magnifyOption: 'modal',
      }),
    ]),
  }
}

function destinationCard (m) {
  return {
    title: 'Назначение',
    handle: 'destination',
    moduleID: String(m.destinations),
    visible: false,
    weight: 16,
    config: pageIcon('fas warehouse'),
    blocks: withBlockIDs([
      recordBlock('Назначение', [0, 0, 32, 28], [
        'name', 'handle', 'type', 'enabled', 'default',
        'endpoint', 'port', 'bucket', 'path', 'share', 'remote',
        'prefix', 'options', 'region', 'secure', 'credential', 'username', 'password',
        'last_check', 'last_check_status', 'last_check_error', 'notes',
      ], {
        hideEmptyFields: true,
        readOnlyFields: ['last_check', 'last_check_status', 'last_check_error'],
        fieldRoles: { name: 'title', endpoint: 'subtitle', enabled: 'badge', default: 'badge', last_check_status: 'badge', type: 'badge' },
        fieldConditions: [
          { field: 'username', condition: 'record.values.type == "sftp" || record.values.type == "smb" || record.values.type == "webdav"', clearOnHide: true },
          { field: 'password', condition: 'record.values.type == "sftp" || record.values.type == "smb" || record.values.type == "webdav"', clearOnHide: true },
        ],
      }),
      ruleChain('Проверить', [32, 0, 16, 8], {
        chainID: 'backup-dest-test',
        label: 'Проверить',
        icon: 'plug',
        context: { destinationID: recID },
      }),
      recordList('Политики', [32, 8, 16, 20], m.policies, ['name', 'cron', 'enabled'], {
        prefilter: `destination = ${recID}`,
        refField: 'destination',
        perPage: 8,
        magnifyOption: 'modal',
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
        'name', 'source', 'destination', 'cron', 'retention_days', 'incremental', 'compression', 'enabled', 'last_run', 'dest_prefix', 'notes',
        'vault', 'vault_retention_days', 'vault_legal_hold',
        'verify', 'verify_cron', 'verify_level', 'engine', 'exclude', 'dest_bucket',
      ], {
        hideEmptyFields: true,
        readOnlyFields: ['last_run'],
        fieldRoles: { name: 'title', enabled: 'badge', cron: 'subtitle' },
        fieldConditions: [
          { field: 'vault_retention_days', condition: 'record.values.vault', clearOnHide: true },
          { field: 'vault_legal_hold', condition: 'record.values.vault', clearOnHide: true },
          { field: 'verify_cron', condition: 'record.values.verify', clearOnHide: true },
          { field: 'verify_level', condition: 'record.values.verify', clearOnHide: true },
        ],
      }),
      ruleChain('Run', [32, 0, 16, 8], {
        chainID: 'backup-run-policy',
        label: 'Запустить',
        icon: 'play',
        context: { policyID: recID },
      }),
      ruleChain('Prune', [32, 8, 16, 8], {
        chainID: 'backup-prune',
        label: 'Prune',
        variant: 'warning',
        icon: 'trash',
        context: { policyID: recID },
      }),
      ruleChain('Verify', [32, 16, 16, 6], {
        chainID: 'backup-verify',
        label: 'Верифицировать',
        variant: 'info',
        icon: 'vial',
        context: { recordID: recID, verifyLevel: 'l1' },
      }),
      recordList('Джобы', [0, 22, 48, 22], m.jobs, ['status', 'kind', 'progress', 'started_at', 'error'], {
        prefilter: `policy = ${recID}`,
        refField: 'policy',
        magnifyOption: 'modal',
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
        hideEmptyFields: true,
        readOnly: true,
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
        'source', 'policy', 'job', 'destination', 's3_endpoint', 's3_bucket', 's3_key', 'size_bytes', 'checksum',
        'files_count', 'kind', 'engine', 'restic_id', 'expires_at', 'restorable', 'verified',
        'scan_status', 'scanned_at', 'scan_log', 'vault_key', 'vault_locked_until',
        'verified_at', 'verify_duration_sec', 'verify_error',
      ], {
        hideEmptyFields: true,
        readOnly: true,
        fieldRoles: { s3_key: 'title', kind: 'badge', restorable: 'badge', scan_status: 'badge', checksum: 'subtitle' },
      }),
      ruleChain('Restore', [32, 0, 16, 8], {
        chainID: 'backup-restore',
        label: 'Восстановить на место',
        icon: 'history',
        context: { snapshotID: recID, destType: 'original' },
      }),
      ruleChain('Vault', [32, 8, 16, 8], {
        chainID: 'backup-vault',
        label: 'В сейф',
        variant: 'success',
        icon: 'vault',
        context: { recordID: recID },
      }),
      ruleChain('Manifest', [32, 16, 16, 8], {
        chainID: 'backup-manifest',
        label: 'Манифест',
        variant: 'secondary',
        icon: 'file-contract',
        context: { recordID: recID },
      }),
      recordList('Восстановления', [32, 24, 16, 10], m.restores, ['status', 'dest_type', 'dest_path', 'started_at'], {
        prefilter: `snapshot = ${recID}`,
        refField: 'snapshot',
        magnifyOption: 'modal',
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
      recordBlock('Восстановление', [0, 0, 32, 22], [
        'snapshot', 'dest_type', 'dest_path', 'status', 'progress', 'error', 'message', 'started_at', 'finished_at',
      ], {
        hideEmptyFields: true,
        readOnlyFields: ['status', 'progress', 'error', 'message', 'started_at', 'finished_at'],
        fieldRoles: { status: 'badge', dest_type: 'badge', error: 'body' },
        fieldConditions: [
          { field: 'dest_path', condition: 'record.values.dest_type == "path"', clearOnHide: true },
        ],
      }),
      ruleChain('Запустить', [32, 0, 16, 8], {
        chainID: 'backup-restore-run',
        label: 'Запустить восстановление',
        icon: 'play',
        context: { recordID: recID },
      }),
    ]),
  }
}

function requestCard (m) {
  return {
    title: 'Заявка',
    handle: 'backup-request',
    moduleID: String(m.backup_requests),
    visible: false,
    weight: 81,
    config: pageIcon('fas inbox'),
    blocks: withBlockIDs([
      recordBlock('Заявка', [0, 0, 32, 24], [
        'name', 'owner', 'path', 'data_class', 'cron', 'retention_days', 'status', 'comment', 'created_at',
      ], {
        hideEmptyFields: true,
        fieldRoles: { name: 'title', status: 'badge', data_class: 'badge', owner: 'subtitle' },
      }),
      ruleChain('Approve', [32, 0, 16, 10], {
        chainID: 'backup-request-approve',
        label: 'Одобрить',
        variant: 'success',
        icon: 'check',
        context: { recordID: recID, engine: '' },
      }),
      ruleChain('ApprovePdn', [32, 10, 16, 10], {
        chainID: 'backup-request-approve',
        label: 'Одобрить (ПДн, kopia)',
        variant: 'primary',
        icon: 'lock',
        context: { recordID: recID, engine: 'kopia' },
      }),
    ]),
  }
}

function auditScopeCard (m) {
  return {
    title: 'Область аудита',
    handle: 'audit-scope',
    moduleID: String(m.audit_scopes),
    visible: false,
    weight: 93,
    config: pageIcon('fas compass'),
    blocks: withBlockIDs([
      recordBlock('Область аудита', [0, 0, 32, 26], [
        'name', 'type', 'enabled', 'path', 'host', 'share', 'smb_path',
        'credential', 'masks', 'min_size_mb', 'content_scan', 'agent',
      ], {
        hideEmptyFields: true,
        fieldRoles: { name: 'title', type: 'badge', enabled: 'badge', path: 'subtitle' },
        fieldConditions: [
          { field: 'host', condition: 'record.values.type == "smb"', clearOnHide: true },
          { field: 'share', condition: 'record.values.type == "smb"', clearOnHide: true },
          { field: 'smb_path', condition: 'record.values.type == "smb"', clearOnHide: true },
          { field: 'credential', condition: 'record.values.type == "smb"', clearOnHide: true },
        ],
      }),
      ruleChain('Аудит', [32, 0, 16, 6], {
        chainID: 'backup-audit-scope',
        label: 'Запустить аудит',
        variant: 'info',
        icon: 'radar',
        context: { scopeID: recID },
      }),
      recordList('Находки', [32, 6, 16, 20], m.audit_findings, ['path', 'criticality', 'status', 'created_at'], {
        prefilter: `scope = ${recID}`,
        refField: 'scope',
        perPage: 10,
        magnifyOption: 'modal',
      }),
    ]),
  }
}

function auditFindingCard (m) {
  return {
    title: 'Находка аудита',
    handle: 'audit-finding',
    moduleID: String(m.audit_findings),
    visible: false,
    weight: 94,
    config: pageIcon('fas triangle-exclamation'),
    blocks: withBlockIDs([
      recordBlock('Находка аудита', [0, 0, 48, 28], [
        'scope', 'path', 'size_bytes', 'mod_time', 'guessed_class', 'criticality', 'reasons', 'reason', 'status', 'created_at',
      ], {
        hideEmptyFields: true,
        readOnlyFields: ['scope', 'path', 'size_bytes', 'mod_time', 'guessed_class', 'criticality', 'reasons', 'reason', 'created_at'],
        fieldRoles: { path: 'title', criticality: 'badge', guessed_class: 'badge', status: 'badge', scope: 'subtitle', reasons: 'body' },
      }),
    ]),
  }
}
