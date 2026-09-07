import fs from "node:fs/promises";
import { SpreadsheetFile, Workbook } from "@oai/artifact-tool";

const dir = "E:/Academic Navigation-test/Academic Navigation - 副本/outputs/019ff4f7-71d7-7e53-b1d2-4e8997fefabf";
const raw = await fs.readFile(`${dir}/t_user.tsv`, "utf8");
const lines = raw.split(/\r?\n/).filter((l) => l.trim().length > 0);
const rows = lines.slice(1).map((l) => l.split("\t"));

const roleMap = {
  student: "学生",
  teacher: "班主任",
  course_teacher: "任课教师",
  department: "系主任",
  dean: "院长",
};
const roleOrder = { student: 0, teacher: 1, course_teacher: 2, department: 3, dean: 4 };
rows.sort((a, b) => {
  const ra = roleOrder[a[3]] ?? 9;
  const rb = roleOrder[b[3]] ?? 9;
  return ra - rb || String(a[1]).localeCompare(String(b[1]));
});

const data = rows.map((r, i) => [
  i + 1,
  r[1],
  "123456",
  r[2],
  roleMap[r[3]] || r[3],
  r[4] === "NULL" ? "" : r[4] || "",
  r[5] === "NULL" ? "" : r[5] || "",
  r[6] === "NULL" ? "" : r[6] || "",
  r[7] === "NULL" ? "" : r[7] || "",
]);
const n = data.length;

const wb = Workbook.create();
const info = wb.worksheets.add("说明");
const list = wb.worksheets.add("账号列表");

info.getRange("A1:C1").merge();
info.getRange("A1").values = [["Academic Navigation 登录账号"]];
info.getRange("A1").format = { font: { bold: true, size: 16, color: "#FFFFFF" }, fill: "#101d3e" };
info.getRange("A2:C2").merge();
info.getRange("A2").values = [["数据来源：academic_nav.t_user · 导出时间 2026-08-12"]];
info.getRange("A2").format = { font: { color: "#6b7280" } };

info.getRange("A4:B4").values = [["统计项", "数量"]];
info.getRange("A5:B8").values = [
  ["总账号数", null],
  ["学生", null],
  ["班主任/任课教师", null],
  ["系主任/院长", null],
];
info.getRange("B5").formulas = [[`=COUNTA('账号列表'!$A$2:$A$${n + 1})`]];
info.getRange("B6").formulas = [[`=COUNTIF('账号列表'!$E$2:$E$${n + 1},"学生")`]];
info.getRange("B7").formulas = [[`=COUNTIFS('账号列表'!$E$2:$E$${n + 1},"班主任")+COUNTIFS('账号列表'!$E$2:$E$${n + 1},"任课教师")`]];
info.getRange("B8").formulas = [[`=COUNTIFS('账号列表'!$E$2:$E$${n + 1},"系主任")+COUNTIFS('账号列表'!$E$2:$E$${n + 1},"院长")`]];
info.getRange("A4:B8").format = { borders: { preset: "all", style: "thin", color: "#D9D9D9" } };
info.getRange("A4:B4").format = { fill: "#2A4D99", font: { bold: true, color: "#FFFFFF" } };
info.getRange("A10:C11").merge();
info.getRange("A10").values = [["所有账号密码统一为 123456。角色对应：student=学生、teacher=班主任、course_teacher=任课教师、department=系主任、dean=院长。"]]; 
info.getRange("A10").format = { font: { color: "#374151" } };
info.showGridLines = false;

list.getRange("A1:I1").values = [["序号", "账号", "密码", "姓名", "角色", "学院", "专业", "班级", "年级"]];
list.getRange(`A2:I${n + 1}`).values = data;
list.tables.add(`A1:I${n + 1}`, true, "AccountsTable");
list.getRange("A1:I1").format = { font: { bold: true, color: "#FFFFFF" }, fill: "#101d3e" };
list.getRange(`A2:A${n + 1}`).format.numberFormat = "0";
list.getRange(`B2:B${n + 1}`).format.numberFormat = "@";
list.getRange(`C2:C${n + 1}`).format.numberFormat = "@";
list.freezePanes.freezeRows(1);
const widths = { 1: 6, 2: 14, 3: 10, 4: 12, 5: 12, 6: 26, 7: 22, 8: 12, 9: 8 };
for (const [col, w] of Object.entries(widths)) {
  list.getRangeByIndexes(1, Number(col), 1, 1).format.columnWidth = w;
}
list.showGridLines = false;

const previewList = await wb.render({ sheetName: "账号列表", range: "A1:I15", scale: 2, format: "png" });
await fs.writeFile(`${dir}/preview_list.png`, new Uint8Array(await previewList.arrayBuffer()));
const previewInfo = await wb.render({ sheetName: "说明", range: "A1:C12", scale: 2, format: "png" });
await fs.writeFile(`${dir}/preview_info.png`, new Uint8Array(await previewInfo.arrayBuffer()));

const xlsx = await SpreadsheetFile.exportXlsx(wb);
await xlsx.save(`${dir}/登录账号密码.xlsx`);
console.log(`rows=${n}`);
