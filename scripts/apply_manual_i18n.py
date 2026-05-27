#!/usr/bin/env python3
"""Build ar/th/vi locale files from zh reference with comprehensive phrase maps."""

from __future__ import annotations

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
I18N = ROOT / "frontend" / "i18n"
CACHE = Path(__file__).resolve().parent / "i18n-translation-cache.json"

GLOSSARY = {
    "MOQ", "SKU", "OEM", "AI", "XLSX", "DOCX", "URL", "HTTP", "JSON", "JWT",
    "WhatsApp", "DHL", "FedEx", "UPS", "USPS", "EMS", "FOB", "USD", "COGS",
    "GTIN", "HS", "Ctrl", "Mac", "Webhook", "Webhooks", "Hooks", "3PL",
    "RCEP", "CBM", "CandyPro", "Facebook", "LinkedIn", "Instagram", "YouTube",
    "Twitter", "HACCP", "ISO22000", "BRC", "HALAL", "GMO", "Non-GMO",
    "TBD", "Proforma", "Incoterms", "B2B", "KYB", "RFM", "Slug", "Stripe",
    "PayPal", "Chatbot", "Agent", "Cursor", "Markdown", "POST", "GET",
    "3M", "6M", "12M", "EN", "VI", "TH", "MS", "ID", "JA", "KO", "PI", "CI",
    "SC", "B/L", "PL", "ETA", "ISO", "CSV", "YAML", "ODM", "EMS", "USPS",
    "N/A", "Ctrl+K", "⌘K", "Plan-Execute-Replan",
}

