#!/usr/bin/env python3
"""Build id/ms admin.json from en with offline translations (no network)."""

from __future__ import annotations

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
SCRIPTS = Path(__file__).resolve().parent

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Temperature",
    "Plan-Execute-Replan", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA", "ISO",
    "CSV", "YAML", "Stripe", "PayPal", "Chatbot", "Agent", "Cursor", "Markdown",
    "Super", "Admin", "Superadmin", "Diff", "Trade", "translate_content",
    "Grade", "A", "EN", "ID", "MS", "PI", "CI", "SC", "PL", "B/L", "RFM",
    "Ctrl+K", "⌘K", "1200×630", "aʷ", "→", "…", "CRUD", "3M", "6M", "12M",
    "B2B Coordinator", "Trade Assistant", "Product Recommend", "AI Search",
    "Analyze Inquiry", "Generate Quotation", "Health Certificate", "Bill of Lading",
    "Commercial Invoice", "Sales Contract", "Packing List", "Certificate of Origin",
    "Credit Note", "Proforma Invoice", "Tax Invoice", "INVOICE", "PROFORMA INVOICE",
    "TAX INVOICE", "CREDIT NOTE", "FAKTUR", "INVOIS", "NOTA KREDIT",
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


def unflatten(flat: dict[str, str]) -> dict:
    root: dict = {}
    for path, value in flat.items():
        parts = path.split(".")
        cur = root
        for part in parts[:-1]:
            cur = cur.setdefault(part, {})
        cur[parts[-1]] = value
    return root


def protect_glossary(text: str) -> tuple[str, list[tuple[str, str]]]:
    reps: list[tuple[str, str]] = []

    def sub(m: re.Match[str]) -> str:
        token = m.group(0)
        ph = f"__G{len(reps)}__"
        reps.append((ph, token))
        return ph

    protected = re.sub(r"\{[a-zA-Z0-9_]+\}", sub, text)
    for term in sorted(GLOSSARY, key=len, reverse=True):
        if term in protected:
            ph = f"__G{len(reps)}__"
            reps.append((ph, term))
            protected = protected.replace(term, ph)
    return protected, reps


def restore(text: str, reps: list[tuple[str, str]]) -> str:
    for ph, orig in reps:
        text = text.replace(ph, orig)
    return text


def apply_rules(text: str, rules: list[tuple[str, str]]) -> str:
    protected, reps = protect_glossary(text)
    for src, dst in sorted(rules, key=lambda x: -len(x[0])):
        protected = protected.replace(src, dst)
    return restore(protected, reps)


