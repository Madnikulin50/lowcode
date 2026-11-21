import { computed, reactive, unref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import MarkdownIt from 'markdown-it'

const helpMarkdown = new MarkdownIt({ html: false, linkify: true, breaks: true })

function markdownToHtml (markdown) {
  if (!markdown) return ''
  return helpMarkdown.render(markdown)
}

function topicKey (topic, part) {
  return `help.topics.${topic}.${part}`
}

function translate (t, _te, key) {
  if (!key) return ''
  const value = t(key)
  return !value || value === key ? '' : value
}

/**
 * Resolves product help (i18n topic) and optional app-authored docs.
 * @param {string|import('vue').MaybeRef<string>} [topic]
 * @param {object|import('vue').MaybeRef<object>} [appDocs]
 */
export function useHelp (topic, appDocs = {}, options = {}) {
  const { t, te } = useI18n({ useScope: 'global' })
  const route = useRoute()
  const includeProduct = computed(() => unref(options.includeProduct) !== false)

  const labels = computed(() => ({
    title: translate(t, te, 'help.panel.title') || 'Help',
    app: translate(t, te, 'help.panel.app') || 'About this',
    product: translate(t, te, 'help.panel.product') || 'How to use',
    close: translate(t, te, 'help.panel.close') || 'Close',
  }))

  const resolvedTopic = computed(() => unref(topic) || route.meta?.helpTopic || '')

  const product = computed(() => {
    const id = resolvedTopic.value
    if (!id) return { title: '', hint: '', body: '', html: '' }
    const title = translate(t, te, topicKey(id, 'title'))
    const hint = translate(t, te, topicKey(id, 'hint'))
    const body = translate(t, te, topicKey(id, 'body'))
    return {
      title,
      hint,
      body,
      html: body ? markdownToHtml(body) : '',
    }
  })

  const app = computed(() => {
    const docs = unref(appDocs) || {}
    const description = docs.description || docs.hint || ''
    const body = String(docs.help || docs.body || '').replace(/<!--\s*compose-help:v\d+\s*-->\s*/g, '')
    return {
      title: docs.title || '',
      hint: docs.hint || '',
      description,
      body,
      html: body ? markdownToHtml(body) : '',
    }
  })

  const hasApp = computed(() => !!(app.value.description || app.value.html))
  const hasProduct = computed(() => includeProduct.value && !!(product.value.hint || product.value.html))
  const hasAny = computed(() => hasApp.value || hasProduct.value || !!app.value.hint || !!app.value.html)
  const hint = computed(() => app.value.hint || app.value.description || (includeProduct.value ? product.value.hint : ''))

  const panelProps = computed(() => ({
    title: app.value.title || (includeProduct.value ? product.value.title : ''),
    description: app.value.description,
    bodyHtml: app.value.html,
    productHint: includeProduct.value ? product.value.hint : '',
    productHtml: includeProduct.value ? product.value.html : '',
    labels: labels.value,
  }))

  const triggerProps = computed(() => ({
    ...panelProps.value,
    hint: hint.value,
  }))

  // reactive() unwraps nested computeds so v-bind="help.triggerProps" in
  // templates spreads the props object, not the ComputedRef internals.
  return reactive({
    labels,
    product,
    app,
    hasApp,
    hasProduct,
    hasAny,
    hint,
    panelProps,
    triggerProps,
  })
}
