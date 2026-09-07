# -*- coding: utf-8 -*-
"""Generate 学业领航智能体系统-详细功能及技术参数（招标版）.docx"""
from docx import Document
from docx.shared import Pt, Cm, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH, WD_TAB_ALIGNMENT, WD_BREAK
from docx.enum.table import WD_TABLE_ALIGNMENT, WD_ALIGN_VERTICAL
from docx.oxml.ns import qn
from docx.oxml import OxmlElement

# ---------------------------------------------------------------- tokens
PAGE_W, PAGE_H = Cm(21.0), Cm(29.7)
MARGIN_TB, MARGIN_LR = Cm(2.5), Cm(2.6)
CONTENT_CM = 15.8
CONTENT_DXA = 8958
BODY_EAST = "宋体"
HEAD_EAST = "微软雅黑"
ASCII_FONT = "Calibri"
MONO = "Consolas"
INK = "0B2545"
BLUE = "1F4D78"
BLUE2 = "2E74B5"
GRAY = "5A6B85"
MUTED = "9AA5B8"
HDR_FILL = "E8EEF5"
NOTE_FILL = "F4F6F9"
CODE_FILL = "F3F5F9"

doc = Document()
cp = doc.core_properties
cp.title = "学业领航智能体系统-详细功能及技术参数（招标版）"
cp.author = "项目组"
cp.subject = "招标功能、技术参数与验收基准"
cp.keywords = "学业领航, 智能体, 招标, 技术参数"

# ---------------------------------------------------------------- page
sec = doc.sections[0]
sec.page_width = PAGE_W
sec.page_height = PAGE_H
sec.top_margin = MARGIN_TB
sec.bottom_margin = MARGIN_TB
sec.left_margin = MARGIN_LR
sec.right_margin = MARGIN_LR
sec.header_distance = Cm(1.5)
sec.footer_distance = Cm(1.5)
sec.different_first_page_header_footer = True


def cm_to_dxa(value):
    return int(round(value * 566.929))


def fmt(run, size=11, bold=False, color=None, east=BODY_EAST, ascii_font=ASCII_FONT):
    run.font.name = ascii_font
    run.font.size = Pt(size)
    run.font.bold = bold
    if color:
        run.font.color.rgb = RGBColor.from_string(color)
    rPr = run._element.get_or_add_rPr()
    rFonts = rPr.get_or_add_rFonts()
    rFonts.set(qn("w:ascii"), ascii_font)
    rFonts.set(qn("w:hAnsi"), ascii_font)
    rFonts.set(qn("w:eastAsia"), east)


def para(text="", size=11, bold=False, color=None, align=None, before=None, after=6,
         line=1.25, east=BODY_EAST, style=None):
    p = doc.add_paragraph(style=style)
    pf = p.paragraph_format
    if before is not None:
        pf.space_before = Pt(before)
    pf.space_after = Pt(after)
    pf.line_spacing = line
    if align is not None:
        p.alignment = align
    if text:
        r = p.add_run(text)
        fmt(r, size=size, bold=bold, color=color, east=east)
    return p


def para_parts(parts, size=11, align=None, before=None, after=6, line=1.25, east=BODY_EAST):
    p = doc.add_paragraph()
    pf = p.paragraph_format
    if before is not None:
        pf.space_before = Pt(before)
    pf.space_after = Pt(after)
    pf.line_spacing = line
    if align is not None:
        p.alignment = align
    for text, bold, color in parts:
        r = p.add_run(text)
        fmt(r, size=size, bold=bold, color=color or INK if bold else None, east=east)
    return p


def bullet(text, prefix=None, size=11):
    p = doc.add_paragraph(style="List Bullet")
    pf = p.paragraph_format
    pf.space_after = Pt(4)
    pf.line_spacing = 1.25
    if prefix:
        r = p.add_run(prefix)
        fmt(r, size=size, bold=True, color=INK)
    r = p.add_run(text)
    fmt(r, size=size)
    return p


def numbered(text, prefix=None, size=11):
    p = doc.add_paragraph(style="List Number")
    pf = p.paragraph_format
    pf.space_after = Pt(4)
    pf.line_spacing = 1.25
    if prefix:
        r = p.add_run(prefix)
        fmt(r, size=size, bold=True, color=INK)
    r = p.add_run(text)
    fmt(r, size=size)
    return p


def h1(text):
    p = doc.add_paragraph(style="Heading 1")
    p.add_run(text)
    p.paragraph_format.space_before = Pt(18)
    p.paragraph_format.space_after = Pt(10)
    p.paragraph_format.keep_with_next = True
    return p


def h2(text):
    p = doc.add_paragraph(style="Heading 2")
    p.add_run(text)
    p.paragraph_format.space_before = Pt(14)
    p.paragraph_format.space_after = Pt(7)
    p.paragraph_format.keep_with_next = True
    return p


def h3(text):
    p = doc.add_paragraph(style="Heading 3")
    p.add_run(text)
    p.paragraph_format.space_before = Pt(10)
    p.paragraph_format.space_after = Pt(5)
    p.paragraph_format.keep_with_next = True
    return p


def page_break():
    p = doc.add_paragraph()
    p.paragraph_format.space_after = Pt(0)
    p.add_run().add_break(WD_BREAK.PAGE)


def shade_cell(cell, fill):
    tcPr = cell._tc.get_or_add_tcPr()
    shd = OxmlElement("w:shd")
    shd.set(qn("w:val"), "clear")
    shd.set(qn("w:color"), "auto")
    shd.set(qn("w:fill"), fill)
    tcPr.append(shd)


def set_cell_margins(table, top=80, bottom=80, start=120, end=120):
    tblPr = table._tbl.tblPr
    tblCellMar = OxmlElement("w:tblCellMar")
    for name, val in (("top", top), ("left", start), ("bottom", bottom), ("right", end)):
        el = OxmlElement("w:" + name)
        el.set(qn("w:w"), str(val))
        el.set(qn("w:type"), "dxa")
        tblCellMar.append(el)
    tblPr.append(tblCellMar)


def set_table_borders(table, grid=True):
    tblPr = table._tbl.tblPr
    borders = tblPr.find(qn("w:tblBorders"))
    if borders is None:
        borders = OxmlElement("w:tblBorders")
        tblPr.append(borders)
    if grid:
        for name in ("top", "left", "bottom", "right", "insideH", "insideV"):
            el = borders.find(qn("w:" + name))
            if el is None:
                el = OxmlElement("w:" + name)
                borders.append(el)
            el.set(qn("w:val"), "single")
            el.set(qn("w:sz"), "4")
            el.set(qn("w:space"), "0")
            el.set(qn("w:color"), "B7C3D4")
    else:
        for name in ("top", "left", "bottom", "right", "insideH", "insideV"):
            el = borders.find(qn("w:" + name))
            if el is None:
                el = OxmlElement("w:" + name)
                borders.append(el)
            el.set(qn("w:val"), "nil")


