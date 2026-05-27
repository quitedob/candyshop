#!/usr/bin/env python3
"""Regenerate translation data blocks for i18n_build_id_ms_offline.py."""
from __future__ import annotations

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPT = Path(__file__).resolve().parent / "i18n_build_id_ms_offline.py"
I18N = ROOT / "frontend" / "i18n"
TRANSLATE_SRC = Path(__file__).resolve().parent / "i18n_translate_id_ms.py"

PRIORITY_PREFIXES = (
    "admin.nav.",
    "admin.dashboard.",
    "admin.products.",
    "admin.inventory.",
    "admin.categories.",
    "admin.coupons.",
    "admin.shippingRates.",
    "admin.taxRates.",
    "admin.status_options.",
    "enum.",
)

GLOSSARY_KEEP = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT", "COGS",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "GTIN", "HS",
    "Ctrl", "Mac", "Webhook", "Webhooks", "3PL", "RCEP", "CBM", "CandyPro",
    "Facebook", "LinkedIn", "Instagram", "YouTube", "Twitter", "HACCP", "ISO22000",
    "BRC", "HALAL", "GMO", "N/A", "TBD", "Proforma", "Incoterms", "B2B", "KYB",
    "RFM", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA", "ISO", "CSV", "Stripe",
    "PayPal", "Chatbot", "Agent", "Markdown", "Super Admin", "Superadmin", "Slug",
    "Non-GMO", "Halal", "Vegan", "Kosher", "Organic", "Edit", "Admin", "Email",
    "Status", "Role", "ID", "EN", "MS", "Thumbnail", "WhatsApp", "RFM", "Qty",
    "3M", "6M", "12M", "#", "0", "global", "USD", "flow-wrap", "foil", "box", "bag", "jar",
}


def flatten(obj: dict, prefix: str = "") -> dict[str, str]:
    out: dict[str, str] = {}
    for key, value in obj.items():
        path = f"{prefix}.{key}" if prefix else key
        if isinstance(value, dict):
            out.update(flatten(value, path))
        else:
            out[path] = value
    return out


# Import phrase data from first generator module content
from _gen_id_ms_data import WORD_PAIRS, WORDS  # noqa: E402


def build_exact_maps() -> tuple[dict[str, str], dict[str, str]]:
    exact_id = {en: id_v for en, id_v, _ in WORD_PAIRS}
    exact_ms = {en: ms_v for en, _, ms_v in WORD_PAIRS}
    for en, id_v, ms_v in WORDS:
        if len(en) > 20 or ". " in en or en.endswith((".", "...", "…", "?")):
            exact_id.setdefault(en, id_v)
            exact_ms.setdefault(en, ms_v)
    return exact_id, exact_ms


def build_lexicons() -> tuple[dict[str, str], dict[str, str]]:
    id_lex = {en: id_v for en, id_v, _ in WORDS if len(en) <= 20 and ". " not in en}
    ms_lex = {en: ms_v for en, _, ms_v in WORDS if len(en) <= 20 and ". " not in en}
    return id_lex, ms_lex


def protect(text: str) -> tuple[str, list[tuple[str, str]]]:
    token_re = re.compile(
        r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|1200×630|aʷ|→|…|\.\.\.|&|[0-9]+(?:\.[0-9]+)?%?)"
    )
    replacements: list[tuple[str, str]] = []

    def repl(match: re.Match[str]) -> str:
        token = match.group(0)
        placeholder = f"__TOK{len(replacements)}__"
        replacements.append((placeholder, token))
        return placeholder

    return token_re.sub(repl, text), replacements


def restore(text: str, replacements: list[tuple[str, str]]) -> str:
    for placeholder, original in replacements:
        text = text.replace(placeholder, original)
    return text


def apply_lexicon(text: str, lexicon: dict[str, str]) -> str:
    protected, replacements = protect(text)
    result = protected
    for phrase, translation in sorted(lexicon.items(), key=lambda item: len(item[0]), reverse=True):
        if phrase in result:
            result = result.replace(phrase, translation)
    return restore(result, replacements)


