#!/usr/bin/env python3
"""Rebuild locales from en Lingva cache + term overrides + gap fill."""

from __future__ import annotations

import json
import re
import sys
import time
import urllib.parse
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

import requests

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache.json"
LOCALES = ("ar", "th", "vi")
LANG = {"ar": "ar", "th": "th", "vi": "vi"}
WORKERS = 8

sys.path.insert(0, str(Path(__file__).resolve().parent))
from apply_priority_i18n import TERM, flatten, unflatten  # noqa: E402

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO", "N/A",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Stripe",
    "PayPal", "Chatbot", "Agent", "Cursor", "Markdown", "POST", "GET",
    "3M", "6M", "12M", "EN", "VI", "TH", "MS", "ID", "JA", "KO", "PI", "CI",
    "SC", "B/L", "PL", "ETA", "ISO", "CSV", "YAML", "ODM", "Plan-Execute-Replan",
    "⌘K", "Ctrl+K",
}

MANUAL = {
    "vi": {
        "admin.brand": "CandyPro Admin",
        "admin.langLabel": "VI", "admin.langLabelShort": "VI", "admin.langTitle": "Tiếng Việt",
        "admin.products.base_price": "Giá cơ bản (USD)",
        "admin.products.field_energy_kcal": "Năng lượng (kcal)",
        "admin.users.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "admin.staff.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "admin.auditLog.showing": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "admin.inventory.batch_edit_result": "Đã cập nhật {updated}, {errors} lỗi",
        "admin.inventory.showing": "Hiển thị {from} đến {to} trong tổng số {total}",
        "admin.products.showing": "Hiển thị {from} đến {to} trong tổng số {total} sản phẩm",
    },
    "ar": {
        "admin.langLabel": "العربية", "admin.langLabelShort": "ع", "admin.langTitle": "العربية",
        "admin.brand": "CandyPro Admin",
    },
    "th": {
        "admin.langLabel": "TH", "admin.langLabelShort": "TH", "admin.langTitle": "ไทย",
        "admin.brand": "CandyPro Admin",
    },
}

ENUM_ZH = json.loads((Path(__file__).resolve().parents[1] / "frontend" / "i18n" / "zh" / "common.json").read_text(encoding="utf-8"))["enum"]