ID_RULES: list[tuple[str, str]] = [
    ("CandyPro Admin", "CandyPro Admin"),
    ("Super Admin", "Super Admin"),
    ("Manage ", "Kelola "),
    ("Track and manage ", "Lacak dan kelola "),
    ("Track all ", "Lacak semua "),
    ("View and manage ", "Lihat dan kelola "),
    ("Review and process ", "Tinjau dan proses "),
    ("Configure ", "Konfigurasi "),
    ("Create ", "Buat "),
    ("Add ", "Tambah "),
    ("Edit ", "Edit "),
    ("Delete ", "Hapus "),
    ("Update ", "Perbarui "),
    ("Loading ", "Memuat "),
    ("Saving", "Menyimpan"),
    ("Creating", "Membuat"),
    ("Updating", "Memperbarui"),
    ("Deleting", "Menghapus"),
    ("Submitting", "Mengirim"),
    ("Assigning", "Menetapkan"),
    ("Translating", "Menerjemahkan"),
    ("Generating", "Membuat"),
    ("Exporting", "Mengekspor"),
    ("Importing", "Mengimpor"),
    ("Applying", "Menerapkan"),
    ("Confirm ", "Konfirmasi "),
    ("Cancel", "Batal"),
    ("Save", "Simpan"),
    ("Search", "Cari"),
    ("Filter", "Filter"),
    ("Previous", "Sebelumnya"),
    ("Next", "Berikutnya"),
    ("Showing {from} to {to} of {total}", "Menampilkan {from} hingga {to} dari {total}"),
    ("Showing {from} to {to} of {total} results", "Menampilkan {from} hingga {to} dari {total} hasil"),
    ("{total} total", "{total} total"),
    ("No data available", "Tidak ada data"),
    ("No recent activity", "Tidak ada aktivitas terbaru"),
    ("Error loading ", "Gagal memuat "),
    ("Failed to load ", "Gagal memuat "),
    ("Failed to ", "Gagal "),
    (" successfully.", " berhasil."),
    (" successfully", " berhasil"),
    ("Are you sure you want to ", "Apakah Anda yakin ingin "),
    ("This action cannot be undone.", "Tindakan ini tidak dapat dibatalkan."),
    (" permanently?", " secara permanen?"),
    (" is required.", " wajib diisi."),
    (" are required.", " wajib diisi."),
    ("optional", "opsional"),
    ("All Statuses", "Semua Status"),
    ("All Status", "Semua Status"),
    ("All Categories", "Semua Kategori"),
    ("All Stock", "Semua Stok"),
    ("All Types", "Semua Tipe"),
    ("All Groups", "Semua Grup"),
    ("All Locales", "Semua Locale"),
    ("All Entity Types", "Semua Tipe Entitas"),
    ("All Actions", "Semua Aksi"),
    ("Dashboard", "Dasbor"),
    ("Products", "Produk"),
    ("Product", "Produk"),
    ("Inventory", "Inventaris"),
    ("Categories", "Kategori"),
    ("Category", "Kategori"),
    ("Coupons", "Kupon"),
    ("Gift Cards", "Kartu Hadiah"),
    ("Coupons & Gift Cards", "Kupon & Kartu Hadiah"),
    ("Shipping Rates", "Tarif Pengiriman"),
    ("Tax Rates", "Tarif Pajak"),
    ("Shipping & Tax", "Pengiriman & Pajak"),
    ("Draft", "Draf"),
    ("Sent", "Terkirim"),
    ("Paid", "Lunas"),
    ("Overdue", "Jatuh tempo"),
    ("Cancelled", "Dibatalkan"),
    ("Active", "Aktif"),
    ("Inactive", "Nonaktif"),
    ("Pending", "Menunggu"),
    ("Confirmed", "Dikonfirmasi"),
    ("Delivered", "Diterima"),
    ("Description", "Deskripsi"),
    ("Status", "Status"),
    ("Actions", "Aksi"),
    ("Notes", "Catatan"),
    ("Amount", "Jumlah"),
    ("Date", "Tanggal"),
    ("Name", "Nama"),
    ("Email", "Email"),
    ("Phone", "Telepon"),
    ("Company", "Perusahaan"),
    ("Customer", "Pelanggan"),
    ("Order", "Pesanan"),
    ("Orders", "Pesanan"),
    ("Users", "Pengguna"),
    ("User", "Pengguna"),
    ("Settings", "Pengaturan"),
    ("Analytics", "Analitik"),
    ("Financial", "Keuangan"),
    ("Management", "Manajemen"),
    ("Translations", "Terjemahan"),
    ("Organizations", "Organisasi"),
    ("Channels", "Saluran"),
    ("Staff", "Staf"),
    ("Audit Log", "Log Audit"),
    ("Certifications", "Sertifikasi"),
    ("Shipments", "Pengiriman"),
    ("Invoices", "Faktur"),
    ("Returns", "Retur"),
    ("Pricing", "Harga"),
    ("Trades", "Perdagangan"),
    ("Companies", "Perusahaan"),
    ("Inquiries", "Permintaan"),
    ("Content", "Konten"),
    ("just now", "baru saja"),
    ("{count} minutes ago", "{count} menit lalu"),
    ("{count} hours ago", "{count} jam lalu"),
    ("{count} days ago", "{count} hari lalu"),
    ("Never", "Tidak pernah"),
    ("Yes", "Ya"),
    ("No ", "Tidak "),
    ("Back to ", "Kembali ke "),
    ("View ", "Lihat "),
    ("Select ", "Pilih "),
    ("Type ", "Ketik "),
    ("From ", "Dari "),
    ("To ", "Ke "),
    ("From Date", "Tanggal Mulai"),
    ("To Date", "Tanggal Akhir"),
    ("Due Date", "Tanggal Jatuh Tempo"),
    ("Invoice Date", "Tanggal Faktur"),
    ("Created", "Dibuat"),
    ("Updated", "Diperbarui"),
    ("Deleted", "Dihapus"),
    ("Errors", "Kesalahan"),
    ("Warehouse", "Gudang"),
    ("Warehouses", "Gudang"),
    ("Transfers", "Transfer"),
    ("Quantity", "Jumlah"),
    ("Qty", "Jml"),
    ("Stock", "Stok"),
    ("Low Stock", "Stok Rendah"),
    ("Out of Stock", "Stok Habis"),
    ("In Stock", "Tersedia"),
    ("Total Value", "Total Nilai"),
    ("Unit Value", "Nilai Satuan"),
    ("Adjust Stock", "Sesuaikan Stok"),
    ("Batch Edit", "Edit Massal"),
    ("Batch Delete", "Hapus Massal"),
    ("Import XLSX", "Impor XLSX"),
    ("Export Selected", "Ekspor Terpilih"),
    ("Download Template", "Unduh Template"),
    ("Clear filters", "Hapus filter"),
    ("Clear selection", "Hapus pilihan"),
    ("Select All", "Pilih Semua"),
    ("Percentage", "Persentase"),
    ("Fixed Amount", "Jumlah Tetap"),
    ("Min Order", "Pesanan Min."),
    ("Starts At", "Mulai"),
    ("Expires At", "Berakhir"),
    ("Balance", "Saldo"),
    ("Currency", "Mata uang"),
    ("Destination", "Tujuan"),
    ("Carrier", "Kurir"),
    ("Base Cost", "Biaya Dasar"),
    ("Cost per kg", "Biaya per kg"),
    ("Est. Days", "Perkiraan Hari"),
    ("Country (ISO)", "Negara (ISO)"),
    ("Region / State", "Wilayah / Provinsi"),
    ("Rate", "Tarif"),
    ("Bahasa Indonesia", "Bahasa Indonesia"),
]

