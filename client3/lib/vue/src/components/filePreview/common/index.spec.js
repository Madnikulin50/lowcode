import { describe, expect, it } from 'vitest'
import { getComponent, getFileExt, hintKind, canPreview } from './index.js'

describe('filePreview registry', () => {
  it('picks office and CAD/BIM components by extension', () => {
    expect(getComponent({ name: 'a.pdf' })).toBe('PDF')
    expect(getComponent({ name: 'a.docx' })).toBe('DOCX')
    expect(getComponent({ name: 'a.xlsx' })).toBe('XLSX')
    expect(getComponent({ name: 'a.pptx' })).toBe('PPTX')
    expect(getComponent({ name: 'a.dxf' })).toBe('CAD')
    expect(getComponent({ name: 'a.dwg' })).toBe('Hint')
    expect(getComponent({ name: 'a.ifc' })).toBe('BIM')
    expect(getComponent({ name: 'a.frag' })).toBe('BIM')
    expect(getComponent({ name: 'a.pln' })).toBe('Hint')
    expect(getComponent({ name: 'a.pla' })).toBe('Hint')
    expect(getComponent({ name: 'a.bimx' })).toBe('Hint')
  })

  it('does not treat DXF/DWG as images despite image/vnd MIME', () => {
    expect(getComponent({ type: 'image/vnd.dxf', name: 'x.dxf' })).toBe('CAD')
    expect(getComponent({ type: 'image/vnd.dwg', name: 'x.dwg' })).toBe('Hint')
    expect(getComponent({ type: 'image/png', name: 'x.png' })).toBe('IMG')
  })

  it('prefers meta.original.ext over a misleading URL', () => {
    expect(getComponent({
      src: '/attachment/123/original/file?sign=1',
      meta: { original: { ext: 'ifc' } },
    })).toBe('BIM')
    expect(getFileExt({ meta: { original: { ext: '.DXF' } } })).toBe('dxf')
  })

  it('classifies hint kinds', () => {
    expect(hintKind('dwg')).toBe('dwg')
    expect(hintKind('pln')).toBe('archicad')
    expect(hintKind('pla')).toBe('archicad')
    expect(hintKind('bimx')).toBe('bimx')
  })

  it('canPreview is true for supported and fallback types', () => {
    expect(canPreview({ name: 'a.zip' })).toBe(true)
    expect(canPreview({ name: 'a.docx' })).toBe(true)
    expect(canPreview({})).toBe(false)
  })
})
