import fs from "node:fs/promises";
import path from "node:path";
import { SpreadsheetFile, Workbook } from "@oai/artifact-tool";

const outputDir = path.resolve(".");

const rows = [
  {
    question: "我最近一边学习一边很焦虑，怎么平衡学业规划和心理压力？",
    answer: "先把任务缩小到本阶段最重要的 1-3 件事，不要把考研、就业、竞赛全部同时铺开。每周设置固定的休息和复盘时间，必要时联系辅导员、班主任或学校心理健康教育中心。长期睡眠、食欲或情绪明显异常时，应优先寻求专业支持。",
  },
  {
    question: "我长期情绪低落，对学习提不起兴趣，怎么办？",
    answer: "长期情绪低落、兴趣减退、睡眠或食欲异常持续存在时，建议尽快联系学校心理健康教育中心、辅导员或校医院。不要只把它当成学习动力问题。",
  },
  {
    question: "如果我有自伤或轻生念头，应该怎么办？",
    answer: "请立即联系可立即提供帮助的人，并尽量不独处。优先联系辅导员、班主任、学校心理健康教育中心或拨打当地专业援助渠道。专业诊断和治疗必须由专业机构完成。",
  },
  {
    question: "GPA 持续下降，怎么判断是否需要调整规划？",
    answer: "先看下降幅度、挂科数量和持续时间。如果连续多个学期明显下降，应优先处理学业恢复，再谈考研或就业。可以查看学业预警、咨询班主任或学业导师，制定补强计划。",
  },
  {
    question: "多门课程挂科了，还能继续准备考研吗？",
    answer: "建议先确保能按学校规定完成补考或重修，恢复当前学业稳定。挂科会影响 GPA 和毕业进度，应先把短期学业风险控制住，再重新评估考研时间线。",
  },
  {
    question: "家庭经济困难会影响我考研或留学吗？",
    answer: "不自动排除升学机会。可以先了解奖助学金、助学贷款、勤工助学、学校专项支持和社会资助，同时把家庭预算纳入路径选择。留学还要评估学费、生活费与汇率变化。",
  },
  {
    question: "我不确定考研还是就业，怎么快速缩小选项？",
    answer: "从三个问题开始：更重视学历和深度研究，还是实践和尽快进入行业；目标岗位是否要求硕士学历；当前 GPA 和核心课基础是否支撑长期备考。可先保留 1 个主路径和 1 个备选路径，而不是无限纠结。",
  },
  {
    question: "师范生只能当老师吗？",
    answer: "不是。除了教师招聘，还可考虑教育培训、教育科技、内容策划、公共管理、人力资源等方向。关键是把专业能力、实践经历和个人偏好结合起来。",
  },
  {
    question: "考研一般什么时候开始准备？",
    answer: "通常在大三上学期完成初步目标筛选，大三寒假前后开始系统复习。跨专业或基础薄弱的学生可适当提前；具体还要看目标院校、考试科目和自身基础。",
  },
  {
    question: "看到政策、分数线或报考信息，应该怎么核实？",
    answer: "以学校官方系统、招生简章、就业或教务部门通知为准。不要只依据网络二手信息。涉及报名截止、材料要求、成绩认定时，务必逐条核对官方最新版本。",
  },
  {
    question: "家长强烈反对我的发展方向，怎么办？",
    answer: "先把客观数据整理清楚：目标路径的就业、成本、成功率和备选方案。用事实沟通，而不是只表达情绪。也邀请家长一起了解学校职业发展、就业指导等资源；最终决定应由学生本人负责。",
  },
  {
    question: "学业困难、家庭变故和心理压力同时出现，怎么处理？",
    answer: "不要试图一次性解决所有问题。先保证基本作息和情绪稳定，主动联系班主任、辅导员、心理中心或校医院。可暂时降低短期目标，优先处理危机和健康，再逐步恢复学业节奏。",
  },
];

const workbook = Workbook.create();
const sheet = workbook.worksheets.add("人生规划安全FAQ");

sheet.getRange("A1:B1").values = [["question", "answer"]];
sheet.getRange("A2:B13").values = rows.map((r) => [r.question, r.answer]);

sheet.getRange("A1:B1").format = {
  fill: "#1F3864",
  font: { bold: true, color: "#FFFFFF" },
  horizontalAlignment: "center",
  verticalAlignment: "middle",
};
sheet.getRange("A1:B13").format = {
  borders: { preset: "all", style: "thin", color: "#D9D9D9" },
  wrapText: true,
  verticalAlignment: "top",
};
sheet.getRange("A1:A13").format.horizontalAlignment = "left";
sheet.getRange("B1:B13").format.horizontalAlignment = "left";
sheet.getRange("A2:B13").format.font = { color: "#111827" };
sheet.getRange("A2:B13").format.rowHeightPx = 48;
sheet.getRange("A1:A13").format.columnWidthPx = 260;
sheet.getRange("B1:B13").format.columnWidthPx = 620;
sheet.freezePanes.freezeRows(1);
sheet.showGridLines = false;

const preview = await workbook.render({
  sheetName: "人生规划安全FAQ",
  range: "A1:B13",
  scale: 1,
  format: "png",
});
await fs.writeFile(
  path.join(outputDir, "qa_preview.png"),
  new Uint8Array(await preview.arrayBuffer()),
);

const exported = await SpreadsheetFile.exportXlsx(workbook);
await exported.save(path.join(outputDir, "人生规划安全FAQ.xlsx"));
console.log("QA workbook exported");
