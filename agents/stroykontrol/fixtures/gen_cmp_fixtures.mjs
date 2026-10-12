#!/usr/bin/env node
/**
 * Generates the PDF packs used by the pack-comparison fixtures and the cmp
 * tests (fixtures/pack-1, fixtures/pack-2). Needs google-chrome (headless
 * print-to-PDF — gives real text layers with Cyrillic).
 *
 *   node fixtures/gen_cmp_fixtures.mjs
 *
 * Each volume is a list of sheets; every sheet carries a title block
 * (основная надпись) in the bottom-right corner with the volume шифр, the
 * sheet number and the stage, like a real drawing sheet.
 */
import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const HERE = dirname(fileURLToPath(import.meta.url))
const CHROME = process.env.CHROME_BIN || 'google-chrome'

const esc = s => String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

function sheet (stage, code, n, total, title, lines) {
  const rows = lines.filter(Array.isArray)
  const text = lines.filter(l => !Array.isArray(l)).map(l => `<p>${esc(l)}</p>`).join('\n')
  const table = rows.length
    ? `<table>${rows.map(r => `<tr>${r.map(c => `<td>${esc(c)}</td>`).join('')}</tr>`).join('')}</table>`
    : ''
  return `<div class="page">
${title ? `<h2>${esc(title)}</h2>` : ''}
${text}
${table}
<div class="stamp"><span>${esc(code)}</span><span>Стадия ${esc(stage)}</span><span>Лист ${n}</span><span>Листов ${total}</span></div>
</div>`
}

function volume (stage, code, title, sheets) {
  return sheets.map((lines, i) => sheet(stage, code, i + 1, sheets.length, i === 0 ? title : '', lines)).join('\n')
}

function html (bodies) {
  return `<!doctype html><html lang="ru"><head><meta charset="utf-8"><style>
@page { size: 297mm 210mm; margin: 0 }
body { margin: 0; font: 13px/1.5 "DejaVu Sans", sans-serif }
.page { position: relative; width: 297mm; height: 210mm; box-sizing: border-box; padding: 14mm 16mm; page-break-after: always; overflow: hidden }
h2 { font-size: 16px; margin: 0 0 10px }
p { margin: 2px 0 }
table { border-collapse: collapse; margin-top: 8px }
td { border: 1px solid #000; padding: 2px 10px }
.stamp { position: absolute; right: 8mm; bottom: 8mm; width: 160mm; height: 12mm; border: 1px solid #000; display: flex; justify-content: space-around; align-items: center }
</style></head><body>${bodies.join('\n')}</body></html>`
}

function pdf (file, bodies) {
  const dir = mkdtempSync(join(tmpdir(), 'cmp-fx-'))
  try {
    const src = join(dir, 'in.html')
    writeFileSync(src, html(bodies))
    mkdirSync(dirname(file), { recursive: true })
    execFileSync(CHROME, ['--headless=new', '--no-sandbox', '--disable-gpu', '--no-pdf-header-footer',
      `--print-to-pdf=${file}`, `file://${src}`], { stdio: 'ignore' })
  } finally {
    rmSync(dir, { recursive: true, force: true })
  }
  console.log('wrote', file)
}

// ---------------------------------------------------------------- pack-1 --
// ПД: one PDF holding three volumes (КЖ, АР, ТХ).
// РД: КЖ with edits + extra detail, АР with a renumbered clause; ТХ is not
// released yet (no file).
const kzhPD = [
  [
    '1. Общие данные',
    '1.1 Фундаменты выполнить из бетона класса В25, W6, F150.',
    '1.2 Рабочая арматура класса А500С.',
    '1.3 Глубина заложения фундамента: 2,4 м.',
    '1.4 Толщина стен подвала: 300 мм.',
    '1.5 Защитный слой бетона: 40 мм.',
  ],
  [
    '2. Ведомость элементов',
    ['Поз.', 'Наименование', 'Ед.', 'Кол-во'],
    ['Ф-1', 'Фундаментная плита', 'шт', '1'],
    ['Ф-2', 'Ленточный фундамент', 'м', '48'],
    ['С-1', 'Стена подвала', 'м3', '36,5'],
    '2.1 Монолитные участки выполнить по узлам листа 3.',
  ],
  [
    '3. Узлы',
    '3.1 Узел примыкания стены к плите армировать сетками С-1.',
    '3.2 Гидроизоляция: обмазочная в два слоя.',
    '3.3 Бетонную подготовку выполнить толщиной 100 мм.',
  ],
]
const arPD = [
  [
    '1. Архитектурные решения',
    '1.1 Наружные стены: кирпич облицовочный 120 мм.',
    '1.2 Утеплитель: минераловатные плиты 150 мм.',
    '1.3 Кровля плоская с внутренним водостоком.',
  ],
  [
    '2. Отделка',
    '2.1 Полы: керамогранит в санузлах.',
    '2.2 Потолки: подвесные типа Армстронг.',
    '2.3 Окна: ПВХ двухкамерные.',
  ],
]
const txPD = [
  [
    '1. Технологические решения',
    '1.1 Производительность линии: 120 т/сутки.',
    '1.2 Давление в магистрали: 1,6 МПа.',
    '1.3 Резервуар хранения: 50 м3.',
  ],
  [
    '2. Оборудование',
    ['Поз.', 'Оборудование', 'Кол-во'],
    ['Н-1', 'Насос центробежный', '2'],
    ['Р-1', 'Резервуар вертикальный', '1'],
  ],
]

