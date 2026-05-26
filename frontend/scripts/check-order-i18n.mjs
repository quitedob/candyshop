import fs from 'fs'

function get(obj, keyPath) {
  return keyPath.split('.').reduce((o, k) => o?.[k], obj)
}

const locales = ['en', 'zh', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms']
const vue = fs.readFileSync('pages/customer/orders/[id].vue', 'utf8')
const usedKeys = [...new Set([...vue.matchAll(/t\('([^']+)'/g)].map((m) => m[1]))]

const files = Object.fromEntries(
  locales.map((loc) => [loc, JSON.parse(fs.readFileSync(`i18n/${loc}/customer.json`, 'utf8'))]),
)
const commons = Object.fromEntries(
  locales.map((loc) => [loc, JSON.parse(fs.readFileSync(`i18n/${loc}/common.json`, 'utf8'))]),
)

const enOrderKeys = Object.keys(files.en.customer.orders)

console.log('=== Untranslated (same as EN) or missing in customer.orders ===')
for (const loc of locales) {
  if (loc === 'en') continue
  const bad = []
  for (const k of enOrderKeys) {
    const v = files[loc].customer.orders?.[k]
    const ev = files.en.customer.orders[k]
    if (v === undefined) bad.push(`${k}:MISSING`)
    else if (v === ev && typeof ev === 'string') bad.push(k)
  }
  if (bad.length) console.log(loc, bad.length, bad.join(', '))
}

console.log('\n=== Page keys missing in zh ===')
for (const k of usedKeys) {
  let v
  if (k.startsWith('customer.')) v = get(files.zh, k)
  else if (k.startsWith('enum.') || k.startsWith('display.')) v = get(commons.zh, k)
  else v = get(commons.zh, k) ?? get(files.zh, k)
  if (v === undefined) console.log('MISSING', k)
}

console.log('\n=== Missing upload/payment keys in en orders ===')
const extra = ['upload_payment_proof', 'payment_proof', 'uploading', 'upload', 'payment_uploaded', 'cancel_confirm']
for (const k of extra) {
  console.log(k, files.en.customer.orders[k] ?? 'MISSING')
}