def set_table_geometry(table, widths_cm):
    widths_dxa = [cm_to_dxa(w) for w in widths_cm]
    total = sum(widths_dxa)
    assert 0 < total <= CONTENT_DXA, (widths_cm, total)
    table.autofit = False
    tblPr = table._tbl.tblPr
    tblW = tblPr.find(qn("w:tblW"))
    if tblW is None:
        tblW = OxmlElement("w:tblW")
        tblPr.append(tblW)
    tblW.set(qn("w:w"), str(total))
    tblW.set(qn("w:type"), "dxa")
    tblInd = tblPr.find(qn("w:tblInd"))
    if tblInd is None:
        tblInd = OxmlElement("w:tblInd")
        tblPr.append(tblInd)
    tblInd.set(qn("w:w"), "0")
    tblInd.set(qn("w:type"), "dxa")
    tblLayout = tblPr.find(qn("w:tblLayout"))
    if tblLayout is None:
        tblLayout = OxmlElement("w:tblLayout")
        tblPr.append(tblLayout)
    tblLayout.set(qn("w:type"), "fixed")
    grid = table._tbl.find(qn("w:tblGrid"))
    cols = grid.findall(qn("w:gridCol"))
    for gc, w in zip(cols, widths_dxa):
        gc.set(qn("w:w"), str(w))
    for row in table.rows:
        for cell, w in zip(row.cells, widths_dxa):
            tcPr = cell._tc.get_or_add_tcPr()
            tcW = tcPr.find(qn("w:tcW"))
            if tcW is None:
                tcW = OxmlElement("w:tcW")
                tcPr.append(tcW)
            tcW.set(qn("w:w"), str(w))
            tcW.set(qn("w:type"), "dxa")


def repeat_header_row(table):
    trPr = table.rows[0]._tr.get_or_add_trPr()
    tblHeader = OxmlElement("w:tblHeader")
    tblHeader.set(qn("w:val"), "true")
    trPr.append(tblHeader)


def cell_text(cell, text, size=10.5, bold=False, color=None, align="left"):
    cell.vertical_alignment = WD_ALIGN_VERTICAL.CENTER
    p = cell.paragraphs[0]
    p.paragraph_format.space_after = Pt(2)
    p.paragraph_format.space_before = Pt(2)
    p.paragraph_format.line_spacing = 1.15
    if align == "center":
        p.alignment = WD_ALIGN_PARAGRAPH.CENTER
    elif align == "right":
        p.alignment = WD_ALIGN_PARAGRAPH.RIGHT
    else:
        p.alignment = WD_ALIGN_PARAGRAPH.LEFT
    for run in list(p.runs):
        run._element.getparent().remove(run._element)
    r = p.add_run(text)
    fmt(r, size=size, bold=bold, color=color)


def caption(text):
    return para(text, size=10.5, bold=True, color=BLUE, before=10, after=4)


def spacer(pts=2):
    p = doc.add_paragraph()
    p.paragraph_format.space_after = Pt(pts)
    p.paragraph_format.space_before = Pt(0)
    return p


def add_table(headers, rows, widths_cm, aligns=None, font_size=10.5, caption_text=None):
    if caption_text:
        caption(caption_text)
    table = doc.add_table(rows=1 + len(rows), cols=len(headers))
    table.alignment = WD_TABLE_ALIGNMENT.LEFT
    set_table_borders(table, grid=True)
    set_cell_margins(table)
    set_table_geometry(table, widths_cm)
    if aligns is None:
        aligns = ["left"] * len(headers)
    for j, htext in enumerate(headers):
        cell = table.rows[0].cells[j]
        cell_text(cell, htext, size=font_size, bold=True, color=INK, align="center")
        shade_cell(cell, HDR_FILL)
    repeat_header_row(table)
    for i, row in enumerate(rows):
        for j, value in enumerate(row):
            cell = table.rows[i + 1].cells[j]
            cell_text(cell, value, size=font_size, align=aligns[j])
    spacer(3)
    return table


def add_note(text, title="说明"):
    table = doc.add_table(rows=1, cols=1)
    table.alignment = WD_TABLE_ALIGNMENT.LEFT
    set_table_borders(table, grid=False)
    set_cell_margins(table, top=100, bottom=100, start=140, end=140)
    set_table_geometry(table, [CONTENT_CM])
    cell = table.rows[0].cells[0]
    cell.vertical_alignment = WD_ALIGN_VERTICAL.CENTER
    shade_cell(cell, NOTE_FILL)
    p = cell.paragraphs[0]
    p.paragraph_format.space_after = Pt(2)
    p.paragraph_format.line_spacing = 1.2
    r = p.add_run(title + "：")
    fmt(r, size=10.5, bold=True, color=BLUE)
    r = p.add_run(text)
    fmt(r, size=10.5)
    spacer(3)
    return table


def add_code(lines):
    table = doc.add_table(rows=1, cols=1)
    table.alignment = WD_TABLE_ALIGNMENT.LEFT
    set_table_borders(table, grid=False)
    set_cell_margins(table, top=100, bottom=100, start=140, end=140)
    set_table_geometry(table, [CONTENT_CM])
    cell = table.rows[0].cells[0]
    shade_cell(cell, CODE_FILL)
    first = True
    for line in lines:
        p = cell.paragraphs[0] if first else cell.add_paragraph()
        first = False
        p.paragraph_format.space_after = Pt(0)
        p.paragraph_format.line_spacing = 1.15
        r = p.add_run(line)
        fmt(r, size=9, color=INK, east=BODY_EAST, ascii_font=MONO)
    spacer(3)
    return table


def add_field(paragraph, instr, size=9, color=None):
    run = paragraph.add_run()
    fmt(run, size=size, color=color)
    fld1 = OxmlElement("w:fldChar")
    fld1.set(qn("w:fldCharType"), "begin")
    it = OxmlElement("w:instrText")
    it.set(qn("xml:space"), "preserve")
    it.text = instr
    fld2 = OxmlElement("w:fldChar")
    fld2.set(qn("w:fldCharType"), "end")
    run._r.append(fld1)
    run._r.append(it)
    run._r.append(fld2)
    return run


def setup_header_footer():
    header = sec.header
    hp = header.paragraphs[0]
    hp.text = ""
    pf = hp.paragraph_format
    pf.space_after = Pt(0)
    pf.line_spacing = 1.0
    pf.tab_stops.add_tab_stop(Cm(15.4), WD_TAB_ALIGNMENT.RIGHT)
    r = hp.add_run("学业领航智能体系统 · 详细功能及技术参数（招标版）")
    fmt(r, size=9, color=MUTED)
    r2 = hp.add_run("\t第 ")
    fmt(r2, size=9, color=MUTED)
    add_field(hp, " PAGE ", size=9, color=MUTED)
    r3 = hp.add_run(" 页")
    fmt(r3, size=9, color=MUTED)
    pPr = hp._p.get_or_add_pPr()
    pBdr = OxmlElement("w:pBdr")
    bottom = OxmlElement("w:bottom")
    bottom.set(qn("w:val"), "single")
    bottom.set(qn("w:sz"), "4")
    bottom.set(qn("w:space"), "2")
    bottom.set(qn("w:color"), "C9D3E0")
    pBdr.append(bottom)
    pPr.append(pBdr)

    footer = sec.footer
    fp = footer.paragraphs[0]
    fp.text = ""
    fp.alignment = WD_ALIGN_PARAGRAPH.CENTER
    fp.paragraph_format.space_after = Pt(0)
    fp.paragraph_format.line_spacing = 1.0
    r = fp.add_run("第 ")
    fmt(r, size=9, color=MUTED)
    add_field(fp, " PAGE ", size=9, color=MUTED)
    r = fp.add_run(" 页，共 ")
    fmt(r, size=9, color=MUTED)
    add_field(fp, " NUMPAGES ", size=9, color=MUTED)
    r = fp.add_run(" 页")
    fmt(r, size=9, color=MUTED)


