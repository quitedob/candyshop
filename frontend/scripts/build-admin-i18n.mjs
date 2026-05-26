#!/usr/bin/env node
/**
 * Generate fully translated ja/ko admin.json and patch common.json.
 * Source of truth: en/admin.json keys; semantic reference: zh/admin.json.
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  COMMON_PATCHES,
  EXPLICIT,
  PRIORITY_SECTIONS,
  PRIORITY_ZH,
} from './build-admin-i18n-data.mjs'
import { EXACT_ZH_EXTRA, translateZhTerms } from './build-admin-i18n-zh-terms.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const I18N = path.join(__dirname, '..', 'i18n')
const LOCALES = ['ja', 'ko']

const TECHNICAL_ONLY = new RegExp(
  String.raw`^(MOQ|OEM|AI|XLSX|URL|Slug|SKU|HTTP|JSON|WhatsApp|Incoterms|3PL|Halal|DOCX|USD|Ctrl\+K|Mac|⌘K|PI|CI|SC|B\/L|PL|CBM|DHL|FedEx|UPS|USPS|EMS|RCEP|Facebook|LinkedIn|Instagram|YouTube|Twitter|ID|EN|KO|JA|N\/A|TBD|g|kg|mm|kJ|kcal|aʷ|Webhook|Webhooks|Hooks|B2B|ODM|GMO|Kosher|Vegan|FORM E|CandyPro|Admin|Super Admin|GTIN|HS|HACCP|ISO22000|BRC|HALAL|KYB|MD|CBM|FOB|RCEP|FORM E|Trade Agent|translate_content|Chatbot|Agent|Diff|Enter|Mac|Ctrl|JSON|ISO|PI No\.|CI No\.|B\/L No\.|Contract No\.|PL No\.|Cert No\.|Quick ODM|Full OEM|Plan-Execute-Replan|B2B Coordinator|Trade Assistant|Product Recommend|AI Search|Analyze Inquiry|Generate Quotation|Stream stopped|Tool completed|Proforma Invoice|Commercial Invoice|Bill of Lading|Sales Contract|Packing List|Certificate of Origin|Health Certificate|Credit Note|TAX INVOICE|PROFORMA INVOICE|INVOICE|CREDIT NOTE|Self-fulfill|3PL|Marketplace|Webstore|Distributor|Quick ODM|Full OEM|RFM)$`,
  'i',
)

/** Simplified-Chinese leakage into ja/ko output. */
const SIMPLIFIED_MARKERS = /[显示加载删除编辑创建更新筛选确认暂无全部这户询览载条页别称缩图]/u

/** @param {unknown} v */
function looksEnglish(v) {
  if (typeof v !== 'string' || !/[A-Za-z]/.test(v)) return false
  if (TECHNICAL_ONLY.test(v.trim())) return false
  return /\b[A-Za-z]{3,}\b/.test(v)
}

/** @param {unknown} v */
function hasSimplifiedChinese(v) {
  return typeof v === 'string' && SIMPLIFIED_MARKERS.test(v)
}

/** @param {Record<string, unknown>} obj @param {string} [prefix] */
function flatten(obj, prefix = '', out = {}) {
  for (const [key, value] of Object.entries(obj)) {
    const pathKey = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      flatten(value, pathKey, out)
    } else {
      out[pathKey] = value
    }
  }
  return out
}

/** @param {Record<string, string>} flat */
function unflatten(flat) {
  /** @type {Record<string, unknown>} */
  const root = {}
  for (const [pathKey, value] of Object.entries(flat)) {
    const parts = pathKey.split('.')
    let cur = root
    for (let i = 0; i < parts.length - 1; i++) {
      cur[parts[i]] = cur[parts[i]] ?? {}
      cur = cur[parts[i]]
    }
    cur[parts.at(-1)] = value
  }
  return root
}

/** @param {Array<[string, string, string]>} triples @param {Locale} locale */
function mapForLocale(triples, locale) {
  const idx = locale === 'ja' ? 1 : 2
  return triples
    .slice()
    .sort((a, b) => b[0].length - a[0].length)
    .map(([en, ja, ko]) => [en, locale === 'ja' ? ja : ko])
}

/** @typedef {'ja'|'ko'} Locale */

