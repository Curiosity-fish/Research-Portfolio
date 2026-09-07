<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { agents, getMockResponse } from '@/mock/agents'
import { getStudentDashboard, getStudentProfile, getDevelopmentPaths, chatWithLifePlanningAgent, getLifePlanningMemories, deleteLifePlanningMemory } from '@/api'
import type { AgentMemoryItem } from '@/api'
import { Close, Delete, Document, Paperclip, Promotion } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const role = computed(() => auth.user?.role || 'student')
const agentId = computed(() => route.params.id as string)
const agent = computed(() => agents.find(a => a.id === agentId.value))

interface ChatMsg {
  id: string
  from: 'user' | 'ai'
  text: string
  time: string
  attachments?: ChatAttachment[]
}

interface ChatAttachment {
  name: string
  size: number
}

interface PendingAttachment extends ChatAttachment {
  id: string
  file: File
  path?: string
  uploading: boolean
  error?: string
}

const messages = ref<ChatMsg[]>([])
const inputText = ref('')
const isTyping = ref(false)
const attachmentInputRef = ref<HTMLInputElement>()
const pendingAttachments = ref<PendingAttachment[]>([])
const maxAttachmentSize = 20 * 1024 * 1024
const maxAttachmentCount = 5
const supportsFileUpload = computed(() => agent.value?.id === 'policy-advisor')
const hasAttachmentError = computed(() => pendingAttachments.value.some(item => item.error))
const hasAttachmentUpload = computed(() => pendingAttachments.value.some(item => item.uploading))
const canSend = computed(() => (
  !isTyping.value
  && !hasAttachmentUpload.value
  && !hasAttachmentError.value
  && Boolean(inputText.value.trim() || pendingAttachments.value.some(item => item.path))
))
const showTypingIndicator = computed(() => {
  const lastMessage = messages.value.at(-1)
  return isTyping.value && (!lastMessage || lastMessage.from !== 'ai' || !lastMessage.text)
})
const streamingMessageId = computed(() => {
  const lastMessage = messages.value.at(-1)
  return isTyping.value && lastMessage?.from === 'ai' && lastMessage.text ? lastMessage.id : ''
})
const chatBodyRef = ref<HTMLDivElement>()
const studentContextCache = ref('')
const eagentWorkflowSessionId = ref('')
const eagentWorkflowInputNodeId = ref('')
const eagentWorkflowMessageId = ref<number | null>(null)
const rememberCurrentMessage = ref(false)
const memoryDialogVisible = ref(false)
const memories = ref<AgentMemoryItem[]>([])
const memoriesLoading = ref(false)

// Reset chat when agent changes
watch(agentId, () => {
  messages.value = []
  inputText.value = ''
  resetAttachments()
  rememberCurrentMessage.value = false
  difyConversationId.value = ''
  difyInThinking.value = false
  eagentWorkflowSessionId.value = ''
  eagentWorkflowInputNodeId.value = ''
  eagentWorkflowMessageId.value = null
})

const recentAgents = computed(() =>
  agents
    .filter(a => a.roles.includes(role.value) && a.status === 'active' && a.id !== agentId.value)
    .slice(0, 4)
)

const mockHistory = [
  { id: 'h1', label: '本周学业数据分析', time: '今天' },
  { id: 'h2', label: '期末复习策略建议', time: '昨天' },
  { id: 'h3', label: '考研备考时间规划', time: '6月10日' },
]

function getTime() {
  const d = new Date()
  return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
}

function formatFileSize(size: number): string {
  if (size < 1024 * 1024) return `${Math.max(1, Math.round(size / 1024))} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

function resetAttachments() {
  pendingAttachments.value = []
  if (attachmentInputRef.value) attachmentInputRef.value.value = ''
}

function openAttachmentPicker() {
  if (!supportsFileUpload.value || isTyping.value) return
  attachmentInputRef.value?.click()
}

function removeAttachment(id: string) {
  pendingAttachments.value = pendingAttachments.value.filter(item => item.id !== id)
}

async function uploadAttachment(attachment: PendingAttachment) {
  const externalApi = agent.value?.externalApi
  if (!externalApi || externalApi.type !== 'eagent-workflow') {
    attachment.uploading = false
    attachment.error = '当前智能体未配置文件上传服务'
    return
  }

  try {
    const formData = new FormData()
    formData.append('file', attachment.file)
    const response = await fetch(`${externalApi.baseUrl}/api/v1/knowledge/upload`, {
      method: 'POST',
      headers: externalApi.apiKey ? { 'X-API-Key': externalApi.apiKey } : undefined,
      body: formData,
    })
    const result = await response.json().catch(() => null)
    const path = result?.data?.file_path || result?.file_path
    if (!response.ok || !path) {
      throw new Error(result?.message || `HTTP ${response.status}`)
    }

    const current = pendingAttachments.value.find(item => item.id === attachment.id)
    if (current) {
      current.path = path
      current.uploading = false
    }
  } catch (error: any) {
    const current = pendingAttachments.value.find(item => item.id === attachment.id)
    if (current) {
      current.uploading = false
      current.error = error?.message || '上传失败'
    }
  }
}

function handleAttachmentSelection(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (!files.length) return

  const availableSlots = Math.max(0, maxAttachmentCount - pendingAttachments.value.length)
  const acceptedFiles = files.slice(0, availableSlots)
  for (const file of acceptedFiles) {
    const attachment: PendingAttachment = {
      id: `${Date.now()}-${crypto.randomUUID()}`,
      file,
      name: file.name,
      size: file.size,
      uploading: file.size <= maxAttachmentSize,
      error: file.size > maxAttachmentSize ? '文件不能超过 20 MB' : undefined,
    }
    pendingAttachments.value.push(attachment)
    if (attachment.uploading) void uploadAttachment(attachment)
  }
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function renderInline(text: string): string {
  return escapeHtml(text)
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/(^|[^*])\*([^*\n]+)\*(?!\*)/g, '$1<em>$2</em>')
}

function formatText(text: string): string {
  const lines = text.split('\n')
  const out: string[] = []
  let inCode = false
  let codeBuffer: string[] = []
  let listType: 'ul' | 'ol' | null = null

  const flushList = () => {
    if (listType) {
      out.push(`</${listType}>`)
      listType = null
    }
  }

  const flushCode = () => {
    out.push(`<pre><code>${codeBuffer.join('\n')}</code></pre>`)
    codeBuffer = []
  }

  for (const rawLine of lines) {
    const line = rawLine.trimEnd()
    if (line.startsWith('```')) {
      if (inCode) {
        flushCode()
        inCode = false
      } else {
        flushList()
        inCode = true
      }
      continue
    }

    if (inCode) {
      codeBuffer.push(escapeHtml(rawLine))
      continue
    }

    if (!line.trim()) {
      flushList()
      continue
    }

    const heading = /^(#{1,6})\s+(.*)$/.exec(line)
    if (heading) {
      flushList()
      const level = heading[1].length
      out.push(`<h${level}>${renderInline(heading[2])}</h${level}>`)
      continue
    }

    if (/^(-{3,}|\*{3,})$/.test(line)) {
      flushList()
      out.push('<hr/>')
      continue
    }

    const quote = /^>\s?(.*)$/.exec(line)
    if (quote) {
      flushList()
      out.push(`<blockquote>${renderInline(quote[1])}</blockquote>`)
      continue
    }

    const ul = /^[-*+]\s+(.*)$/.exec(line)
    if (ul) {
      if (listType !== 'ul') {
        flushList()
        out.push('<ul>')
        listType = 'ul'
      }
      out.push(`<li>${renderInline(ul[1])}</li>`)
      continue
    }

    const ol = /^\d+\.\s+(.*)$/.exec(line)
    if (ol) {
      if (listType !== 'ol') {
        flushList()
        out.push('<ol>')
        listType = 'ol'
      }
      out.push(`<li>${renderInline(ol[1])}</li>`)
      continue
    }

    flushList()
    out.push(`<p>${renderInline(line)}</p>`)
  }

  if (inCode) flushCode()
  flushList()
  return out.join('')
}