def patch_list_numbering():
    try:
        numbering = doc.part.numbering_part.element
        nums = {int(n.get(qn("w:numId"))): n for n in numbering.findall(qn("w:num"))}
        for num_id in (1, 5):
            n = nums.get(num_id)
            if n is None:
                continue
            abs_id = n.find(qn("w:abstractNumId")).get(qn("w:val"))
            for an in numbering.findall(qn("w:abstractNum")):
                if an.get(qn("w:abstractNumId")) != abs_id:
                    continue
                for lvl in an.findall(qn("w:lvl")):
                    if lvl.get(qn("w:ilvl")) != "0":
                        continue
                    ind = lvl.find(qn("w:ind"))
                    if ind is None:
                        ind = OxmlElement("w:ind")
                        lvl.append(ind)
                    ind.set(qn("w:left"), "720")
                    ind.set(qn("w:hanging"), "360")
                    sp = lvl.find(qn("w:spacing"))
                    if sp is None:
                        sp = OxmlElement("w:spacing")
                        lvl.append(sp)
                    sp.set(qn("w:after"), "80")
    except Exception:
        pass


# ---------------------------------------------------------------- styles
normal = doc.styles["Normal"]
normal.font.name = ASCII_FONT
normal.font.size = Pt(11)
rpr = normal.element.get_or_add_rPr()
rf = rpr.get_or_add_rFonts()
rf.set(qn("w:ascii"), ASCII_FONT)
rf.set(qn("w:hAnsi"), ASCII_FONT)
rf.set(qn("w:eastAsia"), BODY_EAST)
npf = normal.paragraph_format
npf.space_after = Pt(6)
npf.line_spacing = 1.25

for name, size, color, before, after in (
    ("Heading 1", 16, BLUE, 18, 10),
    ("Heading 2", 13, BLUE2, 14, 7),
    ("Heading 3", 12, BLUE, 10, 5),
):
    st = doc.styles[name]
    st.font.name = ASCII_FONT
    st.font.size = Pt(size)
    st.font.bold = True
    st.font.color.rgb = RGBColor.from_string(color)
    st.element.get_or_add_rPr()
    rF = st.element.rPr.get_or_add_rFonts()
    rF.set(qn("w:ascii"), ASCII_FONT)
    rF.set(qn("w:hAnsi"), ASCII_FONT)
    rF.set(qn("w:eastAsia"), HEAD_EAST)
    st.paragraph_format.space_before = Pt(before)
    st.paragraph_format.space_after = Pt(after)
    st.paragraph_format.keep_with_next = True

patch_list_numbering()
setup_header_footer()

# ---------------------------------------------------------------- cover
for _ in range(4):
    para("", after=14)
para("ACADEMIC NAVIGATION SYSTEM", size=11, bold=True, color=BLUE2,
     align=WD_ALIGN_PARAGRAPH.CENTER, after=12)
para("学业领航智能体系统", size=26, bold=True, color=INK,
     align=WD_ALIGN_PARAGRAPH.CENTER, after=6)
para("详细功能及技术参数说明（招标版）", size=15, bold=True, color=BLUE2,
     align=WD_ALIGN_PARAGRAPH.CENTER, after=20)
para("版本 v1.0  |  编制日期：2026年8月11日", size=11, color=GRAY,
     align=WD_ALIGN_PARAGRAPH.CENTER, after=26)

meta = doc.add_table(rows=5, cols=2)
meta.alignment = WD_TABLE_ALIGNMENT.CENTER
set_table_borders(meta, grid=False)
set_cell_margins(meta, top=60, bottom=60, start=120, end=120)
set_table_geometry(meta, [3.6, 9.2])
meta_rows = [
    ("文档版本", "v1.0（招标技术参数基准版）"),
    ("编制依据", "《项目进度说明》《后端架构设计》"),
    ("适用对象", "高校教务管理部门、采购方、投标方"),
    ("文档用途", "系统详细功能、技术参数与验收标准说明"),
    ("交付形态", "软件系统（含前端五端、后端服务、AI 智能体）"),
]
for i, (k, v) in enumerate(meta_rows):
    c0 = meta.rows[i].cells[0]
    c1 = meta.rows[i].cells[1]
    cell_text(c0, k, size=11, bold=True, color=INK, align="right")
    cell_text(c1, v, size=11, color=GRAY, align="left")

para("", after=22)
rule = doc.add_paragraph()
rule.paragraph_format.space_after = Pt(4)
pPr = rule._p.get_or_add_pPr()
pBdr = OxmlElement("w:pBdr")
bottom = OxmlElement("w:bottom")
bottom.set(qn("w:val"), "single")
bottom.set(qn("w:sz"), "8")
bottom.set(qn("w:space"), "1")
bottom.set(qn("w:color"), BLUE2)
pBdr.append(bottom)
pPr.append(pBdr)
page_break()

# ---------------------------------------------------------------- chapter 1
h1("第一章 系统概述")

h2("1.1 项目背景与系统定位")
para("学业领航智能体系统是面向高校的全链路学业管理与智能决策支持平台。系统通过 AI 智能体技术，"
     "把分散在教务、学工、体测、就业等系统中的学业数据汇聚为统一的学生全景档案，为学生、教师和各级管理者"
     "提供个性化的学业诊断、预警干预和发展引导服务，实现从“成绩管理”到“过程管理、智能干预、数据决策”的升级。")
para_parts([("核心定位：", True, INK),
            ("以学生学业健康为主线，以五端门户为入口，以 12 个角色化 AI 智能体为服务载体，"
             "构建“数据统一采集—多维画像—智能预警—闭环干预—发展引导—决策支持”的一体化能力。", False, None)])

h2("1.2 建设目标")
bullet("数据统一：对接教务、学工、体测、就业等系统，统一学生成绩、GPA、出勤、画像、健康与就业数据口径。")
bullet("五端协同：学生端、班主任端、任课教师端、系主任端、院领导端按角色提供差异化功能与数据权限。")
bullet("智能预警：基于 GPA、成绩、出勤等多维规则自动触发黄、橙、红三级预警，并支持确认、干预、解除的闭环管理。")
bullet("AI 赋能：提供学业咨询、发展路径规划、学情分析、课程健康度诊断等 12 个角色化智能体，支持流式对话。")
bullet("安全合规：支持学校统一身份认证（CAS/OAuth2）、JWT 鉴权、数据权限隔离与等保三级合规。")

