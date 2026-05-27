#!/usr/bin/env python3
"""Force-regenerate th/admin.json and th/common.json with fresh Thai translations."""

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
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache.json"
LOCALE = "th"
FILES = ("admin.json", "common.json")
SOURCE = "en"
WORKERS = 6

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Temperature",
    "Plan-Execute-Replan", "ODM", "PI", "CI", "SC", "B/L", "PL", "ETA", "ISO",
    "CSV", "YAML", "Stripe", "PayPal", "Chatbot", "Agent", "Cursor", "Markdown",
    "Grade", "FORM", "E-numbers", "display", "layer", "cases", "Mac:",
}

PLACEHOLDER_RE = re.compile(r"(\{[a-zA-Z0-9_]+\})")
TOKEN_RE = re.compile(
    r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|1200×630|aʷ|→|…|\.\.\.|&|[0-9]+(?:\.[0-9]+)?%?|[A-Za-z][A-Za-z0-9+#./\-]{1,})"
)

MANUAL: dict[str, str] = {
    # admin meta
    "admin.langLabel": "TH",
    "admin.langLabelShort": "TH",
    "admin.langTitle": "ไทย",
    "admin.brand": "CandyPro Admin",
    # nav priority
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
    # dashboard
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
    # status_options
    "admin.status_options.draft": "ฉบับร่าง",
    "admin.status_options.sent": "ส่งแล้ว",
    "admin.status_options.paid": "ชำระแล้ว",
    "admin.status_options.overdue": "เกินกำหนด",
    "admin.status_options.cancelled": "ยกเลิกแล้ว",
    "admin.status_options.active": "ใช้งาน",
    "admin.status_options.inactive": "ไม่ใช้งาน",
    # shippingRates
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
    # taxRates
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
    # coupons
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
}

MANUAL_COMMON: dict[str, str] = {
    "skip_to_content": "ข้ามไปที่เนื้อหาหลัก",
    "javascript_required": "ไซต์นี้ทำงานได้ดีที่สุดเมื่อเปิดใช้งาน JavaScript ใช้ลิงก์ด้านล่างเพื่อเรียกดูต่อ",
    "loading": "กำลังโหลด…",
    "close": "ปิด",
    "whatsapp.us": "ติดต่อเราทาง WhatsApp",
    "whatsapp.message": "สวัสดี ฉันสนใจบริการ OEM ขนมของคุณ ช่วยให้ข้อมูลเพิ่มเติมได้ไหม?",
    "whatsapp.product_inquiry": "สวัสดี ฉันสนใจ \"{product}\"{categoryPart} ช่วยให้ข้อมูลเพิ่มเติมได้ไหม?",
    "whatsapp.category_part": " จากหมวดหมู่ {name}",
    "whatsapp.inquiry_intro": "สวัสดี ฉันต้องการสอบถามเกี่ยวกับบริการ OEM ขนมของคุณ\n\n",
    "whatsapp.inquiry_followup": "โปรดให้ข้อมูลเพิ่มเติมเกี่ยวกับราคาและระยะเวลาจัดส่ง",
    "whatsapp.field_name": "ชื่อ",
    "whatsapp.field_company": "บริษัท",
    "whatsapp.field_email": "อีเมล",
    "whatsapp.field_country": "ประเทศ",
    "whatsapp.field_product": "สินค้า",
    "whatsapp.field_quantity": "ปริมาณ",
    "pagination.prev": "ก่อนหน้า",
    "pagination.next": "ถัดไป",
    "pagination.page": "หน้า {current} จาก {total}",
    "pagination.showing": "แสดง {from} ถึง {to} จาก {total} สินค้า",
    "pagination.nav_label": "การนำทางแบบแบ่งหน้า",
    "errors.404.title": "ไม่พบหน้า",
    "errors.404.description": "ขออภัย ไม่พบหน้าที่คุณต้องการ",
    "errors.404.back_home": "กลับหน้าแรก",
    "errors.500.title": "ข้อผิดพลาดของเซิร์ฟเวอร์",
    "errors.500.description": "เกิดข้อผิดพลาดในระบบของเรา โปรดลองอีกครั้งในภายหลัง",
    "errors.500.refresh": "รีเฟรชหน้า",
    "errors.403.title": "ไม่มีสิทธิ์เข้าถึง",
    "errors.403.description": "คุณไม่มีสิทธิ์เข้าถึงหน้านี้",
    "errors.default": "เกิดข้อผิดพลาด",
    "errors.forbidden_insufficient_role": "คุณไม่มีสิทธิ์เข้าถึงทรัพยากรนี้",
    "errors.boundary": "เกิดข้อผิดพลาดบางอย่าง",
    "errors.goHome": "ไปหน้าแรก",
    "errors.tryAgain": "ลองอีกครั้ง",
    "errors.retry": "ลองอีกครั้ง",
    "roles.customer": "ลูกค้า",
    "roles.admin": "ผู้ดูแลระบบ",
    "roles.superadmin": "ผู้ดูแลระบบสูงสุด",
    "display.na": "N/A",
    "display.em_dash": "—",
    "display.tbd": "TBD",
    "display.zero": "0",
    "common.previous": "ก่อนหน้า",
    "common.next": "ถัดไป",
    "common.cancel": "ยกเลิก",
    "common.save": "บันทึก",
    "common.edit": "แก้ไข",
    "common.delete": "ลบ",
    "common.loading": "กำลังโหลด...",
    "common.actions": "การดำเนินการ",
    "common.close": "ปิด",
    "common.clear": "ล้างข้อมูล",
    "enum.order_status.pending": "รอดำเนินการ",
    "enum.order_status.confirmed": "ยืนยันแล้ว",
    "enum.order_status.production": "กำลังผลิต",
    "enum.order_status.shipped": "จัดส่งแล้ว",
    "enum.order_status.delivered": "ส่งมอบแล้ว",
    "enum.order_status.cancelled": "ยกเลิกแล้ว",
    "enum.order_status.processing": "กำลังดำเนินการ",
    "enum.order_status.pending_confirmation": "รอยืนยัน",
    "enum.order_status.pending_approval": "รออนุมัติ",
    "enum.order_status.partially_shipped": "จัดส่งบางส่วน",
    "enum.order_status.partially_delivered": "ส่งมอบบางส่วน",
    "enum.order_status.partially_returned": "คืนบางส่วน",
    "enum.order_status.returned": "คืนสินค้าแล้ว",
    "enum.order_status.expired": "หมดอายุ",
    "enum.order_status.unknown": "ไม่ทราบ",
    "enum.payment_status.unpaid": "ยังไม่ชำระ",
    "enum.payment_status.paid": "ชำระแล้ว",
    "enum.payment_status.partial": "ชำระบางส่วน",
    "enum.payment_status.refunded": "คืนเงินแล้ว",
    "enum.payment_status.pending": "รอดำเนินการ",
    "enum.payment_status.confirmed": "ยืนยันแล้ว",
    "enum.payment_status.unknown": "ไม่ทราบ",
    "enum.product_status.active": "ใช้งาน",
    "enum.product_status.inactive": "ไม่ใช้งาน",
    "enum.product_status.draft": "ฉบับร่าง",
    "enum.product_status.unknown": "ไม่ทราบ",
    "enum.company_status.pending": "รอดำเนินการ",
    "enum.company_status.verified": "ยืนยันแล้ว",
    "enum.company_status.rejected": "ปฏิเสธแล้ว",
    "enum.company_status.active": "ใช้งาน",
    "enum.company_status.suspended": "ระงับ",
    "enum.company_status.inactive": "ไม่ใช้งาน",
    "enum.company_status.unknown": "ไม่ทราบ",
    "enum.kyb_status.pending": "รอดำเนินการ",
    "enum.kyb_status.approved": "อนุมัติแล้ว",
    "enum.kyb_status.rejected": "ปฏิเสธแล้ว",
    "enum.kyb_status.unknown": "ไม่ทราบ",
    "enum.inquiry_status.pending": "รอดำเนินการ",
    "enum.inquiry_status.contacted": "ติดต่อแล้ว",
    "enum.inquiry_status.quoted": "เสนอราคาแล้ว",
    "enum.inquiry_status.negotiating": "กำลังเจรจา",
    "enum.inquiry_status.won": "ชนะ",
    "enum.inquiry_status.lost": "แพ้",
    "enum.inquiry_status.open": "เปิด",
    "enum.inquiry_status.closed": "ปิด",
    "enum.inquiry_status.converted": "แปลงแล้ว",
    "enum.inquiry_status.unknown": "ไม่ทราบ",
    "enum.inquiry_priority.normal": "ปกติ",
    "enum.inquiry_priority.high": "สูง",
    "enum.inquiry_priority.low": "ต่ำ",
    "enum.inquiry_priority.urgent": "เร่งด่วน",
    "enum.inquiry_priority.unknown": "ไม่ทราบ",
    "enum.trade_status.draft": "ฉบับร่าง",
    "enum.trade_status.pending": "รอดำเนินการ",
    "enum.trade_status.confirmed": "ยืนยันแล้ว",
    "enum.trade_status.in_progress": "กำลังดำเนินการ",
    "enum.trade_status.paid": "ชำระแล้ว",
    "enum.trade_status.shipped": "จัดส่งแล้ว",
    "enum.trade_status.active": "ใช้งาน",
    "enum.trade_status.completed": "เสร็จสมบูรณ์",
    "enum.trade_status.closed": "ปิด",
    "enum.trade_status.cancelled": "ยกเลิกแล้ว",
    "enum.trade_status.unknown": "ไม่ทราบ",
    "enum.shipment_status.pending": "รอดำเนินการ",
    "enum.shipment_status.dispatched": "ส่งออกแล้ว",
    "enum.shipment_status.in_transit": "กำลังขนส่ง",
    "enum.shipment_status.delivered": "ส่งมอบแล้ว",
    "enum.shipment_status.exception": "มีปัญหา",
    "enum.shipment_status.unknown": "ไม่ทราบ",
    "enum.shipment_event_type.dispatched": "ส่งออกแล้ว",
    "enum.shipment_event_type.picked_up": "รับพัสดุแล้ว",
    "enum.shipment_event_type.in_transit": "กำลังขนส่ง",
    "enum.shipment_event_type.arrived_at_port": "ถึงท่าเรือ",
    "enum.shipment_event_type.customs_cleared": "ผ่านศุลกากรแล้ว",
    "enum.shipment_event_type.out_for_delivery": "กำลังจัดส่ง",
    "enum.shipment_event_type.delivered": "ส่งมอบแล้ว",
    "enum.shipment_event_type.unknown": "ไม่ทราบ",
    "enum.carrier.standard": "มาตรฐาน",
    "enum.carrier.sea": "ขนส่งทางเรือ",
    "enum.carrier.air": "ขนส่งทางอากาศ",
    "enum.carrier.sea_freight": "ขนส่งทางเรือ",
    "enum.carrier.air_freight": "ขนส่งทางอากาศ",
    "enum.carrier.unknown": "ไม่ทราบ",
    "enum.tracking_provider.standard": "มาตรฐาน",
    "enum.tracking_provider.unknown": "ไม่ทราบ",
    "enum.invoice_type.invoice": "ใบแจ้งหนี้",
    "enum.invoice_type.proforma": "ใบแจ้งหนี้ Proforma",
    "enum.invoice_type.commercial": "ใบแจ้งหนี้เชิงพาณิชย์",
    "enum.invoice_type.credit_note": "ใบลดหนี้",
    "enum.invoice_type.unknown": "ไม่ทราบ",
    "enum.content_type.post": "บทความ",
    "enum.content_type.case": "กรณีศึกษา",
    "enum.content_type.unknown": "ไม่ทราบ",
    "enum.oem_status.inquiry": "สอบถาม",
    "enum.oem_status.sampling": "ตัวอย่าง",
    "enum.oem_status.formulation": "สูตร",
    "enum.oem_status.quotation": "ใบเสนอราคา",
    "enum.oem_status.contract": "สัญญา",
    "enum.oem_status.production": "การผลิต",
    "enum.oem_status.delivery": "การจัดส่ง",
    "enum.oem_status.completed": "เสร็จสมบูรณ์",
    "enum.oem_status.cancelled": "ยกเลิกแล้ว",
    "enum.oem_status.unknown": "ไม่ทราบ",
    "enum.sample_status.requested": "ขอแล้ว",
    "enum.sample_status.shipped": "จัดส่งแล้ว",
    "enum.sample_status.received": "ได้รับแล้ว",
    "enum.sample_status.approved": "อนุมัติแล้ว",
    "enum.sample_status.rejected": "ปฏิเสธแล้ว",
    "enum.sample_status.unknown": "ไม่ทราบ",
    "enum.document_status.draft": "ฉบับร่าง",
    "enum.document_status.pending": "รอดำเนินการ",
    "enum.document_status.confirmed": "ยืนยันแล้ว",
    "enum.document_status.cancelled": "ยกเลิกแล้ว",
    "enum.document_status.completed": "เสร็จสมบูรณ์",
    "enum.document_status.unknown": "ไม่ทราบ",
    "enum.document_type.quotation": "ใบเสนอราคา",
    "enum.document_type.proforma_invoice": "ใบแจ้งหนี้ Proforma",
    "enum.document_type.sales_contract": "สัญญาขาย",
    "enum.document_type.commercial_invoice": "ใบแจ้งหนี้เชิงพาณิชย์",
    "enum.document_type.packing_list": "รายการบรรจุ",
    "enum.document_type.bill_of_lading": "ใบตราส่งสินค้า",
    "enum.document_type.health_certificate": "ใบรับรองสุขอนามัย",
    "enum.document_type.origin_certificate": "ใบรับรองแหล่งกำเนิด",
    "enum.document_type.unknown": "ไม่ทราบ",
    "enum.negotiation_status.pending": "รอดำเนินการ",
    "enum.negotiation_status.accepted": "ยอมรับแล้ว",
    "enum.negotiation_status.rejected": "ปฏิเสธแล้ว",
    "enum.negotiation_status.expired": "หมดอายุ",
    "enum.negotiation_status.unknown": "ไม่ทราบ",
    "enum.invoice_status.draft": "ฉบับร่าง",
    "enum.invoice_status.sent": "ส่งแล้ว",
    "enum.invoice_status.paid": "ชำระแล้ว",
    "enum.invoice_status.overdue": "เกินกำหนด",
    "enum.invoice_status.voided": "เป็นโมฆะ",
    "enum.invoice_status.cancelled": "ยกเลิกแล้ว",
    "enum.invoice_status.unknown": "ไม่ทราบ",
    "enum.user_status.active": "ใช้งาน",
    "enum.user_status.inactive": "ไม่ใช้งาน",
    "enum.user_status.suspended": "ระงับ",
    "enum.user_status.pending": "รอดำเนินการ",
    "enum.user_status.deleted": "ลบแล้ว",
    "enum.user_status.unknown": "ไม่ทราบ",
    "enum.tool_status.success": "สำเร็จ",
    "enum.tool_status.error": "ข้อผิดพลาด",
    "enum.tool_status.running": "กำลังทำงาน",
    "enum.tool_status.unknown": "ไม่ทราบ",
    "enum.compliance_status.passed": "ผ่าน",
    "enum.compliance_status.failed": "ไม่ผ่าน",
    "enum.compliance_status.warnings": "คำเตือน",
    "enum.compliance_status.pending": "รอดำเนินการ",
    "enum.compliance_status.unknown": "ไม่ทราบ",
    "languages.en": "English",
    "languages.zh": "中文",
    "languages.ko": "한국어",
    "languages.ar": "العربية",
    "languages.ja": "日本語",
    "languages.th": "ไทย",
    "languages.vi": "Tiếng Việt",
    "languages.id": "Bahasa Indonesia",
    "languages.ms": "Bahasa Melayu",
    "yes": "ใช่",
    "no": "ไม่",
    "form.yes": "ใช่",
    "form.no": "ไม่",
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


def is_mostly_english(text: str) -> bool:
    if not isinstance(text, str) or len(text) <= 2:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&]+", text):
        return False
    latin = len(re.findall(r"[A-Za-z]", text))
    return latin / max(len(text), 1) > 0.55


def has_cjk(text: str) -> bool:
    return bool(re.search(r"[\u4e00-\u9fff\u3040-\u30ff\uac00-\ud7af]", text))


def has_thai(text: str) -> bool:
    return bool(re.search(r"[\u0E00-\u0E7F]", text))


def needs_retranslate(en_val: str, cached: str | None) -> bool:
    if cached is None:
        return True
    if cached == en_val:
        return True
    if has_cjk(cached):
        return True
    if is_mostly_english(cached) and not is_mostly_english(en_val):
        return True
    if not has_thai(cached) and len(en_val) > 3 and not is_mostly_english(en_val):
        return False
    if not has_thai(cached) and len(en_val) > 3:
        return True
    return False


def translate_one(text: str) -> str:
    protected, replacements = protect(text)
    translator = MyMemoryTranslator(source="en-US", target="th-TH")
    for attempt in range(5):
        try:
            result = translator.translate(protected)
            out = restore(result or text, replacements)
            if out and out != text:
                return out
        except Exception:
            time.sleep(0.8 * (attempt + 1))
    return text


def load_all_cache() -> dict:
    if CACHE_PATH.exists():
        return json.loads(CACHE_PATH.read_text(encoding="utf-8"))
    return {}


def save_all_cache(all_caches: dict) -> None:
    CACHE_PATH.write_text(json.dumps(all_caches, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    all_caches = load_all_cache()
    cache: dict[str, str] = all_caches.get(LOCALE, {})

    en_data = {f: json.loads((ROOT / SOURCE / f).read_text(encoding="utf-8")) for f in FILES}

    needed: set[str] = set()
    for filename in FILES:
        for en_val in flatten(en_data[filename]).values():
            if needs_retranslate(en_val, cache.get(en_val)):
                needed.add(en_val)

    print(f"Retranslating {len(needed)} strings...", flush=True)
    done = 0
    total = len(needed)

    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futures = {pool.submit(translate_one, text): text for text in sorted(needed, key=len)}
        for future in as_completed(futures):
            text = futures[future]
            try:
                result = future.result()
            except Exception:
                result = text
            with _cache_lock:
                cache[text] = result
            done += 1
            if done % 25 == 0 or done == total:
                print(f"  {done}/{total}", flush=True)
                all_caches[LOCALE] = cache
                save_all_cache(all_caches)

    all_caches[LOCALE] = cache
    save_all_cache(all_caches)

    for filename in FILES:
        en_flat = flatten(en_data[filename])
        out_flat = {}
        if filename == "admin.json":
            overrides = MANUAL
        elif filename == "common.json":
            overrides = MANUAL_COMMON
        else:
            overrides = {}
        for key, en_val in en_flat.items():
            if key in overrides:
                out_flat[key] = overrides[key]
            else:
                out_flat[key] = cache.get(en_val, en_val)
        out_path = ROOT / LOCALE / filename
        out_path.write_text(json.dumps(unflatten(out_flat), ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print(f"Wrote {LOCALE}/{filename}", flush=True)

    print("Done.", flush=True)


if __name__ == "__main__":
    main()
