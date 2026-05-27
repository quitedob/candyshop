#!/usr/bin/env python3
"""Generate fully translated ja/ko admin.json and common.json from en reference."""

from __future__ import annotations

import json
import re
import sys
import threading
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

from deep_translator import GoogleTranslator

ROOT = Path(__file__).resolve().parents[1] / "frontend" / "i18n"
CACHE_PATH = Path(__file__).resolve().parent / "i18n-translation-cache-ja-ko-v2.json"
LOCALES = ("ja", "ko")
FILES = ("admin.json", "common.json")
SOURCE = "en"
REF = "zh"
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
    "Super", "Admin", "Superadmin", "Diff", "Trade", "translate_content",
    "INVOICE", "CREDIT", "NOTE", "TAX", "PROFORMA", "COMMERCIAL",
}

PLACEHOLDER_RE = re.compile(r"(\{[a-zA-Z0-9_]+\})")
TOKEN_RE = re.compile(
    r"(\{[a-zA-Z0-9_]+\}|⌘K|Ctrl\+K|1200×630|aʷ|→|…|\.\.\.|&|[0-9]+(?:\.[0-9]+)?%?|[A-Za-z][A-Za-z0-9+#./\-]{1,})"
)

LANG_TARGETS = {
    "ja": "ja",
    "ko": "ko",
}

