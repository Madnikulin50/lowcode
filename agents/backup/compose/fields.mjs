import {
  selectOptions, field, recordRel, numberField, boolSwitch,
} from './helpers.mjs'

export const SOURCE_TYPES = [
  ['fs', 'Локальная папка', { backgroundColor: 'info', textColor: 'white' }],
  ['smb', 'SMB-шара', { backgroundColor: 'primary', textColor: 'white' }],
  ['database', 'База данных', { backgroundColor: 'warning', textColor: 'dark' }],
  ['s3', 'Внешний S3', { backgroundColor: 'secondary', textColor: 'white' }],
]

export const JOB_STATUS = [
  ['pending', 'Ожидает', { backgroundColor: 'light', textColor: 'dark' }],
  ['running', 'Идёт', { backgroundColor: 'info', textColor: 'white' }],
  ['completed', 'Готово', { backgroundColor: 'success', textColor: 'white' }],
  ['failed', 'Ошибка', { backgroundColor: 'danger', textColor: 'white' }],
]

export const JOB_KIND = [
  ['full', 'Полный'],
  ['incremental', 'Инкремент'],
  ['restore', 'Восстановление'],
  ['prune', 'Очистка'],
]

export const DB_ENGINE = [
  ['postgres', 'PostgreSQL'],
  ['mysql', 'MySQL / MariaDB'],
]

export const CRED_KIND = [
  ['smb', 'SMB'],
  ['postgres', 'PostgreSQL'],
  ['mysql', 'MySQL'],
  ['s3', 'S3'],
  ['restic', 'Restic'],
]

export const AGENT_STATUS = [
  ['online', 'Online', { backgroundColor: 'success', textColor: 'white' }],
  ['offline', 'Offline', { backgroundColor: 'secondary', textColor: 'white' }],
]

export const DEST_TYPE = [
  ['original', 'Исходное место'],
  ['path', 'Путь / prefix'],
  ['download', 'Только ключ MinIO'],
]

export const COMPRESSION = [
  ['gzip', 'gzip'],
  ['none', 'без сжатия'],
]

export function dateTimeField (name, label) {
  return field(name, label, 'DateTime', {
    options: {
      format: '',
      onlyDate: false,
      onlyTime: false,
      onlyPastValues: false,
      onlyFutureValues: false,
      outputRelative: true,
      multiDelimiter: '\n',
    },
  })
}

export function agentFields () {
  return [
    field('name', 'Имя', 'String', { required: true }),
    field('url', 'URL', 'String'),
    field('hostname', 'Hostname', 'String'),
    field('capabilities', 'Возможности', 'String'),
    dateTimeField('last_seen', 'Последний контакт'),
    field('status', 'Статус', 'Select', { options: selectOptions(AGENT_STATUS) }),
  ]
}

export function credentialFields () {
  return [
    field('name', 'Имя', 'String', { required: true }),
    field('handle', 'Handle секрета', 'String', { required: true }),
    field('kind', 'Тип', 'Select', { options: selectOptions(CRED_KIND) }),
    field('username', 'Username / Access key', 'String'),
    field('extra', 'Extra (JSON)', 'String'),
  ]
}

export function sourceFields (agents, credentials) {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('type', 'Тип', 'Select', { options: selectOptions(SOURCE_TYPES), required: true }),
    recordRel('agent', 'Агент', agents, 'name', ['name', 'hostname']),
    recordRel('credential', 'Учётные данные', credentials, 'name', ['name', 'handle']),
    field('path', 'Путь / UNC', 'String'),
    field('host', 'Хост / S3 endpoint', 'String'),
    field('share', 'SMB share', 'String'),
    field('smb_path', 'Путь внутри шары', 'String'),
    field('db_engine', 'СУБД', 'Select', { options: selectOptions(DB_ENGINE) }),
    field('db_name', 'Имя БД', 'String'),
    numberField('db_port', 'Порт БД', { precision: 0, format: '0' }),
    field('s3_bucket', 'S3 bucket', 'String'),
    field('s3_prefix', 'S3 prefix', 'String'),
    field('s3_region', 'S3 region', 'String'),
    boolSwitch('s3_secure', 'S3 HTTPS'),
    boolSwitch('enabled', 'Включён'),
    field('notes', 'Заметки', 'String'),
  ]
}

export function policyFields (sources) {
  return [
    field('name', 'Название', 'String', { required: true }),
    recordRel('source', 'Источник', sources, 'name', ['name', 'type'], true),
    field('cron', 'Cron', 'String', { required: true }),
    numberField('retention_days', 'Хранить, дней', { precision: 0, format: '0' }),
    boolSwitch('incremental', 'Инкремент (restic)'),
    field('dest_prefix', 'Prefix в MinIO', 'String'),
    field('compression', 'Сжатие', 'Select', { options: selectOptions(COMPRESSION) }),
    boolSwitch('enabled', 'Включена'),
    dateTimeField('last_run', 'Последний запуск'),
    field('notes', 'Заметки', 'String'),
  ]
}

export function jobFields (policies, sources) {
  return [
    recordRel('policy', 'Политика', policies, 'name', ['name']),
    recordRel('source', 'Источник', sources, 'name', ['name', 'type']),
    field('status', 'Статус', 'Select', { options: selectOptions(JOB_STATUS) }),
    numberField('progress', 'Прогресс', { precision: 1, display: 'progress', suffix: ' %' }),
    numberField('bytes_read', 'Прочитано, байт', { precision: 0, format: '0' }),
    numberField('bytes_written', 'Записано, байт', { precision: 0, format: '0' }),
    numberField('files_count', 'Файлов', { precision: 0, format: '0' }),
    field('kind', 'Вид', 'Select', { options: selectOptions(JOB_KIND) }),
    field('engine', 'Движок', 'String'),
    field('error', 'Ошибка', 'String'),
    field('message', 'Сообщение', 'String'),
    dateTimeField('started_at', 'Начало'),
    dateTimeField('finished_at', 'Конец'),
  ]
}

export function snapshotFields (jobs, sources, policies) {
  return [
    recordRel('job', 'Джоб', jobs, 'started_at', ['started_at', 'status']),
    recordRel('source', 'Источник', sources, 'name', ['name']),
    recordRel('policy', 'Политика', policies, 'name', ['name']),
    field('s3_bucket', 'Bucket', 'String'),
    field('s3_key', 'Ключ', 'String'),
    numberField('size_bytes', 'Размер', { precision: 0, format: '0' }),
    field('checksum', 'SHA-256', 'String'),
    numberField('files_count', 'Файлов', { precision: 0, format: '0' }),
    field('kind', 'Вид', 'Select', { options: selectOptions(JOB_KIND) }),
    field('engine', 'Движок', 'String'),
    field('restic_id', 'Restic snapshot', 'String'),
    dateTimeField('expires_at', 'Истекает'),
    boolSwitch('restorable', 'Можно восстановить'),
    boolSwitch('verified', 'Проверен'),
  ]
}

export function restoreFields (snapshots) {
  return [
    recordRel('snapshot', 'Снапшот', snapshots, 's3_key', ['s3_key', 'checksum'], true),
    field('dest_type', 'Куда', 'Select', { options: selectOptions(DEST_TYPE) }),
    field('dest_path', 'Путь назначения', 'String'),
    field('status', 'Статус', 'Select', { options: selectOptions(JOB_STATUS) }),
    numberField('progress', 'Прогресс', { precision: 1, display: 'progress', suffix: ' %' }),
    field('error', 'Ошибка', 'String'),
    dateTimeField('started_at', 'Начало'),
    dateTimeField('finished_at', 'Конец'),
  ]
}
