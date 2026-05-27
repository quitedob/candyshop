#!/usr/bin/env python3
"""Generate id/ms locale files from en with manual overrides + cached MT."""
from __future__ import annotations

import json
import re
import time
from copy import deepcopy
from pathlib import Path

from deep_translator import MyMemoryTranslator

ROOT = Path(__file__).resolve().parents[1]
I18N = ROOT / "frontend" / "i18n"
CACHE_DIR = ROOT / "scripts" / "i18n_cache"
CACHE_DIR.mkdir(exist_ok=True)

TARGET_MAP = {"id": "id-ID", "ms": "ms-MY"}

PLACEHOLDER_RE = re.compile(
    r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|\d+×\d+|1200×630|aʷ|PI|CI|SC|B/L|MOQ|OEM|SKU|COGS|KYB|"
    r"GTIN|HS|GMO|RCEP|FORM E|HACCP|ISO22000|BRC|HALAL|USD|FOB|CBM|3PL|DHL|FedEx|UPS|USPS|EMS|"
    r"WhatsApp|Facebook|LinkedIn|Instagram|YouTube|Twitter|XLSX|DOCX|JSON|HTTP|URL|API|JWT|"
    r"N/A|TBD|EN|ID|MS|zh|en/ko/ar/ja/th/vi/id/ms|translate_content|Plan-Execute-Replan|"
    r"CandyPro|Incoterms|Proforma|ODM|Webhook|Webhooks|Hooks|Super Admin|Superadmin|RFM|"
    r"Diff|Cursor-style|Trade Agent|Chatbot|B2B|AI|Agent|Mac:|Windows)"
)

MANUAL_ID: dict[str, str] = {}
MANUAL_MS: dict[str, str] = {}


def _add(prefix: str, id_val: str, ms_val: str) -> None:
    MANUAL_ID[prefix] = id_val
    MANUAL_MS[prefix] = ms_val


_nav_id = {
    "dashboard": "Dasbor",
    "users": "Pengguna",
    "inquiries": "Permintaan Penawaran",
    "orders": "Pesanan",
    "returns": "Retur",
    "products": "Produk",
    "categories": "Kategori",
    "companies": "Perusahaan",
    "trades": "Perdagangan",
    "pricing": "Harga",
    "oemProjects": "Proyek OEM",
    "content": "Konten",
    "inventory": "Inventaris",
    "xlsx": "Alat XLSX",
    "shipments": "Pengiriman",
    "invoices": "Faktur",
    "management": "Manajemen",
    "superadmin": "Super Admin",
    "analytics": "Analitik",
    "financial": "Keuangan",
    "certifications": "Sertifikasi",
    "staff": "Staf",
    "auditLog": "Log Audit",
    "settings": "Pengaturan",
    "translations": "Terjemahan",
    "ai": "Aplikasi AI",
    "organizations": "Organisasi Pembeli",
    "channels": "Saluran",
    "webhooks": "Webhook",
    "hooks": "Hook & Acara",
    "coupons": "Kupon",
    "shipping_rates": "Pengiriman & Pajak",
    "tax_rates": "Tarif Pajak",
}
_nav_ms = {
    "dashboard": "Papan Pemuka",
    "users": "Pengguna",
    "inquiries": "Pertanyaan",
    "orders": "Pesanan",
    "returns": "Pemulangan",
    "products": "Produk",
    "categories": "Kategori",
    "companies": "Syarikat",
    "trades": "Perdagangan",
    "pricing": "Harga",
    "oemProjects": "Projek OEM",
    "content": "Kandungan",
    "inventory": "Inventori",
    "xlsx": "Alat XLSX",
    "shipments": "Penghantaran",
    "invoices": "Invois",
    "management": "Pengurusan",
    "superadmin": "Super Admin",
    "analytics": "Analitik",
    "financial": "Kewangan",
    "certifications": "Pensijilan",
    "staff": "Kakitangan",
    "auditLog": "Log Audit",
    "settings": "Tetapan",
    "translations": "Terjemahan",
    "ai": "Aplikasi AI",
    "organizations": "Organisasi Pembeli",
    "channels": "Saluran",
    "webhooks": "Webhook",
    "hooks": "Hook & Acara",
    "coupons": "Kupon",
    "shipping_rates": "Penghantaran & Cukai",
    "tax_rates": "Kadar Cukai",
}
for k, v in _nav_id.items():
    _add(f"admin.nav.{k}", v, _nav_ms[k])

