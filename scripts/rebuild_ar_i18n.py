#!/usr/bin/env python3
"""Rebuild ar/admin.json and ar/common.json from en with full Arabic translations."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache.json"

MANUAL_OVERRIDES = {
    "admin.langLabel": "العربية",
    "admin.langLabelShort": "ع",
    "admin.langTitle": "العربية",
    "admin.brand": "CandyPro Admin",
}

CHINESE_RE = re.compile(r"[\u4e00-\u9fff]")
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


def is_mostly_english(text: str) -> bool:
    if not isinstance(text, str) or len(text) <= 2:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&⌘]+", text):
        return False
    keep = {
        "CandyPro Admin", "CandyPro", "N/A", "TBD", "MOQ", "SKU", "OEM", "AI",
        "XLSX", "DOCX", "URL", "HTTP", "JSON", "WhatsApp", "DHL", "FedEx", "UPS",
        "USPS", "EMS", "FOB", "USD", "COGS", "GTIN", "HS", "Ctrl+K", "⌘K",
        "Webhook", "Webhooks", "Hooks", "3PL", "RCEP", "CBM", "B2B", "KYB", "RFM",
        "Slug", "Incoterms", "Proforma", "Plan-Execute-Replan", "ODM", "PI", "CI",
        "SC", "B/L", "PL", "ETA", "ISO", "Chatbot", "Agent", "Cursor", "Facebook",
        "LinkedIn", "Instagram", "YouTube", "Twitter", "HACCP", "ISO22000", "BRC",
        "HALAL", "GMO", "Non-GMO", "Stripe", "PayPal",
    }
    if text in keep:
        return False
    latin = len(re.findall(r"[A-Za-z]", text))
    return latin / max(len(text), 1) > 0.55


def needs_translation(current: str, en_val: str) -> bool:
    if current == en_val:
        return True
    if CHINESE_RE.search(current):
        return True
    if HTML_ENTITY_RE.search(current):
        return True
    if is_mostly_english(current):
        return True
    return False


def fix_html_entities(text: str) -> str:
    return HTML_ENTITY_RE.sub(" ", text).replace("  ", " ").strip()


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def save_json(path: Path, data: dict) -> None:
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    cache = load_json(CACHE_PATH).get("ar", {}) if CACHE_PATH.exists() else {}

    for filename in ("admin.json", "common.json"):
        en_tree = load_json(ROOT / "en" / filename)
        en_flat = flatten(en_tree)
        existing_flat = flatten(load_json(ROOT / "ar" / filename)) if (ROOT / "ar" / filename).exists() else {}
        overrides = MANUAL_OVERRIDES if filename == "admin.json" else {}

        out_flat: dict[str, str] = {}
        translated = kept = overridden = 0

        for key, en_val in en_flat.items():
            if key in overrides:
                out_flat[key] = overrides[key]
                overridden += 1
                continue

            current = existing_flat.get(key, en_val)
            if key in existing_flat and not needs_translation(current, en_val):
                out_flat[key] = fix_html_entities(current)
                kept += 1
            else:
                val = cache.get(en_val, en_val)
                out_flat[key] = fix_html_entities(val)
                translated += 1

        save_json(ROOT / "ar" / filename, unflatten(out_flat))
        print(f"Wrote ar/{filename}: {len(out_flat)} keys (kept={kept}, translated={translated}, overrides={overridden})")


if __name__ == "__main__":
    main()