# Paths use file-relative flatten: admin.json keys prefixed with "admin.", common.json without wrapper.
MANUAL_OVERRIDES: dict[str, dict[str, str]] = {
    "ja": {
        "admin.brand": "CandyPro 管理画面",
        "admin.logout": "ログアウト",
        "admin.theme.dark": "ダークモード",
        "admin.theme.light": "ライトモード",
        "admin.langLabel": "日本語",
        "admin.langLabelShort": "日",
        "admin.langTitle": "日本語",
        "admin.nav.dashboard": "ダッシュボード",
        "admin.nav.users": "ユーザー管理",
        "admin.nav.inquiries": "お問い合わせ管理",
        "admin.nav.orders": "注文管理",
        "admin.nav.returns": "返品管理",
        "admin.nav.products": "製品管理",
        "admin.nav.categories": "カテゴリ管理",
        "admin.nav.companies": "企業管理",
        "admin.nav.trades": "貿易管理",
        "admin.nav.pricing": "価格管理",
        "admin.nav.oemProjects": "OEMプロジェクト",
        "admin.nav.content": "コンテンツ管理",
        "admin.nav.inventory": "在庫管理",
        "admin.nav.xlsx": "XLSXツール",
        "admin.nav.shipments": "出荷管理",
        "admin.nav.invoices": "請求書管理",
        "admin.nav.management": "運営管理",
        "admin.nav.superadmin": "スーパー管理者",
        "admin.nav.analytics": "データ分析",
        "admin.nav.financial": "財務管理",
        "admin.nav.certifications": "認証管理",
        "admin.nav.staff": "スタッフ管理",
        "admin.nav.auditLog": "監査ログ",
        "admin.nav.settings": "システム設定",
        "admin.nav.translations": "翻訳管理",
        "admin.nav.ai": "AIアプリ",
        "admin.nav.organizations": "購買組織",
        "admin.nav.channels": "販売チャネル",
        "admin.nav.webhooks": "Webhook",
        "admin.nav.hooks": "フックとイベント",
        "admin.nav.coupons": "クーポン",
        "admin.nav.shipping_rates": "送料・税率",
        "admin.nav.tax_rates": "税率管理",
        "admin.dashboard.title": "ダッシュボード",
        "admin.dashboard.subtitle": "プラットフォーム概要と主要指標を一覧で確認できます。",
        "admin.dashboard.loading": "ダッシュボード統計を読み込み中...",
        "admin.dashboard.error": "ダッシュボードの読み込みに失敗しました",
        "admin.dashboard.total_users": "ユーザー総数",
        "admin.dashboard.total_orders": "注文総数",
        "admin.dashboard.total_inquiries": "お問い合わせ総数",
        "admin.dashboard.total_revenue": "総売上",
        "admin.dashboard.this_week": "今週",
        "admin.dashboard.new_users_this_week": "今週 {count} 件の新規",
        "admin.dashboard.pending": "保留中",
        "admin.dashboard.recent_activity": "最近のアクティビティ",
        "admin.dashboard.col_activity": "アクティビティ",
        "admin.dashboard.col_time": "時間",
        "admin.dashboard.no_activity": "最近のアクティビティはありません",
        "admin.dashboard.revenue_this_month": "今月の売上",
        "admin.dashboard.active_customers": "アクティブ顧客",
        "admin.dashboard.avg_order_value": "平均注文額",
        "admin.dashboard.conversion_rate": "コンバージョン率",
        "admin.dashboard.revenue_chart": "売上推移（30日間）",
        "admin.dashboard.revenue": "売上",
        "admin.dashboard.date": "日付",
        "admin.products.title": "製品管理",
        "admin.products.description": "製品カタログを完全な CRUD 操作で管理します。",
        "admin.inventory.title": "在庫管理",
        "admin.inventory.description": "製品の在庫数量・評価額・調整を追跡・管理します。",
        "admin.categories.title": "カテゴリ管理",
        "admin.categories.description": "多言語名称・別名・説明付きで製品カテゴリを管理します。",
        "admin.coupons.title": "クーポンとギフトカード",
        "admin.coupons.description": "プロモーションコードとプリペイド残高。",
        "admin.shippingRates.title": "送料",
        "admin.shippingRates.description": "目的地別運賃表。予想日数は見積参考値であり、配送期間を保証するものではありません。",
        "admin.taxRates.title": "税率管理",
        "admin.taxRates.description": "カート決済時に適用する目的国/地域の税率。",
        "admin.taxRates.rate": "税率",
        "admin.taxRates.col_rate": "税率",
        "admin.status_options.draft": "下書き",
        "admin.status_options.sent": "送信済み",
        "admin.status_options.paid": "支払済み",
        "admin.status_options.overdue": "期限超過",
        "admin.status_options.cancelled": "キャンセル済み",
        "admin.status_options.active": "有効",
        "admin.status_options.inactive": "無効",
        "admin.docLabels.proforma_invoice": "プロフォーマインボイス",
        "admin.docLabels.commercial_invoice": "コマーシャルインボイス",
        "admin.docLabels.bill_of_lading": "船荷証券",
        "admin.docLabels.sales_contract": "販売契約",
        "admin.docLabels.packing_list": "梱包明細",
        "admin.docLabels.cert_of_origin": "原産地証明書",
        "admin.docLabels.health_cert": "衛生証明書",
        "admin.docLabels.total_net_weight": "総正味重量 (kg)",
        "admin.docLabels.total_volume": "総容積 (CBM)",
        "admin.invoices.invoice_heading": "請求書",
        "admin.invoices.proforma_heading": "プロフォーマインボイス",
        "admin.invoices.tax_heading": "税務請求書",
        "admin.invoices.credit_note_heading": "クレジットノート",
        "admin.invoices.col_invoice": "請求書",
        "admin.invoices.export_filename": "請求書",
        "admin.ai.mode_b2b": "B2Bコーディネーター",
        "admin.ai.mode_plan": "計画-実行-再計画",
        "admin.ai.mode_trade": "貿易アシスタント",
        "admin.ai.mode_chatbot": "チャットボット（任意で注文）",
        "admin.ai.mode_recommend": "製品レコメンド",
        "admin.ai.mode_search": "AI検索",
        "admin.ai.mode_analyze_inquiry": "お問い合わせ分析",
        "admin.ai.mode_generate_quotation": "見積生成",
        "admin.ai.stream_stopped": "[ストリーム停止]",
        "admin.oem.type_quick_odm": "クイックODM",
        "admin.oem.type_full_oem": "フルOEM",
        "admin.xlsx.description": "ビジネスレポートのエクスポートと AI によるスプレッドシート翻訳。",
        "admin.xlsx.translate_hint": ".xlsx ファイルをアップロードすると、テキストセルが Trade Agent（translate_content）経由で翻訳されます。",
        "admin.xlsx.inventory_link": "AI 列マッピング付きの製品取り込みは、在庫管理 → XLSX 取り込みをご利用ください。",
        "admin.products.locale_tab_zh": "中国語",
        "skip_to_content": "メインコンテンツへスキップ",
        "errors.api.generic_failed": "操作に失敗しました。しばらくしてからもう一度お試しください。",
        "enum.product_status.active": "販売中",
        "enum.product_status.inactive": "非公開",
        "enum.product_status.draft": "下書き",
        "enum.product_status.unknown": "不明",
        "enum.user_status.active": "有効",
        "enum.user_status.inactive": "無効",
        "enum.user_status.suspended": "停止中",
        "enum.user_status.pending": "保留中",
        "enum.user_status.deleted": "削除済み",
        "enum.user_status.unknown": "不明",
        "enum.order_status.pending": "保留中",
        "enum.order_status.confirmed": "確認済み",
        "enum.order_status.production": "生産中",
        "enum.order_status.shipped": "出荷済み",
        "enum.order_status.delivered": "配達済み",
        "enum.order_status.cancelled": "キャンセル済み",
        "enum.order_status.processing": "処理中",
        "enum.order_status.pending_confirmation": "確認待ち",
        "enum.order_status.pending_approval": "承認待ち",
        "enum.order_status.partially_shipped": "一部出荷",
        "enum.order_status.partially_delivered": "一部配達",
        "enum.order_status.partially_returned": "一部返品",
        "enum.order_status.returned": "返品済み",
        "enum.order_status.expired": "期限切れ",
        "enum.order_status.unknown": "不明",
        "enum.kyb_status.approved": "承認済み",
        "enum.sample_status.approved": "承認済み",
        "enum.compliance_status.warnings": "警告",
        "languages.zh": "中国語",
        "languages.zh_short": "中",
        "languages.ko": "韓国語",
        "languages.ja": "日本語",
        "rte.ordered_list": "番号付きリスト",
        "a11y.dark_mode": "ダークモードに切り替え",
        "a11y.light_mode": "ライトモードに切り替え",
    },
    "ko": {
        "admin.brand": "CandyPro 관리자",
        "admin.logout": "로그아웃",
        "admin.theme.dark": "다크 모드",
        "admin.theme.light": "라이트 모드",
        "admin.langLabel": "한국어",
        "admin.langLabelShort": "한",
        "admin.langTitle": "한국어",
        "admin.nav.dashboard": "대시보드",
        "admin.nav.users": "사용자 관리",
        "admin.nav.inquiries": "문의 관리",
        "admin.nav.orders": "주문 관리",
        "admin.nav.returns": "반품 관리",
        "admin.nav.products": "제품 관리",
        "admin.nav.categories": "카테고리 관리",
        "admin.nav.companies": "기업 관리",
        "admin.nav.trades": "무역 관리",
        "admin.nav.pricing": "가격 관리",
        "admin.nav.oemProjects": "OEM 프로젝트",
        "admin.nav.content": "콘텐츠 관리",
        "admin.nav.inventory": "재고 관리",
        "admin.nav.xlsx": "XLSX 도구",
        "admin.nav.shipments": "배송 관리",
        "admin.nav.invoices": "송장 관리",
        "admin.nav.management": "운영 관리",
        "admin.nav.superadmin": "슈퍼 관리자",
        "admin.nav.analytics": "데이터 분석",
        "admin.nav.financial": "재무 관리",
        "admin.nav.certifications": "인증 관리",
        "admin.nav.staff": "직원 관리",
        "admin.nav.auditLog": "감사 로그",
        "admin.nav.settings": "시스템 설정",
        "admin.nav.translations": "번역 관리",
        "admin.nav.ai": "AI 앱",
        "admin.nav.organizations": "구매 조직",
        "admin.nav.channels": "판매 채널",
        "admin.nav.webhooks": "Webhook",
        "admin.nav.hooks": "훅 및 이벤트",
        "admin.nav.coupons": "쿠폰",
        "admin.nav.shipping_rates": "배송비·세율",
        "admin.nav.tax_rates": "세율 관리",
        "admin.dashboard.title": "대시보드",
        "admin.dashboard.subtitle": "플랫폼 개요와 주요 지표를 한눈에 확인하세요.",
        "admin.dashboard.loading": "대시보드 통계를 불러오는 중...",
        "admin.dashboard.error": "대시보드 로드에 실패했습니다",
        "admin.dashboard.total_users": "총 사용자",
        "admin.dashboard.total_orders": "총 주문",
        "admin.dashboard.total_inquiries": "총 문의",
        "admin.dashboard.total_revenue": "총 매출",
        "admin.dashboard.this_week": "이번 주",
        "admin.dashboard.new_users_this_week": "이번 주 신규 {count}명",
        "admin.dashboard.pending": "대기 중",
        "admin.dashboard.recent_activity": "최근 활동",
        "admin.dashboard.col_activity": "활동",
        "admin.dashboard.col_time": "시간",
        "admin.dashboard.no_activity": "최근 활동이 없습니다",
        "admin.dashboard.revenue_this_month": "이번 달 매출",
        "admin.dashboard.active_customers": "활성 고객",
        "admin.dashboard.avg_order_value": "평균 주문 금액",
        "admin.dashboard.conversion_rate": "전환율",
        "admin.dashboard.revenue_chart": "매출 추이(30일)",
        "admin.dashboard.revenue": "매출",
        "admin.dashboard.date": "날짜",
        "admin.products.title": "제품 관리",
        "admin.products.description": "제품 카탈로그를 완전한 CRUD 작업으로 관리합니다.",
        "admin.inventory.title": "재고 관리",
        "admin.inventory.description": "제품 재고 수준, 가치, 조정을 추적하고 관리합니다.",
        "admin.categories.title": "카테고리 관리",
        "admin.categories.description": "다국어 이름, 별칭, 설명으로 제품 카테고리를 관리합니다.",
        "admin.coupons.title": "쿠폰 및 기프트 카드",
        "admin.coupons.description": "프로모션 코드 및 선불 크레딧.",
        "admin.shippingRates.title": "배송비",
        "admin.shippingRates.description": "목적지별 운임표. 예상 일수는 견적 참고용이며 보장 배송 기간이 아닙니다.",
        "admin.taxRates.title": "세율 관리",
        "admin.taxRates.description": "장바구니 결제 시 적용되는 목적 국가/지역 세율.",
        "admin.taxRates.rate": "세율",
        "admin.taxRates.col_rate": "세율",
        "admin.status_options.draft": "초안",
        "admin.status_options.sent": "발송됨",
        "admin.status_options.paid": "결제 완료",
        "admin.status_options.overdue": "연체",
        "admin.status_options.cancelled": "취소됨",
        "admin.status_options.active": "활성",
        "admin.status_options.inactive": "비활성",
        "admin.docLabels.proforma_invoice": "견적 송장",
        "admin.docLabels.commercial_invoice": "상업 송장",
        "admin.docLabels.bill_of_lading": "선하증권",
        "admin.docLabels.sales_contract": "판매 계약",
        "admin.docLabels.packing_list": "포장 명세서",
        "admin.docLabels.cert_of_origin": "원산지 증명서",
        "admin.docLabels.health_cert": "위생 증명서",
        "admin.docLabels.total_net_weight": "총 순중량 (kg)",
        "admin.docLabels.total_volume": "총 용적 (CBM)",
        "admin.invoices.invoice_heading": "송장",
        "admin.invoices.proforma_heading": "견적 송장",
        "admin.invoices.tax_heading": "세금계산서",
        "admin.invoices.credit_note_heading": "대변표",
        "admin.invoices.col_invoice": "송장",
        "admin.invoices.export_filename": "송장",
        "admin.ai.mode_b2b": "B2B 코디네이터",
        "admin.ai.mode_plan": "계획-실행-재계획",
        "admin.ai.mode_trade": "무역 어시스턴트",
        "admin.ai.mode_chatbot": "챗봇(선택적 주문)",
        "admin.ai.mode_recommend": "제품 추천",
        "admin.ai.mode_search": "AI 검색",
        "admin.ai.mode_analyze_inquiry": "문의 분석",
        "admin.ai.mode_generate_quotation": "견적 생성",
        "admin.ai.stream_stopped": "[스트림 중지됨]",
        "admin.oem.type_quick_odm": "퀵 ODM",
        "admin.oem.type_full_oem": "풀 OEM",
        "admin.xlsx.description": "비즈니스 보고서 내보내기 및 AI 스프레드시트 번역.",
        "admin.xlsx.translate_hint": ".xlsx 파일을 업로드하면 텍스트 셀이 Trade Agent(translate_content)를 통해 번역됩니다.",
        "admin.xlsx.inventory_link": "AI 열 매핑이 포함된 제품 가져오기는 재고 관리 → XLSX 가져오기를 사용하세요.",
        "admin.products.locale_tab_zh": "중국어",
        "skip_to_content": "본문으로 건너뛰기",
        "errors.api.generic_failed": "작업에 실패했습니다. 잠시 후 다시 시도해 주세요.",
        "enum.product_status.active": "판매 중",
        "enum.product_status.inactive": "비공개",
        "enum.product_status.draft": "초안",
        "enum.product_status.unknown": "알 수 없음",
        "enum.user_status.active": "정상",
        "enum.user_status.inactive": "비활성",
        "enum.user_status.suspended": "정지",
        "enum.user_status.pending": "승인 대기",
        "enum.user_status.deleted": "삭제됨",
        "enum.user_status.unknown": "알 수 없음",
        "enum.order_status.pending": "대기 중",
        "enum.order_status.confirmed": "확인됨",
        "enum.order_status.production": "생산 중",
        "enum.order_status.shipped": "배송됨",
        "enum.order_status.delivered": "배달 완료",
        "enum.order_status.cancelled": "취소됨",
        "enum.order_status.processing": "처리 중",
        "enum.order_status.pending_confirmation": "확인 대기",
        "enum.order_status.pending_approval": "승인 대기",
        "enum.order_status.partially_shipped": "부분 배송",
        "enum.order_status.partially_delivered": "부분 배달",
        "enum.order_status.partially_returned": "부분 반품",
        "enum.order_status.returned": "반품됨",
        "enum.order_status.expired": "만료됨",
        "enum.order_status.unknown": "알 수 없음",
        "enum.kyb_status.approved": "승인됨",
        "enum.sample_status.approved": "승인됨",
        "enum.compliance_status.warnings": "경고",
        "languages.zh": "중국어",
        "languages.zh_short": "中",
        "languages.ko": "한국어",
        "languages.ja": "일본어",
        "rte.ordered_list": "번호 목록",
        "a11y.dark_mode": "다크 모드로 전환",
        "a11y.light_mode": "라이트 모드로 전환",
    },
}

