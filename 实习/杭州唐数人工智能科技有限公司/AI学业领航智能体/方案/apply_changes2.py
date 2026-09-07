import re

with open('D:/Academic Navigation/方案/unpacked_proj/word/document.xml', encoding='utf-8') as f:
    content = f.read()

DATE = "2026-06-22T00:00:00Z"
AUTHOR = "Claude"

# Current max id is 20 from previous run
next_id = [21]

def get_next_id():
    i = next_id[0]
    next_id[0] += 1
    return i

def extract_rpr(run_xml):
    m = re.search(r'<w:rPr>.*?</w:rPr>', run_xml, re.DOTALL)
    return m.group(0) if m else ''

def space_attr(text):
    if text.startswith(' ') or text.endswith(' '):
        return ' xml:space="preserve"'
    return ''

def apply_change(content, anchor_text, old_part, new_part):
    idx = content.find(anchor_text)
    if idx < 0:
        print(f"  [SKIP] Not found: {anchor_text[:60]}")
        return content

    run_start = max(
        content.rfind('<w:r>', 0, idx),
        content.rfind('<w:r ', 0, idx)
    )
    run_end = content.find('</w:r>', idx) + len('</w:r>')
    full_run = content[run_start:run_end]

    rpr = extract_rpr(full_run)

    t_match = re.search(r'<w:t[^>]*>(.*?)</w:t>', full_run, re.DOTALL)
    if not t_match:
        print(f"  [SKIP] No <w:t> in run")
        return content
    actual_text = t_match.group(1)

    if old_part not in actual_text:
        print(f"  [SKIP] old_part not in text.")
        print(f"    actual='{actual_text[:100]}'")
        print(f"    old='{old_part[:60]}'")
        return content

    old_idx = actual_text.find(old_part)
    prefix = actual_text[:old_idx]
    suffix = actual_text[old_idx + len(old_part):]

    del_id = get_next_id()
    ins_id = get_next_id()

    parts = []
    if prefix:
        parts.append(f'<w:r>{rpr}<w:t{space_attr(prefix)}>{prefix}</w:t></w:r>')
    parts.append(
        f'<w:del w:id="{del_id}" w:author="{AUTHOR}" w:date="{DATE}">'
        f'<w:r>{rpr}<w:delText{space_attr(old_part)}>{old_part}</w:delText></w:r>'
        f'</w:del>'
    )
    parts.append(
        f'<w:ins w:id="{ins_id}" w:author="{AUTHOR}" w:date="{DATE}">'
        f'<w:r>{rpr}<w:t{space_attr(new_part)}>{new_part}</w:t></w:r>'
        f'</w:ins>'
    )
    if suffix:
        parts.append(f'<w:r>{rpr}<w:t{space_attr(suffix)}>{suffix}</w:t></w:r>')

    replacement = ''.join(parts)
    new_content = content[:run_start] + replacement + content[run_end:]
    print(f"  [OK] '{old_part[:55]}' => '{new_part[:55]}'")
    return new_content

# Change 9: RAG架构 → Milvus + BGE-M3
content = apply_change(
    content,
    'RAG架构，向量化存储学业相关知识文档',
    'RAG架构，向量化存储学业相关知识文档',
    'RAG 架构（向量数据库：Milvus 2.4 Standalone；Embedding 模型：BGE-M3，BAAI 出品，中文召回率最优），向量化存储学业相关知识文档'
)

# Change 10: 智能体管理后台 → Dify
# Only change the first run (keep "调试、运行效果监控" runs as-is)
content = apply_change(
    content,
    '提供智能体管理后台，支持 Prompt 版本管理、',
    '提供智能体管理后台，支持 Prompt 版本管理、',
    '采用 Dify 开源版（v0.10.x）作为智能体编排与管理平台，支持 Prompt 版本管理、'
)

with open('D:/Academic Navigation/方案/unpacked_proj/word/document.xml', 'w', encoding='utf-8') as f:
    f.write(content)

print(f'\nDone. Total IDs used now: {next_id[0]-1}')
