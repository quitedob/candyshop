/**
 * HTML sanitizer composable.
 * Uses DOMPurify to strip XSS vectors (script, event handlers, javascript: URLs)
 * before rendering user-generated or AI-generated HTML via v-html.
 */

import DOMPurify from 'dompurify'

// Configure once: allow only safe tags/attributes for content rendering
const purifyConfig: DOMPurify.Config = {
  ALLOWED_TAGS: [
    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
    'p', 'br', 'hr',
    'ul', 'ol', 'li',
    'blockquote', 'pre', 'code',
    'strong', 'b', 'em', 'i', 'u', 's', 'del', 'ins',
    'a', 'img',
    'table', 'thead', 'tbody', 'tfoot', 'tr', 'th', 'td',
    'div', 'span',
    'sub', 'sup', 'small', 'mark',
  ],
  ALLOWED_ATTR: [
    'href', 'target', 'rel',
    'src', 'alt', 'width', 'height', 'loading',
    'class', 'id',
    'colspan', 'rowspan',
    'start', 'type',
  ],
  ALLOWED_URI_REGEXP: /^(?:(?:https?|mailto|tel):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i,
  ALLOW_DATA_ATTR: false,
}

const purify = DOMPurify()

export function useSanitizer() {
  const sanitize = (html: string): string => {
    if (!html) return ''
    return purify.sanitize(html, purifyConfig) as string
  }

  // Strip all HTML, returning plain text only
  const stripHtml = (html: string): string => {
    if (!html) return ''
    return purify.sanitize(html, { ALLOWED_TAGS: [], ALLOWED_ATTR: [] }) as string
  }

  return { sanitize, stripHtml }
}
