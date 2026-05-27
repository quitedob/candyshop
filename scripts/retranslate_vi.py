#!/usr/bin/env python3
"""Retranslate untranslated vi strings and rebuild admin.json + common.json."""

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

OVERRIDES = {
    "admin.brand": "Quản trị CandyPro",
    "admin.langLabel": "VI",
    "admin.langLabelShort": "VI",
    "admin.langTitle": "Tiếng Việt",
    "admin.users.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
    "admin.staff.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
    "admin.auditLog.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
    "admin.inventory.batch_edit_result": "Đã cập nhật {updated}, {errors} lỗi",
    "admin.products.base_price": "Giá cơ bản (USD)",
    "admin.products.field_energy_kcal": "Năng lượng (kcal)",
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


def needs_translation(text: str, cached: str) -> bool:
    if cached != text:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&]+", text):
        return False
    if len(text) <= 3 and text.isupper():
        return False
    return bool(re.search(r"[A-Za-z]", text))


def translate_one(text: str) -> str:
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source="en-US", target="vi-VN")
    for attempt in range(5):
        try:
            result = translator.translate(protected)
            if result and result != text:
                return restore(result, replacements)
        except Exception:
            time.sleep(0.8 * (attempt + 1))
    return text


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")

    all_cache = json.loads(CACHE_PATH.read_text(encoding="utf-8"))
    vi_cache: dict[str, str] = all_cache.setdefault("vi", {})

    needed: set[str] = set()
    for filename in ("admin.json", "common.json"):
        en_flat = flatten(json.loads((ROOT / "en" / filename).read_text(encoding="utf-8")))
        needed.update(en_flat.values())

    todo = sorted([s for s in needed if needs_translation(s, vi_cache.get(s, s))], key=len)
    print(f"Retranslating {len(todo)} strings...", flush=True)

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
                vi_cache[text] = result
            done += 1
            if done % 50 == 0 or done == len(todo):
                print(f"  {done}/{len(todo)}", flush=True)
                all_cache["vi"] = vi_cache
                CACHE_PATH.write_text(
                    json.dumps(all_cache, ensure_ascii=False, indent=2) + "\n",
                    encoding="utf-8",
                )

    all_cache["vi"] = vi_cache
    CACHE_PATH.write_text(json.dumps(all_cache, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    for filename in ("admin.json", "common.json"):
        en_flat = flatten(json.loads((ROOT / "en" / filename).read_text(encoding="utf-8")))
        out_flat = {}
        for key, en_val in en_flat.items():
            if key in OVERRIDES:
                out_flat[key] = OVERRIDES[key]
            else:
                out_flat[key] = vi_cache.get(en_val, en_val)
        out_path = ROOT / "vi" / filename
        out_path.write_text(
            json.dumps(unflatten(out_flat), ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
        print(f"Wrote vi/{filename}", flush=True)

    remaining = sum(1 for s in needed if vi_cache.get(s, s) == s and re.search(r"[A-Za-z]", s))
    print(f"Done. Still untranslated (latin): {remaining}", flush=True)


if __name__ == "__main__":
    main()