h2("1.3 系统总体架构")
para("系统采用前后端分离的分层架构，自上而下分为前端五端、后端服务、数据层、外部数据适配层与 AI 服务层：")
bullet("前端五端：学生端、班主任端、任课教师端、系主任端、院领导端，基于 Vue 3 + TypeScript + Element Plus + ECharts 构建。")
bullet("后端服务：Spring Boot 3.x 应用服务，统一接口、权限校验、业务编排与预警规则引擎，API 文档由 Knife4j/Swagger 提供。")
bullet("数据层：MySQL 8.0 主库 + Redis 7.x 缓存，Flyway 管理数据库版本；RabbitMQ 承担异步通知，XXL-Job 承担定时同步与报表。")
bullet("外部数据适配层：以 AcademicDataSource 抽象接口统一对接教务、学工、体测、就业、统一认证等外部系统，支持 Mock 与真实数据源切换。")
bullet("AI 服务层：通义千问、文心一言、GPT-4 等大模型可配置接入，支持 SSE 流式对话、RAG 检索增强与教育场景 Prompt 模板。")

h2("1.4 用户角色体系")
add_table(
    ["角色", "默认落地页", "核心职责", "数据范围"],
    [
        ["学生", "AI 领航助手", "查看个人学业、接收预警、发展引导", "本人数据"],
        ["班主任", "班级驾驶舱", "班级学情监控、预警处置、干预跟踪", "所负责班级"],
        ["任课教师", "我的课程", "课程成绩分析、课程维度学情预警", "所授课程"],
        ["系主任", "专业态势", "专业学情分析、课程健康度监控", "所属专业"],
        ["院领导", "全院仪表盘", "全院质量监控、决策支持", "全院数据"],
    ],
    [1.9, 2.6, 6.6, 4.7],
    aligns=["center", "center", "left", "center"],
)

# ---------------------------------------------------------------- chapter 2
h1("第二章 角色与权限说明")

h2("2.1 角色总览")
add_table(
    ["角色", "登录凭证", "功能模块", "AI 智能体权限", "数据权限"],
    [
        ["学生", "学号", "学业总览、学业画像、预警通知、发展引导、AI 广场、快捷访问", "8 个学生端智能体", "仅本人"],
        ["班主任", "工号", "班级驾驶舱、GPA 进退、课程健康度、预警管理、学生详情、AI 广场", "共享 2 个智能体", "所负责班级"],
        ["任课教师", "工号", "我的课程、学情预警、AI 广场", "共享 2 个智能体", "所授课程"],
        ["系主任", "工号", "专业态势、课程分析、预警管理、AI 广场", "共享 2 个智能体", "所属专业"],
        ["院领导", "工号", "全院仪表盘、全院预警管理、AI 广场", "2 个专属智能体", "全院"],
    ],
    [1.7, 1.5, 5.9, 3.0, 3.7],
    aligns=["center", "center", "left", "center", "center"],
)

h2("2.2 学生")
para_parts([("功能权限：", True, INK),
            ("查看个人学业总览、五维学业画像、预警通知与改进建议；使用发展引导功能进行考研、留学、考公考编、就业等多路径规划；"
             "使用 AI 领航助手与 8 个角色化智能体进行多轮对话；通过快捷访问进入常用系统。", False, None)])
para_parts([("AI 智能体权限：", True, INK),
            ("AI 校策知询、AI 人生规划、AI 学业顾问、AI 就业助手、AI 考研助手、AI 留学助手、AI 考编助手、AI 考公助手。", False, None)])
para_parts([("数据权限：", True, INK),
            ("仅可访问本人学业数据，预警、画像与历史记录均按学号隔离。", False, None)])

h2("2.3 班主任")
para_parts([("功能权限：", True, INK),
            ("查看所负责班级的驾驶舱、GPA 进退情况、课程健康度、预警列表与学生详情；对预警进行 AI 根因分析并提交干预回填记录。", False, None)])
para_parts([("AI 智能体权限：", True, INK),
            ("学情分析、育心导学（与任课教师端、系主任端共享）。", False, None)])
para_parts([("数据权限：", True, INK),
            ("可访问所负责班级全部学生数据。", False, None)])

h2("2.4 任课教师")
para_parts([("功能权限：", True, INK),
            ("查看所授课程列表、成绩分布、班级对比、挂科名单与学情预警，并生成干预建议。", False, None)])
para_parts([("AI 智能体权限：", True, INK),
            ("学情分析、育心导学（与班主任端、系主任端共享）。", False, None)])
para_parts([("数据权限：", True, INK),
            ("可访问所授课程班级学生的成绩与预警数据。", False, None)])

h2("2.5 系主任（专业负责人）")
para_parts([("功能权限：", True, INK),
            ("查看专业态势、课程通过率热力图、课程健康度明细、专业预警列表，支持按年级与等级筛选。", False, None)])
para_parts([("AI 智能体权限：", True, INK),
            ("学情分析、育心导学（共享）。", False, None)])
para_parts([("数据权限：", True, INK),
            ("可访问所属专业的全部学生、课程与成绩数据。", False, None)])

h2("2.6 院领导")
para_parts([("功能权限：", True, INK),
            ("查看全院仪表盘、各专业通过率排名、GPA 对比与全院预警总览，支持按专业/年级分组和趋势分析。", False, None)])
para_parts([("AI 智能体权限：", True, INK),
            ("AI 课程健康度诊断、AI 跨专业对比分析（院领导专属）。", False, None)])
para_parts([("数据权限：", True, INK),
            ("可访问全院数据。", False, None)])

h2("2.7 数据权限矩阵")
add_table(
    ["角色", "可访问数据范围"],
    [
        ["学生", "仅自己的数据"],
        ["班主任", "所负责班级的学生数据"],
        ["任课教师", "所授课程班级的学生成绩数据"],
        ["系主任", "所属专业的所有数据"],
        ["院领导", "全院数据"],
    ],
    [3.2, 12.6],
    aligns=["center", "left"],
    caption_text="表 2-1  数据权限控制矩阵",
)

# ---------------------------------------------------------------- chapter 3
h1("第三章 学生端功能详细说明")

h2("3.1 AI 领航助手（智能体广场）")
para("功能描述：为学生提供统一的 AI 智能体入口，包含智能体广场、搜索/分类筛选、对话工作台与历史会话。")
para("核心功能：", bold=True, before=4)
bullet("智能体广场：按角色展示 8 个学生端智能体，支持名称/描述搜索与分类筛选。")
bullet("智能体卡片：名称、描述、图标、使用次数、数据来源说明与预设问题入口。")
bullet("对话工作台：多轮对话、流式输出、快捷问题填充、新会话与历史会话管理。")
bullet("外部平台接入：支持 Dify、E-Agent 等外部智能体平台嵌入与对话代理。")
para("学生端 8 个智能体：", bold=True, before=4)
numbered("AI 校策知询：奖助学金、助学贷款、创业补贴、就业优惠、落户政策等解读。")
numbered("AI 人生规划：长期职业发展与人生目标规划，考研/留学/就业决策支持。")
numbered("AI 学业顾问：7×24 全天候学业咨询，成绩查询、GPA 分析、选课建议。")
numbered("AI 就业助手：技能匹配分析、简历优化、面试准备、职业路径规划。")
numbered("AI 考研助手：院校匹配、竞争力评估、备考时间线、复习策略。")
numbered("AI 留学助手：海外院校推荐、申请材料优化、软背景提升建议。")
numbered("AI 考编助手：事业单位岗位匹配、职测辅导、综合应用能力训练。")
numbered("AI 考公助手：公务员考试规划、岗位智能匹配、行测申论辅导。")

