#!/usr/bin/env python3
"""Final i18n build: zh-first Lingva translation + term overrides."""

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
ZH_MAP = Path(__file__).resolve().parent / "zh-locale-map.json"
WORKERS = 8
LOCALES = ("ar", "th", "vi")
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

# Import term overrides from apply_priority_i18n
sys.path.insert(0, str(Path(__file__).resolve().parent))
from apply_priority_i18n import TERM, flatten, unflatten  # noqa: E402

SESSION = requests.Session()
SESSION.trust_env = False
for k in ("HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy"):
    os.environ.pop(k, None)


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


def lingva(text: str, source: str, target: str) -> str:
    if not text or not text.strip():
        return text
    protected, tokens = protect(text)
    encoded = urllib.parse.quote(protected, safe="")
    url = f"https://lingva.ml/api/v1/{source}/{target}/{encoded}"
    for attempt in range(4):
        try:
            r = SESSION.get(url, timeout=25)
            if r.status_code == 200:
                result = r.json().get("translation") or text
                return restore(result, tokens)
        except Exception:
            pass
        time.sleep(0.4 * (attempt + 1))
    return text


def is_english(text: str) -> bool:
    if not text or len(text) <= 2:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&]+", text):
        return False
    return len(re.findall(r"[A-Za-z]", text)) / max(len(text), 1) > 0.55


def has_cjk(text: str) -> bool:
    return bool(re.search(r"[\u4e00-\u9fff]", text))


def translate_zh_strings(zh_strings: list[str], locale: str, zmap: dict[str, str]) -> None:
    todo = [z for z in zh_strings if z and z not in zmap]
    print(f"  zh->{locale}: {len(todo)} strings", flush=True)
    done = 0
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futs = {pool.submit(lingva, z, "zh", LANG[locale]): z for z in todo}
        for fut in as_completed(futs):
            z = futs[fut]
            try:
                zmap[z] = fut.result()
            except Exception:
                zmap[z] = z
            done += 1
            if done % 100 == 0 or done == len(todo):
                print(f"    {done}/{len(todo)}", flush=True)
            time.sleep(0.05)


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")

    zh_map: dict[str, dict[str, str]] = {loc: {} for loc in LOCALES}
    if ZH_MAP.exists():
        loaded = json.loads(ZH_MAP.read_text(encoding="utf-8"))
        for loc in LOCALES:
            zh_map[loc] = loaded.get(loc, {})

    en_admin = json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8"))
    en_common = json.loads((ROOT / "en" / "common.json").read_text(encoding="utf-8"))
    zh_admin = json.loads((ROOT / "zh" / "admin.json").read_text(encoding="utf-8"))
    zh_common = json.loads((ROOT / "zh" / "common.json").read_text(encoding="utf-8"))

    en_flat = flatten(en_admin) | flatten(en_common)
    zh_flat = flatten(zh_admin) | flatten(zh_common)

    unique_zh = sorted({zh_flat[k] for k in en_flat if zh_flat.get(k)})

    for locale in LOCALES:
        translate_zh_strings(unique_zh, locale, zh_map[locale])
        ZH_MAP.write_text(json.dumps(zh_map, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    MANUAL = {
        "vi": {
            "admin.langLabel": "VI", "admin.langLabelShort": "VI", "admin.langTitle": "Tiếng Việt",
            "admin.products.base_price": "Giá cơ bản (USD)",
            "admin.products.field_energy_kcal": "Năng lượng (kcal)",
        },
        "ar": {
            "admin.langLabel": "العربية", "admin.langLabelShort": "ع", "admin.langTitle": "العربية",
            "admin.brand": "CandyPro Admin",
        },
        "th": {
            "admin.langLabel": "TH", "admin.langLabelShort": "TH", "admin.langTitle": "ไทย",
            "admin.brand": "CandyPro Admin",
        },
    }

    for locale in LOCALES:
        for fname, en_tree in [("admin.json", en_admin), ("common.json", en_common)]:
            zh_tree = zh_admin if fname == "admin.json" else zh_common
            en_f = flatten(en_tree)
            zh_f = flatten(zh_tree)
            out_f: dict[str, str] = {}
            for key, en_val in en_f.items():
                if key in MANUAL.get(locale, {}):
                    out_f[key] = MANUAL[locale][key]
                    continue
                if en_val in TERM:
                    out_f[key] = TERM[en_val][locale]
                    continue
                zh_val = zh_f.get(key, en_val)
                if zh_val in zh_map[locale]:
                    out_f[key] = zh_map[locale][zh_val]
                elif has_cjk(zh_val):
                    out_f[key] = zh_map[locale].get(zh_val, zh_val)
                else:
                    out_f[key] = en_val if not is_english(en_val) else TERM.get(en_val, {}).get(locale, en_val)
            (ROOT / locale / fname).write_text(
                json.dumps(unflatten(out_f), ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            print(f"Wrote {locale}/{fname}", flush=True)

    print("Done.", flush=True)


if __name__ == "__main__":
    main()
