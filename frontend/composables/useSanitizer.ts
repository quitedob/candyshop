/**
 * HTML sanitizer composable.
 *
 * Strips XSS vectors (script, event handlers, javascript:/data: URLs) from
 * user-generated or AI-generated HTML before it is rendered via v-html.
 *
 * Environment handling (SSR safety — finding H16):
 *  - Client: DOMPurify (full DOM support).
 *  - Server: DOMPurify without a DOM returns an unsupported factory whose
 *    `sanitize()` returns the input UNCHANGED (the "isSupported" guard in
 *    dompurify/dist/purify.cjs.js). Relying on it alone would ship raw stored HTML
 *    to the first paint on SSR pages (e.g. /blog/[slug], which is server-rendered).
 *    We therefore use `sanitize-html` (isomorphic, Node-safe) on the server, mirroring
 *    the same allowlist so the SSR payload never contains executable HTML.
 *
 * Note: `sanitize-html` is bundled on both server and client here (a shared composable
 * is auto-imported in both bundles; Nuxt cannot split composables per environment in
 * this version). On the client the DOMPurify branch wins, so `sanitize-html` is never
 * invoked there — it is dead, isomorphic code (no security/correctness impact).
 */

import DOMPurify, { type Config as DOMPurifyConfig } from 'dompurify'
import sanitizeHtml from 'sanitize-html'

// Shared allowlist: only safe tags/attributes are kept on BOTH server and client.
const ALLOWED_TAGS = [
  'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
  'p', 'br', 'hr',
  'ul', 'ol', 'li',
  'blockquote', 'pre', 'code',
  'strong', 'b', 'em', 'i', 'u', 's', 'del', 'ins',
  'a', 'img',
  'table', 'thead', 'tbody', 'tfoot', 'tr', 'th', 'td',
  'div', 'span',
  'sub', 'sup', 'small', 'mark',
]

const ALLOWED_ATTR = [
  'href', 'target', 'rel',
  'src', 'alt', 'width', 'height', 'loading',
  'class', 'id',
  'colspan', 'rowspan',
  'start', 'type',
]

// Client-side config (DOMPurify)
const purifyConfig: DOMPurifyConfig = {
  ALLOWED_TAGS,
  ALLOWED_ATTR,
  // http(s), mailto, tel plus relative/protocol-relative URLs only; blocks javascript:, data:, ftp:, etc.
  ALLOWED_URI_REGEXP: /^(?:(?:https?|mailto|tel):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i,
  ALLOW_DATA_ATTR: false,
}

// Server-side config (sanitize-html) — mirrors the DOMPurify allowlist above.
const serverSanitizeConfig = {
  allowedTags: [...ALLOWED_TAGS],
  // sanitize-html maps attributes per tag; '*' applies the same global allowlist
  // that DOMPurify applies to every tag.
  allowedAttributes: { '*': [...ALLOWED_ATTR] },
  allowedSchemes: ['http', 'https', 'mailto', 'tel'],
  allowProtocolRelative: true,
}

const stripAllServerConfig = {
  allowedTags: [],
  allowedAttributes: {},
}

export function useSanitizer() {
  const sanitize = (html: string): string => {
    if (!html) return ''
    if (import.meta.server) {
      return sanitizeHtml(html, serverSanitizeConfig)
    }
    return DOMPurify.sanitize(html, purifyConfig) as unknown as string
  }

  // Strip all HTML, returning plain text only
  const stripHtml = (html: string): string => {
    if (!html) return ''
    if (import.meta.server) {
      return sanitizeHtml(html, stripAllServerConfig)
    }
    return DOMPurify.sanitize(html, { ALLOWED_TAGS: [], ALLOWED_ATTR: [] }) as unknown as string
  }

  return { sanitize, stripHtml }
}