_status_id = {
    "draft": "Draf",
    "sent": "Terkirim",
    "paid": "Lunas",
    "overdue": "Jatuh tempo",
    "cancelled": "Dibatalkan",
    "active": "Aktif",
    "inactive": "Nonaktif",
}
_status_ms = {
    "draft": "Draf",
    "sent": "Dihantar",
    "paid": "Dibayar",
    "overdue": "Lewat tempoh",
    "cancelled": "Dibatalkan",
    "active": "Aktif",
    "inactive": "Tidak aktif",
}
for k, v in _status_id.items():
    _add(f"admin.status_options.{k}", v, _status_ms[k])

_add("admin.langLabel", "ID", "MS")
_add("admin.langLabelShort", "ID", "MS")
_add("admin.langTitle", "Bahasa Indonesia", "Bahasa Melayu")
_add("admin.brand", "CandyPro Admin", "CandyPro Admin")
_add("admin.logout", "Keluar", "Log keluar")

# Priority: dashboard
_add("admin.dashboard.title", "Dasbor", "Papan Pemuka")
_add("admin.dashboard.subtitle", "Ikhtisar platform dan metrik utama sekilas.", "Gambaran platform dan metrik utama sepintas lalu.")
_add("admin.dashboard.loading", "Memuat statistik dasbor...", "Memuatkan statistik papan pemuka...")
_add("admin.dashboard.error", "Gagal memuat dasbor", "Ralat memuatkan papan pemuka")
_add("admin.dashboard.total_users", "Total Pengguna", "Jumlah Pengguna")
_add("admin.dashboard.total_orders", "Total Pesanan", "Jumlah Pesanan")
_add("admin.dashboard.total_inquiries", "Total Permintaan", "Jumlah Pertanyaan")
_add("admin.dashboard.total_revenue", "Total Pendapatan", "Jumlah Hasil")
_add("admin.dashboard.this_week", "minggu ini", "minggu ini")
_add("admin.dashboard.new_users_this_week", "{count} baru minggu ini", "{count} baharu minggu ini")
_add("admin.dashboard.pending", "menunggu", "menunggu")
_add("admin.dashboard.recent_activity", "Aktivitas Terbaru", "Aktiviti Terkini")
_add("admin.dashboard.col_activity", "Aktivitas", "Aktiviti")
_add("admin.dashboard.col_time", "Waktu", "Masa")
_add("admin.dashboard.no_activity", "Tidak ada aktivitas terbaru", "Tiada aktiviti terkini")
_add("admin.dashboard.revenue_this_month", "Pendapatan Bulan Ini", "Hasil Bulan Ini")
_add("admin.dashboard.active_customers", "Pelanggan Aktif", "Pelanggan Aktif")
_add("admin.dashboard.avg_order_value", "Nilai Pesanan Rata-rata", "Nilai Pesanan Purata")
_add("admin.dashboard.conversion_rate", "Tingkat Konversi", "Kadar Penukaran")
_add("admin.dashboard.revenue_chart", "Pendapatan (30 Hari)", "Hasil (30 Hari)")
_add("admin.dashboard.revenue", "Pendapatan", "Hasil")
_add("admin.dashboard.date", "Tanggal", "Tarikh")

# Priority sections - shared UI tokens
for path, id_v, ms_v in [
    ("admin.products.title", "Produk", "Produk"),
    ("admin.inventory.title", "Inventaris", "Inventori"),
    ("admin.categories.title", "Kategori", "Kategori"),
    ("admin.coupons.title", "Kupon & Kartu Hadiah", "Kupon & Kad Hadiah"),
    ("admin.shippingRates.title", "Tarif Pengiriman", "Kadar Penghantaran"),
    ("admin.taxRates.title", "Tarif Pajak", "Kadar Cukai"),
    ("admin.shippingRates.description", "Tabel tarif pengiriman berdasarkan tujuan. Perkiraan hari hanya referensi penawaran — bukan jaminan waktu transit.", "Jadual kadar penghantaran mengikut destinasi. Anggaran hari adalah rujukan sebut harga — bukan jaminan masa transit."),
    ("admin.taxRates.description", "Tarif pajak negara/wilayah tujuan yang diterapkan saat checkout keranjang.", "Kadar cukai negara/wilayah destinasi yang digunakan semasa checkout troli."),
]:
    _add(path, id_v, ms_v)

