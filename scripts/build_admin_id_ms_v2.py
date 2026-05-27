#!/usr/bin/env python3
"""Build fully translated id/ms admin.json offline (no external APIs)."""

from __future__ import annotations

import json
import re
from copy import deepcopy
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
I18N = ROOT / "frontend" / "i18n"

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "ODM", "PI",
    "CI", "SC", "B/L", "PL", "ETA", "ISO", "CSV", "YAML", "Stripe", "PayPal",
    "Chatbot", "Agent", "Markdown", "Super", "Admin", "Superadmin", "Diff",
    "Trade", "CRUD", "Ctrl+K", "⌘K", "1200×630", "aʷ", "EN", "ID", "MS",
    "12M", "3M", "6M", "3PL", "CREDIT NOTE", "ISO", "mm", "kg", "g", "kJ", "kcal",
}

PHRASES_PATH = Path(__file__).resolve().parent / "i18n_admin_phrases_id_ms.json"
PHRASES: dict[str, dict[str, str]] = {}
if PHRASES_PATH.exists():
    PHRASES = json.loads(PHRASES_PATH.read_text(encoding="utf-8"))

# Path-level overrides (highest priority)
PATH_ID: dict[str, str] = {
    "admin.langLabel": "ID",
    "admin.langLabelShort": "ID",
    "admin.langTitle": "Bahasa Indonesia",
    "admin.brand": "CandyPro Admin",
    "admin.logout": "Keluar",
    "admin.nav.organizations": "Organisasi Pembeli",
    "admin.nav.inquiries": "Permintaan Penawaran",
}
PATH_MS: dict[str, str] = {
    "admin.langLabel": "MS",
    "admin.langLabelShort": "MS",
    "admin.langTitle": "Bahasa Melayu",
    "admin.brand": "CandyPro Admin",
    "admin.logout": "Log keluar",
    "admin.nav.organizations": "Organisasi Pembeli",
}

