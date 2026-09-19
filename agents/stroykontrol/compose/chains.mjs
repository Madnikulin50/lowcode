/**
 * Rule chains that move all document-comparison work out of seed data and into
 * an agent: a deterministic `script` node (real diff/matching) plus an `ai`
 * node (LLM verdict, agent "assistant"), combined by a `score.weighted` node
 * whose factor weights are plain JSON — tune det vs ai influence right in the
 * chain, no code change needed.
 *
 * Shared skeleton per task: [prep nodes] → det (script) → ai → ai_parse
 * (script, parses the ai JSON reply) → combine (score.weighted) → finalize
 * (script, writes everything back via runtime.mcp.*).
 */

function scoreNode (id, label, factors, outScore) {
  return {
    id,
    type: 'score.weighted',
    label,
    config: { factors, normalize: true, scaleMax: 100, outScore },
  }
}

function scriptNode (id, label, code) {
  return { id, type: 'script', label, config: { code } }
}

function aiNode (id, label, prompt) {
  return { id, type: 'ai', label, config: { agent: 'assistant', prompt } }
}

function conditionNode (id, label, field, operator, value) {
  return { id, type: 'condition', label, config: { field, operator, value } }
}

function extractNode (id, label, attachmentField, outText, maxChars = 60000) {
  return { id, type: 'document.extract', label, config: { attachmentField, maxChars, outText } }
}

function crudUpdateNode (id, label, ns, moduleID, moduleHandle, fields) {
  return {
    id,
    type: 'crud',
    label,
    config: {
      operation: 'update',
      namespaceID: ns,
      moduleID: String(moduleID),
      moduleHandle,
      recordID: '{{recordID}}',
      omitEmpty: true,
      continueOnError: true,
      fields,
    },
  }
}

function crudSearchNode (id, label, ns, moduleID, moduleHandle, query, limit = 500) {
  return {
    id,
    type: 'crud',
    label,
    config: { operation: 'search', namespaceID: ns, moduleID: String(moduleID), moduleHandle, query, limit },
  }
}

// Parse the ai node's JSON reply defensively; fall back to the deterministic
// score (from `detNode`) if the model didn't return valid JSON.
function aiParseCode (detNode) {
  return `
var out = { ai_score: null, ai_notes: '', ai_discrepancies: [] };
try {
  var raw = String(ai_response || '').trim();
  var m = raw.match(/\\{[\\s\\S]*\\}/);
  var parsed = JSON.parse(m ? m[0] : raw);
  var score = Number(parsed.score);
  out.ai_score = isNaN(score) ? null : Math.max(0, Math.min(100, score));
  out.ai_notes = String(parsed.notes || '').slice(0, 500);
  out.ai_discrepancies = Array.isArray(parsed.discrepancies) ? parsed.discrepancies.slice(0, 10).map(function (d) {
    return {
      source: 'ai',
      type: String((d && d.type) || 'content'),
      severity: String((d && d.severity) || 'medium'),
      description: String((d && d.description) || '').slice(0, 300),
    };
  }) : [];
} catch (e) {
  // User-facing (this lands verbatim in the record's "comment" field via
  // finalizeCode below) - never leak e.message here: JS/goja exception text
  // for a JSON parse failure on empty input literally reads "EOF" to a
  // human, which is meaningless noise, not a real explanation.
  out.ai_notes = raw
    ? 'вердикт ИИ недоступен: ответ модели не в ожидаемом формате JSON'
    : 'вердикт ИИ недоступен: модель не вернула ответ';
}
if (out.ai_score == null) {
  out.ai_score = (${detNode} && ${detNode}.output && typeof ${detNode}.output.det_score === 'number') ? ${detNode}.output.det_score : 50;
}
return out;
`.trim()
}

function edge (from, to, condition) {
  return condition ? { from, to, condition } : { from, to }
}