ENUM_OVERRIDES = {
    "vi": {
        "order_status.pending": "Đang chờ", "order_status.confirmed": "Đã xác nhận",
        "order_status.production": "Đang sản xuất", "order_status.shipped": "Đã giao hàng",
        "order_status.delivered": "Đã nhận", "order_status.cancelled": "Đã hủy",
        "order_status.processing": "Đang xử lý", "order_status.pending_confirmation": "Chờ xác nhận",
        "order_status.pending_approval": "Chờ phê duyệt", "order_status.partially_shipped": "Giao hàng một phần",
        "order_status.partially_delivered": "Nhận hàng một phần", "order_status.partially_returned": "Trả hàng một phần",
        "order_status.returned": "Đã trả hàng", "order_status.expired": "Đã hết hạn", "order_status.unknown": "Không xác định",
        "payment_status.unpaid": "Chưa thanh toán", "payment_status.paid": "Đã thanh toán",
        "payment_status.partial": "Thanh toán một phần", "payment_status.refunded": "Đã hoàn tiền",
        "payment_status.pending": "Đang chờ", "payment_status.confirmed": "Đã xác nhận", "payment_status.unknown": "Không xác định",
        "product_status.active": "Đang bán", "product_status.inactive": "Ngừng bán",
        "product_status.draft": "Bản nháp", "product_status.unknown": "Không xác định",
        "company_status.pending": "Chờ duyệt", "company_status.verified": "Đã xác minh",
        "company_status.rejected": "Đã từ chối", "company_status.active": "Hoạt động",
        "company_status.suspended": "Đã tạm ngưng", "company_status.inactive": "Không hoạt động", "company_status.unknown": "Không xác định",
        "kyb_status.pending": "Chờ duyệt", "kyb_status.approved": "Đã phê duyệt",
        "kyb_status.rejected": "Đã từ chối", "kyb_status.unknown": "Không xác định",
        "inquiry_status.pending": "Đang chờ", "inquiry_status.contacted": "Đã liên hệ",
        "inquiry_status.quoted": "Đã báo giá", "inquiry_status.negotiating": "Đang thương lượng",
        "inquiry_status.won": "Thành công", "inquiry_status.lost": "Thất bại",
        "inquiry_status.open": "Đang mở", "inquiry_status.closed": "Đã đóng",
        "inquiry_status.converted": "Đã chuyển đổi", "inquiry_status.unknown": "Không xác định",
        "inquiry_priority.normal": "Bình thường", "inquiry_priority.high": "Cao",
        "inquiry_priority.low": "Thấp", "inquiry_priority.urgent": "Khẩn cấp", "inquiry_priority.unknown": "Không xác định",
    },
    "th": {
        "order_status.pending": "รอดำเนินการ", "order_status.confirmed": "ยืนยันแล้ว",
        "order_status.production": "กำลังผลิต", "order_status.shipped": "จัดส่งแล้ว",
        "order_status.delivered": "ส่งมอบแล้ว", "order_status.cancelled": "ยกเลิกแล้ว",
        "order_status.processing": "กำลังดำเนินการ", "order_status.pending_confirmation": "รอการยืนยัน",
        "order_status.pending_approval": "รอการอนุมัติ", "order_status.partially_shipped": "จัดส่งบางส่วน",
        "order_status.partially_delivered": "ส่งมอบบางส่วน", "order_status.partially_returned": "คืนบางส่วน",
        "order_status.returned": "คืนแล้ว", "order_status.expired": "หมดอายุ", "order_status.unknown": "ไม่ทราบ",
        "payment_status.unpaid": "ยังไม่ชำระ", "payment_status.paid": "ชำระแล้ว",
        "payment_status.partial": "ชำระบางส่วน", "payment_status.refunded": "คืนเงินแล้ว",
        "payment_status.pending": "รอดำเนินการ", "payment_status.confirmed": "ยืนยันแล้ว", "payment_status.unknown": "ไม่ทราบ",
        "product_status.active": "วางขาย", "product_status.inactive": "ไม่วางขาย",
        "product_status.draft": "ฉบับร่าง", "product_status.unknown": "ไม่ทราบ",
    },
    "ar": {
        "order_status.pending": "قيد الانتظار", "order_status.confirmed": "مؤكد",
        "order_status.production": "قيد الإنتاج", "order_status.shipped": "تم الشحن",
        "order_status.delivered": "تم التسليم", "order_status.cancelled": "ملغى",
        "order_status.processing": "قيد المعالجة", "order_status.pending_confirmation": "بانتظار التأكيد",
        "order_status.pending_approval": "بانتظار الموافقة", "order_status.partially_shipped": "شُحن جزئيًا",
        "order_status.partially_delivered": "تسليم جزئي", "order_status.partially_returned": "إرجاع جزئي",
        "order_status.returned": "مُرتجع", "order_status.expired": "منتهي", "order_status.unknown": "غير معروف",
        "payment_status.unpaid": "غير مدفوع", "payment_status.paid": "مدفوع",
        "payment_status.partial": "مدفوع جزئيًا", "payment_status.refunded": "مسترد",
        "payment_status.pending": "قيد الانتظار", "payment_status.confirmed": "مؤكد", "payment_status.unknown": "غير معروف",
        "product_status.active": "نشط", "product_status.inactive": "غير نشط",
        "product_status.draft": "مسودة", "product_status.unknown": "غير معروف",
    },
}

SESSION = requests.Session()
SESSION.trust_env = False
import os
for k in ("HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy"):
    os.environ.pop(k, None)


