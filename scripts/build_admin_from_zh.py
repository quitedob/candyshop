#!/usr/bin/env python3
"""Build id/ms admin.json using zh semantic reference + phrase dictionary."""

from __future__ import annotations

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"

GLOSSARY_TERMS = [
    "CandyPro", "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS", "GTIN",
    "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL", "RCEP", "CBM",
    "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "Incoterms", "B2B", "KYB",
    "RFM", "Slug", "Plan-Execute-Replan", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA",
    "Stripe", "PayPal", "Chatbot", "Agent", "Cursor", "translate_content", "CRUD",
    "Facebook", "LinkedIn", "Super Admin", "Superadmin", "Grade A", "Ctrl+K", "⌘K",
    "1200×630", "aʷ", "Diff", "Trade Assistant", "B2B Coordinator", "Product Recommend",
]

# Chinese phrase -> (Indonesian, Malay)
PHRASES: list[tuple[str, str, str]] = [
    ("管理后台", "Admin", "Admin"),
    ("退出登录", "Keluar", "Log keluar"),
    ("暗色模式", "Mode gelap", "Mod gelap"),
    ("浅色模式", "Mode terang", "Mod terang"),
    ("打开侧边导航", "Buka menu sidebar", "Buka menu sidebar"),
    ("关闭侧边导航", "Tutup menu sidebar", "Tutup menu sidebar"),
    ("展开或收起侧边栏", "Perluas atau ciutkan sidebar", "Kembangkan atau runtuhkan sidebar"),
    ("仪表盘", "Dasbor", "Papan Pemuka"),
    ("用户管理", "Pengguna", "Pengguna"),
    ("询价管理", "Permintaan Penawaran", "Pertanyaan"),
    ("订单管理", "Pesanan", "Pesanan"),
    ("退货管理", "Retur", "Pemulangan"),
    ("产品管理", "Produk", "Produk"),
    ("分类管理", "Kategori", "Kategori"),
    ("企业管理", "Perusahaan", "Syarikat"),
    ("贸易管理", "Perdagangan", "Perdagangan"),
    ("价格管理", "Harga", "Harga"),
    ("OEM项目", "Proyek OEM", "Projek OEM"),
    ("内容管理", "Konten", "Kandungan"),
    ("库存管理", "Inventaris", "Inventori"),
    ("XLSX 工具", "Alat XLSX", "Alat XLSX"),
    ("发货管理", "Pengiriman", "Penghantaran"),
    ("发票管理", "Faktur", "Invois"),
    ("运营管理", "Manajemen", "Pengurusan"),
    ("超级管理员", "Super Admin", "Super Admin"),
    ("数据分析", "Analitik", "Analitik"),
    ("财务管理", "Keuangan", "Kewangan"),
    ("认证管理", "Sertifikasi", "Pensijilan"),
    ("员工管理", "Staf", "Kakitangan"),
    ("操作日志", "Log Audit", "Log Audit"),
    ("系统设置", "Pengaturan", "Tetapan"),
    ("翻译管理", "Terjemahan", "Terjemahan"),
    ("AI 应用", "Aplikasi AI", "Aplikasi AI"),
    ("采购组织", "Organisasi Pembeli", "Organisasi Pembeli"),
    ("渠道", "Saluran", "Saluran"),
    ("钩子与事件", "Hook & Acara", "Hook & Acara"),
    ("优惠券", "Kupon", "Kupon"),
    ("运费与税率", "Pengiriman & Pajak", "Penghantaran & Cukai"),
    ("税率管理", "Tarif Pajak", "Kadar Cukai"),
    ("平台概览与关键指标一览。", "Ikhtisar platform dan metrik utama sekilas.", "Gambaran platform dan metrik utama sepintas lalu."),
    ("正在加载仪表盘数据...", "Memuat statistik dasbor...", "Memuatkan statistik papan pemuka..."),
    ("加载仪表盘失败", "Gagal memuat dasbor", "Ralat memuatkan papan pemuka"),
    ("用户总数", "Total Pengguna", "Jumlah Pengguna"),
    ("订单总数", "Total Pesanan", "Jumlah Pesanan"),
    ("询价总数", "Total Permintaan", "Jumlah Pertanyaan"),
    ("总收入", "Total Pendapatan", "Jumlah Hasil"),
    ("本周", "minggu ini", "minggu ini"),
    ("本周新增 {count}", "{count} baru minggu ini", "{count} baharu minggu ini"),
    ("待处理", "menunggu", "menunggu"),
    ("最近活动", "Aktivitas Terbaru", "Aktiviti Terkini"),
    ("活动", "Aktivitas", "Aktiviti"),
    ("时间", "Waktu", "Masa"),
    ("暂无最近活动", "Tidak ada aktivitas terbaru", "Tiada aktiviti terkini"),
    ("本月营收", "Pendapatan Bulan Ini", "Hasil Bulan Ini"),
    ("活跃客户", "Pelanggan Aktif", "Pelanggan Aktif"),
    ("平均订单额", "Nilai Pesanan Rata-rata", "Nilai Pesanan Purata"),
    ("转化率", "Tingkat Konversi", "Kadar Penukaran"),
    ("营收趋势（近30天）", "Pendapatan (30 Hari)", "Hasil (30 Hari)"),
    ("营收", "Pendapatan", "Hasil"),
    ("日期", "Tanggal", "Tarikh"),
    ("管理产品目录，支持完整的增删改查操作。", "Kelola katalog produk dengan operasi CRUD lengkap.", "Urus katalog produk dengan operasi CRUD lengkap."),
    ("添加产品", "Tambah Produk", "Tambah Produk"),
    ("批量翻译缺失语言", "Terjemahkan locale yang belum ada", "Terjemah locale yang tiada"),
    ("翻译中…", "Menerjemahkan…", "Menterjemah…"),
    ("搜索", "Cari", "Cari"),
    ("按产品名称或 SKU 搜索...", "Cari berdasarkan nama produk atau SKU...", "Cari mengikut nama produk atau SKU..."),
    ("状态筛选", "Filter Status", "Tapis Status"),
    ("已发布（默认）", "Diterbitkan (default)", "Diterbitkan (lalai)"),
    ("全部状态", "Semua status", "Semua status"),
    ("草稿", "Draf", "Draf"),
    ("已下架", "Nonaktif", "Tidak aktif"),
    ("正在加载产品...", "Memuat produk...", "Memuatkan produk..."),
    ("暂无产品。", "Tidak ada produk.", "Tiada produk."),
    ("产品", "Produk", "Produk"),
    ("分类", "Kategori", "Kategori"),
    ("最小起订量", "MOQ", "MOQ"),
    ("库存", "Stok", "Stok"),
    ("状态", "Status", "Status"),
    ("编辑", "Edit", "Edit"),
    ("删除", "Hapus", "Padam"),
    ("编辑产品", "Edit Produk", "Edit Produk"),
    ("创建产品", "Buat Produk", "Cipta Produk"),
    ("名称", "Nama", "Nama"),
    ("URL 别名", "Slug", "Slug"),
    ("别名", "Alias", "Alias"),
    ("分类 Slug", "Slug Kategori", "Slug Kategori"),
    ("分类别名", "Alias Kategori", "Alias Kategori"),
    ("基础价格 (美元)", "Harga Dasar (USD)", "Harga Asas (USD)"),
    ("交货周期", "Waktu Pengiriman", "Masa Penghantaran"),
    ("库存数量", "Jumlah Stok", "Kuantiti Stok"),
    ("清真", "Halal", "Halal"),
    ("推荐", "Unggulan", "Pilihan"),
    ("摘要", "Ringkasan", "Ringkasan"),
    ("描述", "Deskripsi", "Penerangan"),
    ("缩略图", "Thumbnail", "Thumbnail"),
    ("OG 分享图", "Gambar OG (Berbagi Sosial)", "Imej OG (Kongsi Sosial)"),
    ("上传图片", "Unggah Gambar", "Muat Naik Imej"),
    ("配料", "Bahan", "Bahan"),
    ("过敏原", "Alergen", "Alergen"),
    ("保质期", "Masa simpan", "Jangka hayat"),
    ("储存条件", "Penyimpanan", "Penyimpanan"),
    ("取消", "Batal", "Batal"),
    ("保存中...", "Menyimpan...", "Menyimpan..."),
    ("更新", "Perbarui", "Kemas kini"),
    ("创建", "Buat", "Cipta"),
    ("产品名称为必填项。", "Nama produk wajib diisi.", "Nama produk diperlukan."),
    ("确定删除此产品？", "Hapus produk ini?", "Padam produk ini?"),
    ("产品创建成功。", "Produk berhasil dibuat.", "Produk berjaya dicipta."),
    ("产品更新成功。", "Produk berhasil diperbarui.", "Produk berjaya dikemas kini."),
    ("产品删除成功。", "Produk berhasil dihapus.", "Produk berjaya dipadam."),
    ("显示第 {from} 到 {to} 条，共 {total} 条", "Menampilkan {from} hingga {to} dari {total}", "Memaparkan {from} hingga {to} daripada {total}"),
    ("上一页", "Sebelumnya", "Sebelum"),
    ("下一页", "Berikutnya", "Seterusnya"),
    ("追踪和管理产品库存水平、价值和调整记录。", "Lacak dan kelola tingkat stok, nilai, dan penyesuaian produk.", "Jejak dan urus tahap stok, nilai, dan pelarasan produk."),
    ("导出", "Ekspor", "Eksport"),
    ("调整库存", "Sesuaikan Stok", "Laraskan Stok"),
    ("产品总数", "Total Produk", "Jumlah Produk"),
    ("低库存", "Stok Rendah", "Stok Rendah"),
    ("缺货", "Stok Habis", "Stok Habis"),
    ("有库存", "Tersedia", "Ada Stok"),
    ("总价值", "Total Nilai", "Jumlah Nilai"),
    ("按产品名称或SKU搜索...", "Cari berdasarkan nama produk atau SKU...", "Cari mengikut nama produk atau SKU..."),
    ("按库存状态筛选", "Filter berdasarkan status stok", "Tapis mengikut status stok"),
    ("按分类筛选", "Filter berdasarkan kategori", "Tapis mengikut kategori"),
    ("全部库存", "Semua Stok", "Semua Stok"),
    ("全部分类", "Semua Kategori", "Semua Kategori"),
    ("单价", "Nilai Satuan", "Nilai Unit"),
    ("操作", "Aksi", "Tindakan"),
    ("正在加载库存...", "Memuat inventaris...", "Memuatkan inventori..."),
    ("暂无库存数据。", "Tidak ada data inventaris.", "Tiada data inventori."),
    ("历史记录", "Riwayat", "Sejarah"),
    ("调整类型", "Tipe Penyesuaian", "Jenis Pelarasan"),
    ("设置为", "Atur Ke", "Tetapkan Ke"),
    ("增加", "Tambah", "Tambah"),
    ("减少", "Kurangi", "Tolak"),
    ("原因", "Alasan", "Sebab"),
    ("补货", "Restok", "Stok Semula"),
    ("修正", "Koreksi", "Pembetulan"),
    ("损耗/损坏", "Kerusakan/Kehilangan", "Kerosakan/Kehilangan"),
    ("客户退货", "Retur Pelanggan", "Pemulangan Pelanggan"),
    ("库存盘点", "Audit Inventaris", "Audit Inventori"),
    ("其他", "Lainnya", "Lain"),
    ("备注", "Catatan", "Nota"),
    ("保存", "Simpan", "Simpan"),
    ("库存历史", "Riwayat Stok", "Sejarah Stok"),
    ("暂无调整记录。", "Tidak ada riwayat penyesuaian.", "Tiada sejarah pelarasan."),
    ("全选", "Pilih Semua", "Pilih Semua"),
    ("已选 {count} 项", "{count} dipilih", "{count} dipilih"),
    ("导出选中", "Ekspor Terpilih", "Eksport Terpilih"),
    ("导入 XLSX", "Impor XLSX", "Import XLSX"),
    ("管理产品分类及多语言名称、别名与描述。", "Kelola kategori produk dengan nama, alias, dan deskripsi multibahasa.", "Urus kategori produk dengan nama, alias, dan penerangan pelbagai bahasa."),
    ("新增分类", "Tambah Kategori", "Tambah Kategori"),
    ("编辑分类", "Edit Kategori", "Edit Kategori"),
    ("创建分类", "Buat Kategori", "Cipta Kategori"),
    ("产品数", "Produk", "Produk"),
    ("加载中...", "Memuat...", "Memuatkan..."),
    ("暂无分类", "Tidak ada kategori", "Tiada kategori"),
    ("图标", "Ikon", "Ikon"),
    ("多语言翻译", "Terjemahan", "Terjemahan"),
    ("请填写中文分类名称", "Nama kategori Tionghoa wajib diisi", "Nama kategori Cina diperlukan"),
    ("请填写 Slug", "Slug wajib diisi", "Slug diperlukan"),
    ("确定删除此分类？", "Hapus kategori ini?", "Padam kategori ini?"),
    ("分类创建成功", "Kategori dibuat", "Kategori dicipta"),
    ("分类更新成功", "Kategori diperbarui", "Kategori dikemas kini"),
    ("分类已删除", "Kategori dihapus", "Kategori dipadam"),
    ("加载分类失败", "Gagal memuat kategori", "Gagal memuatkan kategori"),
    ("保存失败", "Gagal menyimpan", "Gagal menyimpan"),
    ("优惠券与礼品卡", "Kupon & Kartu Hadiah", "Kupon & Kad Hadiah"),
    ("促销码与预付余额。", "Kode promosi dan kredit prabayar.", "Kod promosi dan kredit prabayar."),
    ("创建优惠券", "Buat Kupon", "Cipta Kupon"),
    ("创建礼品卡", "Buat Kartu Hadiah", "Cipta Kad Hadiah"),
    ("礼品卡", "Kartu Hadiah", "Kad Hadiah"),
    ("暂无优惠券。", "Tidak ada kupon.", "Tiada kupon."),
    ("暂无礼品卡。", "Tidak ada kartu hadiah.", "Tiada kad hadiah."),
    ("代码", "Kode", "Kod"),
    ("类型", "Tipe", "Jenis"),
    ("面值", "Nilai", "Nilai"),
    ("最低订单", "Pesanan Min.", "Pesanan Min."),
    ("开始时间", "Mulai", "Bermula"),
    ("过期时间", "Berakhir", "Tamat"),
    ("余额", "Saldo", "Baki"),
    ("货币", "Mata uang", "Mata wang"),
    ("百分比", "Persentase", "Peratusan"),
    ("固定金额", "Jumlah Tetap", "Jumlah Tetap"),
    ("已用", "Digunakan", "Digunakan"),
    ("初始", "Awal", "Permulaan"),
    ("共 {total} 条", "{total} total", "{total} jumlah"),
    ("运费费率", "Tarif Pengiriman", "Kadar Penghantaran"),
    ("按目的地的运费表。预计天数仅供报价参考，非承诺运输时效。", "Tabel tarif pengiriman berdasarkan tujuan. Perkiraan hari hanya referensi penawaran — bukan jaminan waktu transit.", "Jadual kadar penghantaran mengikut destinasi. Anggaran hari adalah rujukan sebut harga — bukan jaminan masa transit."),
    ("添加费率", "Tambah Tarif", "Tambah Kadar"),
    ("暂无运费费率。", "Tidak ada tarif pengiriman.", "Tiada kadar penghantaran."),
    ("目的地", "Tujuan", "Destinasi"),
    ("承运商", "Kurir", "Kurier"),
    ("最小重量 (kg)", "Berat Min. (kg)", "Berat Min. (kg)"),
    ("最大重量 (kg)", "Berat Max. (kg)", "Berat Max. (kg)"),
    ("基础运费", "Biaya Dasar", "Kos Asas"),
    ("每公斤费用", "Biaya per kg", "Kos per kg"),
    ("预计天数", "Perkiraan Hari", "Angg. Hari"),
    ("内部报价用的参考运输天数，可能变动，下单时确认。", "Perkiraan transit referensi untuk penawaran internal. Dapat berubah; konfirmasi saat pesanan.", "Anggaran transit rujukan untuk sebut harga dalaman. Boleh berubah; sahkan semasa pesanan."),
    ("启用", "Aktif", "Aktif"),
    ("天数", "Hari", "Hari"),
    ("税率管理", "Tarif Pajak", "Kadar Cukai"),
    ("按目的国/地区配置的税率，用于购物车结账估算。", "Tarif pajak negara/wilayah tujuan yang diterapkan saat checkout keranjang.", "Kadar cukai negara/wilayah destinasi yang digunakan semasa checkout troli."),
    ("添加税率", "Tambah Tarif Pajak", "Tambah Kadar Cukai"),
    ("暂无税率配置。", "Tidak ada tarif pajak.", "Tiada kadar cukai dikonfigurasi."),
    ("国家 (ISO)", "Negara (ISO)", "Negara (ISO)"),
    ("地区/州", "Wilayah / Provinsi", "Wilayah / Negeri"),
    ("税率", "Tarif", "Kadar"),
    ("小数值，如 0.08875 = 8.875%", "Pecahan desimal, mis. 0,08875 = 8,875%", "Pecahan perpuluhan, cth. 0.08875 = 8.875%"),
    ("已发送", "Terkirim", "Dihantar"),
    ("已付款", "Lunas", "Dibayar"),
    ("已逾期", "Jatuh tempo", "Lewat tempoh"),
    ("已取消", "Dibatalkan", "Dibatalkan"),
    ("禁用", "Nonaktif", "Tidak aktif"),
    ("确认删除？", "Konfirmasi hapus?", "Sahkan padam?"),
    ("中文", "Tionghoa", "Cina"),
    ("印尼文", "Indonesia", "Indonesia"),
    ("马来文", "Melayu", "Melayu"),
    ("英文", "Inggris", "Inggeris"),
    ("ID", "ID", "ID"),
    ("MS", "MS", "MS"),
    ("Bahasa Indonesia", "Bahasa Indonesia", "Bahasa Indonesia"),
    ("Bahasa Melayu", "Bahasa Melayu", "Bahasa Melayu"),
]