# enum groups
_enum_id = {
    "order_status.pending": "Menunggu",
    "order_status.confirmed": "Dikonfirmasi",
    "order_status.production": "Produksi",
    "order_status.shipped": "Dikirim",
    "order_status.delivered": "Diterima",
    "order_status.cancelled": "Dibatalkan",
    "order_status.processing": "Diproses",
    "order_status.pending_confirmation": "Menunggu konfirmasi",
    "order_status.pending_approval": "Menunggu persetujuan",
    "order_status.partially_shipped": "Sebagian dikirim",
    "order_status.partially_delivered": "Sebagian diterima",
    "order_status.partially_returned": "Sebagian dikembalikan",
    "order_status.returned": "Dikembalikan",
    "order_status.expired": "Kedaluwarsa",
    "order_status.unknown": "Tidak diketahui",
    "payment_status.unpaid": "Belum dibayar",
    "payment_status.paid": "Lunas",
    "payment_status.partial": "Sebagian",
    "payment_status.refunded": "Dikembalikan",
    "payment_status.pending": "Menunggu",
    "payment_status.confirmed": "Dikonfirmasi",
    "payment_status.unknown": "Tidak diketahui",
    "product_status.active": "Aktif",
    "product_status.inactive": "Nonaktif",
    "product_status.draft": "Draf",
    "product_status.unknown": "Tidak diketahui",
    "company_status.pending": "Menunggu",
    "company_status.verified": "Terverifikasi",
    "company_status.rejected": "Ditolak",
    "company_status.active": "Aktif",
    "company_status.suspended": "Ditangguhkan",
    "company_status.inactive": "Nonaktif",
    "company_status.unknown": "Tidak diketahui",
    "kyb_status.pending": "Menunggu",
    "kyb_status.approved": "Disetujui",
    "kyb_status.rejected": "Ditolak",
    "kyb_status.unknown": "Tidak diketahui",
    "inquiry_status.pending": "Menunggu",
    "inquiry_status.contacted": "Dihubungi",
    "inquiry_status.quoted": "Dikutip",
    "inquiry_status.negotiating": "Negosiasi",
    "inquiry_status.won": "Menang",
    "inquiry_status.lost": "Kalah",
    "inquiry_status.open": "Terbuka",
    "inquiry_status.closed": "Ditutup",
    "inquiry_status.converted": "Dikonversi",
    "inquiry_status.unknown": "Tidak diketahui",
    "inquiry_priority.normal": "Normal",
    "inquiry_priority.high": "Tinggi",
    "inquiry_priority.low": "Rendah",
    "inquiry_priority.urgent": "Mendesak",
    "inquiry_priority.unknown": "Tidak diketahui",
    "trade_status.draft": "Draf",
    "trade_status.pending": "Menunggu",
    "trade_status.confirmed": "Dikonfirmasi",
    "trade_status.in_progress": "Sedang berlangsung",
    "trade_status.paid": "Lunas",
    "trade_status.shipped": "Dikirim",
    "trade_status.active": "Aktif",
    "trade_status.completed": "Selesai",
    "trade_status.closed": "Ditutup",
    "trade_status.cancelled": "Dibatalkan",
    "trade_status.unknown": "Tidak diketahui",
    "shipment_status.pending": "Menunggu pengiriman",
    "shipment_status.dispatched": "Telah dikirim",
    "shipment_status.in_transit": "Dalam transit",
    "shipment_status.delivered": "Diterima",
    "shipment_status.exception": "Pengecualian",
    "shipment_status.unknown": "Tidak diketahui",
    "shipment_event_type.dispatched": "Telah dikirim",
    "shipment_event_type.picked_up": "Diambil kurir",
    "shipment_event_type.in_transit": "Dalam transit",
    "shipment_event_type.arrived_at_port": "Tiba di pelabuhan",
    "shipment_event_type.customs_cleared": "Bea cukai selesai",
    "shipment_event_type.out_for_delivery": "Dalam pengiriman",
    "shipment_event_type.delivered": "Diterima",
    "shipment_event_type.unknown": "Tidak diketahui",
    "carrier.standard": "Standar",
    "carrier.sea": "Pengiriman laut",
    "carrier.air": "Pengiriman udara",
    "carrier.sea_freight": "Pengiriman laut",
    "carrier.air_freight": "Pengiriman udara",
    "carrier.unknown": "Tidak diketahui",
    "tracking_provider.standard": "Standar",
    "tracking_provider.unknown": "Tidak diketahui",
    "invoice_type.invoice": "Faktur",
    "invoice_type.proforma": "Proforma",
    "invoice_type.commercial": "Faktur komersial",
    "invoice_type.credit_note": "Nota kredit",
    "invoice_type.unknown": "Tidak diketahui",
    "content_type.post": "Artikel",
    "content_type.case": "Studi kasus",
    "content_type.unknown": "Tidak diketahui",
    "oem_status.inquiry": "Permintaan",
    "oem_status.sampling": "Sampel",
    "oem_status.formulation": "Formulasi",
    "oem_status.quotation": "Penawaran",
    "oem_status.contract": "Kontrak",
    "oem_status.production": "Produksi",
    "oem_status.delivery": "Pengiriman",
    "oem_status.completed": "Selesai",
    "oem_status.cancelled": "Dibatalkan",
    "oem_status.unknown": "Tidak diketahui",
    "sample_status.requested": "Diminta",
    "sample_status.shipped": "Dikirim",
    "sample_status.received": "Diterima",
    "sample_status.approved": "Disetujui",
    "sample_status.rejected": "Ditolak",
    "sample_status.unknown": "Tidak diketahui",
    "document_status.draft": "Draf",
    "document_status.pending": "Menunggu",
    "document_status.confirmed": "Dikonfirmasi",
    "document_status.cancelled": "Dibatalkan",
    "document_status.completed": "Selesai",
    "document_status.unknown": "Tidak diketahui",
    "document_type.quotation": "Penawaran",
    "document_type.proforma_invoice": "Faktur proforma",
    "document_type.sales_contract": "Kontrak penjualan",
    "document_type.commercial_invoice": "Faktur komersial",
    "document_type.packing_list": "Daftar kemasan",
    "document_type.bill_of_lading": "Bill of lading",
    "document_type.health_certificate": "Sertifikat kesehatan",
    "document_type.origin_certificate": "Sertifikat asal",
    "document_type.unknown": "Tidak diketahui",
    "negotiation_status.pending": "Menunggu",
    "negotiation_status.accepted": "Diterima",
    "negotiation_status.rejected": "Ditolak",
    "negotiation_status.expired": "Kedaluwarsa",
    "negotiation_status.unknown": "Tidak diketahui",
    "invoice_status.draft": "Draf",
    "invoice_status.sent": "Terkirim",
    "invoice_status.paid": "Lunas",
    "invoice_status.overdue": "Jatuh tempo",
    "invoice_status.voided": "Dibatalkan",
    "invoice_status.cancelled": "Dibatalkan",
    "invoice_status.unknown": "Tidak diketahui",
    "user_status.active": "Aktif",
    "user_status.inactive": "Nonaktif",
    "user_status.suspended": "Ditangguhkan",
    "user_status.pending": "Menunggu",
    "user_status.deleted": "Dihapus",
    "user_status.unknown": "Tidak diketahui",
    "tool_status.success": "Berhasil",
    "tool_status.error": "Gagal",
    "tool_status.running": "Berjalan",
    "tool_status.unknown": "Tidak diketahui",
    "compliance_status.passed": "Lulus",
    "compliance_status.failed": "Gagal",
    "compliance_status.warnings": "Peringatan",
    "compliance_status.pending": "Menunggu",
    "compliance_status.unknown": "Tidak diketahui",
}
_enum_ms = {
    "order_status.pending": "Menunggu",
    "order_status.confirmed": "Disahkan",
    "order_status.production": "Pengeluaran",
    "order_status.shipped": "Dihantar",
    "order_status.delivered": "Diterima",
    "order_status.cancelled": "Dibatalkan",
    "order_status.processing": "Diproses",
    "order_status.pending_confirmation": "Menunggu pengesahan",
    "order_status.pending_approval": "Menunggu kelulusan",
    "order_status.partially_shipped": "Sebahagian dihantar",
    "order_status.partially_delivered": "Sebahagian diterima",
    "order_status.partially_returned": "Sebahagian dipulangkan",
    "order_status.returned": "Dipulangkan",
    "order_status.expired": "Tamat tempoh",
    "order_status.unknown": "Tidak diketahui",
    "payment_status.unpaid": "Belum dibayar",
    "payment_status.paid": "Dibayar",
    "payment_status.partial": "Separa",
    "payment_status.refunded": "Dibayar balik",
    "payment_status.pending": "Menunggu",
    "payment_status.confirmed": "Disahkan",
    "payment_status.unknown": "Tidak diketahui",
    "product_status.active": "Aktif",
    "product_status.inactive": "Tidak aktif",
    "product_status.draft": "Draf",
    "product_status.unknown": "Tidak diketahui",
    "company_status.pending": "Menunggu",
    "company_status.verified": "Disahkan",
    "company_status.rejected": "Ditolak",
    "company_status.active": "Aktif",
    "company_status.suspended": "Digantung",
    "company_status.inactive": "Tidak aktif",
    "company_status.unknown": "Tidak diketahui",
    "kyb_status.pending": "Menunggu",
    "kyb_status.approved": "Diluluskan",
    "kyb_status.rejected": "Ditolak",
    "kyb_status.unknown": "Tidak diketahui",
    "inquiry_status.pending": "Menunggu",
    "inquiry_status.contacted": "Dihubungi",
    "inquiry_status.quoted": "Sebut harga",
    "inquiry_status.negotiating": "Rundingan",
    "inquiry_status.won": "Menang",
    "inquiry_status.lost": "Kalah",
    "inquiry_status.open": "Terbuka",
    "inquiry_status.closed": "Ditutup",
    "inquiry_status.converted": "Ditukar",
    "inquiry_status.unknown": "Tidak diketahui",
    "inquiry_priority.normal": "Biasa",
    "inquiry_priority.high": "Tinggi",
    "inquiry_priority.low": "Rendah",
    "inquiry_priority.urgent": "Segera",
    "inquiry_priority.unknown": "Tidak diketahui",
    "trade_status.draft": "Draf",
    "trade_status.pending": "Menunggu",
    "trade_status.confirmed": "Disahkan",
    "trade_status.in_progress": "Sedang dijalankan",
    "trade_status.paid": "Dibayar",
    "trade_status.shipped": "Dihantar",
    "trade_status.active": "Aktif",
    "trade_status.completed": "Selesai",
    "trade_status.closed": "Ditutup",
    "trade_status.cancelled": "Dibatalkan",
    "trade_status.unknown": "Tidak diketahui",
    "shipment_status.pending": "Menunggu penghantaran",
    "shipment_status.dispatched": "Dihantar",
    "shipment_status.in_transit": "Dalam transit",
    "shipment_status.delivered": "Diterima",
    "shipment_status.exception": "Pengecualian",
    "shipment_status.unknown": "Tidak diketahui",
    "shipment_event_type.dispatched": "Dihantar",
    "shipment_event_type.picked_up": "Diambil",
    "shipment_event_type.in_transit": "Dalam transit",
    "shipment_event_type.arrived_at_port": "Tiba di pelabuhan",
    "shipment_event_type.customs_cleared": "Kastam selesai",
    "shipment_event_type.out_for_delivery": "Dalam penghantaran",
    "shipment_event_type.delivered": "Diterima",
    "shipment_event_type.unknown": "Tidak diketahui",
    "carrier.standard": "Standard",
    "carrier.sea": "Penghantaran laut",
    "carrier.air": "Penghantaran udara",
    "carrier.sea_freight": "Penghantaran laut",
    "carrier.air_freight": "Penghantaran udara",
    "carrier.unknown": "Tidak diketahui",
    "tracking_provider.standard": "Standard",
    "tracking_provider.unknown": "Tidak diketahui",
    "invoice_type.invoice": "Invois",
    "invoice_type.proforma": "Proforma",
    "invoice_type.commercial": "Invois komersial",
    "invoice_type.credit_note": "Nota kredit",
    "invoice_type.unknown": "Tidak diketahui",
    "content_type.post": "Artikel",
    "content_type.case": "Kajian kes",
    "content_type.unknown": "Tidak diketahui",
    "oem_status.inquiry": "Pertanyaan",
    "oem_status.sampling": "Sampel",
    "oem_status.formulation": "Formulasi",
    "oem_status.quotation": "Sebut harga",
    "oem_status.contract": "Kontrak",
    "oem_status.production": "Pengeluaran",
    "oem_status.delivery": "Penghantaran",
    "oem_status.completed": "Selesai",
    "oem_status.cancelled": "Dibatalkan",
    "oem_status.unknown": "Tidak diketahui",
    "sample_status.requested": "Diminta",
    "sample_status.shipped": "Dihantar",
    "sample_status.received": "Diterima",
    "sample_status.approved": "Diluluskan",
    "sample_status.rejected": "Ditolak",
    "sample_status.unknown": "Tidak diketahui",
    "document_status.draft": "Draf",
    "document_status.pending": "Menunggu",
    "document_status.confirmed": "Disahkan",
    "document_status.cancelled": "Dibatalkan",
    "document_status.completed": "Selesai",
    "document_status.unknown": "Tidak diketahui",
    "document_type.quotation": "Sebut harga",
    "document_type.proforma_invoice": "Invois proforma",
    "document_type.sales_contract": "Kontrak jualan",
    "document_type.commercial_invoice": "Invois komersial",
    "document_type.packing_list": "Senarai pembungkusan",
    "document_type.bill_of_lading": "Bill of lading",
    "document_type.health_certificate": "Sijil kesihatan",
    "document_type.origin_certificate": "Sijil asal-usul",
    "document_type.unknown": "Tidak diketahui",
    "negotiation_status.pending": "Menunggu",
    "negotiation_status.accepted": "Diterima",
    "negotiation_status.rejected": "Ditolak",
    "negotiation_status.expired": "Tamat tempoh",
    "negotiation_status.unknown": "Tidak diketahui",
    "invoice_status.draft": "Draf",
    "invoice_status.sent": "Dihantar",
    "invoice_status.paid": "Dibayar",
    "invoice_status.overdue": "Lewat tempoh",
    "invoice_status.voided": "Dibatalkan",
    "invoice_status.cancelled": "Dibatalkan",
    "invoice_status.unknown": "Tidak diketahui",
    "user_status.active": "Aktif",
    "user_status.inactive": "Tidak aktif",
    "user_status.suspended": "Digantung",
    "user_status.pending": "Menunggu",
    "user_status.deleted": "Dipadam",
    "user_status.unknown": "Tidak diketahui",
    "tool_status.success": "Berjaya",
    "tool_status.error": "Ralat",
    "tool_status.running": "Sedang berjalan",
    "tool_status.unknown": "Tidak diketahui",
    "compliance_status.passed": "Lulus",
    "compliance_status.failed": "Gagal",
    "compliance_status.warnings": "Amaran",
    "compliance_status.pending": "Menunggu",
    "compliance_status.unknown": "Tidak diketahui",
}
for k, v in _enum_id.items():
    _add(f"enum.{k}", v, _enum_ms[k])

