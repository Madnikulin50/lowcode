/**
 * Apply the user's color mode to <html>.
 *
 * Corteza themes are selected by `data-color-mode`. Bootstrap 5 dark-mode
 * tokens (`--bs-body-bg`, form/modal/dropdown surfaces, …) are selected by
 * `data-bs-theme`. Both must be set together, otherwise Bootstrap paints
 * its own cool-gray dark over the Corteza palette.
 *
 * Keep ALMOST_BLACK in sync with client3/lib/vue/src/scss/dark.scss and
 * server/pkg/provision/stylesheet.go.
 */

const ALMOST_BLACK: Record<string, string> = {
  black: '#EDEDED',
  white: '#121212',
  primary: '#6E8FF0',
  secondary: '#9A9A9A',
  success: '#43AA8B',
  warning: '#E27646',
  danger: '#F2555A',
  light: '#1A1A1A',
  'extra-light': '#242424',
  dark: '#EDEDED',
  'body-bg': '#0A0A0A',
  'sidebar-bg': '#121212',
  'topbar-bg': '#121212',
}

const LEGACY_NAVY_BODY_BG = '#092B40'
const LEGACY_NAVY_WHITE = '#0B344E'

function hexEq (a: string, b: string): boolean {
  return a.trim().toLowerCase() === b.trim().toLowerCase()
}

function clearInlinePalette (html: HTMLElement): void {
  for (const name of Object.keys(ALMOST_BLACK)) {
    html.style.removeProperty(`--${name}`)
  }
}

/**
 * The old Corteza dark theme (navy #092B40 / #0B344E) used to be the
 * branding-editor default and was persisted into ui.studio.themes →
 * custom.css. That stylesheet beats the client fallback, so the blue
 * background kept returning. If the computed palette is still exactly
 * that old default, stamp the almost-black variables onto <html> so
 * they win regardless of custom.css / minified-custom.css.
 */
function neutralizeLegacyNavy (html: HTMLElement): void {
  const bodyBg = getComputedStyle(html).getPropertyValue('--body-bg')
  const white = getComputedStyle(html).getPropertyValue('--white')
  if (!hexEq(bodyBg, LEGACY_NAVY_BODY_BG) || !hexEq(white, LEGACY_NAVY_WHITE)) {
    return
  }

  for (const [name, value] of Object.entries(ALMOST_BLACK)) {
    html.style.setProperty(`--${name}`, value)
  }
}

export function applyColorMode (theme?: string | null): void {
  if (typeof document === 'undefined' || !theme) {
    return
  }

  const html = document.documentElement
  html.setAttribute('data-color-mode', theme)
  html.setAttribute('data-bs-theme', theme)

  if (theme === 'dark') {
    neutralizeLegacyNavy(html)
  } else {
    clearInlinePalette(html)
  }
}
