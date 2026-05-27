#!/usr/bin/env python3
"""Rebuild ar/th/vi locales using Lingva translate API (no MyMemory quota)."""

from __future__ import annotations

import json
import os
import re
import sys
import time
import urllib.parse
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

import requests

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE = Path(__file__).resolve().parent / "i18n-translation-cache.json"
LOCALES = ("ar", "th", "vi")
FILES = ("admin.json", "common.json")
WORKERS = 6
LANG = {"ar": "ar", "th": "th", "vi": "vi"}

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
    "⌘K", "Ctrl+K",
}

MANUAL = {
    "vi": {
        "admin.users.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "admin.staff.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "admin.auditLog.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "admin.inventory.batch_edit_result": "Đã cập nhật {updated}, {errors} lỗi",
        "admin.products.base_price": "Giá cơ bản (USD)",
        "admin.products.field_energy_kcal": "Năng lượng (kcal)",
        "admin.langLabel": "VI",
        "admin.langLabelShort": "VI",
        "admin.langTitle": "Tiếng Việt",
    },
    "ar": {
        "admin.langLabel": "العربية",
        "admin.langLabelShort": "ع",
        "admin.langTitle": "العربية",
        "admin.brand": "CandyPro Admin",
    },
    "th": {
        "admin.langLabel": "TH",
        "admin.langLabelShort": "TH",
        "admin.langTitle": "ไทย",
        "admin.brand": "CandyPro Admin",
    },
}

SESSION = requests.Session()
SESSION.trust_env = False
for k in ("HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy"):
    os.environ.pop(k, None)


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


def fix_broken(text: str) -> str:
    return re.sub(r"__KEEP\s*_?\s*\d+\s*__", lambda m: "", text).strip()


def is_english(text: str) -> bool:
    if not text or len(text) <= 2:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&]+", text):
        return False
    return len(re.findall(r"[A-Za-z]", text)) / max(len(text), 1) > 0.55


def lingva(text: str, target: str) -> str:
    protected, tokens = protect(text)
    encoded = urllib.parse.quote(protected, safe="")
    url = f"https://lingva.ml/api/v1/en/{target}/{encoded}"
    for attempt in range(4):
        try:
            r = SESSION.get(url, timeout=20)
            if r.status_code == 200:
                data = r.json()
                result = data.get("translation") or text
                return fix_broken(restore(result, tokens))
        except Exception:
            pass
        time.sleep(0.5 * (attempt + 1))
    return text


def translate_batch(strings: list[str], locale: str, cache: dict[str, str]) -> None:
    todo = [s for s in strings if s not in cache or cache[s] == s or "__KEEP" in cache.get(s, "")]
    print(f"  {locale}: {len(todo)} to translate", flush=True)
    done = 0
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futs = {pool.submit(lingva, s, LANG[locale]): s for s in todo}
        for fut in as_completed(futs):
            s = futs[fut]
            try:
                cache[s] = fut.result()
            except Exception:
                cache[s] = s
            done += 1
            if done % 50 == 0 or done == len(todo):
                print(f"    {done}/{len(todo)}", flush=True)
            time.sleep(0.08)


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    cache = json.loads(CACHE.read_text(encoding="utf-8")) if CACHE.exists() else {loc: {} for loc in LOCALES}
    for loc in LOCALES:
        cache.setdefault(loc, {})

    en_data = {f: json.loads((ROOT / "en" / f).read_text(encoding="utf-8")) for f in FILES}

    all_en: set[str] = set()
    for f in FILES:
        all_en.update(flatten(en_data[f]).values())

    for locale in LOCALES:
        translate_batch(sorted(all_en, key=len), locale, cache[locale])
        CACHE.write_text(json.dumps(cache, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    for locale in LOCALES:
        overrides = MANUAL.get(locale, {})
        for fname in FILES:
            en_flat = flatten(en_data[fname])
            out = {}
            for key, en_val in en_flat.items():
                if key in overrides:
                    out[key] = overrides[key]
                else:
                    val = cache[locale].get(en_val, en_val)
                    out[key] = fix_broken(val) if val else en_val
            (ROOT / locale / fname).write_text(
                json.dumps(unflatten(out), ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            print(f"Wrote {locale}/{fname}", flush=True)

    print("Done.", flush=True)


if __name__ == "__main__":
    main()
