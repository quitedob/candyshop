const fs = require('fs');
const path = 'e:/go/website/frontend/i18n';

const allowedExact = new Set([
  'WhatsApp', 'OEM', 'MOQ', 'XLSX', 'Webhook', 'Webhooks', 'AI', 'CandyPro', 'B2B', 'FOB',
  'DHL', 'FedEx', 'UPS', 'USPS', 'EMS', 'Halal', 'URL', 'JSON', 'HTTP', 'IP', 'SKU', 'GTIN',
  'GMO', 'Non-GMO', 'Vegan', 'Kosher', 'Incoterms', '3PL', 'Quick ODM', 'Full OEM', 'Agent',
  'B2B Coordinator', 'Plan-Execute-Replan', 'Trade Assistant', 'Product Recommend', 'AI Search',
  'Facebook', 'LinkedIn', 'Instagram', 'YouTube', 'Twitter', 'JavaScript', 'Cookie', 'TBD',
  'USD', 'global', 'RFM', 'RCEP', 'FORM E', 'CBM', 'EN', 'ID', 'MS', 'Slug', 'Proforma Invoice',
  'Commercial Invoice', 'Bill of Lading', 'Sales Contract', 'Packing List', 'Certificate of Origin',
  'Health Certificate', 'INVOICE', 'PROFORMA INVOICE', 'TAX INVOICE', 'CREDIT NOTE', 'PI No.',
  'CI No.', 'B/L No.', 'Contract No.', 'PL No.', 'Cert No.', 'Analyze Inquiry', 'Generate Quotation',
  'Hook ID', 'DOCX', 'MD', 'Ctrl+K', 'Mac: ⌘K', 'Diff', 'Cursor', 'N/A', '—', '0',
  '+86 123 4567 890', 'CandyPro OEM', 'B2B Candy Manufacturing', 'HACCP', 'BRC', 'English',
  'Bahasa Indonesia', 'Bahasa Melayu', 'Tiếng Việt', 'العربية', '日本語', 'ไทย', 'g', 'kg', 'mm',
  'kJ', 'kcal', 'aʷ', 'HS Code', 'GMO-Free Certified', 'GMO-Free 인증', 'Chatbot (optional order)',
  'Mode_b2b' // typo guard
]);

function flatten(obj, prefix = '') {
  const out = {};
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === 'object' && !Array.isArray(v)) Object.assign(out, flatten(v, key));
    else out[key] = String(v);
  }
  return out;
}

function compare(file) {
  const en = JSON.parse(fs.readFileSync(`${path}/en/${file}`, 'utf8'));
  const ko = JSON.parse(fs.readFileSync(`${path}/ko/${file}`, 'utf8'));
  const ef = flatten(en);
  const kf = flatten(ko);
  return Object.keys(ef)
    .filter((k) => k in kf && ef[k] === kf[k] && !allowedExact.has(ef[k]))
    .map((k) => ({ k, v: ef[k] }));
}

for (const f of ['admin.json', 'common.json']) {
  const m = compare(f);
  console.log(`=== ${f} ${m.length} ===`);
  m.forEach((x) => console.log(`${x.k}\t${JSON.stringify(x.v)}`));
}
