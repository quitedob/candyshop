<template>
  <div class="ai-dashboard">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 inline-flex items-center">
          {{ t('admin.ai.title') }}
          <AiHelpHint topic="ai_console" />
        </h1>
        <p class="text-gray-600 text-sm mt-1">{{ t('admin.ai.description') }}</p>
      </div>
      <button
        class="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
        @click="clearChat"
      >
        <Icon name="heroicons:trash" class="h-4 w-4" />
        {{ t('admin.ai.clear_chat') }}
      </button>
    </div>

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Chat Panel -->
      <div class="lg:col-span-2 rounded-2xl border border-gray-200 bg-white flex flex-col" style="min-height: 500px;">
        <div ref="chatContainer" class="flex-1 overflow-y-auto p-6 space-y-4">
          <div v-if="messages.length === 0" class="flex items-center justify-center h-full text-gray-400">
            <div class="text-center">
              <Icon name="heroicons:chat-bubble-left-right" class="h-12 w-12 mx-auto mb-3 text-orange-300" />
              <p>{{ t('admin.ai.start_conversation') }}</p>
              <div class="mt-4 flex flex-wrap gap-2 justify-center">
                <button
                  v-for="suggestion in suggestions"
                  :key="suggestion"
                  class="rounded-full border border-orange-200 bg-orange-50 px-3 py-1 text-xs text-orange-700 hover:bg-orange-100 transition-colors"
                  @click="sendMessage(suggestion)"
                >
                  {{ suggestion }}
                </button>
              </div>
            </div>
          </div>
          <div
            v-for="(msg, idx) in messages"
            :key="idx"
            :class="['flex gap-3', msg.role === 'user' ? 'justify-end' : '']"
          >
            <div
              v-if="msg.role === 'assistant'"
              class="flex-shrink-0 w-8 h-8 rounded-full bg-orange-100 flex items-center justify-center"
            >
              <Icon name="heroicons:sparkles" class="h-4 w-4 text-orange-500" />
            </div>
            <div
              :class="[
                'max-w-[80%] rounded-2xl px-4 py-3 text-sm',
                msg.role === 'user'
                  ? 'bg-orange-500 text-white rounded-br-md'
                  : 'bg-gray-100 text-gray-800 rounded-bl-md'
              ]"
            >
              <div v-if="msg.toolCalls && msg.toolCalls.length" class="mb-2">
                <div
                  v-for="tc in msg.toolCalls"
                  :key="tc.id"
                  class="text-xs bg-white/80 rounded-lg px-3 py-2 mb-1 font-mono"
                >
                  <span class="font-semibold text-orange-600">{{ tc.name }}</span>
                  <pre class="text-gray-500 mt-1 text-[11px] overflow-x-auto">{{ formatToolArgs(tc.args) }}</pre>
                </div>
              </div>
              <template v-if="msg.content">
                <div v-html="renderContent(msg.content)" />
              </template>
              <span v-else-if="msg.streaming" class="inline-flex gap-1">
                <span class="w-2 h-2 rounded-full bg-gray-400 animate-bounce" style="animation-delay:0ms" />
                <span class="w-2 h-2 rounded-full bg-gray-400 animate-bounce" style="animation-delay:150ms" />
                <span class="w-2 h-2 rounded-full bg-gray-400 animate-bounce" style="animation-delay:300ms" />
              </span>
            </div>
            <div
              v-if="msg.role === 'user'"
              class="flex-shrink-0 w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center"
            >
              <Icon name="heroicons:user" class="h-4 w-4 text-gray-500" />
            </div>
          </div>
        </div>
        <!-- Input -->
        <div class="border-t border-gray-200 p-4">
          <form class="flex gap-3" @submit.prevent="handleSubmit">
            <input
              ref="inputEl"
              v-model="input"
              type="text"
              :placeholder="t('admin.ai.input_placeholder')"
              :disabled="isStreaming"
              class="flex-1 rounded-xl border border-gray-300 px-4 py-2.5 text-sm focus:border-orange-400 focus:outline-none focus:ring-2 focus:ring-orange-100 disabled:bg-gray-50"
              autocomplete="off"
            />
            <button
              type="submit"
              :disabled="!input.trim() || isStreaming"
              class="inline-flex items-center gap-2 rounded-xl bg-orange-500 px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-orange-600 disabled:opacity-50"
            >
              <Icon v-if="isStreaming" name="heroicons:stop" class="h-4 w-4" />
              <Icon v-else name="heroicons:paper-airplane" class="h-4 w-4" />
              {{ isStreaming ? t('admin.ai.stop') : t('admin.ai.send') }}
            </button>
          </form>
        </div>
      </div>

      <!-- Sidebar: HITL + History -->
      <div class="space-y-6">
        <!-- Pending Approvals -->
        <div class="rounded-2xl border border-gray-200 bg-white p-5">
          <h3 class="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-2">
            <Icon name="heroicons:clipboard-document-check" class="h-4 w-4 text-orange-500" />
            {{ t('admin.ai.pending_approvals') }}
          </h3>
          <div v-if="pendingActions.length === 0" class="text-xs text-gray-400 text-center py-4">
            {{ t('admin.ai.no_pending_approvals') }}
          </div>
          <div v-for="action in pendingActions" :key="action.id" class="border border-gray-100 rounded-lg p-3 mb-2">
            <p class="text-xs font-medium text-gray-700">{{ action.label }}</p>
            <p class="text-xs text-gray-400 mt-1">{{ action.detail }}</p>
            <div class="flex gap-2 mt-3">
              <button
                class="flex-1 rounded-lg bg-green-500 px-3 py-1.5 text-xs font-medium text-white hover:bg-green-600 transition-colors"
                @click="approveAction(action.id)"
              >
                {{ t('admin.ai.approve') }}
              </button>
              <button
                class="flex-1 rounded-lg bg-red-100 px-3 py-1.5 text-xs font-medium text-red-700 hover:bg-red-200 transition-colors"
                @click="rejectAction(action.id)"
              >
                {{ t('admin.ai.reject') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Tool Call History -->
        <div class="rounded-2xl border border-gray-200 bg-white p-5">
          <h3 class="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-2">
            <Icon name="heroicons:wrench-screwdriver" class="h-4 w-4 text-orange-500" />
            {{ t('admin.ai.tool_call_history') }}
          </h3>
          <div v-if="toolHistory.length === 0" class="text-xs text-gray-400 text-center py-4">
            {{ t('admin.ai.no_tool_calls') }}
          </div>
          <div
            v-for="th in toolHistory.slice(0, 10)"
            :key="th.id"
            class="text-xs border-b border-gray-50 py-2 last:border-0"
          >
            <span class="font-mono font-semibold text-orange-600">{{ th.name }}</span>
            <span :class="['ml-2 px-1.5 py-0.5 rounded text-[10px] font-medium', th.status === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700']">
              {{ enumLabel('tool_status', th.status) }}
            </span>
            <div class="text-gray-400 mt-0.5">{{ th.time }}</div>
          </div>
        </div>

        <!-- Model Config -->
        <div class="rounded-2xl border border-gray-200 bg-white p-5">
          <h3 class="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-2">
            <Icon name="heroicons:cog-6-tooth" class="h-4 w-4 text-orange-500" />
            {{ t('admin.ai.configuration') }}
          </h3>
          <div class="space-y-3 text-sm">
            <div>
              <label class="block text-xs text-gray-500 mb-1">{{ t('admin.ai.model_label') }}</label>
              <input
                v-model="config.model"
                type="text"
                class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs focus:border-orange-400 focus:outline-none focus:ring-1 focus:ring-orange-100"
              />
            </div>
            <div>
              <label class="block text-xs text-gray-500 mb-1">{{ t('admin.ai.temperature_label') }}</label>
              <input
                v-model.number="config.temperature"
                type="range"
                min="0"
                max="2"
                step="0.1"
                class="w-full accent-orange-500"
              />
              <div class="text-xs text-gray-400 text-right">{{ config.temperature }}</div>
            </div>
            <div>
              <label class="block text-xs text-gray-500 mb-1">{{ t('admin.ai.agent_mode') }}</label>
              <select
                v-model="config.agentMode"
                class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs focus:border-orange-400 focus:outline-none"
              >
                <option value="b2b-coordinator">{{ t('admin.ai.mode_b2b') }}</option>
                <option value="plan-order">{{ t('admin.ai.mode_plan') }}</option>
                <option value="trade-assistant">{{ t('admin.ai.mode_trade') }}</option>
                <option value="chatbot">{{ t('admin.ai.mode_chatbot') }}</option>
                <option value="recommend">{{ t('admin.ai.mode_recommend') }}</option>
                <option value="search">{{ t('admin.ai.mode_search') }}</option>
                <option value="translate">{{ t('admin.ai.mode_translate') }}</option>
                <option value="analyze-inquiry">{{ t('admin.ai.mode_analyze_inquiry') }}</option>
                <option value="generate-quotation">{{ t('admin.ai.mode_generate_quotation') }}</option>
              </select>
            </div>
            <div v-if="showOrderPicker">
              <label class="block text-xs text-gray-500 mb-1">{{ t('admin.ai.order_context') }}</label>
              <select
                v-model="selectedOrderId"
                class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-xs focus:border-orange-400 focus:outline-none"
              >
                <option value="">{{ t('admin.ai.order_none') }}</option>
                <option v-for="o in orderOptions" :key="o.id" :value="o.id">
                  {{ o.label }}
                </option>
              </select>
              <p class="mt-1 text-[10px] text-gray-400">{{ t('admin.ai.order_context_hint') }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })
const { t } = useI18n()
const { enumLabel } = useDisplay()
const { sanitize } = useSanitizer()
const api = useApi()
const runtimeConfig = useRuntimeConfig()
const apiBase = runtimeConfig.public.apiBase || '/api/v1'

/** 按 Agent 模式映射后端端点 */
function agentStreamPath(mode: string): string | null {
  switch (mode) {
    case 'plan-order':
      return `${apiBase}/system/ai/order-processing`
    case 'trade-assistant':
      return `${apiBase}/system/ai/trade-assistant`
    case 'b2b-coordinator':
      return `${apiBase}/system/ai/b2b-coordinator`
    default:
      return null
  }
}

/** 非流式 JSON 模式（Cookie 鉴权，勿发送无效 Bearer 以免覆盖 HttpOnly JWT） */
async function callJsonMode(mode: string, text: string, orderId?: string): Promise<string> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const init = (body: unknown) => ({
    method: 'POST' as const,
    headers,
    credentials: 'include' as RequestCredentials,
    body: JSON.stringify(body),
  })
  if (mode === 'chatbot') {
    if (orderId) {
      const res = await fetch(`${apiBase}/system/chatbot/order`, init({ message: text, orderId }))
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      return data.reply || data.message || JSON.stringify(data)
    }
    const res = await fetch(`${apiBase}/system/chatbot`, init({ message: text }))
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    return data.reply || data.message || JSON.stringify(data)
  }
  if (mode === 'recommend') {
    const res = await fetch(`${apiBase}/system/recommend-products`, init({ query: text }))
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    return JSON.stringify(data.recommendations || data.products || data, null, 2)
  }
  if (mode === 'search') {
    const res = await fetch(`${apiBase}/system/search`, init({ query: text }))
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    return JSON.stringify(data.results || data, null, 2)
  }
  if (mode === 'translate') {
    const res = await fetch(`${apiBase}/system/translate`, init({ text, targetLang: 'en' }))
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    return data.translated || data.translation || JSON.stringify(data)
  }
  if (mode === 'analyze-inquiry') {
    const res = await fetch(`${apiBase}/system/analyze-inquiry`, init({ inquiryText: text }))
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    return JSON.stringify(await res.json(), null, 2)
  }
  if (mode === 'generate-quotation') {
    const res = await fetch(`${apiBase}/system/generate-quotation`, init({ customerRequirements: text }))
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    return JSON.stringify(await res.json(), null, 2)
  }
  throw new Error('unknown mode')
}

