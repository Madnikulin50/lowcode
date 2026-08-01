import TurndownService from 'turndown'
import MarkdownIt from 'markdown-it'

const turndownService = new TurndownService({
  headingStyle: 'atx',
  codeBlockStyle: 'fenced',
})

const md = new MarkdownIt({
  html: true,
  linkify: true,
  breaks: true,
})

export function htmlToMarkdown(html) {
  if (!html) return ''
  return turndownService.turndown(html)
}

export function markdownToHtml(markdown) {
  if (!markdown) return ''
  return md.render(markdown)
}
