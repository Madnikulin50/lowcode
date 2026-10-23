#!/usr/bin/env node
/**
 * Provision the Data Security Platform namespace.
 *
 *   COMPOSE_API=http://127.0.0.1:3334/api/compose \
 *   COMPOSE_DSN='postgres://postgres@127.0.0.1:5432/test12?sslmode=disable' \
 *   TOKEN=... \
 *   node apply.mjs
 *
 * Idempotent by handle. Discovery calls the local catalog agent (agents/dspm/discover.mjs).
 */
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { writeFileSync } from 'node:fs'

import {
  mintToken, detectBase, apiFactory, ensureNamespace, ensureModule, ensureChart,
  ensurePage, ensureRuleChain, parentPages, doughnutChart,
  selectOptions, field, recordRel, userField, dateField, numberField, boolSwitch,
  block, recordList, recordBlock, metricBlock, metricItem, ruleChain, organizer,
  withBlockIDs, pageIcon, setOf,
} from '../../invest/compose/helpers.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

const danger = { backgroundColor: 'danger', textColor: 'white' }
const warning = { backgroundColor: 'warning', textColor: 'dark' }
const info = { backgroundColor: 'info', textColor: 'white' }
const success = { backgroundColor: 'success', textColor: 'white' }
const secondary = { backgroundColor: 'secondary', textColor: 'white' }
const light = { backgroundColor: 'light', textColor: 'dark' }

const SEVERITY = selectOptions([
  ['critical', 'Критическая', danger],
  ['high', 'Высокая', warning],
  ['medium', 'Средняя', info],
  ['low', 'Низкая', secondary],
])

const SENSITIVITY = selectOptions([
  ['public', 'Публичная', success],
  ['internal', 'Внутренняя', info],
  ['confidential', 'Конфиденциальная', warning],
  ['restricted', 'Ограниченная', danger],
])

const CRITICALITY = selectOptions([
  ['low', 'Низкая', secondary],
  ['medium', 'Средняя', info],
  ['high', 'Высокая', warning],
  ['critical', 'Критическая', danger],
])

function sel (pairs) {
  return selectOptions(pairs)
}

function dt (name, label) {
  return dateField(name, label, { onlyDate: false })
}

function str (name, label, required = false) {
  return field(name, label, 'String', { required })
}

function progressBlock (title, xywh, opts) {
  const empty = { default: 0, moduleID: '', filter: '', field: '', operation: '' }
  return block('Progress', title, xywh, {
    value: opts.value,
    minValue: opts.minValue || { ...empty, default: 0 },
    maxValue: opts.maxValue || { ...empty, default: 100 },
    display: {
      showValue: true,
      showRelative: true,
      showProgress: true,
      animated: true,
      variant: 'info',
      thresholds: [],
    },
    refreshRate: 15,
    showRefresh: true,
    magnifyOption: '',
  })
}

function connectorFields () {
  return [
    str('name', 'Название', true),
    field('kind', 'Тип', 'Select', { options: sel([['cloud', 'Облако'], ['saas', 'SaaS'], ['database', 'СУБД'], ['warehouse', 'Warehouse'], ['lake', 'Lake'], ['fileshare', 'Файлы'], ['idp', 'IdP']]) }),
    str('provider', 'Провайдер'),
    field('environment', 'Среда', 'Select', { options: sel([['prod', 'Prod', danger], ['stage', 'Stage', warning], ['dev', 'Dev', info]]) }),
    field('mode', 'Режим', 'Select', { options: sel([['monitor', 'Наблюдение', info], ['enforce', 'Исполнение', warning]]) }),
    field('status', 'Статус', 'Select', { options: sel([['pending', 'Ожидает', light], ['connected', 'Подключён', success], ['error', 'Ошибка', danger], ['disabled', 'Выключен', secondary]]) }),
    str('secret_ref', 'Ссылка на секрет'),
    dt('last_scan', 'Последнее сканирование'),
    str('notes', 'Заметки'),
  ]
}

function scanFields (connectors) {
  return [
    recordRel('connector', 'Коннектор', connectors, 'name', ['name', 'provider']),
    field('kind', 'Тип', 'Select', { options: sel([['discovery', 'Обнаружение'], ['classification', 'Классификация'], ['access_sync', 'Синхронизация доступа'], ['score', 'Оценка']]) }),
    field('status', 'Статус', 'Select', { options: sel([['pending', 'Ожидает', light], ['running', 'Выполняется', info], ['completed', 'Готово', success], ['failed', 'Ошибка', danger]]) }),
    numberField('progress', 'Прогресс', { precision: 0, display: 'progress', min: 0, max: 100 }),
    numberField('found', 'Найдено', { precision: 0 }),
    dt('started_at', 'Старт'),
    dt('finished_at', 'Финиш'),
    str('error', 'Ошибка'),
  ]
}

function storeFields (connectors) {
  return [
    str('name', 'Название', true),
    recordRel('connector', 'Коннектор', connectors, 'name', ['name', 'provider']),
    field('kind', 'Тип', 'Select', { options: sel([['bucket', 'Bucket'], ['database', 'База'], ['warehouse', 'Warehouse'], ['lake', 'Lake'], ['site', 'Сайт'], ['mailbox', 'Почта'], ['fileshare', 'Файлы'], ['saas_app', 'SaaS']]) }),
    str('location', 'Расположение'),
    str('residency', 'Резидентность'),
    userField('owner', 'Владелец'),
    field('criticality', 'Критичность', 'Select', { options: CRITICALITY }),
    field('exposure', 'Экспозиция', 'Select', { options: sel([['private', 'Закрытое', success], ['internal', 'Внутреннее', info], ['org_wide', 'На организацию', warning], ['public', 'Публичное', danger]]) }),
    field('sensitivity', 'Чувствительность', 'Select', { options: SENSITIVITY }),
    str('external_id', 'Внешний ID'),
    str('username', 'Логин'),
    str('password', 'Пароль'),
    str('database', 'База'),
    dt('last_seen', 'Видели'),
  ]
}

