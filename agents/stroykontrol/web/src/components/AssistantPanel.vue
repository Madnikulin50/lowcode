<script setup>
// Ассистент по документу на экране: специалист спрашивает о листе (и о выбранной находке), модель
// отвечает только по тексту листа и приложенной картинке, подсказывает листы, к которым можно перейти.
// Один компонент для просмотрщика сравнения, проверки ИД и отдельного просмотра документа.
import { ref, reactive, computed, nextTick, watch } from 'vue'
import { names, nm } from '../constants.js'
import { errText } from '../utils.js'

const props = defineProps({
  namespaceId: String,
  // () => context of the request (or null when no document is open); receives the chosen scope
  getContext: Function,
  chips: { type: Array, default: () => [] }, // what is in the context: «ПД: файл, стр. 11»
  scopes: { type: Array, default: () => [] }, // [{ key, label }] — what the question is about
  defaultScope: { type: String, default: '' },
  quick: { type: Array, default: null },
  title: { type: String, default: 'Ассистент по документу' },
  hint: { type: String, default: '' },
  activeLabel: { type: String, default: '' }, // the selected finding, if any
  activeId: { type: String, default: '' },
  startOpen: { type: Boolean, default: false },
  canImage: { type: Boolean, default: true },
})
const emit = defineEmits(['goto', 'to-remark'])

const open = ref(props.startOpen)
const state = reactive({ busy: false, error: '', input: '', withImage: false, elapsed: 0 })
const scope = ref(props.defaultScope || (props.scopes[0] && props.scopes[0].key) || '')
const messages = ref([]) // { role, content, jumps?, meta? }
const box = ref(null)

const QUICK = computed(() => props.quick || [
  { label: 'Что на этом листе?', text: 'Что показано на открытом листе? Назови тип листа, основные данные основной надписи и главное содержание.' },
  { label: 'Размеры и отметки', text: 'Извлеки с открытого листа размеры, отметки уровней, марки материалов и другие числовые данные. Если читаешь чертёж — следуй методике чтения чертежей.' },
  { label: 'Объясни расхождение', text: 'Объясни простыми словами, в чём различие в выбранном расхождении и чем оно может быть вызвано. Не оценивай проектное решение.' },
  { label: 'Сформулируй замечание', text: 'Сформулируй текст замечания по выбранному расхождению: где, в чём расхождение и что требуется.' },
])

async function send (text) {
  const content = (text || state.input).trim()
  if (!content || state.busy) return
  const ctx = props.getContext ? props.getContext(scope.value) : null
  if (!ctx) { state.error = 'Сначала откройте документ.'; return }
  messages.value.push({ role: 'user', content })
  state.input = ''
  state.busy = true
  state.elapsed = 0
  state.error = ''
  try {
    // the model may think for minutes, longer than the gateway waits for one request: the question is
    // started as a job and the answer is collected by polling
    const res = await fetch('api/assist', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        namespaceID: props.namespaceId, async: true,
        context: { ...ctx, withImage: state.withImage },
        messages: messages.value.map(m => ({ role: m.role, content: m.content })),
      }),
    })
    let data = await res.json().catch(() => ({}))
    if (!res.ok || data.error) throw new Error(errText(data.error, res.status) || res.statusText)
    if (data.jobID) {
      const started = Date.now()
      for (;;) {
        await new Promise(r => setTimeout(r, 2000))
        state.elapsed = Math.round((Date.now() - started) / 1000)
        if (state.elapsed > 14 * 60) throw new Error('Модель не ответила за 14 минут.')
        const pr = await fetch(`api/assist/result?job=${encodeURIComponent(data.jobID)}`)
        const pd = await pr.json().catch(() => ({}))
        if (!pr.ok) throw new Error(errText(pd.error, pr.status))
        if (pd.status === 'running') continue
        if (pd.status === 'error') throw new Error(errText(pd.error))
        data = pd
        break
      }
    }
    messages.value.push({ role: 'assistant', content: data.reply, jumps: data.jumps || [], meta: { image: data.image, skill: data.skill } })
  } catch (e) {
    state.error = e.message
  } finally {
    state.busy = false
    await nextTick()
    if (box.value) box.value.scrollTop = box.value.scrollHeight
  }
}
const clear = () => { messages.value = []; state.error = '' }
watch(() => props.startOpen, v => { if (v) open.value = true })
defineExpose({ openAndFocus () { open.value = true } })
</script>

