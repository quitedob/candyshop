#!/usr/bin/env node
/** One-off helper: emit priority zh->ja/ko map skeleton from zh strings. */
import fs from 'node:fs'

const zh = JSON.parse(fs.readFileSync('frontend/i18n/zh/admin.json', 'utf8'))
const en = JSON.parse(fs.readFileSync('frontend/i18n/en/admin.json', 'utf8'))
const sections = ['products', 'inventory', 'categories', 'coupons', 'shippingRates', 'taxRates']

/** @type {Record<string, {en:string, zh:string}>} */
const entries = {}

function walk(enNode, zhNode, prefix) {
  for (const key of Object.keys(enNode)) {
    const p = `${prefix}.${key}`
    const ev = enNode[key]
    const zv = zhNode?.[key]
    if (ev && typeof ev === 'object') walk(ev, zv, p)
    else entries[p] = { en: String(ev ?? ''), zh: String(zv ?? ev ?? '') }
  }
}

for (const s of sections) walk(en.admin[s], zh.admin[s], `admin.${s}`)

fs.writeFileSync('frontend/scripts/_priority-entries.json', JSON.stringify(entries, null, 2))
console.log('wrote', Object.keys(entries).length, 'entries')