function scrollToBottom() {
  nextTick(() => {
    if (chatBodyRef.value) {
      chatBodyRef.value.scrollTop = chatBodyRef.value.scrollHeight
    }
  })
}

function getOrCreateAiMessage(id: string): ChatMsg {
  const existing = messages.value.find(m => m.id === id)
  if (existing) return existing

  const aiMsg: ChatMsg = { id, from: 'ai', text: '', time: getTime() }
  messages.value.push(aiMsg)
  return aiMsg
}

async function animateAiText(aiMsgId: string, fullText: string): Promise<void> {
  const aiMsg = messages.value.find(m => m.id === aiMsgId)
  if (!aiMsg) return
  let index = 0
  const charsPerTick = Math.max(1, Math.ceil(fullText.length / 240))
  await new Promise<void>((resolve) => {
    const tick = () => {
      if (index >= fullText.length) {
        aiMsg.text = fullText
        scrollToBottom()
        resolve()
        return
      }
      index = Math.min(fullText.length, index + charsPerTick)
      aiMsg.text = fullText.slice(0, index)
      scrollToBottom()
      window.setTimeout(tick, 12)
    }
    tick()
  })
}

// E-Agent API 调用（无需鉴权，直接使用 model ID）

// Dify 多轮对话需要记录 conversation_id
const difyConversationId = ref<string>('')
const difyInThinking = ref(false)  // 标记当前是否在 <think> 块内

async function sendToDify(baseUrl: string, apiKey: string, message: string) {
  const aiMsgId = `a${Date.now()}`

  try {
    const res = await fetch(`${baseUrl}/v1/chat-messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${apiKey}`,
      },
      body: JSON.stringify({
        inputs: {},
        query: message,
        response_mode: 'streaming',
        conversation_id: difyConversationId.value || undefined,
        user: 'student',
      }),
    })

    if (!res.ok) {
      getOrCreateAiMessage(aiMsgId).text = `⚠️ Dify 返回错误：${res.status}`
      return
    }

    const reader = res.body?.getReader()
    if (!reader) {
      getOrCreateAiMessage(aiMsgId).text = '⚠️ 无法读取响应流'
      return
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() ?? ''
      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const chunk = line.slice(6).trim()
        try {
          const json = JSON.parse(chunk)
          // 保存 conversation_id 用于多轮对话
          if (json.conversation_id) difyConversationId.value = json.conversation_id
          if (json.event === 'message' && json.answer) {
            let chunk: string = json.answer
            // 进入思考块
            if (chunk.includes('<think>')) { difyInThinking.value = true; chunk = chunk.split('<think>')[0] }
            // 离开思考块
            if (difyInThinking.value && chunk.includes('</think>')) {
              difyInThinking.value = false
              chunk = chunk.split('</think>').slice(1).join('</think>')
            }
            if (!difyInThinking.value && chunk) {
              const aiMsg = getOrCreateAiMessage(aiMsgId)
              aiMsg.text += chunk
              scrollToBottom()
            }
          }
        } catch { /* 跳过非 JSON 行 */ }
      }
    }

    if (!messages.value.some(m => m.id === aiMsgId)) {
      getOrCreateAiMessage(aiMsgId).text = '⚠️ 未收到有效回复'
    }
  } catch (e: any) {
    getOrCreateAiMessage(aiMsgId).text = `⚠️ 连接失败：${e.message}`
  } finally {
    isTyping.value = false
    scrollToBottom()
  }
}
async function buildStudentContext(): Promise<string> {
  if (auth.user?.role !== 'student') return ''
  if (studentContextCache.value) return studentContextCache.value

  try {
    const [dashboard, profile, paths] = await Promise.all([
      getStudentDashboard(),
      getStudentProfile(),
      getDevelopmentPaths(),
    ])
    const dashboardData = dashboard as any
    const profileData = profile as any
    const pathData = paths as any[]
    const dimensionText = Array.isArray(profileData?.dimensions)
      ? profileData.dimensions
          .slice(0, 6)
          .map((item: any) => `${item.label || item.key || ''} ${item.score ?? item.avgScore ?? ''}`.trim())
          .filter(Boolean)
          .join('；')
      : ''
    const pathText = Array.isArray(pathData)
      ? pathData
          .slice(0, 5)
          .map((item: any) => `${item.label || item.key || ''} 匹配度 ${item.matchScore ?? ''}/100`.trim())
          .filter(Boolean)
          .join('；')
      : ''
    const alertText = Array.isArray(dashboardData?.recentAlerts)
      ? dashboardData.recentAlerts
          .slice(0, 3)
          .map((item: any) => `${item.level || ''} ${item.title || item.description || ''}`.trim())
          .filter(Boolean)
          .join('；')
      : ''

    const student = auth.user as any
    studentContextCache.value = [
      `姓名：${student?.name || '未知'}`,
      `账号：${student?.account || '未知'}`,
      `年级专业班级：${student?.grade || ''} ${student?.major || ''} ${student?.className || ''}`.trim(),
      `GPA：${dashboardData?.gpa ?? '未知'}，排名：${dashboardData?.rank ?? '未知'}/${dashboardData?.totalStudents ?? '未知'}`,
      `预警数：${dashboardData?.alertCount ?? 0}，健康分（0-100）：${dashboardData?.healthScore ?? '未知'}`,
      dimensionText ? `五维画像：${dimensionText}` : '',
      pathText ? `发展路径：${pathText}` : '',
      alertText ? `近期预警：${alertText}` : '近期预警：无',
    ]
      .filter(Boolean)
      .join('\n')
    return studentContextCache.value
  } catch {
    return ''
  }
}

