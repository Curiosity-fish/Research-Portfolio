import fs from "node:fs/promises";
import { FileBlob, SpreadsheetFile } from "@oai/artifact-tool";

const dir = "E:/Academic Navigation-test/Academic Navigation - 副本/outputs/019ff4f7-71d7-7e53-b1d2-4e8997fefabf";
const input = await FileBlob.load(`${dir}/登录账号密码.xlsx`);
const wb = await SpreadsheetFile.importXlsx(input);

const list = wb.worksheets.getItem("账号列表");
const info = wb.worksheets.getItem("说明");

console.log("=== 账号列表 A1:I6 ===");
const top = await wb.inspect({
  kind: "table",
  sheetId: "账号列表",
  range: "A1:I6",
  include: "values",
  tableMaxRows: 6,
  tableMaxCols: 9,
});
console.log(top.ndjson.slice(0, 2200));

console.log("=== 说明 B4:B8 ===");
const summary = await wb.inspect({
  kind: "table",
  sheetId: "说明",
  range: "A4:B8",
  include: "values,formulas",
  tableMaxRows: 8,
  tableMaxCols: 2,
});
console.log(summary.ndjson.slice(0, 1800));

console.log("=== 公式错误扫描 ===");
const errors = await wb.inspect({
  kind: "match",
  searchTerm: "#REF!|#DIV/0!|#VALUE!|#NAME\\?|#N/A",
  options: { useRegex: true, maxResults: 100 },
});
console.log(errors.ndjson.slice(0, 500));

console.log("=== 末行抽查 ===");
const lastRow = await wb.inspect({
  kind: "table",
  sheetId: "账号列表",
  range: `A${1025}:I${1027}`,
  include: "values",
  tableMaxRows: 3,
  tableMaxCols: 9,
});
console.log(lastRow.ndjson.slice(0, 1200));
