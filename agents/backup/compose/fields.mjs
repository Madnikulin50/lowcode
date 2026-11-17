import {
  selectOptions, field, recordRel, numberField, boolSwitch,
} from './helpers.mjs'

export const SOURCE_TYPES = [
  ['fs', 'Локальная папка', { backgroundColor: 'info', textColor: 'white' }],
  ['smb', 'SMB-шара', { backgroundColor: 'primary', textColor: 'white' }],
  ['database', 'База данных', { backgroundColor: 'warning', textColor: 'dark' }],
  ['s3', 'Внешний S3', { backgroundColor: 'secondary', textColor: 'white' }],
  ['saas', 'SaaS (Bitrix24/почта/DAV)', { backgroundColor: 'success', textColor: 'white' }],
  ['kubernetes', 'Kubernetes', { backgroundColor: 'danger', textColor: 'white' }],
]

export const SAAS_PROVIDER = [
  ['bitrix24', 'Bitrix24'],
  ['imap', 'Почта (IMAP)'],
  ['caldav', 'Календарь (CalDAV)'],
  ['carddav', 'Контакты (CardDAV)'],
]

export const SCAN_STATUS = [
  ['queued', 'В очереди', { backgroundColor: 'light', textColor: 'dark' }],
  ['clean', 'Чисто', { backgroundColor: 'success', textColor: 'white' }],
  ['infected', 'Заражён', { backgroundColor: 'danger', textColor: 'white' }],
  ['unknown', 'Не проверен', { backgroundColor: 'warning', textColor: 'dark' }],
]

export const VERIFY_LEVEL = [
  ['l1', 'L1: целостность'],
  ['l2', 'L2: подъём в docker'],
]

export const ENGINE = [
  ['', 'auto'],
  ['archive', 'Полный архив'],
  ['restic', 'restic (инкремент)'],
  ['kopia', 'kopia (дедуп + шифрование)'],
]

export const DATA_CLASS = [
  ['public', 'Публичные'],
  ['internal', 'Внутренние'],
  ['confidential', 'Конфиденциальные'],
  ['personal_special', 'ПДн спец. категорий'],
]

export const REQUEST_STATUS = [
  ['new', 'Новая', { backgroundColor: 'light', textColor: 'dark' }],
  ['approved', 'Одобрена', { backgroundColor: 'success', textColor: 'white' }],
  ['rejected', 'Отклонена', { backgroundColor: 'danger', textColor: 'white' }],
]

export const FINDING_STATUS = [
  ['open', 'Открыта', { backgroundColor: 'danger', textColor: 'white' }],
  ['acknowledged', 'Принята', { backgroundColor: 'warning', textColor: 'dark' }],
  ['resolved', 'Закрыта', { backgroundColor: 'success', textColor: 'white' }],
]

export const CRIT_LEVEL = [
  ['low', 'Низкая', { backgroundColor: 'light', textColor: 'dark' }],
  ['medium', 'Средняя', { backgroundColor: 'info', textColor: 'white' }],
  ['high', 'Высокая', { backgroundColor: 'warning', textColor: 'dark' }],
  ['critical', 'Критичная', { backgroundColor: 'danger', textColor: 'white' }],
]

export const ACT_STATUS = [
  ['registered', 'Зарегистрирован', { backgroundColor: 'light', textColor: 'dark' }],
  ['executed', 'Исполнен', { backgroundColor: 'success', textColor: 'white' }],
  ['blocked', 'Заблокирован (legal hold)', { backgroundColor: 'danger', textColor: 'white' }],
]

export const VERIFY_STATUS = [
  ['running', 'Идёт', { backgroundColor: 'info', textColor: 'white' }],
  ['clean', 'Восстановился', { backgroundColor: 'success', textColor: 'white' }],
  ['failed', 'Не восстановился', { backgroundColor: 'danger', textColor: 'white' }],
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
  ['vault', 'Сейф'],
  ['scan', 'Скан'],
  ['verify', 'Верификация'],
  ['extract', 'Извлечение'],
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
  ['saas', 'SaaS-токен'],
  ['kubeconfig', 'kubeconfig (base64)'],
]

export const AGENT_STATUS = [
  ['online', 'Online', { backgroundColor: 'success', textColor: 'white' }],
  ['offline', 'Offline', { backgroundColor: 'secondary', textColor: 'white' }],
]

export const DEST_TYPE = [
  ['original', 'Исходное место'],
  ['path', 'Путь / prefix'],
  ['download', 'Только ключ MinIO'],
  ['cluster', 'K8s-кластер (kubectl apply)'],
  ['vault', 'Из WORM-сейфа (чистая копия)'],
]

export const DESTINATION_TYPE = [
  ['minio', 'MinIO / S3-совместимое'],
  ['local', 'Локальная папка агента'],
  ['nfs', 'NFS (server:/export → mountpoint)'],
  ['sftp', 'SSH/SFTP'],
  ['smb', 'SMB-шара'],
  ['rclone', 'rclone (70+ облаков)'],
  ['webdav', 'WebDAV'],
]