MS_RULES: list[tuple[str, str]] = [
    ("CandyPro Admin", "CandyPro Admin"),
    ("Super Admin", "Super Admin"),
    ("Manage ", "Urus "),
    ("Track and manage ", "Jejak dan urus "),
    ("Track all ", "Jejak semua "),
    ("View and manage ", "Lihat dan urus "),
    ("Review and process ", "Semak dan proses "),
    ("Configure ", "Konfigurasi "),
    ("Create ", "Cipta "),
    ("Add ", "Tambah "),
    ("Edit ", "Edit "),
    ("Delete ", "Padam "),
    ("Update ", "Kemas kini "),
    ("Loading ", "Memuatkan "),
    ("Saving", "Menyimpan"),
    ("Creating", "Mencipta"),
    ("Updating", "Mengemas kini"),
    ("Deleting", "Memadam"),
    ("Submitting", "Menghantar"),
    ("Assigning", "Menetapkan"),
    ("Translating", "Menterjemah"),
    ("Generating", "Menjana"),
    ("Exporting", "Mengeksport"),
    ("Importing", "Mengimport"),
    ("Applying", "Menggunakan"),
    ("Confirm ", "Sahkan "),
    ("Cancel", "Batal"),
    ("Save", "Simpan"),
    ("Search", "Cari"),
    ("Filter", "Tapis"),
    ("Previous", "Sebelum"),
    ("Next", "Seterusnya"),
    ("Showing {from} to {to} of {total}", "Memaparkan {from} hingga {to} daripada {total}"),
    ("Showing {from} to {to} of {total} results", "Memaparkan {from} hingga {to} daripada {total} keputusan"),
    ("{total} total", "{total} jumlah"),
    ("No data available", "Tiada data"),
    ("No recent activity", "Tiada aktiviti terkini"),
    ("Error loading ", "Ralat memuatkan "),
    ("Failed to load ", "Gagal memuatkan "),
    ("Failed to ", "Gagal "),
    (" successfully.", " berjaya."),
    (" successfully", " berjaya"),
    ("Are you sure you want to ", "Adakah anda pasti mahu "),
    ("This action cannot be undone.", "Tindakan ini tidak boleh dibuat asal."),
    (" permanently?", " secara kekal?"),
    (" is required.", " diperlukan."),
    (" are required.", " diperlukan."),
    ("optional", "pilihan"),
    ("All Statuses", "Semua Status"),
    ("All Status", "Semua Status"),
    ("All Categories", "Semua Kategori"),
    ("All Stock", "Semua Stok"),
    ("All Types", "Semua Jenis"),
    ("All Groups", "Semua Kumpulan"),
    ("All Locales", "Semua Locale"),
    ("All Entity Types", "Semua Jenis Entiti"),
    ("All Actions", "Semua Tindakan"),
    ("Dashboard", "Papan Pemuka"),
    ("Products", "Produk"),
    ("Product", "Produk"),
    ("Inventory", "Inventori"),
    ("Categories", "Kategori"),
    ("Category", "Kategori"),
    ("Coupons", "Kupon"),
    ("Gift Cards", "Kad Hadiah"),
    ("Coupons & Gift Cards", "Kupon & Kad Hadiah"),
    ("Shipping Rates", "Kadar Penghantaran"),
    ("Tax Rates", "Kadar Cukai"),
    ("Shipping & Tax", "Penghantaran & Cukai"),
    ("Draft", "Draf"),
    ("Sent", "Dihantar"),
    ("Paid", "Dibayar"),
    ("Overdue", "Lewat tempoh"),
    ("Cancelled", "Dibatalkan"),
    ("Active", "Aktif"),
    ("Inactive", "Tidak aktif"),
    ("Pending", "Menunggu"),
    ("Confirmed", "Disahkan"),
    ("Delivered", "Diterima"),
    ("Description", "Penerangan"),
    ("Status", "Status"),
    ("Actions", "Tindakan"),
    ("Notes", "Nota"),
    ("Amount", "Jumlah"),
    ("Date", "Tarikh"),
    ("Name", "Nama"),
    ("Email", "E-mel"),
    ("Phone", "Telefon"),
    ("Company", "Syarikat"),
    ("Customer", "Pelanggan"),
    ("Order", "Pesanan"),
    ("Orders", "Pesanan"),
    ("Users", "Pengguna"),
    ("User", "Pengguna"),
    ("Settings", "Tetapan"),
    ("Analytics", "Analitik"),
    ("Financial", "Kewangan"),
    ("Management", "Pengurusan"),
    ("Translations", "Terjemahan"),
    ("Organizations", "Organisasi"),
    ("Channels", "Saluran"),
    ("Staff", "Kakitangan"),
    ("Audit Log", "Log Audit"),
    ("Certifications", "Pensijilan"),
    ("Shipments", "Penghantaran"),
    ("Invoices", "Invois"),
    ("Returns", "Pemulangan"),
    ("Pricing", "Harga"),
    ("Trades", "Perdagangan"),
    ("Companies", "Syarikat"),
    ("Inquiries", "Pertanyaan"),
    ("Content", "Kandungan"),
    ("just now", "baru sahaja"),
    ("{count} minutes ago", "{count} minit lalu"),
    ("{count} hours ago", "{count} jam lalu"),
    ("{count} days ago", "{count} hari lalu"),
    ("Never", "Tidak pernah"),
    ("Yes", "Ya"),
    ("No ", "Tiada "),
    ("Back to ", "Kembali ke "),
    ("View ", "Lihat "),
    ("Select ", "Pilih "),
    ("Type ", "Taip "),
    ("From ", "Dari "),
    ("To ", "Ke "),
    ("From Date", "Tarikh Mula"),
    ("To Date", "Tarikh Akhir"),
    ("Due Date", "Tarikh Tamat Tempo"),
    ("Invoice Date", "Tarikh Invois"),
    ("Created", "Dicipta"),
    ("Updated", "Dikemas kini"),
    ("Deleted", "Dipadam"),
    ("Errors", "Ralat"),
    ("Warehouse", "Gudang"),
    ("Warehouses", "Gudang"),
    ("Transfers", "Pemindahan"),
    ("Quantity", "Kuantiti"),
    ("Qty", "Ktl"),
    ("Stock", "Stok"),
    ("Low Stock", "Stok Rendah"),
    ("Out of Stock", "Stok Habis"),
    ("In Stock", "Ada Stok"),
    ("Total Value", "Jumlah Nilai"),
    ("Unit Value", "Nilai Unit"),
    ("Adjust Stock", "Laraskan Stok"),
    ("Batch Edit", "Edit Pukal"),
    ("Batch Delete", "Padam Pukal"),
    ("Import XLSX", "Import XLSX"),
    ("Export Selected", "Eksport Terpilih"),
    ("Download Template", "Muat Turun Templat"),
    ("Clear filters", "Kosongkan penapis"),
    ("Clear selection", "Kosongkan pilihan"),
    ("Select All", "Pilih Semua"),
    ("Percentage", "Peratusan"),
    ("Fixed Amount", "Jumlah Tetap"),
    ("Min Order", "Pesanan Min."),
    ("Starts At", "Bermula"),
    ("Expires At", "Tamat"),
    ("Balance", "Baki"),
    ("Currency", "Mata wang"),
    ("Destination", "Destinasi"),
    ("Carrier", "Kurier"),
    ("Base Cost", "Kos Asas"),
    ("Cost per kg", "Kos per kg"),
    ("Est. Days", "Angg. Hari"),
    ("Country (ISO)", "Negara (ISO)"),
    ("Region / State", "Wilayah / Negeri"),
    ("Rate", "Kadar"),
    ("Bahasa Melayu", "Bahasa Melayu"),
]