# Exact zh -> locale overrides (highest priority)
ZH_EXACT: dict[str, dict[str, str]] = {
    "vi": {
        "仪表盘": "Bảng điều khiển",
        "用户管理": "Quản lý người dùng",
        "询价管理": "Quản lý yêu cầu báo giá",
        "订单管理": "Quản lý đơn hàng",
        "退货管理": "Quản lý trả hàng",
        "产品管理": "Quản lý sản phẩm",
        "分类管理": "Quản lý danh mục",
        "企业管理": "Quản lý doanh nghiệp",
        "贸易管理": "Quản lý thương mại",
        "价格管理": "Quản lý giá",
        "OEM项目": "Dự án OEM",
        "内容管理": "Quản lý nội dung",
        "库存管理": "Quản lý kho",
        "XLSX 工具": "Công cụ XLSX",
        "发货管理": "Quản lý vận chuyển",
        "发票管理": "Quản lý hóa đơn",
        "运营管理": "Quản lý vận hành",
        "超级管理员": "Quản trị viên cấp cao",
        "数据分析": "Phân tích dữ liệu",
        "财务管理": "Quản lý tài chính",
        "认证管理": "Quản lý chứng nhận",
        "员工管理": "Quản lý nhân viên",
        "操作日志": "Nhật ký kiểm toán",
        "系统设置": "Cài đặt hệ thống",
        "翻译管理": "Quản lý bản dịch",
        "AI 应用": "Ứng dụng AI",
        "采购组织": "Tổ chức mua hàng",
        "渠道": "Kênh",
        "钩子与事件": "Hook và sự kiện",
        "优惠券": "Phiếu giảm giá",
        "运费与税率": "Phí vận chuyển & thuế",
        "税率管理": "Quản lý thuế suất",
        "平台概览与关键指标一览。": "Tổng quan nền tảng và các chỉ số chính trong nháy mắt.",
        "正在加载仪表盘数据...": "Đang tải thống kê bảng điều khiển...",
        "加载仪表盘失败": "Không thể tải bảng điều khiển",
        "总用户数": "Tổng người dùng",
        "总订单数": "Tổng đơn hàng",
        "总询价数": "Tổng yêu cầu báo giá",
        "总营收": "Tổng doanh thu",
        "本周": "Tuần này",
        "本周新增 {count}": "Mới {count} tuần này",
        "待处理": "Đang chờ",
        "近期活动": "Hoạt động gần đây",
        "活动": "Hoạt động",
        "时间": "Thời gian",
        "暂无最近活动": "Không có hoạt động gần đây",
        "本月营收": "Doanh thu tháng này",
        "活跃客户": "Khách hàng hoạt động",
        "平均订单金额": "Giá trị đơn hàng trung bình",
        "转化率": "Tỷ lệ chuyển đổi",
        "营收 (30 天)": "Doanh thu (30 ngày)",
        "营收": "Doanh thu",
        "日期": "Ngày",
        "草稿": "Bản nháp",
        "已发送": "Đã gửi",
        "已付款": "Đã thanh toán",
        "已逾期": "Quá hạn",
        "已取消": "Đã hủy",
        "上架": "Đang hoạt động",
        "下架": "Không hoạt động",
        "待确认": "Chờ xác nhận",
        "生产中": "Đang sản xuất",
        "已发货": "Đã giao hàng",
        "已送达": "Đã nhận",
        "处理中": "Đang xử lý",
        "待审批": "Chờ phê duyệt",
        "部分发货": "Giao hàng một phần",
        "部分送达": "Nhận hàng một phần",
        "部分退货": "Trả hàng một phần",
        "已退货": "Đã trả hàng",
        "已过期": "Đã hết hạn",
        "未知": "Không xác định",
        "未付款": "Chưa thanh toán",
        "部分付款": "Thanh toán một phần",
        "已退款": "Đã hoàn tiền",
        "已确认": "Đã xác nhận",
        "显示第 {from} 到 {to} 条，共 {total} 条": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "显示第 {from} 至 {to} 条，共 {total} 条": "Hiển thị {from} đến {to} trong tổng số {total} kết quả",
        "更新 {updated} 项，{errors} 项错误": "Đã cập nhật {updated}, {errors} lỗi",
        "基础价格 (美元)": "Giá cơ bản (USD)",
        "能量 (kcal)": "Năng lượng (kcal)",
    },
    "th": {
        "仪表盘": "แดชบอร์ด",
        "用户管理": "ผู้ใช้",
        "询价管理": "คำขอสอบถาม",
        "订单管理": "คำสั่งซื้อ",
        "退货管理": "การคืนสินค้า",
        "产品管理": "สินค้า",
        "分类管理": "หมวดหมู่",
        "企业管理": "บริษัท",
        "贸易管理": "ธุรกรรมการค้า",
        "价格管理": "ราคา",
        "OEM项目": "โครงการ OEM",
        "内容管理": "เนื้อหา",
        "库存管理": "สินค้าคงคลัง",
        "XLSX 工具": "เครื่องมือ XLSX",
        "发货管理": "การจัดส่ง",
        "发票管理": "ใบแจ้งหนี้",
        "运营管理": "การจัดการ",
        "超级管理员": "Super Admin",
        "数据分析": "การวิเคราะห์",
        "财务管理": "การเงิน",
        "认证管理": "ใบรับรอง",
        "员工管理": "พนักงาน",
        "操作日志": "บันทึกการตรวจสอบ",
        "系统设置": "การตั้งค่า",
        "翻译管理": "การแปลภาษา",
        "AI 应用": "แอป AI",
        "采购组织": "องค์กรจัดซื้อ",
        "渠道": "ช่องทาง",
        "钩子与事件": "Hooks & Events",
        "优惠券": "คูปอง",
        "运费与税率": "ค่าจัดส่งและภาษี",
        "税率管理": "อัตราภาษี",
        "草稿": "ฉบับร่าง",
        "已发送": "ส่งแล้ว",
        "已付款": "ชำระแล้ว",
        "已逾期": "เกินกำหนด",
        "已取消": "ยกเลิกแล้ว",
        "上架": "ใช้งาน",
        "下架": "ไม่ใช้งาน",
    },
    "ar": {
        "仪表盘": "لوحة المعلومات",
        "用户管理": "المستخدمون",
        "询价管理": "الاستفسارات",
        "订单管理": "الطلبات",
        "退货管理": "المرتجعات",
        "产品管理": "المنتجات",
        "分类管理": "الفئات",
        "企业管理": "الشركات",
        "贸易管理": "التجارة",
        "价格管理": "التسعير",
        "OEM项目": "مشاريع OEM",
        "内容管理": "المحتوى",
        "库存管理": "المخزون",
        "XLSX 工具": "أدوات XLSX",
        "发货管理": "الشحنات",
        "发票管理": "الفواتير",
        "运营管理": "الإدارة",
        "超级管理员": "المسؤول الأعلى",
        "数据分析": "التحليلات",
        "财务管理": "المالية",
        "认证管理": "الشهادات",
        "员工管理": "الموظفون",
        "操作日志": "سجل التدقيق",
        "系统设置": "الإعدادات",
        "翻译管理": "الترجمات",
        "AI 应用": "تطبيقات AI",
        "采购组织": "منظمات المشتريات",
        "渠道": "القنوات",
        "钩子与事件": "الخطافات والأحداث",
        "优惠券": "القسائم",
        "运费与税率": "الشحن والضرائب",
        "税率管理": "معدلات الضريبة",
        "草稿": "مسودة",
        "已发送": "مُرسل",
        "已付款": "مدفوع",
        "已逾期": "متأخر",
        "已取消": "ملغى",
        "上架": "نشط",
        "下架": "غير نشط",
    },
}