function datasetFields (stores, policies) {
  return [
    str('name', 'Название', true),
    recordRel('data_store', 'Хранилище', stores, 'name', ['name', 'location']),
    str('path', 'Путь'),
    str('sample', 'Фрагмент'),
    userField('owner', 'Владелец'),
    field('purpose', 'Назначение', 'Select', { options: sel([['production', 'Прод'], ['analytics', 'Аналитика'], ['ai', 'AI'], ['archive', 'Архив']]) }),
    field('criticality', 'Критичность', 'Select', { options: CRITICALITY }),
    recordRel('policy', 'Политика по умолчанию', policies, 'name', ['name']),
    field('status', 'Статус', 'Select', { options: sel([['active', 'Активен', success], ['shadow', 'Теневой', warning], ['retired', 'Выведен', secondary]]) }),
  ]
}

function classFields () {
  return [
    str('code', 'Код', true),
    str('name', 'Название', true),
    field('sensitivity', 'Чувствительность', 'Select', { options: SENSITIVITY }),
    str('regulations', 'Регламенты'),
    boolSwitch('builtin', 'Встроенный'),
    str('description', 'Описание'),
  ]
}

function maskFields () {
  return [
    str('name', 'Название', true),
    field('method', 'Метод', 'Select', { options: sel([['redact', 'Скрыть'], ['partial', 'Частично'], ['hash', 'Хеш'], ['nullify', 'Очистить'], ['tokenize', 'Токен']]) }),
    str('pattern', 'Шаблон'),
    str('description', 'Описание'),
  ]
}

function filterFields () {
  return [
    str('name', 'Название', true),
    str('expression', 'Выражение'),
    str('description', 'Описание'),
  ]
}

function policyFields (classes, masks, filters) {
  return [
    str('name', 'Название', true),
    field('mode', 'Режим', 'Select', { options: sel([['monitor', 'Наблюдение', info], ['alert', 'Предупреждение', warning], ['mask', 'Маска', warning], ['block', 'Блок', danger], ['revoke', 'Отзыв', danger]]) }),
    recordRel('data_class', 'Класс данных', classes, 'name', ['name', 'code']),
    field('identity_kind', 'Тип идентичности', 'Select', { options: sel([['any', 'Любая'], ['human', 'Человек'], ['service', 'Сервис'], ['ai_agent', 'AI-агент']]) }),
    recordRel('masking_profile', 'Профиль маски', masks, 'name', ['name']),
    recordRel('row_filter', 'Фильтр строк', filters, 'name', ['name']),
    boolSwitch('enabled', 'Включена'),
    str('description', 'Описание'),
  ]
}

function bindingFields (policies, datasets) {
  return [
    recordRel('policy', 'Политика', policies, 'name', ['name'], true),
    recordRel('dataset', 'Датасет', datasets, 'name', ['name', 'path'], true),
    boolSwitch('enabled', 'Включена'),
  ]
}

function identityFields () {
  return [
    str('name', 'Имя', true),
    field('kind', 'Тип', 'Select', { options: sel([['human', 'Человек'], ['service', 'Сервис'], ['ai_agent', 'AI-агент']]) }),
    field('source', 'Источник', 'Select', { options: sel([['idp', 'IdP'], ['directory', 'Каталог'], ['local', 'Локальная'], ['discovered', 'Обнаружена']]) }),
    str('external_id', 'Внешний ID'),
    str('email', 'Почта'),
    str('department', 'Подразделение'),
    field('status', 'Статус', 'Select', { options: sel([['active', 'Активна', success], ['stale', 'Устаревшая', warning], ['ghost', 'Призрак', danger], ['disabled', 'Отключена', secondary]]) }),
    dt('last_seen', 'Видели'),
  ]
}

function entitlementFields (identities, datasets) {
  return [
    recordRel('identity', 'Идентичность', identities, 'name', ['name', 'email'], true),
    recordRel('dataset', 'Датасет', datasets, 'name', ['name', 'path'], true),
    field('access_level', 'Уровень', 'Select', { options: sel([['read', 'Чтение'], ['write', 'Запись', warning], ['admin', 'Админ', danger], ['share', 'Шаринг', warning]]) }),
    field('origin', 'Источник права', 'Select', { options: sel([['direct', 'Прямое'], ['group', 'Группа'], ['role', 'Роль'], ['link', 'Ссылка']]) }),
    str('origin_name', 'Имя источника'),
    boolSwitch('used', 'Использовалось'),
    dt('last_used', 'Последнее использование'),
    field('status', 'Статус', 'Select', { options: sel([['active', 'Активно', success], ['unused', 'Не используется', warning], ['excessive', 'Избыточно', danger], ['revoked', 'Отозвано', secondary]]) }),
  ]
}

function classificationFields (datasets, classes) {
  return [
    recordRel('dataset', 'Датасет', datasets, 'name', ['name', 'path'], true),
    recordRel('data_class', 'Класс', classes, 'name', ['name', 'code'], true),
    numberField('confidence', 'Точность', { precision: 2, min: 0, max: 1 }),
    field('source', 'Источник', 'Select', { options: sel([['rule', 'Правило'], ['model', 'Модель'], ['manual', 'Вручную']]) }),
    field('status', 'Статус', 'Select', { options: sel([['pending', 'На проверке', warning], ['confirmed', 'Подтверждён', success], ['rejected', 'Отклонён', secondary]]) }),
    userField('reviewer', 'Проверил'),
    dt('reviewed_at', 'Когда проверили'),
    str('location_hint', 'Где найдено'),
    str('masked_sample', 'Маскированный фрагмент'),
  ]
}

function requestFields (identities, datasets, policies) {
  return [
    userField('requester', 'Заявитель'),
    recordRel('identity', 'Идентичность', identities, 'name', ['name', 'email']),
    recordRel('dataset', 'Датасет', datasets, 'name', ['name', 'path'], true),
    str('justification', 'Обоснование'),
    dt('starts_at', 'Начало'),
    dt('ends_at', 'Конец'),
    userField('approver', 'Согласующий'),
    field('status', 'Статус', 'Select', { options: sel([['draft', 'Черновик', light], ['pending', 'Ожидает', warning], ['approved', 'Согласована', success], ['rejected', 'Отклонена', danger], ['expired', 'Истекла', secondary], ['revoked', 'Отозвана', secondary]]) }),
    recordRel('policy', 'Политика', policies, 'name', ['name']),
  ]
}

