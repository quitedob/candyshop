#!/usr/bin/env node
/** Export zh strings still producing English in ja admin */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const I18N = path.join(__dirname, '..', 'i18n')

const TECH = /\b(CandyPro|OEM|XLSX|RFM|CRUD|USD|MOQ|Slug|SKU|AI|MD|WhatsApp|JavaScript|Facebook|LinkedIn|FOB|global|Province|Gluten|Sugar|Organic|Fat|Thumbnail|Images|Flavors|Shapes|Weight|Length|Width|Height|Pack|Carton|Piece|Offer|Convert|Generate|Packaging|Net|Gross|Halal|HACCP|BRC|Cookie|English|Bahasa|https|example|Comma|Incoterms|DHL|FedEx|Webhook|Hooks|B2B|ODM|GTIN|HS|KYB|Admin|Agent|Diff|Proforma|Commercial|Certificate|Invoice|Credit|INVOICE|3PL|Marketplace|Webstore|Distributor|Coordinator|Chatbot|Stream|Ctrl|Mac|ISO|PI|CI|SC|PL|CBM|RCEP|FORM|Kosher|Vegan|N\/A|TBD|EMS|UPS|USPS|Instagram|YouTube|Twitter|LinkedIn|kcal|kJ|mm|kg|g|aʷ|flow-wrap|foil|jar|GMO|E-numbers|Solids|Pallet|Arabic|Korean|Japanese|Traditional|Chinese|Vietnamese|Thai|Indonesian|Malay|Sweetener|Carbohydrates|Protein|Fiber|Additives|Cocoa|Milk|Accept|Reject|Counter-offers|Are you sure|No files|units|cases|layer|layers|display box|GMO-Free|Certified)\b/i

function flatten(obj, prefix = '', out = {}) {
  for (const [key, value] of Object.entries(obj)) {
    const pathKey = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) flatten(value, pathKey, out)
    else out[pathKey] = value
  }
  return out
}

const en = flatten(JSON.parse(fs.readFileSync(path.join(I18N, 'en', 'admin.json'), 'utf8')))
const zh = flatten(JSON.parse(fs.readFileSync(path.join(I18N, 'zh', 'admin.json'), 'utf8')))
const ja = flatten(JSON.parse(fs.readFileSync(path.join(I18N, 'ja', 'admin.json'), 'utf8')))

const need = {}
for (const [k, v] of Object.entries(ja)) {
  if (typeof v !== 'string' || !/\b[A-Za-z]{4,}\b/.test(v)) continue
  if (TECH.test(v)) continue
  const z = zh[k]
  if (typeof z === 'string') need[z] = { key: k, en: en[k], ja: v }
}
fs.writeFileSync(path.join(__dirname, '_need-zh-admin.json'), JSON.stringify(need, null, 2) + '\n')
console.log(Object.keys(need).length)