export const CHECK_STATUS = [
  ['ok', 'OK', { backgroundColor: 'success', textColor: 'white' }],
  ['failed', 'Ошибка', { backgroundColor: 'danger', textColor: 'white' }],
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
    field('provider', 'SaaS провайдер', 'Select', { options: selectOptions(SAAS_PROVIDER) }),
    field('base_url', 'Базовый URL / endpoint', 'String'),
    field('account', 'Аккаунт / ящик', 'String'),
    field('namespaces', 'K8s namespaces (csv, пусто = все)', 'String'),
    boolSwitch('include_pv', 'K8s: PV-манифесты'),
    field('data_class', 'Класс данных', 'Select', { options: selectOptions(DATA_CLASS) }),
    dateTimeField('last_sync', 'SaaS: последняя синхронизация'),
    boolSwitch('enabled', 'Включён'),
    field('notes', 'Заметки', 'String'),
  ]
}

export function destinationFields (credentials) {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('handle', 'Handle', 'String', { required: true }),
    field('type', 'Тип', 'Select', { options: selectOptions(DESTINATION_TYPE), required: true }),
    field('endpoint', 'Endpoint / хост / URL (sftp, smb, nfs server:/export, webdav)', 'String'),
    numberField('port', 'Порт (sftp, default 22)', { precision: 0, format: '0' }),
    field('bucket', 'Бакет (minio)', 'String'),
    field('path', 'Путь: корень local / mountpoint nfs / base dir sftp,webdav / подкаталог smb,rclone', 'String'),
    field('share', 'SMB share', 'String'),
    field('remote', 'rclone remote (имя из rclone config агента)', 'String'),
    field('prefix', 'Префикс по умолчанию', 'String'),
    field('options', 'Опции: nfs mount opts / smb JSON {"port","domain"}', 'String'),
    field('region', 'Region (minio)', 'String'),
    boolSwitch('secure', 'HTTPS (minio)'),
    recordRel('credential', 'Учётные данные (credentials, для minio/sftp/smb)', credentials, 'name', ['name', 'handle']),
    field('username', 'Логин (sftp/smb/webdav; перекрывает credentials)', 'String'),
    field('password', 'Пароль / PEM-ключ (перекрывает BACKUP_SECRET; kind Secret — шифруется)', 'Secret'),
    boolSwitch('default', 'Назначение по умолчанию'),
    boolSwitch('enabled', 'Включено'),
    field('notes', 'Заметки', 'String'),
    dateTimeField('last_check', 'Проверено'),
    field('last_check_status', 'Проверка', 'Select', { options: selectOptions(CHECK_STATUS) }),
    field('last_check_error', 'Ошибка проверки', 'String'),
  ]
}