/** @type {Array<[string, string, string]>} */
const PHRASE_TRIPLES = [
  ['Manage product catalog with complete CRUD operations.', '製品カタログを完全な CRUD 操作で管理します。', '제품 카탈로그를 완전한 CRUD 작업으로 관리합니다.'],
  ['Manage product categories with multilingual names, aliases, and descriptions.', '多言語名称・別名・説明付きで製品カテゴリを管理します。', '다국어 이름, 별칭, 설명으로 제품 카테고리를 관리합니다.'],
  ['Promotional codes and prepaid credits.', 'プロモーションコードとプリペイド残高。', '프로모션 코드 및 선불 크레딧.'],
  ['Destination-based freight rate tables. Estimated days are reference values for quoting — not guaranteed transit times.', '目的地別運賃表。予想日数は見積参考値であり、配送期間を保証するものではありません。', '목적지별 운임표. 예상 일수는 견적 참고용이며 보장 배송 기간이 아닙니다.'],
  ['Destination country/region tax rates applied at cart checkout.', 'カート決済時に適用する目的国/地域の税率。', '장바구니 결제 시 적용되는 목적 국가/지역 세율.'],
  ['Track and manage product stock levels, values, and adjustments.', '製品の在庫数量・評価額・調整を追跡・管理します。', '제품 재고 수준, 가치, 조정을 추적하고 관리합니다.'],
  ['Payment must be confirmed (paid or partial) before moving to execution stages (confirmed, production, shipped, delivered).', '実行段階（確認済み、生産中、出荷済み、配達済み）に進む前に、支払い（全額または一部）の確認が必要です。', '실행 단계(확인됨, 생산 중, 배송됨, 배달 완료)로 진행하기 전에 결제(전액 또는 일부) 확인이 필요합니다.'],
  ['Reference transit estimate for internal quoting. Subject to change; confirm at order time.', '社内見積用の参考輸送日数。変更される場合があります。注文時に確認してください。', '내부 견적용 참고 운송 일수. 변경될 수 있으며 주문 시 확인하세요.'],
  ['Decimal fraction, e.g. 0.08875 = 8.875%', '小数値。例：0.08875 = 8.875%', '소수 값. 예: 0.08875 = 8.875%'],
  ['Chinese category name is required', '中国語のカテゴリ名は必須です', '중국어 카테고리 이름은 필수입니다'],
  ['All customer-facing text lives here only. Chinese is the source locale; scalar fields sync on save.', '顧客向けテキストはすべてここでのみ管理します。中国語がソース言語で、保存時にスカラーフィールドへ同期されます。', '고객 대면 텍스트는 여기에서만 관리합니다. 중국어가 소스 로케일이며 저장 시 스칼라 필드와 동기화됩니다.'],
  ['Category name, alias, and description are maintained here only. Chinese is the source locale.', 'カテゴリ名・別名・説明はここでのみ管理します。中国語がソース言語です。', '카테고리 이름, 별칭, 설명은 여기에서만 관리합니다. 중국어가 소스 로케일입니다.'],
  ['Enter after sales verification for catalog display. You may state a specific reference (e.g. "25 business days after order confirmation") — do not use vague promises like "approx. X weeks". Leave blank or use "confirmed at order time" until verified.', '販売確認後にカタログ表示用として入力します。具体的な参照（例：「注文確認後25営業日」）を記載でき、「約X週」などの曖昧な表現は使用しないでください。未確認の場合は空欄または「注文時に確認」。', '판매 확인 후 카탈로그 표시용으로 입력합니다. 구체적 기준(예: "주문 확인 후 25영업일")을 명시하고 "약 X주" 같은 모호한 표현은 사용하지 마세요. 확인 전에는 비워두거나 "주문 시 확인".'],
  ['Sample lead time reference — confirmed by sales in the sample quote; do not use "approx. X weeks".', 'サンプル参考リードタイム — サンプル見積で営業が確認。 「約X週」は使用しないでください。', '샘플 참고 납기 — 샘플 견적에서 영업이 확인. "약 X주" 사용 금지.'],
  ['Search by product name or SKU...', '製品名または SKU で検索...', '제품명 또는 SKU로 검색...'],
  ['Batch translate missing locales', '未翻訳ロケールを一括翻訳', '누락된 로케일 일괄 번역'],
  ['Published (default)', '公開済み（デフォルト）', '게시됨(기본)'],
  ['All statuses', 'すべてのステータス', '모든 상태'],
  ['Add Product', '製品を追加', '제품 추가'],
  ['Edit Product', '製品を編集', '제품 편집'],
  ['Create Product', '製品を作成', '제품 생성'],
  ['Base Price (USD)', '基本価格 (USD)', '기본 가격 (USD)'],
  ['Lead Time', 'リードタイム', '납기'],
  ['Stock Quantity', '在庫数量', '재고 수량'],
  ['Add Category', 'カテゴリを追加', '카테고리 추가'],
  ['Edit Category', 'カテゴリを編集', '카테고리 편집'],
  ['Create Category', 'カテゴリを作成', '카테고리 생성'],
  ['No categories found', 'カテゴリが見つかりません', '카테고리가 없습니다'],
  ['Category created', 'カテゴリを作成しました', '카테고리가 생성되었습니다'],
  ['Category updated', 'カテゴリを更新しました', '카테고리가 업데이트되었습니다'],
  ['Category deleted', 'カテゴリを削除しました', '카테고리가 삭제되었습니다'],
  ['Failed to load categories', 'カテゴリの読み込みに失敗しました', '카테고리 로드 실패'],
  ['Create Coupon', 'クーポンを作成', '쿠폰 생성'],
  ['Create Gift Card', 'ギフトカードを作成', '기프트 카드 생성'],
  ['Gift Cards', 'ギフトカード', '기프트 카드'],
  ['Coupons & Gift Cards', 'クーポンとギフトカード', '쿠폰 및 기프트 카드'],
  ['No coupons.', 'クーポンがありません。', '쿠폰이 없습니다.'],
  ['No gift cards.', 'ギフトカードがありません。', '기프트 카드가 없습니다.'],
  ['Fixed Amount', '固定金額', '고정 금액'],
  ['Min Order', '最低注文額', '최소 주문'],
  ['Starts At', '開始日時', '시작일'],
  ['Expires At', '有効期限', '만료일'],
  ['Add Rate', '料金を追加', '요금 추가'],
  ['Add Tax Rate', '税率を追加', '세율 추가'],
  ['No shipping rates.', '送料設定がありません。', '배송비 요금이 없습니다.'],
  ['No tax rates configured.', '税率が設定されていません。', '세율이 구성되지 않았습니다.'],
  ['Min Weight (kg)', '最小重量 (kg)', '최소 중량 (kg)'],
  ['Max Weight (kg)', '最大重量 (kg)', '최대 중량 (kg)'],
  ['Base Cost', '基本料金', '기본 비용'],
  ['Cost per kg', 'kg あたりの料金', 'kg당 비용'],
  ['Est. Days', '予想日数', '예상 일수'],
  ['Country (ISO)', '国 (ISO)', '국가 (ISO)'],
  ['Region / State', '地域 / 州', '지역 / 주'],
  ['Adjust Stock', '在庫調整', '재고 조정'],
  ['Total Products', '製品総数', '총 제품'],
  ['Low Stock', '在庫不足', '재고 부족'],
  ['Out of Stock', '在庫切れ', '품절'],
  ['In Stock', '在庫あり', '재고 있음'],
  ['Total Value', '総評価額', '총 가치'],
  ['Filter by stock status', '在庫ステータスで絞り込み', '재고 상태별 필터'],
  ['Filter by category', 'カテゴリで絞り込み', '카테고리별 필터'],
  ['All Categories', 'すべてのカテゴリ', '모든 카테고리'],
  ['Unit Value', '単価', '단가'],
  ['Adjust Stock', '在庫調整', '재고 조정'],
  ['Current Stock', '現在庫', '현재 재고'],
  ['New Stock', '新在庫', '새 재고'],
  ['Adjustment Type', '調整タイプ', '조정 유형'],
  ['Set To', '設定', '설정'],
  ['Subtract', '減算', '차감'],
  ['Restock', '補充', '재입고'],
  ['Correction', '修正', '수정'],
  ['Damage/Loss', '破損/ロス', '파손/손실'],
  ['Customer Return', '顧客返品', '고객 반품'],
  ['Inventory Audit', '在庫監査', '재고 실사'],
  ['Stock History', '在庫履歴', '재고 이력'],
  ['Import XLSX', 'XLSX 取り込み', 'XLSX 가져오기'],
  ['Export Selected', '選択項目をエクスポート', '선택 항목 내보내기'],
  ['Batch Edit', '一括編集', '일괄 편집'],
  ['Batch Delete', '一括削除', '일괄 삭제'],
  ['Add Warehouse', '倉庫を追加', '창고 추가'],
  ['Create Transfer', '移管を作成', '이동 생성'],
  ['From Warehouse', '出庫倉庫', '출발 창고'],
  ['To Warehouse', '入庫倉庫', '도착 창고'],
  ['Dark mode', 'ダークモード', '다크 모드'],
  ['Light mode', 'ライトモード', '라이트 모드'],
  ['Dashboard', 'ダッシュボード', '대시보드'],
  ['Management', '管理', '관리'],
  ['Categories', 'カテゴリ', '카테고리'],
  ['Coupons', 'クーポン', '쿠폰'],
  ['Organizations', '組織', '조직'],
  ['Channels', 'チャネル', '채널'],
  ['Webhooks', 'Webhook', 'Webhook'],
  ['Hooks & Events', 'フックとイベント', '훅 및 이벤트'],
  ['Shipping & Tax', '送料・税', '배송비·세금'],
  ['Tax Rates', '税率', '세율'],
  ['Shipping Rates', '送料', '배송비'],
  ['Loading...', '読み込み中...', '로딩 중...'],
  ['Saving...', '保存中...', '저장 중...'],
  ['Creating...', '作成中...', '생성 중...'],
  ['Updating...', '更新中...', '업데이트 중...'],
  ['Deleting...', '削除中...', '삭제 중...'],
  ['Translating…', '翻訳中…', '번역 중…'],
  ['Exporting…', 'エクスポート中…', '내보내는 중…'],
  ['No data available', 'データがありません', '데이터 없음'],
  ['No data found.', 'データが見つかりません。', '데이터를 찾을 수 없습니다.'],
  ['Save failed', '保存に失敗しました', '저장 실패'],
  ['Load failed', '読み込みに失敗しました', '로드 실패'],
  ['Delete failed', '削除に失敗しました', '삭제 실패'],
  ['Confirm delete?', '削除しますか？', '삭제하시겠습니까?'],
  ['Confirm delete this content?', 'このコンテンツを削除しますか？', '이 콘텐츠를 삭제하시겠습니까?'],
  ['Previous', '前へ', '이전'],
  ['Next', '次へ', '다음'],
  ['Cancel', 'キャンセル', '취소'],
  ['Save', '保存', '저장'],
  ['Edit', '編集', '편집'],
  ['Delete', '削除', '삭제'],
  ['Create', '作成', '생성'],
  ['Update', '更新', '업데이트'],
  ['View', '表示', '보기'],
  ['Search', '検索', '검색'],
  ['Export', 'エクスポート', '내보내기'],
  ['Actions', '操作', '작업'],
  ['Status', 'ステータス', '상태'],
  ['Name', '名称', '이름'],
  ['Description', '説明', '설명'],
  ['Title', 'タイトル', '제목'],
  ['Type', 'タイプ', '유형'],
  ['Amount', '金額', '금액'],
  ['Currency', '通貨', '통화'],
  ['Customer', '顧客', '고객'],
  ['Product', '製品', '제품'],
  ['Products', '製品', '제품'],
  ['Category', 'カテゴリ', '카테고리'],
  ['Order', '注文', '주문'],
  ['Orders', '注文', '주문'],
  ['Company', '会社', '회사'],
  ['Email', 'メール', '이메일'],
  ['Phone', '電話', '전화'],
  ['Password', 'パスワード', '비밀번호'],
  ['Notes', '備考', '메모'],
  ['Date', '日付', '날짜'],
  ['Quantity', '数量', '수량'],
  ['Qty', '数量', '수량'],
  ['Total', '合計', '합계'],
  ['Subtotal', '小計', '소계'],
  ['Tax', '税', '세금'],
  ['Shipping', '送料', '배송'],
  ['Pending', '保留中', '대기 중'],
  ['Active', '有効', '활성'],
  ['Inactive', '無効', '비활성'],
  ['Draft', '下書き', '초안'],
  ['Confirmed', '確認済み', '확인됨'],
  ['Completed', '完了', '완료'],
  ['Cancelled', 'キャンセル済み', '취소됨'],
  ['Approved', '承認済み', '승인됨'],
  ['Rejected', '却下', '거절됨'],
  ['Paid', '支払済み', '결제 완료'],
  ['Overdue', '期限超過', '연체'],
  ['Sent', '送信済み', '발송됨'],
  ['Percentage', 'パーセント', '비율'],
  ['Balance', '残高', '잔액'],
  ['Code', 'コード', '코드'],
  ['Value', '値', '값'],
  ['Used', '使用済み', '사용됨'],
  ['Initial', '初期', '초기'],
  ['Carrier', '配送業者', '운송사'],
  ['Destination', '目的地', '목적지'],
  ['Warehouse', '倉庫', '창고'],
  ['Transfer', '移管', '이동'],
  ['Reason', '理由', '사유'],
  ['Other', 'その他', '기타'],
  ['Success', '成功', '성공'],
  ['Failed', '失敗', '실패'],
  ['Error', 'エラー', '오류'],
  ['Warning', '警告', '경고'],
  ['Required', '必須', '필수'],
  ['Optional', '任意', '선택'],
  ['All', 'すべて', '전체'],
  ['Filter', 'フィルター', '필터'],
  ['Showing {from} to {to} of {total}', '{total} 件中 {from} ～ {to} 件を表示', '전체 {total}건 중 {from}–{to} 표시'],
  ['Showing {from} to {to} of {total} results', '{total} 件中 {from} ～ {to} 件を表示', '전체 {total}건 중 {from}–{to} 표시'],
  ['{total} total', '合計 {total} 件', '총 {total}건'],
  ['Buyer Organizations', '購買組織', '구매 조직'],
  ['Sales Channels', '販売チャネル', '판매 채널'],
  ['Event Hooks', 'イベントフック', '이벤트 훅'],
  ['Outbound event notifications.', 'アウトバウンドイベント通知。', '아웃바운드 이벤트 알림.'],
  ['Plugin and internal hook configurations.', 'プラグインおよび内部フック設定。', '플러그인 및 내부 훅 구성.'],
  ['B2B org credit limits and approval tiers.', 'B2B 組織の与信限度と承認階層。', 'B2B 조직 신용 한도 및 승인 단계.'],
  ['Manage marketplace and webstore channels.', 'マーケットプレイスとウェブストアチャネルを管理。', '마켓플레이스 및 웹스토어 채널 관리.'],
  ['Add Channel', 'チャネルを追加', '채널 추가'],
  ['Add Webhook', 'Webhook を追加', 'Webhook 추가'],
  ['Add Hook', 'フックを追加', '훅 추가'],
  ['Add Organization', '組織を追加', '조직 추가'],
  ['No channels yet.', 'チャネルがありません。', '채널이 없습니다.'],
  ['No webhooks configured.', 'Webhook が設定されていません。', 'Webhook이 구성되지 않았습니다.'],
  ['No hooks configured.', 'フックが設定されていません。', '훅이 구성되지 않았습니다.'],
  ['No organizations.', '組織がありません。', '조직이 없습니다.'],
  ['Approval Threshold', '承認しきい値', '승인 임계값'],
  ['Payment Terms', '支払条件', '결제 조건'],
  ['Credit Limit', '与信限度', '신용 한도'],
  ['Price Multiplier', '価格倍率', '가격 배수'],
  ['Self-fulfill', '自社配送', '자체 이행'],
  ['Select at least one event.', 'イベントを1つ以上選択してください。', '이벤트를 하나 이상 선택하세요.'],
  ['Secret (save now)', 'シークレット（今すぐ保存）', '시크릿(지금 저장)'],
  ['Invalid JSON config.', 'JSON 設定が無効です。', 'JSON 구성이 유효하지 않습니다.'],
  ['Config (JSON)', '設定 (JSON)', '구성 (JSON)'],
  ['Event Name', 'イベント名', '이벤트 이름'],
  ['Delivery logs', '配信ログ', '전달 로그'],
  ['Configurations', '設定', '구성'],
  ['Deliveries', '配信', '전달'],
  ['Executions', '実行', '실행'],
  ['Events', 'イベント', '이벤트'],
  ['Hooks', 'フック', '훅'],
  ['Members', 'メンバー', '멤버'],
  ['Add Member', 'メンバーを追加', '멤버 추가'],
  ['Remove', '削除', '제거'],
  ['Buyer', '購買担当', '구매자'],
  ['Approver', '承認者', '승인자'],
  ['Review and process customer return requests.', '顧客の返品リクエストを審査・処理します。', '고객 반품 요청을 검토하고 처리합니다.'],
  ['All Statuses', 'すべてのステータス', '모든 상태'],
  ['Return ID', '返品 ID', '반품 ID'],
  ['Reject Return', '返品を拒否', '반품 거절'],
  ['Rejection reason', '拒否理由', '거절 사유'],
  ['Return approved.', '返品を承認しました。', '반품이 승인되었습니다.'],
  ['Return received.', '返品を受領しました。', '반품이 접수되었습니다.'],
  ['Refund processed.', '返金を処理しました。', '환불이 처리되었습니다.'],
  ['Return rejected.', '返品を拒否しました。', '반품이 거절되었습니다.'],
  ['Trade BOL Tracking', '貿易 BOL 追跡', '무역 BOL 추적'],
  ['Manage trade bill-of-lading and cross-border logistics.', '貿易船荷証券と越境物流を管理します。', '무역 선하증권 및 국경 간 물류를 관리합니다.'],
  ['New BOL', '新規 BOL', '새 BOL'],
  ['Trade ID', '貿易 ID', '무역 ID'],
  ['Translation Management', '翻訳管理', '번역 관리'],
  ['Add Translation', '翻訳を追加', '번역 추가'],
  ['Search key or value...', 'キーまたは値を検索...', '키 또는 값 검색...'],
  ['All Groups', 'すべてのグループ', '모든 그룹'],
  ['All Locales', 'すべてのロケール', '모든 로케일'],
  ['No translations found', '翻訳が見つかりません', '번역 없음'],
  ['Create Translation', '翻訳を作成', '번역 생성'],
  ['Key', 'キー', '키'],
  ['Locale', 'ロケール', '로케일'],
  ['Group', 'グループ', '그룹'],
  ['Entity Type', 'エンティティタイプ', '엔티티 유형'],
  ['Entity ID', 'エンティティ ID', '엔티티 ID'],
  ['IP Address', 'IP アドレス', 'IP 주소'],
  ['Audit Log', '監査ログ', '감사 로그'],
  ['System Settings', 'システム設定', '시스템 설정'],
  ['Configure platform settings and preferences.', 'プラットフォーム設定と環境設定を構成します。', '플랫폼 설정 및 환경설정을 구성합니다.'],
  ['No settings found for this category.', 'このカテゴリの設定項目がありません。', '이 카테고리에 설정 항목이 없습니다.'],
  ['Setting saved successfully.', '設定を保存しました。', '설정이 저장되었습니다.'],
  ['General', '一般', '일반'],
  ['Notifications', '通知', '알림'],
  ['Staff Management', 'スタッフ管理', '직원 관리'],
  ['Financial Overview', '財務概要', '재무 개요'],
  ['Analytics & Reports', '分析とレポート', '분석 및 보고서'],
  ['Business Reports', 'ビジネスレポート', '비즈니스 보고서'],
  ['OEM Management', 'OEM 管理', 'OEM 관리'],
  ['XLSX Tools', 'XLSX ツール', 'XLSX 도구'],
  ['AI Applications', 'AI アプリケーション', 'AI 애플리케이션'],
  ['Clear Chat', 'チャットをクリア', '채팅 지우기'],
  ['Send', '送信', '전송'],
  ['Stop', '停止', '중지'],
  ['Approve', '承認', '승인'],
  ['Reject', '拒否', '거절'],
  ['Model', 'モデル', '모델'],
  ['Temperature', '温度', '온도'],
  ['Agent Mode', 'Agent モード', 'Agent 모드'],
  ['Pending Approvals', '承認待ち', '승인 대기'],
  ['Tool Call History', 'ツール呼び出し履歴', '도구 호출 기록'],
  ['Configuration', '設定', '구성'],
  ['Order context', '注文コンテキスト', '주문 컨텍스트'],
  ['No order (general chat)', '注文なし（一般チャット）', '주문 없음(일반 채팅)'],
  ['Proforma Invoice', 'プロフォーマインボイス', '프로forma 송장'],
  ['Commercial Invoice', '商業インボイス', '상업 송장'],
  ['Bill of Lading', '船荷証券', '선하증권'],
  ['Sales Contract', '売買契約', '매매 계약'],
  ['Packing List', '梱包明細書', '포장 명세서'],
  ['Certificate of Origin', '原産地証明書', '원산지 증명서'],
  ['Health Certificate', '衛生証明書', '위생 증명서'],
  ['Buyer Name', '買主名', '구매자명'],
  ['Seller Name', '売主名', '판매자명'],
  ['Exporter Name', '輸出者名', '수출자명'],
  ['Importer Name', '輸入者名', '수입자명'],
  ['Bank Details', '銀行詳細', '은행 정보'],
  ['Shipper', '荷送人', '송하인'],
  ['Consignee', '荷受人', '수하인'],
  ['Notify Party', '通知先', '통지처'],
  ['Port of Loading', '積出港', '적재항'],
  ['Port of Discharge', '荷揚港', '양하항'],
  ['Freight Terms', '運賃条件', '운임 조건'],
  ['Goods Description', '品名', '품목 설명'],
  ['Gross Weight (kg)', '総重量 (kg)', '총중량 (kg)'],
  ['Measurement (CBM)', '容積 (CBM)', '용적 (CBM)'],
  ['No. of Packages', '梱包数', '포장 개수'],
  ['Delivery Date', '納期', '납기일'],
  ['Special Terms', '特記事項', '특별 조건'],
  ['Country of Origin', '原産国', '원산국'],
  ['Destination Country', '仕向国', '목적국'],
  ['Inspection Result', '検査結果', '검사 결과'],
  ['Batch Number', 'ロット番号', '배치 번호'],
  ['Production Date', '製造日', '생산일'],
  ['Expiry Date', '有効期限', '유효기간'],
  ['Issuing Authority', '発行機関', '발급 기관'],
  ['Shipping Mark', '荷印', '운송 표시'],
  ['Thumbnail URL', 'サムネイル URL', '썸네일 URL'],
  ['Icon', 'アイコン', '아이콘'],
  ['Alias', '別名', '별칭'],
  ['Featured', 'おすすめ', '추천'],
  ['Summary', '概要', '요약'],
  ['Ingredients', '原材料', '원재료'],
  ['Allergens', 'アレルゲン', '알레르겐'],
  ['Shelf Life', '賞味期限', '유통기한'],
  ['Storage', '保管条件', '보관'],
  ['Upload Image', '画像をアップロード', '이미지 업로드'],
  ['Upload Images', '画像をアップロード', '이미지 업로드'],
  ['Generate with AI', 'AI で生成', 'AI로 생성'],
  ['AI Translate', 'AI 翻訳', 'AI 번역'],
  ['AI Import', 'AI 取り込み', 'AI 가져오기'],
  ['Weight & Measurement', '重量と計量', '중량 및 계량'],
  ['Product Dimensions (mm)', '製品寸法 (mm)', '제품 치수 (mm)'],
  ['Dietary Labels', '食事ラベル', '식이 라벨'],
  ['Nutrition (per 100g)', '栄養成分 (100g あたり)', '영양 (100g당)'],
  ['Ingredient Compliance', '成分コンプライアンス', '성분 규정 준수'],
  ['Trade & Packaging', '貿易と包装', '무역 및 포장'],
  ['Sample Specs', 'サンプル仕様', '샘플 사양'],
  ['Market Profile', '市場プロファイル', '시장 프로필'],
  ['Cost Stack', 'コスト構成', '비용 구조'],
  ['Material', '材料', '재료'],
  ['Labor', '人件費', '인건비'],
  ['Logistics', '物流', '물류'],
  ['Flows', 'フロー', '플로우'],
  ['Solutions', 'ソリューション', '솔루션'],
  ['Projects', 'プロジェクト', '프로젝트'],
  ['New Flow', '新規フロー', '새 플로우'],
  ['New Solution', '新規ソリューション', '새 솔루션'],
  ['Quick ODM', 'クイック ODM', '퀵 ODM'],
  ['Full OEM', 'フル OEM', '풀 OEM'],
  ['Marketplace', 'マーケットプレイス', '마켓플레이스'],
  ['Webstore', 'ウェブストア', '웹스토어'],
  ['Distributor', 'ディストリビューター', '유통업체'],
  ['Data Export', 'データエクスポート', '데이터 내보내기'],
  ['Product catalog', '製品カタログ', '제품 카탈로그'],
  ['Revenue report', '売上レポート', '매출 보고서'],
  ['Trade transactions', '貿易取引', '무역 거래'],
  ['Customers', '顧客', '고객'],
  ['Download started', 'ダウンロードを開始しました', '다운로드 시작됨'],
  ['Export failed', 'エクスポートに失敗しました', '내보내기 실패'],
  ['Translation failed', '翻訳に失敗しました', '번역 실패'],
  ['Please select a .xlsx file', '.xlsx ファイルを選択してください', '.xlsx 파일을 선택하세요'],
  ['Select at least one target language', '対象言語を1つ以上選択してください', '대상 언어를 하나 이상 선택하세요'],
  ['Auto detect', '自動検出', '자동 감지'],
  ['Target language', '対象言語', '대상 언어'],
  ['Target languages', '対象言語', '대상 언어'],
  ['Source language (optional)', 'ソース言語（任意）', '원본 언어(선택)'],
  ['Single language', '単一言語', '단일 언어'],
  ['Multiple languages', '複数言語', '다국어'],
  ['Translate & download', '翻訳してダウンロード', '번역 및 다운로드'],
  ['Choose .xlsx file', '.xlsx ファイルを選択', '.xlsx 파일 선택'],
  ['Selected: {name}', '選択：{name}', '선택: {name}'],
  ['Order confirmed. Inform customer and proceed to production.', '注文が確認されました。顧客に連絡し、生産に進んでください。', '주문이 확인되었습니다. 고객에게 알리고 생산을 진행하세요.'],
  ['Order is in production. Ensure materials are available.', '注文は生産中です。資材が揃っていることを確認してください。', '주문이 생산 중입니다. 자재가 준비되었는지 확인하세요.'],
  ['Order shipped. Update tracking information.', '注文が出荷されました。追跡情報を更新してください。', '주문이 배송되었습니다. 추적 정보를 업데이트하세요.'],
  ['Order delivered. Request feedback from customer.', '注文が配達されました。顧客からフィードバックを依頼してください。', '주문이 배달되었습니다. 고객 피드백을 요청하세요.'],
  ['Order cancelled. Process refund if payment was made.', '注文がキャンセルされました。支払い済みの場合は返金処理を行ってください。', '주문이 취소되었습니다. 결제가 있었다면 환불을 처리하세요.'],
  ['RFM Analysis', 'RFM分析', 'RFM 분석'],
  ['OEM Needed', 'OEM 必要', 'OEM 필요'],
  ['OEM Needed', 'OEM 対応', 'OEM 필요'],
  ['URL or upload', 'URL またはアップロード', 'URL 또는 업로드'],
  ['Author Avatar URL', '著者アバター URL', '작성자 아바타 URL'],
  ['Images (comma separated URLs)', '画像（URL をカンマ区切り）', '이미지(쉼표로 구분된 URL)'],
  ['Slug is required', 'Slug は必須です', 'Slug는 필수입니다'],
  ['Manage factory certifications (HACCP, ISO22000, BRC, HALAL, etc.)', '工場認証を管理（HACCP、ISO22000、BRC、HALAL など）', '공장 인증 관리(HACCP, ISO22000, BRC, HALAL 등)'],
  ['Manage company KYB verification, credit limits, and pricing tiers.', '企業 KYB 認証、与信限度、価格体系を管理します。', '기업 KYB 인증, 신용 한도, 가격 체계를 관리합니다.'],
  ['Manage OEM project lifecycle from inquiry to delivery.', 'お問い合わせから納品までの OEM プロジェクトライフサイクルを管理します。', '문의부터 납품까지 OEM 프로젝트 수명 주기를 관리합니다.'],
  ['Back to OEM Projects', 'OEM プロジェクト一覧へ戻る', 'OEM 프로젝트 목록으로'],
  ['Type a message, press Enter to send...', 'メッセージを入力し、Enter で送信...', '메시지를 입력하고 Enter로 전송...'],
  ['Generate a {type} for this trade transaction.', 'この貿易取引の {type} を生成します。', '이 무역 거래의 {type}을(를) 생성합니다.'],
  ['Export XLSX', 'XLSX エクスポート', 'XLSX 내보내기'],
  ['Select text in title, excerpt, or body, then press Ctrl+K (Mac: ⌘K) for AI edit with diff preview — Accept or Reject.', 'タイトル、抜粋、本文でテキストを選択し、Ctrl+K（Mac：⌘K）で AI 編集（Diff プレビュー）— 承認または拒否。', '제목, 요약, 본문에서 텍스트를 선택한 뒤 Ctrl+K(Mac: ⌘K)로 AI 편집(Diff 미리보기) — 수락 또는 거절.'],
  ['Select text in the body or pick a scope, then describe the fix (Cursor-style partial edit).', '本文でテキストを選択するか範囲を選び、修正内容を記述します（Cursor 式部分編集）。', '본문에서 텍스트를 선택하거나 범위를 고른 뒤 수정 방법을 설명하세요(Cursor 스타일 부분 편집).'],
  ['e.g., 150g milk chocolate bar with almonds, vegan, 12-month shelf life, individually flow-wrapped, 24 pieces per display box, halal certified...', '例：150g アーモンド入りミルクチョコレート、ヴィーガン、賞味期限12か月、個包装、ディスプレイボックス24個入り、Halal 認証...', '예: 150g 아몬드 밀크 초콜릿, 비건, 유통기한 12개월, 개별 포장, 디스플레이 박스 24개, Halal 인증...'],
  ['Optional. Overrides thumbnail for Facebook/LinkedIn previews. 1200×630 recommended.', '任意。Facebook/LinkedIn プレビュー用にサムネイルを上書き。1200×630 推奨。', '선택. Facebook/LinkedIn 미리보기용 썸네일 덮어쓰기. 1200×630 권장.'],
  ['Search by product name or SKU...', '製品名または SKU で検索...', '제품명 또는 SKU로 검색...'],
  ['Category Slug', 'カテゴリ Slug', '카테고리 Slug'],
  ['Energy (kcal)', 'エネルギー (kcal)', '에너지 (kcal)'],
  ['GTIN / Barcode', 'GTIN / バーコード', 'GTIN / 바코드'],
  ['Showing {from} to {to} of {total}', '{total} 件中 {from} ～ {to} 件を表示', '전체 {total}건 중 {from}–{to} 표시'],
  ['Sales Velocity', '販売速度', '판매 속도'],
  ['Customer Churn', '顧客離反', '고객 이탈'],
  ['Inventory Health', '在庫健全性', '재고 건강도'],
  ['Profit & Loss', '損益', '손익'],
  ['Replenishment', '補充', '보충'],
  ['Inquiry Conversion', 'お問い合わせ転換', '문의 전환'],
  ['Loading report...', 'レポートを読み込み中...', '보고서 로딩 중...'],
  ['Failed to load report', 'レポートの読み込みに失敗しました', '보고서 로드 실패'],
  ['Velocity', '速度', '속도'],
  ['Recency', '最新購入', '최근 구매'],
  ['Frequency', '購入頻度', '구매 빈도'],
  ['Monetary', '購入金額', '구매 금액'],
  ['Segment', 'セグメント', '세그먼트'],
  ['Days Inactive', '非アクティブ日数', '미활동 일수'],
  ['Reorder Qty', '補充数量', '보충 수량'],
  ['Converted', '転換', '전환'],
  ['Rate', '転換率', '전환율'],
  ['Business intelligence dashboards with revenue, order, and conversion metrics.', '収益、注文、コンバージョン指標を含む BI ダッシュボード。', '매출, 주문, 전환율 지표를 제공하는 BI 대시보드.'],
  ['Loading analytics...', '分析データを読み込み中...', '분석 데이터를 불러오는 중...'],
  ['Error loading analytics data', '分析データの読み込みエラー', '분석 데이터 로드 오류'],
  ['No data available for the selected period.', '選択した期間にデータがありません。', '선택한 기간에 데이터가 없습니다.'],
  ['Revenue Trends', '売上トレンド', '매출 추세'],
  ['Order Trends', '注文トレンド', '주문 추세'],
  ['Top Products', '人気製品', '인기 제품'],
  ['3 Months', '3か月', '3개월'],
  ['6 Months', '6か月', '6개월'],
  ['12 Months', '12か月', '12개월'],
  ['Accounts Receivable', '売掛金', '매출채권'],
  ['Overdue Amount', '期限超過金額', '연체 금액'],
  ['Revenue This Month', '今月の売上', '이번 달 매출'],
  ['Outstanding Invoices', '未払い請求書', '미결 송장'],
  ['Payment Breakdown', '支払内訳', '결제 내역'],
  ['Due Soon', '期限間近', '만기 임박'],
  ['Invoice No', '請求書番号', '송장 번호'],
  ['Due Date', '支払期日', '만기일'],
  ['No outstanding invoices.', '未払い請求書はありません。', '미결 송장이 없습니다.'],
  ['No financial data available.', '財務データがありません。', '재무 데이터가 없습니다.'],
  ['Manage inquiry records with complete CRUD and conversion actions.', '完全な CRUD とコンバージョン操作でお問い合わせを管理します。', '전체 CRUD 및 전환 작업으로 문의 기록을 관리합니다.'],
  ['Manage order lifecycle with complete CRUD operations.', '完全な CRUD 操作で注文ライフサイクルを管理します。', '전체 CRUD 작업으로 주문 생명주기를 관리합니다.'],
  ['New User', '新規ユーザー', '새 사용자'],
  ['Loading users...', 'ユーザーを読み込み中...', '사용자를 불러오는 중...'],
  ['No users found.', 'ユーザーが見つかりません。', '사용자를 찾을 수 없습니다.'],
  ['First Name', '名', '이름'],
  ['Last Name', '姓', '성'],
  ['Superadmin', 'スーパー管理者', '슈퍼 관리자'],
  ['New Inquiry', '新規お問い合わせ', '새 문의'],
  ['Loading inquiries...', 'お問い合わせを読み込み中...', '문의를 불러오는 중...'],
  ['No inquiries found.', 'お問い合わせが見つかりません。', '문의를 찾을 수 없습니다.'],
  ['Company & Contact', '会社と連絡先', '회사 및 담당자'],
  ['Contact Person', '担当者', '담당자'],
  ['Target Country', '対象国', '대상 국가'],
  ['Estimated Quantity', '見積数量', '예상 수량'],
  ['Expected Delivery', '希望納期', '예상 납기'],
  ['Priority', '優先度', '우선순위'],
  ['Customer Notes', '顧客メモ', '고객 메모'],
  ['Submit Quote', '見積を送信', '견적 제출'],
  ['Submitting...', '送信中...', '제출 중...'],
  ['New Order', '新規注文', '새 주문'],
  ['Loading orders...', '注文を読み込み中...', '주문을 불러오는 중...'],
  ['No orders found.', '注文が見つかりません。', '주문을 찾을 수 없습니다.'],
  ['Payment Status', '支払ステータス', '결제 상태'],
  ['Tracking Number', '追跡番号', '운송장 번호'],
  ['Order Items', '注文明細', '주문 품목'],
  ['Add Item', '明細を追加', '품목 추가'],
  ['Unit Price', '単価', '단가'],
  ['Specifications', '仕様', '사양'],
  ['Shipping Address', '配送先住所', '배송 주소'],
  ['Street', '番地', '도로명'],
  ['City', '市区町村', '도시'],
  ['State', '都道府県', '주/도'],
  ['Zip Code', '郵便番号', '우편번호'],
  ['Country', '国', '국가'],
  ['Filter by user...', 'ユーザーで絞り込み...', '사용자로 필터...'],
  ['From', '開始', '시작'],
  ['To', '終了', '종료'],
  ['Placed on', '注文日', '주문일'],
  ['No items found', '明細がありません', '품목 없음'],
  ['Update Status', 'ステータスを更新', '상태 업데이트'],
  ['Dismiss', '閉じる', '닫기'],
  ['Compliance verified', 'コンプライアンス確認済み', '규정 준수 확인됨'],
  ['Manual compliance ack', '手動コンプライアンス確認', '수동 규정 준수 확인'],
  ['Inventory warnings', '在庫警告', '재고 경고'],
  ['Payment created', '支払を作成', '결제 생성됨'],
  ['Payment confirmed', '支払を確認', '결제 확인됨'],
  ['Payment refunded', '支払を返金', '결제 환불됨'],
  ['Invoice auto-created', '請求書を自動作成', '송장 자동 생성됨'],
  ['Order Fulfillment', '注文フルフィルメント', '주문 이행'],
  ['Create Fulfillment', 'フルフィルメントを作成', '이행 생성'],
  ['Confirm Delivery', '配達を確認', '배달 확인'],
  ['No fulfillments yet.', 'フルフィルメントがありません。', '이행 기록이 없습니다.'],
  ['Fulfillment created.', 'フルフィルメントを作成しました。', '이행이 생성되었습니다.'],
  ['Marked as shipped.', '出荷済みにしました。', '출고 처리되었습니다.'],
  ['Delivery confirmed.', '配達を確認しました。', '배달이 확인되었습니다.'],
  ['Linked Product IDs', '関連製品 ID', '연결된 제품 ID'],
  ['Product IDs, comma separated', '製品 ID（カンマ区切り）', '제품 ID, 쉼표로 구분'],
  ['Structured product references for compliance checks and quoting', 'コンプライアンスチェックと見積用の構造化製品参照', '규정 준수 검사 및 견적용 구조화된 제품 참조'],
  ['Skip to main content', 'メインコンテンツへスキップ', '본문으로 바로가기'],
  ['This site works best with JavaScript enabled. Use the links below to continue browsing.', 'このサイトは JavaScript を有効にすると最適に動作します。以下のリンクから引き続き閲覧できます。', '이 사이트는 JavaScript를 사용할 때 가장 잘 작동합니다. 아래 링크로 계속 둘러보실 수 있습니다.'],
  ['Page Not Found', 'ページが見つかりません', '페이지를 찾을 수 없습니다'],
  ['Sorry, the page you are looking for doesn\'t exist.', 'お探しのページは存在しません。', '죄송합니다. 요청하신 페이지가 존재하지 않습니다.'],
  ['Back to Home', 'ホームに戻る', '홈으로 돌아가기'],
  ['Server Error', 'サーバーエラー', '서버 오류'],
  ['Something went wrong on our end. Please try again later.', 'サーバー側で問題が発生しました。しばらくしてからもう一度お試しください。', '서버에 문제가 발생했습니다. 잠시 후 다시 시도해 주세요.'],
  ['Refresh Page', 'ページを更新', '페이지 새로고침'],
  ['Access Denied', 'アクセス拒否', '접근 거부'],
  ['You do not have permission to access this page.', 'このページにアクセスする権限がありません。', '이 페이지에 접근할 권한이 없습니다.'],
  ['An error occurred', 'エラーが発生しました', '오류가 발생했습니다'],
  ['Something went wrong', '問題が発生しました', '문제가 발생했습니다'],
  ['Go to Home', 'ホームへ', '홈으로 이동'],
  ['Try Again', '再試行', '다시 시도'],
  ['Try again', '再試行', '다시 시도'],
  ['Something went wrong. Please try again.', '問題が発生しました。もう一度お試しください。', '문제가 발생했습니다. 다시 시도해 주세요.'],
  ['Could not load data. Please try again.', 'データを読み込めませんでした。もう一度お試しください。', '데이터를 불러올 수 없습니다. 다시 시도해 주세요.'],
  ['Online payment failed. Please try again.', 'オンライン決済に失敗しました。もう一度お試しください。', '온라인 결제에 실패했습니다. 다시 시도해 주세요.'],
  ['Pending confirmation', '確認待ち', '확인 대기'],
  ['Pending approval', '承認待ち', '승인 대기'],
  ['Partially shipped', '一部出荷', '부분 배송'],
  ['Partially delivered', '一部配達', '부분 배달'],
  ['Partially returned', '一部返品', '부분 반품'],
  ['Switch to dark mode', 'ダークモードに切り替え', '다크 모드로 전환'],
  ['Switch to light mode', 'ライトモードに切り替え', '라이트 모드로 전환'],
  ['B2B Candy Manufacturing', 'B2B キャンディ製造', 'B2B 캔디 제조'],
  ['Manage', '管理', '관리'],
  ['List', '一覧', '목록'],
  ['Overview', '概要', '개요'],
  ['Settings', '設定', '설정'],
  ['Reports', 'レポート', '보고서'],
  ['Analytics', '分析', '분석'],
  ['Financial', '財務', '재무'],
  ['Staff', 'スタッフ', '직원'],
  ['Certifications', '認証', '인증'],
  ['Content', 'コンテンツ', '콘텐츠'],
  ['Pricing', '価格', '가격'],
  ['Shipments', '出荷', '배송'],
  ['Invoices', '請求書', '송장'],
  ['Returns', '返品', '반품'],
  ['Trades', '貿易', '무역'],
  ['Companies', '企業', '기업'],
  ['Users', 'ユーザー', '사용자'],
  ['Inquiries', 'お問い合わせ', '문의'],
  ['Super Admin', 'スーパー管理者', '슈퍼 관리자'],
  ['Logout', 'ログアウト', '로그아웃'],
  ['Open sidebar menu', 'サイドバーメニューを開く', '사이드바 메뉴 열기'],
  ['Close sidebar menu', 'サイドバーメニューを閉じる', '사이드바 메뉴 닫기'],
  ['{count} new this week', '今週 {count} 件の新規', '이번 주 신규 {count}명'],
  ['this week', '今週', '이번 주'],
  ['Recent Activity', '最近のアクティビティ', '최근 활동'],
  ['Activity', 'アクティビティ', '활동'],
  ['Time', '時間', '시간'],
  ['No recent activity', '最近のアクティビティはありません', '최근 활동 없음'],
  ['Avg Order Value', '平均注文額', '평균 주문 금액'],
  ['Conversion Rate', 'コンバージョン率', '전환율'],
  ['Active Customers', 'アクティブ顧客', '활성 고객'],
  ['Total Revenue', '総売上', '총 매출'],
  ['Total Users', 'ユーザー総数', '총 사용자'],
  ['Total Orders', '注文総数', '총 주문'],
  ['Total Inquiries', 'お問い合わせ総数', '총 문의'],
  ['Error loading dashboard', 'ダッシュボードの読み込みに失敗しました', '대시보드 로드 오류'],
  ['Loading dashboard statistics...', 'ダッシュボード統計を読み込み中...', '대시보드 통계를 불러오는 중...'],
  ['Platform overview and key metrics at a glance.', 'プラットフォーム概要と主要指標を一覧で確認できます。', '플랫폼 개요와 주요 지표를 한눈에 확인하세요.'],
]