h2("3.2 学业总览")
para("功能描述：实时呈现学生当前学期与历史学期的学业核心指标，是学生端默认数据驾驶舱。")
bullet("核心 KPI：累计 GPA、专业排名、已修学分/总学分、健康分、预警数量。")
bullet("GPA 趋势图：折线图对比本人 GPA 与年级均值，支持学期切换。")
bullet("AI 周报：自动生成学业状态摘要、关注点与行动建议。")
bullet("近期通知：展示最近预警/通知，无预警时显示正常状态。")
bullet("课程概况：本学期在修课程数量、学分进度与课程分布。")

h2("3.3 学业画像")
para("功能描述：以五维雷达图与明细数据完整呈现学生学业与发展画像。")
bullet("五维画像：学业成绩、实践能力、综合素质、人文素养、身心健康，每维包含得分、专业均值、满分与描述。")
bullet("维度详情：每维度的关键指标标签（如 GPA、排名、项目经历、体测、心理测评）。")
bullet("趋势分析：GPA 历史趋势与专业排名趋势，支持学期切换。")
bullet("画像数据来源：成绩、考勤、体测、竞赛/项目、心理测评等多源数据聚合。")

h2("3.4 预警通知")
para("功能描述：集中展示系统触发的个人学业预警，支持查看详情、确认已读与改进建议。")
bullet("三级预警：黄色（GPA 2.3-2.5 或成绩边缘）、橙色（GPA 2.0-2.3 或多门挂科）、红色（GPA < 2.0 或严重出勤问题）。")
bullet("预警明细：类型、触发事件、推送时间、详细描述、改进建议。")
bullet("挂科课程明细：以标签形式展示挂科课程名称列表（failedCourses 数组）。")
bullet("状态同步：学生确认已读后，班主任端同步显示“学生已查看”。")
bullet("数据字段：id、studentId、studentName、type、level、title、description、failedCourses、date、status、suggestion、triggerEvent、pushedAt。")

h2("3.5 发展引导")
para("功能描述：基于学生画像与专业属性，提供多路径发展引导与个性化规划。")
bullet("路径选择：考研深造、出国留学、公务员考试、事业单位考编、直接就业五类路径。")
bullet("路径匹配度：显示每条路径的匹配分数与适合程度。")
bullet("差距分析：逐维度对比“当前水平”与“目标要求”，标注紧迫度。")
bullet("行动清单：按紧迫度排序的具体提升行动与截止时间。")
bullet("里程碑：关键节点时间线与完成状态（已完成/进行中/待开始）。")
bullet("岗位/院校推荐：就业路径展示职位推荐，考研/留学路径展示三梯度院校推荐，支持分类与梯度筛选。")

h2("3.6 快捷访问")
para("功能描述：提供常用系统与功能的快速入口，可按学校实际系统配置维护。")

# ---------------------------------------------------------------- chapter 4
h1("第四章 班主任端功能详细说明")

h2("4.1 班级驾驶舱")
para("功能描述：班主任查看所负责班级的整体学业健康状态。")
bullet("班级 KPI：班级人数、平均 GPA、当前预警人数、课程通过率。")
bullet("成绩分布：优秀/良好/中等/预警分段的环形图。")
bullet("预警分布：黄、橙、红三级预警数量与人员列表。")
bullet("班级五维画像：班级维度平均得分的雷达图。")
bullet("最新预警列表：学生姓名、学号、等级、内容与日期，可跳转学生详情。")
bullet("AI 班级诊断：自动生成班级学业诊断摘要与关注点。")

h2("4.2 GPA 进退情况")
para("功能描述：按学期查看全班学生 GPA 明细，识别进步与退步显著的学生。")
bullet("学期 GPA 明细表：每名学生各学期 GPA 与变化幅度。")
bullet("进步/退步榜单：Top 5 进步与退步学生，支持点击查看个人趋势。")
bullet("预警阈值标记：GPA < 2.5 自动标红提示。")

h2("4.3 课程健康度")
para("功能描述：从课程维度分析班级学业表现。")
bullet("课程概览：平均分、通过率、挂科率与学分信息。")
bullet("挂科学生名单：按课程展开显示挂科学生姓名、学号与成绩。")
bullet("重点学生汇总：挂科 2 门及以上的学生聚合展示。")

h2("4.4 预警管理")
para("功能描述：统一的班级预警处理工作台。")
bullet("列表筛选：按预警等级、处理状态筛选，支持关键词搜索。")
bullet("预警详情：学生信息、预警内容、触发事件、挂科课程、处置建议。")
bullet("AI 根因分析：一键生成预警根因分析与面谈建议。")
bullet("干预回填：记录干预时间、方式（电话/面谈/消息/其他）、内容、学生反馈与后续计划。")
bullet("状态流转：待处理 → 处理中 → 已解决，处理完成后同步学生端。")

h2("4.5 学生详情")
para("功能描述：查看单个学生的完整学业画像。")
bullet("学生档案：姓名、学号、专业、班级、GPA、排名、预警等级。")
bullet("五维画像与 GPA 趋势：雷达图 + 学期 GPA 柱状/折线图。")
bullet("预警记录：该生全部预警及处理状态。")
bullet("AI 画像叙述：自动生成学生画像解读与谈话建议。")

h2("4.6 班主任端 AI 广场")
para("班主任端提供 2 个共享智能体：")
bullet("学情分析：班级/专业整体学情分析，识别高风险学生，生成学情报告。")
bullet("育心导学：心理健康评估、谈心话术建议、情绪异常预警。")

# ---------------------------------------------------------------- chapter 5
h1("第五章 任课教师端功能详细说明")

h2("5.1 我的课程")
para("功能描述：任课教师查看本学期所授课程的成绩与学情。")
bullet("课程列表：课程名称、课程代码、授课班级、学生人数、均分、通过率、高风险人数、健康分。")
bullet("成绩分布：按分数段展示学生人数柱状图。")
bullet("班级对比：各班级均分、通过率、最高/最低分、人数、风险人数与可视化对比条。")
bullet("挂科名单：按班级展示挂科学生姓名、学号与成绩。")
bullet("AI 学情诊断：自动生成课程学情摘要与重点关注建议。")

h2("5.2 学情预警")
para("功能描述：以课程维度展示需要关注的学生。")
bullet("预警学生列表：姓名、学号、班级、课程、均分、最近成绩、缺勤次数、趋势与原因。")
bullet("筛选统计：按预警等级、课程筛选，顶部展示红/橙/黄/全部数量卡片。")
bullet("一键生成干预建议：跳转对应 AI 智能体生成谈话/辅导建议。")

h2("5.3 任课教师端 AI 广场")
para("任课教师端共享班主任端 2 个智能体：学情分析、育心导学。")

# ---------------------------------------------------------------- chapter 6
h1("第六章 系主任端功能详细说明")

h2("6.1 专业态势")
para("功能描述：系主任查看所属专业的整体学业态势。")
bullet("专业 KPI：专业总人数、平均 GPA、预警比例、课程达标率。")
bullet("各年级 GPA 对比：按年级与学期展示 GPA 对比柱状图。")
bullet("预警分布：黄、橙、红三级预警数量与占比。")
bullet("重点关注学生：专业范围内预警学生列表，支持按等级筛选。")