interface Message {
  role: 'user' | 'assistant'
  content: string
  toolCalls?: ToolCall[]
  streaming?: boolean
}

interface ToolCall {
  id: string
  name: string
  args: Record<string, unknown>
}

interface HITLAction {
  id: string
  label: string
  detail: string
}

interface ToolHistoryEntry {
  id: string
  name: string
  status: 'success' | 'error'
  time: string
}

const input = ref('')
const isStreaming = ref(false)
const messages = ref<Message[]>([])
const pendingActions = ref<HITLAction[]>([])
const toolHistory = ref<ToolHistoryEntry[]>([])
const chatContainer = ref<HTMLElement>()
const inputEl = ref<HTMLInputElement>()
const abortController = ref<AbortController | null>(null)

const config = reactive({
  model: 'MiniMax-M2.7',
  temperature: 0.7,
  agentMode: 'b2b-coordinator',
})

/** Chatbot / 订单处理模式可选绑定订单上下文 */
const selectedOrderId = ref('')
const orderOptions = ref<{ id: string; label: string }[]>([])

const showOrderPicker = computed(() =>
  config.agentMode === 'chatbot' || config.agentMode === 'plan-order',
)

const loadOrderOptions = async () => {
  try {
    const res = await api.get<any>('/admin/orders', { page: 1, limit: 80 })
    const rows = Array.isArray(res?.data) ? res.data : []
    orderOptions.value = rows.map((o: any) => ({
      id: String(o.id),
      label: `${o.orderNumber || o.id} · ${o.status || ''} · ${o.totalAmount ?? ''} ${o.currency || ''}`.trim(),
    }))
  } catch {
    orderOptions.value = []
  }
}

