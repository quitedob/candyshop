#!/usr/bin/env node
/**
 * Generate exact zh->ja/ko translation maps from zh reference files.
 * Run: node frontend/scripts/generate-exact-zh-map.mjs
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { PRIORITY_ZH } from './build-admin-i18n-data.mjs'
import { EXACT_ZH_EXTRA } from './build-admin-i18n-exact-zh.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const I18N = path.join(__dirname, '..', 'i18n')

function flatten(obj, prefix = '', out = {}) {
  for (const [key, value] of Object.entries(obj)) {
    const pathKey = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) flatten(value, pathKey, out)
    else out[pathKey] = value
  }
  return out
}

/** @param {string} zh @param {string} en @param {'ja'|'ko'} locale */
function translate(zh, en, locale) {
  if (PRIORITY_ZH[locale][zh]) return PRIORITY_ZH[locale][zh]
  if (EXACT_ZH_EXTRA[locale][zh]) return EXACT_ZH_EXTRA[locale][zh]
  return null
}

const zhAdmin = flatten(JSON.parse(fs.readFileSync(path.join(I18N, 'zh', 'admin.json'), 'utf8')))
const enAdmin = flatten(JSON.parse(fs.readFileSync(path.join(I18N, 'en', 'admin.json'), 'utf8')))
const zhCommon = flatten(JSON.parse(fs.readFileSync(path.join(I18N, 'zh', 'common.json'), 'utf8')))
const enCommon = flatten(JSON.parse(fs.readFileSync(path.join(I18N, 'en', 'common.json'), 'utf8')))

/** @type {Record<'ja'|'ko', Record<string, string>>} */
const out = { ja: { ...PRIORITY_ZH.ja, ...EXACT_ZH_EXTRA.ja }, ko: { ...PRIORITY_ZH.ko, ...EXACT_ZH_EXTRA.ko } }

for (const [key, zh] of Object.entries({ ...zhAdmin, ...zhCommon })) {
  if (typeof zh !== 'string') continue
  const en = (key in zhAdmin ? enAdmin : enCommon)[key]
  for (const locale of ['ja', 'ko']) {
    if (out[locale][zh]) continue
    const t = translate(zh, String(en ?? ''), locale)
    if (t) out[locale][zh] = t
  }
}

const missing = { ja: [], ko: [] }
for (const [key, zh] of Object.entries({ ...zhAdmin, ...zhCommon })) {
  if (typeof zh !== 'string') continue
  for (const locale of ['ja', 'ko']) {
    if (!out[locale][zh]) missing[locale].push({ key, zh, en: (key in zhAdmin ? enAdmin : enCommon)[key] })
  }
}

fs.writeFileSync(path.join(__dirname, 'exact-zh-map.ja.json'), JSON.stringify(out.ja, null, 2) + '\n')
fs.writeFileSync(path.join(__dirname, 'exact-zh-map.ko.json'), JSON.stringify(out.ko, null, 2) + '\n')
fs.writeFileSync(path.join(__dirname, '_missing-zh-translations.json'), JSON.stringify(missing, null, 2) + '\n')
console.log('ja mapped', Object.keys(out.ja).length, 'missing', missing.ja.length)
console.log('ko mapped', Object.keys(out.ko).length, 'missing', missing.ko.length)