<template>
  <div class="assistant">
    <h2 class="assistant-h" @click="open = !open">
      <span class="caret">{{ open ? '▾' : '▸' }}</span> {{ title }}
      <span class="muted">{{ hint || 'спросить модель об открытом листе · отвечает только по тексту и картинке листа' }}</span>
    </h2>
    <template v-if="open">
      <div v-if="scopes.length" class="as-scopes">
        <label v-for="s in scopes" :key="s.key"><input v-model="scope" type="radio" :value="s.key"> {{ s.label }}</label>
      </div>
      <div class="as-ctx">
        <span v-for="c in chips" :key="c" class="as-chip">{{ c }}</span>
        <span v-if="!chips.length" class="muted">Документ не открыт.</span>
        <span v-if="activeLabel" class="as-chip sel" :title="activeLabel">выбрана находка: {{ activeLabel }}</span>
      </div>
      <div ref="box" class="as-log">
        <div v-if="!messages.length" class="empty">Задайте вопрос по открытому листу или возьмите быстрый вариант ниже.</div>
        <div v-for="(m, i) in messages" :key="i" class="as-msg" :class="m.role">
          <div class="as-role">{{ m.role === 'user' ? 'Вы' : 'Модель' }}<template v-if="m.meta && m.meta.image"> · смотрела изображение листа</template><template v-if="m.meta && m.meta.skill"> · методика чтения чертежей</template></div>
          <div class="as-text">{{ m.content }}</div>
          <div v-if="m.role === 'assistant'" class="as-acts">
            <button v-for="(j, k) in m.jumps" :key="k" class="small" @click="emit('goto', j)">перейти: {{ j.name }}</button>
            <button v-if="activeId" class="small" title="Положить этот текст в карточку замечания по выбранному расхождению" @click="emit('to-remark', { id: activeId, text: m.content })">в карточку замечания</button>
          </div>
        </div>
        <div v-if="state.busy" class="muted">модель думает… {{ state.elapsed ? state.elapsed + ' с' : '' }}<template v-if="state.elapsed > 40"> — локальная модель на процессоре отвечает долго, окно можно не закрывать</template></div>
      </div>
      <div v-if="state.error" class="err">{{ state.error }}</div>
      <div class="as-quick">
        <button v-for="q in QUICK" :key="q.label" class="small" :disabled="state.busy" @click="send(nm(q.text))">{{ q.label }}</button>
        <span class="grow"></span>
        <label v-if="canImage" class="as-img" title="Передать модели изображение листа (для чертежей и сканов оно передаётся само, если у листа почти нет текста)"><input v-model="state.withImage" type="checkbox"> смотреть на изображение</label>
        <button v-if="messages.length" class="small" @click="clear">очистить</button>
      </div>
      <div class="as-input">
        <textarea v-model="state.input" rows="2" :placeholder="`Вопрос об открытом листе (${names.ref} — эталон)`" @keydown.ctrl.enter="send()"></textarea>
        <button class="small primary" :disabled="state.busy || !state.input.trim()" @click="send()">Отправить</button>
      </div>
      <div class="muted">Ctrl+Enter — отправить. Модель не правит документы и не оценивает проектные решения; чертежи без мультимодальной модели читаются слабо.</div>
    </template>
  </div>
</template>

<style scoped>
.assistant { margin-top: 16px; }
.assistant-h { font-size: 13px; margin: 0 0 6px; cursor: pointer; display: flex; flex-wrap: wrap; gap: 4px 12px; align-items: baseline; }
.muted { font-size: 11.5px; font-weight: 400; color: var(--muted); }
.caret { width: 12px; display: inline-block; }
.as-scopes { display: flex; flex-wrap: wrap; gap: 4px 14px; font-size: 12px; padding: 2px 0 4px; }
.as-ctx { display: flex; flex-wrap: wrap; gap: 4px; padding: 2px 0 6px; align-items: center; }
.as-chip { font-size: 11.5px; border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 1px 8px; background: var(--panel); }
.as-chip.sel { border-color: var(--accent); max-width: 420px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.as-log { max-height: 360px; overflow: auto; border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 6px 8px; background: var(--panel); }
.as-msg { padding: 4px 0; border-bottom: 1px solid var(--border); font-size: 12.5px; }
.as-msg:last-child { border-bottom: 0; }
.as-role { font-size: 11px; color: var(--muted); }
.as-msg.user .as-role { color: var(--accent); }
.as-text { white-space: pre-wrap; }
.as-acts { margin-top: 4px; display: flex; flex-wrap: wrap; gap: 4px; }
.as-quick { display: flex; flex-wrap: wrap; gap: 4px; margin: 6px 0; align-items: center; }
.as-quick .grow { flex: 1; }
.as-img { font-size: 11.5px; color: var(--muted); }
.as-input { display: flex; gap: 6px; align-items: flex-end; }
.as-input textarea { flex: 1; border: 1px solid var(--border); background: var(--bg); color: var(--fg); border-radius: var(--radius); padding: 4px 8px; font: inherit; font-size: 12.5px; resize: vertical; }
button.small { border: 1px solid var(--border); background: var(--panel); color: var(--fg); border-radius: var(--radius); padding: 2px 10px; font-size: 11.5px; cursor: pointer; }
button.small:disabled { opacity: .5; cursor: default; }
button.small.primary { background: var(--accent); border-color: var(--accent); color: #fff; }
</style>
