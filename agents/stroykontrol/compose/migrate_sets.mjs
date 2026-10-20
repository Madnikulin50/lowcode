#!/usr/bin/env node
/**
 * Раскладывает существующие документы сравнений по комплектам (document_sets).
 *
 * Для каждого рабочего (не демо) сравнения с объектом: документы одной стороны (ПД/РД)
 * собираются в комплект с версией = метка doc_version ("Изм. 2", "Исходная (2023)"; несколько
 * разных меток склеиваются через « + »; нет метки — «v1»). Одинаковые (объект, сторона, версия)
 * образуют один комплект. Документам проставляется set, сравнению — set_pd/set_rd.
 * Файлы и записи не удаляются; повторный запуск ничего не дублирует.
 *
 *   COMPOSE_DSN=… node migrate_sets.mjs            # показать, что будет сделано
 *   COMPOSE_DSN=… node migrate_sets.mjs --apply    # выполнить
 */
import { mintToken, detectBase, apiFactory, setOf } from './helpers.mjs'

const APPLY = process.argv.includes('--apply')
const STAGE_OF = { ref: 'ref', target: 'target' }

const val = (rec, name) => (rec.values || []).find(v => v.name === name)?.value || ''

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const ns = setOf(await api('GET', '/namespace/?limit=200')).find(n => n.slug === 'stroykontrol' && !n.deletedAt)?.namespaceID
  if (!ns) throw new Error('namespace stroykontrol not found')
  const mod = {}
  for (const h of ['pd_rd_comparisons', 'cmp_documents', 'document_sets']) {
    mod[h] = setOf(await api('GET', `/namespace/${ns}/module/?handle=${h}&limit=5`)).find(x => x.handle === h && !x.deletedAt)?.moduleID
    if (!mod[h]) throw new Error(`module ${h} not found — run apply.mjs first`)
  }
  const list = async (h, query) => {
    const out = []
    for (let page = 1; ; page++) {
      const res = await api('GET', `/namespace/${ns}/module/${mod[h]}/record/?limit=200&page=${page}${query ? '&query=' + encodeURIComponent(query) : ''}`)
      const set = setOf(res)
      out.push(...set)
      if (set.length < 200) break
    }
    return out
  }
  const save = async (h, rec, patch) => {
    const values = (rec.values || []).filter(v => !(v.name in patch))
    for (const [name, value] of Object.entries(patch)) values.push({ name, value: String(value) })
    return api('POST', `/namespace/${ns}/module/${mod[h]}/record/${rec.recordID}`, { values })
  }

  const comps = (await list('pd_rd_comparisons', "(data_kind != 'demo' OR data_kind IS NULL)")).filter(c => val(c, 'object'))
  const docs = await list('cmp_documents')
  const sets = await list('document_sets')
  const setKey = s => [val(s, 'object'), val(s, 'stage'), val(s, 'version')].join('|')
  const setByKey = new Map(sets.map(s => [setKey(s), s]))
  console.log(`сравнений с объектом: ${comps.length}, документов: ${docs.length}, комплектов уже есть: ${sets.length}`)

  let created = 0; let linkedDocs = 0; let linkedCmps = 0
  for (const c of comps) {
    const mine = docs.filter(d => val(d, 'comparison') === c.recordID)
    const patch = {}
    for (const side of ['ref', 'target']) {
      const sideDocs = mine.filter(d => (STAGE_OF[val(d, 'side')] || 'target') === side)
      if (!sideDocs.length) continue
      const labels = [...new Set(sideDocs.map(d => val(d, 'doc_version').trim()).filter(Boolean))]
      const version = labels.length ? labels.join(' + ') : 'v1'
      const key = [val(c, 'object'), side, version].join('|')
      let set = setByKey.get(key)
      if (!set) {
        console.log(`  + комплект ${side === 'ref' ? 'ПД' : 'РД'} «${version}» объекта ${val(c, 'object')} (из сравнения «${val(c, 'title')}»)`)
        created++
        if (APPLY) {
          set = await api('POST', `/namespace/${ns}/module/${mod.document_sets}/record/`, {
            values: [
              { name: 'object', value: val(c, 'object') }, { name: 'stage', value: side },
              { name: 'version', value: version }, { name: 'status', value: 'accepted' },
            ],
          })
        } else {
          set = { recordID: `(новый ${key})`, values: [] }
        }
        setByKey.set(key, set)
      }
      for (const d of sideDocs) {
        if (val(d, 'set') === set.recordID) continue
        linkedDocs++
        if (APPLY) await save('cmp_documents', d, { set: set.recordID })
      }
      patch[side === 'ref' ? 'set_pd' : 'set_rd'] = set.recordID
    }
    const todo = Object.entries(patch).filter(([k, v]) => val(c, k) !== v)
    if (todo.length) {
      linkedCmps++
      console.log(`  ~ «${val(c, 'title')}»: ${todo.map(([k, v]) => `${k}=${v}`).join(', ')}`)
      if (APPLY) await save('pd_rd_comparisons', c, Object.fromEntries(todo))
    }
  }
  console.log(`${APPLY ? 'Готово' : 'Будет сделано'}: комплектов создано ${created}, документов привязано ${linkedDocs}, сравнений обновлено ${linkedCmps}.`)
  if (!APPLY) console.log('Это пробный прогон; для выполнения добавьте --apply.')
}

main().catch(e => { console.error(e); process.exit(1) })