function findingFields (stores, datasets, identities, policies) {
  return [
    str('title', 'Заголовок', true),
    field('kind', 'Тип', 'Select', { options: sel([['public_exposure', 'Публичная экспозиция'], ['excessive_access', 'Избыточные права'], ['misconfig', 'Конфигурация'], ['stale_link', 'Устаревшая ссылка'], ['shadow_ai', 'Теневой AI'], ['policy_violation', 'Нарушение политики'], ['unused_entitlement', 'Неиспользуемое право']]) }),
    field('severity', 'Серьёзность', 'Select', { options: SEVERITY }),
    field('status', 'Статус', 'Select', { options: sel([['open', 'Открыта', danger], ['with_owner', 'У владельца', warning], ['in_progress', 'В работе', info], ['closed', 'Закрыта', success], ['accepted', 'Риск принят', secondary]]) }),
    recordRel('data_store', 'Хранилище', stores, 'name', ['name']),
    recordRel('dataset', 'Датасет', datasets, 'name', ['name', 'path']),
    recordRel('identity', 'Идентичность', identities, 'name', ['name', 'email']),
    recordRel('policy', 'Политика', policies, 'name', ['name']),
    userField('owner', 'Владелец'),
    dt('detected_at', 'Обнаружена'),
    str('summary', 'Суть'),
  ]
}

function remediationFields (findings) {
  return [
    recordRel('finding', 'Находка', findings, 'title', ['title', 'kind'], true),
    field('action', 'Действие', 'Select', { options: sel([['revoke', 'Отозвать'], ['remove_link', 'Снять ссылку'], ['mask', 'Замаскировать'], ['label', 'Поставить метку'], ['delete', 'Удалить'], ['ticket', 'Завести задачу'], ['notify_owner', 'Владельцу']]) }),
    field('status', 'Статус', 'Select', { options: sel([['proposed', 'Предложено', light], ['approved', 'Подтверждено', info], ['done', 'Сделано', success], ['failed', 'Ошибка', danger], ['rolled_back', 'Откат', warning]]) }),
    userField('actor', 'Кто выполнил'),
    boolSwitch('reversible', 'Можно откатить'),
    str('result', 'Результат'),
    dt('executed_at', 'Когда'),
  ]
}

function eventFields (identities, datasets, classes, policies) {
  return [
    dt('occurred_at', 'Когда'),
    recordRel('identity', 'Идентичность', identities, 'name', ['name', 'email']),
    recordRel('dataset', 'Датасет', datasets, 'name', ['name', 'path']),
    recordRel('data_class', 'Класс', classes, 'name', ['name', 'code']),
    str('tool', 'Инструмент'),
    recordRel('policy', 'Политика', policies, 'name', ['name']),
    field('outcome', 'Исход', 'Select', { options: sel([['allow', 'Разрешено', success], ['mask', 'Замаскировано', warning], ['block', 'Заблокировано', danger], ['alert', 'Предупреждение', info]]) }),
    str('summary', 'Суть'),
  ]
}

function alertFields (identities, datasets, events) {
  return [
    str('title', 'Заголовок', true),
    field('severity', 'Серьёзность', 'Select', { options: SEVERITY }),
    field('status', 'Статус', 'Select', { options: sel([['open', 'Открыт', danger], ['investigating', 'Расследование', warning], ['closed', 'Закрыт', success], ['false_positive', 'Ложный', secondary]]) }),
    recordRel('identity', 'Идентичность', identities, 'name', ['name', 'email']),
    recordRel('dataset', 'Датасет', datasets, 'name', ['name', 'path']),
    recordRel('activity_event', 'Событие', events, 'summary', ['summary']),
    dt('detected_at', 'Обнаружен'),
    str('summary', 'Суть'),
  ]
}

function aiFields (identities) {
  return [
    str('name', 'Название', true),
    field('kind', 'Тип', 'Select', { options: sel([['copilot', 'Copilot'], ['agent', 'Агент'], ['model', 'Модель'], ['knowledge_base', 'База знаний'], ['public_tool', 'Публичный инструмент']]) }),
    boolSwitch('sanctioned', 'Санкционирован'),
    recordRel('identity', 'Идентичность', identities, 'name', ['name', 'email']),
    field('status', 'Статус', 'Select', { options: sel([['discovered', 'Обнаружен', warning], ['approved', 'Одобрен', success], ['blocked', 'Заблокирован', danger]]) }),
    userField('owner', 'Владелец'),
    str('notes', 'Заметки'),
  ]
}

function controlFields (classes, policies) {
  return [
    field('framework', 'Регламент', 'Select', { options: sel([['152-fz', '152-ФЗ'], ['gdpr', 'GDPR'], ['hipaa', 'HIPAA'], ['pci-dss', 'PCI DSS'], ['soc2', 'SOC 2'], ['sox', 'SOX'], ['iso27001', 'ISO 27001']]) }),
    str('name', 'Название', true),
    recordRel('data_class', 'Класс', classes, 'name', ['name', 'code']),
    recordRel('policy', 'Политика', policies, 'name', ['name']),
    field('status', 'Покрытие', 'Select', { options: sel([['covered', 'Покрыт', success], ['partial', 'Частично', warning], ['gap', 'Пробел', danger]]) }),
    str('evidence', 'Доказательство'),
  ]
}

function crudUpdate (id, label, ns, moduleID, handle, fields) {
  return {
    id,
    type: 'crud',
    label,
    config: {
      operation: 'update',
      namespaceID: ns,
      moduleID,
      moduleHandle: handle,
      recordID: '{{recordID}}',
      fields,
      omitEmpty: true,
    },
  }
}

function crudCreateScan (id, label, ns, scans, kind) {
  return {
    id,
    type: 'crud',
    label,
    config: {
      operation: 'create',
      namespaceID: ns,
      moduleID: scans,
      moduleHandle: 'scans',
      fields: {
        connector: '{{recordID}}',
        kind,
        status: 'pending',
        progress: '0',
        found: '0',
      },
    },
  }
}

