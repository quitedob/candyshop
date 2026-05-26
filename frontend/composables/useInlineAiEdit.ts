/** Cursor 式 inline edit：选区 + 指令 → 只返回替换片段 */
export type InlineAiFieldType = 'html' | 'plain' | 'title'

export interface InlineAiEditRequest {
  instruction: string
  selectedText: string
  selectedHtml?: string
  contextBefore?: string
  contextAfter?: string
  fieldType: InlineAiFieldType
  language?: string
}

export interface InlineAiEditResponse {
  replacement: string
  format: 'html' | 'plain'
}

export function clampPopoverPosition(top: number, left: number, width = 380) {
  const pad = 12
  const maxLeft = window.innerWidth - width - pad
  return {
    top: Math.max(pad, top),
    left: Math.max(pad, Math.min(left, maxLeft)),
  }
}

export function useInlineAiEdit(defaultLanguage = 'zh') {
  const api = useApi()

  const requestInlineEdit = async (payload: InlineAiEditRequest): Promise<InlineAiEditResponse> => {
    const res = await api.adminAIInlineEditContent({
      ...payload,
      language: payload.language || defaultLanguage,
    })
    if (res?.parseError) {
      throw new Error('parse_error')
    }
    return {
      replacement: res.replacement || '',
      format: res.format === 'html' ? 'html' : 'plain',
    }
  }

  return { requestInlineEdit, clampPopoverPosition }
}
