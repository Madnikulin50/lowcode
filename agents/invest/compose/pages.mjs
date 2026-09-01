import {
  block, recordList, recordBlock, metricBlock, metricItem, ruleChain,
  organizer, commentBlock, withBlockIDs, pageIcon, withRoles,
} from './helpers.mjs'
import { INVEST_PAGE_DOCS } from '../../../client3/web/compose/src/help/appDocs.js'

const recID = '${recordID}'

export function buildPages ({ modules, charts, roles = {} }) {
  const m = modules
  const finance = [roles.investor, roles.pmo, roles.bank]
  const pmo = [roles.investor, roles.pmo]
  return [
    dashboardPage(m, charts, { finance, pmo }),
    listPage('Проекты', 'projects', 10, 'fas industry', m.projects, ['name', 'code', 'phase', 'status', 'budget_planned', 'spi', 'cpi', 'eac']),
    projectCard(m, charts),
    listPage('WBS', 'wbs', 20, 'fas sitemap', m.wbs_items, ['code', 'name', 'project', 'level', 'percent_complete', 'spi', 'cpi', 'is_critical'], { extraBlocks: [
      block('Chart', 'График работ', [0, 0, 48, 22], { chartID: String(charts.wbsGantt) }),
    ], listY: 22 }),
    wbsCard(m),
    documentsPage(m, charts),
    documentCard(m),
    listPage('Договоры', 'contracts', 40, 'fas handshake', m.contracts, ['number', 'title', 'project', 'counterparty', 'amount', 'status', 'end_date']),
    contractCard(m),
    risksPage(m),
    riskCard(m),
    changesPage(m),
    rfcCard(m),
    budgetPage(m, charts),
    budgetLineCard(m),
    listPage('Прогресс', 'progress', 80, 'fas camera', m.progress_facts, ['wbs', 'project', 'quantity', 'percent', 'recorded_at', 'author']),
    progressCard(m),
    nsiPage(m),
    advisorsPage(),
    hiddenRecord('Тип документа', 'document-type', m.document_types, 91, ['name', 'code', 'description']),
    hiddenRecord('Тип конструкции', 'construction-type', m.construction_types, 95, ['name', 'code', 'description']),
    hiddenRecord('Шаблон WBS', 'wbs-template', m.wbs_templates, 96, ['construction_type', 'code', 'name', 'level', 'parent_code', 'budget_planned']),
    hiddenRecord('Требование фазы', 'phase-requirement', m.phase_requirements, 97, ['phase', 'doc_type', 'required']),
    hiddenRecord('Контрагент', 'counterparty', m.counterparties, 92, ['name', 'inn', 'kpp', 'role', 'notes']),
    hiddenRecord('Материал', 'material', m.materials, 93, ['name', 'unit', 'unit_price', 'gost']),
    hiddenRecord('Норма времени', 'labor-norm', m.labor_norms, 94, ['name', 'unit', 'hours', 'notes']),
    hiddenRecord('Версия документа', 'document-version', m.document_versions, 32, ['document', 'version', 'author', 'comment', 'created_on', 'file']),
    hiddenRecord('Согласование', 'approval', m.approvals, 33, ['document', 'approver', 'decision', 'step', 'role', 'due_date', 'decided_at', 'comment']),
    hiddenRecord('Участник', 'project-member', m.project_members, 12, ['project', 'user', 'role']),
    hiddenRecord('Платёж', 'cashflow-item', m.cashflow_items, 72, ['project', 'budget_line', 'date', 'amount', 'direction', 'description']),
    hiddenRecord('Запись журнала', 'change-log-item', m.change_log, 62, ['rfc', 'project', 'summary', 'old_budget', 'new_budget', 'old_end', 'new_end', 'actor', 'changed_at']),
  ].map(withPageDocs)
}

function withPageDocs (page) {
  const docs = INVEST_PAGE_DOCS[page.handle]
  if (!docs) return page
  return {
    ...page,
    description: docs.description,
    config: { ...(page.config || {}), help: docs.help },
  }
}