function buildDiscoveryChain (ns, modules, agentUrl) {
  const stores = String(modules.data_stores)
  const datasets = String(modules.datasets)
  const scans = String(modules.scans)
  const connectors = String(modules.connectors)
  return {
    id: 'dspm-run-discovery',
    name: 'DSPM: обнаружение',
    description: 'Читает каталог PostgreSQL и записывает хранилища и датасеты.',
    entryNode: 'http_discover',
    namespaceID: ns,
    nodes: [
      {
        id: 'http_discover',
        type: 'http',
        label: 'Каталог PostgreSQL',
        config: { url: `${agentUrl}/discover`, method: 'GET', timeout: 120 },
      },
      {
        id: 'upsert_connector',
        type: 'crud.upsert',
        label: 'Коннектор',
        config: {
          namespaceID: ns,
          moduleID: connectors,
          moduleHandle: 'connectors',
          matchBy: ['name'],
          resultVar: 'connectorRecordID',
          fields: {
            name: '{{connectorName}}',
            kind: 'database',
            provider: '{{provider}}',
            environment: 'dev',
            mode: 'monitor',
            status: 'connected',
            secret_ref: 'env:DSPM_PG_DSN',
            last_scan: '{{scannedAt}}',
          },
        },
      },
      {
        id: 'record_scan',
        type: 'crud',
        label: 'Задание',
        config: {
          operation: 'create',
          namespaceID: ns,
          moduleID: scans,
          moduleHandle: 'scans',
          fields: {
            connector: '{{connectorRecordID}}',
            kind: 'discovery',
            status: 'completed',
            progress: '100',
            found: '{{found}}',
            started_at: '{{scannedAt}}',
            finished_at: '{{scannedAt}}',
          },
        },
      },
      {
        id: 'foreach_stores',
        type: 'foreach',
        label: 'Каждая база',
        config: { items: 'items', itemVar: 'item', maxItems: 80 },
      },
      {
        id: 'upsert_store',
        type: 'crud.upsert',
        label: 'Хранилище',
        config: {
          namespaceID: ns,
          moduleID: stores,
          moduleHandle: 'data_stores',
          matchBy: ['external_id'],
          resultVar: 'storeRecordID',
          omitEmpty: true,
          fields: {
            name: '{{item.name}}',
            connector: '{{connectorRecordID}}',
            kind: '{{item.kind}}',
            location: '{{item.location}}',
            residency: '{{item.residency}}',
            exposure: '{{item.exposure}}',
            sensitivity: '{{item.sensitivity}}',
            external_id: '{{item.externalId}}',
            last_seen: '{{item.lastSeen}}',
          },
        },
      },
      {
        id: 'foreach_datasets',
        type: 'foreach',
        label: 'Каждая таблица',
        config: { items: 'item.datasets', itemVar: 'ds', maxItems: 20 },
      },
      {
        id: 'upsert_dataset',
        type: 'crud.upsert',
        label: 'Датасет',
        config: {
          namespaceID: ns,
          moduleID: datasets,
          moduleHandle: 'datasets',
          matchBy: ['data_store', 'path'],
          matchAll: true,
          omitEmpty: true,
          continueOnError: true,
          fields: {
            name: '{{ds.name}}',
            data_store: '{{storeRecordID}}',
            path: '{{ds.path}}',
            sample: '{{ds.sample}}',
            purpose: '{{ds.purpose}}',
            status: '{{ds.status}}',
          },
        },
      },
    ],
    edges: [
      { from: 'http_discover', to: 'upsert_connector' },
      { from: 'upsert_connector', to: 'record_scan' },
      { from: 'record_scan', to: 'foreach_stores' },
      { from: 'foreach_stores', to: 'upsert_store' },
      { from: 'foreach_stores', to: 'foreach_datasets' },
      { from: 'foreach_datasets', to: 'upsert_dataset' },
    ],
  }
}

function buildInspectChain (ns, modules, agentUrl) {
  const datasets = String(modules.datasets)
  return {
    id: 'dspm-inspect-store',
    name: 'DSPM: инспекция актива',
    description: 'Подключается к активу логином и паролем записи и сохраняет маскированные фрагменты.',
    entryNode: 'http_inspect',
    namespaceID: ns,
    nodes: [
      {
        id: 'http_inspect',
        type: 'http',
        label: 'Инспекция',
        config: {
          url: `${agentUrl}/inspect`,
          method: 'POST',
          timeout: 60,
          body: '{"location":"{{location}}","kind":"{{kind}}","username":"{{username}}","password":"{{password}}","database":"{{database}}"}',
        },
      },
      {
        id: 'foreach_samples',
        type: 'foreach',
        label: 'Каждый фрагмент',
        config: { items: 'items', itemVar: 'ds', maxItems: 12 },
      },
      {
        id: 'upsert_sample',
        type: 'crud.upsert',
        label: 'Фрагмент датасета',
        config: {
          namespaceID: ns,
          moduleID: datasets,
          moduleHandle: 'datasets',
          matchBy: ['data_store', 'path'],
          matchAll: true,
          omitEmpty: true,
          continueOnError: true,
          fields: {
            name: '{{ds.name}}',
            data_store: '{{recordID}}',
            path: '{{ds.path}}',
            sample: '{{ds.sample}}',
            purpose: '{{ds.purpose}}',
            status: 'shadow',
          },
        },
      },
    ],
    edges: [
      { from: 'http_inspect', to: 'foreach_samples' },
      { from: 'foreach_samples', to: 'upsert_sample' },
    ],
  }
}

