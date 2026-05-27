#!/usr/bin/env python3
"""Translate remaining English strings to Arabic and rebuild ar i18n files."""

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
WORKERS = 8

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Temperature",
    "Plan-Execute-Replan", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA", "ISO",
    "CSV", "YAML", "Stripe", "PayPal", "Chatbot", "Agent", "Cursor", "Markdown",
    "CandyPro Admin", "3M", "6M", "12M", "Trade", "MS", "0", "—", "한", "日",
    "한국어", "ع", "العربية", "Ctrl+K", "⌘K",
}

PLACEHOLDER_RE = re.compile(r"(\{[a-zA-Z0-9_]+\})")
TOKEN_RE = re.compile(
    r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|1200×630|aʷ|→|…|\.\.\.|&|[0-9]+(?:\.[0-9]+)?%?|[A-Za-z][A-Za-z0-9+#./\-]{1,})"
)

MANUAL_OVERRIDES = {
    "admin.langLabel": "العربية",
    "admin.langLabelShort": "ع",
    "admin.langTitle": "العربية",
    "admin.brand": "CandyPro Admin",
}

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
                placeholder = f"__KEEP_{len(replacements)}__"
                replacements.append((placeholder, token))
                return placeholder
        if PLACEHOLDER_RE.fullmatch(token):
            placeholder = f"__PH_{len(replacements)}__"
            replacements.append((placeholder, token))
            return placeholder
        return token

    return TOKEN_RE.sub(repl, text), replacements


def restore(text: str, replacements: list[tuple[str, str]]) -> str:
    for placeholder, original in replacements:
        text = text.replace(placeholder, original)
    return text


def translate_one(text: str) -> str:
    if text in GLOSSARY or len(text) <= 2:
        return text
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source="en-US", target="ar-SA")
    for attempt in range(4):
        try:
            result = translator.translate(protected)
            if result and result != protected:
                return restore(result, replacements)
            return restore(result or text, replacements)
        except Exception:
            time.sleep(0.5 * (attempt + 1))
    return text


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def save_json(path: Path, data: dict) -> None:
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    cache_all = load_json(CACHE_PATH) if CACHE_PATH.exists() else {"ar": {}, "th": {}, "vi": {}}
    ar_cache: dict[str, str] = cache_all.setdefault("ar", {})

    needed: set[str] = set()
    for filename in ("admin.json", "common.json"):
        en_flat = flatten(load_json(ROOT / "en" / filename))
        ar_flat = flatten(load_json(ROOT / "ar" / filename))
        for key, en_val in en_flat.items():
            if ar_flat.get(key, en_val) == en_val and en_val not in GLOSSARY:
                cached = ar_cache.get(en_val, en_val)
                if cached == en_val and len(en_val) > 2:
                    needed.add(en_val)

    todo = sorted(needed, key=len)
    print(f"Translating {len(todo)} strings...", flush=True)
    done = 0
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futures = {pool.submit(translate_one, text): text for text in todo}
        for future in as_completed(futures):
            text = futures[future]
            try:
                result = future.result()
            except Exception:
                result = text
            with _cache_lock:
                ar_cache[text] = result
            done += 1
            if done % 50 == 0 or done == len(todo):
                print(f"  {done}/{len(todo)}", flush=True)

    save_json(CACHE_PATH, cache_all)

    for filename in ("admin.json", "common.json"):
        en_flat = flatten(load_json(ROOT / "en" / filename))
        overrides = MANUAL_OVERRIDES if filename == "admin.json" else {}
        out_flat: dict[str, str] = {}
        for key, en_val in en_flat.items():
            if key in overrides:
                out_flat[key] = overrides[key]
            elif en_val in GLOSSARY:
                out_flat[key] = en_val
            else:
                out_flat[key] = ar_cache.get(en_val, en_val)
        save_json(ROOT / "ar" / filename, unflatten(out_flat))
        print(f"Wrote ar/{filename}: {len(out_flat)} keys", flush=True)


if __name__ == "__main__":
    main()
