#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const I18N = path.join(__dirname, '..', 'i18n')

const TECH = new RegExp(
  String.raw`\b(CandyPro|OEM|XLSX|RFM|CRUD|USD|MOQ|Slug|SKU|AI|MD|WhatsApp|JavaScript|Facebook|LinkedIn|FOB|global|Province|Gluten|Sugar|Organic|Fat|Thumbnail|Images|Flavors|Shapes|Weight|Length|Width|Height|Pack|Carton|Piece|Offer|Convert|Generate|Packaging|Net|Gross|Halal|HACCP|BRC|Cookie|English|Bahasa|https|example|Comma|Incoterms|DHL|FedEx|Webhook|Hooks|B2B|ODM|GTIN|HS|KYB|Admin|Agent|Diff|Proforma|Commercial|Certificate|Invoice|Credit|INVOICE|3PL|Marketplace|Webstore|Distributor|Coordinator|Chatbot|Stream|Ctrl|Mac|ISO|PI|CI|SC|PL|CBM|RCEP|FORM|Kosher|Vegan|N/A|TBD|EMS|UPS|USPS|Instagram|YouTube|Twitter|LinkedIn)\b`,
  'i',
)

function flatten(obj, prefix = '', out = {}) {
  for (const [key, value] of Object.entries(obj)) {
    const pathKey = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) flatten(value, pathKey, out)
    else out[pathKey] = value
  }
  return out
}

for (const locale of ['ja', 'ko']) {
  for (const ns of ['admin', 'common']) {
    const loc = flatten(JSON.parse(fs.readFileSync(path.join(I18N, locale, `${ns}.json`), 'utf8')))
    const zh = flatten(JSON.parse(fs.readFileSync(path.join(I18N, 'zh', `${ns}.json`), 'utf8')))
    const bad = Object.entries(loc).filter(([k, v]) => {
      if (typeof v !== 'string' || !/\b[A-Za-z]{4,}\b/.test(v)) return false
      if (TECH.test(v)) return false
      return true
    })
    console.log(`\n${locale}/${ns}: ${bad.length} non-tech english-like`)
    for (const [k, v] of bad.slice(0, 30)) console.log(`  ${k}: ${v} (zh: ${zh[k] ?? 'n/a'})`)
  }
}