h2("6.2 课程分析")
para("功能描述：从课程维度评估专业教学质量与课程健康度。")
bullet("课程通过率热力图：课程 × 年级的通过率热力图，支持点击查看详情。")
bullet("课程明细：每门课程在各年级的均分、通过率、高分率（≥90）、不及格率。")
bullet("AI 风险标注：自动识别“评分偏松”“通过率偏低”等风险课程。")
bullet("重点关注提示：通过率 < 75% 或不及格率 > 25% 的课程自动提示。")

h2("6.3 预警管理")
para("功能描述：专业范围内的学生预警统一管理。")
bullet("预警列表：专业学生预警记录，支持按年级、等级、状态筛选。")
bullet("视图切换：按预警等级、按年级两种视图切换。")
bullet("学生信息：预警学生姓名、学号、年级、GPA、预警等级与专业信息。")

h2("6.4 系主任端 AI 广场")
para("系主任端共享 2 个智能体：学情分析、育心导学。")

# ---------------------------------------------------------------- chapter 7
h1("第七章 院领导端功能详细说明")

h2("7.1 全院仪表盘")
para("功能描述：院领导查看全院学业质量的核心仪表盘。")
bullet("全院 KPI：全院在校生、平均 GPA、综合预警率、课程达标率。")
bullet("专业通过率排名：横向条形图按课程通过率排名，颜色分级。")
bullet("专业 GPA 对比：柱状图 + 院均 GPA 参考线。")
bullet("专业预警分布：各专业黄、橙、红预警数量明细表。")

h2("7.2 全院预警管理")
para("功能描述：从院级视角查看各专业、各年级的预警分布与趋势。")
bullet("KPI 卡片：红色、橙色、黄色、合计四类预警总数。")
bullet("专业预警率对比：横向条形图展示各专业预警率。")
bullet("近四学期趋势：红/橙/黄预警数量的堆叠折线趋势。")
bullet("分组明细：按专业或按年级切换查看预警明细与风险等级。")

h2("7.3 院领导端 AI 广场")
para("院领导端提供 2 个专属智能体：")
bullet("AI 课程健康度诊断：全院课程质量评估，识别系统性问题课程，输出改进建议。")
bullet("AI 跨专业对比分析：各专业横向对比，识别异常低位专业，辅助资源配置决策。")

# ---------------------------------------------------------------- chapter 8
h1("第八章 AI 智能体清单汇总")

h2("8.1 智能体总览")
add_table(
    ["智能体", "学生", "班主任", "任课教师", "系主任", "院领导", "核心能力"],
    [
        ["AI 校策知询", "√", "—", "—", "—", "—", "奖助、贷款、创业、落户等政策解读"],
        ["AI 人生规划", "√", "—", "—", "—", "—", "职业发展与人生目标规划"],
        ["AI 学业顾问", "√", "—", "—", "—", "—", "7×24 学业咨询、GPA 与选课建议"],
        ["AI 就业助手", "√", "—", "—", "—", "—", "技能匹配、简历优化、面试准备"],
        ["AI 考研助手", "√", "—", "—", "—", "—", "院校匹配、竞争力评估、备考规划"],
        ["AI 留学助手", "√", "—", "—", "—", "—", "海外院校推荐、申请材料优化"],
        ["AI 考编助手", "√", "—", "—", "—", "—", "事业单位岗位匹配、职测辅导"],
        ["AI 考公助手", "√", "—", "—", "—", "—", "公务员考试规划、岗位匹配、行测申论"],
        ["学情分析", "—", "√", "√", "√", "—", "班级/专业学情分析与高风险识别"],
        ["育心导学", "—", "√", "√", "√", "—", "心理评估、谈心话术、情绪预警"],
        ["AI 课程健康度诊断", "—", "—", "—", "—", "√", "课程质量评估与改进建议"],
        ["AI 跨专业对比分析", "—", "—", "—", "—", "√", "专业横向对比与决策支持"],
    ],
    [2.8, 1.15, 1.15, 1.15, 1.15, 1.15, 7.25],
    aligns=["left"] + ["center"] * 5 + ["left"],
    font_size=10,
    caption_text="表 8-1  AI 智能体角色权限与核心能力总览",
)
add_note("系统共 12 个独立智能体（部分跨角色共享），智能体配置支持按角色动态扩展，"
         "可通过管理端新增、启停与调整提示词、模型类型和适用角色。")

h2("8.2 智能体对话交互参数")
bullet("多轮对话：携带最近 20 条会话上下文，支持多轮追问。")
bullet("流式输出：SSE 协议实时返回增量内容，首字延迟 < 3 秒。")
bullet("历史持久化：会话与消息落库，刷新后历史可恢复，支持新会话。")
bullet("预设问题：每个智能体配置 4 条快捷问题，点击自动填充发送。")
bullet("卡牌数据：对话可返回成绩、雷达、预警、趋势等结构化卡片数据。")
bullet("模型可配置：通义千问、文心一言、GPT-4 等模型按智能体配置切换。")
bullet("使用统计：记录使用次数与 token 消耗，支持运营分析。")

# ---------------------------------------------------------------- chapter 9
h1("第九章 后端架构与技术参数")

h2("9.1 系统架构")
para("后端采用 Spring Boot 3.x 分层架构（可按业务域演进为微服务），核心分层如下：")
bullet("Controller 层：参数校验、权限验证、统一响应返回。")
bullet("Service 层：学业数据计算、预警规则引擎、AI 对话管理、统计聚合。")
bullet("Mapper 层：MyBatis-Plus 数据访问，支持代码生成与数据权限拦截。")
bullet("数据源适配层：AcademicDataSource 接口统一封装教务、学工、体测、就业等外部系统调用。")
bullet("基础设施层：MySQL 8.0、Redis 7.x、RabbitMQ、XXL-Job、Prometheus/Grafana 监控。")

h2("9.2 后端技术栈参数")
add_table(
    ["类别", "技术选型", "用途说明"],
    [
        ["后端框架", "Spring Boot 3.x（Java 21）", "应用服务、依赖管理、自动配置"],
        ["ORM", "MyBatis-Plus 3.5.x", "CRUD 简化、代码生成、数据权限"],
        ["数据库", "MySQL 8.0 + Flyway", "主库存储与数据库版本管理"],
        ["缓存", "Redis 7.x", "热点数据缓存、JWT 黑名单、限流"],
        ["认证鉴权", "Spring Security + JWT", "无状态认证与角色鉴权"],
        ["API 文档", "Knife4j / SpringDoc Swagger", "接口文档与在线调试"],
        ["消息队列", "RabbitMQ", "预警通知、异步任务、数据同步"],
        ["任务调度", "XXL-Job", "成绩同步、预警扫描、报表生成、缓存预热"],
        ["监控", "Actuator + Prometheus + Grafana", "健康检查、指标采集与可视化"],
        ["日志", "SLF4J + Logback", "滚动日志、按日归档、保留 30 天"],
    ],
    [2.5, 5.2, 8.1],
    aligns=["center", "left", "left"],
    font_size=10,
    caption_text="表 9-1  后端技术栈与参数",
)