export function policyFields (sources, destinations) {
  return [
    field('name', 'Название', 'String', { required: true }),
    recordRel('source', 'Источник', sources, 'name', ['name', 'type'], true),
    recordRel('destination', 'Назначение (пусто = default/env)', destinations, 'name', ['name', 'handle']),
    field('cron', 'Cron', 'String', { required: true }),
    numberField('retention_days', 'Хранить, дней', { precision: 0, format: '0' }),
    boolSwitch('incremental', 'Инкремент (restic)'),
    field('dest_prefix', 'Prefix в MinIO (пусто = префикс назначения)', 'String'),
    field('compression', 'Сжатие', 'Select', { options: selectOptions(COMPRESSION) }),
    boolSwitch('vault', 'Сейф (WORM-копия + скан)'),
    numberField('vault_retention_days', 'Сейф: хранить дней', { precision: 0, format: '0' }),
    boolSwitch('vault_legal_hold', 'Сейф: legal hold на чистые'),
    boolSwitch('verify', 'Верификация восстановления'),
    field('verify_cron', 'Верификация: cron', 'String'),
    field('verify_level', 'Верификация: уровень', 'Select', { options: selectOptions(VERIFY_LEVEL) }),
    field('engine', 'Движок (fs)', 'Select', { options: selectOptions(ENGINE) }),
    field('exclude', 'Исключения (glob, через запятую или строку)', 'String'),
    field('dest_bucket', 'Бакет назначения (legacy, пусто = default)', 'String'),
    boolSwitch('archive', 'Юрзначимый архив'),
    numberField('keep_years', 'Архив: хранить лет', { precision: 0, format: '0' }),
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
    numberField('progress', 'Прогресс', { precision: 1, display: 'progress', suffix: ' %', max: 100 }),
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

export function snapshotFields (jobs, sources, policies, destinations) {
  return [
    recordRel('job', 'Джоб', jobs, 'started_at', ['started_at', 'status']),
    recordRel('source', 'Источник', sources, 'name', ['name']),
    recordRel('policy', 'Политика', policies, 'name', ['name']),
    recordRel('destination', 'Назначение', destinations, 'name', ['name']),
    field('dest_type', 'Тип назначения', 'String'),
    field('s3_endpoint', 'Endpoint / локатор', 'String'),
    field('s3_bucket', 'Bucket / корень', 'String'),
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
    dateTimeField('verified_at', 'Верифицирован'),
    numberField('verify_duration_sec', 'Верификация, сек', { precision: 0, format: '0' }),
    field('verify_error', 'Ошибка верификации', 'String'),
    field('vault_key', 'Ключ в сейфе', 'String'),
    dateTimeField('vault_locked_until', 'Сейф: заблокировано до'),
    field('scan_status', 'Скан', 'Select', { options: selectOptions(SCAN_STATUS) }),
    dateTimeField('scanned_at', 'Сканирован'),
    field('scan_log', 'Лог скана', 'String'),
    field('data_class', 'Класс данных', 'Select', { options: selectOptions(DATA_CLASS) }),
    field('manifest_key', 'Манифест архива', 'String'),
    field('manifest_sha', 'SHA манифеста', 'String'),
  ]
}

export function verifyReportFields (snapshots, sources) {
  return [
    recordRel('snapshot', 'Снапшот', snapshots, 's3_key', ['s3_key'], true),
    recordRel('source', 'Источник', sources, 'name', ['name']),
    field('level', 'Уровень', 'Select', { options: selectOptions(VERIFY_LEVEL) }),
    field('status', 'Статус', 'Select', { options: selectOptions(VERIFY_STATUS) }),
    numberField('rto_sec', 'RTO, сек', { precision: 0, format: '0' }),
    field('log', 'Лог', 'String'),
    dateTimeField('created_at', 'Создан'),
  ]
}

export function backupRequestFields () {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('owner', 'Владелец данных', 'String', { required: true }),
    field('path', 'Путь к данным', 'String', { required: true }),
    field('data_class', 'Класс данных', 'Select', { options: selectOptions(DATA_CLASS) }),
    field('cron', 'Периодичность (cron)', 'String'),
    numberField('retention_days', 'Хранить, дней', { precision: 0, format: '0' }),
    field('status', 'Статус', 'Select', { options: selectOptions(REQUEST_STATUS) }),
    field('comment', 'Комментарий', 'String'),
    dateTimeField('created_at', 'Создана'),
  ]
}

export function auditScopeFields (credentials, agents) {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('type', 'Тип', 'Select', { options: selectOptions([['fs', 'Локальная папка'], ['smb', 'SMB-шара']]) }),
    field('path', 'Путь / UNC', 'String', { required: true }),
    field('host', 'Хост / endpoint', 'String'),
    field('share', 'SMB share', 'String'),
    field('smb_path', 'Путь внутри шары', 'String'),
    recordRel('credential', 'Учётные данные', credentials, 'name', ['name', 'handle']),
    field('masks', 'Маски файлов (через запятую)', 'String'),
    numberField('min_size_mb', 'Мин. размер, МБ', { precision: 0, format: '0' }),
    recordRel('agent', 'Агент', agents, 'name', ['name', 'hostname']),
    boolSwitch('content_scan', 'Сканировать содержимое (ПДн, fs)'),
    boolSwitch('enabled', 'Включён'),
  ]
}

export function auditFindingFields (scopes) {
  return [
    recordRel('scope', 'Область', scopes, 'name', ['name']),
    field('path', 'Путь', 'String'),
    numberField('size_bytes', 'Размер', { precision: 0, format: '0' }),
    dateTimeField('mod_time', 'Изменён'),
    field('guessed_class', 'Класс данных', 'Select', { options: selectOptions(DATA_CLASS) }),
    field('criticality', 'Критичность', 'Select', { options: selectOptions(CRIT_LEVEL) }),
    field('reasons', 'Основания оценки', 'String'),
    field('reason', 'Причина', 'String'),
    field('status', 'Статус', 'Select', { options: selectOptions(FINDING_STATUS) }),
    dateTimeField('created_at', 'Создана'),
  ]
}

export function destructionActFields (snapshots) {
  return [
    recordRel('snapshot', 'Снапшот', snapshots, 's3_key', ['s3_key'], true),
    field('manifest_sha', 'SHA манифеста', 'String'),
    numberField('files_count', 'Файлов', { precision: 0, format: '0' }),
    field('basis', 'Основание (срок хранения / приказ)', 'String', { required: true }),
    field('destroyed_by', 'Кто уничтожил', 'String'),
    field('status', 'Статус', 'Select', { options: selectOptions(ACT_STATUS) }),
    dateTimeField('created_at', 'Создан'),
  ]
}

export function restoreFields (snapshots) {
  return [
    recordRel('snapshot', 'Снапшот', snapshots, 's3_key', ['s3_key', 'checksum'], true),
    field('dest_type', 'Куда', 'Select', { options: selectOptions(DEST_TYPE) }),
    field('dest_path', 'Путь назначения', 'String'),
    field('status', 'Статус', 'Select', { options: selectOptions(JOB_STATUS) }),
    numberField('progress', 'Прогресс', { precision: 1, display: 'progress', suffix: ' %', max: 100 }),
    field('error', 'Ошибка', 'String'),
    field('message', 'Сообщение / куда восстановлено', 'String'),
    dateTimeField('started_at', 'Начало'),
    dateTimeField('finished_at', 'Конец'),
  ]
}
