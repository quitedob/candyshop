#!/usr/bin/env python3
"""Build ar/admin.json and ar/common.json from en sources, cache, and manual overrides."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

from ar_manual_translations import MANUAL

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache.json"

ADMIN_OVERRIDES = {
    "admin.langLabel": "العربية",
    "admin.langLabelShort": "ع",
    "admin.langTitle": "العربية",
    "admin.brand": "CandyPro Admin",
}

ARABIC_RE = re.compile(r"[\u0600-\u06FF]")
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


def fix_html_entities(text: str) -> str:
    return HTML_ENTITY_RE.sub(" ", text).replace("  ", " ").strip()


def has_arabic(text: str) -> bool:
    return bool(ARABIC_RE.search(text))


def needs_manual(en_val: str, current: str) -> bool:
    if current == en_val:
        return True
    if CHINESE_RE.search(current):
        return True
    if HTML_ENTITY_RE.search(current):
        return True
    if not has_arabic(current):
        return True
    return False


def resolve_translation(en_val: str, cache: dict[str, str]) -> str:
    if en_val in MANUAL:
        return MANUAL[en_val]
    cached = fix_html_entities(cache.get(en_val, en_val))
    if needs_manual(en_val, cached):
        return MANUAL.get(en_val, cached)
    return cached


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def save_json(path: Path, data: dict) -> None:
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    cache = load_json(CACHE_PATH).get("ar", {}) if CACHE_PATH.exists() else {}

    totals = {"keys": 0, "manual": 0, "cache": 0, "overrides": 0, "still_english": 0, "no_arabic": 0}

    for filename in ("admin.json", "common.json"):
        en_flat = flatten(load_json(ROOT / "en" / filename))
        overrides = ADMIN_OVERRIDES if filename == "admin.json" else {}
        out_flat: dict[str, str] = {}

        for key, en_val in en_flat.items():
            if key in overrides:
                out_flat[key] = overrides[key]
                totals["overrides"] += 1
                continue

            cached = cache.get(en_val, en_val)
            if en_val in MANUAL or needs_manual(en_val, cached):
                val = resolve_translation(en_val, cache)
                totals["manual"] += 1
            else:
                val = fix_html_entities(cached)
                totals["cache"] += 1

            out_flat[key] = fix_html_entities(val)
            totals["keys"] += 1

            if val == en_val and len(en_val) > 3 and re.search(r"[A-Za-z]{3,}", en_val):
                totals["still_english"] += 1
            if not has_arabic(val) and en_val != val and not re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&⌘]+", val):
                # allow glossary-only strings
                latin = len(re.findall(r"[A-Za-z]", val))
                if latin / max(len(val), 1) > 0.7:
                    totals["no_arabic"] += 1

        save_json(ROOT / "ar" / filename, unflatten(out_flat))
        print(f"Wrote ar/{filename}: {len(out_flat)} keys")

    print(
        f"Stats: keys={totals['keys']} manual={totals['manual']} cache={totals['cache']} "
        f"overrides={totals['overrides']} still_english={totals['still_english']} "
        f"no_arabic={totals['no_arabic']}"
    )


if __name__ == "__main__":
    main()
