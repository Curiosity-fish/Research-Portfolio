import { FileBlob, SpreadsheetFile } from '@oai/artifact-tool'

const filePath = process.argv[2]
if (!filePath) {
  console.error('usage: node verify_workbook.mjs <path-to.xlsx>')
  process.exit(1)
}

const input = await FileBlob.load(filePath)
const workbook = await SpreadsheetFile.importXlsx(input)

const sheetInfo = await workbook.inspect({
  kind: 'sheet',
  include: 'id,name',
  maxChars: 4000,
})
console.log('--- sheets ---')
console.log(sheetInfo.ndjson)

console.log('--- table sizes ---')
for (const sheet of workbook.worksheets.items) {
  const used = sheet.getUsedRange(true)
  if (!used) continue
  const values = used.values
  const rows = values.length
  const cols = values[0]?.length ?? 0
  const header = values[0]?.join(' | ') ?? ''
  console.log(`${sheet.name}: ${rows} rows x ${cols} cols :: ${header.slice(0, 120)}`)
}

console.log('--- formula error scan ---')
const errors = await workbook.inspect({
  kind: 'match',
  searchTerm: '#REF!|#DIV/0!|#VALUE!|#NAME\\?|#N/A',
  options: { useRegex: true, maxResults: 50 },
  summary: 'final formula error scan',
})
console.log(errors.ndjson || 'no formula errors')