# Longest-first phrase pairs: (english, indonesian, malay)
PHRASE_TRIPLES: list[tuple[str, str, str]] = [
    ("Manage product catalog with complete CRUD operations.", "Kelola katalog produk dengan operasi CRUD lengkap.", "Urus katalog produk dengan operasi CRUD lengkap."),
    ("Track and manage product stock levels, values, and adjustments.", "Lacak dan kelola tingkat stok, nilai, dan penyesuaian produk.", "Jejak dan urus tahap stok, nilai, dan pelarasan produk."),
    ("Manage product categories with multilingual names, aliases, and descriptions.", "Kelola kategori produk dengan nama, alias, dan deskripsi multibahasa.", "Urus kategori produk dengan nama, alias, dan penerangan pelbagai bahasa."),
    ("Promotional codes and prepaid credits.", "Kode promosi dan kredit prabayar.", "Kod promosi dan kredit prabayar."),
    ("Destination-based freight rate tables. Estimated days are reference values for quoting — not guaranteed transit times.", "Tabel tarif pengiriman berdasarkan tujuan. Perkiraan hari hanya referensi penawaran — bukan jaminan waktu transit.", "Jadual kadar penghantaran mengikut destinasi. Anggaran hari adalah rujukan sebut harga — bukan jaminan masa transit."),
    ("Destination country/region tax rates applied at cart checkout.", "Tarif pajak negara/wilayah tujuan yang diterapkan saat checkout keranjang.", "Kadar cukai negara/wilayah destinasi yang digunakan semasa checkout troli."),
    ("Platform overview and key metrics at a glance.", "Ikhtisar platform dan metrik utama sekilas.", "Gambaran platform dan metrik utama sepintas lalu."),
    ("Loading dashboard statistics...", "Memuat statistik dasbor...", "Memuatkan statistik papan pemuka..."),
    ("Failed to load dashboard", "Gagal memuat dasbor", "Ralat memuatkan papan pemuka"),
    ("Coupons & Gift Cards", "Kupon & Kartu Hadiah", "Kupon & Kad Hadiah"),
    ("Shipping Rates", "Tarif Pengiriman", "Kadar Penghantaran"),
    ("Tax Rates", "Tarif Pajak", "Kadar Cukai"),
    ("Shipping & Tax", "Pengiriman & Pajak", "Penghantaran & Cukai"),
    ("Search by product name or SKU...", "Cari berdasarkan nama produk atau SKU...", "Cari mengikut nama produk atau SKU..."),
    ("Loading products...", "Memuat produk...", "Memuatkan produk..."),
    ("No products found.", "Tidak ada produk.", "Tiada produk."),
    ("Add Product", "Tambah Produk", "Tambah Produk"),
    ("Edit Product", "Edit Produk", "Edit Produk"),
    ("Create Product", "Buat Produk", "Cipta Produk"),
    ("Delete this product?", "Hapus produk ini?", "Padam produk ini?"),
    ("Product created successfully.", "Produk berhasil dibuat.", "Produk berjaya dicipta."),
    ("Product updated successfully.", "Produk berhasil diperbarui.", "Produk berjaya dikemas kini."),
    ("Product deleted successfully.", "Produk berhasil dihapus.", "Produk berjaya dipadam."),
    ("Product name is required.", "Nama produk wajib diisi.", "Nama produk diperlukan."),
    ("Published (default)", "Diterbitkan (default)", "Diterbitkan (lalai)"),
    ("All statuses", "Semua status", "Semua status"),
    ("Translate missing locales", "Terjemahkan locale yang belum ada", "Terjemah locale yang tiada"),
    ("Translating…", "Menerjemahkan…", "Menterjemah…"),
    ("Translating...", "Menerjemahkan...", "Menterjemah..."),
    ("Loading inventory...", "Memuat inventaris...", "Memuatkan inventori..."),
    ("No inventory data found.", "Tidak ada data inventaris.", "Tiada data inventori."),
    ("Adjust Stock", "Sesuaikan Stok", "Laraskan Stok"),
    ("Low Stock", "Stok Rendah", "Stok Rendah"),
    ("Out of Stock", "Stok Habis", "Stok Habis"),
    ("In Stock", "Tersedia", "Ada Stok"),
    ("Total Value", "Total Nilai", "Jumlah Nilai"),
    ("Filter by stock status", "Filter berdasarkan status stok", "Tapis mengikut status stok"),
    ("Filter by category", "Filter berdasarkan kategori", "Tapis mengikut kategori"),
    ("All Categories", "Semua Kategori", "Semua Kategori"),
    ("Stock History", "Riwayat Stok", "Sejarah Stok"),
    ("No adjustment history found.", "Tidak ada riwayat penyesuaian.", "Tiada sejarah pelarasan."),
    ("Add Category", "Tambah Kategori", "Tambah Kategori"),
    ("Edit Category", "Edit Kategori", "Edit Kategori"),
    ("Create Category", "Buat Kategori", "Cipta Kategori"),
    ("No categories found", "Tidak ada kategori", "Tiada kategori"),
    ("Delete this category?", "Hapus kategori ini?", "Padam kategori ini?"),
    ("Category created", "Kategori dibuat", "Kategori dicipta"),
    ("Category updated", "Kategori diperbarui", "Kategori dikemas kini"),
    ("Category deleted", "Kategori dihapus", "Kategori dipadam"),
    ("Failed to load categories", "Gagal memuat kategori", "Gagal memuatkan kategori"),
    ("Create Coupon", "Buat Kupon", "Cipta Kupon"),
    ("Create Gift Card", "Buat Kartu Hadiah", "Cipta Kad Hadiah"),
    ("No coupons.", "Tidak ada kupon.", "Tiada kupon."),
    ("No gift cards.", "Tidak ada kartu hadiah.", "Tiada kad hadiah."),
    ("Gift Cards", "Kartu Hadiah", "Kad Hadiah"),
    ("Fixed Amount", "Jumlah Tetap", "Jumlah Tetap"),
    ("Min Order", "Pesanan Min.", "Pesanan Min."),
    ("Starts At", "Mulai", "Bermula"),
    ("Expires At", "Berakhir", "Tamat"),
    ("Add Rate", "Tambah Tarif", "Tambah Kadar"),
    ("No shipping rates.", "Tidak ada tarif pengiriman.", "Tiada kadar penghantaran."),
    ("Add Tax Rate", "Tambah Tarif Pajak", "Tambah Kadar Cukai"),
    ("No tax rates configured.", "Tidak ada tarif pajak.", "Tiada kadar cukai dikonfigurasi."),
    ("Min Weight (kg)", "Berat Min. (kg)", "Berat Min. (kg)"),
    ("Max Weight (kg)", "Berat Max. (kg)", "Berat Max. (kg)"),
    ("Base Cost", "Biaya Dasar", "Kos Asas"),
    ("Cost per kg", "Biaya per kg", "Kos per kg"),
    ("Est. Days", "Perkiraan Hari", "Angg. Hari"),
    ("Country (ISO)", "Negara (ISO)", "Negara (ISO)"),
    ("Region / State", "Wilayah / Provinsi", "Wilayah / Negeri"),
    ("Decimal fraction, e.g. 0.08875 = 8.875%", "Pecahan desimal, mis. 0,08875 = 8,875%", "Pecahan perpuluhan, cth. 0.08875 = 8.875%"),
    ("Showing {from} to {to} of {total}", "Menampilkan {from} hingga {to} dari {total}", "Memaparkan {from} hingga {to} daripada {total}"),
    ("Showing {from} to {to} of {total} results", "Menampilkan {from} hingga {to} dari {total} hasil", "Memaparkan {from} hingga {to} daripada {total} keputusan"),
    ("{count} new this week", "{count} baru minggu ini", "{count} baharu minggu ini"),
    ("Revenue This Month", "Pendapatan Bulan Ini", "Hasil Bulan Ini"),
    ("Active Customers", "Pelanggan Aktif", "Pelanggan Aktif"),
    ("Avg Order Value", "Nilai Pesanan Rata-rata", "Nilai Pesanan Purata"),
    ("Conversion Rate", "Tingkat Konversi", "Kadar Penukaran"),
    ("Recent Activity", "Aktivitas Terbaru", "Aktiviti Terkini"),
    ("No recent activity", "Tidak ada aktivitas terbaru", "Tiada aktiviti terkini"),
    ("Open sidebar menu", "Buka menu sidebar", "Buka menu sidebar"),
    ("Close sidebar menu", "Tutup menu sidebar", "Tutup menu sidebar"),
    ("Expand or collapse sidebar", "Perluas atau ciutkan sidebar", "Kembangkan atau runtuhkan sidebar"),
    ("Dark mode", "Mode gelap", "Mod gelap"),
    ("Light mode", "Mode terang", "Mod terang"),
    ("Buyer Organizations", "Organisasi Pembeli", "Organisasi Pembeli"),
    ("Hooks & Events", "Hook & Acara", "Hook & Acara"),
    ("Audit Log", "Log Audit", "Log Audit"),
    ("Super Admin", "Super Admin", "Super Admin"),
    ("AI Applications", "Aplikasi AI", "Aplikasi AI"),
    ("AI Translate All", "Terjemahan AI Semua", "Terjemah AI Semua"),
    ("Generate with AI", "Buat dengan AI", "Jana dengan AI"),
    ("Save Product", "Simpan Produk", "Simpan Produk"),
    ("AI Import", "Impor AI", "Import AI"),
    ("Translation saved successfully", "Terjemahan berhasil disimpan", "Terjemahan berjaya disimpan"),
    ("Translation is taking longer than usual. Please wait or try again later.", "Terjemahan membutuhkan waktu lebih lama dari biasanya. Harap tunggu atau coba lagi nanti.", "Terjemahan mengambil masa lebih lama daripada biasa. Sila tunggu atau cuba lagi kemudian."),
    ("Save failed", "Gagal menyimpan", "Gagal menyimpan"),
    ("Load failed", "Gagal memuat", "Gagal memuatkan"),
    ("Delete failed", "Gagal menghapus", "Gagal memadam"),
    ("Confirm delete?", "Konfirmasi hapus?", "Sahkan padam?"),
    ("Are you sure", "Apakah Anda yakin", "Adakah anda pasti"),
    ("Failed to load", "Gagal memuat", "Gagal memuatkan"),
    ("Failed to ", "Gagal ", "Gagal "),
    (" successfully.", " berhasil.", " berjaya."),
    ("successfully.", "berhasil.", "berjaya."),
    ("Loading...", "Memuat...", "Memuatkan..."),
    ("Saving...", "Menyimpan...", "Menyimpan..."),
    ("Updating...", "Memperbarui...", "Mengemas kini..."),
    ("Deleting...", "Menghapus...", "Memadam..."),
    ("Applying...", "Menerapkan...", "Menggunakan..."),
    ("Importing...", "Mengimpor...", "Mengimport..."),
    ("Exporting…", "Mengekspor…", "Mengeksport…"),
    ("Analyzing...", "Menganalisis...", "Menganalisis..."),
    ("Assigning...", "Menetapkan...", "Menetapkan..."),
    ("No data available", "Tidak ada data", "Tiada data"),
    ("No data found", "Tidak ada data", "Tiada data dijumpai"),
    ("All Statuses", "Semua Status", "Semua Status"),
    ("All Status", "Semua Status", "Semua Status"),
    ("All Types", "Semua Tipe", "Semua Jenis"),
    ("All Groups", "Semua Grup", "Semua Kumpulan"),
    ("All Locales", "Semua Locale", "Semua Locale"),
    ("All Stock", "Semua Stok", "Semua Stok"),
    ("In Progress", "Sedang Berlangsung", "Sedang Dijalankan"),
    ("In Transit", "Dalam Transit", "Dalam Transit"),
    ("Due Soon", "Segera Jatuh Tempo", "Akan Tamat Tempoh"),
    ("Payment Terms", "Syarat Pembayaran", "Terma Pembayaran"),
    ("Payment Status", "Status Pembayaran", "Status Pembayaran"),
    ("Payment Method", "Metode Pembayaran", "Kaedah Pembayaran"),
    ("Credit Limit", "Batas Kredit", "Had Kredit"),
    ("Price List", "Daftar Harga", "Senarai Harga"),
    ("Order Items", "Item Pesanan", "Item Pesanan"),
    ("Order Number", "Nomor Pesanan", "Nombor Pesanan"),
    ("Shipping Address", "Alamat Pengiriman", "Alamat Penghantaran"),
    ("Base Price", "Harga Dasar", "Harga Asas"),
    ("Base Price (USD)", "Harga Dasar (USD)", "Harga Asas (USD)"),
    ("Lead Time", "Waktu Pengiriman", "Masa Penghantaran"),
    ("Stock Quantity", "Jumlah Stok", "Kuantiti Stok"),
    ("Batch Edit", "Edit Massal", "Edit Pukal"),
    ("Batch Delete", "Hapus Massal", "Padam Pukal"),
    ("Import XLSX", "Impor XLSX", "Import XLSX"),
    ("Export XLSX", "Ekspor XLSX", "Eksport XLSX"),
    ("Content Management", "Manajemen Konten", "Pengurusan Kandungan"),
    ("Translation Management", "Manajemen Terjemahan", "Pengurusan Terjemahan"),
    ("System Settings", "Pengaturan Sistem", "Tetapan Sistem"),
    ("Business Reports", "Laporan Bisnis", "Laporan Perniagaan"),
    ("Analytics & Reports", "Analitik & Laporan", "Analitik & Laporan"),
    ("Financial Overview", "Ikhtisar Keuangan", "Gambaran Kewangan"),
    ("Staff Management", "Manajemen Staf", "Pengurusan Kakitangan"),
    ("OEM Projects", "Proyek OEM", "Projek OEM"),
    ("Sales Channels", "Saluran Penjualan", "Saluran Jualan"),
    ("XLSX Tools", "Alat XLSX", "Alat XLSX"),
    ("Manage ", "Kelola ", "Urus "),
    ("Track and manage", "Lacak dan kelola", "Jejak dan urus"),
    ("Create ", "Buat ", "Cipta "),
    ("Edit ", "Edit ", "Edit "),
    ("Delete ", "Hapus ", "Padam "),
    ("Add ", "Tambah ", "Tambah "),
    ("New ", "Baru ", "Baharu "),
    ("Update ", "Perbarui ", "Kemas kini "),
    ("Search by ", "Cari berdasarkan ", "Cari mengikut "),
    ("Filter by ", "Filter berdasarkan ", "Tapis mengikut "),
    ("Search ", "Cari ", "Cari "),
    ("Filter ", "Filter ", "Tapis "),
    ("Export ", "Ekspor ", "Eksport "),
    ("Import ", "Impor ", "Import "),
    ("Loading ", "Memuat ", "Memuatkan "),
    ("No ", "Tidak ada ", "Tiada "),
    ("Save", "Simpan", "Simpan"),
    ("Cancel", "Batal", "Batal"),
    ("Delete", "Hapus", "Padam"),
    ("Edit", "Edit", "Edit"),
    ("Create", "Buat", "Cipta"),
    ("Update", "Perbarui", "Kemas kini"),
    ("Confirm", "Konfirmasi", "Sahkan"),
    ("Approve", "Setujui", "Luluskan"),
    ("Reject", "Tolak", "Tolak"),
    ("Send", "Kirim", "Hantar"),
    ("View", "Lihat", "Lihat"),
    ("Export", "Ekspor", "Eksport"),
    ("Import", "Impor", "Import"),
    ("Search", "Cari", "Cari"),
    ("Filter", "Filter", "Tapis"),
    ("Actions", "Aksi", "Tindakan"),
    ("Status", "Status", "Status"),
    ("Active", "Aktif", "Aktif"),
    ("Inactive", "Nonaktif", "Tidak aktif"),
    ("Draft", "Draf", "Draf"),
    ("Pending", "Menunggu", "Menunggu"),
    ("Paid", "Lunas", "Dibayar"),
    ("Sent", "Terkirim", "Dihantar"),
    ("Overdue", "Jatuh tempo", "Lewat tempoh"),
    ("Cancelled", "Dibatalkan", "Dibatalkan"),
    ("Completed", "Selesai", "Selesai"),
    ("Confirmed", "Dikonfirmasi", "Disahkan"),
    ("Delivered", "Diterima", "Diterima"),
    ("Shipped", "Dikirim", "Dihantar"),
    ("Production", "Produksi", "Pengeluaran"),
    ("Verified", "Terverifikasi", "Disahkan"),
    ("Rejected", "Ditolak", "Ditolak"),
    ("Approved", "Disetujui", "Diluluskan"),
    ("Customer", "Pelanggan", "Pelanggan"),
    ("Product", "Produk", "Produk"),
    ("Products", "Produk", "Produk"),
    ("Category", "Kategori", "Kategori"),
    ("Categories", "Kategori", "Kategori"),
    ("Order", "Pesanan", "Pesanan"),
    ("Orders", "Pesanan", "Pesanan"),
    ("Company", "Perusahaan", "Syarikat"),
    ("Companies", "Perusahaan", "Syarikat"),
    ("User", "Pengguna", "Pengguna"),
    ("Users", "Pengguna", "Pengguna"),
    ("Name", "Nama", "Nama"),
    ("Description", "Deskripsi", "Penerangan"),
    ("Amount", "Jumlah", "Jumlah"),
    ("Total", "Total", "Jumlah"),
    ("Date", "Tanggal", "Tarikh"),
    ("Notes", "Catatan", "Nota"),
    ("Email", "Email", "E-mel"),
    ("Phone", "Telepon", "Telefon"),
    ("Address", "Alamat", "Alamat"),
    ("Country", "Negara", "Negara"),
    ("Region", "Wilayah", "Wilayah"),
    ("Currency", "Mata uang", "Mata wang"),
    ("Quantity", "Jumlah", "Kuantiti"),
    ("Price", "Harga", "Harga"),
    ("Tax", "Pajak", "Cukai"),
    ("Shipping", "Pengiriman", "Penghantaran"),
    ("Invoice", "Faktur", "Invois"),
    ("Invoices", "Faktur", "Invois"),
    ("Shipment", "Pengiriman", "Penghantaran"),
    ("Shipments", "Pengiriman", "Penghantaran"),
    ("Inventory", "Inventaris", "Inventori"),
    ("Warehouse", "Gudang", "Gudang"),
    ("Carrier", "Kurir", "Kurier"),
    ("Destination", "Tujuan", "Destinasi"),
    ("Tracking", "Pelacakan", "Penjejakan"),
    ("Previous", "Sebelumnya", "Sebelum"),
    ("Next", "Berikutnya", "Seterusnya"),
    ("Yes", "Ya", "Ya"),
    ("No", "Tidak", "Tidak"),
    ("All", "Semua", "Semua"),
    ("None", "Tidak ada", "Tiada"),
    ("Type", "Tipe", "Jenis"),
    ("Value", "Nilai", "Nilai"),
    ("Balance", "Saldo", "Baki"),
    ("Code", "Kode", "Kod"),
    ("Rate", "Tarif", "Kadar"),
    ("Reason", "Alasan", "Sebab"),
    ("Notes", "Catatan", "Nota"),
    ("Summary", "Ringkasan", "Ringkasan"),
    ("Details", "Detail", "Butiran"),
    ("History", "Riwayat", "Sejarah"),
    ("Settings", "Pengaturan", "Tetapan"),
    ("Management", "Manajemen", "Pengurusan"),
    ("Analytics", "Analitik", "Analitik"),
    ("Financial", "Keuangan", "Kewangan"),
    ("Revenue", "Pendapatan", "Hasil"),
    ("Profit", "Laba", "Keuntungan"),
    ("Cost", "Biaya", "Kos"),
    ("Dashboard", "Dasbor", "Papan Pemuka"),
    ("Coupons", "Kupon", "Kupon"),
    ("Certifications", "Sertifikasi", "Pensijilan"),
    ("Organizations", "Organisasi", "Organisasi"),
    ("Channels", "Saluran", "Saluran"),
    ("Translations", "Terjemahan", "Terjemahan"),
    ("Inquiries", "Permintaan", "Pertanyaan"),
    ("Returns", "Retur", "Pemulangan"),
    ("Trades", "Perdagangan", "Perdagangan"),
    ("Staff", "Staf", "Kakitangan"),
    ("Content", "Konten", "Kandungan"),
    ("Pricing", "Harga", "Harga"),
    ("Percentage", "Persentase", "Peratusan"),
    ("Thumbnail", "Thumbnail", "Thumbnail"),
    ("Featured", "Unggulan", "Pilihan"),
    ("Halal", "Halal", "Halal"),
    ("Ingredients", "Bahan", "Bahan"),
    ("Allergens", "Alergen", "Alergen"),
    ("Certifications", "Sertifikasi", "Pensijilan"),
    ("Shelf Life", "Masa simpan", "Jangka hayat"),
    ("Storage", "Penyimpanan", "Penyimpanan"),
    ("Summary", "Ringkasan", "Ringkasan"),
    ("Title", "Judul", "Tajuk"),
    ("Subtitle", "Subjudul", "Sari kata"),
    ("Loading", "Memuat", "Memuatkan"),
    ("Error", "Kesalahan", "Ralat"),
    ("Success", "Berhasil", "Berjaya"),
    ("Warning", "Peringatan", "Amaran"),
    ("Failed", "Gagal", "Gagal"),
    ("Unknown", "Tidak diketahui", "Tidak diketahui"),
    ("Exception", "Pengecualian", "Pengecualian"),
    ("General", "Umum", "Am"),
    ("Remove", "Hapus", "Buang"),
    ("Apply", "Terapkan", "Guna"),
    ("Clear", "Hapus", "Kosongkan"),
    ("Select", "Pilih", "Pilih"),
    ("Back", "Kembali", "Kembali"),
    ("Close", "Tutup", "Tutup"),
    ("Open", "Buka", "Buka"),
    ("Print", "Cetak", "Cetak"),
    ("Expand", "Perluas", "Kembangkan"),
    ("Collapse", "Ciutkan", "Runtuhkan"),
    ("Refund", "Pengembalian dana", "Bayaran balik"),
    ("Assign", "Tetapkan", "Tetapkan"),
    ("Analyze", "Analisis", "Analisis"),
    ("Dispatch", "Kirim", "Hantar"),
    ("Delivered", "Diterima", "Diterima"),
]

