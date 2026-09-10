import { brotliDecompressSync, gunzipSync } from 'node:zlib'
import { getResponseHeader, setResponseHeader } from 'h3'
import { withInlineScriptHashes } from '../utils/content-security-policy'

export default defineNitroPlugin((nitroApp) => {
  // render:response runs after Nuxt has emitted its route/locale-specific payload.
  nitroApp.hooks.hook('render:response', (response, { event }) => {
    if (typeof response.body !== 'string' || !String(response.headers?.['content-type']).includes('text/html')) return
    response.headers ||= {}
    response.headers['content-security-policy'] = withInlineScriptHashes(
      useRuntimeConfig(event).contentSecurityPolicy, response.body,
    )
  })

  // Prerendered files and SWR cache hits bypass the renderer. Apply the same
  // policy to their actual bytes, without rewriting assets or caching a nonce.
  nitroApp.hooks.hook('beforeResponse', (event, response) => {
    if (!String(getResponseHeader(event, 'content-type')).includes('text/html')) return
    if (typeof response.body !== 'string' && !Buffer.isBuffer(response.body)) return
    let contents = response.body
    if (!contents.length) return // Preserve the cached policy on HEAD/304.
    const encoding = getResponseHeader(event, 'content-encoding')
    if (Buffer.isBuffer(contents) && encoding === 'gzip') contents = gunzipSync(contents)
    if (Buffer.isBuffer(contents) && encoding === 'br') contents = brotliDecompressSync(contents)
    setResponseHeader(event, 'content-security-policy', withInlineScriptHashes(
      useRuntimeConfig(event).contentSecurityPolicy, Buffer.isBuffer(contents) ? contents.toString('utf8') : contents,
    ))
  })
})
