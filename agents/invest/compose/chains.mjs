export function buildRuleChains ({ nsID, modules, engineUrl }) {
  const ns = String(nsID)
  const docs = String(modules.documents)
  const rfc = String(modules.change_requests)
  const log = String(modules.change_log)
  const risks = String(modules.risks)

  const upd = (id, name, moduleID, handle, fields, description) => ({
    id,
    name,
    description,
    entryNode: 'upd',
    namespaceID: ns,
    nodes: [{
      id: 'upd',
      type: 'crud',
      label: name,
      config: {
        operation: 'update',
        namespaceID: ns,
        moduleID,
        moduleHandle: handle,
        recordID: '{{recordID}}',
        fields,
      },
    }],
    edges: [],
  })

  return [
    upd('invest-submit-approval', 'Инвест: отправить документ на согласование', docs, 'documents',
      { status: 'in_review' }, 'Статус документа → на согласовании.'),
    upd('invest-approve-document', 'Инвест: утвердить документ', docs, 'documents',
      { status: 'approved' }, 'Статус документа → утверждён.'),
    upd('invest-reject-document', 'Инвест: отклонить документ', docs, 'documents',
      { status: 'rejected' }, 'Статус документа → отклонён.'),
    upd('invest-escalate-approval', 'Инвест: эскалировать согласование', docs, 'documents',
      { status: 'in_review', notes: 'Эскалация: {{notes}}' }, 'Помечает согласование как эскалированное (статус остаётся на согласовании).'),
    upd('invest-submit-rfc', 'Инвест: отправить RFC', rfc, 'change_requests',
      { status: 'in_review' }, 'RFC → на согласовании.'),
    {
      id: 'invest-approve-rfc',
      name: 'Инвест: утвердить RFC',
      description: 'RFC → утверждён, строка в журнале изменений.',
      entryNode: 'upd',
      namespaceID: ns,
      nodes: [
        {
          id: 'upd',
          type: 'crud',
          label: 'Утвердить RFC',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: rfc,
            moduleHandle: 'change_requests',
            recordID: '{{recordID}}',
            fields: { status: 'approved' },
          },
        },
        {
          id: 'log',
          type: 'crud',
          label: 'Журнал изменений',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: log,
            moduleHandle: 'change_log',
            fields: {
              rfc: '{{recordID}}',
              project: '{{project}}',
              summary: 'Утверждён RFC: {{title}}',
              new_budget: '{{eac_after}}',
              old_budget: '{{eac_before}}',
            },
          },
        },
      ],
      edges: [{ from: 'upd', to: 'log' }],
    },
    upd('invest-reject-rfc', 'Инвест: отклонить RFC', rfc, 'change_requests',
      { status: 'rejected' }, 'RFC → отклонён.'),
    {
      id: 'invest-recalculate-evm',
      name: 'Инвест: пересчитать EVM',
      description: 'POST /recalculate-evm на invest-engine. Опционально projectID в контексте.',
      entryNode: 'http',
      namespaceID: ns,
      nodes: [{
        id: 'http',
        type: 'http',
        label: 'engine EVM',
        config: {
          url: engineUrl + '/recalculate-evm',
          method: 'POST',
          body: '{"namespaceID":{{namespaceID}},"projectID":"{{projectID}}","token":"{{authToken}}"}',
          timeout: 60,
        },
      }],
      edges: [],
    },
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
    {
      id: 'invest-threshold-alert',
      name: 'Инвест: пороговые алерты',
      description: 'Просроченные документы на согласовании и WBS с CPI < 0.9 → риски. Движок + поиск.',
      entryNode: 'http',
      namespaceID: ns,
      nodes: [
        {
          id: 'http',
          type: 'http',
          label: 'engine alerts',
          config: {
            url: engineUrl + '/alerts',
            method: 'POST',
            body: '{"namespaceID":{{namespaceID}},"projectID":"{{projectID}}","cpiThreshold":0.9,"token":"{{authToken}}"}',
            timeout: 60,
          },
        },
        {
          id: 'overdue',
          type: 'crud',
          label: 'Документы на согласовании',
          config: {
            operation: 'search',
            namespaceID: ns,
            moduleID: docs,
            moduleHandle: 'documents',
            query: "status = 'in_review'",
            limit: 50,
          },
        },
      ],
      edges: [{ from: 'http', to: 'overdue' }],
    },
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
              status: 'open',
              description: 'Автоалерт: CPI={{cpi}} на работе {{code}} {{name}}.',
            },
          },
        },
      ],
      edges: [{ from: 'check', to: 'risk', condition: 'check_result' }],
    },
  ]
}
