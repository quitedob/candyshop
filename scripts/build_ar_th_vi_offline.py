#!/usr/bin/env python3
"""Build ar/th/vi admin.json + common.json from zh reference via Lingva mirrors."""

from __future__ import annotations

import json
import os
import re
import sys
import time
import urllib.parse
from pathlib import Path

import requests

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
ZH_MAP = Path(__file__).resolve().parent / "zh-locale-map.json"
FILES = ("admin.json", "common.json")
LOCALES = ("ar", "th", "vi")
LANG = {"ar": "ar", "th": "th", "vi": "vi"}
MIRRORS = [
    "https://lingva.ml",
    "https://translate.plausibility.cloud",
    "https://lingva.lunar.icu",
]

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Stripe",
    "PayPal", "Chatbot", "Agent", "Cursor", "Markdown", "POST", "GET",
    "3M", "6M", "12M", "EN", "VI", "TH", "MS", "ID", "JA", "KO", "PI", "CI",
    "SC", "B/L", "PL", "ETA", "ISO", "CSV", "YAML", "ODM", "Plan-Execute-Replan",
    "⌘K", "Ctrl+K", "Super Admin", "Superadmin",
}

sys.path.insert(0, str(Path(__file__).resolve().parent))
from apply_priority_i18n import TERM  # noqa: E402

MANUAL: dict[str, dict[str, str]] = {
    "ar": {
        "admin.langLabel": "العربية",
        "admin.langLabelShort": "ع",
        "admin.langTitle": "العربية",
        "admin.brand": "CandyPro Admin",
        "admin.nav.webhooks": "Webhooks",
    },
    "th": {
        "admin.langLabel": "TH",
        "admin.langLabelShort": "TH",
        "admin.langTitle": "ไทย",
        "admin.brand": "CandyPro Admin",
        "admin.nav.webhooks": "Webhooks",
        "admin.nav.hooks": "Hooks & Events",
    },
    "vi": {
        "admin.langLabel": "VI",
        "admin.langLabelShort": "VI",
        "admin.langTitle": "Tiếng Việt",
        "admin.brand": "CandyPro Admin",
        "admin.nav.webhooks": "Webhooks",
        "admin.nav.hooks": "Hook và sự kiện",
        "admin.users.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "admin.staff.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "admin.auditLog.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "admin.inventory.batch_edit_result": "Đã cập nhật {updated}, {errors} lỗi",
        "admin.products.base_price": "Giá cơ bản (USD)",
        "admin.products.field_energy_kcal": "Năng lượng (kcal)",
    },
}

SESSION = requests.Session()
SESSION.trust_env = False
for k in ("HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy"):
    os.environ.pop(k, None)

_mirror_idx = 0
HTML_ENTITY_RE = re.compile(r"&#\d+;")


def flatten(obj: dict, prefix: str = "") -> dict[str, str]:
    out: dict[str, str] = {}
    for key, value in obj.items():
        path = f"{prefix}.{key}" if prefix else key
        if isinstance(value, dict):
            out.update(flatten(value, path))
        else:
            out[path] = value
    return out


def unflatten(flat: dict[str, str]) -> dict:
    root: dict = {}
    for path, value in flat.items():
        parts = path.split(".")
        cur = root
        for part in parts[:-1]:
            cur = cur.setdefault(part, {})
        cur[parts[-1]] = value
    return root


def protect(text: str) -> tuple[str, list[tuple[str, str]]]:
    tokens: list[tuple[str, str]] = []
    out = text
    for ph in re.findall(r"\{[a-zA-Z0-9_]+\}", text):
        tok = f"ZZPH{len(tokens)}ZZ"
        tokens.append((tok, ph))
        out = out.replace(ph, tok, 1)
    for term in sorted(GLOSSARY, key=len, reverse=True):
        if term in out:
            tok = f"ZZK{len(tokens)}ZZ"
            tokens.append((tok, term))
            out = out.replace(term, tok)
    return out, tokens


def restore(text: str, tokens: list[tuple[str, str]]) -> str:
    for tok, orig in tokens:
        text = text.replace(tok, orig)
    return text


def fix_text(text: str) -> str:
    text = HTML_ENTITY_RE.sub("\n", text)
    text = re.sub(r"__KEEP\s*_?\s*\d+\s*__", "", text)
    return re.sub(r" +", " ", text).strip()


def has_cjk(text: str) -> bool:
    return bool(re.search(r"[\u4e00-\u9fff]", text))


