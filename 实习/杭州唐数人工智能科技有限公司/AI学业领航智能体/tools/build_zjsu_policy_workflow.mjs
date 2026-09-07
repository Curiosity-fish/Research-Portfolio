import fs from 'node:fs';
import path from 'node:path';

const sourcePath = 'C:/Users/BenBen/Downloads/文档知识库问答-40159.json';
const outputPath = path.resolve('outputs/浙工商AI校策知询-稳定版.json');
const workflow = JSON.parse(fs.readFileSync(sourcePath, 'utf8'));

const nodeById = new Map(workflow.nodes.map((node) => [node.id, node]));
const start = nodeById.get('start_4a3f1');
const input = nodeById.get('input_1ca67');
const retriever = nodeById.get('knowledge_retriever_ea7b5');
const answer = nodeById.get('llm_9ac29');
const output = nodeById.get('output_5e46e');
const end = nodeById.get('end_daacc');

if (![start, input, retriever, answer, output, end].every(Boolean)) {
  throw new Error('Source workflow is missing one or more required nodes.');
}

workflow.name = '浙工商AI校策知询-稳定版';
workflow.description = '浙江工商大学全日制普通本科生学生手册政策问答。用户问题直连知识库检索，回答仅基于检索证据。';

const retrieverSettings = retriever.data.group_params[0].params;
const query = retrieverSettings.find((param) => param.key === 'user_question');
query.value = ['input_1ca67.user_input'];
query.varZh = { 'input_1ca67.user_input': '接收政策问题/user_input' };

const retrievalOptions = retrieverSettings.find((param) => param.key === 'advanced_retrieval_switch');
retrievalOptions.value = {
  ...retrievalOptions.value,
  keyword_weight: 0.5,
  vector_weight: 0.5,
  search_switch: true,
  rerank_flag: false,
  max_chunk_size: 15000,
};

const answerPrompt = answer.data.group_params.find((group) => group.name === '提示词').params;
const systemPrompt = answerPrompt.find((param) => param.key === 'system_prompt');
const userPrompt = answerPrompt.find((param) => param.key === 'user_prompt');

systemPrompt.value = `你是“浙江工商大学 AI 校策知询助手”。服务对象是全日制普通本科生。当前知识范围仅为《浙江工商大学学生手册（2025 汇编）》。

你只能依据输入中的“手册检索证据”作答，不能使用训练记忆补充校内规则，不能把用户描述当作已核实事实。回答的目标是帮助学生理解规定、判断自己还需确认什么，并给出下一步行动；你不能替学校作审批决定。

回答规则：
1. 证据直接覆盖问题时，先给简短结论，再说明适用条件、限制或后果和建议动作。关键结论必须在“依据”中标注制度名称、条款或检索片段中实际存在的页码。没有定位信息时，不要编页码。
2. 不要因为问题涉及年级、申请或个案就拒答。先根据证据说明通用规则；只有一个事实会改变规则分支且确实无法判断时，最后只追问一个问题。
3. 证据为空、不相关、问题超出本科学生手册范围、或用户要求确认本学期日期、金额、材料、名单、审批结果时，明确说明当前手册无法确认，并建议联系所在学院教务办、教务处、学生处或学生资助中心等对应部门。不得编造联系人、金额、日期或办理入口。
4. 手册是“2025 汇编”，其中制度可能后续修订。涉及办理时间、收费、资助金额、申请材料、名单和最终资格时，结尾必须写“请以学校及相关部门最新通知为准”。
5. 禁止“你一定可以”“肯定不能”“保证获批”等承诺性语言。处分、申诉、退学、资助等高影响事项保持中立和尊重。
6. 不索取或处理身份证号、银行卡、病历、家庭收入等敏感信息。

直接输出给学生看的 Markdown。不要输出 JSON、代码块、route、字段名或任何程序化结构。

证据充分时按此格式回答：
**结论**
一句准确、非承诺性的结论。

**适用条件或规则**
- 要点 1
- 要点 2

**你现在可以做什么**
1. 行动 1
2. 行动 2

**依据**
- 《制度名称》：第X条或检索片段中存在的页码

证据不足或超范围时，用两三句话说明原因和建议确认的部门类别，不要编造政策。`;

userPrompt.value = `用户问题：{{#input_1ca67.user_input#}}

仅可使用的手册检索证据：{{#knowledge_retriever_ea7b5.retrieved_output#}}

请直接输出给学生看的 Markdown 答复。`;
userPrompt.varZh = {
  'input_1ca67.user_input': '接收政策问题/user_input',
  'knowledge_retriever_ea7b5.retrieved_output': '文档知识库检索/retrieved_output',
};

const outputMessage = output.data.group_params[0].params.find((param) => param.key === 'message');
outputMessage.value = { msg: '{{#llm_9ac29.output#}}', files: [] };
outputMessage.varZh = { 'llm_9ac29.output': '证据约束政策答复/output' };

workflow.nodes = [start, input, retriever, answer, output, end];
workflow.edges = [
  {
    id: 'xy-edge__start_4a3f1right_handle-input_1ca67left_handle',
    type: 'customEdge',
    source: start.id,
    target: input.id,
    animated: true,
    sourceHandle: 'right_handle',
    targetHandle: 'left_handle',
  },
  {
    id: 'xy-edge__input_1ca67right_handle-knowledge_retriever_ea7b5left_handle',
    type: 'customEdge',
    source: input.id,
    target: retriever.id,
    animated: true,
    sourceHandle: 'right_handle',
    targetHandle: 'left_handle',
  },
  {
    id: 'xy-edge__knowledge_retriever_ea7b5right_handle-llm_9ac29left_handle',
    type: 'customEdge',
    source: retriever.id,
    target: answer.id,
    animated: true,
    sourceHandle: 'right_handle',
    targetHandle: 'left_handle',
  },
  {
    id: 'xy-edge__llm_9ac29right_handle-output_5e46eleft_handle',
    type: 'customEdge',
    source: answer.id,
    target: output.id,
    animated: true,
    sourceHandle: 'right_handle',
    targetHandle: 'left_handle',
  },
  {
    id: 'xy-edge__output_5e46eright_handle-end_daaccleft_handle',
    type: 'customEdge',
    source: output.id,
    target: end.id,
    animated: true,
    sourceHandle: 'right_handle',
    targetHandle: 'left_handle',
  },
];

fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, `${JSON.stringify(workflow, null, 2)}\n`, 'utf8');
console.log(outputPath);