PATTERNS = {
    "id": [
        (r"^Loading (.+)\.\.\.$", r"Memuat \1..."),
        (r"^Loading (.+)…$", r"Memuat \1…"),
        (r"^Error loading (.+)$", r"Gagal memuat \1"),
        (r"^Failed to load (.+)$", r"Gagal memuat \1"),
        (r"^No (.+)\.$", r"Tidak ada \1."),
        (r"^No (.+)$", r"Tidak ada \1"),
        (r"^Manage (.+)$", r"Kelola \1"),
        (r"^Add (.+)$", r"Tambah \1"),
        (r"^Create (.+)$", r"Buat \1"),
        (r"^Edit (.+)$", r"Edit \1"),
        (r"^Delete (.+)$", r"Hapus \1"),
        (r"^Update (.+)$", r"Perbarui \1"),
        (r"^Save (.+)$", r"Simpan \1"),
        (r"^Could not (.+)\. Please try again\.$", r"Tidak dapat \1. Silakan coba lagi."),
        (r"^A list of all (.+)$", r"Daftar semua \1"),
    ],
    "ms": [
        (r"^Loading (.+)\.\.\.$", r"Memuatkan \1..."),
        (r"^Loading (.+)…$", r"Memuatkan \1…"),
        (r"^Error loading (.+)$", r"Ralat memuatkan \1"),
        (r"^Failed to load (.+)$", r"Gagal memuatkan \1"),
        (r"^No (.+)\.$", r"Tiada \1."),
        (r"^No (.+)$", r"Tiada \1"),
        (r"^Manage (.+)$", r"Urus \1"),
        (r"^Add (.+)$", r"Tambah \1"),
        (r"^Create (.+)$", r"Cipta \1"),
        (r"^Edit (.+)$", r"Edit \1"),
        (r"^Delete (.+)$", r"Padam \1"),
        (r"^Update (.+)$", r"Kemas kini \1"),
        (r"^Save (.+)$", r"Simpan \1"),
        (r"^Could not (.+)\. Please try again\.$", r"Tidak dapat \1. Sila cuba lagi."),
        (r"^A list of all (.+)$", r"Senarai semua \1"),
    ],
}


def translate_value(text: str, locale: str, exact: dict[str, str], lexicon: dict[str, str]) -> str:
    if text in exact:
        return exact[text]
    result = apply_lexicon(text, lexicon)
    for pattern, repl in PATTERNS[locale]:
        result = re.sub(pattern, repl, result)
    return result


def load_overrides_block() -> str:
    from _gen_id_ms_data import OVERRIDES  # noqa: WPS433

    src = TRANSLATE_SRC.read_text(encoding="utf-8")
    enum_start = src.index("_enum_id = {")
    enum_end = src.index("# Brand carriers kept as-is")
    enum_section = src[enum_start:enum_end]
    lines = [
        'PATH_OVERRIDES["id"].clear(); PATH_OVERRIDES["ms"].clear()',
        OVERRIDES.strip(),
        enum_section.strip(),
        'for c in ["carrier.dhl", "carrier.fedex", "carrier.ups", "carrier.usps", "carrier.ems"]:',
        '    brand = {"carrier.dhl": "DHL", "carrier.fedex": "FedEx", "carrier.ups": "UPS", "carrier.usps": "USPS", "carrier.ems": "EMS"}[c]',
        '    _add(f"enum.{c}", brand, brand)',
        '_add("enum.inquiry_priority.normal", "Biasa", "Biasa")',
        '_add("enum.carrier.standard", "Standar", "Standard")',
        '_add("enum.tracking_provider.standard", "Standar", "Standard")',
    ]
    return "\n".join(lines)