/** @type {Record<'ja'|'ko', Record<string, string>>} */
let FULL_ZH_MAP = { ja: {}, ko: {} }

const EN_JA = mapForLocale(PHRASE_TRIPLES, 'ja')
const EN_KO = mapForLocale(PHRASE_TRIPLES, 'ko')

/** Build zh->locale map: hand-crafted > existing good translations > en exact phrases. */
function buildFullZhMap(enAdmin, zhAdmin, existingAdmin, locale) {
  const fe = flatten(enAdmin)
  const fz = flatten(zhAdmin)
  const fx = flatten(existingAdmin)
  /** @type {Record<string, string>} */
  const map = { ...PRIORITY_ZH[locale], ...EXACT_ZH_EXTRA[locale] }

  for (const [pathKey, enStr] of Object.entries(fe)) {
    const zhStr = fz[pathKey]
    if (typeof zhStr !== 'string') continue

    if (map[zhStr] && !looksEnglish(map[zhStr]) && !hasSimplifiedChinese(map[zhStr])) continue

    if (PRIORITY_ZH[locale][zhStr]) {
      map[zhStr] = PRIORITY_ZH[locale][zhStr]
      continue
    }

    const existing = fx[pathKey]
    if (typeof existing === 'string' && !looksEnglish(existing) && !hasSimplifiedChinese(existing)) {
      map[zhStr] = existing
      continue
    }

    const exact = EN_EXACT[locale][String(enStr)]
    if (exact && exact !== enStr && !looksEnglish(exact)) {
      map[zhStr] = exact
      continue
    }

    const phrase = translateEnPhrase(String(enStr), locale)
    if (phrase !== enStr && !looksEnglish(phrase)) {
      map[zhStr] = phrase
      continue
    }

    const zhPhrase = translateZhTerms(zhStr, locale)
    if (zhPhrase !== zhStr && !looksEnglish(zhPhrase) && !hasSimplifiedChinese(zhPhrase)) {
      map[zhStr] = zhPhrase
    }
  }
  return map
}

