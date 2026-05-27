#!/usr/bin/env python3
"""Apply high-quality priority translations and fill gaps from zh reference."""

from __future__ import annotations

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"

# en string -> {ar, th, vi} for high-frequency UI terms Lingva misses
TERM: dict[str, dict[str, str]] = {
    "Dashboard": {"ar": "لوحة المعلومات", "th": "แดชบอร์ด", "vi": "Bảng điều khiển"},
    "Users": {"ar": "المستخدمون", "th": "ผู้ใช้", "vi": "Người dùng"},
    "Inquiries": {"ar": "الاستفسارات", "th": "คำขอสอบถาม", "vi": "Yêu cầu báo giá"},
    "Orders": {"ar": "الطلبات", "th": "คำสั่งซื้อ", "vi": "Đơn hàng"},
    "Returns": {"ar": "المرتجعات", "th": "การคืนสินค้า", "vi": "Trả hàng"},
    "Products": {"ar": "المنتجات", "th": "สินค้า", "vi": "Sản phẩm"},
    "Categories": {"ar": "الفئات", "th": "หมวดหมู่", "vi": "Danh mục"},
    "Companies": {"ar": "الشركات", "th": "บริษัท", "vi": "Doanh nghiệp"},
    "Trades": {"ar": "التجارة", "th": "ธุรกรรมการค้า", "vi": "Thương mại"},
    "Pricing": {"ar": "التسعير", "th": "ราคา", "vi": "Giá"},
    "OEM Projects": {"ar": "مشاريع OEM", "th": "โครงการ OEM", "vi": "Dự án OEM"},
    "Content": {"ar": "المحتوى", "th": "เนื้อหา", "vi": "Nội dung"},
    "Inventory": {"ar": "المخزون", "th": "สินค้าคงคลัง", "vi": "Kho hàng"},
    "XLSX Tools": {"ar": "أدوات XLSX", "th": "เครื่องมือ XLSX", "vi": "Công cụ XLSX"},
    "Shipments": {"ar": "الشحنات", "th": "การจัดส่ง", "vi": "Vận chuyển"},
    "Invoices": {"ar": "الفواتير", "th": "ใบแจ้งหนี้", "vi": "Hóa đơn"},
    "Management": {"ar": "الإدارة", "th": "การจัดการ", "vi": "Quản lý"},
    "Super Admin": {"ar": "المسؤول الأعلى", "th": "ผู้ดูแลระบบสูงสุด", "vi": "Quản trị viên cấp cao"},
    "Analytics": {"ar": "التحليلات", "th": "การวิเคราะห์", "vi": "Phân tích"},
    "Financial": {"ar": "المالية", "th": "การเงิน", "vi": "Tài chính"},
    "Certifications": {"ar": "الشهادات", "th": "ใบรับรอง", "vi": "Chứng nhận"},
    "Staff": {"ar": "الموظفون", "th": "พนักงาน", "vi": "Nhân viên"},
    "Audit Log": {"ar": "سجل التدقيق", "th": "บันทึกการตรวจสอบ", "vi": "Nhật ký kiểm toán"},
    "Settings": {"ar": "الإعدادات", "th": "การตั้งค่า", "vi": "Cài đặt"},
    "Translations": {"ar": "الترجمات", "th": "การแปลภาษา", "vi": "Bản dịch"},
    "AI Apps": {"ar": "تطبيقات AI", "th": "แอป AI", "vi": "Ứng dụng AI"},
    "Organizations": {"ar": "المنظمات", "th": "องค์กร", "vi": "Tổ chức"},
    "Channels": {"ar": "القنوات", "th": "ช่องทาง", "vi": "Kênh"},
    "Webhooks": {"ar": "Webhooks", "th": "Webhooks", "vi": "Webhooks"},
    "Hooks & Events": {"ar": "الخطافات والأحداث", "th": "Hooks & Events", "vi": "Hook và sự kiện"},
    "Coupons": {"ar": "القسائم", "th": "คูปอง", "vi": "Phiếu giảm giá"},
    "Shipping & Tax": {"ar": "الشحن والضرائب", "th": "ค่าจัดส่งและภาษี", "vi": "Phí vận chuyển & thuế"},
    "Tax Rates": {"ar": "معدلات الضريبة", "th": "อัตราภาษี", "vi": "Thuế suất"},
    "Logout": {"ar": "تسجيل الخروج", "th": "ออกจากระบบ", "vi": "Đăng xuất"},
    "Dark mode": {"ar": "الوضع الداكن", "th": "โหมดมืด", "vi": "Chế độ tối"},
    "Light mode": {"ar": "الوضع الفاتح", "th": "โหมดสว่าง", "vi": "Chế độ sáng"},
    "Open sidebar menu": {"ar": "فتح قائمة الشريط الجانبي", "th": "เปิดเมนูแถบด้านข้าง", "vi": "Mở menu thanh bên"},
    "Close sidebar menu": {"ar": "إغلاق قائمة الشريط الجانبي", "th": "ปิดเมนูแถบด้านข้าง", "vi": "Đóng menu thanh bên"},
    "Expand or collapse sidebar": {"ar": "توسيع أو طي الشريط الجانبي", "th": "ขยายหรือยุบแถบด้านข้าง", "vi": "Mở rộng hoặc thu gọn thanh bên"},
    "Platform overview and key metrics at a glance.": {
        "ar": "نظرة عامة على المنصة والمقاييس الرئيسية في لمحة.",
        "th": "ภาพรวมแพลตฟอร์มและตัวชี้วัดสำคัญในหน้าเดียว",
        "vi": "Tổng quan nền tảng và các chỉ số chính trong nháy mắt.",
    },
    "Loading dashboard statistics...": {
        "ar": "جاري تحميل إحصائيات لوحة المعلومات...",
        "th": "กำลังโหลดสถิติแดชบอร์ด...",
        "vi": "Đang tải thống kê bảng điều khiển...",
    },
    "Error loading dashboard": {
        "ar": "خطأ في تحميل لوحة المعلومات",
        "th": "เกิดข้อผิดพลาดในการโหลดแดชบอร์ด",
        "vi": "Lỗi tải bảng điều khiển",
    },
    "Total Users": {"ar": "إجمالي المستخدمين", "th": "ผู้ใช้ทั้งหมด", "vi": "Tổng người dùng"},
    "Total Orders": {"ar": "إجمالي الطلبات", "th": "คำสั่งซื้อทั้งหมด", "vi": "Tổng đơn hàng"},
    "Total Inquiries": {"ar": "إجمالي الاستفسارات", "th": "คำขอสอบถามทั้งหมด", "vi": "Tổng yêu cầu báo giá"},
    "Total Revenue": {"ar": "إجمالي الإيرادات", "th": "รายได้รวม", "vi": "Tổng doanh thu"},
    "this week": {"ar": "هذا الأسبوع", "th": "สัปดาห์นี้", "vi": "Tuần này"},
    "{count} new this week": {"ar": "{count} جديد هذا الأسبوع", "th": "ใหม่ {count} รายการสัปดาห์นี้", "vi": "Mới {count} tuần này"},
    "pending": {"ar": "قيد الانتظار", "th": "รอดำเนินการ", "vi": "Đang chờ"},
    "Recent Activity": {"ar": "النشاط الأخير", "th": "กิจกรรมล่าสุด", "vi": "Hoạt động gần đây"},
    "Activity": {"ar": "النشاط", "th": "กิจกรรม", "vi": "Hoạt động"},
    "Time": {"ar": "الوقت", "th": "เวลา", "vi": "Thời gian"},
    "No recent activity": {"ar": "لا يوجد نشاط حديث", "th": "ไม่มีกิจกรรมล่าสุด", "vi": "Không có hoạt động gần đây"},
    "Revenue This Month": {"ar": "الإيرادات هذا الشهر", "th": "รายได้เดือนนี้", "vi": "Doanh thu tháng này"},
    "Active Customers": {"ar": "العملاء النشطون", "th": "ลูกค้าที่ใช้งานอยู่", "vi": "Khách hàng đang hoạt động"},
    "Avg Order Value": {"ar": "متوسط قيمة الطلب", "th": "มูลค่าคำสั่งซื้อเฉลี่ย", "vi": "Giá trị đơn hàng trung bình"},
    "Conversion Rate": {"ar": "معدل التحويل", "th": "อัตราการแปลง", "vi": "Tỷ lệ chuyển đổi"},
    "Revenue (30 Days)": {"ar": "الإيرادات (30 يومًا)", "th": "รายได้ (30 วัน)", "vi": "Doanh thu (30 ngày)"},
    "Revenue": {"ar": "الإيرادات", "th": "รายได้", "vi": "Doanh thu"},
    "Date": {"ar": "التاريخ", "th": "วันที่", "vi": "Ngày"},
    "Draft": {"ar": "مسودة", "th": "ฉบับร่าง", "vi": "Bản nháp"},
    "Sent": {"ar": "مُرسل", "th": "ส่งแล้ว", "vi": "Đã gửi"},
    "Paid": {"ar": "مدفوع", "th": "ชำระแล้ว", "vi": "Đã thanh toán"},
    "Overdue": {"ar": "متأخر", "th": "เกินกำหนด", "vi": "Quá hạn"},
    "Cancelled": {"ar": "ملغى", "th": "ยกเลิกแล้ว", "vi": "Đã hủy"},
    "Active": {"ar": "نشط", "th": "ใช้งาน", "vi": "Đang hoạt động"},
    "Inactive": {"ar": "غير نشط", "th": "ไม่ใช้งาน", "vi": "Không hoạt động"},
    "Cancel": {"ar": "إلغاء", "th": "ยกเลิก", "vi": "Hủy"},
    "Save": {"ar": "حفظ", "th": "บันทึก", "vi": "Lưu"},
    "Delete": {"ar": "حذف", "th": "ลบ", "vi": "Xóa"},
    "Edit": {"ar": "تعديل", "th": "แก้ไข", "vi": "Chỉnh sửa"},
    "Create": {"ar": "إنشاء", "th": "สร้าง", "vi": "Tạo"},
    "Search": {"ar": "بحث", "th": "ค้นหา", "vi": "Tìm kiếm"},
    "Loading...": {"ar": "جاري التحميل...", "th": "กำลังโหลด...", "vi": "Đang tải..."},
    "Loading…": {"ar": "جاري التحميل…", "th": "กำลังโหลด…", "vi": "Đang tải…"},
    "Previous": {"ar": "السابق", "th": "ก่อนหน้า", "vi": "Trước"},
    "Next": {"ar": "التالي", "th": "ถัดไป", "vi": "Tiếp"},
    "Status": {"ar": "الحالة", "th": "สถานะ", "vi": "Trạng thái"},
    "Actions": {"ar": "الإجراءات", "th": "การดำเนินการ", "vi": "Thao tác"},
    "Pending": {"ar": "قيد الانتظار", "th": "รอดำเนินการ", "vi": "Đang chờ"},
    "Unknown": {"ar": "غير معروف", "th": "ไม่ทราบ", "vi": "Không xác định"},
    "Name": {"ar": "الاسم", "th": "ชื่อ", "vi": "Tên"},
    "Product": {"ar": "المنتج", "th": "สินค้า", "vi": "Sản phẩm"},
    "Customer": {"ar": "العميل", "th": "ลูกค้า", "vi": "Khách hàng"},
    "Order": {"ar": "الطلب", "th": "คำสั่งซื้อ", "vi": "Đơn hàng"},
    "Close": {"ar": "إغلاق", "th": "ปิด", "vi": "Đóng"},
    "Showing {from} to {to} of {total} results": {
        "ar": "عرض {from} إلى {to} من {total} نتيجة",
        "th": "แสดง {from} ถึง {to} จาก {total} รายการ",
        "vi": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
    },
    "Showing {from} to {to} of {total}": {
        "ar": "عرض {from} إلى {to} من {total}",
        "th": "แสดง {from} ถึง {to} จาก {total}",
        "vi": "Hiển thị {from} đến {to} trong tổng số {total}",
    },
    "Updated {updated}, {errors} errors": {
        "ar": "تم تحديث {updated}، {errors} أخطاء",
        "th": "อัปเดต {updated} รายการ, {errors} ข้อผิดพลาด",
        "vi": "Đã cập nhật {updated}, {errors} lỗi",
    },
    "Business Reports": {"ar": "تقارير الأعمال", "th": "รายงานธุรกิจ", "vi": "Báo cáo kinh doanh"},
    "Inquiry Conversion": {"ar": "تحويل الاستفسارات", "th": "การแปลงคำขอสอบถาม", "vi": "Chuyển đổi yêu cầu báo giá"},
    "Failed to load report": {"ar": "فشل تحميل التقرير", "th": "โหลดรายงานไม่สำเร็จ", "vi": "Không thể tải báo cáo"},
    "No data available": {"ar": "لا توجد بيانات", "th": "ไม่มีข้อมูล", "vi": "Không có dữ liệu"},
    "Analytics & Reports": {"ar": "التحليلات والتقارير", "th": "การวิเคราะห์และรายงาน", "vi": "Phân tích & báo cáo"},
    "Financial Overview": {"ar": "نظرة مالية عامة", "th": "ภาพรวมการเงิน", "vi": "Tổng quan tài chính"},
    "Accounts Receivable": {"ar": "الذمم المدينة", "th": "ลูกหนี้การค้า", "vi": "Phải thu"},
    "Outstanding Invoices": {"ar": "الفواتير المستحقة", "th": "ใบแจ้งหนี้ค้างชำระ", "vi": "Hóa đơn chưa thanh toán"},
    "Profit & Loss": {"ar": "الأرباح والخسائر", "th": "กำไรและขาดทุน", "vi": "Lãi & lỗ"},
    "Inventory Health": {"ar": "صحة المخزون", "th": "สุขภาพสินค้าคงคลัง", "vi": "Tình trạng kho"},
    "Loading users...": {"ar": "جاري تحميل المستخدمين...", "th": "กำลังโหลดผู้ใช้...", "vi": "Đang tải người dùng..."},
    "Loading analytics...": {"ar": "جاري تحميل التحليلات...", "th": "กำลังโหลดการวิเคราะห์...", "vi": "Đang tải phân tích..."},
    "Error loading analytics data": {"ar": "خطأ في تحميل بيانات التحليلات", "th": "เกิดข้อผิดพลาดในการโหลดข้อมูลวิเคราะห์", "vi": "Lỗi tải dữ liệu phân tích"},
    "Revenue Trends": {"ar": "اتجاهات الإيرادات", "th": "แนวโน้มรายได้", "vi": "Xu hướng doanh thu"},
    "Top Products": {"ar": "أفضل المنتجات", "th": "สินค้ายอดนิยม", "vi": "Sản phẩm hàng đầu"},
    "Payment Breakdown": {"ar": "تفصيل المدفوعات", "th": "รายละเอียดการชำระเงิน", "vi": "Phân tích thanh toán"},
    "Due Date": {"ar": "تاريخ الاستحقاق", "th": "วันครบกำหนด", "vi": "Ngày đến hạn"},
    "Amount": {"ar": "المبلغ", "th": "จำนวนเงิน", "vi": "Số tiền"},
    "Qty": {"ar": "الكمية", "th": "จำนวน", "vi": "SL"},
    "Stock": {"ar": "المخزون", "th": "สต็อก", "vi": "Tồn kho"},
    "Profit": {"ar": "الربح", "th": "กำไร", "vi": "Lợi nhuận"},
    "Month": {"ar": "الشهر", "th": "เดือน", "vi": "Tháng"},
    "Recency": {"ar": "الحداثة", "th": "ความล่าสุด", "vi": "Gần đây"},
    "Frequency": {"ar": "التكرار", "th": "ความถี่", "vi": "Tần suất"},
    "Monetary": {"ar": "القيمة النقدية", "th": "มูลค่า", "vi": "Giá trị"},
    "Segment": {"ar": "الشريحة", "th": "กลุ่ม", "vi": "Phân khúc"},
    "Rate": {"ar": "المعدل", "th": "อัตรา", "vi": "Tỷ lệ"},
    "Quantity": {"ar": "الكمية", "th": "ปริมาณ", "vi": "Số lượng"},
    "Email": {"ar": "البريد الإلكتروني", "th": "อีเมล", "vi": "Email"},
    "ID": {"ar": "المعرف", "th": "รหัส", "vi": "ID"},
    "Reorder Qty": {"ar": "كمية إعادة الطلب", "th": "ปริมาณสั่งซื้อใหม่", "vi": "SL đặt lại"},
    "Manage product catalog with complete CRUD operations.": {
        "ar": "إدارة كatalog المنتجات مع عمليات CRUD كاملة.",
        "th": "จัดการแคตตาล็อกสินค้าพร้อม CRUD ครบถ้วน",
        "vi": "Quản lý danh mục sản phẩm với đầy đủ thao tác CRUD.",
    },
    "Add Product": {"ar": "إضافة منتج", "th": "เพิ่มสินค้า", "vi": "Thêm sản phẩm"},
    "Category": {"ar": "الفئة", "th": "หมวดหมู่", "vi": "Danh mục"},
    "Total Products": {"ar": "إجمالي المنتجات", "th": "สินค้าทั้งหมด", "vi": "Tổng sản phẩm"},
    "Total Value": {"ar": "القيمة الإجمالية", "th": "มูลค่ารวม", "vi": "Tổng giá trị"},
    "Loading inventory...": {"ar": "جاري تحميل المخزون...", "th": "กำลังโหลดสินค้าคงคลัง...", "vi": "Đang tải kho hàng..."},
    "Export": {"ar": "تصدير", "th": "ส่งออก", "vi": "Xuất"},
    "History": {"ar": "السجل", "th": "ประวัติ", "vi": "Lịch sử"},
    "Reason": {"ar": "السبب", "th": "เหตุผล", "vi": "Lý do"},
    "Subtract": {"ar": "طرح", "th": "ลบ", "vi": "Trừ"},
    "Coupons & Gift Cards": {"ar": "القسائم وبطاقات الهدايا", "th": "คูปองและบัตรของขวัญ", "vi": "Phiếu giảm giá & thẻ quà"},
    "Shipping Rates": {"ar": "أسعار الشحن", "th": "อัตราค่าจัดส่ง", "vi": "Biểu phí vận chuyển"},
    "Tax Rates": {"ar": "معدلات الضريبة", "th": "อัตราภาษี", "vi": "Thuế suất"},
}

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

