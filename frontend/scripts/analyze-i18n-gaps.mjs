import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const i18nDir = path.join(__dirname, '..', 'i18n')

function flatten(obj, prefix = '', res = {}) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k
    if (v && typeof v === 'object' && !Array.isArray(v)) flatten(v, key, res)
    else res[key] = v
  }
  return res
}

function looksEnglish(v) {
  if (typeof v !== 'string') return false
  if (!/[A-Za-z]/.test(v)) return false
  // Allow common technical tokens
  if (/^(MOQ|OEM|AI|XLSX|URL|Slug|SKU|HTTP|JSON|WhatsApp|Incoterms|3PL|COGS|Halal|Kosher|Vegan|GMO|ODM|B2B|Webhook|Webhooks|Hooks|DOCX|USD|Ctrl\+K|Mac|⌘K|PI|CI|SC|B\/L|PL|CBM|DHL|FedEx|UPS|USPS|EMS|RCEP|Facebook|LinkedIn|Instagram|YouTube|Twitter|ID|EN|KO|JA|N\/A|TBD|g|kg|mm|kJ|kcal|aʷ)$/i.test(v.trim())) return false
  // Has 3+ letter English word
  return /\b[A-Za-z]{3,}\b/.test(v)
}

for (const file of ['admin', 'common']) {
  const en = JSON.parse(fs.readFileSync(path.join(i18nDir, 'en', `${file}.json`), 'utf8'))
  for (const locale of ['ja', 'ko']) {
    const loc = JSON.parse(fs.readFileSync(path.join(i18nDir, locale, `${file}.json`), 'utf8'))
    const root = file === 'admin' ? 'admin' : null
    const fe = flatten(root ? en.admin : en)
    const fl = flatten(root ? loc.admin : loc)
    const missing = Object.keys(fe).filter((k) => !(k in fl))
    const english = Object.entries(fl).filter(([, v]) => looksEnglish(v)).map(([k]) => k)
    console.log(`\n=== ${locale}/${file}.json ===`)
    console.log('total keys en:', Object.keys(fe).length)
    console.log('total keys loc:', Object.keys(fl).length)
    console.log('missing:', missing.length)
    if (missing.length) console.log('missing sample:', missing.slice(0, 15).join(', '))
    console.log('english-like:', english.length)
    if (english.length) console.log('english sample:', english.slice(0, 30).join(', '))
  }
}
