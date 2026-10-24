#!/usr/bin/env node
/**
 * Generates the PDF exports of remarks used by the tests of the PDF import (cmp/expertise_pdf_test.go):
 *   cmp/testdata/remarks-pdf/table.pdf — a landscape table «№ | Раздел | Замечание | Ответ проектировщика»
 *     over two pages (the header repeated), with wrapped cells;
 *   cmp/testdata/remarks-pdf/list.pdf  — a letter with numbered remarks and «Ответ:» paragraphs.
 * Needs a headless Chrome / Chromium (CHROME_BIN):
 *
 *   CHROME_BIN=/opt/pw-browsers/chromium-1194/chrome-linux/chrome node fixtures/gen_remarks_pdf.mjs
 */
import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const HERE = dirname(fileURLToPath(import.meta.url))
const CHROME = process.env.CHROME_BIN || 'google-chrome'
const OUT = join(HERE, '..', 'cmp', 'testdata', 'remarks-pdf')
const esc = s => String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

function print (name, html) {
  const dir = mkdtempSync(join(tmpdir(), 'rem-pdf-'))
  try {
    const src = join(dir, 'in.html')
    writeFileSync(src, html)
    execFileSync(CHROME, ['--headless=new', '--no-sandbox', '--disable-gpu', '--no-pdf-header-footer',
      `--print-to-pdf=${join(OUT, name)}`, `file://${src}`], { stdio: 'ignore' })
  } finally {
    rmSync(dir, { recursive: true, force: true })
  }
  console.log('wrote', join(OUT, name))
}

const REMARKS = [
  ['1', '2024-15-КЖ2', 'Глубина заложения фундамента в пояснительной записке (2,4 м) не соответствует чертежу КЖ2 лист 3 (2,6 м). Привести в соответствие.', 'Исправлено: в ПЗ принято 2,6 м, том 2024-15-КЖ2 изм. 1.'],
  ['2', '2024-15-КЖ2', 'Не указан класс бетона по водонепроницаемости для фундаментной плиты.', 'Принято W6, см. лист 2.'],
  ['3', '2024-15-АР', 'Отсутствует ведомость отделки помещений.', ''],
  ['4', '2024-15-АР', 'Предел огнестойкости стен лестничной клетки REI 90 не соответствует требуемому REI 120 по СП 2.13130.2020, п. 5.4.3. Обосновать или исправить.', 'Предел огнестойкости увеличен до REI 120, лист АР-7.'],
  ['5', '2024-15-ТХ', 'В технологических решениях не приведены сведения о численности персонала.', 'Сведения приведены в разделе 3.'],
  ['6', '2024-15-КЖ2', 'Не обоснован защитный слой бетона 40 мм для конструкций, соприкасающихся с грунтом.', ''],
  ['7', '2024-15-КЖ2', 'Шаг арматуры фундаментной плиты на листе 5 (200 мм) не совпадает со спецификацией (150 мм).', 'Исправлено, шаг 200 мм в спецификации.'],
  ['8', '2024-15-АР', 'Не указан уклон кровли.', 'Уклон кровли 2 %, лист АР-9.'],
]

const css = `@page { size: A4 landscape; margin: 15mm }
body { font-family: 'DejaVu Sans', Arial, sans-serif; font-size: 10pt }
table { border-collapse: collapse; width: 100% }
thead { display: table-header-group }
td, th { border: 1px solid #000; padding: 3pt 4pt; vertical-align: top; text-align: left }
tr { page-break-inside: avoid }`

mkdirSync(OUT, { recursive: true })
const rows = [...REMARKS, ...REMARKS.map(([n, s, t, a]) => [String(Number(n) + 8), s, t.replace('.', ' (повторно).'), a])]
print('table.pdf', `<!doctype html><meta charset="utf-8"><style>${css}</style>
<h3>Замечания государственной экспертизы по объекту «Административное здание»</h3>
<table><colgroup><col style="width:5%"><col style="width:14%"><col style="width:46%"><col style="width:35%"></colgroup>
<thead><tr><th>№</th><th>Раздел</th><th>Замечание</th><th>Ответ проектировщика</th></tr></thead>
<tbody>${rows.map(r => `<tr>${r.map(c => `<td>${esc(c)}</td>`).join('')}</tr>`).join('')}</tbody></table>`)

print('list.pdf', `<!doctype html><meta charset="utf-8"><style>@page { size: A4; margin: 20mm } body { font-family: 'DejaVu Sans', Arial, sans-serif; font-size: 11pt } p { margin: 4pt 0 }</style>
<p>Перечень замечаний экспертизы по проектной документации</p>
<p>Раздел 4 КЖ</p>
${REMARKS.slice(0, 3).map(([n, s, t, a]) => `<p>${n}. ${esc(t)} Том ${s}.</p>${a ? `<p>Ответ: ${esc(a)}</p>` : ''}`).join('\n')}
<p>Раздел 3 АР</p>
${REMARKS.slice(3, 5).map(([n, s, t, a], i) => `<p>${i + 4}. ${esc(t)}</p>${a ? `<p>Ответ: ${esc(a)}</p>` : ''}`).join('\n')}`)
