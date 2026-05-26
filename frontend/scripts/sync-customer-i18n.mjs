#!/usr/bin/env node
/**
 * Deep-merge missing keys from i18n/en/customer.json into other locale customer.json files.
 * Existing translations are preserved; only missing keys are filled from English.
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.join(__dirname, '..', 'i18n')
const source = JSON.parse(fs.readFileSync(path.join(root, 'en', 'customer.json'), 'utf8'))
const targets = ['ko', 'ja', 'vi', 'th', 'id', 'ms', 'ar']

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

for (const locale of targets) {
  const file = path.join(root, locale, 'customer.json')
  const current = JSON.parse(fs.readFileSync(file, 'utf8'))
  const merged = deepMerge(current, source)
  fs.writeFileSync(file, JSON.stringify(merged, null, 2) + '\n', 'utf8')
  console.log(`Synced ${locale}/customer.json`)
}