# Build EXACT lookup from triples + external phrases file
EXACT_ID: dict[str, str] = {}
EXACT_MS: dict[str, str] = {}
for en, id_v, ms_v in PHRASE_TRIPLES:
    EXACT_ID[en] = id_v
    EXACT_MS[en] = ms_v
for en, pair in PHRASES.items():
    EXACT_ID[en] = pair["id"]
    EXACT_MS[en] = pair["ms"]

PLACEHOLDER_RE = re.compile(r"(\{[a-zA-Z0-9_]+\})")
TOKEN_RE = re.compile(
    r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|1200×630|aʷ|→|…|\.\.\.|&|[0-9]+(?:\.[0-9]+)?%?|[A-Za-z][A-Za-z0-9+#./\-]{1,})"
)
LATIN_RE = re.compile(r"[A-Za-z]{3,}")


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


def translate_text(text: str, locale: str) -> str:
    if not isinstance(text, str):
        return text
    exact = EXACT_ID if locale == "id" else EXACT_MS
    if text in exact:
        return exact[text]

    protected, reps = protect(text)
    out = protected

    # Longest phrase replacement on protected text
    triples = sorted(PHRASE_TRIPLES, key=lambda x: len(x[0]), reverse=True)
    for en, id_v, ms_v in triples:
        tr = id_v if locale == "id" else ms_v
        if en in out:
            out = out.replace(en, tr)

    # Single-word EXACT from triples (whole words)
    for en, id_v, ms_v in sorted(PHRASE_TRIPLES, key=lambda x: len(x[0]), reverse=True):
        if " " in en or len(en) <= 2:
            continue
        tr = id_v if locale == "id" else ms_v
        out = re.sub(rf"\b{re.escape(en)}\b", tr, out)

    out = restore(out, reps)

    if out in exact:
        return exact[out]
    return out


def walk(obj: dict, prefix: str, locale: str) -> None:
    path_map = PATH_ID if locale == "id" else PATH_MS
    for key, val in obj.items():
        path = f"{prefix}.{key}" if prefix else key
        if isinstance(val, dict):
            walk(val, path, locale)
        elif path in path_map:
            obj[key] = path_map[path]
        else:
            obj[key] = translate_text(val, locale)


def flatten(obj: dict, prefix: str = "") -> dict[str, str]:
    out: dict[str, str] = {}
    for k, v in obj.items():
        p = f"{prefix}.{k}" if prefix else k
        if isinstance(v, dict):
            out.update(flatten(v, p))
        else:
            out[p] = v
    return out


def main() -> None:
    en_admin = json.loads((I18N / "en" / "admin.json").read_text(encoding="utf-8"))
    en_flat = flatten(en_admin)

    # Seed path overrides from nav/dashboard/products priority (from generate script)
    priority_prefixes = (
        "admin.nav.", "admin.dashboard.", "admin.products.", "admin.inventory.",
        "admin.categories.", "admin.coupons.", "admin.shippingRates.", "admin.taxRates.",
        "admin.status_options.",
    )
    for path, en_val in en_flat.items():
        if any(path.startswith(p) for p in priority_prefixes):
            PATH_ID[path] = translate_text(en_val, "id")
            PATH_MS[path] = translate_text(en_val, "ms")

    for locale in ("id", "ms"):
        data = deepcopy(en_admin)
        walk(data, "", locale)
        out = I18N / locale / "admin.json"
        out.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print(f"Wrote {out}")


if __name__ == "__main__":
    main()
