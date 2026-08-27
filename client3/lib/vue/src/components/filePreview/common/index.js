import mime from 'mime'

/**
 * More specific MIME types must come before prefix matches like `image/`.
 * `image/vnd.dxf` / `image/vnd.dwg` would otherwise be treated as bitmaps.
 */
const mimeTypes = [
  { type: 'application/pdf', component: 'PDF' },
  { type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', component: 'DOCX' },
  { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', component: 'XLSX' },
  { type: 'application/vnd.openxmlformats-officedocument.presentationml.presentation', component: 'PPTX' },
  { type: 'image/vnd.dxf', component: 'CAD' },
  { type: 'application/dxf', component: 'CAD' },
  { type: 'application/x-dxf', component: 'CAD' },
  { type: 'image/vnd.dwg', component: 'Hint' },
  { type: 'application/acad', component: 'Hint' },
  { type: 'application/x-dwg', component: 'Hint' },
  { type: 'application/dwg', component: 'Hint' },
  { type: 'application/ifc', component: 'BIM' },
  { type: 'application/x-ifc', component: 'BIM' },
  { type: 'model/ifc', component: 'BIM' },
  { type: 'application/x-step', component: 'BIM' },
  { type: 'application/ifczip', component: 'BIM' },
  { type: 'image/', component: 'IMG' },
]

const extComponents = {
  pdf: 'PDF',
  png: 'IMG',
  jpg: 'IMG',
  jpeg: 'IMG',
  gif: 'IMG',
  webp: 'IMG',
  bmp: 'IMG',
  svg: 'IMG',
  docx: 'DOCX',
  xlsx: 'XLSX',
  pptx: 'PPTX',
  dxf: 'CAD',
  dwg: 'Hint',
  ifc: 'BIM',
  ifczip: 'BIM',
  frag: 'BIM',
  pln: 'Hint',
  pla: 'Hint',
  bimx: 'Hint',
}

export function getFileExt ({ src, name, meta } = {}) {
  const fromMeta = meta?.original?.ext || meta?.preview?.ext
  if (fromMeta) {
    return String(fromMeta).replace(/^\./, '').toLowerCase()
  }
  const base = String(name || src || '').split('?')[0]
  const m = base.match(/\.([a-z0-9]+)$/i)
  return m ? m[1].toLowerCase() : ''
}

function mimeComponent (srcType) {
  if (!srcType) {
    return
  }
  const lower = String(srcType).toLowerCase()
  for (const { type, component } of mimeTypes) {
    if (lower.indexOf(type) >= 0) {
      return component
    }
  }
}

/**
 * Tells what component (if any) can preview the given file
 * @returns {String|undefined} preview component or undefined
 */
export const getComponent = ({ type, src, name, meta } = {}) => {
  const ext = getFileExt({ src, name, meta })
  if (ext && extComponents[ext]) {
    return extComponents[ext]
  }

  const srcType = type || mime.getType(src) || mime.getType(name)
  const fromMime = mimeComponent(srcType)
  if (fromMime) {
    return fromMime
  }

  if (srcType || ext || name || src) {
    return 'NoPreview'
  }
}

/**
 * Tells if we support the given file type preview
 */
export const canPreview = ({ type, src, name, meta } = {}) => {
  return !!getComponent({ type, src, name, meta })
}

export const getExtensionIconType = (ext) => {
  switch (String(ext || '').replace(/^\./, '').toLowerCase()) {
    case 'odt':
    case 'doc':
    case 'docx':
      return 'word'
    case 'pdf':
      return 'pdf'
    case 'ppt':
    case 'pptx':
      return 'powerpoint'
    case 'zip':
    case 'rar':
      return 'archive'
    case 'xls':
    case 'xlsx':
    case 'csv':
      return 'excel'
    case 'mov':
    case 'mp3':
    case 'mp4':
      return 'video'
    case 'png':
    case 'jpg':
    case 'jpeg':
    case 'gif':
    case 'webp':
    case 'bmp':
    case 'svg':
      return 'image'
    case 'dxf':
    case 'dwg':
    case 'ifc':
    case 'ifczip':
    case 'frag':
    case 'pln':
    case 'pla':
    case 'bimx':
      return 'alt'
    default: return 'alt'
  }
}

export const hintKind = (ext) => {
  switch (String(ext || '').replace(/^\./, '').toLowerCase()) {
    case 'dwg':
      return 'dwg'
    case 'pln':
    case 'pla':
      return 'archicad'
    case 'bimx':
      return 'bimx'
    default:
      return 'generic'
  }
}