ZH_MAP_PATH = ROOT.parent / "scripts" / "zh-locale-map.json"


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


def is_priority(key: str) -> bool:
    return any(key.startswith(p) or f".{p}" in key for p in PRIORITY_PREFIXES) or key.startswith("enum.")


def is_english(text: str) -> bool:
    if not text or len(text) <= 2:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&]+", text):
        return False
    return len(re.findall(r"[A-Za-z]", text)) / max(len(text), 1) > 0.55


def load_zh_map() -> dict[str, dict[str, str]]:
    if ZH_MAP_PATH.exists():
        return json.loads(ZH_MAP_PATH.read_text(encoding="utf-8"))
    return {}


def main() -> None:
    en_admin = json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8"))
    en_common = json.loads((ROOT / "en" / "common.json").read_text(encoding="utf-8"))
    zh_admin = json.loads((ROOT / "zh" / "admin.json").read_text(encoding="utf-8"))
    zh_common = json.loads((ROOT / "zh" / "common.json").read_text(encoding="utf-8"))

    en_flat = flatten(en_admin) | flatten(en_common)
    zh_flat = flatten(zh_admin) | flatten(zh_common)
    zh_map = load_zh_map()

    cache_path = ROOT.parent / "scripts" / "i18n-translation-cache.json"
    cache = json.loads(cache_path.read_text(encoding="utf-8")) if cache_path.exists() else {}

    for locale in ("ar", "th", "vi"):
        loc_cache = cache.get(locale, {})
        fname_map = {"admin.json": en_admin, "common.json": en_common}
        zh_map_data = {"admin.json": zh_admin, "common.json": zh_common}

        for fname, tree in fname_map.items():
            en_f = flatten(tree)
            zh_f = flatten(zh_map_data[fname])
            current = json.loads((ROOT / locale / fname).read_text(encoding="utf-8"))
            cur_f = flatten(current)
            out_f = dict(cur_f)

            for key, en_val in en_f.items():
                zh_val = zh_f.get(key, "")
                # Priority 1: term dict
                if en_val in TERM:
                    out_f[key] = TERM[en_val][locale]
                    continue
                # Priority 2: zh map file
                if zh_val and zh_val in zh_map.get(locale, {}):
                    out_f[key] = zh_map[locale][zh_val]
                    continue
                # Priority 3: cache if not english
                cached = loc_cache.get(en_val)
                if cached and cached != en_val and "__KEEP" not in cached:
                    if not is_english(cached) or not is_english(en_val):
                        out_f[key] = cached
                        continue
                # Priority 4: keep current if not english
                cur = cur_f.get(key, en_val)
                if cur != en_val and not is_english(cur):
                    out_f[key] = cur

            (ROOT / locale / fname).write_text(
                json.dumps(unflatten(out_f), ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            print(f"Patched {locale}/{fname}")


if __name__ == "__main__":
    main()
