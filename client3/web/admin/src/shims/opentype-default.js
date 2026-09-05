export {
  BoundingBox,
  Font,
  Glyph,
  Path,
  _parse,
  load,
  loadSync,
  parse,
} from 'opentype.js/dist/opentype.mjs'

import { load, loadSync, parse } from 'opentype.js/dist/opentype.mjs'

const opentype = { parse, load, loadSync }

export default opentype
