import { describe, expect, it } from 'vitest'
import JSZip from 'jszip'
import { borderToCss, colWidthToPx, decodeCell, ptToPx, readLayout } from './layout.js'

const STYLES = `<?xml version="1.0" encoding="UTF-8"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <borders count="3">
    <border><left/><right/><top/><bottom/><diagonal/></border>
    <border>
      <left style="thin"><color rgb="FFFF0000"/></left>
      <right style="medium"><color indexed="64"/></right>
      <top/>
      <bottom style="double"><color rgb="FF00FF00"/></bottom>
    </border>
    <border><left/><right/><top style="dashed"/><bottom/></border>
  </borders>
  <cellStyleXfs count="1"><xf numFmtId="0" borderId="2"/></cellStyleXfs>
  <cellXfs count="3">
    <xf numFmtId="0" borderId="0"/>
    <xf numFmtId="0" borderId="1" applyBorder="1"><alignment wrapText="1"/></xf>
    <xf numFmtId="0" borderId="2" applyBorder="1"/>
  </cellXfs>
</styleSheet>`

const WORKBOOK = `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Data &amp; more" sheetId="1" r:id="rId1"/>
    <sheet name="Second" sheetId="2" r:id="rId2"/>
  </sheets>
</workbook>`

const RELS = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="worksheet" Target="/xl/worksheets/sheet2.xml"/>
  <Relationship Id="rId3" Type="styles" Target="styles.xml"/>
</Relationships>`

const SHEET1 = `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetFormatPr defaultRowHeight="20.25" defaultColWidth="10"/>
  <cols><col min="1" max="1" width="20" customWidth="1"/></cols>
  <sheetData>
    <row r="1"><c r="A1" s="1" t="s"><v>0</v></c><c r="B1" s="0"/></row>
    <row r="2" ht="30" customHeight="1"><c r="C2" s="2"/><c r="D2"><v>1</v></c></row>
  </sheetData>
</worksheet>`

const SHEET2 = `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData><row r="1"><c r="AA10" s="1"/></row></sheetData>
</worksheet>`

async function buildXlsx () {
  const zip = new JSZip()
  zip.file('xl/workbook.xml', WORKBOOK)
  zip.file('xl/_rels/workbook.xml.rels', RELS)
  zip.file('xl/styles.xml', STYLES)
  zip.file('xl/worksheets/sheet1.xml', SHEET1)
  zip.file('xl/worksheets/sheet2.xml', SHEET2)
  return zip.generateAsync({ type: 'uint8array' })
}

describe('xlsx layout', () => {
  it('converts units', () => {
    expect(ptToPx(15)).toBe(20)
    expect(ptToPx(30)).toBe(40)
    expect(colWidthToPx({ wpx: 185 })).toBe(185)
    expect(colWidthToPx({ wch: 10 })).toBe(75)
    expect(colWidthToPx(undefined)).toBeUndefined()
  })

  it('maps border edges to CSS', () => {
    expect(borderToCss({ style: 'thin', color: { rgb: 'FF112233' } })).toBe('1px solid #112233')
    expect(borderToCss({ style: 'double' })).toBe('3px double #000')
    expect(borderToCss({ style: 'hair', color: { theme: '1' } })).toBe('1px dotted #000')
    expect(borderToCss({ style: 'none' })).toBeUndefined()
    expect(borderToCss({})).toBeUndefined()
  })

  it('decodes cell references', () => {
    expect(decodeCell('A1')).toEqual({ r: 0, c: 0 })
    expect(decodeCell('AA10')).toEqual({ r: 9, c: 26 })
    expect(decodeCell('$B$3')).toEqual({ r: 2, c: 1 })
    expect(decodeCell('')).toBeNull()
  })

  it('reads borders and defaults per sheet', async () => {
    const layout = await readLayout(await buildXlsx())

    const first = layout['Data & more']
    expect(first.defaultRowHeightPt).toBe(20.25)
    expect(first.defaultColWidthPx).toBe(75)
    expect(first.borders.get('0:0')).toEqual({
      borderLeft: '1px solid #FF0000',
      borderRight: '2px solid #000',
      borderBottom: '3px double #00FF00',
    })
    expect(first.borders.get('1:2')).toEqual({ borderTop: '1px dashed #000' })
    // s="0" → border without edges, and cells without s → nothing
    expect(first.borders.has('0:1')).toBe(false)
    expect(first.borders.has('1:3')).toBe(false)

    const second = layout.Second
    expect(second.defaultRowHeightPt).toBe(15)
    expect(second.borders.get('9:26')).toMatchObject({ borderLeft: '1px solid #FF0000' })
  })

  it('returns an empty layout for non-xlsx input', async () => {
    expect(await readLayout(new Uint8Array([1, 2, 3]))).toEqual({})
  })
})