function dashboardPage (m, charts, vis = {}) {
  return {
    title: 'Дашборд',
    handle: 'dashboard',
    visible: true,
    weight: 0,
    description: 'Сводка инвестора',
    config: pageIcon('fas chart-pie'),
    blocks: withBlockIDs([
      metricBlock('Сводка', [0, 0, 48, 14], [
        metricItem('Проекты', m.projects, "status = 'active'", { role: 'hero', color: '#2e59d9', fontSize: '28' }),
        metricItem('На согласовании', m.documents, "status = 'in_review'", { role: 'balloon', color: '#f6c23e' }),
        metricItem('Открытые RFC', m.change_requests, "status = 'in_review'", { role: 'balloon', color: '#e74a3b' }),
        metricItem('Открытые риски', m.risks, "status = 'open'", { role: 'meta', color: '#e74a3b' }),
      ]),
      withRoles(ruleChain('EVM', [0, 14, 16, 10], {
        chainID: 'invest-recalculate-evm',
        label: 'Пересчитать EVM',
        variant: 'primary',
        icon: 'percent',
      }), vis.pmo),
      withRoles(ruleChain('CPM', [16, 14, 16, 10], {
        chainID: 'invest-critical-path',
        label: 'Критический путь',
        variant: 'info',
        icon: 'sitemap',
      }), vis.pmo),
      ruleChain('Алерты', [32, 14, 16, 10], {
        chainID: 'invest-threshold-alert',
        label: 'Проверить пороги',
        variant: 'warning',
        icon: 'bell',
      }),
      block('Chart', 'Документы по статусу', [0, 24, 16, 18], { chartID: String(charts.docsByStatus) }),
      block('Chart', 'RFC по статусу', [16, 24, 16, 18], { chartID: String(charts.rfcByStatus) }),
      block('Chart', 'Риски по баллу', [32, 24, 16, 18], { chartID: String(charts.risksByScore || charts.risksByImpact) }),
      block('Chart', 'График WBS', [0, 42, 48, 22], { chartID: String(charts.wbsGantt) }),
      withRoles(recordList('Портфель', [0, 64, 48, 18], m.projects, ['name', 'phase', 'status', 'spi', 'cpi', 'eac', 'budget_planned'], {
        presort: 'spi ASC',
        perPage: 12,
        hideAddButton: true,
      }), vis.finance),
      recordList('Документы на мне', [0, 82, 24, 18], m.documents, ['title', 'status', 'assignee', 'due_date'], {
        prefilter: "status = 'in_review' AND assignee = ${userID}",
        perPage: 8,
        hideAddButton: true,
      }),
      recordList('Открытые RFC', [24, 82, 24, 18], m.change_requests, ['title', 'rfc_type', 'status', 'delta_budget', 'eac_after'], {
        prefilter: "status = 'in_review'",
        perPage: 8,
        hideAddButton: true,
      }),
    ]),
  }
}

function listPage (title, handle, weight, icon, moduleID, fields, extra = {}) {
  const blocks = []
  if (extra.extraBlocks) blocks.push(...extra.extraBlocks)
  blocks.push(recordList(title, [0, extra.listY || 0, 48, extra.listY ? 36 : 50], moduleID, fields))
  return {
    title,
    handle,
    visible: true,
    weight,
    config: pageIcon(icon),
    blocks: withBlockIDs(blocks),
  }
}

