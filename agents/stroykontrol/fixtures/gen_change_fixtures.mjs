#!/usr/bin/env node
/**
 * Generates fixtures/changes-1: two editions of one volume (Изм. 0 → Изм. 1) for the check of the
 * change registration (ТЗ 3.3.7, 3.3.8). Needs google-chrome or chromium (headless print-to-PDF).
 *
 *   CHROME_BIN=/path/to/chrome node fixtures/gen_change_fixtures.mjs
 *
 * The title block of every sheet is the form for the next sheets of a text document (ГОСТ Р 21.101):
 * the change rows above the column heads «Изм. Кол.уч. Лист № док. Подп. Дата» and the sheet number.
 *
 * Edition 1 against edition 0:
 *   sheet 1 — unchanged;
 *   sheet 2 — changed, declared («изменённых 2», «1 - Изм.» in the title block);
 *   sheet 3 — changed, NOT declared;
 *   sheet 4 — annulled (declared «аннулированных 4»);
 *   sheet 5 — declared replaced («Зам.»), but nothing differs;
 *   sheet 6 — new (declared), sheet 7 — the table of the change registration itself (new, declared).
 */
import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const HERE = dirname(fileURLToPath(import.meta.url))
const CHROME = process.env.CHROME_BIN || 'google-chrome'
const CODE = '2024-15-КЖ2'

const esc = s => String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

// one sheet: body lines (an array is a table row), the sheet number and the change rows of its title block
function sheet ({ n, title, lines, changes = [] }) {
  const rows = lines.filter(Array.isArray)
  const text = lines.filter(l => !Array.isArray(l)).map(l => `<p>${esc(l)}</p>`).join('\n')
  const table = rows.length
    ? `<table>${rows.map(r => `<tr>${r.map(c => `<td>${esc(c)}</td>`).join('')}</tr>`).join('')}</table>`
    : ''
  const ch = changes.map(c => `<div class="row">${c.map(x => `<span>${esc(x)}</span>`).join('')}</div>`).join('')
  return `<div class="page">
${title ? `<h2>${esc(title)}</h2>` : ''}
${text}
${table}
<div class="stamp">${ch}<div class="row"><span>Изм.</span><span>Кол.уч.</span><span>Лист</span><span>№ док.</span><span>Подп.</span><span>Дата</span><span class="code">${esc(CODE)}</span><span class="no">${n}</span></div></div>
</div>`
}

function html (bodies) {
  return `<!doctype html><html lang="ru"><head><meta charset="utf-8"><style>
@page { size: 210mm 297mm; margin: 0 }
body { margin: 0; font: 13px/1.5 "DejaVu Sans", sans-serif }
.page { position: relative; width: 210mm; height: 297mm; box-sizing: border-box; padding: 16mm 14mm 30mm 22mm; page-break-after: always; overflow: hidden }
h2 { font-size: 16px; margin: 0 0 10px }
p { margin: 2px 0 }
table { border-collapse: collapse; margin-top: 8px }
td { border: 1px solid #000; padding: 2px 8px }
.stamp { position: absolute; right: 6mm; bottom: 6mm; width: 180mm; border: 1px solid #000; font-size: 10px }
.stamp .row { display: flex; gap: 6mm; padding: 1mm 2mm }
.stamp .code { margin-left: auto; font-size: 12px }
.stamp .no { width: 10mm; text-align: right }
</style></head><body>${bodies.join('\n')}</body></html>`
}

function pdf (file, sheets) {
  const dir = mkdtempSync(join(tmpdir(), 'cmp-chg-'))
  try {
    const src = join(dir, 'in.html')
    writeFileSync(src, html(sheets.map(sheet)))
    mkdirSync(dirname(file), { recursive: true })
    execFileSync(CHROME, ['--headless=new', '--no-sandbox', '--disable-gpu', '--no-pdf-header-footer',
      `--print-to-pdf=${file}`, `file://${src}`], { stdio: 'ignore' })
  } finally {
    rmSync(dir, { recursive: true, force: true })
  }
  console.log('wrote', file)
}

