#!/usr/bin/env python3
"""Fix failed i18n translations and rebuild ar/th/vi locale files."""

from __future__ import annotations

import json
import re
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

from deep_translator import MyMemoryTranslator

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache.json"
LOCALES = ("ar", "th", "vi")
FILES = ("admin.json", "common.json")
WORKERS = 4

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Temperature",
    "Plan-Execute-Replan", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA", "ISO",
    "CSV", "YAML", "Stripe", "PayPal", "Chatbot", "Agent", "Cursor", "Markdown",
    "POST", "GET", "EN", "VI", "TH", "MS", "ID", "JA", "KO",
}

LANG_CODES = {
    "ar": ("en-US", "ar-SA"),
    "th": ("en-US", "th-TH"),
    "vi": ("en-US", "vi-VN"),
}

BROKEN_KEEP_RE = re.compile(r"__KEEP\s*_?\s*(\d+)\s*__", re.I)
PLACEHOLDER_RE = re.compile(r"(\{[a-zA-Z0-9_]+\})")


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
    replacements: list[tuple[str, str]] = []
    result = text

    for ph in PLACEHOLDER_RE.findall(text):
        token = f"ZZPH{len(replacements)}ZZ"
        replacements.append((token, ph))
        result = result.replace(ph, token, 1)

    for term in sorted(GLOSSARY, key=len, reverse=True):
        if term in result:
            token = f"ZZKEEP{len(replacements)}ZZ"
            replacements.append((token, term))
            result = result.replace(term, token)

    return result, replacements


def restore(text: str, replacements: list[tuple[str, str]]) -> str:
    for token, original in replacements:
        text = text.replace(token, original)
    return text


def fix_broken_glossary(text: str, source: str) -> str:
    if "__KEEP" not in text and "ZZKEEP" not in text:
        return text

    def repl(match: re.Match[str]) -> str:
        idx = int(match.group(1))
        tokens = []
        protected, replacements = protect(source)
        for token, original in replacements:
            if token.startswith("ZZKEEP"):
                tokens.append(original)
        if idx < len(tokens):
            return tokens[idx]
        return match.group(0)

    text = BROKEN_KEEP_RE.sub(repl, text)
    text = re.sub(r"ZZKEEP(\d+)ZZ", repl, text)
    return text


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def save_json(path: Path, data: dict) -> None:
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def needs_retranslate(key: str, value: str) -> bool:
    if key == value and len(key) > 2 and re.search(r"[A-Za-z]{2,}", key):
        return True
    if "__KEEP" in value or "ZZKEEP" in value:
        return True
    if "&#10;" in value or "&amp;" in value:
        return True
    return False


def translate_one(text: str, locale: str) -> str:
    src, tgt = LANG_CODES[locale]
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source=src, target=tgt)
    for attempt in range(5):
        try:
            result = translator.translate(protected)
            if result:
                restored = restore(result, replacements)
                restored = fix_broken_glossary(restored, text)
                if restored != text or attempt == 4:
                    return restored
        except Exception:
            pass
        time.sleep(0.8 * (attempt + 1))
    return fix_broken_glossary(text, text)


def retranslate_locale(locale: str, cache: dict[str, str]) -> None:
    todo = [k for k, v in list(cache.items()) if needs_retranslate(k, v)]
    print(f"  {locale}: retranslating {len(todo)} entries...", flush=True)
    done = 0
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futures = {pool.submit(translate_one, text, locale): text for text in todo}
        for future in as_completed(futures):
            text = futures[future]
            try:
                cache[text] = future.result()
            except Exception:
                cache[text] = fix_broken_glossary(text, text)
            done += 1
            if done % 50 == 0 or done == len(todo):
                print(f"    {locale}: {done}/{len(todo)}", flush=True)
            time.sleep(0.05)


def build_files(cache: dict[str, dict[str, str]]) -> None:
    en_data = {f: load_json(ROOT / "en" / f) for f in FILES}
    for locale in LOCALES:
        for filename in FILES:
            en_flat = flatten(en_data[filename])
            out_flat = {key: cache[locale].get(val, val) for key, val in en_flat.items()}
            save_json(ROOT / locale / filename, unflatten(out_flat))
            print(f"Wrote {locale}/{filename}", flush=True)


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    cache = json.loads(CACHE_PATH.read_text(encoding="utf-8"))

    en_strings: set[str] = set()
    for filename in FILES:
        en_strings.update(flatten(load_json(ROOT / "en" / filename)).values())

    for locale in LOCALES:
        cache.setdefault(locale, {})
        for s in en_strings:
            cache[locale].setdefault(s, s)
        print(f"Fixing {locale}...", flush=True)
        retranslate_locale(locale, cache[locale])

    CACHE_PATH.write_text(json.dumps(cache, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    build_files(cache)
    print("Done.", flush=True)


if __name__ == "__main__":
    main()