# Brand carriers kept as-is
for c in ["carrier.dhl", "carrier.fedex", "carrier.ups", "carrier.usps", "carrier.ems"]:
    brand = c.split(".")[1].upper() if c != "carrier.fedex" else "FedEx"
    if c == "carrier.dhl":
        brand = "DHL"
    elif c == "carrier.ups":
        brand = "UPS"
    elif c == "carrier.usps":
        brand = "USPS"
    elif c == "carrier.ems":
        brand = "EMS"
    _add(f"enum.{c}", brand, brand)


def protect(text: str) -> tuple[str, list[str]]:
    tokens: list[str] = []

    def repl(m: re.Match[str]) -> str:
        tokens.append(m.group(0))
        return f"__TOK{len(tokens)-1}__"

    return PLACEHOLDER_RE.sub(repl, text), tokens


def restore(text: str, tokens: list[str]) -> str:
    for i, tok in enumerate(tokens):
        text = text.replace(f"__TOK{i}__", tok)
    return text


def leaf_paths(obj: dict, prefix: str = "") -> list[tuple[str, str]]:
    out: list[tuple[str, str]] = []
    for k, v in obj.items():
        p = f"{prefix}.{k}" if prefix else k
        if isinstance(v, dict):
            out.extend(leaf_paths(v, p))
        else:
            out.append((p, v))
    return out


