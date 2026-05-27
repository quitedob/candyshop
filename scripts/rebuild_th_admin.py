#!/usr/bin/env python3
"""Rebuild th/admin.json: manual overrides + en cache (Thai only) + en->th for remainder."""

from __future__ import annotations

import json
import re
import sys
import time
from pathlib import Path

from deep_translator import MyMemoryTranslator

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache.json"

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

PLACEHOLDER_RE = re.compile(r"(\{[a-zA-Z0-9_]+\})")
TOKEN_RE = re.compile(
    r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|1200×630|aʷ|→|…|\.\.\.|&|[0-9]+(?:\.[0-9]+)?%?|[A-Za-z][A-Za-z0-9+#./\-]{1,})"
)

# Load manual overrides from build script
exec((Path(__file__).parent / "build_th_admin_from_zh.py").read_text(encoding="utf-8").split("def flatten")[0])


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


def has_thai(text: str) -> bool:
    return bool(re.search(r"[\u0E00-\u0E7F]", text))


def has_cjk(text: str) -> bool:
    return bool(re.search(r"[\u4e00-\u9fff]", text))


def good_translation(en_val: str, candidate: str) -> bool:
    if not candidate or candidate == en_val:
        return False
    if has_cjk(candidate):
        return False
    if has_thai(candidate):
        return True
    return False


def translate_en_to_th(text: str, cache: dict[str, str]) -> str:
    if text in cache and good_translation(text, cache[text]):
        return cache[text]
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source="en-US", target="th-TH")
    for attempt in range(5):
        try:
            result = translator.translate(protected)
            out = restore(result or text, replacements)
            if good_translation(text, out):
                cache[text] = out
                return out
        except Exception:
            time.sleep(0.5 * (attempt + 1))
    return cache.get(text, text)


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    all_caches = json.loads(CACHE_PATH.read_text(encoding="utf-8")) if CACHE_PATH.exists() else {}
    cache: dict[str, str] = all_caches.get("th", {})

    en_flat = flatten(json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8")))
    out_flat: dict[str, str] = {}
    todo: list[str] = []

    for key, en_val in en_flat.items():
        if key in MANUAL_BY_KEY:
            out_flat[key] = MANUAL_BY_KEY[key]
            continue
        cached = cache.get(en_val, "")
        if good_translation(en_val, cached):
            out_flat[key] = cached
        else:
            todo.append(en_val)

    unique_todo = sorted(set(todo), key=len)
    print(f"Translating {len(unique_todo)} strings en->th", flush=True)
    for i, text in enumerate(unique_todo, 1):
        th_text = translate_en_to_th(text, cache)
        if i % 25 == 0 or i == len(unique_todo):
            print(f"  {i}/{len(unique_todo)}", flush=True)
            all_caches["th"] = cache
            CACHE_PATH.write_text(json.dumps(all_caches, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    for key, en_val in en_flat.items():
        if key in MANUAL_BY_KEY:
            continue
        if key not in out_flat or not good_translation(en_val, out_flat[key]):
            out_flat[key] = cache.get(en_val, en_val)

    (ROOT / "th" / "admin.json").write_text(
        json.dumps(unflatten(out_flat), ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    all_caches["th"] = cache
    CACHE_PATH.write_text(json.dumps(all_caches, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print("Wrote th/admin.json", flush=True)


if __name__ == "__main__":
    main()