def is_english(text: str) -> bool:
    if not text or len(text) <= 2:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&⌘]+", text):
        return False
    if text in GLOSSARY:
        return False
    latin = len(re.findall(r"[A-Za-z]", text))
    return latin / max(len(text), 1) > 0.55


def lingva(text: str, source: str, target: str) -> str:
    global _mirror_idx
    if not text or not text.strip():
        return text
    protected, tokens = protect(text)
    encoded = urllib.parse.quote(protected, safe="")
    for attempt in range(len(MIRRORS) * 3):
        host = MIRRORS[_mirror_idx % len(MIRRORS)]
        _mirror_idx += 1
        url = f"{host}/api/v1/{source}/{target}/{encoded}"
        try:
            r = SESSION.get(url, timeout=25)
            if r.status_code == 200:
                result = r.json().get("translation") or text
                return fix_text(restore(result, tokens))
            if r.status_code == 429:
                time.sleep(2 + attempt * 0.5)
                continue
        except Exception:
            time.sleep(1)
    return text


def needs_translate(s: str, zmap: dict[str, str]) -> bool:
    if not s:
        return False
    cached = zmap.get(s)
    if not cached or cached == s:
        return True
    if is_english(cached) and has_cjk(s):
        return True
    return False


def translate_all(strings: list[str], source: str, locale: str, zmap: dict[str, str]) -> None:
    from concurrent.futures import ThreadPoolExecutor, as_completed

    todo = [s for s in strings if needs_translate(s, zmap)]
    print(f"  {source}->{locale}: {len(todo)}/{len(strings)} to translate", flush=True)
    if not todo:
        return
    done = 0
    with ThreadPoolExecutor(max_workers=6) as pool:
        futs = {pool.submit(lingva, s, source, LANG[locale]): s for s in todo}
        for fut in as_completed(futs):
            s = futs[fut]
            try:
                zmap[s] = fut.result()
            except Exception:
                zmap[s] = s
            done += 1
            if done % 50 == 0 or done == len(todo):
                print(f"    {done}/{len(todo)}", flush=True)
                ZH_MAP.write_text(json.dumps(zh_map, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
            time.sleep(0.05)


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")

    global zh_map
    zh_map = {loc: {} for loc in LOCALES}
    if ZH_MAP.exists():
        loaded = json.loads(ZH_MAP.read_text(encoding="utf-8"))
        for loc in LOCALES:
            zh_map[loc] = loaded.get(loc, {})

    en_data = {f: json.loads((ROOT / "en" / f).read_text(encoding="utf-8")) for f in FILES}
    zh_data = {f: json.loads((ROOT / "zh" / f).read_text(encoding="utf-8")) for f in FILES}
    en_flat = flatten(en_data["admin.json"]) | flatten(en_data["common.json"])
    zh_flat = flatten(zh_data["admin.json"]) | flatten(zh_data["common.json"])

    unique_zh = sorted({zh_flat[k] for k in en_flat if has_cjk(zh_flat.get(k, ""))})
    unique_en = sorted({en_flat[k] for k in en_flat if is_english(en_flat[k]) and not has_cjk(zh_flat.get(k, ""))})

    for locale in LOCALES:
        translate_all(unique_zh, "zh", locale, zh_map[locale])
        translate_all(unique_en, "en", locale, zh_map[locale])
        ZH_MAP.write_text(json.dumps(zh_map, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    for locale in LOCALES:
        overrides = MANUAL.get(locale, {})
        for fname in FILES:
            en_f = flatten(en_data[fname])
            zh_f = flatten(zh_data[fname])
            out_f: dict[str, str] = {}
            for key, en_val in en_f.items():
                if key in overrides:
                    out_f[key] = overrides[key]
                    continue
                if en_val in TERM:
                    out_f[key] = TERM[en_val][locale]
                    continue
                zh_val = zh_f.get(key, en_val)
                if has_cjk(zh_val) and zh_val in zh_map[locale]:
                    out_f[key] = zh_map[locale][zh_val]
                elif en_val in zh_map[locale]:
                    out_f[key] = zh_map[locale][en_val]
                elif has_cjk(zh_val):
                    out_f[key] = zh_map[locale].get(zh_val, zh_val)
                else:
                    out_f[key] = zh_map[locale].get(en_val, en_val)
                out_f[key] = fix_text(out_f[key])
            (ROOT / locale / fname).write_text(
                json.dumps(unflatten(out_f), ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            print(f"Wrote {locale}/{fname}", flush=True)

    print("Done.", flush=True)


if __name__ == "__main__":
    main()