def set_path(obj: dict, path: str, value: str) -> None:
    parts = path.split(".")
    cur = obj
    for part in parts[:-1]:
        cur = cur[part]
    cur[parts[-1]] = value


def load_cache(locale: str) -> dict[str, str]:
    path = CACHE_DIR / f"{locale}.json"
    if path.exists():
        return json.loads(path.read_text(encoding="utf-8"))
    return {}


def save_cache(locale: str, cache: dict[str, str]) -> None:
    path = CACHE_DIR / f"{locale}.json"
    path.write_text(json.dumps(cache, ensure_ascii=False, indent=2), encoding="utf-8")


def translate_text(text: str, locale: str, cache: dict[str, str]) -> str:
    if text in cache:
        return cache[text]
    protected, tokens = protect(text)
    translator = MyMemoryTranslator(source="en-GB", target=TARGET_MAP[locale])
    tr = text
    for attempt in range(5):
        try:
            tr = translator.translate(protected) or text
            break
        except Exception:
            time.sleep(0.4 * (attempt + 1))
    result = restore(tr, tokens)
    cache[text] = result
    return result


def fill_cache(strings: list[str], locale: str) -> dict[str, str]:
    cache = load_cache(locale)
    todo = [s for s in strings if s not in cache]
    total = len(todo)
    for i, src in enumerate(todo, 1):
        translate_text(src, locale, cache)
        if i % 25 == 0 or i == total:
            save_cache(locale, cache)
            print(f"  [{locale}] {i}/{total}")
        time.sleep(0.08)
    save_cache(locale, cache)
    return cache


