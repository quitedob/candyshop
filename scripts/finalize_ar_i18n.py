#!/usr/bin/env python3
"""Finalize ar/admin.json and ar/common.json with complete Arabic translations."""

from __future__ import annotations

import json
import re
import sys
import threading
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

from deep_translator import MyMemoryTranslator

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache.json"
WORKERS = 6

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl+K", "⌘K", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug",
    "Plan-Execute-Replan", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA", "ISO",
    "CSV", "YAML", "Stripe", "PayPal", "Chatbot", "Agent", "Cursor", "Markdown",
    "CandyPro Admin", "3M", "6M", "12M", "Trade", "MS",
}

MANUAL_OVERRIDES = {
    "admin.langLabel": "العربية",
    "admin.langLabelShort": "ع",
    "admin.langTitle": "العربية",
    "admin.brand": "CandyPro Admin",
}

ARABIC_RE = re.compile(r"[\u0600-\u06FF]")
CHINESE_RE = re.compile(r"[\u4e00-\u9fff]")
HTML_ENTITY_RE = re.compile(r"&#\d+;")
PLACEHOLDER_RE = re.compile(r"(\{[a-zA-Z0-9_]+\})")
TOKEN_RE = re.compile(
    r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|1200×630|→|…|\.\.\.|&|[0-9]+(?:\.[0-9]+)?%?|[A-Za-z][A-Za-z0-9+#./\-]{1,})"
)

_cache_lock = threading.Lock()


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

    def repl(match: re.Match[str]) -> str:
        token = match.group(0)
        for keep in GLOSSARY:
            if token == keep:
                ph = f"__K{len(replacements)}__"
                replacements.append((ph, token))
                return ph
        if PLACEHOLDER_RE.fullmatch(token):
            ph = f"__P{len(replacements)}__"
            replacements.append((ph, token))
            return ph
        return token

    return TOKEN_RE.sub(repl, text), replacements


def restore(text: str, replacements: list[tuple[str, str]]) -> str:
    for ph, orig in replacements:
        text = text.replace(ph, orig)
    return text


def is_good_arabic(text: str, en_val: str) -> bool:
    if text == en_val:
        return False
    if CHINESE_RE.search(text):
        return False
    if HTML_ENTITY_RE.search(text):
        return False
    if not ARABIC_RE.search(text):
        return False
    # Reject reversed/broken RTL (Arabic chars but Latin word order artifacts)
    if text.startswith("...") or text.endswith("..."):
        return False
    return True


def translate_one(text: str) -> str:
    if text in GLOSSARY:
        return text
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source="en-US", target="ar-SA")
    for attempt in range(5):
        try:
            result = translator.translate(protected)
            if result:
                restored = restore(result, replacements)
                if is_good_arabic(restored, text) or text in GLOSSARY:
                    return restored
        except Exception:
            time.sleep(0.8 * (attempt + 1))
    return text


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def save_json(path: Path, data: dict) -> None:
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    cache_all = load_json(CACHE_PATH) if CACHE_PATH.exists() else {"ar": {}, "th": {}, "vi": {}}
    ar_cache: dict[str, str] = cache_all.setdefault("ar", {})

    todo: set[str] = set()
    for filename in ("admin.json", "common.json"):
        for en_val in flatten(load_json(ROOT / "en" / filename)).values():
            if en_val in GLOSSARY:
                continue
            cached = ar_cache.get(en_val, en_val)
            if not is_good_arabic(cached, en_val):
                todo.add(en_val)

    todo_list = sorted(todo, key=len)
    print(f"Translating {len(todo_list)} strings...", flush=True)
    done = 0
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futures = {pool.submit(translate_one, s): s for s in todo_list}
        for future in as_completed(futures):
            src = futures[future]
            try:
                result = future.result()
            except Exception:
                result = src
            with _cache_lock:
                if is_good_arabic(result, src):
                    ar_cache[src] = result
            done += 1
            if done % 40 == 0 or done == len(todo_list):
                print(f"  {done}/{len(todo_list)}", flush=True)
                save_json(CACHE_PATH, cache_all)

    save_json(CACHE_PATH, cache_all)

    for filename in ("admin.json", "common.json"):
        en_flat = flatten(load_json(ROOT / "en" / filename))
        overrides = MANUAL_OVERRIDES if filename == "admin.json" else {}
        out: dict[str, str] = {}
        missing = 0
        for key, en_val in en_flat.items():
            if key in overrides:
                out[key] = overrides[key]
            elif en_val in GLOSSARY:
                out[key] = en_val
            else:
                val = ar_cache.get(en_val, en_val)
                if not is_good_arabic(val, en_val):
                    missing += 1
                    val = translate_one(en_val)
                    if is_good_arabic(val, en_val):
                        ar_cache[en_val] = val
                out[key] = val
        save_json(ROOT / "ar" / filename, unflatten(out))
        print(f"Wrote ar/{filename}: {len(out)} keys, still-missing={missing}", flush=True)

    save_json(CACHE_PATH, cache_all)


if __name__ == "__main__":
    main()
