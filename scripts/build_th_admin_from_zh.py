#!/usr/bin/env python3
"""Build th/admin.json from en keys using zh semantic source -> th translation."""

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
CACHE_PATH = Path(__file__).resolve().parent / "i18n-zh-th-cache.json"
WORKERS = 8

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

MANUAL_BY_KEY: dict[str, str] = {
    "admin.langLabel": "TH",
    "admin.langLabelShort": "TH",
    "admin.langTitle": "ไทย",
    "admin.brand": "CandyPro Admin",
    "admin.nav.dashboard": "แดชบอร์ด",
    "admin.nav.users": "ผู้ใช้",
    "admin.nav.inquiries": "คำขอสอบถาม",
    "admin.nav.orders": "คำสั่งซื้อ",
    "admin.nav.returns": "การคืนสินค้า",
    "admin.nav.products": "สินค้า",
    "admin.nav.categories": "หมวดหมู่",
    "admin.nav.companies": "บริษัท",
    "admin.nav.trades": "ธุรกรรมการค้า",
    "admin.nav.pricing": "ราคา",
    "admin.nav.oemProjects": "โครงการ OEM",
    "admin.nav.content": "เนื้อหา",
    "admin.nav.inventory": "สินค้าคงคลัง",
    "admin.nav.xlsx": "เครื่องมือ XLSX",
    "admin.nav.shipments": "การจัดส่ง",
    "admin.nav.invoices": "ใบแจ้งหนี้",
    "admin.nav.management": "การจัดการ",
    "admin.nav.superadmin": "ผู้ดูแลระบบสูงสุด",
    "admin.nav.analytics": "การวิเคราะห์",
    "admin.nav.financial": "การเงิน",
    "admin.nav.certifications": "ใบรับรอง",
    "admin.nav.staff": "พนักงาน",
    "admin.nav.auditLog": "บันทึกการตรวจสอบ",
    "admin.nav.settings": "การตั้งค่า",
    "admin.nav.translations": "การแปลภาษา",
    "admin.nav.ai": "แอป AI",
    "admin.nav.organizations": "องค์กร",
    "admin.nav.channels": "ช่องทาง",
    "admin.nav.webhooks": "Webhooks",
    "admin.nav.hooks": "Hooks และ Events",
    "admin.nav.coupons": "คูปอง",
    "admin.nav.shipping_rates": "ค่าจัดส่งและภาษี",
    "admin.nav.tax_rates": "อัตราภาษี",
    "admin.logout": "ออกจากระบบ",
    "admin.theme.dark": "โหมดมืด",
    "admin.theme.light": "โหมดสว่าง",
    "admin.a11y.openNav": "เปิดเมนูแถบด้านข้าง",
    "admin.a11y.closeNav": "ปิดเมนูแถบด้านข้าง",
    "admin.a11y.toggleSidebar": "ขยายหรือยุบแถบด้านข้าง",
    "admin.dashboard.title": "แดชบอร์ด",
    "admin.dashboard.subtitle": "ภาพรวมแพลตฟอร์มและตัวชี้วัดสำคัญในหน้าเดียว",
    "admin.dashboard.loading": "กำลังโหลดสถิติแดชบอร์ด...",
    "admin.dashboard.error": "เกิดข้อผิดพลาดในการโหลดแดชบอร์ด",
    "admin.dashboard.total_users": "ผู้ใช้ทั้งหมด",
    "admin.dashboard.total_orders": "คำสั่งซื้อทั้งหมด",
    "admin.dashboard.total_inquiries": "คำขอสอบถามทั้งหมด",
    "admin.dashboard.total_revenue": "รายได้รวม",
    "admin.dashboard.this_week": "สัปดาห์นี้",
    "admin.dashboard.new_users_this_week": "ใหม่ {count} รายการสัปดาห์นี้",
    "admin.dashboard.pending": "รอดำเนินการ",
    "admin.dashboard.recent_activity": "กิจกรรมล่าสุด",
    "admin.dashboard.col_activity": "กิจกรรม",
    "admin.dashboard.col_time": "เวลา",
    "admin.dashboard.no_activity": "ไม่มีกิจกรรมล่าสุด",
    "admin.dashboard.revenue_this_month": "รายได้เดือนนี้",
    "admin.dashboard.active_customers": "ลูกค้าที่ใช้งานอยู่",
    "admin.dashboard.avg_order_value": "มูลค่าคำสั่งซื้อเฉลี่ย",
    "admin.dashboard.conversion_rate": "อัตราการแปลง",
    "admin.dashboard.revenue_chart": "รายได้ (30 วัน)",
    "admin.dashboard.revenue": "รายได้",
    "admin.dashboard.date": "วันที่",
    "admin.dashboard.activity_order": "ลูกค้า {user} สั่งซื้อใหม่ {order_no}",
    "admin.dashboard.activity_inquiry": "คำขอสอบถาม B2B ใหม่จาก {company}",
    "admin.dashboard.activity_user": "ผู้ใช้ใหม่ลงทะเบียน: {user}",
    "admin.dashboard.unknown_customer": "ลูกค้าไม่ทราบชื่อ",
    "admin.dashboard.unknown_company": "บริษัทไม่ทราบชื่อ",
    "admin.dashboard.unknown_user": "ผู้ใช้ไม่ทราบชื่อ",
    "admin.dashboard.time_just_now": "เมื่อสักครู่",
    "admin.dashboard.time_minutes_ago": "{count} นาทีที่แล้ว",
    "admin.dashboard.time_hours_ago": "{count} ชั่วโมงที่แล้ว",
    "admin.dashboard.time_days_ago": "{count} วันที่แล้ว",
    "admin.dashboard.time_unknown": "ไม่ทราบ",
    "admin.status_options.draft": "ฉบับร่าง",
    "admin.status_options.sent": "ส่งแล้ว",
    "admin.status_options.paid": "ชำระแล้ว",
    "admin.status_options.overdue": "เกินกำหนด",
    "admin.status_options.cancelled": "ยกเลิกแล้ว",
    "admin.status_options.active": "ใช้งาน",
    "admin.status_options.inactive": "ไม่ใช้งาน",
    "admin.shippingRates.title": "อัตราค่าจัดส่ง",
    "admin.shippingRates.description": "ตารางอัตราค่าขนส่งตามปลายทาง จำนวนวันโดยประมาณเป็นค่าอ้างอิงสำหรับการเสนอราคาเท่านั้น — ไม่ใช่เวลาจัดส่งที่รับประกัน",
    "admin.shippingRates.create": "เพิ่มอัตรา",
    "admin.shippingRates.edit": "แก้ไข",
    "admin.shippingRates.delete": "ลบ",
    "admin.shippingRates.save": "บันทึก",
    "admin.shippingRates.cancel": "ยกเลิก",
    "admin.shippingRates.no_data": "ไม่มีอัตราค่าจัดส่ง",
    "admin.shippingRates.destination": "ปลายทาง",
    "admin.shippingRates.carrier": "ผู้ให้บริการขนส่ง",
    "admin.shippingRates.min_weight": "น้ำหนักขั้นต่ำ (kg)",
    "admin.shippingRates.max_weight": "น้ำหนักสูงสุด (kg)",
    "admin.shippingRates.base_cost": "ค่าใช้จ่ายพื้นฐาน",
    "admin.shippingRates.cost_per_kg": "ค่าใช้จ่ายต่อ kg",
    "admin.shippingRates.currency": "สกุลเงิน",
    "admin.shippingRates.estimated_days": "วันโดยประมาณ",
    "admin.shippingRates.estimated_days_hint": "ค่าประมาณเวลาจัดส่งสำหรับการเสนอราคาภายใน อาจเปลี่ยนแปลงได้ ยืนยันเมื่อสั่งซื้อ",
    "admin.shippingRates.active": "ใช้งาน",
    "admin.shippingRates.col_destination": "ปลายทาง",
    "admin.shippingRates.col_carrier": "ผู้ให้บริการ",
    "admin.shippingRates.col_base_cost": "ค่าใช้จ่ายพื้นฐาน",
    "admin.shippingRates.col_days": "วัน",
    "admin.shippingRates.col_status": "สถานะ",
    "admin.shippingRates.actions": "การดำเนินการ",
    "admin.taxRates.title": "อัตราภาษี",
    "admin.taxRates.description": "อัตราภาษีตามประเทศ/ภูมิภาคปลายทางที่ใช้เมื่อชำระเงินในตะกร้า",
    "admin.taxRates.create": "เพิ่มอัตราภาษี",
    "admin.taxRates.edit": "แก้ไข",
    "admin.taxRates.delete": "ลบ",
    "admin.taxRates.save": "บันทึก",
    "admin.taxRates.cancel": "ยกเลิก",
    "admin.taxRates.no_data": "ยังไม่ได้ตั้งค่าอัตราภาษี",
    "admin.taxRates.country": "ประเทศ (ISO)",
    "admin.taxRates.region": "ภูมิภาค / รัฐ",
    "admin.taxRates.rate": "อัตรา",
    "admin.taxRates.rate_hint": "ทศนิยม เช่น 0.08875 = 8.875%",
    "admin.taxRates.name": "ชื่อ",
    "admin.taxRates.active": "ใช้งาน",
    "admin.taxRates.col_country": "ประเทศ",
    "admin.taxRates.col_region": "ภูมิภาค",
    "admin.taxRates.col_name": "ชื่อ",
    "admin.taxRates.col_rate": "อัตรา",
    "admin.taxRates.col_status": "สถานะ",
    "admin.taxRates.actions": "การดำเนินการ",
    "admin.coupons.title": "คูปองและบัตรของขวัญ",
    "admin.coupons.description": "รหัสโปรโมชันและเครดิตล่วงหน้า",
    "admin.coupons.create": "สร้างคูปอง",
    "admin.coupons.create_giftcard": "สร้างบัตรของขวัญ",
    "admin.coupons.delete": "ลบ",
    "admin.coupons.save": "บันทึก",
    "admin.coupons.cancel": "ยกเลิก",
    "admin.coupons.tab_coupons": "คูปอง",
    "admin.coupons.tab_giftcards": "บัตรของขวัญ",
    "admin.coupons.no_data": "ไม่มีคูปอง",
    "admin.coupons.no_giftcards": "ไม่มีบัตรของขวัญ",
    "admin.coupons.code": "รหัส",
    "admin.coupons.type": "ประเภท",
    "admin.coupons.value": "มูลค่า",
    "admin.coupons.min_order": "ยอดสั่งซื้อขั้นต่ำ",
    "admin.coupons.starts_at": "เริ่มต้น",
    "admin.coupons.expires_at": "หมดอายุ",
    "admin.coupons.balance": "ยอดคงเหลือ",
    "admin.coupons.currency": "สกุลเงิน",
    "admin.coupons.type_percentage": "เปอร์เซ็นต์",
    "admin.coupons.type_fixed": "จำนวนเงินคงที่",
    "admin.coupons.col_code": "รหัส",
    "admin.coupons.col_type": "ประเภท",
    "admin.coupons.col_value": "มูลค่า",
    "admin.coupons.col_used": "ใช้แล้ว",
    "admin.coupons.col_status": "สถานะ",
    "admin.coupons.col_initial": "ยอดเริ่มต้น",
    "admin.coupons.col_balance": "ยอดคงเหลือ",
    "admin.coupons.col_currency": "สกุลเงิน",
    "admin.coupons.actions": "การดำเนินการ",
    "admin.coupons.showing": "ทั้งหมด {total} รายการ",
    "admin.coupons.previous": "ก่อนหน้า",
    "admin.coupons.next": "ถัดไป",
    "admin.categories.title": "หมวดหมู่",
    "admin.categories.description": "จัดการหมวดหมู่สินค้าพร้อมชื่อ นามแฝง และคำอธิบายหลายภาษา",
    "admin.categories.add_category": "เพิ่มหมวดหมู่",
    "admin.categories.edit_category": "แก้ไขหมวดหมู่",
    "admin.categories.create_category": "สร้างหมวดหมู่",
    "admin.products.title": "สินค้า",
    "admin.products.description": "จัดการแคตตalog สินค้าพร้อมการดำเนินการ CRUD ครบถ้วน",
    "admin.products.add_product": "เพิ่มสินค้า",
    "admin.products.moq": "MOQ",
    "admin.products.col_moq": "MOQ",
    "admin.inventory.title": "สินค้าคงคลัง",
    "admin.inventory.description": "ติดตามและจัดการระดับสต็อก มูลค่า และการปรับปรุงสินค้า",
    "admin.inventory.col_sku": "SKU",
    "admin.inventory.col_moq": "MOQ",
    "admin.inventory.import_xlsx": "นำเข้า XLSX",
}