_cache_lock = threading.Lock()
_save_counter = 0
all_caches: dict[str, dict[str, str]] = {}


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


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def save_json(path: Path, data: dict) -> None:
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def load_cache() -> dict[str, dict[str, str]]:
    if CACHE_PATH.exists():
        return json.loads(CACHE_PATH.read_text(encoding="utf-8"))
    return {loc: {} for loc in LOCALES}


def save_cache(cache: dict[str, dict[str, str]]) -> None:
    CACHE_PATH.write_text(json.dumps(cache, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def translate_one(text: str, locale: str) -> str:
    tgt = LANG_TARGETS[locale]
    protected, replacements = protect(text)
    translator = GoogleTranslator(source="zh-CN", target=tgt)
    for attempt in range(5):
        try:
            result = translator.translate(protected)
            return restore(result or text, replacements)
        except Exception:
            time.sleep(1.2 * (attempt + 1))
    return text


def translate_strings(strings: list[str], locale: str, cache: dict[str, str]) -> None:
    global _save_counter
    todo = [s for s in strings if s not in cache]
    total = len(todo)
    if total == 0:
        print(f"  {locale}: all {len(strings)} cached", flush=True)
        return

    print(f"  {locale}: translating {total} new strings ({len(strings) - total} cached)...", flush=True)
    tgt = LANG_TARGETS[locale]
    translator = GoogleTranslator(source="zh-CN", target=tgt)
    batch_size = 40
    done = 0

    for i in range(0, total, batch_size):
        batch = todo[i : i + batch_size]
        protected_batch: list[str] = []
        repl_batch: list[list[tuple[str, str]]] = []
        for text in batch:
            protected, replacements = protect(text)
            protected_batch.append(protected)
            repl_batch.append(replacements)
        results = batch
        for attempt in range(5):
            try:
                results = translator.translate_batch(protected_batch)
                break
            except Exception:
                time.sleep(1.5 * (attempt + 1))
        for text, protected, replacements, raw in zip(batch, protected_batch, repl_batch, results):
            out = restore(raw or text, replacements)
            cache[text] = out
        done += len(batch)
        _save_counter += len(batch)
        if _save_counter % 200 == 0 or done == total:
            save_cache(all_caches)
        if done % 200 == 0 or done == total:
            print(f"    {locale}: {done}/{total}", flush=True)
        time.sleep(0.3)


def build_locale_file(
    en_tree: dict,
    zh_flat: dict[str, str],
    value_map: dict[str, str],
    overrides: dict[str, str],
) -> dict:
    en_flat = flatten(en_tree)
    out_flat: dict[str, str] = {}

    for key, en_val in en_flat.items():
        if key in overrides:
            out_flat[key] = overrides[key]
            continue
        source = zh_flat.get(key, en_val)
        out_flat[key] = value_map.get(source, value_map.get(en_val, en_val))

    return unflatten(out_flat)


def main() -> None:
    global all_caches
    sys.stdout.reconfigure(encoding="utf-8")

    all_caches = load_cache()
    for loc in LOCALES:
        all_caches.setdefault(loc, {})

    en_data = {f: load_json(ROOT / SOURCE / f) for f in FILES}
    zh_data = {f: load_json(ROOT / REF / f) for f in FILES}
    zh_flat_all: dict[str, str] = {}
    for filename in FILES:
        zh_flat_all.update(flatten(zh_data[filename]))

    needed: set[str] = set()
    for filename in FILES:
        en_flat = flatten(en_data[filename])
        for key, en_val in en_flat.items():
            needed.add(zh_flat_all.get(key, en_val))
    print(f"Unique source strings: {len(needed)}", flush=True)

    for locale in LOCALES:
        print(f"Locale {locale}: translating...", flush=True)
        translate_strings(sorted(needed, key=len), locale, all_caches[locale])

    save_cache(all_caches)

    for locale in LOCALES:
        overrides = MANUAL_OVERRIDES.get(locale, {})
        for filename in FILES:
            existing_path = ROOT / locale / filename
            zh_flat = flatten(zh_data[filename])
            translated = build_locale_file(
                en_data[filename],
                zh_flat,
                all_caches[locale],
                overrides,
            )
            save_json(existing_path, translated)
            print(f"Wrote {locale}/{filename}", flush=True)

    print("Done.", flush=True)


if __name__ == "__main__":
    main()