// ---------------------------------------------------------------------------
// Chain 1 — ТЗ №1: сравнение ПД и РД
// ---------------------------------------------------------------------------
function compareChain (ns, m, weights) {
  const detCode = `
function normLine (s) { return String(s || '').toLowerCase().replace(/\\s+/g, ' ').trim(); }
function toLines (s) {
  return String(s || '')
    .split(/\\n+|(?<=[.!?])\\s+/)
    .map(normLine)
    .filter(function (l) { return l.length > 2; });
}
var a = toLines(pd_text);
var b = toLines(rd_text);
var bagA = {}, bagB = {};
a.forEach(function (l) { bagA[l] = (bagA[l] || 0) + 1; });
b.forEach(function (l) { bagB[l] = (bagB[l] || 0) + 1; });
var common = 0;
Object.keys(bagA).forEach(function (l) { if (bagB[l]) common += Math.min(bagA[l], bagB[l]); });
var total = Math.max(a.length, b.length, 1);
var detScore = Math.round((common / total) * 1000) / 10;

var bB = Object.assign({}, bagB);
var missing = [];
a.forEach(function (l) { if (bB[l] > 0) { bB[l]--; } else { missing.push(l); } });
var bA = Object.assign({}, bagA);
var extra = [];
b.forEach(function (l) { if (bA[l] > 0) { bA[l]--; } else { extra.push(l); } });

var discrepancies = [];
missing.slice(0, 6).forEach(function (l) { discrepancies.push({ source: 'det', type: 'content', severity: 'medium', description: 'Есть в ПД, нет в РД: ' + l.slice(0, 220) }); });
extra.slice(0, 6).forEach(function (l) { discrepancies.push({ source: 'det', type: 'content', severity: 'medium', description: 'Есть в РД, нет в ПД: ' + l.slice(0, 220) }); });

return {
  det_score: detScore,
  det_lines_pd: a.length,
  det_lines_rd: b.length,
  det_missing_count: missing.length,
  det_extra_count: extra.length,
  det_discrepancies: discrepancies,
  pd_text_short: String(pd_text || '').slice(0, 4000),
  rd_text_short: String(rd_text || '').slice(0, 4000),
};
`.trim()

  const finalizeCode = `
var detOut = (det && det.output) || {};
var aiOut = (ai_parse && ai_parse.output) || {};
var discrepancies = (detOut.det_discrepancies || []).concat(aiOut.ai_discrepancies || []);
var errors = [];

var upd = runtime.mcp.updateRecord(namespaceID, '${m.pd_rd_comparisons}', recordID, {
  status: 'done',
  similarity_percent: final_score,
  matching_pages: Math.max((detOut.det_lines_pd || 0) - (detOut.det_missing_count || 0), 0),
  differing_pages: (detOut.det_missing_count || 0) + (detOut.det_extra_count || 0),
  comment: 'Агент: ' + final_score + '% схожести. ' + (aiOut.ai_notes || ''),
});
if (upd && upd.error) errors.push('updateRecord: ' + upd.error);

discrepancies.slice(0, 12).forEach(function (d) {
  var created = runtime.mcp.createRecord(namespaceID, '${m.pd_rd_discrepancies}', {
    comparison: recordID,
    discrepancy_type: d.type,
    severity: d.severity,
    description: (d.source === 'ai' ? '[ИИ] ' : '[детерм.] ') + d.description,
  });
  if (created && created.error) errors.push('createRecord discrepancy: ' + created.error);
});

return { final_score: final_score, discrepancy_count: discrepancies.length, errors: errors };
`.trim()

  return {
    id: 'stroykontrol-compare-pd-rd',
    name: 'Стройконтроль: сравнить ПД/РД',
    description: 'Агент сравнивает pd_file и rd_file: детерминированный диф текста + вердикт LLM, вес каждого — в узле combine.',
    entryNode: 'cond_pd',
    namespaceID: ns,
    config: {
      triggers: [{
        resourceType: 'compose:record',
        eventType: 'afterCreate,afterUpdate',
        moduleHandle: 'pd_rd_comparisons',
        async: true,
        fileField: 'pd_file',
      }],
    },
    nodes: [
      conditionNode('cond_pd', 'Есть файл ПД', 'pd_file', 'notEmpty'),
      conditionNode('cond_rd', 'Есть файл РД', 'rd_file', 'notEmpty'),
      crudUpdateNode('mark_processing', 'Статус: обрабатывается', ns, m.pd_rd_comparisons, 'pd_rd_comparisons', { status: 'processing' }),
      extractNode('extract_pd', 'Извлечь текст ПД', 'pd_file', 'pd_text'),
      extractNode('extract_rd', 'Извлечь текст РД', 'rd_file', 'rd_text'),
      scriptNode('det', 'Детерминированное сравнение', detCode),
      aiNode('ai', 'Агент: сравнить ПД/РД', 'Ты эксперт производственно-технического отдела. Сравни текст проектной документации (ПД) и рабочей документации (РД) одного объекта. Дай процент схожести 0-100 и до 5 содержательных расхождений (не форматирование). Ответь СТРОГО JSON без markdown: {"score": <0-100>, "notes": "<1-3 предложения по-русски>", "discrepancies": [{"type": "content|numbering|missing_page|extra_page|drawing", "severity": "low|medium|high", "description": "..."}]}\n\nПД:\n{{det.output.pd_text_short}}\n\nРД:\n{{det.output.rd_text_short}}'),
      scriptNode('ai_parse', 'Разбор ответа агента', aiParseCode('det')),
      scoreNode('combine', 'Итоговый вердикт', weights.map(([field, weight]) => ({ field, weight, max: 100 })), 'final_score'),
      scriptNode('finalize', 'Записать результат', finalizeCode),
    ],
    edges: [
      edge('cond_pd', 'cond_rd', 'cond_pd_result'),
      edge('cond_rd', 'mark_processing', 'cond_rd_result'),
      edge('mark_processing', 'extract_pd'),
      edge('extract_pd', 'extract_rd'),
      edge('extract_rd', 'det'),
      edge('det', 'ai'),
      edge('ai', 'ai_parse'),
      edge('ai_parse', 'combine'),
      edge('combine', 'finalize'),
    ],
  }
}

