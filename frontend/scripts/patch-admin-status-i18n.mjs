#!/usr/bin/env node
/** Patch status_options.active/inactive and ai_translate_slow across all admin.json locales. */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.join(__dirname, '..', 'i18n')
const locales = ['en', 'zh', 'ko', 'ja', 'ar', 'th', 'vi', 'id', 'ms']

const statusLabels = {
  en: { active: 'Active', inactive: 'Inactive' },
  zh: { active: '启用', inactive: '禁用' },
  ko: { active: '활성', inactive: '비활성' },
  ja: { active: '有効', inactive: '無効' },
  ar: { active: 'نشط', inactive: 'غير نشط' },
  th: { active: 'ใช้งาน', inactive: 'ไม่ใช้งาน' },
  vi: { active: 'Hoạt động', inactive: 'Không hoạt động' },
  id: { active: 'Aktif', inactive: 'Nonaktif' },
  ms: { active: 'Aktif', inactive: 'Tidak aktif' },
}

const aiSlow = {
  en: 'Translation is taking longer than usual. Please wait or try again later.',
  zh: '翻译耗时较长，请稍候或稍后重试。',
  ko: '번역에 시간이 더 걸리고 있습니다. 잠시 기다리거나 나중에 다시 시도하세요.',
  ja: '翻訳に通常より時間がかかっています。しばらくお待ちいただくか、後でもう一度お試しください。',
  ar: 'تستغرق الترجمة وقتًا أطول من المعتاد. يرجى الانتظار أو المحاولة لاحقًا.',
  th: 'การแปลใช้เวลานานกว่าปกติ โปรดรอหรือลองใหม่ภายหลัง',
  vi: 'Dịch đang mất nhiều thời gian hơn bình thường. Vui lòng đợi hoặc thử lại sau.',
  id: 'Terjemahan membutuhkan waktu lebih lama dari biasanya. Harap tunggu atau coba lagi nanti.',
  ms: 'Terjemahan mengambil masa lebih lama daripada biasa. Sila tunggu atau cuba lagi kemudian.',
}

for (const locale of locales) {
  const file = path.join(root, locale, 'admin.json')
  const data = JSON.parse(fs.readFileSync(file, 'utf8'))
  if (!data.admin.status_options) data.admin.status_options = {}
  data.admin.status_options.active = statusLabels[locale].active
  data.admin.status_options.inactive = statusLabels[locale].inactive
  if (!data.admin.products) data.admin.products = {}
  data.admin.products.ai_translate_slow = aiSlow[locale]
  fs.writeFileSync(file, JSON.stringify(data, null, 2) + '\n', 'utf8')
  console.log(`Patched ${locale}/admin.json`)
}
