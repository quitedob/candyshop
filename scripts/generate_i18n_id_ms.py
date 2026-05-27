#!/usr/bin/env python3
"""Generate fully translated id/ms admin.json and common.json from en reference."""

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
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache-id-ms.json"
LOCALES = ("id", "ms")
FILES = ("admin.json",)
SOURCE = "en"
WORKERS = 4

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Temperature",
    "Plan-Execute-Replan", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA", "ISO",
    "CSV", "YAML", "Stripe", "PayPal", "Chatbot", "Agent", "Cursor", "Markdown",
    "Super", "Admin", "Superadmin", "Diff", "Trade", "Agent", "translate_content",
}

PLACEHOLDER_RE = re.compile(r"(\{[a-zA-Z0-9_]+\})")
TOKEN_RE = re.compile(
    r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|1200×630|aʷ|→|…|\.\.\.|&|[0-9]+(?:\.[0-9]+)?%?|[A-Za-z][A-Za-z0-9+#./\-]{1,})"
)

LANG_CODES = {
    "id": ("en-GB", "id-ID"),
    "ms": ("en-GB", "ms-MY"),
}

MANUAL_OVERRIDES: dict[str, dict[str, str]] = {
    "id": {
        "admin.brand": "CandyPro Admin",
        "admin.logout": "Keluar",
        "admin.langLabel": "ID",
        "admin.langLabelShort": "ID",
        "admin.langTitle": "Bahasa Indonesia",
        "admin.nav.dashboard": "Dasbor",
        "admin.nav.users": "Pengguna",
        "admin.nav.inquiries": "Permintaan Penawaran",
        "admin.nav.orders": "Pesanan",
        "admin.nav.returns": "Retur",
        "admin.nav.products": "Produk",
        "admin.nav.categories": "Kategori",
        "admin.nav.companies": "Perusahaan",
        "admin.nav.trades": "Perdagangan",
        "admin.nav.pricing": "Harga",
        "admin.nav.oemProjects": "Proyek OEM",
        "admin.nav.content": "Konten",
        "admin.nav.inventory": "Inventaris",
        "admin.nav.xlsx": "Alat XLSX",
        "admin.nav.shipments": "Pengiriman",
        "admin.nav.invoices": "Faktur",
        "admin.nav.management": "Manajemen",
        "admin.nav.superadmin": "Super Admin",
        "admin.nav.analytics": "Analitik",
        "admin.nav.financial": "Keuangan",
        "admin.nav.certifications": "Sertifikasi",
        "admin.nav.staff": "Staf",
        "admin.nav.auditLog": "Log Audit",
        "admin.nav.settings": "Pengaturan",
        "admin.nav.translations": "Terjemahan",
        "admin.nav.ai": "Aplikasi AI",
        "admin.nav.organizations": "Organisasi Pembeli",
        "admin.nav.channels": "Saluran",
        "admin.nav.webhooks": "Webhook",
        "admin.nav.hooks": "Hook & Acara",
        "admin.nav.coupons": "Kupon",
        "admin.nav.shipping_rates": "Pengiriman & Pajak",
        "admin.nav.tax_rates": "Tarif Pajak",
        "admin.dashboard.title": "Dasbor",
        "admin.dashboard.subtitle": "Ikhtisar platform dan metrik utama sekilas.",
        "admin.dashboard.loading": "Memuat statistik dasbor...",
        "admin.dashboard.error": "Gagal memuat dasbor",
        "admin.dashboard.total_users": "Total Pengguna",
        "admin.dashboard.total_orders": "Total Pesanan",
        "admin.dashboard.total_inquiries": "Total Permintaan",
        "admin.dashboard.total_revenue": "Total Pendapatan",
        "admin.dashboard.this_week": "minggu ini",
        "admin.dashboard.new_users_this_week": "{count} baru minggu ini",
        "admin.dashboard.pending": "menunggu",
        "admin.dashboard.recent_activity": "Aktivitas Terbaru",
        "admin.dashboard.col_activity": "Aktivitas",
        "admin.dashboard.col_time": "Waktu",
        "admin.dashboard.no_activity": "Tidak ada aktivitas terbaru",
        "admin.dashboard.revenue_this_month": "Pendapatan Bulan Ini",
        "admin.dashboard.active_customers": "Pelanggan Aktif",
        "admin.dashboard.avg_order_value": "Nilai Pesanan Rata-rata",
        "admin.dashboard.conversion_rate": "Tingkat Konversi",
        "admin.dashboard.revenue_chart": "Pendapatan (30 Hari)",
        "admin.dashboard.revenue": "Pendapatan",
        "admin.dashboard.date": "Tanggal",
        "admin.products.title": "Produk",
        "admin.products.description": "Kelola katalog produk dengan operasi CRUD lengkap.",
        "admin.inventory.title": "Inventaris",
        "admin.inventory.description": "Lacak dan kelola tingkat stok, nilai, dan penyesuaian produk.",
        "admin.categories.title": "Kategori",
        "admin.categories.description": "Kelola kategori produk dengan nama, alias, dan deskripsi multibahasa.",
        "admin.coupons.title": "Kupon & Kartu Hadiah",
        "admin.coupons.description": "Kode promosi dan kredit prabayar.",
        "admin.shippingRates.title": "Tarif Pengiriman",
        "admin.shippingRates.description": "Tabel tarif pengiriman berdasarkan tujuan. Perkiraan hari hanya referensi penawaran — bukan jaminan waktu transit.",
        "admin.taxRates.title": "Tarif Pajak",
        "admin.taxRates.description": "Tarif pajak negara/wilayah tujuan yang diterapkan saat checkout keranjang.",
        "admin.status_options.draft": "Draf",
        "admin.status_options.sent": "Terkirim",
        "admin.status_options.paid": "Lunas",
        "admin.status_options.overdue": "Jatuh tempo",
        "admin.status_options.cancelled": "Dibatalkan",
        "admin.status_options.active": "Aktif",
        "admin.status_options.inactive": "Nonaktif",
        "admin.products.ai_translate_slow": "Terjemahan membutuhkan waktu lebih lama dari biasanya. Harap tunggu atau coba lagi nanti.",
        "admin.products.locale_tab_zh": "Tionghoa",
        "admin.products.locale_tab_en": "Inggris",
        "admin.products.locale_tab_ar": "Arab",
        "admin.products.locale_tab_ko": "Korea",
        "admin.products.locale_tab_ja": "Jepang",
        "admin.products.locale_tab_zh-tw": "Tionghoa Tradisional",
        "admin.products.locale_tab_th": "Thailand",
        "admin.products.locale_tab_vi": "Vietnam",
        "admin.products.locale_tab_id": "Indonesia",
        "admin.products.locale_tab_ms": "Melayu",
        "admin.taxRates.col_country": "Negara",
        "admin.taxRates.col_region": "Wilayah",
        "admin.coupons.showing": "Total {total}",
        "admin.coupons.create": "Buat Kupon",
        "admin.coupons.create_giftcard": "Buat Kartu Hadiah",
        "admin.coupons.no_data": "Tidak ada kupon.",
        "admin.coupons.no_giftcards": "Tidak ada kartu hadiah.",
        "admin.coupons.tab_coupons": "Kupon",
        "admin.coupons.tab_giftcards": "Kartu Hadiah",
        "admin.coupons.type_percentage": "Persentase",
        "admin.coupons.type_fixed": "Jumlah Tetap",
        "admin.coupons.min_order": "Pesanan Min.",
        "admin.coupons.starts_at": "Mulai",
        "admin.coupons.expires_at": "Berakhir",
        "admin.coupons.col_used": "Digunakan",
        "admin.coupons.col_initial": "Awal",
        "admin.coupons.col_balance": "Saldo",
        "admin.coupons.currency": "Mata uang",
        "admin.coupons.col_currency": "Mata uang",
        "admin.shippingRates.create": "Tambah Tarif",
        "admin.shippingRates.no_data": "Tidak ada tarif pengiriman.",
        "admin.shippingRates.destination": "Tujuan",
        "admin.shippingRates.carrier": "Kurir",
        "admin.shippingRates.min_weight": "Berat Min. (kg)",
        "admin.shippingRates.max_weight": "Berat Max. (kg)",
        "admin.shippingRates.base_cost": "Biaya Dasar",
        "admin.shippingRates.cost_per_kg": "Biaya per kg",
        "admin.shippingRates.currency": "Mata uang",
        "admin.shippingRates.estimated_days": "Perkiraan Hari",
        "admin.shippingRates.estimated_days_hint": "Perkiraan transit referensi untuk penawaran internal. Dapat berubah; konfirmasi saat pesanan.",
        "admin.shippingRates.col_destination": "Tujuan",
        "admin.shippingRates.col_carrier": "Kurir",
        "admin.shippingRates.col_base_cost": "Biaya Dasar",
        "admin.shippingRates.col_days": "Hari",
        "admin.taxRates.create": "Tambah Tarif Pajak",
        "admin.taxRates.no_data": "Tidak ada tarif pajak.",
        "admin.taxRates.country": "Negara (ISO)",
        "admin.taxRates.region": "Wilayah / Provinsi",
        "admin.taxRates.rate": "Tarif",
        "admin.taxRates.rate_hint": "Pecahan desimal, mis. 0,08875 = 8,875%",
        "admin.nav.organizations": "Organisasi",
    },
    "ms": {
        "admin.brand": "CandyPro Admin",
        "admin.logout": "Log keluar",
        "admin.langLabel": "MS",
        "admin.langLabelShort": "MS",
        "admin.langTitle": "Bahasa Melayu",
        "admin.nav.dashboard": "Papan Pemuka",
        "admin.nav.users": "Pengguna",
        "admin.nav.inquiries": "Pertanyaan",
        "admin.nav.orders": "Pesanan",
        "admin.nav.returns": "Pemulangan",
        "admin.nav.products": "Produk",
        "admin.nav.categories": "Kategori",
        "admin.nav.companies": "Syarikat",
        "admin.nav.trades": "Perdagangan",
        "admin.nav.pricing": "Harga",
        "admin.nav.oemProjects": "Projek OEM",
        "admin.nav.content": "Kandungan",
        "admin.nav.inventory": "Inventori",
        "admin.nav.xlsx": "Alat XLSX",
        "admin.nav.shipments": "Penghantaran",
        "admin.nav.invoices": "Invois",
        "admin.nav.management": "Pengurusan",
        "admin.nav.superadmin": "Super Admin",
        "admin.nav.analytics": "Analitik",
        "admin.nav.financial": "Kewangan",
        "admin.nav.certifications": "Pensijilan",
        "admin.nav.staff": "Kakitangan",
        "admin.nav.auditLog": "Log Audit",
        "admin.nav.settings": "Tetapan",
        "admin.nav.translations": "Terjemahan",
        "admin.nav.ai": "Aplikasi AI",
        "admin.nav.organizations": "Organisasi Pembeli",
        "admin.nav.channels": "Saluran",
        "admin.nav.webhooks": "Webhook",
        "admin.nav.hooks": "Hook & Acara",
        "admin.nav.coupons": "Kupon",
        "admin.nav.shipping_rates": "Penghantaran & Cukai",
        "admin.nav.tax_rates": "Kadar Cukai",
        "admin.dashboard.title": "Papan Pemuka",
        "admin.dashboard.subtitle": "Gambaran platform dan metrik utama sepintas lalu.",
        "admin.dashboard.loading": "Memuatkan statistik papan pemuka...",
        "admin.dashboard.error": "Ralat memuatkan papan pemuka",
        "admin.dashboard.total_users": "Jumlah Pengguna",
        "admin.dashboard.total_orders": "Jumlah Pesanan",
        "admin.dashboard.total_inquiries": "Jumlah Pertanyaan",
        "admin.dashboard.total_revenue": "Jumlah Hasil",
        "admin.dashboard.this_week": "minggu ini",
        "admin.dashboard.new_users_this_week": "{count} baharu minggu ini",
        "admin.dashboard.pending": "menunggu",
        "admin.dashboard.recent_activity": "Aktiviti Terkini",
        "admin.dashboard.col_activity": "Aktiviti",
        "admin.dashboard.col_time": "Masa",
        "admin.dashboard.no_activity": "Tiada aktiviti terkini",
        "admin.dashboard.revenue_this_month": "Hasil Bulan Ini",
        "admin.dashboard.active_customers": "Pelanggan Aktif",
        "admin.dashboard.avg_order_value": "Nilai Pesanan Purata",
        "admin.dashboard.conversion_rate": "Kadar Penukaran",
        "admin.dashboard.revenue_chart": "Hasil (30 Hari)",
        "admin.dashboard.revenue": "Hasil",
        "admin.dashboard.date": "Tarikh",
        "admin.products.title": "Produk",
        "admin.products.description": "Urus katalog produk dengan operasi CRUD lengkap.",
        "admin.inventory.title": "Inventori",
        "admin.inventory.description": "Jejak dan urus tahap stok, nilai, dan pelarasan produk.",
        "admin.categories.title": "Kategori",
        "admin.categories.description": "Urus kategori produk dengan nama, alias, dan penerangan pelbagai bahasa.",
        "admin.coupons.title": "Kupon & Kad Hadiah",
        "admin.coupons.description": "Kod promosi dan kredit prabayar.",
        "admin.shippingRates.title": "Kadar Penghantaran",
        "admin.shippingRates.description": "Jadual kadar penghantaran mengikut destinasi. Anggaran hari adalah rujukan sebut harga — bukan jaminan masa transit.",
        "admin.taxRates.title": "Kadar Cukai",
        "admin.taxRates.description": "Kadar cukai negara/wilayah destinasi yang digunakan semasa checkout troli.",
        "admin.status_options.draft": "Draf",
        "admin.status_options.sent": "Dihantar",
        "admin.status_options.paid": "Dibayar",
        "admin.status_options.overdue": "Lewat tempoh",
        "admin.status_options.cancelled": "Dibatalkan",
        "admin.status_options.active": "Aktif",
        "admin.status_options.inactive": "Tidak aktif",
        "admin.products.ai_translate_slow": "Terjemahan mengambil masa lebih lama daripada biasa. Sila tunggu atau cuba lagi kemudian.",
        "admin.products.locale_tab_zh": "Cina",
        "admin.products.locale_tab_en": "Inggeris",
        "admin.products.locale_tab_ar": "Arab",
        "admin.products.locale_tab_ko": "Korea",
        "admin.products.locale_tab_ja": "Jepun",
        "admin.products.locale_tab_zh-tw": "Cina Tradisional",
        "admin.products.locale_tab_th": "Thai",
        "admin.products.locale_tab_vi": "Vietnam",
        "admin.products.locale_tab_id": "Indonesia",
        "admin.products.locale_tab_ms": "Melayu",
        "admin.taxRates.col_country": "Negara",
        "admin.taxRates.col_region": "Kawasan",
        "admin.coupons.showing": "Jumlah {total}",
        "admin.coupons.create": "Cipta Kupon",
        "admin.coupons.create_giftcard": "Cipta Kad Hadiah",
        "admin.coupons.no_data": "Tiada kupon.",
        "admin.coupons.no_giftcards": "Tiada kad hadiah.",
        "admin.coupons.tab_coupons": "Kupon",
        "admin.coupons.tab_giftcards": "Kad Hadiah",
        "admin.coupons.title": "Kupon & Kad Hadiah",
        "admin.coupons.type_percentage": "Peratusan",
        "admin.coupons.type_fixed": "Jumlah Tetap",
        "admin.coupons.min_order": "Pesanan Min.",
        "admin.coupons.starts_at": "Mula",
        "admin.coupons.expires_at": "Tamat",
        "admin.coupons.col_used": "Digunakan",
        "admin.coupons.col_initial": "Permulaan",
        "admin.coupons.col_balance": "Baki",
        "admin.coupons.currency": "Mata wang",
        "admin.coupons.col_currency": "Mata wang",
        "admin.shippingRates.title": "Kadar Penghantaran",
        "admin.shippingRates.description": "Jadual kadar penghantaran mengikut destinasi. Anggaran hari adalah rujukan sebut harga — bukan jaminan masa transit.",
        "admin.shippingRates.create": "Tambah Kadar",
        "admin.shippingRates.no_data": "Tiada kadar penghantaran.",
        "admin.shippingRates.destination": "Destinasi",
        "admin.shippingRates.carrier": "Pengangkut",
        "admin.shippingRates.min_weight": "Berat Min. (kg)",
        "admin.shippingRates.max_weight": "Berat Maks. (kg)",
        "admin.shippingRates.base_cost": "Kos Asas",
        "admin.shippingRates.cost_per_kg": "Kos per kg",
        "admin.shippingRates.currency": "Mata wang",
        "admin.shippingRates.estimated_days": "Anggaran Hari",
        "admin.shippingRates.estimated_days_hint": "Anggaran transit rujukan untuk sebut harga dalaman. Tertakluk perubahan; sahkan semasa pesanan.",
        "admin.shippingRates.col_destination": "Destinasi",
        "admin.shippingRates.col_carrier": "Pengangkut",
        "admin.shippingRates.col_base_cost": "Kos Asas",
        "admin.shippingRates.col_days": "Hari",
        "admin.taxRates.create": "Tambah Kadar Cukai",
        "admin.taxRates.no_data": "Tiada kadar cukai.",
        "admin.taxRates.country": "Negara (ISO)",
        "admin.taxRates.region": "Kawasan / Negeri",
        "admin.taxRates.rate": "Kadar",
        "admin.taxRates.rate_hint": "Pecahan perpuluhan, cth. 0.08875 = 8.875%",
        "admin.nav.organizations": "Organisasi",
    },
}

_cache_lock = threading.Lock()
_save_counter = 0
all_caches: dict[str, dict[str, str]] = {}


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


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def save_json(path: Path, data: dict) -> None:
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def load_cache() -> dict[str, dict[str, str]]:
    if CACHE_PATH.exists():
        return json.loads(CACHE_PATH.read_text(encoding="utf-8"))
    return {loc: {} for loc in LOCALES}


def save_cache(cache: dict[str, dict[str, str]]) -> None:
    CACHE_PATH.write_text(json.dumps(cache, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def translate_one(text: str, locale: str) -> str:
    src, tgt = LANG_CODES[locale]
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source=src, target=tgt)
    for attempt in range(5):
        try:
            result = translator.translate(protected)
            return restore(result or text, replacements)
        except Exception:
            time.sleep(0.8 * (attempt + 1))
    return text


def translate_strings(strings: list[str], locale: str, cache: dict[str, str]) -> None:
    global _save_counter
    todo = [s for s in strings if s not in cache]
    total = len(todo)
    if total == 0:
        print(f"  {locale}: all {len(strings)} cached", flush=True)
        return

    print(f"  {locale}: translating {total} new strings ({len(strings) - total} cached)...", flush=True)
    done = 0

    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futures = {pool.submit(translate_one, text, locale): text for text in todo}
        for future in as_completed(futures):
            text = futures[future]
            try:
                result = future.result()
            except Exception:
                result = text
            with _cache_lock:
                cache[text] = result
                _save_counter += 1
                if _save_counter % 50 == 0:
                    save_cache(all_caches)
            done += 1
            if done % 50 == 0 or done == total:
                print(f"    {locale}: {done}/{total}", flush=True)


def build_locale_file(
    en_tree: dict,
    value_map: dict[str, str],
    overrides: dict[str, str],
) -> dict:
    en_flat = flatten(en_tree)
    out_flat: dict[str, str] = {}

    for key, en_val in en_flat.items():
        if key in overrides:
            out_flat[key] = overrides[key]
        else:
            out_flat[key] = value_map.get(en_val, en_val)

    return unflatten(out_flat)


def main() -> None:
    global all_caches
    sys.stdout.reconfigure(encoding="utf-8")

    all_caches = load_cache()
    for loc in LOCALES:
        all_caches.setdefault(loc, {})

    en_data = {f: load_json(ROOT / SOURCE / f) for f in FILES}

    for locale in LOCALES:
        needed: set[str] = set()
        for filename in FILES:
            en_flat = flatten(en_data[filename])
            for _key, en_val in en_flat.items():
                needed.add(en_val)
        print(f"Locale {locale}: {len(needed)} strings needed", flush=True)
        translate_strings(sorted(needed, key=len), locale, all_caches[locale])

    save_cache(all_caches)

    for locale in LOCALES:
        overrides = MANUAL_OVERRIDES.get(locale, {})
        for filename in FILES:
            existing_path = ROOT / locale / filename
            file_overrides = overrides if filename == "admin.json" else {}
            translated = build_locale_file(
                en_data[filename],
                all_caches[locale],
                file_overrides,
            )
            save_json(existing_path, translated)
            print(f"Wrote {locale}/{filename}", flush=True)

    print("Done.", flush=True)


if __name__ == "__main__":
    main()
