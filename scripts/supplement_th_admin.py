#!/usr/bin/env python3
"""Translate remaining admin en strings to th and merge into admin.json."""

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
SUPPLEMENT = Path(__file__).resolve().parent / "th_admin_supplement.json"
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


def has_thai(text: str) -> bool:
    return bool(re.search(r"[\u0E00-\u0E7F]", text))


def translate(text: str) -> str:
    protected, reps = protect(text)
    tr = MyMemoryTranslator(source="en-US", target="th-TH")
    for attempt in range(5):
        try:
            out = restore(tr.translate(protected) or text, reps)
            if has_thai(out):
                return out
        except Exception:
            time.sleep(0.6 * (attempt + 1))
    return text


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    supplement: dict[str, str] = {}
    if SUPPLEMENT.exists():
        supplement = json.loads(SUPPLEMENT.read_text(encoding="utf-8"))

    en = flatten(json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8")))
    th = flatten(json.loads((ROOT / "th" / "admin.json").read_text(encoding="utf-8")))

    todo = sorted({v for k, v in en.items() if th.get(k, v) == v and len(v) > 2})
    print(f"Need translation for {len(todo)} strings", flush=True)

    done = 0
    lock = threading.Lock()

    def work(text: str) -> tuple[str, str]:
        if text in supplement and has_thai(supplement[text]):
            return text, supplement[text]
        return text, translate(text)

    with ThreadPoolExecutor(max_workers=8) as pool:
        futures = {pool.submit(work, text): text for text in todo}
        for future in as_completed(futures):
            text, result = future.result()
            with lock:
                supplement[text] = result
                done += 1
                if done % 25 == 0 or done == len(todo):
                    print(f"  {done}/{len(todo)}", flush=True)
                    SUPPLEMENT.write_text(
                        json.dumps(supplement, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8",
                    )

    out: dict[str, str] = {}
    for k, v in en.items():
        if th.get(k) != v and has_thai(th.get(k, "")):
            out[k] = th[k]
        elif v in supplement and has_thai(supplement[v]):
            out[k] = supplement[v]
        else:
            out[k] = th.get(k, v)

    (ROOT / "th" / "admin.json").write_text(
        json.dumps(unflatten(out), ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    SUPPLEMENT.write_text(json.dumps(supplement, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print("Done", flush=True)


if __name__ == "__main__":
    main()
