// A page picture that cannot be shown must say why, in words: the server answers an image request
// with a JSON error (writeErr), the browser only sees a broken <img>. probeImage asks for the
// same URL again and turns the answer into a readable reason.
import { reactive, watch } from 'vue'
import { names } from '../constants.js'

export function friendlyReason (status, message) {
  const m = String(message || '')
  if (/not a PDF|нет изображений страниц/i.test(m)) return 'Для этого типа файла нет картинок страниц — смотрите текст или скачайте файл.'
  if (/has no document|no such fixture|не найден среди документов|нет такого документа/i.test(m)) return 'Файл не найден среди документов этого сравнения: возможно, его удалили или заменили. Обновите страницу.'
  if (/out of range|нет такой страницы|не найден|not found/i.test(m) || status === 404) return 'В файле нет такой страницы: возможно, документ короче, чем ожидалось.'
  if (/token is expired|unauthorized/i.test(m) || status === 401) return 'Агент потерял доступ к Compose (токен истёк). Обновите страницу или перезапустите агент.'
  if (/rasteriz|pdftoppm/i.test(m)) return 'Не удалось подготовить картинку страницы (ошибка растеризации PDF). Файл может быть повреждён.'
  if (status >= 500) return 'Сервер не смог подготовить картинку страницы' + (m ? ` (${m.slice(0, 120)}).` : '.')
  if (status === 0) return 'Нет связи с агентом: картинка не загрузилась. Проверьте, что агент запущен.'
  return m ? `Картинка страницы не загрузилась: ${m.slice(0, 140)}` : 'Картинка страницы не загрузилась.'
}

export async function probeImage (url) {
  try {
    const res = await fetch(url, { cache: 'no-store' })
    if (res.ok) return 'Картинка получена, но браузер не смог её показать (повреждённое изображение).'
    let msg = ''
    try { msg = (await res.json()).error || '' } catch (e) { msg = (await res.text().catch(() => '')) || '' }
    return friendlyReason(res.status, msg)
  } catch (e) {
    return friendlyReason(0, '')
  }
}

// Failure state of a pair of page pictures (side by side, overlay, difference): which one failed and why.
export function usePageErrors (props) {
  const st = reactive({ pd: '', rd: '', tick: 0 })
  watch(() => props.pdUrl, () => { st.pd = '' })
  watch(() => props.rdUrl, () => { st.rd = '' })
  const fail = async (side, url) => { st[side] = await probeImage(url) }
  const retry = () => { st.pd = ''; st.rd = ''; st.tick++ }
  return { st, fail, retry }
}

const NO_PAGE = 'В этом файле нет такой страницы (номер больше, чем страниц в документе) или файл не выбран.'

// The messages that replace the overlay / difference stage: a side with no page or a picture that failed.
export function pageProblems (props, st) {
  const out = []
  if (!props.pdUrl) out.push({ label: names.ref, reason: NO_PAGE })
  else if (st.pd) out.push({ label: names.ref, reason: st.pd })
  if (!props.rdUrl) out.push({ label: names.target, reason: NO_PAGE })
  else if (st.rd) out.push({ label: names.target, reason: st.rd })
  return out.length ? out : null
}
