import { readFileSync, writeFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = import.meta.dirname
const files = [
  resolve(root, '../node_modules/nuxt-og-image/dist/module.mjs'),
  resolve(root, '../node_modules/nuxt-og-image/dist/module.cjs'),
]

const replacements = [
  ['unenv/runtime/mock/empty', 'unenv/mock/empty'],
  ['unenv/runtime/mock/proxy-cjs', 'unenv/mock/proxy-cjs'],
]

for (const file of files) {
  try {
    let content = readFileSync(file, 'utf8')
    let changed = false
    for (const [from, to] of replacements) {
      if (content.includes(from)) {
        content = content.replaceAll(from, to)
        changed = true
      }
    }
    if (changed) {
      writeFileSync(file, content, 'utf8')
      console.log(`[patch-nuxt-og-image] Patched: ${file}`)
    }
  } catch (e) {
    if (e.code === 'ENOENT') {
      // File doesn't exist — skip silently
      continue
    }
    console.warn(`[patch-nuxt-og-image] Warning: could not patch ${file}: ${e.message}`)
  }
}