function projectCard (m, charts) {
  return {
    title: 'Проект',
    handle: 'project',
    moduleID: String(m.projects),
    visible: false,
    weight: 11,
    config: pageIcon('fas industry'),
    blocks: withBlockIDs([
      recordBlock('Проект', [0, 0, 32, 28], [
        'name', 'code', 'status', 'phase', 'construction_type', 'investor',
        'start_planned', 'end_planned', 'budget_planned', 'budget_actual',
        'spi', 'cpi', 'eac', 'description',
      ], {
        fieldRoles: {
          name: 'title',
          code: 'subtitle',
          status: 'badge',
          phase: 'badge',
          investor: 'meta',
          construction_type: 'badge',
          start_planned: 'meta',
          end_planned: 'meta',
          budget_planned: 'meta',
          spi: 'badge',
          cpi: 'badge',
          eac: 'meta',
          description: 'body',
        },
        sections: [{ title: 'Описание', fields: ['description'] }],
      }),
      ruleChain('EVM', [32, 0, 16, 10], {
        chainID: 'invest-recalculate-evm',
        label: 'Пересчитать EVM',
        icon: 'percent',
        context: { projectID: recID },
      }),
      ruleChain('WBS из шаблона', [32, 10, 16, 8], {
        chainID: 'invest-clone-wbs',
        label: 'Создать WBS из шаблона',
        variant: 'info',
        icon: 'sitemap',
        context: { projectID: recID },
      }),
      metricBlock('По проекту', [32, 18, 16, 10], [
        metricItem('Документы', m.documents, `project = ${recID}`, { role: 'hero', color: '#2e59d9' }),
        metricItem('Утверждены', m.documents, `project = ${recID} AND status = 'approved'`, { role: 'balloon', color: '#1cc88a' }),
        metricItem('RFC', m.change_requests, `project = ${recID} AND status = 'in_review'`, { role: 'meta', color: '#e74a3b' }),
      ], { itemsPerRow: '1' }),
      recordList('WBS', [0, 28, 24, 22], m.wbs_items, ['code', 'name', 'level', 'percent_complete', 'spi'], {
        prefilter: `project = ${recID}`,
        refField: 'project',
      }),
      recordList('Документы', [24, 28, 24, 22], m.documents, ['title', 'status', 'due_date'], {
        prefilter: `project = ${recID}`,
        refField: 'project',
      }),
      recordList('Договоры', [0, 50, 24, 18], m.contracts, ['number', 'title', 'amount', 'status'], {
        prefilter: `project = ${recID}`,
        refField: 'project',
      }),
      recordList('Участники', [24, 50, 24, 18], m.project_members, ['user', 'role'], {
        prefilter: `project = ${recID}`,
        refField: 'project',
      }),
      recordList('Обязательные документы фазы', [0, 68, 48, 16], m.phase_requirements, ['phase', 'doc_type', 'required'], {
        prefilter: "phase = '${record.values.phase}'",
        hideAddButton: true,
        perPage: 10,
      }),
    ]),
  }
}

function wbsCard (m) {
  return {
    title: 'Элемент WBS',
    handle: 'wbs-item',
    moduleID: String(m.wbs_items),
    visible: false,
    weight: 21,
    config: pageIcon('fas sitemap'),
    blocks: withBlockIDs([
      recordBlock('Работа', [0, 0, 32, 32], [
        'code', 'name', 'level', 'project', 'parent', 'predecessor',
        'start_planned', 'end_planned', 'percent_complete', 'is_critical',
        'budget_planned', 'actual_cost', 'pv', 'ev', 'spi', 'cpi', 'eac', 'notes',
      ], {
        fieldRoles: {
          name: 'title',
          code: 'subtitle',
          level: 'badge',
          is_critical: 'badge',
          percent_complete: 'meta',
          spi: 'badge',
          cpi: 'badge',
          project: 'meta',
          notes: 'body',
        },
        sections: [{ title: 'EVM', fields: ['budget_planned', 'actual_cost', 'pv', 'ev', 'spi', 'cpi', 'eac'] }],
      }),
      recordList('Факты прогресса', [32, 0, 16, 32], m.progress_facts, ['quantity', 'percent', 'recorded_at'], {
        prefilter: `wbs = ${recID}`,
        refField: 'wbs',
        perPage: 10,
      }),
      recordList('Документы', [0, 32, 24, 20], m.documents, ['title', 'status'], {
        prefilter: `wbs = ${recID}`,
        refField: 'wbs',
      }),
      recordList('Статьи бюджета', [24, 32, 24, 20], m.budget_lines, ['article', 'planned', 'actual'], {
        prefilter: `wbs = ${recID}`,
        refField: 'wbs',
      }),
    ]),
  }
}

