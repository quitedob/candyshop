#!/usr/bin/env python3
"""Fill remaining English strings in th i18n using Google Translate."""

from __future__ import annotations

import json
import re
import sys
import time
from pathlib import Path

from deep_translator import GoogleTranslator

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache.json"
FILES = ("admin.json", "common.json")

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Temperature",
    "Plan-Execute-Replan", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA", "ISO",
    "CSV", "YAML", "Stripe", "PayPal", "Chatbot", "Agent", "Cursor", "Markdown",
}

MANUAL_BY_KEY = {
    "admin.langLabel": "TH",
    "admin.langLabelShort": "TH",
    "admin.langTitle": "ไทย",
    "admin.brand": "CandyPro Admin",
}

PLACEHOLDER_RE = re.compile(r"(\{[a-zA-Z0-9_]+\})")
TOKEN_RE = re.compile(
    r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|1200×630|aʷ|→|…|\.\.\.|&|[0-9]+(?:\.[0-9]+)?%?|[A-Za-z][A-Za-z0-9+#./\-]{1,})"
)


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


def is_mostly_english(text: str) -> bool:
    if not isinstance(text, str) or len(text) <= 2:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&]+", text):
        return False
    latin = len(re.findall(r"[A-Za-z]", text))
    return latin / max(len(text), 1) > 0.55


def has_thai(text: str) -> bool:
    return bool(re.search(r"[\u0E00-\u0E7F]", text))


def needs_translation(en_val: str, cur: str) -> bool:
    if cur == en_val:
        return True
    if is_mostly_english(cur) and len(en_val) > 3:
        return True
    if not has_thai(cur) and len(en_val) > 3 and re.search(r"[A-Za-z]{3,}", en_val):
        return True
    return False


def translate_one(text: str, translator: GoogleTranslator) -> str:
    protected, replacements = protect(text)
    for attempt in range(4):
        try:
            result = translator.translate(protected)
            out = restore(result or text, replacements)
            if out and has_thai(out):
                return out
            if out and out != text and not is_mostly_english(out):
                return out
        except Exception:
            time.sleep(1.0 * (attempt + 1))
    return text


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    all_caches = json.loads(CACHE_PATH.read_text(encoding="utf-8")) if CACHE_PATH.exists() else {}
    cache = all_caches.setdefault("th", {})
    translator = GoogleTranslator(source="en", target="th")

    for filename in FILES:
        en_flat = flatten(json.loads((ROOT / "en" / filename).read_text(encoding="utf-8")))
        th_path = ROOT / "th" / filename
        th_flat = flatten(json.loads(th_path.read_text(encoding="utf-8"))) if th_path.exists() else {}

        todo: set[str] = set()
        for key, en_val in en_flat.items():
            cur = th_flat.get(key, en_val)
            if key in MANUAL_BY_KEY:
                continue
            if needs_translation(en_val, cur):
                todo.add(en_val)

        print(f"{filename}: translating {len(todo)} unique strings", flush=True)
        done = 0
        for text in sorted(todo, key=len):
            if text in cache and has_thai(cache[text]) and not needs_translation(text, cache[text]):
                continue
            cache[text] = translate_one(text, translator)
            done += 1
            if done % 20 == 0:
                print(f"  {done}/{len(todo)}", flush=True)
                all_caches["th"] = cache
                CACHE_PATH.write_text(json.dumps(all_caches, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
            time.sleep(0.05)

        out_flat: dict[str, str] = {}
        for key, en_val in en_flat.items():
            if key in MANUAL_BY_KEY:
                out_flat[key] = MANUAL_BY_KEY[key]
            else:
                cur = th_flat.get(key, en_val)
                if needs_translation(en_val, cur):
                    out_flat[key] = cache.get(en_val, en_val)
                else:
                    out_flat[key] = cur

        th_path.write_text(json.dumps(unflatten(out_flat), ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print(f"Wrote th/{filename}", flush=True)

    all_caches["th"] = cache
    CACHE_PATH.write_text(json.dumps(all_caches, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print("Done.", flush=True)


if __name__ == "__main__":
    main()