function buildRuleChains ({ nsID, modules, agentUrl }) {
  const ns = String(nsID)
  const scans = String(modules.scans)
  const chain = (id, name, description, node) => ({
    id,
    name,
    description,
    entryNode: node.id,
    nodes: [node],
    edges: [],
    namespaceID: ns,
  })
  return [
    buildDiscoveryChain(ns, modules, agentUrl),
    buildInspectChain(ns, modules, agentUrl),
    chain('dspm-run-classification', 'DSPM: классификация', 'Создаёт задание classification. Агент ещё не подключён.', crudCreateScan('job', 'Задание', ns, scans, 'classification')),
    chain('dspm-sync-access', 'DSPM: синхронизация доступа', 'Создаёт задание access_sync. Агент ещё не подключён.', crudCreateScan('job', 'Задание', ns, scans, 'access_sync')),
    chain('dspm-score-findings', 'DSPM: оценка находок', 'Создаёт задание score. Агент ещё не подключён.', crudCreateScan('job', 'Задание', ns, scans, 'score')),
    chain('dspm-confirm-class', 'DSPM: подтвердить класс', 'Ставит классификации статус confirmed.', crudUpdate('upd', 'Подтвердить', ns, String(modules.classifications), 'classifications', { status: 'confirmed' })),
    chain('dspm-approve-request', 'DSPM: согласовать заявку', 'Ставит заявке статус approved.', crudUpdate('upd', 'Согласовать', ns, String(modules.access_requests), 'access_requests', { status: 'approved' })),
    chain('dspm-reject-request', 'DSPM: отклонить заявку', 'Ставит заявке статус rejected.', crudUpdate('upd', 'Отклонить', ns, String(modules.access_requests), 'access_requests', { status: 'rejected' })),
    chain('dspm-revoke-request', 'DSPM: отозвать заявку', 'Ставит заявке статус revoked.', crudUpdate('upd', 'Отозвать', ns, String(modules.access_requests), 'access_requests', { status: 'revoked' })),
    chain('dspm-route-owner', 'DSPM: находка владельцу', 'Ставит находке статус with_owner.', crudUpdate('upd', 'Владельцу', ns, String(modules.findings), 'findings', { status: 'with_owner' })),
    chain('dspm-remediate', 'DSPM: взять в работу', 'Ставит находке статус in_progress. Исполнение на источнике делает агент.', crudUpdate('upd', 'В работу', ns, String(modules.findings), 'findings', { status: 'in_progress' })),
    chain('dspm-accept-risk', 'DSPM: принять риск', 'Ставит находке статус accepted.', crudUpdate('upd', 'Принять риск', ns, String(modules.findings), 'findings', { status: 'accepted' })),
  ]
}

