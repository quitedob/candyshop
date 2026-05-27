#!/usr/bin/env python3
"""Check ar/th/vi i18n files for untranslated English strings."""

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
FILES = ("admin.json", "common.json")
LOCALES = ("ar", "th", "vi")

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Plan-Execute-Replan",
    "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA", "ISO", "Chatbot", "Agent", "Cursor",
    "Superadmin", "Super Admin", "Admin", "ID", "Qty", "Ctrl+K", "⌘K",
}

PRIORITY = (
    "admin.nav.", "admin.dashboard.", "admin.products.", "admin.inventory.",
    "admin.categories.", "admin.coupons.", "admin.shippingRates.", "admin.taxRates.",
    "admin.status_options.", "enum.",
)


def flatten(obj, prefix=""):
    out = {}
    for key, value in obj.items():
        path = f"{prefix}.{key}" if prefix else key
        if isinstance(value, dict):
            out.update(flatten(value, path))
        else:
            out[path] = value
    return out


def is_untranslated(key: str, en_val: str, loc_val: str) -> bool:
    if loc_val == en_val:
        return True
    if not isinstance(loc_val, str) or len(loc_val) < 3:
        return False
    # Strip glossary and placeholders
    cleaned = loc_val
    for term in GLOSSARY:
        cleaned = re.sub(rf"\b{re.escape(term)}\b", "", cleaned)
    cleaned = re.sub(r"\{[a-zA-Z0-9_]+\}", "", cleaned)
    cleaned = re.sub(r"[^A-Za-z]", "", cleaned)
    if len(cleaned) >= 8:
        latin = len(re.findall(r"[A-Za-z]", loc_val))
        if latin / max(len(loc_val), 1) > 0.6:
            return True
    return False


def main():
    en_flat = {}
    for fname in FILES:
        data = json.loads((ROOT / "en" / fname).read_text(encoding="utf-8"))
        for k, v in flatten(data).items():
            en_flat[f"{fname}:{k}"] = v

    for loc in LOCALES:
        for fname in FILES:
            data = json.loads((ROOT / loc / fname).read_text(encoding="utf-8"))
            flat = flatten(data)
            untrans = []
            pri_untrans = []
            for k, v in flat.items():
                ek = f"{fname}:{k}"
                en_v = en_flat.get(ek, "")
                if is_untranslated(k, en_v, v):
                    untrans.append((k, v))
                    if any(p in k for p in PRIORITY):
                        pri_untrans.append((k, v))
            print(f"{loc}/{fname}: {len(untrans)} untranslated ({len(pri_untrans)} priority)")
            for k, v in pri_untrans[:20]:
                print(f"  {k}: {v}")
            if len(pri_untrans) > 20:
                print(f"  ... +{len(pri_untrans) - 20} more")


if __name__ == "__main__":
    main()