def protect(text: str) -> tuple[str, list[tuple[str, str]]]:
    tokens: list[tuple[str, str]] = []
    out = text
    for ph in re.findall(r"\{[a-zA-Z0-9_]+\}", text):
        tok = f"ZZPH{len(tokens)}ZZ"
        tokens.append((tok, ph))
        out = out.replace(ph, tok, 1)
    for term in sorted(GLOSSARY, key=len, reverse=True):
        if term in out:
            tok = f"ZZK{len(tokens)}ZZ"
            tokens.append((tok, term))
            out = out.replace(term, tok)
    return out, tokens


def restore(text: str, tokens: list[tuple[str, str]]) -> str:
    for tok, orig in tokens:
        text = text.replace(tok, orig)
    return text


def lingva_en(text: str, target: str) -> str:
    protected, tokens = protect(text)
    url = f"https://lingva.ml/api/v1/en/{target}/{urllib.parse.quote(protected, safe='')}"
    for attempt in range(4):
        try:
            r = SESSION.get(url, timeout=25)
            if r.status_code == 200:
                return restore(r.json().get("translation") or text, tokens)
        except Exception:
            pass
        time.sleep(0.4 * (attempt + 1))
    return text


def is_english(text: str) -> bool:
    if not text or len(text) <= 2:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&]+", text):
        return False
    return len(re.findall(r"[A-Za-z]", text)) / max(len(text), 1) > 0.55


def has_cjk(text: str) -> bool:
    return bool(re.search(r"[\u4e00-\u9fff]", text))


def fill_cache(strings: list[str], locale: str, cache: dict[str, str]) -> None:
    todo = [s for s in strings if s not in cache or cache[s] == s or has_cjk(cache.get(s, ""))]
    print(f"  fill {locale}: {len(todo)}", flush=True)
    done = 0
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futs = {pool.submit(lingva_en, s, LANG[locale]): s for s in todo}
        for fut in as_completed(futs):
            s = futs[fut]
            try:
                cache[s] = fut.result()
            except Exception:
                cache[s] = s
            done += 1
            if done % 50 == 0 or done == len(todo):
                print(f"    {done}/{len(todo)}", flush=True)
            time.sleep(0.06)


def resolve(en_val: str, locale: str, cache: dict[str, str]) -> str:
    if en_val in TERM:
        return TERM[en_val][locale]
    val = cache.get(en_val, en_val)
    if has_cjk(val):
        val = cache.get(en_val, en_val)
    return val


def apply_enum_overrides(tree: dict, locale: str) -> dict:
    flat = flatten(tree)
    overrides = ENUM_OVERRIDES.get(locale, {})
    for key, val in overrides.items():
        flat[f"enum.{key}"] = val
    return unflatten(flat)


def main() -> None:
    sys.stdout.reconfigure(encoding="utf-8")
    cache = json.loads(CACHE_PATH.read_text(encoding="utf-8"))

    en_admin = json.loads((ROOT / "en" / "admin.json").read_text(encoding="utf-8"))
    en_common = json.loads((ROOT / "en" / "common.json").read_text(encoding="utf-8"))
    en_flat = flatten(en_admin) | flatten(en_common)
    all_en = sorted(set(en_flat.values()), key=len)

    for locale in LOCALES:
        cache.setdefault(locale, {})
        fill_cache(all_en, locale, cache[locale])
    CACHE_PATH.write_text(json.dumps(cache, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    for locale in LOCALES:
        for fname, en_tree in [("admin.json", en_admin), ("common.json", en_common)]:
            en_f = flatten(en_tree)
            out_f = {}
            for key, en_val in en_f.items():
                if key in MANUAL.get(locale, {}):
                    out_f[key] = MANUAL[locale][key]
                else:
                    out_f[key] = resolve(en_val, locale, cache[locale])
            tree = unflatten(out_f)
            if fname == "common.json":
                tree = apply_enum_overrides(tree, locale)
            (ROOT / locale / fname).write_text(
                json.dumps(tree, ensure_ascii=False, indent=2) + "\n", encoding="utf-8",
            )
            print(f"Wrote {locale}/{fname}", flush=True)

    print("Done.", flush=True)


if __name__ == "__main__":
    main()
