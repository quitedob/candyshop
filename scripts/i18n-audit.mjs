import fs from 'fs';
import path from 'path';

const root = path.resolve('frontend/i18n');

function flat(obj, prefix = '', out = {}) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === 'object' && !Array.isArray(v)) flat(v, key, out);
    else out[key] = v;
  }
  return out;
}

function isMostlyEnglish(s) {
  if (typeof s !== 'string') return false;
  if (s.length <= 2) return false;
  if (/^[\d\s.,\-+×→/:{}$#%]+$/.test(s)) return false;
  if (/^(MOQ|SKU|ID|URL|HTTP|JSON|XLSX|DOCX|PI|CI|SC|B\/L|CBM|GTIN|HS|OEM|AI|Ctrl|Mac|WhatsApp|Facebook|LinkedIn|Instagram|YouTube|Twitter|DHL|FedEx|UPS|USPS|EMS|FOB|USD|RCEP|FORM E|3PL|HACCP|ISO|BRC|HALAL|GMO|Non-GMO|Proforma|Incoterms|Webhook|Hooks|Slug|Temperature|Temperature|N\/A|TBD|EMS)$/i.test(s.trim())) return false;
  const latin = (s.match(/[A-Za-z]/g) || []).length;
  return latin / s.length > 0.55;
}

for (const locale of ['ar', 'th', 'vi']) {
  for (const file of ['admin', 'common']) {
    const en = JSON.parse(fs.readFileSync(path.join(root, 'en', `${file}.json`), 'utf8'));
    const loc = JSON.parse(fs.readFileSync(path.join(root, locale, `${file}.json`), 'utf8'));
    const fe = flat(en);
    const fl = flat(loc);
    let missing = 0;
    let same = 0;
    let englishish = 0;
    for (const k of Object.keys(fe)) {
      if (!(k in fl)) missing++;
      else if (fl[k] === fe[k]) same++;
      else if (isMostlyEnglish(fl[k])) englishish++;
    }
    console.log(`${locale}/${file}: total=${Object.keys(fe).length} missing=${missing} same=${same} englishish=${englishish}`);
  }
}
