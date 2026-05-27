#!/usr/bin/env python3
"""Build th/admin.json from en using manual overrides, phrase map, and en->th cache."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache.json"

# --- manual key overrides (priority sections) ---
MANUAL_BY_KEY: dict[str, str] = {}
exec(  # noqa: S102
    (Path(__file__).parent / "build_th_admin_from_zh.py")
    .read_text(encoding="utf-8")
    .split("_cache_lock = threading.Lock()")[0]
    .split("MANUAL_BY_KEY: dict[str, str] = ", 1)[1]
)
# exec above loads MANUAL_BY_KEY from sibling script header

COMMON_EN: dict[str, str] = {
    "Logout": "ออกจากระบบ",
    "Dark mode": "โหมดมืด",
    "Light mode": "โหมดสว่าง",
    "Open sidebar menu": "เปิดเมนูแถบด้านข้าง",
    "Close sidebar menu": "ปิดเมนูแถบด้านข้าง",
    "Expand or collapse sidebar": "ขยายหรือยุบแถบด้านข้าง",
    "Dashboard": "แดชบอร์ด",
    "Users": "ผู้ใช้",
    "Inquiries": "คำขอสอบถาม",
    "Orders": "คำสั่งซื้อ",
    "Returns": "การคืนสินค้า",
    "Products": "สินค้า",
    "Categories": "หมวดหมู่",
    "Companies": "บริษัท",
    "Trades": "ธุรกรรมการค้า",
    "Pricing": "ราคา",
    "Content": "เนื้อหา",
    "Inventory": "สินค้าคงคลัง",
    "Shipments": "การจัดส่ง",
    "Invoices": "ใบแจ้งหนี้",
    "Management": "การจัดการ",
    "Analytics": "การวิเคราะห์",
    "Financial": "การเงิน",
    "Certifications": "ใบรับรอง",
    "Staff": "พนักงาน",
    "Settings": "การตั้งค่า",
    "Translations": "การแปลภาษา",
    "Organizations": "องค์กร",
    "Channels": "ช่องทาง",
    "Edit": "แก้ไข",
    "Delete": "ลบ",
    "Cancel": "ยกเลิก",
    "Save": "บันทึก",
    "Create": "สร้าง",
    "Update": "อัปเดต",
    "Loading...": "กำลังโหลด...",
    "Loading…": "กำลังโหลด…",
    "Actions": "การดำเนินการ",
    "Previous": "ก่อนหน้า",
    "Next": "ถัดไป",
    "View": "ดู",
    "Search": "ค้นหา",
    "Status": "สถานะ",
    "Name": "ชื่อ",
    "Email": "อีเมล",
    "Phone": "โทรศัพท์",
    "Company": "บริษัท",
    "Customer": "ลูกค้า",
    "Amount": "จำนวนเงิน",
    "Date": "วันที่",
    "Notes": "หมายเหตุ",
    "Yes": "ใช่",
    "No": "ไม่",
    "All": "ทั้งหมด",
    "Active": "ใช้งาน",
    "Inactive": "ไม่ใช้งาน",
    "Pending": "รอดำเนินการ",
    "Confirmed": "ยืนยันแล้ว",
    "Paid": "ชำระแล้ว",
    "Draft": "ฉบับร่าง",
    "Sent": "ส่งแล้ว",
    "Overdue": "เกินกำหนด",
    "Cancelled": "ยกเลิกแล้ว",
    "Unknown": "ไม่ทราบ",
    "Total": "รวม",
    "Export": "ส่งออก",
    "Import": "นำเข้า",
    "Filter": "กรอง",
    "Confirm": "ยืนยัน",
    "Reject": "ปฏิเสธ",
    "Approve": "อนุมัติ",
    "Refund": "คืนเงิน",
    "Description": "คำอธิบาย",
    "Title": "หัวข้อ",
    "Type": "ประเภท",
    "Currency": "สกุลเงิน",
    "Quantity": "ปริมาณ",
    "Product": "สินค้า",
    "Category": "หมวดหมู่",
    "Order": "คำสั่งซื้อ",
    "Invoice": "ใบแจ้งหนี้",
    "Payment": "การชำระเงิน",
    "Shipping": "การจัดส่ง",
    "Tax": "ภาษี",
    "Subtotal": "ยอดรวมย่อย",
    "Revenue": "รายได้",
    "Cost": "ต้นทุน",
    "Profit": "กำไร",
    "Month": "เดือน",
    "Loading report...": "กำลังโหลดรายงาน...",
    "Failed to load report": "โหลดรายงานไม่สำเร็จ",
    "No data available": "ไม่มีข้อมูล",
    "Business Reports": "รายงานธุรกิจ",
    "Sales Velocity": "ความเร็วการขาย",
    "RFM Analysis": "การวิเคราะห์ RFM",
    "Customer Churn": "การสูญเสียลูกค้า",
    "Inventory Health": "สุขภาพสินค้าคงคลัง",
    "Profit & Loss": "กำไรและขาดทุน",
    "Replenishment": "การเติมสต็อก",
    "Inquiry Conversion": "การแปลงคำขอสอบถาม",
    "Analytics & Reports": "การวิเคราะห์และรายงาน",
    "Business intelligence dashboards with revenue, order, and conversion metrics.": "แดชบอร์ด Business Intelligence พร้อมตัวชี้วัดรายได้ คำสั่งซื้อ และการแปลง",
    "Loading analytics...": "กำลังโหลดการวิเคราะห์...",
    "Error loading analytics data": "เกิดข้อผิดพลาดในการโหลดข้อมูลการวิเคราะห์",
    "No data available for the selected period.": "ไม่มีข้อมูลสำหรับช่วงเวลาที่เลือก",
    "Revenue Trends": "แนวโน้มรายได้",
    "Order Trends": "แนวโน้มคำสั่งซื้อ",
    "Top Products": "สินค้ายอดนิยม",
    "3 Months": "3 เดือน",
    "6 Months": "6 เดือน",
    "12 Months": "12 เดือน",
    "Financial Overview": "ภาพรวมการเงิน",
    "Accounts receivable, payment tracking, and revenue analytics.": "ลูกหนี้การค้า การติดตามการชำระเงิน และการวิเคราะห์รายได้",
    "Loading financial data...": "กำลังโหลดข้อมูลการเงิน...",
    "Error loading financial data": "เกิดข้อผิดพลาดในการโหลดข้อมูลการเงิน",
    "Accounts Receivable": "ลูกหนี้การค้า",
    "Overdue Amount": "ยอดเกินกำหนด",
    "Outstanding Invoices": "ใบแจ้งหนี้ค้างชำระ",
    "Payment Breakdown": "รายละเอียดการชำระเงิน",
    "Invoice No": "เลขที่ใบแจ้งหนี้",
    "Due Date": "วันครบกำหนด",
    "Due Soon": "ใกล้ครบกำหนด",
    "No outstanding invoices.": "ไม่มีใบแจ้งหนี้ค้างชำระ",
    "No financial data available.": "ไม่มีข้อมูลการเงิน",
    "New User": "ผู้ใช้ใหม่",
    "Loading users...": "กำลังโหลดผู้ใช้...",
    "No users found.": "ไม่พบผู้ใช้",
    "Create User": "สร้างผู้ใช้",
    "First Name": "ชื่อ",
    "Last Name": "นามสกุล",
    "Password": "รหัสผ่าน",
    "Role": "บทบาท",
    "Admin": "ผู้ดูแลระบบ",
    "Superadmin": "ผู้ดูแลระบบสูงสุด",
    "Suspended": "ระงับ",
    "Creating...": "กำลังสร้าง...",
    "Delete this user permanently?": "ลบผู้ใช้นี้ถาวรหรือไม่?",
    "Showing {from} to {to} of {total} results": "แสดง {from} ถึง {to} จาก {total} รายการ",
    "Showing {from} to {to} of {total}": "แสดง {from} ถึง {to} จาก {total}",
    "Manage inquiry records with complete CRUD and conversion actions.": "จัดการบันทึกคำขอสอบถามพร้อม CRUD และการแปลงครบถ้วน",
    "New Inquiry": "คำขอสอบถามใหม่",
    "Loading inquiries...": "กำลังโหลดคำขอสอบถาม...",
    "No inquiries found.": "ไม่พบคำขอสอบถาม",
    "Company & Contact": "บริษัทและผู้ติดต่อ",
    "Edit Inquiry": "แก้ไขคำขอสอบถาม",
    "Create Inquiry": "สร้างคำขอสอบถาม",
    "Contact Person": "ผู้ติดต่อ",
    "Target Country": "ประเทศเป้าหมาย",
    "Estimated Quantity": "ปริมาณโดยประมาณ",
    "Expected Delivery": "วันจัดส่งที่คาดหวัง",
    "Priority": "ความสำคัญ",
    "Message": "ข้อความ",
    "Customer Notes": "หมายเหตุลูกค้า",
    "Assignment": "การมอบหมาย",
    "Assign": "มอบหมาย",
    "Assigning...": "กำลังมอบหมาย...",
    "AI Analysis": "การวิเคราะห์ AI",
    "Analyze": "วิเคราะห์",
    "Analyzing...": "กำลังวิเคราะห์...",
    "Risk": "ความเสี่ยง",
    "Action": "การดำเนินการ",
    "Summary": "สรุป",
    "Quote": "ใบเสนอราคา",
    "Submit Quote": "ส่งใบเสนอราคา",
    "Submitting...": "กำลังส่ง...",
    "Saving...": "กำลังบันทึก...",
    "Updating...": "กำลังอัปเดต...",
    "Manage order lifecycle with complete CRUD operations.": "จัดการวงจรชีวิตคำสั่งซื้อพร้อม CRUD ครบถ้วน",
    "New Order": "คำสั่งซื้อใหม่",
    "Loading orders...": "กำลังโหลดคำสั่งซื้อ...",
    "No orders found.": "ไม่พบคำสั่งซื้อ",
    "Edit Order": "แก้ไขคำสั่งซื้อ",
    "Create Order": "สร้างคำสั่งซื้อ",
    "Payment Status": "สถานะการชำระเงิน",
    "Tracking Number": "หมายเลขติดตาม",
    "Order Items": "รายการสั่งซื้อ",
    "Add Item": "เพิ่มรายการ",
    "Unit Price": "ราคาต่อหน่วย",
    "Shipping Address": "ที่อยู่จัดส่ง",
    "Street": "ถนน",
    "City": "เมือง",
    "State": "รัฐ",
    "Zip Code": "รหัสไปรษณีย์",
    "Country": "ประเทศ",
    "Manage product catalog with complete CRUD operations.": "จัดการแคตตalog สินค้าพร้อม CRUD ครบถ้วน",
    "Add Product": "เพิ่มสินค้า",
    "Translate missing locales": "แปลภาษาที่ขาด",
    "Translating…": "กำลังแปล…",
    "Search by product name or SKU...": "ค้นหาตามชื่อสินค้าหรือ SKU...",
    "Published (default)": "เผยแพร่แล้ว (ค่าเริ่มต้น)",
    "All statuses": "ทุกสถานะ",
    "Loading products...": "กำลังโหลดสินค้า...",
    "No products found.": "ไม่พบสินค้า",
    "Edit Product": "แก้ไขสินค้า",
    "Create Product": "สร้างสินค้า",
    "Slug": "Slug",
    "Alias": "นามแฝง",
    "Base Price (USD)": "ราคาฐาน (USD)",
    "Lead Time": "ระยะเวลาจัดส่ง",
    "Stock Quantity": "จำนวนสต็อก",
    "Halal": "Halal",
    "Featured": "แนะนำ",
    "Thumbnail": "ภาพย่อ",
    "Upload Image": "อัปโหลดรูปภาพ",
    "Upload Images": "อัปโหลดรูปภาพ",
    "Ingredients": "ส่วนผสม",
    "Allergens": "สารก่อภูมิแพ้",
    "Shelf Life": "อายุการเก็บ",
    "Storage": "การจัดเก็บ",
    "Track and manage product stock levels, values, and adjustments.": "ติดตามและจัดการระดับสต็อก มูลค่า และการปรับปรุงสินค้า",
    "Adjust Stock": "ปรับสต็อก",
    "Total Products": "สินค้าทั้งหมด",
    "Low Stock": "สต็อกต่ำ",
    "Out of Stock": "สินค้าหมด",
    "In Stock": "มีสินค้า",
    "Total Value": "มูลค่ารวม",
    "Unit Value": "มูลค่าต่อหน่วย",
    "History": "ประวัติ",
    "Reason": "เหตุผล",
    "Restock": "เติมสต็อก",
    "Correction": "แก้ไข",
    "Coupons & Gift Cards": "คูปองและบัตรของขวัญ",
    "Promotional codes and prepaid credits.": "รหัสโปรโมชันและเครดิตล่วงหน้า",
    "Create Coupon": "สร้างคูปอง",
    "Create Gift Card": "สร้างบัตรของขวัญ",
    "Gift Cards": "บัตรของขวัญ",
    "No coupons.": "ไม่มีคูปอง",
    "No gift cards.": "ไม่มีบัตรของขวัญ",
    "Code": "รหัส",
    "Value": "มูลค่า",
    "Percentage": "เปอร์เซ็นต์",
    "Fixed Amount": "จำนวนเงินคงที่",
    "Balance": "ยอดคงเหลือ",
    "Shipping Rates": "อัตราค่าจัดส่ง",
    "Tax Rates": "อัตราภาษี",
    "Add Rate": "เพิ่มอัตรา",
    "Add Tax Rate": "เพิ่มอัตราภาษี",
    "Destination": "ปลายทาง",
    "Carrier": "ผู้ให้บริการขนส่ง",
    "Region / State": "ภูมิภาค / รัฐ",
    "Rate": "อัตรา",
    "Confirm delete?": "ยืนยันการลบ?",
    "Super Admin": "ผู้ดูแลระบบสูงสุด",
    "Audit Log": "บันทึกการตรวจสอบ",
    "AI Apps": "แอป AI",
    "Hooks & Events": "Hooks และ Events",
    "Shipping & Tax": "ค่าจัดส่งและภาษี",
    "XLSX Tools": "เครื่องมือ XLSX",
    "OEM Projects": "โครงการ OEM",
    "Qty": "จำนวน",
    "Velocity": "ความเร็ว",
    "Recency": "ความใหม่",
    "Frequency": "ความถี่",
    "Monetary": "มูลค่า",
    "Segment": "เซ็กเมนต์",
    "Days Inactive": "วันที่ไม่ได้ใช้งาน",
    "Stock": "สต็อก",
    "Reorder Qty": "ปริมาณสั่งซื้อใหม่",
    "Inquiries": "คำขอสอบถาม",
    "Converted": "แปลงแล้ว",
    "Converted": "แปลงแล้ว",
    "Rate": "อัตรา",
    "ID": "ID",
    "g": "g",
    "Just now": "เมื่อสักครู่",
    "Unknown customer": "ลูกค้าไม่ทราบชื่อ",
    "Unknown company": "บริษัทไม่ทราบชื่อ",
    "Unknown user": "ผู้ใช้ไม่ทราบชื่อ",
    "{count} minutes ago": "{count} นาทีที่แล้ว",
    "{count} hours ago": "{count} ชั่วโมงที่แล้ว",
    "{count} days ago": "{count} วันที่แล้ว",
    "New order {order_no} placed by {user}": "ลูกค้า {user} สั่งซื้อใหม่ {order_no}",
    "New B2B inquiry from {company}": "คำขอสอบถาม B2B ใหม่จาก {company}",
    "New user registered: {user}": "ผู้ใช้ใหม่ลงทะเบียน: {user}",
    "Platform overview and key metrics at a glance.": "ภาพรวมแพลตฟอร์มและตัวชี้วัดสำคัญในหน้าเดียว",
    "Loading dashboard statistics...": "กำลังโหลดสถิติแดชบอร์ด...",
    "Error loading dashboard": "เกิดข้อผิดพลาดในการโหลดแดชบอร์ด",
    "Total Users": "ผู้ใช้ทั้งหมด",
    "Total Orders": "คำสั่งซื้อทั้งหมด",
    "Total Inquiries": "คำขอสอบถามทั้งหมด",
    "Total Revenue": "รายได้รวม",
    "this week": "สัปดาห์นี้",
    "{count} new this week": "ใหม่ {count} รายการสัปดาห์นี้",
    "pending": "รอดำเนินการ",
    "Recent Activity": "กิจกรรมล่าสุด",
    "Activity": "กิจกรรม",
    "Time": "เวลา",
    "No recent activity": "ไม่มีกิจกรรมล่าสุด",
    "Revenue This Month": "รายได้เดือนนี้",
    "Active Customers": "ลูกค้าที่ใช้งานอยู่",
    "Avg Order Value": "มูลค่าคำสั่งซื้อเฉลี่ย",
    "Conversion Rate": "อัตราการแปลง",
    "Revenue (30 Days)": "รายได้ (30 วัน)",
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


def has_thai(text: str) -> bool:
    return bool(re.search(r"[\u0E00-\u0E7F]", text))


def has_cjk(text: str) -> bool:
    return bool(re.search(r"[\u4e00-\u9fff]", text))


def is_glossary_only(text: str) -> bool:
    return bool(re.fullmatch(r"[A-Za-z0-9/:.+\-\s#×→&]+", text))


def translate_value(en_val: str, cache: dict[str, str]) -> str:
    if en_val in MANUAL_BY_KEY.values():
        return en_val
    if en_val in COMMON_EN:
        return COMMON_EN[en_val]
    cached = cache.get(en_val, "")
    if cached and has_thai(cached) and not has_cjk(cached):
        return cached
    return en_val


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    cache = json.loads(CACHE_PATH.read_text(encoding="utf-8")).get("th", {})
    en_flat = flatten(json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8")))

    out_flat: dict[str, str] = {}
    for key, en_val in en_flat.items():
        if key in MANUAL_BY_KEY:
            out_flat[key] = MANUAL_BY_KEY[key]
        else:
            out_flat[key] = translate_value(en_val, cache)

    (ROOT / "th" / "admin.json").write_text(
        json.dumps(unflatten(out_flat), ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )

    untrans = sum(
        1
        for k, v in out_flat.items()
        if v == en_flat[k] and not is_glossary_only(v) and len(v) > 3
    )
    cjk = sum(1 for v in out_flat.values() if has_cjk(v))
    thai = sum(1 for v in out_flat.values() if has_thai(v))
    print(f"Wrote th/admin.json: thai={thai} cjk={cjk} untranslated={untrans}")


if __name__ == "__main__":
    main()