const suggestions = computed(() => [
  t('admin.ai.suggestion_1'),
  t('admin.ai.suggestion_2'),
  t('admin.ai.suggestion_3'),
  t('admin.ai.suggestion_4'),
])

function clearChat() {
  messages.value = []
  toolHistory.value = []
}

function formatToolArgs(args: unknown): string {
  if (!args) return ''
  try {
    return JSON.stringify(args, null, 2).slice(0, 300)
  } catch {
    return String(args).slice(0, 300)
  }
}

function renderContent(content: string): string {
  // First apply markdown-like formatting, then sanitize for XSS defense
  const html = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/\n/g, '<br>')
    .replace(/`([^`]+)`/g, '<code class="bg-gray-200 px-1 rounded text-xs">$1</code>')
  return sanitize(html)
}

function scrollToBottom() {
  nextTick(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight
    }
  })
}

async function handleSubmit() {
  if (!input.value.trim() || isStreaming.value) {
    if (isStreaming.value && abortController.value) {
      abortController.value.abort()
      isStreaming.value = false
    }
    return
  }
  await sendMessage(input.value.trim())
}

async function sendMessage(text: string) {
  messages.value.push({ role: 'user', content: text, streaming: false })
  input.value = ''
  const assistantMsg: Message = { role: 'assistant', content: '', streaming: true, toolCalls: [] }
  messages.value.push(assistantMsg)
  scrollToBottom()

  isStreaming.value = true
  abortController.value = new AbortController()

  const streamUrl = agentStreamPath(config.agentMode)
  if (!streamUrl) {
    try {
      assistantMsg.content = await callJsonMode(
        config.agentMode,
        text,
        selectedOrderId.value || undefined,
      )
    } catch (err: unknown) {
      assistantMsg.content = `Error: ${err instanceof Error ? err.message : 'Unknown error'}`
    }
    assistantMsg.streaming = false
    isStreaming.value = false
    scrollToBottom()
    return
  }

  try {
    let url = `${streamUrl}?query=${encodeURIComponent(text)}`
    if (selectedOrderId.value && config.agentMode === 'plan-order') {
      url += `&orderId=${encodeURIComponent(selectedOrderId.value)}`
    }
    url += `&temperature=${config.temperature}`
    url += `&model=${encodeURIComponent(config.model)}`
    const response = await fetch(url, {
      method: 'GET',
      credentials: 'include',
      signal: abortController.value.signal,
    })

    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    if (!response.body) throw new Error('No response body')

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let partialEvent = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      partialEvent += decoder.decode(value, { stream: true })
      const parts = partialEvent.split('\n\n')
      partialEvent = parts.pop() || ''

      for (const part of parts) {
        if (!part.startsWith('data: ')) continue
        try {
          handleSSEEvent(JSON.parse(part.slice(6)), assistantMsg)
        } catch {
          // 不完整 JSON，跳过
        }
      }
    }

    assistantMsg.streaming = false
  } catch (err: unknown) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      assistantMsg.content += `\n\n${t('admin.ai.stream_stopped')}`
    } else {
      assistantMsg.content = `Error: ${err instanceof Error ? err.message : 'Unknown error'}`
    }
    assistantMsg.streaming = false
  } finally {
    isStreaming.value = false
    abortController.value = null
    scrollToBottom()
  }
}

function handleSSEEvent(event: Record<string, unknown>, msg: Message) {
  const type = event.type as string
  const content = (event.content as string) || ''

  if (type === 'message' || type === 'tool_result') {
    if (content) {
      msg.content += (msg.content ? '\n\n' : '') + content
    }
    const rawCalls = event.tool_calls as Array<{ id?: string; function?: { name?: string; arguments?: string } }> | undefined
    if (rawCalls?.length) {
      if (!msg.toolCalls) msg.toolCalls = []
      for (const tc of rawCalls) {
        const name = tc.function?.name || 'tool'
        msg.toolCalls.push({
          id: tc.id || name,
          name,
          args: tc.function?.arguments ? { raw: tc.function.arguments } : {},
        })
        toolHistory.value.unshift({
          id: tc.id || `${name}-${Date.now()}`,
          name,
          status: 'success',
          time: new Date().toLocaleTimeString(),
        })
      }
    }
    scrollToBottom()
    return
  }

  if (type === 'stream_chunk' || type === 'tool_result_chunk') {
    msg.content += content
    scrollToBottom()
    return
  }

  if (type === 'action') {
    if (event.action_type === 'interrupted') {
      pendingActions.value.push({
        id: `hitl-${Date.now()}`,
        label: (event.action_type as string) || 'approval',
        detail: content,
      })
    }
    if (event.action_type === 'exit') {
      msg.streaming = false
    }
    return
  }

  if (type === 'error') {
    msg.content += `\n\nError: ${event.error}`
    scrollToBottom()
  }
}

async function approveAction(id: string) {
  const action = pendingActions.value.find((a) => a.id === id)
  if (!action) return
  pendingActions.value = pendingActions.value.filter((a) => a.id !== id)
  await sendMessage(`${t('admin.ai.approve')}: ${action.label}`)
}

async function rejectAction(id: string) {
  pendingActions.value = pendingActions.value.filter((a) => a.id !== id)
}

onMounted(() => {
  inputEl.value?.focus()
  loadOrderOptions()
})
</script>

<style scoped>
.ai-dashboard {
  @apply max-w-7xl mx-auto;
}
</style>