_cache_lock = threading.Lock()


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


def has_thai(text: str) -> bool:
    return bool(re.search(r"[\u0E00-\u0E7F]", text))


def has_cjk(text: str) -> bool:
    return bool(re.search(r"[\u4e00-\u9fff]", text))


def translate_zh_to_th(text: str) -> str:
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source="zh-CN", target="th-TH")
    for attempt in range(4):
        try:
            result = translator.translate(protected)
            out = restore(result or text, replacements)
            if has_thai(out) and not has_cjk(out):
                return out
        except Exception:
            time.sleep(0.4 * (attempt + 1))
    return text


def load_cache() -> dict[str, str]:
    if CACHE_PATH.exists():
        return json.loads(CACHE_PATH.read_text(encoding="utf-8"))
    return {}


def save_cache(cache: dict[str, str]) -> None:
    CACHE_PATH.write_text(json.dumps(cache, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    cache = load_cache()
    en_flat = flatten(json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8")))
    zh_flat = flatten(json.loads((ROOT / "zh" / "admin.json").read_text(encoding="utf-8")))

    todo = set()
    for key in en_flat:
        if key in MANUAL_BY_KEY:
            continue
        zh_val = zh_flat.get(key, en_flat[key])
        if zh_val not in cache or not has_thai(cache[zh_val]) or has_cjk(cache.get(zh_val, "")):
            todo.add(zh_val)

    print(f"zh->th {len(todo)} strings", flush=True)
    done = 0
    total = len(todo)
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futures = {pool.submit(translate_zh_to_th, text): text for text in sorted(todo, key=len)}
        for future in as_completed(futures):
            text = futures[future]
            try:
                cache[text] = future.result()
            except Exception:
                cache[text] = text
            done += 1
            if done % 50 == 0 or done == total:
                print(f"  {done}/{total}", flush=True)
                save_cache(cache)

    out_flat: dict[str, str] = {}
    for key, en_val in en_flat.items():
        if key in MANUAL_BY_KEY:
            out_flat[key] = MANUAL_BY_KEY[key]
        else:
            zh_val = zh_flat.get(key, en_val)
            out_flat[key] = cache.get(zh_val, zh_val)

    (ROOT / "th" / "admin.json").write_text(
        json.dumps(unflatten(out_flat), ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    save_cache(cache)
    print("Wrote th/admin.json", flush=True)


if __name__ == "__main__":
    main()
