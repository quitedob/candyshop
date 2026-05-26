/**
 * 为除 zh/en 外的 customer.json 批量补齐购物车/订单费用边界 i18n
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const i18nRoot = path.join(__dirname, '..', 'i18n')

const cartInsertAfter = '"estimated_shipping":'
const cartBlockByLocale = {
  ar: {
    estimated_tax_ref: 'Reference Tax',
    estimated_shipping_ref: 'Reference Shipping',
    total_merchandise_only: 'Merchandise Subtotal',
    total_partial_estimates: 'Subtotal incl. partial estimates',
    total_with_estimates: 'Total incl. reference fees',
    fees_reference_disclaimer: 'Tax and shipping shown are system reference estimates, not a final quote. Final amounts will be confirmed on the proforma invoice (PI).',
    fees_partial_disclaimer: 'Some fees could not be estimated yet. Amounts shown are reference only; the PI is final.',
    fees_pending_disclaimer: 'Select a shipping country to see reference fees. Tax and shipping will be finalized on the PI by our sales team.',
    fees_pending_destination: 'Select shipping country first',
    fees_pending_rates: 'No matching rate — sales will confirm',
    fees_pending_weight: 'Missing weight data — sales will confirm',
    fees_computed_zero_hint: '(ref. $0)',
    incoterms_label: 'Incoterms',
    incoterms_hint: 'Affects reference shipping estimate (EXW = buyer arranges freight).',
    incoterms_exw_hint: 'EXW: reference shipping is $0; buyer arranges export pickup.',
    incoterms_fob_hint: 'FOB (default): reference includes freight to export port.',
    checkout_pricing_section: 'Amount at checkout',
    checkout_pricing_notice_reference: 'Tax and shipping are reference values; the PI is final.',
    checkout_pricing_notice_partial: 'Some fees were not estimated at checkout; the PI is final.',
    checkout_pricing_notice_pending: 'Tax and shipping were not fully estimated; sales will confirm on the PI.',
    checkout_pricing_subtotal: 'Merchandise subtotal',
    checkout_pricing_tax: 'Tax',
    checkout_pricing_shipping: 'Shipping',
    checkout_pricing_total: 'Order total'
  },
  id: null,
  ja: null,
  ko: {
    estimated_tax_ref: '참고 세금',
    estimated_shipping_ref: '참고 배송비',
    total_merchandise_only: '상품 소계',
    total_partial_estimates: '일부 참고 비용 포함',
    total_with_estimates: '참고 비용 포함 합계',
    fees_reference_disclaimer: '표시된 세금/배송비는 시스템 참고 추정치이며 최종 견적이 아닙니다. 최종 금액은 PI(Proforma Invoice)에서 확인됩니다.',
    fees_partial_disclaimer: '일부 비용은 아직 추정되지 않았습니다. 표시 금액은 참고용이며 PI가 최종입니다.',
    fees_pending_disclaimer: '배송 국가를 선택하면 참고 비용을 확인할 수 있습니다. 세금과 배송비는 PI에서 영업팀이 최종 확인합니다.',
    fees_pending_destination: '먼저 배송 국가를 선택하세요',
    fees_pending_rates: '일치하는 요율 없음 — 영업팀 확인',
    fees_pending_weight: '중량 데이터 없음 — 영업팀 확인',
    fees_computed_zero_hint: '(참고 $0)',
    incoterms_label: '인코텀즈',
    incoterms_hint: 'Affects reference shipping estimate (EXW = buyer arranges freight).',
    incoterms_exw_hint: 'EXW: reference shipping is $0; buyer arranges export pickup.',
    incoterms_fob_hint: 'FOB (default): reference includes freight to export port.',
    checkout_pricing_section: '결제 시 금액',
    checkout_pricing_notice_reference: '세금/배송비는 참고값이며 PI가 최종입니다.',
    checkout_pricing_notice_partial: '일부 비용이 추정되지 않았습니다. PI가 최종입니다.',
    checkout_pricing_notice_pending: '세금/배송비가 완전히 추정되지 않았습니다. PI에서 영업팀이 확인합니다.',
    checkout_pricing_subtotal: '상품 소계',
    checkout_pricing_tax: '세금',
    checkout_pricing_shipping: '배송비',
    checkout_pricing_total: '주문 합계'
  },
  ms: null,
  th: {
    estimated_tax_ref: 'ภาษีอ้างอิง',
    estimated_shipping_ref: 'ค่าจัดส่งอ้างอิง',
    total_merchandise_only: 'ยอดสินค้า',
    total_partial_estimates: 'รวมค่าอ้างอิงบางส่วน',
    total_with_estimates: 'รวมค่าอ้างอิง',
    fees_reference_disclaimer: 'ภาษี/ค่าจัดส่งที่แสดงเป็นการประมาณอ้างอิง ไม่ใช่ใบเสนอราคาสุดท้าย ยอดสุดท้ายจะยืนยันใน PI',
    fees_partial_disclaimer: 'ยังประมาณค่าใช้จ่ายบางส่วนไม่ได้ ยอดที่แสดงเป็นการอ้างอิง PI เป็นที่สุดท้าย',
    fees_pending_disclaimer: 'เลือกประเทศจัดส่งเพื่อดูค่าอ้างอิง ภาษีและค่าจัดส่งจะยืนยันใน PI',
    fees_pending_destination: 'เลือกประเทศจัดส่งก่อน',
    fees_pending_rates: 'ไม่มีอัตราที่ตรง — รอทีมขายยืนยัน',
    fees_pending_weight: 'ไม่มีข้อมูลน้ำหนัก — รอทีมขายยืนยัน',
    fees_computed_zero_hint: '(อ้างอิง $0)',
    incoterms_label: 'Incoterms',
    incoterms_hint: 'มีผลต่อการประมาณค่าจัดส่ง (EXW = ผู้ซื้อจัดการขนส่ง)',
    incoterms_exw_hint: 'EXW: ค่าจัดส่งอ้างอิง $0 ผู้ซื้อรับสินค้าที่โรงงาน',
    incoterms_fob_hint: 'FOB (ค่าเริ่มต้น): รวมค่าขนส่งถึงท่าส่งออก',
    checkout_pricing_section: 'ยอดตอนสั่งซื้อ',
    checkout_pricing_notice_reference: 'ภาษี/ค่าจัดส่งเป็นค่าอ้างอิง PI เป็นที่สุดท้าย',
    checkout_pricing_notice_partial: 'ประมาณค่าไม่ครบ PI เป็นที่สุดท้าย',
    checkout_pricing_notice_pending: 'ยังประมาณภาษี/ค่าจัดส่งไม่ครบ ทีมขายจะยืนยันใน PI',
    checkout_pricing_subtotal: 'ยอดสินค้า',
    checkout_pricing_tax: 'ภาษี',
    checkout_pricing_shipping: 'ค่าจัดส่ง',
    checkout_pricing_total: 'ยอดรวม'
  },
  vi: null,
  ja: {
    estimated_tax_ref: '参考税金',
    estimated_shipping_ref: '参考送料',
    total_merchandise_only: '商品小計',
    total_partial_estimates: '一部参考費用込み',
    total_with_estimates: '参考費用込み合計',
    fees_reference_disclaimer: '表示の税金・送料はシステム参考見積であり最終見積ではありません。最終金額はPI（Proforma Invoice）で確定します。',
    fees_partial_disclaimer: '一部費用は未見積です。表示額は参考値で、PIが最終です。',
    fees_pending_disclaimer: '配送国を選択すると参考費用を表示します。税金・送料はPIで営業が最終確認します。',
    fees_pending_destination: '配送国を先に選択してください',
    fees_pending_rates: '該当レートなし — 営業確認',
    fees_pending_weight: '重量データなし — 営業確認',
    fees_computed_zero_hint: '（参考 $0）',
    incoterms_label: 'Incoterms',
    incoterms_hint: '参考送料見積に影響（EXW=買主手配）',
    incoterms_exw_hint: 'EXW: 参考送料$0、工場渡し',
    incoterms_fob_hint: 'FOB（既定）: 輸出港までの参考運賃込み',
    checkout_pricing_section: '注文時の金額',
    checkout_pricing_notice_reference: '税金・送料は参考値。PIが最終です。',
    checkout_pricing_notice_partial: '一部未見積。PIが最終です。',
    checkout_pricing_notice_pending: '税金・送料は未見積。PIで営業が確定します。',
    checkout_pricing_subtotal: '商品小計',
    checkout_pricing_tax: '税金',
    checkout_pricing_shipping: '送料',
    checkout_pricing_total: '注文合計'
  }
}

// id/ms/vi/ar fallback to English-style keys
for (const loc of ['id', 'ms', 'vi', 'ar']) {
  if (!cartBlockByLocale[loc]) cartBlockByLocale[loc] = { ...cartBlockByLocale.ar }
}

const ordersBlockByLocale = {
  zh: null,
  en: null,
  ko: {
    pricing_subtotal: '상품 소계',
    pricing_tax: '세금',
    pricing_shipping: '배송비',
    pricing_pi_notice: '아래 금액은 주문 시 시스템 스냅샷입니다. PI(Proforma Invoice)가 최종입니다.',
    pricing_pi_confirmed: 'PI 확인 금액',
    pricing_snapshot_reference: '주문 스냅샷 (참고)',
    pricing_view_pi: 'PI 보기',
    pricing_pi_total: 'PI 합계'
  },
  th: {
    pricing_subtotal: 'ยอดสินค้า',
    pricing_tax: 'ภาษี',
    pricing_shipping: 'ค่าจัดส่ง',
    pricing_pi_notice: 'ยอดด้านล่างเป็นสแนปช็อตตอนสั่งซื้อ PI เป็นที่สุดท้าย',
    pricing_pi_confirmed: 'ยอดยืนยัน PI',
    pricing_snapshot_reference: 'สแนปช็อตคำสั่ง (อ้างอิง)',
    pricing_view_pi: 'ดู PI',
    pricing_pi_total: 'ยอด PI'
  },
  ja: {
    pricing_subtotal: '商品小計',
    pricing_tax: '税金',
    pricing_shipping: '送料',
    pricing_pi_notice: '以下は注文時のスナップショット。PIが最終です。',
    pricing_pi_confirmed: 'PI確定金額',
    pricing_snapshot_reference: '注文スナップショット（参考）',
    pricing_view_pi: 'PIを表示',
    pricing_pi_total: 'PI合計'
  }
}

for (const loc of ['id', 'ms', 'vi', 'ar']) {
  ordersBlockByLocale[loc] = {
    pricing_subtotal: 'Merchandise subtotal',
    pricing_tax: 'Tax',
    pricing_shipping: 'Shipping',
    pricing_pi_notice: 'Amounts below are a system snapshot at order placement. The proforma invoice (PI) is final.',
    pricing_pi_confirmed: 'PI confirmed amount',
    pricing_snapshot_reference: 'Order snapshot (reference)',
    pricing_view_pi: 'View PI',
    pricing_pi_total: 'PI total'
  }
}

function insertCartKeys(root, block) {
  const cart = root?.customer?.cart
  if (!cart) return false
  if (cart.fees_reference_disclaimer) return false
  for (const [k, v] of Object.entries(block)) {
    if (cart[k] === undefined) cart[k] = v
  }
  return true
}

function insertOrdersKeys(root, block) {
  if (!block) return false
  const orders = root?.customer?.orders
  if (!orders) return false
  if (orders.pricing_pi_confirmed) return false
  for (const [k, v] of Object.entries(block)) {
    if (orders[k] === undefined) orders[k] = v
  }
  return true
}

const locales = ['ar', 'id', 'ja', 'ko', 'ms', 'th', 'vi']
for (const loc of locales) {
  const file = path.join(i18nRoot, loc, 'customer.json')
  const raw = fs.readFileSync(file, 'utf8')
  const obj = JSON.parse(raw)
  const cartChanged = insertCartKeys(obj, cartBlockByLocale[loc])
  const ordersChanged = insertOrdersKeys(obj, ordersBlockByLocale[loc])
  if (cartChanged || ordersChanged) {
    fs.writeFileSync(file, JSON.stringify(obj, null, 2) + '\n', 'utf8')
    console.log('patched', loc)
  } else {
    console.log('skip', loc)
  }
}

// zh/en orders + cart extra keys
for (const loc of ['zh', 'en']) {
  const file = path.join(i18nRoot, loc, 'customer.json')
  const obj = JSON.parse(fs.readFileSync(file, 'utf8'))
  const zhEnCart = loc === 'zh' ? {
    incoterms_label: '贸易术语 (Incoterms)',
    incoterms_hint: '影响参考运费估算（EXW 表示买方自提，参考运费为 $0）',
    incoterms_exw_hint: 'EXW：参考运费 $0，买方在工厂提货',
    incoterms_fob_hint: 'FOB（默认）：含至出口港的参考运费',
    checkout_pricing_section: '下单金额',
    checkout_pricing_notice_reference: '税费/运费为参考值，最终以 PI 为准。',
    checkout_pricing_notice_partial: '部分费用下单时未能估算，最终以 PI 为准。',
    checkout_pricing_notice_pending: '税费/运费尚未完整估算，销售将在 PI 中确认。',
    checkout_pricing_subtotal: '商品小计',
    checkout_pricing_tax: '税费',
    checkout_pricing_shipping: '运费',
    checkout_pricing_total: '订单合计'
  } : {
    incoterms_label: 'Incoterms',
    incoterms_hint: 'Affects reference shipping estimate (EXW = buyer arranges freight).',
    incoterms_exw_hint: 'EXW: reference shipping is $0; buyer arranges export pickup.',
    incoterms_fob_hint: 'FOB (default): reference includes freight to export port.',
    checkout_pricing_section: 'Amount at checkout',
    checkout_pricing_notice_reference: 'Tax and shipping are reference values; the PI is final.',
    checkout_pricing_notice_partial: 'Some fees were not estimated at checkout; the PI is final.',
    checkout_pricing_notice_pending: 'Tax and shipping were not fully estimated; sales will confirm on the PI.',
    checkout_pricing_subtotal: 'Merchandise subtotal',
    checkout_pricing_tax: 'Tax',
    checkout_pricing_shipping: 'Shipping',
    checkout_pricing_total: 'Order total'
  }
  const zhEnOrders = loc === 'zh' ? {
    pricing_pi_confirmed: 'PI 确认金额',
    pricing_snapshot_reference: '下单快照（仅供参考）',
    pricing_view_pi: '查看形式发票',
    pricing_pi_total: 'PI 合计'
  } : {
    pricing_pi_confirmed: 'PI confirmed amount',
    pricing_snapshot_reference: 'Order snapshot (reference)',
    pricing_view_pi: 'View proforma invoice',
    pricing_pi_total: 'PI total'
  }
  insertCartKeys(obj, zhEnCart)
  insertOrdersKeys(obj, zhEnOrders)
  fs.writeFileSync(file, JSON.stringify(obj, null, 2) + '\n', 'utf8')
  console.log('patched', loc, 'extras')
}
