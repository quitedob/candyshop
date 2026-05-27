#!/usr/bin/env python3
"""Build complete th i18n from en structure using vi semantic reference + manual overrides."""

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
CACHE_PATH = Path(__file__).resolve().parent / "i18n-vi-th-cache.json"
FILES = ("admin.json", "common.json")
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
}

MANUAL_EN: dict[str, str] = {
    "Dark mode": "โหมดมืด",
    "Light mode": "โหมดสว่าง",
    "Close sidebar menu": "ปิดเมนูแถบด้านข้าง",
    "Open sidebar menu": "เปิดเมนูแถบด้านข้าง",
    "Expand or collapse sidebar": "ขยายหรือยุบแถบด้านข้าง",
    "Skip to main content": "ข้ามไปที่เนื้อหาหลัก",
    "This site works best with JavaScript enabled. Use the links below to continue browsing.": "ไซต์นี้ทำงานได้ดีที่สุดเมื่อเปิดใช้งาน JavaScript ใช้ลิงก์ด้านล่างเพื่อเรียกดูต่อ",
    "Loading…": "กำลังโหลด…",
    "Close": "ปิด",
    "WhatsApp Us": "ติดต่อเราทาง WhatsApp",
    "Previous": "ก่อนหน้า",
    "Next": "ถัดไป",
    "Pagination": "การนำทางแบบแบ่งหน้า",
    "Page Not Found": "ไม่พบหน้า",
    "Sorry, the page you are looking for doesn't exist.": "ขออภัย ไม่พบหน้าที่คุณต้องการ",
    "Back to Home": "กลับหน้าแรก",
    "Server Error": "ข้อผิดพลาดของเซิร์ฟเวอร์",
    "Something went wrong on our end. Please try again later.": "เกิดข้อผิดพลาดในระบบของเรา โปรดลองอีกครั้งในภายหลัง",
    "Refresh Page": "รีเฟรชหน้า",
    "Access Denied": "ไม่มีสิทธิ์เข้าถึง",
    "You do not have permission to access this page.": "คุณไม่มีสิทธิ์เข้าถึงหน้านี้",
    "An error occurred": "เกิดข้อผิดพลาด",
    "You do not have permission to access this resource.": "คุณไม่มีสิทธิ์เข้าถึงทรัพยากรนี้",
    "Something went wrong": "เกิดข้อผิดพลาดบางอย่าง",
    "Go to Home": "ไปหน้าแรก",
    "Try Again": "ลองอีกครั้ง",
    "Try again": "ลองอีกครั้ง",
    "Customer": "ลูกค้า",
    "Admin": "ผู้ดูแลระบบ",
    "Super Admin": "ผู้ดูแลระบบสูงสุด",
    "Yes": "ใช่",
    "No": "ไม่",
    "Cancel": "ยกเลิก",
    "Save": "บันทึก",
    "Edit": "แก้ไข",
    "Delete": "ลบ",
    "Loading...": "กำลังโหลด...",
    "Actions": "การดำเนินการ",
    "Clear input": "ล้างข้อมูล",
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


def translate_vi_to_th(text: str) -> str:
    if text in MANUAL_EN:
        return MANUAL_EN[text]
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source="vi-VN", target="th-TH")
    for attempt in range(4):
        try:
            result = translator.translate(protected)
            out = restore(result or text, replacements)
            if has_thai(out):
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

    for filename in FILES:
        en_flat = flatten(json.loads((ROOT / "en" / filename).read_text(encoding="utf-8")))
        vi_flat = flatten(json.loads((ROOT / "vi" / filename).read_text(encoding="utf-8")))

        todo: set[str] = set()
        for key in en_flat:
            if key in MANUAL_BY_KEY:
                continue
            vi_val = vi_flat.get(key, en_flat[key])
            if vi_val not in cache or not has_thai(cache[vi_val]):
                todo.add(vi_val)

        print(f"{filename}: vi->th {len(todo)} strings", flush=True)
        done = 0
        total = len(todo)
        with ThreadPoolExecutor(max_workers=WORKERS) as pool:
            futures = {pool.submit(translate_vi_to_th, text): text for text in sorted(todo, key=len)}
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
            elif en_val in MANUAL_EN:
                out_flat[key] = MANUAL_EN[en_val]
            else:
                vi_val = vi_flat.get(key, en_val)
                out_flat[key] = cache.get(vi_val, vi_val)

        (ROOT / "th" / filename).write_text(
            json.dumps(unflatten(out_flat), ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
        print(f"Wrote th/{filename}", flush=True)

    save_cache(cache)
    print("Done.", flush=True)


if __name__ == "__main__":
    main()