h2("9.3 数据库设计参数")
add_table(
    ["数据表", "用途", "关键字段"],
    [
        ["user / student / teacher", "用户、学生、教师档案", "account、role、student_id、teacher_id、class_id、major_id"],
        ["class / major / department", "班级、专业、学院组织", "name、grade、major_id、advisor_id、dean_id"],
        ["course / course_class / grade", "课程、开课班级、成绩", "code、credits、term、teacher_id、score、grade_point、status"],
        ["alert / intervention_record", "预警与干预记录", "student_id、level、type、reason、failed_courses、status、methods、content"],
        ["gpa_history / profile_score", "GPA 历史与画像", "student_id、term、gpa、rank、dimension_key、score、avg_score"],
        ["ai_agent / ai_conversation / ai_message", "AI 智能体与会话", "code、roles、system_prompt、conversation_id、role、content、tokens"],
        ["job_cache / school_cache", "职位与院校缓存", "platform、match_score、tier、admission_gpa、source_url"],
    ],
    [3.7, 3.9, 8.2],
    aligns=["left", "left", "left"],
    font_size=10,
    caption_text="表 9-2  核心数据表设计参数",
)
para("数据字典：", bold=True, before=6)
bullet("预警等级：yellow（GPA 2.3-2.5 或成绩边缘波动）；orange（GPA 2.0-2.3 或多门挂科风险）；red（GPA < 2.0 或严重出勤问题）。")
bullet("成绩状态：pending（待出分）、passed（通过）、failed（挂科）、retake（重修）。")
bullet("课程类型：required（必修）、elective（选修）、public（公共课）。")

h2("9.4 API 接口规范")
para("接口采用 RESTful 风格，统一响应格式如下：")
add_code([
    "{",
    '  "code": 200,',
    '  "message": "success",',
    '  "data": {},',
    '  "timestamp": 1750000000000',
    "}",
])
bullet("认证方式：请求头携带 Authorization: Bearer <JWT_TOKEN>。")
bullet("错误码：200 成功；400 参数错误；401 未授权；403 无权限；404 资源不存在；500 服务内部错误。")
bullet("分页格式：PageResult 返回 list、total、page、size。")
bullet("接口文档：Swagger UI 在线调试，接口按模块分组。")

add_table(
    ["模块", "核心接口", "说明"],
    [
        ["认证", "POST /api/auth/login；GET /api/auth/me；POST /api/auth/logout", "登录返回 token + 用户信息，支持当前用户查询与登出"],
        ["学生端", "GET /api/student/dashboard；/api/student/profile；/api/student/alerts；/api/student/gpa-history", "学业总览、画像、预警与 GPA 历史"],
        ["班主任端", "GET /api/teacher/dashboard；/api/teacher/gpa-progress；/api/teacher/course-stats", "班级驾驶舱、GPA 进退与课程统计"],
        ["任课教师端", "GET /api/course-teacher/courses；/api/course-teacher/alerts", "我的课程与课程预警"],
        ["系主任端", "GET /api/department/overview；/api/department/courses；/api/department/alerts", "专业态势、课程分析与专业预警"],
        ["院领导端", "GET /api/dean/dashboard；/api/dean/alerts", "全院仪表盘与预警总览"],
        ["预警", "GET /api/alerts；POST /api/alerts/{id}/confirm；POST /api/alerts/{id}/intervention", "预警列表、确认已读、干预回填"],
        ["AI", "GET /api/agents；POST /api/agents/{id}/chat（SSE）；GET /api/agents/{id}/history", "智能体列表、流式对话与会话历史"],
    ],
    [1.7, 7.3, 6.8],
    aligns=["center", "left", "left"],
    font_size=10,
    caption_text="表 9-3  核心接口清单",
)

h2("9.5 外部数据对接与数据形式")
add_table(
    ["对接系统", "数据内容", "关键字段", "同步方式"],
    [
        ["教务系统", "成绩、GPA、课程、排名", "courseCode、courseName、credit、score、gradePoint、term", "定时增量同步 + 事件触发"],
        ["学工系统", "考勤、请假、辅导员", "date、courseName、status（present/late/absent）", "每日增量同步"],
        ["体测系统", "体测报告", "term、totalScore、items[]", "学期同步"],
        ["就业平台", "招聘职位缓存", "platform、jobTitle、salaryRange、matchScore", "XXL-Job 每日刷新"],
        ["统一认证", "身份与角色", "uid、工号/学号、角色属性", "登录时实时校验"],
    ],
    [2.4, 3.3, 5.5, 4.6],
    aligns=["center", "left", "left", "center"],
    font_size=10,
    caption_text="表 9-4  外部数据对接参数",
)
para("外部数据以 JSON 形式返回，适配层统一映射为内部 DTO 后落库。成绩数据示例：", before=4)
add_code([
    "{",
    '  "studentId": "2022001",',
    '  "term": "2024-2025-1",',
    '  "grades": [',
    "    {",
    '      "courseCode": "CS2021",',
    '      "courseName": "数据结构",',
    '      "credit": 4.0,',
    '      "score": 86.5,',
    '      "gradePoint": 3.7,',
    '      "status": "passed",',
    '      "term": "2024-2025-1",',
    '      "examDate": "2025-01-15"',
    "    }",
    "  ]",
    "}",
])
para("预警数据中的挂科课程列表（failedCourses）以 JSON 数组存储并原样返回，非挂科类预警返回空数组。")

h2("9.6 AI 模型接入参数")
bullet("模型支持：通义千问（qwen-max）、文心一言、GPT-4，可按智能体配置 model_type。")
bullet("流式协议：SSE（text/event-stream），增量消息格式 data: {\"event\":\"message\",\"content\":\"...\"}，结束标记 data: [DONE]。")
bullet("上下文构建：系统提示词（agent.system_prompt） + 最近 20 条对话历史 + 当前用户消息。")
bullet("安全控制：API Key 服务端保管、按用户/接口限流、token 消耗统计。")
bullet("外部平台：支持 Dify（v1/chat-messages）、E-Agent（assistant/chat/completions）代理接入。")

h2("9.7 缓存与性能指标")
add_table(
    ["数据类型", "缓存键", "TTL"],
    [
        ["用户信息", "user:{userId}", "30 分钟"],
        ["学生 GPA", "student:gpa:{studentId}", "1 小时"],
        ["班级统计", "class:stats:{classId}:{term}", "2 小时"],
        ["课程统计", "course:stats:{courseId}:{term}", "2 小时"],
        ["专业排名", "department:ranking:{term}", "4 小时"],
        ["AI 对话历史", "ai:conversation:{conversationId}", "24 小时"],
    ],
    [3.4, 8.0, 4.4],
    aligns=["left", "left", "center"],
    font_size=10,
    caption_text="表 9-5  Redis 缓存设计",
)
add_table(
    ["性能指标", "技术参数"],
    [
        ["并发用户数", "支持 10000+ 在线用户同时访问"],
        ["页面响应时间", "常规页面加载 < 2 秒，复杂报表 < 5 秒"],
        ["AI 响应时间", "流式输出首字延迟 < 3 秒，完整回答 < 30 秒"],
        ["数据库容量", "支持 50000+ 学生数据存储与查询"],
        ["系统可用性", "99.5% 以上（年停机时间 < 44 小时）"],
        ["数据备份", "每日全量备份 + 实时增量备份，保留 30 天"],
    ],
    [3.4, 12.4],
    aligns=["center", "left"],
    font_size=10,
    caption_text="表 9-6  性能指标",
)

