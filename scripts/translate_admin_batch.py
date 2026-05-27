#!/usr/bin/env python3
"""Translate admin.json en strings to id/ms via Google Translate (offline cache)."""

from __future__ import annotations

import json
import re
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE = Path(__file__).resolve().parent / "admin_translation_cache.json"

GLOSSARY = sorted(
    {
        "CandyPro", "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON",
        "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS", "GTIN",
        "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL", "RCEP", "CBM",
        "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "Incoterms", "B2B", "KYB",
        "RFM", "Slug", "Plan-Execute-Replan", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA",
        "Stripe", "PayPal", "Chatbot", "Agent", "Cursor", "translate_content", "CRUD",
        "Facebook", "LinkedIn", "Super Admin", "Superadmin", "Grade A", "Ctrl+K", "⌘K",
        "1200×630", "aʷ", "Diff", "Trade Assistant", "B2B Coordinator", "Product Recommend",
        "Bahasa Indonesia", "Bahasa Melayu", "EN", "ID", "MS",
    },
    key=len,
    reverse=True,
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
    reps: list[tuple[str, str]] = []
    protected = text
    for term in GLOSSARY:
        if term in protected:
            ph = f"__G{len(reps)}__"
            reps.append((ph, term))
            protected = protected.replace(term, ph)
    for m in re.finditer(r"\{[a-zA-Z0-9_]+\}", text):
        token = m.group(0)
        ph = f"__P{len(reps)}__"
        reps.append((ph, token))
        protected = protected.replace(token, ph, 1)
    return protected, reps


def restore(text: str, reps: list[tuple[str, str]]) -> str:
    for ph, orig in reps:
        text = text.replace(ph, orig)
    return text


def load_cache() -> dict:
    if CACHE.exists():
        return json.loads(CACHE.read_text(encoding="utf-8"))
    return {"id": {}, "ms": {}}


def save_cache(cache: dict) -> None:
    CACHE.write_text(json.dumps(cache, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def translate_text(text: str, target: str, cache: dict) -> str:
    if text in cache[target]:
        return cache[target][text]
    from deep_translator import GoogleTranslator

    protected, reps = protect(text)
    translated = GoogleTranslator(source="en", target=target).translate(protected)
    if not translated:
        translated = protected
    result = restore(translated, reps)
    cache[target][text] = result
    return result


def main() -> None:
    en_admin = json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8"))["admin"]
    en_flat = flatten(en_admin)
    unique = sorted(set(en_flat.values()))

    cache = load_cache()
    for locale in ("id", "ms"):
        cache.setdefault(locale, {})

    total = len(unique)
    for i, text in enumerate(unique, 1):
        for locale in ("id", "ms"):
            if text not in cache[locale]:
                try:
                    translate_text(text, locale, cache)
                except Exception as exc:  # noqa: BLE001
                    print(f"ERR [{locale}] {text[:60]!r}: {exc}")
                    cache[locale][text] = text
                time.sleep(0.08)
        if i % 10 == 0:
            save_cache(cache)
            print(f"Progress {i}/{total}", flush=True)
    save_cache(cache)

    for locale in ("id", "ms"):
        out_flat = {k: cache[locale].get(v, v) for k, v in en_flat.items()}
        data = {"admin": unflatten(out_flat)}
        path = ROOT / locale / "admin.json"
        path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        same = sum(1 for k, v in en_flat.items() if out_flat[k] == v)
        print(f"Wrote {path} ({same} identical to en)")


if __name__ == "__main__":
    main()