function documentsPage (m, charts) {
  return {
    title: 'Документы',
    handle: 'documents',
    visible: true,
    weight: 30,
    config: pageIcon('fas file'),
    blocks: withBlockIDs([
      metricBlock('Реестр', [0, 0, 48, 12], [
        metricItem('Всего', m.documents, '', { role: 'hero', color: '#2e59d9', fontSize: '28' }),
        metricItem('Черновики', m.documents, "status = 'draft'", { role: 'meta', color: '#858796' }),
        metricItem('На согласовании', m.documents, "status = 'in_review'", { role: 'balloon', color: '#f6c23e' }),
        metricItem('Утверждены', m.documents, "status = 'approved'", { role: 'meta', color: '#1cc88a' }),
      ]),
      organizer('Черновик', [0, 12, 12, 26], m.documents, {
        labelField: 'title', descriptionField: 'number', groupField: 'status', group: 'draft',
      }),
      organizer('На согласовании', [12, 12, 12, 26], m.documents, {
        labelField: 'title', descriptionField: 'number', groupField: 'status', group: 'in_review',
      }),
      organizer('Утверждён', [24, 12, 12, 26], m.documents, {
        labelField: 'title', descriptionField: 'number', groupField: 'status', group: 'approved',
      }),
      organizer('Отклонён', [36, 12, 12, 26], m.documents, {
        labelField: 'title', descriptionField: 'number', groupField: 'status', group: 'rejected',
      }),
      block('Chart', 'По статусу', [0, 38, 16, 18], { chartID: String(charts.docsByStatus) }),
      recordList('Реестр документов', [16, 38, 32, 22], m.documents, ['title', 'number', 'doc_type', 'project', 'status', 'extract_status', 'assignee', 'due_date']),
    ]),
  }
}

function documentCard (m) {
  return {
    title: 'Документ',
    handle: 'document',
    moduleID: String(m.documents),
    visible: false,
    weight: 31,
    config: pageIcon('fas file'),
    blocks: withBlockIDs([
      recordBlock('Документ', [0, 0, 32, 28], [
        'title', 'number', 'status', 'sign_status', 'extract_status', 'doc_type', 'project', 'wbs', 'contract',
        'author', 'assignee', 'due_date', 'file', 'summary', 'notes',
      ], {
        fieldRoles: {
          title: 'title',
          number: 'subtitle',
          status: 'badge',
          sign_status: 'badge',
          extract_status: 'badge',
          doc_type: 'badge',
          project: 'meta',
          assignee: 'meta',
          due_date: 'meta',
          summary: 'body',
          notes: 'body',
        },
        sections: [
          { title: 'Файл', fields: ['file'] },
          { title: 'ИИ', fields: ['summary', 'extracted_text', 'extract_error', 'extracted_at'] },
        ],
      }),
      ruleChain('На согласование', [32, 0, 16, 9], {
        chainID: 'invest-submit-approval',
        label: 'Отправить',
        variant: 'warning',
        icon: 'share-alt',
      }),
      ruleChain('Согласовать шаг', [32, 9, 16, 9], {
        chainID: 'invest-approve-document',
        label: 'Согласовать мой шаг',
        variant: 'success',
        icon: 'check',
      }),
      ruleChain('Отклонить', [32, 18, 16, 9], {
        chainID: 'invest-reject-document',
        label: 'Отклонить',
        variant: 'danger',
        icon: 'times',
      }),
      ruleChain('Эскалация', [32, 27, 16, 9], {
        chainID: 'invest-escalate-approval',
        label: 'Эскалировать',
        variant: 'info',
        icon: 'arrow-up',
      }),
      ruleChain('Summary', [32, 36, 16, 9], {
        chainID: 'invest-summarize-document',
        label: 'Обновить summary',
        variant: 'primary',
        icon: 'align-left',
      }),
      recordList('Версии', [0, 46, 24, 20], m.document_versions, ['version', 'author', 'comment', 'created_on'], {
        prefilter: `document = ${recID}`,
        refField: 'document',
        presort: 'version DESC',
      }),
      recordList('Маршрут', [24, 46, 24, 20], m.approvals, ['step', 'approver', 'role', 'decision', 'due_date', 'comment'], {
        prefilter: `document = ${recID}`,
        refField: 'document',
        presort: 'step ASC',
      }),
      commentBlock('Комментарии', [0, 66, 48, 18], m.document_comments, {
        referenceField: 'document',
        filter: `document = ${recID}`,
        titleField: '',
      }),
    ]),
  }
}