function buildPages (m, charts) {
  const list = (title, handle, weight, icon, moduleID, fields, extra = {}) => ({
    title,
    handle,
    visible: true,
    weight,
    config: pageIcon(icon),
    blocks: withBlockIDs([
      ...(extra.before || []),
      recordList(extra.listTitle || title, [0, extra.listY || 0, 48, extra.listH || 40], moduleID, fields, extra.list || {}),
      ...(extra.after || []),
    ]),
  })
  const card = (title, handle, moduleID, fields, blocks, weight) => ({
    title,
    handle,
    moduleID: String(moduleID),
    visible: false,
    weight,
    config: pageIcon('fas file'),
    blocks: withBlockIDs([
      recordBlock(title, [0, 0, 32, 26], fields, { fieldRoles: { name: 'title', title: 'title', status: 'badge', severity: 'badge' } }),
      ...blocks,
    ]),
  })

  return [
    {
      title: 'Сводка',
      handle: 'dashboard',
      visible: true,
      weight: 0,
      config: pageIcon('fas shield-halved'),
      blocks: withBlockIDs([
        metricBlock('Сводка', [0, 0, 48, 14], [
          metricItem('Хранилища', m.data_stores, '', { role: 'hero', color: '#2e59d9', fontSize: '28' }),
          metricItem('Ограниченные', m.data_stores, "sensitivity = 'restricted'", { role: 'balloon', color: '#e74a3b' }),
          metricItem('Открытые находки', m.findings, "status = 'open'", { role: 'balloon', color: '#e74a3b' }),
          metricItem('Заявки', m.access_requests, "status = 'pending'", { role: 'meta', color: '#f6c23e' }),
        ]),
        ruleChain('Обнаружение', [0, 14, 16, 10], { chainID: 'dspm-run-discovery', label: 'Обнаружение', icon: 'search' }),
        ruleChain('Синхронизация доступа', [16, 14, 16, 10], { chainID: 'dspm-sync-access', label: 'Синхронизация доступа', icon: 'key' }),
        ruleChain('Оценка находок', [32, 14, 16, 10], { chainID: 'dspm-score-findings', label: 'Оценка находок', icon: 'exclamation-triangle' }),
        block('Chart', 'Находки по серьёзности', [0, 24, 16, 18], { chartID: String(charts.findingsBySeverity) }),
        block('Chart', 'Хранилища по экспозиции', [16, 24, 16, 18], { chartID: String(charts.storesByExposure) }),
        block('Chart', 'Заявки по статусу', [32, 24, 16, 18], { chartID: String(charts.requestsByStatus) }),
        recordList('Обнаруженные хранилища', [0, 42, 48, 18], m.data_stores, ['name', 'kind', 'location', 'exposure', 'last_seen'], { hideAddButton: true }),
        recordList('Открытые находки', [0, 60, 24, 22], m.findings, ['title', 'kind', 'severity', 'status', 'owner'], { prefilter: "status = 'open'", hideAddButton: true }),
        recordList('Заявки на согласовании', [24, 60, 24, 22], m.access_requests, ['requester', 'dataset', 'status', 'ends_at'], { prefilter: "status = 'pending'", hideAddButton: true }),
      ]),
    },
    list('Инвентарь', 'inventory', 10, 'fas database', m.data_stores, ['name', 'kind', 'residency', 'exposure', 'sensitivity', 'owner', 'last_seen']),
    card('Хранилище', 'store', m.data_stores, ['name', 'connector', 'kind', 'location', 'username', 'password', 'database', 'residency', 'owner', 'criticality', 'exposure', 'sensitivity', 'external_id', 'last_seen'], [
      ruleChain('Инспекция', [32, 0, 16, 12], { chainID: 'dspm-inspect-store', label: 'Инспектировать', icon: 'search' }),
      recordList('Датасеты', [0, 26, 32, 22], m.datasets, ['name', 'path', 'sample', 'purpose', 'status', 'owner'], { refField: 'data_store' }),
      recordList('Находки', [32, 0, 16, 48], m.findings, ['title', 'severity', 'status'], { refField: 'data_store' }),
    ], 11),
    list('Датасеты', 'datasets', 20, 'fas table', m.datasets, ['name', 'data_store', 'path', 'purpose', 'criticality', 'policy', 'status']),
    {
      title: 'Датасет',
      handle: 'dataset',
      moduleID: String(m.datasets),
      visible: false,
      weight: 21,
      config: pageIcon('fas table'),
      blocks: withBlockIDs([
        recordBlock('Датасет', [0, 0, 32, 24], ['name', 'data_store', 'path', 'owner', 'purpose', 'criticality', 'policy', 'status'], { fieldRoles: { name: 'title', status: 'badge', purpose: 'badge' } }),
        ruleChain('Классификация', [32, 0, 16, 12], { chainID: 'dspm-run-classification', label: 'Классификация', icon: 'tag' }),
        ruleChain('Подтвердить класс', [32, 12, 16, 12], { chainID: 'dspm-confirm-class', label: 'Подтвердить класс', icon: 'check' }),
        recordList('Классы', [0, 24, 24, 22], m.classifications, ['data_class', 'confidence', 'source', 'status'], { refField: 'dataset' }),
        recordList('Кто имеет доступ', [24, 24, 24, 22], m.entitlements, ['identity', 'access_level', 'origin', 'used', 'status'], { refField: 'dataset' }),
        recordList('Политики', [0, 46, 48, 18], m.policy_bindings, ['policy', 'enabled'], { refField: 'dataset' }),
      ]),
    },
    {
      title: 'Классификация',
      handle: 'classification',
      visible: true,
      weight: 30,
      config: pageIcon('fas tag'),
      blocks: withBlockIDs([
        recordList('На проверке', [0, 0, 48, 28], m.classifications, ['dataset', 'data_class', 'confidence', 'source', 'status', 'reviewer'], { prefilter: "status = 'pending'" }),
        recordList('Классы', [0, 28, 48, 24], m.data_classes, ['code', 'name', 'sensitivity', 'regulations', 'builtin']),
      ]),
    },
    card('Классификация', 'classification-item', m.classifications, ['dataset', 'data_class', 'confidence', 'source', 'status', 'reviewer', 'reviewed_at', 'location_hint', 'masked_sample'], [
      ruleChain('Подтвердить', [32, 0, 16, 12], { chainID: 'dspm-confirm-class', label: 'Подтвердить', icon: 'check' }),
    ], 31),
    list('Доступ', 'access', 40, 'fas key', m.entitlements, ['identity', 'dataset', 'access_level', 'origin', 'used', 'last_used', 'status']),
    card('Право', 'entitlement', m.entitlements, ['identity', 'dataset', 'access_level', 'origin', 'origin_name', 'used', 'last_used', 'status'], [], 41),
    list('Идентичности', 'identities', 50, 'fas users', m.identities, ['name', 'kind', 'source', 'department', 'status', 'last_seen']),
    {
      title: 'Идентичность',
      handle: 'identity',
      moduleID: String(m.identities),
      visible: false,
      weight: 51,
      config: pageIcon('fas user'),
      blocks: withBlockIDs([
        recordBlock('Идентичность', [0, 0, 32, 22], ['name', 'kind', 'source', 'external_id', 'email', 'department', 'status', 'last_seen'], { fieldRoles: { name: 'title', status: 'badge', kind: 'badge' } }),
        recordList('Куда есть доступ', [0, 22, 48, 24], m.entitlements, ['dataset', 'access_level', 'origin', 'used', 'status'], { refField: 'identity' }),
        recordList('События', [0, 46, 48, 20], m.activity_events, ['occurred_at', 'dataset', 'tool', 'outcome'], { refField: 'identity' }),
      ]),
    },
    {
      title: 'Политики',
      handle: 'policies',
      visible: true,
      weight: 60,
      config: pageIcon('fas scale-balanced'),
      blocks: withBlockIDs([
        recordList('Политики', [0, 0, 48, 28], m.policies, ['name', 'mode', 'data_class', 'identity_kind', 'enabled']),
        recordList('Профили маски', [0, 28, 24, 22], m.masking_profiles, ['name', 'method']),
        recordList('Фильтры строк', [24, 28, 24, 22], m.row_filters, ['name', 'expression']),
      ]),
    },
    {
      title: 'Политика',
      handle: 'policy',
      moduleID: String(m.policies),
      visible: false,
      weight: 61,
      config: pageIcon('fas scale-balanced'),
      blocks: withBlockIDs([
        recordBlock('Политика', [0, 0, 32, 26], ['name', 'mode', 'data_class', 'identity_kind', 'masking_profile', 'row_filter', 'enabled', 'description'], { fieldRoles: { name: 'title', mode: 'badge' } }),
        recordList('Где действует', [0, 26, 48, 22], m.policy_bindings, ['dataset', 'enabled'], { refField: 'policy' }),
      ]),
    },
    {
      title: 'Заявки',
      handle: 'requests',
      visible: true,
      weight: 70,
      config: pageIcon('fas clock'),
      blocks: withBlockIDs([
        organizer('Ожидают', [0, 0, 16, 28], m.access_requests, { labelField: 'justification', descriptionField: 'dataset', groupField: 'status', group: 'pending' }),
        organizer('Согласованы', [16, 0, 16, 28], m.access_requests, { labelField: 'justification', descriptionField: 'dataset', groupField: 'status', group: 'approved' }),
        organizer('Отклонены', [32, 0, 16, 28], m.access_requests, { labelField: 'justification', descriptionField: 'dataset', groupField: 'status', group: 'rejected' }),
        recordList('Все заявки', [0, 28, 48, 24], m.access_requests, ['requester', 'dataset', 'status', 'starts_at', 'ends_at', 'approver']),
      ]),
    },
    {
      title: 'Заявка',
      handle: 'request',
      moduleID: String(m.access_requests),
      visible: false,
      weight: 71,
      config: pageIcon('fas clock'),
      blocks: withBlockIDs([
        recordBlock('Заявка', [0, 0, 32, 28], ['requester', 'identity', 'dataset', 'justification', 'starts_at', 'ends_at', 'approver', 'status', 'policy'], { fieldRoles: { justification: 'title', status: 'badge' } }),
        ruleChain('Согласовать', [32, 0, 16, 10], { chainID: 'dspm-approve-request', label: 'Согласовать', icon: 'check', variant: 'success' }),
        ruleChain('Отклонить', [32, 10, 16, 10], { chainID: 'dspm-reject-request', label: 'Отклонить', icon: 'times', variant: 'danger' }),
        ruleChain('Отозвать', [32, 20, 16, 10], { chainID: 'dspm-revoke-request', label: 'Отозвать', icon: 'lock', variant: 'secondary' }),
      ]),
    },
    {
      title: 'Находки',
      handle: 'findings',
      visible: true,
      weight: 80,
      config: pageIcon('fas exclamation-triangle'),
      blocks: withBlockIDs([
        organizer('Открыты', [0, 0, 12, 28], m.findings, { labelField: 'title', descriptionField: 'kind', groupField: 'status', group: 'open' }),
        organizer('У владельца', [12, 0, 12, 28], m.findings, { labelField: 'title', descriptionField: 'kind', groupField: 'status', group: 'with_owner' }),
        organizer('В работе', [24, 0, 12, 28], m.findings, { labelField: 'title', descriptionField: 'kind', groupField: 'status', group: 'in_progress' }),
        organizer('Закрыты', [36, 0, 12, 28], m.findings, { labelField: 'title', descriptionField: 'kind', groupField: 'status', group: 'closed' }),
        recordList('Все находки', [0, 28, 48, 24], m.findings, ['title', 'kind', 'severity', 'status', 'dataset', 'owner', 'detected_at']),
      ]),
    },
    {
      title: 'Находка',
      handle: 'finding',
      moduleID: String(m.findings),
      visible: false,
      weight: 81,
      config: pageIcon('fas exclamation-triangle'),
      blocks: withBlockIDs([
        recordBlock('Находка', [0, 0, 32, 28], ['title', 'kind', 'severity', 'status', 'data_store', 'dataset', 'identity', 'policy', 'owner', 'detected_at', 'summary'], { fieldRoles: { title: 'title', severity: 'badge', status: 'badge' } }),
        ruleChain('Владельцу', [32, 0, 16, 10], { chainID: 'dspm-route-owner', label: 'Владельцу', icon: 'user' }),
        ruleChain('Исправить', [32, 10, 16, 10], { chainID: 'dspm-remediate', label: 'В работу', icon: 'wrench' }),
        ruleChain('Принять риск', [32, 20, 16, 10], { chainID: 'dspm-accept-risk', label: 'Принять риск', icon: 'check', variant: 'secondary' }),
        recordList('Действия', [0, 28, 48, 20], m.remediations, ['action', 'status', 'actor', 'reversible', 'executed_at'], { refField: 'finding' }),
      ]),
    },
    {
      title: 'Активность',
      handle: 'activity',
      visible: true,
      weight: 90,
      config: pageIcon('fas timeline'),
      blocks: withBlockIDs([
        recordList('События', [0, 0, 48, 28], m.activity_events, ['occurred_at', 'identity', 'dataset', 'tool', 'policy', 'outcome']),
        recordList('Алерты', [0, 28, 48, 24], m.alerts, ['title', 'severity', 'status', 'identity', 'dataset', 'detected_at']),
      ]),
    },
    card('Алерт', 'alert', m.alerts, ['title', 'severity', 'status', 'identity', 'dataset', 'activity_event', 'detected_at', 'summary'], [], 91),
    list('AI', 'ai', 100, 'fas robot', m.ai_assets, ['name', 'kind', 'sanctioned', 'identity', 'status', 'owner']),
    {
      title: 'Актив AI',
      handle: 'ai-asset',
      moduleID: String(m.ai_assets),
      visible: false,
      weight: 101,
      config: pageIcon('fas robot'),
      blocks: withBlockIDs([
        recordBlock('Актив', [0, 0, 32, 22], ['name', 'kind', 'sanctioned', 'identity', 'status', 'owner', 'notes'], { fieldRoles: { name: 'title', status: 'badge', kind: 'badge' } }),
        recordList('Доступные датасеты', [0, 22, 48, 24], m.entitlements, ['dataset', 'access_level', 'used', 'status'], { refField: 'identity' }),
      ]),
    },
    {
      title: 'Соответствие',
      handle: 'compliance',
      visible: true,
      weight: 110,
      config: pageIcon('fas clipboard-check'),
      blocks: withBlockIDs([
        block('Chart', 'Покрытие по регламентам', [0, 0, 24, 20], { chartID: String(charts.controlsByStatus) }),
        recordList('Контроли', [0, 20, 48, 32], m.compliance_controls, ['framework', 'name', 'data_class', 'policy', 'status']),
      ]),
    },
    {
      title: 'Подключения',
      handle: 'connectors',
      visible: true,
      weight: 120,
      config: pageIcon('fas plug'),
      blocks: withBlockIDs([
        ruleChain('Обнаружение', [0, 0, 16, 12], { chainID: 'dspm-run-discovery', label: 'Обнаружение', icon: 'search' }),
        progressBlock('Задания в работе', [16, 0, 32, 12], {
          value: { default: 0, moduleID: String(m.scans), filter: "status = 'running'", field: 'count', operation: 'count' },
          maxValue: { default: 1, moduleID: String(m.scans), filter: '', field: 'count', operation: 'count' },
        }),
        recordList('Коннекторы', [0, 12, 48, 24], m.connectors, ['name', 'kind', 'provider', 'environment', 'mode', 'status', 'last_scan']),
        recordList('Задания', [0, 36, 48, 22], m.scans, ['connector', 'kind', 'status', 'progress', 'found', 'started_at']),
      ]),
    },
    {
      title: 'Коннектор',
      handle: 'connector',
      moduleID: String(m.connectors),
      visible: false,
      weight: 121,
      config: pageIcon('fas plug'),
      blocks: withBlockIDs([
        recordBlock('Коннектор', [0, 0, 32, 24], ['name', 'kind', 'provider', 'environment', 'mode', 'status', 'secret_ref', 'last_scan', 'notes'], { fieldRoles: { name: 'title', status: 'badge', mode: 'badge' } }),
        ruleChain('Запустить обнаружение', [32, 0, 16, 12], { chainID: 'dspm-run-discovery', label: 'Обнаружение', icon: 'search' }),
        recordList('Задания', [0, 24, 48, 22], m.scans, ['kind', 'status', 'progress', 'found', 'started_at', 'error'], { refField: 'connector' }),
      ]),
    },
  ]
}

