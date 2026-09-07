const {
  Document, Packer, Paragraph, TextRun, Table, TableRow, TableCell,
  Header, Footer, AlignmentType, HeadingLevel, BorderStyle, WidthType,
  ShadingType, PageNumber, PageBreak, VerticalAlign
} = require('docx');
const fs = require('fs');

// ── helpers ──────────────────────────────────────────────────────────────────
const W = 11906; // A4 width DXA
const ML = 1200, MR = 1200, MT = 1134, MB = 1134;
const CW = W - ML - MR; // 9506 content width

const border = { style: BorderStyle.SINGLE, size: 4, color: "BFBFBF" };
const borders = { top: border, bottom: border, left: border, right: border };
const hdrBorder = { style: BorderStyle.SINGLE, size: 4, color: "1F3864" };
const hdrBorders = { top: hdrBorder, bottom: hdrBorder, left: hdrBorder, right: hdrBorder };

const cm = (top=80, bottom=80, left=120, right=120) => ({ top, bottom, left, right });

function hdrCell(text, w, span) {
  return new TableCell({
    width: { size: w, type: WidthType.DXA },
    columnSpan: span,
    borders: hdrBorders,
    shading: { fill: "1F3864", type: ShadingType.CLEAR },
    margins: cm(),
    verticalAlign: VerticalAlign.CENTER,
    children: [new Paragraph({
      alignment: AlignmentType.CENTER,
      children: [new TextRun({ text, bold: true, color: "FFFFFF", size: 20, font: "仿宋" })]
    })]
  });
}

function dataCell(text, w, shade, bold=false, center=false) {
  return new TableCell({
    width: { size: w, type: WidthType.DXA },
    borders,
    shading: { fill: shade || "FFFFFF", type: ShadingType.CLEAR },
    margins: cm(),
    verticalAlign: VerticalAlign.CENTER,
    children: [new Paragraph({
      alignment: center ? AlignmentType.CENTER : AlignmentType.LEFT,
      children: [new TextRun({ text: text || "", bold, size: 19, font: "仿宋" })]
    })]
  });
}

function multiPara(texts, shade) {
  return new TableCell({
    borders,
    shading: { fill: shade || "FFFFFF", type: ShadingType.CLEAR },
    margins: cm(),
    children: texts.map(t => new Paragraph({
      children: [new TextRun({ text: t, size: 19, font: "仿宋" })]
    }))
  });
}

function h1(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_1,
    pageBreakBefore: true,
    children: [new TextRun({ text, font: "黑体" })]
  });
}
function h2(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_2,
    children: [new TextRun({ text, font: "黑体" })]
  });
}
function h3(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_3,
    children: [new TextRun({ text, font: "黑体" })]
  });
}
function body(text, bold=false) {
  return new Paragraph({
    spacing: { before: 60, after: 60, line: 360, lineRule: "auto" },
    children: [new TextRun({ text, bold, size: 24, font: "仿宋" })]
  });
}
function note(text) {
  return new Paragraph({
    spacing: { before: 40, after: 40 },
    children: [new TextRun({ text, size: 20, color: "7F7F7F", font: "仿宋", italics: true })]
  });
}
function gap(sz=100) {
  return new Paragraph({ spacing: { before: sz, after: sz }, children: [new TextRun("")] });
}

// simple 2-col table: col1 label (gray), col2 value
function twoCol(rows, col1W=2800) {
  const col2W = CW - col1W;
  return new Table({
    width: { size: CW, type: WidthType.DXA },
    columnWidths: [col1W, col2W],
    rows: rows.map(([k, v], i) => new TableRow({
      children: [
        dataCell(k, col1W, i % 2 === 0 ? "EBF0F9" : "F5F8FD", true),
        dataCell(v, col2W, i % 2 === 0 ? "FFFFFF" : "FAFBFE"),
      ]
    }))
  });
}

