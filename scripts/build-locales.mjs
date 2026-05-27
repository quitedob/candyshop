#!/usr/bin/env node
/**
 * Build ar/th/vi locale files from en structure using translation dictionaries.
 * Dictionaries: scripts/i18n-dict/{ar,th,vi}.json  (en string -> localized string)
 */
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.join(__dirname, '..', 'frontend', 'i18n');
const DICT_DIR = path.join(__dirname, 'i18n-dict');
const LOCALES = ['ar', 'th', 'vi'];
const FILES = ['admin.json', 'common.json'];

const GLOSSARY = new Set([
  'MOQ', 'SKU', 'OEM', 'AI', 'XLSX', 'DOCX', 'URL', 'HTTP', 'JSON', 'JWT',
  'WhatsApp', 'DHL', 'FedEx', 'UPS', 'USPS', 'EMS', 'FOB', 'USD', 'COGS',
  'GTIN', 'HS', 'Ctrl', 'Mac', 'Webhook', 'Webhooks', 'Hooks', '3PL',
  'RCEP', 'CBM', 'CandyPro', 'Facebook', 'LinkedIn', 'Instagram', 'YouTube',
  'Twitter', 'HACCP', 'ISO22000', 'BRC', 'HALAL', 'GMO', 'Non-GMO', 'N/A',
  'TBD', 'Proforma', 'Incoterms', 'B2B', 'KYB', 'RFM', 'Slug', 'POST', 'GET',
  'Stripe', 'PayPal', 'Chatbot', 'Agent', 'Markdown', 'Cursor', '3M', '6M', '12M',
  'EN', 'VI', 'TH', 'MS', 'ID', 'JA', 'KO', 'PI', 'CI', 'SC', 'B/L', 'PL', 'ETA',
  'ISO', 'CSV', 'YAML', 'ODM', 'Plan-Execute-Replan',
]);

function loadJson(p) {
  return JSON.parse(fs.readFileSync(p, 'utf8'));
}

function saveJson(p, data) {
  fs.writeFileSync(p, JSON.stringify(data, null, 2) + '\n', 'utf8');
}

function flatten(obj, prefix = '') {
  const out = {};
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === 'object' && !Array.isArray(v)) Object.assign(out, flatten(v, key));
    else out[key] = v;
  }
  return out;
}

function unflatten(flat) {
  const root = {};
  for (const [path, value] of Object.entries(flat)) {
    const parts = path.split('.');
    let cur = root;
    for (const part of parts.slice(0, -1)) cur = cur[part] ??= {};
    cur[parts[parts.length - 1]] = value;
  }
  return root;
}

function protect(text) {
  const tokens = [];
  let out = text;
  const phRe = /\{[a-zA-Z0-9_]+\}/g;
  out = out.replace(phRe, (m) => {
    const t = `⟦P${tokens.length}⟧`;
    tokens.push([t, m]);
    return t;
  });
  for (const term of [...GLOSSARY].sort((a, b) => b.length - a.length)) {
    if (out.includes(term)) {
      const t = `⟦K${tokens.length}⟧`;
      tokens.push([t, term]);
      out = out.split(term).join(t);
    }
  }
  return { out, tokens };
}

function restore(text, tokens) {
  let r = text;
  for (const [t, orig] of tokens) r = r.split(t).join(orig);
  return r;
}

function translatePhrase(text, dict, zhDict, zhText) {
  if (!text || typeof text !== 'string') return text;
  if (dict[text]) return dict[text];
  if (GLOSSARY.has(text)) return text;

  const { out, tokens } = protect(text);
  if (dict[out]) return restore(dict[out], tokens);

  // Try zh-based lookup
  if (zhText && zhDict[zhText]) return zhDict[zhText];

  return dict[text] ?? text;
}

function buildTree(enTree, dict, zhFlat, enFlat, prefix = '') {
  const out = {};
  for (const [k, v] of Object.entries(enTree)) {
    const keyPath = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === 'object' && !Array.isArray(v)) {
      out[k] = buildTree(v, dict, zhFlat, enFlat, keyPath);
    } else {
      const zhText = zhFlat[keyPath];
      out[k] = translatePhrase(v, dict, {}, zhText);
    }
  }
  return out;
}

function loadDicts() {
  const dicts = {};
  for (const loc of LOCALES) {
    const p = path.join(DICT_DIR, `${loc}.json`);
    dicts[loc] = fs.existsSync(p) ? loadJson(p) : {};
  }
  return dicts;
}

function mergeCache(dicts) {
  const cachePath = path.join(__dirname, 'i18n-translation-cache.json');
  if (!fs.existsSync(cachePath)) return;
  const cache = loadJson(cachePath);
  for (const loc of LOCALES) {
    for (const [en, val] of Object.entries(cache[loc] || {})) {
      if (val && val !== en && !val.includes('__KEEP') && !val.includes('ZZKEEP')) {
        dicts[loc][en] = val;
      }
    }
  }
}

function main() {
  const dicts = loadDicts();
  mergeCache(dicts);

  const zhData = Object.fromEntries(FILES.map((f) => [f, loadJson(path.join(ROOT, 'zh', f))]));
  const enData = Object.fromEntries(FILES.map((f) => [f, loadJson(path.join(ROOT, 'en', f))]));

  for (const loc of LOCALES) {
    for (const file of FILES) {
      const zhFlat = flatten(zhData[file]);
      const enFlat = flatten(enData[file]);
      const tree = buildTree(enData[file], dicts[loc], zhFlat, enFlat);
      saveJson(path.join(ROOT, loc, file), tree);
      console.log(`Wrote ${loc}/${file}`);
    }
    saveJson(path.join(DICT_DIR, `${loc}.json`), dicts[loc]);
  }
}

main();
