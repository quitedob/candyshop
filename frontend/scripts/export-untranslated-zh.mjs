#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const I18N = path.join(__dirname, '..', 'i18n')

const TECHNICAL_ONLY = /^(MOQ|OEM|AI|XLSX|URL|Slug|SKU|HTTP|JSON|WhatsApp|Incoterms|3PL|COGS|Halal|DOCX|USD|Ctrl\+K|Mac|⌘K|PI|CI|SC|B\/L|PL|CBM|DHL|FedEx|UPS|USPS|EMS|RCEP|Facebook|LinkedIn|Instagram|YouTube|Twitter|ID|EN|KO|JA|N\/A|TBD|g|kg|mm|kJ|kcal|aʷ|Webhook|Webhooks|Hooks|B2B|ODM|GMO|Kosher|Vegan|FORM E|CandyPro|Admin|Super Admin|GTIN|HS|HACCP|ISO22000|BRC|HALAL|KYB|MD|FOB|Trade Agent|translate_content|Chatbot|Agent|Diff|Accept|Reject|Enter|ISO|RFM)$/i

function looksEnglish(v) {
  if (typeof v !== 'string' || !/[A-Za-z]/.test(v)) return false
  if (TECHNICAL_ONLY.test(v.trim())) return false
  return /\b[A-Za-z]{3,}\b/.test(v)
}

function flatten(obj, prefix = '', out = {}) {
  for (const [key, value] of Object.entries(obj)) {
    const pathKey = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) flatten(value, pathKey, out)
    else out[pathKey] = value
  }
  return out
}

for (const ns of ['admin', 'common']) {
  for (const locale of ['ja', 'ko']) {
    const en = JSON.parse(fs.readFileSync(path.join(I18N, 'en', `${ns}.json`), 'utf8'))
    const zh = JSON.parse(fs.readFileSync(path.join(I18N, 'zh', `${ns}.json`), 'utf8'))
    const loc = JSON.parse(fs.readFileSync(path.join(I18N, locale, `${ns}.json`), 'utf8'))
    const fe = flatten(en)
    const fz = flatten(zh)
    const fl = flatten(loc)
    const zhNeed = new Map()
    for (const [k, enVal] of Object.entries(fe)) {
      const locVal = fl[k]
      if (!looksEnglish(locVal)) continue
      const zhVal = fz[k]
      if (typeof zhVal === 'string') zhNeed.set(zhVal, { key: k, en: enVal, loc: locVal })
    }
    const outPath = path.join(__dirname, `_untranslated-${locale}-${ns}.json`)
    fs.writeFileSync(outPath, JSON.stringify(Object.fromEntries(zhNeed), null, 2))
    console.log(`${locale}/${ns}: ${zhNeed.size} unique zh strings -> ${outPath}`)
  }
}