# Sort longest first for replacement
PHRASES.sort(key=lambda x: -len(x[0]))


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
    reps: list[tuple[str, str]] = []
    protected = text
    for i, term in enumerate(sorted(GLOSSARY_TERMS, key=len, reverse=True)):
        if term in protected:
            ph = f"__G{i}__"
            reps.append((ph, term))
            protected = protected.replace(term, ph)
    for m in re.finditer(r"\{[a-zA-Z0-9_]+\}", text):
        token = m.group(0)
        ph = f"__P{len(reps)}__"
        reps.append((ph, token))
        protected = protected.replace(token, ph, 1)
    return protected, reps


def restore(text: str, reps: list[tuple[str, str]]) -> str:
    for ph, orig in reps:
        text = text.replace(ph, orig)
    return text


def translate_zh(zh: str, locale: str) -> str:
    protected, reps = protect(zh)
    result = protected
    for zh_phrase, id_text, ms_text in PHRASES:
        if zh_phrase in result:
            dst = id_text if locale == "id" else ms_text
            result = result.replace(zh_phrase, dst)
    # If still has CJK, fall back to en via key - leave as mixed; prefer id/ms words for remaining CJK
    result = restore(result, reps)
    return result


def build(locale: str) -> dict:
    from build_admin_id_ms_offline import translate as rule_translate

    en_flat = flatten(json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8"))["admin"])
    zh_flat = flatten(json.loads((ROOT / "zh" / "admin.json").read_text(encoding="utf-8"))["admin"])

    # Map each English source string to its Chinese equivalent (when available)
    en_to_zh: dict[str, str] = {}
    for key, en_val in en_flat.items():
        zh_val = zh_flat.get(key, en_val)
        if zh_val != en_val:
            en_to_zh[en_val] = zh_val

    def translate_en(en_val: str) -> str:
        if en_val in en_to_zh:
            return translate_zh(en_to_zh[en_val], locale)
        return rule_translate(en_val, locale)

    out: dict[str, str] = {key: translate_en(en_val) for key, en_val in en_flat.items()}
    return {"admin": unflatten(out)}


def main() -> None:
    for locale in ("id", "ms"):
        data = build(locale)
        path = ROOT / locale / "admin.json"
        path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print(f"Wrote {path}")


if __name__ == "__main__":
    main()