function contractCard (m) {
  return {
    title: 'Договор',
    handle: 'contract',
    moduleID: String(m.contracts),
    visible: false,
    weight: 41,
    config: pageIcon('fas handshake'),
    blocks: withBlockIDs([
      recordBlock('Договор', [0, 0, 32, 26], [
        'title', 'number', 'status', 'project', 'counterparty', 'amount', 'start_date', 'end_date', 'file', 'terms',
      ], {
        fieldRoles: {
          title: 'title',
          number: 'subtitle',
          status: 'badge',
          amount: 'meta',
          counterparty: 'meta',
          end_date: 'meta',
          terms: 'body',
        },
        sections: [{ title: 'Условия', fields: ['terms'] }],
      }),
      ruleChain('Юрист', [32, 0, 16, 12], {
        chainID: 'invest-lawyer-review',
        label: 'Спросить юриста',
        variant: 'primary',
        icon: 'scale-balanced',
      }),
      recordList('Документы пакета', [0, 26, 48, 22], m.documents, ['title', 'number', 'status', 'due_date'], {
        prefilter: `contract = ${recID}`,
        refField: 'contract',
      }),
    ]),
  }
}

function risksPage (m) {
  return {
    title: 'Риски',
    handle: 'risks',
    visible: true,
    weight: 50,
    config: pageIcon('fas triangle-exclamation'),
    blocks: withBlockIDs([
      organizer('Открыт', [0, 0, 12, 26], m.risks, {
        labelField: 'title', descriptionField: 'impact', groupField: 'status', group: 'open',
      }),
      organizer('Митигация', [12, 0, 12, 26], m.risks, {
        labelField: 'title', descriptionField: 'impact', groupField: 'status', group: 'mitigating',
      }),
      organizer('Закрыт', [24, 0, 12, 26], m.risks, {
        labelField: 'title', descriptionField: 'impact', groupField: 'status', group: 'closed',
      }),
      organizer('Принят', [36, 0, 12, 26], m.risks, {
        labelField: 'title', descriptionField: 'impact', groupField: 'status', group: 'accepted',
      }),
      recordList('Реестр рисков', [0, 26, 48, 24], m.risks, ['title', 'project', 'probability', 'impact', 'score', 'status', 'owner']),
    ]),
  }
}

function riskCard (m) {
  return {
    title: 'Риск',
    handle: 'risk',
    moduleID: String(m.risks),
    visible: false,
    weight: 51,
    config: pageIcon('fas triangle-exclamation'),
    blocks: withBlockIDs([
      recordBlock('Риск', [0, 0, 48, 28], [
        'title', 'status', 'probability', 'impact', 'score', 'project', 'wbs', 'owner', 'mitigation', 'description',
      ], {
        fieldRoles: {
          title: 'title',
          status: 'badge',
          probability: 'badge',
          impact: 'badge',
          score: 'badge',
          owner: 'meta',
          mitigation: 'body',
          description: 'body',
        },
        sections: [
          { title: 'Митигация', fields: ['mitigation'] },
          { title: 'Описание', fields: ['description'] },
        ],
      }),
    ]),
  }
}

function changesPage (m) {
  return {
    title: 'Изменения',
    handle: 'changes',
    visible: true,
    weight: 60,
    config: pageIcon('fas shuffle'),
    blocks: withBlockIDs([
      organizer('Черновик', [0, 0, 12, 24], m.change_requests, {
        labelField: 'title', descriptionField: 'rfc_type', groupField: 'status', group: 'draft',
      }),
      organizer('На согласовании', [12, 0, 12, 24], m.change_requests, {
        labelField: 'title', descriptionField: 'rfc_type', groupField: 'status', group: 'in_review',
      }),
      organizer('Утверждён', [24, 0, 12, 24], m.change_requests, {
        labelField: 'title', descriptionField: 'rfc_type', groupField: 'status', group: 'approved',
      }),
      organizer('Отклонён', [36, 0, 12, 24], m.change_requests, {
        labelField: 'title', descriptionField: 'rfc_type', groupField: 'status', group: 'rejected',
      }),
      recordList('RFC', [0, 24, 48, 22], m.change_requests, ['title', 'project', 'rfc_type', 'status', 'delta_budget', 'delta_days', 'eac_after']),
    ]),
  }
}