/** @type {Record<'ja'|'ko', Record<string, string>>} */
const EN_EXACT = { ja: Object.fromEntries(PHRASE_TRIPLES.map(([en, ja]) => [en, ja])), ko: Object.fromEntries(PHRASE_TRIPLES.map(([en, , ko]) => [en, ko])) }

/** @param {string} en @param {Locale} locale */
function translateEnPhrase(en, locale) {
  let result = en
  const pairs = locale === 'ja' ? EN_JA : EN_KO
  for (const [from, to] of pairs) {
    if (from && result.includes(from)) result = result.split(from).join(to)
  }
  return result
}

/** @param {Record<string, unknown>} enTree @param {Record<string, unknown>} zhTree @param {Record<string, unknown>} existingTree @param {Locale} locale @param {string} [prefix] */
function buildAdminTree(enTree, zhTree, existingTree, locale, prefix = 'admin') {
  /** @type {Record<string, string>} */
  const flat = {}

  /** @param {unknown} enVal @param {unknown} zhVal @param {unknown} existingVal @param {string} path */
  function walk(enVal, zhVal, existingVal, path) {
    if (enVal && typeof enVal === 'object' && !Array.isArray(enVal)) {
      const enObj = enVal
      const zhObj = zhVal && typeof zhVal === 'object' ? zhVal : {}
      const exObj = existingVal && typeof existingVal === 'object' ? existingVal : {}
      for (const key of Object.keys(enObj)) {
        walk(enObj[key], zhObj[key], exObj[key], `${path}.${key}`)
      }
      return
    }

    const explicit = EXPLICIT[locale][path]
    if (explicit) {
      flat[path] = explicit
      return
    }

    if (
      typeof existingVal === 'string' &&
      !looksEnglish(existingVal) &&
      !hasSimplifiedChinese(existingVal)
    ) {
      flat[path] = existingVal
      return
    }

    const enStr = String(enVal ?? '')
    const zhStr = typeof zhVal === 'string' ? zhVal : undefined

    if (zhStr && FULL_ZH_MAP[locale][zhStr]) {
      flat[path] = FULL_ZH_MAP[locale][zhStr]
      return
    }

    const exact = EN_EXACT[locale][enStr]
    if (exact && exact !== enStr) {
      flat[path] = exact
      return
    }

    if (zhStr && PRIORITY_ZH[locale][zhStr]) {
      flat[path] = PRIORITY_ZH[locale][zhStr]
      return
    }

    if (zhStr) {
      const zhTerm = translateZhTerms(zhStr, locale)
      if (zhTerm !== zhStr && !looksEnglish(zhTerm) && !hasSimplifiedChinese(zhTerm)) {
        flat[path] = zhTerm
        return
      }
    }

    const phrase = translateEnPhrase(enStr, locale)
    if (phrase !== enStr && !looksEnglish(phrase)) {
      flat[path] = phrase
      return
    }

    flat[path] = typeof existingVal === 'string' && !looksEnglish(existingVal) ? existingVal : phrase
  }

  walk(enTree, zhTree, existingTree, prefix)
  return unflatten(flat)
}