async function seedClasses (api, nsID, moduleID) {
  const list = setOf(await api('GET', `/namespace/${nsID}/module/${moduleID}/record/?limit=5`))
  if (list.length) {
    console.log('data classes already present, skip seed')
    return
  }
  const samples = [
    { code: 'pii', name: 'Персональные данные', sensitivity: 'restricted', regulations: '152-fz,gdpr', builtin: '1' },
    { code: 'phi', name: 'Медицинские данные', sensitivity: 'restricted', regulations: 'hipaa', builtin: '1' },
    { code: 'pci', name: 'Платёжные данные', sensitivity: 'restricted', regulations: 'pci-dss', builtin: '1' },
    { code: 'credentials', name: 'Учётные данные', sensitivity: 'restricted', regulations: '', builtin: '1' },
    { code: 'financial', name: 'Финансовые данные', sensitivity: 'confidential', regulations: 'sox', builtin: '1' },
    { code: 'ip', name: 'Интеллектуальная собственность', sensitivity: 'confidential', regulations: '', builtin: '1' },
  ]
  for (const sample of samples) {
    const values = Object.entries(sample).map(([name, value]) => ({ name, value: String(value) }))
    await api('POST', `/namespace/${nsID}/module/${moduleID}/record/`, { values })
  }
  console.log('seeded', samples.length, 'data classes')
}

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  console.log('API', base)

  const nsID = await ensureNamespace(api, {
    name: 'Data Security Platform',
    slug: 'dspm',
    meta: {
      subtitle: 'Управление безопасностью данных',
      description: 'Инвентарь, классификация, эффективный доступ, политики, заявки, находки и аудит. Исполнение на источнике делает агент.',
    },
  })

  const connectors = await ensureModule(api, nsID, { name: 'Коннекторы', handle: 'connectors', fields: connectorFields() })
  const data_classes = await ensureModule(api, nsID, { name: 'Классы данных', handle: 'data_classes', fields: classFields() })
  const masking_profiles = await ensureModule(api, nsID, { name: 'Профили маскирования', handle: 'masking_profiles', fields: maskFields() })
  const row_filters = await ensureModule(api, nsID, { name: 'Фильтры строк', handle: 'row_filters', fields: filterFields() })
  const policies = await ensureModule(api, nsID, { name: 'Политики', handle: 'policies', fields: policyFields(data_classes, masking_profiles, row_filters) })
  const data_stores = await ensureModule(api, nsID, { name: 'Хранилища', handle: 'data_stores', fields: storeFields(connectors) })
  const datasets = await ensureModule(api, nsID, { name: 'Датасеты', handle: 'datasets', fields: datasetFields(data_stores, policies) })
  const scans = await ensureModule(api, nsID, { name: 'Задания', handle: 'scans', fields: scanFields(connectors) })
  const identities = await ensureModule(api, nsID, { name: 'Идентичности', handle: 'identities', fields: identityFields() })
  const policy_bindings = await ensureModule(api, nsID, { name: 'Привязки политик', handle: 'policy_bindings', fields: bindingFields(policies, datasets) })
  const entitlements = await ensureModule(api, nsID, { name: 'Права', handle: 'entitlements', fields: entitlementFields(identities, datasets) })
  const classifications = await ensureModule(api, nsID, { name: 'Классификации', handle: 'classifications', fields: classificationFields(datasets, data_classes) })
  const access_requests = await ensureModule(api, nsID, { name: 'Заявки на доступ', handle: 'access_requests', fields: requestFields(identities, datasets, policies) })
  const findings = await ensureModule(api, nsID, { name: 'Находки', handle: 'findings', fields: findingFields(data_stores, datasets, identities, policies) })
  const remediations = await ensureModule(api, nsID, { name: 'Исправления', handle: 'remediations', fields: remediationFields(findings) })
  const activity_events = await ensureModule(api, nsID, { name: 'События доступа', handle: 'activity_events', fields: eventFields(identities, datasets, data_classes, policies) })
  const alerts = await ensureModule(api, nsID, { name: 'Алерты', handle: 'alerts', fields: alertFields(identities, datasets, activity_events) })
  const ai_assets = await ensureModule(api, nsID, { name: 'Активы AI', handle: 'ai_assets', fields: aiFields(identities) })
  const compliance_controls = await ensureModule(api, nsID, { name: 'Контроли соответствия', handle: 'compliance_controls', fields: controlFields(data_classes, policies) })

  const modules = {
    connectors, scans, data_stores, datasets, data_classes, classifications,
    identities, entitlements, masking_profiles, row_filters, policies, policy_bindings,
    access_requests, findings, remediations, activity_events, alerts, ai_assets,
    compliance_controls,
  }

  const charts = {
    findingsBySeverity: await ensureChart(api, nsID, doughnutChart('Находки по серьёзности', 'findings-by-severity', findings, 'severity')),
    storesByExposure: await ensureChart(api, nsID, doughnutChart('Хранилища по экспозиции', 'stores-by-exposure', data_stores, 'exposure')),
    requestsByStatus: await ensureChart(api, nsID, doughnutChart('Заявки по статусу', 'requests-by-status', access_requests, 'status')),
    controlsByStatus: await ensureChart(api, nsID, doughnutChart('Контроли по покрытию', 'controls-by-status', compliance_controls, 'status')),
  }

  const pageIDs = {}
  for (const page of buildPages(modules, charts)) {
    pageIDs[page.handle] = await ensurePage(api, nsID, page)
  }
  await parentPages(api, nsID, pageIDs, {
    store: 'inventory',
    dataset: 'datasets',
    'classification-item': 'classification',
    entitlement: 'access',
    identity: 'identities',
    policy: 'policies',
    request: 'requests',
    finding: 'findings',
    alert: 'activity',
    'ai-asset': 'ai',
    connector: 'connectors',
  })

  const agentUrl = (process.env.DSPM_AGENT_URL || 'http://127.0.0.1:8091').replace(/\/$/, '')
  for (const chain of buildRuleChains({ nsID, modules, agentUrl })) {
    await ensureRuleChain(api, chain)
  }

  await seedClasses(api, nsID, data_classes)

  const summary = {
    namespaceID: String(nsID),
    slug: 'dspm',
    api: base,
    modules: Object.fromEntries(Object.entries(modules).map(([k, v]) => [k, String(v)])),
    charts: Object.fromEntries(Object.entries(charts).map(([k, v]) => [k, String(v)])),
    pages: pageIDs,
  }
  writeFileSync(join(HERE, 'applied.json'), JSON.stringify(summary, null, 2))
  console.log('\nDSPM ready', nsID)
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