// ---------------------------------------------------------------------------
// Chain 2 — ТЗ №2: проверка ИД по реестру (по объекту, кнопка)
// ---------------------------------------------------------------------------
function checkIdChain (ns, m, weights) {
  const detCode = `
function val (rec, name) {
  // crud search / runtime.mcp.searchRecords return flat {field: value, ...}
  // objects (see recordSetToMap in compose/mcp/bridge.go), not the
  // {values:[{name,value}]} shape the REST API uses.
  var v = rec ? rec[name] : undefined;
  return v == null ? '' : String(v);
}
function norm (s) { return String(s || '').toLowerCase().replace(/[^a-zа-яё0-9]+/gi, ' ').trim(); }
function sim (a, b) {
  a = norm(a); b = norm(b);
  if (!a || !b) return 0;
  if (a === b) return 1;
  var wa = a.split(' ').filter(Boolean), wb = b.split(' ').filter(Boolean);
  var setB = {}; wb.forEach(function (w) { setB[w] = true; });
  var hit = 0; wa.forEach(function (w) { if (setB[w]) hit++; });
  return hit / Math.max(wa.length, wb.length, 1);
}
var registry = (search_registry.records || []).map(function (r) {
  var reqVal = val(r, 'is_required');
  return { id: r.recordID, name: val(r, 'doc_name'), number: val(r, 'doc_number'), version: val(r, 'doc_version'), required: reqVal === 'true' || reqVal === '1' };
});
var files = (search_files.records || []).map(function (r) {
  return { id: r.recordID, name: val(r, 'recognized_name'), number: val(r, 'recognized_number'), version: val(r, 'recognized_version') };
});
var used = {};
var results = [];
files.forEach(function (f) {
  var best = null, bestScore = 0;
  registry.forEach(function (r) {
    if (used[r.id]) return;
    var s = sim(f.name, r.name) * 0.7 + (f.number && f.number === r.number ? 0.2 : 0) + (f.version && f.version === r.version ? 0.1 : 0);
    if (s > bestScore) { bestScore = s; best = r; }
  });
  var status;
  if (!best || bestScore < 0.35) {
    status = 'not_in_registry';
  } else {
    used[best.id] = true;
    if (bestScore >= 0.9 && (!f.number || f.number === best.number) && (!f.version || f.version === best.version)) status = 'matched';
    else if (best.number && f.number && best.number !== f.number) status = 'number_mismatch';
    else if (best.version && f.version && best.version !== f.version) status = 'version_mismatch';
    else status = 'name_mismatch';
  }
  results.push({ fileID: f.id, registryID: best ? best.id : '', status: status, score: Math.round(bestScore * 100) });
});
var matched = results.filter(function (r) { return r.status === 'matched'; }).length;
var mismatched = results.filter(function (r) { return r.status !== 'matched' && r.status !== 'not_in_registry'; }).length + results.filter(function(r){return r.status==='not_in_registry';}).length;
var missing = registry.filter(function (r) { return r.required && !used[r.id]; }).length;
var detScore = registry.length ? Math.round((matched / registry.length) * 1000) / 10 : 100;
var summary = results.slice(0, 15).map(function (r) {
  var f = files.filter(function (x) { return x.id === r.fileID; })[0];
  return (f ? f.name : '?') + ' → ' + r.status + ' (' + r.score + '%)';
}).join('; ');
return {
  det_results: results,
  det_registry_count: registry.length,
  det_files_count: files.length,
  det_matched: matched,
  det_mismatched: mismatched,
  det_missing: missing,
  det_score: detScore,
  det_summary: summary,
};
`.trim()

  const finalizeCode = `
var detOut = (det && det.output) || {};
var aiOut = (ai_parse && ai_parse.output) || {};
var errors = [];

(detOut.det_results || []).forEach(function (r) {
  var upd = runtime.mcp.updateRecord(namespaceID, '${m.id_files}', r.fileID, {
    match_status: r.status,
    matched_registry: r.registryID || '',
  });
  if (upd && upd.error) errors.push('updateRecord id_files ' + r.fileID + ': ' + upd.error);
});

var run = runtime.mcp.createRecord(namespaceID, '${m.id_check_runs}', {
  object: recordID,
  run_date: new Date().toISOString(),
  total_registry_docs: detOut.det_registry_count || 0,
  uploaded_docs: detOut.det_matched || 0,
  missing_docs: detOut.det_missing || 0,
  mismatched_docs: detOut.det_mismatched || 0,
  overall_status: (detOut.det_missing === 0 && detOut.det_mismatched === 0) ? 'success' : 'fail',
});
if (run && run.error) errors.push('createRecord id_check_runs: ' + run.error);

return { final_score: final_score, checked: (detOut.det_results || []).length, notes: aiOut.ai_notes || '', errors: errors };
`.trim()

  return {
    id: 'stroykontrol-check-id',
    name: 'Стройконтроль: проверить ИД по реестру',
    description: 'Агент сопоставляет загруженные файлы ИД объекта с реестром: детерминированное сравнение наименований/номеров/версий + вердикт LLM.',
    entryNode: 'search_registry',
    namespaceID: ns,
    nodes: [
      crudSearchNode('search_registry', 'Реестр объекта', ns, m.id_registry, 'id_registry', "object = '{{recordID}}'"),
      crudSearchNode('search_files', 'Файлы объекта', ns, m.id_files, 'id_files', "object = '{{recordID}}'"),
      scriptNode('det', 'Детерминированное сопоставление', detCode),
      aiNode('ai', 'Агент: комплектность ИД', 'Ты эксперт ПТО. По сводке сопоставления файлов исполнительной документации с реестром оцени уверенность в комплектности (0-100) и дай короткий вывод. Сводка (файл → статус сопоставления): {{det.output.det_summary}}\nВсего в реестре: {{det.output.det_registry_count}}, загружено файлов: {{det.output.det_files_count}}, совпало: {{det.output.det_matched}}.\nОтветь СТРОГО JSON без markdown: {"score": <0-100>, "notes": "<1-3 предложения по-русски>", "discrepancies": []}'),
      scriptNode('ai_parse', 'Разбор ответа агента', aiParseCode('det')),
      scoreNode('combine', 'Итоговый вердикт', weights.map(([field, weight]) => ({ field, weight, max: 100 })), 'final_score'),
      scriptNode('finalize', 'Записать результат', finalizeCode),
    ],
    edges: [
      edge('search_registry', 'search_files'),
      edge('search_files', 'det'),
      edge('det', 'ai'),
      edge('ai', 'ai_parse'),
      edge('ai_parse', 'combine'),
      edge('combine', 'finalize'),
    ],
  }
}