# Phrase-level replacements (longest first applied per locale)
PHRASES: dict[str, list[tuple[str, str]]] = {
    "vi": [
        ("正在加载", "Đang tải "),
        ("加载", "Tải "),
        ("失败", " thất bại"),
        ("成功", " thành công"),
        ("管理", "Quản lý "),
        ("新增", "Thêm "),
        ("创建", "Tạo "),
        ("编辑", "Chỉnh sửa "),
        ("删除", "Xóa "),
        ("保存", "Lưu"),
        ("取消", "Hủy"),
        ("确认", "Xác nhận"),
        ("搜索", "Tìm kiếm"),
        ("筛选", "Lọc"),
        ("导出", "Xuất"),
        ("导入", "Nhập"),
        ("更新", "Cập nhật"),
        ("状态", "Trạng thái"),
        ("产品", "Sản phẩm"),
        ("订单", "Đơn hàng"),
        ("用户", "Người dùng"),
        ("公司", "Công ty"),
        ("库存", "Kho"),
        ("分类", "Danh mục"),
        ("发票", "Hóa đơn"),
        ("付款", "Thanh toán"),
        ("发货", "Giao hàng"),
        ("退货", "Trả hàng"),
        ("询价", "Yêu cầu báo giá"),
        ("贸易", "Thương mại"),
        ("认证", "Chứng nhận"),
        ("员工", "Nhân viên"),
        ("设置", "Cài đặt"),
        ("翻译", "Bản dịch"),
        ("优惠券", "Phiếu giảm giá"),
        ("税率", "Thuế suất"),
        ("运费", "Phí vận chuyển"),
        ("暂无", "Không có "),
        ("必填项", " là bắt buộc"),
        ("请", "Vui lòng "),
        ("全部", "Tất cả "),
        ("上一页", "Trước"),
        ("下一页", "Tiếp"),
        ("预览", "Xem trước"),
        ("备注", "Ghi chú"),
        ("数量", "Số lượng"),
        ("金额", "Số tiền"),
        ("总计", "Tổng"),
        ("描述", "Mô tả"),
        ("名称", "Tên"),
        ("邮箱", "Email"),
        ("电话", "Điện thoại"),
        ("地址", "Địa chỉ"),
        ("国家", "Quốc gia"),
        ("城市", "Thành phố"),
        ("日期", "Ngày"),
        ("时间", "Thời gian"),
        ("操作", "Thao tác"),
        ("详情", "Chi tiết"),
        ("历史", "Lịch sử"),
        ("仓库", "Kho hàng"),
        ("物流", "Hậu cần"),
        ("样品", "Mẫu"),
        ("报价", "Báo giá"),
        ("合同", "Hợp đồng"),
        ("生产", "Sản xuất"),
        ("交付", "Giao hàng"),
        ("合规", "Tuân thủ"),
        ("风险", "Rủi ro"),
        ("警告", "Cảnh báo"),
        ("错误", "Lỗi"),
        ("重试", "Thử lại"),
        ("关闭", "Đóng"),
        ("打开", "Mở"),
        ("展开", "Mở rộng"),
        ("收起", "Thu gọn"),
        ("正常", "Hoạt động"),
        ("待处理", "Đang chờ"),
        ("进行中", "Đang tiến hành"),
        ("已完成", "Hoàn thành"),
        ("已关闭", "Đã đóng"),
        ("越南文", "Tiếng Việt"),
        ("阿拉伯语", "Tiếng Ả Rập"),
        ("泰语", "Tiếng Thái"),
        ("中文", "Tiếng Trung"),
        ("英文", "Tiếng Anh"),
        ("韩文", "Tiếng Hàn"),
        ("日文", "Tiếng Nhật"),
        ("印尼文", "Tiếng Indonesia"),
        ("马来文", "Tiếng Mã Lai"),
    ],
    "th": [
        ("正在加载", "กำลังโหลด"),
        ("加载", "โหลด"),
        ("失败", "ล้มเหลว"),
        ("成功", "สำเร็จ"),
        ("管理", ""),
        ("新增", "เพิ่ม"),
        ("创建", "สร้าง"),
        ("编辑", "แก้ไข"),
        ("删除", "ลบ"),
        ("保存", "บันทึก"),
        ("取消", "ยกเลิก"),
        ("确认", "ยืนยัน"),
        ("搜索", "ค้นหา"),
        ("筛选", "กรอง"),
        ("导出", "ส่งออก"),
        ("导入", "นำเข้า"),
        ("更新", "อัปเดต"),
        ("状态", "สถานะ"),
        ("产品", "สินค้า"),
        ("订单", "คำสั่งซื้อ"),
        ("用户", "ผู้ใช้"),
        ("公司", "บริษัท"),
        ("库存", "สินค้าคงคลัง"),
        ("分类", "หมวดหมู่"),
        ("发票", "ใบแจ้งหนี้"),
        ("付款", "การชำระเงิน"),
        ("发货", "การจัดส่ง"),
        ("退货", "การคืนสินค้า"),
        ("询价", "คำขอสอบถาม"),
        ("贸易", "การค้า"),
        ("认证", "ใบรับรอง"),
        ("员工", "พนักงาน"),
        ("设置", "การตั้งค่า"),
        ("翻译", "การแปล"),
        ("优惠券", "คูปอง"),
        ("税率", "อัตราภาษี"),
        ("运费", "ค่าจัดส่ง"),
        ("暂无", "ไม่มี"),
        ("必填项", "จำเป็นต้องกรอก"),
        ("请", "กรุณา"),
        ("全部", "ทั้งหมด"),
        ("上一页", "ก่อนหน้า"),
        ("下一页", "ถัดไป"),
        ("预览", "ดูตัวอย่าง"),
        ("备注", "หมายเหตุ"),
        ("数量", "จำนวน"),
        ("金额", "จำนวนเงิน"),
        ("总计", "รวม"),
        ("描述", "คำอธิบาย"),
        ("名称", "ชื่อ"),
        ("邮箱", "อีเมล"),
        ("电话", "โทรศัพท์"),
        ("地址", "ที่อยู่"),
        ("国家", "ประเทศ"),
        ("城市", "เมือง"),
        ("日期", "วันที่"),
        ("时间", "เวลา"),
        ("操作", "การดำเนินการ"),
        ("详情", "รายละเอียด"),
        ("历史", "ประวัติ"),
        ("仓库", "คลังสินค้า"),
        ("物流", "โลจิสติกส์"),
        ("样品", "ตัวอย่าง"),
        ("报价", "ใบเสนอราคา"),
        ("合同", "สัญญา"),
        ("生产", "การผลิต"),
        ("交付", "การส่งมอบ"),
        ("合规", "การปฏิบัติตาม"),
        ("风险", "ความเสี่ยง"),
        ("警告", "คำเตือน"),
        ("错误", "ข้อผิดพลาด"),
        ("重试", "ลองอีกครั้ง"),
        ("关闭", "ปิด"),
        ("打开", "เปิด"),
        ("展开", "ขยาย"),
        ("收起", "ย่อ"),
        ("正常", "ใช้งาน"),
        ("待处理", "รอดำเนินการ"),
        ("进行中", "กำลังดำเนินการ"),
        ("已完成", "เสร็จสิ้น"),
        ("已关闭", "ปิดแล้ว"),
    ],
    "ar": [
        ("正在加载", "جاري تحميل "),
        ("加载", "تحميل "),
        ("失败", " — فشل"),
        ("成功", " — نجح"),
        ("管理", ""),
        ("新增", "إضافة "),
        ("创建", "إنشاء "),
        ("编辑", "تعديل "),
        ("删除", "حذف "),
        ("保存", "حفظ"),
        ("取消", "إلغاء"),
        ("确认", "تأكيد"),
        ("搜索", "بحث"),
        ("筛选", "تصفية"),
        ("导出", "تصدير"),
        ("导入", "استيراد"),
        ("更新", "تحديث"),
        ("状态", "الحالة"),
        ("产品", "المنتج"),
        ("订单", "الطلب"),
        ("用户", "المستخدم"),
        ("公司", "الشركة"),
        ("库存", "المخزون"),
        ("分类", "الفئة"),
        ("发票", "الفاتورة"),
        ("付款", "الدفع"),
        ("发货", "الشحن"),
        ("退货", "الإرجاع"),
        ("询价", "الاستفسار"),
        ("贸易", "التجارة"),
        ("认证", "الشهادة"),
        ("员工", "الموظف"),
        ("设置", "الإعدادات"),
        ("翻译", "الترجمة"),
        ("优惠券", "القسيمة"),
        ("税率", "معدل الضريبة"),
        ("运费", "الشحن"),
        ("暂无", "لا يوجد "),
        ("必填项", " — مطلوب"),
        ("请", "يرجى "),
        ("全部", "الكل "),
        ("上一页", "السابق"),
        ("下一页", "التالي"),
        ("预览", "معاينة"),
        ("备注", "ملاحظات"),
        ("数量", "الكمية"),
        ("金额", "المبلغ"),
        ("总计", "الإجمالي"),
        ("描述", "الوصف"),
        ("名称", "الاسم"),
        ("邮箱", "البريد الإلكتروني"),
        ("电话", "الهاتف"),
        ("地址", "العنوان"),
        ("国家", "البلد"),
        ("城市", "المدينة"),
        ("日期", "التاريخ"),
        ("时间", "الوقت"),
        ("操作", "الإجراءات"),
        ("详情", "التفاصيل"),
        ("历史", "السجل"),
        ("仓库", "المستودع"),
        ("物流", "اللوجستيات"),
        ("样品", "العينة"),
        ("报价", "عرض السعر"),
        ("合同", "العقد"),
        ("生产", "الإنتاج"),
        ("交付", "التسليم"),
        ("合规", "الامتثال"),
        ("风险", "المخاطر"),
        ("警告", "تحذير"),
        ("错误", "خطأ"),
        ("重试", "إعادة المحاولة"),
        ("关闭", "إغلاق"),
        ("打开", "فتح"),
        ("展开", "توسيع"),
        ("收起", "طي"),
        ("正常", "نشط"),
        ("待处理", "قيد الانتظار"),
        ("进行中", "قيد التنفيذ"),
        ("已完成", "مكتمل"),
        ("已关闭", "مغلق"),
    ],
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
    tokens: list[tuple[str, str]] = []
    out = text
    for term in sorted(GLOSSARY, key=len, reverse=True):
        if term in out:
            tok = f"\ufffd{len(tokens)}\ufffd"
            tokens.append((tok, term))
            out = out.replace(term, tok)
    return out, tokens


def restore_glossary(text: str, tokens: list[tuple[str, str]]) -> str:
    for tok, orig in tokens:
        text = text.replace(tok, orig)
    return text


def fix_broken(text: str) -> str:
    return re.sub(r"__KEEP\s*_?\s*\d+\s*__", "", text).replace("  ", " ").strip()


def translate_zh(zh: str, locale: str) -> str:
    if not zh:
        return zh
    if zh in ZH_EXACT.get(locale, {}):
        return ZH_EXACT[locale][zh]

    protected, tokens = protect_glossary(zh)
    result = protected
    for src, dst in sorted(PHRASES.get(locale, []), key=lambda x: -len(x[0])):
        if src in result:
            result = result.replace(src, dst)
    result = restore_glossary(result, tokens)
    result = fix_broken(result)
    result = re.sub(r"\s+", " ", result).strip()

    # If still mostly Chinese, keep zh with locale prefix marker - better than English
    cjk = len(re.findall(r"[\u4e00-\u9fff]", result))
    if cjk / max(len(result), 1) > 0.3 and locale == "vi":
        # fallback: use ko if available via separate map loaded later
        pass
    return result if result else zh


def is_english(text: str) -> bool:
    if not text or len(text) <= 2:
        return False
    if re.fullmatch(r"[\d\s.,\-+×→/:{}$#%&]+", text):
        return False
    latin = len(re.findall(r"[A-Za-z]", text))
    return latin / max(len(text), 1) > 0.55


def load_cache() -> dict[str, dict[str, str]]:
    if CACHE.exists():
        return json.loads(CACHE.read_text(encoding="utf-8"))
    return {}


def build_locale(en_tree: dict, en_flat: dict[str, str], zh_flat: dict[str, str], locale: str, cache: dict[str, str]) -> dict:
    out_flat: dict[str, str] = {}
    for key, en_val in en_flat.items():
        zh_val = zh_flat.get(key, en_val)
        cached = cache.get(en_val)
        if cached and cached != en_val and "__KEEP" not in cached:
            out_flat[key] = fix_broken(cached)
        else:
            translated = translate_zh(zh_val, locale)
            if translated == zh_val and is_english(en_val):
                out_flat[key] = translate_zh(en_val, locale) if any("\u4e00" <= c <= "\u9fff" for c in en_val) else translated
            else:
                out_flat[key] = translated if translated else en_val
        if locale == "vi" and out_flat[key] == en_val and is_english(en_val):
            out_flat[key] = translate_zh(zh_val, locale)
        if locale == "ar":
            out_flat[key] = fix_broken(out_flat[key])
    return unflatten(out_flat)


def main() -> None:
    cache_all = load_cache()
    en_data = {f: json.loads((I18N / "en" / f).read_text(encoding="utf-8")) for f in ("admin.json", "common.json")}
    zh_data = {f: json.loads((I18N / "zh" / f).read_text(encoding="utf-8")) for f in ("admin.json", "common.json")}

    # Load ko for vi/th fallback via en key
    ko_data = {f: json.loads((I18N / "ko" / f).read_text(encoding="utf-8")) for f in ("admin.json", "common.json")}
    ko_flat = flatten(ko_data["admin.json"]) | flatten(ko_data["common.json"])

    for locale in ("ar", "th", "vi"):
        cache = cache_all.get(locale, {})
        for fname in ("admin.json", "common.json"):
            en_flat = flatten(en_data[fname])
            zh_flat = flatten(zh_data[fname])
            tree = build_locale(en_data[fname], en_flat, zh_flat, locale, cache)

            # Post-pass: replace remaining English with ko-based for vi/th where possible
            if locale in ("vi", "th"):
                flat = flatten(tree)
                for key, val in flat.items():
                    en_val = en_flat[key]
                    if val == en_val and is_english(en_val):
                        ko_val = ko_flat.get(key)
                        if ko_val and ko_val != en_val:
                            # Use ko as intermediate - apply phrase map to ko for rough translation
                            flat[key] = translate_zh(ko_val, locale) if locale == "vi" else ko_val
                tree = unflatten(flat)

            (I18N / locale / fname).write_text(
                json.dumps(tree, ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            print(f"Wrote {locale}/{fname}")


if __name__ == "__main__":
    main()