def main() -> None:
    id_lex, ms_lex = build_lexicons()
    exact_id, exact_ms = build_exact_maps()

    en_admin = json.loads((I18N / "en" / "admin.json").read_text(encoding="utf-8"))
    en_common = json.loads((I18N / "en" / "common.json").read_text(encoding="utf-8"))
    en_flat = {**flatten(en_admin), **flatten(en_common)}
    all_values = sorted(set(en_flat.values()))

    # Expand exact maps for every unique string
    for val in all_values:
        if val not in exact_id:
            exact_id[val] = translate_value(val, "id", exact_id, id_lex)
        if val not in exact_ms:
            exact_ms[val] = translate_value(val, "ms", exact_ms, ms_lex)

    SECTION_ONLY_PREFIXES = (
        "admin.products.",
        "admin.inventory.",
        "admin.categories.",
        "admin.coupons.",
        "admin.shippingRates.",
        "admin.taxRates.",
    )

    base = SCRIPT.read_text(encoding="utf-8").split("OVERRIDES_BLOCK =")[0].rstrip() + "\n\n"
    lines = [base]
    lines.append('OVERRIDES_BLOCK = """')
    lines.append(load_overrides_block().strip())
    lines.append('"""')
    lines.append("")
    lines.append("EXACT_ENTRIES = [")
    for val in all_values:
        lines.append(f"    ({val!r}, {exact_id[val]!r}, {exact_ms[val]!r}),")
    lines.append("]")
    lines.append("")
    lines.append("LEXICON_ID = {")
    for en, id_v, _ in WORDS:
        lines.append(f"    {en!r}: {id_v!r},")
    lines.append("}")
    lines.append("")
    lines.append("LEXICON_MS = {")
    for en, _, ms_v in WORDS:
        lines.append(f"    {en!r}: {ms_v!r},")
    lines.append("}")
    lines.append("")
    lines.append("PATTERNS = {")
    for locale, rules in PATTERNS.items():
        lines.append(f'    "{locale}": [')
        for pattern, repl in rules:
            lines.append(f"        ({pattern!r}, {repl!r}),")
        lines.append("    ],")
    lines.append("}")
    lines.append("")
    lines.append(
        '''def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    register_path_overrides()
    register_exact_strings()
    en_data = {f: load_json(I18N / "en" / f) for f in FILES}
    for locale in LOCALES:
        (I18N / locale).mkdir(parents=True, exist_ok=True)
    results = []
    for filename in FILES:
        for locale in LOCALES:
            translated = build_locale(en_data[filename], locale)
            out_path = I18N / locale / filename
            save_json(out_path, translated)
            report = validate(en_data[filename], translated, locale)
            report["file"] = filename
            results.append(report)
            print(f"Wrote {locale}/{filename} ({report['leaf_count']} leaves, {report['same_as_en']} same as en)")
    print("\\n=== Spot check ===")
    for locale in LOCALES:
        admin = load_json(I18N / locale / "admin.json")
        common = load_json(I18N / locale / "common.json")
        samples = [
            ("nav.dashboard", admin["admin"]["nav"]["dashboard"]),
            ("dashboard.title", admin["admin"]["dashboard"]["title"]),
            ("status_options.active", admin["admin"]["status_options"]["active"]),
            ("enum.order_status.pending", common["enum"]["order_status"]["pending"]),
        ]
        print(f"  [{locale}] " + " | ".join(f"{k}={v}" for k, v in samples))
    print("\\n=== Priority English leftovers ===")
    for r in results:
        leftovers = r["priority_english_leftovers"]
        print(f"{r['locale']}/{r['file']}: {len(leftovers)} priority English leftovers")
        for item in leftovers[:8]:
            print(f"  - {item}")
        if len(leftovers) > 8:
            print(f"  ... and {len(leftovers) - 8} more")


if __name__ == "__main__":
    main()
'''
    )
    SCRIPT.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"Regenerated {SCRIPT} with {len(all_values)} exact entries")


if __name__ == "__main__":
    main()
