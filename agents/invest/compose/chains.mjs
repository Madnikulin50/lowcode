export function buildRuleChains ({ nsID, modules, engineUrl, evmUrl }) {
  const ns = String(nsID)
  const rfc = String(modules.change_requests)
  const risks = String(modules.risks)
  const projects = String(modules.projects)
  const wbs = String(modules.wbs_items)
  const facts = String(modules.progress_facts)
  const documents = String(modules.documents)
  const calc = (evmUrl || 'http://localhost:8088/api').replace(/\/$/, '')

  const jobBody = extra => {
    const fields = {
      namespaceID: '{{namespaceID}}',
      recordID: '{{recordID}}',
      documentID: '{{documentID}}',
      projectID: '{{projectID}}',
      token: '{{authToken}}',
      userID: '{{userID}}',
      ...extra,
    }
    const inner = Object.entries(fields).map(([k, v]) => {
      if (k === 'namespaceID') return `"${k}":${v}`
      return `"${k}":"${v}"`
    }).join(',')
    return `{${inner}}`
  }

  const http = (id, name, path, description, extra = {}) => ({
    id,
    name,
    description,
    entryNode: 'http',
    namespaceID: ns,
    nodes: [{
      id: 'http',
      type: 'http',
      label: name,
      config: {
        url: engineUrl + path,
        method: 'POST',
        body: jobBody(extra),
        timeout: 60,
      },
    }],
    edges: extra.edges || [],
  })

  const evmChain = (id, name, description, projectExpr) => ({
    id,
    name,
    description,
    entryNode: 'search_wbs',
    namespaceID: ns,
    nodes: [
      {
        id: 'search_wbs',
        type: 'crud',
        label: 'WBS',
        config: {
          operation: 'search',
          namespaceID: ns,
          moduleID: wbs,
          moduleHandle: 'wbs_items',
          query: '',
          limit: 500,
        },
      },
      {
        id: 'search_facts',
        type: 'crud',
        label: 'Факты прогресса',
        config: {
          operation: 'search',
          namespaceID: ns,
          moduleID: facts,
          moduleHandle: 'progress_facts',
          query: '',
          limit: 500,
        },
      },
      {
        id: 'http',
        type: 'http',
        label: 'calc-evm',
        config: {
          url: calc + '/call/evm',
          method: 'POST',
          body: '{"projectID":"' + projectExpr + '","items":{{search_wbs.records}},"facts":{{search_facts.records}}}',
          timeout: 60,
        },
      },
      {
        id: 'foreach_wbs',
        type: 'foreach',
        label: 'Каждая работа',
        config: { items: 'items', itemVar: 'item' },
      },
      {
        id: 'upd_wbs',
        type: 'crud',
        label: 'EVM на WBS',
        config: {
          operation: 'update',
          namespaceID: ns,
          moduleID: wbs,
          moduleHandle: 'wbs_items',
          recordID: '{{item.id}}',
          omitEmpty: true,
          fields: {
            percent_complete: '{{item.percentComplete}}',
            actual_cost: '{{item.actualCost}}',
            pv: '{{item.pv}}',
            ev: '{{item.ev}}',
            spi: '{{item.spi}}',
            cpi: '{{item.cpi}}',
            eac: '{{item.eac}}',
          },
        },
      },
      {
        id: 'foreach_projects',
        type: 'foreach',
        label: 'Каждый проект',
        config: { items: 'projects', itemVar: 'proj' },
      },
      {
        id: 'upd_project',
        type: 'crud',
        label: 'EVM на проект',
        config: {
          operation: 'update',
          namespaceID: ns,
          moduleID: projects,
          moduleHandle: 'projects',
          recordID: '{{proj.projectID}}',
          omitEmpty: true,
          fields: {
            spi: '{{proj.spi}}',
            cpi: '{{proj.cpi}}',
            eac: '{{proj.eac}}',
            budget_actual: '{{proj.ac}}',
          },
        },
      },
    ],
    edges: [
      { from: 'search_wbs', to: 'search_facts' },
      { from: 'search_facts', to: 'http' },
      { from: 'http', to: 'foreach_wbs' },
      { from: 'foreach_wbs', to: 'upd_wbs' },
      { from: 'http', to: 'foreach_projects' },
      { from: 'foreach_projects', to: 'upd_project' },
    ],
  })

  const httpThenMail = (id, name, path, description, extra = {}) => ({
    id,
    name,
    description,
    entryNode: 'http',
    namespaceID: ns,
    nodes: [
      {
        id: 'http',
        type: 'http',
        label: name,
        config: {
          url: engineUrl + path,
          method: 'POST',
          body: jobBody(extra),
          timeout: 60,
        },
      },
      { id: 'check', type: 'condition', label: 'Есть email', config: { field: 'notifyEmail', operator: 'notEmpty' } },
      {
        id: 'mail',
        type: 'mail',
        label: 'Письмо',
        config: {
          to: '{{notifyEmail}}',
          subject: extra.mailSubject || 'Инвестпроекты: {{title}}',
          body: extra.mailBody || '<p>{{message}}</p>',
          contentType: 'html',
        },
      },
    ],
    edges: [
      { from: 'http', to: 'check' },
      { from: 'check', to: 'mail', condition: 'check_result' },
    ],
  })

  return [
    httpThenMail('invest-submit-approval', 'Инвест: отправить документ на согласование',
      '/submit-approval', 'Статус in_review, автоверсия, шаги approvals по ролям PMO/инвестор.',
      { mailSubject: 'На согласование: {{title}}', mailBody: '<p>Документ отправлен на согласование.</p><p>{{message}}</p>' }),
    httpThenMail('invest-approve-document', 'Инвест: согласовать мой шаг',
      '/decide-approval', 'Закрывает текущий шаг маршрута. Документ утверждён, когда все шаги пройдены.',
      { decision: 'approved', mailSubject: 'Согласовано: {{title}}' }),
    http('invest-reject-document', 'Инвест: отклонить документ', '/decide-approval',
      'Отклоняет текущий шаг и документ.', { decision: 'rejected' }),
    httpThenMail('invest-escalate-approval', 'Инвест: эскалировать согласование',
      '/escalate-approval', 'Текущий шаг → escalated, новый шаг на PMO.',
      { mailSubject: 'Эскалация: {{title}}', mailBody: '<p>Согласование эскалировано на PMO.</p>' }),
    http('invest-submit-rfc', 'Инвест: отправить RFC', '/simulate-rfc',
      'Сначала симулирует EAC; статус на согласовании выставляет отдельная кнопка или вручную.'),
    http('invest-simulate-rfc', 'Инвест: симулировать RFC', '/simulate-rfc',
      'Пишет eac_before/eac_after и прогноз финиша без изменения baseline.'),
    http('invest-approve-rfc', 'Инвест: утвердить RFC', '/approve-rfc',
      'Требует симуляцию. Двигает бюджет/срок проекта, журнал, пересчёт EVM.'),
    http('invest-reject-rfc', 'Инвест: отклонить RFC', '/reject-rfc', 'RFC → отклонён.'),
    http('invest-clone-wbs', 'Инвест: WBS из шаблона', '/clone-wbs',
      'Копирует wbs_templates выбранного типа конструкции в проект.'),
    evmChain('invest-recalculate-evm', 'Инвест: пересчитать EVM',
      'Search WBS/факты → calc-evm → SPI/CPI/EAC на работы и проекты.', '{{projectID}}'),
    evmChain('invest-recalculate-evm-fact', 'Инвест: EVM по факту прогресса',
      'То же умение; projectID берётся из поля project записи факта.', '{{project}}'),
    {
      id: 'invest-critical-path',
      name: 'Инвест: критический путь',
      description: 'POST /critical-path на invest-engine, выставляет is_critical на WBS.',
      entryNode: 'http',
      namespaceID: ns,
      nodes: [{
        id: 'http',
        type: 'http',
        label: 'engine CPM',
        config: {
          url: engineUrl + '/critical-path',
          method: 'POST',
          body: '{"namespaceID":{{namespaceID}},"projectID":"{{projectID}}","token":"{{authToken}}"}',
          timeout: 60,
        },
      }],
      edges: [],
    },
    httpThenMail('invest-threshold-alert', 'Инвест: пороговые алерты',
      '/alerts', 'Просроченные документы/RFC, CPI < 0.9, резерв ≤ 0 → риски.',
      { mailSubject: 'Алерты Инвестпроекты', mailBody: '<p>Создано рисков: {{created}}. Всего сигналов: {{alerts}}.</p>' }),
    {
      id: 'invest-lawyer-review',
      name: 'Инвест: юрист по договору',
      description: 'ИИ-советчик Юрист. Рекомендация, не решение.',
      entryNode: 'ai',
      namespaceID: ns,
      nodes: [{
        id: 'ai',
        type: 'ai',
        label: 'Юрист',
        config: {
          agent: 'assistant',
          prompt: 'Ты юрист-советчик (human-in-the-loop). Проанализируй договор «{{title}}» № {{number}}, сумма {{amount}}, статус {{status}}, сроки {{start_date}}–{{end_date}}, условия: {{terms}}. Дай риски и что проверить человеку. Не меняй данные.',
        },
      }],
      edges: [],
    },
    {
      id: 'invest-fin-review',
      name: 'Инвест: финконтролёр',
      description: 'ИИ-советчик Финконтролёр. Рекомендация по бюджету/EVM.',
      entryNode: 'ai',
      namespaceID: ns,
      nodes: [{
        id: 'ai',
        type: 'ai',
        label: 'Финконтролёр',
        config: {
          agent: 'assistant',
          prompt: 'Ты финконтролёр-советчик (human-in-the-loop). Посмотри статью/проект: {{article}} план {{planned}} факт {{actual}} резерв {{reserve}}. Если полей нет — используй SPI {{spi}} CPI {{cpi}} EAC {{eac}}. Укажи отклонения и что проверить. Не меняй данные.',
        },
      }],
      edges: [],
    },
    {
      id: 'invest-flag-low-cpi',
      name: 'Инвест: флаг низкого CPI на работе WBS',
      description: 'Если CPI работы < 0.9 — открыть риск.',
      entryNode: 'check',
      namespaceID: ns,
      nodes: [
        { id: 'check', type: 'condition', label: 'CPI < 0.9', config: { field: 'cpi', operator: 'lt', value: '0.9' } },
        {
          id: 'risk',
          type: 'crud',
          label: 'Создать риск',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: risks,
            moduleHandle: 'risks',
            fields: {
              project: '{{project}}',
              wbs: '{{recordID}}',
              title: 'CPI ниже порога: {{name}}',
              probability: 'high',
              impact: 'high',
              score: '9',
              status: 'open',
              description: 'Автоалерт: CPI={{cpi}} на работе {{code}} {{name}}.',
            },
          },
        },
      ],
      edges: [{ from: 'check', to: 'risk', condition: 'check_result' }],
    },
    {
      id: 'invest-submit-rfc-status',
      name: 'Инвест: RFC на согласование',
      description: 'Только статус in_review (симуляцию делайте отдельной кнопкой).',
      entryNode: 'upd',
      namespaceID: ns,
      nodes: [{
        id: 'upd',
        type: 'crud',
        label: 'RFC на согласование',
        config: {
          operation: 'update',
          namespaceID: ns,
          moduleID: rfc,
          moduleHandle: 'change_requests',
          recordID: '{{recordID}}',
          fields: { status: 'in_review' },
        },
      }],
      edges: [],
    },
    {
      id: 'invest-summarize-document',
      name: 'Инвест: summary документа',
      description: 'Извлечь текст вложения (docx/xlsx/pdf/dxf/dwg/ifc/ArchiCAD) и записать summary в запись. Автозапуск после смены файла.',
      entryNode: 'has_file',
      namespaceID: ns,
      config: {
        triggers: [{
          resourceType: 'compose:record',
          eventType: 'afterCreate,afterUpdate',
          moduleHandle: 'documents',
          async: true,
          fileField: 'file',
        }],
      },
      nodes: [
        { id: 'has_file', type: 'condition', label: 'Есть файл', config: { field: 'file', operator: 'notEmpty' } },
        {
          id: 'pending',
          type: 'crud',
          label: 'Статус pending',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: documents,
            moduleHandle: 'documents',
            recordID: '{{recordID}}',
            omitEmpty: true,
            continueOnError: true,
            fields: { extract_status: 'pending' },
          },
        },
        {
          id: 'extract',
          type: 'document.extract',
          label: 'Извлечь текст',
          config: { attachmentField: 'file', maxChars: 64000, outText: 'extracted_text' },
        },
        { id: 'check_ok', type: 'condition', label: 'Текст есть', config: { field: 'extract_ok', operator: 'eq', value: 'true' } },
        {
          id: 'ai',
          type: 'ai',
          label: 'Summary',
          config: {
            agent: 'assistant',
            prompt: 'Составь краткое summary (5–10 предложений) инвестиционного документа «{{title}}» № {{number}} (тип {{doc_type}}). Пиши по-русски, только текст резюме, без префикса FINAL и без markdown-заголовков. Укажи стороны, даты, суммы, обязательства; для чертежей/BIM — листы, слои, этажи, штамп, если они есть в тексте.\n\nТекст документа:\n{{extracted_text}}',
          },
        },
        {
          id: 'save_ok',
          type: 'crud',
          label: 'Записать summary',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: documents,
            moduleHandle: 'documents',
            recordID: '{{recordID}}',
            fields: {
              summary: '{{ai_response}}',
              extracted_text: '{{extracted_text}}',
              extract_status: '{{extract_status}}',
              extract_error: '',
              extracted_at: '{{extracted_at}}',
              file_hash: '{{file_hash}}',
            },
          },
        },
        { id: 'check_fail', type: 'condition', label: 'Извлечение не удалось', config: { field: 'extract_ok', operator: 'eq', value: 'false' } },
        {
          id: 'save_fail',
          type: 'crud',
          label: 'Записать ошибку',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: documents,
            moduleHandle: 'documents',
            recordID: '{{recordID}}',
            omitEmpty: true,
            continueOnError: true,
            fields: {
              extract_status: '{{extract_status}}',
              extract_error: '{{extract_error}}',
              extracted_text: '{{extracted_text}}',
              extracted_at: '{{extracted_at}}',
              file_hash: '{{file_hash}}',
            },
          },
        },
      ],
      edges: [
        { from: 'has_file', to: 'pending', condition: 'has_file_result' },
        { from: 'pending', to: 'extract' },
        { from: 'extract', to: 'check_ok' },
        { from: 'extract', to: 'check_fail' },
        { from: 'check_ok', to: 'ai', condition: 'check_ok_result' },
        { from: 'ai', to: 'save_ok' },
        { from: 'check_fail', to: 'save_fail', condition: 'check_fail_result' },
      ],
    },
  ]
}