h2("9.8 安全设计")
bullet("传输安全：全站 HTTPS 加密传输。")
bullet("身份认证：支持学校统一身份认证（CAS/OAuth2）对接，系统内部使用 JWT 无状态令牌。")
bullet("权限控制：基于 RBAC 的细粒度权限管理 + 数据权限拦截器，保证角色数据隔离。")
bullet("敏感数据：密码 BCrypt 加密，证件号等敏感字段加密存储，接口不返回明文敏感信息。")
bullet("操作审计：登录、预警处理、干预回填、智能体调用等关键操作完整记录，可追溯。")
bullet("接口防护：参数校验、SQL 注入防护、XSS 过滤、CSRF 防护、Redis 限流。")
bullet("合规要求：支持等保三级测评（按学校要求配合提供）。")

h2("9.9 部署方案")
add_table(
    ["组件", "技术参数", "说明"],
    [
        ["MySQL", "mysql:8.0", "主数据库，数据卷持久化"],
        ["Redis", "redis:7-alpine", "缓存、黑名单与限流"],
        ["RabbitMQ", "rabbitmq:3-management-alpine", "异步通知与任务"],
        ["后端服务", "Spring Boot 3.x 容器镜像", "Java 21，端口 8080"],
        ["Nginx", "反向代理 + 静态资源", "HTTPS、SSE 关闭缓冲、限流"],
    ],
    [2.5, 5.4, 7.9],
    aligns=["center", "left", "left"],
    font_size=10,
    caption_text="表 9-7  容器化部署组件",
)
bullet("云端部署：阿里云/腾讯云/华为云，Docker + Kubernetes 容器化部署。")
bullet("本地部署：学校私有服务器，支持离线运行（AI 功能需外网或本地模型服务）。")
bullet("SSE 支持：Nginx 配置 proxy_http_version 1.1、Connection 空、proxy_buffering off。")

# ---------------------------------------------------------------- chapter 10
h1("第十章 验收标准")

h2("10.1 功能验收")
add_table(
    ["验收项", "验收标准", "验收方法"],
    [
        ["登录与身份识别", "学号登录进入学生端；工号登录自动识别班主任/任课教师/系主任/院领导角色并跳转对应界面", "实际登录测试"],
        ["学生端功能", "学业总览、学业画像、预警通知、发展引导、AI 广场、快捷访问功能完整，学期切换生效", "功能逐一测试"],
        ["班主任端功能", "班级驾驶舱、GPA 进退、课程健康度、预警管理、干预记录、学生详情功能完整", "功能逐一测试"],
        ["任课教师端功能", "我的课程、成绩分布、班级对比、学情预警功能完整", "功能逐一测试"],
        ["系主任端功能", "专业态势、课程分析热力图、专业预警管理功能完整", "功能逐一测试"],
        ["院领导端功能", "全院仪表盘、全院预警管理功能完整", "功能逐一测试"],
        ["AI 智能体", "12 个智能体全部可正常对话，流式输出、历史恢复与预设问题可用，回答准确率 ≥ 85%", "对话测试 + 专家评估"],
        ["数据对接", "教务/学工/体测等外部数据同步正常，字段映射与前端展示一致", "联调测试 + 数据核对"],
    ],
    [2.7, 8.0, 5.1],
    aligns=["center", "left", "center"],
    font_size=10,
    caption_text="表 10-1  功能验收标准",
)

h2("10.2 性能验收")
bullet("并发测试：1000 用户同时在线，系统响应正常，无崩溃。")
bullet("响应时间：90% 的请求响应时间符合 9.7 章节性能指标。")
bullet("稳定性测试：连续运行 7 天，无重大故障。")

h2("10.3 安全验收")
bullet("通过等保三级测评（如学校要求）。")
bullet("通过渗透测试，无高危漏洞。")
bullet("数据权限隔离验证：学生只能查看本人数据，教师只能查看所负责班级/课程数据，系主任仅专业数据，院领导可查看全院。")

h2("10.4 文档交付")
bullet("系统部署手册")
bullet("用户操作手册（学生端、班主任端、任课教师端、系主任端、院领导端）")
bullet("系统维护手册")
bullet("数据库设计文档")
bullet("API 接口文档（Swagger/Knife4j）")

h2("10.5 培训与支持")
bullet("提供不少于 2 次系统使用培训（管理员培训、教师培训）。")
bullet("提供 1 年免费运维支持。")
bullet("提供 7×12 小时技术支持热线。")
bullet("重大故障 4 小时内响应，24 小时内解决。")

# ---------------------------------------------------------------- chapter 11
h1("第十一章 项目实施与交付")

h2("11.1 实施阶段")
add_table(
    ["阶段", "主要工作", "交付成果"],
    [
        ["M1 基础平台", "认证、统一响应、角色鉴权、数据库初始化", "可登录、接口文档、初始化数据"],
        ["M2 预警与学生/教师核心", "预警规则与 API、学生端、班主任端核心页面", "学生端、班主任端功能联调通过"],
        ["M3 任课教师/系主任/院长", "课程预警、专业态势、全院仪表盘与预警总览", "五端核心功能联调通过"],
        ["M4 AI 智能体", "智能体配置、SSE 对话、会话持久化、外部模型接入", "12 个智能体可对话"],
        ["M5 实时推送与上线", "WebSocket 预警推送、数据同步、性能安全加固、部署上线", "生产环境交付与验收"],
    ],
    [2.8, 7.0, 6.0],
    aligns=["center", "left", "left"],
    font_size=10,
    caption_text="表 11-1  实施阶段与交付",
)

h2("11.2 服务保障")
para("交付后提供 1 年免费运维支持，包含系统监控、故障处理、数据备份检查与安全补丁更新；"
     "根据学校需要提供二次开发支持与功能扩展服务。")

# ---------------------------------------------------------------- appendix
h1("附录A 默认测试账号")
add_table(
    ["角色", "账号", "密码", "默认落地页"],
    [
        ["学生", "student", "123456", "AI 领航助手"],
        ["班主任", "teacher", "123456", "班级驾驶舱"],
        ["任课教师", "course_teacher", "123456", "我的课程"],
        ["系主任", "department", "123456", "专业态势"],
        ["院领导", "dean", "123456", "全院仪表盘"],
    ],
    [3.0, 4.2, 3.0, 5.6],
    aligns=["center", "center", "center", "center"],
    caption_text="表 A-1  演示环境默认账号",
)
add_note("本表仅用于演示与验收环境；生产环境启用学校统一身份认证后，测试账号将失效。")

para("— 文档结束 —", align=WD_ALIGN_PARAGRAPH.CENTER, before=18, after=0, color=GRAY)

OUT = "学业领航智能体系统-详细功能及技术参数（招标版）.docx"
doc.save(OUT)
print("SAVED", OUT)
