#!/usr/bin/env python3
"""Apply final manual Thai patches to th/admin.json."""

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"

PATCH: dict[str, str] = {
    "admin.categories.description": "จัดการหมวดหมู่สินค้าพร้อมชื่อ นามแฝง และคำอธิบายหลายภาษา",
    "admin.certifications.description": "จัดการใบรับรองโรงงาน (HACCP, ISO22000, BRC, HALAL ฯลฯ)",
    "admin.shipments.description": "จัดการการจัดส่งคำสั่งซื้อ การติดตาม และสถานะการส่งมอบ",
    "admin.shippingRates.description": "ตารางอัตราค่าขนส่งตามปลายทาง จำนวนวันโดยประมาณเป็นค่าอ้างอิงสำหรับการเสนอราคาเท่านั้น — ไม่ใช่เวลาจัดส่งที่รับประกัน",
    "admin.shippingRates.base_cost": "ค่าใช้จ่ายพื้นฐาน",
    "admin.shippingRates.col_base_cost": "ค่าใช้จ่ายพื้นฐาน",
    "admin.shippingRates.estimated_days": "วันโดยประมาณ",
    "admin.taxRates.country": "ประเทศ (ISO)",
    "admin.inquiries.user_id": "รหัสผู้ใช้ (ไม่บังคับ)",
    "admin.inquiries.product_ids_placeholder": "รหัสสินค้า คั่นด้วยจุลภาค",
    "admin.inquiries.validation_required": "ต้องระบุบริษัท ผู้ติดต่อ และอีเมล",
    "admin.inquiryDetail.confirm_specs_desc": "ยืนยันประเภทบรรจุภัณฑ์ น้ำหนัก ขนาด และมาตรฐานคุณภาพ เมื่อทั้งสองฝ่ายยืนยันแล้ว คำขอสอบถามจะเปลี่ยนเป็นสถานะ 'confirmed'",
    "admin.inventory.batch_delete_result": "ลบ {deleted} รายการ, {errors} ข้อผิดพลาด",
    "admin.inventory.batch_edit_result": "อัปเดต {updated} รายการ, {errors} ข้อผิดพลาด",
    "admin.orders.confirm_refund": "คุณแน่ใจหรือไม่ว่าต้องการคืนเงินการชำระนี้?",
    "admin.organizations.select_org_placeholder": "เลือกองค์กร…",
    "admin.payment_policy.requires_payment_before_execution": "ต้องยืนยันการชำระเงิน (ชำระเต็มหรือบางส่วน) ก่อนเข้าสู่ขั้นตอนดำเนินการ (confirmed, production, shipped, delivered)",
    "admin.pricing.confirm_delete_price_list": "ลบรายการราคานี้หรือไม่?",
    "admin.products.ai_section_nutrition": "โภชนาการ (ต่อ 100g)",
    "admin.products.ai_translate_reloaded": "การแปลเสร็จสมบูรณ์บนเซิร์ฟเวอร์ โหลดคำแปลล่าสุดจากฐานข้อมูลแล้ว",
    "admin.products.field_carbohydrates_g": "คาร์โบไฮเดรต (g)",
    "admin.products.field_cocoa_solids_pct": "โกโก้ %",
    "admin.products.field_energy_kcal": "พลังงาน (kcal)",
    "admin.products.field_fiber_g": "ใยอาหาร (g)",
    "admin.products.field_gross_weight_per_carton": "น้ำหนักรวม/ลัง (kg)",
    "admin.products.field_milk_solids_pct": "ของแข็งจากนม %",
    "admin.products.field_net_weight_per_pack": "น้ำหนักสุทธิ/แพ็ค (g)",
    "admin.products.field_net_weight_per_piece": "น้ำหนักสุทธิ/ชิ้น (g)",
    "admin.products.field_sample_lead_time_hint": "อ้างอิงระยะเวลาตัวอย่าง — ยืนยันโดยฝ่ายขายในใบเสนอราคาตัวอย่าง ห้ามใช้ \"approx. X weeks\"",
    "admin.products.field_saturated_fat_g": "ไขมันอิ่มตัว (g)",
    "admin.products.field_sugars_g": "น้ำตาล (g)",
    "admin.products.field_water_activity": "Water Activity (aʷ)",
    "admin.products.flavors": "รสชาติ (คั่นด้วยจุลภาค)",
    "admin.products.images": "รูปภาพ (คั่นด้วยจุลภาค)",
    "admin.products.lead_time_hint": "กรอกหลังฝ่ายขายยืนยันเพื่อแสดงในแคตตalog อาจระบุอ้างอิงชัดเจน (เช่น \"25 วันทำการหลังยืนยันคำสั่งซื้อ\") — ห้ามใช้คำกำกวม เช่น \"approx. X weeks\" ว่างไว้หรือใช้ \"confirmed at order time\" จนกว่าจะยืนยัน",
    "admin.products.packaging_placeholder": "flow-wrap, foil, box, bag, jar",
    "admin.products.section_dimensions": "ขนาดสินค้า (mm)",
    "admin.products.section_nutrition": "โภชนาการ (ต่อ 100g)",
    "admin.products.shapes": "รูปทรง (คั่นด้วยจุลภาค)",
    "admin.products.translations_hint": "ข้อความที่ลูกค้าเห็นทั้งหมดอยู่ที่นี่เท่านั้น ภาษาจีนเป็นภาษาต้นฉบับ ฟิลด์สเกลาร์ซิงค์เมื่อบันทึก",
    "admin.docLabels.gross_weight": "น้ำหนักรวม (kg)",
    "admin.docLabels.total_net_weight": "น้ำหนักสุทธิรวม (kg)",
    "admin.content.case_services": "บริการ (คั่นด้วยจุลภาค)",
    "admin.content.post_read_time": "เวลาอ่าน (นาที)",
    "admin.content.post_tags": "แท็ก (คั่นด้วยจุลภาค)",
    "admin.content.translations_hint": "ข้อความที่ลูกค้าเห็นทั้งหมดอยู่ที่นี่เท่านั้น ภาษาจีนเป็นภาษาต้นฉบับ ฟิลด์สเกลาร์ซิงค์เมื่อบันทึก",
    "admin.content.ai_empty_preview": "ใส่หัวข้อแล้วคลิก Generate Chinese Draft ตรวจสอบ จากนั้นเผยแพร่พร้อมคำแปลครบทุกภาษา",
    "admin.content.ai_review_banner": "ตรวจสอบและแก้ไขฉบับร่างภาษาจีนด้านล่าง เมื่อเผยแพร่ ทุกภาษาจะถูกแปลและซิงค์ไปยังบล็อกสาธารณะ",
    "admin.content.ai_revise_placeholder": "เช่น ย่อหัวข้อ เพิ่มรายละเอียด EU compliance ในย่อหน้า 2 ใช้โทนมืออาชีพมากขึ้น…",
    "admin.content.ai_revising": "กำลังแก้ไข…",
    "admin.content.ai_topic_placeholder": "อธิบายหัวข้อหรือประเด็นสำคัญ (ภาษาใดก็ได้) AI จะร่างเป็นภาษาจีนก่อน…",
    "admin.content.inline_ai_hint": "เลือกข้อความในหัวข้อ บทคัดย่อ หรือเนื้อหา แล้วกด Ctrl+K (Mac: ⌘K) เพื่อแก้ไขด้วย AI พร้อม diff — Accept หรือ Reject",
    "admin.content.inline_ai_loading": "กำลังแก้ไข…",
    "admin.translations.placeholder_key": "เช่น errors.my_new_code",
    "admin.xlsx.exporting": "กำลังส่งออก…",
    "admin.xlsx.trade_status": "สถานะการค้า (ไม่บังคับ)",
    "admin.ai.mode_chatbot": "Chatbot (optional order)",
    "admin.ai_help.content_modal_ai.body": "ขณะแก้ไขเนื้อหา ใส่หัวข้อแล้วคลิก Generate with AI เพื่อเติมหัวข้อ เนื้อหา ฯลฯ (ฉบับร่างจีน ไม่แปลอัตโนมัติ)\nใช้ AI Translate ในส่วน Translations สำหรับภาษาอื่นหลังบันทึก",
    "admin.ai_help.content_translate.body": "เมื่อแก้ไขเนื้อหาที่บันทึกแล้ว คลิก AI Translate ใต้ Translations\nฟิลด์ต้นฉบับ (มักเป็นจีน) จะถูกแปลเป็น en/ko/ar/ja/th/vi/id/ms\nตรวจสอบแต่ละแท็บภาษาแล้วบันทึก",
    "admin.ai_help.inventory_import_ai.body": "หลังอัปโหลดไฟล์ Excel สินค้าคงคลัง/สินค้า AI จะแนะนำการแมปคอลัมน์ → ฟิลด์\nตรวจสอบแบนเนอร์สีน้ำเงินและปรับการแมปด้านล่างก่อนยืนยันนำเข้า",
    "admin.ai_help.products_translate.body": "เมื่อแก้ไขสินค้า คลิก AI Translate ใต้ Translations\nชื่อ สรุป คำอธิบาย ส่วนผสม ฯลฯ จะถูกแปลจากภาษาหลัก\nตรวจสอบแท็บภาษาก่อนบันทึก",
    "admin.ai_help.trades_assistant.body": "แชทเกี่ยวกับการค้านี้ — เอกสาร การปฏิบัติตาม โลจิสติกส์\nAgent อาจเรียกเครื่องมือการค้า คำตอบสตรีมในแผงแชท\nใช้ AI Generate Document ด้านล่างสำหรับร่าง PI/CI/SC",
    "admin.ai_help.trades_generate_doc.body": "เลือกประเภทเอกสาร (PI, CI, SC ฯลฯ) และบริบท (ถ้ามี)\nAI ร่างจากข้อมูลการค้าปัจจุบันและบันทึกในระบบ\nตรวจสอบฟิลด์ในรายการเอกสารก่อนส่งให้ลูกค้าเสมอ",
    "admin.ai_help.xlsx_translate.body": "อัปโหลด Excel เลือกโหมดแปลภาษาเดียวหรือหลายภาษา\nAI แปลหัวคอลัมน์หรือคีย์คอลัมน์ที่เลือก — มีประโยชน์สำหรับรายงานและ term sheets\nดาวน์โหลดผลลัพธ์ ลองตัวอย่างเล็กก่อนงาน batch",
    "admin.ai_help.ai_console.body": "เลือกโหมด Agent ทางขวา แล้วแชท:\n• Chatbot: Q&A อาจมีบริบทคำสั่งซื้อ\n• Plan-Execute-Replan / B2B Coordinator / Trade Assistant: streaming agents\n• Recommend / Search / Translate / Analyze inquiry / Generate quotation: one-shot JSON\nสำหรับ Chatbot เลือกคำสั่งซื้อในแถบด้านข้างเพื่อพูดคุยเกี่ยวกับคำสั่งนั้น",
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


def main() -> None:
    flat = flatten(json.loads((ROOT / "th" / "admin.json").read_text(encoding="utf-8")))
    for key, value in PATCH.items():
        flat[key] = value
    (ROOT / "th" / "admin.json").write_text(
        json.dumps(unflatten(flat), ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    print(f"Patched {len(PATCH)} keys")


if __name__ == "__main__":
    main()