const kzhRD = [
  [
    '1. Общие данные',
    '1.1 Фундаменты выполнить из бетона класса В30, W6, F150.',
    '1.2 Рабочая арматура класса А500С.',
    '1.3 Глубина заложения фундамента: 2,6 м.',
    '1.4 Толщина стен подвала: 300 мм.',
    '1.5 Гидроизоляция по рабочим чертежам марки КЖ.2.',
  ],
  [
    '2. Ведомость элементов',
    ['Поз.', 'Наименование', 'Ед.', 'Кол-во'],
    ['Ф-1', 'Фундаментная плита', 'шт', '1'],
    ['Ф-2', 'Ленточный фундамент', 'м', '52'],
    ['С-1', 'Стена подвала', 'м3', '36,5'],
    '2.1 Монолитные участки выполнить по узлам листа 3.',
    '2.2 Закладные детали М-1 устанавливать по схеме расстановки.',
  ],
  [
    '3. Узлы',
    '3.1 Узел примыкания стены к плите армировать сетками С-1.',
    '3.2 Гидроизоляция: обмазочная в два слоя.',
    '3.3 Бетонную подготовку выполнить толщиной 100 мм.',
  ],
]
const arRD = [
  [
    '1. Архитектурные решения',
    '1.1 Наружные стены: кирпич облицовочный 120 мм.',
    '1.2 Утеплитель: минераловатные плиты 150 мм.',
    '1.3 Кровля плоская с внутренним водостоком.',
  ],
  [
    '2. Отделка',
    '2.5 Полы: керамогранит в санузлах.',
    '2.2 Потолки: подвесные типа Армстронг.',
    '2.3 Окна: ПВХ двухкамерные.',
  ],
]

pdf(join(HERE, 'pack-1/pd-package.pdf'), [
  volume('П', '2024-15-КЖ', 'Конструктивные решения', kzhPD),
  volume('П', '2024-15-АР', 'Архитектурные решения', arPD),
  volume('П', '2024-15-ТХ', 'Технологические решения', txPD),
])
pdf(join(HERE, 'pack-1/rd-kzh.pdf'), [volume('Р', '2024-15-КЖ1', 'Конструктивные решения', kzhRD)])
pdf(join(HERE, 'pack-1/rd-ar.pdf'), [volume('Р', '2024-15-АР1', 'Архитектурные решения', arRD)])

// pack-1-v2: the next iteration of pack-1 (meta.json: parent = pack-1). The
// designer fixed the foundation depth and the concrete grade, left the
// ленточный фундамент quantity (52) as it was and made a new slip (wall 350).
const kzhRD2 = kzhRD.map(sheet => sheet.map(l => {
  if (typeof l !== 'string') return l
  return l
    .replace('класса В30', 'класса В25')
    .replace('2,6 м', '2,4 м')
    .replace('Толщина стен подвала: 300 мм', 'Толщина стен подвала: 350 мм')
}))
pdf(join(HERE, 'pack-1-v2/pd-package.pdf'), [
  volume('П', '2024-15-КЖ', 'Конструктивные решения', kzhPD),
  volume('П', '2024-15-АР', 'Архитектурные решения', arPD),
  volume('П', '2024-15-ТХ', 'Технологические решения', txPD),
])
pdf(join(HERE, 'pack-1-v2/rd-kzh.pdf'), [volume('Р', '2024-15-КЖ1', 'Конструктивные решения', kzhRD2)])
pdf(join(HERE, 'pack-1-v2/rd-ar.pdf'), [volume('Р', '2024-15-АР1', 'Архитектурные решения', arRD)])

// ---------------------------------------------------------------- pack-2 --
// One ПД volume covers two structures; РД exists for structure 1 only and
// is more detailed. Nothing here may be reported above "low".
const isPD = [
  [
    '1. Искусственные сооружения',
    '1.1 Сооружение 1: путепровод через автодорогу, длина 84 м.',
    '1.2 Пролётное строение: сталежелезобетонное, класс бетона В35.',
    '1.3 Опоры: монолитные железобетонные, арматура А500С.',
  ],
  [
    '2. Сооружение 2',
    '2.1 Сооружение 2: мост через реку, длина 126 м.',
    '2.2 Пролётное строение: железобетонное, класс бетона В40.',
    '2.3 Опоры на буронабивных сваях диаметром 1,2 м.',
  ],
]
const isRD = [
  [
    '1. Путепровод',
    '1.1 Сооружение 1: путепровод через автодорогу, длина 84 м.',
    '1.2 Пролётное строение: сталежелезобетонное, класс бетона В35.',
    '1.3 Опоры: монолитные железобетонные, арматура А500С.',
    '1.4 Деформационные швы: резиновые, тип ДШ-60.',
    '1.5 Ограждение: барьерное, высота 1,1 м.',
  ],
]
pdf(join(HERE, 'pack-2/pd-is.pdf'), [volume('П', '2024-20-ИС', 'Искусственные сооружения', isPD)])
pdf(join(HERE, 'pack-2/rd-is1.pdf'), [volume('Р', '2024-20-ИС1', 'Искусственные сооружения', isRD)])

