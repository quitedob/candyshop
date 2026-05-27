#!/usr/bin/env python3
"""Build vi admin/common from zh semantic source + cache, translating zh->vi."""

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
_zh_vi_cache: dict[str, str] = {}


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


def has_vietnamese(text: str) -> bool:
    return bool(re.search(r"[\u0100-\u024f\u1e00-\u1eff]", text)) or bool(
        re.search(r"[àáảãạăằắẳẵặâầấẩẫậèéẻẽẹêềếểễệìíỉĩịòóỏõọôồốổỗộơờớởỡợùúủũụưừứửữựỳýỷỹỵđ]", text, re.I)
    )


def translate_zh_vi(text: str) -> str:
    if text in _zh_vi_cache:
        return _zh_vi_cache[text]
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source="zh-CN", target="vi-VN")
    for attempt in range(4):
        try:
            result = translator.translate(protected)
            if result:
                out = restore(result, replacements)
                _zh_vi_cache[text] = out
                return out
        except Exception:
            time.sleep(0.5 * (attempt + 1))
    _zh_vi_cache[text] = text
    return text


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")

    for filename in ("admin.json", "common.json"):
        en_flat = flatten(json.loads((ROOT / "en" / filename).read_text(encoding="utf-8")))
        zh_flat = flatten(json.loads((ROOT / "zh" / filename).read_text(encoding="utf-8")))

        todo_zh: list[str] = []
        plan: dict[str, str] = {}

        for key, en_val in en_flat.items():
            if key in OVERRIDES:
                plan[key] = OVERRIDES[key]
                continue
            zh_val = zh_flat.get(key, en_val)
            if zh_val != en_val and re.search(r"[\u4e00-\u9fff]", zh_val):
                todo_zh.append(zh_val)
                plan[key] = ("__ZH__", zh_val)
            else:
                plan[key] = ("__EN__", en_val)

        unique_zh = sorted(set(todo_zh), key=len)
        print(f"{filename}: translating {len(unique_zh)} zh strings...", flush=True)

        done = 0
        with ThreadPoolExecutor(max_workers=WORKERS) as pool:
            futures = {pool.submit(translate_zh_vi, text): text for text in unique_zh}
            for future in as_completed(futures):
                future.result()
                done += 1
                if done % 100 == 0 or done == len(unique_zh):
                    print(f"  {done}/{len(unique_zh)}", flush=True)

        out_flat: dict[str, str] = {}
        for key, item in plan.items():
            if isinstance(item, str):
                out_flat[key] = item
            elif item[0] == "__ZH__":
                out_flat[key] = _zh_vi_cache.get(item[1], item[1])
            else:
                out_flat[key] = item[1]

        out_path = ROOT / "vi" / filename
        out_path.write_text(
            json.dumps(unflatten(out_flat), ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )

        untranslated = sum(
            1
            for k, v in out_flat.items()
            if v == en_flat[k] and re.search(r"[A-Za-z]{3,}", v)
        )
        no_vi = sum(1 for v in out_flat.values() if not has_vietnamese(v) and re.search(r"[A-Za-z]", v))
        print(f"Wrote vi/{filename} — identical_to_en={untranslated}, no_vietnamese_chars={no_vi}", flush=True)


if __name__ == "__main__":
    main()