# Key-path overrides for priority sections and known tricky strings
ID_KEY_OVERRIDES: dict[str, str] = {}
MS_KEY_OVERRIDES: dict[str, str] = {}

# Load hand-crafted overrides if present
for loc, var in (("id", "ID_KEY_OVERRIDES"), ("ms", "MS_KEY_OVERRIDES")):
    path = SCRIPTS / f"admin_overrides_{loc}.json"
    if path.exists():
        data = json.loads(path.read_text(encoding="utf-8"))
        if loc == "id":
            ID_KEY_OVERRIDES.update(data)
        else:
            MS_KEY_OVERRIDES.update(data)

# Load full string-level overrides
ID_STRING_OVERRIDES: dict[str, str] = {}
MS_STRING_OVERRIDES: dict[str, str] = {}
for loc, target in (("id", ID_STRING_OVERRIDES), ("ms", MS_STRING_OVERRIDES)):
    path = SCRIPTS / f"admin_strings_{loc}.json"
    if path.exists():
        target.update(json.loads(path.read_text(encoding="utf-8")))


def translate(text: str, locale: str) -> str:
    if locale == "id":
        if text in ID_STRING_OVERRIDES:
            return ID_STRING_OVERRIDES[text]
        return apply_rules(text, ID_RULES)
    if text in MS_STRING_OVERRIDES:
        return MS_STRING_OVERRIDES[text]
    return apply_rules(text, MS_RULES)


def build_admin(locale: str) -> dict:
    en_tree = json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8"))["admin"]
    en_flat = flatten(en_tree)
    key_overrides = ID_KEY_OVERRIDES if locale == "id" else MS_KEY_OVERRIDES
    out: dict[str, str] = {}
    for key, en_val in en_flat.items():
        if key in key_overrides:
            out[key] = key_overrides[key]
        else:
            out[key] = translate(en_val, locale)
    return {"admin": unflatten(out)}


def main() -> None:
    for locale in ("id", "ms"):
        data = build_admin(locale)
        path = ROOT / locale / "admin.json"
        path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print(f"Wrote {path}")


if __name__ == "__main__":
    main()