// 3-col table with header row
function threeCol(headers, rows, widths) {
  const [w1, w2, w3] = widths;
  return new Table({
    width: { size: CW, type: WidthType.DXA },
    columnWidths: widths,
    rows: [
      new TableRow({ children: headers.map((h, i) => hdrCell(h, widths[i])) }),
      ...rows.map(([c1, c2, c3], i) => new TableRow({
        children: [
          dataCell(c1, w1, i % 2 === 0 ? "EBF0F9" : "F5F8FD", true),
          dataCell(c2, w2, i % 2 === 0 ? "FFFFFF" : "FAFBFE"),
          dataCell(c3, w3, i % 2 === 0 ? "FFFFFF" : "FAFBFE"),
        ]
      }))
    ]
  });
}

// 4-col table
function fourCol(headers, rows, widths) {
  return new Table({
    width: { size: CW, type: WidthType.DXA },
    columnWidths: widths,
    rows: [
      new TableRow({ children: headers.map((h, i) => hdrCell(h, widths[i])) }),
      ...rows.map(([c1, c2, c3, c4], i) => new TableRow({
        children: [
          dataCell(c1, widths[0], i%2===0?"EBF0F9":"F5F8FD", true, true),
          dataCell(c2, widths[1], i%2===0?"FFFFFF":"FAFBFE"),
          dataCell(c3, widths[2], i%2===0?"FFFFFF":"FAFBFE"),
          dataCell(c4, widths[3], i%2===0?"FFFFFF":"FAFBFE"),
        ]
      }))
    ]
  });
}

