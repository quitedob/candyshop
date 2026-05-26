#!/usr/bin/env node
/**
 * Deep-merge missing keys from i18n/en/admin.json + common.json into other locales.
 * Preserves existing translations; only adds keys missing in target files.
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.join(__dirname, '..', 'i18n')
const targets = ['ko', 'ja', 'ar', 'th', 'vi', 'id', 'ms']
const files = ['admin.json', 'common.json']

function deepMerge(base, fill) {
  if (!fill || typeof fill !== 'object' || Array.isArray(fill)) return base ?? fill
  const out = { ...(base && typeof base === 'object' && !Array.isArray(base) ? base : {}) }
  for (const [key, value] of Object.entries(fill)) {
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      out[key] = deepMerge(out[key], value)
    } else if (!(key in out)) {
      out[key] = value
    }
  }
  return out
}

for (const file of files) {
  const source = JSON.parse(fs.readFileSync(path.join(root, 'en', file), 'utf8'))
  for (const locale of targets) {
    const filePath = path.join(root, locale, file)
    const current = JSON.parse(fs.readFileSync(filePath, 'utf8'))
    const merged = deepMerge(current, source)
    fs.writeFileSync(filePath, JSON.stringify(merged, null, 2) + '\n', 'utf8')
    console.log(`Synced ${locale}/${file}`)
  }
}