async function sendToEagent(baseUrl: string, assistantId: string, context: string, apiKey?: string) {
  const aiMsgId = `a${Date.now()}`
  const historyMessages = messages.value
    .filter(m => m.text.trim())
    .map(m => ({
      role: m.from === 'user' ? 'user' : 'assistant',
      content: m.text,
    }))
  const apiMessages: Array<{ role: string; content: string }> = []
  if (context) {
    apiMessages.push({ role: 'system', content: context })
  }
  apiMessages.push(...historyMessages)

  try {
    const res = await fetch(`${baseUrl}/api/v2/assistant/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(apiKey ? { 'X-API-Key': apiKey } : {}),
      },
      body: JSON.stringify({
        model: assistantId,
        messages: apiMessages,
        temperature: 0,
        stream: true,
      }),
    })

    if (!res.ok) {
      getOrCreateAiMessage(aiMsgId).text = `⚠️ E-Agent 返回错误：${res.status}（${res.statusText}）`
      return
    }

    const reader = res.body?.getReader()
    if (!reader) {
      getOrCreateAiMessage(aiMsgId).text = '⚠️ 无法读取响应流'
      return
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() ?? ''
      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const chunk = line.slice(6).trim()
        if (chunk === '[DONE]') break
        try {
          const json = JSON.parse(chunk)
          const delta = json?.choices?.[0]?.delta?.content ?? ''
          if (delta) {
            const aiMsg = getOrCreateAiMessage(aiMsgId)
            aiMsg.text += delta
            scrollToBottom()
          }
        } catch { /* 跳过非 JSON 行 */ }
      }
    }

    if (!messages.value.some(m => m.id === aiMsgId)) {
      getOrCreateAiMessage(aiMsgId).text = '⚠️ 未收到有效回复'
    }
  } catch (e: any) {
    getOrCreateAiMessage(aiMsgId).text = `⚠️ 连接失败：${e.message}\n请检查 SSH 隧道是否已开启。`
  } finally {
    isTyping.value = false
    scrollToBottom()
  }
}

async function sendToManagedLifePlanning(message: string, remember: boolean) {
  const aiMsgId = `a${Date.now()}`
  const history = messages.value
    .slice(0, -1)
    .filter(item => item.text.trim())
    .slice(-16)
    .map(item => ({
      role: item.from === 'user' ? 'user' as const : 'assistant' as const,
      content: item.text,
    }))

  try {
    const result = await chatWithLifePlanningAgent({ message, remember, history })
    getOrCreateAiMessage(aiMsgId).text = result.answer || '暂未收到有效回复'
    if (result.memoryNotice) {
      result.remembered ? ElMessage.success(result.memoryNotice) : ElMessage.warning(result.memoryNotice)
    }
  } catch (error: any) {
    getOrCreateAiMessage(aiMsgId).text = `⚠️ 人生规划智能体暂时不可用：${error?.message || '请稍后重试'}`
  } finally {
    isTyping.value = false
    scrollToBottom()
  }
}

async function openMemoryDialog() {
  if (agentId.value !== 'life-planning') return
  memoryDialogVisible.value = true
  memoriesLoading.value = true
  try {
    memories.value = await getLifePlanningMemories()
  } catch {
    memories.value = []
  } finally {
    memoriesLoading.value = false
  }
}

async function removeMemory(memoryId: string) {
  try {
    await deleteLifePlanningMemory(memoryId)
    memories.value = memories.value.filter(item => item.id !== memoryId)
  } catch {
    // The shared request interceptor displays the server error.
  }
}

function extractWorkflowText(events: any[]): string {
  const parts: string[] = []
  for (const event of events) {
    if (!['output_msg', 'output_with_input_msg', 'output_with_choose_msg'].includes(event?.event)) {
      continue
    }
    const message = event?.output_schema?.message
    if (Array.isArray(message)) {
      parts.push(...message.filter((item): item is string => typeof item === 'string'))
    } else if (typeof message === 'string') {
      parts.push(message)
    }
  }
  return parts.filter(Boolean).join('\n')
}

function parseJsonValue(value: unknown): any {
  if (typeof value !== 'string') return value
  try {
    return JSON.parse(value)
  } catch {
    return value
  }
}

function resolveWorkflowEvent(sse: any): any {
  const data = parseJsonValue(sse?.data)
  const nestedData = parseJsonValue(data?.data)

  if (nestedData && typeof nestedData === 'object') return nestedData
  if (data && typeof data === 'object') return data
  return sse
}

function messageText(value: unknown): string {
  if (typeof value === 'string') return value
  if (Array.isArray(value)) return value.map(messageText).filter(Boolean).join('')
  if (!value || typeof value !== 'object') return ''

  const record = value as Record<string, unknown>
  return messageText(record.content ?? record.text ?? record.message)
}

function extractStreamText(sse: any, event: any): string {
  const candidates = [
    event?.choices?.[0]?.delta?.content,
    sse?.choices?.[0]?.delta?.content,
    event?.delta?.content,
    event?.output_schema?.message,
    event?.output?.message,
    event?.answer,
    event?.content,
  ]

  for (const candidate of candidates) {
    const text = messageText(candidate)
    if (text) return text
  }
  return ''
}

function findWorkflowInputNodeId(events: any[]): string {
  const inputEvent = events.find(event => event?.event === 'input')
  return inputEvent?.node_id || ''
}

function findWorkflowInputMessageId(events: any[]): number | null {
  const inputEvent = events.find(event => event?.event === 'input')
  return typeof inputEvent?.message_id === 'number' ? inputEvent.message_id : null
}

function resetWorkflowSessionIfClosed(events: any[]) {
  if (!events.some(event => event?.event === 'close')) return
  if (findWorkflowInputNodeId(events)) return
  eagentWorkflowSessionId.value = ''
  eagentWorkflowInputNodeId.value = ''
  eagentWorkflowMessageId.value = null
}

function buildWorkflowInput(
  message: string,
  context: string,
  inputNodeId: string,
  filePaths: string[] = [],
): Record<string, unknown> {
  const userInput = context
    ? `【学生客观数据，来自学校数据库，请优先使用；缺失时不要编造】\n${context}\n\n【学生本次输入】\n${message}`
    : message
  const value: Record<string, unknown> = {
    user_input: userInput,
    ...(filePaths.length ? { dialog_files_content: filePaths } : {}),
  }
  if (inputNodeId) {
    return { [inputNodeId]: value }
  }
  return value
}

async function invokeEagentWorkflow(
  baseUrl: string,
  apiKey: string,
  body: Record<string, unknown>,
  onEvent: (sse: any, event: any) => void,
) {
  const res = await fetch(`${baseUrl}/api/v2/workflow/invoke`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'text/event-stream',
      'Cache-Control': 'no-cache',
      'X-API-Key': apiKey,
    },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }

  const reader = res.body?.getReader()
  if (!reader) {
    throw new Error('Unable to read stream')
  }

  const decoder = new TextDecoder()
  let buffer = ''
  let dataLines: string[] = []

  const emitEvent = () => {
    if (!dataLines.length) return
    const payload = dataLines.join('\n').trim()
    dataLines = []
    if (!payload || payload === '[DONE]') return
    try {
      const sse = JSON.parse(payload)
      onEvent(sse, resolveWorkflowEvent(sse))
    } catch {
      // Ignore malformed SSE events without interrupting the remaining stream.
    }
  }

  const processLine = (rawLine: string) => {
    const line = rawLine.endsWith('\r') ? rawLine.slice(0, -1) : rawLine
    if (!line) {
      emitEvent()
      return
    }
    if (line.startsWith(':')) return
    if (line.startsWith('data:')) {
      // A few E-Agent deployments separate JSON events with a single newline.
      // Flush a complete pending JSON value before accepting the next data line.
      if (dataLines.length) {
        const pendingPayload = dataLines.join('\n').trim()
        if (pendingPayload === '[DONE]') {
          emitEvent()
        } else {
          try {
            JSON.parse(pendingPayload)
            emitEvent()
          } catch {
            // The current SSE event contains multiple data lines.
          }
        }
      }
      dataLines.push(line.slice(5).trimStart())
    }
  }

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    let newlineIndex = buffer.indexOf('\n')
    while (newlineIndex >= 0) {
      processLine(buffer.slice(0, newlineIndex))
      buffer = buffer.slice(newlineIndex + 1)
      newlineIndex = buffer.indexOf('\n')
    }
  }

  buffer += decoder.decode()
  if (buffer) processLine(buffer)
  emitEvent()
}

async function sendToEagentWorkflow(
  baseUrl: string,
  workflowId: string,
  context: string,
  apiKey: string,
  filePaths: string[] = [],
) {
  const aiMsgId = `a${Date.now()}`

  try {
    const latestUser = messages.value
      .filter(m => m.from === 'user')
      .at(-1)?.text ?? ''
    const workflowInput = buildWorkflowInput(
      latestUser,
      context,
      eagentWorkflowInputNodeId.value,
      filePaths,
    )

    let streamed = false
    let renderedText = ''
    let events: any[] = []
    const appendChunk = (incomingText: string) => {
      if (!incomingText) return

      let chunk = incomingText
      if (incomingText.startsWith(renderedText)) {
        chunk = incomingText.slice(renderedText.length)
      } else if (renderedText.endsWith(incomingText)) {
        chunk = ''
      }
      if (!chunk) return

      renderedText += chunk
      const aiMsg = getOrCreateAiMessage(aiMsgId)
      aiMsg.text = renderedText
      scrollToBottom()
    }
    const handleStreamEvent = (sse: any, event: any) => {
      const sessionId = sse?.session_id || event?.session_id
      if (sessionId) {
        eagentWorkflowSessionId.value = sessionId
      }
      if (!event) return

      const eventName = event.event || event.type || sse?.event || sse?.type
      if (eventName === 'guide_word' || eventName === 'guide_question') return

      events.push(event)
      if (eventName === 'input') {
        if (event.node_id) eagentWorkflowInputNodeId.value = event.node_id
        if (typeof event.message_id === 'number') eagentWorkflowMessageId.value = event.message_id
        return
      }

      const isStreamEvent = eventName === 'stream_msg' || eventName === 'message_delta' || eventName === 'chat.completion.chunk' || Boolean(event?.choices?.[0]?.delta)
      if (!isStreamEvent) return

      const chunk = extractStreamText(sse, event)
      if (chunk) {
        streamed = true
        appendChunk(chunk)
      }
    }

    const firstBody: Record<string, unknown> = {
      workflow_id: workflowId,
      stream: true,
      input: workflowInput,
    }
    if (eagentWorkflowSessionId.value) {
      firstBody.session_id = eagentWorkflowSessionId.value
    }
    if (eagentWorkflowMessageId.value) {
      firstBody.message_id = eagentWorkflowMessageId.value
    }

    await invokeEagentWorkflow(baseUrl, apiKey, firstBody, handleStreamEvent)
    const firstInputNodeId = findWorkflowInputNodeId(events)
    if (firstInputNodeId) {
      eagentWorkflowInputNodeId.value = firstInputNodeId
    }
    const firstInputMessageId = findWorkflowInputMessageId(events)
    if (firstInputMessageId) {
      eagentWorkflowMessageId.value = firstInputMessageId
    }
    let text = extractWorkflowText(events)
    resetWorkflowSessionIfClosed(events)

    // 第一次调用只初始化流程并返回输入节点，需再用同一 session 提交用户输入。
    if (!streamed && !text) {
      events = []
      const secondBody: Record<string, unknown> = {
        workflow_id: workflowId,
        session_id: eagentWorkflowSessionId.value,
        stream: true,
        input: buildWorkflowInput(
          latestUser,
          context,
          eagentWorkflowInputNodeId.value,
          filePaths,
        ),
      }
      if (eagentWorkflowMessageId.value) {
        secondBody.message_id = eagentWorkflowMessageId.value
      }
      await invokeEagentWorkflow(baseUrl, apiKey, secondBody, handleStreamEvent)
      const secondInputNodeId = findWorkflowInputNodeId(events)
      if (secondInputNodeId) {
        eagentWorkflowInputNodeId.value = secondInputNodeId
      }
      const secondInputMessageId = findWorkflowInputMessageId(events)
      if (secondInputMessageId) {
        eagentWorkflowMessageId.value = secondInputMessageId
      }
      text = extractWorkflowText(events)
      resetWorkflowSessionIfClosed(events)
    }

    const aiMsg = messages.value.find(m => m.id === aiMsgId)
    const finalText = aiMsg?.text || text || '（工作流未返回内容）'
    if (!streamed && finalText) {
      getOrCreateAiMessage(aiMsgId)
      await animateAiText(aiMsgId, finalText)
    } else if (!aiMsg && finalText) {
      getOrCreateAiMessage(aiMsgId).text = finalText
    }
  } catch (e: any) {
    getOrCreateAiMessage(aiMsgId).text = `⚠️ E-Agent 工作流调用失败：${e.message}`
  } finally {
    isTyping.value = false
    scrollToBottom()
  }
}

async function sendMessage(text?: string) {
  const attachmentPaths = pendingAttachments.value
    .map(item => item.path)
    .filter((path): path is string => Boolean(path))
  const msg = (text ?? inputText.value).trim() || (attachmentPaths.length ? '请结合我上传的文件进行解读。' : '')
  if (!msg || isTyping.value || hasAttachmentUpload.value || hasAttachmentError.value || !agent.value) return
  const messageAttachments = pendingAttachments.value.map(({ name, size }) => ({ name, size }))
  inputText.value = ''
  resetAttachments()

  messages.value.push({
    id: `u${Date.now()}`,
    from: 'user',
    text: msg,
    time: getTime(),
    attachments: messageAttachments,
  })
  scrollToBottom()

  isTyping.value = true

  if (agentId.value === 'life-planning') {
    const shouldRemember = rememberCurrentMessage.value
    rememberCurrentMessage.value = false
    void sendToManagedLifePlanning(msg, shouldRemember)
    return
  }

  // 有外部 API 配置时走真实接口
  if (agent.value.externalApi) {
    const { type, baseUrl, assistantId, workflowId, apiKey } = agent.value.externalApi
    const studentContext = await buildStudentContext()
    if (type === 'dify') {
      sendToDify(baseUrl, assistantId || '', msg)
    } else if (type === 'eagent-workflow') {
      sendToEagentWorkflow(baseUrl, workflowId || '', studentContext, apiKey || '', attachmentPaths)
    } else {
      sendToEagent(baseUrl, assistantId || '', studentContext, apiKey)
    }
    return
  }

  // 否则走 mock 回答
  const delay = 900 + (msg.length % 3) * 200
  setTimeout(() => {
    isTyping.value = false
    const response = getMockResponse(agentId.value, msg, auth.user)
    messages.value.push({ id: `a${Date.now()}`, from: 'ai', text: response, time: getTime() })
    scrollToBottom()
  }, delay)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

const routePrefix: Record<string, string> = {
  student:        'student',
  teacher:        'teacher',
  course_teacher: 'course-teacher',
  department:     'department',
  dean:           'dean',
}
const prefix = computed(() => routePrefix[role.value] ?? role.value)

function goToAgent(id: string) {
  router.push(`/${prefix.value}/agents/${id}`)
}

function backToPlaza() {
  router.push(`/${prefix.value}/agents`)
}

function newChat() {
  messages.value = []
  inputText.value = ''
  resetAttachments()
  rememberCurrentMessage.value = false
  eagentWorkflowSessionId.value = ''
  eagentWorkflowInputNodeId.value = ''
  eagentWorkflowMessageId.value = null
}
</script>

<template>
  <div v-if="agent" class="chat-root">

    <!-- iframe 模式：外部平台嵌入，保持和普通模式相同的 sidebar 布局 -->
    <template v-if="agent.iframeUrl && !agent.externalApi">
      <aside class="chat-sidebar">
        <button class="back-btn" @click="backToPlaza">
          <el-icon><ArrowLeft /></el-icon>
          <span>AI 领航助手</span>
        </button>
        <div class="cur-agent">
          <div class="cur-agent-icon" :style="{ background: agent.iconBg }">{{ agent.icon }}</div>
          <div class="cur-agent-info">
            <div class="cur-agent-name">{{ agent.name }}</div>
            <div class="cur-agent-desc">{{ agent.description }}</div>
          </div>
        </div>
        <div class="sidebar-divider" />
        <div class="sidebar-section">
          <div class="sidebar-section-title">最近打开的智能体</div>
          <div v-for="a in recentAgents" :key="a.id" class="recent-agent-item" @click="goToAgent(a.id)">
            <span class="recent-agent-icon" :style="{ background: a.iconBg }">{{ a.icon }}</span>
            <span class="recent-agent-name">{{ a.name }}</span>
          </div>
        </div>
      </aside>
      <div class="chat-main" style="padding:0;overflow:hidden;">
        <iframe
          :src="agent.iframeUrl"
          class="agent-iframe"
          frameborder="0"
          allow="microphone"
        />
      </div>

    </template>
    <!-- 普通对话模式 -->
    <template v-else>
    <!-- Left sidebar -->
    <aside class="chat-sidebar">
      <!-- Back -->
      <button class="back-btn" @click="backToPlaza">
        <el-icon><ArrowLeft /></el-icon>
        <span>AI 领航助手</span>
      </button>

      <!-- Current agent -->
      <div class="cur-agent">
        <div class="cur-agent-icon" :style="{ background: agent.iconBg }">
          {{ agent.icon }}
        </div>
        <div class="cur-agent-info">
          <div class="cur-agent-name">{{ agent.name }}</div>
          <div class="cur-agent-desc">{{ agent.description }}</div>
        </div>
      </div>

      <div class="sidebar-divider" />

      <!-- Recently opened -->
      <div class="sidebar-section">
        <div class="sidebar-section-title">最近打开的智能体</div>
        <div
          v-for="a in recentAgents" :key="a.id"
          class="recent-agent-item"
          @click="goToAgent(a.id)"
        >
          <span class="recent-icon" :style="{ background: a.iconBg }">{{ a.icon }}</span>
          <div class="recent-info">
            <div class="recent-name">{{ a.name }}</div>
            <div class="recent-cat">{{ a.category }}</div>
          </div>
        </div>
      </div>

      <div class="sidebar-divider" />

      <!-- History -->
      <div class="sidebar-section flex-1">
        <div class="sidebar-section-title">历史对话</div>
        <div v-for="h in mockHistory" :key="h.id" class="history-item">
          <el-icon class="history-icon"><ChatDotRound /></el-icon>
          <div class="history-info">
            <div class="history-label">{{ h.label }}</div>
            <div class="history-time">{{ h.time }}</div>
          </div>
        </div>
      </div>

      <div v-if="agentId === 'life-planning'" class="sidebar-memory">
        <button class="memory-manage-btn" @click="openMemoryDialog">
          <el-icon><Document /></el-icon>
          <span>管理长期记忆</span>
        </button>
      </div>

      <!-- New conversation button -->
      <div class="sidebar-bottom">
        <button class="new-chat-btn" @click="newChat">
          <el-icon><Plus /></el-icon>
          <span>新建会话</span>
        </button>
      </div>
    </aside>

    <!-- Main chat area -->
    <div class="chat-main">
      <!-- Top header -->
      <div class="chat-header">
        <div class="chat-header-icon" :style="{ background: agent.iconBg }">{{ agent.icon }}</div>
        <div>
          <div class="chat-header-name">{{ agent.name }}</div>
          <div class="chat-header-desc">{{ agent.description }}</div>
        </div>
      </div>

      <!-- Chat body -->
      <div ref="chatBodyRef" class="chat-body">
        <!-- Welcome / preset questions -->
        <div v-if="messages.length === 0" class="welcome-screen">
          <div class="welcome-icon" :style="{ background: agent.iconBg }">{{ agent.icon }}</div>
          <div class="welcome-title">{{ agent.name }}</div>
          <div class="welcome-msg">{{ agent.welcomeMessage }}</div>
          <div class="preset-list">
            <button
              v-for="q in agent.presetQuestions" :key="q"
              class="preset-item"
              @click="sendMessage(q)"
            >
              <span class="preset-text">{{ q }}</span>
              <el-icon class="preset-arrow"><ArrowRight /></el-icon>
            </button>
          </div>
        </div>

        <!-- Messages -->
        <template v-else>
          <div v-for="msg in messages" :key="msg.id" class="msg-row" :class="msg.from">
            <!-- AI avatar -->
            <div v-if="msg.from === 'ai'" class="msg-avatar ai-avatar" :style="{ background: agent.iconBg }">
              {{ agent.icon }}
            </div>

            <div class="msg-bubble-wrap">
              <div class="msg-bubble" :class="msg.from">
                <div v-if="msg.from === 'ai'" class="markdown-body" v-html="formatText(msg.text)"></div>
                <span v-else>{{ msg.text }}</span>
                <div v-if="msg.from === 'user' && msg.attachments?.length" class="message-attachments">
                  <div v-for="file in msg.attachments" :key="file.name" class="message-attachment">
                    <el-icon><Document /></el-icon>
                    <span>{{ file.name }}</span>
                    <small>{{ formatFileSize(file.size) }}</small>
                  </div>
                </div>
                <span v-if="msg.id === streamingMessageId" class="stream-cursor" aria-hidden="true" />
              </div>
              <div class="msg-time">{{ msg.time }}</div>
            </div>

            <!-- User avatar -->
            <div v-if="msg.from === 'user'" class="msg-avatar user-avatar">
              {{ auth.user?.name?.[0] || '我' }}
            </div>
          </div>

          <!-- Typing indicator -->
          <div v-if="showTypingIndicator" class="msg-row ai">
            <div class="msg-avatar ai-avatar" :style="{ background: agent.iconBg }">{{ agent.icon }}</div>
            <div class="typing-bubble">
              <span class="typing-dot" />
              <span class="typing-dot" />
              <span class="typing-dot" />
            </div>
          </div>
        </template>
      </div>

      <!-- Input area -->
      <div class="chat-input-area">
        <!-- Quick preset chips (shown after first message) -->
        <div v-if="messages.length > 0 && messages.length < 5" class="quick-chips">
          <button
            v-for="q in agent.presetQuestions.slice(0,3)" :key="q"
            class="quick-chip"
            @click="sendMessage(q)"
          >{{ q }}</button>
        </div>

        <div v-if="supportsFileUpload && pendingAttachments.length" class="attachment-list">
          <div v-for="file in pendingAttachments" :key="file.id" class="attachment-item" :class="{ error: file.error }">
            <el-icon><Document /></el-icon>
            <span class="attachment-name">{{ file.name }}</span>
            <span class="attachment-status">
              {{ file.error || (file.uploading ? '上传中...' : formatFileSize(file.size)) }}
            </span>
            <button type="button" class="remove-attachment-btn" :disabled="file.uploading" :title="`移除 ${file.name}`" @click="removeAttachment(file.id)">
              <el-icon><Close /></el-icon>
            </button>
          </div>
        </div>

        <div class="input-row">
          <input
            v-if="supportsFileUpload"
            ref="attachmentInputRef"
            class="attachment-input"
            type="file"
            multiple
            @change="handleAttachmentSelection"
          />
          <button
            v-if="supportsFileUpload"
            type="button"
            class="attachment-btn"
            :disabled="isTyping"
            title="添加附件"
            aria-label="添加附件"
            @click="openAttachmentPicker"
          >
            <el-icon><Paperclip /></el-icon>
          </button>
          <textarea
            v-model="inputText"
            class="chat-textarea"
            :placeholder="`和 ${agent.name} 对话...`"
            rows="1"
            @keydown="onKeydown"
          />
          <button class="send-btn" :class="{ active: canSend }" :disabled="!canSend" @click="sendMessage()">
            <el-icon><Promotion /></el-icon>
          </button>
        </div>

        <el-checkbox
          v-if="agentId === 'life-planning'"
          v-model="rememberCurrentMessage"
          class="remember-checkbox"
          :disabled="isTyping"
        >将本次明确的长期目标或偏好保存为记忆</el-checkbox>

        <div class="input-hint">按 Enter 发送，Shift+Enter 换行 <span v-if="supportsFileUpload">· 单个文件最大 20 MB</span></div>
      </div>
    </div>
    </template>

    <el-dialog v-model="memoryDialogVisible" title="我的长期记忆" width="520px" class="memory-dialog">
      <div v-loading="memoriesLoading" class="memory-dialog-body">
        <p class="memory-dialog-hint">仅显示你确认保存的低敏长期目标、偏好和行动计划。</p>
        <div v-if="!memoriesLoading && !memories.length" class="memory-empty">暂无已保存的长期记忆。</div>
        <div v-for="item in memories" :key="item.id" class="memory-row">
          <span>{{ item.memory }}</span>
          <button type="button" class="memory-delete-btn" title="删除这条记忆" @click="removeMemory(item.id)">
            <el-icon><Delete /></el-icon>
          </button>
        </div>
      </div>
    </el-dialog>
  </div>

  <div v-else class="not-found">
    <div style="font-size:40px">🤖</div>
    <div style="font-size:16px;font-weight:700;color:#374151;margin-top:12px">智能体不存在</div>
    <el-button style="margin-top:16px" @click="backToPlaza">返回 AI 广场</el-button>
  </div>
</template>

<style scoped>
.chat-root {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: #f6f8fb;
}

.agent-iframe {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
}

/* ── Left sidebar ── */
.chat-sidebar {
  width: 248px;
  flex-shrink: 0;
  background: #fff;
  border-right: 1px solid #edf0f6;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px;
  font-size: 13px;
  font-weight: 600;
  color: #6b7280;
  background: transparent;
  border: none;
  cursor: pointer;
  font-family: inherit;
  transition: color 0.15s;
  border-bottom: 1px solid #f0f2f7;
}
.back-btn:hover { color: #2A4D99; }

.cur-agent {
  padding: 14px 16px;
  display: flex;
  gap: 10px;
  border-bottom: 1px solid #f0f2f7;
}
.cur-agent-icon {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}
.cur-agent-name { font-size: 13px; font-weight: 700; color: #111827; line-height: 1.3; }
.cur-agent-desc {
  font-size: 11px;
  color: #9facc5;
  margin-top: 3px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}

.sidebar-divider { height: 1px; background: #f0f2f7; margin: 0; }

.sidebar-section {
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
}
.sidebar-section.flex-1 { flex: 1; }
.sidebar-section-title {
  font-size: 11px;
  font-weight: 700;
  color: #9facc5;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  margin-bottom: 6px;
}

.recent-agent-item {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 7px 8px;
  border-radius: 9px;
  cursor: pointer;
  transition: background 0.12s;
}
.recent-agent-item:hover { background: #f6f8fb; }
.recent-icon {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  flex-shrink: 0;
}
.recent-name { font-size: 12px; font-weight: 600; color: #374151; }
.recent-cat { font-size: 10.5px; color: #9facc5; margin-top: 1px; }

.history-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 7px 8px;
  border-radius: 9px;
  cursor: pointer;
  transition: background 0.12s;
}
.history-item:hover { background: #f6f8fb; }
.history-icon { font-size: 14px; color: #d1d5db; margin-top: 2px; flex-shrink: 0; }
.history-label { font-size: 12px; color: #374151; line-height: 1.4; }
.history-time { font-size: 10.5px; color: #9facc5; margin-top: 2px; }
.sidebar-memory { padding: 0 14px 12px; }
.memory-manage-btn { width: 100%; display: flex; align-items: center; gap: 8px; border: 1px solid #d9e2f3; background: #fff; color: #52647e; padding: 8px 10px; font-size: 12px; cursor: pointer; border-radius: 4px; }
.memory-manage-btn:hover { border-color: #8aa4c8; color: #2b5c9a; }
.remember-checkbox { display: flex; margin: 8px 2px 0; font-size: 12px; color: #69778c; }
.memory-dialog-body { min-height: 100px; }
.memory-dialog-hint { margin: 0 0 14px; color: #667085; font-size: 13px; line-height: 1.6; }
.memory-empty { padding: 20px 0; text-align: center; color: #98a2b3; font-size: 13px; }
.memory-row { display: flex; align-items: flex-start; gap: 10px; padding: 11px 0; border-top: 1px solid #edf0f4; color: #344054; font-size: 13px; line-height: 1.6; }
.memory-row span { flex: 1; }
.memory-delete-btn { border: 0; background: transparent; color: #98a2b3; cursor: pointer; padding: 3px; }
.memory-delete-btn:hover { color: #d92d20; }

.sidebar-bottom {
  padding: 12px 16px;
  border-top: 1px solid #f0f2f7;
}
.new-chat-btn {
  width: 100%;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  border: 1.5px dashed #d1d5db;
  border-radius: 10px;
  background: transparent;
  font-size: 13px;
  font-weight: 600;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}
.new-chat-btn:hover { border-color: #2A4D99; color: #2A4D99; background: rgba(42,77,153,0.04); }

/* ── Main chat ── */
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 24px;
  background: #fff;
  border-bottom: 1px solid #edf0f6;
  flex-shrink: 0;
}
.chat-header-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}
.chat-header-name { font-size: 15px; font-weight: 800; color: #111827; }
.chat-header-desc { font-size: 12px; color: #9facc5; margin-top: 2px; }

/* ── Chat body ── */
.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Welcome */
.welcome-screen {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 24px;
  max-width: 560px;
  margin: 0 auto;
  width: 100%;
}
.welcome-icon {
  width: 68px;
  height: 68px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 34px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
/* Markdown output */
.msg-bubble.ai .markdown-body {
  line-height: 1.65;
}
.msg-bubble.ai .markdown-body :deep(p) {
  margin: 0 0 8px;
}
.msg-bubble.ai .markdown-body :deep(p:last-child) {
  margin-bottom: 0;
}
.msg-bubble.ai .markdown-body :deep(h1),
.msg-bubble.ai .markdown-body :deep(h2),
.msg-bubble.ai .markdown-body :deep(h3),
.msg-bubble.ai .markdown-body :deep(h4) {
  font-weight: 700;
  color: #111827;
  margin: 12px 0 6px;
  line-height: 1.35;
}
.msg-bubble.ai .markdown-body :deep(h1) { font-size: 15px; }
.msg-bubble.ai .markdown-body :deep(h2) { font-size: 14px; }
.msg-bubble.ai .markdown-body :deep(h3),
.msg-bubble.ai .markdown-body :deep(h4) { font-size: 13.5px; }
.msg-bubble.ai .markdown-body :deep(ul),
.msg-bubble.ai .markdown-body :deep(ol) {
  margin: 0 0 8px;
  padding-left: 18px;
}
.msg-bubble.ai .markdown-body :deep(li) {
  margin: 2px 0;
}
.msg-bubble.ai .markdown-body :deep(strong) {
  color: #111827;
}
.msg-bubble.ai .markdown-body :deep(code) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  background: #f1f3f7;
  color: #2A4D99;
  padding: 2px 5px;
  border-radius: 4px;
}
.msg-bubble.ai .markdown-body :deep(pre) {
  background: #111827;
  color: #e5e7eb;
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}
.msg-bubble.ai .markdown-body :deep(pre code) {
  background: transparent;
  color: inherit;
  padding: 0;
  font-size: 12px;
}
.msg-bubble.ai .markdown-body :deep(blockquote) {
  border-left: 3px solid #2A4D99;
  padding-left: 10px;
  color: #6b7280;
  margin: 8px 0;
}
.msg-bubble.ai .markdown-body :deep(hr) {
  border: none;
  border-top: 1px solid #edf0f6;
  margin: 12px 0;
}

  margin-bottom: 16px;
}
.welcome-title {
  font-size: 18px;
  font-weight: 800;
  color: #111827;
  margin-bottom: 8px;
}
.welcome-msg {
  font-size: 13.5px;
  color: #6b7280;
  text-align: center;
  line-height: 1.6;
  margin-bottom: 24px;
  max-width: 400px;
}
.preset-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
.preset-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 13px 16px;
  border: 1px solid #edf0f6;
  border-radius: 12px;
  background: #fff;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
  text-align: left;
}
.preset-item:hover {
  border-color: #2A4D99;
  background: rgba(42,77,153,0.03);
}
.preset-text { font-size: 13.5px; font-weight: 500; color: #374151; }
.preset-arrow { font-size: 13px; color: #9facc5; flex-shrink: 0; }
.preset-item:hover .preset-arrow { color: #2A4D99; }

/* Messages */
.msg-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}
.msg-row.user { flex-direction: row-reverse; }

.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  flex-shrink: 0;
  margin-top: 2px;
}
.ai-avatar { }
.user-avatar {
  background: linear-gradient(135deg, #2A4D99, #5580d4);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}

.msg-bubble-wrap {
  display: flex;
  flex-direction: column;
  gap: 3px;
  max-width: 72%;
}
.msg-row.user .msg-bubble-wrap { align-items: flex-end; }

.msg-bubble {
  padding: 12px 16px;
  border-radius: 16px;
  font-size: 13.5px;
  line-height: 1.65;
  word-break: break-word;
}
.msg-bubble.ai {
  background: #fff;
  color: #1f2937;
  border: 1px solid #edf0f6;
  border-radius: 4px 16px 16px 16px;
  box-shadow: 0 1px 4px rgba(10,19,41,0.05);
}
.msg-bubble.user {
  background: #2A4D99;
  color: #fff;
  border-radius: 16px 4px 16px 16px;
}
.message-attachments {
  display: flex;
  flex-direction: column;
  gap: 5px;
  margin-top: 8px;
}
.message-attachment {
  display: flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  padding: 6px 8px;
  border-radius: 5px;
  background: rgba(255,255,255,0.14);
  font-size: 12px;
}
.message-attachment span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.message-attachment small {
  margin-left: auto;
  white-space: nowrap;
  color: rgba(255,255,255,0.72);
}
.msg-time {
  font-size: 10.5px;
  color: #c4cad4;
  padding: 0 4px;
}

/* Typing indicator */
.typing-bubble {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 14px 18px;
  background: #fff;
  border: 1px solid #edf0f6;
  border-radius: 4px 16px 16px 16px;
  box-shadow: 0 1px 4px rgba(10,19,41,0.05);
}
.typing-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #9facc5;
  animation: bounce 1.2s infinite;
}
.typing-dot:nth-child(2) { animation-delay: 0.2s; }
.typing-dot:nth-child(3) { animation-delay: 0.4s; }

.stream-cursor {
  display: inline-block;
  width: 2px;
  height: 1em;
  margin-left: 3px;
  vertical-align: -0.12em;
  background: currentColor;
  animation: stream-cursor-blink 0.8s steps(1) infinite;
}

@keyframes stream-cursor-blink {
  50% { opacity: 0; }
}
@keyframes bounce {
  0%, 80%, 100% { transform: scale(0.7); opacity: 0.5; }
  40% { transform: scale(1); opacity: 1; }
}

/* ── Input area ── */
.chat-input-area {
  padding: 12px 24px 16px;
  background: #fff;
  border-top: 1px solid #edf0f6;
  flex-shrink: 0;
}

.quick-chips {
  display: flex;
  gap: 7px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
.quick-chip {
  font-size: 12px;
  font-weight: 500;
  color: #2A4D99;
  background: rgba(42,77,153,0.07);
  border: 1px solid rgba(42,77,153,0.18);
  padding: 5px 11px;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.12s;
  font-family: inherit;
}
.quick-chip:hover { background: rgba(42,77,153,0.14); }

.attachment-list {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-bottom: 10px;
}
.attachment-item {
  display: flex;
  align-items: center;
  min-width: 0;
  max-width: min(100%, 360px);
  gap: 7px;
  padding: 6px 7px 6px 9px;
  border: 1px solid #d8e2f2;
  border-radius: 6px;
  color: #35527d;
  background: #f7faff;
  font-size: 12px;
}
.attachment-item.error {
  border-color: #fecaca;
  color: #b42318;
  background: #fff7f7;
}
.attachment-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.attachment-status {
  flex-shrink: 0;
  color: #7c8da7;
  white-space: nowrap;
}
.attachment-item.error .attachment-status { color: inherit; }
.remove-attachment-btn,
.attachment-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  cursor: pointer;
  font-family: inherit;
}
.remove-attachment-btn {
  width: 20px;
  height: 20px;
  padding: 0;
  flex-shrink: 0;
  color: inherit;
}
.remove-attachment-btn:disabled { cursor: wait; opacity: 0.45; }
.attachment-input { display: none; }
.attachment-btn {
  width: 42px;
  height: 42px;
  flex-shrink: 0;
  border: 1.5px solid #e5e7eb;
  border-radius: 8px;
  color: #55709b;
  background: #fafbfd;
  font-size: 17px;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}
.attachment-btn:hover:not(:disabled) {
  border-color: #2A4D99;
  color: #2A4D99;
  background: #f4f7fc;
}
.attachment-btn:disabled { cursor: not-allowed; opacity: 0.55; }

.input-row {
  display: flex;
  gap: 10px;
  align-items: flex-end;
}
.chat-textarea {
  flex: 1;
  padding: 10px 14px;
  border: 1.5px solid #e5e7eb;
  border-radius: 12px;
  font-size: 13.5px;
  font-family: inherit;
  color: #1f2937;
  resize: none;
  outline: none;
  min-height: 42px;
  max-height: 120px;
  line-height: 1.5;
  transition: border-color 0.15s;
  background: #fafbfd;
}
.chat-textarea:focus { border-color: #2A4D99; background: #fff; }
.chat-textarea::placeholder { color: #b4bcc8; }

.send-btn {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  border: none;
  background: #d1d5db;
  color: #fff;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s, transform 0.1s;
  flex-shrink: 0;
}
.send-btn.active { background: #2A4D99; }
.send-btn.active:hover { background: #1e3a7a; }
.send-btn.active:active { transform: scale(0.95); }
.send-btn:disabled { cursor: not-allowed; }

.input-hint {
  font-size: 11px;
  color: #c4cad4;
  margin-top: 7px;
  padding-left: 2px;
}

/* Not found */
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
}
</style>