function rfcCard (m) {
  return {
    title: 'RFC',
    handle: 'change-request',
    moduleID: String(m.change_requests),
    visible: false,
    weight: 61,
    config: pageIcon('fas shuffle'),
    blocks: withBlockIDs([
      recordBlock('Запрос на изменение', [0, 0, 32, 28], [
        'title', 'status', 'rfc_type', 'project', 'wbs', 'delta_budget', 'delta_days',
        'eac_before', 'eac_after', 'end_after', 'simulated', 'author', 'justification',
      ], {
        fieldRoles: {
          title: 'title',
          status: 'badge',
          rfc_type: 'badge',
          simulated: 'badge',
          delta_budget: 'meta',
          delta_days: 'meta',
          eac_after: 'meta',
          justification: 'body',
        },
        sections: [{ title: 'Обоснование', fields: ['justification'] }],
      }),
      ruleChain('Симулировать', [32, 0, 16, 9], {
        chainID: 'invest-simulate-rfc',
        label: 'Симулировать EAC',
        variant: 'info',
        icon: 'calculator',
      }),
      ruleChain('На согласование', [32, 9, 16, 9], {
        chainID: 'invest-submit-rfc-status',
        label: 'Отправить RFC',
        variant: 'warning',
        icon: 'share-alt',
      }),
      ruleChain('Утвердить', [32, 18, 16, 9], {
        chainID: 'invest-approve-rfc',
        label: 'Утвердить RFC',
        variant: 'success',
        icon: 'check',
      }),
      ruleChain('Отклонить', [32, 27, 16, 9], {
        chainID: 'invest-reject-rfc',
        label: 'Отклонить RFC',
        variant: 'danger',
        icon: 'times',
      }),
      recordList('Журнал изменений', [0, 36, 48, 20], m.change_log, ['summary', 'old_budget', 'new_budget', 'old_end', 'new_end', 'actor', 'changed_at'], {
        prefilter: `rfc = ${recID}`,
        refField: 'rfc',
      }),
    ]),
  }
}

function budgetPage (m, charts) {
  return {
    title: 'Бюджет',
    handle: 'budget',
    visible: true,
    weight: 70,
    config: pageIcon('fas wallet'),
    blocks: withBlockIDs([
      metricBlock('Финансы', [0, 0, 48, 12], [
        metricItem('Статей', m.budget_lines, '', { role: 'hero', color: '#2e59d9', fontSize: '28' }),
        metricItem('План', m.budget_lines, '', { role: 'meta', field: 'planned', operation: 'sum', color: '#2e59d9', prefix: '₽ ' }),
        metricItem('Факт', m.budget_lines, '', { role: 'meta', field: 'actual', operation: 'sum', color: '#e74a3b', prefix: '₽ ' }),
        metricItem('Резерв', m.budget_lines, '', { role: 'balloon', field: 'reserve', operation: 'sum', color: '#1cc88a', prefix: '₽ ' }),
      ]),
      ruleChain('Финконтролёр', [0, 12, 16, 10], {
        chainID: 'invest-fin-review',
        label: 'Спросить финконтролёра',
        icon: 'comments',
      }),
      recordList('Статьи бюджета', [0, 22, 24, 16], m.budget_lines, ['article', 'project', 'wbs', 'planned', 'actual', 'reserve']),
      block('Chart', 'Приход / расход', [24, 22, 24, 16], { chartID: String(charts.cashflowByDir) }),
      block('Chart', 'Поток по датам', [0, 38, 48, 16], { chartID: String(charts.cashflowByDate) }),
      recordList('Денежный поток', [0, 54, 48, 18], m.cashflow_items, ['date', 'project', 'direction', 'amount', 'description'], {
        presort: 'date DESC',
      }),
    ]),
  }
}

function budgetLineCard (m) {
  return {
    title: 'Статья бюджета',
    handle: 'budget-line',
    moduleID: String(m.budget_lines),
    visible: false,
    weight: 71,
    config: pageIcon('fas wallet'),
    blocks: withBlockIDs([
      recordBlock('Статья', [0, 0, 32, 22], ['article', 'project', 'wbs', 'planned', 'actual', 'reserve', 'notes'], {
        fieldRoles: {
          article: 'title',
          planned: 'meta',
          actual: 'meta',
          reserve: 'badge',
          notes: 'body',
        },
      }),
      recordList('Платежи', [32, 0, 16, 22], m.cashflow_items, ['date', 'amount', 'direction'], {
        prefilter: `budget_line = ${recID}`,
        refField: 'budget_line',
      }),
    ]),
  }
}

