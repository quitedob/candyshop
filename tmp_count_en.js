const fs = require('fs');
function findEnglish(path) {
  const data = JSON.parse(fs.readFileSync(path, 'utf8'));
  const enPath = path.replace('/ar/', '/en/').replace('\\ar\\', '\\en\\');
  const en = JSON.parse(fs.readFileSync(enPath, 'utf8'));
  const results = [];
  function walk(a, e, p = '') {
    if (typeof a === 'string' && typeof e === 'string') {
      if (a === e) results.push({ p, v: a });
      return;
    }
    if (Array.isArray(a) && Array.isArray(e)) {
      a.forEach((x, i) => walk(x, e[i], p + '[' + i + ']'));
      return;
    }
    if (a && e && typeof a === 'object') {
      for (const k of Object.keys(a)) walk(a[k], e[k], p ? p + '.' + k : k);
    }
  }
  walk(data, en);
  return results;
}
for (const f of ['frontend/i18n/ar/admin.json', 'frontend/i18n/ar/common.json']) {
  const r = findEnglish('e:/go/website/' + f);
  console.log(f + ': ' + r.length);
  r.forEach((x) => console.log('  ' + x.p));
}