// ------------------------------------------------------------- pack-scan --
// ПД as text, РД as a scan: the same sheets rendered to pictures and printed
// back to a PDF without a text layer. Recognition (tesseract) must read them.
import { execFileSync as run } from 'node:child_process'
import { readdirSync, readFileSync, writeFileSync as write } from 'node:fs'

function scanned (file, bodies) {
  const dir = mkdtempSync(join(tmpdir(), 'cmp-scan-'))
  try {
    const text = join(dir, 'text.pdf')
    pdf(text, bodies)
    run('pdftoppm', ['-r', '150', '-png', text, join(dir, 'p')])
    const imgs = readdirSync(dir).filter(f => f.endsWith('.png')).sort()
    const pages = imgs.map(f => `<div class="page" style="padding:0"><img src="data:image/png;base64,${readFileSync(join(dir, f)).toString('base64')}" style="width:297mm;height:210mm;display:block"></div>`)
    pdf(file, pages)
  } finally {
    rmSync(dir, { recursive: true, force: true })
  }
}
pdf(join(HERE, 'pack-scan/pd.pdf'), [volume('П', '2024-15-КЖ', 'Конструктивные решения', kzhPD)])
scanned(join(HERE, 'pack-scan/rd.pdf'), [volume('Р', '2024-15-КЖ1', 'Конструктивные решения', kzhRD)])

// -------------------------------------------------------------- pack-vor --
// ВОР and specification tables. ПД is consistent (the bill's reinforcement is
// the sum of two specification lines); РД changes the slab concrete (+12%),
// changes the waterproofing in the specification only (so the bill and the
// specification disagree) and adds a drain pipe.
function tablesVolume (stage, code, vorRows, specRows) {
  const head = ['№ п/п', 'Наименование работ', 'Ед. изм.', 'Кол-во', 'Примечание']
  const shead = ['Поз.', 'Наименование', 'Ед. изм.', 'Кол.', 'Примечание']
  return [
    sheet(stage, code, 1, 2, 'Ведомость объемов работ', [head, ...vorRows]),
    sheet(stage, code, 2, 2, 'Спецификация оборудования, изделий и материалов', [shead, ...specRows]),
  ].join('\n')
}
pdf(join(HERE, 'pack-vor/pd.pdf'), [tablesVolume('П', '2024-30-КЖ', [
  ['1', 'Устройство бетонной подготовки из бетона B7,5', 'м3', '6,0', ''],
  ['2', 'Бетонирование фундаментной плиты из бетона B25', 'м3', '12,5', ''],
  ['3', 'Армирование фундаментной плиты арматурой A500C', 'т', '3,2', ''],
  ['4', 'Гидроизоляция обмазочная стен подвала', 'м2', '40', ''],
], [
  ['1', 'Бетон B7,5 для бетонной подготовки', 'м3', '6,0', ''],
  ['2', 'Бетон B25 для фундаментной плиты', 'м3', '12,5', ''],
  ['3', 'Арматура A500C d16 для фундаментной плиты', 'т', '2,0', ''],
  ['4', 'Арматура A500C d12 для фундаментной плиты', 'т', '1,2', ''],
  ['5', 'Гидроизоляция обмазочная стен подвала', 'м2', '40', ''],
])])
pdf(join(HERE, 'pack-vor/rd.pdf'), [tablesVolume('Р', '2024-30-КЖ1', [
  ['1', 'Устройство бетонной подготовки из бетона B7,5', 'м3', '6,0', ''],
  ['2', 'Бетонирование фундаментной плиты из бетона B25', 'м3', '14,0', ''],
  ['3', 'Армирование фундаментной плиты арматурой A500C', 'т', '3,2', ''],
  ['4', 'Гидроизоляция обмазочная стен подвала', 'м2', '40', ''],
  ['5', 'Устройство дренажной трубы ПНД вдоль фундамента', 'м', '25', ''],
], [
  ['1', 'Бетон B7,5 для бетонной подготовки', 'м3', '6,0', ''],
  ['2', 'Бетон B25 для фундаментной плиты', 'м3', '14,0', ''],
  ['3', 'Арматура A500C d16 для фундаментной плиты', 'т', '2,0', ''],
  ['4', 'Арматура A500C d12 для фундаментной плиты', 'т', '1,2', ''],
  ['5', 'Гидроизоляция обмазочная стен подвала', 'м2', '45', ''],
  ['6', 'Труба дренажная ПНД вдоль фундамента', 'м', '25', ''],
])])