/** @param {Record<string, unknown>} target @param {string} dotPath @param {string} value */
function setNested(target, dotPath, value) {
  const parts = dotPath.split('.')
  let cur = target
  for (let i = 0; i < parts.length - 1; i++) {
    cur[parts[i]] = cur[parts[i]] ?? {}
    cur = cur[parts[i]]
  }
  cur[parts.at(-1)] = value
}

/** @param {Record<string, unknown>} common @param {Locale} locale */
function patchCommon(common, locale) {
  for (const [dotPath, value] of Object.entries(COMMON_PATCHES[locale])) {
    setNested(common, dotPath, value)
  }
  return common
}

/** @param {Record<string, unknown>} enCommon @param {Record<string, unknown>} zhCommon @param {Locale} locale */
function buildCommonOut(enCommon, zhCommon, locale) {
  const fe = flatten(enCommon)
  const fz = flatten(zhCommon)
  /** @type {Record<string, string>} */
  const flat = {}
  for (const [key, enStr] of Object.entries(fe)) {
    if (COMMON_PATCHES[locale][key]) {
      flat[key] = COMMON_PATCHES[locale][key]
      continue
    }
    const zhStr = fz[key]
    if (typeof zhStr === 'string' && PRIORITY_ZH[locale][zhStr]) {
      flat[key] = PRIORITY_ZH[locale][zhStr]
      continue
    }
    if (typeof zhStr === 'string' && FULL_ZH_MAP[locale][zhStr]) {
      flat[key] = FULL_ZH_MAP[locale][zhStr]
      continue
    }
    if (typeof zhStr === 'string') {
      const zhTerm = translateZhTerms(zhStr, locale)
      if (zhTerm !== zhStr && !looksEnglish(zhTerm) && !hasSimplifiedChinese(zhTerm)) {
        flat[key] = zhTerm
        continue
      }
    }
    const phrase = translateEnPhrase(String(enStr), locale)
    flat[key] = phrase
  }
  return unflatten(flat)
}