// ── Document ─────────────────────────────────────────────────────────────────
const doc = new Document({
  styles: {
    default: {
      document: { run: { font: "仿宋", size: 24, color: "000000" } }
    },
    paragraphStyles: [
      {
        id: "Heading1", name: "Heading 1", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 36, bold: true, font: "黑体", color: "1F3864" },
        paragraph: { spacing: { before: 400, after: 200 }, outlineLevel: 0,
          border: { bottom: { style: BorderStyle.SINGLE, size: 8, color: "1F3864", space: 4 } } }
      },
      {
        id: "Heading2", name: "Heading 2", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 28, bold: true, font: "黑体", color: "1F3864" },
        paragraph: { spacing: { before: 280, after: 120 }, outlineLevel: 1 }
      },
      {
        id: "Heading3", name: "Heading 3", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 24, bold: true, font: "黑体", color: "2E5FA3" },
        paragraph: { spacing: { before: 200, after: 80 }, outlineLevel: 2 }
      },
    ]
  },

  sections: [{
    properties: {
      page: {
        size: { width: W, height: 16838 },
        margin: { top: MT, right: MR, bottom: MB, left: ML }
      }
    },
    headers: {
      default: new Header({
        children: [new Paragraph({
          border: { bottom: { style: BorderStyle.SINGLE, size: 6, color: "1F3864", space: 1 } },
          tabStops: [{ type: "right", position: 9506 }],
          children: [
            new TextRun({ text: "学生学业领航智能体系统招标技术参数文件", size: 18, color: "1F3864", font: "仿宋" }),
            new TextRun({ text: "\tV1.0  2026年7月", size: 18, color: "9E9E9E", font: "仿宋" }),
          ]
        })]
      })
    },
    footers: {
      default: new Footer({
        children: [new Paragraph({
          border: { top: { style: BorderStyle.SINGLE, size: 6, color: "1F3864", space: 1 } },
          alignment: AlignmentType.CENTER,
          children: [
            new TextRun({ text: "第 ", size: 18, color: "7F7F7F", font: "仿宋" }),
            new TextRun({ children: [PageNumber.CURRENT], size: 18, color: "1F3864", font: "仿宋", bold: true }),
            new TextRun({ text: " 页", size: 18, color: "7F7F7F", font: "仿宋" }),
          ]
        })]
      })
    },

    children: [
      // ── 封面 ──────────────────────────────────────────────────────────────
      gap(2000),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        spacing: { before: 0, after: 200 },
        children: [new TextRun({ text: "学生学业领航智能体系统", size: 64, bold: true, font: "黑体", color: "1F3864" })]
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        spacing: { before: 0, after: 600 },
        children: [new TextRun({ text: "招标技术参数文件（功能需求说明书）", size: 36, font: "仿宋", color: "2E5FA3" })]
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        border: { top: { style: BorderStyle.SINGLE, size: 8, color: "1F3864", space: 6 } },
        spacing: { before: 400, after: 100 },
        children: [new TextRun({ text: "", size: 24 })]
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "版本：V1.0", size: 24, font: "仿宋", color: "5A5A5A" })]
      }),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "日期：2026年7月", size: 24, font: "仿宋", color: "5A5A5A" })]
      }),
      gap(600),
      new Paragraph({ children: [new PageBreak()] }),

      // ── 一、项目概述 ──────────────────────────────────────────────────────
      h1("一、项目概述"),
      body("项目名称：学生学业领航智能体系统", true),
      gap(60),
      body("建设目标：面向高校全体学生、教师（含辅导员）、院系管理人员，构建一套以AI为核心引擎的学业数据分析与主动干预平台。通过接入学校统一身份认证及各业务数据系统，实现学业预警实时推送、个性化发展路径引导两大核心功能，辅以学业总览、学业画像、体检报告等基础功能模块，形成"数据感知—预警推送—智能引导"的完整学业支持闭环。"),
      gap(60),
      body("建设范围：学生端Web应用、教师/辅导员端Web应用、院系领导端Web应用、AI智能体服务层、数据采集与集成层，共五个子系统。"),
      gap(200),

      // ── 二、基本参数 ──────────────────────────────────────────────────────
      h1("二、基本参数"),
      note("基本参数为必须满足的最低要求，投标方须逐条承诺响应，不响应任意一项视为实质性不符合。"),
      gap(100),

      h2("2.1 系统架构基本参数"),
      fourCol(
        ["序号","参数项","最低要求",""],
        [
          ["A-01","前端框架","Vue 3.4及以上版本，采用TypeScript，禁止使用Vue 2",""],
          ["A-02","后端框架","Spring Boot 3.x（Java 17+），提供RESTful API",""],
          ["A-03","AI服务层","独立微服务，支持Python FastAPI或同等方案，与主应用解耦",""],
          ["A-04","数据库","MySQL 8.0及以上",""],
          ["A-05","缓存","Redis 7.0及以上，用于会话管理与消息推送",""],
          ["A-06","部署方式","支持Docker容器化部署，提供Docker Compose或K8s编排文件",""],
          ["A-07","浏览器兼容","Chrome 100+、Edge 100+、Firefox 100+正常运行",""],
          ["A-08","响应式支持","支持1280px以上宽屏分辨率自适应显示",""],
        ],
        [800, 2200, 5506, 1000]
      ),
      gap(200),

      h2("2.2 身份认证基本参数"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["B-01","认证方式","接入学校统一身份认证系统（CAS协议），不得设置独立账号密码登录入口"],
          ["B-02","角色识别","系统从统一身份认证返回的用户属性中自动识别角色，支持：学生、教师、辅导员、院系管理人员、院领导，共5类角色"],
          ["B-03","自动分流","用户登录后根据角色自动跳转对应端界面，无需用户手动选择端口"],
          ["B-04","会话管理","支持Token机制，会话有效期不少于8小时，支持无感续期"],
          ["B-05","单点登出","在统一身份认证平台退出后，本系统同步注销会话"],
        ],
        [800, 2400, 6306]
      ),
      gap(200),

      h2("2.3 学生端功能模块基本参数"),
      body("学生端须包含以下7个功能模块，不得缺失："),
      gap(80),

      h3("2.3.1 学业总览"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["C-01","KPI指标","展示累计GPA、专业排名、已修学分/应修学分、当前预警状态，共4项核心指标"],
          ["C-02","GPA趋势图","折线图展示近5个学期GPA变化趋势"],
          ["C-03","预警状态提示","首页顶部展示当前最高等级预警状态，无预警时显示"学业状态正常"绿色提示"],
        ],
        [800, 2400, 6306]
      ),
      gap(120),

      h3("2.3.2 AI广场"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["D-01","智能体数量","学生端不少于6个AI智能体"],
          ["D-02","卡片信息","每个智能体卡片须展示：名称、功能描述、数据来源标签、预置问题示例、使用状态（已上线/开发中）、使用人次"],
          ["D-03","搜索过滤","支持按关键词搜索、按功能分类筛选"],
          ["D-04","对话能力","每个智能体支持多轮对话，响应时间不超过5秒"],
        ],
        [800, 2400, 6306]
      ),
      gap(120),

      h3("2.3.3 学业画像"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["E-01","五维雷达图","包含学业成绩、学习行为、实践能力、发展意向、健康状态五个维度，与全院均值双系列对比"],
          ["E-02","维度详情","点击每个维度可展开具体子指标数据"],
          ["E-03","数据来源","各维度数据须注明数据来源与更新时间"],
        ],
        [800, 2400, 6306]
      ),
      gap(120),

      h3("2.3.4 体检报告"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["F-01","学期切换","支持查看近3年内各学期体测报告"],
          ["F-02","综合评分","展示综合健康评分及各项目（BMI、肺活量、跑步等）明细"],
          ["F-03","健康状态","红黄绿三档状态标识，超标项目高亮提示"],
        ],
        [800, 2400, 6306]
      ),
      gap(120),

      h3("2.3.5 预警通知（核心模块，详见第三章控标参数）"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["G-01","预警分级","支持黄色（关注）、橙色（警示）、红色（严重）三级预警展示"],
          ["G-02","触发信息","每条预警须展示：触发事件、推送时间、具体课程/行为原因"],
          ["G-03","确认机制","橙色及以上预警需学生主动确认已查看，系统记录确认时间"],
          ["G-04","历史记录","保留当学年全部预警历史，可按等级/类型/时间筛选"],
        ],
        [800, 2400, 6306]
      ),
      gap(120),

      h3("2.3.6 发展引导（核心模块，详见第三章控标参数）"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["H-01","路径选择","支持考研深造、出国留学、教师招聘（师范专业）、社会就业（非师范专业）等发展路径"],
          ["H-02","匹配分析","基于学业画像与目标要求对比，生成差距分析报告"],
          ["H-03","行动清单","输出带优先级标注的可执行行动项"],
          ["H-04","AI对话","支持就发展规划问题进行多轮对话咨询"],
        ],
        [800, 2400, 6306]
      ),
      gap(120),

      h3("2.3.7 快捷访问"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["I-01","快捷入口","不少于6个常用功能快捷入口，支持自定义排序"],
        ],
        [800, 2400, 6306]
      ),
      gap(200),

      h2("2.4 教师/辅导员端功能模块基本参数"),
      threeCol(
        ["序号","模块","最低要求"],
        [
          ["J-01","班级驾驶舱","展示班级KPI（总人数、均GPA、预警人数、通过率）、成绩分布饼图、预警分布图、近期预警列表"],
          ["J-02","预警管理","列表展示全班预警，支持按等级/状态筛选，可查看挂科具体科目，支持干预记录回填"],
          ["J-03","学生画像","查看个体学生五维画像、成绩趋势、预警历史"],
          ["J-04","预警推送接收","实时接收系统推送预警通知，辅导员须在教师端同步接收学生预警信息"],
          ["J-05","AI辅助","提供预警根因分析、面谈提纲生成等AI助手功能"],
        ],
        [800, 2400, 6306]
      ),
      gap(200),

      h2("2.5 院系管理端功能模块基本参数"),
      threeCol(
        ["序号","模块","最低要求"],
        [
          ["K-01","专业态势","各年级GPA分布、预警率、通过率等宏观数据"],
          ["K-02","培养方案偏差分析","按年级对比培养方案标准学分与实际选课学分差异，分2023/2024/2025/2026级展示"],
          ["K-03","课程达成度","各课程历年通过率热力图"],
          ["K-04","院领导视图","全院级数据汇总，专业健康度横向对比"],
        ],
        [800, 2400, 6306]
      ),
      gap(200),

      h2("2.6 数据集成基本参数"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["L-01","成绩数据","对接教务系统，获取学生各学期各课程成绩、GPA、挂科记录"],
          ["L-02","出勤数据","对接考勤系统，获取课堂出勤记录，精确到课次"],
          ["L-03","作业数据","对接教学平台，获取作业提交率、完成情况"],
          ["L-04","学籍数据","对接学籍系统，获取学生基本信息、专业类别（师范/非师范）、年级"],
          ["L-05","体测数据","对接体测系统，获取历年体测成绩与健康评分"],
          ["L-06","数据更新","成绩、出勤类数据更新周期不超过24小时，实时预警触发延迟不超过1小时"],
        ],
        [800, 2400, 6306]
      ),
      gap(200),

      h2("2.7 性能基本参数"),
      threeCol(
        ["序号","参数项","最低要求"],
        [
          ["M-01","页面加载","普通页面首屏加载时间不超过3秒（千兆内网环境）"],
          ["M-02","AI响应","智能体对话首字符响应时间不超过5秒"],
          ["M-03","并发支持","支持不少于500人同时在线使用"],
          ["M-04","可用性","系统年可用率不低于99.5%"],
        ],
        [800, 2400, 6306]
      ),
      gap(200),

      // ── 三、控标参数 ─────────────────────────────────────────────────────
      h1("三、控标参数"),
      note("控标参数为本项目核心竞争力要求，评分时重点考察，须提供实现方案说明及可验证的技术证明材料。"),
      gap(100),

      // 3.1 预警推送
      h2("3.1 【控标参数一】预警实时推送闭环机制"),
      body("背景：预警不能仅停留在系统内被动查阅，必须在触发时主动送达相关人员并形成处置闭环。"),
      gap(100),

      h3("3.1.1 推送触发条件（须全部支持）"),
      fourCol(
        ["触发场景","预警等级","推送对象",""],
        [
          ["单门课程成绩不及格（期末）","黄色","学生本人",""],
          ["累计挂科达2门","橙色","学生本人 + 班级辅导员",""],
          ["累计挂科达3门","橙色","学生本人 + 辅导员 + 专业负责人",""],
          ["累计挂科超5门","红色","学生本人 + 辅导员 + 专业负责人 + 院领导",""],
          ["单门课程出勤率低于75%","黄色","学生本人 + 任课教师",""],
          ["连续3周出勤率低于75%","橙色","学生本人 + 辅导员",""],
          ["GPA跌破2.5","黄色","学生本人",""],
          ["GPA跌破2.0","橙色","学生本人 + 辅导员",""],
          ["作业未提交率连续2周超40%","黄色","学生本人 + 任课教师",""],
        ],
        [3400, 1500, 3606, 1000]
      ),
      gap(140),

      h3("3.1.2 推送通道要求"),
      body("须同时支持以下推送通道中的至少2种："),
      gap(80),
      threeCol(
        ["通道","说明","备注"],
        [
          ["系统内推送","用户登录后首页顶部弹出预警提示，预警通知模块红点计数","必须支持"],
          ["短信推送","红色预警强制触发短信通知，手机号来源学籍系统","必须支持"],
          ["企业微信/钉钉推送","接入学校已有IM系统，支持工作通知卡片推送","投标时须说明拟对接平台"],
        ],
        [2000, 5000, 2506]
      ),
      gap(140),

      h3("3.1.3 学生端与教师端信息同步要求【关键控标点】"),
      body("同一条预警在学生端和教师/辅导员端须状态同步，具体要求："),
      gap(80),
      threeCol(
        ["同步场景","技术要求",""],
        [
          ["预警触发","学生端与教师端在同一条预警触发后30分钟内同步显示",""],
          ["学生确认已查看","教师端预警列表对应记录实时更新"学生已查看"状态",""],
          ["教师完成干预回填","学生端预警详情页显示"班主任已联系"提示及联系时间",""],
          ["预警解除","学生端与教师端同步标记为"已解除"，历史记录保留",""],
        ],
        [2800, 4700, 2006]
      ),
      gap(80),
      note("投标方须提供该功能的技术实现方案，说明采用WebSocket长连接/SSE/轮询的具体策略及其可靠性保障措施。"),
      gap(140),

      h3("3.1.4 预警内容详细度要求【关键控标点】"),
      body("挂科类预警必须展示具体课程名称，不得仅显示挂科数量。具体要求："),
      gap(80),
      twoCol([
        ["挂科预警","红橙色挂科预警须列出全部挂科课程名称（以标签形式逐课展示）"],
        ["出勤预警","须展示具体课程名称与缺课次数"],
        ["成绩预警","须展示具体课程成绩与班级均分对比"],
      ], 2400),
      gap(200),

      // 3.2 发展引导
      h2("3.2 【控标参数二】智能发展引导系统"),
      body("背景：发展引导是本项目最具差异化价值的核心功能，要求系统能主动采集外部真实岗位/考研要求，结合趋势预测，对每个学生进行个性化反推分析。"),
      gap(100),

      h3("3.2.1 专业类别区分机制【关键控标点】"),
      body("系统须从学籍数据中自动识别专业类别，并采用不同的信息采集与推荐策略："),
      gap(80),
      threeCol(
        ["专业类别","识别方式","发展路径差异"],
        [
          ["师范专业","学籍系统专业属性字段","重点采集中小学、幼儿园教师招聘信息；推送教师资格证考试要求；对接当地教育局及各校招聘公告"],
          ["非师范专业","同上","重点对接BOSS直聘、前程无忧、智联招聘、猎聘等社会招聘平台；按专业方向分类采集岗位要求"],
        ],
        [1800, 2200, 5506]
      ),
      gap(80),
      note("两类专业均须支持：考研院校要求采集、出国留学院校要求采集（作为可选路径）。"),
      gap(140),

      h3("3.2.2 外部数据采集能力【关键控标点】"),
      body("系统须具备对以下外部来源进行结构化信息采集的能力："),
      gap(80),
      body("师范专业采集来源（须支持）：", true),
      gap(60),
      threeCol(
        ["采集来源","采集内容",""],
        [
          ["省级教育厅/局官网","教师招聘公告、学历学位要求、专业方向要求",""],
          ["各地市中小学官网","校招公告、岗位要求、薪资区间",""],
          ["教师资格证报名官网","考试科目、报名条件、合格标准变化",""],
          ["高校附属学校官网","实验学校、附属学校招聘要求",""],
        ],
        [3000, 5006, 1500]
      ),
      gap(100),
      body("非师范专业采集来源（须支持）：", true),
      gap(60),
      threeCol(
        ["采集来源","采集内容",""],
        [
          ["BOSS直聘","岗位名称、学历要求、专业要求、技能标签、薪资区间、城市",""],
          ["前程无忧","同上",""],
          ["智联招聘","同上",""],
          ["猎聘","高端岗位、技能要求",""],
          ["国聘网","国企/央企校招岗位要求",""],
        ],
        [2400, 5606, 1500]
      ),
      gap(100),
      body("考研信息采集来源（须支持，所有专业）：", true),
      gap(60),
      threeCol(
        ["采集来源","采集内容",""],
        [
          ["各高校研究生招生信息网","报考条件、分数线历年变化、推免比例",""],
          ["中国研究生招生信息网","全国分数线、报录比",""],
        ],
        [3000, 5006, 1500]
      ),
      gap(80),
      note("投标方须说明数据采集的技术方案，包括：采集频率、反反爬策略、数据清洗与结构化方法，以及在平台改版情况下的维护机制。"),
      gap(140),

      h3("3.2.3 要求趋势分析能力【关键控标点】"),
      body("系统须对采集到的招聘/考研要求进行时间序列分析，识别要求变化趋势，并对大一、大二学生进行未来预测提示。"),
      gap(80),
      body("趋势识别维度：", true),
      gap(60),
      fourCol(
        ["趋势类型","说明","示例",""],
        [
          ["学历要求趋势","识别某类岗位对学历要求是否在逐年提高","近3年该类岗位本科占比从65%降至40%，硕士要求持续提升",""],
          ["分数线趋势","识别目标院校专业分数线年际变化方向","该专业国家线3年上涨12分，预测3年后约需380分",""],
          ["技能要求趋势","识别新兴技能标签出现频率变化","AI相关技能词出现频率3年增长340%，已成必备项",""],
          ["竞争激烈度趋势","岗位数量与投递量变化","该岗位近2年报录比持续上升，竞争加剧",""],
        ],
        [2000, 2500, 3506, 1500]
      ),
      gap(100),
      body("趋势预测展示要求：", true),
      gap(60),
      twoCol([
        ["大一学生","显示"3年后预测要求""],
        ["大二学生","显示"2年后预测要求""],
        ["大三学生","显示"1年后预测要求""],
        ["置信度说明","预测结论须注明置信度与数据来源时间"],
      ], 2000),
      gap(140),

      h3("3.2.4 个性化差距反推分析【关键控标点】"),
      body("基于外部要求与个人学业画像，系统须能对每个学生进行反推计算，输出以下维度（须全部覆盖）："),
      gap(80),
      threeCol(
        ["分析维度","输出内容示例",""],
        [
          ["学分差距","目标岗位要求计算机类专业必修学分≥80，你目前已修52分，还需28分，建议3-4学期内补足",""],
          ["GPA差距","目标院校近3年录取均GPA 3.7，你当前3.62，尚需提升0.08，可重点优化下学期专业核心课",""],
          ["竞赛/证书差距","目标岗位高频要求：计算机等级证书（NCRE）、数学建模获奖经历，你尚未取得，建议大三前完成",""],
          ["技能差距","岗位高频技能词与你画像对比：Python（欠缺）、数据库操作（已具备）、机器学习（欠缺）",""],
          ["实践经历差距","目标岗位80%要求1段以上实习经历，你尚无实习记录，建议大三暑假参与校企合作项目",""],
          ["教师资格差距（师范专业）","目标学校要求持有教师资格证，你尚未参加考试，下次笔试时间为XX月，建议本学期开始备考",""],
        ],
        [2400, 5506, 1600]
      ),
      gap(140),

      h3("3.2.5 AI对话智能体能力"),
      body("发展引导模块须内置AI对话智能体，支持以下对话场景："),
      gap(80),
      threeCol(
        ["对话场景","示例输入","期望输出能力"],
        [
          ["路径询问","我想做小学老师，需要准备什么？","结合学生专业类型（师范/非师范）、当前画像、最新采集的招聘要求，输出个性化建议"],
          ["差距询问","我考浙大教育学研究生差多少？","调取目标院校近3年分数线与该生当前GPA对比，给出量化差距与行动建议"],
          ["趋势询问","3年后找前端开发工作需要会什么？","基于趋势分析结果，预测技能要求变化，输出学习路线图"],
          ["计划制定","帮我制定大三的考研备考计划","结合课程安排、考试时间节点，生成个性化时间规划"],
          ["竞赛咨询","有哪些竞赛对我的就业目标有帮助？","匹配岗位高频竞赛要求，推荐适合该生当前阶段参加的竞赛"],
        ],
        [1600, 2800, 5106]
      ),
      gap(80),
      body("AI对话技术要求：", true),
      gap(60),
      twoCol([
        ["知识库技术","须基于检索增强生成（RAG）技术，知识库包含实时采集的招聘数据与考研数据"],
        ["更新频率","招聘数据不少于每周1次，考研数据不少于每月1次"],
        ["数据引用","对话须引用具体数据来源（如"数据来源：XX招聘平台，采集时间：XXXX年XX月"）"],
        ["响应时间","单次对话响应时间不超过8秒"],
      ], 2400),
      gap(200),

      // 3.3 学业画像
      h2("3.3 【控标参数三】学业画像数据深度"),
      threeCol(
        ["序号","控标要求","说明"],
        [
          ["N-01","五维画像原始数据可查","点击任意维度可展开查看该维度所有原始计算数据，不可仅展示分数"],
          ["N-02","画像历史对比","支持查看近3个学期画像变化对比，以折线或雷达叠加图形式展示"],
          ["N-03","同辈对比","各维度分数须与班级均值、年级均值进行对比展示"],
          ["N-04","预警联动","画像中某维度异常时，自动关联到预警模块对应预警条目"],
        ],
        [800, 3000, 5706]
      ),
      gap(200),

      // 3.4 安全合规
      h2("3.4 【控标参数四】系统安全与数据合规"),
      threeCol(
        ["序号","控标要求","最低标准"],
        [
          ["O-01","数据加密","学生成绩、GPA等敏感字段存储加密（AES-256-GCM），传输使用HTTPS/TLS 1.3"],
          ["O-02","权限隔离","学生只能查看本人数据；教师只能查看本班学生数据；院系管理员只能查看本院数据，代码层面强制执行，不依赖前端控制"],
          ["O-03","操作审计","教师查看学生画像、导出数据等敏感操作须记录操作日志，日志保留不少于1年"],
          ["O-04","外部数据合规","采集外部招聘/考研数据须符合相关平台用户协议，采集频率须在合理范围内，须提供合规说明"],
          ["O-05","等保要求","系统须满足网络安全等级保护二级（等保2.0）要求，提供合规承诺书"],
        ],
        [800, 2800, 5906]
      ),
      gap(200),

      // 3.5 可交付成果
      h2("3.5 【控标参数五】可交付成果要求"),
      threeCol(
        ["序号","交付物","要求"],
        [
          ["P-01","系统源代码","前后端完整源代码，含注释，移交时须通过代码质量检查"],
          ["P-02","数据库设计文档","完整ER图及表结构说明文档"],
          ["P-03","API接口文档","全量RESTful API文档，含请求/响应示例"],
          ["P-04","部署手册","从零到上线的完整部署操作手册"],
          ["P-05","用户操作手册","学生端、教师端、管理端分角色操作手册"],
          ["P-06","接口对接规范","与学校各业务系统（教务/学籍/体测）对接的技术接口规范文档"],
          ["P-07","培训服务","系统上线前须提供不少于2次现场培训（教师端培训、管理员培训）"],
          ["P-08","质保期","系统上线验收后质保期不少于12个月，质保期内缺陷免费修复"],
        ],
        [800, 2400, 6306]
      ),
      gap(200),

      // ── 四、评分参考 ─────────────────────────────────────────────────────
      h1("四、评分参考（技术分建议权重）"),
      threeCol(
        ["评分项","分值","说明"],
        [
          ["基本参数响应完整性","20分","逐条响应，不可有未应答项"],
          ["预警推送闭环方案","25分","重点评审学生端/教师端同步机制的技术可行性"],
          ["发展引导与外部数据采集方案","30分","重点评审师范/非师范区分机制、趋势分析能力、AI对话RAG实现方案"],
          ["数据安全与合规方案","10分","等保材料、隐私保护方案"],
          ["类似项目业绩","10分","近3年同类高校数据平台或AI智能体项目经验"],
          ["服务与保障方案","5分","响应时效承诺、本地化服务能力"],
        ],
        [3800, 1500, 4206]
      ),
      gap(200),

      // ── 说明 ─────────────────────────────────────────────────────────────
      new Paragraph({
        border: { top: { style: BorderStyle.SINGLE, size: 6, color: "1F3864", space: 4 } },
        spacing: { before: 200, after: 60 },
        children: [new TextRun({ text: "文件说明", bold: true, size: 22, font: "黑体", color: "1F3864" })]
      }),
      note("本技术参数文件版本：V1.0，编制日期：2026年7月。"),
      note("如与招标文件其他部分存在冲突，以本技术参数文件为准。"),
      note("投标方须对基本参数逐条作出明确承诺响应（"响应"/"不响应"），控标参数须附技术实现方案说明。"),
    ]
  }]
});

Packer.toBuffer(doc).then(buf => {
  fs.writeFileSync("D:\\Academic Navigation\\方案\\AI学业领航项目招标技术参数文件.docx", buf);
  console.log("DONE");
});