def build_locale(en_data: dict, locale: str, manual: dict[str, str], cache: dict[str, str]) -> dict:
    result = deepcopy(en_data)
    for path, en_val in leaf_paths(en_data):
        if path in manual:
            set_path(result, path, manual[path])
        elif isinstance(en_val, str):
            set_path(result, path, cache.get(en_val, en_val))
    return result


def main() -> None:
    import sys

    only = sys.argv[1:] if len(sys.argv) > 1 else ["admin", "common"]
    en_admin = json.loads((I18N / "en" / "admin.json").read_text(encoding="utf-8"))
    en_common = json.loads((I18N / "en" / "common.json").read_text(encoding="utf-8"))

    datasets = []
    if "admin" in only:
        datasets.append(("admin.json", en_admin))
    if "common" in only:
        datasets.append(("common.json", en_common))

    all_strings: set[str] = set()
    for _, data in datasets:
        for _, v in leaf_paths(data):
            if isinstance(v, str):
                all_strings.add(v)
    manual_en = set(MANUAL_ID.values()) | set(MANUAL_MS.values())
    # strings covered by path manual use en source; still need en->id/ms for others
    to_translate = sorted(all_strings)
    print(f"Unique strings: {len(to_translate)}")

    for locale in ("id", "ms"):
        print(f"Filling cache for {locale}...")
        fill_cache(to_translate, locale)

    if "admin" in only:
        id_admin = build_locale(en_admin, "id", MANUAL_ID, load_cache("id"))
        ms_admin = build_locale(en_admin, "ms", MANUAL_MS, load_cache("ms"))
        (I18N / "id" / "admin.json").write_text(
            json.dumps(id_admin, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
        )
        (I18N / "ms" / "admin.json").write_text(
            json.dumps(ms_admin, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
        )
        print("Wrote admin.json for id/ms")

    if "common" in only:
        id_common = build_locale(en_common, "id", MANUAL_ID, load_cache("id"))
        ms_common = build_locale(en_common, "ms", MANUAL_MS, load_cache("ms"))
        (I18N / "id" / "common.json").write_text(
            json.dumps(id_common, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
        )
        (I18N / "ms" / "common.json").write_text(
            json.dumps(ms_common, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
        )
        print("Wrote common.json for id/ms")


if __name__ == "__main__":
    main()