/** @param {Record<string, unknown>} adminRoot @param {Locale} locale */
function reportEnglishKeys(adminRoot, locale) {
  const flat = flatten(adminRoot)
  return Object.entries(flat)
    .filter(([, v]) => looksEnglish(v))
    .map(([k, v]) => `${k}: ${v}`)
}

function main() {
  const enAdmin = JSON.parse(fs.readFileSync(path.join(I18N, 'en', 'admin.json'), 'utf8'))
  const zhAdmin = JSON.parse(fs.readFileSync(path.join(I18N, 'zh', 'admin.json'), 'utf8'))
  const enCommon = JSON.parse(fs.readFileSync(path.join(I18N, 'en', 'common.json'), 'utf8'))
  const zhCommon = JSON.parse(fs.readFileSync(path.join(I18N, 'zh', 'common.json'), 'utf8'))

  /** @type {Record<'ja'|'ko', Record<string, string>>} */
  const fullMaps = {
    ja: buildFullZhMap(
      enAdmin.admin,
      zhAdmin.admin,
      JSON.parse(fs.readFileSync(path.join(I18N, 'ja', 'admin.json'), 'utf8')).admin,
      'ja',
    ),
    ko: buildFullZhMap(
      enAdmin.admin,
      zhAdmin.admin,
      JSON.parse(fs.readFileSync(path.join(I18N, 'ko', 'admin.json'), 'utf8')).admin,
      'ko',
    ),
  }

  FULL_ZH_MAP = fullMaps

  /** @type {Record<string, {admin: number, common: number, english: string[]}>} */
  const stats = {}

  for (const locale of LOCALES) {
    const existingAdmin = JSON.parse(fs.readFileSync(path.join(I18N, locale, 'admin.json'), 'utf8'))
    const existingCommon = JSON.parse(fs.readFileSync(path.join(I18N, locale, 'common.json'), 'utf8'))

    const adminOut = buildAdminTree(enAdmin.admin, zhAdmin.admin, existingAdmin.admin, locale)

    const adminFlat = flatten(adminOut.admin)
    fs.writeFileSync(
      path.join(I18N, locale, 'admin.json'),
      `${JSON.stringify(adminOut, null, 2)}\n`,
      'utf8',
    )

    const commonOut = buildCommonOut(enCommon, zhCommon, locale)
    fs.writeFileSync(
      path.join(I18N, locale, 'common.json'),
      `${JSON.stringify(commonOut, null, 2)}\n`,
      'utf8',
    )

    JSON.parse(fs.readFileSync(path.join(I18N, locale, 'admin.json'), 'utf8'))
    JSON.parse(fs.readFileSync(path.join(I18N, locale, 'common.json'), 'utf8'))

    const english = reportEnglishKeys(adminOut.admin, locale)
    stats[locale] = {
      admin: Object.keys(adminFlat).length,
      common: Object.keys(flatten(commonOut)).length,
      english,
    }
    console.log(`✓ ${locale}/admin.json (${Object.keys(adminFlat).length} keys)`)
    console.log(`✓ ${locale}/common.json patched`)
    console.log(`  English-like admin keys remaining: ${english.length}`)
  }

  console.log('\n=== Summary ===')
  for (const locale of LOCALES) {
    console.log(`${locale}: admin keys=${stats[locale].admin}, common keys=${stats[locale].common}, english admin=${stats[locale].english.length}`)
    if (stats[locale].english.length) {
      console.log(`  Sample English keys (${locale}):`)
      for (const line of stats[locale].english.slice(0, 40)) console.log(`    ${line}`)
      if (stats[locale].english.length > 40) console.log(`    ... and ${stats[locale].english.length - 40} more`)
    }
  }
}

main()
