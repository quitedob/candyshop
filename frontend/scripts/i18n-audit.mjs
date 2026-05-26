#!/usr/bin/env node
/**
 * Audit admin + common i18n files against en/zh baselines.
 * Reports missing keys and values that still match English source (untranslated).
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.join(__dirname, '..', 'i18n')
const locales = ['en', 'zh', 'ko', 'ja', 'ar', 'th', 'vi', 'id', 'ms']
const files = ['admin.json', 'common.json']
const baselines = ['en', 'zh']

function flatten(obj, prefix = '') {
  const out = {}
  if (!obj || typeof obj !== 'object' || Array.isArray(obj)) return out
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k
    if (v && typeof v === 'object' && !Array.isArray(v)) {
      Object.assign(out, flatten(v, key))
    } else {
      out[key] = v
    }
  }
  return out
}

function loadFlat(locale, file) {
  const p = path.join(root, locale, file)
  if (!fs.existsSync(p)) return {}
  return flatten(JSON.parse(fs.readFileSync(p, 'utf8')))
}

let exitCode = 0

for (const file of files) {
  const enFlat = loadFlat('en', file)
  const zhFlat = loadFlat('zh', file)
  const allKeys = new Set([...Object.keys(enFlat), ...Object.keys(zhFlat)])

  console.log(`\n=== ${file} ===`)
  console.log(`Baseline keys: ${allKeys.size} (en: ${Object.keys(enFlat).length}, zh: ${Object.keys(zhFlat).length})`)

  for (const locale of locales) {
    if (locale === 'en') continue
    const flat = loadFlat(locale, file)
    const missing = [...allKeys].filter(k => !(k in flat))
    const untranslated = [...allKeys].filter(k => {
      if (!(k in flat)) return false
      const val = flat[k]
      if (typeof val !== 'string') return false
      if (locale === 'zh') return false
      return val === enFlat[k] && enFlat[k] && /[A-Za-z]/.test(enFlat[k])
    })

    if (missing.length || untranslated.length) {
      console.log(`\n[${locale}] missing: ${missing.length}, likely untranslated (EN copy): ${untranslated.length}`)
      if (missing.length) {
        console.log('  Missing sample:', missing.slice(0, 15).join(', '), missing.length > 15 ? '...' : '')
      }
      if (untranslated.length && locale !== 'en') {
        console.log('  Untranslated sample:', untranslated.slice(0, 10).join(', '), untranslated.length > 10 ? '...' : '')
      }
      exitCode = 1
    } else {
      console.log(`[${locale}] OK`)
    }
  }
}

// Nav key leak check (SNAKE_CASE raw keys)
const adminEn = loadFlat('en', 'admin.json')
const navKeys = Object.keys(adminEn).filter(k => k.startsWith('nav.'))
console.log('\n=== admin.nav keys ===')
for (const locale of locales.filter(l => l !== 'en')) {
  const flat = loadFlat(locale, 'admin.json')
  const missingNav = navKeys.filter(k => !(k in flat))
  if (missingNav.length) {
    console.log(`[${locale}] missing nav: ${missingNav.join(', ')}`)
    exitCode = 1
  }
}

process.exit(exitCode)