function progressCard (m) {
  return {
    title: 'Факт прогресса',
    handle: 'progress-fact',
    moduleID: String(m.progress_facts),
    visible: false,
    weight: 81,
    config: pageIcon('fas camera'),
    blocks: withBlockIDs([
      recordBlock('Фиксация', [0, 0, 32, 30], [
        'wbs', 'project', 'quantity', 'unit', 'percent', 'cost', 'photo', 'geo', 'author', 'recorded_at', 'offline', 'notes',
      ], {
        fieldRoles: {
          wbs: 'title',
          percent: 'badge',
          quantity: 'subtitle',
          recorded_at: 'meta',
          author: 'meta',
          notes: 'body',
        },
        sections: [
          { title: 'Фото и геолокация', fields: ['photo', 'geo'] },
        ],
      }),
      ruleChain('EVM', [32, 0, 16, 12], {
        chainID: 'invest-recalculate-evm-fact',
        label: 'Зафиксировать и пересчитать',
        icon: 'percent',
      }),
    ]),
  }
}

function nsiPage (m) {
  return {
    title: 'НСИ',
    handle: 'nsi',
    visible: true,
    weight: 90,
    config: pageIcon('fas book'),
    blocks: withBlockIDs([
      recordList('Типы документов', [0, 0, 24, 18], m.document_types, ['name', 'code']),
      recordList('Контрагенты', [24, 0, 24, 18], m.counterparties, ['name', 'inn', 'role']),
      recordList('Типы конструкций', [0, 18, 24, 16], m.construction_types, ['name', 'code']),
      recordList('Шаблоны WBS', [24, 18, 24, 16], m.wbs_templates, ['construction_type', 'code', 'name', 'level']),
      recordList('Материалы', [0, 34, 16, 16], m.materials, ['name', 'unit', 'unit_price']),
      recordList('Нормы времени', [16, 34, 16, 16], m.labor_norms, ['name', 'unit', 'hours']),
      recordList('Документы по фазе', [32, 34, 16, 16], m.phase_requirements, ['phase', 'doc_type', 'required']),
    ]),
  }
}

const LAWYER_PROMPT = `Ты ИИ-советчик «Юрист» пространства Инвестпроекты. Human-in-the-loop: не утверждай документы и не меняй статусы сам — дай рекомендацию.
Работай с модулями contracts и documents (инструменты module_contracts_* и module_documents_*).
Проверяй договоры и пакеты документов: сроки, сумма, отсутствие файла, статус УКЭП, риски формулировок.
Ответ: (1) вывод, (2) риски списком, (3) что проверить человеку, (4) прецедент/ссылка на запись. Цифры только из инструментов.`

const FIN_PROMPT = `Ты ИИ-советчик «Финконтролёр» пространства Инвестпроекты. Human-in-the-loop: не пересчитывай бюджет сам — дай рекомендацию.
Работай с модулями budget_lines, cashflow_items, wbs_items, change_requests (module_*_records / search).
Смотри план/факт/резерв, SPI/CPI/EAC, отклонения, влияние RFC.
Ответ: (1) вывод по деньгам, (2) отклонения, (3) что проверить, (4) ссылка на записи. Цифры только из инструментов.`

function advisorsPage () {
  return {
    title: 'ИИ-советчики',
    handle: 'advisors',
    visible: true,
    weight: 100,
    config: {
      ...pageIcon('fas brain'),
      prompt: 'ИИ-советчики Юрист и Финконтролёр. Рекомендации проверяет человек.',
    },
    blocks: withBlockIDs([
      block('AiChat', 'Юрист', [0, 0, 24, 36], {
        prompt: LAWYER_PROMPT,
        startPrompt: LAWYER_PROMPT,
        model: '',
      }),
      block('AiChat', 'Финконтролёр', [24, 0, 24, 36], {
        prompt: FIN_PROMPT,
        startPrompt: FIN_PROMPT,
        model: '',
      }),
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
    blocks: withBlockIDs([
      recordBlock(title, [0, 0, 48, 28], fields, {
        fieldRoles: { [fields[0]]: 'title' },
      }),
    ]),
  }
}
