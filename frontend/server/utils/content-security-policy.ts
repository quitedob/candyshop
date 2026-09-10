import { createHash } from 'node:crypto'

/** Hash the final inline script text, including Nuxt payloads and module scripts. */
export function withInlineScriptHashes(policy: string, html: string): string {
  const hashes = new Set<string>()
  // Quoted attributes can contain >. Preserve script whitespace; HTML normalizes
  // line endings before CSP compares the script's text with its allowed hash.
  const scriptPattern = /<script\b(?:"[^"]*"|'[^']*'|[^'">])*?>([\s\S]*?)<\/script\s*>/gi
  for (const script of html.matchAll(scriptPattern)) {
    const contents = script[1]!.replace(/\r\n?/g, '\n')
    if (contents) {
      hashes.add(`'sha256-${createHash('sha256').update(contents, 'utf8').digest('base64')}'`)
    }
  }
  return policy.replace(/(^|;)\s*script-src\s+[^;]*/i, (_directive, separator: string) =>
    `${separator} script-src 'self' ${[...hashes].join(' ')}`.trimEnd(),
  )
}
