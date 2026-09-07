import re

with open('D:/Academic Navigation/方案/unpacked_proj/word/document.xml', encoding='utf-8') as f:
    content = f.read()

DATE = "2026-06-22T00:00:00Z"
AUTHOR = "Claude"
next_id = [1]

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
        print(f"  [SKIP] No <w:t> in run for: {anchor_text[:60]}")
        return content
    actual_text = t_match.group(1)

    if old_part not in actual_text:
        print(f"  [SKIP] old_part not in text. actual='{actual_text[:80]}', old='{old_part[:60]}'")
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

# 1. 前端框架
content = apply_change(
    content,
    'Vue 3 / React 18',
    'Vue 3 / React 18 及以上主流前端框架',
    'Vue 3.4 + TypeScript + Vite 5，配套 Element Plus 2.14+ 组件库、ECharts 5.x 及 Pinia 2.x 状态管理'
)

# 2. 后端框架
content = apply_change(
    content,
    'RESTful API 架构，支持接口版本管理',
    'RESTful API 架构',
    'Spring Boot 3.3（Java 17）+ Spring Security + MyBatis-Plus 3.5 构建的 RESTful API 服务；AI 服务层独立为 Python FastAPI 微服务（内网调用）'
)

# 3a. 数据库：去掉 PostgreSQL 选项
content = apply_change(
    content,
    'MySQL 8.0+ / PostgreSQL 14+',
    'MySQL 8.0+ / PostgreSQL 14+',
    'MySQL 8.0'
)

# 3b. Redis 末尾添加 Doris 说明
content = apply_change(
    content,
    'Redis 7.0+）。',
    'Redis 7.0+）。',
    'Redis 7.0）；分析层采用 Apache Doris 2.x OLAP 引擎，支撑数仓 ODS→DWD→DWS→ADS 四层查询。'
)

# 4. 部署方式
content = apply_change(
    content,
    'Docker Compose / Kubernetes',
    'Docker Compose / Kubernetes',
    'Docker Compose'
)

# 5. ETL 调度平台明确为 DolphinScheduler
content = apply_change(
    content,
    '调度平台支持 DAG 依赖管理',
    '调度平台支持',
    '采用 Apache DolphinScheduler 3.x 作为 ETL 调度引擎，支持'
)

# 6. ETL 采集工具明确为 DataX + Canal
content = apply_change(
    content,
    'ETL 作业实现日常自动调度，执行策略为每日 1 次全量核对',
    'ETL 作业实现日常自动调度，执行策略为每日 1 次全量核对 + 准实时增量同步。',
    'ETL 作业由 DataX（批量全量采集）与 Canal（MySQL Binlog CDC 准实时增量）双引擎驱动，实现每日 1 次全量核对 + 准实时增量同步。'
)

# 7. 上下文窗口：8000 → 64K
content = apply_change(
    content,
    '8000 tokens',
    '8000 tokens',
    '64K tokens（主力模型 DeepSeek-V3 原生支持 64K context，远超此最低要求）'
)

# 8. 大模型：明确 DeepSeek + Qwen
content = apply_change(
    content,
    '支持接入不少于 1 个经国家网信办备案的大语言模型，支持私有化部署或 API 调用两种接入模式。',
    '支持接入不少于 1 个经国家网信办备案的大语言模型，支持私有化部署或 API 调用两种接入模式。',
    '接入经国家网信办备案的 DeepSeek-V3（主力，API 调用模式）与 Qwen2.5-14B（离线备选，Ollama 私有化部署）；两种接入模式可按数据安全要求动态切换。'
)

# 9. 知识库框架：明确 Milvus + BGE-M3
content = apply_change(
    content,
    'RAG框构，向量化存储学业相关知识文档',
    'RAG框构，向量化存储学业相关知识文档',
    'RAG 框构（向量数据库：Milvus 2.4 Standalone；Embedding 模型：BGE-M3，BAAI 出品，中文召回率最优），向量化存储学业相关知识文档'
)

# 10. 智能体管理后台：明确 Dify
content = apply_change(
    content,
    '提供智能体管理后台，支持 Prompt 版本管理、调试、运行效果监控。',
    '提供智能体管理后台，支持 Prompt 版本管理、调试、运行效果监控。',
    '采用 Dify 开源版（v0.10.x）作为智能体编排与管理平台，原生提供 Prompt 版本管理、调试控制台、Workflow 多步工具调用编排及运行效果监控大盘。'
)

# 11. SSO 认证：明确 CAS
content = apply_change(
    content,
    '支持对接学校统一身份认证，实现单点登录。',
    '支持对接学校统一身份认证，实现单点登录。',
    '对接学校现有 CAS Server，采用 Spring Security CAS Client 实现单点登录；CAS ticket 验证成功后转换为系统内 JWT（jjwt 库），前端无感知携带。'
)

with open('D:/Academic Navigation/方案/unpacked_proj/word/document.xml', 'w', encoding='utf-8') as f:
    f.write(content)

print(f'\nAll done. Total tracked change IDs used: {next_id[0]-1}')