// ---------------------------------------------------------------------------
// Chain 3 — ТЗ №3: проверка ВОИСР против сметы (по позиции ВОИСР)
// ---------------------------------------------------------------------------
function checkVoisrChain (ns, m, weights) {
  const detCode = `
function val (rec, name) {
  // crud search / runtime.mcp.searchRecords return flat {field: value, ...}
  // objects (see recordSetToMap in compose/mcp/bridge.go), not the
  // {values:[{name,value}]} shape the REST API uses.
  var v = rec ? rec[name] : undefined;
  return v == null ? '' : String(v);
}
var smetaID = String(smeta_ref || '');
var smetaRows = smetaID ? runtime.mcp.searchRecords(namespaceID, '${m.smeta_items}', "recordID = '" + smetaID + "'", 1) : [];
var notFound = !smetaRows || !smetaRows.length;
var smetaVolume = notFound ? null : (Number(val(smetaRows[0], 'volume')) || 0);
var smetaAmount = notFound ? null : (Number(val(smetaRows[0], 'total_amount')) || 0);
var vVolume = Number(volume) || 0;
var vAmount = Number(amount) || 0;
var volumeDiff = notFound ? null : Math.round((vVolume - smetaVolume) * 1000) / 1000;
var amountDiff = notFound ? null : Math.round((vAmount - smetaAmount) * 100) / 100;
var amountDiffPct = (notFound || !smetaAmount) ? null : Math.round((amountDiff / smetaAmount) * 1000) / 10;
var detScore = notFound ? 0 : Math.max(0, 100 - Math.min(100, Math.abs(amountDiffPct)));
return {
  det_not_found: notFound,
  det_smeta_volume: smetaVolume,
  det_smeta_amount: smetaAmount,
  det_volume_diff: volumeDiff,
  det_amount_diff: amountDiff,
  det_amount_diff_percent: amountDiffPct,
  det_score: Math.round(detScore * 10) / 10,
  det_work_name: notFound ? '' : val(smetaRows[0], 'work_name'),
};
`.trim()

  const finalizeCode = `
var detOut = (det && det.output) || {};
var aiOut = (ai_parse && ai_parse.output) || {};
var status = detOut.det_not_found ? 'not_found' : (final_score >= 97 ? 'match' : 'mismatch');

var created = runtime.mcp.createRecord(namespaceID, '${m.voisr_check_results}', {
  voisr_item: recordID,
  smeta_volume: detOut.det_smeta_volume,
  smeta_amount: detOut.det_smeta_amount,
  volume_diff: detOut.det_volume_diff,
  amount_diff: detOut.det_amount_diff,
  amount_diff_percent: detOut.det_amount_diff_percent,
  status: status,
  note: 'Агент (' + final_score + '/100): ' + (aiOut.ai_notes || ''),
});

return { status: status, final_score: final_score, error: (created && created.error) || '' };
`.trim()

  return {
    id: 'stroykontrol-check-voisr',
    name: 'Стройконтроль: проверить позицию ВОИСР',
    description: 'Агент сверяет позицию ВОИСР со сметой: арифметика (объём/сумма) + вердикт LLM о характере расхождения.',
    entryNode: 'det',
    namespaceID: ns,
    config: {
      triggers: [{
        resourceType: 'compose:record',
        eventType: 'afterCreate,afterUpdate',
        moduleHandle: 'voisr_items',
        async: true,
      }],
    },
    nodes: [
      scriptNode('det', 'Сверка со сметой', detCode),
      aiNode('ai', 'Агент: оценка расхождения ВОИСР', 'Ты сметчик-контролёр. Позиция ВОИСР «{{det.output.det_work_name}}»: объём по смете {{det.output.det_smeta_volume}}, по ВОИСР {{volume}}; сумма по смете {{det.output.det_smeta_amount}}, по ВОИСР {{amount}}; разница {{det.output.det_amount_diff_percent}}%. Оцени по шкале 0-100 (100 = полностью соответствует смете) и в 1-2 предложениях скажи, похоже ли расхождение на техническую погрешность или на завышение объёмов/стоимости. Ответь СТРОГО JSON без markdown: {"score": <0-100>, "notes": "...", "discrepancies": []}'),
      scriptNode('ai_parse', 'Разбор ответа агента', aiParseCode('det')),
      scoreNode('combine', 'Итоговый вердикт', weights.map(([field, weight]) => ({ field, weight, max: 100 })), 'final_score'),
      scriptNode('finalize', 'Записать результат', finalizeCode),
    ],
    edges: [
      edge('det', 'ai'),
      edge('ai', 'ai_parse'),
      edge('ai_parse', 'combine'),
      edge('combine', 'finalize'),
    ],
  }
}

/**
 * @param weights per-chain [detWeight, aiWeight] pairs (0..1, need not sum to 1 —
 *   score.weighted normalizes). Tune directly here, or edit the deployed
 *   chain's `combine` node `factors[].weight` in place.
 */
export function buildRuleChains ({ nsID, modules, weights = {} }) {
  const w = {
    compare: weights.compare || [0.5, 0.5],
    checkId: weights.checkId || [0.6, 0.4],
    checkVoisr: weights.checkVoisr || [0.7, 0.3],
  }
  return [
    compareChain(nsID, modules, [['det.output.det_score', w.compare[0]], ['ai_parse.output.ai_score', w.compare[1]]]),
    checkIdChain(nsID, modules, [['det.output.det_score', w.checkId[0]], ['ai_parse.output.ai_score', w.checkId[1]]]),
    checkVoisrChain(nsID, modules, [['det.output.det_score', w.checkVoisr[0]], ['ai_parse.output.ai_score', w.checkVoisr[1]]]),
  ]
}