const s1 = ['1.1 Фундаменты выполнить из бетона класса В25, W6, F150 по подготовке толщиной 100 мм.', '1.2 Рабочая арматура класса А500С, конструктивная арматура класса А240.',
  '1.3 Отметка низа фундаментной плиты минус 3,200 соответствует абсолютной отметке 152,40.', '1.4 Основанием служат суглинки тугопластичные с модулем деформации 18 МПа.']
const s2 = ['2.1 Толщина фундаментной плиты 600 мм, защитный слой нижней арматуры 50 мм.', '2.2 Шаг стержней нижней сетки 200 мм в обоих направлениях по всей площади плиты.',
  '2.3 Дополнительное армирование под колоннами выполнить по схеме узла.', '2.4 Перепуск стержней выполнить внахлёст длиной 40 диаметров.']
const s2b = ['2.1 Толщина фундаментной плиты 700 мм, защитный слой нижней арматуры 50 мм.', s2[1], s2[2], s2[3]]
const s3 = ['3.1 Стены подвала толщиной 300 мм из бетона класса В25 по водонепроницаемости W6.', '3.2 Вертикальная гидроизоляция обмазочная в два слоя по огрунтованной поверхности.',
  '3.3 Деформационные швы устраивать через 30 м с установкой гидрошпонок.', '3.4 Обратную засыпку пазух выполнять послойно с уплотнением до коэффициента 0,95.']
const s3b = [s3[0], s3[1], '3.3 Деформационные швы устраивать через 24 м с установкой гидрошпонок.', s3[3]]
const s4 = ['4.1 Приямки выполнить из монолитного железобетона по месту с привязкой к осям здания.', '4.2 Крышки приямков стальные рифлёные с антикоррозионным покрытием.',
  '4.3 Отвод воды из приямков выполнить в дренажную систему здания по проекту НВК.', '4.4 Ограждение приямков высотой 1,2 м по серии типовых узлов.']
const s5 = ['5.1 Закладные детали М-1 устанавливать до бетонирования по схеме расстановки.', '5.2 Сварные швы выполнять электродами Э42А с катетом не менее 6 мм.',
  '5.3 Антикоррозионную защиту закладных деталей выполнить горячим цинкованием.', '5.4 Отклонение закладных деталей от проектного положения не более 5 мм.']
const s6 = ['6.1 Пандус у входа в подвал выполнить из монолитного бетона класса В25.', '6.2 Уклон пандуса не более 1:12, покрытие с противоскользящей обработкой.',
  '6.3 Армирование пандуса сеткой из арматуры класса А500С с шагом 150 мм.', '6.4 Поручни пандуса стальные на высоте 0,9 м с двух сторон по всей длине.']

const before = [
  { n: 1, title: 'Конструктивные решения. Фундаменты', lines: s1 },
  { n: 2, lines: s2 },
  { n: 3, lines: s3 },
  { n: 4, lines: s4 },
  { n: 5, lines: s5 },
]
const after = [
  { n: 1, title: 'Конструктивные решения. Фундаменты', lines: s1 },
  { n: 2, lines: s2b, changes: [['1', '-', 'Изм.', '15-24', '', '01.24']] },
  { n: 3, lines: s3b },
  { n: 5, lines: s5, changes: [['1', '-', 'Зам.', '15-24', '', '01.24']] },
  { n: 6, lines: s6, changes: [['1', '-', 'Нов.', '15-24', '', '01.24']] },
  {
    n: 7,
    title: 'Таблица регистрации изменений',
    lines: [
      ['Изм.', 'Номера листов (страниц)', '', '', '', 'Всего листов (страниц) в док.', 'Номер док.', 'Подп.', 'Дата'],
      ['', 'изменённых', 'заменённых', 'новых', 'аннулированных', '', '', '', ''],
      ['1', '2', '5', '6, 7', '4', '6', '15-24', '', '01.24'],
    ],
  },
]

pdf(join(HERE, 'changes-1', 'pd-kzh2.pdf'), before)
pdf(join(HERE, 'changes-1', 'rd-kzh2 изм. 1.pdf'), after)
writeFileSync(join(HERE, 'changes-1', 'meta.json'), JSON.stringify({ title: 'КЖ2 · изм. 0 → изм. 1 (регистрация изменений)', kind: 'versions' }) + '\n')
